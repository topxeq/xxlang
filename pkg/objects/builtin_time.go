// pkg/objects/builtin_time.go
// Time related built-in functions for Xxlang
package objects

import (
	"fmt"
	"time"
)

func init() {
	// Time enhancement functions
	Builtins["getNowStr"] = &Builtin{Fn: builtinGetNowStr}
	Builtins["getNowTimeStamp"] = &Builtin{Fn: builtinGetNowTimeStamp}
	Builtins["formatTime"] = &Builtin{Fn: builtinFormatTime}
	Builtins["timeToTick"] = &Builtin{Fn: builtinTimeToTick}
	Builtins["timeAddSecs"] = &Builtin{Fn: builtinTimeAddSecs}
	Builtins["timeAddDate"] = &Builtin{Fn: builtinTimeAddDate}
	Builtins["timeBefore"] = &Builtin{Fn: builtinTimeBefore}
	Builtins["strToTime"] = &Builtin{Fn: builtinStrToTime}
	Builtins["timeAfter"] = &Builtin{Fn: builtinTimeAfter}
	Builtins["timeEqual"] = &Builtin{Fn: builtinTimeEqual}
	Builtins["timeDiff"] = &Builtin{Fn: builtinTimeDiff}
	Builtins["timeDiffSecs"] = &Builtin{Fn: builtinTimeDiffSecs}
	Builtins["parseTime"] = &Builtin{Fn: builtinParseTime}
	Builtins["isTime"] = &Builtin{Fn: builtinIsTime}
}

// getNowStr - get current time as formatted string
// Usage: getNowStr() -> string (default format: "2006-01-02 15:04:05")
//
//	getNowStr(format) -> string
func builtinGetNowStr(args ...Object) Object {
	if len(args) > 1 {
		return newError("wrong number of arguments for getNowStr. got=%d, want=0 or 1", len(args))
	}

	format := "2006-01-02 15:04:05"
	if len(args) == 1 {
		f, ok := args[0].(*String)
		if !ok {
			return newError("argument to 'getNowStr' must be STRING, got %s", args[0].Type())
		}
		format = f.Value
	}

	return NewString(time.Now().Format(format))
}

// getNowTimeStamp - get current Unix timestamp
// Usage: getNowTimeStamp() -> int
func builtinGetNowTimeStamp(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getNowTimeStamp. got=%d, want=0", len(args))
	}
	return NewInt(time.Now().Unix())
}

// formatTime - format timestamp to string
// Usage: formatTime(timestamp) -> string
//
//	formatTime(timestamp, format) -> string
func builtinFormatTime(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for formatTime. got=%d, want=1 or 2", len(args))
	}

	var timestamp int64
	switch arg := args[0].(type) {
	case *Int:
		timestamp = arg.Value
	case *Float:
		timestamp = int64(arg.Value)
	default:
		return newError("first argument to 'formatTime' must be INT or FLOAT, got %s", args[0].Type())
	}

	format := "2006-01-02 15:04:05"
	if len(args) == 2 {
		f, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'formatTime' must be STRING, got %s", args[1].Type())
		}
		format = f.Value
	}

	t := time.Unix(timestamp, 0)
	return NewString(t.Format(format))
}

// timeToTick - convert time to Unix timestamp
// Usage: timeToTick(year, month, day) -> int
//
//	timeToTick(year, month, day, hour, minute, second) -> int
func builtinTimeToTick(args ...Object) Object {
	if len(args) != 3 && len(args) != 6 {
		return newError("wrong number of arguments for timeToTick. got=%d, want=3 or 6", len(args))
	}

	year, ok := args[0].(*Int)
	if !ok {
		return newError("first argument to 'timeToTick' must be INT, got %s", args[0].Type())
	}

	month, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'timeToTick' must be INT, got %s", args[1].Type())
	}

	day, ok := args[2].(*Int)
	if !ok {
		return newError("third argument to 'timeToTick' must be INT, got %s", args[2].Type())
	}

	hour := int64(0)
	minute := int64(0)
	second := int64(0)

	if len(args) == 6 {
		h, ok := args[3].(*Int)
		if !ok {
			return newError("fourth argument to 'timeToTick' must be INT, got %s", args[3].Type())
		}
		hour = h.Value

		m, ok := args[4].(*Int)
		if !ok {
			return newError("fifth argument to 'timeToTick' must be INT, got %s", args[4].Type())
		}
		minute = m.Value

		s, ok := args[5].(*Int)
		if !ok {
			return newError("sixth argument to 'timeToTick' must be INT, got %s", args[5].Type())
		}
		second = s.Value
	}

	t := time.Date(int(year.Value), time.Month(month.Value), int(day.Value),
		int(hour), int(minute), int(second), 0, time.Local)
	return NewInt(t.Unix())
}

