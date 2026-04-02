// pkg/objects/sync_test.go
package objects

import (
	"testing"
)

func TestNewMutex(t *testing.T) {
	m := NewMutex()
	if m == nil {
		t.Fatal("expected mutex instance")
	}
}

func TestMutexLockUnlock(t *testing.T) {
	m := NewMutex()
	m.Lock()
	m.Unlock()
}

func TestMutexTryLock(t *testing.T) {
	m := NewMutex()
	if !m.TryLock() {
		t.Error("expected TryLock to succeed")
	}
	m.Unlock()

	m.Lock()
	if m.TryLock() {
		t.Error("expected TryLock to fail when locked")
	}
	m.Unlock()
}

func TestMutexInspect(t *testing.T) {
	m := NewMutex()
	if m.Inspect() != "Mutex" {
		t.Errorf("expected 'Mutex', got '%s'", m.Inspect())
	}
}

func TestNewRWMutex(t *testing.T) {
	m := NewRWMutex()
	if m == nil {
		t.Fatal("expected rwmutex instance")
	}
}

func TestRWMutexLockUnlock(t *testing.T) {
	m := NewRWMutex()
	m.Lock()
	m.Unlock()
}

func TestRWMutexRLockRUnlock(t *testing.T) {
	m := NewRWMutex()
	m.RLock()
	m.RUnlock()
}

func TestRWMutexTryLock(t *testing.T) {
	m := NewRWMutex()
	if !m.TryLock() {
		t.Error("expected TryLock to succeed")
	}
	m.Unlock()
}

func TestRWMutexTryRLock(t *testing.T) {
	m := NewRWMutex()
	if !m.TryRLock() {
		t.Error("expected TryRLock to succeed")
	}
	m.RUnlock()
}

func TestNewWaitGroup(t *testing.T) {
	w := NewWaitGroup()
	if w == nil {
		t.Fatal("expected waitgroup instance")
	}
}

func TestWaitGroupAddDone(t *testing.T) {
	w := NewWaitGroup()
	w.Add(1)
	w.Done()
	w.Wait()
}

func TestNewOnce(t *testing.T) {
	o := NewOnce()
	if o == nil {
		t.Fatal("expected once instance")
	}
}

func TestOnceDo(t *testing.T) {
	o := NewOnce()
	count := 0
	o.Do(func() { count++ })
	o.Do(func() { count++ })
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestNewAtomicInt(t *testing.T) {
	a := NewAtomicInt(10)
	if a == nil {
		t.Fatal("expected atomic int instance")
	}
	if a.Load() != 10 {
		t.Errorf("expected 10, got %d", a.Load())
	}
}

func TestAtomicIntAdd(t *testing.T) {
	a := NewAtomicInt(10)
	result := a.Add(5)
	if result != 15 {
		t.Errorf("expected 15, got %d", result)
	}
}

func TestAtomicIntStore(t *testing.T) {
	a := NewAtomicInt(10)
	a.Store(20)
	if a.Load() != 20 {
		t.Errorf("expected 20, got %d", a.Load())
	}
}

func TestAtomicIntSwap(t *testing.T) {
	a := NewAtomicInt(10)
	old := a.Swap(20)
	if old != 10 {
		t.Errorf("expected old value 10, got %d", old)
	}
	if a.Load() != 20 {
		t.Errorf("expected 20, got %d", a.Load())
	}
}

func TestAtomicIntCompareAndSwap(t *testing.T) {
	a := NewAtomicInt(10)
	if !a.CompareAndSwap(10, 20) {
		t.Error("expected CAS to succeed")
	}
	if a.Load() != 20 {
		t.Errorf("expected 20, got %d", a.Load())
	}
	if a.CompareAndSwap(10, 30) {
		t.Error("expected CAS to fail")
	}
}

func TestNewGoroutine(t *testing.T) {
	g := NewGoroutine()
	if g == nil {
		t.Fatal("expected goroutine instance")
	}
	if g.ID() == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestGoroutineStatus(t *testing.T) {
	g := NewGoroutine()
	if g.Status() != GoroutineRunning {
		t.Error("expected running status")
	}
}

func TestGoroutineMarkFinished(t *testing.T) {
	g := NewGoroutine()
	g.MarkFinished()
	if g.Status() != GoroutineFinished {
		t.Error("expected finished status")
	}
}

func TestGoroutineMarkPanicked(t *testing.T) {
	g := NewGoroutine()
	g.MarkPanicked("test panic")
	if g.Status() != GoroutinePanicked {
		t.Error("expected panicked status")
	}
	if g.PanicValue() != "test panic" {
		t.Error("expected panic value")
	}
}

func TestNewCond(t *testing.T) {
	m := NewMutex()
	c := NewCond(m)
	if c == nil {
		t.Fatal("expected cond instance")
	}
}
