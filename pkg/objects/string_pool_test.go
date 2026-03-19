// pkg/objects/string_pool_test.go
package objects

import (
	"runtime"
	"sync"
	"testing"
)

// TestNewStringCaching tests that cached values return the same pointer
func TestNewStringCaching(t *testing.T) {
	// Test values within cache range
	cachedValues := []string{"", "true", "false", "null", "INT", "FLOAT", "BOOL",
		"STRING", "ARRAY", "MAP", "FUNCTION", "BUILTIN", "ERROR", "nil", "0", "1", " ", "\n"}
	for _, val := range cachedValues {
		obj1 := NewString(val)
		obj2 := NewString(val)
		if obj1 != obj2 {
			t.Errorf("NewString(%q) should return cached pointer, got different pointers", val)
		}
		if obj1.Value != val {
			t.Errorf("NewString(%q).Value = %q, want %q", val, obj1.Value, val)
		}
	}
}

// TestNewStringPooling tests that values outside cache range use the pool
func TestNewStringPooling(t *testing.T) {
	// Test values outside cache range
	val := "unique_test_string_12345"
	obj1 := NewString(val)
	obj2 := NewString(val)

	// They should have the correct value
	if obj1.Value != val {
		t.Errorf("NewString(%q).Value = %q, want %q", val, obj1.Value, val)
	}
	if obj2.Value != val {
		t.Errorf("NewString(%q).Value = %q, want %q", val, obj2.Value, val)
	}

	// Release and verify pool reuse
	ReleaseString(obj1)
	obj3 := NewString("another_unique_string")
	_ = obj3
	// obj3 may or may not be the same pointer as obj1 depending on pool behavior
}

// TestReleaseString tests the release functionality
func TestReleaseString(t *testing.T) {
	// Releasing a cached value should do nothing
	cachedVal := "true"
	obj1 := NewString(cachedVal)
	ReleaseString(obj1) // Should not panic or cause issues

	// Releasing a pooled value should return it to the pool
	pooledVal := "test_string_for_pool"
	obj2 := NewString(pooledVal)
	ReleaseString(obj2) // Should return to pool

	// Get stats to verify
	stats := GetStringPoolStats()
	if stats.Released < 1 {
		t.Error("Expected at least 1 release in stats")
	}
}

// TestReleaseStringSlice tests batch release
func TestReleaseStringSlice(t *testing.T) {
	objs := []*String{
		NewString("unique_string_1"),
		NewString("unique_string_2"),
		NewString("true"), // This is cached, won't be pooled
	}

	ReleaseStringSlice(objs)

	stats := GetStringPoolStats()
	// At least 2 should be released (the pooled ones)
	if stats.Released < 2 {
		t.Errorf("Expected at least 2 releases, got %d", stats.Released)
	}
}

// TestStringPoolStats tests statistics tracking
func TestStringPoolStats(t *testing.T) {
	ResetStringPoolStats()

	// Create some cached strings
	_ = NewString("")
	_ = NewString("true")
	_ = NewString("INT")

	stats := GetStringPoolStats()
	if stats.CacheHits != 3 {
		t.Errorf("Expected 3 cache hits, got %d", stats.CacheHits)
	}

	// Create pooled strings
	pooled := NewString("test_pooled_string")
	ReleaseString(pooled)

	stats = GetStringPoolStats()
	if stats.Released != 1 {
		t.Errorf("Expected 1 release, got %d", stats.Released)
	}
}

// TestBufferedStringPool tests the buffered pool operations
func TestBufferedStringPool(t *testing.T) {
	// Get and put through buffer
	obj1 := GetBufferedString("test_buffered_string")
	if obj1.Value != "test_buffered_string" {
		t.Errorf("GetBufferedString value mismatch")
	}

	PutBufferedString(obj1)

	// Get another, might be reused
	obj2 := GetBufferedString("another_buffered_string")
	if obj2.Value != "another_buffered_string" {
		t.Errorf("GetBufferedString value mismatch")
	}
}

// TestConcurrentNewString tests concurrent access to the pool
func TestConcurrentNewString(t *testing.T) {
	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				// Mix of cached and pooled values
				var val string
				if i%2 == 0 {
					val = string(rune('a' + (i % 26))) // Single char, likely cached
				} else {
					val = string(rune('a'+(i%26))) + "_unique_suffix_" + string(rune('0'+id%10))
				}
				obj := NewString(val)
				if obj.Value != val {
					t.Errorf("NewString(%q).Value = %q", val, obj.Value)
				}
				// Release non-cached values
				switch val {
				case "", "true", "false", "null", "INT", "FLOAT", "BOOL", "STRING",
					"ARRAY", "MAP", "FUNCTION", "BUILTIN", "ERROR", "nil", "0", "1", " ", "\n":
					// Don't release cached values
				default:
					ReleaseString(obj)
				}
			}
		}(g)
	}

	wg.Wait()
}

