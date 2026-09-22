package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unser-elefant/remnawave-go/domain"
	"github.com/unser-elefant/remnawave-go/internal/api/model"
	realmapper "github.com/unser-elefant/remnawave-go/internal/mapper"
	mock_deviceapi "github.com/unser-elefant/remnawave-go/internal/mock/deviceAPI"
)

func newTestDeviceService(api *mock_deviceapi.MockdeviceAPI) *DeviceService {
	return NewDeviceService(api, slog.Default())
}

func TestDeviceService_GetDevices(t *testing.T) {
	t.Run("successful get devices", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		platform := "Windows"
		osVersion := "11"
		apiDevices := []model.Device{{
			HWID:      "device-1",
			UserID:    123,
			Platform:  &platform,
			OSVersion: &osVersion,
		}}
		mockResponse := &model.DevicesResponse{
			Response: model.DevicesData{Devices: apiDevices, Total: 1},
		}
		domainDevices := []domain.Device{{
			HWID:      "device-1",
			UserID:    int64(123),
			Platform:  "Windows",
			OSVersion: "11",
		}}

		ctx := context.Background()
		mockAPI.EXPECT().GetByUserID(ctx, uint64(123)).Return(mockResponse, nil)

		devices, err := service.GetDevices(ctx, 123)

		require.NoError(t, err)
		assert.Equal(t, domainDevices, devices)
	})

	t.Run("api error", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		ctx := context.Background()
		mockAPI.EXPECT().GetByUserID(ctx, uint64(123)).Return(nil, errors.New("api error"))

		devices, err := service.GetDevices(ctx, 123)

		assert.Error(t, err)
		assert.Nil(t, devices)
		assert.Contains(t, err.Error(), "failed to get devices")
	})

	t.Run("empty devices list", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		mockResponse := &model.DevicesResponse{
			Response: model.DevicesData{Devices: []model.Device{}, Total: 0},
		}

		ctx := context.Background()
		mockAPI.EXPECT().GetByUserID(ctx, uint64(123)).Return(mockResponse, nil)

		devices, err := service.GetDevices(ctx, 123)

		require.NoError(t, err)
		assert.Empty(t, devices)
	})
}

func TestDeviceService_DeleteAllDevices(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		ctx := context.Background()
		mockAPI.EXPECT().DeleteAll(ctx, uint64(123)).Return(nil)

		err := service.DeleteAllDevices(ctx, 123)

		assert.NoError(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		ctx := context.Background()
		mockAPI.EXPECT().DeleteAll(ctx, uint64(123)).Return(errors.New("delete failed"))

		err := service.DeleteAllDevices(ctx, 123)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete devices")
	})
}

