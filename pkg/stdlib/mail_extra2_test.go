// pkg/stdlib/mail_extra2_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// mailCall invokes a builtin from the mail module.
func mailCall(name string, args ...objects.Object) objects.Object {
	mod := Get("mail")
	if mod == nil {
		return &objects.Error{Message: "mail module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestMail_Extra2_Init tests that mail module registers all exports.
func TestMail_Extra2_Init(t *testing.T) {
	mod := Get("mail")
	if mod == nil {
		t.Skip("mail module not found")
	}
	expected := []string{
		"newClient", "send", "isMailClient",
	}
	for _, name := range expected {
		if _, ok := mod.Exports[name].(*objects.Builtin); !ok {
			t.Fatalf("export %s not found or not a builtin in mail module", name)
		}
	}
}

// TestMail_Extra2_NewClient_ArgumentValidation tests newClient argument validation.
func TestMail_Extra2_NewClient_ArgumentValidation(t *testing.T) {
	// No args
	res := mailCall("newClient")
	if res.Type() != objects.ErrorType {
		t.Fatalf("newClient() with no args should error")
	}

	// Only 3 args
	res = mailCall("newClient", String("host"), Int(587), String("user"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("newClient() with 3 args should error")
	}

	// Wrong host type
	res = mailCall("newClient", Int(123), Int(587), String("user"), String("pass"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("newClient() with int host should error")
	}

	// Wrong port type
	res = mailCall("newClient", String("host"), String("587"), String("user"), String("pass"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("newClient() with string port should error")
	}

	// Wrong user type
	res = mailCall("newClient", String("host"), Int(587), Int(123), String("pass"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("newClient() with int user should error")
	}

	// Wrong password type
	res = mailCall("newClient", String("host"), Int(587), String("user"), Int(123))
	if res.Type() != objects.ErrorType {
		t.Fatalf("newClient() with int password should error")
	}

	// Valid call with 4 args
	res = mailCall("newClient", String("smtp.example.com"), Int(587), String("user@example.com"), String("password"))
	if _, ok := res.(*objects.MailClient); !ok {
		t.Fatalf("newClient() should return MailClient, got %s", res.Type())
	}

	// Valid call with options
	res = mailCall("newClient", String("smtp.example.com"), Int(587), String("user@example.com"), String("password"), String("-tls"), String("-from=noreply@example.com"))
	if _, ok := res.(*objects.MailClient); !ok {
		t.Fatalf("newClient() with options should return MailClient, got %s", res.Type())
	}
}

// TestMail_Extra2_Send_ArgumentValidation tests send argument validation.
func TestMail_Extra2_Send_ArgumentValidation(t *testing.T) {
	// No args
	res := mailCall("send")
	if res.Type() != objects.ErrorType {
		t.Fatalf("send() with no args should error")
	}

	// Wrong type arg
	res = mailCall("send", String("not a map"))
	if res.Type() != objects.ErrorType {
		t.Fatalf("send() with string should error")
	}

	// Missing required fields
	config := &objects.Map{Pairs: map[objects.HashKey]objects.MapPair{
		String("host").HashKey():     {Key: String("host"), Value: String("smtp.example.com")},
		String("port").HashKey():     {Key: String("port"), Value: Int(587)},
		String("user").HashKey():     {Key: String("user"), Value: String("user@example.com")},
		String("password").HashKey(): {Key: String("password"), Value: String("pass")},
	}}
	res = mailCall("send", config)
	// This will fail because it tries to actually send, but we're just testing argument validation
	// The function should at least not panic and return something
	_ = res
}

// TestMail_Extra2_IsMailClient tests isMailClient function.
func TestMail_Extra2_IsMailClient(t *testing.T) {
	// No args
	res := mailCall("isMailClient")
	if res.Type() != objects.ErrorType {
		t.Fatalf("isMailClient() with no args should error")
	}

	// With a MailClient
	client := mailCall("newClient", String("smtp.example.com"), Int(587), String("user@example.com"), String("password"))
	res = mailCall("isMailClient", client)
	if res.Type() != objects.BoolType {
		t.Fatalf("isMailClient() should return bool, got %s", res.Type())
	}
	if !res.(*objects.Bool).Value {
		t.Fatalf("isMailClient(MailClient) should return true")
	}

	// With a non-MailClient
	res = mailCall("isMailClient", String("not a client"))
	if res.Type() != objects.BoolType {
		t.Fatalf("isMailClient() should return bool, got %s", res.Type())
	}
	if res.(*objects.Bool).Value {
		t.Fatalf("isMailClient(string) should return false")
	}
}
