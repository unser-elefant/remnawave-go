package mapper

import (
	"github.com/unser-elefant/remnawave-go/domain"
	"github.com/unser-elefant/remnawave-go/webhook/model"
)

func ToDomainUser(m *model.RemnawaveWebhookUserEventsData) *domain.User {
	if m == nil {
		return nil
	}

	d := &domain.User{
		ID:                    m.ID,
		ShortUUID:             m.ShortUUID,
		Username:              m.Username,
		Status:                domain.UserStatus(m.Status),
		SubscriptionEnd:       m.ExpireAt,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
		TrafficUsed:           m.UserTraffic.UsedTrafficBytes,
		TrafficLimit:          m.TrafficLimitBytes,
		TrafficStrategy:       domain.TrafficLimitStrategy(m.TrafficLimitStrategy),
		LastTrafficReset:      m.LastTrafficResetAt,
		SubscriptionURL:       m.SubscriptionURL,
		OnlineAt:              m.UserTraffic.OnlineAt,
		FirstConnectedAt:      m.UserTraffic.FirstConnectedAt,
		LastConnectedNodeUUID: m.UserTraffic.LastConnectedNodeUUID,
	}

	if m.TelegramID != nil {
		d.TelegramID = *m.TelegramID
	}

	if m.HWIDDeviceLimit != nil {
		d.DeviceLimit = *m.HWIDDeviceLimit
	}

	if m.Description != nil {
		d.Description = *m.Description
	}

	if m.Tag != nil {
		d.Tag = *m.Tag
	}

	if m.Email != nil {
		d.Email = *m.Email
	}

	return d
}

func ToDomainUsers(models []model.RemnawaveWebhookUserEventsData) []*domain.User {
	result := make([]*domain.User, 0, len(models))
	for i := range models {
		result = append(result, ToDomainUser(&models[i]))
	}

	return result
}
