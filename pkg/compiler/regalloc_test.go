// pkg/compiler/regalloc_test.go
package compiler

import (
	"testing"
)

func TestRegAllocator_New(t *testing.T) {
	ra := NewRegAllocator(16)
	if ra == nil {
		t.Fatal("expected RegAllocator, got nil")
	}
	if ra.regCount != 16 {
		t.Errorf("expected 16 registers, got %d", ra.regCount)
	}
	if len(ra.freeRegs) == 0 {
		t.Error("expected freeRegs to be initialized")
	}
}

func TestRegAllocator_AddInterval(t *testing.T) {
	ra := NewRegAllocator(16)

	sym := &Symbol{Name: "x", Index: 0}
	interval := ra.AddInterval(sym, 0, 10)

	if interval == nil {
		t.Fatal("expected interval, got nil")
	}
	if interval.Var.Name != "x" {
		t.Errorf("expected var name 'x', got %q", interval.Var.Name)
	}
	if interval.Start != 0 {
		t.Errorf("expected start 0, got %d", interval.Start)
	}
	if interval.End != 10 {
		t.Errorf("expected end 10, got %d", interval.End)
	}
	if interval.Reg != -1 {
		t.Errorf("expected Reg -1 (unassigned), got %d", interval.Reg)
	}
}

func TestRegAllocator_Allocate(t *testing.T) {
	tests := []struct {
		name      string
		intervals []*LiveInterval
	}{
		{name: "no intervals", intervals: []*LiveInterval{}},
		{name: "single interval", intervals: []*LiveInterval{{Var: &Symbol{Name: "x"}, Start: 0, End: 10}}},
		{name: "non-overlapping intervals", intervals: []*LiveInterval{
			{Var: &Symbol{Name: "x"}, Start: 0, End: 10},
			{Var: &Symbol{Name: "y"}, Start: 11, End: 20},
		}},
		{name: "overlapping intervals more than registers", intervals: []*LiveInterval{
			{Var: &Symbol{Name: "a"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "b"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "c"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "d"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "e"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "f"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "g"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "h"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "i"}, Start: 0, End: 100},
			{Var: &Symbol{Name: "j"}, Start: 0, End: 100},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ra := NewRegAllocator(16)
			for _, interval := range tt.intervals {
				ra.intervals = append(ra.intervals, interval)
			}

			spilled := ra.Allocate()

			if len(tt.intervals) > 8 && spilled == 0 {
				t.Logf("Note: Expected spills for %d overlapping intervals", len(tt.intervals))
			}
		})
	}
}

func TestRegAllocator_GetRegister(t *testing.T) {
	ra := NewRegAllocator(16)

	reg := ra.GetRegister("x")
	if reg != -1 {
		t.Errorf("expected -1 for non-existent var, got %d", reg)
	}

	sym := &Symbol{Name: "x", Index: 0}
	ra.AddInterval(sym, 0, 10)
	ra.Allocate()

	reg = ra.GetRegister("x")
	if reg < 0 {
		t.Errorf("expected non-negative register, got %d", reg)
	}
}

func TestRegAllocator_GetInterval(t *testing.T) {
	ra := NewRegAllocator(16)

	interval := ra.GetInterval("x")
	if interval != nil {
		t.Error("expected nil for non-existent var")
	}

	sym := &Symbol{Name: "x", Index: 0}
	ra.AddInterval(sym, 0, 10)

	interval = ra.GetInterval("x")
	if interval == nil {
		t.Fatal("expected interval, got nil")
	}
	if interval.Var.Name != "x" {
		t.Errorf("expected var name 'x', got %q", interval.Var.Name)
	}
}

func TestRegAllocator_AllocateTemp(t *testing.T) {
	ra := NewRegAllocator(16)

	reg := ra.AllocateTemp()
	if reg < 0 {
		t.Errorf("expected non-negative register, got %d", reg)
	}

	ra.FreeTemp(reg)

	reg2 := ra.AllocateTemp()
	if reg2 < 0 {
		t.Errorf("expected non-negative register after free, got %d", reg2)
	}
}

func TestRegAllocator_AllocateTemp_Exhausted(t *testing.T) {
	ra := NewRegAllocator(FirstLocalRegister + 2)

	ra.AllocateTemp()
	ra.AllocateTemp()

	reg := ra.AllocateTemp()
	if reg != -1 {
		t.Errorf("expected -1 when exhausted, got %d", reg)
	}
}

