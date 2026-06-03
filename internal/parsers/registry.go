package parsers

// ForName returns the parser bound to a given package-manager binary name.
// Returns nil for unknown names (the shim layer treats nil → passthrough).
func ForName(name string) Parser {
	switch name {
	// npm ecosystem
	case "npm":
		return NPM()
	case "pnpm":
		return PNPM()
	case "yarn":
		return Yarn()
	case "bun":
		return Bun()
	case "npx":
		return NPX()
	// PyPI ecosystem
	case "pip", "pip3":
		return Pip()
	case "uv":
		return UV()
	case "poetry":
		return Poetry()
	case "pipenv":
		return Pipenv()
	case "pdm":
		return PDM()
	case "pipx":
		return Pipx()
	// crates.io
	case "cargo":
		return Cargo()
	// RubyGems
	case "gem":
		return Gem()
	case "bundle", "bundler":
		return Bundle()
	// Go modules
	case "go":
		return Go()
	// Packagist (PHP)
	case "composer":
		return Composer()
	// NuGet
	case "dotnet":
		return Dotnet()
	}
	return nil
}
