// pkg/objects/int_test.go
package objects

import (
	"runtime"
	"sync"
	"testing"
)

// TestNewIntCaching tests that cached values return the same pointer
func TestNewIntCaching(t *testing.T) {
	// Test values within cache range
	for _, val := range []int64{0, 1, 100, 1000, 10000, IntCacheMin, IntCacheMax} {
		obj1 := NewInt(val)
		obj2 := NewInt(val)
		if obj1 != obj2 {
			t.Errorf("NewInt(%d) should return cached pointer, got different pointers", val)
		}
		if obj1.Value != val {
			t.Errorf("NewInt(%d).Value = %d, want %d", val, obj1.Value, val)
		}
	}
}

// TestNewIntPooling tests that values outside cache range use the pool
func TestNewIntPooling(t *testing.T) {
	// Test values outside cache range
	val := int64(IntCacheMax + 1000)
	obj1 := NewInt(val)
	obj2 := NewInt(val)

	// They should have the correct value
	if obj1.Value != val {
		t.Errorf("NewInt(%d).Value = %d, want %d", val, obj1.Value, val)
	}
	if obj2.Value != val {
		t.Errorf("NewInt(%d).Value = %d, want %d", val, obj2.Value, val)
	}

	// Release and verify pool reuse
	ReleaseInt(obj1)
	obj3 := NewInt(val + 1)
	_ = obj3
	// obj3 may or may not be the same pointer as obj1 depending on pool behavior
}

// TestReleaseInt tests the release functionality
func TestReleaseInt(t *testing.T) {
	// Releasing a cached value should do nothing
	cachedVal := int64(100)
	obj1 := NewInt(cachedVal)
	ReleaseInt(obj1) // Should not panic or cause issues

	// Releasing a pooled value should return it to the pool
	pooledVal := int64(IntCacheMax + 5000)
	obj2 := NewInt(pooledVal)
	ReleaseInt(obj2) // Should return to pool

	// Get stats to verify
	stats := GetIntPoolStats()
	if stats.Released < 1 {
		t.Error("Expected at least 1 release in stats")
	}
}

// TestReleaseIntSlice tests batch release
func TestReleaseIntSlice(t *testing.T) {
	objs := []*Int{
		NewInt(IntCacheMax + 1000),
		NewInt(IntCacheMax + 2000),
		NewInt(100), // This is cached, won't be pooled
	}

	ReleaseIntSlice(objs)

	stats := GetIntPoolStats()
	// At least 2 should be released (the pooled ones)
	if stats.Released < 2 {
		t.Errorf("Expected at least 2 releases, got %d", stats.Released)
	}
}

// TestIsCachedInt tests the cache range check
func TestIsCachedInt(t *testing.T) {
	tests := []struct {
		val      int64
		expected bool
	}{
		{IntCacheMin, true},
		{IntCacheMax, true},
		{0, true},
		{100, true},
		{IntCacheMin - 1, false},
		{IntCacheMax + 1, false},
		{-10000, false},
		{1000000, false},
	}

	for _, tt := range tests {
		result := IsCachedInt(tt.val)
		if result != tt.expected {
			t.Errorf("IsCachedInt(%d) = %v, want %v", tt.val, result, tt.expected)
		}
	}
}

// TestNewIntSlice tests batch creation
func TestNewIntSlice(t *testing.T) {
	values := []int64{0, 100, IntCacheMax + 100, -50, IntCacheMin - 100}
	result := NewIntSlice(values)

	if len(result) != len(values) {
		t.Fatalf("NewIntSlice returned %d elements, want %d", len(result), len(values))
	}

	for i, obj := range result {
		if obj.Value != values[i] {
			t.Errorf("result[%d].Value = %d, want %d", i, obj.Value, values[i])
		}
	}
}

// TestIntPoolStats tests statistics tracking
func TestIntPoolStats(t *testing.T) {
	ResetIntPoolStats()

	// Create some cached integers
	_ = NewInt(0)
	_ = NewInt(100)
	_ = NewInt(1000)

	stats := GetIntPoolStats()
	if stats.CacheHits != 3 {
		t.Errorf("Expected 3 cache hits, got %d", stats.CacheHits)
	}

	// Create pooled integers
	pooled := NewInt(IntCacheMax + 10000)
	ReleaseInt(pooled)

	stats = GetIntPoolStats()
	if stats.Released != 1 {
		t.Errorf("Expected 1 release, got %d", stats.Released)
	}
}

// TestBufferedIntPool tests the buffered pool operations
func TestBufferedIntPool(t *testing.T) {
	// Get and put through buffer
	obj1 := GetBufferedInt(1000000)
	if obj1.Value != 1000000 {
		t.Errorf("GetBufferedInt(1000000).Value = %d, want 1000000", obj1.Value)
	}

	PutBufferedInt(obj1)

	// Get another, might be reused
	obj2 := GetBufferedInt(2000000)
	if obj2.Value != 2000000 {
		t.Errorf("GetBufferedInt(2000000).Value = %d, want 2000000", obj2.Value)
	}
}

// TestConcurrentNewInt tests concurrent access to the pool
func TestConcurrentNewInt(t *testing.T) {
	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				// Mix of cached and pooled values
				val := int64(i + id*iterations)
				if val%2 == 0 {
					val = val % 100000 // Likely cached
				}
				obj := NewInt(val)
				if obj.Value != val {
					t.Errorf("NewInt(%d).Value = %d", val, obj.Value)
				}
				// Release pooled values
				if !IsCachedInt(val) {
					ReleaseInt(obj)
				}
			}
		}(g)
	}

	wg.Wait()
}

// BenchmarkNewIntCached benchmarks creating cached integers
func BenchmarkNewIntCached(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewInt(int64(i % 100000))
	}
}

// BenchmarkNewIntPooled benchmarks creating pooled integers
func BenchmarkNewIntPooled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		obj := NewInt(int64(IntCacheMax + i + 1))
		ReleaseInt(obj)
	}
}

// BenchmarkNewIntPooledNoRelease benchmarks pooled integers without release
func BenchmarkNewIntPooledNoRelease(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewInt(int64(IntCacheMax + i + 1))
	}
}

// BenchmarkNewIntDirect benchmarks direct struct creation (for comparison)
func BenchmarkNewIntDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &Int{Value: int64(i)}
	}
}

// BenchmarkNewIntSlice benchmarks batch creation
func BenchmarkNewIntSlice(b *testing.B) {
	values := make([]int64, 100)
	for i := range values {
		values[i] = int64(i * 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewIntSlice(values)
	}
}

// BenchmarkBufferedInt benchmarks buffered pool access
func BenchmarkBufferedInt(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := GetBufferedInt(1000000)
			PutBufferedInt(obj)
		}
	})
}

// BenchmarkPoolContention tests pool performance under contention
func BenchmarkPoolContention(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := NewInt(int64(IntCacheMax + 1000000))
			ReleaseInt(obj)
		}
	})
}

// TestWarmIntPool tests pool warming
func TestWarmIntPool(t *testing.T) {
	// Warm the pool
	WarmIntPool(100)

	// Force GC to clear any existing pooled objects
	runtime.GC()

	// The warm should have added objects
	// We can't directly verify this, but we can check that NewInt still works
	obj := NewInt(IntCacheMax + 100000)
	if obj.Value != IntCacheMax+100000 {
		t.Error("Pool warming affected NewInt behavior")
	}
}
