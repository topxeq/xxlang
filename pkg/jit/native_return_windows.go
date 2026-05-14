//go:build windows && amd64
// +build windows,amd64

package jit

import "github.com/topxeq/xxlang/pkg/compiler"

// NativeReturnType describes how a native result should be boxed back into a VM value.
type NativeReturnType int

const (
	ReturnTypeUnknown NativeReturnType = iota
	ReturnTypeInt
	ReturnTypeBool
	ReturnTypeNull
)

func analyzeReturnType(code []byte) NativeReturnType {
	lastOp := compiler.Opcode(0)
	haveLast := false
	firstType := ReturnTypeUnknown

	for i := 0; i < len(code); {
		op := compiler.Opcode(code[i])

		if op == compiler.OpRegReturn && haveLast {
			// If the previous opcode is also OpRegReturn, this return is
			// unreachable (a default fallthrough appended by the compiler).
			// Skip it so it doesn't poison the type analysis.
			if lastOp == compiler.OpRegReturn {
				def, err := compiler.Lookup(byte(op))
				if err != nil {
					break
				}
				width := 1
				for _, w := range def.OperandWidths {
					width += w
				}
				lastOp = op
				i += width
				continue
			}

			thisType := ReturnTypeUnknown
			switch lastOp {
			case compiler.OpRegTrue, compiler.OpRegFalse, compiler.OpRegAnd, compiler.OpRegOr, compiler.OpRegNot,
				compiler.OpRegEqual, compiler.OpRegNotEqual, compiler.OpRegLess, compiler.OpRegLessEqual,
				compiler.OpRegGreater, compiler.OpRegGreaterEqual:
				thisType = ReturnTypeBool
			case compiler.OpRegNull:
				thisType = ReturnTypeNull
			case compiler.OpRegAdd, compiler.OpRegSub, compiler.OpRegMul, compiler.OpRegDiv, compiler.OpRegMod,
				compiler.OpRegAddConst, compiler.OpRegSubConst, compiler.OpRegMulConst,
				compiler.OpRegLoadConst, compiler.OpRegMove, compiler.OpRegIncLocal, compiler.OpRegDecLocal,
				compiler.OpRegLoopCountAdd, compiler.OpRegAddLocalCheck:
				thisType = ReturnTypeInt
			}

			if firstType == ReturnTypeUnknown {
				firstType = thisType
			} else if thisType != firstType {
				// Inconsistent return types across different return paths
				return ReturnTypeUnknown
			}
		}

		def, err := compiler.Lookup(byte(op))
		if err != nil {
			break
		}

		lastOp = op
		haveLast = true

		width := 1
		for _, w := range def.OperandWidths {
			width += w
		}
		i += width
	}

	return firstType
}
