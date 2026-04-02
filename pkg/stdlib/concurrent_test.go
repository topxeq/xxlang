package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callConcurrentFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("concurrent")
	if mod == nil {
		panic("concurrent module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestConcurrentMakeTube(t *testing.T) {
	result := callConcurrentFunc("makeTube")
	if _, ok := result.(*objects.Tube); !ok {
		t.Errorf("expected *objects.Tube, got %T", result)
	}

	result = callConcurrentFunc("makeTube", String("int"))
	if _, ok := result.(*objects.Tube); !ok {
		t.Errorf("expected *objects.Tube, got %T", result)
	}

	result = callConcurrentFunc("makeTube", objects.NewInt(10))
	if _, ok := result.(*objects.Tube); !ok {
		t.Errorf("expected *objects.Tube, got %T", result)
	}

	result = callConcurrentFunc("makeTube", String("int"), objects.NewInt(5))
	if _, ok := result.(*objects.Tube); !ok {
		t.Errorf("expected *objects.Tube, got %T", result)
	}
}

func TestConcurrentCloseTube(t *testing.T) {
	tube := callConcurrentFunc("makeTube")
	result := callConcurrentFunc("closeTube", tube)
	if result != objects.NULL {
		t.Errorf("expected NULL, got %T", result)
	}

	result = callConcurrentFunc("closeTube")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for missing argument")
	}

	result = callConcurrentFunc("closeTube", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for wrong type")
	}
}

func TestConcurrentTubeLen(t *testing.T) {
	tube := callConcurrentFunc("makeTube", objects.NewInt(10))
	result := callConcurrentFunc("tubeLen", tube)
	if _, ok := result.(*objects.Int); !ok {
		t.Errorf("expected *objects.Int, got %T", result)
	}

	result = callConcurrentFunc("tubeLen")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for missing argument")
	}
}

func TestConcurrentTubeCap(t *testing.T) {
	tube := callConcurrentFunc("makeTube", objects.NewInt(10))
	result := callConcurrentFunc("tubeCap", tube)
	if _, ok := result.(*objects.Int); !ok {
		t.Errorf("expected *objects.Int, got %T", result)
	}

	result = callConcurrentFunc("tubeCap")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for missing argument")
	}
}

func TestConcurrentTubeClosed(t *testing.T) {
	tube := callConcurrentFunc("makeTube")
	result := callConcurrentFunc("tubeClosed", tube)
	if _, ok := result.(*objects.Bool); !ok {
		t.Errorf("expected *objects.Bool, got %T", result)
	}

	callConcurrentFunc("closeTube", tube)
	result = callConcurrentFunc("tubeClosed", tube)
	if b, ok := result.(*objects.Bool); !ok || !b.Value {
		t.Errorf("expected true for closed tube")
	}
}

func TestConcurrentNewMutex(t *testing.T) {
	result := callConcurrentFunc("newMutex")
	if _, ok := result.(*objects.Mutex); !ok {
		t.Errorf("expected *objects.Mutex, got %T", result)
	}
}

func TestConcurrentNewRWMutex(t *testing.T) {
	result := callConcurrentFunc("newRWMutex")
	if _, ok := result.(*objects.RWMutex); !ok {
		t.Errorf("expected *objects.RWMutex, got %T", result)
	}
}

func TestConcurrentNewWaitGroup(t *testing.T) {
	result := callConcurrentFunc("newWaitGroup")
	if _, ok := result.(*objects.WaitGroup); !ok {
		t.Errorf("expected *objects.WaitGroup, got %T", result)
	}
}

func TestConcurrentNewOnce(t *testing.T) {
	result := callConcurrentFunc("newOnce")
	if _, ok := result.(*objects.Once); !ok {
		t.Errorf("expected *objects.Once, got %T", result)
	}
}

func TestConcurrentNewAtomic(t *testing.T) {
	result := callConcurrentFunc("newAtomic")
	if _, ok := result.(*objects.AtomicInt); !ok {
		t.Errorf("expected *objects.AtomicInt, got %T", result)
	}

	result = callConcurrentFunc("newAtomic", objects.NewInt(100))
	if atomic, ok := result.(*objects.AtomicInt); !ok {
		t.Errorf("expected *objects.AtomicInt, got %T", result)
	} else {
		if atomic.Load() != 100 {
			t.Errorf("expected 100, got %d", atomic.Load())
		}
	}
}

func TestConcurrentNewCond(t *testing.T) {
	mutex := callConcurrentFunc("newMutex")
	result := callConcurrentFunc("newCond", mutex)
	if _, ok := result.(*objects.Cond); !ok {
		t.Errorf("expected *objects.Cond, got %T", result)
	}

	result = callConcurrentFunc("newCond")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for missing argument")
	}

	result = callConcurrentFunc("newCond", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for wrong type")
	}
}

func TestConcurrentTubeSendRecv(t *testing.T) {
	tube := callConcurrentFunc("makeTube", objects.NewInt(10))

	result := callConcurrentFunc("tubeSend", tube, objects.NewInt(42))
	if _, ok := result.(*objects.Bool); !ok {
		t.Errorf("expected *objects.Bool, got %T", result)
	}

	result = callConcurrentFunc("tubeRecv", tube)
	if arr, ok := result.(*objects.Array); !ok {
		t.Errorf("expected *objects.Array, got %T", result)
	} else {
		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements, got %d", len(arr.Elements))
		}
	}

	result = callConcurrentFunc("tubeSend")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for missing arguments")
	}
}

func TestConcurrentTubeTrySendRecv(t *testing.T) {
	tube := callConcurrentFunc("makeTube", objects.NewInt(10))

	result := callConcurrentFunc("tubeTrySend", tube, objects.NewInt(42))
	if arr, ok := result.(*objects.Array); !ok {
		t.Errorf("expected *objects.Array, got %T", result)
	} else {
		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements, got %d", len(arr.Elements))
		}
	}

	result = callConcurrentFunc("tubeTryRecv", tube)
	if arr, ok := result.(*objects.Array); !ok {
		t.Errorf("expected *objects.Array, got %T", result)
	} else {
		if len(arr.Elements) != 3 {
			t.Errorf("expected 3 elements, got %d", len(arr.Elements))
		}
	}

	result = callConcurrentFunc("tubeTrySend")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for missing arguments")
	}

	result = callConcurrentFunc("tubeTryRecv")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("expected error for missing arguments")
	}
}
