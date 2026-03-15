# Awesome-Go Submission Guide for Xxlang

## Current Status

| Requirement | Status | Notes |
|-------------|--------|-------|
| 5+ months history | ❌ **BLOCKING** | First commit: 2026-03-07, need to wait until ~2026-08-07 |
| Open source license | ✅ Pass | MIT License |
| go.mod present | ✅ Pass | `go 1.25.5` |
| SemVer release | ✅ Pass | v0.3.2 - v0.3.9 |
| English documentation | ✅ Pass | README.md + code comments |
| Test coverage ≥80% | ⚠️ 38.1% | Core packages: lexer 97.2%, interpreter 91.7% |
| Go Report Card A-/A/A+ | ⏳ Pending | Check: https://goreportcard.com/report/github.com/topxeq/xxlang |
| Active maintenance | ✅ Pass | Recent commits, responding to issues |

## Action Items Before Submission

### 1. Wait for 5-Month Requirement
The repository needs to be at least 5 months old. Submit after **2026-08-07**.

### 2. Improve Test Coverage (Recommended)
Current coverage is 38.1%, below the 80% requirement. Focus on:
- `pkg/objects` (44.0%)
- `pkg/vm` (57.9%)
- `pkg/compiler` (62.6%)
- `pkg/parser` (71.4%)

### 3. Add Codecov Integration (Recommended)
1. Connect repo to https://codecov.io
2. Add coverage badge to README
3. Ensure CI runs tests with coverage

## Submission Materials

### Category
**Languages** - Xxlang is a scripting language implemented in Go.

### Entry to Add
```markdown
- [Xxlang](https://github.com/topxeq/xxlang) - Bytecode VM-based scripting language with closures, classes, and module system.
```

### PR Body Template
```markdown
Forge link: https://github.com/topxeq/xxlang
pkg.go.dev: https://pkg.go.dev/github.com/topxeq/xxlang
goreportcard.com: https://goreportcard.com/report/github.com/topxeq/xxlang
Coverage: https://app.codecov.io/gh/topxeq/xxlang
```

## Submission Checklist

Before opening the PR, verify:

- [ ] Repository is 5+ months old
- [ ] Go Report Card grade is A-, A, or A+
- [ ] Test coverage ≥80%
- [ ] pkg.go.dev documentation is accessible
- [ ] Entry is in correct category (Languages)
- [ ] Entry is in alphabetical order within category
- [ ] Description is concise, non-promotional, ends with period
- [ ] Only README.md is modified in the PR

## Post-Acceptance Badge

After acceptance, add to README:

```markdown
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
```

## Reference Links

- [Awesome-Go README](https://github.com/avelino/awesome-go)
- [Awesome-Go Contributing Guide](https://github.com/avelino/awesome-go/blob/main/CONTRIBUTING.md)
- [Go Report Card](https://goreportcard.com/report/github.com/topxeq/xxlang)
- [pkg.go.dev](https://pkg.go.dev/github.com/topxeq/xxlang)
