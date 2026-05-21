package parsers

import "strings"

// Cargo handles `cargo add` and `cargo install`.
//   cargo add serde                    → direct
//   cargo add serde@1.0.150            → direct, pinned (@ form, cargo ≥ 1.62)
//   cargo add serde --vers 1.0         → direct, pinned (--vers form)
//   cargo add serde --features derive  → direct
//   cargo install ripgrep              → direct (binary crate)
//   cargo build / test / run / ...     → passthrough
//
// `cargo install --path .` / `--git <url>` install from local source or a
// repo; the gate has nothing useful to say, so those positionals are skipped.

var cargoFlagsTakingArg = map[string]bool{
	"--vers":     true,
	"--version":  true,
	"--features": true,
	"-F":         true,
	"--path":     true,
	"--git":      true,
	"--branch":   true,
	"--tag":      true,
	"--rev":      true,
	"--registry": true,
	"--index":    true,
	"--root":     true,
	"--profile":  true,
	"--target":   true,
	"--bin":      true,
	"--example":  true,
	"-j":         true,
	"--jobs":     true,
}

// Cargo returns a parser for `cargo` argv.
func Cargo() Parser { return cargoParser{} }

type cargoParser struct{}

func (cargoParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	verb := args[0]
	if verb != "add" && verb != "install" {
		return ParseResult{}
	}
	rest := args[1:]

	// A --vers/--version flag (if present) applies to the crate(s) named in
	// the same command. Capture it so `cargo add serde --vers 1.0` pins.
	var flagVersion string
	for i, a := range rest {
		switch {
		case (a == "--vers" || a == "--version") && i+1 < len(rest):
			flagVersion = rest[i+1]
		case strings.HasPrefix(a, "--vers="):
			flagVersion = strings.TrimPrefix(a, "--vers=")
		case strings.HasPrefix(a, "--version="):
			flagVersion = strings.TrimPrefix(a, "--version=")
		}
	}

	positionals := splitPositionals(rest, cargoFlagsTakingArg)
	pkgs := make([]PkgRef, 0, len(positionals))
	for _, spec := range positionals {
		// Local path installs aren't registry crates.
		if strings.Contains(spec, "/") || strings.HasPrefix(spec, ".") {
			continue
		}
		name, version := splitNameVersion(spec)
		if version == "" {
			version = flagVersion
		}
		if name == "" {
			continue
		}
		pkgs = append(pkgs, PkgRef{Ecosystem: "crates.io", Name: name, Version: version})
	}
	if len(pkgs) == 0 {
		return ParseResult{}
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "cargo " + verb}
}
