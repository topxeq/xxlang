//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains the main WebView2 implementation.
package webview2

import (
	"fmt"
	"sync"
	"sync/atomic"
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

	// Message queue for messages from JavaScript (thread-safe)
	messageQueue []string

	// Handlers (stored to prevent GC)
	envHandler        *EnvironmentHandler
	controllerHandler *ControllerHandler
	messageHandler    *MessageHandler
	scriptHandlers    map[int64]*ScriptHandler // Script handlers by ID (prevent GC)

	// Atomic flag for initialization
	inited uintptr

	// Closed flag (atomic for thread safety)
	closed uintptr
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
		callbacks:      make(map[string]func(args map[string]interface{})),
		eventTokens:    make(map[uintptr]EventRegistrationToken),
		scriptHandlers: make(map[int64]*ScriptHandler),
	}

	// Create and store handlers (prevents GC)
	wv.envHandler = newEnvironmentHandler(wv)
	wv.controllerHandler = newControllerHandler(wv)

	// Create window
	hwnd, err := wv.createWindow(config.Title, config.Width, config.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	wv.hwnd = hwnd

	// Create WebView2 environment
	var userDataPtr *uint16
	if config.UserDataFolder != "" {
		userDataPtr = StringToLPCWSTR(config.UserDataFolder)
	}

	hr, _, _ := syscall.Syscall6(
		createEnvWithOps,
		4,
		0, // browserExecutableFolder
		uintptr(unsafe.Pointer(userDataPtr)),
		0, // environmentOptions
		uintptr(unsafe.Pointer(wv.envHandler)),
		0, 0,
	)

	if hr != 0 {
		wv.destroyWindow()
		return nil, fmt.Errorf("CreateCoreWebView2EnvironmentWithOptions failed: hr=0x%x", hr)
	}

	// Process messages until initialization completes
	var msg Msg
	for {
		if atomic.LoadUintptr(&wv.inited) != 0 {
			break
		}
		ret, _, _ := getMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 {
			wv.destroyWindow()
			return nil, fmt.Errorf("message loop quit during initialization")
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}

	// Check if we have a controller
	if wv.controller == nil {
		wv.destroyWindow()
		return nil, fmt.Errorf("controller creation failed")
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

// setupMessageHandler sets up the web message handler.
func (wv *WebView2) setupMessageHandler() error {
	if wv.webview == nil {
		return fmt.Errorf("webview not created")
	}

	// Create and store message handler
	wv.messageHandler = newMessageHandler(wv)
	var token EventRegistrationToken

	hr, _, _ := syscall.Syscall(
		wv.webview.vtbl.AddWebMessageReceived,
		3,
		uintptr(unsafe.Pointer(wv.webview)),
		uintptr(unsafe.Pointer(wv.messageHandler)),
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

	// Get settings - GetSettings returns an ICoreWebView2Settings* (interface pointer)
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

	// Cast the interface pointer directly to *ICoreWebView2Settings
	wv.settings = (*ICoreWebView2Settings)(unsafe.Pointer(settingsPtr))

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

	// Create wrapper callback that cleans up handler after invocation
	wrapperCallback := func(result string, err error) {
		// Clean up handler from map after callback completes
		wv.mu.Lock()
		delete(wv.scriptHandlers, id)
		wv.mu.Unlock()
		// Call original callback if provided
		if callback != nil {
			callback(result, err)
		}
	}

	handler := wv.createExecuteScriptHandler(id, wrapperCallback)

	// Store handler to prevent GC until callback is invoked
	wv.mu.Lock()
	wv.scriptHandlers[id] = handler
	wv.mu.Unlock()

	hr, _, _ := syscall.Syscall(
		wv.webview.vtbl.ExecuteScript,
		3,
		uintptr(unsafe.Pointer(wv.webview)),
		uintptr(scriptBSTR),
		uintptr(unsafe.Pointer(handler)),
	)

	if hr != 0 {
		// Remove handler on error
		wv.mu.Lock()
		delete(wv.scriptHandlers, id)
		wv.mu.Unlock()
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

// Poll processes a single message without blocking.
// Returns true if a message was processed, false otherwise.
// Use this for non-blocking GUI updates while doing computation.
func (wv *WebView2) Poll() bool {
	return processSingleMessage()
}

// PopMessage returns the next message from the queue.
// Returns empty string if no messages are available.
func (wv *WebView2) PopMessage() string {
	wv.mu.Lock()
	defer wv.mu.Unlock()
	if len(wv.messageQueue) == 0 {
		return ""
	}
	msg := wv.messageQueue[0]
	wv.messageQueue = wv.messageQueue[1:]
	return msg
}

// HasMessages returns true if there are messages waiting in the queue.
func (wv *WebView2) HasMessages() bool {
	wv.mu.Lock()
	defer wv.mu.Unlock()
	return len(wv.messageQueue) > 0
}

// IsClosed returns true if the WebView2 window has been closed.
func (wv *WebView2) IsClosed() bool {
	return atomic.LoadUintptr(&wv.closed) != 0
}

// SetClosed sets the closed state.
func (wv *WebView2) SetClosed(closed bool) {
	if closed {
		atomic.StoreUintptr(&wv.closed, 1)
	} else {
		atomic.StoreUintptr(&wv.closed, 0)
	}
}

// Close closes the WebView2 instance.
func (wv *WebView2) Close() {
	// Mark as closed first
	wv.SetClosed(true)

	if wv.controller != nil {
		syscall.Syscall(wv.controller.vtbl.Close, 1,
			uintptr(unsafe.Pointer(wv.controller)), 0, 0)
		wv.controller = nil
	}

	wv.destroyWindow()
	COMUninitialize()
}

// HandleMessage handles a web message from JavaScript.
// This is called from a COM callback, so we must be careful:
// - Only do minimal work here
// - Don't call Go callbacks or do JSON parsing
// - Just queue the message for later processing
func (wv *WebView2) handleMessage(msg string) {
	wv.mu.Lock()
	// Limit queue size to prevent memory issues
	if len(wv.messageQueue) < 1000 {
		wv.messageQueue = append(wv.messageQueue, msg)
	}
	wv.mu.Unlock()
}
