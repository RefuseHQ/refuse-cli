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

		// Idempotency: if target already points at us, nothing to do.
		if cur, err := os.Stat(target); err == nil {
			if self, sErr := os.Stat(selfExe); sErr == nil && os.SameFile(cur, self) {
				res.Skipped = append(res.Skipped, name+": already installed")
				// Still try the legacy cleanup so a re-run after upgrade tidies stragglers.
				if cleanupLegacyBareName(dir, name, selfExe) {
					res.LegacyRemoved = append(res.LegacyRemoved, name)
				}
				continue
			}
			// Foreign file at our spot — remove it before relink.
			_ = os.Remove(target)
		}
		// Stale entry that isn't a file but isn't a known link either (e.g., broken symlink).
		// os.Stat above fails for broken symlinks; nothing to do then.

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
		info, err := os.Stat(target)
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
		if !os.SameFile(info, selfInfo) {
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
