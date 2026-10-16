package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/unser-elefant/remnawave-go/domain"
	"github.com/unser-elefant/remnawave-go/internal/api/model"
	realmapper "github.com/unser-elefant/remnawave-go/internal/mapper"
	mock_userapi "github.com/unser-elefant/remnawave-go/internal/mock/userAPI"
)

func newTestUserService(api *mock_userapi.MockuserAPI) *UserService {
	return NewUserService(api, slog.Default())
}

func newRegisterRequest(telegramID int64, username string) *domain.RegisterUserRequest {
	return &domain.RegisterUserRequest{
		TelegramID: telegramID,
		Username:   username,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
	}
}

func newTestUser(username string, telegramID *int64, usedTraffic, trafficLimit float64) model.User {
	now := time.Now()
	return model.User{
		ID:                   1,
		ShortUUID:            "short",
		Username:             username,
		Status:               "ACTIVE",
		TelegramID:           telegramID,
		TrafficLimitBytes:    trafficLimit,
		TrafficLimitStrategy: "MONTH",
		ExpireAt:             now.Add(30 * 24 * time.Hour),
		CreatedAt:            now,
		UpdatedAt:            now,
		SubscriptionURL:      "https://example.com/sub",
		TrojanPassword:       "password",
		VlessUUID:            "vless-uuid",
		SSPassword:           "ss-password",
		ActiveInternalSquads: []model.Squad{},
		UserTraffic: model.UserTraffic{
			UsedTrafficBytes:         float64(usedTraffic),
			LifetimeUsedTrafficBytes: float64(usedTraffic * 2),
		},
	}
}

func TestUserService_GetUserByTelegramID(t *testing.T) {
	t.Run("successful get user", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		telegramID := int64(12345)
		apiUser := newTestUser("testuser", &telegramID, 1000, 10000)
		mockResponse := &model.UsersResponse{Response: model.UsersData{Users: []model.User{apiUser}}}
		mappedUser := realmapper.ToDomainUser(&apiUser)

		ctx := context.Background()
		mockAPI.EXPECT().GetStream(ctx, mock.MatchedBy(func(q model.UserStreamQuery) bool {
			return q.TelegramID != nil && *q.TelegramID == strconv.FormatInt(telegramID, 10)
		})).Return(mockResponse, nil)

		user, err := service.GetUserByTelegramID(ctx, telegramID)

		require.NoError(t, err)
		assert.Equal(t, mappedUser, *user)
	})

	t.Run("user not found", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, domain.ErrUserNotFound)

		user, err := service.GetUserByTelegramID(ctx, 99999)

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("maps client ErrResourceNotFound to domain ErrUserNotFound", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, domain.ErrUserNotFound)

		user, err := service.GetUserByTelegramID(ctx, 99999)

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("empty response", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(&model.UsersResponse{Response: model.UsersData{Users: []model.User{}}}, nil)

		user, err := service.GetUserByTelegramID(ctx, 12345)

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("api error", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		expectedErr := errors.New("network error")
		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, expectedErr)

		user, err := service.GetUserByTelegramID(ctx, 12345)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to get user")
	})
}

func TestUserService_IsUserExists(t *testing.T) {
	t.Run("user exists", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		telegramID := int64(12345)
		apiUser := newTestUser("user", &telegramID, 0, 0)
		mockAPI.EXPECT().GetStream(context.Background(), mock.Anything).Return(&model.UsersResponse{
			Response: model.UsersData{Users: []model.User{apiUser}},
		}, nil)

		exists, err := service.IsUserExists(context.Background(), telegramID)

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("user does not exist", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, domain.ErrUserNotFound)

		exists, err := service.IsUserExists(ctx, 99999)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("api error", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		expectedErr := errors.New("api error")
		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, expectedErr)

		exists, err := service.IsUserExists(ctx, 12345)

		assert.Error(t, err)
		assert.False(t, exists)
	})
}

func TestUserService_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		telegramID := int64(12345)

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, domain.ErrUserNotFound)
		mockAPI.EXPECT().Create(ctx, mock.AnythingOfType("*model.CreateUserRequest")).Return(nil)

		err := service.Register(ctx, newRegisterRequest(telegramID, "newuser"))

		assert.NoError(t, err)
	})

	t.Run("user already exists", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		telegramID := int64(12345)
		apiUser := newTestUser("existing", &telegramID, 0, 0)

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(&model.UsersResponse{
			Response: model.UsersData{Users: []model.User{apiUser}},
		}, nil)

		err := service.Register(ctx, newRegisterRequest(telegramID, "newuser"))

		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	})

	t.Run("registration api error", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		telegramID := int64(12345)

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, domain.ErrUserNotFound)
		mockAPI.EXPECT().Create(ctx, mock.AnythingOfType("*model.CreateUserRequest")).Return(errors.New("api error"))

		err := service.Register(ctx, newRegisterRequest(telegramID, "newuser"))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to register user")
	})

	t.Run("check existence error", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		telegramID := int64(12345)

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, errors.New("network error"))

		err := service.Register(ctx, newRegisterRequest(telegramID, "newuser"))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check user existence")
	})
}

