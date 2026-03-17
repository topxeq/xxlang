package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// compileFlags holds flags for the compile command
type compileFlags struct {
	output   string
	target   string
	bytecode bool
}

// compileCmd implements the compile subcommand
func compileCmd(args []string) error {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)

	var flags compileFlags
	fs.StringVar(&flags.output, "o", "", "Output file path")
	fs.StringVar(&flags.target, "", "", "Cross-compile target (os/arch)")
	fs.BoolVar(&flags.bytecode, "bytecode", false, "Output as bytecode (.xxb) instead of executable")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	if len(fs.Args()) == 0 {
		fmt.Println("Usage: xxlang compile [options] <file>")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -o, --output path     Output file path")
		fmt.Println("      --target os/arch  Cross-compile for target OS/architecture")
		fmt.Println("      --bytecode         Output as bytecode (.xxb) instead of executable")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  xxlang compile script.xxl")
		fmt.Println("  xxlang compile -o program script.xxl")
		fmt.Println("  xxlang compile -o program.exe --target windows/amd64 script.xxl")
		fmt.Println("  xxlang compile --bytecode script.xxl")
		return nil
	}

	inputPath := fs.Args()[0]

	// Validate input file exists
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}
	if info.IsDir() {
		return fmt.Errorf("input path is a directory: %s", inputPath)
	}

	// Read the source file
	code, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", inputPath, err)
	}

	// Determine output path
	var outputPath string
	if flags.output != "" {
		outputPath = flags.output
	} else {
		// Default: use input base name with appropriate extension
		baseName := strings.TrimSuffix(filepath.Base(inputPath), ".xxl")
		if flags.bytecode {
			outputPath = baseName + ".xxb"
		} else {
			outputPath = baseName
			if runtime.GOOS == "windows" {
				outputPath += ".exe"
			}
		}
	}

	// Parse the source
	l := lexer.New(string(code))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors in %s", formatErrors(p.Errors()))
	}

	// Compile
	c := compiler.New()
	// Define preset global variables before compilation
	c.DefineGlobal("argsG")
	c.DefineGlobal("scriptPathG")
	if err := c.Compile(program); err != nil {
		return fmt.Errorf("compile error: %v", err)
	}
	bytecode := c.Bytecode()

	// Handle different output formats
	if flags.bytecode {
		// Serialize to bytecode file
		if err := bytecode.SerializeToFile(outputPath); err != nil {
			return fmt.Errorf("error writing bytecode: %v", err)
		}
		fmt.Printf("Compiled %s -> %s (bytecode)\n", inputPath, outputPath)
	} else {
		// Create executable with embedded launcher
		if err := createExecutable(bytecode, outputPath, flags.target); err != nil {
			return fmt.Errorf("error creating executable: %v", err)
		}
		fmt.Printf("Compiled %s -> %s (executable)\n", inputPath, outputPath)
	}

	return nil
}

// createExecutable generates a standalone executable from bytecode
// This approach creates a shell script launcher that runs xxlang with the bytecode
func createExecutable(bytecode *compiler.Bytecode, outputPath, target string) error {
	// Serialize bytecode to data
	bytecodeData, err := bytecode.Serialize()
	if err != nil {
		return fmt.Errorf("error serializing bytecode: %v", err)
	}

	// Get path to xxlang binary
	xxlangPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting xxlang path: %v", err)
	}

	// Write bytecode to a file next to the executable
	bytecodeFile := outputPath + ".xxb"
	if err := os.WriteFile(bytecodeFile, bytecodeData, 0644); err != nil {
		return fmt.Errorf("error writing bytecode file: %v", err)
	}

	// Create a shell script launcher that runs xxlang with the bytecode file
	var launcherScript string
	if runtime.GOOS == "windows" {
		// Windows batch script
		launcherScript = fmt.Sprintf(`@echo off
"%s" run "%%~dp0.xxb"
`, filepath.Base(xxlangPath))
	} else {
		// Unix shell script
		launcherScript = `#!/bin/sh
exec "$(dirname "$0")/` + filepath.Base(xxlangPath) + `" run "$(dirname "$0")/` + filepath.Base(outputPath) + `.xxb"
`
	}

	if err := os.WriteFile(outputPath, []byte(launcherScript), 0755); err != nil {
		return fmt.Errorf("error writing launcher script: %v", err)
	}

	// Make it executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(outputPath, 0755); err != nil {
			return fmt.Errorf("error making executable: %v", err)
		}
	}

	return nil
}

// generateBytecodeData generates Go source for embedded bytecode data
func generateBytecodeData(data []byte) string {
	var sb strings.Builder
	sb.WriteString("// Code generated by xxlang compile. DO NOT EDIT.\n")
	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"fmt\"\n")
	sb.WriteString("\t\"os\"\n\n")
	sb.WriteString("\t\"github.com/topxeq/xxlang/pkg/compiler\"\n")
	sb.WriteString("\t\"github.com/topxeq/xxlang/pkg/vm\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("var bytecodeData = []byte{\n")
	for _, b := range data {
		sb.WriteString(fmt.Sprintf("\t0x%02x,\n", b))
	}
	sb.WriteString("}\n\n")
	sb.WriteString("func main() {\n")
	sb.WriteString("\t// Load and run the bytecode\n")
	sb.WriteString("\tbytecode, err := compiler.Deserialize(bytecodeData)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\tfmt.Fprintf(os.Stderr, \"Error loading bytecode: %%v\\n\", err)\n")
	sb.WriteString("\t\tos.Exit(1)\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tv := vm.New(bytecode)\n")
	sb.WriteString("\tif err := v.Run(); err != nil {\n")
	sb.WriteString("\t\tfmt.Fprintf(os.Stderr, \"Runtime error: %%v\\n\", err)\n")
	sb.WriteString("\t\tos.Exit(1)\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// launcherTemplate is the template for the generated launcher (not used with shell approach)
const launcherTemplate = `// Code generated by xxlang compile. DO NOT EDIT.
// This file is replaced by data.go which contains the actual bytecode.

package main

func main() {
	// Main entry point - bytecode is in data.go
}
`

// formatErrors formats parser errors into a single error
func formatErrors(errors []string) error {
	var sb strings.Builder
	sb.WriteString("parser errors:\n")
	for _, err := range errors {
		sb.WriteString("  ")
		sb.WriteString(err)
		sb.WriteString("\n")
	}
	return fmt.Errorf("%s", sb.String())
}
