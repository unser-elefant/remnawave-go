package domain

import (
	"strings"
	"time"
)

type Device struct {
	HWID        string
	UserID      int64
	Platform    string
	OSVersion   string
	DeviceModel string
	UserAgent   string
	RequestIP   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (d *Device) GetPlatformInfo() string {
	if d == nil {
		return ""
	}
	if d.Platform == "" {
		return "Неизвестная платформа"
	}

	info := d.Platform
	if d.OSVersion != "" {
		info += " " + d.OSVersion
	}
	return info
}

func (d *Device) GetDisplayName() string {
	if d == nil {
		return ""
	}
	name := d.GetPlatformInfo()
	if d.DeviceModel != "" {
		name += " (" + d.DeviceModel + ")"
	}
	return name
}

func (d *Device) GetTruncatedUserAgent(maxLen int) string {
	if d == nil || d.UserAgent == "" || maxLen <= 0 {
		return ""
	}
	runes := []rune(d.UserAgent)
	if len(runes) <= maxLen {
		return d.UserAgent
	}
	if maxLen <= 3 {
		return strings.Repeat(".", maxLen)
	}
	return string(runes[:maxLen-3]) + strings.Repeat(".", 3)
}

func (d *Device) HasUserAgent() bool {
	return d != nil && d.UserAgent != ""
}

func (d *Device) HasRequestIP() bool {
	return d != nil && d.RequestIP != ""
}

func (d *Device) GetFormattedCreatedAt() string {
	if d == nil {
		return ""
	}
	return d.CreatedAt.Format("02/01/2006 15:04")
}

func (d *Device) GetDeviceAge() int {
	if d == nil {
		return 0
	}
	duration := time.Since(d.CreatedAt)
	return int(duration.Hours() / 24)
}
