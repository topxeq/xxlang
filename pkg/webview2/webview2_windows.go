//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains the main WebView2 implementation.
package webview2

import (
	"encoding/json"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

// WebView2 represents a WebView2 instance.
type WebView2 struct {
	hwnd       uintptr
	controller *ICoreWebView2Controller
	webview    *ICoreWebView2
	settings   *ICoreWebView2Settings
	env        *ICoreWebView2Environment

	mu          sync.Mutex
	callbacks   map[string]func(args map[string]interface{})
	scriptID    int64
	eventTokens map[uintptr]EventRegistrationToken

	// Channels for async operations
	envCreated    chan error
	controllerCreated chan error
}

// ICoreWebView2Controller wrapper.
type ICoreWebView2Controller struct {
	vtbl *ICoreWebView2ControllerVTable
}

// ICoreWebView2 wrapper.
type ICoreWebView2 struct {
	vtbl *ICoreWebView2VTable
}

// ICoreWebView2Environment wrapper.
type ICoreWebView2Environment struct {
	vtbl *ICoreWebView2EnvironmentVTable
}

// ICoreWebView2Settings wrapper.
type ICoreWebView2Settings struct {
	vtbl *ICoreWebView2SettingsVTable
}

// WebView2Config holds configuration for WebView2.
type WebView2Config struct {
	Title          string
	Width          int
	Height         int
	UserDataFolder string
	Debug          bool
}

var (
	loaderModule     *MemoryModule
	loaderOnce       sync.Once
	loaderErr        error
	createEnvWithOps uintptr
	getInstalledVersion uintptr
	compareVersions  uintptr
)

// initLoader initializes the WebView2Loader from embedded DLL.
// It tries memory loading first, falls back to temp file if needed.
func initLoader() error {
	loaderOnce.Do(func() {
		// Try memory loading first
		var err error
		loaderModule, err = loadFromMemory(webview2LoaderDLL)
		if err == nil {
			// Memory loading succeeded, get procedure addresses
			createEnvWithOps, err = loaderModule.GetProc("CreateCoreWebView2EnvironmentWithOptions")
			if err == nil {
				getInstalledVersion, err = loaderModule.GetProc("GetAvailableCoreWebView2BrowserVersionString")
			}
			if err == nil {
				compareVersions, _ = loaderModule.GetProc("CompareBrowserVersions")
			}
		}

		if err != nil {
			// Memory loading failed, fall back to temp file
			loaderModule = nil
			if loadErr := loadFromTempFile(); loadErr != nil {
				loaderErr = fmt.Errorf("memory loading failed: %w, temp file loading failed: %v", err, loadErr)
				return
			}
		}
	})
	return loaderErr
}

// loadFromTempFile loads the DLL from a temporary file.
func loadFromTempFile() error {
	tempDir := LPCWSTRToString(getTempPath())
	if tempDir == "" {
		tempDir = "C:\\Windows\\Temp"
	}

	dllPath := tempDir + "\\WebView2Loader_xxl.dll"

	// Write DLL to temp file
	if err := writeDLLToFile(dllPath); err != nil {
		return fmt.Errorf("failed to write DLL to temp file: %w", err)
	}

	// Load with system LoadLibrary
	dllNamePtr, _ := syscall.UTF16PtrFromString(dllPath)
	handle, _, err := loadLibraryW.Call(uintptr(unsafe.Pointer(dllNamePtr)))
	if handle == 0 {
		return fmt.Errorf("failed to load DLL: %w", err)
	}

	// Get procedure addresses
	namePtr, _ := syscall.BytePtrFromString("CreateCoreWebView2EnvironmentWithOptions")
	createEnvWithOps, _, _ = getProcAddress.Call(handle, uintptr(unsafe.Pointer(namePtr)))
	if createEnvWithOps == 0 {
		return fmt.Errorf("failed to get CreateCoreWebView2EnvironmentWithOptions")
	}

	namePtr, _ = syscall.BytePtrFromString("GetAvailableCoreWebView2BrowserVersionString")
	getInstalledVersion, _, _ = getProcAddress.Call(handle, uintptr(unsafe.Pointer(namePtr)))
	if getInstalledVersion == 0 {
		return fmt.Errorf("failed to get GetAvailableCoreWebView2BrowserVersionString")
	}

	namePtr, _ = syscall.BytePtrFromString("CompareBrowserVersions")
	compareVersions, _, _ = getProcAddress.Call(handle, uintptr(unsafe.Pointer(namePtr)))

	return nil
}

// writeDLLToFile writes the embedded DLL to a file.
func writeDLLToFile(path string) error {
	f, err := syscall.CreateFile(
		syscall.StringToUTF16Ptr(path),
		syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.CREATE_ALWAYS,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(f)

	var written uint32
	err = syscall.WriteFile(f, webview2LoaderDLL, &written, nil)
	if err != nil {
		return err
	}

	return nil
}

// getTempPath gets the system temp directory.
func getTempPath() *uint16 {
	var path [syscall.MAX_PATH + 1]uint16
	getTempPathProc := kernel32DLL.NewProc("GetTempPathW")
	getTempPathProc.Call(uintptr(len(path)), uintptr(unsafe.Pointer(&path[0])))
	return &path[0]
}

// GetInstalledVersion returns the installed WebView2 runtime version.
func GetInstalledVersion() (string, error) {
	if err := initLoader(); err != nil {
		return "", err
	}

	var result *uint16
	hr, _, _ := syscall.Syscall(
		getInstalledVersion,
		2,
		0,
		uintptr(unsafe.Pointer(&result)),
		0,
	)

	if hr != 0 {
		return "", fmt.Errorf("failed to get version: hr=0x%x", hr)
	}
	defer CoTaskMemFree(uintptr(unsafe.Pointer(result)))

	return LPCWSTRToString(result), nil
}

// NewWebView2 creates a new WebView2 instance.
func NewWebView2(config WebView2Config) (*WebView2, error) {
	if err := COMInitialize(); err != nil {
		return nil, fmt.Errorf("COM initialization failed: %w", err)
	}

	if err := initLoader(); err != nil {
		return nil, fmt.Errorf("WebView2Loader initialization failed: %w", err)
	}

	wv := &WebView2{
		callbacks:       make(map[string]func(args map[string]interface{})),
		eventTokens:     make(map[uintptr]EventRegistrationToken),
		envCreated:      make(chan error, 1),
		controllerCreated: make(chan error, 1),
	}

	// Create window
	hwnd, err := wv.createWindow(config.Title, config.Width, config.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	wv.hwnd = hwnd

	// Create WebView2 environment
	if err := wv.createEnvironment(config.UserDataFolder); err != nil {
		wv.destroyWindow()
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Create controller
	if err := wv.createController(); err != nil {
		wv.destroyWindow()
		return nil, fmt.Errorf("failed to create controller: %w", err)
	}

	// Setup message handler
	if err := wv.setupMessageHandler(); err != nil {
		wv.Close()
		return nil, fmt.Errorf("failed to setup message handler: %w", err)
	}

	// Configure settings
	if err := wv.configureSettings(config.Debug); err != nil {
		wv.Close()
		return nil, fmt.Errorf("failed to configure settings: %w", err)
	}

	return wv, nil
}

// createEnvironment creates the WebView2 environment.
func (wv *WebView2) createEnvironment(userDataFolder string) error {
	// Create handler callback
	handler := wv.createEnvironmentCompletedHandler()

	var userDataPtr *uint16
	if userDataFolder != "" {
		userDataPtr = StringToLPCWSTR(userDataFolder)
	}

	hr, _, _ := syscall.Syscall6(
		createEnvWithOps,
		4,
		0, // browserExecutableFolder
		uintptr(unsafe.Pointer(userDataPtr)),
		0, // environmentOptions
		uintptr(unsafe.Pointer(handler)),
		0, 0,
	)

	if hr != 0 {
		return fmt.Errorf("CreateCoreWebView2EnvironmentWithOptions failed: hr=0x%x", hr)
	}

	// Wait for async completion
	return <-wv.envCreated
}

// createController creates the WebView2 controller.
func (wv *WebView2) createController() error {
	if wv.env == nil {
		return fmt.Errorf("environment not created")
	}

	handler := wv.createControllerCompletedHandler()

	hr, _, _ := syscall.Syscall(
		wv.env.vtbl.CreateCoreWebView2Controller,
		3,
		uintptr(unsafe.Pointer(wv.env)),
		wv.hwnd,
		uintptr(unsafe.Pointer(handler)),
	)

	if hr != 0 {
		return fmt.Errorf("CreateCoreWebView2Controller failed: hr=0x%x", hr)
	}

	return <-wv.controllerCreated
}

// setupMessageHandler sets up the web message handler.
func (wv *WebView2) setupMessageHandler() error {
	if wv.webview == nil {
		return fmt.Errorf("webview not created")
	}

	handler := wv.createWebMessageHandler()
	var token EventRegistrationToken

	hr, _, _ := syscall.Syscall(
		wv.webview.vtbl.AddWebMessageReceived,
		3,
		uintptr(unsafe.Pointer(wv.webview)),
		uintptr(unsafe.Pointer(handler)),
		uintptr(unsafe.Pointer(&token)),
	)

	if hr != 0 {
		return fmt.Errorf("AddWebMessageReceived failed: hr=0x%x", hr)
	}

	wv.eventTokens[wv.webview.vtbl.AddWebMessageReceived] = token
	return nil
}

// configureSettings configures WebView2 settings.
func (wv *WebView2) configureSettings(debug bool) error {
	if wv.controller == nil {
		return fmt.Errorf("controller not created")
	}

	// Get settings
	var settingsPtr uintptr
	hr, _, _ := syscall.Syscall(
		wv.webview.vtbl.GetSettings,
		2,
		uintptr(unsafe.Pointer(wv.webview)),
		uintptr(unsafe.Pointer(&settingsPtr)),
		0,
	)

	if hr != 0 || settingsPtr == 0 {
		return fmt.Errorf("GetSettings failed: hr=0x%x", hr)
	}

	wv.settings = &ICoreWebView2Settings{
		vtbl: (*ICoreWebView2SettingsVTable)(unsafe.Pointer(settingsPtr)),
	}

	// Configure settings
	debugVal := uintptr(0)
	if debug {
		debugVal = 1
	}

	// Enable/disable dev tools
	syscall.Syscall(wv.settings.vtbl.PutAreDevToolsEnabled, 2,
		uintptr(unsafe.Pointer(wv.settings)), debugVal, 0)

	// Enable/disable context menus
	syscall.Syscall(wv.settings.vtbl.PutAreDefaultContextMenusEnabled, 2,
		uintptr(unsafe.Pointer(wv.settings)), debugVal, 0)

	// Enable script and web messages
	syscall.Syscall(wv.settings.vtbl.PutIsScriptEnabled, 2,
		uintptr(unsafe.Pointer(wv.settings)), 1, 0)
	syscall.Syscall(wv.settings.vtbl.PutIsWebMessageEnabled, 2,
		uintptr(unsafe.Pointer(wv.settings)), 1, 0)

	// Set bounds to fill the window
	var rect Rect
	getClientRect := user32DLL.NewProc("GetClientRect")
	getClientRect.Call(wv.hwnd, uintptr(unsafe.Pointer(&rect)))

	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top

	syscall.Syscall(wv.controller.vtbl.PutBounds, 2,
		uintptr(unsafe.Pointer(wv.controller)),
		uintptr(unsafe.Pointer(&rect)), 0)

	_ = width
	_ = height

	return nil
}

// Navigate navigates to the specified URL.
func (wv *WebView2) Navigate(url string) error {
	if wv.webview == nil {
		return fmt.Errorf("webview not initialized")
	}

	urlPtr := StringToLPCWSTR(url)
	hr, _, _ := syscall.Syscall(
		wv.webview.vtbl.Navigate,
		2,
		uintptr(unsafe.Pointer(wv.webview)),
		uintptr(unsafe.Pointer(urlPtr)),
		0,
	)

	if hr != 0 {
		return fmt.Errorf("Navigate failed: hr=0x%x", hr)
	}
	return nil
}

// NavigateToString navigates to the HTML content string.
func (wv *WebView2) NavigateToString(html string) error {
	if wv.webview == nil {
		return fmt.Errorf("webview not initialized")
	}

	htmlBSTR := StringToBSTR(html)
	defer FreeBSTR(htmlBSTR)

	hr, _, _ := syscall.Syscall(
		wv.webview.vtbl.NavigateToString,
		2,
		uintptr(unsafe.Pointer(wv.webview)),
		uintptr(htmlBSTR),
		0,
	)

	if hr != 0 {
		return fmt.Errorf("NavigateToString failed: hr=0x%x", hr)
	}
	return nil
}

// ExecuteScript executes JavaScript and returns result via callback.
func (wv *WebView2) ExecuteScript(script string, callback func(result string, err error)) error {
	if wv.webview == nil {
		return fmt.Errorf("webview not initialized")
	}

	scriptBSTR := StringToBSTR(script)
	defer FreeBSTR(scriptBSTR)

	wv.mu.Lock()
	wv.scriptID++
	id := wv.scriptID
	wv.mu.Unlock()

	handler := wv.createExecuteScriptHandler(id, callback)

	hr, _, _ := syscall.Syscall(
		wv.webview.vtbl.ExecuteScript,
		3,
		uintptr(unsafe.Pointer(wv.webview)),
		uintptr(scriptBSTR),
		uintptr(unsafe.Pointer(handler)),
	)

	if hr != 0 {
		return fmt.Errorf("ExecuteScript failed: hr=0x%x", hr)
	}
	return nil
}

// PostWebMessageAsJson posts a web message as JSON.
func (wv *WebView2) PostWebMessageAsJson(json string) error {
	if wv.webview == nil {
		return fmt.Errorf("webview not initialized")
	}

	jsonBSTR := StringToBSTR(json)
	defer FreeBSTR(jsonBSTR)

	hr, _, _ := syscall.Syscall(
		wv.webview.vtbl.PostWebMessageAsJson,
		2,
		uintptr(unsafe.Pointer(wv.webview)),
		uintptr(jsonBSTR),
		0,
	)

	if hr != 0 {
		return fmt.Errorf("PostWebMessageAsJson failed: hr=0x%x", hr)
	}
	return nil
}

// BindFunction binds a function to be callable from JavaScript.
func (wv *WebView2) BindFunction(name string, fn func(args map[string]interface{})) {
	wv.mu.Lock()
	wv.callbacks[name] = fn
	wv.mu.Unlock()

	// Inject JavaScript bridge code
	bridgeCode := fmt.Sprintf(`
		if (!window.xxlang) {
			window.xxlang = {
				call: function(name, args) {
					return new Promise((resolve, reject) => {
						const msg = JSON.stringify({__fn: name, __args: args || {}});
						window.chrome.webview.postMessage(msg);
						window._xxlang_resolvers = window._xxlang_resolvers || {};
						window._xxlang_resolvers[name] = {resolve: resolve, reject: reject};
					});
				}
			};
		}
	`)
	wv.ExecuteScript(bridgeCode, nil)
}

// Run runs the WebView2 window message loop (blocking).
func (wv *WebView2) Run() {
	runMessageLoop(wv.hwnd)
}

// Close closes the WebView2 instance.
func (wv *WebView2) Close() {
	if wv.controller != nil {
		syscall.Syscall(wv.controller.vtbl.Close, 1,
			uintptr(unsafe.Pointer(wv.controller)), 0, 0)
		wv.controller = nil
	}

	wv.destroyWindow()
	COMUninitialize()
}

// HandleMessage handles a web message from JavaScript.
func (wv *WebView2) handleMessage(msg string) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		return
	}

	if fnName, ok := data["__fn"].(string); ok {
		wv.mu.Lock()
		fn, exists := wv.callbacks[fnName]
		wv.mu.Unlock()

		if exists {
			args, _ := data["__args"].(map[string]interface{})
			go fn(args)
		}
	}
}
