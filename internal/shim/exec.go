package shim

import (
	"fmt"
	"os"
	"path/filepath"
)

// osExecutable is a seam over os.Executable so tests can simulate a
// symlinked-shim launch topology without needing to actually re-exec.
var osExecutable = os.Executable

// SelfDirs returns the directories we must NOT pick a "real" package
// manager from when resolving via PATH — both the directory of the running
// binary as launched (the symlink path, e.g. ~/.refuse/bin/npm), and the
// directory of the real binary the symlink points at (e.g. the Homebrew
// cellar). The Homebrew/cask install puts the shim on PATH as a symlink
// into ~/.refuse/bin pointing at the actual refuse binary somewhere else;
// only excluding the resolved path lets ~/.refuse/bin (which IS on PATH)
// match again and recurse.
func SelfDirs() ([]string, error) {
	exe, err := osExecutable()
	if err != nil {
		return nil, err
	}
	dirs := []string{filepath.Dir(exe)}
	if resolved, rerr := filepath.EvalSymlinks(exe); rerr == nil {
		if rd := filepath.Dir(resolved); rd != dirs[0] {
			dirs = append(dirs, rd)
		}
	}
	return dirs, nil
}

// SelfDir is kept for backwards-compatible callers that only need the
// resolved binary directory. Prefer SelfDirs for any PATH-walking.
func SelfDir() (string, error) {
	exe, err := osExecutable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	return filepath.Dir(resolved), nil
}

// findReal walks PATH (minus the shim dirs — both the launched symlink dir
// and the resolved target dir) and returns the first executable named
// `name`. If a candidate is itself a symlink that points back at the refuse
// binary, it's treated as a shim and skipped too — covers the case where
// the user has symlinked the shim somewhere besides ~/.refuse/bin.
func findReal(name string) (string, error) {
	dirs, err := SelfDirs()
	if err != nil {
		return "", err
	}
	skip := make(map[string]struct{}, len(dirs))
	for _, d := range dirs {
		skip[d] = struct{}{}
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		if _, ok := skip[dir]; ok {
			continue
		}
		cand := filepath.Join(dir, name)
		info, err := os.Stat(cand)
		if err != nil || info.IsDir() || info.Mode()&0o111 == 0 {
			continue
		}
		// Defence in depth: even if a non-shim PATH dir happens to hold a
		// symlink pointing back at us, treat it as another shim and skip.
		if resolved, rerr := filepath.EvalSymlinks(cand); rerr == nil {
			if _, ok := skip[filepath.Dir(resolved)]; ok {
				continue
			}
		}
		return cand, nil
	}
	return "", fmt.Errorf("refuse: real %q not found on PATH (self dirs %v excluded)", name, dirs)
}
