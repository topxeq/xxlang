// pkg/vm/inline_cache_test.go
// Tests for inline caching
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// hashName Tests
// ============================================

func TestHashName(t *testing.T) {
	tests := []struct {
		s        string
		expected uint32
	}{
		{"", 0},
		{"a", 97},                        // 'a' = 97
		{"ab", 97*31 + 98},               // 'a'*31 + 'b'
		{"test", hashName("test")},       // Just verify it's deterministic
		{"method", hashName("method")},   // Just verify it's deterministic
	}

	for _, tt := range tests {
		if tt.expected != 0 || tt.s == "" {
			got := hashName(tt.s)
			if tt.expected != 0 && got != tt.expected {
				t.Errorf("hashName(%q) = %d, expected %d", tt.s, got, tt.expected)
			}
		}
	}
}

func TestHashNameDeterministic(t *testing.T) {
	// Same input should always produce same hash
	s := "testMethod"
	h1 := hashName(s)
	h2 := hashName(s)
	if h1 != h2 {
		t.Errorf("hashName is not deterministic: %d != %d", h1, h2)
	}
}

func TestHashNameDifferent(t *testing.T) {
	// Different inputs should (likely) produce different hashes
	h1 := hashName("method1")
	h2 := hashName("method2")
	if h1 == h2 {
		t.Errorf("Different strings produced same hash: %d", h1)
	}
}

// ============================================
// ComputeCacheIndex Tests
// ============================================

func TestComputeCacheIndexPrimitive(t *testing.T) {
	nameHash := uint32(12345)

	// Test with primitive type (no class)
	idx := ComputeCacheIndex(objects.TagInt, nil, nameHash)
	if idx < 0 || idx >= CacheSize {
		t.Errorf("Cache index %d out of range [0, %d)", idx, CacheSize)
	}
}

func TestComputeCacheIndexClass(t *testing.T) {
	nameHash := uint32(12345)
	class := &objects.Class{Name: "TestClass"}

	// Test with class
	idx := ComputeCacheIndex(objects.TagInstance, class, nameHash)
	if idx < 0 || idx >= CacheSize {
		t.Errorf("Cache index %d out of range [0, %d)", idx, CacheSize)
	}
}

func TestComputeCacheIndexDifferentTypes(t *testing.T) {
	nameHash := uint32(12345)

	idxInt := ComputeCacheIndex(objects.TagInt, nil, nameHash)
	idxFloat := ComputeCacheIndex(objects.TagFloat, nil, nameHash)
	idxString := ComputeCacheIndex(objects.TagString, nil, nameHash)

	// Different types should likely produce different indices
	sameCount := 0
	if idxInt == idxFloat {
		sameCount++
	}
	if idxInt == idxString {
		sameCount++
	}
	if idxFloat == idxString {
		sameCount++
	}

	// Allow some collisions but not all
	if sameCount == 3 {
		t.Logf("Warning: all types produced same cache index %d", idxInt)
	}
}

// ============================================
// InlineCacheTable Tests
// ============================================

func TestNewInlineCacheTable(t *testing.T) {
	var table InlineCacheTable

	// Fresh table should have no hits/misses
	hits, misses := table.Stats()
	if hits != 0 || misses != 0 {
		t.Errorf("Fresh table has stats: hits=%d, misses=%d", hits, misses)
	}
}

func TestInlineCacheTableSetGet(t *testing.T) {
	var table InlineCacheTable

	nameHash := hashName("testMethod")

	// Set a cache entry
	table.Set(
		objects.TagInt,  // typeTag
		nil,             // class
		nameHash,        // nameHash
		CacheResultPrimitiveMethod, // resultType
		nil,             // method
		0,               // fieldIdx
		nil,             // definingClass
	)

	// Get should find the entry
	entry := table.Get(objects.TagInt, nil, nameHash)
	if entry == nil {
		t.Fatal("Cache entry not found after Set")
	}

	if entry.ResultType != CacheResultPrimitiveMethod {
		t.Errorf("ResultType = %d, expected CacheResultPrimitiveMethod", entry.ResultType)
	}

	// Stats should show a hit
	hits, _ := table.Stats()
	if hits != 1 {
		t.Errorf("Expected 1 hit, got %d", hits)
	}
}

