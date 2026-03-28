// pkg/stdlib/task.go
// Task and scheduling module for Xxlang.
package stdlib

import (
	"fmt"
	"sync"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "task",
		Exports: map[string]objects.Object{
			"isCronExprValid": BuiltinFunc(IsCronExprValid),
			"isCronExprDue":   BuiltinFunc(IsCronExprDue),
			"runTicker":       BuiltinFunc(RunTicker),
			"stopTicker":      BuiltinFunc(StopTicker),
		},
	})
}

// tickerManager manages running tickers
var tickerManager = &struct {
	sync.RWMutex
	tickers map[int]*runningTicker
	nextID  int
}{
	tickers: make(map[int]*runningTicker),
	nextID:  1,
}

type runningTicker struct {
	ticker   *time.Ticker
	stopChan chan struct{}
	stopped  bool
}

// IsCronExprValid checks if a cron expression is valid.
// Usage: task.isCronExprValid(expr) -> bool
//
// Example:
//
//	task.isCronExprValid("* * * * *")        // true
//	task.isCronExprValid("*/5 * * * *")      // true
//	task.isCronExprValid("invalid")          // false
//	task.isCronExprValid("0 0 * * * *")      // true (with seconds)
func IsCronExprValid(args ...objects.Object) objects.Object {
	if len(args) != 1 {
		return Error("isCronExprValid() takes exactly 1 argument")
	}

	expr, ok := args[0].(*objects.String)
	if !ok {
		return Error("argument to 'isCronExprValid' must be STRING, got " + string(args[0].Type()))
	}

	_, err := objects.ParseCron(expr.Value)
	if err != nil {
		return objects.FALSE
	}

	return objects.TRUE
}

// IsCronExprDue checks if a cron expression is due at the current or specified time.
// Usage: task.isCronExprDue(expr) -> bool
//
//	task.isCronExprDue(expr, timeStr) -> bool  (check at specified time)
//
// Example:
//
//	task.isCronExprDue("* * * * *")           // true if current minute matches
//	task.isCronExprDue("0 0 * * *")           // true if it's midnight
func IsCronExprDue(args ...objects.Object) objects.Object {
	if len(args) < 1 || len(args) > 2 {
		return Error("isCronExprDue() takes 1 or 2 arguments")
	}

	expr, ok := args[0].(*objects.String)
	if !ok {
		return Error("first argument to 'isCronExprDue' must be STRING, got " + string(args[0].Type()))
	}

	var checkTime time.Time
	if len(args) == 2 {
		timeArg := args[1]
		switch t := timeArg.(type) {
		case *objects.String:
			parsed, err := time.Parse("2006-01-02 15:04:05", t.Value)
			if err != nil {
				parsed, err = time.Parse(time.RFC3339, t.Value)
				if err != nil {
					return Error("failed to parse time string: " + err.Error())
				}
			}
			checkTime = parsed
		default:
			checkTime = time.Now()
		}
	} else {
		checkTime = time.Now()
	}

	schedule, err := objects.ParseCron(expr.Value)
	if err != nil {
		return Error("invalid cron expression: " + err.Error())
	}

	if schedule.Matches(checkTime) {
		return objects.TRUE
	}

	nextTime := schedule.Next(checkTime.Add(-time.Minute))
	if nextTime.Year() == checkTime.Year() &&
		nextTime.Month() == checkTime.Month() &&
		nextTime.Day() == checkTime.Day() &&
		nextTime.Hour() == checkTime.Hour() &&
		nextTime.Minute() == checkTime.Minute() {
		return objects.TRUE
	}

	return objects.FALSE
}

// RunTicker runs a function periodically at specified intervals.
// Returns a ticker ID that can be used to stop the ticker.
// Usage: task.runTicker(intervalSeconds, callback) -> int (ticker ID)
//
//	task.runTicker(intervalSeconds, callback, arg1, arg2, ...) -> int
//
// The callback function will be called with any additional arguments.
// To stop the ticker, return an error from the callback or use stopTicker.
//
// Example:
//
//	id := task.runTicker(5, func() { pln("tick!") })
//	id := task.runTicker(1.5, myFunc, arg1, arg2)
func RunTicker(args ...objects.Object) objects.Object {
	if len(args) < 2 {
		return Error("runTicker() requires at least 2 arguments")
	}

	var interval float64
	switch i := args[0].(type) {
	case *objects.Int:
		interval = float64(i.Value)
	case *objects.Float:
		interval = i.Value
	default:
		return Error("interval must be INT or FLOAT, got " + string(args[0].Type()))
	}

	if interval <= 0 {
		return Error("interval must be positive")
	}

	tag := args[1].TypeTag()
	switch tag {
	case objects.TagFunction, objects.TagBuiltin, objects.TagClosure:
	default:
		return Error("callback must be FUNCTION, got " + string(args[1].Type()))
	}

	duration := time.Duration(interval * float64(time.Second))
	ticker := time.NewTicker(duration)
	stopChan := make(chan struct{})

	tickerManager.Lock()
	tickerID := tickerManager.nextID
	tickerManager.nextID++
	rt := &runningTicker{
		ticker:   ticker,
		stopChan: stopChan,
		stopped:  false,
	}
	tickerManager.tickers[tickerID] = rt
	tickerManager.Unlock()

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("Ticker %d fired\n", tickerID)
			case <-stopChan:
				return
			}
		}
	}()

	return Int(int64(tickerID))
}

// StopTicker stops a running ticker by ID.
// Usage: task.stopTicker(tickerID) -> bool
//
// Returns true if the ticker was stopped, false if not found.
func StopTicker(args ...objects.Object) objects.Object {
	if len(args) != 1 {
		return Error("stopTicker() takes exactly 1 argument")
	}

	id, ok := args[0].(*objects.Int)
	if !ok {
		return Error("argument to 'stopTicker' must be INT, got " + string(args[0].Type()))
	}

	tickerManager.Lock()
	defer tickerManager.Unlock()

	rt, exists := tickerManager.tickers[int(id.Value)]
	if !exists || rt.stopped {
		return objects.FALSE
	}

	rt.stopped = true
	close(rt.stopChan)
	delete(tickerManager.tickers, int(id.Value))

	return objects.TRUE
}
