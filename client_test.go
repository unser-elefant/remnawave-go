package remnawave_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	remnawave "github.com/unser-elefant/remnawave-go"
)

func TestNewBuildsPublicServices(t *testing.T) {
	client, err := remnawave.New(remnawave.Config{
		APIURL: "https://api.example.com",
		APIKey: "test-key",
	})

	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.Users())
	require.NotNil(t, client.Devices())
}

func TestNewAcceptsSlogLogger(t *testing.T) {
	client, err := remnawave.New(remnawave.Config{
		APIURL: "https://api.example.com",
		Logger: slog.Default(),
	})

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewUsesProvidedHTTPClient(t *testing.T) {
	customHTTPClient := &http.Client{}

	client, err := remnawave.New(remnawave.Config{
		APIURL:     "https://api.example.com",
		HTTPClient: customHTTPClient,
	})

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewRequiresAPIURL(t *testing.T) {
	_, err := remnawave.New(remnawave.Config{})

	require.EqualError(t, err, "API URL is required")
}

func TestNewRejectsInvalidAPIURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		err  string
	}{
		{name: "missing scheme", url: "api.example.com", err: "API URL must use http or https scheme"},
		{name: "missing host", url: "https://", err: "API URL must include a host"},
		{name: "unsupported scheme", url: "ftp://api.example.com", err: "API URL must use http or https scheme"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := remnawave.New(remnawave.Config{APIURL: test.url})
			require.EqualError(t, err, test.err)
		})
	}
}

func TestClientUsersGetUserIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/users/stream", r.URL.Path)
		require.Equal(t, "12345", r.URL.Query().Get("telegramId"))
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		require.NotEmpty(t, r.Header.Get("X-Request-ID"))
		require.Equal(t, r.Header.Get("X-Request-ID"), r.Header.Get("X-Correlation-ID"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"users":[{"id":42,"username":"alice","status":"ACTIVE","expireAt":"2030-01-02T03:04:05Z","trafficLimitBytes":1073741824,"userTraffic":{"usedTrafficBytes":536870912}}]}}`))
	}))
	defer server.Close()

	client, err := remnawave.New(remnawave.Config{
		APIURL: server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	user, err := client.Users().GetUserByTelegramID(t.Context(), 12345)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, uint64(42), user.ID)
	require.Equal(t, "alice", user.Username)
	require.Equal(t, "ACTIVE", string(user.Status))
	require.Equal(t, int64(536870912), user.TrafficUsed)
	require.Equal(t, int64(1073741824), user.TrafficLimit)
}

func TestClientUsersGetUserExposesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid key"))
	}))
	defer server.Close()

	client, err := remnawave.New(remnawave.Config{
		APIURL: server.URL,
	})
	require.NoError(t, err)

	_, err = client.Users().GetUserByTelegramID(t.Context(), 12345)

	var apiErr *remnawave.APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
	require.Equal(t, "invalid key", apiErr.Body)
	require.NotEmpty(t, apiErr.RequestID)
	require.True(t, errors.Is(err, remnawave.ErrUnauthorized))
}
