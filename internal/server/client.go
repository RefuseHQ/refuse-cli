package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the typed HTTP client for the refuse REST API.
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

// New returns a Client with sensible defaults (1.5 s timeout — install gates
// run interactively and shouldn't wait long).
func New(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: 1500 * time.Millisecond},
	}
}

// ErrUnauthorized is returned on a 401 from the server (missing or invalid key).
var ErrUnauthorized = errors.New("server: unauthorized")

// ErrRateLimited is returned on a 429.
var ErrRateLimited = errors.New("server: rate limited")

// ErrServerUnreachable is returned on connection / timeout errors.
var ErrServerUnreachable = errors.New("server: unreachable")

// CheckPackage runs a single check_package request.
func (c *Client) CheckPackage(ctx context.Context, req CheckPackageRequest) (CheckPackageResponse, error) {
	var out CheckPackageResponse
	err := c.post(ctx, "/api/v1/check/package", req, &out)
	return out, err
}

// CheckBatch runs a batch_check request.
func (c *Client) CheckBatch(ctx context.Context, req BatchCheckRequest) (BatchCheckResponse, error) {
	var out BatchCheckResponse
	err := c.post(ctx, "/api/v1/check/batch", req, &out)
	return out, err
}

// CheckLockfile runs a check_lockfile request.
func (c *Client) CheckLockfile(ctx context.Context, req CheckLockfileRequest) (BatchCheckResponse, error) {
	var out BatchCheckResponse
	err := c.post(ctx, "/api/v1/check/lockfile", req, &out)
	return out, err
}

func (c *Client) post(ctx context.Context, path string, in any, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrServerUnreachable, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return json.NewDecoder(resp.Body).Decode(out)
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	}
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("server: %d %s: %s", resp.StatusCode, resp.Status, string(b))
}
