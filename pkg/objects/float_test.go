// pkg/objects/float_test.go
package objects

import (
	"runtime"
	"sync"
	"testing"
)

// TestNewFloatCaching tests that cached values return the same pointer
func TestNewFloatCaching(t *testing.T) {
	// Test values within cache range
	cachedValues := []float64{0.0, 1.0, -1.0, 2.0, 0.5, 10.0, 100.0, 1000.0}
	for _, val := range cachedValues {
		obj1 := NewFloat(val)
		obj2 := NewFloat(val)
		if obj1 != obj2 {
			t.Errorf("NewFloat(%f) should return cached pointer, got different pointers", val)
		}
		if obj1.Value != val {
			t.Errorf("NewFloat(%f).Value = %f, want %f", val, obj1.Value, val)
		}
	}
}

// TestNewFloatPooling tests that values outside cache range use the pool
func TestNewFloatPooling(t *testing.T) {
	// Test values outside cache range
	val := 999999.123456
	obj1 := NewFloat(val)
	obj2 := NewFloat(val)

	// They should have the correct value
	if obj1.Value != val {
		t.Errorf("NewFloat(%f).Value = %f, want %f", val, obj1.Value, val)
	}
	if obj2.Value != val {
		t.Errorf("NewFloat(%f).Value = %f, want %f", val, obj2.Value, val)
	}

	// Release and verify pool reuse
	ReleaseFloat(obj1)
	obj3 := NewFloat(val + 1)
	_ = obj3
	// obj3 may or may not be the same pointer as obj1 depending on pool behavior
}

// TestReleaseFloat tests the release functionality
func TestReleaseFloat(t *testing.T) {
	// Releasing a cached value should do nothing
	cachedVal := 1.0
	obj1 := NewFloat(cachedVal)
	ReleaseFloat(obj1) // Should not panic or cause issues

	// Releasing a pooled value should return it to the pool
	pooledVal := 12345.6789
	obj2 := NewFloat(pooledVal)
	ReleaseFloat(obj2) // Should return to pool

	// Get stats to verify
	stats := GetFloatPoolStats()
	if stats.Released < 1 {
		t.Error("Expected at least 1 release in stats")
	}
}

// TestReleaseFloatSlice tests batch release
func TestReleaseFloatSlice(t *testing.T) {
	objs := []*Float{
		NewFloat(99999.123),
		NewFloat(88888.456),
		NewFloat(1.0), // This is cached, won't be pooled
	}

	ReleaseFloatSlice(objs)

	stats := GetFloatPoolStats()
	// At least 2 should be released (the pooled ones)
	if stats.Released < 2 {
		t.Errorf("Expected at least 2 releases, got %d", stats.Released)
	}
}

// TestFloatPoolStats tests statistics tracking
func TestFloatPoolStats(t *testing.T) {
	ResetFloatPoolStats()

	// Create some cached floats
	_ = NewFloat(0.0)
	_ = NewFloat(1.0)
	_ = NewFloat(2.0)

	stats := GetFloatPoolStats()
	if stats.CacheHits != 3 {
		t.Errorf("Expected 3 cache hits, got %d", stats.CacheHits)
	}

	// Create pooled floats
	pooled := NewFloat(99999.999)
	ReleaseFloat(pooled)

	stats = GetFloatPoolStats()
	if stats.Released != 1 {
		t.Errorf("Expected 1 release, got %d", stats.Released)
	}
}

// TestBufferedFloatPool tests the buffered pool operations
func TestBufferedFloatPool(t *testing.T) {
	// Get and put through buffer
	obj1 := GetBufferedFloat(12345.6789)
	if obj1.Value != 12345.6789 {
		t.Errorf("GetBufferedFloat(12345.6789).Value = %f, want 12345.6789", obj1.Value)
	}

	PutBufferedFloat(obj1)

	// Get another, might be reused
	obj2 := GetBufferedFloat(98765.4321)
	if obj2.Value != 98765.4321 {
		t.Errorf("GetBufferedFloat(98765.4321).Value = %f, want 98765.4321", obj2.Value)
	}
}