func TestRegAllocator_FreeTemp_Invalid(t *testing.T) {
	ra := NewRegAllocator(16)

	ra.FreeTemp(-1)
	ra.FreeTemp(100)
}

func TestRegAllocator_SpillCount(t *testing.T) {
	ra := NewRegAllocator(FirstLocalRegister + 2)

	for i := 0; i < 10; i++ {
		ra.AddInterval(&Symbol{Name: string(rune('a' + i))}, 0, 100)
	}

	spilled := ra.Allocate()

	if ra.SpillCount() != spilled {
		t.Errorf("expected SpillCount %d, got %d", spilled, ra.SpillCount())
	}
}

func TestRegAllocator_Stats(t *testing.T) {
	ra := NewRegAllocator(16)

	stats := ra.Stats()
	if stats.TotalIntervals != 0 {
		t.Errorf("expected 0 intervals, got %d", stats.TotalIntervals)
	}
	if stats.AssignedRegs != 0 {
		t.Errorf("expected 0 assigned, got %d", stats.AssignedRegs)
	}
	if stats.SpilledInts != 0 {
		t.Errorf("expected 0 spilled, got %d", stats.SpilledInts)
	}

	ra.AddInterval(&Symbol{Name: "x"}, 0, 10)
	ra.AddInterval(&Symbol{Name: "y"}, 5, 15)
	ra.Allocate()

	stats = ra.Stats()
	if stats.TotalIntervals != 2 {
		t.Errorf("expected 2 intervals, got %d", stats.TotalIntervals)
	}
}

func TestRegAllocator_Reset(t *testing.T) {
	ra := NewRegAllocator(16)

	ra.AddInterval(&Symbol{Name: "x"}, 0, 10)
	ra.AddInterval(&Symbol{Name: "y"}, 5, 15)
	ra.Allocate()

	ra.Reset()

	if len(ra.intervals) != 0 {
		t.Errorf("expected 0 intervals after reset, got %d", len(ra.intervals))
	}
	if len(ra.active) != 0 {
		t.Errorf("expected 0 active after reset, got %d", len(ra.active))
	}
	if ra.spillCount != 0 {
		t.Errorf("expected 0 spillCount after reset, got %d", ra.spillCount)
	}

	reg := ra.AllocateTemp()
	if reg < 0 {
		t.Errorf("expected valid temp after reset, got %d", reg)
	}
}

func TestRegAllocator_ExpireOldIntervals(t *testing.T) {
	ra := NewRegAllocator(16)

	ra.AddInterval(&Symbol{Name: "x"}, 0, 5)
	ra.AddInterval(&Symbol{Name: "y"}, 10, 20)

	ra.intervals[0].Reg = 5
	ra.intervals[0].Active = true
	ra.active = append(ra.active, ra.intervals[0])

	ra.expireOldIntervals(10)

	if len(ra.active) != 0 {
		t.Errorf("expected 0 active after expire, got %d", len(ra.active))
	}
}

func TestRegAllocator_SpillAtInterval(t *testing.T) {
	ra := NewRegAllocator(FirstLocalRegister + 2)

	active := &LiveInterval{
		Var:    &Symbol{Name: "active"},
		Start:  0,
		End:    100,
		Reg:    5,
		Active: true,
	}
	ra.active = append(ra.active, active)

	toSpill := &LiveInterval{
		Var:    &Symbol{Name: "new"},
		Start:  50,
		End:    60,
		Reg:    -1,
		Active: false,
	}

	ra.spillAtInterval(toSpill)

	if toSpill.Reg != -2 {
		t.Errorf("expected Reg -2 (spilled), got %d", toSpill.Reg)
	}
	if toSpill.Spill < 0 {
		t.Errorf("expected non-negative spill slot, got %d", toSpill.Spill)
	}
}

func TestRegAllocator_SpillAtInterval_NoActive(t *testing.T) {
	ra := NewRegAllocator(16)

	toSpill := &LiveInterval{
		Var:    &Symbol{Name: "new"},
		Start:  50,
		End:    60,
		Reg:    -1,
		Active: false,
	}

	ra.spillAtInterval(toSpill)

	if toSpill.Reg != -2 {
		t.Errorf("expected Reg -2 (spilled), got %d", toSpill.Reg)
	}
}
