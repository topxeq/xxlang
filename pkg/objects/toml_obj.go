// pkg/objects/toml_obj.go
// TOML document object for Xxlang - pure Go implementation without CGO
package objects

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// TomlValueType represents the type of a TOML value
type TomlValueType int

const (
	TomlString TomlValueType = iota
	TomlInteger
	TomlFloat
	TomlBoolean
	TomlDatetime
	TomlDate
	TomlTime
	TomlArray
	TomlTable
	TomlNull
)

func (t TomlValueType) String() string {
	switch t {
	case TomlString:
		return "STRING"
	case TomlInteger:
		return "INTEGER"
	case TomlFloat:
		return "FLOAT"
	case TomlBoolean:
		return "BOOLEAN"
	case TomlDatetime:
		return "DATETIME"
	case TomlDate:
		return "DATE"
	case TomlTime:
		return "TIME"
	case TomlArray:
		return "ARRAY"
	case TomlTable:
		return "TABLE"
	default:
		return "NULL"
	}
}

// TomlValue represents a value in TOML document
type TomlValue struct {
	mu     sync.RWMutex
	Type   TomlValueType
	Value  interface{}
	parent *TomlValue
}

// TomlDocument represents a parsed TOML document
type TomlDocument struct {
	mu       sync.RWMutex
	root     *TomlValue
	modified bool
}

// NewTomlDocument creates a new empty TOML document
func NewTomlDocument() *TomlDocument {
	return &TomlDocument{
		root: &TomlValue{
			Type:  TomlTable,
			Value: make(map[string]*TomlValue),
		},
	}
}

// Type returns the object type
func (d *TomlDocument) Type() ObjectType { return TomlDocumentType }

// TypeTag returns the type tag
func (d *TomlDocument) TypeTag() TypeTag { return TypeTag(102) }

// Inspect returns a string representation
func (d *TomlDocument) Inspect() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.root == nil {
		return "TomlDocument(empty)"
	}
	keys := d.root.Keys()
	return fmt.Sprintf("TomlDocument(keys=%d)", len(keys))
}

// ToBool returns true
func (d *TomlDocument) ToBool() *Bool { return TRUE }

// HashKey returns a hash key
func (d *TomlDocument) HashKey() HashKey {
	return HashKey{Type: "TOML_DOCUMENT", Value: uint64(uintptr(unsafe.Pointer(d)))}
}

// Root returns the root table
func (d *TomlDocument) Root() *TomlValue {
	return d.root
}

// Get retrieves a value by path
// Path format: "section.key", "section.subsection.key", "array[0].key"
func (d *TomlDocument) Get(path string) *TomlValue {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getByPath(path)
}

// Set sets a value at the specified path
func (d *TomlDocument) Set(path string, value *TomlValue) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.modified = true
	return d.setByPath(path, value)
}

// Remove removes a value at the specified path
func (d *TomlDocument) Remove(path string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.modified = true
	return d.removeByPath(path)
}

// Has checks if a path exists
func (d *TomlDocument) Has(path string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getByPath(path) != nil
}

// Keys returns all top-level keys
func (d *TomlDocument) Keys() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.root.Keys()
}

// Sections returns all section names (top-level tables)
func (d *TomlDocument) Sections() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.root.Keys()
}

// ToMap converts the document to an Xxlang Map
func (d *TomlDocument) ToMap() *Map {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.root.ToMap()
}

// ToString serializes the document to a TOML string
func (d *TomlDocument) ToString() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return EncodeToml(d.root)
}

// ToIndented serializes with indentation
func (d *TomlDocument) ToIndented() string {
	return d.ToString()
}

// Save saves the document to a file
func (d *TomlDocument) Save(path string) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	content := d.ToString()
	return os.WriteFile(path, []byte(content), 0644)
}

// Merge merges another TOML document into this one
func (d *TomlDocument) Merge(other *TomlDocument) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.modified = true
	return mergeTomlTables(d.root, other.root)
}

// getByPath retrieves a value by path
func (d *TomlDocument) getByPath(path string) *TomlValue {
	if path == "" {
		return d.root
	}

	parts := parseTomlPath(path)
	if len(parts) == 0 {
		return d.root
	}

	current := d.root
	for _, part := range parts {
		if current == nil {
			return nil
		}

		switch current.Type {
		case TomlTable:
			table := current.Value.(map[string]*TomlValue)
			var ok bool
			current, ok = table[part.key]
			if !ok {
				return nil
			}
		case TomlArray:
			if part.index < 0 {
				return nil
			}
			arr := current.Value.([]*TomlValue)
			if part.index >= len(arr) {
				return nil
			}
			current = arr[part.index]
		default:
			return nil
		}
	}

	return current
}

