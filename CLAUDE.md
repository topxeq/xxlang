# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
Xxlang (Chinese: 现象语言) is a line-by-line interpreted scripting language implemented in Go. This is a greenfield project.

## Key Requirements
- **Go standard library preferred** - minimize external dependencies
- **English for all documentation and code comments**
- **Repository**: `github/topxeq/xxlang`
- **Naming**: camelCase for all built-in functions and variables (e.g., `typeOf`, `runCode`)
- **File extension**: `.xxl`

## Core Features to Implement
- Modern programming language features with lightweight OOP (avoid complex OOP for performance)
- All data types are objects with a common base class supporting universal methods (`typeOf`, `toStr`)
- Type-specific methods (e.g., `parseStr` for integers)
- Module loading during execution
- Plugin system for extending functionality (loaded at runtime, syntax similar to modules)
- Embedding support: callable as a Go library with parameter passing and return values
- `runCode` function: execute Xxlang code from within Xxlang with parameters and returns
- Compile to executable: package VM + code files for target platform (not native compilation)
- Performance benchmark: recursive Fibonacci(35) vs Go, C, Java, Python

## Planned Structure
```
cmd/xxlang/     - Main executable entry point
pkg/interpreter/ - Interpreter package (importable for embedding)
```
