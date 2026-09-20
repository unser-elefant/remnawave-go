package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/unser-elefant/remnawave-go/webhook/payload"
)

type logger interface {
	WithContext(ctx context.Context) *slog.Logger
}

type validator interface {
	ValidateSignature(signature string, body []byte) bool
	ValidateIP(addr string) bool
}

type dispatcher interface {
	Publish(ctx context.Context, p *payload.RemnawaveWebhook) (bool, error)
}

type unhandledFunc func(ctx context.Context, payload *payload.RemnawaveWebhook)

type Processor struct {
	l  logger
	v  validator
	d  dispatcher
	uf unhandledFunc
}

func New(l logger, v validator, d dispatcher, uf unhandledFunc) *Processor {
	return &Processor{l: l, v: v, d: d, uf: uf}
}

func (p *Processor) Process(ctx context.Context, body []byte, signature string, remoteAddr string) error {
	log := p.l.WithContext(ctx)

	if !p.v.ValidateSignature(signature, body) {
		return fmt.Errorf("invalid signature")
	}

	if !p.v.ValidateIP(remoteAddr) {
		return fmt.Errorf("ip is not allowed: %s", remoteAddr)
	}

	var webhookPayload payload.RemnawaveWebhook
	if err := json.Unmarshal(body, &webhookPayload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	log.Info("Received webhook",
		"scope", webhookPayload.Scope,
		"event", webhookPayload.Event)

	if err := ctx.Err(); err != nil {
		return err
	}

	handled := false
	var err error
	handled, err = p.d.Publish(ctx, &webhookPayload)
	if err != nil {
		return fmt.Errorf("publish webhook event: %w", err)
	}

	if !handled {
		p.uf(ctx, &webhookPayload)
		return fmt.Errorf("unknown webhook event: scope=%s event=%s", webhookPayload.Scope, webhookPayload.Event)
	}

	return nil
}
