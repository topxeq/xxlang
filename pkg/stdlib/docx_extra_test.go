// pkg/stdlib/docx_extra_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func createTestMap() *objects.Map {
	pairs := make(map[objects.HashKey]objects.MapPair)
	pairs[objects.NewString("key1").HashKey()] = objects.MapPair{
		Key:   objects.NewString("key1"),
		Value: objects.NewString("value1"),
	}
	pairs[objects.NewString("key2").HashKey()] = objects.MapPair{
		Key:   objects.NewString("key2"),
		Value: objects.NewInt(123),
	}
	pairs[objects.NewString("key3").HashKey()] = objects.MapPair{
		Key:   objects.NewString("key3"),
		Value: objects.TRUE,
	}
	return objects.NewMap(pairs)
}

func TestGetMapString(t *testing.T) {
	m := createTestMap()

	result := getMapString(m, "key1", "default")
	if result != "value1" {
		t.Errorf("expected 'value1', got '%s'", result)
	}

	result = getMapString(m, "missing", "default")
	if result != "default" {
		t.Errorf("expected 'default', got '%s'", result)
	}

	result = getMapString(nil, "key", "default")
	if result != "default" {
		t.Errorf("expected 'default' for nil map, got '%s'", result)
	}
}

func TestGetMapInt(t *testing.T) {
	m := createTestMap()

	result := getMapInt(m, "key2", 0)
	if result != 123 {
		t.Errorf("expected 123, got %d", result)
	}

	result = getMapInt(m, "missing", 999)
	if result != 999 {
		t.Errorf("expected 999, got %d", result)
	}

	result = getMapInt(nil, "key", 999)
	if result != 999 {
		t.Errorf("expected 999 for nil map, got %d", result)
	}
}

func TestGetMapBool(t *testing.T) {
	m := createTestMap()

	result := getMapBool(m, "key3", false)
	if !result {
		t.Error("expected true")
	}

	result = getMapBool(m, "missing", true)
	if !result {
		t.Error("expected default true")
	}

	result = getMapBool(nil, "key", false)
	if result {
		t.Error("expected false for nil map")
	}
}

func TestIntToStr(t *testing.T) {
	result := intToStr(42)
	if result != "42" {
		t.Errorf("expected '42', got '%s'", result)
	}

	result = intToStr(0)
	if result != "0" {
		t.Errorf("expected '0', got '%s'", result)
	}

	result = intToStr(-123)
	if result != "-123" {
		t.Errorf("expected '-123', got '%s'", result)
	}
}
