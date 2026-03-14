// pkg/stdlib/time.go
// Time utilities for the Xxlang standard library.
package stdlib

import (
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "std/time",
		Exports: map[string]objects.Object{
			// unix returns the current Unix timestamp in seconds
			"unix": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(time.Now().Unix())
			}),

			// unixMs returns the current Unix timestamp in milliseconds
			"unixMs": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(time.Now().UnixMilli())
			}),

			// unixNano returns the current Unix timestamp in nanoseconds
			"unixNano": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(time.Now().UnixNano())
			}),

			// now returns the current time as a map with year, month, day, hour, minute, second
			"now": BuiltinFunc(func(args ...objects.Object) objects.Object {
				now := time.Now()
				pairs := make(map[objects.HashKey]objects.MapPair)

				yearKey := String("year")
				pairs[yearKey.HashKey()] = objects.MapPair{
					Key:   yearKey,
					Value: Int(int64(now.Year())),
				}

				monthKey := String("month")
				pairs[monthKey.HashKey()] = objects.MapPair{
					Key:   monthKey,
					Value: Int(int64(now.Month())),
				}

				dayKey := String("day")
				pairs[dayKey.HashKey()] = objects.MapPair{
					Key:   dayKey,
					Value: Int(int64(now.Day())),
				}

				hourKey := String("hour")
				pairs[hourKey.HashKey()] = objects.MapPair{
					Key:   hourKey,
					Value: Int(int64(now.Hour())),
				}

				minuteKey := String("minute")
				pairs[minuteKey.HashKey()] = objects.MapPair{
					Key:   minuteKey,
					Value: Int(int64(now.Minute())),
				}

				secondKey := String("second")
				pairs[secondKey.HashKey()] = objects.MapPair{
					Key:   secondKey,
					Value: Int(int64(now.Second())),
				}

				nanoKey := String("nanosecond")
				pairs[nanoKey.HashKey()] = objects.MapPair{
					Key:   nanoKey,
					Value: Int(int64(now.Nanosecond())),
				}

				return &objects.Map{Pairs: pairs}
			}),

			// year returns the current year
			"year": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(time.Now().Year()))
			}),

			// month returns the current month (1-12)
			"month": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(time.Now().Month()))
			}),

			// day returns the current day of month
			"day": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(time.Now().Day()))
			}),

			// hour returns the current hour (0-23)
			"hour": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(time.Now().Hour()))
			}),

			// minute returns the current minute (0-59)
			"minute": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(time.Now().Minute()))
			}),

			// second returns the current second (0-59)
			"second": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(time.Now().Second()))
			}),

			// weekday returns the current weekday (0=Sunday, 1=Monday, ..., 6=Saturday)
			"weekday": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return Int(int64(time.Now().Weekday()))
			}),

			// sleep pauses execution for the specified duration in milliseconds
			"sleep": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sleep() takes exactly 1 argument (milliseconds)")
				}
				ms, ok := args[0].(*objects.Int)
				if !ok {
					return Error("sleep() requires an integer argument (milliseconds)")
				}
				time.Sleep(time.Duration(ms.Value) * time.Millisecond)
				return Null()
			}),

			// sleepSec pauses execution for the specified duration in seconds
			"sleepSec": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sleepSec() takes exactly 1 argument (seconds)")
				}
				sec, ok := args[0].(*objects.Int)
				if !ok {
					return Error("sleepSec() requires an integer argument (seconds)")
				}
				time.Sleep(time.Duration(sec.Value) * time.Second)
				return Null()
			}),

			// format formats the current time using Go's time format
			// Common layouts: "2006-01-02", "15:04:05", "2006-01-02 15:04:05"
			"format": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("format() takes at least 1 argument (layout)")
				}
				layout, ok := args[0].(*objects.String)
				if !ok {
					return Error("format() requires a string layout argument")
				}
				return String(time.Now().Format(layout.Value))
			}),

			// formatUnix formats a Unix timestamp using Go's time format
			"formatUnix": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("formatUnix() takes 2 arguments (timestamp, layout)")
				}
				ts, ok1 := args[0].(*objects.Int)
				layout, ok2 := args[1].(*objects.String)
				if !ok1 || !ok2 {
					return Error("formatUnix() requires an integer timestamp and string layout")
				}
				t := time.Unix(ts.Value, 0)
				return String(t.Format(layout.Value))
			}),

			// parse parses a time string and returns Unix timestamp
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("parse() takes exactly 2 arguments (layout, value)")
				}
				layout, ok1 := args[0].(*objects.String)
				value, ok2 := args[1].(*objects.String)
				if !ok1 || !ok2 {
					return Error("parse() requires two string arguments (layout, value)")
				}
				t, err := time.Parse(layout.Value, value.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Int(t.Unix())
			}),

			// since returns the duration in milliseconds since the given Unix timestamp
			"since": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("since() takes exactly 1 argument (Unix timestamp in milliseconds)")
				}
				startMs, ok := args[0].(*objects.Int)
				if !ok {
					return Error("since() requires an integer argument (Unix timestamp in milliseconds)")
				}
				elapsed := time.Since(time.UnixMilli(startMs.Value))
				return Int(elapsed.Milliseconds())
			}),

			// addDays adds days to current time and returns Unix timestamp
			"addDays": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("addDays() takes exactly 1 argument (days)")
				}
				days, ok := args[0].(*objects.Int)
				if !ok {
					return Error("addDays() requires an integer argument (days)")
				}
				result := time.Now().AddDate(0, 0, int(days.Value))
				return Int(result.Unix())
			}),

			// addMonths adds months to current time and returns Unix timestamp
			"addMonths": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("addMonths() takes exactly 1 argument (months)")
				}
				months, ok := args[0].(*objects.Int)
				if !ok {
					return Error("addMonths() requires an integer argument (months)")
				}
				result := time.Now().AddDate(0, int(months.Value), 0)
				return Int(result.Unix())
			}),

			// addYears adds years to current time and returns Unix timestamp
			"addYears": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("addYears() takes exactly 1 argument (years)")
				}
				years, ok := args[0].(*objects.Int)
				if !ok {
					return Error("addYears() requires an integer argument (years)")
				}
				result := time.Now().AddDate(int(years.Value), 0, 0)
				return Int(result.Unix())
			}),

			// isLeapYear checks if the given year is a leap year
			"isLeapYear": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isLeapYear() takes exactly 1 argument (year)")
				}
				year, ok := args[0].(*objects.Int)
				if !ok {
					return Error("isLeapYear() requires an integer argument (year)")
				}
				y := int(year.Value)
				isLeap := (y%4 == 0 && y%100 != 0) || (y%400 == 0)
				return Bool(isLeap)
			}),

			// daysInMonth returns the number of days in the given month
			"daysInMonth": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("daysInMonth() takes exactly 2 arguments (year, month)")
				}
				year, ok1 := args[0].(*objects.Int)
				month, ok2 := args[1].(*objects.Int)
				if !ok1 || !ok2 {
					return Error("daysInMonth() requires two integer arguments (year, month)")
				}
				// time.Date normalizes the day, so we use day 0 of the next month
				// to get the last day of the current month
				days := time.Date(int(year.Value), time.Month(month.Value+1), 0, 0, 0, 0, 0, time.UTC).Day()
				return Int(int64(days))
			}),
		},
	})
}