// setByPath sets a value at the specified path
func (d *TomlDocument) setByPath(path string, value *TomlValue) error {
	parts := parseTomlPath(path)
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := d.root

	for i, part := range parts {
		isLast := i == len(parts)-1

		if isLast {
			// Set the value
			switch current.Type {
			case TomlTable:
				table := current.Value.(map[string]*TomlValue)
				table[part.key] = value
				value.parent = current
				return nil
			case TomlArray:
				arr := current.Value.([]*TomlValue)
				if part.index < 0 || part.index >= len(arr) {
					return fmt.Errorf("array index out of bounds: %d", part.index)
				}
				arr[part.index] = value
				value.parent = current
				return nil
			default:
				return fmt.Errorf("cannot set value on non-table/array type")
			}
		}

		// Navigate to next level
		switch current.Type {
		case TomlTable:
			table := current.Value.(map[string]*TomlValue)
			next, ok := table[part.key]
			if !ok {
				// Create intermediate table
				next = &TomlValue{
					Type:   TomlTable,
					Value:  make(map[string]*TomlValue),
					parent: current,
				}
				table[part.key] = next
			}
			current = next
		case TomlArray:
			arr := current.Value.([]*TomlValue)
			if part.index < 0 || part.index >= len(arr) {
				return fmt.Errorf("array index out of bounds: %d", part.index)
			}
			current = arr[part.index]
		default:
			return fmt.Errorf("cannot navigate through non-table/array type")
		}
	}

	return nil
}

// removeByPath removes a value at the specified path
func (d *TomlDocument) removeByPath(path string) bool {
	parts := parseTomlPath(path)
	if len(parts) == 0 {
		return false
	}

	current := d.root

	// Navigate to parent
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		switch current.Type {
		case TomlTable:
			table := current.Value.(map[string]*TomlValue)
			var ok bool
			current, ok = table[part.key]
			if !ok {
				return false
			}
		case TomlArray:
			arr := current.Value.([]*TomlValue)
			if part.index < 0 || part.index >= len(arr) {
				return false
			}
			current = arr[part.index]
		default:
			return false
		}
	}

	// Remove the key
	lastPart := parts[len(parts)-1]
	switch current.Type {
	case TomlTable:
		table := current.Value.(map[string]*TomlValue)
		if _, ok := table[lastPart.key]; ok {
			delete(table, lastPart.key)
			return true
		}
	}

	return false
}

// ============================================================
// TomlValue Methods
// ============================================================

// Keys returns keys of a table
func (v *TomlValue) Keys() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.Type != TomlTable {
		return nil
	}

	table := v.Value.(map[string]*TomlValue)
	keys := make([]string, 0, len(table))
	for k := range table {
		keys = append(keys, k)
	}
	return keys
}

// Get gets a value from a table by key
func (v *TomlValue) Get(key string) *TomlValue {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.Type != TomlTable {
		return nil
	}

	table := v.Value.(map[string]*TomlValue)
	return table[key]
}

// Set sets a value in a table
func (v *TomlValue) Set(key string, value *TomlValue) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.Type != TomlTable {
		return
	}

	table := v.Value.(map[string]*TomlValue)
	table[key] = value
	value.parent = v
}

// ToMap converts TomlValue to Xxlang Map
func (v *TomlValue) ToMap() *Map {
	v.mu.RLock()
	defer v.mu.RUnlock()

	switch v.Type {
	case TomlTable:
		table := v.Value.(map[string]*TomlValue)
		result := NewMapWithCapacity(len(table))
		for key, val := range table {
			keyObj := NewString(key)
			result.Pairs[keyObj.HashKey()] = MapPair{
				Key:   keyObj,
				Value: val.ToXxlangObject(),
			}
		}
		return result
	case TomlArray:
		arr := v.Value.([]*TomlValue)
		elements := make([]Object, len(arr))
		for i, val := range arr {
			elements[i] = val.ToXxlangObject()
		}
		return &Map{
			Pairs: map[HashKey]MapPair{
				NewString("type").HashKey(): {
					Key:   NewString("type"),
					Value: NewString("array"),
				},
				NewString("items").HashKey(): {
					Key:   NewString("items"),
					Value: NewArray(elements),
				},
			},
		}
	default:
		return NewMapWithCapacity(0)
	}
}

// ToXxlangObject converts TomlValue to Xxlang Object
func (v *TomlValue) ToXxlangObject() Object {
	switch v.Type {
	case TomlString:
		return NewString(v.Value.(string))
	case TomlInteger:
		return NewInt(v.Value.(int64))
	case TomlFloat:
		return NewFloat(v.Value.(float64))
	case TomlBoolean:
		if v.Value.(bool) {
			return TRUE
		}
		return FALSE
	case TomlDatetime:
		return NewString(v.Value.(time.Time).Format(time.RFC3339))
	case TomlDate:
		return NewString(v.Value.(time.Time).Format("2006-01-02"))
	case TomlTime:
		return NewString(v.Value.(time.Time).Format("15:04:05"))
	case TomlArray:
		arr := v.Value.([]*TomlValue)
		elements := make([]Object, len(arr))
		for i, val := range arr {
			elements[i] = val.ToXxlangObject()
		}
		return NewArray(elements)
	case TomlTable:
		return v.ToMap()
	default:
		return NULL
	}
}

