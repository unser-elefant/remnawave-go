package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/unser-elefant/remnawave-go/internal/api/model"
)

type deviceAPI struct {
	baseURL string
	c       HTTPClient
}

func NewDeviceAPI(baseURL string, c HTTPClient) *deviceAPI {
	return &deviceAPI{
		baseURL: baseURL,
		c:       c,
	}
}

func (d *deviceAPI) GetByUserID(ctx context.Context, userID uint64) (*model.DevicesResponse, error) {
	if d == nil || d.c == nil {
		return nil, errors.New("device API client is not initialized")
	}
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	url := fmt.Sprintf("%s/api/hwid/devices/%d", d.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return nil, errors.New("failed to get devices: empty response")
	}
	defer resp.Body.Close()

	var result *model.DevicesResponse
	err = unmarshalData(resp.Body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	if result == nil {
		return nil, errors.New("response data is null")
	}

	return result, nil
}

func (d *deviceAPI) DeleteAll(ctx context.Context, userID uint64) error {
	if d == nil || d.c == nil {
		return errors.New("device API client is not initialized")
	}
	if ctx == nil {
		return errors.New("context is nil")
	}
	url := fmt.Sprintf("%s/api/hwid/devices/delete-all", d.baseURL)

	r, err := marshalData(map[string]uint64{
		"userId": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete all devices: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return errors.New("failed to delete all devices: empty response")
	}
	defer resp.Body.Close()

	return nil
}
