package parsers

// Pipx handles `pipx install` and `pipx run` (which fetches transiently).
//   pipx install <pkg>          → direct
//   pipx install <pkg>==1.0     → direct, pinned
//   pipx run <pkg>              → direct (fetches the package to run it)
//   pipx run <pkg>==1.0         → direct, pinned
//   pipx upgrade / list / uninstall → passthrough

var pipxFlagsTakingArg = map[string]bool{
	"--python":    true,
	"--index-url": true,
	"--pip-args":  true,
	"--spec":      true,
	"--suffix":    true,
}

// Pipx returns a parser for `pipx` argv.
func Pipx() Parser { return pipxParser{} }

type pipxParser struct{}

func (pipxParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	verb := args[0]
	if verb != "install" && verb != "run" {
		return ParseResult{}
	}
	positionals := splitPositionals(args[1:], pipxFlagsTakingArg)
	pkgs := pyPositionalsToPkgs(positionals)
	if len(pkgs) == 0 {
		return ParseResult{}
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "pipx " + verb}
}