// timeAddSecs - add seconds to timestamp
// Usage: timeAddSecs(timestamp, seconds) -> int
func builtinTimeAddSecs(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for timeAddSecs. got=%d, want=2", len(args))
	}

	var timestamp int64
	switch arg := args[0].(type) {
	case *Int:
		timestamp = arg.Value
	case *Float:
		timestamp = int64(arg.Value)
	default:
		return newError("first argument to 'timeAddSecs' must be INT or FLOAT, got %s", args[0].Type())
	}

	var seconds float64
	switch arg := args[1].(type) {
	case *Int:
		seconds = float64(arg.Value)
	case *Float:
		seconds = arg.Value
	default:
		return newError("second argument to 'timeAddSecs' must be INT or FLOAT, got %s", args[1].Type())
	}

	t := time.Unix(timestamp, 0)
	newTime := t.Add(time.Duration(seconds * float64(time.Second)))
	return NewInt(newTime.Unix())
}

// timeAddDate - add years, months, days to timestamp
// Usage: timeAddDate(timestamp, years, months, days) -> int
func builtinTimeAddDate(args ...Object) Object {
	if len(args) != 4 {
		return newError("wrong number of arguments for timeAddDate. got=%d, want=4", len(args))
	}

	var timestamp int64
	switch arg := args[0].(type) {
	case *Int:
		timestamp = arg.Value
	case *Float:
		timestamp = int64(arg.Value)
	default:
		return newError("first argument to 'timeAddDate' must be INT or FLOAT, got %s", args[0].Type())
	}

	years, ok := args[1].(*Int)
	if !ok {
		return newError("second argument to 'timeAddDate' must be INT, got %s", args[1].Type())
	}

	months, ok := args[2].(*Int)
	if !ok {
		return newError("third argument to 'timeAddDate' must be INT, got %s", args[2].Type())
	}

	days, ok := args[3].(*Int)
	if !ok {
		return newError("fourth argument to 'timeAddDate' must be INT, got %s", args[3].Type())
	}

	t := time.Unix(timestamp, 0)
	newTime := t.AddDate(int(years.Value), int(months.Value), int(days.Value))
	return NewInt(newTime.Unix())
}

// timeBefore - check if time1 is before time2
// Usage: timeBefore(timestamp1, timestamp2) -> bool
func builtinTimeBefore(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for timeBefore. got=%d, want=2", len(args))
	}

	var t1, t2 int64
	switch arg := args[0].(type) {
	case *Int:
		t1 = arg.Value
	case *Float:
		t1 = int64(arg.Value)
	default:
		return newError("first argument to 'timeBefore' must be INT or FLOAT, got %s", args[0].Type())
	}

	switch arg := args[1].(type) {
	case *Int:
		t2 = arg.Value
	case *Float:
		t2 = int64(arg.Value)
	default:
		return newError("second argument to 'timeBefore' must be INT or FLOAT, got %s", args[1].Type())
	}

	return &Bool{Value: t1 < t2}
}

// timeAfter - check if time1 is after time2
// Usage: timeAfter(timestamp1, timestamp2) -> bool
func builtinTimeAfter(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for timeAfter. got=%d, want=2", len(args))
	}

	var t1, t2 int64
	switch arg := args[0].(type) {
	case *Int:
		t1 = arg.Value
	case *Float:
		t1 = int64(arg.Value)
	default:
		return newError("first argument to 'timeAfter' must be INT or FLOAT, got %s", args[0].Type())
	}

	switch arg := args[1].(type) {
	case *Int:
		t2 = arg.Value
	case *Float:
		t2 = int64(arg.Value)
	default:
		return newError("second argument to 'timeAfter' must be INT or FLOAT, got %s", args[1].Type())
	}

	return &Bool{Value: t1 > t2}
}

// timeEqual - check if two timestamps are equal
// Usage: timeEqual(timestamp1, timestamp2) -> bool
func builtinTimeEqual(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for timeEqual. got=%d, want=2", len(args))
	}

	var t1, t2 int64
	switch arg := args[0].(type) {
	case *Int:
		t1 = arg.Value
	case *Float:
		t1 = int64(arg.Value)
	default:
		return newError("first argument to 'timeEqual' must be INT or FLOAT, got %s", args[0].Type())
	}

	switch arg := args[1].(type) {
	case *Int:
		t2 = arg.Value
	case *Float:
		t2 = int64(arg.Value)
	default:
		return newError("second argument to 'timeEqual' must be INT or FLOAT, got %s", args[1].Type())
	}

	return &Bool{Value: t1 == t2}
}

