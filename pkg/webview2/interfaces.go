//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains WebView2 interface GUIDs and VTable definitions.
package webview2

import "syscall"

// WebView2 interface GUIDs
var (
	// IID_ICoreWebView2 is the GUID for ICoreWebView2 interface.
	IID_ICoreWebView2 = syscall.GUID{
		Data1: 0x76eceabd,
		Data2: 0x461d,
		Data3: 0x4f7a,
		Data4: [8]byte{0xb1, 0x2e, 0x2e, 0xed, 0x29, 0x73, 0x72, 0xa9},
	}

	// IID_ICoreWebView2Controller is the GUID for ICoreWebView2Controller interface.
	IID_ICoreWebView2Controller = syscall.GUID{
		Data1: 0x4d00c4c1,
		Data2: 0xcf8c,
		Data3: 0x4a30,
		Data4: [8]byte{0xb0, 0x5d, 0x22, 0x3e, 0xab, 0x22, 0x89, 0x55},
	}

	// IID_ICoreWebView2Environment is the GUID for ICoreWebView2Environment interface.
	IID_ICoreWebView2Environment = syscall.GUID{
		Data1: 0xb96d755e,
		Data2: 0x0b59,
		Data3: 0x4d96,
		Data4: [8]byte{0xa7, 0xac, 0x8b, 0xe3, 0xee, 0x67, 0x1b, 0x27},
	}

	// IID_ICoreWebView2Settings is the GUID for ICoreWebView2Settings interface.
	IID_ICoreWebView2Settings = syscall.GUID{
		Data1: 0xe5622c44,
		Data2: 0xa61d,
		Data3: 0x46f3,
		Data4: [8]byte{0xb0, 0x66, 0xfc, 0x9e, 0x64, 0x0a, 0x15, 0x7a},
	}

	// IID_ICoreWebView2WebMessageReceivedEventHandler is the GUID for web message handler.
	IID_ICoreWebView2WebMessageReceivedEventHandler = syscall.GUID{
		Data1: 0x5720f181,
		Data2: 0x667c,
		Data3: 0x4e4a,
		Data4: [8]byte{0x94, 0x6d, 0x71, 0xd6, 0x3f, 0x3f, 0x97, 0x79},
	}

	// IID_ICoreWebView2ExecuteScriptCompletedHandler is the GUID for script execution handler.
	IID_ICoreWebView2ExecuteScriptCompletedHandler = syscall.GUID{
		Data1: 0x49511172,
		Data2: 0x67d3,
		Data3: 0x469a,
		Data4: [8]byte{0xb6, 0x8f, 0x0b, 0x70, 0x1d, 0x8d, 0x4f, 0x4d},
	}

	// IID_ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandler is for environment creation callback.
	IID_ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandler = syscall.GUID{
		Data1: 0x4c1c4cd7,
		Data2: 0xcb97,
		Data3: 0x4a95,
		Data4: [8]byte{0xa4, 0x7b, 0x9a, 0x69, 0x7e, 0xa5, 0x6b, 0xe7},
	}

	// IID_ICoreWebView2CreateCoreWebView2ControllerCompletedHandler is for controller creation callback.
	IID_ICoreWebView2CreateCoreWebView2ControllerCompletedHandler = syscall.GUID{
		Data1: 0x6c4819f3,
		Data2: 0xc9b7,
		Data3: 0x4260,
		Data4: [8]byte{0x8c, 0x26, 0x6b, 0xa3, 0x2a, 0x1c, 0x77, 0x9b},
	}

	// IID_ICoreWebView2NavigationCompletedEventHandler is for navigation completed callback.
	IID_ICoreWebView2NavigationCompletedEventHandler = syscall.GUID{
		Data1: 0xd1e3619a,
		Data2: 0x6347,
		Data3: 0x4588,
		Data4: [8]byte{0x84, 0xd6, 0x69, 0x9f, 0x85, 0x2d, 0x0c, 0x17},
	}
)

// ICoreWebView2VTable defines the vtable for ICoreWebView2 interface.
type ICoreWebView2VTable struct {
	QueryInterface                     uintptr
	AddRef                             uintptr
	Release                            uintptr
	GetSettings                        uintptr
	GetSource                          uintptr
	Navigate                           uintptr
	NavigateToString                   uintptr
	AddNavigationStarting              uintptr
	RemoveNavigationStarting           uintptr
	AddContentLoading                  uintptr
	RemoveContentLoading               uintptr
	AddSourceChanged                   uintptr
	RemoveSourceChanged                uintptr
	AddHistoryChanged                  uintptr
	RemoveHistoryChanged               uintptr
	AddNavigationCompleted             uintptr
	RemoveNavigationCompleted          uintptr
	AddFrameNavigationStarting         uintptr
	RemoveFrameNavigationStarting      uintptr
	AddFrameNavigationCompleted        uintptr
	RemoveFrameNavigationCompleted     uintptr
	AddScriptDialogOpening             uintptr
	RemoveScriptDialogOpening          uintptr
	AddPermissionRequested             uintptr
	RemovePermissionRequested          uintptr
	AddProcessFailed                   uintptr
	RemoveProcessFailed                uintptr
	AddScriptToExecuteOnDocumentCreated uintptr
	RemoveScriptToExecuteOnDocumentCreated uintptr
	ExecuteScript                      uintptr
	CapturePreview                     uintptr
	Reload                             uintptr
	PostWebMessageAsJson               uintptr
	PostWebMessageAsString             uintptr
	AddWebMessageReceived              uintptr
	RemoveWebMessageReceived           uintptr
	CallDevToolsProtocolMethod         uintptr
	GetBrowserProcessId                uintptr
	GetCanGoBack                       uintptr
	GetCanGoForward                    uintptr
	GoBack                             uintptr
	GoForward                          uintptr
	GetDevToolsProtocolEventReceiver   uintptr
	Stop                               uintptr
	AddNewWindowRequested              uintptr
	RemoveNewWindowRequested           uintptr
	AddDocumentTitleChanged            uintptr
	RemoveDocumentTitleChanged         uintptr
	GetDocumentTitle                   uintptr
	AddHostObjectToScript              uintptr
	RemoveHostObjectFromScript         uintptr
	OpenDevToolsWindow                 uintptr
	AddContainsFullScreenElementChanged uintptr
	RemoveContainsFullScreenElementChanged uintptr
	GetContainsFullScreenElement       uintptr
	AddWebResourceRequested            uintptr
	RemoveWebResourceRequested         uintptr
	AddWebResourceRequestedFilter      uintptr
	RemoveWebResourceRequestedFilter   uintptr
	AddWindowCloseRequested            uintptr
	RemoveWindowCloseRequested         uintptr
}

