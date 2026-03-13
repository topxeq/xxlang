// pkg/compiler/optimizer_test.go
// Tests for bytecode optimizer
package compiler

import "testing"

func TestOptimizer_FoldConstants(t *testing.T) {
    // Test: push const 1; push const 2; add should become push const 3
    // TODO: implement test
    t.Skip("not implemented")
}

func TestOptimizer_PeepholeOptimization(t *testing.T) {
    // Test: push X; pop should be removed
    // TODO: implement test
    t.Skip("not implemented")
}
