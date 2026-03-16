// pkg/stdlib/time_test.go
package stdlib

import (
	"testing"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestTimeModuleRegistered(t *testing.T) {
	mod := Get("time")
	if mod == nil {
		t.Fatal("time module not registered")
	}
	if mod.Name != "time" {
		t.Errorf("module name = %s, want time", mod.Name)
	}
}

func TestTimeUnix(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["unix"].(*objects.Builtin)
	if !ok {
		t.Fatal("unix not found or not a builtin")
	}

	before := time.Now().Unix()
	result := fn.Fn()
	after := time.Now().Unix()

	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if intResult.Value < before || intResult.Value > after {
		t.Errorf("unix() = %d, want between %d and %d", intResult.Value, before, after)
	}
}

func TestTimeUnixMs(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["unixMs"].(*objects.Builtin)
	if !ok {
		t.Fatal("unixMs not found or not a builtin")
	}

	before := time.Now().UnixMilli()
	result := fn.Fn()
	after := time.Now().UnixMilli()

	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if intResult.Value < before || intResult.Value > after {
		t.Errorf("unixMs() = %d, want between %d and %d", intResult.Value, before, after)
	}
}

func TestTimeNow(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["now"].(*objects.Builtin)
	if !ok {
		t.Fatal("now not found or not a builtin")
	}

	result := fn.Fn()
	mapResult, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	// Check that all expected keys exist
	expectedKeys := []string{"year", "month", "day", "hour", "minute", "second", "nanosecond"}
	for _, key := range expectedKeys {
		found := false
		for _, pair := range mapResult.Pairs {
			if k, ok := pair.Key.(*objects.String); ok && k.Value == key {
				found = true
				if _, ok := pair.Value.(*objects.Int); !ok {
					t.Errorf("now()[%s] is not Int", key)
				}
				break
			}
		}
		if !found {
			t.Errorf("now() missing key: %s", key)
		}
	}
}

func TestTimeYear(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["year"].(*objects.Builtin)
	if !ok {
		t.Fatal("year not found or not a builtin")
	}

	result := fn.Fn()
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	expectedYear := int64(time.Now().Year())
	if intResult.Value != expectedYear {
		t.Errorf("year() = %d, want %d", intResult.Value, expectedYear)
	}
}

func TestTimeMonth(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["month"].(*objects.Builtin)
	if !ok {
		t.Fatal("month not found or not a builtin")
	}

	result := fn.Fn()
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	expectedMonth := int64(time.Now().Month())
	if intResult.Value != expectedMonth {
		t.Errorf("month() = %d, want %d", intResult.Value, expectedMonth)
	}
}

func TestTimeDay(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["day"].(*objects.Builtin)
	if !ok {
		t.Fatal("day not found or not a builtin")
	}

	result := fn.Fn()
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	expectedDay := int64(time.Now().Day())
	if intResult.Value != expectedDay {
		t.Errorf("day() = %d, want %d", intResult.Value, expectedDay)
	}
}

func TestTimeHour(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["hour"].(*objects.Builtin)
	if !ok {
		t.Fatal("hour not found or not a builtin")
	}

	result := fn.Fn()
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if intResult.Value < 0 || intResult.Value > 23 {
		t.Errorf("hour() = %d, want 0-23", intResult.Value)
	}
}

func TestTimeMinute(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["minute"].(*objects.Builtin)
	if !ok {
		t.Fatal("minute not found or not a builtin")
	}

	result := fn.Fn()
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if intResult.Value < 0 || intResult.Value > 59 {
		t.Errorf("minute() = %d, want 0-59", intResult.Value)
	}
}

func TestTimeSecond(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["second"].(*objects.Builtin)
	if !ok {
		t.Fatal("second not found or not a builtin")
	}

	result := fn.Fn()
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if intResult.Value < 0 || intResult.Value > 59 {
		t.Errorf("second() = %d, want 0-59", intResult.Value)
	}
}

func TestTimeWeekday(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["weekday"].(*objects.Builtin)
	if !ok {
		t.Fatal("weekday not found or not a builtin")
	}

	result := fn.Fn()
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if intResult.Value < 0 || intResult.Value > 6 {
		t.Errorf("weekday() = %d, want 0-6", intResult.Value)
	}
}

func TestTimeSleep(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["sleep"].(*objects.Builtin)
	if !ok {
		t.Fatal("sleep not found or not a builtin")
	}

	// Test sleep with 10ms
	start := time.Now()
	result := fn.Fn(Int(10))
	elapsed := time.Since(start)

	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("sleep() should return NULL, got %T", result)
	}

	if elapsed < 10*time.Millisecond {
		t.Errorf("sleep(10) took %v, want at least 10ms", elapsed)
	}
}

func TestTimeSleepErrors(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["sleep"].(*objects.Builtin)
	if !ok {
		t.Fatal("sleep not found or not a builtin")
	}

	// No arguments
	result := fn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sleep() with no args should return Error")
	}

	// Wrong type
	result = fn.Fn(String("100"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("sleep(string) should return Error")
	}
}

func TestTimeSleepSec(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["sleepSec"].(*objects.Builtin)
	if !ok {
		t.Fatal("sleepSec not found or not a builtin")
	}

	// Test sleepSec with 0 seconds (just check it works)
	result := fn.Fn(Int(0))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("sleepSec() should return NULL, got %T", result)
	}
}

