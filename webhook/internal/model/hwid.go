package model

import "github.com/unser-elefant/remnawave-go/internal/api/model"

type RemnawaveWebhookUserHwidDevicesEventsData struct {
	User           model.User   `json:"user"`
	HwidUserDevice model.Device `json:"hwidUserDevice"`
}
