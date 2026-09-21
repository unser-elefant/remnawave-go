package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_DaysLeft(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		subscriptionEnd time.Time
		expectedMin     int
		expectedMax     int
	}{
		{name: "expires in 30 days", subscriptionEnd: now.Add(30 * 24 * time.Hour), expectedMin: 29, expectedMax: 30},
		{name: "expires in 7 days", subscriptionEnd: now.Add(7 * 24 * time.Hour), expectedMin: 6, expectedMax: 7},
		{name: "expires today", subscriptionEnd: now.Add(12 * time.Hour), expectedMin: 0, expectedMax: 0},
		{name: "expired 5 hours ago", subscriptionEnd: now.Add(-5 * time.Hour), expectedMin: -1, expectedMax: -1},
		{name: "expired 5 days ago", subscriptionEnd: now.Add(-5 * 24 * time.Hour), expectedMin: -1, expectedMax: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionEnd: tt.subscriptionEnd}
			result := user.DaysLeft()
			assert.GreaterOrEqual(t, result, tt.expectedMin)
			assert.LessOrEqual(t, result, tt.expectedMax)
		})
	}
}

func TestUser_HasTrafficLimit(t *testing.T) {
	tests := []struct {
		name         string
		trafficLimit float64
		expected     bool
	}{
		{name: "has limit", trafficLimit: 10 * 1024 * 1024 * 1024, expected: true},
		{name: "no limit (unlimited)", trafficLimit: 0, expected: false},
		{name: "negative (should be false)", trafficLimit: -1, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{TrafficLimit: tt.trafficLimit}
			assert.Equal(t, tt.expected, user.HasTrafficLimit())
		})
	}
}

func TestUser_TrafficUsagePercent(t *testing.T) {
	tests := []struct {
		name         string
		trafficUsed  float64
		trafficLimit float64
		expected     float64
	}{
		{name: "50% used", trafficUsed: 5 * 1024 * 1024 * 1024, trafficLimit: 10 * 1024 * 1024 * 1024, expected: 50.0},
		{name: "100% used", trafficUsed: 10 * 1024 * 1024 * 1024, trafficLimit: 10 * 1024 * 1024 * 1024, expected: 100.0},
		{name: "no limit (should return 0)", trafficUsed: 5 * 1024 * 1024 * 1024, trafficLimit: 0, expected: 0.0},
		{name: "over 100% used", trafficUsed: 15 * 1024 * 1024 * 1024, trafficLimit: 10 * 1024 * 1024 * 1024, expected: 150.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{TrafficUsed: tt.trafficUsed, TrafficLimit: tt.trafficLimit}
			assert.InDelta(t, tt.expected, user.TrafficUsagePercent(), 0.01)
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   UserStatus
		expected bool
	}{
		{name: "active user", status: UserStatusActive, expected: true},
		{name: "disabled user", status: UserStatusDisabled, expected: false},
		{name: "limited user", status: UserStatusLimited, expected: false},
		{name: "expired user", status: UserStatusExpired, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}
			assert.Equal(t, tt.expected, user.IsActive())
		})
	}
}

func TestUser_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		subscriptionEnd time.Time
		expected        bool
	}{
		{name: "not expired", subscriptionEnd: now.Add(30 * 24 * time.Hour), expected: false},
		{name: "expired", subscriptionEnd: now.Add(-5 * 24 * time.Hour), expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionEnd: tt.subscriptionEnd}
			assert.Equal(t, tt.expected, user.IsExpired())
		})
	}
}

func TestUser_GetFormattedExpirationDate(t *testing.T) {
	testTime := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)
	user := &User{SubscriptionEnd: testTime}

	assert.Equal(t, "15/06/2025 17:30", user.GetFormattedExpirationDate())
}

func TestUser_GetFormattedTimeLeft(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		addDuration  time.Duration
		expected     string
		containsText string
	}{
		{name: "expired", addDuration: -5 * time.Hour, expected: "Истекла"},
		{name: "shown in hours", addDuration: 5*time.Hour + 30*time.Minute, containsText: "ч."},
		{name: "shown in days", addDuration: 3*24*time.Hour + time.Minute, expected: "3 дн."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionEnd: now.Add(tt.addDuration)}
			result := user.GetFormattedTimeLeft()

			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			}
			if tt.containsText != "" {
				assert.Contains(t, result, tt.containsText)
				assert.Contains(t, result, "мин.")
			}
		})
	}
}

