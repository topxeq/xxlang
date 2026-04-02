//go:build amd64 && !windows
// +build amd64,!windows

// pkg/jit/jit_collections.go
// JIT collection operations with Go runtime callbacks
// This file implements array and map operations through callbacks to Go runtime
package jit

import (
	"reflect"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/objects"
)

// CollectionCodeGenerator generates x86-64 code for collection operations
// with proper callback support to Go runtime
type CollectionCodeGenerator struct {
	*SimpleCodeGenerator

	// Callback pointer for collection operations
	collectionCallbackPtr uintptr
}

// NewCollectionCodeGenerator creates a new code generator with collection support
func NewCollectionCodeGenerator() *CollectionCodeGenerator {
	return &CollectionCodeGenerator{
		SimpleCodeGenerator: NewSimpleCodeGenerator(),
	}
}

// SetCollectionCallback sets the callback pointer for collection operations
func (cg *CollectionCodeGenerator) SetCollectionCallback(ptr uintptr) {
	cg.collectionCallbackPtr = ptr
}

// compileArrayCreateWithCallback creates an array using Go runtime callback
// The callback returns an object handle that can be used for subsequent operations
func (cg *CollectionCodeGenerator) compileArrayCreateWithCallback(dst, startReg, count int) {
	// For each element, push it to the stack for the callback
	// Then call the collection callback with OpArrayCreate

	// Push elements onto stack (in reverse order for correct callback order)
	for i := count - 1; i >= 0; i-- {
		cg.emitMovSlotToRax(startReg + i)
		cg.emitByte(0x50) // push rax
	}

	// Prepare callback arguments
	// rdi = opKind (OpArrayCreate)
	// rsi = numArgs (count)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(OpArrayCreate))
	cg.emitBytes([]byte{0x48, 0xC7, 0xC6}) // mov rsi, imm32
	cg.emitUint32(uint32(count))

	// Call callback
	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		// No callback - return 0
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	}

	// Clean up stack (pop elements)
	for i := 0; i < count; i++ {
		cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x08}) // add rsp, 8
	}

	// Store result (object handle) in destination register
	cg.emitMovRaxToSlot(dst)
}

// compileArrayEmptyWithCallback creates an empty array using callback
func (cg *CollectionCodeGenerator) compileArrayEmptyWithCallback(dst int) {
	// rdi = opKind (OpArrayEmpty)
	// rsi = numArgs (0)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(OpArrayEmpty))
	cg.emitBytes([]byte{0x48, 0x31, 0xF6}) // xor rsi, rsi

	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	}

	cg.emitMovRaxToSlot(dst)
}

// compileArrayAppendWithCallback appends an element to an array
func (cg *CollectionCodeGenerator) compileArrayAppendWithCallback(dst, arrReg, elemReg int) {
	// Push array handle and element
	cg.emitMovSlotToRax(arrReg)
	cg.emitByte(0x50) // push rax (array handle)
	cg.emitMovSlotToRax(elemReg)
	cg.emitByte(0x50) // push rax (element)

	// rdi = opKind (OpArrayAppend)
	// rsi = numArgs (2)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(OpArrayAppend))
	cg.emitBytes([]byte{0x48, 0xC7, 0xC6, 0x02, 0x00, 0x00, 0x00}) // mov rsi, 2

	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		cg.emitMovSlotToRax(arrReg)
	}

	// Clean up stack
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x10}) // add rsp, 16

	cg.emitMovRaxToSlot(dst)
}

// compileIndexWithCallback gets an element from an array/map
func (cg *CollectionCodeGenerator) compileIndexWithCallback(dst, objReg, keyReg int) {
	// Push object handle and key
	cg.emitMovSlotToRax(objReg)
	cg.emitByte(0x50) // push rax (object handle)
	cg.emitMovSlotToRax(keyReg)
	cg.emitByte(0x50) // push rax (key)

	// rdi = opKind (OpArrayGet)
	// rsi = numArgs (2)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(OpArrayGet))
	cg.emitBytes([]byte{0x48, 0xC7, 0xC6, 0x02, 0x00, 0x00, 0x00}) // mov rsi, 2

	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	}

	// Clean up stack
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x10}) // add rsp, 16

	cg.emitMovRaxToSlot(dst)
}

// compileSetIndexWithCallback sets an element in an array/map
func (cg *CollectionCodeGenerator) compileSetIndexWithCallback(objReg, keyReg, valReg int) {
	// Push object handle, key, and value
	cg.emitMovSlotToRax(objReg)
	cg.emitByte(0x50) // push rax (object handle)
	cg.emitMovSlotToRax(keyReg)
	cg.emitByte(0x50) // push rax (key)
	cg.emitMovSlotToRax(valReg)
	cg.emitByte(0x50) // push rax (value)

	// rdi = opKind (OpArraySet)
	// rsi = numArgs (3)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(OpArraySet))
	cg.emitBytes([]byte{0x48, 0xC7, 0xC6, 0x03, 0x00, 0x00, 0x00}) // mov rsi, 3

	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	}

	// Clean up stack
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x18}) // add rsp, 24
}

