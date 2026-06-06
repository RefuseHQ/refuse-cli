package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RefuseHQ/refuse-cli/internal/config"
)

// TestWriteDefaultConfigIfMissing covers the install-time auto-config path.
// The whole point is that a user can `refuse install` without first running
// `refuse init` and still get a working config on disk that the shims can
// read.
func TestWriteDefaultConfigIfMissing(t *testing.T) {
	t.Run("writes the default when nothing on disk", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("REFUSE_HOME", tmp)

		wrote, path, err := writeDefaultConfigIfMissing()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !wrote {
			t.Fatal("expected wrote=true on a clean filesystem, got false")
		}
		if path != filepath.Join(tmp, "config.yaml") {
			t.Fatalf("unexpected path: %q", path)
		}

		// The file should be readable + parse back to Defaults().
		got, err := config.Load()
		if err != nil {
			t.Fatalf("Load after write: %v", err)
		}
		want := config.Defaults()
		if got.ServerURL != want.ServerURL {
			t.Errorf("ServerURL = %q, want %q", got.ServerURL, want.ServerURL)
		}
		if got.Version != want.Version {
			t.Errorf("Version = %d, want %d", got.Version, want.Version)
		}
		if got.Policy.SeverityThreshold != want.Policy.SeverityThreshold {
			t.Errorf("SeverityThreshold = %q, want %q",
				got.Policy.SeverityThreshold, want.Policy.SeverityThreshold)
		}
	})

	t.Run("leaves an existing config alone", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("REFUSE_HOME", tmp)

		existing := []byte(
			"server_url: https://my.refuse.example\n" +
				"api_key: rfs_existing\n" +
				"policy:\n  severity_threshold: critical\n",
		)
		if err := os.WriteFile(filepath.Join(tmp, "config.yaml"), existing, 0o600); err != nil {
			t.Fatal(err)
		}

		wrote, _, err := writeDefaultConfigIfMissing()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wrote {
			t.Fatal("expected wrote=false when a config already exists")
		}

		// File on disk must be byte-identical to what we wrote.
		got, err := os.ReadFile(filepath.Join(tmp, "config.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(existing) {
			t.Errorf("existing config was clobbered\nwant:\n%s\n\ngot:\n%s",
				existing, got)
		}
	})
}

// TestDefaultsStampCurrentVersion guards against forgetting to update
// Defaults() when CurrentConfigVersion is bumped — the migrate() one-liner
// is the other half of the forward-compat story; this side anchors new
// installs to the latest schema.
func TestDefaultsStampCurrentVersion(t *testing.T) {
	d := config.Defaults()
	if d.Version != config.CurrentConfigVersion {
		t.Fatalf("Defaults().Version = %d, want %d (forward-compat invariant)",
			d.Version, config.CurrentConfigVersion)
	}
}