// AsString converts to string
func (v *TomlValue) AsString() (string, bool) {
	if v.Type == TomlString {
		return v.Value.(string), true
	}
	return "", false
}

// AsInt converts to int
func (v *TomlValue) AsInt() (int64, bool) {
	if v.Type == TomlInteger {
		return v.Value.(int64), true
	}
	return 0, false
}

// AsFloat converts to float
func (v *TomlValue) AsFloat() (float64, bool) {
	if v.Type == TomlFloat {
		return v.Value.(float64), true
	}
	if v.Type == TomlInteger {
		return float64(v.Value.(int64)), true
	}
	return 0, false
}

// AsBool converts to bool
func (v *TomlValue) AsBool() (bool, bool) {
	if v.Type == TomlBoolean {
		return v.Value.(bool), true
	}
	return false, false
}

// AsArray converts to array
func (v *TomlValue) AsArray() ([]*TomlValue, bool) {
	if v.Type == TomlArray {
		return v.Value.([]*TomlValue), true
	}
	return nil, false
}

// AsTable converts to table
func (v *TomlValue) AsTable() (map[string]*TomlValue, bool) {
	if v.Type == TomlTable {
		return v.Value.(map[string]*TomlValue), true
	}
	return nil, false
}

// TypeName returns the type name
func (v *TomlValue) TypeName() string {
	return v.Type.String()
}

// ============================================================
// Path parsing
// ============================================================

type tomlPathPart struct {
	key   string
	index int
}

func parseTomlPath(path string) []tomlPathPart {
	var parts []tomlPathPart

	// Handle array index syntax: key[0]
	i := 0
	for i < len(path) {
		// Skip dots
		for i < len(path) && path[i] == '.' {
			i++
		}
		if i >= len(path) {
			break
		}

		// Parse key (handle quoted keys)
		var key string
		if path[i] == '"' || path[i] == '\'' {
			// Quoted key
			quote := path[i]
			i++
			start := i
			for i < len(path) && path[i] != quote {
				i++
			}
			key = path[start:i]
			i++ // Skip closing quote
		} else {
			// Unquoted key
			start := i
			for i < len(path) && path[i] != '.' && path[i] != '[' {
				i++
			}
			key = path[start:i]
		}

		parts = append(parts, tomlPathPart{key: key, index: -1})

		// Check for array index
		if i < len(path) && path[i] == '[' {
			i++ // Skip [
			start := i
			for i < len(path) && path[i] != ']' {
				i++
			}
			indexStr := path[start:i]
			index := 0
			for _, c := range indexStr {
				if c >= '0' && c <= '9' {
					index = index*10 + int(c-'0')
				}
			}
			if len(parts) > 0 {
				parts[len(parts)-1].index = index
			}
			i++ // Skip ]
		}
	}

	return parts
}

// ============================================================
// Helper functions
// ============================================================

func mergeTomlTables(dst, src *TomlValue) error {
	if dst.Type != TomlTable || src.Type != TomlTable {
		return fmt.Errorf("both values must be tables")
	}

	dstTable := dst.Value.(map[string]*TomlValue)
	srcTable := src.Value.(map[string]*TomlValue)

	for key, srcVal := range srcTable {
		dstVal, exists := dstTable[key]
		if !exists {
			// Copy the value
			dstTable[key] = copyTomlValue(srcVal)
		} else if dstVal.Type == TomlTable && srcVal.Type == TomlTable {
			// Recursively merge tables
			mergeTomlTables(dstVal, srcVal)
		} else {
			// Overwrite
			dstTable[key] = copyTomlValue(srcVal)
		}
	}

	return nil
}

func copyTomlValue(v *TomlValue) *TomlValue {
	if v == nil {
		return nil
	}

	switch v.Type {
	case TomlString, TomlInteger, TomlFloat, TomlBoolean, TomlDatetime, TomlDate, TomlTime:
		return &TomlValue{Type: v.Type, Value: v.Value}
	case TomlArray:
		arr := v.Value.([]*TomlValue)
		newArr := make([]*TomlValue, len(arr))
		for i, val := range arr {
			newArr[i] = copyTomlValue(val)
		}
		return &TomlValue{Type: TomlArray, Value: newArr}
	case TomlTable:
		table := v.Value.(map[string]*TomlValue)
		newTable := make(map[string]*TomlValue)
		for k, val := range table {
			newTable[k] = copyTomlValue(val)
		}
		return &TomlValue{Type: TomlTable, Value: newTable}
	default:
		return &TomlValue{Type: TomlNull}
	}
}

