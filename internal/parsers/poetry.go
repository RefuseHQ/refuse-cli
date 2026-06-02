package parsers

import "strings"

// Poetry handles `poetry add` and `poetry install`.
//   poetry add <pkg>             → direct
//   poetry add <pkg>@1.2.3       → direct, pinned (@ form)
//   poetry add <pkg>==1.2.3      → direct, pinned (pip form)
//   poetry add <pkg> --extras foo → direct
//   poetry install               → lockfile mode (poetry.lock)
//   poetry update / run / lock   → passthrough

var poetryFlagsTakingArg = map[string]bool{
	"--group":    true,
	"-G":         true,
	"--source":   true,
	"--extras":   true,
	"-E":         true,
	"--python":   true,
	"--platform": true,
	"--editable": false,
	"--with":     true,
	"--without":  true,
}

// Poetry returns a parser for `poetry` argv.
func Poetry() Parser { return poetryParser{} }

type poetryParser struct{}

func (poetryParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	verb := args[0]
	switch verb {
	case "add":
		positionals := splitPositionals(args[1:], poetryFlagsTakingArg)
		pkgs := make([]PkgRef, 0, len(positionals))
		for _, spec := range positionals {
			// Skip path / URL specs.
			if strings.Contains(spec, "/") || strings.HasPrefix(spec, ".") ||
				strings.HasPrefix(spec, "git+") {
				continue
			}
			// poetry accepts both @ and pip-style operators. Try @ first.
			name, version := splitNameVersion(spec)
			if version == "" {
				// Fallback to pip-style ==, ~=, >=, etc.
				name, version = splitPipSpec(spec)
			}
			if name == "" {
				continue
			}
			pkgs = append(pkgs, PkgRef{Ecosystem: "PyPI", Name: name, Version: version})
		}
		if len(pkgs) == 0 {
			return ParseResult{}
		}
		return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "poetry add"}
	case "install":
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "poetry.lock",
			Reason:       "poetry install → lockfile",
		}
	}
	return ParseResult{}
}
