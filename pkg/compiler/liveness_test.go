// pkg/compiler/liveness_test.go
package compiler

import (
	"testing"
)

func TestLivenessAnalyzer_New(t *testing.T) {
	la := NewLivenessAnalyzer()
	if la == nil {
		t.Fatal("expected LivenessAnalyzer, got nil")
	}
	if la.intervals == nil {
		t.Error("expected intervals map to be initialized")
	}
}

func TestLivenessAnalyzer_GetIntervals(t *testing.T) {
	la := NewLivenessAnalyzer()

	intervals := la.GetIntervals()
	if len(intervals) != 0 {
		t.Errorf("expected 0 intervals, got %d", len(intervals))
	}

	la.intervals["x"] = &LiveInterval{Var: &Symbol{Name: "x"}, Start: 0, End: 10}
	la.intervals["y"] = &LiveInterval{Var: &Symbol{Name: "y"}, Start: 5, End: 15}

	intervals = la.GetIntervals()
	if len(intervals) != 2 {
		t.Errorf("expected 2 intervals, got %d", len(intervals))
	}
}

func TestLivenessAnalyzer_RecordDefinition(t *testing.T) {
	la := NewLivenessAnalyzer()

	sym := &Symbol{Name: "x", Index: 0}

	la.recordDefinition(sym, 5)

	if len(la.intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(la.intervals))
	}
	interval := la.intervals["x"]
	if interval.Start != 5 {
		t.Errorf("expected start 5, got %d", interval.Start)
	}
	if interval.End != 5 {
		t.Errorf("expected end 5, got %d", interval.End)
	}

	la.recordDefinition(sym, 2)

	if interval.Start != 2 {
		t.Errorf("expected start 2, got %d", interval.Start)
	}
}

func TestLivenessAnalyzer_RecordUse(t *testing.T) {
	la := NewLivenessAnalyzer()

	sym := &Symbol{Name: "x", Index: 0}

	la.recordUse(sym, 10)

	if len(la.intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(la.intervals))
	}
	interval := la.intervals["x"]
	if interval.End != 10 {
		t.Errorf("expected end 10, got %d", interval.End)
	}

	la.recordUse(sym, 20)

	if interval.End != 20 {
		t.Errorf("expected end 20, got %d", interval.End)
	}
}

func TestLivenessAnalyzer_RecordUseAfterDefinition(t *testing.T) {
	la := NewLivenessAnalyzer()

	sym := &Symbol{Name: "x", Index: 0}

	la.recordDefinition(sym, 5)
	la.recordUse(sym, 15)

	interval := la.intervals["x"]
	if interval.Start != 5 {
		t.Errorf("expected start 5, got %d", interval.Start)
	}
	if interval.End != 15 {
		t.Errorf("expected end 15, got %d", interval.End)
	}
}

func TestLiveInterval_Fields(t *testing.T) {
	sym := &Symbol{Name: "x", Index: 0}
	interval := &LiveInterval{
		Var:    sym,
		Start:  5,
		End:    15,
		Reg:    3,
		Spill:  1,
		Active: true,
	}

	if interval.Var.Name != "x" {
		t.Errorf("expected Var.Name 'x', got %q", interval.Var.Name)
	}
	if interval.Start != 5 {
		t.Errorf("expected Start 5, got %d", interval.Start)
	}
	if interval.End != 15 {
		t.Errorf("expected End 15, got %d", interval.End)
	}
	if interval.Reg != 3 {
		t.Errorf("expected Reg 3, got %d", interval.Reg)
	}
	if interval.Spill != 1 {
		t.Errorf("expected Spill 1, got %d", interval.Spill)
	}
	if !interval.Active {
		t.Error("expected Active true")
	}
}
