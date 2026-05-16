// Package hook handles per-agent pre-tool-use hook installation.
//
// Each agent (Claude Code, Codex, Cursor, ...) has its own config shape;
// agents register an Adapter that knows where the config file is and how to
// idempotently insert/remove the refuse hook.
package hook

import (
	"fmt"
	"sort"
)

// Status reports whether the adapter found its config file + whether a refuse
// hook is currently installed.
type Status struct {
	Agent     string
	Found     bool
	ConfigPath string
	Installed bool
	Detail    string
}

// Adapter is implemented by each agent's hook installer.
type Adapter interface {
	Name() string
	// Install writes the refuse hook into the agent's config, idempotently.
	Install(refuseBin string) error
	// Remove strips out the refuse hook entry, if present.
	Remove() error
	// Status reports the current install state.
	Status() (Status, error)
}

var adapters = map[string]Adapter{}

// register adds an adapter to the lookup table. Called from init() in each
// per-agent file.
func register(a Adapter) { adapters[a.Name()] = a }

// Get returns the adapter for a given agent name.
func Get(name string) (Adapter, bool) {
	a, ok := adapters[name]
	return a, ok
}

// Names returns the registered agent names in sorted order.
func Names() []string {
	out := make([]string, 0, len(adapters))
	for n := range adapters {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// All returns adapters in name order — convenient for `hook list`.
func All() []Adapter {
	names := Names()
	out := make([]Adapter, 0, len(names))
	for _, n := range names {
		out = append(out, adapters[n])
	}
	return out
}

// ErrUnknownAgent — convenience error for the CLI layer.
func ErrUnknownAgent(name string) error {
	return fmt.Errorf("unknown agent %q (supported: %v)", name, Names())
}
