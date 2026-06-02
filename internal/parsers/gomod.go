package parsers

import "strings"

// Go handles `go get` and `go install` for module-version specs.
//   go get github.com/foo/bar@v1.2.3      → direct, pinned
//   go install github.com/foo/cmd@latest  → direct (latest → unpinned)
//   go get golang.org/x/text              → direct, unpinned
//   go get ./...   /  go build / test ... → passthrough
//
// Go's version syntax is always module@version. `@latest`, `@upgrade`,
// `@patch`, `@none` are pseudo-selectors — we treat them as unpinned so the
// gate resolves the concrete latest.

// Go returns a parser for `go` argv.
func Go() Parser { return goParser{} }

type goParser struct{}

func (goParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	verb := args[0]
	if verb != "get" && verb != "install" {
		return ParseResult{}
	}

	positionals := splitPositionals(args[1:], nil)
	pkgs := make([]PkgRef, 0, len(positionals))
	for _, spec := range positionals {
		// Local paths and package-pattern meta-args aren't registry installs.
		if spec == "all" || strings.HasPrefix(spec, ".") || strings.HasPrefix(spec, "/") ||
			strings.Contains(spec, "...") {
			continue
		}
		name := spec
		version := ""
		// Module paths never contain '@'; the last '@' separates the version.
		if at := strings.LastIndexByte(spec, '@'); at != -1 {
			name = spec[:at]
			version = spec[at+1:]
			switch version {
			case "latest", "upgrade", "patch", "none":
				version = ""
			}
		}
		if name == "" {
			continue
		}
		pkgs = append(pkgs, PkgRef{Ecosystem: "Go", Name: name, Version: version})
	}
	if len(pkgs) == 0 {
		return ParseResult{}
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "go " + verb}
}
