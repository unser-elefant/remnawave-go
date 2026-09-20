package typed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/unser-elefant/remnawave-go/webhook/payload"
)

type WebhookParser[T any] func(*payload.RemnawaveWebhook) (T, error)

func TypedWebhookHandler[T any](parser WebhookParser[T], h func(context.Context, T) error) func(context.Context, *payload.RemnawaveWebhook) error {
	return func(ctx context.Context, p *payload.RemnawaveWebhook) error {
		value, err := parser(p)
		if err != nil {
			return err
		}
		return h(ctx, value)
	}
}

func ParseWebhookData[T any](p *payload.RemnawaveWebhook) (T, error) {
	var data T
	if err := json.Unmarshal(p.Data, &data); err != nil {
		return data, fmt.Errorf("parse webhook data: %w", err)
	}
	return data, nil
}
