// pkg/compiler/options_test.go
// Tests for optimization flags
package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOptimizations(t *testing.T) {
	flags := DefaultOptimizations()

	assert.True(t, flags.BytecodeOptimizer)
	assert.True(t, flags.InlineCache)
	assert.True(t, flags.ClosurePool)
	assert.True(t, flags.TypeSpecialization)
}
