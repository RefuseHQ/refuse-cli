package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/server"
)

// auditConcurrency caps in-flight HTTP requests against the server.
const auditConcurrency = 6

type fileKind int

const (
	kindLockfile fileKind = iota
	kindDockerfile
	kindWorkflow
)

type auditFile struct {
	path string
	kind fileKind
}

// Skipped during the audit walk — these never contain project lockfiles
// worth scanning and they're often massive.
var excludeDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"venv":         true,
	"target":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".nuxt":        true,
	".cache":       true,
	".terraform":   true,
	".tox":         true,
	"__pycache__":  true,
}

// Basenames the server's check_lockfile understands; kept in lock-step
// with the help text on `refuse check-lockfile`.
var lockfileNames = map[string]bool{
	"package-lock.json":  true,
	"yarn.lock":          true,
	"pnpm-lock.yaml":     true,
	"bun.lock":           true,
	"bun.lockb":          true,
	"requirements.txt":   true,
	"Pipfile.lock":       true,
	"poetry.lock":        true,
	"pdm.lock":           true,
	"uv.lock":            true,
	"Cargo.lock":         true,
	"Gemfile.lock":       true,
	"go.sum":             true,
	"composer.lock":      true,
	"pom.xml":            true,
	"mix.lock":           true,
	"pubspec.lock":       true,
	"packages.lock.json": true,
}

func classify(path string) (fileKind, bool) {
	base := filepath.Base(path)
	if lockfileNames[base] {
		return kindLockfile, true
	}
	if strings.HasSuffix(base, ".csproj") {
		return kindLockfile, true
	}
	if base == "Dockerfile" || strings.HasSuffix(base, ".dockerfile") || strings.HasPrefix(base, "Dockerfile.") {
		return kindDockerfile, true
	}
	if (strings.HasSuffix(base, ".yaml") || strings.HasSuffix(base, ".yml")) &&
		strings.Contains(filepath.ToSlash(path), ".github/workflows/") {
		return kindWorkflow, true
	}
	return 0, false
}

func discover(root string) ([]auditFile, error) {
	var out []auditFile
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if excludeDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if k, ok := classify(path); ok {
			out = append(out, auditFile{path: path, kind: k})
		}
		return nil
	})
	return out, err
}

type auditResult struct {
	File       string                    `json:"file"`
	Kind       string                    `json:"kind"`
	Response   server.BatchCheckResponse `json:"response,omitempty"`
	Error      string                    `json:"error,omitempty"`
	DurationMs int64                     `json:"duration_ms"`
}

func kindLabel(k fileKind) string {
	switch k {
	case kindLockfile:
		return "lockfile"
	case kindDockerfile:
		return "dockerfile"
	case kindWorkflow:
		return "workflow"
	}
	return "unknown"
}

func scanOne(ctx context.Context, c *server.Client, f auditFile) auditResult {
	start := time.Now()
	res := auditResult{File: f.path, Kind: kindLabel(f.kind)}
	content, err := os.ReadFile(f.path)
	if err != nil {
		res.Error = err.Error()
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}
	filename := filepath.Base(f.path)
	var resp server.BatchCheckResponse
	switch f.kind {
	case kindLockfile:
		resp, err = c.CheckLockfile(ctx, server.CheckLockfileRequest{Filename: filename, Content: string(content)})
	case kindDockerfile:
		resp, err = c.CheckDockerfile(ctx, server.CheckDockerfileRequest{Filename: filename, Content: string(content)})
	case kindWorkflow:
		resp, err = c.CheckWorkflow(ctx, server.CheckWorkflowRequest{Filename: filename, Content: string(content)})
	}
	if err != nil {
		res.Error = err.Error()
	} else {
		res.Response = resp
	}
	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// renderAudit prints one line per file plus a totals block. Returns
// the number of files that had at least one vulnerable hit.
func renderAudit(out io.Writer, results []auditResult) int {
	var (
		totalPkgs    int
		totalVuln    int
		filesVulnHit int
		filesErrors  int
		bySev        = map[string]int{}
	)
	fmt.Fprintln(out, "refuse audit")
	for _, r := range results {
		summary := r.Response.Summary
		totalPkgs += summary.Total
		totalVuln += summary.Vulnerable
		for k, v := range summary.BySeverity {
			bySev[k] += v
		}
		var verdict string
		switch {
		case r.Error != "":
			verdict = fmt.Sprintf("ERROR %s", r.Error)
			filesErrors++
		case summary.Vulnerable > 0:
			verdict = fmt.Sprintf("%d/%d VULNERABLE", summary.Vulnerable, summary.Total)
			filesVulnHit++
		default:
			verdict = fmt.Sprintf("clean (%d)", summary.Total)
		}
		fmt.Fprintf(out, "  %-10s %-50s %s  (%dms)\n", r.Kind, truncatePath(r.File, 50), verdict, r.DurationMs)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintf(out, "Summary: %d files scanned, %d packages, %d vulnerable",
		len(results), totalPkgs, totalVuln)
	if filesErrors > 0 {
		fmt.Fprintf(out, ", %d errors", filesErrors)
	}
	fmt.Fprintln(out)
	if totalVuln > 0 {
		parts := []string{}
		for _, s := range []string{"critical", "high", "medium", "low"} {
			if n := bySev[s]; n > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", n, s))
			}
		}
		fmt.Fprintln(out, "  by severity: "+strings.Join(parts, ", "))
	}
	return filesVulnHit
}

func truncatePath(p string, max int) string {
	if len(p) <= max {
		return p
	}
	return "…" + p[len(p)-max+1:]
}

func realAuditCmd() *cobra.Command {
	var jsonOut bool
	var rootDir string
	cmd := &cobra.Command{
		Use:   "audit [path]",
		Short: "Scan every lockfile / Dockerfile / GH-Actions workflow in a tree",
		Long: `Walks the given directory (default: cwd), finds every supported
manifest, and runs check-lockfile / check-dockerfile / check-workflow on
each in parallel. Prints a per-file verdict plus a roll-up summary.
Exits 2 if any file contains a vulnerable package.

Common build / VCS / cache directories (.git, node_modules, vendor,
.venv, target, dist, build, __pycache__, etc.) are skipped.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir
			if len(args) > 0 {
				root = args[0]
			}
			if root == "" {
				root = "."
			}
			if _, err := os.Stat(root); err != nil {
				return fmt.Errorf("audit path %q: %w", root, err)
			}

			files, err := discover(root)
			if err != nil {
				return fmt.Errorf("walk: %w", err)
			}
			if len(files) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "refuse audit: no scannable files found")
				return nil
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			c := server.New(cfg.ServerURL, cfg.APIKey)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			results := make([]auditResult, len(files))
			sem := make(chan struct{}, auditConcurrency)
			var wg sync.WaitGroup
			for i, f := range files {
				wg.Add(1)
				sem <- struct{}{}
				go func(i int, f auditFile) {
					defer wg.Done()
					defer func() { <-sem }()
					results[i] = scanOne(ctx, c, f)
				}(i, f)
			}
			wg.Wait()

			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(results)
			}
			if renderAudit(cmd.OutOrStdout(), results) > 0 {
				os.Exit(2)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print raw JSON instead of the rendered summary")
	cmd.Flags().StringVarP(&rootDir, "dir", "C", "", "Directory to scan (defaults to the positional arg or cwd)")
	return cmd
}
