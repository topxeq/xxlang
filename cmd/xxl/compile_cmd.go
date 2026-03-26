package main

import (
	"encoding/binary"
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
	gui      bool // Windows only: create GUI executable (no console)
}

// Magic marker for embedded bytecode
const bytecodeMagic = "XXLANG_BYTECODE_V1"

// compileCmd implements the compile subcommand
func compileCmd(args []string) error {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)

	var flags compileFlags
	fs.StringVar(&flags.output, "o", "", "Output file path")
	fs.StringVar(&flags.target, "", "", "Cross-compile target (os/arch)")
	fs.BoolVar(&flags.bytecode, "bytecode", false, "Output as bytecode (.xxb) instead of executable")
	fs.BoolVar(&flags.gui, "gui", false, "Create GUI executable (Windows only, no console window)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	if len(fs.Args()) == 0 {
		fmt.Println("Usage: xxlang compile [options] <file>")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -o, --output path     Output file path")
		fmt.Println("      --target os/arch  Cross-compile for target OS/architecture")
		fmt.Println("      --bytecode        Output as bytecode (.xxb) instead of executable")
		fmt.Println("      --gui             Create GUI executable (Windows only, no console)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  xxlang compile script.xxl")
		fmt.Println("  xxlang compile -o program script.xxl")
		fmt.Println("  xxlang compile -o program.exe --target windows/amd64 script.xxl")
		fmt.Println("  xxlang compile --gui -o app.exe script.xxl    # GUI app, no console")
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
			if runtime.GOOS == "windows" || strings.Contains(flags.target, "windows") {
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

	// Compile using the register compiler (same as the default VM)
	c := compiler.NewRegCompiler()
	// Define preset global variables before compilation
	c.SymbolTable().Define("argsG")
	c.SymbolTable().Define("scriptPathG")
	if _, err := c.Compile(program); err != nil {
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
		// Create executable with embedded bytecode
		isGUI, err := createEmbeddedExecutable(bytecode, outputPath, flags.target, flags.gui)
		if err != nil {
			return fmt.Errorf("error creating executable: %v", err)
		}
		exeType := "executable"
		if isGUI {
			exeType = "GUI executable (no console)"
		}
		fmt.Printf("Compiled %s -> %s (%s)\n", inputPath, outputPath, exeType)
	}

	return nil
}

// createEmbeddedExecutable creates a standalone executable with embedded bytecode
// by appending the bytecode to a copy of the xxlang binary
// Returns true if the output is a GUI executable (no console window)
func createEmbeddedExecutable(bytecode *compiler.Bytecode, outputPath, target string, gui bool) (bool, error) {
	// Serialize bytecode to data
	bytecodeData, err := bytecode.Serialize()
	if err != nil {
		return false, fmt.Errorf("error serializing bytecode: %v", err)
	}

	// Get path to current xxlang binary
	xxlangPath, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("error getting xxlang path: %v", err)
	}

	// Check if we need GUI mode on Windows
	isGUI := gui || isGUIExecutable(xxlangPath)

	// Read the current executable
	exeData, err := os.ReadFile(xxlangPath)
	if err != nil {
		return false, fmt.Errorf("error reading executable: %v", err)
	}

	// Create the output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return false, fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	// Write the executable data
	if _, err := outFile.Write(exeData); err != nil {
		return false, fmt.Errorf("error writing executable data: %v", err)
	}

	// Write the magic marker
	if _, err := outFile.Write([]byte(bytecodeMagic)); err != nil {
		return false, fmt.Errorf("error writing magic marker: %v", err)
	}

	// Write bytecode length (8 bytes, little-endian)
	lengthBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(lengthBuf, uint64(len(bytecodeData)))
	if _, err := outFile.Write(lengthBuf); err != nil {
		return false, fmt.Errorf("error writing bytecode length: %v", err)
	}

	// Write the bytecode
	if _, err := outFile.Write(bytecodeData); err != nil {
		return false, fmt.Errorf("error writing bytecode: %v", err)
	}

	// Make it executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(outputPath, 0755); err != nil {
			return false, fmt.Errorf("error making executable: %v", err)
		}
	}

	return isGUI, nil
}

// isGUIExecutable checks if the current executable is a GUI (no console) version
func isGUIExecutable(path string) bool {
	// Check if the executable name contains 'w' before the extension (e.g., xxlw.exe)
	base := filepath.Base(path)
	return strings.HasPrefix(base, "xxlw") || strings.Contains(strings.ToLower(base), "xxlw")
}

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