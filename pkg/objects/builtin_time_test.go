// pkg/objects/builtin_time_test.go
package objects

import (
	"testing"
	"time"
)

func TestBuiltinGetNowStr(t *testing.T) {
	fn, ok := Builtins["getNowStr"]
	if !ok {
		t.Fatal("getNowStr builtin not found")
	}

	result := fn.Fn()
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn(NewString("2006"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string arg")
	}
}

func TestBuiltinGetNowStrCompact(t *testing.T) {
	fn, ok := Builtins["getNowStrCompact"]
	if !ok {
		t.Fatal("getNowStrCompact builtin not found")
	}

	result := fn.Fn()
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}
}

func TestBuiltinGetNowTimeStamp(t *testing.T) {
	fn, ok := Builtins["getNowTimeStamp"]
	if !ok {
		t.Fatal("getNowTimeStamp builtin not found")
	}

	result := fn.Fn()
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value == 0 {
		t.Error("expected non-zero timestamp")
	}

	now := time.Now().Unix()
	if intResult.Value < now-1 || intResult.Value > now+1 {
		t.Errorf("expected ~%d, got %d", now, intResult.Value)
	}

	result = fn.Fn(NewInt(1))
	if !isError(result) {
		t.Error("expected error for args")
	}
}

func TestBuiltinFormatTime(t *testing.T) {
	fn, ok := Builtins["formatTime"]
	if !ok {
		t.Fatal("formatTime builtin not found")
	}

	result := fn.Fn(NewInt(1704067200))
	strResult, ok := result.(*String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if strResult.Value == "" {
		t.Error("expected non-empty result")
	}

	result = fn.Fn(NewFloat(1704067200.0))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String for Float, got %T", result)
	}

	result = fn.Fn(NewString("not a number"))
	if !isError(result) {
		t.Error("expected error for string arg")
	}

	result = fn.Fn(NewInt(1704067200), NewString("2006"))
	strResult, ok = result.(*String)
	if !ok {
		t.Fatalf("expected String with format, got %T", result)
	}
}

func TestBuiltinTimeToTick(t *testing.T) {
	fn, ok := Builtins["timeToTick"]
	if !ok {
		t.Fatal("timeToTick builtin not found")
	}

	result := fn.Fn(NewInt(2024), NewInt(1), NewInt(1))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value == 0 {
		t.Error("expected non-zero tick")
	}

	result = fn.Fn(NewInt(2024), NewInt(1), NewInt(1), NewInt(0), NewInt(0), NewInt(0))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	result = fn.Fn(NewInt(2024))
	if !isError(result) {
		t.Error("expected error for wrong arg count")
	}
}