func TestUser_GetTimeUntilExpiration(t *testing.T) {
	tests := []struct {
		name         string
		addDuration  time.Duration
		checkMinutes bool
	}{
		{name: "5 hours until expiration", addDuration: 5*time.Hour + 30*time.Minute, checkMinutes: true},
		{name: "exactly 24 hours", addDuration: 24 * time.Hour, checkMinutes: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionEnd: time.Now().Add(tt.addDuration)}
			hours, minutes := user.GetTimeUntilExpiration()
			expectedHours := int(tt.addDuration.Hours())

			assert.InDelta(t, expectedHours, hours, 1)
			if tt.checkMinutes {
				assert.GreaterOrEqual(t, minutes, 0)
				assert.Less(t, minutes, 60)
			}
		})
	}
}

func TestUser_ShouldShowTimeInHours(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		subscriptionEnd time.Time
		expected        bool
	}{
		{name: "less than 24 hours", subscriptionEnd: now.Add(5 * time.Hour), expected: true},
		{name: "more than 24 hours", subscriptionEnd: now.Add(30 * time.Hour), expected: false},
		{name: "expired recently (less than 24 hours ago)", subscriptionEnd: now.Add(-5 * time.Hour), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionEnd: tt.subscriptionEnd}
			assert.Equal(t, tt.expected, user.ShouldShowTimeInHours())
		})
	}
}

func TestUser_GetTrafficUsedGB(t *testing.T) {
	tests := []struct {
		name        string
		trafficUsed float64
		expected    float64
	}{
		{name: "1 GB", trafficUsed: 1 * 1024 * 1024 * 1024, expected: 1.0},
		{name: "5.5 GB", trafficUsed: 5.5 * 1024 * 1024 * 1024, expected: 5.5},
		{name: "0 GB", trafficUsed: 0, expected: 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{TrafficUsed: tt.trafficUsed}
			assert.InDelta(t, tt.expected, user.GetTrafficUsedGB(), 0.01)
		})
	}
}

func TestUser_GetTrafficLimitGB(t *testing.T) {
	tests := []struct {
		name         string
		trafficLimit float64
		expected     float64
	}{
		{name: "10 GB", trafficLimit: 10 * 1024 * 1024 * 1024, expected: 10.0},
		{name: "unlimited", trafficLimit: 0, expected: 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{TrafficLimit: tt.trafficLimit}
			assert.InDelta(t, tt.expected, user.GetTrafficLimitGB(), 0.01)
		})
	}
}

func TestUser_HasLastConnection(t *testing.T) {
	tests := []struct {
		name     string
		nodeUUID *string
		expected bool
	}{
		{name: "has connection info", nodeUUID: new("node-1"), expected: true},
		{name: "empty connection info", nodeUUID: new(""), expected: false},
		{name: "no connection info", nodeUUID: nil, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{LastConnectedNodeUUID: tt.nodeUUID}
			assert.Equal(t, tt.expected, user.HasLastConnection())
		})
	}
}

func TestUser_GetFormattedLastConnection(t *testing.T) {
	tests := []struct {
		name     string
		nodeUUID *string
		onlineAt *time.Time
		expected string
	}{
		{
			name:     "has connection",
			nodeUUID: new("node-1"),
			onlineAt: new(time.Date(2025, 3, 15, 14, 30, 0, 0, time.UTC)),
			expected: "15/03/2025 17:30",
		},
		{name: "no connection", nodeUUID: nil, onlineAt: nil, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{LastConnectedNodeUUID: tt.nodeUUID, OnlineAt: tt.onlineAt}
			assert.Equal(t, tt.expected, user.GetFormattedLastConnection())
		})
	}
}

func TestUser_GetFormattedLastConnectionWithoutOnlineAt(t *testing.T) {
	user := &User{LastConnectedNodeUUID: new("node-1")}

	assert.Empty(t, user.GetFormattedLastConnection())
}

func TestNilUserMethodsAreSafe(t *testing.T) {
	var user *User

	assert.Equal(t, -1, user.DaysLeft())
	assert.False(t, user.HasTrafficLimit())
	assert.Zero(t, user.TrafficUsagePercent())
	assert.False(t, user.IsActive())
	assert.True(t, user.IsExpired())
	assert.Empty(t, user.GetFormattedExpirationDate())
	assert.Empty(t, user.GetFormattedTimeLeft())
	assert.Zero(t, user.GetTrafficUsedGB())
	assert.Zero(t, user.GetTrafficLimitGB())
	assert.False(t, user.HasLastConnection())
	assert.Equal(t, "❌", user.GetStatusEmoji())
	assert.Empty(t, user.GetExpirationEmoji())
	assert.False(t, user.HasSubscriptionURL())
	assert.Empty(t, user.GetSubscriptionURLWithPort(nil))
	assert.Empty(t, user.GetSubscriptionJSONURLWithPort(nil))
}

