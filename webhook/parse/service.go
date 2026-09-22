package parse

import (
	"fmt"

	"github.com/unser-elefant/remnawave-go/webhook/domain"
	"github.com/unser-elefant/remnawave-go/webhook/internal/model"
	"github.com/unser-elefant/remnawave-go/webhook/payload"
	"github.com/unser-elefant/remnawave-go/webhook/typed"
)

func ParseLoginWebhook(p *payload.RemnawaveWebhook) (domain.LoginAttempt, error) {
	dto, err := typed.ParseWebhookData[model.RemnawaveWebhookServiceEventsData](p)
	if err != nil {
		return domain.LoginAttempt{}, fmt.Errorf("parse user data: %w", err)
	}

	return toDomainLogin(&dto), nil
}

func toDomainLogin(m *model.RemnawaveWebhookServiceEventsData) domain.LoginAttempt {
	if m == nil {
		return domain.LoginAttempt{}
	}

	return domain.LoginAttempt{
		Username:    m.LoginAttempt.Username,
		Ip:          m.LoginAttempt.Ip,
		UserAgent:   m.LoginAttempt.UserAgent,
		Description: m.LoginAttempt.Description,
		Password:    m.LoginAttempt.Password,
	}

}
