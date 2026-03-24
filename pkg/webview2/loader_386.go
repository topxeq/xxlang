//go:build windows && 386

package webview2

import _ "embed"

//go:embed WebView2Loader_x86.dll
var webview2LoaderDLL []byte
