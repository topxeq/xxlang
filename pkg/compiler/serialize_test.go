// pkg/compiler/serialize_test.go
package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestSerialize_Int(t *testing.T) {
	bc := &Bytecode{
		Constants:    []objects.Object{&objects.Int{Value: 42}},
		Instructions: []byte{0x01, 0x02, 0x03},
	}

	data, err := bc.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected serialized data")
	}
}

func TestDeserialize_Int(t *testing.T) {
	bc := &Bytecode{
		Constants:    []objects.Object{&objects.Int{Value: 42}},
		Instructions: []byte{0x01, 0x02, 0x03},
	}

	data, err := bc.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bc2.Constants) != 1 {
		t.Fatalf("expected 1 constant, got %d", len(bc2.Constants))
	}

	intObj, ok := bc2.Constants[0].(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", bc2.Constants[0])
	}
	if intObj.Value != 42 {
		t.Fatalf("expected 42, got %d", intObj.Value)
	}
}

func TestSerialize_Float(t *testing.T) {
	bc := &Bytecode{
		Constants:    []objects.Object{&objects.Float{Value: 3.14}},
		Instructions: []byte{},
	}

	data, err := bc.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	floatObj, ok := bc2.Constants[0].(*objects.Float)
	if !ok {
		t.Fatalf("expected Float, got %T", bc2.Constants[0])
	}
	if floatObj.Value != 3.14 {
		t.Fatalf("expected 3.14, got %f", floatObj.Value)
	}
}

func TestSerialize_String(t *testing.T) {
	bc := &Bytecode{
		Constants:    []objects.Object{&objects.String{Value: "hello world"}},
		Instructions: []byte{},
	}

	data, err := bc.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	strObj, ok := bc2.Constants[0].(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", bc2.Constants[0])
	}
	if strObj.Value != "hello world" {
		t.Fatalf("expected 'hello world', got %q", strObj.Value)
	}
}

func TestSerialize_Bool(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  bool
	}{
		{"true", true, true},
		{"false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var boolObj objects.Object = objects.FALSE
			if tt.value {
				boolObj = objects.TRUE
			}

			bc := &Bytecode{
				Constants:    []objects.Object{boolObj},
				Instructions: []byte{},
			}

			data, err := bc.Serialize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			bc2, err := Deserialize(data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result, ok := bc2.Constants[0].(*objects.Bool)
			if !ok {
				t.Fatalf("expected Bool, got %T", bc2.Constants[0])
			}
			if result.Value != tt.want {
				t.Errorf("expected %v, got %v", tt.want, result.Value)
			}
		})
	}
}

func TestSerialize_Null(t *testing.T) {
	bc := &Bytecode{
		Constants:    []objects.Object{objects.NULL},
		Instructions: []byte{},
	}

	data, err := bc.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bc2.Constants[0] != objects.NULL {
		t.Error("expected NULL")
	}
}

func TestSerialize_Array(t *testing.T) {
	bc := &Bytecode{
		Constants: []objects.Object{
			&objects.Array{
				Elements: []objects.Object{
					&objects.Int{Value: 1},
					&objects.Int{Value: 2},
					&objects.String{Value: "three"},
				},
			},
		},
		Instructions: []byte{},
	}

	data, err := bc.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := bc2.Constants[0].(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", bc2.Constants[0])
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestSerialize_Map(t *testing.T) {
	bc := &Bytecode{
		Constants: []objects.Object{
			&objects.Map{
				Pairs: map[objects.HashKey]objects.MapPair{
					objects.NewString("a").HashKey(): {
						Key:   objects.NewString("a"),
						Value: &objects.Int{Value: 1},
					},
					objects.NewString("b").HashKey(): {
						Key:   objects.NewString("b"),
						Value: &objects.Int{Value: 2},
					},
				},
			},
		},
		Instructions: []byte{},
	}

	data, err := bc.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := bc2.Constants[0].(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", bc2.Constants[0])
	}
	if len(m.Pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(m.Pairs))
	}
}

func TestSerialize_MultipleConstants(t *testing.T) {
	bc := &Bytecode{
		Constants: []objects.Object{
			&objects.Int{Value: 42},
			&objects.Float{Value: 3.14},
			&objects.String{Value: "hello"},
			objects.TRUE,
			objects.NULL,
			&objects.Array{Elements: []objects.Object{&objects.Int{Value: 1}}},
		},
		Instructions: []byte{0x01, 0x02},
	}

	data, err := bc.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bc2.Constants) != 6 {
		t.Fatalf("expected 6 constants, got %d", len(bc2.Constants))
	}
}

