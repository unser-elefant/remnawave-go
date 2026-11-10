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

func TestUserAPI_Revoke(t *testing.T) {
	tests := []struct {
		name     string
		call     func(u *userAPI) error
		wantBody string
	}{
		{
			name:     "revoke subscription sends no body",
			call:     func(u *userAPI) error { return u.RevokeSubscription(context.Background(), 7) },
			wantBody: "",
		},
		{
			name:     "revoke passwords sends revokeOnlyPasswords",
			call:     func(u *userAPI) error { return u.RevokePasswords(context.Background(), 7) },
			wantBody: `{"revokeOnlyPasswords":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mock_httpclient.NewMockHTTPClient(t)
			userAPI := NewUserAPI("https://api.example.com", client)
			var requestBody string

			client.EXPECT().Do(mock.Anything).RunAndReturn(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "https://api.example.com/api/users/7/actions/revoke", req.URL.String())
				if req.Body != nil {
					body, err := io.ReadAll(req.Body)
					require.NoError(t, err)
					requestBody = string(body)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil
			})

			require.NoError(t, tt.call(userAPI))
			require.Equal(t, tt.wantBody, requestBody)
		})
	}
}
