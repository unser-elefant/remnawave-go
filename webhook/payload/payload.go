package payload

import (
	"encoding/json"
	"time"
)

type Meta struct {
	NotConnectedAfterHours *int `json:"notConnectedAfterHours"`
	Expiration             *int `json:"expiration"`
}

type RemnawaveWebhook struct {
	Scope     Scope           `json:"scope"`
	Event     Event           `json:"event"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	Meta      *Meta           `json:"meta"`
}
