// pkg/objects/builtin_task_test.go
// Tests for task/cron parsing functions
package objects

import (
	"testing"
	"time"
)

func TestParseCron(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		wantError bool
		validate  func(t *testing.T, schedule *CronSchedule)
	}{
		{
			name:      "standard 5-field cron",
			expr:      "*/5 * * * *",
			wantError: false,
			validate: func(t *testing.T, s *CronSchedule) {
				if s.hasSeconds {
					t.Errorf("expected hasSeconds=false for 5-field cron")
				}
			},
		},
		{
			name:      "cron with seconds (6 fields)",
			expr:      "0 */5 * * * *",
			wantError: false,
			validate: func(t *testing.T, s *CronSchedule) {
				if !s.hasSeconds {
					t.Errorf("expected hasSeconds=true for 6-field cron")
				}
			},
		},
		{
			name:      "every minute",
			expr:      "* * * * *",
			wantError: false,
		},
		{
			name:      "every hour",
			expr:      "0 * * * *",
			wantError: false,
		},
		{
			name:      "daily at midnight",
			expr:      "0 0 * * *",
			wantError: false,
		},
		{
			name:      "weekly",
			expr:      "0 0 * * 0",
			wantError: false,
		},
		{
			name:      "monthly",
			expr:      "0 0 1 * *",
			wantError: false,
		},
		{
			name:      "specific time",
			expr:      "30 14 * * 1-5",
			wantError: false,
		},
		{
			name:      "valid expression with range and list",
			expr:      "0 9-17,19-21 * * 1-5",
			wantError: false,
		},
		{
			name:      "step values with range",
			expr:      "*/15 */2 * * *",
			wantError: false,
		},
		{
			name:      "too few fields",
			expr:      "* * * *",
			wantError: true,
		},
		{
			name:      "too many fields",
			expr:      "* * * * * * *",
			wantError: true,
		},
		{
			name:      "invalid step syntax",
			expr:      "*/ * * * *",
			wantError: true,
		},
		{
			name:      "empty expression",
			expr:      "",
			wantError: true,
		},
		{
			name:      "list of values",
			expr:      "0 0 1,15 * *",
			wantError: false,
		},
		{
			name:      "step values with range",
			expr:      "*/15 */2 * * *",
			wantError: false,
		},
		{
			name:      "too few fields",
			expr:      "* * * *",
			wantError: true,
		},
		{
			name:      "too many fields",
			expr:      "* * * * * * *",
			wantError: true,
		},
		{
			name:      "invalid field value - non-numeric",
			expr:      "a * * * *",
			wantError: true,
		},
		{
			name:      "invalid step syntax",
			expr:      "*/ * * * *",
			wantError: true,
		},
		{
			name:      "empty expression",
			expr:      "",
			wantError: true,
		},
		{
			name:      "wildcard in seconds with 5-field",
			expr:      "* * * * *",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCron(tt.expr)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("expected non-nil schedule")
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestCronSchedule_Next(t *testing.T) {
	// Note: Next() method is not shown in the snippet but should exist
	// This is a basic test for schedule parsing only
	schedule, err := ParseCron("*/5 * * * *")
	if err != nil {
		t.Fatalf("failed to parse cron: %v", err)
	}

	if schedule == nil {
		t.Errorf("expected non-nil schedule")
	}

	// Test field presence
	_ = schedule.second
	_ = schedule.minute
	_ = schedule.hour
	_ = schedule.day
	_ = schedule.month
	_ = schedule.dayOfWeek
}

func TestCronField(t *testing.T) {
	// Test CronField parsing (indirectly through ParseCron)
	expr := "1-5,10-15/2,20,25-30/3"
	// This would be parsed internally by ParseCron
	// We test the overall parsing here

	schedule, err := ParseCron(expr + " * * * *")
	if err != nil {
		t.Fatalf("failed to parse complex cron: %v", err)
	}

	if schedule == nil {
		t.Errorf("expected non-nil schedule for complex expression")
	}
}

func TestCronScheduleMatches(t *testing.T) {
	schedule, err := ParseCron("30 14 * * 1-5")
	if err != nil {
		t.Fatalf("failed to parse cron: %v", err)
	}

	tests := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{"match weekday 14:30", "2024-01-08T14:30:00", true},
		{"wrong minute", "2024-01-08T14:00:00", false},
		{"wrong hour", "2024-01-08T15:30:00", false},
		{"weekend", "2024-01-07T14:30:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse time manually for cross-platform compatibility
			year, month, day := 2024, 1, 8
			hour, min, sec := 14, 30, 0
			if tt.timeStr == "2024-01-08T14:00:00" {
				hour, min, sec = 14, 0, 0
			} else if tt.timeStr == "2024-01-08T15:30:00" {
				hour, min, sec = 15, 30, 0
			} else if tt.timeStr == "2024-01-07T14:30:00" {
				year, month, day = 2024, 1, 7
			}
			testTime := parseTimeSimple(year, time.Month(month), day, hour, min, sec)
			result := schedule.Matches(testTime)
			if result != tt.expected {
				t.Errorf("Matches(%s) = %v, want %v", tt.timeStr, result, tt.expected)
			}
		})
	}
}

