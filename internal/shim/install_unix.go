//go:build !windows

package shim

import (
	"os"
	"path/filepath"
)

// shimTargetPath returns the path of the shim file for the given manager name.
// On POSIX we keep the bare name — the resolver doesn't need an extension and
// historical shims have always been extensionless.
func shimTargetPath(dir, name string) string {
	return filepath.Join(dir, name)
}

// createShim drops a symlink at target pointing at selfExe. POSIX symlinks
// are unprivileged; this should succeed for any user with write access to dir.
func createShim(selfExe, target string) error {
	return os.Symlink(selfExe, target)
}

// cleanupLegacyBareName is a no-op on POSIX — bare names ARE the canonical
// shim layout here.
func cleanupLegacyBareName(_ string, _ string, _ string) bool {
	return false
}
