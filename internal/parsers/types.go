// Package parsers turns a package manager's argv into a `ParseResult` that
// the gate engine can decide on. Each manager has its own file and is
// covered by table-driven tests.
package parsers

// Mode describes what kind of install operation the argv represents.
type Mode int

const (
	// ModePassthrough — not an install verb (test/build/run/etc.).
	ModePassthrough Mode = iota
	// ModeDirect — argv names specific packages to install.
	ModeDirect
	// ModeLockfile — the call resolves from a lockfile in cwd (npm ci,
	// pip install -r, yarn install with no args, ...). The shim hands the
	// lockfile path to `refuse check-lockfile` if discoverable.
	ModeLockfile
)

// PkgRef is a single package extracted from argv. Version may be empty when
// the user invoked the install without pinning (e.g. `npm install lodash`),
// in which case the gate resolves "latest" against the registry.
type PkgRef struct {
	Ecosystem string
	Name      string
	Version   string // empty when unpinned
}

// ParseResult is the parser's output.
type ParseResult struct {
	// IsInstall is true when the argv represents an install-flavored verb.
	IsInstall bool
	// Mode is the kind of install (Direct or Lockfile). Meaningful only when
	// IsInstall is true.
	Mode Mode
	// Packages — populated when Mode is ModeDirect.
	Packages []PkgRef
	// LockfileHint — relative path to the lockfile when Mode is ModeLockfile;
	// the shim resolves it against cwd before calling check-lockfile.
	LockfileHint string
	// Reason is human-readable, used in --verbose output and block messages.
	Reason string
}

// Parser is implemented by each package-manager file. The argv slice is
// everything *after* argv[0] (the manager name itself).
type Parser interface {
	Parse(args []string) ParseResult
}
