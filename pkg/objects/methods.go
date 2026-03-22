// pkg/objects/methods.go
package objects

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// TypeMethods maps ObjectType -> methodName -> *Builtin
var TypeMethods = map[ObjectType]map[string]*Builtin{
	IntType:         intMethods,
	FloatType:       floatMethods,
	StringType:      stringMethods,
	ArrayType:       arrayMethods,
	MapType:         mapMethods,
	BoolType:        boolMethods,
	NullType:        nullMethods,
	StringBuilderType: stringBuilderMethods,
}

// GetMethod returns the builtin method for the given object type and method name
func GetMethod(objType ObjectType, name string) (*Builtin, bool) {
	methods, ok := TypeMethods[objType]
	if !ok {
		return nil, false
	}
	method, ok := methods[name]
	return method, ok
}

// ============================================================
// Universal Methods (available on all types)
// ============================================================

// universalTypeOf returns the type of any object
func universalTypeOf(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for typeOf. got=%d, want=1", len(args))
	}
	return NewString(string(args[0].Type()))
}

// universalToStr returns the string representation of any object
func universalToStr(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for toStr. got=%d, want=1", len(args))
	}
	return NewString(args[0].Inspect())
}

// ============================================================
// Int Methods
// ============================================================

var intMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"toFloat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toFloat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Int)
		if !ok {
			return newError("receiver for toFloat must be INT, got %s", args[0].Type())
		}
		return NewFloat(float64(self.Value))
	}},
	"abs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for abs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Int)
		if !ok {
			return newError("receiver for abs must be INT, got %s", args[0].Type())
		}
		if self.Value < 0 {
			return NewInt(-self.Value)
		}
		return self
	}},
}

// ============================================================
// Float Methods
// ============================================================

var floatMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"toInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for toInt must be FLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(self.Value))
	}},
	"abs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for abs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for abs must be FLOAT, got %s", args[0].Type())
		}
		if self.Value < 0 {
			return NewFloat(-self.Value)
		}
		return self
	}},
	"floor": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for floor. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for floor must be FLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(math.Floor(self.Value)))
	}},
	"ceil": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for ceil. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for ceil must be FLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(math.Ceil(self.Value)))
	}},
	"round": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for round. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Float)
		if !ok {
			return newError("receiver for round must be FLOAT, got %s", args[0].Type())
		}
		return NewInt(int64(math.Round(self.Value)))
	}},
}

// ============================================================
// String Methods
// ============================================================

var stringMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for len must be STRING, got %s", args[0].Type())
		}
		return NewInt(int64(len(self.Value)))
	}},
	"upper": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for upper. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for upper must be STRING, got %s", args[0].Type())
		}
		return NewString(strings.ToUpper(self.Value))
	}},
	"lower": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for lower. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for lower must be STRING, got %s", args[0].Type())
		}
		return NewString(strings.ToLower(self.Value))
	}},
	"trim": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for trim. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for trim must be STRING, got %s", args[0].Type())
		}
		return NewString(strings.TrimSpace(self.Value))
	}},
	"trimLeft": {Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newError("wrong number of arguments for trimLeft. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for trimLeft must be STRING, got %s", args[0].Type())
		}
		cutset := " \t\n\r"
		if len(args) == 2 {
			cs, ok := args[1].(*String)
			if !ok {
				return newError("argument for trimLeft must be STRING, got %s", args[1].Type())
			}
			cutset = cs.Value
		}
		return NewString(strings.TrimLeft(self.Value, cutset))
	}},
	"trimRight": {Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newError("wrong number of arguments for trimRight. got=%d, want=1 or 2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for trimRight must be STRING, got %s", args[0].Type())
		}
		cutset := " \t\n\r"
		if len(args) == 2 {
			cs, ok := args[1].(*String)
			if !ok {
				return newError("argument for trimRight must be STRING, got %s", args[1].Type())
			}
			cutset = cs.Value
		}
		return NewString(strings.TrimRight(self.Value, cutset))
	}},
	"split": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for split. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for split must be STRING, got %s", args[0].Type())
		}
		sep, ok := args[1].(*String)
		if !ok {
			return newError("separator for split must be STRING, got %s", args[1].Type())
		}
		parts := strings.Split(self.Value, sep.Value)
		elements := make([]Object, len(parts))
		for i, part := range parts {
			elements[i] = NewString(part)
		}
		return NewArray(elements)
	}},
	"contains": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for contains. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for contains must be STRING, got %s", args[0].Type())
		}
		substr, ok := args[1].(*String)
		if !ok {
			return newError("argument for contains must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: strings.Contains(self.Value, substr.Value)}
	}},
	"indexOf": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for indexOf. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for indexOf must be STRING, got %s", args[0].Type())
		}
		substr, ok := args[1].(*String)
		if !ok {
			return newError("argument for indexOf must be STRING, got %s", args[1].Type())
		}
		return NewInt(int64(strings.Index(self.Value, substr.Value)))
	}},
	"startsWith": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for startsWith. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for startsWith must be STRING, got %s", args[0].Type())
		}
		prefix, ok := args[1].(*String)
		if !ok {
			return newError("argument for startsWith must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: strings.HasPrefix(self.Value, prefix.Value)}
	}},
	"endsWith": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for endsWith. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for endsWith must be STRING, got %s", args[0].Type())
		}
		suffix, ok := args[1].(*String)
		if !ok {
			return newError("argument for endsWith must be STRING, got %s", args[1].Type())
		}
		return &Bool{Value: strings.HasSuffix(self.Value, suffix.Value)}
	}},
	"subStr": {Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("wrong number of arguments for subStr. got=%d, want=2 or 3", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for subStr must be STRING, got %s", args[0].Type())
		}
		start, ok := args[1].(*Int)
		if !ok {
			return newError("start index for subStr must be INT, got %s", args[1].Type())
		}
		// Convert to runes for proper Unicode handling
		runes := []rune(self.Value)
		runeLen := len(runes)
		startIdx := int(start.Value)
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > runeLen {
			startIdx = runeLen
		}
		if len(args) == 3 {
			end, ok := args[2].(*Int)
			if !ok {
				return newError("end index for subStr must be INT, got %s", args[2].Type())
			}
			endIdx := int(end.Value)
			if endIdx < startIdx {
				endIdx = startIdx
			}
			if endIdx > runeLen {
				endIdx = runeLen
			}
			return NewString(string(runes[startIdx:endIdx]))
		}
		return NewString(string(runes[startIdx:]))
	}},
	"toInt": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toInt. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for toInt must be STRING, got %s", args[0].Type())
		}
		val, err := strconv.ParseInt(self.Value, 10, 64)
		if err != nil {
			return newError("could not convert string to int: %s", self.Value)
		}
		return NewInt(val)
	}},
	"toFloat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toFloat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*String)
		if !ok {
			return newError("receiver for toFloat must be STRING, got %s", args[0].Type())
		}
		val, err := strconv.ParseFloat(self.Value, 64)
		if err != nil {
			return newError("could not convert string to float: %s", self.Value)
		}
		return NewFloat(val)
	}},
}

