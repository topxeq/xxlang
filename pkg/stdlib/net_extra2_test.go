// pkg/stdlib/net_extra2_test.go
// Additional tests for net module to improve coverage of init and builtins.
package stdlib

import (
	"testing"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

// netCall invokes a function from the net module, returning an objects.Object.
func netCall(name string, args ...objects.Object) objects.Object {
	mod := Get("net")
	if mod == nil {
		return &objects.Error{Message: "net module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// Test that the net module is registered and exposes the expected builtins.
func TestNetModuleExportsExist(t *testing.T) {
	t.Log("Verifying net module is registered with expected exports")
	mod := Get("net")
	if mod == nil {
		t.Fatalf("net module not registered in init")
	}
	expected := []string{
		"get", "post", "request", "head", "download",
		"setTimeout", "isOK", "isRedirect", "isClientError", "isServerError",
		"getJson", "postJson",
	}
	for _, name := range expected {
		if _, ok := mod.Exports[name].(*objects.Builtin); !ok {
			t.Fatalf("net module missing builtin: %s", name)
		}
	}
}

// Comprehensive argument validation tests to exercise init-related code paths
func TestNetArgumentValidation(t *testing.T) {
	// Do not run in parallel to avoid mutating shared httpClient timeout state across tests.
	t.Run("Get_ArgumentValidation", func(t *testing.T) {
		// No args -> error
		res := netCall("get")
		if _, ok := res.(*objects.Error); !ok {
			t.Error("get() with no args should error")
		}
		// Non-string URL -> error
		res = netCall("get", Int(123))
		if _, ok := res.(*objects.Error); !ok {
			t.Error("get() with non-string URL should error")
		}
	})

	t.Run("Post_ArgumentValidation", func(t *testing.T) {
		// No args -> error
		res := netCall("post")
		if res == nil {
			t.Fatalf("post() returned nil")
		}
		if _, ok := res.(*objects.Error); !ok {
			t.Error("post() with no args should error")
		}
		// URL ok, but body not a string -> error
		res = netCall("post", String("http://example.invalid"), Int(1))
		if _, ok := res.(*objects.Error); !ok {
			t.Error("post() with non-string body should error")
		}
	})

	t.Run("Request_ArgumentValidation", func(t *testing.T) {
		// Not enough args -> error
		res := netCall("request")
		if _, ok := res.(*objects.Error); !ok {
			t.Error("request() with 0 args should error")
		}
		// Wrong method type -> error (empty method invalid)
		res = netCall("request", String(""))
		if _, ok := res.(*objects.Error); !ok {
			t.Error("request() with invalid method type should error")
		}
	})

	t.Run("Head_ArgumentValidation", func(t *testing.T) {
		// 0 args error
		res := netCall("head")
		if _, ok := res.(*objects.Error); !ok {
			t.Error("head() with no args should error")
		}
		// Non-string arg error
		res = netCall("head", Int(1))
		if _, ok := res.(*objects.Error); !ok {
			t.Error("head() with non-string arg should error")
		}
	})

	t.Run("SetTimeout_Validation_and_Behavior", func(t *testing.T) {
		// Save current timeout to restore later
		old := httpClientTimeoutSafe()
		defer restoreHttpClientTimeoutSafe(old)

		// Valid set
		res := netCall("setTimeout", Int(1))
		if s, ok := res.(*objects.Null); ok && s == nil {
			// nothing
		} else {
			// Accepts Null return via Inspect
			if res.Inspect() != "null" {
				t.Fatalf("setTimeout should return null, got: %v", res.Inspect())
			}
		}
		if httpClientTimeoutSafe() != time.Second {
			t.Fatalf("expected timeout to be 1s, got %v", httpClientTimeoutSafe())
		}

		// Invalid type -> error
		res = netCall("setTimeout", String("not an int"))
		if _, ok := res.(*objects.Error); !ok {
			t.Error("setTimeout() with non-integer should error")
		}
	})

	t.Run("IsOKAndOtherHelpers", func(t *testing.T) {
		// isOK with valid int
		res := netCall("isOK", Int(200))
		if b, ok := res.(*objects.Bool); !ok || !b.Value {
			t.Fatalf("isOK(200) should be true, got %T %v", res, res.Inspect())
		}
		// isOK with non-int -> error
		res = netCall("isOK", String("200"))
		if _, ok := res.(*objects.Error); !ok {
			t.Error("isOK() with non-int should error")
		}
		// isRedirect
		res = netCall("isRedirect", Int(301))
		if b, ok := res.(*objects.Bool); !ok || !b.Value {
			t.Fatalf("isRedirect(301) should be true, got %T %v", res, res.Inspect())
		}
		// isClientError
		res = netCall("isClientError", Int(404))
		if b, ok := res.(*objects.Bool); !ok || !b.Value {
			t.Fatalf("isClientError(404) should be true, got %T %v", res, res.Inspect())
		}
		// isServerError
		res = netCall("isServerError", Int(500))
		if b, ok := res.(*objects.Bool); !ok || !b.Value {
			t.Fatalf("isServerError(500) should be true, got %T %v", res, res.Inspect())
		}
	})

	t.Run("JsonHelpers_Validation", func(t *testing.T) {
		// getJson and postJson should error on wrong arg counts
		res := netCall("getJson")
		if _, ok := res.(*objects.Error); !ok {
			t.Error("getJson() with no args should error")
		}
		res = netCall("postJson", String("http://example.invalid"))
		if _, ok := res.(*objects.Error); !ok {
			t.Error("postJson() with insufficient args should error")
		}
	})
}

// Helpers to access and restore httpClient timeout safely without leaking state across tests
func httpClientTimeoutSafe() time.Duration {
	// net.go defines httpClient with Timeout; we rely on its value here
	// Access via the same package to avoid exporting
	return httpClient.Timeout
}

func restoreHttpClientTimeoutSafe(d time.Duration) {
	httpClient.Timeout = d
}
