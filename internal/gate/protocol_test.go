package gate

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitShellCommand(t *testing.T) {
	cases := []struct {
		in   string
		want [][]string
	}{
		{"", nil},
		{"npm install foo", [][]string{{"npm", "install", "foo"}}},
		{"npm install foo && npm test", [][]string{{"npm", "install", "foo"}, {"npm", "test"}}},
		{"pip install requests; pip freeze", [][]string{{"pip", "install", "requests"}, {"pip", "freeze"}}},
		{`echo "a && b" && npm install foo`, [][]string{{"echo", "a && b"}, {"npm", "install", "foo"}}},
		{"cd /tmp && pnpm add lodash", [][]string{{"cd", "/tmp"}, {"pnpm", "add", "lodash"}}},
		{"npm install foo | tee log", [][]string{{"npm", "install", "foo"}, {"tee", "log"}}},
	}
	for _, c := range cases {
		got, err := SplitShellCommand(c.in)
		if err != nil {
			t.Fatalf("split %q: %v", c.in, err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("split %q:\n  got:  %v\n  want: %v", c.in, got, c.want)
		}
	}
}

func TestParseHookInputBash(t *testing.T) {
	in := strings.NewReader(`{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"npm install foo"}}`)
	cmd, err := ParseHookInput(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cmd != "npm install foo" {
		t.Errorf("got %q, want %q", cmd, "npm install foo")
	}
}

func TestParseHookInputNonBash(t *testing.T) {
	in := strings.NewReader(`{"hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{"file":"foo"}}`)
	cmd, err := ParseHookInput(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cmd != "" {
		t.Errorf("got %q, want empty (non-Bash tool)", cmd)
	}
}