// ============================================================
// Array Methods
// ============================================================

var arrayMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for len must be ARRAY, got %s", args[0].Type())
		}
		return NewInt(int64(len(self.Elements)))
	}},
	"push": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for push. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for push must be ARRAY, got %s", args[0].Type())
		}
		newElements := make([]Object, len(self.Elements)+1)
		copy(newElements, self.Elements)
		newElements[len(self.Elements)] = args[1]
		return NewArray(newElements)
	}},
	"pop": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for pop. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for pop must be ARRAY, got %s", args[0].Type())
		}
		if len(self.Elements) == 0 {
			return newError("cannot pop from empty array")
		}
		lastElem := self.Elements[len(self.Elements)-1]
		newElements := make([]Object, len(self.Elements)-1)
		copy(newElements, self.Elements[:len(self.Elements)-1])
		result := NewArray(newElements)
		result.LastPopped = lastElem
		return result
	}},
	"first": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for first. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for first must be ARRAY, got %s", args[0].Type())
		}
		if len(self.Elements) == 0 {
			return NULL
		}
		return self.Elements[0]
	}},
	"last": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for last. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for last must be ARRAY, got %s", args[0].Type())
		}
		if len(self.Elements) == 0 {
			return NULL
		}
		return self.Elements[len(self.Elements)-1]
	}},
	"indexOf": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for indexOf. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for indexOf must be ARRAY, got %s", args[0].Type())
		}
		for i, elem := range self.Elements {
			if compareObjects(elem, args[1]) {
				return NewInt(int64(i))
			}
		}
		return NewInt(-1)
	}},
	"contains": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for contains. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for contains must be ARRAY, got %s", args[0].Type())
		}
		for _, elem := range self.Elements {
			if compareObjects(elem, args[1]) {
				return TRUE
			}
		}
		return FALSE
	}},
	"reverse": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for reverse. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for reverse must be ARRAY, got %s", args[0].Type())
		}
		if len(self.Elements) == 0 {
			return self
		}
		reversed := make([]Object, len(self.Elements))
		for i := 0; i < len(self.Elements); i++ {
			reversed[i] = self.Elements[len(self.Elements)-1-i]
		}
		return NewArray(reversed)
	}},
	"join": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for join. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for join must be ARRAY, got %s", args[0].Type())
		}
		sep, ok := args[1].(*String)
		if !ok {
			return newError("separator for join must be STRING, got %s", args[1].Type())
		}
		parts := make([]string, len(self.Elements))
		for i, elem := range self.Elements {
			if s, ok := elem.(*String); ok {
				parts[i] = s.Value
			} else {
				parts[i] = elem.Inspect()
			}
		}
		return NewString(strings.Join(parts, sep.Value))
	}},
	// sortByFunc sorts the array in-place using a custom comparator function.
	// The comparator function receives two indices (idx1, idx2) and returns true
	// if the element at idx1 should come before the element at idx2.
	// Returns the array itself (sorted in-place).
	"sortByFunc": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for sortByFunc. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Array)
		if !ok {
			return newError("receiver for sortByFunc must be ARRAY, got %s", args[0].Type())
		}

		if len(self.Elements) <= 1 {
			return self
		}

		// The comparator can be a Function, Closure, or Builtin
		comparator := args[1]

		// Sort using the comparator
		sort.Slice(self.Elements, func(i, j int) bool {
			// Call the comparator with two indices
			result, err := CallUserFunc(comparator, NewInt(int64(i)), NewInt(int64(j)))
			if err != nil {
				// If there's an error, maintain original order
				return false
			}
			// Convert result to boolean
			if b, ok := result.(*Bool); ok {
				return b.Value
			}
			// Non-boolean result: treat truthy values as true
			if result.Type() == NullType {
				return false
			}
			return true
		})

		return self
	}},
}

