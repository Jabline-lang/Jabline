package compiler

import (
	"jabline/pkg/code"
	"jabline/pkg/object"
	"testing"
)

func TestEliminateDeadCodeNoChange(t *testing.T) {
	ins := code.Make(code.OpConstant, 42)
	ins = append(ins, code.Make(code.OpReturnValue)...)
	result := EliminateDeadCode(ins)
	if len(result) != len(ins) {
		t.Errorf("expected length %d, got %d", len(ins), len(result))
	}
	for i, b := range ins {
		if result[i] != b {
			t.Errorf("byte %d: expected %d, got %d", i, b, result[i])
		}
	}
}

func TestEliminateDeadCodeAfterReturn(t *testing.T) {
	ins := code.Make(code.OpConstant, 1)
	ins = append(ins, code.Make(code.OpReturnValue)...)
	ins = append(ins, code.Make(code.OpConstant, 2)...)
	ins = append(ins, code.Make(code.OpPop)...)
	result := EliminateDeadCode(ins)
	if len(result) != 4 {
		t.Fatalf("expected length 4, got %d:\n%v", len(result), result)
	}
	if result[0] != byte(code.OpConstant) {
		t.Errorf("expected OpConstant, got %d", result[0])
	}
	if result[3] != byte(code.OpReturnValue) {
		t.Errorf("expected OpReturnValue, got %d", result[3])
	}
}

func TestEliminateDeadCodeAfterJump(t *testing.T) {
	ins := code.Make(code.OpJump, 7)
	ins = append(ins, code.Make(code.OpConstant, 99)...)
	ins = append(ins, code.Make(code.OpPop)...)
	ins = append(ins, code.Make(code.OpConstant, 42)...)
	ins = append(ins, code.Make(code.OpReturnValue)...)
	result := EliminateDeadCode(ins)
	if len(result) != 7 {
		t.Fatalf("expected length 7, got %d:\n%v", len(result), result)
	}
	relocated := int(code.ReadUint16(result[1:]))
	if relocated != 3 {
		t.Errorf("expected jump target 3, got %d", relocated)
	}
	if code.Opcode(result[3]) != code.OpConstant {
		t.Errorf("expected OpConstant at 3, got %d", result[3])
	}
	if code.Opcode(result[6]) != code.OpReturnValue {
		t.Errorf("expected OpReturnValue at 6, got %d", result[6])
	}
}

func TestEliminateDeadCodeWithTry(t *testing.T) {
	ins := code.Make(code.OpTry, 10, 17)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpReturnValue)...)
	ins = append(ins, code.Make(code.OpPop)...)
	ins = append(ins, code.Make(code.OpConstant, 2)...)
	ins = append(ins, code.Make(code.OpReturnValue)...)
	ins = append(ins, code.Make(code.OpConstant, 3)...)
	ins = append(ins, code.Make(code.OpReturnValue)...)

	result := EliminateDeadCode(ins)
	if len(result) != 14 {
		t.Fatalf("expected length 14, got %d:\n%v", len(result), result)
	}
	if code.Opcode(result[0]) != code.OpTry {
		t.Fatalf("expected OpTry at 0, got %d", result[0])
	}
	newTryBody := int(code.ReadUint16(result[1:]))
	newCatchBody := int(code.ReadUint16(result[3:]))
	if newTryBody != 9 {
		t.Errorf("expected try body at 9, got %d", newTryBody)
	}
	if newCatchBody != 13 {
		t.Errorf("expected catch body at 13, got %d", newCatchBody)
	}
}

func TestEliminateDeadCodeAfterThrow(t *testing.T) {
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpThrow)...)
	ins = append(ins, code.Make(code.OpPop)...)
	ins = append(ins, code.Make(code.OpReturnValue)...)
	result := EliminateDeadCode(ins)
	if len(result) != 4 {
		t.Fatalf("expected length 4, got %d:\n%v", len(result), result)
	}
}

func TestEliminateDeadCodeMultipleJumps(t *testing.T) {
	ins := code.Make(code.OpJumpNotTruthy, 7)
	ins = append(ins, code.Make(code.OpJump, 7)...)
	ins = append(ins, code.Make(code.OpPop)...)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpReturnValue)...)
	result := EliminateDeadCode(ins)
	if len(result) != 10 {
		t.Fatalf("expected length 10, got %d:\n%v", len(result), result)
	}
	t1 := int(code.ReadUint16(result[1:]))
	if t1 != 6 {
		t.Errorf("expected JumpNotTruthy target 6, got %d", t1)
	}
	t2 := int(code.ReadUint16(result[4:]))
	if t2 != 6 {
		t.Errorf("expected Jump target 6, got %d", t2)
	}
}

