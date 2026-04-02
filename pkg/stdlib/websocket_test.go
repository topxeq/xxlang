// pkg/stdlib/websocket_test.go
// Tests for WebSocket module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestWebSocketModule_Exists(t *testing.T) {
	mod := Get("websocket")
	if mod == nil {
		t.Fatal("websocket module not found")
	}
	if mod.Name != "websocket" {
		t.Errorf("expected module name 'websocket', got %s", mod.Name)
	}
}

func TestWebSocketModule_Exports(t *testing.T) {
	mod := Get("websocket")
	if mod == nil {
		t.Skip("websocket module not found")
	}

	expectedExports := []string{"upgrade", "readMsg", "sendText", "sendBinary", "close"}
	for _, name := range expectedExports {
		if _, ok := mod.Exports[name]; !ok {
			t.Errorf("expected export '%s' in websocket module", name)
		}
	}
}

func TestWebSocketModule_UpgradeArgCount(t *testing.T) {
	mod := Get("websocket")
	if mod == nil {
		t.Skip("websocket module not found")
	}

	fn, ok := mod.Exports["upgrade"]
	if !ok {
		t.Fatal("upgrade not found")
	}

	builtin := fn.(*objects.Builtin)

	// Wrong arg count
	result := builtin.Fn(objects.NewString("test"))
	err, ok := result.(*objects.Error)
	if !ok {
		t.Error("expected error for wrong arg count")
	}
	_ = err
}

func TestWebSocketModule_UpgradeWrongTypes(t *testing.T) {
	mod := Get("websocket")
	if mod == nil {
		t.Skip("websocket module not found")
	}

	fn := mod.Exports["upgrade"].(*objects.Builtin)

	// Pass wrong types
	result := fn.Fn(objects.NewInt(1), objects.NewInt(2))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong argument types")
	}
}

func TestWebSocketModule_ReadMsgArgCount(t *testing.T) {
	mod := Get("websocket")
	if mod == nil {
		t.Skip("websocket module not found")
	}

	fn := mod.Exports["readMsg"].(*objects.Builtin)

	result := fn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong arg count")
	}
}

func TestWebSocketModule_SendTextArgCount(t *testing.T) {
	mod := Get("websocket")
	if mod == nil {
		t.Skip("websocket module not found")
	}

	fn := mod.Exports["sendText"].(*objects.Builtin)

	result := fn.Fn(objects.NewString("test"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong arg count")
	}
}
