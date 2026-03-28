// pkg/objects/builtin_task.go
// Task and scheduling functions for Xxlang.
// Note: These functions have been moved to the 'task' stdlib module.
// This file retains the cron parser for use by the task module.
package objects

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronField represents a cron field with valid values
type CronField struct {
	values map[int]bool
}

// CronSchedule represents a parsed cron schedule
type CronSchedule struct {
	second     *CronField
	minute     *CronField
	hour       *CronField
	day        *CronField
	month      *CronField
	dayOfWeek  *CronField
	hasSeconds bool
}

// ParseCron parses a cron expression and returns a schedule
// This is exported for use by the task stdlib module
func ParseCron(expr string) (*CronSchedule, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 && len(fields) != 6 {
		return nil, fmt.Errorf("cron expression must have 5 or 6 fields, got %d", len(fields))
	}

	schedule := &CronSchedule{
		hasSeconds: len(fields) == 6,
	}

	var err error
	var minuteField, hourField, domField, monthField, dowField, secondField []string

	if schedule.hasSeconds {
		secondField = strings.Split(fields[0], ",")
		minuteField = strings.Split(fields[1], ",")
		hourField = strings.Split(fields[2], ",")
		domField = strings.Split(fields[3], ",")
		monthField = strings.Split(fields[4], ",")
		dowField = strings.Split(fields[5], ",")
	} else {
		minuteField = strings.Split(fields[0], ",")
		hourField = strings.Split(fields[1], ",")
		domField = strings.Split(fields[2], ",")
		monthField = strings.Split(fields[3], ",")
		dowField = strings.Split(fields[4], ",")
	}

	if schedule.hasSeconds {
		schedule.second, err = ParseCronField(secondField, 0, 59)
		if err != nil {
			return nil, err
		}
	}
	schedule.minute, err = ParseCronField(minuteField, 0, 59)
	if err != nil {
		return nil, err
	}
	schedule.hour, err = ParseCronField(hourField, 0, 23)
	if err != nil {
		return nil, err
	}
	schedule.day, err = ParseCronField(domField, 1, 31)
	if err != nil {
		return nil, err
	}
	schedule.month, err = ParseCronField(monthField, 1, 12)
	if err != nil {
		return nil, err
	}
	schedule.dayOfWeek, err = ParseCronField(dowField, 0, 6)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

// ParseCronField parses a single cron field
// This is exported for use by the task stdlib module
func ParseCronField(parts []string, min, max int) (*CronField, error) {
	cf := &CronField{values: make(map[int]bool)}

	for _, part := range parts {
		if strings.Contains(part, "/") {
			if err := ParseCronStep(cf, part, min, max); err != nil {
				return nil, err
			}
		} else if strings.Contains(part, "-") {
			if err := ParseCronRange(cf, part, min, max); err != nil {
				return nil, err
			}
		} else if part == "*" {
			for i := min; i <= max; i++ {
				cf.values[i] = true
			}
		} else {
			val, err := ParseCronValue(part, min, max)
			if err != nil {
				return nil, err
			}
			cf.values[val] = true
		}
	}

	return cf, nil
}

// ParseCronStep parses a cron step expression (e.g., "*/5" or "0-30/5")
func ParseCronStep(cf *CronField, part string, min, max int) error {
	components := strings.Split(part, "/")
	if len(components) != 2 {
		return fmt.Errorf("invalid step: %s", part)
	}

	var rangeMin, rangeMax int
	if components[0] == "*" {
		rangeMin, rangeMax = min, max
	} else {
		rangeParts := strings.Split(components[0], "-")
		if len(rangeParts) != 2 {
			return fmt.Errorf("invalid range in step: %s", components[0])
		}
		var err error
		rangeMin, err = ParseCronValue(rangeParts[0], min, max)
		if err != nil {
			return err
		}
		rangeMax, err = ParseCronValue(rangeParts[1], min, max)
		if err != nil {
			return err
		}
	}

	step, err := ParseCronValue(components[1], 1, max)
	if err != nil {
		return err
	}

	for i := rangeMin; i <= rangeMax; i += step {
		cf.values[i] = true
	}

	return nil
}

// ParseCronRange parses a cron range expression (e.g., "1-5")
func ParseCronRange(cf *CronField, part string, min, max int) error {
	rangeParts := strings.Split(part, "-")
	if len(rangeParts) != 2 {
		return fmt.Errorf("invalid range: %s", part)
	}

	start, err := ParseCronValue(rangeParts[0], min, max)
	if err != nil {
		return err
	}
	end, err := ParseCronValue(rangeParts[1], min, max)
	if err != nil {
		return err
	}

	for i := start; i <= end; i++ {
		cf.values[i] = true
	}

	return nil
}

// ParseCronValue parses a single cron value
func ParseCronValue(s string, min, max int) (int, error) {
	val := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid value: %s", s)
		}
		val = val*10 + int(c-'0')
	}
	if val < min || val > max {
		return 0, fmt.Errorf("value %d out of range [%d, %d]", val, min, max)
	}
	return val, nil
}