func parseTimeSimple(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}

func TestCronScheduleNext(t *testing.T) {
	schedule, err := ParseCron("0 * * * *")
	if err != nil {
		t.Fatalf("failed to parse cron: %v", err)
	}

	after := time.Date(2024, 1, 8, 10, 30, 0, 0, time.UTC)
	next := schedule.Next(after)

	if next.Minute() != 0 {
		t.Errorf("expected minute 0, got %d", next.Minute())
	}
	if next.Hour() != 11 {
		t.Errorf("expected hour 11, got %d", next.Hour())
	}
}

func TestCronScheduleString(t *testing.T) {
	schedule, err := ParseCron("30 14 * * 1-5")
	if err != nil {
		t.Fatalf("failed to parse cron: %v", err)
	}

	result := schedule.String()
	if result == "" {
		t.Error("expected non-empty string representation")
	}
}

func TestCronScheduleStringWithSeconds(t *testing.T) {
	schedule, err := ParseCron("0 30 14 * * 1-5")
	if err != nil {
		t.Fatalf("failed to parse cron: %v", err)
	}

	result := schedule.String()
	if result == "" {
		t.Error("expected non-empty string representation")
	}
}

func TestValidateCronFields(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"* * * * *", true},
		{"0 * * * *", true},
		{"0 0 * * *", true},
		{"0 0 0 * *", true},
		{"0 0 0 0 *", true},
		{"0 0 0 0 0", true},
		{"*/5 * * * *", true},
		{"0 9-17 * * 1-5", true},
		{"0 0 1,15 * *", true},
		{"* * * *", false},
		{"* * * * * * *", false},
		{"", false},
		{"a * * * *", false},
		{"* * * * * *", true},
	}

	for _, tt := range tests {
		result := validateCronFields(tt.expr)
		if result != tt.expected {
			t.Errorf("validateCronFields(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestIsValidCronField(t *testing.T) {
	tests := []struct {
		field    string
		expected bool
	}{
		{"*", true},
		{"*/5", true},
		{"1-5", true},
		{"1,2,3", true},
		{"10", true},
		{"*/", false},
		{"a", false},
		{"1-", false},
		{"-5", false},
	}

	for _, tt := range tests {
		result := isValidCronField(tt.field)
		if result != tt.expected {
			t.Errorf("isValidCronField(%q) = %v, want %v", tt.field, result, tt.expected)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		s        string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"", false},
		{"abc", false},
		{"12a", false},
		{"-5", false},
	}

	for _, tt := range tests {
		result := isNumeric(tt.s)
		if result != tt.expected {
			t.Errorf("isNumeric(%q) = %v, want %v", tt.s, result, tt.expected)
		}
	}
}

func TestFieldToString(t *testing.T) {
	cf := &CronField{values: make(map[int]bool)}
	for i := 0; i < 60; i++ {
		cf.values[i] = true
	}
	result := fieldToString(cf, 0, 59)
	if result != "*" {
		t.Errorf("expected '*', got %q", result)
	}

	cf2 := &CronField{values: make(map[int]bool)}
	cf2.values[0] = true
	cf2.values[30] = true
	result = fieldToString(cf2, 0, 59)
	if result != "0,30" {
		t.Errorf("expected '0,30', got %q", result)
	}
}