func TestDeserialize_InvalidMagic(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00}

	_, err := Deserialize(data)
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestDeserialize_InvalidVersion(t *testing.T) {
	data := []byte{
		'X', 'X', 'L', 'B',
		0x00, 0x00, 0x00, 99,
	}

	_, err := Deserialize(data)
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestSerializeToFile_Basic(t *testing.T) {
	bc := &Bytecode{
		Constants:    []objects.Object{&objects.Int{Value: 42}},
		Instructions: []byte{0x01},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.xxlbc")

	err := bc.SerializeToFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestDeserializeFromFile(t *testing.T) {
	bc := &Bytecode{
		Constants:    []objects.Object{&objects.Int{Value: 42}},
		Instructions: []byte{0x01},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.xxlbc")

	err := bc.SerializeToFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc2, err := DeserializeFromFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	intObj, ok := bc2.Constants[0].(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", bc2.Constants[0])
	}
	if intObj.Value != 42 {
		t.Fatalf("expected 42, got %d", intObj.Value)
	}
}

func TestDeserializeFromFile_NotFound(t *testing.T) {
	_, err := DeserializeFromFile("/nonexistent/path/file.xxlbc")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestObjectToSerializable_Builtin(t *testing.T) {
	obj := &objects.Builtin{}
	serial, err := objectToSerializable(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial.Type != "builtin" {
		t.Errorf("expected type 'builtin', got %q", serial.Type)
	}
}

func TestSerializableToObject_UnknownType(t *testing.T) {
	serial := serializableObject{Type: "unknown", Value: nil}
	_, err := serializableToObject(serial)
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestObjectToSerializable_Nil(t *testing.T) {
	serial, err := objectToSerializable(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial.Type != "null" {
		t.Errorf("expected type 'null', got %q", serial.Type)
	}
}

func TestObjectToSerializable_CompiledFunction(t *testing.T) {
	fn := &CompiledFunction{
		Instructions:  []byte{0x01, 0x02},
		NumLocals:     2,
		NumParameters: 1,
		FreeVariables: []Symbol{
			{Name: "x", Scope: GlobalScope, Index: 0},
		},
	}

	serial, err := objectToSerializable(fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial.Type != "compiled_function" {
		t.Errorf("expected type 'compiled_function', got %q", serial.Type)
	}
}

func TestObjectToSerializable_Unsupported(t *testing.T) {
	obj := &objects.Error{}
	_, err := objectToSerializable(obj)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestSerializableToObject_InvalidValueTypes(t *testing.T) {
	tests := []struct {
		name  string
		serial serializableObject
	}{
		{"invalid int", serializableObject{Type: "int", Value: "not an int"}},
		{"invalid float", serializableObject{Type: "float", Value: "not a float"}},
		{"invalid string", serializableObject{Type: "string", Value: 123}},
		{"invalid bool", serializableObject{Type: "bool", Value: "not a bool"}},
		{"invalid array", serializableObject{Type: "array", Value: "not an array"}},
		{"invalid map", serializableObject{Type: "map", Value: "not a map"}},
		{"invalid compiled_function", serializableObject{Type: "compiled_function", Value: 123}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := serializableToObject(tt.serial)
			if err == nil {
				t.Fatal("expected error for invalid value type")
			}
		})
	}
}

func TestDecodeCompiledFunctionFromMap_Basic(t *testing.T) {
	fn, err := decodeCompiledFunctionFromMap(map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fn.NumLocals != 0 {
		t.Errorf("expected 0 locals, got %d", fn.NumLocals)
	}
}

func TestDecodeCompiledFunctionFromMap_NumericTypes(t *testing.T) {
	data := map[string]any{"NumLocals": int(5)}
	fn, err := decodeCompiledFunctionFromMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fn.NumLocals != 5 {
		t.Errorf("expected 5 locals, got %d", fn.NumLocals)
	}

	data2 := map[string]any{"NumLocals": int64(10)}
	fn2, err := decodeCompiledFunctionFromMap(data2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fn2.NumLocals != 10 {
		t.Errorf("expected 10 locals, got %d", fn2.NumLocals)
	}

	data3 := map[string]any{"NumLocals": uint64(15)}
	fn3, err := decodeCompiledFunctionFromMap(data3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fn3.NumLocals != 15 {
		t.Errorf("expected 15 locals, got %d", fn3.NumLocals)
	}
}
