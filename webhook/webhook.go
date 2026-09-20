package webhook

import (
	"context"
	"errors"
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
	if err := validateConfig(cfg); err != nil {
		return err
	}

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

	shutdownErrCh := make(chan error, 2)
	go func() {
		shutdownErrCh <- server.Shutdown(shutdownCtx)
	}()
	go func() {
		shutdownErrCh <- h.Shutdown(shutdownCtx)
	}()

	var shutdownErrs []error
	for range 2 {
		if err := <-shutdownErrCh; err != nil {
			shutdownErrs = append(shutdownErrs, err)
		}
	}
	if len(shutdownErrs) > 0 {
		return fmt.Errorf("webhook shutdown failed: %w", errors.Join(shutdownErrs...))
	}

	l.Info("Webhook server stopped")
	return nil
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("webhook config is nil")
	}
	if cfg.ServerPort < 0 || cfg.ServerPort > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.ServerPort)
	}
	if cfg.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive: %s", cfg.ShutdownTimeout)
	}
	if cfg.ReadTimeout < 0 {
		return fmt.Errorf("read timeout cannot be negative: %s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout < 0 {
		return fmt.Errorf("write timeout cannot be negative: %s", cfg.WriteTimeout)
	}
	return nil
}
