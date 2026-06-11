package server

import (
	"errors"
	"testing"
	"time"
)

func TestParseRateLimitBody_FullPayload(t *testing.T) {
	body := []byte(`{
		"error": "Quota exceeded",
		"quota": {
			"used": 100000,
			"limit": 100000,
			"plan": "free",
			"period": "2026-05-28",
			"period_end": "2026-06-27T08:11:17.540Z"
		},
		"upgrade": "https://refuse.dev/pricing"
	}`)

	re := parseRateLimitBody(body)
	if re == nil {
		t.Fatal("expected RateLimitedError, got nil")
	}
	if got, want := re.Used, int64(100000); got != want {
		t.Errorf("Used = %d, want %d", got, want)
	}
	if got, want := re.Limit, int64(100000); got != want {
		t.Errorf("Limit = %d, want %d", got, want)
	}
	if got, want := re.Plan, "free"; got != want {
		t.Errorf("Plan = %q, want %q", got, want)
	}
	if got, want := re.Period, "2026-05-28"; got != want {
		t.Errorf("Period = %q, want %q", got, want)
	}
	wantEnd, _ := time.Parse(time.RFC3339, "2026-06-27T08:11:17.540Z")
	if !re.PeriodEnd.Equal(wantEnd) {
		t.Errorf("PeriodEnd = %s, want %s", re.PeriodEnd, wantEnd)
	}
	if got, want := re.UpgradeURL, "https://refuse.dev/pricing"; got != want {
		t.Errorf("UpgradeURL = %q, want %q", got, want)
	}
}

func TestParseRateLimitBody_UnwrapsToErrRateLimited(t *testing.T) {
	body := []byte(`{"quota":{"used":1,"limit":2}}`)
	re := parseRateLimitBody(body)
	if re == nil {
		t.Fatal("expected RateLimitedError")
	}
	if !errors.Is(re, ErrRateLimited) {
		t.Error("errors.Is(re, ErrRateLimited) should be true")
	}
}

func TestParseRateLimitBody_RejectsNonQuotaBody(t *testing.T) {
	cases := map[string][]byte{
		"empty":        nil,
		"invalid json": []byte("not json at all"),
		"no quota":     []byte(`{"error":"rate limit exceeded by upstream proxy"}`),
		"zero limit":   []byte(`{"quota":{"used":5,"limit":0}}`),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if got := parseRateLimitBody(body); got != nil {
				t.Errorf("expected nil for %q body, got %+v", name, got)
			}
		})
	}
}

func TestRateLimitedError_ErrorString(t *testing.T) {
	re := &RateLimitedError{Used: 1, Limit: 2}
	// Intentionally short — callers prepend "refuse: " themselves.
	if got, want := re.Error(), "rate limited"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
