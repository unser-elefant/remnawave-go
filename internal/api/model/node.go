package model

import (
	"encoding/json"
	"time"
)

type NodeIPStatus string

const (
	NodeIPStatusInbound    NodeIPStatus = "INBOUND"
	NodeIPStatusOutbound   NodeIPStatus = "OUTBOUND"
	NodeIPStatusManagement NodeIPStatus = "MANAGEMENT"
	NodeIPStatusTransit    NodeIPStatus = "TRANSIT"
	NodeIPStatusMonitoring NodeIPStatus = "MONITORING"
	NodeIPStatusReserve    NodeIPStatus = "RESERVE"
	NodeIPStatusBlocked    NodeIPStatus = "BLOCKED"
	NodeIPStatusFlagged    NodeIPStatus = "FLAGGED"
	NodeIPStatusDeprecated NodeIPStatus = "DEPRECATED"
	NodeIPStatusUnknown    NodeIPStatus = "UNKNOWN"
)

type NodeIP struct {
	IP     string       `json:"ip"`
	Status NodeIPStatus `json:"status"`
}

type NodeInbound struct {
	UUID        string          `json:"uuid"`
	ProfileUUID string          `json:"profileUuid"`
	Tag         string          `json:"tag"`
	Type        string          `json:"type"`
	Network     *string         `json:"network"`
	Security    *string         `json:"security"`
	Port        *float64        `json:"port"`
	RawInbound  json.RawMessage `json:"rawInbound"`
}

type NodeConfigProfile struct {
	ActiveConfigProfileUUID *string       `json:"activeConfigProfileUuid"`
	ActiveInbounds          []NodeInbound `json:"activeInbounds"`
}

type NodeProvider struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	FaviconLink *string   `json:"faviconLink"`
	LoginUrl    *string   `json:"loginUrl"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type NodeInfo struct {
	Arch              string   `json:"arch"`
	Cpus              int64    `json:"cpus"`
	CpuModel          string   `json:"cpuModel"`
	MemoryTotal       float64  `json:"memoryTotal"`
	HostName          string   `json:"hostName"`
	Platfrom          string   `json:"platform"`
	Release           string   `json:"relaese"`
	Type              string   `json:"type"`
	Version           string   `json:"verison"`
	NetworkInterfaces []string `json:"networkInterfaces"`
}

type NodeInterface struct {
	Interface     string  `json:"interface"`
	RxBytesPerSec float64 `json:"rxBytesPerSec"`
	TxBytesPerSec float64 `json:"txBytesPerSec"`
	RxTotal       float64 `json:"rxTotal"`
	TxTotal       float64 `json:"txTotal"`
}

type NodeStats struct {
	MemoryFree float64        `json:"memoryFree"`
	MemoryUsed float64        `json:"memoryUsed"`
	Uptime     float64        `json:"uptime"`
	LoadAvg    []float64      `json:"loadAvg"`
	Interface  *NodeInterface `json:"interface"`
}

type NodeSystem struct {
	Info  NodeInfo  `json:"info"`
	Stats NodeStats `json:"stats"`
}

type NodeVersions struct {
	Xray string `json:"xray"`
	Node string `json:"node"`
}

type Node struct {
	UUID                      string            `json:"uuid"`
	ID                        int64             `json:"id"`
	Name                      string            `json:"name"`
	Address                   string            `json:"address"`
	Port                      *int64            `json:"port"`
	ProxyURL                  *string           `json:"proxyUrl"`
	IsConnected               bool              `json:"isConnected"`
	IsDisabled                bool              `json:"isDisabled"`
	IsConnecting              bool              `json:"isConnecting"`
	LastStatusChange          *time.Time        `json:"lastStatusChange"`
	LastStatusMessage         *string           `json:"lastStatusMessage"`
	IsTrafficTrackingActive   bool              `json:"isTrafficTrackingActive"`
	TrafficResetDay           *int64            `json:"trafficResetDay"`
	TrafficLimitBytes         *float64          `json:"trafficLimitBytes"`
	TrafficUsedBytes          *float64          `json:"trafficUsedBytes"`
	NotifyPercent             *int64            `json:"notifyPercent"`
	ViewPosition              int64             `json:"viewPosition"`
	CountryCode               string            `json:"countryCode"`
	ConsumptionMultiplier     float64           `json:"consumptionMultiplier"`
	NodeConsumptionMultiplier float64           `json:"nodeConsumptionMultiplier"`
	Tags                      []string          `json:"tags"`
	IntegrationUUIDs          []string          `json:"integrationUuids"`
	IPs                       []NodeIP          `json:"ips"`
	CreatedAt                 time.Time         `json:"createdAt"`
	UpdatedAt                 time.Time         `json:"updatedAt"`
	ConfigProfile             NodeConfigProfile `json:"configProfile"`
	ProviderUUID              *string           `json:"providerUuid"`
	Provider                  *NodeProvider     `json:"provider"`
	ActivePluginUUID          *string           `json:"activePluginUuid"`
	System                    *NodeSystem       `json:"system"`
	Versions                  *NodeVersions     `json:"versions"`
	XrayUptime                float64           `json:"xrayUptime"`
	UsersOnline               float64           `json:"usersOnline"`
	Note                      *string           `json:"note"`
}
