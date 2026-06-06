// Will move into refuse-cli/internal/shim/exec_test.go
package shim

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestFindRealSkipsSymlinkedShimDir exercises the cask-install topology:
// PATH has ~/.refuse/bin and the standard system bin. The refuse binary
// lives somewhere else (e.g. /opt/homebrew/Caskroom/refuse/x.y.z), and
// ~/.refuse/bin/npm is a symlink to it. findReal must NOT return that
// symlink (which would recurse back into shim mode) — it must keep
// walking PATH and find the real npm in the system bin.
func TestFindRealSkipsSymlinkedShimDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink topology not exercised on windows")
	}

	tmp := t.TempDir()
	// Layout:
	//   <tmp>/cask/refuse           — the "real" refuse binary
	//   <tmp>/shim/npm -> <tmp>/cask/refuse
	//   <tmp>/realbin/npm           — the manager we want findReal to pick
	cask := filepath.Join(tmp, "cask")
	shim := filepath.Join(tmp, "shim")
	real := filepath.Join(tmp, "realbin")
	for _, d := range []string{cask, shim, real} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	refuseBin := filepath.Join(cask, "refuse")
	if err := os.WriteFile(refuseBin, []byte("#!/bin/sh\necho refuse\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	npmShim := filepath.Join(shim, "npm")
	if err := os.Symlink(refuseBin, npmShim); err != nil {
		t.Fatal(err)
	}
	realNpm := filepath.Join(real, "npm")
	if err := os.WriteFile(realNpm, []byte("#!/bin/sh\necho real\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Pretend we were invoked via the symlinked shim.
	t.Setenv("PATH", shim+string(os.PathListSeparator)+real)
	origExe := osExecutable
	t.Cleanup(func() { osExecutable = origExe })
	osExecutable = func() (string, error) { return npmShim, nil }

	got, err := findReal("npm")
	if err != nil {
		t.Fatalf("findReal: %v", err)
	}
	if got != realNpm {
		t.Fatalf("findReal returned %q, want %q (the shim dir should be skipped)", got, realNpm)
	}
}

// TestFindRealSkipsExternalSymlinkToRefuse covers a defence-in-depth case:
// a PATH dir we don't recognise as the shim dir contains a file that's
// itself a symlink to the refuse binary. We must still skip it.
func TestFindRealSkipsExternalSymlinkToRefuse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink topology not exercised on windows")
	}
	tmp := t.TempDir()
	cask := filepath.Join(tmp, "cask")
	shim := filepath.Join(tmp, "shim")
	extra := filepath.Join(tmp, "extra")
	real := filepath.Join(tmp, "realbin")
	for _, d := range []string{cask, shim, extra, real} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	refuseBin := filepath.Join(cask, "refuse")
	if err := os.WriteFile(refuseBin, []byte("ref"), 0o755); err != nil {
		t.Fatal(err)
	}
	npmShim := filepath.Join(shim, "npm")
	if err := os.Symlink(refuseBin, npmShim); err != nil {
		t.Fatal(err)
	}
	// Extra PATH entry with a rogue symlink to the refuse binary.
	rogue := filepath.Join(extra, "npm")
	if err := os.Symlink(refuseBin, rogue); err != nil {
		t.Fatal(err)
	}
	realNpm := filepath.Join(real, "npm")
	if err := os.WriteFile(realNpm, []byte("#!/bin/sh\necho real\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH",
		shim+string(os.PathListSeparator)+
			extra+string(os.PathListSeparator)+
			real)
	origExe := osExecutable
	t.Cleanup(func() { osExecutable = origExe })
	osExecutable = func() (string, error) { return npmShim, nil }

	got, err := findReal("npm")
	if err != nil {
		t.Fatalf("findReal: %v", err)
	}
	if got != realNpm {
		t.Fatalf("findReal returned %q, want %q (rogue symlink to refuse should be skipped)", got, realNpm)
	}
}
