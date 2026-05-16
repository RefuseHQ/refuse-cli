package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() { register(&claudeCode{}) }

// Sentinel substring we look for to identify "this is our hook". Survives
// users renaming the binary location as long as the args stay the same.
const claudeRefuseTag = "refuse gate"

type claudeCode struct{}

func (claudeCode) Name() string { return "claude-code" }

func (c claudeCode) configPath() (string, error) {
	if p := os.Getenv("CLAUDE_CODE_SETTINGS"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// Claude Code's settings.json shape (we only touch the hooks key):
//
//   {
//     "hooks": {
//       "PreToolUse": [
//         { "matcher": "Bash", "hooks": [
//           { "type": "command", "command": "...", "timeout": 5000 }
//         ] }
//       ]
//     }
//   }
//
// We preserve everything else in the file. JSON encoding keeps a stable key
// order via the struct field tags.

type claudeHookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

type claudeHookGroup struct {
	Matcher string          `json:"matcher,omitempty"`
	Hooks   []claudeHookCmd `json:"hooks"`
}

func (c claudeCode) Install(refuseBin string) error {
	p, err := c.configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}

	settings := map[string]any{}
	if raw, err := os.ReadFile(p); err == nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, &settings); err != nil {
			return fmt.Errorf("parse %s: %w", p, err)
		}
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}

	pre, _ := hooks["PreToolUse"].([]any)
	// Strip any existing refuse entry so we don't duplicate.
	pre = withoutRefuseEntries(pre)

	desired := claudeHookGroup{
		Matcher: "Bash",
		Hooks: []claudeHookCmd{{
			Type:    "command",
			Command: fmt.Sprintf("%s gate --agent=claude-code", refuseBin),
			Timeout: 5000,
		}},
	}
	desiredAny, err := toGenericMap(desired)
	if err != nil {
		return err
	}
	pre = append(pre, desiredAny)

	hooks["PreToolUse"] = pre
	settings["hooks"] = hooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(p, append(out, '\n'), 0o600)
}

func (c claudeCode) Remove() error {
	p, err := c.configPath()
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	settings := map[string]any{}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return fmt.Errorf("parse %s: %w", p, err)
	}
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return nil
	}
	pre, _ := hooks["PreToolUse"].([]any)
	stripped := withoutRefuseEntries(pre)
	if len(stripped) == 0 {
		delete(hooks, "PreToolUse")
	} else {
		hooks["PreToolUse"] = stripped
	}
	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(p, append(out, '\n'), 0o600)
}

func (c claudeCode) Status() (Status, error) {
	st := Status{Agent: c.Name()}
	p, err := c.configPath()
	if err != nil {
		return st, err
	}
	st.ConfigPath = p
	raw, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			st.Detail = "settings.json not found"
			return st, nil
		}
		return st, err
	}
	st.Found = true
	var settings map[string]any
	if err := json.Unmarshal(raw, &settings); err != nil {
		st.Detail = "settings.json malformed: " + err.Error()
		return st, nil
	}
	hooks, _ := settings["hooks"].(map[string]any)
	pre, _ := hooks["PreToolUse"].([]any)
	for _, g := range pre {
		if entryHasRefuse(g) {
			st.Installed = true
			break
		}
	}
	if !st.Installed {
		st.Detail = "no refuse hook"
	}
	return st, nil
}

func withoutRefuseEntries(in []any) []any {
	out := make([]any, 0, len(in))
	for _, entry := range in {
		if entryHasRefuse(entry) {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func entryHasRefuse(entry any) bool {
	m, ok := entry.(map[string]any)
	if !ok {
		return false
	}
	hs, _ := m["hooks"].([]any)
	for _, h := range hs {
		hm, ok := h.(map[string]any)
		if !ok {
			continue
		}
		if cmd, _ := hm["command"].(string); strings.Contains(cmd, claudeRefuseTag) {
			return true
		}
	}
	return false
}

// toGenericMap round-trips a struct through JSON so it merges back into the
// untyped settings map cleanly (preserves field order, drops zero values
// where the tag says omitempty).
func toGenericMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// atomicWrite writes via a temp file + rename so a half-write can't leave
// the user's settings.json corrupt.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".refuse-claude-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