func TestDeviceService_GetDevicesCount(t *testing.T) {
	t.Run("successful get count", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		mockResponse := &model.DevicesResponse{
			Response: model.DevicesData{Devices: []model.Device{}, Total: 5},
		}

		ctx := context.Background()
		mockAPI.EXPECT().GetByUserID(ctx, uint64(123)).Return(mockResponse, nil)

		count, err := service.GetDevicesCount(ctx, 123)

		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("count error", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		ctx := context.Background()
		mockAPI.EXPECT().GetByUserID(ctx, uint64(123)).Return(nil, errors.New("api error"))

		count, err := service.GetDevicesCount(ctx, 123)

		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get devices count")
	})

	t.Run("zero count", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		mockResponse := &model.DevicesResponse{
			Response: model.DevicesData{Devices: []model.Device{}, Total: 0},
		}

		ctx := context.Background()
		mockAPI.EXPECT().GetByUserID(ctx, uint64(123)).Return(mockResponse, nil)

		count, err := service.GetDevicesCount(ctx, 123)

		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestDeviceService_Integration_Mapping(t *testing.T) {
	t.Run("correct mapping from API to domain", func(t *testing.T) {
		mockAPI := mock_deviceapi.NewMockdeviceAPI(t)
		service := newTestDeviceService(mockAPI)

		platform := "Android"
		osVersion := "13"
		deviceModel := "Pixel 7"
		userAgent := "Chrome/Mobile"
		apiDevices := []model.Device{{
			HWID:        "hwid-android",
			UserID:      789,
			Platform:    &platform,
			OSVersion:   &osVersion,
			DeviceModel: &deviceModel,
			UserAgent:   &userAgent,
		}}
		ctx := context.Background()
		mockAPI.EXPECT().GetByUserID(ctx, uint64(789)).Return(&model.DevicesResponse{
			Response: model.DevicesData{Devices: apiDevices, Total: 1},
		}, nil)

		devices, err := service.GetDevices(ctx, 789)

		require.NoError(t, err)
		require.Len(t, devices, 1)

		device := devices[0]
		assert.Equal(t, "hwid-android", device.HWID)
		assert.Equal(t, int64(789), device.UserID)
		assert.Equal(t, "Android", device.Platform)
		assert.Equal(t, "13", device.OSVersion)
		assert.Equal(t, "Pixel 7", device.DeviceModel)
		assert.Equal(t, "Chrome/Mobile", device.UserAgent)
		assert.Contains(t, device.GetDisplayName(), "Android 13")
		assert.Contains(t, device.GetDisplayName(), "Pixel 7")
		assert.True(t, device.HasUserAgent())
	})
}

func BenchmarkDeviceService_MapperToDomain(b *testing.B) {
	platform := "Windows"
	osVersion := "11"
	deviceModel := "Desktop PC"
	userAgent := "Mozilla/5.0"

	apiDevice := &model.Device{
		HWID:        "device-hwid-123",
		UserID:      456,
		Platform:    &platform,
		OSVersion:   &osVersion,
		DeviceModel: &deviceModel,
		UserAgent:   &userAgent,
	}

	for b.Loop() {
		_ = realmapper.ToDomainDevice(apiDevice)
	}
}

func BenchmarkDeviceService_MapperToDomainSlice(b *testing.B) {
	platform := "Android"
	osVersion := "13"
	deviceModel := "Pixel 7"
	userAgent := "Chrome Mobile"

	apiDevices := []model.Device{
		{HWID: "device-1", UserID: 1, Platform: &platform, OSVersion: &osVersion, DeviceModel: &deviceModel, UserAgent: &userAgent},
		{HWID: "device-2", UserID: 1, Platform: &platform, OSVersion: &osVersion, DeviceModel: &deviceModel, UserAgent: &userAgent},
		{HWID: "device-3", UserID: 1, Platform: &platform, OSVersion: &osVersion, DeviceModel: &deviceModel, UserAgent: &userAgent},
	}

	for b.Loop() {
		domainDevices := make([]domain.Device, 0, len(apiDevices))
		for i := range apiDevices {
			domainDevices = append(domainDevices, realmapper.ToDomainDevice(&apiDevices[i]))
		}
	}
}

func BenchmarkDeviceService_DeviceOperations(b *testing.B) {
	b.Run("GetDisplayName", func(b *testing.B) {
		device := &domain.Device{HWID: "device-ios", UserID: int64(123), Platform: "iOS", OSVersion: "17.0", DeviceModel: "iPhone 15 Pro"}
		for b.Loop() {
			_ = device.GetDisplayName()
		}
	})

	b.Run("HasUserAgent", func(b *testing.B) {
		device := &domain.Device{HWID: "device-test", UserID: int64(123), UserAgent: "Mozilla/5.0"}
		for b.Loop() {
			_ = device.HasUserAgent()
		}
	})

	b.Run("GetPlatformInfo", func(b *testing.B) {
		device := &domain.Device{HWID: "device-test", UserID: int64(123), Platform: "Android", OSVersion: "13"}
		for b.Loop() {
			_ = device.GetPlatformInfo()
		}
	})
}

func BenchmarkDeviceService_CountValidation(b *testing.B) {
	mockResponse := &model.DevicesResponse{
		Response: model.DevicesData{Devices: []model.Device{}, Total: 5},
	}

	for b.Loop() {
		count := mockResponse.Response.Total
		if count < 0 {
			b.Fatal("negative count")
		}
	}
}

func BenchmarkDeviceService_MultipleDevicesMapping(b *testing.B) {
	platforms := []string{"Windows", "Android", "iOS", "macOS", "Linux"}
	osVersions := []string{"11", "13", "17.0", "14.0", "Ubuntu 22.04"}
	models := []string{"Desktop", "Pixel 7", "iPhone 15", "MacBook Pro", "ThinkPad"}
	apiDevices := make([]model.Device, 10)
	for i := range apiDevices {
		idx := i % len(platforms)
		apiDevices[i] = model.Device{
			HWID:        fmt.Sprintf("device-%d", i),
			UserID:      1,
			Platform:    &platforms[idx],
			OSVersion:   &osVersions[idx],
			DeviceModel: &models[idx],
		}
	}

	for b.Loop() {
		domainDevices := make([]domain.Device, 0, len(apiDevices))
		for i := range apiDevices {
			dev := realmapper.ToDomainDevice(&apiDevices[i])
			domainDevices = append(domainDevices, dev)
			_ = dev.GetDisplayName()
			_ = dev.HasUserAgent()
			_ = dev.GetPlatformInfo()
		}
	}
}
