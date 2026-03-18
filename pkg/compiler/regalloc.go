// pkg/compiler/regalloc.go
package compiler

import (
	"sort"
)

// LiveInterval represents the lifetime of a variable in the code
type LiveInterval struct {
	Var    *Symbol    // The variable this interval represents
	Start  int        // Definition point (instruction index)
	End    int        // Last use point (instruction index)
	Reg    int        // Assigned register (-1 if not assigned, -2 if spilled)
	Spill  int        // Spill slot index if spilled
	Active bool       // Whether this interval is currently active
}

// RegAllocator performs linear scan register allocation
type RegAllocator struct {
	intervals   []*LiveInterval
	freeRegs    []int
	regCount    int
	spillCount  int
	active      []*LiveInterval // Currently active intervals
	firstFree   int             // First register available for allocation
}

// NewRegAllocator creates a new register allocator
func NewRegAllocator(numRegs int) *RegAllocator {
	freeRegs := make([]int, 0, numRegs)
	for i := FirstLocalRegister; i < numRegs; i++ {
		freeRegs = append(freeRegs, i)
	}

	return &RegAllocator{
		intervals: make([]*LiveInterval, 0),
		freeRegs:  freeRegs,
		regCount:  numRegs,
		active:    make([]*LiveInterval, 0),
		firstFree: FirstLocalRegister,
	}
}

// AddInterval adds a new live interval for a variable
func (ra *RegAllocator) AddInterval(v *Symbol, start, end int) *LiveInterval {
	interval := &LiveInterval{
		Var:    v,
		Start:  start,
		End:    end,
		Reg:    -1,
		Spill:  -1,
		Active: false,
	}
	ra.intervals = append(ra.intervals, interval)
	return interval
}

// Allocate performs linear scan register allocation
// Returns the number of spilled intervals
func (ra *RegAllocator) Allocate() int {
	// Sort intervals by start position
	sort.Slice(ra.intervals, func(i, j int) bool {
		return ra.intervals[i].Start < ra.intervals[j].Start
	})

	spilled := 0

	for _, interval := range ra.intervals {
		// Expire old intervals that are no longer active
		ra.expireOldIntervals(interval.Start)

		// Try to allocate a register
		if len(ra.freeRegs) == 0 {
			// No free registers, need to spill
			ra.spillAtInterval(interval)
			spilled++
		} else {
			// Allocate a free register
			reg := ra.freeRegs[0]
			ra.freeRegs = ra.freeRegs[1:]
			interval.Reg = reg
			ra.active = append(ra.active, interval)
			interval.Active = true
		}
	}

	return spilled
}

// expireOldIntervals removes intervals that are no longer active
func (ra *RegAllocator) expireOldIntervals(currentPos int) {
	newActive := make([]*LiveInterval, 0, len(ra.active))

	for _, interval := range ra.active {
		if interval.End >= currentPos {
			newActive = append(newActive, interval)
		} else {
			// This interval is no longer active, free its register
			if interval.Reg >= 0 {
				ra.freeRegs = append(ra.freeRegs, interval.Reg)
			}
			interval.Active = false
		}
	}

	ra.active = newActive
}

// spillAtInterval handles spilling when no registers are available
func (ra *RegAllocator) spillAtInterval(interval *LiveInterval) {
	// Find the interval with the furthest end point in active
	if len(ra.active) == 0 {
		// No active intervals, spill current
		interval.Reg = -2
		interval.Spill = ra.spillCount
		ra.spillCount++
		return
	}

	// Find the active interval with the furthest end
	spillCandidate := ra.active[0]
	spillIdx := 0
	for i, active := range ra.active {
		if active.End > spillCandidate.End {
			spillCandidate = active
			spillIdx = i
		}
	}

	// If the current interval ends before the spill candidate,
	// it's better to spill the current one
	if interval.End < spillCandidate.End {
		// Spill current interval
		interval.Reg = -2
		interval.Spill = ra.spillCount
		ra.spillCount++
	} else {
		// Spill the active interval and give its register to current
		spillCandidate.Reg = -2
		spillCandidate.Spill = ra.spillCount
		spillCandidate.Active = false
		ra.spillCount++

		// Remove from active
		ra.active = append(ra.active[:spillIdx], ra.active[spillIdx+1:]...)

		// Give the register to current interval
		interval.Reg = spillCandidate.Reg
		ra.active = append(ra.active, interval)
		interval.Active = true
	}
}

// GetRegister returns the register assigned to a variable, or -1 if not found
func (ra *RegAllocator) GetRegister(varName string) int {
	for _, interval := range ra.intervals {
		if interval.Var != nil && interval.Var.Name == varName {
			return interval.Reg
		}
	}
	return -1
}

// GetInterval returns the live interval for a variable
func (ra *RegAllocator) GetInterval(varName string) *LiveInterval {
	for _, interval := range ra.intervals {
		if interval.Var != nil && interval.Var.Name == varName {
			return interval
		}
	}
	return nil
}

// AllocateTemp allocates a temporary register
func (ra *RegAllocator) AllocateTemp() int {
	if len(ra.freeRegs) == 0 {
		return -1 // No free registers
	}
	reg := ra.freeRegs[0]
	ra.freeRegs = ra.freeRegs[1:]
	return reg
}

// FreeTemp frees a temporary register
func (ra *RegAllocator) FreeTemp(reg int) {
	if reg >= FirstLocalRegister && reg < ra.regCount {
		ra.freeRegs = append(ra.freeRegs, reg)
	}
}

// SpillCount returns the number of spill slots used
func (ra *RegAllocator) SpillCount() int {
	return ra.spillCount
}

// RegAllocStats holds statistics about register allocation
type RegAllocStats struct {
	TotalIntervals int
	AssignedRegs   int
	SpilledInts    int
	MaxActive      int
}

// Stats returns statistics about the allocation
func (ra *RegAllocator) Stats() RegAllocStats {
	stats := RegAllocStats{
		TotalIntervals: len(ra.intervals),
	}

	for _, interval := range ra.intervals {
		if interval.Reg >= 0 {
			stats.AssignedRegs++
		} else if interval.Reg == -2 {
			stats.SpilledInts++
		}
	}

	stats.MaxActive = len(ra.active)

	return stats
}

// Reset clears the allocator for reuse
func (ra *RegAllocator) Reset() {
	ra.intervals = make([]*LiveInterval, 0)
	ra.active = make([]*LiveInterval, 0)
	ra.spillCount = 0

	// Rebuild free register list
	ra.freeRegs = make([]int, 0, ra.regCount-FirstLocalRegister)
	for i := FirstLocalRegister; i < ra.regCount; i++ {
		ra.freeRegs = append(ra.freeRegs, i)
	}
}