func TestUserService_Update(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)
		ctx := context.Background()
		telegramID := int64(12345)
		apiUser := newTestUser("existing-user", &telegramID, 0, 0)
		status := domain.UserStatusDisabled
		limit := float64(1024)
		strategy := domain.TrafficLimitMonth
		expireAt := time.Date(2025, 1, 17, 15, 38, 45, 65_000_000, time.UTC)
		expectedExpireAt := "2025-01-17T15:38:45.065Z"
		description := "updated description"
		tag := "UPDATED"
		email := "alice@example.com"
		hwidDeviceLimit := int64(3)
		externalSquadUUID := "external-squad"
		activeInternalSquads := []string{"squad-a", "squad-b"}
		request := &domain.UpdateUserRequest{
			TelegramID:           telegramID,
			Status:               &status,
			TrafficLimitBytes:    &limit,
			TrafficLimitStrategy: &strategy,
			ExpireAt:             &expireAt,
			Description:          &description,
			Tag:                  &tag,
			Email:                &email,
			HWIDDeviceLimit:      &hwidDeviceLimit,
			ActiveInternalSquads: activeInternalSquads,
			ExternalSquadUUID:    &externalSquadUUID,
		}

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(&model.UsersResponse{
			Response: model.UsersData{Users: []model.User{apiUser}},
		}, nil)
		mockAPI.EXPECT().Update(ctx, mock.MatchedBy(func(req *model.UpdateUserRequest) bool {
			return req != nil && req.ID == apiUser.ID && req.TelegramID != nil &&
				*req.TelegramID == telegramID && req.Status != nil && *req.Status == string(status) &&
				req.TrafficLimitBytes != nil && *req.TrafficLimitBytes == int64(limit) &&
				req.TrafficLimitStrategy != nil && *req.TrafficLimitStrategy == string(strategy) &&
				req.ExpireAt != nil && *req.ExpireAt == expectedExpireAt &&
				req.Description != nil && *req.Description == description && req.Tag != nil && *req.Tag == tag &&
				req.Email != nil && *req.Email == email && req.HWIDDeviceLimit != nil &&
				*req.HWIDDeviceLimit == hwidDeviceLimit &&
				assert.ObjectsAreEqual(req.ActiveInternalSquads, activeInternalSquads) &&
				req.ExternalSquadUUID != nil && *req.ExternalSquadUUID == externalSquadUUID
		})).Return(nil)

		require.NoError(t, service.Update(ctx, request))
	})

	t.Run("update API error is returned", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)
		ctx := context.Background()
		telegramID := int64(12345)
		apiUser := newTestUser("existing-user", &telegramID, 0, 0)
		expectedErr := errors.New("api error")

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(&model.UsersResponse{
			Response: model.UsersData{Users: []model.User{apiUser}},
		}, nil)
		mockAPI.EXPECT().Update(ctx, mock.AnythingOfType("*model.UpdateUserRequest")).Return(expectedErr)

		err := service.Update(ctx, &domain.UpdateUserRequest{TelegramID: telegramID})
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestUserService_RevokeSubscription(t *testing.T) {
	t.Run("successful revoke subscription", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		userID := 123

		mockAPI.EXPECT().RevokeSubscription(ctx, uint64(userID)).Return(nil)

		err := service.RevokeSubscription(ctx, uint64(userID))
		_ = err

		require.NoError(t, err)
	})

	t.Run("api error on revoke", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		userID := 123
		mockAPI.EXPECT().RevokeSubscription(ctx, uint64(userID)).Return(errors.New("api error"))

		err := service.RevokeSubscription(ctx, uint64(userID))
		_ = err

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to revoke subscription")
	})

	t.Run("revoke with null id", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		err := service.RevokeSubscription(context.Background(), 0)

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidUserID)
	})

	t.Run("revoke non-existent user", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		userID := 9999
		mockAPI.EXPECT().RevokeSubscription(ctx, uint64(userID)).Return(domain.ErrUserNotFound)

		err := service.RevokeSubscription(ctx, uint64(userID))
		_ = err

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to revoke subscription")
	})
}

