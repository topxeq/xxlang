// pkg/objects/sync.go
// Synchronization primitives for concurrent programming
package objects

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

// ============================================
// Mutex - Mutual Exclusion Lock
// ============================================

// Mutex wraps sync.Mutex for use in Xxlang
type Mutex struct {
	mu sync.Mutex
}

// NewMutex creates a new Mutex
func NewMutex() *Mutex {
	return &Mutex{}
}

// Lock acquires the mutex, blocking if necessary
func (m *Mutex) Lock() {
	m.mu.Lock()
}

// Unlock releases the mutex
func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

// TryLock attempts to acquire the mutex without blocking
// Returns true if successful, false otherwise
func (m *Mutex) TryLock() bool {
	return m.mu.TryLock()
}

// Object interface implementation
func (m *Mutex) Type() ObjectType { return MutexType }
func (m *Mutex) TypeTag() TypeTag  { return TagMutex }
func (m *Mutex) Inspect() string   { return "Mutex" }
func (m *Mutex) ToBool() *Bool     { return TRUE }
func (m *Mutex) HashKey() HashKey {
	return HashKey{Type: MutexType, Value: uint64(uintptr(unsafe.Pointer(m)))}
}

// ============================================
// RWMutex - Read/Write Mutual Exclusion Lock
// ============================================

// RWMutex wraps sync.RWMutex for use in Xxlang
type RWMutex struct {
	mu sync.RWMutex
}

// NewRWMutex creates a new RWMutex
func NewRWMutex() *RWMutex {
	return &RWMutex{}
}

// Lock acquires the write lock
func (m *RWMutex) Lock() {
	m.mu.Lock()
}

// Unlock releases the write lock
func (m *RWMutex) Unlock() {
	m.mu.Unlock()
}

// RLock acquires the read lock
func (m *RWMutex) RLock() {
	m.mu.RLock()
}

// RUnlock releases the read lock
func (m *RWMutex) RUnlock() {
	m.mu.RUnlock()
}

// TryLock attempts to acquire the write lock without blocking
func (m *RWMutex) TryLock() bool {
	return m.mu.TryLock()
}

// TryRLock attempts to acquire the read lock without blocking
func (m *RWMutex) TryRLock() bool {
	return m.mu.TryRLock()
}

// Object interface implementation
func (m *RWMutex) Type() ObjectType { return RWMutexType }
func (m *RWMutex) TypeTag() TypeTag  { return TagRWMutex }
func (m *RWMutex) Inspect() string   { return "RWMutex" }
func (m *RWMutex) ToBool() *Bool     { return TRUE }
func (m *RWMutex) HashKey() HashKey {
	return HashKey{Type: RWMutexType, Value: uint64(uintptr(unsafe.Pointer(m)))}
}

// ============================================
// WaitGroup - Wait for collection of goroutines
// ============================================

// WaitGroup wraps sync.WaitGroup for use in Xxlang
type WaitGroup struct {
	wg sync.WaitGroup
}

// NewWaitGroup creates a new WaitGroup
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{}
}

// Add increments the WaitGroup counter
func (w *WaitGroup) Add(delta int) {
	w.wg.Add(delta)
}

// Done decrements the WaitGroup counter
func (w *WaitGroup) Done() {
	w.wg.Done()
}

// Wait blocks until the WaitGroup counter is zero
func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

// Object interface implementation
func (w *WaitGroup) Type() ObjectType { return WaitGroupType }
func (w *WaitGroup) TypeTag() TypeTag  { return TagWaitGroup }
func (w *WaitGroup) Inspect() string   { return "WaitGroup" }
func (w *WaitGroup) ToBool() *Bool     { return TRUE }
func (w *WaitGroup) HashKey() HashKey {
	return HashKey{Type: WaitGroupType, Value: uint64(uintptr(unsafe.Pointer(w)))}
}

// ============================================
// Once - Ensure action is performed only once
// ============================================

// Once wraps sync.Once for use in Xxlang
type Once struct {
	once sync.Once
}

// NewOnce creates a new Once
func NewOnce() *Once {
	return &Once{}
}

// Do executes the function only once
func (o *Once) Do(fn func()) {
	o.once.Do(fn)
}

// Object interface implementation
func (o *Once) Type() ObjectType { return OnceType }
func (o *Once) TypeTag() TypeTag  { return TagOnce }
func (o *Once) Inspect() string   { return "Once" }
func (o *Once) ToBool() *Bool     { return TRUE }
func (o *Once) HashKey() HashKey {
	return HashKey{Type: OnceType, Value: uint64(uintptr(unsafe.Pointer(o)))}
}

// ============================================
// Cond - Condition Variable
// ============================================

// Cond wraps sync.Cond for use in Xxlang
type Cond struct {
	cond *sync.Cond
	mu   *Mutex // Reference to the underlying mutex for lifecycle
}

