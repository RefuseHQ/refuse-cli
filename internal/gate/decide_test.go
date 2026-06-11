package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/parsers"
	"github.com/RefuseHQ/refuse-cli/internal/server"
)

func fakeServer(t *testing.T, batch server.BatchCheckResponse, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(batch)
	}))
}

func TestGateAllowsClean(t *testing.T) {
	srv := fakeServer(t, server.BatchCheckResponse{
		Results: []server.CheckPackageResponse{{Vulnerable: false, Package: "playwright", Version: "1.60.0"}},
		Summary: server.BatchSummary{Total: 1, Vulnerable: 0},
	}, 200)
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high"}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "playwright", Version: "1.60.0"}},
	})
	if res.Decision != DecisionAllow {
		t.Errorf("expected allow, got block: %s", res.Message)
	}
}

func TestGateBlocksHighSeverity(t *testing.T) {
	srv := fakeServer(t, server.BatchCheckResponse{
		Results: []server.CheckPackageResponse{{
			Vulnerable: true, Package: "lodash", Version: "4.17.10",
			Vulnerabilities: []server.VulnerabilityRef{{
				RefuseID: "rfs-test-1", SeverityLabel: "high",
				Summary: "Prototype pollution in defaultsDeep",
			}},
			SuggestedFixes: []server.SuggestedFix{{Version: "4.17.21", Type: "minimum_safe"}},
		}},
		Summary: server.BatchSummary{Total: 1, Vulnerable: 1},
	}, 200)
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high"}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "lodash", Version: "4.17.10"}},
	})
	if res.Decision != DecisionBlock {
		t.Errorf("expected block, got allow")
	}
	if !strings.Contains(res.Message, "lodash") || !strings.Contains(res.Message, "4.17.21") {
		t.Errorf("block message missing expected content: %q", res.Message)
	}
}

func TestGateAllowsBelowThreshold(t *testing.T) {
	srv := fakeServer(t, server.BatchCheckResponse{
		Results: []server.CheckPackageResponse{{
			Vulnerable: true, Package: "low-vuln-pkg", Version: "1.0.0",
			Vulnerabilities: []server.VulnerabilityRef{{SeverityLabel: "low", Summary: "minor"}},
		}},
		Summary: server.BatchSummary{Total: 1, Vulnerable: 1},
	}, 200)
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high"}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "low-vuln-pkg", Version: "1.0.0"}},
	})
	if res.Decision != DecisionAllow {
		t.Errorf("expected allow (low < high threshold), got block")
	}
}

func TestGateFailOpenOnServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", 500)
	}))
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high"}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "x", Version: "1.0.0"}},
	})
	if res.Decision != DecisionAllow {
		t.Errorf("expected fail-open (allow), got block")
	}
}

func TestGateFailClosedOnServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", 500)
	}))
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high", FailClosed: true}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "x", Version: "1.0.0"}},
	})
	if res.Decision != DecisionBlock {
		t.Errorf("expected fail-closed (block), got allow")
	}
}

// quotaServer returns a server whose POSTs all 429 with a quota body shaped
// like mcp.refuse.dev's. periodEnd is rendered fresh on each call so the
// "resets in N days" math always reads sensibly relative to the test's
// wall clock.
func quotaServer(t *testing.T, used, limit int64, plan string, periodEnd time.Time, upgrade string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, `{
			"error":"Quota exceeded",
			"quota":{"used":%d,"limit":%d,"plan":%q,"period":"2026-05-28","period_end":%q},
			"upgrade":%q
		}`, used, limit, plan, periodEnd.Format(time.RFC3339), upgrade)
	}))
}

func TestGateRateLimitMessage_FailOpen(t *testing.T) {
	resetAt := time.Now().Add(16 * 24 * time.Hour)
	srv := quotaServer(t, 100_000, 100_000, "free", resetAt, "https://refuse.dev/pricing")
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high"}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "x", Version: "1.0.0"}},
	})
	if res.Decision != DecisionAllow {
		t.Fatalf("expected DecisionAllow, got %v", res.Decision)
	}
	must := func(s string) {
		t.Helper()
		if !strings.Contains(res.Message, s) {
			t.Errorf("message missing %q\n--- got ---\n%s", s, res.Message)
		}
	}
	must("refuse: rate limited")
	must("account quota 100,000/100,000 used")
	must("https://refuse.dev/pricing")
	must("REFUSE_FAIL_CLOSED=1 to block")
	if strings.Contains(res.Message, "fail-open") || strings.Contains(res.Message, "fail-closed") {
		t.Errorf("avoid jargon in the rendered message\n--- got ---\n%s", res.Message)
	}
}

func TestGateRateLimitMessage_FailClosed(t *testing.T) {
	resetAt := time.Now().Add(2 * 24 * time.Hour)
	srv := quotaServer(t, 5_000, 5_000, "free", resetAt, "")
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high", FailClosed: true}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "x", Version: "1.0.0"}},
	})
	if res.Decision != DecisionBlock {
		t.Fatalf("expected DecisionBlock under fail-closed, got %v", res.Decision)
	}
	if !strings.Contains(res.Message, "install blocked") {
		t.Errorf("expected 'install blocked' in message:\n%s", res.Message)
	}
	// No upgrade URL was supplied — message shouldn't make one up.
	if strings.Contains(res.Message, "https://") {
		t.Errorf("message synthesized an upgrade URL the server didn't return:\n%s", res.Message)
	}
}

func TestGateRateLimit_FallbackWhenNoBody(t *testing.T) {
	// Some 429s come from upstream proxies with no quota body — we should
	// still fall back to the plain "rate limited" message and fail-open
	// without crashing.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high"}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "x", Version: "1.0.0"}},
	})
	if res.Decision != DecisionAllow {
		t.Fatalf("expected DecisionAllow on bodyless 429, got %v", res.Decision)
	}
	if !strings.Contains(res.Message, "rate limited") {
		t.Errorf("expected 'rate limited' in fallback message:\n%s", res.Message)
	}
}

func TestGateAllowsLockfileMode(t *testing.T) {
	srv := fakeServer(t, server.BatchCheckResponse{}, 200)
	defer srv.Close()

	eng := Engine{Client: server.New(srv.URL, ""), Policy: config.Policy{SeverityThreshold: "high"}}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeLockfile, LockfileHint: "package-lock.json",
	})
	if res.Decision != DecisionAllow {
		t.Errorf("lockfile mode should allow (v1 passes through)")
	}
}

func TestGateAllowsOverrideEnv(t *testing.T) {
	srv := fakeServer(t, server.BatchCheckResponse{
		Results: []server.CheckPackageResponse{{
			Vulnerable: true, Package: "lodash", Version: "4.17.10",
			Vulnerabilities: []server.VulnerabilityRef{{SeverityLabel: "critical", Summary: "x"}},
		}},
		Summary: server.BatchSummary{Total: 1, Vulnerable: 1},
	}, 200)
	defer srv.Close()

	t.Setenv("REFUSE_ALLOW_VULNERABLE", "1")
	eng := Engine{
		Client: server.New(srv.URL, ""),
		Policy: config.Policy{SeverityThreshold: "high", OverrideEnv: "REFUSE_ALLOW_VULNERABLE"},
	}
	res := eng.Decide(context.Background(), parsers.ParseResult{
		IsInstall: true, Mode: parsers.ModeDirect,
		Packages: []parsers.PkgRef{{Ecosystem: "npm", Name: "lodash", Version: "4.17.10"}},
	})
	if res.Decision != DecisionAllow {
		t.Errorf("expected override to allow, got block")
	}
}
