package shim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const legacyCliBlock = "# >>> refuse cli (managed) >>>\n" +
	"export PATH=\"/home/u/.refuse/bin:$PATH\"\n" +
	"# <<< refuse cli (managed) <<<\n"

// Uninstall (enable=false) must remove BOTH the canonical "shim" block and the
// legacy "cli" block written by the old curl|sh installer — and leave the rest
// of the file untouched. Regression for the marker-mismatch that let an
// uninstall strand a refuse PATH export behind.
func TestPatchFileStripsLegacyAndCanonicalOnUninstall(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	content := "export FOO=1\n\n" + legacyCliBlock + "\n" + managedBlock("/home/u/.refuse/bin")
	if err := os.WriteFile(rc, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := patchFile(rc, "/home/u/.refuse/bin", false); err != nil {
		t.Fatal(err)
	}

	got := readFile(t, rc)
	if strings.Contains(got, "refuse cli (managed)") {
		t.Errorf("legacy cli block survived uninstall:\n%s", got)
	}
	if strings.Contains(got, "refuse shim (managed)") {
		t.Errorf("canonical shim block survived uninstall:\n%s", got)
	}
	if !strings.Contains(got, "export FOO=1") {
		t.Errorf("uninstall clobbered unrelated content:\n%s", got)
	}
}

// Install (enable=true) over a file that already has the legacy block (and even
// a duplicate) must consolidate to exactly one canonical block.
func TestPatchFileConsolidatesDuplicatesOnInstall(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	content := "export FOO=1\n\n" + legacyCliBlock + "\n" + legacyCliBlock
	if err := os.WriteFile(rc, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := patchFile(rc, "/home/u/.refuse/bin", true); err != nil {
		t.Fatal(err)
	}

	got := readFile(t, rc)
	if n := strings.Count(got, ">>> refuse"); n != 1 {
		t.Errorf("expected exactly one managed block, found %d:\n%s", n, got)
	}
	if strings.Contains(got, "refuse cli (managed)") {
		t.Errorf("legacy cli block not migrated to canonical:\n%s", got)
	}
	if !strings.Contains(got, "export FOO=1") {
		t.Errorf("install clobbered unrelated content:\n%s", got)
	}
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