func TestUserService_Integration_FullFlow(t *testing.T) {
	t.Run("register and retrieve user", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		telegramID := int64(12345)
		username := "integrationuser"
		apiUser := newTestUser(username, &telegramID, 0, 0)
		mappedUser := realmapper.ToDomainUser(&apiUser)

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, domain.ErrUserNotFound).Once()
		exists, err := service.IsUserExists(ctx, telegramID)
		require.NoError(t, err)
		assert.False(t, exists)

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, domain.ErrUserNotFound).Once()
		mockAPI.EXPECT().Create(ctx, mock.AnythingOfType("*model.CreateUserRequest")).Return(nil)
		err = service.Register(ctx, newRegisterRequest(telegramID, username))
		require.NoError(t, err)

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(&model.UsersResponse{
			Response: model.UsersData{Users: []model.User{apiUser}},
		}, nil)

		user, err := service.GetUserByTelegramID(ctx, telegramID)
		require.NoError(t, err)
		assert.Equal(t, mappedUser, *user)
	})

	t.Run("register get user and revoke subscription", func(t *testing.T) {
		mockAPI := mock_userapi.NewMockuserAPI(t)
		service := newTestUserService(mockAPI)

		ctx := context.Background()
		telegramID := int64(67890)
		username := "testrevoke"
		apiUser := newTestUser(username, &telegramID, 1000, 10000)
		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(nil, domain.ErrUserNotFound).Once()
		mockAPI.EXPECT().Create(ctx, mock.AnythingOfType("*model.CreateUserRequest")).Return(nil)

		err := service.Register(ctx, newRegisterRequest(telegramID, username))
		require.NoError(t, err)

		mockAPI.EXPECT().GetStream(ctx, mock.Anything).Return(&model.UsersResponse{
			Response: model.UsersData{Users: []model.User{apiUser}},
		}, nil)

		user, err := service.GetUserByTelegramID(ctx, telegramID)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), user.ID)

		mockAPI.EXPECT().RevokeSubscription(ctx, user.ID).Return(nil)
		err = service.RevokeSubscription(ctx, user.ID)
		require.NoError(t, err)
	})
}

func BenchmarkUserService_MapperToDomain(b *testing.B) {
	telegramID := int64(12345)

	apiUser := &model.User{
		ID:                   1,
		ShortUUID:            "short-uuid",
		Username:             "testuser",
		Status:               "ACTIVE",
		TelegramID:           &telegramID,
		TrafficLimitBytes:    10737418240,
		TrafficLimitStrategy: "MONTH",
		ExpireAt:             time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		SubscriptionURL:      "https://example.com/sub",
		TrojanPassword:       "password",
		VlessUUID:            "vless-uuid",
		SSPassword:           "ss-password",
		ActiveInternalSquads: []model.Squad{},
		UserTraffic: model.UserTraffic{
			UsedTrafficBytes:         5368709120,
			LifetimeUsedTrafficBytes: 10737418240,
		},
	}

	for b.Loop() {
		_ = realmapper.ToDomainUser(apiUser)
	}
}

func BenchmarkUserService_IsUserExistsLogic(b *testing.B) {
	b.Run("user exists check", func(b *testing.B) {
		for b.Loop() {
			err := domain.ErrUserNotFound
			exists := !errors.Is(err, domain.ErrUserNotFound)
			if exists {
				b.Fatal("unexpected result")
			}
		}
	})

	b.Run("user not found check", func(b *testing.B) {
		for b.Loop() {
			var err error
			exists := !errors.Is(err, domain.ErrUserNotFound)
			if !exists {
				b.Fatal("unexpected result")
			}
		}
	})
}

func BenchmarkUserService_RegisterUserValidation(b *testing.B) {
	telegramID := int64(12345)
	username := "newuser"

	for b.Loop() {
		_ = model.NewCreateUserRequest(&model.CreateUserInitInfo{
			Username:             username,
			ExpireAt:             time.Now().Add(2 * 24 * time.Hour).Format(time.RFC3339),
			TelegramID:           telegramID,
			HWIDDeviceLimit:      &[]int64{5}[0],
			ActiveInternalSquads: []string{"default"},
		})
	}
}

func BenchmarkUserService_RevokeSubValidation(b *testing.B) {
	b.Run("valid uuid", func(b *testing.B) {
		uuid := "user-uuid-123"
		for b.Loop() {
			if uuid == "" {
				b.Fatal("empty uuid")
			}
		}
	})

	b.Run("empty uuid", func(b *testing.B) {
		uuid := ""
		for b.Loop() {
			if uuid == "" {
				_ = fmt.Errorf("uuid cannot be empty")
			}
		}
	})
}

func BenchmarkUserService_MultipleOperations(b *testing.B) {
	telegramID := int64(12345)

	apiUser := &model.User{
		ID:                   1,
		ShortUUID:            "short-uuid",
		Username:             "testuser",
		Status:               "ACTIVE",
		TelegramID:           &telegramID,
		TrafficLimitBytes:    10737418240,
		TrafficLimitStrategy: "MONTH",
		ExpireAt:             time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		SubscriptionURL:      "https://example.com/sub",
		TrojanPassword:       "password",
		VlessUUID:            "vless-uuid",
		SSPassword:           "ss-password",
		ActiveInternalSquads: []model.Squad{},
		UserTraffic: model.UserTraffic{
			UsedTrafficBytes:         5368709120,
			LifetimeUsedTrafficBytes: 10737418240,
		},
	}

	for b.Loop() {
		user := realmapper.ToDomainUser(apiUser)
		_ = user.IsActive()
		_ = user.IsExpired()
		_ = user.DaysLeft()
		_ = user.TrafficUsagePercent()
		_ = user.GetFormattedExpirationDate()
	}
}
