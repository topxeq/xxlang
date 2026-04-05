// cmd/xxl/service.go
// Windows/Linux/MacOS service support for Xxlang
// Based on kardianos/service library

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// Service related global variables
var (
	serviceNameG    = "xxlang"
	basePathG       = ""
	configFileNameG = ""
	currentOSG      = ""
	serviceModeG    = false
)

// handleServiceCommands handles service-related command line arguments
func handleServiceCommands(args []string) {
	// Check if running in service mode
	if hasSwitch(args, "-service") {
		fmt.Println(serviceNameG, "V", Version, "is running in service mode")
		serviceModeG = true

		s := initSvc()
		if s == nil {
			logWithTime("Failed to init service")
			return
		}

		err := (*s).Run()
		if err != nil {
			logWithTime("Service failed to run: %v", err)
		}
		os.Exit(0)
	}

	// Install service
	if hasSwitch(args, "-installService") {
		s := initSvc()
		if s == nil {
			fmt.Println("failed to initialize service")
			return
		}

		fmt.Println("installing service \"" + serviceNameG + "\"...")
		err := (*s).Install()
		if err != nil {
			fmt.Println("failed to install service:", err.Error())
			return
		}
		fmt.Println("service installed - \"" + serviceNameG + "\"")
		os.Exit(0)
	}

	// Remove/uninstall service
	if hasSwitch(args, "-removeService") || hasSwitch(args, "-uninstallService") {
		s := initSvc()
		if s == nil {
			fmt.Println("failed to init service")
			return
		}

		// Stop first
		err := (*s).Stop()
		if err != nil {
			fmt.Println("failed to stop service:", err.Error())
		} else {
			fmt.Println("service stopped - \"" + serviceNameG + "\"")
		}

		// Uninstall
		err = (*s).Uninstall()
		if err != nil {
			fmt.Println("failed to remove service:", err.Error())
			return
		}
		fmt.Println("service removed - \"" + serviceNameG + "\"")
		os.Exit(0)
	}

	// Start service
	if hasSwitch(args, "-startService") {
		s := initSvc()
		if s == nil {
			fmt.Println("failed to init service")
			return
		}

		fmt.Println("starting service \"" + serviceNameG + "\"...")
		err := (*s).Start()
		if err != nil {
			fmt.Println("failed to start:", err.Error())
			return
		}
		fmt.Println("service started - \"" + serviceNameG + "\"")
		os.Exit(0)
	}

	// Stop service
	if hasSwitch(args, "-stopService") {
		s := initSvc()
		if s == nil {
			fmt.Println("failed to init service")
			return
		}

		err := (*s).Stop()
		if err != nil {
			fmt.Println("failed to stop service:", err.Error())
		} else {
			fmt.Println("service stopped - \"" + serviceNameG + "\"")
		}
		os.Exit(0)
	}

	// Restart service
	if hasSwitch(args, "-restartService") {
		s := initSvc()
		if s == nil {
			fmt.Println("failed to init service")
			return
		}

		// Stop first
		err := (*s).Stop()
		if err != nil {
			fmt.Println("failed to stop service:", err.Error())
		} else {
			fmt.Println("service stopped")
		}

		// Start
		fmt.Println("starting service...")
		err = (*s).Start()
		if err != nil {
			fmt.Println("failed to start:", err.Error())
			return
		}
		fmt.Println("service started - \"" + serviceNameG + "\"")
		os.Exit(0)
	}

	// Reinstall service
	if hasSwitch(args, "-reinstallService") {
		s := initSvc()
		if s == nil {
			fmt.Println("failed to init service")
			return
		}

		// Stop and uninstall
		err := (*s).Stop()
		if err != nil {
			fmt.Println("failed to stop service:", err.Error())
		} else {
			fmt.Println("service stopped")
		}

		err = (*s).Uninstall()
		if err != nil {
			fmt.Println("failed to remove service:", err.Error())
		} else {
			fmt.Println("service removed")
		}

		// Reinstall
		fmt.Println("installing service...")
		err = (*s).Install()
		if err != nil {
			fmt.Println("failed to install service:", err.Error())
			return
		}
		fmt.Println("service installed")

		// Start
		fmt.Println("starting service...")
		err = (*s).Start()
		if err != nil {
			fmt.Println("failed to start:", err.Error())
			return
		}
		fmt.Println("service started - \"" + serviceNameG + "\"")
		os.Exit(0)
	}
}

// hasSwitch checks if a command-line switch exists
func hasSwitch(args []string, switchName string) bool {
	for _, arg := range args {
		if arg == switchName {
			return true
		}
	}
	return false
}

// program implements service.Interface
type program struct {
	BasePath string
}

// Start is called when the service is started
func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

// Stop is called when the service is stopped
func (p *program) Stop(s service.Service) error {
	// Stop should return quickly
	return nil
}

// run is the main service loop
func (p *program) run() {
	doWork()
}

// initSvc initializes the service configuration
func initSvc() *service.Service {
	// Set OS-specific paths
	if getOSName() == "windows" {
		currentOSG = "win"
		if strings.TrimSpace(basePathG) == "" || basePathG == "." {
			basePathG = "C:\\xxlang"
		}
		configFileNameG = serviceNameG + "win.cfg"
	} else {
		currentOSG = "linux"
		if strings.TrimSpace(basePathG) == "" || basePathG == "." {
			basePathG = "/xxlang"
		}
		configFileNameG = serviceNameG + "linux.cfg"
	}

	// Create base directory if it doesn't exist
	if !fileExists(basePathG) {
		os.MkdirAll(basePathG, 0755)
	}

	// Set log file path
	logFile := filepath.Join(basePathG, serviceNameG+".log")
	_ = logFile // Will be used by logWithTime

	// Service configuration
	svcConfig := &service.Config{
		Name:        serviceNameG,
		DisplayName: "Xxlang Service",
		Description: "Xxlang Script Language Service V" + Version,
		Arguments:   []string{"-service"},
	}

	prg := &program{BasePath: basePathG}
	s, err := service.New(prg, svcConfig)

	if err != nil {
		logWithTime("unable to init service: %v", err)
		return nil
	}

	return &s
}

