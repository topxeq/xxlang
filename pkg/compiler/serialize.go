// pkg/compiler/serialize.go
// Bytecode serialization for saving and loading compiled code.
package compiler

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/topxeq/xxlang/pkg/objects"
)

// Magic header for bytecode files
const (
	MagicHeader    = "XXLB" // Xxlang Bytecode
	CurrentVersion = 1
)

// init registers all object types for gob encoding.
func init() {
	// Register all object types that can be serialized
	gob.Register(&objects.Int{})
	gob.Register(&objects.Float{})
	gob.Register(&objects.String{})
	gob.Register(&objects.Bool{})
	gob.Register(&objects.Null{})
	gob.Register(&objects.Array{})
	gob.Register(&objects.Map{})
	gob.Register(&CompiledFunction{})
	gob.Register(serializableObject{})
	gob.Register([]serializableObject{})
	gob.Register(serializableCompiledFunction{})
	gob.Register(serializableFreeVar{})
	gob.Register(map[string]serializableObject{})
}

// serializableObject wraps an object for gob serialization.
// We need this because gob doesn't handle interfaces well.
type serializableObject struct {
	Type  string
	Value interface{}
}

// serializableCompiledFunction holds the serialized form of CompiledFunction.
type serializableCompiledFunction struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
	FreeVariables []serializableFreeVar
}

// serializableFreeVar holds the serialized form of a free variable Symbol.
type serializableFreeVar struct {
	Name  string
	Scope string
	Index int
}

// Serialize encodes bytecode to a binary format.
// Returns the serialized bytes or an error.
func (b *Bytecode) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	// Write magic header
	if _, err := buf.WriteString(MagicHeader); err != nil {
		return nil, fmt.Errorf("failed to write magic header: %v", err)
	}

	// Write version
	version := make([]byte, 4)
	binary.BigEndian.PutUint32(version, CurrentVersion)
	if _, err := buf.Write(version); err != nil {
		return nil, fmt.Errorf("failed to write version: %v", err)
	}

	// Convert constants to serializable format
	serialConstants := make([]serializableObject, len(b.Constants))
	for i, obj := range b.Constants {
		serialConst, err := objectToSerializable(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize constant %d: %v", i, err)
		}
		serialConstants[i] = serialConst
	}

	// Create the serializable bytecode structure
	sb := struct {
		Constants    []serializableObject
		Instructions []byte
	}{
		Constants:    serialConstants,
		Instructions: b.Instructions,
	}

	// Encode with gob
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(sb); err != nil {
		return nil, fmt.Errorf("failed to encode bytecode: %v", err)
	}

	return buf.Bytes(), nil
}

// Deserialize decodes binary data to bytecode.
// Returns the bytecode or an error.
func Deserialize(data []byte) (*Bytecode, error) {
	buf := bytes.NewReader(data)

	// Read and verify magic header
	magic := make([]byte, 4)
	if _, err := buf.Read(magic); err != nil {
		return nil, fmt.Errorf("failed to read magic header: %v", err)
	}
	if string(magic) != MagicHeader {
		return nil, fmt.Errorf("invalid bytecode file: bad magic header")
	}

	// Read and verify version
	versionBytes := make([]byte, 4)
	if _, err := buf.Read(versionBytes); err != nil {
		return nil, fmt.Errorf("failed to read version: %v", err)
	}
	version := binary.BigEndian.Uint32(versionBytes)
	if version > CurrentVersion {
		return nil, fmt.Errorf("unsupported bytecode version %d (current: %d)", version, CurrentVersion)
	}

	// Decode the rest with gob
	sb := struct {
		Constants    []serializableObject
		Instructions []byte
	}{}

	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&sb); err != nil {
		return nil, fmt.Errorf("failed to decode bytecode: %v", err)
	}

	// Convert serializable constants back to objects
	constants := make([]objects.Object, len(sb.Constants))
	for i, serial := range sb.Constants {
		obj, err := serializableToObject(serial)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize constant %d: %v", i, err)
		}
		constants[i] = obj
	}

	return &Bytecode{
		Constants:    constants,
		Instructions: sb.Instructions,
	}, nil
}

// SerializeToFile writes bytecode to a file.
// The format is: magic(4) + version(4) + gob-encoded data
func (b *Bytecode) SerializeToFile(path string) error {
	data, err := b.Serialize()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// DeserializeFromFile reads bytecode from a file.
func DeserializeFromFile(path string) (*Bytecode, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read bytecode file: %v", err)
	}

	return Deserialize(data)
}

