// Package version exposes build metadata injected via ldflags at release time.
package version

// These are overridden at build time via -ldflags="-X ..." in the goreleaser
// config. Defaults are useful for `go run` / dev builds.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
