// pkg/objects/orderedmap.go
// OrderedMap is a map that preserves insertion order and supports reordering.
package objects

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// OrderedMapPair represents a key-value pair in an OrderedMap
type OrderedMapPair struct {
	Key   Object
	Value Object
}

// OrderedMap represents a map that preserves insertion order of key-value pairs.
// It supports O(1) lookup via a hash map and ordered iteration via a slice.
type OrderedMap struct {
	Pairs      map[HashKey]int  // HashKey -> index in orderSlice
	orderSlice []OrderedMapPair // Ordered storage of pairs
	mu         sync.RWMutex     // Mutex for thread safety
}

// NewOrderedMap creates a new empty OrderedMap
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Pairs:      make(map[HashKey]int),
		orderSlice: make([]OrderedMapPair, 0),
	}
}

// NewOrderedMapWithCapacity creates a new OrderedMap with pre-allocated capacity
func NewOrderedMapWithCapacity(capacity int) *OrderedMap {
	return &OrderedMap{
		Pairs:      make(map[HashKey]int, capacity),
		orderSlice: make([]OrderedMapPair, 0, capacity),
	}
}

// Object interface implementation

// Type returns the object type
func (om *OrderedMap) Type() ObjectType { return OrderedMapType }

// TypeTag returns the type tag for fast type checking
func (om *OrderedMap) TypeTag() TypeTag { return TagOrderedMap }

// Inspect returns a string representation of the OrderedMap
func (om *OrderedMap) Inspect() string {
	om.mu.RLock()
	defer om.mu.RUnlock()

	visited := make(map[interface{}]struct{})
	return om.inspect(visited)
}

// inspect with cycle detection (caller must hold lock or be single-threaded)
func (om *OrderedMap) inspect(visited map[interface{}]struct{}) string {
	// Detect cycle
	if _, ok := visited[om]; ok {
		return "{...}"
	}
	visited[om] = struct{}{}

	if len(om.orderSlice) == 0 {
		return "{}"
	}

	var buf strings.Builder
	buf.WriteString("{")
	for i, pair := range om.orderSlice {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(pair.Key.Inspect())
		buf.WriteString(": ")
		// Use cycle-aware inspection for values
		switch v := pair.Value.(type) {
		case *Map:
			buf.WriteString(v.inspect(visited))
		case *Array:
			buf.WriteString(v.inspect(visited))
		case *OrderedMap:
			buf.WriteString(v.inspect(visited))
		default:
			buf.WriteString(pair.Value.Inspect())
		}
	}
	buf.WriteString("}")
	return buf.String()
}

// ToBool returns TRUE if the map has elements, FALSE otherwise
func (om *OrderedMap) ToBool() *Bool {
	om.mu.RLock()
	defer om.mu.RUnlock()
	if len(om.orderSlice) == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns an empty HashKey (OrderedMaps are not hashable)
func (om *OrderedMap) HashKey() HashKey {
	return HashKey{Type: OrderedMapType, Value: 0}
}

// CRUD Operations

// Get retrieves a value by key. Returns NULL if key not found.
func (om *OrderedMap) Get(key Object) Object {
	om.mu.RLock()
	defer om.mu.RUnlock()

	hashKey := key.HashKey()
	idx, ok := om.Pairs[hashKey]
	if !ok {
		return NULL
	}
	return om.orderSlice[idx].Value
}

// Set adds or updates a key-value pair.
// If the key exists, it updates the value but preserves the position.
// If the key is new, it appends to the end.
func (om *OrderedMap) Set(key Object, value Object) {
	om.mu.Lock()
	defer om.mu.Unlock()

	hashKey := key.HashKey()
	if idx, ok := om.Pairs[hashKey]; ok {
		// Key exists: update value, preserve position
		om.orderSlice[idx].Value = value
	} else {
		// New key: append to end
		om.Pairs[hashKey] = len(om.orderSlice)
		om.orderSlice = append(om.orderSlice, OrderedMapPair{Key: key, Value: value})
	}
}

// Delete removes a key-value pair. Returns true if the key was found and removed.
func (om *OrderedMap) Delete(key Object) bool {
	om.mu.Lock()
	defer om.mu.Unlock()

	hashKey := key.HashKey()
	idx, ok := om.Pairs[hashKey]
	if !ok {
		return false
	}

	// Remove from slice
	om.orderSlice = append(om.orderSlice[:idx], om.orderSlice[idx+1:]...)

	// Remove from map
	delete(om.Pairs, hashKey)

	// Update indices for elements after the deleted one
	for k, v := range om.Pairs {
		if v > idx {
			om.Pairs[k] = v - 1
		}
	}

	return true
}

// HasKey checks if a key exists
func (om *OrderedMap) HasKey(key Object) bool {
	om.mu.RLock()
	defer om.mu.RUnlock()

	_, ok := om.Pairs[key.HashKey()]
	return ok
}

// Len returns the number of key-value pairs
func (om *OrderedMap) Len() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return len(om.orderSlice)
}