func TestTimeFormat(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["format"].(*objects.Builtin)
	if !ok {
		t.Fatal("format not found or not a builtin")
	}

	layout := String("2006-01-02")
	result := fn.Fn(layout)

	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	// Check format is YYYY-MM-DD (10 characters)
	if len(strResult.Value) != 10 {
		t.Errorf("format(2006-01-02) = %s, want format YYYY-MM-DD", strResult.Value)
	}
}

func TestTimeFormatUnix(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["formatUnix"].(*objects.Builtin)
	if !ok {
		t.Fatal("formatUnix not found or not a builtin")
	}

	// Known timestamp: 2024-01-01 00:00:00 UTC = 1704067200
	ts := Int(1704067200)
	layout := String("2006-01-02")
	result := fn.Fn(ts, layout)

	strResult, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if strResult.Value != "2024-01-01" {
		t.Errorf("formatUnix(1704067200, 2006-01-02) = %s, want 2024-01-01", strResult.Value)
	}
}

func TestTimeParse(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["parse"].(*objects.Builtin)
	if !ok {
		t.Fatal("parse not found or not a builtin")
	}

	layout := String("2006-01-02")
	value := String("2024-01-01")
	result := fn.Fn(layout, value)

	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	// 2024-01-01 00:00:00 UTC = 1704067200
	if intResult.Value != 1704067200 {
		t.Errorf("parse(2006-01-02, 2024-01-01) = %d, want 1704067200", intResult.Value)
	}
}

func TestTimeParseError(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["parse"].(*objects.Builtin)
	if !ok {
		t.Fatal("parse not found or not a builtin")
	}

	layout := String("2006-01-02")
	value := String("invalid-date")
	result := fn.Fn(layout, value)

	if _, ok := result.(*objects.Error); !ok {
		t.Error("parse() with invalid date should return Error")
	}
}

func TestTimeSince(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["since"].(*objects.Builtin)
	if !ok {
		t.Fatal("since not found or not a builtin")
	}

	// 100ms ago
	start := time.Now().Add(-100 * time.Millisecond).UnixMilli()
	result := fn.Fn(Int(start))

	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	// Should be around 100ms, allow some tolerance
	if intResult.Value < 90 || intResult.Value > 200 {
		t.Errorf("since(100ms ago) = %d, want around 100", intResult.Value)
	}
}

func TestTimeAddDays(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["addDays"].(*objects.Builtin)
	if !ok {
		t.Fatal("addDays not found or not a builtin")
	}

	result := fn.Fn(Int(1))
	_, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	// Just verify it returns a timestamp greater than now
	now := time.Now().Unix()
	if result.(*objects.Int).Value < now {
		t.Error("addDays(1) should return a timestamp in the future")
	}
}

func TestTimeAddMonths(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["addMonths"].(*objects.Builtin)
	if !ok {
		t.Fatal("addMonths not found or not a builtin")
	}

	result := fn.Fn(Int(1))
	_, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
}

func TestTimeAddYears(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["addYears"].(*objects.Builtin)
	if !ok {
		t.Fatal("addYears not found or not a builtin")
	}

	result := fn.Fn(Int(1))
	_, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
}

func TestTimeIsLeapYear(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["isLeapYear"].(*objects.Builtin)
	if !ok {
		t.Fatal("isLeapYear not found or not a builtin")
	}

	tests := []struct {
		year     int64
		expected bool
	}{
		{2000, true},  // Divisible by 400
		{2004, true},  // Divisible by 4, not by 100
		{1900, false}, // Divisible by 100, not by 400
		{2001, false}, // Not divisible by 4
		{2024, true},  // Leap year
		{2023, false}, // Not a leap year
	}

	for _, tt := range tests {
		result := fn.Fn(Int(tt.year))
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if boolResult.Value != tt.expected {
			t.Errorf("isLeapYear(%d) = %v, want %v", tt.year, boolResult.Value, tt.expected)
		}
	}
}

func TestTimeDaysInMonth(t *testing.T) {
	mod := Get("time")
	fn, ok := mod.Exports["daysInMonth"].(*objects.Builtin)
	if !ok {
		t.Fatal("daysInMonth not found or not a builtin")
	}

	tests := []struct {
		year     int64
		month    int64
		expected int64
	}{
		{2024, 1, 31},  // January
		{2024, 2, 29},  // February in leap year
		{2023, 2, 28},  // February in non-leap year
		{2024, 4, 30},  // April
		{2024, 12, 31}, // December
	}

	for _, tt := range tests {
		result := fn.Fn(Int(tt.year), Int(tt.month))
		intResult, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if intResult.Value != tt.expected {
			t.Errorf("daysInMonth(%d, %d) = %d, want %d", tt.year, tt.month, intResult.Value, tt.expected)
		}
	}
}

func TestTimeAllExportsExist(t *testing.T) {
	mod := Get("time")
	expectedExports := []string{
		"unix", "unixMs", "unixNano", "now",
		"year", "month", "day", "hour", "minute", "second", "weekday",
		"sleep", "sleepSec",
		"format", "formatUnix", "parse",
		"since", "addDays", "addMonths", "addYears",
		"isLeapYear", "daysInMonth",
	}

	for _, name := range expectedExports {
		if _, ok := mod.Exports[name]; !ok {
			t.Errorf("missing export: %s", name)
		}
	}
}
