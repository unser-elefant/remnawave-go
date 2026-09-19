package bus

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/unser-elefant/remnawave-go/webhook/payload"
)

type logger interface {
	WithContext(ctx context.Context) *slog.Logger
}

type Handler func(context.Context, *payload.RemnawaveWebhook) error

type Dispatcher struct {
	mu   sync.RWMutex
	subs map[payload.Event][]Handler
	l    logger
}

func NewDispatcher(l logger) *Dispatcher {
	return &Dispatcher{
		subs: make(map[payload.Event][]Handler),
		l:    l,
	}
}

func (d *Dispatcher) Subscribe(e payload.Event, h Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.subs[e] = append(d.subs[e], h)
}

func (d *Dispatcher) Publish(ctx context.Context, p *payload.RemnawaveWebhook) (bool, error) {
	d.mu.RLock()
	handlers := append([]Handler(nil), d.subs[p.Event]...)
	d.mu.RUnlock()

	if len(handlers) == 0 {
		return false, nil
	}

	var errs []error
	for _, h := range handlers {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		if err := h(ctx, p); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return true, errors.Join(errs...)
	}

	return true, nil
}

func (d *Dispatcher) RegisterWebhookHandler(e payload.Event, h Handler) {
	d.Subscribe(e, func(ctx context.Context, p *payload.RemnawaveWebhook) error {
		if err := h(ctx, p); err != nil {
			d.l.WithContext(ctx).Error("Failed to process webhook event",
				"event", p.Event,
				"scope", p.Scope,
				"error", err)
			return err
		}
		return nil
	})
}