// Ordered Access Methods

// GetOrderedKeys returns keys in insertion order
func (om *OrderedMap) GetOrderedKeys() []Object {
	om.mu.RLock()
	defer om.mu.RUnlock()

	keys := make([]Object, len(om.orderSlice))
	for i, pair := range om.orderSlice {
		keys[i] = pair.Key
	}
	return keys
}

// GetOrderedValues returns values in insertion order
func (om *OrderedMap) GetOrderedValues() []Object {
	om.mu.RLock()
	defer om.mu.RUnlock()

	values := make([]Object, len(om.orderSlice))
	for i, pair := range om.orderSlice {
		values[i] = pair.Value
	}
	return values
}

// GetOrderedPairs returns key-value pairs as array of [key, value] arrays
func (om *OrderedMap) GetOrderedPairs() []Object {
	om.mu.RLock()
	defer om.mu.RUnlock()

	pairs := make([]Object, len(om.orderSlice))
	for i, pair := range om.orderSlice {
		pairs[i] = &Array{Elements: []Object{pair.Key, pair.Value}}
	}
	return pairs
}

// Reordering Methods

// MoveToFront moves a key to the front (position 0). Returns error if key not found.
func (om *OrderedMap) MoveToFront(key Object) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	hashKey := key.HashKey()
	idx, ok := om.Pairs[hashKey]
	if !ok {
		return ErrKeyNotFound(key)
	}

	if idx == 0 {
		return nil // Already at front
	}

	// Move element to front
	pair := om.orderSlice[idx]
	om.orderSlice = append(om.orderSlice[:idx], om.orderSlice[idx+1:]...)
	om.orderSlice = append([]OrderedMapPair{pair}, om.orderSlice...)

	// Update indices
	om.updateIndicesAfterMove(idx, 0)
	return nil
}

// MoveToBack moves a key to the back (last position). Returns error if key not found.
func (om *OrderedMap) MoveToBack(key Object) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	hashKey := key.HashKey()
	idx, ok := om.Pairs[hashKey]
	if !ok {
		return ErrKeyNotFound(key)
	}

	lastIdx := len(om.orderSlice) - 1
	if idx == lastIdx {
		return nil // Already at back
	}

	// Move element to back
	pair := om.orderSlice[idx]
	om.orderSlice = append(om.orderSlice[:idx], om.orderSlice[idx+1:]...)
	om.orderSlice = append(om.orderSlice, pair)

	// Update indices
	om.updateIndicesAfterMove(idx, lastIdx)
	return nil
}

// MoveBefore moves key1 to the position immediately before key2.
// Returns error if either key is not found.
func (om *OrderedMap) MoveBefore(key1, key2 Object) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	hashKey1 := key1.HashKey()
	hashKey2 := key2.HashKey()

	idx1, ok1 := om.Pairs[hashKey1]
	idx2, ok2 := om.Pairs[hashKey2]

	if !ok1 {
		return ErrKeyNotFound(key1)
	}
	if !ok2 {
		return ErrKeyNotFound(key2)
	}

	if idx1 == idx2 {
		return nil // Same key
	}

	// Calculate new position
	newIdx := idx2
	if idx1 < idx2 {
		newIdx = idx2 - 1
	}

	if idx1 == newIdx {
		return nil // Already in position
	}

	om.moveElement(idx1, newIdx)
	return nil
}

