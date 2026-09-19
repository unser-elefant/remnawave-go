package bus

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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
