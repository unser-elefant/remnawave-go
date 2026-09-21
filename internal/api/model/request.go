package model

import (
	"net/url"
	"strconv"
)

type UserStreamQuery struct {
	Cursor               *int64
	Size                 *int64
	Status               *UserStatus
	TrafficLimitStrategy *TrafficLimitStrategy
	TelegramID           *string
	Email                *string
	Tag                  *string
	ExternalSquadUUID    *string
}

func (q UserStreamQuery) Encode() string {
	values := url.Values{}

	if q.Cursor != nil {
		values.Set("cursor", strconv.FormatInt(*q.Cursor, 10))
	}
	if q.Size != nil {
		values.Set("size", strconv.FormatInt(*q.Size, 10))
	}
	if q.Status != nil {
		values.Set("status", string(*q.Status))
	}
	if q.TrafficLimitStrategy != nil {
		values.Set("trafficLimitStrategy", string(*q.TrafficLimitStrategy))
	}
	if q.TelegramID != nil {
		values.Set("telegramId", *q.TelegramID)
	}
	if q.Email != nil {
		values.Set("email", *q.Email)
	}
	if q.Tag != nil {
		values.Set("tag", *q.Tag)
	}
	if q.ExternalSquadUUID != nil {
		values.Set("externalSquadUuid", *q.ExternalSquadUUID)
	}

	return values.Encode()
}

type CreateUserRequest struct {
	Username             string   `json:"username"`                       // Required. 3-36 chars, pattern: ^[a-zA-Z0-9_-]+$
	ExpireAt             string   `json:"expireAt"`                       // Required. Format: 2025-01-17T15:38:45.065Z
	Status               *string  `json:"status,omitempty"`               // Optional. Default: ACTIVE. Enum: ACTIVE, DISABLED, LIMITED, EXPIRED
	ShortUUID            *string  `json:"shortUuid,omitempty"`            // Optional. Short UUID identifier
	TrojanPassword       *string  `json:"trojanPassword,omitempty"`       // Optional. 8-32 chars
	VlessUUID            *string  `json:"vlessUuid,omitempty"`            // Optional. Valid UUID format
	SSPassword           *string  `json:"ssPassword,omitempty"`           // Optional. 8-32 chars
	TrafficLimitBytes    *int64   `json:"trafficLimitBytes,omitempty"`    // Optional. Min: 0. 0 = unlimited
	TrafficLimitStrategy *string  `json:"trafficLimitStrategy,omitempty"` // Optional. Default: NO_RESET. Enum: NO_RESET, DAY, WEEK, MONTH
	CreatedAt            *string  `json:"createdAt,omitempty"`            // Optional. Format: 2025-01-17T15:38:45.065Z
	LastTrafficResetAt   *string  `json:"lastTrafficResetAt,omitempty"`   // Optional. Format: 2025-01-17T15:38:45.065Z
	Description          *string  `json:"description,omitempty"`          // Optional. Additional notes
	Tag                  *string  `json:"tag,omitempty"`                  // Optional. Max 16 chars, pattern: ^[A-Z0-9_]+$
	TelegramID           *int64   `json:"telegramId,omitempty"`           // Optional. Telegram user ID
	Email                *string  `json:"email,omitempty"`                // Optional. Valid email format
	HWIDDeviceLimit      *int64   `json:"hwidDeviceLimit,omitempty"`      // Optional. Min: 0
	ActiveInternalSquads []string `json:"activeInternalSquads,omitempty"` // Optional. Array of squad UUIDs
	UUID                 *string  `json:"uuid,omitempty"`                 // Optional. Specific UUID, otherwise auto-generated
	ExternalSquadUUID    *string  `json:"externalSquadUuid,omitempty"`    // Optional. External squad UUID
}

type CreateUserInitInfo struct {
	Username             string   `json:"username"`
	ExpireAt             string   `json:"expireAt"`
	TelegramID           int64    `json:"telegramId"`
	ActiveInternalSquads []string `json:"activeInternalSquads"`
	HWIDDeviceLimit      *int64   `json:"hwidDeviceLimit"`
	Description          *string  `json:"description"`
	Status               *string  `json:"status"`
	TrafficLimitBytes    *int64   `json:"trafficLimitBytes"`
	TrafficLimitStrategy *string  `json:"trafficLimitStrategy"`
}

func NewCreateUserRequest(info *CreateUserInitInfo) *CreateUserRequest {
	if info == nil {
		return nil
	}
	req := &CreateUserRequest{
		Username:             info.Username,
		ExpireAt:             info.ExpireAt,
		TelegramID:           &info.TelegramID,
		ActiveInternalSquads: info.ActiveInternalSquads,
		HWIDDeviceLimit:      info.HWIDDeviceLimit,
		Description:          info.Description,
	}

	if info.Status != nil {
		req.Status = info.Status
	} else {
		req.Status = new("ACTIVE")
	}

	if info.TrafficLimitBytes != nil {
		req.TrafficLimitBytes = info.TrafficLimitBytes
	} else {
		req.TrafficLimitBytes = new(int64(0))
	}

	if info.TrafficLimitStrategy != nil {
		req.TrafficLimitStrategy = info.TrafficLimitStrategy
	} else {
		req.TrafficLimitStrategy = new("NO_RESET")
	}

	return req
}
