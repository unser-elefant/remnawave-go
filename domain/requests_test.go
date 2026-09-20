package domain

import (
	"errors"
	"testing"
	"time"
)

func TestRegisterUserRequestValidate(t *testing.T) {
	valid := RegisterUserRequest{
		TelegramID: 12345,
		Username:   "alice",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("valid request failed validation: %v", err)
	}

	tests := []struct {
		name    string
		request RegisterUserRequest
	}{
		{name: "missing telegram id", request: RegisterUserRequest{Username: "alice", ExpiresAt: valid.ExpiresAt}},
		{name: "missing username", request: RegisterUserRequest{TelegramID: 12345, ExpiresAt: valid.ExpiresAt}},
		{name: "missing expiration", request: RegisterUserRequest{TelegramID: 12345, Username: "alice"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if !errors.Is(err, ErrInvalidUserRequest) {
				t.Fatalf("expected ErrInvalidUserRequest, got %v", err)
			}
		})
	}
}

func TestRegisterUserRequestValidateRejectsNegativeDeviceLimit(t *testing.T) {
	limit := int64(-1)
	err := (RegisterUserRequest{
		TelegramID:      12345,
		Username:        "alice",
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		HWIDDeviceLimit: &limit,
	}).Validate()

	if !errors.Is(err, ErrInvalidUserRequest) {
		t.Fatalf("expected ErrInvalidUserRequest, got %v", err)
	}
}
