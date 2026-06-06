package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadMigratesUnversionedConfig verifies that a config file written
// before the Version field existed is treated as v1 instead of v0. The
// migration scaffolding's only real job today is to never let an old file
// silently re-stamp itself as "newer than current."
func TestLoadMigratesUnversionedConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("REFUSE_HOME", tmp)

	// A pre-versioning config — no config_version key.
	if err := os.WriteFile(
		filepath.Join(tmp, "config.yaml"),
		[]byte("server_url: https://example.invalid\napi_key: rfs_old\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Version != 1 {
		t.Fatalf("migrate() should have stamped Version=1, got %d", got.Version)
	}
	if got.ServerURL != "https://example.invalid" {
		t.Errorf("ServerURL drifted during migration: %q", got.ServerURL)
	}
}

// TestLoadKeepsCurrentVersion verifies that a config already at the current
// version doesn't get rewound or rewritten by migrate(). Cheap regression
// guard for the future when migrate() actually does work.
func TestLoadKeepsCurrentVersion(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("REFUSE_HOME", tmp)

	if err := Save(Defaults()); err != nil {
		t.Fatal(err)
	}

	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != CurrentConfigVersion {
		t.Fatalf("Load() returned Version=%d, want %d", got.Version, CurrentConfigVersion)
	}
}
