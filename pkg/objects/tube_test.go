// pkg/objects/tube_test.go
package objects

import (
	"sync"
	"testing"
	"time"
)

// ============================================================
// Tube Type Tests
// ============================================================

func TestTubeType(t *testing.T) {
	tube := NewTube("", 0)

	if got := tube.Type(); got != TubeType {
		t.Errorf("Tube.Type() = %s, want TUBE", got)
	}

	if got := tube.TypeTag(); got != TagTube {
		t.Errorf("Tube.TypeTag() = %d, want %d", got, TagTube)
	}
}

func TestTubeInspect(t *testing.T) {
	// Untyped tube
	tube := NewTube("", 5)
	inspect := tube.Inspect()
	if inspect == "" {
		t.Error("Tube.Inspect() should not be empty")
	}

	// Typed tube
	typedTube := NewTube(IntType, 10)
	inspect = typedTube.Inspect()
	if inspect == "" {
		t.Error("Typed tube Inspect() should not be empty")
	}

	// Closed tube
	tube.Close()
	inspect = tube.Inspect()
	if inspect == "" {
		t.Error("Closed tube Inspect() should not be empty")
	}
}

func TestTubeToBool(t *testing.T) {
	tube := NewTube("", 0)
	if tube.ToBool() != TRUE {
		t.Error("Tube.ToBool() should be TRUE")
	}

	// Even closed tube is truthy
	tube.Close()
	if tube.ToBool() != TRUE {
		t.Error("Closed tube ToBool() should still be TRUE")
	}
}

func TestTubeHashKey(t *testing.T) {
	tube1 := NewTube("", 0)
	tube2 := NewTube("", 0)

	// Different tubes should have different hash keys
	if tube1.HashKey() == tube2.HashKey() {
		t.Error("Different tubes should have different hash keys")
	}

	// Same tube should return same hash key
	hk := tube1.HashKey()
	if tube1.HashKey() != hk {
		t.Error("Same tube should return same hash key")
	}
}

// ============================================================
// Tube Creation Tests
// ============================================================

func TestNewTubeUnbuffered(t *testing.T) {
	tube := NewTube("", 0)

	if tube.Cap() != 0 {
		t.Errorf("Unbuffered tube Cap() = %d, want 0", tube.Cap())
	}

	if tube.Len() != 0 {
		t.Errorf("Empty tube Len() = %d, want 0", tube.Len())
	}

	if tube.IsClosed() {
		t.Error("New tube should not be closed")
	}
}

func TestNewTubeBuffered(t *testing.T) {
	tube := NewTube("", 10)

	if tube.Cap() != 10 {
		t.Errorf("Buffered tube Cap() = %d, want 10", tube.Cap())
	}

	if tube.Len() != 0 {
		t.Errorf("Empty tube Len() = %d, want 0", tube.Len())
	}
}

func TestNewTubeTyped(t *testing.T) {
	intTube := NewTube(IntType, 5)
	if intTube.ElemType() != IntType {
		t.Errorf("Typed tube ElemType() = %s, want INT", intTube.ElemType())
	}

	stringTube := NewTube(StringType, 5)
	if stringTube.ElemType() != StringType {
		t.Errorf("Typed tube ElemType() = %s, want STRING", stringTube.ElemType())
	}
}

// ============================================================
// Send and Receive Tests
// ============================================================

func TestTubeSendReceiveBuffered(t *testing.T) {
	tube := NewTube("", 2)

	// Send should succeed on buffered tube
	val1 := &Int{Value: 42}
	if !tube.Send(val1) {
		t.Error("Send to buffered tube should succeed")
	}

	if tube.Len() != 1 {
		t.Errorf("After one send, Len() = %d, want 1", tube.Len())
	}

	val2 := &String{Value: "hello"}
	if !tube.Send(val2) {
		t.Error("Second send should succeed")
	}

	// Receive values
	recv1, ok1 := tube.Receive()
	if !ok1 {
		t.Error("Receive should succeed")
	}
	if recv1.(*Int).Value != 42 {
		t.Errorf("Received value = %d, want 42", recv1.(*Int).Value)
	}

	recv2, ok2 := tube.Receive()
	if !ok2 {
		t.Error("Second receive should succeed")
	}
	if recv2.(*String).Value != "hello" {
		t.Errorf("Received value = %s, want hello", recv2.(*String).Value)
	}
}

