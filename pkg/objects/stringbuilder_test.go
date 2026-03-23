// pkg/objects/stringbuilder_test.go
package objects

import (
	"strings"
	"testing"
)

func TestStringBuilder_Type(t *testing.T) {
	sb := NewStringBuilder()
	if sb.Type() != StringBuilderType {
		t.Errorf("expected StringBuilderType, got %s", sb.Type())
	}
	if sb.TypeTag() != TagStringBuilder {
		t.Errorf("expected TagStringBuilder, got %d", sb.TypeTag())
	}
}

func TestStringBuilder_Inspect(t *testing.T) {
	sb := NewStringBuilder()
	inspect := sb.Inspect()
	if !strings.Contains(inspect, "StringBuilder") {
		t.Errorf("expected Inspect to contain 'StringBuilder', got %s", inspect)
	}
}

func TestStringBuilder_ToBool(t *testing.T) {
	sb := NewStringBuilder()
	if sb.ToBool() != TRUE {
		t.Error("expected StringBuilder to be truthy")
	}
}

func TestStringBuilder_Write(t *testing.T) {
	sb := NewStringBuilder()

	n := sb.Write("Hello")
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	if sb.String() != "Hello" {
		t.Errorf("expected 'Hello', got %q", sb.String())
	}

	sb.Write(" World")
	if sb.String() != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", sb.String())
	}
}

func TestStringBuilder_WriteLine(t *testing.T) {
	sb := NewStringBuilder()

	n := sb.WriteLine("Hello")
	if n != 6 { // "Hello" + newline
		t.Errorf("expected 6 bytes written, got %d", n)
	}
	if sb.String() != "Hello\n" {
		t.Errorf("expected 'Hello\\n', got %q", sb.String())
	}
}

func TestStringBuilder_String(t *testing.T) {
	sb := NewStringBuilder()
	sb.Write("test")
	if sb.String() != "test" {
		t.Errorf("expected 'test', got %q", sb.String())
	}
}

func TestStringBuilder_Len(t *testing.T) {
	sb := NewStringBuilder()
	if sb.Len() != 0 {
		t.Errorf("expected empty builder to have len 0, got %d", sb.Len())
	}

	sb.Write("Hello")
	if sb.Len() != 5 {
		t.Errorf("expected len 5, got %d", sb.Len())
	}
}

func TestStringBuilder_Clear(t *testing.T) {
	sb := NewStringBuilder()
	sb.Write("Hello")
	sb.Clear()

	if sb.Len() != 0 {
		t.Errorf("expected cleared builder to have len 0, got %d", sb.Len())
	}
	if sb.String() != "" {
		t.Errorf("expected cleared builder to be empty, got %q", sb.String())
	}
}

func TestStringBuilder_Reset(t *testing.T) {
	sb := NewStringBuilder()
	sb.Write("Hello")
	sb.Reset()

	if sb.Len() != 0 {
		t.Errorf("expected reset builder to have len 0, got %d", sb.Len())
	}
}

func TestStringBuilder_Grow(t *testing.T) {
	sb := NewStringBuilder()
	// Grow should not panic
	sb.Grow(100)
	sb.Write("test")
	if sb.String() != "test" {
		t.Errorf("expected 'test', got %q", sb.String())
	}
}

