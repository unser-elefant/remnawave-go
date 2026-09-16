package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/unser-elefant/remnawave-go/domain"
	"github.com/unser-elefant/remnawave-go/internal/api/model"
	"github.com/unser-elefant/remnawave-go/internal/httpclient"
	"github.com/unser-elefant/remnawave-go/internal/mapper"
)

type userAPI interface {
	GetStream(ctx context.Context, query model.UserStreamQuery) (*model.UsersResponse, error)
	Create(ctx context.Context, req *model.CreateUserRequest) error
	RevokeSubscription(ctx context.Context, id uint64) error
}

type UserService struct {
	api userAPI
	l   httpclient.Logger
}

func NewUserService(
	api userAPI,
	l httpclient.Logger,
) *UserService {
	if l == nil {
		l = slog.Default()
	}
	return &UserService{
		api: api,
		l:   l,
	}
}

func (u *UserService) GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	if u == nil || u.api == nil {
		return nil, errors.New("user service is not initialized")
	}
	query := model.UserStreamQuery{
		TelegramID: new(strconv.FormatInt(telegramID, 10)),
	}

	resp, err := u.api.GetStream(ctx, query)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, httpclient.ErrResourceNotFound) {
			u.l.Debug("User not found", "telegram_id", telegramID)
			return nil, domain.ErrUserNotFound
		}
		u.l.Error("Failed to get user from API", "telegram_id", telegramID, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if resp == nil {
		return nil, errors.New("failed to get user: empty response")
	}

	if len(resp.Response.Users) == 0 {
		u.l.Debug("User not found (empty response)", "telegram_id", telegramID)
		return nil, domain.ErrUserNotFound
	}

	if len(resp.Response.Users) > 1 {
		u.l.Warn("User has 2 or more profiles with the same telegram ID", "telegram_id", telegramID)
	}

	user := mapper.ToDomainUser(&resp.Response.Users[0])
	u.l.Debug("User retrieved successfully",
		"telegram_id", telegramID,
		"id", user.ID,
		"username", user.Username)

	return user, nil
}

func (u *UserService) Register(ctx context.Context, request domain.RegisterUserRequest) error {
	if u == nil || u.api == nil {
		return errors.New("user service is not initialized")
	}
	if err := request.Validate(); err != nil {
		return err
	}
	u.l.Info("Attempting to register user",
		"telegram_id", request.TelegramID,
		"username", request.Username)

	exists, err := u.IsUserExists(ctx, request.TelegramID)
	if err != nil {
		u.l.Error("Failed to check user existence", "telegram_id", request.TelegramID, "error", err)
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		u.l.Warn("User already exists", "telegram_id", request.TelegramID)
		return domain.ErrUserAlreadyExists
	}

	req := model.NewCreateUserRequest(&model.CreateUserInitInfo{
		Username:             request.Username,
		ExpireAt:             request.ExpiresAt.Format(time.RFC3339),
		TelegramID:           request.TelegramID,
		HWIDDeviceLimit:      request.HWIDDeviceLimit,
		ActiveInternalSquads: request.ActiveInternalSquads,
	})

	if err := u.api.Create(ctx, req); err != nil {
		u.l.Error("Failed to register user", "telegram_id", request.TelegramID, "error", err)
		return fmt.Errorf("failed to register user: %w", err)
	}

	u.l.Info("User registered successfully",
		"telegram_id", request.TelegramID,
		"username", request.Username)

	return nil
}

func (u *UserService) IsUserExists(ctx context.Context, telegramID int64) (bool, error) {
	if u == nil {
		return false, errors.New("user service is not initialized")
	}
	_, err := u.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (u *UserService) RevokeSubscription(ctx context.Context, id uint64) error {
	if u == nil || u.api == nil {
		return errors.New("user service is not initialized")
	}
	if id == 0 {
		return domain.ErrInvalidUserID
	}

	u.l.Info("Attempting to revoke subscription", "id", id)

	if err := u.api.RevokeSubscription(ctx, id); err != nil {
		u.l.Error("Failed to revoke subscription", "id", id, "error", err)
		return fmt.Errorf("failed to revoke subscription: %w", err)
	}

	u.l.Info("Subscription revoked successfully", "id", id)

	return nil
}
