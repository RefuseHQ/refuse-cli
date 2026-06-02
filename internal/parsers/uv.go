package parsers

import "strings"

// UV handles Astral's `uv` — a pip-compatible front-end plus a project tool.
//   uv pip install <pkg>             → direct (pip-compat)
//   uv pip install <pkg>==1.0        → direct, pinned
//   uv pip install -r reqs.txt       → lockfile mode
//   uv add <pkg>                     → direct (project add)
//   uv add <pkg>==1.0                → direct, pinned
//   uv tool install <pkg>            → direct (install a CLI tool)
//   uv sync / run / build / lock     → passthrough
//
// All install paths target PyPI.

var uvPipFlagsTakingArg = map[string]bool{
	"--index-url":       true,
	"--extra-index-url": true,
	"--find-links":      true,
	"--python":          true,
	"-p":                true,
	"--target":          true,
	"-t":                true,
	"--prefix":          true,
}

var uvAddFlagsTakingArg = map[string]bool{
	"--group":   true,
	"--package": true,
	"--python":  true,
	"-p":        true,
	"--extra":   true,
}

// UV returns a parser for `uv` argv.
func UV() Parser { return uvParser{} }

type uvParser struct{}

func (uvParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	switch args[0] {
	case "pip":
		return uvPipParse(args[1:])
	case "add":
		return uvDirectParse(args[1:], uvAddFlagsTakingArg, "uv add")
	case "tool":
		// `uv tool install <pkg>` is the only tool subcommand we gate.
		if len(args) > 1 && args[1] == "install" {
			return uvDirectParse(args[2:], uvPipFlagsTakingArg, "uv tool install")
		}
		return ParseResult{}
	}
	return ParseResult{}
}

func uvPipParse(args []string) ParseResult {
	if len(args) == 0 || args[0] != "install" {
		return ParseResult{}
	}
	rest := args[1:]
	positionals := make([]string, 0, len(rest))
	var lockfile string
	skipNext := false
	for i, a := range rest {
		if skipNext {
			skipNext = false
			continue
		}
		switch {
		case a == "-r" || a == "--requirement":
			if i+1 < len(rest) {
				lockfile = rest[i+1]
				skipNext = true
			}
		case strings.HasPrefix(a, "--requirement="):
			lockfile = strings.TrimPrefix(a, "--requirement=")
		case uvPipFlagsTakingArg[a]:
			skipNext = true
		case strings.HasPrefix(a, "-"):
			// boolean flag — skip
		default:
			positionals = append(positionals, a)
		}
	}
	if lockfile != "" {
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: lockfile,
			Reason:       "uv pip install -r " + lockfile,
		}
	}
	pkgs := pyPositionalsToPkgs(positionals)
	if len(pkgs) == 0 {
		return ParseResult{}
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "uv pip install"}
}

func uvDirectParse(rest []string, takeArg map[string]bool, reason string) ParseResult {
	positionals := splitPositionals(rest, takeArg)
	pkgs := pyPositionalsToPkgs(positionals)
	if len(pkgs) == 0 {
		return ParseResult{}
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: reason}
}

// pyPositionalsToPkgs converts pip-style positional install specs to PkgRefs,
// skipping path / URL / wheel specs.
func pyPositionalsToPkgs(positionals []string) []PkgRef {
	pkgs := make([]PkgRef, 0, len(positionals))
	for _, spec := range positionals {
		if strings.Contains(spec, "/") || strings.HasPrefix(spec, ".") ||
			strings.HasPrefix(spec, "git+") || strings.HasSuffix(spec, ".whl") ||
			strings.HasSuffix(spec, ".tar.gz") {
			continue
		}
		name, version := splitPipSpec(spec)
		if name == "" {
			continue
		}
		pkgs = append(pkgs, PkgRef{Ecosystem: "PyPI", Name: name, Version: version})
	}
	return pkgs
}
