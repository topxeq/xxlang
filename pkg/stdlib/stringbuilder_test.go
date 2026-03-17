// pkg/stdlib/stringbuilder_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestStringBuilder_Module(t *testing.T) {
	module := Get("stringbuilder")
	if module == nil {
		t.Fatal("stringbuilder module not found")
	}
	if module.Name != "stringbuilder" {
		t.Errorf("expected module name 'stringbuilder', got %q", module.Name)
	}
}

func TestStringBuilder_Create(t *testing.T) {
	module := Get("stringbuilder")
	createFn, ok := module.Exports["create"]
	if !ok {
		t.Fatal("create function not found")
	}

	builtin, ok := createFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", createFn)
	}

	t.Run("create empty", func(t *testing.T) {
		result := builtin.Fn()
		sb, ok := result.(*objects.StringBuilder)
		if !ok {
			t.Fatalf("expected StringBuilder, got %T", result)
		}
		if sb.Len() != 0 {
			t.Errorf("expected empty StringBuilder, got len %d", sb.Len())
		}
	})

	t.Run("create with capacity", func(t *testing.T) {
		result := builtin.Fn(&objects.Int{Value: 100})
		sb, ok := result.(*objects.StringBuilder)
		if !ok {
			t.Fatalf("expected StringBuilder, got %T", result)
		}
		sb.Write("test")
		if sb.String() != "test" {
			t.Errorf("expected 'test', got %q", sb.String())
		}
	})

	t.Run("create with invalid capacity type", func(t *testing.T) {
		result := builtin.Fn(&objects.String{Value: "100"})
		if result.Type() != objects.ErrorType {
			t.Errorf("expected Error, got %T", result)
		}
	})

	t.Run("create with too many arguments", func(t *testing.T) {
		result := builtin.Fn(&objects.Int{Value: 100}, &objects.Int{Value: 200})
		if result.Type() != objects.ErrorType {
			t.Errorf("expected Error, got %T", result)
		}
	})
}

func TestStringBuilder_IsStringBuilder(t *testing.T) {
	module := Get("stringbuilder")
	isSB, ok := module.Exports["isStringBuilder"]
	if !ok {
		t.Fatal("isStringBuilder function not found")
	}

	builtin, ok := isSB.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", isSB)
	}

	t.Run("is StringBuilder", func(t *testing.T) {
		sb := objects.NewStringBuilder()
		result := builtin.Fn(sb)
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if boolResult.Value != true {
			t.Error("expected true for StringBuilder")
		}
	})

	t.Run("is not StringBuilder - Int", func(t *testing.T) {
		result := builtin.Fn(&objects.Int{Value: 42})
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if boolResult.Value != false {
			t.Error("expected false for Int")
		}
	})

	t.Run("is not StringBuilder - String", func(t *testing.T) {
		result := builtin.Fn(&objects.String{Value: "test"})
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if boolResult.Value != false {
			t.Error("expected false for String")
		}
	})

	t.Run("is not StringBuilder - Array", func(t *testing.T) {
		result := builtin.Fn(&objects.Array{Elements: []objects.Object{}})
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if boolResult.Value != false {
			t.Error("expected false for Array")
		}
	})

	t.Run("wrong argument count", func(t *testing.T) {
		result := builtin.Fn()
		if result.Type() != objects.ErrorType {
			t.Errorf("expected Error, got %T", result)
		}
	})
}
