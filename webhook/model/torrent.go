package model

import "github.com/unser-elefant/remnawave-go/internal/api/model"

type ActionReport struct {
	Blocked       bool    `json:"blocked"`
	Ip            string  `json:"ip"`
	BlockDuraion  float64 `json:"blockDuration"`
	WillUnblockAt string  `json:"willUnblockAt"`
	UserId        string  `json:"userId"`
	ProcessedAt   string  `json:"processedAt"`
}

type XrayReport struct {
	Email          *string  `json:"email"`
	Level          *float64 `json:"level"`
	Protocol       *string  `json:"protocol"`
	Network        string   `json:"network"`
	Source         *string  `json:"source"`
	Destination    string   `json:"destination"`
	RouteTarget    *string  `json:"routeTarget"`
	OriginalTarget *string  `json:"originalTarget"`
	InboundTag     *string  `json:"inboundTag"`
	InboundName    *string  `json:"inboundName"`
	InboundLocal   *string  `json:"inboundLocal"`
	OutboundTag    *string  `json:"outboundTag"`
	Ts             float64  `json:"ts"`
}

type Report struct {
	ActionReport ActionReport `json:"actionReport"`
	XrayReport   XrayReport   `josn:"xrayReport"`
}

type RemnawaveWebhookTorrentBlockerEventsData struct {
	Node   model.Node `json:"node"`
	User   model.User `json:"user"`
	Report Report     `json:"report"`
}