// timeDiff - get difference between two timestamps
// Usage: timeDiff(timestamp1, timestamp2) -> map with days, hours, minutes, seconds, totalSeconds
func builtinTimeDiff(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for timeDiff. got=%d, want=2", len(args))
	}

	var t1, t2 int64
	switch arg := args[0].(type) {
	case *Int:
		t1 = arg.Value
	case *Float:
		t1 = int64(arg.Value)
	default:
		return newError("first argument to 'timeDiff' must be INT or FLOAT, got %s", args[0].Type())
	}

	switch arg := args[1].(type) {
	case *Int:
		t2 = arg.Value
	case *Float:
		t2 = int64(arg.Value)
	default:
		return newError("second argument to 'timeDiff' must be INT or FLOAT, got %s", args[1].Type())
	}

	diff := t2 - t1
	if diff < 0 {
		diff = -diff
	}

	pairs := make(map[HashKey]MapPair)

	totalSeconds := diff
	days := totalSeconds / 86400
	remaining := totalSeconds % 86400
	hours := remaining / 3600
	remaining = remaining % 3600
	minutes := remaining / 60
	seconds := remaining % 60

	pairs[NewString("days").HashKey()] = MapPair{
		Key:   NewString("days"),
		Value: NewInt(days),
	}
	pairs[NewString("hours").HashKey()] = MapPair{
		Key:   NewString("hours"),
		Value: NewInt(hours),
	}
	pairs[NewString("minutes").HashKey()] = MapPair{
		Key:   NewString("minutes"),
		Value: NewInt(minutes),
	}
	pairs[NewString("seconds").HashKey()] = MapPair{
		Key:   NewString("seconds"),
		Value: NewInt(seconds),
	}
	pairs[NewString("totalSeconds").HashKey()] = MapPair{
		Key:   NewString("totalSeconds"),
		Value: NewInt(totalSeconds),
	}

	return NewMap(pairs)
}

// timeDiffSecs - get difference in seconds between two timestamps
// Usage: timeDiffSecs(timestamp1, timestamp2) -> int
func builtinTimeDiffSecs(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for timeDiffSecs. got=%d, want=2", len(args))
	}

	var t1, t2 int64
	switch arg := args[0].(type) {
	case *Int:
		t1 = arg.Value
	case *Float:
		t1 = int64(arg.Value)
	default:
		return newError("first argument to 'timeDiffSecs' must be INT or FLOAT, got %s", args[0].Type())
	}

	switch arg := args[1].(type) {
	case *Int:
		t2 = arg.Value
	case *Float:
		t2 = int64(arg.Value)
	default:
		return newError("second argument to 'timeDiffSecs' must be INT or FLOAT, got %s", args[1].Type())
	}

	return NewInt(t2 - t1)
}

// strToTime - parse string to timestamp
// Usage: strToTime(str) -> int (default format: "2006-01-02 15:04:05")
//
//	strToTime(str, format) -> int
func builtinStrToTime(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for strToTime. got=%d, want=1 or 2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'strToTime' must be STRING, got %s", args[0].Type())
	}

	format := "2006-01-02 15:04:05"
	if len(args) == 2 {
		f, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'strToTime' must be STRING, got %s", args[1].Type())
		}
		format = f.Value
	}

	t, err := time.ParseInLocation(format, str.Value, time.Local)
	if err != nil {
		return newError("strToTime parse error: %v", err)
	}

	return NewInt(t.Unix())
}

// parseTime - parse string to timestamp (alias for strToTime)
// Usage: parseTime(str, format) -> int
func builtinParseTime(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for parseTime. got=%d, want=2", len(args))
	}

	str, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'parseTime' must be STRING, got %s", args[0].Type())
	}

	format, ok := args[1].(*String)
	if !ok {
		return newError("second argument to 'parseTime' must be STRING, got %s", args[1].Type())
	}

	t, err := time.ParseInLocation(format.Value, str.Value, time.Local)
	if err != nil {
		return newError("parseTime error: %v", err)
	}

	return NewInt(t.Unix())
}

// isTime - check if value is a valid timestamp
// Usage: isTime(value) -> bool
func builtinIsTime(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for isTime. got=%d, want=1", len(args))
	}

	switch args[0].(type) {
	case *Int:
		// Any integer could be a timestamp
		return TRUE
	case *Float:
		// Float could also be a timestamp
		return TRUE
	default:
		return FALSE
	}
}

// Helper function to create time info string
func formatDuration(seconds int64) string {
	days := seconds / 86400
	remaining := seconds % 86400
	hours := remaining / 3600
	remaining = remaining % 3600
	minutes := remaining / 60
	secs := remaining % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
