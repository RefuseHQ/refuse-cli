package shim

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/RefuseHQ/refuse-cli/internal/config"
)

// Default managers to install shims for. install/uninstall iterate over this.
// Order is stable for deterministic output.
var DefaultShims = []string{
	// npm family
	"npm", "pnpm", "yarn", "bun", "npx",
	// PyPI family
	"pip", "pip3", "uv", "poetry", "pipenv", "pdm", "pipx",
	// crates.io
	"cargo",
	// RubyGems
	"gem", "bundle",
	// Go modules
	"go",
	// Packagist
	"composer",
	// NuGet
	"dotnet",
}

// InstallResult is the structured summary `refuse install` reports back.
type InstallResult struct {
	BinDir        string
	Installed     []string
	Skipped       []string
	LegacyRemoved []string // bare-name shims from a prior version we cleaned up
	ShellRC       []string // shell rc files we edited
}

// Install drops symlinks for each shim into ~/.refuse/bin (creating the dir
// if needed) and edits the user's shell-rc files to prepend it to PATH.
func Install(shims []string) (InstallResult, error) {
	if len(shims) == 0 {
		shims = DefaultShims
	}
	dir, err := binDir()
	if err != nil {
		return InstallResult{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return InstallResult{}, err
	}

	selfExe, err := os.Executable()
	if err != nil {
		return InstallResult{}, fmt.Errorf("locate self: %w", err)
	}
	if resolved, rerr := filepath.EvalSymlinks(selfExe); rerr == nil {
		selfExe = resolved
	}

	res := InstallResult{BinDir: dir}
	for _, name := range shims {
		if _, ok := KnownManagers[name]; !ok {
			res.Skipped = append(res.Skipped, name+": not a known manager")
			continue
		}
		target := shimTargetPath(dir, name)

		// Detect anything already at target with Lstat, which (unlike os.Stat)
		// does NOT follow the link — so a dangling/broken shim is seen too.
		// os.Stat errors on a broken shim, which used to skip cleanup and make
		// createShim fail with "file exists" on the next install.
		if lfi, lerr := os.Lstat(target); lerr == nil {
			// A symlink that still resolves to us? Idempotent — leave it.
			if lfi.Mode()&os.ModeSymlink != 0 {
				if cur, sErr := os.Stat(target); sErr == nil {
					if self, e := os.Stat(selfExe); e == nil && os.SameFile(cur, self) {
						res.Skipped = append(res.Skipped, name+": already installed")
						// Still try the legacy cleanup so a re-run after upgrade tidies stragglers.
						if cleanupLegacyBareName(dir, name, selfExe) {
							res.LegacyRemoved = append(res.LegacyRemoved, name)
						}
						continue
					}
				}
			}
			// Foreign file, wrong target, or dangling shim — clear it before relinking.
			if rmErr := os.Remove(target); rmErr != nil {
				return res, fmt.Errorf("remove stale shim %s: %w", target, rmErr)
			}
		}

		if cleanupLegacyBareName(dir, name, selfExe) {
			res.LegacyRemoved = append(res.LegacyRemoved, name)
		}

		if err := createShim(selfExe, target); err != nil {
			return res, fmt.Errorf("create shim %s: %w", target, err)
		}
		res.Installed = append(res.Installed, name)
	}

	edited, err := updateShellRC(dir, true)
	if err != nil {
		return res, err
	}
	res.ShellRC = edited
	return res, nil
}

// Uninstall removes shim symlinks (only those that point at us — leaves
// foreign files alone) and reverts the shell-rc PATH edits.
func Uninstall() (InstallResult, error) {
	dir, err := binDir()
	if err != nil {
		return InstallResult{}, err
	}
	selfExe, err := os.Executable()
	if err != nil {
		return InstallResult{}, err
	}
	if resolved, rerr := filepath.EvalSymlinks(selfExe); rerr == nil {
		selfExe = resolved
	}

	res := InstallResult{BinDir: dir}
	selfInfo, err := os.Stat(selfExe)
	if err != nil {
		return res, err
	}
	for name := range KnownManagers {
		target := shimTargetPath(dir, name)

		// Lstat so a dangling shim is detected too — os.Stat reports ENOENT for a
		// broken link, which used to make us treat an existing shim as already gone.
		lfi, err := os.Lstat(target)
		if errors.Is(err, os.ErrNotExist) {
			// Even if the canonical target is gone, a legacy bare-name shim may linger.
			if cleanupLegacyBareName(dir, name, selfExe) {
				res.LegacyRemoved = append(res.LegacyRemoved, name)
			}
			continue
		}
		if err != nil {
			res.Skipped = append(res.Skipped, name+": "+err.Error())
			continue
		}
		// Ours if it resolves to us, or if it's a symlink (valid OR dangling)
		// pointing at the refuse binary in our own bin dir. Leave foreign files alone.
		ours := false
		if cur, sErr := os.Stat(target); sErr == nil {
			ours = os.SameFile(cur, selfInfo)
		} else if lfi.Mode()&os.ModeSymlink != 0 {
			if dest, rlErr := os.Readlink(target); rlErr == nil {
				if !filepath.IsAbs(dest) {
					dest = filepath.Join(dir, dest)
				}
				ours = filepath.Dir(filepath.Clean(dest)) == dir
			}
		}
		if !ours {
			res.Skipped = append(res.Skipped, name+": not our shim")
			continue
		}
		if err := os.Remove(target); err != nil {
			return res, err
		}
		res.Installed = append(res.Installed, name)
		if cleanupLegacyBareName(dir, name, selfExe) {
			res.LegacyRemoved = append(res.LegacyRemoved, name)
		}
	}

	edited, err := updateShellRC(dir, false)
	if err != nil {
		return res, err
	}
	res.ShellRC = edited
	return res, nil
}

func binDir() (string, error) {
	d, err := config.UserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "bin"), nil
}