func TestEliminateDeadCodeEmpty(t *testing.T) {
	result := EliminateDeadCode(nil)
	if result != nil {
		t.Errorf("expected nil for empty input")
	}
	result = EliminateDeadCode(code.Instructions{})
	if len(result) != 0 {
		t.Errorf("expected empty for empty input")
	}
}

// ── OptimizeBytecode tests ──────────────────────────────────────────────────

func TestOptimizeBytecodeEmpty(t *testing.T) {
	ins, consts := OptimizeBytecode(nil, nil)
	if ins != nil {
		t.Errorf("expected nil instructions, got %v", ins)
	}
	if consts != nil {
		t.Errorf("expected nil constants, got %v", consts)
	}

	ins, consts = OptimizeBytecode(code.Instructions{}, []object.Object{})
	if len(ins) != 0 {
		t.Errorf("expected empty instructions, got %v", ins)
	}
	if len(consts) != 0 {
		t.Errorf("expected empty constants, got %v", consts)
	}
}

func TestOptimizeBytecodeConstantFoldAdd(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 2},
		&object.Integer{Value: 3},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpAdd)...)

	result, newConsts := OptimizeBytecode(ins, constants)

	if len(newConsts) != 1 {
		t.Fatalf("expected 1 constant, got %d", len(newConsts))
	}
	intVal, ok := newConsts[0].(*object.Integer)
	if !ok {
		t.Fatalf("expected Integer constant, got %T", newConsts[0])
	}
	if intVal.Value != 5 {
		t.Errorf("expected constant value 5, got %d", intVal.Value)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 bytes, got %d:\n%s", len(result), result)
	}
	if code.Opcode(result[0]) != code.OpConstant {
		t.Errorf("expected OpConstant, got %d", result[0])
	}
	idx := int(code.ReadUint16(result[1:]))
	if idx != 0 {
		t.Errorf("expected constant index 0, got %d", idx)
	}
}

func TestOptimizeBytecodeConstantFoldSub(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 10},
		&object.Integer{Value: 3},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpSub)...)

	result, newConsts := OptimizeBytecode(ins, constants)

	if len(newConsts) != 1 || newConsts[0].(*object.Integer).Value != 7 {
		t.Fatalf("expected 1 constant with value 7, got %v", newConsts)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(result))
	}
}

func TestOptimizeBytecodeConstantFoldMul(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 4},
		&object.Integer{Value: 5},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpMul)...)

	_, newConsts := OptimizeBytecode(ins, constants)
	if len(newConsts) != 1 || newConsts[0].(*object.Integer).Value != 20 {
		t.Fatalf("expected constant 20, got %v", newConsts)
	}
}

func TestOptimizeBytecodeConstantFoldDiv(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 20},
		&object.Integer{Value: 4},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpDiv)...)

	_, newConsts := OptimizeBytecode(ins, constants)
	if len(newConsts) != 1 || newConsts[0].(*object.Integer).Value != 5 {
		t.Fatalf("expected constant 5, got %v", newConsts)
	}
}

func TestOptimizeBytecodeConstantFoldMod(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 10},
		&object.Integer{Value: 3},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpMod)...)

	_, newConsts := OptimizeBytecode(ins, constants)
	if len(newConsts) != 1 || newConsts[0].(*object.Integer).Value != 1 {
		t.Fatalf("expected constant 1, got %v", newConsts)
	}
}

func TestOptimizeBytecodeConstantFoldEqual(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 5},
		&object.Integer{Value: 5},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpEqual)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 1 || code.Opcode(result[0]) != code.OpTrue {
		t.Errorf("expected OpTrue, got %d", result[0])
	}

	// Not equal
	constants2 := []object.Object{
		&object.Integer{Value: 5},
		&object.Integer{Value: 3},
	}
	ins2 := code.Make(code.OpConstant, 0)
	ins2 = append(ins2, code.Make(code.OpConstant, 1)...)
	ins2 = append(ins2, code.Make(code.OpEqual)...)

	result2, _ := OptimizeBytecode(ins2, constants2)
	if len(result2) != 1 || code.Opcode(result2[0]) != code.OpFalse {
		t.Errorf("expected OpFalse, got %d", result2[0])
	}
}