// compileMapCreateWithCallback creates a map from key-value pairs
func (cg *CollectionCodeGenerator) compileMapCreateWithCallback(dst, startReg, pairCount int) {
	// Push pairs (key, value) onto stack in reverse order
	for i := pairCount - 1; i >= 0; i-- {
		// Push value first (so key is on top)
		cg.emitMovSlotToRax(startReg + i*2 + 1)
		cg.emitByte(0x50) // push rax
		// Push key
		cg.emitMovSlotToRax(startReg + i*2)
		cg.emitByte(0x50) // push rax
	}

	// rdi = opKind (OpMapCreate)
	// rsi = numArgs (pairCount * 2)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(OpMapCreate))
	cg.emitBytes([]byte{0x48, 0xC7, 0xC6}) // mov rsi, imm32
	cg.emitUint32(uint32(pairCount * 2))

	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	}

	// Clean up stack
	cg.emitBytes([]byte{0x48, 0x81, 0xC4}) // add rsp, imm32
	cg.emitUint32(uint32(pairCount * 2 * 8))

	cg.emitMovRaxToSlot(dst)
}

// compileMapEmptyWithCallback creates an empty map
func (cg *CollectionCodeGenerator) compileMapEmptyWithCallback(dst int) {
	// rdi = opKind (OpMapEmpty)
	// rsi = numArgs (0)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(OpMapEmpty))
	cg.emitBytes([]byte{0x48, 0x31, 0xF6}) // xor rsi, rsi

	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		cg.emitBytes([]byte{0x48, 0x31, 0xC0}) // xor rax, rax
	}

	cg.emitMovRaxToSlot(dst)
}

// compileMapSetWithCallback sets a key-value pair in a map
func (cg *CollectionCodeGenerator) compileMapSetWithCallback(dst, mapReg, keyReg, valReg int) {
	// Push map handle, key, and value
	cg.emitMovSlotToRax(mapReg)
	cg.emitByte(0x50) // push rax (map handle)
	cg.emitMovSlotToRax(keyReg)
	cg.emitByte(0x50) // push rax (key)
	cg.emitMovSlotToRax(valReg)
	cg.emitByte(0x50) // push rax (value)

	// rdi = opKind (OpMapSet)
	// rsi = numArgs (3)
	cg.emitBytes([]byte{0x48, 0xC7, 0xC7}) // mov rdi, imm32
	cg.emitUint32(uint32(OpMapSet))
	cg.emitBytes([]byte{0x48, 0xC7, 0xC6, 0x03, 0x00, 0x00, 0x00}) // mov rsi, 3

	if cg.collectionCallbackPtr != 0 {
		cg.emitBytes([]byte{0x48, 0xB8}) // mov rax, imm64
		cg.emitUint64(uint64(cg.collectionCallbackPtr))
		cg.emitBytes([]byte{0xFF, 0xD0}) // call rax
	} else {
		cg.emitMovSlotToRax(mapReg)
	}

	// Clean up stack
	cg.emitBytes([]byte{0x48, 0x83, 0xC4, 0x18}) // add rsp, 24

	cg.emitMovRaxToSlot(dst)
}

// ============================================================================
// JIT Collection Context
// ============================================================================

// JITCollectionContext maintains state for collection operations during JIT execution
type JITCollectionContext struct {
	objects     map[int64]objects.Object
	nextHandle  int64
	freeHandles []int64
}

// NewJITCollectionContext creates a new collection context
func NewJITCollectionContext() *JITCollectionContext {
	return &JITCollectionContext{
		objects:     make(map[int64]objects.Object),
		nextHandle:  1,
		freeHandles: make([]int64, 0),
	}
}

// AllocateObject allocates a handle for an object
func (ctx *JITCollectionContext) AllocateObject(obj objects.Object) int64 {
	if len(ctx.freeHandles) > 0 {
		handle := ctx.freeHandles[len(ctx.freeHandles)-1]
		ctx.freeHandles = ctx.freeHandles[:len(ctx.freeHandles)-1]
		ctx.objects[handle] = obj
		return handle
	}

	handle := ctx.nextHandle
	ctx.nextHandle++
	ctx.objects[handle] = obj
	return handle
}

// GetObject retrieves an object by handle
func (ctx *JITCollectionContext) GetObject(handle int64) (objects.Object, bool) {
	obj, ok := ctx.objects[handle]
	return obj, ok
}

