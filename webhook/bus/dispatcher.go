package bus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
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

	if err := ctx.Err(); err != nil {
		return true, err
	}

	errs := make([]error, len(handlers))
	var wg sync.WaitGroup
	for i, h := range handlers {
		wg.Go(func() {
			errs[i] = d.runHandler(ctx, h, p)
		})
	}
	wg.Wait()

	return true, errors.Join(errs...)
}

func (d *Dispatcher) runHandler(ctx context.Context, h Handler, p *payload.RemnawaveWebhook) (err error) {
	defer func() {
		if r := recover(); r != nil {
			d.l.WithContext(ctx).Error("Panic in webhook handler",
				"event", p.Event,
				"scope", p.Scope,
				"panic", r,
				"stack", string(debug.Stack()))
			err = fmt.Errorf("webhook handler panic: %v", r)
		}
	}()

	return h(ctx, p)
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