// MoveAfter moves key1 to the position immediately after key2.
// Returns error if either key is not found.
func (om *OrderedMap) MoveAfter(key1, key2 Object) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	hashKey1 := key1.HashKey()
	hashKey2 := key2.HashKey()

	idx1, ok1 := om.Pairs[hashKey1]
	idx2, ok2 := om.Pairs[hashKey2]

	if !ok1 {
		return ErrKeyNotFound(key1)
	}
	if !ok2 {
		return ErrKeyNotFound(key2)
	}

	if idx1 == idx2 {
		return nil // Same key
	}

	// Calculate new position
	newIdx := idx2 + 1
	if idx1 < idx2 {
		newIdx = idx2
	}

	if idx1 == newIdx {
		return nil // Already in position
	}

	om.moveElement(idx1, newIdx)
	return nil
}

// Swap swaps the positions of two keys. Returns error if either key is not found.
func (om *OrderedMap) Swap(key1, key2 Object) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	hashKey1 := key1.HashKey()
	hashKey2 := key2.HashKey()

	idx1, ok1 := om.Pairs[hashKey1]
	idx2, ok2 := om.Pairs[hashKey2]

	if !ok1 {
		return ErrKeyNotFound(key1)
	}
	if !ok2 {
		return ErrKeyNotFound(key2)
	}

	if idx1 == idx2 {
		return nil // Same key
	}

	// Swap in slice
	om.orderSlice[idx1], om.orderSlice[idx2] = om.orderSlice[idx2], om.orderSlice[idx1]

	// Update indices in map
	om.Pairs[hashKey1] = idx2
	om.Pairs[hashKey2] = idx1

	return nil
}

// InsertAt inserts a key-value pair at a specific index.
// If the key already exists, it updates the value and moves it to the new position.
// Returns error if index is out of bounds.
func (om *OrderedMap) InsertAt(key Object, value Object, index int) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if index < 0 || index > len(om.orderSlice) {
		return ErrIndexOutOfRange(index, len(om.orderSlice))
	}

	hashKey := key.HashKey()

	if existingIdx, ok := om.Pairs[hashKey]; ok {
		// Key exists: update value and move
		om.orderSlice[existingIdx].Value = value

		if existingIdx == index {
			return nil
		}

		om.moveElement(existingIdx, index)
	} else {
		// New key: insert at position
		if index == len(om.orderSlice) {
			// Append to end
			om.Pairs[hashKey] = len(om.orderSlice)
			om.orderSlice = append(om.orderSlice, OrderedMapPair{Key: key, Value: value})
		} else {
			// Insert in middle
			om.orderSlice = append(om.orderSlice, OrderedMapPair{})
			copy(om.orderSlice[index+1:], om.orderSlice[index:])
			om.orderSlice[index] = OrderedMapPair{Key: key, Value: value}

			// Update all indices from index onwards
			for i := index; i < len(om.orderSlice); i++ {
				om.Pairs[om.orderSlice[i].Key.HashKey()] = i
			}
		}
	}

	return nil
}

// GetIndex returns the index of a key (0-based). Returns -1 if key not found.
func (om *OrderedMap) GetIndex(key Object) int {
	om.mu.RLock()
	defer om.mu.RUnlock()

	idx, ok := om.Pairs[key.HashKey()]
	if !ok {
		return -1
	}
	return idx
}

// GetAt returns the key-value pair at a specific index.
// Returns nil, nil, error if index is out of bounds.
func (om *OrderedMap) GetAt(index int) (Object, Object, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if index < 0 || index >= len(om.orderSlice) {
		return nil, nil, ErrIndexOutOfRange(index, len(om.orderSlice))
	}

	pair := om.orderSlice[index]
	return pair.Key, pair.Value, nil
}

// SetAt updates the value at a specific index.
// Returns error if index is out of bounds.
func (om *OrderedMap) SetAt(index int, value Object) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if index < 0 || index >= len(om.orderSlice) {
		return ErrIndexOutOfRange(index, len(om.orderSlice))
	}

	om.orderSlice[index].Value = value
	return nil
}

// Reverse reverses the order of all elements
func (om *OrderedMap) Reverse() {
	om.mu.Lock()
	defer om.mu.Unlock()

	// Reverse the slice
	for i, j := 0, len(om.orderSlice)-1; i < j; i, j = i+1, j-1 {
		om.orderSlice[i], om.orderSlice[j] = om.orderSlice[j], om.orderSlice[i]
	}

	// Update all indices
	for i, pair := range om.orderSlice {
		om.Pairs[pair.Key.HashKey()] = i
	}
}

