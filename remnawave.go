package remnawave

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/unser-elefant/remnawave-go/domain"
	"github.com/unser-elefant/remnawave-go/internal/api"
	"github.com/unser-elefant/remnawave-go/internal/httpclient"
	"github.com/unser-elefant/remnawave-go/internal/service"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type RegisterUserRequest = domain.RegisterUserRequest

type APIError = httpclient.APIError

var (
	ErrResourceNotFound = httpclient.ErrResourceNotFound
	ErrBadRequest       = httpclient.ErrBadRequest
	ErrUnauthorized     = httpclient.ErrUnauthorized
	ErrAPIError         = httpclient.ErrAPIError
)

type Config struct {
	APIURL             string
	APIKey             string
	HTTPClient         *http.Client
	InsecureSkipVerify bool
	TLSServerName      *url.URL
	Timeout            time.Duration
	Logger             Logger
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.APIURL) == "" {
		return errors.New("API URL is required")
	}

	parsed, err := url.Parse(c.APIURL)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("API URL must use http or https scheme")
	}
	if parsed.Host == "" {
		return errors.New("API URL must include a host")
	}

	return nil
}

type UserService interface {
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
	Register(ctx context.Context, request *RegisterUserRequest) error
	IsUserExists(ctx context.Context, telegramID int64) (bool, error)
	RevokeSubscription(ctx context.Context, id uint64) error
}

type DeviceService interface {
	GetDevices(ctx context.Context, userID uint64) ([]domain.Device, error)
	DeleteAllDevices(ctx context.Context, userID uint64) error
	GetDevicesCount(ctx context.Context, userID uint64) (int, error)
}

var (
	_ UserService   = (*service.UserService)(nil)
	_ DeviceService = (*service.DeviceService)(nil)
)

type Client struct {
	users   UserService
	devices DeviceService
}

func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	apiURL, err := url.Parse(cfg.APIURL)
	if err != nil {
		return nil, fmt.Errorf("invalid API URL: %w", err)
	}

	l := cfg.Logger
	if l == nil {
		l = slog.Default()
	}

	httpClient := httpclient.New(l, &httpclient.Config{
		APIURL:             apiURL,
		APIKey:             cfg.APIKey,
		HTTPClient:         cfg.HTTPClient,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		TLSServerName:      cfg.TLSServerName,
		Timeout:            cfg.Timeout,
	})

	userAPI := api.NewUserAPI(apiURL.String(), httpClient)
	deviceAPI := api.NewDeviceAPI(apiURL.String(), httpClient)

	return &Client{
		users:   service.NewUserService(userAPI, l),
		devices: service.NewDeviceService(deviceAPI, l),
	}, nil
}

func (c *Client) Users() UserService {
	if c == nil {
		return nil
	}
	return c.users
}

func (c *Client) Devices() DeviceService {
	if c == nil {
		return nil
	}
	return c.devices
}
