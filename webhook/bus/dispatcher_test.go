package bus

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unser-elefant/remnawave-go/webhook/payload"
)

type testLogger struct{}

func (testLogger) WithContext(context.Context) *slog.Logger { return slog.Default() }

func TestPublish_ReturnsFalseWhenNoSubscribers(t *testing.T) {
	dispatcher := NewDispatcher(testLogger{})
	hook := &payload.RemnawaveWebhook{
		Event: payload.EventUserModified,
	}

	handled, err := dispatcher.Publish(context.Background(), hook)

	assert.False(t, handled)
	assert.NoError(t, err)
}

func TestPublish_FanOutToAllSubscribers(t *testing.T) {
	dispatcher := NewDispatcher(testLogger{})
	hook := &payload.RemnawaveWebhook{
		Event: payload.EventUserModified,
	}

	var (
		mu          sync.Mutex
		calls       []string
		samePointer = true
	)

	dispatcher.Subscribe(payload.EventUserModified, func(_ context.Context, p *payload.RemnawaveWebhook) error {
		mu.Lock()
		samePointer = samePointer && p == hook
		calls = append(calls, "first")
		mu.Unlock()
		return nil
	})

	dispatcher.Subscribe(payload.EventUserModified, func(_ context.Context, p *payload.RemnawaveWebhook) error {
		mu.Lock()
		samePointer = samePointer && p == hook
		calls = append(calls, "second")
		mu.Unlock()
		return nil
	})

	handled, err := dispatcher.Publish(context.Background(), hook)

	assert.True(t, handled)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"first", "second"}, calls)
	assert.True(t, samePointer)
}

func TestPublish_RunsHandlersConcurrently(t *testing.T) {
	dispatcher := NewDispatcher(testLogger{})
	hook := &payload.RemnawaveWebhook{Event: payload.EventUserModified}

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})

	waitFor := func(ch <-chan struct{}) error {
		select {
		case <-ch:
			return nil
		case <-time.After(2 * time.Second):
			return errors.New("other handler did not run concurrently")
		}
	}

	dispatcher.Subscribe(payload.EventUserModified, func(context.Context, *payload.RemnawaveWebhook) error {
		close(firstStarted)
		return waitFor(secondStarted)
	})
	dispatcher.Subscribe(payload.EventUserModified, func(context.Context, *payload.RemnawaveWebhook) error {
		close(secondStarted)
		return waitFor(firstStarted)
	})

	handled, err := dispatcher.Publish(context.Background(), hook)

	assert.True(t, handled)
	assert.NoError(t, err)
}

func TestPublish_HandlerErrorDoesNotStopOthers(t *testing.T) {
	dispatcher := NewDispatcher(testLogger{})
	hook := &payload.RemnawaveWebhook{Event: payload.EventUserModified}

	var called atomic.Int32
	dispatcher.Subscribe(payload.EventUserModified, func(context.Context, *payload.RemnawaveWebhook) error {
		called.Add(1)
		return errors.New("first failed")
	})
	dispatcher.Subscribe(payload.EventUserModified, func(context.Context, *payload.RemnawaveWebhook) error {
		called.Add(1)
		return nil
	})
	dispatcher.Subscribe(payload.EventUserModified, func(context.Context, *payload.RemnawaveWebhook) error {
		called.Add(1)
		return errors.New("third failed")
	})

	handled, err := dispatcher.Publish(context.Background(), hook)

	assert.True(t, handled)
	require.Error(t, err)
	assert.Equal(t, "first failed\nthird failed", err.Error())
	assert.Equal(t, int32(3), called.Load())
}

func TestPublish_HandlerPanicBecomesError(t *testing.T) {
	dispatcher := NewDispatcher(testLogger{})
	hook := &payload.RemnawaveWebhook{Event: payload.EventUserModified}

	var otherCalled atomic.Bool
	dispatcher.Subscribe(payload.EventUserModified, func(context.Context, *payload.RemnawaveWebhook) error {
		panic("boom")
	})
	dispatcher.Subscribe(payload.EventUserModified, func(context.Context, *payload.RemnawaveWebhook) error {
		otherCalled.Store(true)
		return nil
	})

	handled, err := dispatcher.Publish(context.Background(), hook)

	assert.True(t, handled)
	require.Error(t, err)
	assert.ErrorContains(t, err, "webhook handler panic: boom")
	assert.True(t, otherCalled.Load())
}

func TestPublish_CanceledContextSkipsHandlers(t *testing.T) {
	dispatcher := NewDispatcher(testLogger{})
	hook := &payload.RemnawaveWebhook{Event: payload.EventUserModified}

	var called atomic.Bool
	dispatcher.Subscribe(payload.EventUserModified, func(context.Context, *payload.RemnawaveWebhook) error {
		called.Store(true)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handled, err := dispatcher.Publish(ctx, hook)

	assert.True(t, handled)
	require.ErrorIs(t, err, context.Canceled)
	assert.False(t, called.Load())
}
