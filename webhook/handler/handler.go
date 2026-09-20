package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	p               processor
	l               logger
	maxBodySize     int64
	processTimeout  time.Duration
	sem             chan struct{}
	stateMu         sync.Mutex
	wg              sync.WaitGroup
	closing         atomic.Bool
	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc
}

func New(
	p processor,
	l logger,
	maxBodySize int64,
	maxWorkers int,
	processTimeout time.Duration,
) (*Handler, error) {
	if maxBodySize <= 0 {
		return nil, fmt.Errorf("max body size must be positive: %d", maxBodySize)
	}
	if maxWorkers <= 0 {
		return nil, fmt.Errorf("max workers must be positive: %d", maxWorkers)
	}
	if processTimeout <= 0 {
		return nil, fmt.Errorf("process timeout must be positive: %s", processTimeout)
	}

	lifecycleCtx, lifecycleCancel := context.WithCancel(context.Background())

	return &Handler{
		p:               p,
		l:               l,
		maxBodySize:     maxBodySize,
		processTimeout:  processTimeout,
		sem:             make(chan struct{}, maxWorkers),
		lifecycleCtx:    lifecycleCtx,
		lifecycleCancel: lifecycleCancel,
	}, nil
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
		h.stateMu.Lock()
		if h.closing.Load() {
			h.stateMu.Unlock()
			<-h.sem
			log.Warn("Webhook handler is shutting down")
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		h.wg.Add(1)
		h.stateMu.Unlock()
		go h.processWebhook(body, sign, r.RemoteAddr, requestID, context.WithoutCancel(ctx))

	case <-h.lifecycleCtx.Done():
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
	h.stateMu.Lock()
	h.closing.Store(true)
	h.lifecycleCancel()
	h.stateMu.Unlock()

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

	ctx, stop := context.WithCancel(baseCtx)

	shutdownStop := context.AfterFunc(h.lifecycleCtx, stop)
	defer shutdownStop()

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