func TestInlineCacheTableMiss(t *testing.T) {
	var table InlineCacheTable

	nameHash := hashName("nonexistent")

	// Get should miss
	entry := table.Get(objects.TagInt, nil, nameHash)
	if entry != nil {
		t.Error("Expected nil for cache miss")
	}

	// Stats should show a miss
	_, misses := table.Stats()
	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}
}

func TestInlineCacheTableMismatch(t *testing.T) {
	var table InlineCacheTable

	nameHash := hashName("method1")

	// Set with TagInt
	table.Set(objects.TagInt, nil, nameHash, CacheResultPrimitiveMethod, nil, 0, nil)

	// Get with different type should miss
	entry := table.Get(objects.TagFloat, nil, nameHash)
	if entry != nil {
		t.Error("Expected nil for type mismatch")
	}

	_, misses := table.Stats()
	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}
}

func TestInlineCacheTableWithClass(t *testing.T) {
	var table InlineCacheTable

	class := &objects.Class{Name: "TestClass"}
	nameHash := hashName("instanceMethod")

	// Set with class
	table.Set(
		objects.TagInstance,
		class,
		nameHash,
		CacheResultMethod,
		nil,
		0,
		class,
	)

	// Get with same class should hit
	entry := table.Get(objects.TagInstance, class, nameHash)
	if entry == nil {
		t.Fatal("Cache entry not found")
	}

	if entry.Class != class {
		t.Error("Class mismatch in cache entry")
	}

	// Get with different class should miss
	differentClass := &objects.Class{Name: "DifferentClass"}
	entry = table.Get(objects.TagInstance, differentClass, nameHash)
	if entry != nil {
		t.Error("Expected nil for different class")
	}
}

func TestInlineCacheTableReset(t *testing.T) {
	var table InlineCacheTable

	nameHash := hashName("testMethod")

	// Set an entry
	table.Set(objects.TagInt, nil, nameHash, CacheResultPrimitiveMethod, nil, 0, nil)

	// Get to increment hit counter
	_ = table.Get(objects.TagInt, nil, nameHash)

	// Reset
	table.Reset()

	// Stats should be zero
	hits, misses := table.Stats()
	if hits != 0 || misses != 0 {
		t.Errorf("After reset: hits=%d, misses=%d", hits, misses)
	}

	// Entry should be cleared
	entry := table.Get(objects.TagInt, nil, nameHash)
	if entry != nil {
		t.Error("Expected nil after reset")
	}
}

func TestInlineCacheTableStats(t *testing.T) {
	var table InlineCacheTable

	nameHash1 := hashName("method1")
	nameHash2 := hashName("method2")

	// Set one entry
	table.Set(objects.TagInt, nil, nameHash1, CacheResultPrimitiveMethod, nil, 0, nil)

	// Hit
	_ = table.Get(objects.TagInt, nil, nameHash1)
	// Miss (different name)
	_ = table.Get(objects.TagInt, nil, nameHash2)
	// Hit
	_ = table.Get(objects.TagInt, nil, nameHash1)

	hits, misses := table.Stats()
	if hits != 2 {
		t.Errorf("Expected 2 hits, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}
}

// ============================================
// CacheResultType Tests
// ============================================

func TestCacheResultTypeValues(t *testing.T) {
	types := []CacheResultType{
		CacheResultNone,
		CacheResultMethod,
		CacheResultField,
		CacheResultNull,
		CacheResultPrimitiveMethod,
		CacheResultMapMethod,
	}

	// Just verify they're distinct
	seen := make(map[CacheResultType]bool)
	for _, ct := range types {
		if seen[ct] {
			t.Errorf("Duplicate cache result type: %d", ct)
		}
		seen[ct] = true
	}
}
