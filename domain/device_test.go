package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDevice_GetPlatformInfo(t *testing.T) {
	tests := []struct {
		name     string
		device   Device
		expected string
	}{
		{
			name: "full platform info",
			device: Device{
				Platform:  "Windows",
				OSVersion: "11",
			},
			expected: "Windows 11",
		},
		{
			name: "platform without version",
			device: Device{
				Platform: "Linux",
			},
			expected: "Linux",
		},
		{
			name:     "empty platform",
			device:   Device{},
			expected: "Неизвестная платформа",
		},
		{
			name: "platform with empty version",
			device: Device{
				Platform:  "macOS",
				OSVersion: "",
			},
			expected: "macOS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.device.GetPlatformInfo()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDevice_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		device   Device
		expected string
	}{
		{
			name: "full device info",
			device: Device{
				Platform:    "Android",
				OSVersion:   "13",
				DeviceModel: "Samsung Galaxy S23",
			},
			expected: "Android 13 (Samsung Galaxy S23)",
		},
		{
			name: "without device model",
			device: Device{
				Platform:  "iOS",
				OSVersion: "17",
			},
			expected: "iOS 17",
		},
		{
			name: "only platform",
			device: Device{
				Platform: "Linux",
			},
			expected: "Linux",
		},
		{
			name:     "empty device",
			device:   Device{},
			expected: "Неизвестная платформа",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.device.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDevice_GetTruncatedUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		maxLen    int
		expected  string
	}{
		{
			name:      "short user agent",
			userAgent: "Mozilla/5.0",
			maxLen:    20,
			expected:  "Mozilla/5.0",
		},
		{
			name:      "exact length",
			userAgent: "Mozilla/5.0",
			maxLen:    11,
			expected:  "Mozilla/5.0",
		},
		{
			name:      "needs truncation",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			maxLen:    20,
			expected:  "Mozilla/5.0 (Wind...",
		},
		{
			name:      "empty user agent",
			userAgent: "",
			maxLen:    20,
			expected:  "",
		},
		{
			name:      "very small max length",
			userAgent: "Test",
			maxLen:    3,
			expected:  "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := Device{UserAgent: tt.userAgent}
			result := device.GetTruncatedUserAgent(tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNilDeviceMethodsAreSafe(t *testing.T) {
	var device *Device

	assert.Empty(t, device.GetPlatformInfo())
	assert.Empty(t, device.GetDisplayName())
	assert.Empty(t, device.GetTruncatedUserAgent(20))
	assert.False(t, device.HasUserAgent())
	assert.False(t, device.HasRequestIP())
	assert.Empty(t, device.GetFormattedCreatedAt())
	assert.Equal(t, 0, device.GetDeviceAge())
}

func TestDevice_HasUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{
			name:      "has user agent",
			userAgent: "Mozilla/5.0",
			expected:  true,
		},
		{
			name:      "empty user agent",
			userAgent: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := Device{UserAgent: tt.userAgent}
			result := device.HasUserAgent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDevice_GetFormattedCreatedAt(t *testing.T) {
	tests := []struct {
		name      string
		createdAt time.Time
		expected  string
	}{
		{
			name:      "specific date",
			createdAt: time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC),
			expected:  "15/03/2024 14:30",
		},
		{
			name:      "beginning of year",
			createdAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:  "01/01/2025 00:00",
		},
		{
			name:      "end of year",
			createdAt: time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC),
			expected:  "31/12/2024 23:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := Device{CreatedAt: tt.createdAt}
			result := device.GetFormattedCreatedAt()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDevice_GetDeviceAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		createdAt time.Time
		expected  int
	}{
		{
			name:      "created today",
			createdAt: now,
			expected:  0,
		},
		{
			name:      "created 1 day ago",
			createdAt: now.Add(-24 * time.Hour),
			expected:  1,
		},
		{
			name:      "created 7 days ago",
			createdAt: now.Add(-7 * 24 * time.Hour),
			expected:  7,
		},
		{
			name:      "created 30 days ago",
			createdAt: now.Add(-30 * 24 * time.Hour),
			expected:  30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := Device{CreatedAt: tt.createdAt}
			result := device.GetDeviceAge()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ***************************
// ***** Benchmark tests *****
// ***************************

func BenchmarkDevice_GetPlatformInfo(b *testing.B) {
	device := Device{
		Platform:  "Windows",
		OSVersion: "11",
	}

	for b.Loop() {
		device.GetPlatformInfo()
	}
}

func BenchmarkDevice_GetDisplayName(b *testing.B) {
	device := Device{
		Platform:    "Android",
		DeviceModel: "Samsung Galaxy S21",
		OSVersion:   "12",
	}

	for b.Loop() {
		device.GetDisplayName()
	}
}

func BenchmarkDevice_GetTruncatedUserAgent(b *testing.B) {
	device := Device{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}

	for b.Loop() {
		device.GetTruncatedUserAgent(50)
	}
}

func BenchmarkDevice_GetFormattedCreatedAt(b *testing.B) {
	device := Device{
		CreatedAt: time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC),
	}

	for b.Loop() {
		device.GetFormattedCreatedAt()
	}
}

func BenchmarkDevice_GetDeviceAge(b *testing.B) {
	device := Device{
		CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
	}

	for b.Loop() {
		device.GetDeviceAge()
	}
}

func BenchmarkDevice_HasUserAgent(b *testing.B) {
	device := Device{
		UserAgent: "Mozilla/5.0",
	}

	for b.Loop() {
		device.HasUserAgent()
	}
}

func BenchmarkDevice_AllMethods(b *testing.B) {
	device := Device{
		HWID:        "device-hwid-123",
		Platform:    "Android",
		DeviceModel: "Samsung Galaxy S21",
		OSVersion:   "12",
		UserAgent:   "Mozilla/5.0 (Linux; Android 12) AppleWebKit/537.36",
		CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
	}

	for b.Loop() {
		device.GetPlatformInfo()
		device.GetDisplayName()
		device.GetTruncatedUserAgent(50)
		device.GetFormattedCreatedAt()
		device.GetDeviceAge()
		device.HasUserAgent()
	}
}
