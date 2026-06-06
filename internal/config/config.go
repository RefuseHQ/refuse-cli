// Package config loads + validates the refuse CLI config.
//
// Layered resolution (highest priority first):
//  1. command-line flags                       (handled by caller)
//  2. env vars (REFUSE_SERVER_URL, REFUSE_API_KEY, REFUSE_FAIL_CLOSED, REFUSE_POLICY)
//  3. project file:  ./.refuse.yaml  (walk up from cwd)
//  4. user file:     ~/.refuse/config.yaml
//  5. built-in defaults
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultServerURL = "https://mcp.refuse.dev"
	UserConfigFile   = "config.yaml"
	ProjectFile      = ".refuse.yaml"
	// CurrentConfigVersion is the schema version stamped on every config we
	// write. Existing configs without a version are treated as v1. Bump this
	// (and add a one-block migration in Load) when defaults change in a way
	// that should reach existing users (e.g. flipping fail_closed=true).
	CurrentConfigVersion = 1
)

// Config is the resolved config. Nil-able pointers are deliberate so we can
// tell "the user explicitly set X to false" from "X was unset".
type Config struct {
	// Version is the config schema version. 0 / missing means "written before
	// versioning was introduced — treat as v1." Bump CurrentConfigVersion when
	// the meaning of any field changes.
	Version   int    `yaml:"config_version,omitempty"`
	ServerURL string `yaml:"server_url"`
	APIKey    string `yaml:"api_key"`
	Policy    Policy `yaml:"policy"`
}

// Policy is the gate's decision policy.
type Policy struct {
	// SeverityThreshold is the minimum severity at which we block. Valid
	// values: "low", "medium", "high", "critical". Default: "high".
	SeverityThreshold string `yaml:"severity_threshold"`
	// FailClosed = true makes network errors and 5xx responses a block.
	// Default: false (fail open with a stderr warning).
	FailClosed bool `yaml:"fail_closed"`
	// AllowYanked allows installs of yanked versions when no advisory matches.
	// Default: false.
	AllowYanked bool `yaml:"allow_yanked"`
	// AllowPrerelease allows installs of prerelease versions.
	// Default: false.
	AllowPrerelease bool `yaml:"allow_prerelease"`
	// OverrideEnv is the env var name a user can set to force a block to
	// pass through. Default: "REFUSE_ALLOW_VULNERABLE".
	OverrideEnv string `yaml:"override_env"`
}

// Defaults returns the built-in default Config. Exported so the install
// command can stamp it onto disk as the auto-created config — keeping the
// "what defaults look like" answer in one place.
func Defaults() Config {
	return Config{
		Version:   CurrentConfigVersion,
		ServerURL: DefaultServerURL,
		Policy: Policy{
			SeverityThreshold: "high",
			FailClosed:        false,
			AllowYanked:       false,
			AllowPrerelease:   false,
			OverrideEnv:       "REFUSE_ALLOW_VULNERABLE",
		},
	}
}

// UserDir returns ~/.refuse (or $REFUSE_HOME if set).
func UserDir() (string, error) {
	if h := os.Getenv("REFUSE_HOME"); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".refuse"), nil
}

// UserConfigPath returns ~/.refuse/config.yaml.
func UserConfigPath() (string, error) {
	d, err := UserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, UserConfigFile), nil
}

// Load returns the resolved Config. Errors are non-fatal where possible —
// a missing config file is fine, a malformed one returns an error.
func Load() (Config, error) {
	cfg := Defaults()

	// user file
	if path, err := UserConfigPath(); err == nil {
		if err := mergeFromFile(&cfg, path); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return cfg, fmt.Errorf("read %s: %w", path, err)
			}
		}
	}

	// project file (walk up from cwd to git root or filesystem root)
	if p, ok := findProjectFile(); ok {
		if err := mergeFromFile(&cfg, p); err != nil {
			return cfg, fmt.Errorf("read %s: %w", p, err)
		}
	}

	// env vars
	if v := os.Getenv("REFUSE_SERVER_URL"); v != "" {
		cfg.ServerURL = v
	}
	if v := os.Getenv("REFUSE_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("REFUSE_POLICY"); v != "" {
		cfg.Policy.SeverityThreshold = v
	}
	if v := os.Getenv("REFUSE_FAIL_CLOSED"); v != "" {
		cfg.Policy.FailClosed = boolish(v)
	}

	cfg.ServerURL = strings.TrimRight(cfg.ServerURL, "/")
	migrate(&cfg)
	return cfg, nil
}

// migrate brings older on-disk configs up to CurrentConfigVersion. Today it's
// a no-op for v0 (pre-versioning) → v1 because the schema didn't actually
// change. Add a branch here when bumping CurrentConfigVersion so existing
// users converge on new defaults rather than carrying a stale snapshot.
func migrate(cfg *Config) {
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	// Future: switch on cfg.Version and apply per-version bumps here.
}

func mergeFromFile(cfg *Config, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var loaded Config
	if err := yaml.Unmarshal(raw, &loaded); err != nil {
		return err
	}
	if loaded.ServerURL != "" {
		cfg.ServerURL = loaded.ServerURL
	}
	if loaded.APIKey != "" {
		cfg.APIKey = loaded.APIKey
	}
	if loaded.Policy.SeverityThreshold != "" {
		cfg.Policy.SeverityThreshold = loaded.Policy.SeverityThreshold
	}
	cfg.Policy.FailClosed = cfg.Policy.FailClosed || loaded.Policy.FailClosed
	cfg.Policy.AllowYanked = cfg.Policy.AllowYanked || loaded.Policy.AllowYanked
	cfg.Policy.AllowPrerelease = cfg.Policy.AllowPrerelease || loaded.Policy.AllowPrerelease
	if loaded.Policy.OverrideEnv != "" {
		cfg.Policy.OverrideEnv = loaded.Policy.OverrideEnv
	}
	return nil
}

func findProjectFile() (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		p := filepath.Join(cwd, ProjectFile)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", false
		}
		cwd = parent
	}
}

func boolish(s string) bool {
	switch strings.ToLower(s) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// Save writes cfg to the user config path (creating ~/.refuse if needed).
func Save(cfg Config) error {
	d, err := UserDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o755); err != nil {
		return err
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	p, _ := UserConfigPath()
	return os.WriteFile(p, out, 0o600)
}

// SeverityRank converts a severity_label to a comparable rank.
func SeverityRank(s string) int {
	switch strings.ToLower(s) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	}
	return 0
}
