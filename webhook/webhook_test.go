package webhook

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unser-elefant/remnawave-go/webhook/handler"
)

type testLogger struct{}

func (testLogger) Debug(string, ...any)                     {}
func (testLogger) Info(string, ...any)                      {}
func (testLogger) Warn(string, ...any)                      {}
func (testLogger) Error(string, ...any)                     {}
func (testLogger) WithContext(context.Context) *slog.Logger { return slog.Default() }

type processorStub struct{}

func (processorStub) Process(context.Context, []byte, string, string) error { return nil }

func newTestHandler() *handler.Handler {
	h, err := handler.New(processorStub{}, testLogger{}, 1024, 1, time.Second)
	if err != nil {
		panic(err)
	}
	return h
}

func TestRun_HealthAndShutdown(t *testing.T) {
	cfg := &Config{
		ServerPort:      0,
		ShutdownTimeout: time.Second,
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- Run(ctx, cfg, testLogger{}, newTestHandler()) }()

	cancel()
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("webhook.Run did not stop in time")
	}
}

func TestRun_ReturnsErrorOnInvalidAddress(t *testing.T) {
	cfg := &Config{
		ServerPort:      -1,
		ShutdownTimeout: time.Second,
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
	}

	err := Run(context.Background(), cfg, testLogger{}, newTestHandler())

	require.Error(t, err)
	assert.EqualError(t, err, "invalid server port: -1")
}

func TestRun_RejectsInvalidConfig(t *testing.T) {
	config := &Config{ShutdownTimeout: -time.Second}

	err := Run(context.Background(), config, testLogger{}, newTestHandler())

	assert.EqualError(t, err, "shutdown timeout must be positive: -1s")
}