func TestTubeSendReceiveUnbuffered(t *testing.T) {
	tube := NewTube("", 0)

	var sendOk bool
	var wg sync.WaitGroup

	// Send in goroutine (will block until receive)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sendOk = tube.Send(&Int{Value: 100})
	}()

	// Small delay to ensure send is blocked
	time.Sleep(10 * time.Millisecond)

	// Receive should unblock the send
	recv, ok := tube.Receive()
	if !ok {
		t.Error("Receive should succeed")
	}
	if recv.(*Int).Value != 100 {
		t.Errorf("Received value = %d, want 100", recv.(*Int).Value)
	}

	wg.Wait()
	if !sendOk {
		t.Error("Send should have succeeded")
	}
}

func TestTubeSendOnClosed(t *testing.T) {
	tube := NewTube("", 1)
	tube.Close()

	ok := tube.Send(&Int{Value: 42})
	if ok {
		t.Error("Send on closed tube should return false")
	}
}

func TestTubeReceiveFromClosed(t *testing.T) {
	tube := NewTube("", 1)

	// Send a value then close
	tube.Send(&Int{Value: 42})
	tube.Close()

	// Should still be able to receive the value
	recv, ok := tube.Receive()
	if !ok {
		t.Error("Should receive value from closed tube")
	}
	if recv.(*Int).Value != 42 {
		t.Errorf("Received value = %d, want 42", recv.(*Int).Value)
	}

	// Next receive should fail (tube is empty and closed)
	_, ok = tube.Receive()
	if ok {
		t.Error("Receive from empty closed tube should fail")
	}
}

// ============================================================
// TrySend and TryReceive Tests
// ============================================================

func TestTubeTrySendSuccess(t *testing.T) {
	tube := NewTube("", 2)

	sent, ok := tube.TrySend(&Int{Value: 42})
	if !sent || !ok {
		t.Error("TrySend to buffered tube should succeed")
	}

	if tube.Len() != 1 {
		t.Errorf("After TrySend, Len() = %d, want 1", tube.Len())
	}
}

func TestTubeTrySendFull(t *testing.T) {
	tube := NewTube("", 1)

	// Fill the tube
	tube.Send(&Int{Value: 1})

	// TrySend should fail (not block)
	sent, ok := tube.TrySend(&Int{Value: 2})
	if sent {
		t.Error("TrySend to full tube should not send")
	}
	if !ok {
		t.Error("TrySend to full (but open) tube should return ok=true")
	}
}

func TestTubeTrySendClosed(t *testing.T) {
	tube := NewTube("", 1)
	tube.Close()

	sent, ok := tube.TrySend(&Int{Value: 42})
	if sent {
		t.Error("TrySend to closed tube should not send")
	}
	if ok {
		t.Error("TrySend to closed tube should return ok=false")
	}
}

func TestTubeTryReceiveSuccess(t *testing.T) {
	tube := NewTube("", 2)
	tube.Send(&Int{Value: 42})

	val, received, open := tube.TryReceive()
	if !received {
		t.Error("TryReceive should receive")
	}
	if !open {
		t.Error("Tube should still be open")
	}
	if val.(*Int).Value != 42 {
		t.Errorf("Received value = %d, want 42", val.(*Int).Value)
	}
}

func TestTubeTryReceiveEmpty(t *testing.T) {
	tube := NewTube("", 1)

	val, received, open := tube.TryReceive()
	if received {
		t.Error("TryReceive from empty tube should not receive")
	}
	if !open {
		t.Error("Empty (but open) tube should return open=true")
	}
	if val != NULL {
		t.Error("TryReceive from empty tube should return NULL")
	}
}

