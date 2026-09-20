package processor

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unser-elefant/remnawave-go/webhook/payload"
)

type testLogger struct{}

func (testLogger) WithContext(context.Context) *slog.Logger { return slog.Default() }

type validatorStub struct {
	signatureValid bool
	ipValid        bool
}

func (v validatorStub) ValidateSignature(string, []byte) bool { return v.signatureValid }
func (v validatorStub) ValidateIP(string) bool                { return v.ipValid }

type dispatcherStub struct {
	handled bool
	err     error
	payload *payload.RemnawaveWebhook
}

func (d *dispatcherStub) Publish(_ context.Context, p *payload.RemnawaveWebhook) (bool, error) {
	d.payload = p
	return d.handled, d.err
}

func validProcessor(d *dispatcherStub, v validatorStub, unhandled unhandledFunc) *Processor {
	return New(testLogger{}, v, d, unhandled)
}

func TestProcess_PublishesValidWebhook(t *testing.T) {
	dispatcher := &dispatcherStub{handled: true}
	service := validProcessor(dispatcher, validatorStub{signatureValid: true, ipValid: true}, nil)
	body := []byte(`{"scope":"user","event":"user.modified","data":{"username":"john"}}`)

	err := service.Process(context.Background(), body, "valid", "10.0.0.1:9999")

	require.NoError(t, err)
	require.NotNil(t, dispatcher.payload)
	assert.Equal(t, payload.ScopeUser, dispatcher.payload.Scope)
	assert.Equal(t, payload.EventUserModified, dispatcher.payload.Event)
}

func TestProcess_RejectsInvalidSignature(t *testing.T) {
	dispatcher := &dispatcherStub{handled: true}
	service := validProcessor(dispatcher, validatorStub{ipValid: true}, nil)

	err := service.Process(context.Background(), []byte(`{}`), "invalid", "10.0.0.1:9999")

	assert.EqualError(t, err, "invalid signature")
	assert.Nil(t, dispatcher.payload)
}

func TestProcess_RejectsDisallowedIP(t *testing.T) {
	dispatcher := &dispatcherStub{handled: true}
	service := validProcessor(dispatcher, validatorStub{signatureValid: true}, nil)

	err := service.Process(context.Background(), []byte(`{}`), "valid", "10.0.0.1:9999")

	assert.EqualError(t, err, "ip is not allowed: 10.0.0.1:9999")
	assert.Nil(t, dispatcher.payload)
}

func TestProcess_RejectsInvalidPayload(t *testing.T) {
	service := validProcessor(&dispatcherStub{handled: true}, validatorStub{signatureValid: true, ipValid: true}, nil)

	err := service.Process(context.Background(), []byte(`{"scope":`), "valid", "10.0.0.1:9999")

	assert.ErrorContains(t, err, "invalid payload")
}

func TestProcess_RejectsCanceledContextBeforePublish(t *testing.T) {
	dispatcher := &dispatcherStub{handled: true}
	service := validProcessor(dispatcher, validatorStub{signatureValid: true, ipValid: true}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := service.Process(ctx, []byte(`{"scope":"user","event":"user.modified"}`), "valid", "10.0.0.1:9999")

	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, dispatcher.payload)
}

func TestProcess_ReportsUnhandledWebhook(t *testing.T) {
	dispatcher := &dispatcherStub{}
	var unhandledPayload *payload.RemnawaveWebhook
	service := validProcessor(dispatcher, validatorStub{signatureValid: true, ipValid: true}, func(ctx context.Context, p *payload.RemnawaveWebhook) error {
		unhandledPayload = p
		return nil
	})
	body, err := json.Marshal(payload.RemnawaveWebhook{Scope: payload.ScopeUser, Event: payload.Event("user.custom")})
	require.NoError(t, err)

	err = service.Process(context.Background(), body, "valid", "10.0.0.1:9999")

	assert.ErrorContains(t, err, "unknown webhook event: scope=user event=user.custom")
	assert.NotNil(t, unhandledPayload)
}

func TestProcess_WrapsPublishError(t *testing.T) {
	dispatcher := &dispatcherStub{handled: true, err: errors.New("publish failed")}
	service := validProcessor(dispatcher, validatorStub{signatureValid: true, ipValid: true}, nil)

	err := service.Process(context.Background(), []byte(`{"scope":"user","event":"user.modified"}`), "valid", "10.0.0.1:9999")

	assert.ErrorContains(t, err, "publish webhook event: publish failed")
}
