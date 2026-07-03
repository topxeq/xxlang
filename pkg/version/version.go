// Package version holds the single source of truth for the Xxlang version
// string and build number.
//
// Before v0.9.7 the version was injected at build time via -ldflags
// (-X main.Version=...) and the build number was hard-coded in
// cmd/xxl/main.go. That split the version across three places (the workflow,
// main.go, and the tag) and caused the "vdev" health-endpoint symptom: the
// release workflow forgot to propagate the build number, and local/CI builds
// without ldflags saw Version="dev".
//
// Now both values live here. `go build ./cmd/xxl`, `go install`, IDE runs,
// and CI all read the same identifiers, so there is nothing to inject and
// nothing to forget. The release workflow (release.yml) only verifies that
// Version matches the git tag.
//
// These are `var` (not `const`) solely so tests can override them; production
// code must never reassign them at runtime.
package version

// Version is the Xxlang version string (without the leading "v"), e.g.
// "0.9.7". It matches the git tag for releases.
var Version = "0.9.7"

// BuildNumber is a hard-coded build identifier with format YYYYMMDDNN
// (year month day + daily sequence number, e.g. 2026070302). It should be
// bumped together with Version for each release.
var BuildNumber = "2026070302"
