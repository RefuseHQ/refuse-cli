package pyhook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DiscoverInterpreters scans the filesystem for Python interpreters the
// user might `python -m pip install` against. It looks in:
//
//   - every directory on PATH (for `python`, `python3`, `python3.X`)
//   - common venv homes (~/.virtualenvs, $WORKON_HOME, $PWD/.venv, $PWD/venv)
//   - pyenv:   ~/.pyenv/versions/*/bin/python(3)
//   - conda:   ~/miniconda3/envs/*/bin/python, ~/anaconda3/envs/*/bin/python, ~/.conda/envs/*/bin/python
//
// Returns absolute, deduped (by real-path) interpreter paths.
func DiscoverInterpreters() []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		if p == "" {
			return
		}
		// Resolve symlinks so /usr/bin/python → real path; dedupe.
		real, err := filepath.EvalSymlinks(p)
		if err != nil {
			real = p
		}
		abs, err := filepath.Abs(real)
		if err != nil {
			abs = real
		}
		if seen[abs] {
			return
		}
		// Only keep things that look like executable Pythons.
		info, err := os.Stat(abs)
		if err != nil || info.IsDir() {
			return
		}
		if info.Mode()&0o111 == 0 {
			return
		}
		seen[abs] = true
		out = append(out, abs)
	}

	// 1. PATH.
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if !isPythonName(name) {
				continue
			}
			add(filepath.Join(dir, name))
		}
	}

	// 2. venv homes.
	home, _ := os.UserHomeDir()
	venvHomes := []string{
		filepath.Join(home, ".virtualenvs"),
		os.Getenv("WORKON_HOME"),
		".venv",
		"venv",
	}
	for _, h := range venvHomes {
		if h == "" {
			continue
		}
		// `.venv` / `venv` can also BE the env, not a parent. Try both.
		add(filepath.Join(h, "bin", "python"))
		add(filepath.Join(h, "bin", "python3"))
		// Or a parent dir holding multiple named envs.
		entries, err := os.ReadDir(h)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			add(filepath.Join(h, e.Name(), "bin", "python"))
			add(filepath.Join(h, e.Name(), "bin", "python3"))
		}
	}

	// 3. pyenv versions.
	pyenvRoot := os.Getenv("PYENV_ROOT")
	if pyenvRoot == "" {
		pyenvRoot = filepath.Join(home, ".pyenv")
	}
	if entries, err := os.ReadDir(filepath.Join(pyenvRoot, "versions")); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			add(filepath.Join(pyenvRoot, "versions", e.Name(), "bin", "python"))
			add(filepath.Join(pyenvRoot, "versions", e.Name(), "bin", "python3"))
		}
	}

	// 4. conda / mamba envs.
	for _, root := range []string{
		filepath.Join(home, "miniconda3"),
		filepath.Join(home, "anaconda3"),
		filepath.Join(home, ".conda"),
		"/opt/miniconda3",
		"/opt/anaconda3",
		"/opt/homebrew/Caskroom/miniconda/base",
	} {
		// The base env interpreter itself.
		add(filepath.Join(root, "bin", "python"))
		add(filepath.Join(root, "bin", "python3"))
		// Each named env.
		entries, err := os.ReadDir(filepath.Join(root, "envs"))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			add(filepath.Join(root, "envs", e.Name(), "bin", "python"))
			add(filepath.Join(root, "envs", e.Name(), "bin", "python3"))
		}
	}

	return out
}

// isPythonName matches `python`, `python3`, `python3.12`, etc., but not
// `pythonw` (Windows GUI shim) or anything with extra suffixes.
func isPythonName(name string) bool {
	switch {
	case name == "python", name == "python3":
		return true
	case strings.HasPrefix(name, "python3."):
		// `python3.12`, but reject `python3.12-config`, `python3.12m`, …
		rest := name[len("python3."):]
		if rest == "" {
			return false
		}
		for _, r := range rest {
			if r < '0' || r > '9' {
				return false
			}
		}
		return true
	}
	return false
}

// SitePackagesOf calls `python -c 'import site; print(site.getsitepackages()[0])'`
// on a concrete absolute path (skipping PATH resolution). Convenience
// wrapper used by the discover/install flow.
func SitePackagesOf(pythonBin string) (string, error) {
	out, err := exec.Command(pythonBin, "-c",
		`import site,sys; ps=site.getsitepackages(); print(ps[0] if ps else sys.prefix)`,
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