// TestConcurrentNewFloat tests concurrent access to the pool
func TestConcurrentNewFloat(t *testing.T) {
	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				// Mix of cached and pooled values
				val := float64(i+id*iterations) * 0.5
				if i%2 == 0 {
					val = float64(i % 10) // Likely cached
				}
				obj := NewFloat(val)
				if obj.Value != val {
					t.Errorf("NewFloat(%f).Value = %f", val, obj.Value)
				}
				// Release non-cached values
				if val != 0.0 && val != 1.0 && val != -1.0 && val != 2.0 &&
					val != 0.5 && val != 10.0 && val != 100.0 && val != 1000.0 {
					ReleaseFloat(obj)
				}
			}
		}(g)
	}

	wg.Wait()
}

// TestNewFloatSlice tests batch creation
func TestNewFloatSlice(t *testing.T) {
	values := []float64{0.0, 1.0, 999.999, -0.5, 10000.5}
	result := NewFloatSlice(values)

	if len(result) != len(values) {
		t.Fatalf("NewFloatSlice returned %d elements, want %d", len(result), len(values))
	}

	for i, obj := range result {
		if obj.Value != values[i] {
			t.Errorf("result[%d].Value = %f, want %f", i, obj.Value, values[i])
		}
	}
}

// BenchmarkNewFloatCached benchmarks creating cached floats
func BenchmarkNewFloatCached(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewFloat(float64(i % 10))
	}
}

// BenchmarkNewFloatPooled benchmarks creating pooled floats
func BenchmarkNewFloatPooled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		obj := NewFloat(float64(i + 10000))
		ReleaseFloat(obj)
	}
}

// BenchmarkNewFloatPooledNoRelease benchmarks pooled floats without release
func BenchmarkNewFloatPooledNoRelease(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewFloat(float64(i + 10000))
	}
}

// BenchmarkNewFloatDirect benchmarks direct struct creation (for comparison)
func BenchmarkNewFloatDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &Float{Value: float64(i)}
	}
}

// BenchmarkNewFloatSlice benchmarks batch creation
func BenchmarkNewFloatSlice(b *testing.B) {
	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i) * 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFloatSlice(values)
	}
}

// BenchmarkBufferedFloat benchmarks buffered pool access
func BenchmarkBufferedFloat(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := GetBufferedFloat(12345.6789)
			PutBufferedFloat(obj)
		}
	})
}

// BenchmarkFloatPoolContention tests pool performance under contention
func BenchmarkFloatPoolContention(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := NewFloat(99999.999)
			ReleaseFloat(obj)
		}
	})
}

// TestWarmFloatPool tests pool warming
func TestWarmFloatPool(t *testing.T) {
	// Warm the pool
	WarmFloatPool(100)

	// Force GC to clear any existing pooled objects
	runtime.GC()

	// The warm should have added objects
	// We can't directly verify this, but we can check that NewFloat still works
	obj := NewFloat(99999.999)
	if obj.Value != 99999.999 {
		t.Error("Pool warming affected NewFloat behavior")
	}
}

// TestCachedFloatConstants tests that cached float constants are correct
func TestCachedFloatConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant *Float
		expected float64
	}{
		{"FLOAT_ZERO", FLOAT_ZERO, 0.0},
		{"FLOAT_ONE", FLOAT_ONE, 1.0},
		{"FLOAT_NEG_ONE", FLOAT_NEG_ONE, -1.0},
		{"FLOAT_TWO", FLOAT_TWO, 2.0},
		{"FLOAT_HALF", FLOAT_HALF, 0.5},
		{"FLOAT_TEN", FLOAT_TEN, 10.0},
		{"FLOAT_HUNDRED", FLOAT_HUNDRED, 100.0},
		{"FLOAT_THOUSAND", FLOAT_THOUSAND, 1000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant.Value != tt.expected {
				t.Errorf("%s.Value = %f, want %f", tt.name, tt.constant.Value, tt.expected)
			}
			// Verify NewFloat returns the same pointer
			obj := NewFloat(tt.expected)
			if obj != tt.constant {
				t.Errorf("NewFloat(%f) != %s, expected same pointer", tt.expected, tt.name)
			}
		})
	}
}
