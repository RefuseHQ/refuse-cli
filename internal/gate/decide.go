// Package gate is the decision engine for both shims and agent hooks.
//
// Inputs: a ParseResult from the parsers package and a Client+Policy. Output:
// an exit-code-style Decision plus rendered messages.
package gate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/parsers"
	"github.com/RefuseHQ/refuse-cli/internal/server"
)

// Decision is the gate verdict.
type Decision int

const (
	// DecisionAllow — proceed with the original install.
	DecisionAllow Decision = iota
	// DecisionBlock — abort the install with a reason.
	DecisionBlock
)

// Result is the structured verdict + the rendered human message.
type Result struct {
	Decision     Decision
	Message      string                            // human-readable block reason, empty on allow
	Findings     []server.CheckPackageResponse     // per-package result for direct mode
	BatchSummary *server.BatchSummary              // populated for batch / lockfile
	ServerError  error                             // non-nil if the server call failed (fail-open path)
}

// Engine carries the deps the gate needs.
type Engine struct {
	Client *server.Client
	Policy config.Policy
}

// Decide runs the gate against a parser result.
func (e Engine) Decide(ctx context.Context, p parsers.ParseResult) Result {
	if !p.IsInstall || (p.Mode == parsers.ModeDirect && len(p.Packages) == 0) {
		// Not an install verb, or install with no positional packages we
		// can vet (transitive resolution happens on the lockfile side).
		return Result{Decision: DecisionAllow}
	}

	// Override env wins early so users can force a known-bad install through
	// after seeing the block message.
	if e.Policy.OverrideEnv != "" {
		if v, ok := envIfSet(e.Policy.OverrideEnv); ok && truthy(v) {
			return Result{Decision: DecisionAllow}
		}
	}

	if p.Mode == parsers.ModeLockfile {
		// v1: lockfile installs are passed through. The shim layer can hand
		// the lockfile path to `refuse check-lockfile` separately if the
		// user wants. Gating the install itself requires resolving the
		// lockfile against the server, which is its own subcommand.
		return Result{Decision: DecisionAllow}
	}

	pkgs := make([]server.CheckPackageRequest, 0, len(p.Packages))
	for _, ref := range p.Packages {
		pkgs = append(pkgs, server.CheckPackageRequest{
			Ecosystem: ref.Ecosystem,
			Name:      ref.Name,
			Version:   firstNonEmpty(ref.Version, "latest"),
		})
	}
	resp, err := e.Client.CheckBatch(ctx, server.BatchCheckRequest{Packages: pkgs})
	if err != nil {
		return e.failOpenOrClosed(err)
	}

	worst := worstSeverity(resp.Results)
	threshold := config.SeverityRank(e.Policy.SeverityThreshold)
	if worst >= threshold && worst > 0 {
		return Result{
			Decision:     DecisionBlock,
			Message:      renderBlock(p, resp),
			Findings:     resp.Results,
			BatchSummary: &resp.Summary,
		}
	}
	return Result{Decision: DecisionAllow, Findings: resp.Results, BatchSummary: &resp.Summary}
}

func (e Engine) failOpenOrClosed(err error) Result {
	if e.Policy.FailClosed {
		return Result{
			Decision:    DecisionBlock,
			Message:     fmt.Sprintf("refuse: %v (fail-closed)", err),
			ServerError: err,
		}
	}
	hint := ""
	if errors.Is(err, server.ErrServerUnreachable) {
		hint = " (set REFUSE_FAIL_CLOSED=1 to require gate)"
	} else if errors.Is(err, server.ErrRateLimited) {
		hint = " — rate limited, allowing install (upgrade plan to raise the limit)"
	} else if errors.Is(err, server.ErrUnauthorized) {
		hint = " — set REFUSE_API_KEY or run `refuse init`"
	}
	return Result{
		Decision:    DecisionAllow,
		Message:     fmt.Sprintf("refuse: %v%s", err, hint),
		ServerError: err,
	}
}

func renderBlock(p parsers.ParseResult, resp server.BatchCheckResponse) string {
	var b strings.Builder
	b.WriteString("refuse: blocked — vulnerable package(s)\n")
	for _, r := range resp.Results {
		if !r.Vulnerable {
			continue
		}
		for _, v := range r.Vulnerabilities {
			label := v.SeverityLabel
			if label == "" {
				label = "unknown"
			}
			id := v.RefuseID
			if v.CVE != nil && *v.CVE != "" {
				id = *v.CVE
			} else if v.GHSA != nil && *v.GHSA != "" {
				id = *v.GHSA
			}
			fmt.Fprintf(&b, "  %s@%s  %s  %s  %s\n", r.Package, r.Version, id, label, truncate(v.Summary, 80))
		}
		if len(r.SuggestedFixes) > 0 {
			f := r.SuggestedFixes[0]
			fmt.Fprintf(&b, "    Suggested: %s (%s)%s\n", f.Version, f.Type, breakingTag(f.BreakingChange))
		}
	}
	// Override hint
	b.WriteString("\nOverride: set REFUSE_ALLOW_VULNERABLE=1 to bypass.\n")
	return b.String()
}

func breakingTag(b bool) string {
	if b {
		return " ⚠ breaking change"
	}
	return ""
}

func worstSeverity(results []server.CheckPackageResponse) int {
	worst := 0
	for _, r := range results {
		if !r.Vulnerable {
			continue
		}
		for _, v := range r.Vulnerabilities {
			if rank := config.SeverityRank(v.SeverityLabel); rank > worst {
				worst = rank
			}
		}
	}
	return worst
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func truthy(s string) bool {
	switch strings.ToLower(s) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
