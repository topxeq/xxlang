// pkg/stdlib/task_test.go
// Tests for task module functions
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestIsCronExprValid(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{"standard cron", "* * * * *", true},
		{"every 5 minutes", "*/5 * * * *", true},
		{"hourly", "0 * * * *", true},
		{"daily at midnight", "0 0 * * *", true},
		{"weekly on Sunday", "0 0 * * 0", true},
		{"specific time", "30 14 * * *", true},
		{"with seconds", "0 0 0 * * *", true},
		{"invalid expression", "invalid", false},
		{"too few fields", "* * *", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCronExprValid(objects.NewString(tt.expr))

			if b, ok := result.(*objects.Bool); ok {
				if b.Value != tt.expected {
					t.Errorf("isCronExprValid(%q) = %v, want %v", tt.expr, b.Value, tt.expected)
				}
			} else {
				t.Errorf("expected Bool, got %T", result)
			}
		})
	}
}

func TestIsCronExprDue(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		wantNoError bool
	}{
		{"every minute", "* * * * *", true},
		{"specific hour", "0 * * * *", true},
		{"invalid expression", "not valid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCronExprDue(objects.NewString(tt.expr))

			if tt.wantNoError {
				if _, ok := result.(*objects.Bool); !ok {
					if err, isErr := result.(*objects.Error); isErr {
						t.Errorf("unexpected error: %v", err.Message)
					} else {
						t.Errorf("expected Bool, got %T", result)
					}
				}
			} else {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected Error for invalid expression, got %T", result)
				}
			}
		})
	}
}

func TestRunAndStopTicker(t *testing.T) {
	callback := &objects.CompiledFunction{
		Instructions:  []byte{0x00},
		NumLocals:     0,
		NumParameters: 0,
	}

	t.Run("run ticker", func(t *testing.T) {
		result := RunTicker(objects.NewFloat(1.0), callback)

		if id, ok := result.(*objects.Int); ok {
			if id.Value <= 0 {
				t.Error("expected positive ticker ID")
			}

			stopResult := StopTicker(id)
			if b, ok := stopResult.(*objects.Bool); ok {
				if !b.Value {
					t.Error("expected stopTicker to return true")
				}
			}

			stopAgain := StopTicker(id)
			if b, ok := stopAgain.(*objects.Bool); ok {
				if b.Value {
					t.Error("expected stopTicker to return false for already stopped ticker")
				}
			}
		} else {
			t.Errorf("expected Int (ticker ID), got %T", result)
		}
	})

	t.Run("invalid interval", func(t *testing.T) {
		result := RunTicker(objects.NewInt(-1), callback)
		if _, ok := result.(*objects.Error); !ok {
			t.Error("expected error for negative interval")
		}
	})

	t.Run("invalid callback", func(t *testing.T) {
		result := RunTicker(objects.NewFloat(1.0), objects.NewString("not a function"))
		if _, ok := result.(*objects.Error); !ok {
			t.Error("expected error for non-function callback")
		}
	})
}