func TestTubeTryReceiveClosed(t *testing.T) {
	tube := NewTube("", 1)
	tube.Close()

	val, received, open := tube.TryReceive()
	if received {
		t.Error("TryReceive from empty closed tube should not receive")
	}
	if open {
		t.Error("Closed tube should return open=false")
	}
	_ = val
}

// ============================================================
// Close Tests
// ============================================================

func TestTubeClose(t *testing.T) {
	tube := NewTube("", 2)

	if tube.IsClosed() {
		t.Error("New tube should not be closed")
	}

	tube.Close()

	if !tube.IsClosed() {
		t.Error("Tube should be closed after Close()")
	}

	// Double close should be safe
	tube.Close() // Should not panic
}

func TestTubeCloseWithValues(t *testing.T) {
	tube := NewTube("", 2)
	tube.Send(&Int{Value: 1})
	tube.Send(&Int{Value: 2})

	tube.Close()

	// Should still be able to receive buffered values
	val1, ok1 := tube.Receive()
	if !ok1 || val1.(*Int).Value != 1 {
		t.Error("Should receive first value")
	}

	val2, ok2 := tube.Receive()
	if !ok2 || val2.(*Int).Value != 2 {
		t.Error("Should receive second value")
	}

	// After draining, receive should fail
	_, ok3 := tube.Receive()
	if ok3 {
		t.Error("Should not receive from drained closed tube")
	}
}

// ============================================================
// Len and Cap Tests
// ============================================================

func TestTubeLenCap(t *testing.T) {
	tube := NewTube("", 3)

	if tube.Cap() != 3 {
		t.Errorf("Cap() = %d, want 3", tube.Cap())
	}

	if tube.Len() != 0 {
		t.Errorf("Empty tube Len() = %d, want 0", tube.Len())
	}

	tube.Send(&Int{Value: 1})
	tube.Send(&Int{Value: 2})

	if tube.Len() != 2 {
		t.Errorf("After 2 sends, Len() = %d, want 2", tube.Len())
	}

	tube.Receive()

	if tube.Len() != 1 {
		t.Errorf("After 1 receive, Len() = %d, want 1", tube.Len())
	}
}

// ============================================================
// Tube Methods Tests
// ============================================================

func TestTubeMethods(t *testing.T) {
	tube := NewTube("", 2)

	// Test len method
	method, ok := GetMethod(TubeType, "len")
	if !ok {
		t.Fatal("len method not found")
	}
	result := method.Fn(tube)
	if result.(*Int).Value != 0 {
		t.Errorf("len method = %d, want 0", result.(*Int).Value)
	}

	// Test cap method
	method, ok = GetMethod(TubeType, "cap")
	if !ok {
		t.Fatal("cap method not found")
	}
	result = method.Fn(tube)
	if result.(*Int).Value != 2 {
		t.Errorf("cap method = %d, want 2", result.(*Int).Value)
	}

	// Test send method - returns TRUE on success
	method, ok = GetMethod(TubeType, "send")
	if !ok {
		t.Fatal("send method not found")
	}
	result = method.Fn(tube, &Int{Value: 42})
	if result != TRUE {
		t.Errorf("send method should return TRUE on success, got %v", result)
	}

	// Test receive method
	method, ok = GetMethod(TubeType, "receive")
	if !ok {
		t.Fatal("receive method not found")
	}
	result = method.Fn(tube)
	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("receive method should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("receive result should have 2 elements, got %d", len(arr.Elements))
	}

	// Test isClosed method
	method, ok = GetMethod(TubeType, "isClosed")
	if !ok {
		t.Fatal("isClosed method not found")
	}
	result = method.Fn(tube)
	if result != FALSE {
		t.Errorf("isClosed should be FALSE for open tube")
	}

	// Test close method
	method, ok = GetMethod(TubeType, "close")
	if !ok {
		t.Fatal("close method not found")
	}
	result = method.Fn(tube)
	if result != NULL {
		t.Errorf("close method should return NULL")
	}

	// Verify closed
	method, _ = GetMethod(TubeType, "isClosed")
	result = method.Fn(tube)
	if result != TRUE {
		t.Errorf("isClosed should be TRUE after close")
	}
}

