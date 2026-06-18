package shim

import (
	"os"
	"path/filepath"
	"testing"
)

// A dangling shim (a symlink whose target no longer exists — e.g. left over
// after the refuse binary moved) must NOT block a re-install, and must be
// cleaned up by uninstall. Regression for:
//
//	Error: create shim .../bin/npm: symlink ...: file exists
//
// which happened because the old code probed the shim with os.Stat (which
// follows the link and errors on a broken one), so the stale link was never
// cleared before os.Symlink.
func TestInstallReplacesDanglingShim(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("REFUSE_HOME", "") // force the HOME-derived ~/.refuse path

	binPath := filepath.Join(home, ".refuse", "bin")
	if err := os.MkdirAll(binPath, 0o755); err != nil {
		t.Fatal(err)
	}

	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if r, e := filepath.EvalSymlinks(self); e == nil {
		self = r
	}

	npm := filepath.Join(binPath, "npm")
	// Pre-seed a DANGLING npm shim pointing at a refuse that no longer exists.
	if err := os.Symlink(filepath.Join(binPath, "refuse-gone"), npm); err != nil {
		t.Fatal(err)
	}

	if _, err := Install([]string{"npm"}); err != nil {
		t.Fatalf("Install over a dangling shim failed: %v", err)
	}

	resolved, err := filepath.EvalSymlinks(npm)
	if err != nil {
		t.Fatalf("npm shim still dangling after install: %v", err)
	}
	if resolved != self {
		t.Errorf("npm shim resolves to %s, want %s", resolved, self)
	}

	// Uninstall must remove a dangling shim too (one that points into our bin dir).
	if err := os.Remove(npm); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(binPath, "refuse"), npm); err != nil { // dangling, in our dir
		t.Fatal(err)
	}
	if _, err := Uninstall(); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}
	if _, err := os.Lstat(npm); !os.IsNotExist(err) {
		t.Errorf("uninstall left the dangling shim behind (lstat err = %v)", err)
	}
}