// SortByKey sorts the map by key (alphabetically by string representation)
func (om *OrderedMap) SortByKey() {
	om.mu.Lock()
	defer om.mu.Unlock()

	sort.Slice(om.orderSlice, func(i, j int) bool {
		return om.orderSlice[i].Key.Inspect() < om.orderSlice[j].Key.Inspect()
	})

	// Update all indices
	for i, pair := range om.orderSlice {
		om.Pairs[pair.Key.HashKey()] = i
	}
}

// ToMap converts the OrderedMap to a regular Map
func (om *OrderedMap) ToMap() *Map {
	om.mu.RLock()
	defer om.mu.RUnlock()

	pairs := make(map[HashKey]MapPair, len(om.orderSlice))
	for _, pair := range om.orderSlice {
		pairs[pair.Key.HashKey()] = MapPair{Key: pair.Key, Value: pair.Value}
	}

	return &Map{Pairs: pairs}
}

// Clone creates a deep copy of the OrderedMap
func (om *OrderedMap) Clone() *OrderedMap {
	om.mu.RLock()
	defer om.mu.RUnlock()

	newPairs := make(map[HashKey]int, len(om.Pairs))
	newSlice := make([]OrderedMapPair, len(om.orderSlice))

	for k, v := range om.Pairs {
		newPairs[k] = v
	}
	copy(newSlice, om.orderSlice)

	return &OrderedMap{
		Pairs:      newPairs,
		orderSlice: newSlice,
	}
}

// Helper methods

// moveElement moves an element from oldIdx to newIdx
func (om *OrderedMap) moveElement(oldIdx, newIdx int) {
	if oldIdx == newIdx {
		return
	}

	pair := om.orderSlice[oldIdx]

	if oldIdx < newIdx {
		// Moving forward: shift elements left
		copy(om.orderSlice[oldIdx:newIdx], om.orderSlice[oldIdx+1:newIdx+1])
		om.orderSlice[newIdx] = pair

		// Update indices for shifted elements
		for i := oldIdx; i <= newIdx; i++ {
			om.Pairs[om.orderSlice[i].Key.HashKey()] = i
		}
	} else {
		// Moving backward: shift elements right
		copy(om.orderSlice[newIdx+1:oldIdx+1], om.orderSlice[newIdx:oldIdx])
		om.orderSlice[newIdx] = pair

		// Update indices for shifted elements
		for i := newIdx; i <= oldIdx; i++ {
			om.Pairs[om.orderSlice[i].Key.HashKey()] = i
		}
	}
}

// updateIndicesAfterMove updates all indices after moving an element
func (om *OrderedMap) updateIndicesAfterMove(oldIdx, newIdx int) {
	// Rebuild all indices from the affected range
	start := oldIdx
	if newIdx < oldIdx {
		start = newIdx
	}
	end := oldIdx
	if newIdx > oldIdx {
		end = newIdx
	}

	for i := start; i <= end && i < len(om.orderSlice); i++ {
		om.Pairs[om.orderSlice[i].Key.HashKey()] = i
	}
}

// Error helpers

// ErrKeyNotFound creates an error for a missing key
func ErrKeyNotFound(key Object) error {
	return &orderedMapError{msg: "key not found: " + key.Inspect()}
}

// ErrIndexOutOfRange creates an error for an out of range index
func ErrIndexOutOfRange(index, length int) error {
	return &orderedMapError{msg: "index out of range: " + strconv.Itoa(index) + " (length: " + strconv.Itoa(length) + ")"}
}

type orderedMapError struct {
	msg string
}

func (e *orderedMapError) Error() string {
	return e.msg
}

// OrderedMapPool for memory efficiency
var orderedMapPool = sync.Pool{
	New: func() interface{} {
		return NewOrderedMap()
	},
}

// AcquireOrderedMap gets an OrderedMap from the pool
func AcquireOrderedMap() *OrderedMap {
	return orderedMapPool.Get().(*OrderedMap)
}

// ReleaseOrderedMap returns an OrderedMap to the pool
func ReleaseOrderedMap(om *OrderedMap) {
	if om == nil {
		return
	}
	// Clear the map
	om.Pairs = make(map[HashKey]int)
	om.orderSlice = om.orderSlice[:0]
	orderedMapPool.Put(om)
}
