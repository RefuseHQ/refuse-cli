//go:build windows

package shim

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// shimTargetPath returns the .exe-suffixed shim path. PATHEXT on Windows
// requires an executable extension; a bare-name file is invisible to the
// shell's command resolution. We've shipped extensionless shims in
// historical versions — see cleanupLegacyBareName for the upgrade path.
func shimTargetPath(dir, name string) string {
	return filepath.Join(dir, name+".exe")
}

// createShim tries three tiers in order:
//
//  1. os.Link (NTFS hardlink) — shares the same MFT record as refuse.exe,
//     zero extra disk, no special privilege. Fails across volumes and on
//     non-NTFS filesystems.
//  2. os.Symlink — works for unprivileged users with Developer Mode on,
//     or admin. Fails with ERROR_PRIVILEGE_NOT_HELD otherwise.
//  3. io.Copy — last resort, ~8 MB per shim. Always works.
//
// Whichever tier succeeds, the resulting file is a real .exe that the
// Windows shell resolves via PATHEXT.
func createShim(selfExe, target string) error {
	if err := os.Link(selfExe, target); err == nil {
		return nil
	}
	// Symlink tier — covers cross-volume hardlinks failing.
	if err := os.Symlink(selfExe, target); err == nil {
		return nil
	}
	// Copy tier — last resort. Open self for read, create target for write.
	src, err := os.Open(selfExe)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		_ = os.Remove(target)
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return nil
}

// cleanupLegacyBareName checks for a file at <dir>/<name> (no extension)
// left behind by historical refuse install runs that predated the .exe
// suffix. If the file exists and SameFile-matches selfExe, it's a stale
// shim from a prior version — remove it and tell the caller. Anything
// else stays untouched.
func cleanupLegacyBareName(dir, name, selfExe string) bool {
	legacy := filepath.Join(dir, name)
	info, err := os.Stat(legacy)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		return false
	}
	self, err := os.Stat(selfExe)
	if err != nil {
		return false
	}
	if !os.SameFile(info, self) {
		return false
	}
	if err := os.Remove(legacy); err != nil {
		return false
	}
	return true
}
