package mapper

import (
	"time"

	"github.com/unser-elefant/remnawave-go/domain"
	"github.com/unser-elefant/remnawave-go/internal/api/model"
)

func ToDomainUser(m *model.User) domain.User {
	if m == nil {
		return domain.User{}
	}

	d := domain.User{
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
		d.DeviceLimitSet = true
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

func ToDomainUsers(models []model.User) []domain.User {
	result := make([]domain.User, 0, len(models))
	for i := range models {
		result = append(result, ToDomainUser(&models[i]))
	}

	return result
}

func UpdateUserRequestToModel(userID uint64, domain *domain.UpdateUserRequest) model.UpdateUserRequest {
	req := model.UpdateUserRequest{
		ID:                   userID,
		Description:          domain.Description,
		Tag:                  domain.Tag,
		TelegramID:           &domain.TelegramID,
		Email:                domain.Email,
		HWIDDeviceLimit:      domain.HWIDDeviceLimit,
		ActiveInternalSquads: domain.ActiveInternalSquads,
		ExternalSquadUUID:    domain.ExternalSquadUUID,
	}

	if domain.Status != nil {
		status := string(*domain.Status)
		req.Status = &status
	}
	if domain.TrafficLimitBytes != nil {
		trafficLimitBytes := int64(*domain.TrafficLimitBytes)
		req.TrafficLimitBytes = &trafficLimitBytes
	}
	if domain.TrafficLimitStrategy != nil {
		strategy := string(*domain.TrafficLimitStrategy)
		req.TrafficLimitStrategy = &strategy
	}
	if domain.ExpireAt != nil {
		expireAt := domain.ExpireAt.Format(time.RFC3339Nano)
		req.ExpireAt = &expireAt
	}

	return req
}
