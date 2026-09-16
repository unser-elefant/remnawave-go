package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/unser-elefant/remnawave-go/internal/api/model"
)

type userAPI struct {
	baseURL string
	c       HTTPClient
}

func NewUserAPI(baseURL string, c HTTPClient) *userAPI {
	return &userAPI{
		baseURL: baseURL,
		c:       c,
	}
}

func (u *userAPI) GetStream(ctx context.Context, query model.UserStreamQuery) (*model.UsersResponse, error) {
	if u == nil || u.c == nil {
		return nil, errors.New("user API client is not initialized")
	}
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	url := fmt.Sprintf("%s/api/users/stream", u.baseURL)
	if encodedQuery := query.Encode(); encodedQuery != "" {
		url += "?" + encodedQuery
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return nil, errors.New("failed to get user: empty response")
	}
	defer resp.Body.Close()

	var result *model.UsersResponse
	err = unmarshalData(resp.Body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	if result == nil {
		return nil, errors.New("response data is null")
	}

	return result, nil
}

func (u *userAPI) Create(ctx context.Context, req *model.CreateUserRequest) error {
	if u == nil || u.c == nil {
		return errors.New("user API client is not initialized")
	}
	if ctx == nil {
		return errors.New("context is nil")
	}
	if req == nil {
		return errors.New("create user request is nil")
	}
	url := fmt.Sprintf("%s/api/users", u.baseURL)

	r, err := marshalData(req)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.c.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return errors.New("failed to create user: empty response")
	}
	defer resp.Body.Close()

	return nil
}

func (u *userAPI) RevokeSubscription(ctx context.Context, userID uint64) error {
	if u == nil || u.c == nil {
		return errors.New("user API client is not initialized")
	}
	if ctx == nil {
		return errors.New("context is nil")
	}
	url := fmt.Sprintf("%s/api/users/%d/actions/revoke", u.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke subscription: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return errors.New("failed to revoke subscription: empty response")
	}
	defer resp.Body.Close()

	return nil
}
