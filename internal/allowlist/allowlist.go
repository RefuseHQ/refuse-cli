// Package allowlist reads/writes the per-project `.refuse.yaml` that
// teams use to formally accept a known risk. An allowlisted (CVE,
// package?) pair is hidden from the gate's block decision, so day-to-
// day installs don't churn on the same alert.
//
// File shape (YAML, intentionally tiny):
//
//	version: 1
//	allowlist:
//	  - cve: CVE-2019-10744
//	    package: lodash         # optional — match all packages if omitted
//	    ecosystem: npm          # optional — only meaningful with `package`
//	    reason: "We don't use _.merge"
//	    expires: 2026-12-31     # optional ISO date; expired = not allowlisted
//
// Discovery walks from cwd up to find the nearest `.refuse.yaml`.
package allowlist

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// FileName is the canonical project-local allowlist filename.
const FileName = ".refuse.yaml"

// Entry is one accepted-risk row.
type Entry struct {
	CVE       string `yaml:"cve"`
	Package   string `yaml:"package,omitempty"`
	Ecosystem string `yaml:"ecosystem,omitempty"`
	Reason    string `yaml:"reason,omitempty"`
	Expires   string `yaml:"expires,omitempty"` // ISO 8601 date, "YYYY-MM-DD"
}

// File is the parsed `.refuse.yaml`.
type File struct {
	Version   int     `yaml:"version"`
	Allowlist []Entry `yaml:"allowlist"`
}

// Find walks from start up to root looking for FileName, returning the
// first hit or "" if none. Returns the absolute path.
func Find(start string) string {
	dir := start
	if abs, err := filepath.Abs(start); err == nil {
		dir = abs
	}
	for {
		candidate := filepath.Join(dir, FileName)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Load reads + parses `.refuse.yaml` at the given path. A missing file
// returns an empty File without error, so callers can always proceed.
func Load(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &File{Version: 1}, nil
		}
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if f.Version == 0 {
		f.Version = 1
	}
	return &f, nil
}

// Save writes the file with a header comment + canonical YAML.
func Save(path string, f *File) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.WriteString(out,
		"# refuse allowlist — formally accepted risks.\n"+
			"# Each entry suppresses one advisory in the gate's block decision.\n"+
			"# Expired entries (`expires:` in the past) are treated as not allowlisted.\n\n",
	); err != nil {
		return err
	}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	if err := enc.Encode(f); err != nil {
		_ = enc.Close()
		return err
	}
	return enc.Close()
}

// IsAllowed reports whether a given (cve, ecosystem, package) tuple is
// covered by the allowlist on `now`. Empty ecosystem/package on an entry
// means match-any. An entry whose `expires` is before `now` is ignored.
// Returns the matched entry on a hit so callers can render the reason.
func (f *File) IsAllowed(now time.Time, cve, ecosystem, pkg string) (Entry, bool) {
	if f == nil {
		return Entry{}, false
	}
	cveU := strings.ToUpper(strings.TrimSpace(cve))
	pkgL := strings.ToLower(strings.TrimSpace(pkg))
	ecoL := strings.ToLower(strings.TrimSpace(ecosystem))
	for _, e := range f.Allowlist {
		if strings.ToUpper(strings.TrimSpace(e.CVE)) != cveU {
			continue
		}
		if e.Package != "" && strings.ToLower(strings.TrimSpace(e.Package)) != pkgL {
			continue
		}
		if e.Ecosystem != "" && strings.ToLower(strings.TrimSpace(e.Ecosystem)) != ecoL {
			continue
		}
		if e.Expires != "" {
			t, err := time.Parse("2006-01-02", e.Expires)
			if err == nil && t.Before(now) {
				continue
			}
		}
		return e, true
	}
	return Entry{}, false
}

// Add appends an entry, deduping on (CVE, package, ecosystem). Returns
// whether the file was modified.
func (f *File) Add(e Entry) bool {
	for i, existing := range f.Allowlist {
		if matchesKey(existing, e) {
			f.Allowlist[i] = e
			return true
		}
	}
	f.Allowlist = append(f.Allowlist, e)
	return true
}

// Remove drops every entry whose key matches the given (cve, ecosystem,
// package). Returns the number of entries removed.
func (f *File) Remove(cve, ecosystem, pkg string) int {
	cveU := strings.ToUpper(strings.TrimSpace(cve))
	pkgL := strings.ToLower(strings.TrimSpace(pkg))
	ecoL := strings.ToLower(strings.TrimSpace(ecosystem))
	kept := f.Allowlist[:0]
	dropped := 0
	for _, e := range f.Allowlist {
		drop := strings.ToUpper(strings.TrimSpace(e.CVE)) == cveU &&
			strings.ToLower(strings.TrimSpace(e.Package)) == pkgL &&
			strings.ToLower(strings.TrimSpace(e.Ecosystem)) == ecoL
		if drop {
			dropped++
			continue
		}
		kept = append(kept, e)
	}
	f.Allowlist = kept
	return dropped
}

func matchesKey(a, b Entry) bool {
	return strings.EqualFold(strings.TrimSpace(a.CVE), strings.TrimSpace(b.CVE)) &&
		strings.EqualFold(strings.TrimSpace(a.Package), strings.TrimSpace(b.Package)) &&
		strings.EqualFold(strings.TrimSpace(a.Ecosystem), strings.TrimSpace(b.Ecosystem))
}
