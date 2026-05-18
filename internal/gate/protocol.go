package gate

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/shlex"
)

// HookInput is the JSON the gate reads on stdin when invoked by an agent
// hook. The shape follows Claude Code's PreToolUse payload, which the other
// agents we eventually support can normalise onto.
//
// Example:
//
//	{
//	  "hook_event_name": "PreToolUse",
//	  "tool_name": "Bash",
//	  "tool_input": { "command": "npm install foo", "description": "..." }
//	}
type HookInput struct {
	HookEventName string         `json:"hook_event_name"`
	ToolName      string         `json:"tool_name"`
	ToolInput     map[string]any `json:"tool_input"`
}

// ParseHookInput reads stdin (or the supplied reader in tests), parses the
// JSON, and extracts the shell command. Returns empty command for non-Bash
// tool calls — the caller should exit 0 silently in that case.
func ParseHookInput(r io.Reader) (string, error) {
	var in HookInput
	dec := json.NewDecoder(r)
	if err := dec.Decode(&in); err != nil {
		return "", fmt.Errorf("hook: parse stdin: %w", err)
	}
	if in.ToolName != "Bash" {
		// Other tool types — nothing for us to vet.
		return "", nil
	}
	cmd, _ := in.ToolInput["command"].(string)
	return cmd, nil
}

// SplitShellCommand splits a bash one-liner into commands joined by `&&`,
// `;`, or `|`. Each subcommand is then tokenized with shlex.
//
// We deliberately ignore subshells, redirections, and quoting nuances that
// don't matter for our use case — we only want to find the leading binary
// name and its argv.
func SplitShellCommand(cmd string) ([][]string, error) {
	if cmd == "" {
		return nil, nil
	}
	// shlex doesn't know about &&/||/;/|, so we do a simple split first.
	parts := splitOnConnectors(cmd)
	out := make([][]string, 0, len(parts))
	for _, p := range parts {
		toks, err := shlex.Split(p)
		if err != nil || len(toks) == 0 {
			continue
		}
		out = append(out, toks)
	}
	return out, nil
}

// splitOnConnectors splits a bash command line on the top-level connectors
// `&&`, `||`, `;`, `|`. Quoting is respected at the byte level — we don't
// split inside a single- or double-quoted string. Backslash escapes outside
// quotes are honoured.
func splitOnConnectors(s string) []string {
	var (
		out      []string
		buf      []byte
		inS, inD bool
		i        int
	)
	flush := func() {
		t := trimSpace(string(buf))
		if t != "" {
			out = append(out, t)
		}
		buf = buf[:0]
	}
	for i < len(s) {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s) && !inS:
			buf = append(buf, c, s[i+1])
			i += 2
			continue
		case c == '\'' && !inD:
			inS = !inS
		case c == '"' && !inS:
			inD = !inD
		}
		if !inS && !inD {
			if c == ';' || c == '|' || c == '&' {
				// Look at the connector kind.
				if c == ';' {
					flush()
					i++
					continue
				}
				if i+1 < len(s) && s[i+1] == c {
					// && or ||
					flush()
					i += 2
					continue
				}
				if c == '|' {
					// Single pipe — gate each side too.
					flush()
					i++
					continue
				}
				// Single & — background, treat as separator.
				flush()
				i++
				continue
			}
		}
		buf = append(buf, c)
		i++
	}
	flush()
	return out
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && isSpace(s[start]) {
		start++
	}
	for end > start && isSpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func isSpace(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }
