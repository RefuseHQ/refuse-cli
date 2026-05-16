package parsers

import "strings"

// Pip handles `pip` / `pip3`. Verbs that matter:
//   pip install foo               → direct
//   pip install foo==1.2.3        → direct, pinned
//   pip install -r requirements.txt  → lockfile mode
//   pip install --upgrade pip     → direct (pip upgrades itself)
//
// `pip download`, `pip wheel`, `pip uninstall`, `pip freeze`, `pip list`,
// `pip show`, `pip search` all pass through. We only gate `install`.

// Pip returns the parser for pip / pip3 argv.
func Pip() Parser { return pipParser{} }

type pipParser struct{}

func (pipParser) Parse(args []string) ParseResult {
	if len(args) == 0 || args[0] != "install" {
		return ParseResult{}
	}
	// Walk argv ourselves so we can pick up the `-r <file>` lockfile arg.
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
		case a == "-c" || a == "--constraint" || a == "--index-url" ||
			a == "--extra-index-url" || a == "--target" || a == "-t" ||
			a == "--prefix" || a == "--find-links" || a == "-f":
			skipNext = true
		case strings.HasPrefix(a, "-"):
			// boolean flag — ignore
		default:
			positionals = append(positionals, a)
		}
	}

	if lockfile != "" {
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: lockfile,
			Reason:       "pip install -r " + lockfile,
		}
	}
	if len(positionals) == 0 {
		// `pip install` alone is a usage error from pip's perspective; no
		// install happens, so we pass through.
		return ParseResult{}
	}
	pkgs := make([]PkgRef, 0, len(positionals))
	for _, spec := range positionals {
		// Skip path-style installs and editable installs — those install
		// from local source and the gate has nothing useful to say.
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
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "pip install"}
}