func TestBuiltinTimeAddSecs(t *testing.T) {
	fn, ok := Builtins["timeAddSecs"]
	if !ok {
		t.Fatal("timeAddSecs builtin not found")
	}

	result := fn.Fn(NewInt(1704067200), NewInt(3600))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1704070800 {
		t.Errorf("expected 1704070800, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(1704067200), NewInt(-3600))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 1704063600 {
		t.Errorf("expected 1704063600, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTimeAddDate(t *testing.T) {
	fn, ok := Builtins["timeAddDate"]
	if !ok {
		t.Fatal("timeAddDate builtin not found")
	}

	result := fn.Fn(NewInt(1704067200), NewInt(1), NewInt(0), NewInt(0))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value == 0 {
		t.Error("expected non-zero result")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTimeBefore(t *testing.T) {
	fn, ok := Builtins["timeBefore"]
	if !ok {
		t.Fatal("timeBefore builtin not found")
	}

	result := fn.Fn(NewInt(100), NewInt(200))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected TRUE for 100 < 200")
	}

	result = fn.Fn(NewInt(200), NewInt(100))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected FALSE for 200 > 100")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinStrToTime(t *testing.T) {
	fn, ok := Builtins["strToTime"]
	if !ok {
		t.Fatal("strToTime builtin not found")
	}

	result := fn.Fn(NewString("2024-01-01 00:00:00"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value == 0 {
		t.Error("expected non-zero timestamp")
	}

	result = fn.Fn(NewString("invalid"))
	if !isError(result) {
		t.Error("expected error for invalid string")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTimeAfter(t *testing.T) {
	fn, ok := Builtins["timeAfter"]
	if !ok {
		t.Fatal("timeAfter builtin not found")
	}

	result := fn.Fn(NewInt(200), NewInt(100))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected TRUE for 200 > 100")
	}

	result = fn.Fn(NewInt(100), NewInt(200))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected FALSE for 100 < 200")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTimeEqual(t *testing.T) {
	fn, ok := Builtins["timeEqual"]
	if !ok {
		t.Fatal("timeEqual builtin not found")
	}

	result := fn.Fn(NewInt(100), NewInt(100))
	boolResult, ok := result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if !boolResult.Value {
		t.Error("expected TRUE for 100 == 100")
	}

	result = fn.Fn(NewInt(100), NewInt(200))
	boolResult, ok = result.(*Bool)
	if !ok {
		t.Fatalf("expected Bool, got %T", result)
	}
	if boolResult.Value {
		t.Error("expected FALSE for 100 != 200")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTimeDiff(t *testing.T) {
	fn, ok := Builtins["timeDiff"]
	if !ok {
		t.Fatal("timeDiff builtin not found")
	}

	result := fn.Fn(NewInt(100), NewInt(200))
	mapResult, ok := result.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", result)
	}

	totalSecondsPair, ok := mapResult.Pairs[NewString("totalSeconds").HashKey()]
	if !ok {
		t.Fatal("expected totalSeconds key in map")
	}
	intResult, ok := totalSecondsPair.Value.(*Int)
	if !ok {
		t.Fatalf("expected Int for totalSeconds, got %T", totalSecondsPair.Value)
	}
	if intResult.Value != 100 {
		t.Errorf("expected 100, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTimeDiffSecs(t *testing.T) {
	fn, ok := Builtins["timeDiffSecs"]
	if !ok {
		t.Fatal("timeDiffSecs builtin not found")
	}

	result := fn.Fn(NewInt(100), NewInt(200))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 100 {
		t.Errorf("expected 100, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(200), NewInt(100))
	intResult, ok = result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != -100 {
		t.Errorf("expected -100, got %d", intResult.Value)
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinParseTime(t *testing.T) {
	fn, ok := Builtins["parseTime"]
	if !ok {
		t.Fatal("parseTime builtin not found")
	}

	result := fn.Fn(NewString("2024-01-01 00:00:00"), NewString("2006-01-02 15:04:05"))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value == 0 {
		t.Error("expected non-zero timestamp")
	}

	result = fn.Fn(NewString("invalid"), NewString("2006-01-02"))
	if !isError(result) {
		t.Error("expected error for invalid string")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinIsTime(t *testing.T) {
	fn, ok := Builtins["isTime"]
	if !ok {
		t.Fatal("isTime builtin not found")
	}

	result := fn.Fn(NewInt(1704067200))
	if result != TRUE {
		t.Error("expected TRUE for Int")
	}

	result = fn.Fn(NewFloat(1704067200.0))
	if result != TRUE {
		t.Error("expected TRUE for Float")
	}

	result = fn.Fn(NewString("hello"))
	if result != FALSE {
		t.Error("expected FALSE for String")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinNow(t *testing.T) {
	fn, ok := Builtins["now"]
	if !ok {
		t.Fatal("now builtin not found")
	}

	result := fn.Fn()
	timeResult, ok := result.(*Time)
	if !ok {
		t.Fatalf("expected Time, got %T", result)
	}
	if timeResult.Value.IsZero() {
		t.Error("expected non-zero time")
	}

	result = fn.Fn(NewInt(1))
	if !isError(result) {
		t.Error("expected error for args")
	}
}

func TestBuiltinTimeToTimeStamp(t *testing.T) {
	fn, ok := Builtins["timeToTimeStamp"]
	if !ok {
		t.Fatal("timeToTimeStamp builtin not found")
	}

	now := time.Now()
	result := fn.Fn(NewTime(now))
	intResult, ok := result.(*Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value == 0 {
		t.Error("expected non-zero timestamp")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinTimeStampToTime(t *testing.T) {
	fn, ok := Builtins["timeStampToTime"]
	if !ok {
		t.Fatal("timeStampToTime builtin not found")
	}

	result := fn.Fn(NewInt(1704067200))
	timeResult, ok := result.(*Time)
	if !ok {
		t.Fatalf("expected Time, got %T", result)
	}
	if timeResult.Value.IsZero() {
		t.Error("expected non-zero time")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int64
		expected string
	}{
		{0, "0s"},
		{30, "30s"},
		{60, "1m 0s"},
		{90, "1m 30s"},
		{3600, "1h 0m 0s"},
		{3661, "1h 1m 1s"},
		{86400, "1d 0h 0m 0s"},
		{90061, "1d 1h 1m 1s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.seconds)
		if result != tt.expected {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.seconds, result, tt.expected)
		}
	}
}
