package gate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
