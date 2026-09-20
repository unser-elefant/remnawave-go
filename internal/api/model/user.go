package model

import "time"

type Squad struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type UserTraffic struct {
	UsedTrafficBytes         float64    `json:"usedTrafficBytes"`
	LifetimeUsedTrafficBytes float64    `json:"lifetimeUsedTrafficBytes"`
	OnlineAt                 *time.Time `json:"onlineAt"`
	FirstConnectedAt         *time.Time `json:"firstConnectedAt"`
	LastConnectedNodeUUID    *string    `json:"lastConnectedNodeUuid"`
}

type User struct {
	ID                     uint64      `json:"id"`
	ShortUUID              string      `json:"shortUuid"`
	Username               string      `json:"username"`
	Status                 string      `json:"status"`
	TrafficLimitBytes      int64       `json:"trafficLimitBytes"`
	TrafficLimitStrategy   string      `json:"trafficLimitStrategy"`
	ExpireAt               time.Time   `json:"expireAt"`
	TelegramID             *int64      `json:"telegramId"`
	Email                  *string     `json:"email"`
	Description            *string     `json:"description"`
	Tag                    *string     `json:"tag"`
	HWIDDeviceLimit        *int64      `json:"hwidDeviceLimit"`
	ExternalSquadUUID      *string     `json:"externalSquadUuid"`
	TrojanPassword         string      `json:"trojanPassword"`
	VlessUUID              string      `json:"vlessUuid"`
	SSPassword             string      `json:"ssPassword"`
	LastTriggeredThreshold int         `json:"lastTriggeredThreshold"`
	SubRevokedAt           *time.Time  `json:"subRevokedAt"`
	LastTrafficResetAt     *time.Time  `json:"lastTrafficResetAt"`
	CreatedAt              time.Time   `json:"createdAt"`
	UpdatedAt              time.Time   `json:"updatedAt"`
	SubscriptionURL        string      `json:"subscriptionUrl"`
	ActiveInternalSquads   []Squad     `json:"activeInternalSquads"`
	UserTraffic            UserTraffic `json:"userTraffic"`
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
