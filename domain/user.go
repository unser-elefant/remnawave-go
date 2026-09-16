package domain

import (
	"fmt"
	"net/url"
	"time"
)

var moscowLocation = time.FixedZone("MSK", 3*60*60)

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusDisabled UserStatus = "DISABLED"
	UserStatusLimited  UserStatus = "LIMITED"
	UserStatusExpired  UserStatus = "EXPIRED"
)

type TrafficLimitStrategy string

const (
	TrafficLimitNoReset TrafficLimitStrategy = "NO_RESET"
	TrafficLimitDay     TrafficLimitStrategy = "DAY"
	TrafficLimitWeek    TrafficLimitStrategy = "WEEK"
	TrafficLimitMonth   TrafficLimitStrategy = "MONTH"
)

type User struct {
	ID              uint64
	ShortUUID       string
	Username        string
	Status          UserStatus
	TelegramID      int64
	SubscriptionEnd time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time

	TrafficUsed      int64
	TrafficLimit     int64
	TrafficStrategy  TrafficLimitStrategy
	LastTrafficReset *time.Time

	DeviceLimit int

	SubscriptionURL string

	Description string
	Tag         string
	Email       string

	OnlineAt              *time.Time
	FirstConnectedAt      *time.Time
	LastConnectedNodeUUID *string
}

func (u *User) DaysLeft() int {
	if u == nil {
		return -1
	}
	duration := time.Until(u.SubscriptionEnd)
	if duration < 0 {
		return -1
	}
	days := int(duration.Hours() / 24)
	return days
}

func (u *User) HasTrafficLimit() bool {
	return u != nil && u.TrafficLimit > 0
}

func (u *User) TrafficUsagePercent() float64 {
	if u == nil || !u.HasTrafficLimit() {
		return 0
	}
	return (float64(u.TrafficUsed) / float64(u.TrafficLimit)) * 100
}

func (u *User) IsActive() bool {
	return u != nil && u.Status == UserStatusActive
}

func (u *User) IsExpired() bool {
	return u == nil || u.DaysLeft() < 0
}

func (u *User) GetFormattedExpirationDate() string {
	if u == nil {
		return ""
	}
	return u.SubscriptionEnd.In(moscowLocation).Format("02/01/2006 15:04")
}

func (u *User) GetFormattedTimeLeft() string {
	if u == nil {
		return ""
	}
	daysLeft := u.DaysLeft()
	if daysLeft < 0 {
		return "Истекла"
	}
	if u.ShouldShowTimeInHours() {
		hours, minutes := u.GetTimeUntilExpiration()
		return fmt.Sprintf("%d ч. %d мин.", hours, minutes)
	}
	return fmt.Sprintf("%d дн.", daysLeft)
}

func (u *User) GetTimeUntilExpiration() (hours int, minutes int) {
	if u == nil {
		return 0, 0
	}
	duration := time.Until(u.SubscriptionEnd)
	hours = int(duration.Hours())
	minutes = int(duration.Minutes()) % 60
	return hours, minutes
}

func (u *User) ShouldShowTimeInHours() bool {
	if u == nil {
		return false
	}
	duration := time.Until(u.SubscriptionEnd)
	return duration > 0 && duration.Hours() < 24
}

func (u *User) GetTrafficUsedGB() float64 {
	if u == nil {
		return 0
	}
	return float64(u.TrafficUsed) / (1024 * 1024 * 1024)
}

func (u *User) GetTrafficLimitGB() float64 {
	if u == nil {
		return 0
	}
	return float64(u.TrafficLimit) / (1024 * 1024 * 1024)
}

func (u *User) HasLastConnection() bool {
	return u != nil && u.LastConnectedNodeUUID != nil && *u.LastConnectedNodeUUID != ""
}

func (u *User) GetFormattedLastConnection() string {
	if u == nil || !u.HasLastConnection() || u.OnlineAt == nil {
		return ""
	}
	return u.OnlineAt.In(moscowLocation).Format("02/01/2006 15:04")
}

func (u *User) GetStatusEmoji() string {
	if u != nil && u.IsActive() {
		return "✅"
	}
	return "❌"
}

func (u *User) GetExpirationEmoji() string {
	if u == nil {
		return ""
	}
	daysLeft := u.DaysLeft()
	if daysLeft < 0 {
		return "🚫"
	}
	if daysLeft == 0 && time.Until(u.SubscriptionEnd).Hours() < 24 {
		return "‼️"
	}
	if daysLeft <= 3 {
		return "⚠️"
	}
	return "📅"
}

func (u *User) HasSubscriptionURL() bool {
	return u != nil && u.SubscriptionURL != ""
}

func (u *User) GetSubscriptionURLWithPort(port *int) string {
	if u == nil {
		return ""
	}
	if !u.HasSubscriptionURL() || port == nil || *port <= 0 {
		return u.SubscriptionURL
	}

	parsedURL, err := url.Parse(u.SubscriptionURL)
	if err != nil || parsedURL.Host == "" {
		return u.SubscriptionURL
	}

	host := parsedURL.Hostname()
	parsedURL.Host = fmt.Sprintf("%s:%d", host, *port)
	return parsedURL.String()
}

func (u *User) GetSubscriptionJSONURLWithPort(port *int) string {
	if u == nil {
		return ""
	}
	subscriptionURL := u.GetSubscriptionURLWithPort(port)
	if subscriptionURL == "" {
		return ""
	}

	jsonURL, err := url.JoinPath(subscriptionURL, "json")
	if err != nil {
		return subscriptionURL + "/json"
	}

	return jsonURL
}
