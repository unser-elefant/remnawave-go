package payload

import (
	"encoding/json"
	"time"
)

type RemnawaveWebhook struct {
	Scope     Scope           `json:"scope"`
	Event     Event           `json:"event"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	Meta      *Meta           `json:"meta"`
}

type Meta struct {
	NotConnectedAfterHours *int `json:"notConnectedAfterHours"`
	Expiration             *int `json:"expiration"`
}