// NewCond creates a new Cond associated with the given Mutex
func NewCond(m *Mutex) *Cond {
	return &Cond{
		cond: sync.NewCond(&m.mu),
		mu:   m,
	}
}

// Wait atomically unlocks the mutex and suspends execution
func (c *Cond) Wait() {
	c.cond.Wait()
}

// Signal wakes one goroutine waiting on the condition
func (c *Cond) Signal() {
	c.cond.Signal()
}

// Broadcast wakes all goroutines waiting on the condition
func (c *Cond) Broadcast() {
	c.cond.Broadcast()
}

// Object interface implementation
func (c *Cond) Type() ObjectType { return CondType }
func (c *Cond) TypeTag() TypeTag  { return TagCond }
func (c *Cond) Inspect() string   { return "Cond" }
func (c *Cond) ToBool() *Bool     { return TRUE }
func (c *Cond) HashKey() HashKey {
	return HashKey{Type: CondType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}

// ============================================
// AtomicInt - Atomic Integer Operations
// ============================================

// AtomicInt provides atomic operations on an integer
type AtomicInt struct {
	value int64
}

// NewAtomicInt creates a new AtomicInt with the given initial value
func NewAtomicInt(initial int64) *AtomicInt {
	return &AtomicInt{value: initial}
}

// Add atomically adds delta to the value and returns the new value
func (a *AtomicInt) Add(delta int64) int64 {
	return atomic.AddInt64(&a.value, delta)
}

// Load atomically loads and returns the value
func (a *AtomicInt) Load() int64 {
	return atomic.LoadInt64(&a.value)
}

// Store atomically stores the value
func (a *AtomicInt) Store(val int64) {
	atomic.StoreInt64(&a.value, val)
}

// Swap atomically swaps the value and returns the old value
func (a *AtomicInt) Swap(new int64) int64 {
	return atomic.SwapInt64(&a.value, new)
}

// CompareAndSwap atomically compares and swaps
// Returns true if the swap was successful
func (a *AtomicInt) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&a.value, old, new)
}

// Object interface implementation
func (a *AtomicInt) Type() ObjectType { return AtomicIntType }
func (a *AtomicInt) TypeTag() TypeTag  { return TagAtomicInt }
func (a *AtomicInt) Inspect() string   { return fmt.Sprintf("AtomicInt(%d)", a.Load()) }
func (a *AtomicInt) ToBool() *Bool     { return &Bool{Value: a.Load() != 0} }
func (a *AtomicInt) HashKey() HashKey {
	return HashKey{Type: AtomicIntType, Value: uint64(a.Load())}
}

// ============================================
// Goroutine - Represents a running goroutine
// ============================================

// GoroutineStatus represents the state of a goroutine
type GoroutineStatus int

const (
	GoroutineRunning GoroutineStatus = iota
	GoroutineFinished
	GoroutinePanicked
)

// Goroutine represents a goroutine started with 'run'
type Goroutine struct {
	id         uint64
	status     GoroutineStatus
	started    int64 // Unix timestamp in nanoseconds
	waitChan   chan struct{}
	panicValue interface{}
}

var goroutineIDCounter uint64

// NewGoroutine creates a new Goroutine record
func NewGoroutine() *Goroutine {
	return &Goroutine{
		id:       atomic.AddUint64(&goroutineIDCounter, 1),
		status:   GoroutineRunning,
		waitChan: make(chan struct{}),
	}
}

// ID returns the goroutine's unique identifier
func (g *Goroutine) ID() uint64 {
	return g.id
}

// Status returns the current status of the goroutine
func (g *Goroutine) Status() GoroutineStatus {
	return g.status
}

// Wait blocks until the goroutine completes
func (g *Goroutine) Wait() {
	<-g.waitChan
}

// MarkFinished marks the goroutine as finished
func (g *Goroutine) MarkFinished() {
	g.status = GoroutineFinished
	close(g.waitChan)
}

// MarkPanicked marks the goroutine as panicked
func (g *Goroutine) MarkPanicked(v interface{}) {
	g.status = GoroutinePanicked
	g.panicValue = v
	close(g.waitChan)
}

// PanicValue returns the panic value if the goroutine panicked
func (g *Goroutine) PanicValue() interface{} {
	return g.panicValue
}

// Object interface implementation
func (g *Goroutine) Type() ObjectType { return GoroutineType }
func (g *Goroutine) TypeTag() TypeTag  { return TagGoroutine }
func (g *Goroutine) Inspect() string {
	var status string
	switch g.status {
	case GoroutineRunning:
		status = "running"
	case GoroutineFinished:
		status = "finished"
	case GoroutinePanicked:
		status = "panicked"
	}
	return fmt.Sprintf("Goroutine(id:%d, %s)", g.id, status)
}
func (g *Goroutine) ToBool() *Bool { return TRUE }
func (g *Goroutine) HashKey() HashKey {
	return HashKey{Type: GoroutineType, Value: g.id}
}
