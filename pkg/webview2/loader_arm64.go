//go:build windows && arm64

package webview2

import _ "embed"

//go:embed WebView2Loader_arm64.dll
var webview2LoaderDLL []byte