// ICoreWebView2ControllerVTable defines the vtable for ICoreWebView2Controller interface.
type ICoreWebView2ControllerVTable struct {
	QueryInterface                    uintptr
	AddRef                            uintptr
	Release                           uintptr
	GetIsVisible                      uintptr
	PutIsVisible                      uintptr
	GetBounds                         uintptr
	PutBounds                         uintptr
	GetZoomFactor                     uintptr
	PutZoomFactor                     uintptr
	AddZoomFactorChanged              uintptr
	RemoveZoomFactorChanged           uintptr
	SetBoundsAndZoomFactor            uintptr
	MoveFocus                         uintptr
	AddMoveFocusRequested             uintptr
	RemoveMoveFocusRequested          uintptr
	AddGotFocus                       uintptr
	RemoveGotFocus                    uintptr
	AddLostFocus                      uintptr
	RemoveLostFocus                   uintptr
	AddAcceleratorKeyPressed          uintptr
	RemoveAcceleratorKeyPressed       uintptr
	GetParentWindow                   uintptr
	PutParentWindow                   uintptr
	NotifyParentWindowPositionChanged uintptr
	Close                             uintptr
	GetCoreWebView2                   uintptr
}

// ICoreWebView2EnvironmentVTable defines the vtable for ICoreWebView2Environment interface.
type ICoreWebView2EnvironmentVTable struct {
	QueryInterface                           uintptr
	AddRef                                   uintptr
	Release                                  uintptr
	CreateCoreWebView2Controller             uintptr
	CreateWebResourceResponse                uintptr
	GetBrowserVersionString                  uintptr
	AddNewBrowserVersionAvailable            uintptr
	RemoveNewBrowserVersionAvailable         uintptr
	CreateCoreWebView2CompositionController  uintptr
	CreateCoreWebView2Pointer                uintptr
}

// ICoreWebView2SettingsVTable defines the vtable for ICoreWebView2Settings interface.
type ICoreWebView2SettingsVTable struct {
	QueryInterface                        uintptr
	AddRef                                uintptr
	Release                               uintptr
	GetIsScriptEnabled                    uintptr
	PutIsScriptEnabled                    uintptr
	GetIsWebMessageEnabled                uintptr
	PutIsWebMessageEnabled                uintptr
	GetAreDefaultScriptDialogsEnabled     uintptr
	PutAreDefaultScriptDialogsEnabled     uintptr
	GetIsStatusBarEnabled                 uintptr
	PutIsStatusBarEnabled                 uintptr
	GetAreDevToolsEnabled                 uintptr
	PutAreDevToolsEnabled                 uintptr
	GetAreDefaultContextMenusEnabled      uintptr
	PutAreDefaultContextMenusEnabled      uintptr
	GetAreHostObjectsAllowed              uintptr
	PutAreHostObjectsAllowed              uintptr
	GetIsZoomControlEnabled               uintptr
	PutIsZoomControlEnabled               uintptr
	GetIsBuiltInErrorPageEnabled          uintptr
	PutIsBuiltInErrorPageEnabled          uintptr
}

// ICoreWebView2WebMessageReceivedEventArgsVTable defines the vtable for web message event args.
type ICoreWebView2WebMessageReceivedEventArgsVTable struct {
	QueryInterface           uintptr
	AddRef                   uintptr
	Release                  uintptr
	GetSource                uintptr
	GetWebMessageAsJson      uintptr
	TryGetWebMessageAsString uintptr
}

// ICoreWebView2NavigationCompletedEventArgsVTable defines the vtable for navigation completed event args.
type ICoreWebView2NavigationCompletedEventArgsVTable struct {
	QueryInterface   uintptr
	AddRef           uintptr
	Release          uintptr
	GetIsSuccess     uintptr
	GetWebErrorStatus uintptr
	GetNavigationId  uintptr
}

// ICoreWebView2ExecuteScriptCompletedHandlerVTable defines the vtable for script execution callback.
type ICoreWebView2ExecuteScriptCompletedHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

// ICoreWebView2WebMessageReceivedEventHandlerVTable defines the vtable for web message callback.
type ICoreWebView2WebMessageReceivedEventHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

// ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandlerVTable for environment creation callback.
type ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

// ICoreWebView2CreateCoreWebView2ControllerCompletedHandlerVTable for controller creation callback.
type ICoreWebView2CreateCoreWebView2ControllerCompletedHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

// ICoreWebView2NavigationCompletedEventHandlerVTable for navigation completed callback.
type ICoreWebView2NavigationCompletedEventHandlerVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

// Rect represents a Windows RECT structure.
type Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// EventRegistrationToken represents a COM event registration token.
type EventRegistrationToken struct {
	Value int64
}
