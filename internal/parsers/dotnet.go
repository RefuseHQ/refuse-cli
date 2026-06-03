package parsers

import "strings"

// Dotnet handles `dotnet add package` and `dotnet restore`.
//   dotnet add package <pkg>                → direct
//   dotnet add package <pkg> --version 1.0  → direct, pinned
//   dotnet add package <pkg> -v 1.0         → direct, pinned
//   dotnet restore                          → lockfile (packages.lock.json)
//   dotnet build / test / run / pack / …    → passthrough
//
// Ecosystem is NuGet. `dotnet add reference <path>` is project-local and skipped.

var dotnetFlagsTakingArg = map[string]bool{
	"--version":           true,
	"-v":                  true,
	"--framework":         true,
	"-f":                  true,
	"--source":            true,
	"-s":                  true,
	"--package-directory": true,
	"--project":           true,
}

// Dotnet returns a parser for `dotnet` argv.
func Dotnet() Parser { return dotnetParser{} }

type dotnetParser struct{}

func (dotnetParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	switch args[0] {
	case "restore":
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "packages.lock.json",
			Reason:       "dotnet restore → lockfile",
		}
	case "add":
		// Expect `dotnet add package <pkg> [--version <v>]`.
		// Skip any preceding <project> arg before `package`.
		rest := args[1:]
		idx := -1
		for i, a := range rest {
			if a == "package" {
				idx = i
				break
			}
		}
		if idx == -1 {
			return ParseResult{}
		}
		subRest := rest[idx+1:]
		var flagVersion string
		for i, a := range subRest {
			switch {
			case (a == "--version" || a == "-v") && i+1 < len(subRest):
				flagVersion = subRest[i+1]
			case strings.HasPrefix(a, "--version="):
				flagVersion = strings.TrimPrefix(a, "--version=")
			}
		}
		positionals := splitPositionals(subRest, dotnetFlagsTakingArg)
		pkgs := make([]PkgRef, 0, len(positionals))
		for _, spec := range positionals {
			// Skip local project paths.
			if strings.Contains(spec, "/") || strings.HasSuffix(spec, ".csproj") {
				continue
			}
			pkgs = append(pkgs, PkgRef{Ecosystem: "NuGet", Name: spec, Version: flagVersion})
		}
		if len(pkgs) == 0 {
			return ParseResult{}
		}
		return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "dotnet add package"}
	}
	return ParseResult{}
}
