package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/unser-elefant/remnawave-go/internal/observability/requestid"
)

type processor interface {
	Process(ctx context.Context, body []byte, sign string, remoteAddr string) error
}

type logger interface {
	WithContext(ctx context.Context) *slog.Logger
}

type Handler struct {
	p              processor
	l              logger
	maxBodySize    int64
	stopCh         <-chan struct{}
	newBaseCtx     func(string) context.Context
	processTimeout time.Duration
	sem            chan struct{}
	wg             sync.WaitGroup
	closing        atomic.Bool
}

func New(
	p processor,
	l logger,
	maxBodySize int64,
	maxWorkers int,
	stopCh <-chan struct{},
	newBaseCtx func(string) context.Context,
	processTimeout time.Duration,
) *Handler {
	if newBaseCtx == nil {
		newBaseCtx = func(requestID string) context.Context {
			return requestid.With(context.Background(), requestID)
		}
	}

	return &Handler{
		p:              p,
		l:              l,
		maxBodySize:    maxBodySize,
		stopCh:         stopCh,
		newBaseCtx:     newBaseCtx,
		processTimeout: processTimeout,
		sem:            make(chan struct{}, maxWorkers),
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx, requestID := createRequestContext(r)
	log := h.l.WithContext(ctx)
	w.Header().Set(requestid.HeaderRequestID, requestID)

	if h.closing.Load() {
		log.Warn("Webhook handler is shutting down")
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, h.maxBodySize))
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			log.Warn("Webhook request body too large", "error", err)
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}

		log.Error("Failed to read body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	sign := r.Header.Get("X-Remnawave-Signature")

	waitCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	select {
	case h.sem <- struct{}{}:
		if h.closing.Load() {
			<-h.sem
			log.Warn("Webhook handler is shutting down")
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		h.wg.Add(1)
		go h.processWebhook(body, sign, r.RemoteAddr, requestID, h.newBaseCtx(requestID))

	case <-h.stopCh:
		log.Warn("Webhook handler is shutting down")
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return

	case <-waitCtx.Done():
		log.Warn("Too many concurrent webhook requests")
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}

func (h *Handler) Shutdown(ctx context.Context) error {
	h.closing.Store(true)

	done := make(chan struct{})

	go func() {
		defer close(done)
		h.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Handler) processWebhook(body []byte, sign, remoteAddr, requestID string, baseCtx context.Context) {
	defer h.wg.Done()
	defer func() {
		<-h.sem
	}()

	defer func() {
		if r := recover(); r != nil {
			h.l.WithContext(context.Background()).Error("Panic in webhook processor", "panic", r, "request_id", requestID)
		}
	}()

	ctx := requestid.With(baseCtx, requestID)
	ctx, cancel := context.WithTimeout(ctx, h.processTimeout)
	defer cancel()
	log := h.l.WithContext(ctx)

	if err := h.p.Process(ctx, body, sign, remoteAddr); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warn("Process webhook timeout")
			return
		}

		if errors.Is(err, context.Canceled) {
			log.Warn("Process webhook canceled")
			return
		}

		log.Error("Failed to process webhook", "error", err)
	}
}

func createRequestContext(r *http.Request) (context.Context, string) {
	ctx := r.Context()

	requestID := strings.TrimSpace(r.Header.Get(requestid.HeaderRequestID))
	if requestID == "" {
		requestID = strings.TrimSpace(r.Header.Get(requestid.HeaderCorrelationID))
	}

	if requestID == "" {
		ctx, requestID = requestid.GetOrCreate(ctx)
	} else {
		ctx = requestid.With(ctx, requestID)
	}

	return ctx, requestID
}