// FromXxlangObject converts Xxlang Object to TomlValue
func FromXxlangObject(obj Object) *TomlValue {
	switch v := obj.(type) {
	case *String:
		return &TomlValue{Type: TomlString, Value: v.Value}
	case *Int:
		return &TomlValue{Type: TomlInteger, Value: v.Value}
	case *Float:
		return &TomlValue{Type: TomlFloat, Value: v.Value}
	case *Bool:
		return &TomlValue{Type: TomlBoolean, Value: v.Value}
	case *Array:
		arr := make([]*TomlValue, len(v.Elements))
		for i, elem := range v.Elements {
			arr[i] = FromXxlangObject(elem)
		}
		return &TomlValue{Type: TomlArray, Value: arr}
	case *Map:
		table := make(map[string]*TomlValue)
		for _, pair := range v.Pairs {
			if key, ok := pair.Key.(*String); ok {
				table[key.Value] = FromXxlangObject(pair.Value)
			}
		}
		return &TomlValue{Type: TomlTable, Value: table}
	case *Null:
		return &TomlValue{Type: TomlNull}
	default:
		return &TomlValue{Type: TomlString, Value: v.Inspect()}
	}
}

// CreateTomlValueFromInterface creates TomlValue from Go interface{}
func CreateTomlValueFromInterface(v interface{}) *TomlValue {
	switch val := v.(type) {
	case string:
		return &TomlValue{Type: TomlString, Value: val}
	case int:
		return &TomlValue{Type: TomlInteger, Value: int64(val)}
	case int64:
		return &TomlValue{Type: TomlInteger, Value: val}
	case float64:
		return &TomlValue{Type: TomlFloat, Value: val}
	case bool:
		return &TomlValue{Type: TomlBoolean, Value: val}
	case time.Time:
		return &TomlValue{Type: TomlDatetime, Value: val}
	case []interface{}:
		arr := make([]*TomlValue, len(val))
		for i, elem := range val {
			arr[i] = CreateTomlValueFromInterface(elem)
		}
		return &TomlValue{Type: TomlArray, Value: arr}
	case map[string]interface{}:
		table := make(map[string]*TomlValue)
		for k, v := range val {
			table[k] = CreateTomlValueFromInterface(v)
		}
		return &TomlValue{Type: TomlTable, Value: table}
	case map[string]*TomlValue:
		return &TomlValue{Type: TomlTable, Value: val}
	case []*TomlValue:
		return &TomlValue{Type: TomlArray, Value: val}
	case nil:
		return &TomlValue{Type: TomlNull}
	default:
		return &TomlValue{Type: TomlString, Value: fmt.Sprintf("%v", val)}
	}
}

// String methods for debugging
func (v *TomlValue) String() string {
	return fmt.Sprintf("TomlValue(%s)", v.Type)
}

// Ensure key exists in string representation
func (v *TomlValue) Contains(key string) bool {
	if v.Type != TomlTable {
		return false
	}
	table := v.Value.(map[string]*TomlValue)
	_, ok := table[key]
	return ok
}

// Get nested value with multiple keys
func (v *TomlValue) GetPath(keys ...string) *TomlValue {
	current := v
	for _, key := range keys {
		if current == nil || current.Type != TomlTable {
			return nil
		}
		table := current.Value.(map[string]*TomlValue)
		current = table[key]
	}
	return current
}

// NewTomlValue creates a new TomlValue
func NewTomlValue(valType TomlValueType, value interface{}) *TomlValue {
	return &TomlValue{
		Type:  valType,
		Value: value,
	}
}

// IsTable checks if the value is a table
func (v *TomlValue) IsTable() bool {
	return v.Type == TomlTable
}

// IsArray checks if the value is an array
func (v *TomlValue) IsArray() bool {
	return v.Type == TomlArray
}

// IsScalar checks if the value is a scalar (non-container) type
func (v *TomlValue) IsScalar() bool {
	return v.Type != TomlArray && v.Type != TomlTable
}

// Len returns the length of an array or number of keys in a table
func (v *TomlValue) Len() int {
	switch v.Type {
	case TomlArray:
		return len(v.Value.([]*TomlValue))
	case TomlTable:
		return len(v.Value.(map[string]*TomlValue))
	default:
		return 0
	}
}

// KeysString returns keys as a comma-separated string
func (v *TomlValue) KeysString() string {
	keys := v.Keys()
	if keys == nil {
		return ""
	}
	return strings.Join(keys, ", ")
}