func TestOptimizeBytecodeConstantFoldNotEqual(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 5},
		&object.Integer{Value: 3},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpNotEqual)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 1 || code.Opcode(result[0]) != code.OpTrue {
		t.Errorf("expected OpTrue, got %d", result[0])
	}
}

func TestOptimizeBytecodeConstantFoldGreaterThan(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 5},
		&object.Integer{Value: 3},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpGreaterThan)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 1 || code.Opcode(result[0]) != code.OpTrue {
		t.Errorf("expected OpTrue, got %d", result[0])
	}
}

func TestOptimizeBytecodeConstantFoldMinus(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 5},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpMinus)...)

	result, newConsts := OptimizeBytecode(ins, constants)
	if len(newConsts) != 1 {
		t.Fatalf("expected 1 constant, got %d", len(newConsts))
	}
	intVal, ok := newConsts[0].(*object.Integer)
	if !ok || intVal.Value != -5 {
		t.Fatalf("expected constant -5, got %v", newConsts[0])
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(result))
	}
}

func TestOptimizeBytecodeConstantFoldBang(t *testing.T) {
	// OpConstant true; OpBang → OpFalse
	constants := []object.Object{
		&object.Boolean{Value: true},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpBang)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 1 || code.Opcode(result[0]) != code.OpFalse {
		t.Errorf("expected OpFalse, got %d", result[0])
	}

	// OpConstant false; OpBang → OpTrue
	constants2 := []object.Object{
		&object.Boolean{Value: false},
	}
	ins2 := code.Make(code.OpConstant, 0)
	ins2 = append(ins2, code.Make(code.OpBang)...)

	result2, _ := OptimizeBytecode(ins2, constants2)
	if len(result2) != 1 || code.Opcode(result2[0]) != code.OpTrue {
		t.Errorf("expected OpTrue, got %d", result2[0])
	}
}

func TestOptimizeBytecodeRedundantGetSetLocal(t *testing.T) {
	constants := []object.Object{}
	// OpGetLocal 0; OpSetLocal 0 → removed
	ins := code.Make(code.OpGetLocal, 0)
	ins = append(ins, code.Make(code.OpSetLocal, 0)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 0 {
		t.Errorf("expected empty instructions, got %d bytes", len(result))
	}
}

func TestOptimizeBytecodeRedundantGetSetGlobal(t *testing.T) {
	constants := []object.Object{}
	ins := code.Make(code.OpGetGlobal, 5)
	ins = append(ins, code.Make(code.OpSetGlobal, 5)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 0 {
		t.Errorf("expected empty instructions, got %d bytes", len(result))
	}
}

func TestOptimizeBytecodeRedundantConstPop(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 42},
	}
	// OpConstant 0; OpPop → removed (constant not referenced after elimination)
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpPop)...)

	result, newConsts := OptimizeBytecode(ins, constants)
	if len(result) != 0 {
		t.Errorf("expected empty instructions, got %d bytes:\n%s", len(result), result)
	}
	if len(newConsts) != 0 {
		t.Errorf("expected no constants remaining, got %d", len(newConsts))
	}
}

func TestOptimizeBytecodePatternIncLocal(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 1},
	}
	// OpGetLocal 0; OpConstant 0; OpAdd; OpSetLocal 0 → OpIncLocal 0
	ins := code.Make(code.OpGetLocal, 0)
	ins = append(ins, code.Make(code.OpConstant, 0)...)
	ins = append(ins, code.Make(code.OpAdd)...)
	ins = append(ins, code.Make(code.OpSetLocal, 0)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 2 {
		t.Fatalf("expected 2 bytes (OpIncLocal), got %d:\n%s", len(result), result)
	}
	if code.Opcode(result[0]) != code.OpIncLocal {
		t.Errorf("expected OpIncLocal, got %d", result[0])
	}
	if result[1] != 0 {
		t.Errorf("expected local index 0, got %d", result[1])
	}
}

func TestOptimizeBytecodePatternDecLocal(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 1},
	}
	// OpGetLocal 0; OpConstant 0; OpSub; OpSetLocal 0 → OpDecLocal 0
	ins := code.Make(code.OpGetLocal, 0)
	ins = append(ins, code.Make(code.OpConstant, 0)...)
	ins = append(ins, code.Make(code.OpSub)...)
	ins = append(ins, code.Make(code.OpSetLocal, 0)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 2 {
		t.Fatalf("expected 2 bytes (OpDecLocal), got %d:\n%s", len(result), result)
	}
	if code.Opcode(result[0]) != code.OpDecLocal {
		t.Errorf("expected OpDecLocal, got %d", result[0])
	}
	if result[1] != 0 {
		t.Errorf("expected local index 0, got %d", result[1])
	}
}