func TestUser_GetStatusEmoji(t *testing.T) {
	tests := []struct {
		name     string
		status   UserStatus
		expected string
	}{
		{name: "active", status: UserStatusActive, expected: "✅"},
		{name: "disabled", status: UserStatusDisabled, expected: "❌"},
		{name: "expired", status: UserStatusExpired, expected: "❌"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}
			assert.Equal(t, tt.expected, user.GetStatusEmoji())
		})
	}
}

func TestUser_GetExpirationEmoji(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		subscriptionEnd time.Time
		expected        string
	}{
		{name: "expired", subscriptionEnd: now.Add(-5 * 24 * time.Hour), expected: "🚫"},
		{name: "expires in less than 24 hours", subscriptionEnd: now.Add(12 * time.Hour), expected: "‼️"},
		{name: "expires soon (3 days)", subscriptionEnd: now.Add(3 * 24 * time.Hour), expected: "⚠️"},
		{name: "expires in 5 days", subscriptionEnd: now.Add(5 * 24 * time.Hour), expected: "📅"},
		{name: "expires in future (30 days)", subscriptionEnd: now.Add(30 * 24 * time.Hour), expected: "📅"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionEnd: tt.subscriptionEnd}
			assert.Equal(t, tt.expected, user.GetExpirationEmoji())
		})
	}
}

func TestUser_HasSubscriptionURL(t *testing.T) {
	tests := []struct {
		name            string
		subscriptionURL string
		expected        bool
	}{
		{name: "has URL", subscriptionURL: "https://example.com/sub", expected: true},
		{name: "no URL", subscriptionURL: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionURL: tt.subscriptionURL}
			assert.Equal(t, tt.expected, user.HasSubscriptionURL())
		})
	}
}

func TestUser_GetSubscriptionURLWithPort(t *testing.T) {
	tests := []struct {
		name            string
		subscriptionURL string
		port            *int
		expected        string
	}{
		{name: "add port to URL", subscriptionURL: "https://example.com/sub", port: new(8443), expected: "https://example.com:8443/sub"},
		{name: "no port provided", subscriptionURL: "https://example.com/sub", port: nil, expected: "https://example.com/sub"},
		{name: "zero port provided", subscriptionURL: "https://example.com/sub", port: new(0), expected: "https://example.com/sub"},
		{name: "empty URL", subscriptionURL: "", port: new(8443), expected: ""},
		{name: "invalid URL", subscriptionURL: "not-a-valid-url", port: new(8443), expected: "not-a-valid-url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionURL: tt.subscriptionURL}
			assert.Equal(t, tt.expected, user.GetSubscriptionURLWithPort(tt.port))
		})
	}
}

func TestUser_GetSubscriptionJSONURLWithPort(t *testing.T) {
	tests := []struct {
		name            string
		subscriptionURL string
		port            *int
		expected        string
	}{
		{name: "add json path with port", subscriptionURL: "https://example.com/sub", port: new(8443), expected: "https://example.com:8443/sub/json"},
		{name: "add json path without port", subscriptionURL: "https://example.com/sub", port: nil, expected: "https://example.com/sub/json"},
		{name: "preserve single slash", subscriptionURL: "https://example.com/sub/", port: nil, expected: "https://example.com/sub/json"},
		{name: "empty URL", subscriptionURL: "", port: new(8443), expected: ""},
		{name: "invalid URL falls back", subscriptionURL: "not-a-valid-url", port: new(8443), expected: "not-a-valid-url/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionURL: tt.subscriptionURL}
			assert.Equal(t, tt.expected, user.GetSubscriptionJSONURLWithPort(tt.port))
		})
	}
}

func BenchmarkUser_DaysLeft(b *testing.B) {
	user := &User{SubscriptionEnd: time.Now().Add(30 * 24 * time.Hour)}
	for b.Loop() {
		user.DaysLeft()
	}
}

func BenchmarkUser_IsExpired(b *testing.B) {
	user := &User{SubscriptionEnd: time.Now().Add(30 * 24 * time.Hour)}
	for b.Loop() {
		user.IsExpired()
	}
}