func TestTubeTrySendMethod(t *testing.T) {
	tube := NewTube("", 1)

	method, ok := GetMethod(TubeType, "trySend")
	if !ok {
		t.Fatal("trySend method not found")
	}

	// First send should succeed
	result := method.Fn(tube, &Int{Value: 1})
	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("trySend should return Array, got %T", result)
	}
	if arr.Elements[0] != TRUE {
		t.Error("First trySend should succeed (sent=true)")
	}
	if arr.Elements[1] != TRUE {
		t.Error("Tube should be open (ok=true)")
	}

	// Now tube is full, trySend should fail (not block)
	result = method.Fn(tube, &Int{Value: 2})
	arr = result.(*Array)
	if arr.Elements[0] != FALSE {
		t.Error("trySend to full tube should return sent=false")
	}
	if arr.Elements[1] != TRUE {
		t.Error("Tube should still be open (ok=true)")
	}
}

func TestTubeTryReceiveMethod(t *testing.T) {
	tube := NewTube("", 1)
	tube.Send(&Int{Value: 42})

	method, ok := GetMethod(TubeType, "tryReceive")
	if !ok {
		t.Fatal("tryReceive method not found")
	}

	result := method.Fn(tube)
	arr, ok := result.(*Array)
	if !ok {
		t.Fatalf("tryReceive should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("tryReceive result should have 3 elements, got %d", len(arr.Elements))
	}

	// Elements: [value, received, open]
	if arr.Elements[1] != TRUE {
		t.Error("Should have received")
	}
	if arr.Elements[2] != TRUE {
		t.Error("Tube should be open")
	}
}

// ============================================================
// Builtin Functions Tests
// ============================================================

func TestBuiltinMakeTube(t *testing.T) {
	fn, ok := Builtins["makeTube"]
	if !ok {
		t.Fatal("makeTube builtin not found")
	}

	// With buffer only
	result := fn.Fn(&Int{Value: 10})
	tube, ok := result.(*Tube)
	if !ok {
		t.Fatalf("makeTube should return *Tube, got %T", result)
	}
	if tube.Cap() != 10 {
		t.Errorf("Tube capacity = %d, want 10", tube.Cap())
	}

	// With type and buffer
	result = fn.Fn(&String{Value: "INT"}, &Int{Value: 5})
	tube, ok = result.(*Tube)
	if !ok {
		t.Fatalf("makeTube should return *Tube, got %T", result)
	}
	if tube.ElemType() != IntType {
		t.Errorf("Tube elem type = %s, want INT", tube.ElemType())
	}

	// No arguments (unbuffered, untyped)
	result = fn.Fn()
	tube, ok = result.(*Tube)
	if !ok {
		t.Fatalf("makeTube should return *Tube, got %T", result)
	}
	if tube.Cap() != 0 {
		t.Errorf("Unbuffered tube Cap() = %d, want 0", tube.Cap())
	}
}

func TestBuiltinCloseTube(t *testing.T) {
	fn, ok := Builtins["closeTube"]
	if !ok {
		t.Fatal("closeTube builtin not found")
	}

	tube := NewTube("", 1)
	result := fn.Fn(tube)

	if result != NULL {
		t.Errorf("closeTube should return NULL, got %v", result)
	}

	if !tube.IsClosed() {
		t.Error("Tube should be closed")
	}

	// Test error case
	result = fn.Fn(&Int{Value: 42})
	if !isError(result) {
		t.Error("closeTube on non-tube should return error")
	}
}

func TestBuiltinTubeLen(t *testing.T) {
	fn, ok := Builtins["tubeLen"]
	if !ok {
		t.Fatal("tubeLen builtin not found")
	}

	tube := NewTube("", 2)
	tube.Send(&Int{Value: 1})

	result := fn.Fn(tube)
	lenVal, ok := result.(*Int)
	if !ok {
		t.Fatalf("tubeLen should return *Int, got %T", result)
	}
	if lenVal.Value != 1 {
		t.Errorf("tubeLen = %d, want 1", lenVal.Value)
	}
}

func TestBuiltinTubeCap(t *testing.T) {
	fn, ok := Builtins["tubeCap"]
	if !ok {
		t.Fatal("tubeCap builtin not found")
	}

	tube := NewTube("", 5)
	result := fn.Fn(tube)
	capVal, ok := result.(*Int)
	if !ok {
		t.Fatalf("tubeCap should return *Int, got %T", result)
	}
	if capVal.Value != 5 {
		t.Errorf("tubeCap = %d, want 5", capVal.Value)
	}
}

func TestBuiltinTubeClosed(t *testing.T) {
	fn, ok := Builtins["tubeClosed"]
	if !ok {
		t.Fatal("tubeClosed builtin not found")
	}

	tube := NewTube("", 1)

	// Open tube
	result := fn.Fn(tube)
	if result != FALSE {
		t.Errorf("Open tube should return FALSE, got %v", result)
	}

	// Closed tube
	tube.Close()
	result = fn.Fn(tube)
	if result != TRUE {
		t.Errorf("Closed tube should return TRUE, got %v", result)
	}
}

func TestBuiltinTubeSend(t *testing.T) {
	fn, ok := Builtins["tubeSend"]
	if !ok {
		t.Fatal("tubeSend builtin not found")
	}

	tube := NewTube("", 1)
	result := fn.Fn(tube, &Int{Value: 42})

	if result != TRUE {
		t.Errorf("tubeSend should return TRUE, got %v", result)
	}

	// Verify value was sent
	if tube.Len() != 1 {
		t.Errorf("After send, Len() = %d, want 1", tube.Len())
	}

	// Send to closed tube
	tube.Close()
	result = fn.Fn(tube, &Int{Value: 2})
	if result != FALSE {
		t.Errorf("tubeSend to closed tube should return FALSE")
	}
}

func TestBuiltinTubeRecv(t *testing.T) {
	fn, ok := Builtins["tubeRecv"]
	if !ok {
		t.Fatal("tubeRecv builtin not found")
	}

	tube := NewTube("", 1)
	tube.Send(&Int{Value: 42})

	result := fn.Fn(tube)
	m, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("tubeRecv should return OrderedMap, got %T", result)
	}

	// Check value
	val := m.Get(&String{Value: "value"})
	if val == nil || val.(*Int).Value != 42 {
		t.Error("Received value should be 42")
	}

	// Check ok flag
	okVal := m.Get(&String{Value: "ok"})
	if okVal != TRUE {
		t.Error("ok should be TRUE")
	}
}

