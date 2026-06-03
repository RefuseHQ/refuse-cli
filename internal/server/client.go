package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Client is the typed HTTP client for the refuse REST API.
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

// defaultTimeout balances "user is waiting at a terminal" against the
// realities of the hosted server. mcp.refuse.dev scales to zero and
// takes ~4.6s to cold-start; the original 1.5s timeout meant the first
// install after an idle period always failed open with a scary
// "server: unreachable" line. 8s absorbs a cold start with margin while
// still failing reasonably fast when the server is genuinely down.
// Tunable via REFUSE_TIMEOUT_MS.
const defaultTimeout = 8 * time.Second

// New returns a Client with sensible defaults. The HTTP timeout can be
// overridden by setting REFUSE_TIMEOUT_MS=<integer milliseconds>.
func New(baseURL, apiKey string) *Client {
	timeout := defaultTimeout
	if v := os.Getenv("REFUSE_TIMEOUT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: timeout},
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

// CheckDockerfile runs a check_dockerfile request.
func (c *Client) CheckDockerfile(ctx context.Context, req CheckDockerfileRequest) (BatchCheckResponse, error) {
	var out BatchCheckResponse
	err := c.post(ctx, "/api/v1/check/dockerfile", req, &out)
	return out, err
}

// CheckWorkflow runs a check_workflow request.
func (c *Client) CheckWorkflow(ctx context.Context, req CheckWorkflowRequest) (BatchCheckResponse, error) {
	var out BatchCheckResponse
	err := c.post(ctx, "/api/v1/check/workflow", req, &out)
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
		return fmt.Errorf("%w: %w", ErrServerUnreachable, err)
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
