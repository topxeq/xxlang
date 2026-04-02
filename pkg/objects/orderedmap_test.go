// pkg/objects/orderedmap_test.go
package objects

import (
	"testing"
)

func TestNewOrderedMap(t *testing.T) {
	om := NewOrderedMap()
	if om == nil {
		t.Fatal("expected ordered map instance")
	}
}

func TestOrderedMapSetGet(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("key1"), NewInt(1))
	om.Set(NewString("key2"), NewInt(2))

	val := om.Get(NewString("key1"))
	if val == nil {
		t.Error("expected to find key1")
	}
	intVal, ok := val.(*Int)
	if !ok || intVal.Value != 1 {
		t.Errorf("expected 1, got %v", val)
	}
}

func TestOrderedMapDelete(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("key1"), NewInt(1))
	om.Delete(NewString("key1"))

	val := om.Get(NewString("key1"))
	if val != NULL {
		t.Error("expected NULL after delete")
	}
}

func TestOrderedMapLen(t *testing.T) {
	om := NewOrderedMap()
	if om.Len() != 0 {
		t.Error("expected empty map")
	}
	om.Set(NewString("key1"), NewInt(1))
	if om.Len() != 1 {
		t.Errorf("expected 1, got %d", om.Len())
	}
}

func TestOrderedMapGetOrderedKeys(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))
	om.Set(NewString("b"), NewInt(2))

	keys := om.GetOrderedKeys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestOrderedMapGetOrderedValues(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))
	om.Set(NewString("b"), NewInt(2))

	values := om.GetOrderedValues()
	if len(values) != 2 {
		t.Errorf("expected 2 values, got %d", len(values))
	}
}

func TestOrderedMapToMap(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))
	om.Set(NewString("b"), NewInt(2))

	m := om.ToMap()
	if len(m.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(m.Pairs))
	}
}

func TestOrderedMapHasKey(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))

	if !om.HasKey(NewString("a")) {
		t.Error("expected to have key 'a'")
	}
	if om.HasKey(NewString("b")) {
		t.Error("expected not to have key 'b'")
	}
}

func TestOrderedMapMoveToFront(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))
	om.Set(NewString("b"), NewInt(2))
	om.Set(NewString("c"), NewInt(3))

	err := om.MoveToFront(NewString("c"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	keys := om.GetOrderedKeys()
	if keys[0].(*String).Value != "c" {
		t.Error("expected 'c' at front")
	}
}

func TestOrderedMapMoveToBack(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))
	om.Set(NewString("b"), NewInt(2))

	err := om.MoveToBack(NewString("a"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	keys := om.GetOrderedKeys()
	if keys[1].(*String).Value != "a" {
		t.Error("expected 'a' at back")
	}
}

func TestOrderedMapSwap(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))
	om.Set(NewString("b"), NewInt(2))

	err := om.Swap(NewString("a"), NewString("b"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	keys := om.GetOrderedKeys()
	if keys[0].(*String).Value != "b" {
		t.Error("expected 'b' at index 0 after swap")
	}
}

func TestOrderedMapReverse(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))
	om.Set(NewString("b"), NewInt(2))
	om.Set(NewString("c"), NewInt(3))

	om.Reverse()

	keys := om.GetOrderedKeys()
	if keys[0].(*String).Value != "c" {
		t.Error("expected 'c' at index 0 after reverse")
	}
}

func TestOrderedMapSortByKey(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("c"), NewInt(3))
	om.Set(NewString("a"), NewInt(1))
	om.Set(NewString("b"), NewInt(2))

	om.SortByKey()

	keys := om.GetOrderedKeys()
	if keys[0].(*String).Value != "a" {
		t.Error("expected 'a' at index 0 after sort")
	}
}

func TestOrderedMapClone(t *testing.T) {
	om := NewOrderedMap()
	om.Set(NewString("a"), NewInt(1))

	cloned := om.Clone()
	if cloned.Len() != om.Len() {
		t.Error("expected same length")
	}
}

func TestAcquireOrderedMap(t *testing.T) {
	om := AcquireOrderedMap()
	if om == nil {
		t.Fatal("expected ordered map from pool")
	}
	ReleaseOrderedMap(om)
}

func TestNewOrderedMapWithCapacity(t *testing.T) {
	om := NewOrderedMapWithCapacity(10)
	if om == nil {
		t.Fatal("expected ordered map instance")
	}
}
