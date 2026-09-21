package model

import "time"

type LoginAttempt struct {
	Username    string  `json:"username"`
	Ip          string  `json:"ip"`
	UserAgent   string  `json:"userAgent"`
	Description *string `json:"description"`
	Password    *string `json:"password"`
}

type Action string

const (
	ActionCreated Action = "CREATED"
	ActionUpdated Action = "UPDATED"
	ActionDeleted Action = "DELETED"
)

type SubpageConfig struct {
	Action Action `json:"action"`
	Uuid   string `json:"uuid"`
}

type ApiToken struct {
	Name     string    `json:"name"`
	Uuid     string    `json:"uuid"`
	ExpireAt time.Time `json:"expireAt"`
	Scopes   []string  `json:"scopes"`
}

type RemnawaveWebhookServiceEventsData struct {
	LoginAttempt  *LoginAttempt  `json:"loginAttempt"`
	PanelVersion  *string        `json:"panelVersion"`
	SubpageConfig *SubpageConfig `json:"subpageConfig"`
	ApiToken      *ApiToken      `json:"apiToken"`
}