// ReleaseObject releases a handle for reuse
func (ctx *JITCollectionContext) ReleaseObject(handle int64) {
	if _, ok := ctx.objects[handle]; ok {
		delete(ctx.objects, handle)
		ctx.freeHandles = append(ctx.freeHandles, handle)
	}
}

// Clear releases all objects
func (ctx *JITCollectionContext) Clear() {
	ctx.objects = make(map[int64]objects.Object)
	ctx.freeHandles = make([]int64, 0)
	ctx.nextHandle = 1
}

// ============================================================================
// Collection Callback Implementation (Alternative to native_executor.go)
// ============================================================================

// globalJITCollectionContext is the global context for collection operations
var globalJITCollectionContext = NewJITCollectionContext()

// GetGlobalJITCollectionContext returns the global collection context
func GetGlobalJITCollectionContext() *JITCollectionContext {
	return globalJITCollectionContext
}

// ResetGlobalJITCollectionContext resets the global collection context
func ResetGlobalJITCollectionContext() {
	globalJITCollectionContext.Clear()
}

// JITCollectionCallback implements the collection operation callback for JIT
// This is an alternative entry point that uses the JIT-specific context
func JITCollectionCallback(opKind, numArgs int, argsPtr *int64) int64 {
	args := make([]int64, numArgs)
	if numArgs > 0 && argsPtr != nil {
		argsSlice := unsafe.Slice(argsPtr, numArgs)
		copy(args, argsSlice)
	}

	ctx := globalJITCollectionContext

	switch CollectionOpKind(opKind) {
	case OpArrayCreate:
		elements := make([]objects.Object, numArgs)
		for i, v := range args {
			elements[i] = &objects.Int{Value: v}
		}
		arr := &objects.Array{Elements: elements}
		return ctx.AllocateObject(arr)

	case OpArrayEmpty:
		arr := &objects.Array{Elements: []objects.Object{}}
		return ctx.AllocateObject(arr)

	case OpArrayAppend:
		if numArgs < 2 {
			return 0
		}
		handle := args[0]
		elem := args[1]
		obj, ok := ctx.GetObject(handle)
		if !ok {
			return 0
		}
		arr, ok := obj.(*objects.Array)
		if !ok {
			return 0
		}
		arr.Elements = append(arr.Elements, &objects.Int{Value: elem})
		return handle

	case OpArrayGet:
		if numArgs < 2 {
			return 0
		}
		handle := args[0]
		index := int(args[1])
		obj, ok := ctx.GetObject(handle)
		if !ok {
			return 0
		}
		arr, ok := obj.(*objects.Array)
		if !ok || index < 0 || index >= len(arr.Elements) {
			return 0
		}
		if elem, ok := arr.Elements[index].(*objects.Int); ok {
			return elem.Value
		}
		return 0

	case OpArraySet:
		if numArgs < 3 {
			return 0
		}
		handle := args[0]
		index := int(args[1])
		value := args[2]
		obj, ok := ctx.GetObject(handle)
		if !ok {
			return 0
		}
		arr, ok := obj.(*objects.Array)
		if !ok || index < 0 || index >= len(arr.Elements) {
			return 0
		}
		arr.Elements[index] = &objects.Int{Value: value}
		return handle

	case OpMapCreate:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for i := 0; i+1 < numArgs; i += 2 {
			key := &objects.Int{Value: args[i]}
			val := &objects.Int{Value: args[i+1]}
			pairs[key.HashKey()] = objects.MapPair{Key: key, Value: val}
		}
		m := &objects.Map{Pairs: pairs}
		return ctx.AllocateObject(m)

	case OpMapEmpty:
		m := &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
		return ctx.AllocateObject(m)

	case OpMapSet:
		if numArgs < 3 {
			return 0
		}
		handle := args[0]
		key := args[1]
		value := args[2]
		obj, ok := ctx.GetObject(handle)
		if !ok {
			return 0
		}
		m, ok := obj.(*objects.Map)
		if !ok {
			return 0
		}
		k := &objects.Int{Value: key}
		v := &objects.Int{Value: value}
		m.Pairs[k.HashKey()] = objects.MapPair{Key: k, Value: v}
		return handle

	case OpMapGet:
		if numArgs < 2 {
			return 0
		}
		handle := args[0]
		key := args[1]
		obj, ok := ctx.GetObject(handle)
		if !ok {
			return 0
		}
		m, ok := obj.(*objects.Map)
		if !ok {
			return 0
		}
		k := &objects.Int{Value: key}
		if pair, ok := m.Pairs[k.HashKey()]; ok {
			if val, ok := pair.Value.(*objects.Int); ok {
				return val.Value
			}
		}
		return 0

	default:
		return 0
	}
}

// GetJITCollectionCallbackPtr returns the function pointer for JIT collection callbacks
func GetJITCollectionCallbackPtr() uintptr {
	return reflect.ValueOf(JITCollectionCallback).Pointer()
}