func TestBuiltinTubeTrySend(t *testing.T) {
	fn, ok := Builtins["tubeTrySend"]
	if !ok {
		t.Fatal("tubeTrySend builtin not found")
	}

	tube := NewTube("", 1)
	result := fn.Fn(tube, &Int{Value: 42})

	m, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("tubeTrySend should return OrderedMap, got %T", result)
	}

	// {sent, ok}
	sent := m.Get(&String{Value: "sent"})
	if sent != TRUE {
		t.Error("Should have sent")
	}

	okVal := m.Get(&String{Value: "ok"})
	if okVal != TRUE {
		t.Error("Tube should be open")
	}

	// Tube is now full, trySend should fail
	result = fn.Fn(tube, &Int{Value: 3})
	m = result.(*OrderedMap)

	sent = m.Get(&String{Value: "sent"})
	if sent != FALSE {
		t.Error("Should not have sent (full)")
	}

	okVal = m.Get(&String{Value: "ok"})
	if okVal != TRUE {
		t.Error("Tube should still be open")
	}
}

func TestBuiltinTubeTryRecv(t *testing.T) {
	fn, ok := Builtins["tubeTryRecv"]
	if !ok {
		t.Fatal("tubeTryRecv builtin not found")
	}

	tube := NewTube("", 1)
	tube.Send(&Int{Value: 42})

	result := fn.Fn(tube)
	m, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("tubeTryRecv should return OrderedMap, got %T", result)
	}

	// {value, received, open}
	received := m.Get(&String{Value: "received"})
	if received != TRUE {
		t.Error("Should have received")
	}
	open := m.Get(&String{Value: "open"})
	if open != TRUE {
		t.Error("Tube should be open")
	}
}

