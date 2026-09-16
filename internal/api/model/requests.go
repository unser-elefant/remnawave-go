package model

import (
	"net/url"
	"strconv"

	"github.com/unser-elefant/remnawave-go/domain"
)

type UserStreamQuery struct {
	Cursor               *int64
	Size                 *int64
	Status               *domain.UserStatus
	TrafficLimitStrategy *domain.TrafficLimitStrategy
	TelegramID           *string
	Email                *string
	Tag                  *string
	ExternalSquadUUID    *string
}

func (q UserStreamQuery) Encode() string {
	values := url.Values{}

	if q.Cursor != nil {
		values.Set("cursor", strconv.FormatInt(*q.Cursor, 10))
	}
	if q.Size != nil {
		values.Set("size", strconv.FormatInt(*q.Size, 10))
	}
	if q.Status != nil {
		values.Set("status", string(*q.Status))
	}
	if q.TrafficLimitStrategy != nil {
		values.Set("trafficLimitStrategy", string(*q.TrafficLimitStrategy))
	}
	if q.TelegramID != nil {
		values.Set("telegramId", *q.TelegramID)
	}
	if q.Email != nil {
		values.Set("email", *q.Email)
	}
	if q.Tag != nil {
		values.Set("tag", *q.Tag)
	}
	if q.ExternalSquadUUID != nil {
		values.Set("externalSquadUuid", *q.ExternalSquadUUID)
	}

	return values.Encode()
}
