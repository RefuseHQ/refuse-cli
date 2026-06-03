// Package pyhook embeds the Python pip-guard files + helpers for managing
// their installation into a Python env's site-packages.
//
// The hook is two files:
//
//	refuse_pip_guard.py   — module that patches pip.InstallCommand.run
//	refuse_pip_guard.pth  — single `import refuse_pip_guard` line
//
// Python's site initialization processes .pth files at interpreter
// startup; any line starting with `import` is exec'd. That triggers our
// module's _enable() side effect, which monkey-patches pip to first
// shell out to `refuse pip-gate <args>`.
package pyhook

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed refuse_pip_guard.py
var moduleSource []byte

//go:embed refuse_pip_guard.pth
var pthSource []byte

const (
	ModuleName = "refuse_pip_guard.py"
	PthName    = "refuse_pip_guard.pth"
)

// Files returns the two on-disk filenames (basename) and their contents.
// Useful for tests and for status reporting.
func Files() map[string][]byte {
	return map[string][]byte{
		ModuleName: moduleSource,
		PthName:    pthSource,
	}
}

// SitePackages returns the site-packages directory of the given Python
// executable (`python`, `python3`, /opt/.../python, …). It runs the
// interpreter with `-c 'import site; print(site.getsitepackages()[0])'`
// — the first entry is the per-prefix site dir, which is where pip
// installs by default.
func SitePackages(pythonBin string) (string, error) {
	if pythonBin == "" {
		pythonBin = "python3"
	}
	if _, err := exec.LookPath(pythonBin); err != nil {
		return "", fmt.Errorf("python interpreter %q not found on PATH", pythonBin)
	}
	out, err := exec.Command(pythonBin, "-c",
		`import site,sys; ps=site.getsitepackages(); print(ps[0] if ps else sys.prefix)`,
	).Output()
	if err != nil {
		return "", fmt.Errorf("run %s: %w", pythonBin, err)
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return "", errors.New("python returned an empty site-packages path")
	}
	return path, nil
}

// Install writes both files into siteDir, overwriting any existing copy.
// Returns the absolute paths actually written.
func Install(siteDir string) (modulePath, pthPath string, err error) {
	if err := os.MkdirAll(siteDir, 0o755); err != nil {
		return "", "", fmt.Errorf("ensure site dir: %w", err)
	}
	modulePath = filepath.Join(siteDir, ModuleName)
	pthPath = filepath.Join(siteDir, PthName)
	if err := os.WriteFile(modulePath, moduleSource, 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", modulePath, err)
	}
	if err := os.WriteFile(pthPath, pthSource, 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", pthPath, err)
	}
	return modulePath, pthPath, nil
}

// Uninstall removes both files from siteDir. Missing files are ignored;
// any other error is returned. Returns the paths that were removed.
func Uninstall(siteDir string) ([]string, error) {
	var removed []string
	for _, name := range []string{ModuleName, PthName} {
		p := filepath.Join(siteDir, name)
		err := os.Remove(p)
		if err == nil {
			removed = append(removed, p)
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return removed, fmt.Errorf("remove %s: %w", p, err)
		}
	}
	return removed, nil
}

// Status returns whether each of the two files is present under siteDir.
func Status(siteDir string) (moduleOK, pthOK bool) {
	if _, err := os.Stat(filepath.Join(siteDir, ModuleName)); err == nil {
		moduleOK = true
	}
	if _, err := os.Stat(filepath.Join(siteDir, PthName)); err == nil {
		pthOK = true
	}
	return moduleOK, pthOK
}