// doWork is the main service work function
func doWork() {
	logWithTime("%s V%s", serviceNameG, Version)
	logWithTime("os: %s, basePath: %s, config: %s", runtime.GOOS, basePathG, configFileNameG)
	logWithTime("command-line args: %v", os.Args)
	logWithTime("Service started.")

	// Read config file if exists
	cfgFileName := filepath.Join(basePathG, configFileNameG)
	if fileExists(cfgFileName) {
		config := loadSimpleMap(cfgFileName)
		if config != nil {
			if path, ok := config["xxlangBasePath"]; ok && path != "" {
				basePathG = path
			}
		}
	}

	// Start auto-remove task runner
	go runAutoRemoveTask()

	// Start thread task runner
	go runThreadTask()

	// Run regular tasks once at startup
	runRegularTasks()

	// Keep service running
	select {}
}

// runAutoRemoveTask runs scripts matching autoRemoveTask*.xxl pattern
// and deletes them after execution
func runAutoRemoveTask() {
	for {
		taskFiles := getFileList(basePathG, "autoRemoveTask*.xxl")

		if len(taskFiles) > 0 {
			for i, v := range taskFiles {
				code := loadStringFromFile(v)
				if code == "" {
					logWithTime("failed to load auto-remove task [%d] %s", i, v)
					continue
				}

				logWithTime("running auto-remove task: %s", v)

				// Execute the script
				executeCodeSafe(code, v)
				logWithTime("auto-remove task completed: %s", v)

				// Remove the script after execution
				os.Remove(v)
				logWithTime("auto-remove task deleted: %s", v)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// runThreadTask runs scripts matching threadTask*.xxl pattern
// each in a separate goroutine
func runThreadTask() {
	for {
		taskFiles := getFileList(basePathG, "threadTask*.xxl")

		if len(taskFiles) > 0 {
			for i, v := range taskFiles {
				code := loadStringFromFile(v)
				if code == "" {
					logWithTime("failed to load thread task [%d] %s", i, v)
					continue
				}

				logWithTime("starting thread task: %s", v)

				// Execute in separate goroutine
				go func(scriptPath, scriptCode string) {
					executeCodeSafe(scriptCode, scriptPath)
					logWithTime("thread task [%s] completed", scriptPath)
				}(v, code)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// runRegularTasks runs scripts matching task*.xxl pattern once at startup
func runRegularTasks() {
	taskFiles := getFileList(basePathG, "task*.xxl")

	if len(taskFiles) > 0 {
		for i, v := range taskFiles {
			code := loadStringFromFile(v)
			if code == "" {
				logWithTime("failed to load task [%d] %s", i, v)
				continue
			}

			logWithTime("running task: %s", v)

			// Execute the script
			executeCodeSafe(code, v)
			logWithTime("task completed: %s", v)
		}
	}
}

// Helper functions

func getOSName() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	default:
		return "unknown"
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func logWithTime(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("["+timestamp+"] "+format, args...)

	// Print to stdout
	fmt.Println(message)

	// Also append to log file
	logFile := filepath.Join(basePathG, serviceNameG+".log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(message + "\n")
		f.Close()
	}
}

func loadStringFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func loadSimpleMap(path string) map[string]string {
	content := loadStringFromFile(path)
	if content == "" {
		return nil
	}

	result := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

func getFileList(dir, pattern string) []string {
	files, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil
	}
	// Sort files alphabetically
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i] > files[j] {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
	return files
}

// executeCodeSafe compiles and executes xxlang source code without exiting on errors
// This is a safe version for use in service context
func executeCodeSafe(code, sourcePath string) (success bool) {
	defer func() {
		if r := recover(); r != nil {
			logWithTime("executeCodeSafe recovered from panic: %v", r)
			success = false
		}
	}()

	// Lexical analysis
	l := lexer.New(code)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		logWithTime("Parser error in %s: %v", sourcePath, formatParserErrors(p.Errors()))
		return false
	}

	// Compilation
	c := compiler.NewRegCompiler()
	argsGSymbol := c.SymbolTable().Define("argsG")
	scriptPathGSymbol := c.SymbolTable().Define("scriptPathG")
	c.SetSourceFile(sourcePath)

	if _, err := c.Compile(program); err != nil {
		logWithTime("Compiler error in %s: %v", sourcePath, err)
		return false
	}

	// Create main module for exports
	mainModule := &objects.Module{
		Name:    sourcePath,
		Exports: make(map[string]objects.Object),
	}

	// Prepare globals array with preset values
	globals := make([]vm.Value, compiler.GlobalsSize)

	// Set argsG - empty array for service scripts
	argsElements := make([]objects.Object, 0)
	globals[argsGSymbol.Index] = vm.NewObject(&objects.Array{Elements: argsElements})

	// Set scriptPathG - script path
	globals[scriptPathGSymbol.Index] = vm.NewObject(&objects.String{Value: sourcePath})

	// Execution
	bytecode := c.Bytecode()
	machine := vm.NewRegVMWithGlobals(bytecode, globals)
	machine.SetCurrentModule(mainModule)

	if err := machine.Run(); err != nil {
		logWithTime("Runtime error in %s: %v", sourcePath, err)
		return false
	}

	return true
}
