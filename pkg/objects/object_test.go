// pkg/objects/object_test.go
package objects

import "testing"

func TestObjectType(t *testing.T) {
	tests := []struct {
		obj      Object
		expected ObjectType
	}{
		{NULL, NullType},
		{TRUE, BoolType},
		{FALSE, BoolType},
	}

	for _, tt := range tests {
		if got := tt.obj.Type(); got != tt.expected {
			t.Errorf("object.Type() = %s, want %s", got, tt.expected)
		}
	}
}

func TestNullInspect(t *testing.T) {
	if got := NULL.Inspect(); got != "null" {
		t.Errorf("NULL.Inspect() = %s, want null", got)
	}
}

func TestBoolInspect(t *testing.T) {
	if got := TRUE.Inspect(); got != "true" {
		t.Errorf("TRUE.Inspect() = %s, want true", got)
	}
	if got := FALSE.Inspect(); got != "false" {
		t.Errorf("FALSE.Inspect() = %s, want false", got)
	}
}

func TestBoolToBool(t *testing.T) {
	if TRUE.ToBool() != TRUE {
		t.Error("TRUE.ToBool() should return TRUE")
	}
	if FALSE.ToBool() != FALSE {
		t.Error("FALSE.ToBool() should return FALSE")
	}
}
