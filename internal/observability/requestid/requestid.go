package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
)

type contextKey struct{}

var key contextKey

func New() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func With(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, key, id)
}

func FromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	id, ok := ctx.Value(key).(string)
	if !ok || id == "" {
		return "", false
	}

	return id, true
}

func GetOrCreate(ctx context.Context) (context.Context, string) {
	if id, ok := FromContext(ctx); ok {
		return ctx, id
	}

	id := New()
	return With(ctx, id), id
}