func TestOptimizeBytecodePreservesReturnValue(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 42},
	}
	// OpConstant 0; OpReturnValue → should not be optimized away
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpReturnValue)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 4 {
		t.Errorf("expected 4 bytes (OpConstant+OpReturnValue), got %d", len(result))
	}
}

func TestOptimizeBytecodeConstantFoldBooleanEqual(t *testing.T) {
	constants := []object.Object{
		&object.Boolean{Value: true},
		&object.Boolean{Value: true},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpEqual)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 1 || code.Opcode(result[0]) != code.OpTrue {
		t.Errorf("expected OpTrue for true==true, got %d", result[0])
	}

	constants2 := []object.Object{
		&object.Boolean{Value: true},
		&object.Boolean{Value: false},
	}
	ins2 := code.Make(code.OpConstant, 0)
	ins2 = append(ins2, code.Make(code.OpConstant, 1)...)
	ins2 = append(ins2, code.Make(code.OpEqual)...)

	result2, _ := OptimizeBytecode(ins2, constants2)
	if len(result2) != 1 || code.Opcode(result2[0]) != code.OpFalse {
		t.Errorf("expected OpFalse for true==false, got %d", result2[0])
	}
}

func TestOptimizeBytecodeConstantFoldMultiple(t *testing.T) {
	// Test that multiple folds happen: (2+3)+(4+5) → OpConst 5; OpConst 9; OpAdd → OpConst 14
	constants := []object.Object{
		&object.Integer{Value: 2},
		&object.Integer{Value: 3},
		&object.Integer{Value: 4},
		&object.Integer{Value: 5},
	}
	ins := code.Make(code.OpConstant, 0) // 2
	ins = append(ins, code.Make(code.OpConstant, 1)...) // 3
	ins = append(ins, code.Make(code.OpAdd)...)          // → 5
	ins = append(ins, code.Make(code.OpConstant, 2)...) // 4
	ins = append(ins, code.Make(code.OpConstant, 3)...) // 5
	ins = append(ins, code.Make(code.OpAdd)...)          // → 9
	ins = append(ins, code.Make(code.OpAdd)...)          // 5+9 = 14

	result, newConsts := OptimizeBytecode(ins, constants)

	// All folds should reduce to a single OpConstant 14
	if len(newConsts) != 1 {
		t.Fatalf("expected 1 constant (14), got %d: %v", len(newConsts), newConsts)
	}
	intVal, ok := newConsts[0].(*object.Integer)
	if !ok || intVal.Value != 14 {
		t.Fatalf("expected constant 14, got %v", newConsts[0])
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 bytes (OpConstant 0), got %d:\n%s", len(result), result)
	}
}

func TestOptimizeBytecodeJumpToNext(t *testing.T) {
	constants := []object.Object{}
	// OpJump to next instruction → remove the jump
	// OpJump at 0, target=3 (the +3 means the instruction after the jump)
	ins := code.Make(code.OpJump, 3)
	ins = append(ins, code.Make(code.OpReturn)...)

	result, _ := OptimizeBytecode(ins, constants)
	if len(result) != 1 || code.Opcode(result[0]) != code.OpReturn {
		t.Errorf("expected just OpReturn, got %s", result.String())
	}
}



func TestOptimizeBytecodeCombined(t *testing.T) {
	// Test combined optimizations: OpConstant 5; OpConstant 3; OpSub; OpPop → eliminated
	constants := []object.Object{
		&object.Integer{Value: 5},
		&object.Integer{Value: 3},
	}
	ins := code.Make(code.OpConstant, 0)
	ins = append(ins, code.Make(code.OpConstant, 1)...)
	ins = append(ins, code.Make(code.OpSub)...) // folded to OpConstant 2
	ins = append(ins, code.Make(code.OpPop)...) // then OpConst+Pop eliminated

	result, newConsts := OptimizeBytecode(ins, constants)
	if len(result) != 0 {
		t.Errorf("expected empty after full optimization, got %d bytes:\n%s", len(result), result)
	}
	if len(newConsts) != 0 {
		t.Errorf("expected no constants, got %d", len(newConsts))
	}
}
