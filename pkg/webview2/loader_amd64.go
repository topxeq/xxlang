//go:build windows && amd64

package webview2

import _ "embed"

//go:embed WebView2Loader_x64.dll
var webview2LoaderDLL []byte
