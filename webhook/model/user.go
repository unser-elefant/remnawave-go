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

type RemnawaveWebhookUserEventsData struct {
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
	LastTriggeredThreshold int64       `json:"lastTriggeredThreshold"`
	SubRevokedAt           *time.Time  `json:"subRevokedAt"`
	LastTrafficResetAt     *time.Time  `json:"lastTrafficResetAt"`
	CreatedAt              time.Time   `json:"createdAt"`
	UpdatedAt              time.Time   `json:"updatedAt"`
	SubscriptionURL        string      `json:"subscriptionUrl"`
	ActiveInternalSquads   []Squad     `json:"activeInternalSquads"`
	UserTraffic            UserTraffic `json:"userTraffic"`
}