// Matches checks if the given time matches the cron schedule
func (s *CronSchedule) Matches(t time.Time) bool {
	if s.hasSeconds && !s.second.values[t.Second()] {
		return false
	}
	if !s.minute.values[t.Minute()] {
		return false
	}
	if !s.hour.values[t.Hour()] {
		return false
	}
	if !s.day.values[t.Day()] {
		return false
	}
	if !s.month.values[int(t.Month())] {
		return false
	}
	if !s.dayOfWeek.values[int(t.Weekday())] {
		return false
	}
	return true
}

// matches is an alias for Matches (for internal use)
func (s *CronSchedule) matches(t time.Time) bool {
	return s.Matches(t)
}

// Next returns the next time the cron schedule will fire after the given time
func (s *CronSchedule) Next(after time.Time) time.Time {
	t := after.Add(time.Second)

	for i := 0; i < 366*24*60*60; i++ {
		if s.matches(t) {
			return t
		}
		t = t.Add(time.Second)
	}

	return t
}

// String returns the string representation of the cron schedule
func (s *CronSchedule) String() string {
	var secondField, minuteField, hourField, domField, monthField, dowField string

	if s.hasSeconds {
		secondField = fieldToString(s.second, 0, 59)
	}
	minuteField = fieldToString(s.minute, 0, 59)
	hourField = fieldToString(s.hour, 0, 23)
	domField = fieldToString(s.day, 1, 31)
	monthField = fieldToString(s.month, 1, 12)
	dowField = fieldToString(s.dayOfWeek, 0, 6)

	if s.hasSeconds {
		return fmt.Sprintf("%s %s %s %s %s %s", secondField, minuteField, hourField, domField, monthField, dowField)
	}
	return fmt.Sprintf("%s %s %s %s %s", minuteField, hourField, domField, monthField, dowField)
}

// fieldToString converts a cron field to string representation
func fieldToString(cf *CronField, min, max int) string {
	if len(cf.values) == max-min+1 {
		return "*"
	}

	var parts []string
	for i := min; i <= max; i++ {
		if cf.values[i] {
			parts = append(parts, strconv.Itoa(i))
		}
	}

	return strings.Join(parts, ",")
}

// Helper function to validate cron expression fields
func validateCronFields(expr string) bool {
	fields := strings.Fields(expr)

	// Standard cron: 5 fields (minute, hour, day, month, weekday)
	// Extended cron: 6 fields (second, minute, hour, day, month, weekday)
	if len(fields) != 5 && len(fields) != 6 {
		return false
	}

	// Validate each field
	for _, field := range fields {
		if !isValidCronField(field) {
			return false
		}
	}

	return true
}

// isValidCronField validates a single cron field
func isValidCronField(field string) bool {
	if field == "*" {
		return true
	}

	// Handle step values like */5
	if strings.HasPrefix(field, "*/") {
		step := strings.TrimPrefix(field, "*/")
		return isNumeric(step)
	}

	// Handle ranges like 1-5
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) != 2 {
			return false
		}
		return isNumeric(parts[0]) && isNumeric(parts[1])
	}

	// Handle lists like 1,2,3
	if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		for _, p := range parts {
			if !isNumeric(p) {
				return false
			}
		}
		return true
	}

	// Simple number
	return isNumeric(field)
}

// isNumeric checks if a string is a valid number
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
