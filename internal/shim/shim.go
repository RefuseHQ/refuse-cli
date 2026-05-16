// Package shim handles the "multicall argv[0]" path.
//
// On install (`refuse install`) we drop symlinks named after package managers
// into ~/.refuse/bin and prepend that dir to PATH. When the user runs `npm
// install foo`, the shim hands the call to `refuse gate` to vet, then execs
// the *real* npm with the original argv on a pass.
package shim

// IsKnownManager returns true when argv[0] matches one of the package
// managers we wrap. Used by main.go to pick between shim and CLI mode.
func IsKnownManager(name string) bool {
	_, ok := KnownManagers[name]
	return ok
}

// KnownManagers is the set of binary names we install shims for. The value
// is the ecosystem identifier we send to the refuse server.
var KnownManagers = map[string]string{
	"npm":    "npm",
	"pnpm":   "npm",
	"yarn":   "npm",
	"pip":    "PyPI",
	"pip3":   "PyPI",
	"cargo":  "crates.io",
	"gem":    "RubyGems",
	"bun":    "npm",
	"go":     "Go",
	"dotnet": "NuGet",
}

// Run is the shim entry point — vet the call, then exec the real binary.
// For v1 it just execs through; the gate is wired up in the gate task.
func Run(name string, args []string) int {
	// TODO: parse args, call refuse gate, block on vulnerable.
	// For v1 scaffold: execute the real binary directly so nothing breaks
	// while the gate is being implemented.
	return execReal(name, args)
}
