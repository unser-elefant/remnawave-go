package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/unser-elefant/remnawave-go/internal/api/model"
	mock_httpclient "github.com/unser-elefant/remnawave-go/internal/mock/HTTPClient"
)

func TestUserAPI_Update(t *testing.T) {
	client := mock_httpclient.NewMockHTTPClient(t)
	userAPI := NewUserAPI("https://api.example.com", client)
	status := "DISABLED"
	limit := int64(1024)
	var requestBody string

	client.EXPECT().Do(mock.Anything).RunAndReturn(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodPatch, req.Method)
		require.Equal(t, "https://api.example.com/api/users", req.URL.String())
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		requestBody = string(body)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil
	})

	err := userAPI.Update(context.Background(), &model.UpdateUserRequest{
		ID:                7,
		Status:            &status,
		TrafficLimitBytes: &limit,
	})

	require.NoError(t, err)
	require.JSONEq(t, `{"id":7,"status":"DISABLED","trafficLimitBytes":1024}`, requestBody)
}
