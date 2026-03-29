// pkg/objects/time_obj.go
// Time object type for Xxlang
package objects

import (
	"fmt"
	"time"
)

// Time represents a time object with various time components
type Time struct {
	Value time.Time
}

// NewTime creates a new Time object from time.Time
func NewTime(t time.Time) *Time {
	return &Time{Value: t}
}

// NewTimeFromTimestamp creates a new Time object from Unix timestamp (auto-detects seconds/milliseconds)
func NewTimeFromTimestamp(timestamp int64) *Time {
	// Auto-detect: if timestamp > 1e12, treat as milliseconds
	if timestamp > 1e12 {
		return &Time{Value: time.UnixMilli(timestamp)}
	}
	return &Time{Value: time.Unix(timestamp, 0)}
}

// NewTimeFromComponents creates a new Time object from date/time components
func NewTimeFromComponents(year, month, day, hour, minute, second, nano int) *Time {
	return &Time{Value: time.Date(year, time.Month(month), day, hour, minute, second, nano, time.Local)}
}

func (t *Time) Type() ObjectType { return TimeType }
func (t *Time) TypeTag() TypeTag { return TagTime }

func (t *Time) Inspect() string {
	return fmt.Sprintf("<Time: %s>", t.Value.Format("2006-01-02 15:04:05"))
}

func (t *Time) ToBool() *Bool { return TRUE }

func (t *Time) HashKey() HashKey {
	return HashKey{
		Type:  TimeType,
		Value: uint64(t.Value.UnixNano()),
	}
}

// GetYear returns the year
func (t *Time) GetYear() int {
	return t.Value.Year()
}

// GetMonth returns the month (1-12)
func (t *Time) GetMonth() int {
	return int(t.Value.Month())
}

// GetDay returns the day of month
func (t *Time) GetDay() int {
	return t.Value.Day()
}

// GetHour returns the hour (0-23)
func (t *Time) GetHour() int {
	return t.Value.Hour()
}

// GetMinute returns the minute (0-59)
func (t *Time) GetMinute() int {
	return t.Value.Minute()
}

// GetSecond returns the second (0-59)
func (t *Time) GetSecond() int {
	return t.Value.Second()
}

// GetNanosecond returns the nanosecond
func (t *Time) GetNanosecond() int {
	return t.Value.Nanosecond()
}

// GetWeekday returns the weekday (0=Sunday, 1=Monday, ..., 6=Saturday)
func (t *Time) GetWeekday() int {
	return int(t.Value.Weekday())
}

// GetTimestamp returns Unix timestamp in seconds
func (t *Time) GetTimestamp() int64 {
	return t.Value.Unix()
}

// GetTimestampMs returns Unix timestamp in milliseconds
func (t *Time) GetTimestampMs() int64 {
	return t.Value.UnixMilli()
}

// Format returns formatted time string
func (t *Time) Format(layout string) string {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return t.Value.Format(layout)
}

// AddSecs adds seconds to the time
func (t *Time) AddSecs(secs float64) *Time {
	return &Time{Value: t.Value.Add(time.Duration(secs * float64(time.Second)))}
}

// AddDate adds years, months, days to the time
func (t *Time) AddDate(years, months, days int) *Time {
	return &Time{Value: t.Value.AddDate(years, months, days)}
}

// AddDuration adds a duration string (e.g., "1h", "30m", "1h30m")
func (t *Time) AddDuration(durStr string) (*Time, error) {
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return nil, err
	}
	return &Time{Value: t.Value.Add(dur)}, nil
}

// Before checks if this time is before another time
func (t *Time) Before(other *Time) bool {
	return t.Value.Before(other.Value)
}

// After checks if this time is after another time
func (t *Time) After(other *Time) bool {
	return t.Value.After(other.Value)
}

// Equal checks if this time equals another time
func (t *Time) Equal(other *Time) bool {
	return t.Value.Equal(other.Value)
}

// Sub returns the duration between this time and another time
func (t *Time) Sub(other *Time) time.Duration {
	return t.Value.Sub(other.Value)
}

// DiffSecs returns the difference in seconds between this time and another time
func (t *Time) DiffSecs(other *Time) float64 {
	return t.Value.Sub(other.Value).Seconds()
}

// IsZero checks if the time is zero value
func (t *Time) IsZero() bool {
	return t.Value.IsZero()
}

// ToMap converts Time to a Map for backward compatibility
func (t *Time) ToMap() *Map {
	pairs := make(map[HashKey]MapPair)

	pairs[NewString("year").HashKey()] = MapPair{
		Key: NewString("year"), Value: NewInt(int64(t.GetYear())),
	}
	pairs[NewString("month").HashKey()] = MapPair{
		Key: NewString("month"), Value: NewInt(int64(t.GetMonth())),
	}
	pairs[NewString("day").HashKey()] = MapPair{
		Key: NewString("day"), Value: NewInt(int64(t.GetDay())),
	}
	pairs[NewString("hour").HashKey()] = MapPair{
		Key: NewString("hour"), Value: NewInt(int64(t.GetHour())),
	}
	pairs[NewString("minute").HashKey()] = MapPair{
		Key: NewString("minute"), Value: NewInt(int64(t.GetMinute())),
	}
	pairs[NewString("second").HashKey()] = MapPair{
		Key: NewString("second"), Value: NewInt(int64(t.GetSecond())),
	}
	pairs[NewString("nanosecond").HashKey()] = MapPair{
		Key: NewString("nanosecond"), Value: NewInt(int64(t.GetNanosecond())),
	}
	pairs[NewString("timestamp").HashKey()] = MapPair{
		Key: NewString("timestamp"), Value: NewInt(t.GetTimestamp()),
	}
	pairs[NewString("timestampMs").HashKey()] = MapPair{
		Key: NewString("timestampMs"), Value: NewInt(t.GetTimestampMs()),
	}
	pairs[NewString("weekday").HashKey()] = MapPair{
		Key: NewString("weekday"), Value: NewInt(int64(t.GetWeekday())),
	}
	pairs[NewString("str").HashKey()] = MapPair{
		Key: NewString("str"), Value: NewString(t.Format("")),
	}

	return NewMap(pairs)
}
