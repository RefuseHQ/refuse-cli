package parsers

// SplitNameVersionExt is the exported form of the shared splitNameVersion
// helper — handy for CLI subcommands that take a `name@version` argument.
func SplitNameVersionExt(spec string) (name, version string) {
	return splitNameVersion(spec)
}
