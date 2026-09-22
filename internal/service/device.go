package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/unser-elefant/remnawave-go/domain"
	"github.com/unser-elefant/remnawave-go/internal/api/model"
	"github.com/unser-elefant/remnawave-go/internal/httpclient"
	"github.com/unser-elefant/remnawave-go/internal/mapper"
)

type deviceAPI interface {
	GetByUserID(ctx context.Context, userID uint64) (*model.DevicesResponse, error)
	DeleteAll(ctx context.Context, userID uint64) error
}

type DeviceService struct {
	api deviceAPI
	l   httpclient.Logger
}

func NewDeviceService(api deviceAPI, l httpclient.Logger) *DeviceService {
	if l == nil {
		l = slog.Default()
	}
	return &DeviceService{
		api: api,
		l:   l,
	}
}

func (d *DeviceService) GetDevices(ctx context.Context, userID uint64) ([]domain.Device, error) {
	if d == nil || d.api == nil {
		return nil, errors.New("device service is not initialized")
	}
	response, err := d.api.GetByUserID(ctx, userID)
	if err != nil {
		d.l.Error("Failed to get devices from API", "user_uuid", userID, "error", err)
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	if response == nil {
		return nil, errors.New("failed to get devices: empty response")
	}

	devices := mapper.ToDomainDevices(response.Response.Devices)
	d.l.Debug("Devices retrieved successfully",
		"user_uuid", userID,
		"count", len(devices))

	return devices, nil
}

func (d *DeviceService) DeleteAllDevices(ctx context.Context, userID uint64) error {
	if d == nil || d.api == nil {
		return errors.New("device service is not initialized")
	}
	d.l.Info("Deleting all devices", "user_id", userID)

	if err := d.api.DeleteAll(ctx, userID); err != nil {
		d.l.Error("Failed to delete devices", "user_uuid", userID, "error", err)
		return fmt.Errorf("failed to delete devices: %w", err)
	}

	d.l.Info("All devices deleted successfully", "user_uuid", userID)

	return nil
}

func (d *DeviceService) GetDevicesCount(ctx context.Context, userID uint64) (int, error) {
	if d == nil || d.api == nil {
		return 0, errors.New("device service is not initialized")
	}
	response, err := d.api.GetByUserID(ctx, userID)
	if err != nil {

		return 0, fmt.Errorf("failed to get devices count: %w", err)
	}
	if response == nil {
		return 0, errors.New("failed to get devices count: empty response")
	}

	return response.Response.Total, nil
}
