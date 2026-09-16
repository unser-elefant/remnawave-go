package httpclient

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/unser-elefant/remnawave-go/internal/observability/requestid"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrBadRequest       = errors.New("bad request")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrAPIError         = errors.New("API error")
)

const maxAPIErrorBodySize = 64 << 10

type APIError struct {
	StatusCode int
	RequestID  string
	URL        string
	Body       string
	Err        error
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err == nil {
		return fmt.Sprintf("API request failed with status %d", e.StatusCode)
	}
	return fmt.Sprintf("%s: status=%d", e.Err, e.StatusCode)
}

func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type Config struct {
	APIURL             *url.URL
	APIKey             string
	HTTPClient         *http.Client
	InsecureSkipVerify bool
	TLSServerName      *url.URL
	Timeout            time.Duration
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Client struct {
	httpClient    *http.Client
	l             Logger
	baseURL       string
	apiKey        string
	tlsServerName string
}

var errClientNotInitialized = errors.New("client is not initialized")

func New(l Logger, cfg *Config) *Client {
	if l == nil {
		l = slog.Default()
	}
	if cfg == nil {
		cfg = &Config{}
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	baseURL := ""
	if cfg.APIURL != nil {
		baseURL = cfg.APIURL.String()
	}

	httpClient := cfg.HTTPClient
	tlsServerName := ""
	if httpClient == nil {
		if cfg.InsecureSkipVerify {
			l.Warn("TLS certificate verification is disabled for Remnawave client")
		}

		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			MinVersion:         tls.VersionTLS12,
		}

		tlsServerName = serverNameFromURL(cfg.TLSServerName)
		if tlsServerName != "" {
			tlsConfig.ServerName = tlsServerName
		}

		protocols := new(http.Protocols)
		protocols.SetHTTP1(true)
		protocols.SetHTTP2(true)
		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
			Protocols:       protocols,
		}

		httpClient = &http.Client{
			Timeout:   timeout,
			Transport: transport,
		}
	}

	return &Client{
		httpClient:    httpClient,
		l:             l,
		baseURL:       baseURL,
		apiKey:        cfg.APIKey,
		tlsServerName: tlsServerName,
	}
}

func NewClient(l Logger, cfg *Config) *Client {
	return New(l, cfg)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c == nil || c.httpClient == nil || c.l == nil {
		return nil, errClientNotInitialized
	}
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.URL == nil {
		return nil, fmt.Errorf("request URL is nil")
	}

	ctx, reqID := requestid.GetOrCreate(req.Context())
	if ctx != req.Context() {
		req = req.Clone(ctx)
	}

	if req.URL != nil && !req.URL.IsAbs() && c.baseURL != "" {
		base, err := url.Parse(c.baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse base url: %w", err)
		}
		req.URL = base.ResolveReference(req.URL)
	}

	urlValue := ""
	if req.URL != nil {
		urlValue = req.URL.String()
	}

	c.l.Debug("HTTP request", "request_id", reqID, "method", req.Method, "url", urlValue)

	c.setHeaders(req, reqID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.l.Error("HTTP request failed", "request_id", reqID, "method", req.Method, "url", urlValue, "error", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	c.l.Debug("HTTP response received", "request_id", reqID, "method", req.Method, "url", urlValue, "status", resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		statusErr := c.mapStatusError(reqID, urlValue, resp)
		resp.Body.Close()
		return nil, statusErr
	}

	return resp, nil
}

func serverNameFromURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	if u.Host != "" {
		return u.Host
	}
	return u.Path
}

func (c *Client) setHeaders(req *http.Request, requestID string) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(requestid.HeaderRequestID, requestID)
	req.Header.Set(requestid.HeaderCorrelationID, requestID)

	if c.tlsServerName != "" {
		req.Host = c.tlsServerName
	}
}

func (c *Client) mapStatusError(requestID, urlValue string, resp *http.Response) error {
	respData, readErr := io.ReadAll(io.LimitReader(resp.Body, maxAPIErrorBodySize))
	respText := string(respData)

	var kind error = ErrAPIError
	switch resp.StatusCode {
	case http.StatusNotFound:
		kind = ErrResourceNotFound
	case http.StatusBadRequest:
		kind = ErrBadRequest
	case http.StatusUnauthorized:
		kind = ErrUnauthorized
	}

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		RequestID:  requestID,
		URL:        urlValue,
		Body:       respText,
		Err:        kind,
	}
	if readErr != nil {
		apiErr.Body = fmt.Sprintf("%s (failed to read response body: %v)", respText, readErr)
	}

	if errors.Is(kind, ErrBadRequest) {
		c.l.Warn("Bad request", "request_id", requestID, "url", urlValue, "response", respText)
	}

	if !errors.Is(kind, ErrBadRequest) {
		c.l.Warn("API returned non-OK status", "request_id", requestID, "url", urlValue, "status", resp.StatusCode, "response", respText)
	}

	return apiErr
}
