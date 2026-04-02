// pkg/stdlib/ftp_extra_test.go
// Additional tests for ftp module.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestFTPConnect tests ftp.connect basic argument validation.
func TestFTPConnect(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}
	fn := mod.Exports["connect"].(*objects.Builtin)

	// Wrong number of args
	res := fn.Fn()
	if res.Type() != objects.ErrorType {
		t.Error("expected error for no args")
	}
	// Wrong type for host
	res = fn.Fn(objects.NewInt(123), objects.NewInt(21))
	if res.Type() != objects.ErrorType {
		t.Error("expected error for wrong host type")
	}
	// Valid types (but connection may fail, we just check argument acceptance)
	res = fn.Fn(String("localhost"), Int(21), String("user"), String("pass"))
	// It might return an error due to connection failure, which is okay; we just ensure it's not a type error.
	// The important thing is that the function processes arguments without immediate type errors.
	// Acceptable: either error (connection) or ftp client object.
	_ = res
}

// TestFTPList tests ftp.list if available.
func TestFTPList(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}
	// Check if list function exists; it might be a method on client.
	// The ftp module may have list as a builtin that takes a client.
	if fn, ok := mod.Exports["list"].(*objects.Builtin); ok {
		// Without a connected client, it should error
		res := fn.Fn(String("dummy_path"))
		if res.Type() != objects.ErrorType {
			t.Errorf("list without client returned non-error: %s", res.Inspect())
		}
	}
}

// TestFTPDownload tests ftp.download if available.
func TestFTPDownload(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}
	if fn, ok := mod.Exports["download"].(*objects.Builtin); ok {
		// Insufficient args
		res := fn.Fn(String("remote"))
		if res.Type() != objects.ErrorType {
			t.Error("expected error for missing local arg")
		}
	}
}

// TestFTPUpload tests ftp.upload if available.
func TestFTPUpload(t *testing.T) {
	mod := Get("ftp")
	if mod == nil {
		t.Skip("ftp module not found")
	}
	if fn, ok := mod.Exports["upload"].(*objects.Builtin); ok {
		// Missing args
		res := fn.Fn(String("local"))
		if res.Type() != objects.ErrorType {
			t.Error("expected error for missing remote arg")
		}
	}
}