// ============================================================
// Concurrency Tests
// ============================================================

func TestTubeConcurrentSendReceive(t *testing.T) {
	tube := NewTube("", 0) // Unbuffered

	const numValues = 100
	var wg sync.WaitGroup

	// Start receiver
	received := make([]int64, 0, numValues)
	var receivedMu sync.Mutex
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numValues; i++ {
			val, ok := tube.Receive()
			if ok {
				receivedMu.Lock()
				received = append(received, val.(*Int).Value)
				receivedMu.Unlock()
			}
		}
	}()

	// Send values
	for i := 0; i < numValues; i++ {
		tube.Send(&Int{Value: int64(i)})
	}

	wg.Wait()

	if len(received) != numValues {
		t.Errorf("Received %d values, want %d", len(received), numValues)
	}
}

func TestTubeConcurrentClose(t *testing.T) {
	tube := NewTube("", 10)

	var wg sync.WaitGroup

	// Multiple goroutines trying to close
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tube.Close() // Should be safe to call multiple times
		}()
	}

	wg.Wait()

	if !tube.IsClosed() {
		t.Error("Tube should be closed")
	}
}

// ============================================================
// Select-like Behavior Tests (using reflect.Select)
// ============================================================

func TestTubeReflectSelect(t *testing.T) {
	tube1 := NewTube("", 1)
	tube2 := NewTube("", 1)

	// Send to tube1
	tube1.Send(&Int{Value: 1})

	// Use reflect.Select to choose
	// This tests that ReflectValue() works correctly
	ch1 := tube1.ReflectValue()
	ch2 := tube2.ReflectValue()

	if !ch1.IsValid() || !ch2.IsValid() {
		t.Error("ReflectValue should return valid reflect.Value")
	}

	if ch1.Kind() != reflectChan {
		t.Errorf("ReflectValue kind = %v, want Chan", ch1.Kind())
	}

	_ = ch2
}

// reflectChan is the reflect.Kind for channels
const reflectChan = 18 // reflect.Chan

// ============================================================
// Edge Cases Tests
// ============================================================

func TestTubeSendNilValue(t *testing.T) {
	tube := NewTube("", 1)

	// Sending NULL should work
	ok := tube.Send(NULL)
	if !ok {
		t.Error("Send NULL should succeed")
	}

	val, ok := tube.Receive()
	if !ok {
		t.Error("Receive should succeed")
	}
	if val != NULL {
		t.Error("Should receive NULL")
	}
}

func TestTubeZeroBuffer(t *testing.T) {
	tube := NewTube("", 0)

	if tube.Cap() != 0 {
		t.Errorf("Zero buffer tube Cap() = %d, want 0", tube.Cap())
	}

	// Unbuffered tube should have len 0 when empty
	if tube.Len() != 0 {
		t.Errorf("Empty unbuffered tube Len() = %d, want 0", tube.Len())
	}
}

func TestTubeLargeBuffer(t *testing.T) {
	tube := NewTube("", 1000)

	if tube.Cap() != 1000 {
		t.Errorf("Large buffer tube Cap() = %d, want 1000", tube.Cap())
	}

	// Send many values
	for i := 0; i < 500; i++ {
		tube.Send(&Int{Value: int64(i)})
	}

	if tube.Len() != 500 {
		t.Errorf("After 500 sends, Len() = %d, want 500", tube.Len())
	}
}
