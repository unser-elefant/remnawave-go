package parse

import (
	"fmt"

	"github.com/unser-elefant/remnawave-go/internal/mapper"
	"github.com/unser-elefant/remnawave-go/webhook/domain"
	"github.com/unser-elefant/remnawave-go/webhook/internal/model"
	"github.com/unser-elefant/remnawave-go/webhook/payload"
	"github.com/unser-elefant/remnawave-go/webhook/typed"
)

func ParseHwidWebhook(p *payload.RemnawaveWebhook) (domain.UserHwidDevice, error) {
	dto, err := typed.ParseWebhookData[model.RemnawaveWebhookUserHwidDevicesEventsData](p)
	if err != nil {
		return domain.UserHwidDevice{}, fmt.Errorf("parse user data: %w", err)
	}

	return toDomainHwid(&dto), nil
}

func toDomainHwid(m *model.RemnawaveWebhookUserHwidDevicesEventsData) domain.UserHwidDevice {
	if m == nil {
		return domain.UserHwidDevice{}
	}

	return domain.UserHwidDevice{
		User:           mapper.ToDomainUser(&m.User),
		HwidUserDevice: mapper.ToDomainDevice(&m.HwidUserDevice),
	}

}