func BenchmarkUser_IsActive(b *testing.B) {
	user := &User{
		Status:          UserStatusActive,
		SubscriptionEnd: time.Now().Add(30 * 24 * time.Hour),
		TrafficLimit:    10737418240,
		TrafficUsed:     5368709120,
	}
	for b.Loop() {
		user.IsActive()
	}
}

func BenchmarkUser_GetTimeUntilExpiration(b *testing.B) {
	user := &User{SubscriptionEnd: time.Now().Add(5 * time.Hour)}
	for b.Loop() {
		user.GetTimeUntilExpiration()
	}
}

func BenchmarkUser_ShouldShowTimeInHours(b *testing.B) {
	user := &User{SubscriptionEnd: time.Now().Add(12 * time.Hour)}
	for b.Loop() {
		user.ShouldShowTimeInHours()
	}
}

func BenchmarkUser_GetTrafficUsedGB(b *testing.B) {
	user := &User{TrafficUsed: 5905580032}
	for b.Loop() {
		user.GetTrafficUsedGB()
	}
}

func BenchmarkUser_GetTrafficLimitGB(b *testing.B) {
	user := &User{TrafficLimit: 10737418240}
	for b.Loop() {
		user.GetTrafficLimitGB()
	}
}

func BenchmarkUser_TrafficUsagePercent(b *testing.B) {
	user := &User{TrafficLimit: 10737418240, TrafficUsed: 5368709120}
	for b.Loop() {
		user.TrafficUsagePercent()
	}
}

func BenchmarkUser_GetStatusEmoji(b *testing.B) {
	user := &User{Status: UserStatusActive, SubscriptionEnd: time.Now().Add(30 * 24 * time.Hour)}
	for b.Loop() {
		user.GetStatusEmoji()
	}
}

func BenchmarkUser_GetExpirationEmoji(b *testing.B) {
	user := &User{SubscriptionEnd: time.Now().Add(30 * 24 * time.Hour)}
	for b.Loop() {
		user.GetExpirationEmoji()
	}
}

func BenchmarkUser_GetFormattedExpirationDate(b *testing.B) {
	user := &User{SubscriptionEnd: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)}
	for b.Loop() {
		user.GetFormattedExpirationDate()
	}
}

func BenchmarkUser_GetFormattedTimeLeft(b *testing.B) {
	user := &User{SubscriptionEnd: time.Now().Add(5*time.Hour + 30*time.Minute)}
	for b.Loop() {
		user.GetFormattedTimeLeft()
	}
}

func BenchmarkUser_GetFormattedLastConnection(b *testing.B) {
	user := &User{
		LastConnectedNodeUUID: new("node-1"),
		OnlineAt:              new(time.Date(2025, 11, 15, 14, 30, 0, 0, time.UTC)),
	}
	for b.Loop() {
		user.GetFormattedLastConnection()
	}
}

func BenchmarkUser_GetSubscriptionURLWithPort(b *testing.B) {
	user := &User{SubscriptionURL: "https://example.com/subscribe"}
	port := 8443
	for b.Loop() {
		user.GetSubscriptionURLWithPort(&port)
	}
}

func BenchmarkUser_GetSubscriptionJSONURLWithPort(b *testing.B) {
	user := &User{SubscriptionURL: "https://example.com/subscribe"}
	port := 8443
	for b.Loop() {
		user.GetSubscriptionJSONURLWithPort(&port)
	}
}

func BenchmarkUser_AllMethods(b *testing.B) {
	now := time.Now()
	user := &User{
		TelegramID:            12345,
		Username:              "testuser",
		ID:                    1337,
		Status:                UserStatusActive,
		SubscriptionEnd:       now.Add(7 * 24 * time.Hour),
		TrafficLimit:          10737418240,
		TrafficUsed:           5368709120,
		SubscriptionURL:       "https://example.com/subscribe",
		LastConnectedNodeUUID: new("node-1"),
		OnlineAt:              new(now.Add(-2 * time.Hour)),
	}

	for b.Loop() {
		user.DaysLeft()
		user.IsExpired()
		user.IsActive()
		user.GetTimeUntilExpiration()
		user.ShouldShowTimeInHours()
		user.GetTrafficUsedGB()
		user.GetTrafficLimitGB()
		user.TrafficUsagePercent()
		user.GetStatusEmoji()
		user.GetExpirationEmoji()
		user.GetFormattedExpirationDate()
		user.GetFormattedTimeLeft()
		user.GetFormattedLastConnection()
		user.HasTrafficLimit()
		user.HasSubscriptionURL()
		user.HasLastConnection()
	}
}
