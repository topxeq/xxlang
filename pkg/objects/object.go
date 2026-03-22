// pkg/objects/object.go
package objects

import "fmt"

// ObjectType represents the type of an object
type ObjectType string

// Object types
const (
	NullType         ObjectType = "NULL"
	IntType          ObjectType = "INT"
	FloatType        ObjectType = "FLOAT"
	StringType       ObjectType = "STRING"
	CharsType        ObjectType = "CHARS"
	BoolType         ObjectType = "BOOL"
	ArrayType        ObjectType = "ARRAY"
	MapType          ObjectType = "MAP"
	FunctionType     ObjectType = "FUNCTION"
	BuiltinType      ObjectType = "BUILTIN"
	BytesType        ObjectType = "BYTES"
	ClassType        ObjectType = "CLASS"
	InstanceType     ObjectType = "INSTANCE"
	ErrorType        ObjectType = "ERROR"
	ReturnType       ObjectType = "RETURN"
	ClosureType      ObjectType = "CLOSURE"
	ModuleType       ObjectType = "MODULE"
	StringBuilderType ObjectType = "STRING_BUILDER"
	BigIntType       ObjectType = "BIGINT"
	BigFloatType     ObjectType = "BIGFLOAT"
	HttpReqType      ObjectType = "HTTP_REQ"
	HttpRespType     ObjectType = "HTTP_RESP"
	HttpMuxType      ObjectType = "HTTP_MUX"
	WebSocketType    ObjectType = "WEBSOCKET"
	TubeType         ObjectType = "TUBE"
	MutexType        ObjectType = "MUTEX"
	RWMutexType      ObjectType = "RWMUTEX"
	WaitGroupType    ObjectType = "WAITGROUP"
	OnceType         ObjectType = "ONCE"
	CondType         ObjectType = "COND"
	AtomicIntType    ObjectType = "ATOMICINT"
	GoroutineType    ObjectType = "GOROUTINE"
)

// TypeTag is a fast integer type identifier for hot path checks
type TypeTag uint8

// Type tags for fast type checking (must match ObjectType order)
const (
	TagNull TypeTag = iota
	TagInt
	TagFloat
	TagString
	TagChars
	TagBool
	TagArray
	TagMap
	TagFunction
	TagBuiltin
	TagBytes
	TagClass
	TagInstance
	TagError
	TagReturn
	TagClosure
	TagModule
	TagStringBuilder
	TagBigInt
	TagBigFloat
	TagHttpReq
	TagHttpResp
	TagHttpMux
	TagWebSocket
	TagTube
	TagMutex
	TagRWMutex
	TagWaitGroup
	TagOnce
	TagCond
	TagAtomicInt
	TagGoroutine
	TagUnknown
)

// Object is the base interface for all values in Xxlang
type Object interface {
	Type() ObjectType
	TypeTag() TypeTag // Fast type check without string comparison
	Inspect() string
	ToBool() *Bool
	HashKey() HashKey
}

// HashKey is used for map keys
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Null represents the null value
type Null struct{}

func (n *Null) Type() ObjectType { return NullType }
func (n *Null) TypeTag() TypeTag { return TagNull }
func (n *Null) Inspect() string  { return "null" }
func (n *Null) ToBool() *Bool    { return FALSE }
func (n *Null) HashKey() HashKey { return HashKey{Type: NullType, Value: 0} }

// NULL is the singleton null value
var NULL = &Null{}

// Bool represents a boolean value
type Bool struct {
	Value bool
}

func (b *Bool) Type() ObjectType { return BoolType }
func (b *Bool) TypeTag() TypeTag { return TagBool }
func (b *Bool) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}
func (b *Bool) ToBool() *Bool { return b }
func (b *Bool) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	}
	return HashKey{Type: BoolType, Value: value}
}

// TRUE and FALSE are singleton boolean values
var (
	TRUE  = &Bool{Value: true}
	FALSE = &Bool{Value: false}
)

// Error represents a runtime error
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ErrorType }
func (e *Error) TypeTag() TypeTag { return TagError }
func (e *Error) Inspect() string  { return fmt.Sprintf("ERROR: %s", e.Message) }
func (e *Error) ToBool() *Bool    { return FALSE }
func (e *Error) HashKey() HashKey { return HashKey{Type: ErrorType, Value: 0} }

// Return represents a return value (used internally)
type Return struct {
	Value Object
}

func (r *Return) Type() ObjectType { return ReturnType }
func (r *Return) TypeTag() TypeTag { return TagReturn }
func (r *Return) Inspect() string  { return r.Value.Inspect() }
func (r *Return) ToBool() *Bool    { return r.Value.ToBool() }
func (r *Return) HashKey() HashKey { return HashKey{Type: ReturnType, Value: 0} }

// IsTruthy checks if an object is truthy
func IsTruthy(obj Object) bool {
	if obj == NULL {
		return false
	}
	return obj.ToBool().Value
}
