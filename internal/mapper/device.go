package mapper

import (
	"github.com/unser-elefant/remnawave-go/domain"
	"github.com/unser-elefant/remnawave-go/internal/api/model"
)

func ToDomainDevice(m *model.Device) *domain.Device {
	if m == nil {
		return nil
	}

	d := &domain.Device{
		HWID:      m.HWID,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.Platform != nil {
		d.Platform = *m.Platform
	}

	if m.OSVersion != nil {
		d.OSVersion = *m.OSVersion
	}

	if m.DeviceModel != nil {
		d.DeviceModel = *m.DeviceModel
	}

	if m.UserAgent != nil {
		d.UserAgent = *m.UserAgent
	}

	if m.RequestIP != nil {
		d.RequestIP = *m.RequestIP
	}

	return d
}

func ToDomainDevices(models []model.Device) []*domain.Device {
	result := make([]*domain.Device, 0, len(models))
	for i := range models {
		result = append(result, ToDomainDevice(&models[i]))
	}

	return result
}