// objectToSerializable converts an object to a serializable format.
func objectToSerializable(obj objects.Object) (serializableObject, error) {
	if obj == nil {
		return serializableObject{Type: "null", Value: nil}, nil
	}

	switch o := obj.(type) {
	case *objects.Int:
		return serializableObject{Type: "int", Value: o.Value}, nil

	case *objects.Float:
		return serializableObject{Type: "float", Value: o.Value}, nil

	case *objects.String:
		return serializableObject{Type: "string", Value: o.Value}, nil

	case *objects.Bool:
		return serializableObject{Type: "bool", Value: o.Value}, nil

	case *objects.Null:
		return serializableObject{Type: "null", Value: nil}, nil

	case *objects.Array:
		// Convert array elements recursively
		elements := make([]serializableObject, len(o.Elements))
		for i, elem := range o.Elements {
			serialElem, err := objectToSerializable(elem)
			if err != nil {
				return serializableObject{}, err
			}
			elements[i] = serialElem
		}
		return serializableObject{Type: "array", Value: elements}, nil

	case *objects.Map:
		// Convert map pairs to string-keyed map
		pairs := make(map[string]serializableObject)
		for _, pair := range o.Pairs {
			keyStr := pair.Key.Inspect()
			serialValue, err := objectToSerializable(pair.Value)
			if err != nil {
				return serializableObject{}, err
			}
			pairs[keyStr] = serialValue
		}
		return serializableObject{Type: "map", Value: pairs}, nil

	case *objects.Builtin:
		// Builtins can't be serialized - store a placeholder
		// They need to be re-registered at load time
		return serializableObject{Type: "builtin", Value: nil}, nil

	case *CompiledFunction:
		// Convert free variables
		freeVars := make([]serializableFreeVar, len(o.FreeVariables))
		for i, fv := range o.FreeVariables {
			freeVars[i] = serializableFreeVar{
				Name:  fv.Name,
				Scope: string(fv.Scope),
				Index: fv.Index,
			}
		}
		return serializableObject{
			Type: "compiled_function",
			Value: serializableCompiledFunction{
				Instructions:  o.Instructions,
				NumLocals:     o.NumLocals,
				NumParameters: o.NumParameters,
				FreeVariables: freeVars,
			},
		}, nil

	default:
		return serializableObject{}, fmt.Errorf("cannot serialize object of type %T", obj)
	}
}

// serializableToObject converts a serializable object back to an object.
func serializableToObject(serial serializableObject) (objects.Object, error) {
	switch serial.Type {
	case "int":
		if v, ok := serial.Value.(int64); ok {
			return &objects.Int{Value: v}, nil
		}
		return nil, fmt.Errorf("invalid int value type: %T", serial.Value)

	case "float":
		if v, ok := serial.Value.(float64); ok {
			return &objects.Float{Value: v}, nil
		}
		return nil, fmt.Errorf("invalid float value type: %T", serial.Value)

	case "string":
		if v, ok := serial.Value.(string); ok {
			return &objects.String{Value: v}, nil
		}
		return nil, fmt.Errorf("invalid string value type: %T", serial.Value)

	case "bool":
		if v, ok := serial.Value.(bool); ok {
			if v {
				return objects.TRUE, nil
			}
			return objects.FALSE, nil
		}
		return nil, fmt.Errorf("invalid bool value type: %T", serial.Value)

	case "null":
		return objects.NULL, nil

	case "array":
		if elements, ok := serial.Value.([]serializableObject); ok {
			arr := &objects.Array{Elements: make([]objects.Object, len(elements))}
			for i, elem := range elements {
				obj, err := serializableToObject(elem)
				if err != nil {
					return nil, err
				}
				arr.Elements[i] = obj
			}
			return arr, nil
		}
		return nil, fmt.Errorf("invalid array value type: %T", serial.Value)

	case "map":
		if pairs, ok := serial.Value.(map[string]serializableObject); ok {
			m := &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
			for keyStr, serialValue := range pairs {
				value, err := serializableToObject(serialValue)
				if err != nil {
					return nil, err
				}
				key := &objects.String{Value: keyStr}
				m.Pairs[key.HashKey()] = objects.MapPair{
					Key:   key,
					Value: value,
				}
			}
			return m, nil
		}
		return nil, fmt.Errorf("invalid map value type: %T", serial.Value)

	case "builtin":
		// Builtins need to be re-registered at load time
		return objects.NULL, nil

	case "compiled_function":
		switch data := serial.Value.(type) {
		case serializableCompiledFunction:
			return decodeCompiledFunction(data)
		case map[string]interface{}:
			// gob sometimes decodes structs as maps
			return decodeCompiledFunctionFromMap(data)
		default:
			return nil, fmt.Errorf("invalid compiled_function value type: %T", serial.Value)
		}

	default:
		return nil, fmt.Errorf("unknown object type: %s", serial.Type)
	}
}

// decodeCompiledFunction decodes a compiled function from its serializable form.
func decodeCompiledFunction(data serializableCompiledFunction) (*CompiledFunction, error) {
	freeVars := make([]Symbol, len(data.FreeVariables))
	for i, fv := range data.FreeVariables {
		freeVars[i] = Symbol{
			Name:  fv.Name,
			Scope: SymbolScope(fv.Scope),
			Index: fv.Index,
		}
	}
	return &CompiledFunction{
		Instructions:  data.Instructions,
		NumLocals:     data.NumLocals,
		NumParameters: data.NumParameters,
		FreeVariables: freeVars,
	}, nil
}

// decodeCompiledFunctionFromMap decodes a compiled function from gob's map representation.
func decodeCompiledFunctionFromMap(data map[string]interface{}) (*CompiledFunction, error) {
	fn := &CompiledFunction{}

	if v, ok := data["Instructions"]; ok {
		if instr, ok := v.([]byte); ok {
			fn.Instructions = instr
		}
	}
	if v, ok := data["NumLocals"]; ok {
		switch n := v.(type) {
		case int:
			fn.NumLocals = n
		case int64:
			fn.NumLocals = int(n)
		case uint64:
			fn.NumLocals = int(n)
		}
	}
	if v, ok := data["NumParameters"]; ok {
		switch n := v.(type) {
		case int:
			fn.NumParameters = n
		case int64:
			fn.NumParameters = int(n)
		case uint64:
			fn.NumParameters = int(n)
		}
	}
	if v, ok := data["FreeVariables"]; ok {
		if fvs, ok := v.([]serializableFreeVar); ok {
			fn.FreeVariables = make([]Symbol, len(fvs))
			for i, fv := range fvs {
				fn.FreeVariables[i] = Symbol{
					Name:  fv.Name,
					Scope: SymbolScope(fv.Scope),
					Index: fv.Index,
				}
			}
		}
	}

	return fn, nil
}
