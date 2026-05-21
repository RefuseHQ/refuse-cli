package parsers

import "strings"

// Gem handles `gem install` (and its `gem i` alias).
//   gem install rails                  → direct
//   gem install rails -v 6.1.0         → direct, pinned (-v form)
//   gem install rails --version 6.1.0  → direct, pinned
//   gem install rails:6.1.0            → direct, pinned (colon form)
//   gem update / list / build / ...    → passthrough
//
// Local `.gem` file installs are skipped — the gate vets registry gems.

var gemFlagsTakingArg = map[string]bool{
	"-v":            true,
	"--version":     true,
	"-i":            true,
	"--install-dir": true,
	"-n":            true,
	"--bindir":      true,
	"--platform":    true,
	"-g":            true,
	"--file":        true,
	"-s":            true,
	"--source":      true,
}

// Gem returns a parser for `gem` argv.
func Gem() Parser { return gemParser{} }

type gemParser struct{}

func (gemParser) Parse(args []string) ParseResult {
	if len(args) == 0 || (args[0] != "install" && args[0] != "i") {
		return ParseResult{}
	}
	rest := args[1:]

	// -v/--version applies to the gem(s) named in the same command. RubyGems
	// version specs can be operators (`-v '>= 6.0'`); we keep just the digits
	// and let the server resolve. Trim common operator/space noise.
	var flagVersion string
	for i, a := range rest {
		switch {
		case (a == "-v" || a == "--version") && i+1 < len(rest):
			flagVersion = strings.TrimLeft(rest[i+1], "=<>~! ")
		case strings.HasPrefix(a, "--version="):
			flagVersion = strings.TrimLeft(strings.TrimPrefix(a, "--version="), "=<>~! ")
		}
	}

	positionals := splitPositionals(rest, gemFlagsTakingArg)
	pkgs := make([]PkgRef, 0, len(positionals))
	for _, spec := range positionals {
		if strings.HasSuffix(spec, ".gem") || strings.Contains(spec, "/") {
			continue
		}
		name := spec
		version := flagVersion
		// Colon form: `gem install rails:6.1.0`.
		if i := strings.IndexByte(spec, ':'); i != -1 {
			name = spec[:i]
			version = spec[i+1:]
		}
		if name == "" {
			continue
		}
		pkgs = append(pkgs, PkgRef{Ecosystem: "RubyGems", Name: name, Version: version})
	}
	if len(pkgs) == 0 {
		return ParseResult{}
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "gem install"}
}
