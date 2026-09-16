package httpclient

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()

	apiURL, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("failed to parse base URL: %v", err)
	}

	cfg := &Config{
		APIURL: apiURL,
		APIKey: "test-api-key",
	}

	return NewClient(slog.Default(), cfg)
}

func TestNewClient(t *testing.T) {
	t.Run("accepts nil logger and config", func(t *testing.T) {
		assert.NotNil(t, New(nil, nil))
	})

	t.Run("creates client with default timeout", func(t *testing.T) {
		apiURL, err := url.Parse("https://api.example.com")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		cfg := &Config{
			APIURL: apiURL,
			APIKey: "test-key",
		}

		client := NewClient(slog.Default(), cfg)

		assert.NotNil(t, client)
		assert.Equal(t, cfg.APIURL.String(), client.baseURL)
		assert.Equal(t, cfg.APIKey, client.apiKey)
		assert.NotNil(t, client.httpClient)
		assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
	})

	t.Run("creates client with TLS server name", func(t *testing.T) {
		apiURL, err := url.Parse("https://api.example.com")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}
		tlsName, err := url.Parse("custom.example.com")
		if err != nil {
			t.Fatalf("failed to parse TLS server name: %v", err)
		}

		cfg := &Config{
			APIURL:        apiURL,
			APIKey:        "test-key",
			TLSServerName: tlsName,
		}

		client := NewClient(slog.Default(), cfg)

		assert.Equal(t, "custom.example.com", client.tlsServerName)
	})

	t.Run("uses configured timeout", func(t *testing.T) {
		apiURL, err := url.Parse("https://api.example.com")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		client := New(slog.Default(), &Config{
			APIURL:  apiURL,
			Timeout: 3 * time.Second,
		})

		assert.Equal(t, 3*time.Second, client.httpClient.Timeout)
	})

	t.Run("uses provided HTTP client", func(t *testing.T) {
		customHTTPClient := &http.Client{Timeout: 37 * time.Second}
		client := New(slog.Default(), &Config{HTTPClient: customHTTPClient})

		assert.Same(t, customHTTPClient, client.httpClient)
		assert.Equal(t, 37*time.Second, client.httpClient.Timeout)
	})
}

func TestClient_Do(t *testing.T) {
	t.Run("successful request sets headers and returns response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/test", r.URL.Path)
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			reqID := r.Header.Get("X-Request-ID")
			corrID := r.Header.Get("X-Correlation-ID")
			assert.NotEmpty(t, reqID)
			assert.Equal(t, reqID, corrID)

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := setupTestClient(t, server.URL)
		req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		resp.Body.Close()
	})

	t.Run("returns ErrResourceNotFound on 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := setupTestClient(t, server.URL)
		req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)

		assert.ErrorIs(t, err, ErrResourceNotFound)
		var apiErr *APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Nil(t, resp)
	})

	t.Run("returns ErrBadRequest on 400", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid parameters"))
		}))
		defer server.Close()

		client := setupTestClient(t, server.URL)
		req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)

		assert.ErrorIs(t, err, ErrBadRequest)
		var apiErr *APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, "invalid parameters", apiErr.Body)
		assert.NotEmpty(t, apiErr.RequestID)
		assert.Contains(t, apiErr.URL, "/test")
		assert.Nil(t, resp)
	})

	t.Run("returns ErrUnauthorized on 401", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client := setupTestClient(t, server.URL)
		req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)

		assert.ErrorIs(t, err, ErrUnauthorized)
		assert.Nil(t, resp)
	})

	t.Run("returns ErrAPIError on 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := setupTestClient(t, server.URL)
		req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)

		assert.ErrorIs(t, err, ErrAPIError)
		assert.Nil(t, resp)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := setupTestClient(t, server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context canceled"))
		assert.Nil(t, resp)
	})

	t.Run("resolves relative URL using baseURL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := setupTestClient(t, server.URL)
		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.URL = &url.URL{Path: "/test"}

		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		resp.Body.Close()
	})

	t.Run("fails on invalid baseURL with relative request URL", func(t *testing.T) {
		client := setupTestClient(t, "http://example.com")
		client.baseURL = ":\\invalid-url"

		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.URL = &url.URL{Path: "/test"}

		resp, err := client.Do(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestClient_SetHeaders(t *testing.T) {
	t.Run("sets Host header when TLSServerName is configured", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "custom.example.com", r.Host)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		apiURL, err := url.Parse(server.URL)
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}
		tlsName, err := url.Parse("custom.example.com")
		if err != nil {
			t.Fatalf("failed to parse TLS server name: %v", err)
		}

		cfg := &Config{
			APIURL:        apiURL,
			APIKey:        "test-key",
			TLSServerName: tlsName,
		}
		client := NewClient(slog.Default(), cfg)
		req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		resp.Body.Close()
	})

	t.Run("does not set Host header when TLSServerName is empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Host)
			assert.True(t, strings.HasPrefix(r.Host, "127.0.0.1") || strings.HasPrefix(r.Host, "[::1]"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := setupTestClient(t, server.URL)
		req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		resp.Body.Close()
	})
}

func BenchmarkClient_Do(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiURL, err := url.Parse(server.URL)
	if err != nil {
		b.Fatalf("failed to parse URL: %v", err)
	}

	cfg := &Config{
		APIURL: apiURL,
		APIKey: "test-key",
	}
	client := NewClient(slog.Default(), cfg)

	for b.Loop() {
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}
}

func BenchmarkClient_ParallelDo(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiURL, err := url.Parse(server.URL)
	if err != nil {
		b.Fatalf("failed to parse URL: %v", err)
	}

	cfg := &Config{
		APIURL: apiURL,
		APIKey: "test-key",
	}
	client := NewClient(slog.Default(), cfg)

	if transport, ok := client.httpClient.Transport.(*http.Transport); ok {
		transport.MaxIdleConns = 100
		transport.MaxIdleConnsPerHost = 100
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
			resp, err := client.Do(req)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}
	})
}
