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
var DefaultShims = []string{"npm", "pnpm", "yarn", "bun", "pip", "pip3"}

// InstallResult is the structured summary `refuse install` reports back.
type InstallResult struct {
	BinDir    string
	Installed []string
	Skipped   []string
	ShellRC   []string // shell rc files we edited
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
		target := filepath.Join(dir, name)
		// If a symlink already exists pointing at us, treat as installed.
		if cur, err := os.Readlink(target); err == nil {
			if cur == selfExe {
				res.Skipped = append(res.Skipped, name+": already installed")
				continue
			}
			// Overwrite a stale link.
			_ = os.Remove(target)
		}
		// Also remove any leftover regular file (e.g. from a prior bash-script
		// shim).
		if _, err := os.Stat(target); err == nil {
			_ = os.Remove(target)
		}
		if err := os.Symlink(selfExe, target); err != nil {
			return res, fmt.Errorf("symlink %s: %w", target, err)
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
	for name := range KnownManagers {
		target := filepath.Join(dir, name)
		link, err := os.Readlink(target)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			res.Skipped = append(res.Skipped, name+": "+err.Error())
			continue
		}
		if link != selfExe {
			res.Skipped = append(res.Skipped, name+": not our symlink ("+link+")")
			continue
		}
		if err := os.Remove(target); err != nil {
			return res, err
		}
		res.Installed = append(res.Installed, name)
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
