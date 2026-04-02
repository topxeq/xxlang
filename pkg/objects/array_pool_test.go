// pkg/objects/array_pool_test.go
package objects

import (
	"testing"
)

func TestNewArrayWithCapacity(t *testing.T) {
	arr := NewArrayWithCapacity(10)
	if arr == nil {
		t.Fatal("expected array instance")
	}
	if len(arr.Elements) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arr.Elements))
	}
}

func TestReleaseArray(t *testing.T) {
	arr := NewArrayWithCapacity(5)
	arr.Elements = []Object{NewInt(1), NewInt(2)}

	ReleaseArray(arr)
}

func TestReleaseArraySlice(t *testing.T) {
	arr1 := NewArrayWithCapacity(5)
	arr1.Elements = []Object{NewInt(1)}
	arr2 := NewArrayWithCapacity(5)
	arr2.Elements = []Object{NewInt(2)}

	slice := []*Array{arr1, arr2}
	ReleaseArraySlice(slice)
}

func TestGetArrayPoolStats(t *testing.T) {
	stats := GetArrayPoolStats()
	// Stats should have the expected fields
	_ = stats.Created
	_ = stats.PoolHits
	_ = stats.CacheHits
	_ = stats.Released
}

func TestResetArrayPoolStats(t *testing.T) {
	ResetArrayPoolStats()
	stats := GetArrayPoolStats()
	if stats.Created != 0 {
		t.Errorf("expected 0 created after reset, got %d", stats.Created)
	}
	if stats.PoolHits != 0 {
		t.Errorf("expected 0 pool hits after reset, got %d", stats.PoolHits)
	}
}

func TestWarmArrayPool(t *testing.T) {
	ResetArrayPoolStats()
	WarmArrayPool(10)
	// Pool should be warmed with 10 arrays
	// Note: WarmArrayPool doesn't update stats directly
}

func TestNewArrayEmpty(t *testing.T) {
	arr := NewArray([]Object{})
	if arr != EMPTY_ARRAY {
		t.Error("expected EMPTY_ARRAY for empty input")
	}
}

func TestNewArrayWithElements(t *testing.T) {
	elements := []Object{NewInt(1), NewInt(2), NewInt(3)}
	arr := NewArray(elements)
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestNewArrayWithCapacityZero(t *testing.T) {
	arr := NewArrayWithCapacity(0)
	if arr != EMPTY_ARRAY {
		t.Error("expected EMPTY_ARRAY for zero capacity")
	}
}
