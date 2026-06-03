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
	// npm ecosystem
	"npm":  "npm",
	"pnpm": "npm",
	"yarn": "npm",
	"bun":  "npm",
	"npx":  "npm",
	// PyPI ecosystem
	"pip":    "PyPI",
	"pip3":   "PyPI",
	"uv":     "PyPI",
	"poetry": "PyPI",
	"pipenv": "PyPI",
	"pdm":    "PyPI",
	"pipx":   "PyPI",
	// crates.io
	"cargo": "crates.io",
	// RubyGems
	"gem":     "RubyGems",
	"bundle":  "RubyGems",
	"bundler": "RubyGems",
	// Go modules
	"go": "Go",
	// Packagist (PHP)
	"composer": "Packagist",
	// NuGet
	"dotnet": "NuGet",
}

// Run is the shim entry point — vet the call, then exec the real binary.
// On a block exits 2 with a stderr explanation; on allow it exec's the real
// manager so the parent shell sees its exit code directly.
func Run(name string, args []string) int {
	return runFromArgv(name, args)
}
