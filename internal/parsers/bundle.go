package parsers

import "strings"

// Bundle handles `bundle install` / `bundle add` (bundler).
//   bundle install                    → lockfile (Gemfile.lock)
//   bundle add <gem>                  → direct
//   bundle add <gem> --version 1.0    → direct, pinned
//   bundle add <gem> -v 1.0           → direct, pinned
//   bundle update / exec / config     → passthrough

var bundleFlagsTakingArg = map[string]bool{
	"-v":        true,
	"--version": true,
	"--source":  true,
	"-s":        true,
	"--group":   true,
	"-g":        true,
	"--git":     true,
	"--github":  true,
	"--branch":  true,
	"--ref":     true,
	"--tag":     true,
}

// Bundle returns a parser for `bundle` argv.
func Bundle() Parser { return bundleParser{} }

type bundleParser struct{}

func (bundleParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		// `bundle` (no args) is roughly `bundle install`.
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "Gemfile.lock",
			Reason:       "bundle (no args) → install Gemfile.lock",
		}
	}
	switch args[0] {
	case "install":
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "Gemfile.lock",
			Reason:       "bundle install → lockfile",
		}
	case "add":
		rest := args[1:]
		// Capture -v / --version applied to the gem(s) listed.
		var flagVersion string
		for i, a := range rest {
			switch {
			case (a == "-v" || a == "--version") && i+1 < len(rest):
				flagVersion = strings.TrimLeft(rest[i+1], "=<>~! ")
			case strings.HasPrefix(a, "--version="):
				flagVersion = strings.TrimLeft(strings.TrimPrefix(a, "--version="), "=<>~! ")
			}
		}
		positionals := splitPositionals(rest, bundleFlagsTakingArg)
		pkgs := make([]PkgRef, 0, len(positionals))
		for _, spec := range positionals {
			if strings.Contains(spec, "/") {
				continue
			}
			pkgs = append(pkgs, PkgRef{Ecosystem: "RubyGems", Name: spec, Version: flagVersion})
		}
		if len(pkgs) == 0 {
			return ParseResult{}
		}
		return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "bundle add"}
	}
	return ParseResult{}
}
