package parsers

import "strings"

// Composer handles PHP's `composer`.
//   composer require vendor/pkg               → direct
//   composer require vendor/pkg:1.0           → direct, pinned (colon form)
//   composer require vendor/pkg ^1.0          → direct, pinned (separate arg)
//   composer install                          → lockfile (composer.lock)
//   composer update / run-script / dump-autoload → passthrough

var composerFlagsTakingArg = map[string]bool{
	"--dev":                  false,
	"-d":                     true, // working-directory
	"--working-dir":          true,
	"--prefer-stable":        false,
	"--prefer-dist":          false,
	"--prefer-source":        false,
	"--ignore-platform-reqs": false,
}

// Composer returns a parser for `composer` argv.
func Composer() Parser { return composerParser{} }

type composerParser struct{}

func (composerParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	verb := args[0]
	switch verb {
	case "install":
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "composer.lock",
			Reason:       "composer install → lockfile",
		}
	case "require":
		positionals := splitPositionals(args[1:], composerFlagsTakingArg)
		pkgs := make([]PkgRef, 0, len(positionals))
		// Composer accepts either `vendor/pkg:1.0` or two args `vendor/pkg`
		// then `^1.0`. Walk positionals pairing each `vendor/pkg` with the
		// next token if it looks like a version (not another vendor/pkg).
		i := 0
		for i < len(positionals) {
			spec := positionals[i]
			if !strings.Contains(spec, "/") {
				// Not a `vendor/pkg`; skip.
				i++
				continue
			}
			name := spec
			version := ""
			if colon := strings.IndexByte(spec, ':'); colon != -1 {
				name = spec[:colon]
				version = spec[colon+1:]
			} else if i+1 < len(positionals) && !strings.Contains(positionals[i+1], "/") {
				// Next token looks like a version constraint, not a package.
				version = strings.TrimLeft(positionals[i+1], "^~=<>* ")
				i++
			}
			pkgs = append(pkgs, PkgRef{Ecosystem: "Packagist", Name: name, Version: version})
			i++
		}
		if len(pkgs) == 0 {
			return ParseResult{}
		}
		return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "composer require"}
	}
	return ParseResult{}
}
