package model

import "time"

type RemnawaveWebhookCrmEventsData struct {
	ProviderName  string    `json:"providerName"`
	NodeName      string    `json:"nodeName"`
	NextBillingAt time.Time `json:"nextBillingAt"`
	LoginUrl      string    `json:"loginUrl"`
}
