package stdlib

import "testing"

// TestGetNextTestNum verifies that getNextTestNum returns sequential values
// starting from 1. Tests run sequentially; no parallel execution.
func TestGetNextTestNum(t *testing.T) {
	if v := getNextTestNum(); v != 1 {
		t.Fatalf("expected getNextTestNum() to return 1, got %d", v)
	}
	if v := getNextTestNum(); v != 2 {
		t.Fatalf("expected getNextTestNum() to return 2, got %d", v)
	}
	if v := getNextTestNum(); v != 3 {
		t.Fatalf("expected getNextTestNum() to return 3, got %d", v)
	}
	if v := getNextTestNum(); v != 4 {
		t.Fatalf("expected getNextTestNum() to return 4, got %d", v)
	}
}

// TestFindFirstDiffIndex validates the behavior of findFirstDiffIndex on a
// variety of inputs, including ASCII and multi-byte UTF-8 sequences.
func TestFindFirstDiffIndex(t *testing.T) {
	type tc struct {
		a, b string
		want int
	}
	cases := []tc{
		{a: "hello", b: "hello", want: -1}, // equal
		{a: "hello", b: "world", want: 0},  // different at start
		{a: "hello", b: "hexxo", want: 2},  // different at middle
		{a: "hello", b: "hell", want: 4},   // different lengths, prefix match
		{a: "abc", b: "def", want: 0},      // different lengths, no match at start
		{a: "", b: "a", want: 0},           // empty vs non-empty
		{a: "", b: "", want: -1},           // both empty
		{a: "café", b: "cafè", want: 4},    // Unicode bytes differ at index 4 (UTF-8)
	}
	for i, c := range cases {
		got := findFirstDiffIndex(c.a, c.b)
		if got != c.want {
			t.Fatalf("case %d: findFirstDiffIndex(%q, %q) = %d; want %d", i, c.a, c.b, got, c.want)
		}
	}
}

// TestInitRegistration performs a lightweight sanity check that the package's
// init-time registration did not break the exported/unexported API surface.
func TestInitRegistration(t *testing.T) {
	// Sanity check: functions should be usable after init
	if v := getNextTestNum(); v <= 0 {
		t.Fatalf("init sanity: getNextTestNum() returned non-positive value %d", v)
	}
	if idx := findFirstDiffIndex("", "a"); idx != 0 {
		t.Fatalf("init sanity: unexpected diff index for empty vs 'a': %d", idx)
	}
}
