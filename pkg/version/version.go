// Package version holds the single source of truth for the Xxlang version
// string.
//
// Before v0.9.7 the version was injected at build time via -ldflags
// (-X main.Version=...) and a separate BuildNumber was hard-coded in
// cmd/xxl/main.go. That split the version across three places (the workflow,
// main.go, and the tag) and caused the "vdev" health-endpoint symptom.
//
// Since v0.9.7 the version lives here. `go build ./cmd/xxl`, `go install`,
// IDE runs, and CI all read the same identifier, so there is nothing to
// inject and nothing to forget. The release workflow (release.yml) only
// verifies that Version matches the git tag.
//
// In v0.9.8 the BuildNumber was removed entirely — a single Version string
// is enough, and having two identifiers (Version + BuildNumber) just created
// maintenance burden without value.
//
// This is `var` (not `const`) solely so tests can override it; production
// code must never reassign it at runtime.
package version

// Version is the Xxlang version string (without the leading "v"), e.g.
// "0.9.8". It matches the git tag for releases.
var Version = "0.9.8"
