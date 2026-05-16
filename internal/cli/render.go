package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/RefuseHQ/refuse-cli/internal/server"
)

const (
	defaultCheckTimeout    = 5 * time.Second
	defaultLockfileTimeout = 30 * time.Second
)

func filepathBase(p string) string { return filepath.Base(p) }

func renderPackageResult(w io.Writer, r server.CheckPackageResponse) {
	if !r.Vulnerable {
		fmt.Fprintf(w, "%s@%s  clean\n", r.Package, r.Version)
		return
	}
	fmt.Fprintf(w, "%s@%s  VULNERABLE\n", r.Package, r.Version)
	for _, v := range r.Vulnerabilities {
		id := v.RefuseID
		if v.CVE != nil && *v.CVE != "" {
			id = *v.CVE
		} else if v.GHSA != nil && *v.GHSA != "" {
			id = *v.GHSA
		}
		fmt.Fprintf(w, "  %s  %s  %s\n", id, v.SeverityLabel, truncate(v.Summary, 80))
	}
	if len(r.SuggestedFixes) > 0 {
		f := r.SuggestedFixes[0]
		fmt.Fprintf(w, "  Suggested: %s (%s)\n", f.Version, f.Type)
	}
}

func renderBatchResult(w io.Writer, r server.BatchCheckResponse) {
	fmt.Fprintf(w, "Scanned: %d packages, %d vulnerable\n", r.Summary.Total, r.Summary.Vulnerable)
	for _, pr := range r.Results {
		if !pr.Vulnerable {
			continue
		}
		renderPackageResult(w, pr)
	}
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