// ============================================================
// Map Methods
// ============================================================

var mapMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for len must be MAP, got %s", args[0].Type())
		}
		return NewInt(int64(len(self.Pairs)))
	}},
	"keys": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for keys. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for keys must be MAP, got %s", args[0].Type())
		}
		keys := make([]Object, len(self.Pairs))
		i := 0
		for _, pair := range self.Pairs {
			keys[i] = pair.Key
			i++
		}
		// Sort keys for deterministic output
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Inspect() < keys[j].Inspect()
		})
		return NewArray(keys)
	}},
	"values": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for values. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for values must be MAP, got %s", args[0].Type())
		}
		// Get keys and sort them for deterministic order
		keys := make([]Object, len(self.Pairs))
		i := 0
		for _, pair := range self.Pairs {
			keys[i] = pair.Key
			i++
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Inspect() < keys[j].Inspect()
		})
		// Get values in the same order as sorted keys
		vals := make([]Object, len(keys))
		for i, key := range keys {
			vals[i] = self.Pairs[key.HashKey()].Value
		}
		return NewArray(vals)
	}},
	"hasKey": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for hasKey. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for hasKey must be MAP, got %s", args[0].Type())
		}
		_, exists := self.Pairs[args[1].HashKey()]
		return &Bool{Value: exists}
	}},
	"delete": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for delete. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*Map)
		if !ok {
			return newError("receiver for delete must be MAP, got %s", args[0].Type())
		}
		newPairs := make(map[HashKey]MapPair, len(self.Pairs)-1)
		for k, v := range self.Pairs {
			if k != args[1].HashKey() {
				newPairs[k] = v
			}
		}
		return NewMap(newPairs)
	}},
}

// ============================================================
// Bool Methods
// ============================================================

var boolMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
}

// ============================================================
// Null Methods
// ============================================================

var nullMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
}

// ============================================================
// StringBuilder Methods
// ============================================================

var stringBuilderMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"len": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for len. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for len must be STRING_BUILDER, got %s", args[0].Type())
		}
		return NewInt(int64(self.Len()))
	}},
	"write": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for write. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for write must be STRING_BUILDER, got %s", args[0].Type())
		}
		str, ok := args[1].(*String)
		if !ok {
			return newError("argument for write must be STRING, got %s", args[1].Type())
		}
		n := self.Write(str.Value)
		return NewInt(int64(n))
	}},
	"writeLine": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for writeLine. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for writeLine must be STRING_BUILDER, got %s", args[0].Type())
		}
		str, ok := args[1].(*String)
		if !ok {
			return newError("argument for writeLine must be STRING, got %s", args[1].Type())
		}
		n := self.WriteLine(str.Value)
		return NewInt(int64(n))
	}},
	"toString": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toString. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for toString must be STRING_BUILDER, got %s", args[0].Type())
		}
		return NewString(self.String())
	}},
	"clear": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for clear. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for clear must be STRING_BUILDER, got %s", args[0].Type())
		}
		self.Clear()
		return NULL
	}},
	"reset": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for reset. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for reset must be STRING_BUILDER, got %s", args[0].Type())
		}
		self.Reset()
		return NULL
	}},
	"grow": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for grow. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for grow must be STRING_BUILDER, got %s", args[0].Type())
		}
		n, ok := args[1].(*Int)
		if !ok {
			return newError("argument for grow must be INT, got %s", args[1].Type())
		}
		self.Grow(int(n.Value))
		return NULL
	}},
	"isEmpty": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isEmpty. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*StringBuilder)
		if !ok {
			return newError("receiver for isEmpty must be STRING_BUILDER, got %s", args[0].Type())
		}
		return &Bool{Value: self.Len() == 0}
	}},
}
