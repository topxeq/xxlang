// pkg/objects/time_extra_test.go
package objects

import (
	"testing"
	"time"
)

func TestNewTimeFromTimestamp(t *testing.T) {
	ts := time.Now().Unix()
	tm := NewTimeFromTimestamp(ts)
	if tm == nil {
		t.Fatal("expected time instance")
	}
}

func TestNewTimeFromComponents(t *testing.T) {
	tm := NewTimeFromComponents(2024, 1, 15, 10, 30, 0, 0)
	if tm == nil {
		t.Fatal("expected time instance")
	}
}

func TestTimeGetTimestamp(t *testing.T) {
	now := time.Now()
	tm := NewTime(now)

	ts := tm.GetTimestamp()
	if ts == 0 {
		t.Error("expected non-zero timestamp")
	}
}

func TestTimeFormat(t *testing.T) {
	now := time.Now()
	tm := NewTime(now)

	s := tm.Format("2006-01-02")
	if s == "" {
		t.Error("expected formatted string")
	}
}

func TestTimeGetYear(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	if tm.GetYear() != 2024 {
		t.Errorf("expected 2024, got %d", tm.GetYear())
	}
}

func TestTime_AddSecs(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	newTm := tm.AddSecs(3600) // Add 1 hour

	if newTm.GetHour() != 11 {
		t.Errorf("expected hour 11, got %d", newTm.GetHour())
	}
}

func TestTime_AddDate(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	newTm := tm.AddDate(1, 0, 0) // Add 1 year

	if newTm.GetYear() != 2025 {
		t.Errorf("expected year 2025, got %d", newTm.GetYear())
	}
}

func TestTime_AddDuration(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	newTm, err := tm.AddDuration("1h30m")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newTm.GetHour() != 12 {
		t.Errorf("expected hour 12, got %d", newTm.GetHour())
	}
	if newTm.GetMinute() != 0 {
		t.Errorf("expected minute 0, got %d", newTm.GetMinute())
	}
}

func TestTime_AddDuration_Invalid(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	_, err := tm.AddDuration("invalid")

	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestTime_BeforeAfter(t *testing.T) {
	t1 := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	t2 := NewTimeFromComponents(2024, 6, 15, 11, 30, 0, 0)

	if !t1.Before(t2) {
		t.Error("expected t1 to be before t2")
	}
	if t1.After(t2) {
		t.Error("expected t1 not to be after t2")
	}
}

func TestTime_Equal(t *testing.T) {
	t1 := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	t2 := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	t3 := NewTimeFromComponents(2024, 6, 15, 11, 30, 0, 0)

	if !t1.Equal(t2) {
		t.Error("expected t1 to equal t2")
	}
	if t1.Equal(t3) {
		t.Error("expected t1 not to equal t3")
	}
}

func TestTime_Sub(t *testing.T) {
	t1 := NewTimeFromComponents(2024, 6, 15, 11, 30, 0, 0)
	t2 := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)

	diff := t1.Sub(t2)
	if diff != time.Hour {
		t.Errorf("expected 1 hour difference, got %v", diff)
	}
}

func TestTime_DiffSecs(t *testing.T) {
	t1 := NewTimeFromComponents(2024, 6, 15, 11, 30, 0, 0)
	t2 := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)

	diff := t1.DiffSecs(t2)
	if diff != 3600 {
		t.Errorf("expected 3600 seconds difference, got %f", diff)
	}
}

func TestTime_IsZero(t *testing.T) {
	tm := NewTime(time.Time{})
	if !tm.IsZero() {
		t.Error("expected IsZero to return true for zero time")
	}

	tm = NewTime(time.Now())
	if tm.IsZero() {
		t.Error("expected IsZero to return false for non-zero time")
	}
}

func TestTime_ToMap(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 45, 0)
	m := tm.ToMap()

	if m.Type() != MapType {
		t.Fatalf("expected Map, got %s", m.Type())
	}

	// Check some fields
	yearKey := NewString("year").HashKey()
	if pair, ok := m.Pairs[yearKey]; ok {
		if pair.Value.(*Int).Value != 2024 {
			t.Errorf("expected year 2024, got %d", pair.Value.(*Int).Value)
		}
	} else {
		t.Error("expected 'year' key in map")
	}
}

func TestTime_Getters(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 45, 123456789)

	if tm.GetMonth() != 6 {
		t.Errorf("expected month 6, got %d", tm.GetMonth())
	}
	if tm.GetDay() != 15 {
		t.Errorf("expected day 15, got %d", tm.GetDay())
	}
	if tm.GetHour() != 10 {
		t.Errorf("expected hour 10, got %d", tm.GetHour())
	}
	if tm.GetMinute() != 30 {
		t.Errorf("expected minute 30, got %d", tm.GetMinute())
	}
	if tm.GetSecond() != 45 {
		t.Errorf("expected second 45, got %d", tm.GetSecond())
	}
	if tm.GetNanosecond() != 123456789 {
		t.Errorf("expected nanosecond 123456789, got %d", tm.GetNanosecond())
	}
}

func TestTime_TimestampMs(t *testing.T) {
	tm := NewTime(time.Unix(1609459200, 0))
	if tm.GetTimestampMs() != 1609459200000 {
		t.Errorf("expected timestamp ms 1609459200000, got %d", tm.GetTimestampMs())
	}
}

func TestTime_Inspect(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 45, 0)
	inspect := tm.Inspect()

	if inspect == "" {
		t.Error("expected non-empty inspect string")
	}
}

func TestTime_HashKey(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 45, 0)
	key := tm.HashKey()

	if key.Type != TimeType {
		t.Errorf("expected HashKey.Type TIME, got %s", key.Type)
	}
}

func TestTimeGetMonth(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	if tm.GetMonth() != 6 {
		t.Errorf("expected 6, got %d", tm.GetMonth())
	}
}

func TestTimeGetDay(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	if tm.GetDay() != 15 {
		t.Errorf("expected 15, got %d", tm.GetDay())
	}
}

func TestTimeGetHour(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	if tm.GetHour() != 10 {
		t.Errorf("expected 10, got %d", tm.GetHour())
	}
}

func TestTimeGetMinute(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	if tm.GetMinute() != 30 {
		t.Errorf("expected 30, got %d", tm.GetMinute())
	}
}

func TestTimeGetSecond(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 45, 0)
	if tm.GetSecond() != 45 {
		t.Errorf("expected 45, got %d", tm.GetSecond())
	}
}

func TestTimeGetWeekday(t *testing.T) {
	tm := NewTimeFromComponents(2024, 6, 15, 10, 30, 0, 0)
	weekday := tm.GetWeekday()
	if weekday < 0 || weekday > 6 {
		t.Errorf("expected weekday 0-6, got %d", weekday)
	}
}