// TestNewStringSlice tests batch creation
func TestNewStringSlice(t *testing.T) {
	values := []string{"", "true", "unique_1", "INT", "unique_2"}
	result := NewStringSlice(values)

	if len(result) != len(values) {
		t.Fatalf("NewStringSlice returned %d elements, want %d", len(result), len(values))
	}

	for i, obj := range result {
		if obj.Value != values[i] {
			t.Errorf("result[%d].Value = %q, want %q", i, obj.Value, values[i])
		}
	}
}

// BenchmarkNewStringCached benchmarks creating cached strings
func BenchmarkNewStringCached(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewString("true")
	}
}

// BenchmarkNewStringPooled benchmarks creating pooled strings
func BenchmarkNewStringPooled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		obj := NewString("unique_benchmark_string")
		ReleaseString(obj)
	}
}

// BenchmarkNewStringPooledNoRelease benchmarks pooled strings without release
func BenchmarkNewStringPooledNoRelease(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewString("unique_benchmark_string_no_release")
	}
}

// BenchmarkNewStringDirect benchmarks direct struct creation (for comparison)
func BenchmarkNewStringDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &String{Value: "benchmark_string"}
	}
}

// BenchmarkNewStringSlice benchmarks batch creation
func BenchmarkNewStringSlice(b *testing.B) {
	values := []string{"one", "two", "three", "four", "five"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewStringSlice(values)
	}
}

// BenchmarkBufferedString benchmarks buffered pool access
func BenchmarkBufferedString(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := GetBufferedString("benchmark_buffered_string")
			PutBufferedString(obj)
		}
	})
}

// BenchmarkStringPoolContention tests pool performance under contention
func BenchmarkStringPoolContention(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := NewString("contention_test_string")
			ReleaseString(obj)
		}
	})
}

// TestWarmStringPool tests pool warming
func TestWarmStringPool(t *testing.T) {
	// Warm the pool
	WarmStringPool(100)

	// Force GC to clear any existing pooled objects
	runtime.GC()

	// The warm should have added objects
	// We can't directly verify this, but we can check that NewString still works
	obj := NewString("test_after_warm")
	if obj.Value != "test_after_warm" {
		t.Error("Pool warming affected NewString behavior")
	}
}

// TestCachedStringConstants tests that cached string constants are correct
func TestCachedStringConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant *String
		expected string
	}{
		{"STRING_EMPTY", STRING_EMPTY, ""},
		{"STRING_TRUE", STRING_TRUE, "true"},
		{"STRING_FALSE", STRING_FALSE, "false"},
		{"STRING_NULL", STRING_NULL, "null"},
		{"STRING_INT", STRING_INT, "INT"},
		{"STRING_FLOAT", STRING_FLOAT, "FLOAT"},
		{"STRING_BOOL", STRING_BOOL, "BOOL"},
		{"STRING_STRING", STRING_STRING, "STRING"},
		{"STRING_ARRAY", STRING_ARRAY, "ARRAY"},
		{"STRING_MAP", STRING_MAP, "MAP"},
		{"STRING_FUNC", STRING_FUNC, "FUNCTION"},
		{"STRING_BUILTIN", STRING_BUILTIN, "BUILTIN"},
		{"STRING_ERROR", STRING_ERROR, "ERROR"},
		{"STRING_NIL", STRING_NIL, "nil"},
		{"STRING_ZERO", STRING_ZERO, "0"},
		{"STRING_ONE", STRING_ONE, "1"},
		{"STRING_SPACE", STRING_SPACE, " "},
		{"STRING_NEWLINE", STRING_NEWLINE, "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant.Value != tt.expected {
				t.Errorf("%s.Value = %q, want %q", tt.name, tt.constant.Value, tt.expected)
			}
			// Verify NewString returns the same pointer
			obj := NewString(tt.expected)
			if obj != tt.constant {
				t.Errorf("NewString(%q) != %s, expected same pointer", tt.expected, tt.name)
			}
		})
	}
}

// TestInternStringWithStats tests the string interning with stats tracking
func TestInternStringWithStats(t *testing.T) {
	ResetStringPoolStats()

	// First intern should create new
	val := "intern_test_string_unique"
	obj1 := InternString(val)

	// Second intern should return same pointer
	obj2 := InternString(val)
	if obj1 != obj2 {
		t.Error("InternString should return same pointer for same value")
	}

	// Check stats
	stats := GetStringPoolStats()
	if stats.InternHits != 1 {
		t.Errorf("Expected 1 intern hit, got %d", stats.InternHits)
	}
}

// TestInternBatchWithDuplicates tests batch interning with duplicate detection
func TestInternBatchWithDuplicates(t *testing.T) {
	values := []string{"unique_one", "unique_two", "unique_three", "unique_one", "unique_two"} // duplicates
	result := InternBatch(values)

	if len(result) != len(values) {
		t.Fatalf("InternBatch returned %d elements, want %d", len(result), len(values))
	}

	// Verify duplicates return same pointer
	if result[0] != result[3] {
		t.Error("Interned 'unique_one' should return same pointer")
	}
	if result[1] != result[4] {
		t.Error("Interned 'unique_two' should return same pointer")
	}
}
