package objects

import (
	"testing"
)

func TestWebView_NewWebView(t *testing.T) {
	wv := NewWebView(nil)
	if wv == nil {
		t.Fatal("NewWebView() returned nil")
	}
	if wv.closed {
		t.Error("New WebView should not be closed")
	}
}

func TestWebView_Type(t *testing.T) {
	wv := NewWebView(nil)
	if wv.Type() != WebViewType {
		t.Errorf("Type() = %v, want %v", wv.Type(), WebViewType)
	}
}

func TestWebView_TypeTag(t *testing.T) {
	wv := NewWebView(nil)
	if wv.TypeTag() != TagWebView {
		t.Errorf("TypeTag() = %v, want %v", wv.TypeTag(), TagWebView)
	}
}

func TestWebView_Inspect(t *testing.T) {
	wv := NewWebView(nil)
	if wv.Inspect() != "<WEBVIEW active>" {
		t.Errorf("Inspect() = %v, want <WEBVIEW active>", wv.Inspect())
	}

	wv.SetClosed(true)
	if wv.Inspect() != "<WEBVIEW closed>" {
		t.Errorf("Inspect() = %v, want <WEBVIEW closed>", wv.Inspect())
	}
}

func TestWebView_ToBool(t *testing.T) {
	wv := NewWebView(nil)
	if !wv.ToBool().Value {
		t.Error("ToBool() should return true for active WebView")
	}

	wv.SetClosed(true)
	if wv.ToBool().Value {
		t.Error("ToBool() should return false for closed WebView")
	}
}

func TestWebView_HashKey(t *testing.T) {
	wv := NewWebView(nil)
	hk := wv.HashKey()
	if hk.Type != WebViewType {
		t.Errorf("HashKey Type = %v, want %v", hk.Type, WebViewType)
	}
	if hk.Value == 0 {
		t.Error("HashKey Value should not be 0")
	}
}

func TestWebView_Equals(t *testing.T) {
	wv1 := NewWebView("handle1")
	wv2 := NewWebView("handle1")
	wv3 := NewWebView("handle2")
	wv4 := NewWebView(nil)

	if !wv1.Equals(wv2).Value {
		t.Error("Equals() should return true for same handle")
	}
	if wv1.Equals(wv3).Value {
		t.Error("Equals() should return false for different handle")
	}
	if wv4.Equals(nil).Value {
		t.Error("Equals() should return false for nil")
	}
}

func TestWebView_IsClosed(t *testing.T) {
	wv := NewWebView(nil)
	if wv.IsClosed() {
		t.Error("IsClosed() should return false initially")
	}

	wv.SetClosed(true)
	if !wv.IsClosed() {
		t.Error("IsClosed() should return true after SetClosed(true)")
	}
}

func TestWebView_Callbacks(t *testing.T) {
	wv := NewWebView(nil)

	fn := &Function{Name: "testCallback"}
	wv.AddCallback("test", fn)

	got, ok := wv.GetCallback("test")
	if !ok {
		t.Error("GetCallback() should return true for existing callback")
	}
	if got != fn {
		t.Error("GetCallback() returned wrong function")
	}

	_, ok = wv.GetCallback("nonexistent")
	if ok {
		t.Error("GetCallback() should return false for nonexistent callback")
	}
}

func TestWebView_Close(t *testing.T) {
	wv := NewWebView(nil)
	fn := &Function{Name: "testCallback"}
	wv.AddCallback("test", fn)

	wv.Close()

	if !wv.IsClosed() {
		t.Error("Close() should set closed to true")
	}

	_, ok := wv.GetCallback("test")
	if ok {
		t.Error("Close() should clear callbacks")
	}
}
