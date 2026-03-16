# Command Line Argument Processing Examples

This directory contains examples demonstrating how to handle command line arguments in Xxlang.

## Files

### `args_basic.xxl`

A comprehensive example showing 6 different patterns for handling command line arguments:

1. **Simple positional arguments** - Basic access to raw arguments
2. **First argument with default** - Get argument with fallback value
3. **Flag detection** - Check if boolean flags are present
4. **Get option value** - Get value for options like `--name value`
5. **Environment variables as fallback** - CLI args with env var fallback
6. **Collect positional arguments** - Filter out options, keep positional args

```bash
# Run the example
xxlang examples/scripts/cmdline/args_basic.xxl arg1 arg2 --name John --count 5 --verbose
```

### `cli_demo.xxl`

A complete CLI application demonstrating:

- Long and short option parsing (`-h`/`--help`, `-n`/`--name`)
- Boolean flags (`--verbose`)
- Options with values (`--name John`, `--count 5`)
- Positional argument collection
- Help message generation
- Version display

```bash
# Show help
xxlang examples/scripts/cmdline/cli_demo.xxl --help

# Basic usage
xxlang examples/scripts/cmdline/cli_demo.xxl --name John --count 3

# With verbose output
xxlang examples/scripts/cmdline/cli_demo.xxl --verbose --file data.txt file1.txt file2.txt
```

### `echo.xxl`

A simple `echo`-like utility demonstrating practical CLI implementation:

- Combined flag parsing (`-ne` for no newline + escape interpretation)
- Help message
- Backslash escape interpretation

```bash
# Basic echo
xxlang examples/scripts/cmdline/echo.xxl Hello World

# No trailing newline
xxlang examples/scripts/cmdline/echo.xxl -n "Hello"

# With escape sequences
xxlang examples/scripts/cmdline/echo.xxl -e "Line1\nLine2\tTabbed"
```

## Key Functions

### Get Arguments

```xxl
let env = load("env")
let args = env.args()  // Returns array of all arguments
// args[0] is the program name/first script
```

### Parse Options

```xxl
// Get option value with default
let getOption = fn(shortName, longName, default) {
    let args = env.args()
    let i = 0
    while i < array.len(args) {
        if args[i] == shortName || args[i] == longName {
            if i + 1 < array.len(args) {
                return args[i + 1]
            }
        }
        i = i + 1
    }
    return default
}

let name = getOption("-n", "--name", "default")
```

### Check Flags

```xxl
let hasFlag = fn(flagName) {
    let args = env.args()
    let i = 0
    while i < array.len(args) {
        if args[i] == flagName {
            return true
        }
        i = i + 1
    }
    return false
}

let verbose = hasFlag("--verbose")
```

## Related Standard Library Modules

- `env` - Environment variables and process info
- `array` - Array manipulation for argument handling
- `string` - String parsing and manipulation
- `fmt` - Formatted output
