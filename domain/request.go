package domain

import (
	"fmt"
	"strings"
	"time"
)

type RegisterUserRequest struct {
	TelegramID           int64
	Username             string
	ExpiresAt            time.Time
	Description          *string
	HWIDDeviceLimit      *int64
	ActiveInternalSquads []string
}

func (r RegisterUserRequest) Validate() error {
	if r.TelegramID < 0 {
		return fmt.Errorf("%w: telegram ID cannot be negative", ErrInvalidUserRequest)
	}
	if strings.TrimSpace(r.Username) == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidUserRequest)
	}
	if r.ExpiresAt.IsZero() {
		return fmt.Errorf("%w: expiration time is required", ErrInvalidUserRequest)
	}
	if r.HWIDDeviceLimit != nil && *r.HWIDDeviceLimit < 0 {
		return fmt.Errorf("%w: HWID device limit cannot be negative", ErrInvalidUserRequest)
	}

	return nil
}

type UpdateUserRequest struct {
	UserID     uint64
	TelegramID int64

	Status               *UserStatus
	TrafficLimitBytes    *float64
	TrafficLimitStrategy *TrafficLimitStrategy
	ExpireAt             *time.Time
	Description          *string
	Tag                  *string
	Email                *string
	HWIDDeviceLimit      *int64
	ActiveInternalSquads []string
	ExternalSquadUUID    *string
}
