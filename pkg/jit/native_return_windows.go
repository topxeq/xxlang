//go:build windows && amd64
// +build windows,amd64

package jit

import (
	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
)

// NativeReturnType describes how a native result should be boxed back into a VM value.
type NativeReturnType int

const (
	ReturnTypeUnknown NativeReturnType = iota
	ReturnTypeInt
	ReturnTypeBool
	ReturnTypeNull
)

// analyzeReturnType analyzes the bytecode to determine what type of value
// the function returns. It looks at the opcode immediately before each
// OpRegReturn instruction to infer the type.
func analyzeReturnType(code []byte) NativeReturnType {
	return analyzeReturnTypeWithConstants(code, nil)
}

// analyzeReturnTypeWithConstants extends analyzeReturnType to also consider
// OpRegCall/OpRegTailCall by looking up the called function in the constants
// pool and recursively analyzing its return type.
func analyzeReturnTypeWithConstants(code []byte, constants []objects.Object) NativeReturnType {
	lastOp := compiler.Opcode(0)
	haveLast := false
	firstType := ReturnTypeUnknown
	// Track the constIdx associated with each register (from OpRegLoadConst)
	regConstIdx := make(map[int]int)
	// Track the constIdx associated with each global (from OpRegStoreGlobal)
	globalConstIdx := make(map[int]int)

	for i := 0; i < len(code); {
		op := compiler.Opcode(code[i])

		// Track OpRegLoadConst → register → constIdx mapping
		if op == compiler.OpRegLoadConst && i+3 < len(code) {
			dst := int(code[i+1])
			constIdx := int(code[i+2])<<8 | int(code[i+3])
			regConstIdx[dst] = constIdx
		}

		// Track OpRegStoreGlobal: if the source register holds a known constIdx,
		// propagate it to the global index
		if op == compiler.OpRegStoreGlobal && i+3 < len(code) {
			src := int(code[i+1])
			globalIdx := int(code[i+2])<<8 | int(code[i+3])
			if constIdx, ok := regConstIdx[src]; ok {
				globalConstIdx[globalIdx] = constIdx
			}
		}

		// Track OpRegLoadGlobal: if the global holds a known constIdx,
		// propagate it to the destination register
		if op == compiler.OpRegLoadGlobal && i+3 < len(code) {
			dst := int(code[i+1])
			globalIdx := int(code[i+2])<<8 | int(code[i+3])
			if constIdx, ok := globalConstIdx[globalIdx]; ok {
				regConstIdx[dst] = constIdx
			}
		}

		if op == compiler.OpRegReturn && haveLast {
			// Skip unreachable OpRegReturn after OpRegReturn
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
			case compiler.OpRegCall, compiler.OpRegTailCall:
				// Try to resolve the called function's return type from constants
				thisType = ReturnTypeUnknown
				// For OpRegCall, the previous op before OpRegCall should be
				// the setup for the function register. We need the funcReg from OpRegCall.
				// But we don't have it here since OpRegCall is the "lastOp" not the current op.
				// Instead, we look at what was loaded into the register that OpRegCall uses.
				// This is complex — for now, default to Unknown and rely on the
				// enhanced path below.
			}

			if firstType == ReturnTypeUnknown {
				firstType = thisType
			} else if thisType != firstType {
				return ReturnTypeUnknown
			}
		}

		// Enhanced: when we see OpRegCall/OpRegTailCall, check if the called
		// function's return type can be determined from the constant pool.
		// OpRegCall format: OpRegCall funcReg numArgs
		// The funcReg should have been set by OpRegLoadConst earlier,
		// and the constIdx tells us which function in the constants pool.
		if (op == compiler.OpRegCall || op == compiler.OpRegTailCall) && i+2 < len(code) {
			funcReg := int(code[i+1])
			if constIdx, ok := regConstIdx[funcReg]; ok && constants != nil {
				if constIdx < len(constants) {
					fnObj := constants[constIdx]
					if fn, ok := fnObj.(*compiler.CompiledFunction); ok {
						calledType := analyzeReturnTypeWithConstants(fn.Instructions, constants)

						// Propagate the called function's return type as if
						// this OpRegCall produces a value of that type.
						// If this is the last instruction (no following OpRegReturn),
						// treat it as an implicit return.
						switch calledType {
						case ReturnTypeBool:
							lastOp = compiler.OpRegAnd
						case ReturnTypeInt:
							lastOp = compiler.OpRegAdd
						case ReturnTypeNull:
							lastOp = compiler.OpRegNull
						default:
							lastOp = op
						}
						haveLast = true

						// Check if this OpRegCall is at the end of bytecode
						// (no following OpRegReturn) — treat as implicit return
						def, err := compiler.Lookup(byte(op))
						if err != nil {
							break
						}
						width := 1
						for _, w := range def.OperandWidths {
							width += w
						}
						nextIP := i + width

						// If no OpRegReturn follows, treat OpRegCall as implicit return
						if nextIP >= len(code) || compiler.Opcode(code[nextIP]) != compiler.OpRegReturn {
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
							} else if thisType != firstType && thisType != ReturnTypeUnknown {
								return ReturnTypeUnknown
							}
						}

						i += width
						continue
					}
				}
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
