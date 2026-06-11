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

// ErrRateLimited is returned on a 429. When the server provided a parseable
// quota body, a *RateLimitedError is returned instead — it unwraps to this
// sentinel so existing `errors.Is(err, ErrRateLimited)` checks keep working
// while richer callers can `errors.As` to extract used/limit/reset details.
var ErrRateLimited = errors.New("server: rate limited")

// RateLimitedError wraps ErrRateLimited with the per-account quota
// information mcp.refuse.dev returns alongside its 429 response. The body
// looks like:
//
//	{ "error": "Quota exceeded",
//	  "quota": { "used": …, "limit": …, "period": "YYYY-MM-DD",
//	             "period_end": "ISO-8601", "plan": "free" },
//	  "upgrade": "https://refuse.dev/pricing" }
//
// When the response is empty or malformed we fall back to the plain
// ErrRateLimited so callers always know they hit a rate limit.
type RateLimitedError struct {
	Used       int64
	Limit      int64
	Plan       string
	Period     string    // YYYY-MM-DD cycle start
	PeriodEnd  time.Time // zero if the server didn't send a usable date
	UpgradeURL string    // empty when none was offered (e.g. Pro users)
}

// Error implements error. The message intentionally omits "server: " so it
// renders naturally when callers prepend their own prefix ("refuse: %v").
func (e *RateLimitedError) Error() string { return "rate limited" }

// Unwrap lets `errors.Is(err, ErrRateLimited)` see through this wrapper.
func (e *RateLimitedError) Unwrap() error { return ErrRateLimited }

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
		body, _ := io.ReadAll(resp.Body)
		if re := parseRateLimitBody(body); re != nil {
			return re
		}
		return ErrRateLimited
	}
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("server: %d %s: %s", resp.StatusCode, resp.Status, string(b))
}

// parseRateLimitBody pulls the structured quota info out of mcp.refuse.dev's
// 429 body. Returns nil when the body is missing, unparseable, or doesn't
// look like a quota response — callers should fall back to ErrRateLimited.
func parseRateLimitBody(body []byte) *RateLimitedError {
	if len(body) == 0 {
		return nil
	}
	var payload struct {
		Quota struct {
			Used      int64  `json:"used"`
			Limit     int64  `json:"limit"`
			Plan      string `json:"plan"`
			Period    string `json:"period"`
			PeriodEnd string `json:"period_end"`
		} `json:"quota"`
		Upgrade string `json:"upgrade"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	// A response with no `limit` isn't a quota body — could be a generic
	// 429 from a different middleware (Cloudflare, an upstream proxy, etc).
	if payload.Quota.Limit == 0 {
		return nil
	}
	out := &RateLimitedError{
		Used:       payload.Quota.Used,
		Limit:      payload.Quota.Limit,
		Plan:       payload.Quota.Plan,
		Period:     payload.Quota.Period,
		UpgradeURL: payload.Upgrade,
	}
	if payload.Quota.PeriodEnd != "" {
		if t, err := time.Parse(time.RFC3339, payload.Quota.PeriodEnd); err == nil {
			out.PeriodEnd = t
		}
	}
	return out
}
