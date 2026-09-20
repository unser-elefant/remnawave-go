package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unser-elefant/remnawave-go/internal/observability/requestid"
)

type testLogger struct{}

func (testLogger) WithContext(context.Context) *slog.Logger { return slog.Default() }

type processorStub struct {
	process func(ctx context.Context, body []byte, signature string, remoteAddr string) error
}

func (p processorStub) Process(ctx context.Context, body []byte, signature string, remoteAddr string) error {
	if p.process != nil {
		return p.process(ctx, body, signature, remoteAddr)
	}
	return nil
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func newTestHandler(p processor) *Handler {
	h, err := New(
		p,
		testLogger{},
		1024,
		1,
		time.Second,
	)
	if err != nil {
		panic(err)
	}
	return h
}

func TestNewHandler_UsesProvidedMaxBodySize(t *testing.T) {
	h, err := New(
		processorStub{},
		testLogger{},
		2048,
		1,
		time.Second,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(2048), h.maxBodySize)
}

func TestNewHandler_RejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		maxBodySize    int64
		maxWorkers     int
		processTimeout time.Duration
	}{
		{name: "body size", maxBodySize: 0, maxWorkers: 1, processTimeout: time.Second},
		{name: "workers", maxBodySize: 1024, maxWorkers: 0, processTimeout: time.Second},
		{name: "process timeout", maxBodySize: 1024, maxWorkers: 1, processTimeout: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(processorStub{}, testLogger{}, tt.maxBodySize, tt.maxWorkers, tt.processTimeout)
			require.Error(t, err)
		})
	}
}

func TestHandleWebhook_Success(t *testing.T) {
	called := make(chan struct{}, 1)
	h, err := New(processorStub{
		process: func(ctx context.Context, body []byte, signature string, remoteAddr string) error {
			called <- struct{}{}
			id, ok := requestid.FromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, "existing-request-id", id)
			assert.Equal(t, `{"ok":true}`, string(body))
			assert.Empty(t, signature)
			return nil
		},
	}, testLogger{}, 1024, 1, time.Second)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{"ok":true}`))
	req.Header.Set(requestid.HeaderRequestID, "existing-request-id")
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.StatusCode)
	assert.Equal(t, "existing-request-id", res.Header.Get(requestid.HeaderRequestID))
	assert.Contains(t, string(body), `"status":"accepted"`)
	require.Eventually(t, func() bool {
		select {
		case <-called:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestHandleWebhook_GeneratesRequestID(t *testing.T) {
	called := make(chan struct{}, 1)
	h, err := New(processorStub{
		process: func(ctx context.Context, body []byte, signature string, remoteAddr string) error {
			called <- struct{}{}
			id, ok := requestid.FromContext(ctx)
			assert.True(t, ok)
			assert.NotEmpty(t, id)
			assert.Empty(t, signature)
			return nil
		},
	}, testLogger{}, 1024, 1, time.Second)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(requestid.HeaderRequestID))
	require.Eventually(t, func() bool {
		select {
		case <-called:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestHandleWebhook_RequestTooLarge(t *testing.T) {
	h, err := New(processorStub{}, testLogger{}, 4, 1, time.Second)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{"toolarge":true}`))
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func TestHandleWebhook_ReadError(t *testing.T) {
	h, err := New(processorStub{}, testLogger{}, 1024, 1, time.Second)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook", http.NoBody)
	req.Body = io.NopCloser(errReader{})
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleWebhook_UsesStatusFromProcessorError(t *testing.T) {
	called := make(chan struct{}, 1)
	h, err := New(processorStub{
		process: func(ctx context.Context, body []byte, signature string, remoteAddr string) error {
			called <- struct{}{}
			return errors.New("invalid signature")
		},
	}, testLogger{}, 1024, 1, time.Second)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	require.Eventually(t, func() bool {
		select {
		case <-called:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestHandleWebhook_UsesInternalServerErrorFromGenericProcessorError(t *testing.T) {
	called := make(chan struct{}, 1)
	h, err := New(processorStub{
		process: func(ctx context.Context, body []byte, signature string, remoteAddr string) error {
			called <- struct{}{}
			return errors.New("boom")
		},
	}, testLogger{}, 1024, 1, time.Second)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	require.Eventually(t, func() bool {
		select {
		case <-called:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestShutdown_WaitsForBackgroundWork(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	h := newTestHandler(processorStub{
		process: func(ctx context.Context, body []byte, signature string, remoteAddr string) error {
			started <- struct{}{}
			<-release
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	require.Eventually(t, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- h.Shutdown(context.Background())
	}()

	select {
	case err := <-shutdownDone:
		t.Fatalf("shutdown returned before background work completed: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	select {
	case err := <-shutdownDone:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("shutdown did not wait for background work")
	}
}

func TestShutdown_CancelsBackgroundWork(t *testing.T) {
	started := make(chan struct{}, 1)
	canceled := make(chan struct{}, 1)
	h := newTestHandler(processorStub{
		process: func(ctx context.Context, body []byte, signature string, remoteAddr string) error {
			started <- struct{}{}
			<-ctx.Done()
			canceled <- struct{}{}
			return ctx.Err()
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	h.HandleWebhook(rec, req)

	require.Eventually(t, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, h.Shutdown(shutdownCtx))

	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not cancel background work")
	}
}

func TestHandleWebhook_RejectsNewWorkAfterShutdownStarts(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	h := newTestHandler(processorStub{
		process: func(ctx context.Context, body []byte, signature string, remoteAddr string) error {
			started <- struct{}{}
			<-release
			return nil
		},
	})

	firstReq := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	firstRec := httptest.NewRecorder()
	h.HandleWebhook(firstRec, firstReq)

	require.Eventually(t, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- h.Shutdown(context.Background())
	}()

	require.Eventually(t, h.closing.Load, time.Second, 10*time.Millisecond)

	secondReq := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	secondRec := httptest.NewRecorder()
	h.HandleWebhook(secondRec, secondReq)

	assert.Equal(t, http.StatusServiceUnavailable, secondRec.Code)

	close(release)

	select {
	case err := <-shutdownDone:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("shutdown did not complete in time")
	}
}
