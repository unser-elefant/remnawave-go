package webhook

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/unser-elefant/remnawave-go/webhook/handler"
)

type logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	WithContext(ctx context.Context) *slog.Logger
}

type Config struct {
	ServerPort      int
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

func Run(ctx context.Context, cfg *Config, l logger, h *handler.Handler) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", h.HandleWebhook)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	serverErrCh := make(chan error, 1)

	go func() {
		l.Info("Starting server", "port", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			select {
			case serverErrCh <- err:
			case <-ctx.Done():
			}
		}
	}()

	select {
	case <-ctx.Done():
		l.Info("Shutting down webhook server...")
	case err := <-serverErrCh:
		l.Error("Webhook server failed", "error", err)
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("webhook server shutdown failed: %w", err)
	}
	if err := h.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("webhook handler shutdown failed: %w", err)
	}

	l.Info("Webhook server stopped")
	return nil
}