func TestStringBuilder_Concurrent(t *testing.T) {
	sb := NewStringBuilder()
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func() {
			sb.Write("a")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if sb.Len() != 10 {
		t.Errorf("expected len 10, got %d", sb.Len())
	}
}

func TestStringBuilder_HashKey(t *testing.T) {
	sb1 := NewStringBuilder()
	sb2 := NewStringBuilder()

	// Different instances should have different hash keys
	if sb1.HashKey() == sb2.HashKey() {
		t.Error("expected different StringBuilder instances to have different hash keys")
	}
}

// Test StringBuilder methods
func TestStringBuilder_Methods(t *testing.T) {
	t.Run("len", func(t *testing.T) {
		sb := NewStringBuilder()
		sb.Write("Hello")

		method, ok := GetMethod(StringBuilderType, "len")
		if !ok {
			t.Fatal("len method not found")
		}

		result := method.Fn(sb)
		intResult, ok := result.(*Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if intResult.Value != 5 {
			t.Errorf("expected 5, got %d", intResult.Value)
		}
	})

	t.Run("write", func(t *testing.T) {
		sb := NewStringBuilder()

		method, ok := GetMethod(StringBuilderType, "write")
		if !ok {
			t.Fatal("write method not found")
		}

		result := method.Fn(sb, &String{Value: "Hello"})
		intResult, ok := result.(*Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if intResult.Value != 5 {
			t.Errorf("expected 5, got %d", intResult.Value)
		}
		if sb.String() != "Hello" {
			t.Errorf("expected 'Hello', got %q", sb.String())
		}
	})

	t.Run("writeLine", func(t *testing.T) {
		sb := NewStringBuilder()

		method, ok := GetMethod(StringBuilderType, "writeLine")
		if !ok {
			t.Fatal("writeLine method not found")
		}

		result := method.Fn(sb, &String{Value: "Hello"})
		intResult, ok := result.(*Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if intResult.Value != 6 {
			t.Errorf("expected 6, got %d", intResult.Value)
		}
		if sb.String() != "Hello\n" {
			t.Errorf("expected 'Hello\\n', got %q", sb.String())
		}
	})

	t.Run("toString", func(t *testing.T) {
		sb := NewStringBuilder()
		sb.Write("Hello")

		method, ok := GetMethod(StringBuilderType, "toString")
		if !ok {
			t.Fatal("toString method not found")
		}

		result := method.Fn(sb)
		strResult, ok := result.(*String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if strResult.Value != "Hello" {
			t.Errorf("expected 'Hello', got %q", strResult.Value)
		}
	})

	t.Run("clear", func(t *testing.T) {
		sb := NewStringBuilder()
		sb.Write("Hello")

		method, ok := GetMethod(StringBuilderType, "clear")
		if !ok {
			t.Fatal("clear method not found")
		}

		result := method.Fn(sb)
		if result != NULL {
			t.Errorf("expected NULL, got %T", result)
		}
		if sb.Len() != 0 {
			t.Errorf("expected cleared builder to have len 0, got %d", sb.Len())
		}
	})

	t.Run("isEmpty", func(t *testing.T) {
		sb := NewStringBuilder()

		method, ok := GetMethod(StringBuilderType, "isEmpty")
		if !ok {
			t.Fatal("isEmpty method not found")
		}

		result := method.Fn(sb)
		boolResult, ok := result.(*Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if boolResult.Value != true {
			t.Error("expected empty builder to be isEmpty true")
		}

		sb.Write("test")
		result = method.Fn(sb)
		boolResult = result.(*Bool)
		if boolResult.Value != false {
			t.Error("expected non-empty builder to be isEmpty false")
		}
	})

	t.Run("grow", func(t *testing.T) {
		sb := NewStringBuilder()

		method, ok := GetMethod(StringBuilderType, "grow")
		if !ok {
			t.Fatal("grow method not found")
		}

		result := method.Fn(sb, &Int{Value: 100})
		if result != NULL {
			t.Errorf("expected NULL, got %T", result)
		}
	})

	t.Run("typeOf", func(t *testing.T) {
		sb := NewStringBuilder()

		method, ok := GetMethod(StringBuilderType, "typeOf")
		if !ok {
			t.Fatal("typeOf method not found")
		}

		result := method.Fn(sb)
		strResult, ok := result.(*String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if strResult.Value != "STRING_BUILDER" {
			t.Errorf("expected 'STRING_BUILDER', got %q", strResult.Value)
		}
	})

	t.Run("toStr", func(t *testing.T) {
		sb := NewStringBuilder()
		sb.Write("test")

		method, ok := GetMethod(StringBuilderType, "toStr")
		if !ok {
			t.Fatal("toStr method not found")
		}

		result := method.Fn(sb)
		strResult, ok := result.(*String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if !strings.Contains(strResult.Value, "StringBuilder") {
			t.Errorf("expected toStr to contain 'StringBuilder', got %q", strResult.Value)
		}
	})
}

func TestStringBuilder_Cap(t *testing.T) {
	sb := NewStringBuilder()
	sb.Write("Hello")
	// Cap returns the actual capacity of the builder's internal buffer
	// Capacity may be larger than length due to internal growth
	cap := sb.Cap()
	if cap < 5 {
		t.Errorf("expected cap >= 5, got %d", cap)
	}
}

func TestStringBuilder_MethodErrors(t *testing.T) {
	t.Run("write wrong arg type", func(t *testing.T) {
		sb := NewStringBuilder()
		method, _ := GetMethod(StringBuilderType, "write")

		result := method.Fn(sb, &Int{Value: 42})
		if result.Type() != ErrorType {
			t.Errorf("expected Error, got %T", result)
		}
	})

	t.Run("grow wrong arg type", func(t *testing.T) {
		sb := NewStringBuilder()
		method, _ := GetMethod(StringBuilderType, "grow")

		result := method.Fn(sb, &String{Value: "100"})
		if result.Type() != ErrorType {
			t.Errorf("expected Error, got %T", result)
		}
	})

	t.Run("len wrong receiver", func(t *testing.T) {
		method, _ := GetMethod(StringBuilderType, "len")

		result := method.Fn(&String{Value: "test"})
		if result.Type() != ErrorType {
			t.Errorf("expected Error, got %T", result)
		}
	})
}
