package compiler

import (
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

// parsedInst represents a decoded instruction with its original position.
type parsedInst struct {
	opcode   code.Opcode
	operands []int
	width    int
	oldPos   int
	removed  bool // marked for removal
}

// OptimizeBytecode applies peephole optimizations to bytecode instructions.
// It performs constant folding, redundant instruction elimination, and
// pattern replacement. Returns the optimized instructions and updated constants.
func OptimizeBytecode(instructions code.Instructions, constants []object.Object) (code.Instructions, []object.Object) {
	if len(instructions) == 0 {
		return instructions, constants
	}

	// 1. Parse instructions into a mutable list
	insts := parseInstructions(instructions)

	// 2. Normalize OpConstant8 → OpConstant
	for i := range insts {
		if insts[i].opcode == code.OpConstant8 {
			insts[i].opcode = code.OpConstant
			insts[i].width = 3
		}
	}

	// 3. Constant Folding
	insts = foldConstants(insts, &constants)

	// 4. Redundant Instruction Elimination (marks instructions with removed=true)
	eliminateRedundant(insts)

	// 5. Pattern Replacement
	insts = replacePatterns(insts, constants)

	// 6. Rebuild constant pool: remove unreferenced constants and update indices.
	insts, constants = rebuildConstantPool(insts, constants)

	// 7. Assemble new bytecode and compute oldPos→newPos mapping.
	//    Removed instructions' oldPos map to the next alive byte position.
	newIns, oldToNew := assembleInstructions(insts)

	// 8. Relocate jump offsets in the new bytecode
	relocateJumps(newIns, oldToNew)

	// 9. Run dead code elimination (handles further compaction and jump relocation)
	newIns = EliminateDeadCode(newIns)

	return newIns, constants
}

func parseInstructions(instructions code.Instructions) []parsedInst {
	var insts []parsedInst
	ip := 0
	for ip < len(instructions) {
		def, err := code.Lookup(instructions[ip])
		if err != nil {
			insts = append(insts, parsedInst{
				opcode: code.Opcode(instructions[ip]),
				width:  1,
				oldPos: ip,
			})
			ip++
			continue
		}

		totalWidth := 1
		for _, w := range def.OperandWidths {
			totalWidth += w
		}

		op := code.Opcode(instructions[ip])
		operands, _ := code.ReadOperands(def, instructions[ip+1:])

		if op == code.OpSelect {
			numCases := int(instructions[ip+1])
			hasDefault := int(instructions[ip+2])
			totalWidth += numCases * 3
			if hasDefault != 0 {
				totalWidth += 2
			}
		}

		insts = append(insts, parsedInst{
			opcode:   op,
			operands: operands,
			width:    totalWidth,
			oldPos:   ip,
		})
		ip += totalWidth
	}
	return insts
}

func foldConstants(insts []parsedInst, constants *[]object.Object) []parsedInst {
	safeConst := func(idx int) object.Object {
		if idx >= 0 && idx < len(*constants) {
			return (*constants)[idx]
		}
		return nil
	}

	emitFolded := func(folded object.Object, oldPos int) parsedInst {
		idx := findConstant(*constants, folded)
		if idx == -1 {
			idx = len(*constants)
			*constants = append(*constants, folded)
		}
		if b, ok := folded.(*object.Boolean); ok {
			if b.Value {
				return parsedInst{opcode: code.OpTrue, width: 1, oldPos: oldPos}
			}
			return parsedInst{opcode: code.OpFalse, width: 1, oldPos: oldPos}
		}
		return parsedInst{opcode: code.OpConstant, operands: []int{idx}, width: 3, oldPos: oldPos}
	}

	// Run folding passes until no more folds can be applied
	for {
		var result []parsedInst
		changed := false

		for i := 0; i < len(insts); i++ {
			// Binary operation: OpConst a; OpConst b; OpArith
			if i+2 < len(insts) &&
				insts[i].opcode == code.OpConstant &&
				insts[i+1].opcode == code.OpConstant {

				leftIdx := insts[i].operands[0]
				rightIdx := insts[i+1].operands[0]
				op := insts[i+2].opcode

				left := safeConst(leftIdx)
				right := safeConst(rightIdx)
				if left != nil && right != nil {
					leftInt, leftOK := left.(*object.Integer)
					rightInt, rightOK := right.(*object.Integer)

					var folded object.Object

					if leftOK && rightOK {
						switch op {
						case code.OpAdd:
							folded = &object.Integer{Value: leftInt.Value + rightInt.Value}
						case code.OpSub:
							folded = &object.Integer{Value: leftInt.Value - rightInt.Value}
						case code.OpMul:
							folded = &object.Integer{Value: leftInt.Value * rightInt.Value}
						case code.OpDiv:
							if rightInt.Value != 0 {
								folded = &object.Integer{Value: leftInt.Value / rightInt.Value}
							}
						case code.OpMod:
							if rightInt.Value != 0 {
								folded = &object.Integer{Value: leftInt.Value % rightInt.Value}
							}
						case code.OpEqual:
							folded = &object.Boolean{Value: leftInt.Value == rightInt.Value}
						case code.OpNotEqual:
							folded = &object.Boolean{Value: leftInt.Value != rightInt.Value}
						case code.OpGreaterThan:
							folded = &object.Boolean{Value: leftInt.Value > rightInt.Value}
						}
					}

					if folded == nil {
						leftBool, leftBOK := left.(*object.Boolean)
						rightBool, rightBOK := right.(*object.Boolean)
						if leftBOK && rightBOK {
							switch op {
							case code.OpEqual:
								folded = &object.Boolean{Value: leftBool.Value == rightBool.Value}
							case code.OpNotEqual:
								folded = &object.Boolean{Value: leftBool.Value != rightBool.Value}
							}
						}
					}

					if folded != nil {
						result = append(result, emitFolded(folded, insts[i].oldPos))
						i += 2
						changed = true
						continue
					}
				}
			}

			// Unary operation: OpConst a; OpMinus/OpBang
			if i+1 < len(insts) && insts[i].opcode == code.OpConstant {
				constIdx := insts[i].operands[0]
				op := insts[i+1].opcode
				val := safeConst(constIdx)
				if val != nil {
					var folded object.Object
					if intVal, ok := val.(*object.Integer); ok && op == code.OpMinus {
						folded = &object.Integer{Value: -intVal.Value}
					}
					if boolVal, ok := val.(*object.Boolean); ok && op == code.OpBang {
						folded = &object.Boolean{Value: !boolVal.Value}
					}

					if folded != nil {
						result = append(result, emitFolded(folded, insts[i].oldPos))
						i += 1
						changed = true
						continue
					}
				}
			}

			result = append(result, insts[i])
		}

		insts = result
		if !changed {
			break
		}
	}

	return insts
}

func eliminateRedundant(insts []parsedInst) {
	for i := 0; i < len(insts); i++ {
		if insts[i].removed {
			continue
		}

		// OpGetLocal x; OpSetLocal x → remove both
		if i+1 < len(insts) &&
			!insts[i+1].removed &&
			insts[i].opcode == code.OpGetLocal &&
			insts[i+1].opcode == code.OpSetLocal &&
			insts[i].operands[0] == insts[i+1].operands[0] {
			insts[i].removed = true
			insts[i+1].removed = true
			continue
		}

		// OpGetGlobal x; OpSetGlobal x → remove both
		if i+1 < len(insts) &&
			!insts[i+1].removed &&
			insts[i].opcode == code.OpGetGlobal &&
			insts[i+1].opcode == code.OpSetGlobal &&
			insts[i].operands[0] == insts[i+1].operands[0] {
			insts[i].removed = true
			insts[i+1].removed = true
			continue
		}

		// OpConstant i; OpPop → remove both
		if i+1 < len(insts) &&
			!insts[i+1].removed &&
			insts[i].opcode == code.OpConstant &&
			insts[i+1].opcode == code.OpPop {
			insts[i].removed = true
			insts[i+1].removed = true
			continue
		}
	}
}

func replacePatterns(insts []parsedInst, constants []object.Object) []parsedInst {
	var result []parsedInst
	for i := 0; i < len(insts); i++ {
		if insts[i].removed {
			result = append(result, insts[i])
			continue
		}

		// OpGetLocal x; OpConstant 1; OpAdd; OpSetLocal x → OpIncLocal x
		if i+3 < len(insts) &&
			insts[i].opcode == code.OpGetLocal &&
			insts[i+1].opcode == code.OpConstant &&
			insts[i+2].opcode == code.OpAdd &&
			insts[i+3].opcode == code.OpSetLocal &&
			insts[i].operands[0] == insts[i+3].operands[0] {
			constIdx := insts[i+1].operands[0]
			if c, ok := constants[constIdx].(*object.Integer); ok && c.Value == 1 {
				result = append(result, parsedInst{
					opcode:   code.OpIncLocal,
					operands: []int{insts[i].operands[0]},
					width:    2,
					oldPos:   insts[i].oldPos,
				})
				i += 3
				continue
			}
		}

		// OpGetLocal x; OpConstant 1; OpSub; OpSetLocal x → OpDecLocal x
		if i+3 < len(insts) &&
			insts[i].opcode == code.OpGetLocal &&
			insts[i+1].opcode == code.OpConstant &&
			insts[i+2].opcode == code.OpSub &&
			insts[i+3].opcode == code.OpSetLocal &&
			insts[i].operands[0] == insts[i+3].operands[0] {
			constIdx := insts[i+1].operands[0]
			if c, ok := constants[constIdx].(*object.Integer); ok && c.Value == 1 {
				result = append(result, parsedInst{
					opcode:   code.OpDecLocal,
					operands: []int{insts[i].operands[0]},
					width:    2,
					oldPos:   insts[i].oldPos,
				})
				i += 3
				continue
			}
		}

		// OpJump L where L is the next instruction (jump to next) → remove
		if insts[i].opcode == code.OpJump {
			target := insts[i].operands[0]
			nextPos := insts[i].oldPos + insts[i].width
			if target == nextPos && i+1 < len(insts) && insts[i+1].oldPos == nextPos {
				continue
			}
		}

		result = append(result, insts[i])
	}
	return result
}

func rebuildConstantPool(insts []parsedInst, constants []object.Object) ([]parsedInst, []object.Object) {
	// Collect referenced constant indices from OpConstant and OpClosure instructions.
	referenced := make(map[int]bool)
	for _, in := range insts {
		if in.removed {
			continue
		}
		switch in.opcode {
		case code.OpConstant:
			referenced[in.operands[0]] = true
		case code.OpClosure:
			// OpClosure operands: [constant_index (2 bytes), free_count (1 byte)]
			referenced[in.operands[0]] = true
		}
	}

	// Build a mapping from old index → new index for referenced constants.
	oldToNew := make(map[int]int)
	var newConstants []object.Object
	for oldIdx, c := range constants {
		if referenced[oldIdx] {
			oldToNew[oldIdx] = len(newConstants)
			newConstants = append(newConstants, c)
		}
	}

	// If nothing changed, return as-is.
	if len(newConstants) == len(constants) {
		allMatch := true
		for oldIdx, newIdx := range oldToNew {
			if oldIdx != newIdx {
				allMatch = false
				break
			}
		}
		if allMatch {
			return insts, constants
		}
	}

	// Update OpConstant operands to use new indices.
	for i := range insts {
		if insts[i].opcode == code.OpConstant {
			insts[i].operands[0] = oldToNew[insts[i].operands[0]]
		}
	}

	return insts, newConstants
}

func assembleInstructions(insts []parsedInst) (code.Instructions, map[int]int) {
	var newIns code.Instructions
	oldToNew := make(map[int]int)

	// First pass: assemble only non-removed instructions and build old→new map
	for _, in := range insts {
		if in.removed {
			continue
		}
		oldToNew[in.oldPos] = len(newIns)
		newIns = append(newIns, code.Make(in.opcode, in.operands...)...)
	}

	// Second pass: for removed instructions, map their oldPos to the new position
	// of the nearest following non-removed instruction (or end of bytecode).
	// We iterate through insts, tracking the "next alive" position for each gap.
	nextAlive := -1
	// Walk backwards to compute next-alive mapping
	for i := len(insts) - 1; i >= 0; i-- {
		if !insts[i].removed {
			if pos, ok := oldToNew[insts[i].oldPos]; ok {
				nextAlive = pos
			}
		} else {
			if nextAlive >= 0 {
				oldToNew[insts[i].oldPos] = nextAlive
			} else {
				// All instructions after this are also dead; map to end
				oldToNew[insts[i].oldPos] = len(newIns)
			}
		}
	}

	return newIns, oldToNew
}

func relocateJumps(instructions code.Instructions, oldToNew map[int]int) {
	ip := 0
	for ip < len(instructions) {
		def, err := code.Lookup(instructions[ip])
		if err != nil {
			ip++
			continue
		}

		totalWidth := 1
		for _, w := range def.OperandWidths {
			totalWidth += w
		}

		op := code.Opcode(instructions[ip])
		switch op {
		case code.OpJump, code.OpJumpNotTruthy, code.OpJumpNotNull,
			code.OpJumpNotTruthyKeep, code.OpJumpTruthyKeep,
			code.OpJumpIfEqual, code.OpJumpIfTrue:
			oldTarget := int(code.ReadUint16(instructions[ip+1:]))
			if newTarget, ok := oldToNew[oldTarget]; ok {
				code.WriteUint16(instructions, ip+1, uint16(newTarget))
			}
		case code.OpTry:
			oldTryBody := int(code.ReadUint16(instructions[ip+1:]))
			if newTryBody, ok := oldToNew[oldTryBody]; ok {
				code.WriteUint16(instructions, ip+1, uint16(newTryBody))
			}
			oldCatchBody := int(code.ReadUint16(instructions[ip+3:]))
			if newCatchBody, ok := oldToNew[oldCatchBody]; ok {
				code.WriteUint16(instructions, ip+3, uint16(newCatchBody))
			}
		case code.OpSelect:
			numCases := int(instructions[ip+1])
			hasDefault := int(instructions[ip+2])
			off := ip + 3
			for i := 0; i < numCases; i++ {
				oldTarget := int(code.ReadUint16(instructions[off+1:]))
				if newTarget, ok := oldToNew[oldTarget]; ok {
					code.WriteUint16(instructions, off+1, uint16(newTarget))
				}
				off += 3
			}
			if hasDefault != 0 {
				oldTarget := int(code.ReadUint16(instructions[off:]))
				if newTarget, ok := oldToNew[oldTarget]; ok {
					code.WriteUint16(instructions, off, uint16(newTarget))
				}
			}
			totalWidth += numCases * 3
			if hasDefault != 0 {
				totalWidth += 2
			}
		}

		ip += totalWidth
	}
}

func findConstant(constants []object.Object, obj object.Object) int {
	for i, c := range constants {
		switch a := c.(type) {
		case *object.Integer:
			if b, ok := obj.(*object.Integer); ok && a.Value == b.Value {
				return i
			}
		case *object.Boolean:
			if b, ok := obj.(*object.Boolean); ok && a.Value == b.Value {
				return i
			}
		}
	}
	return -1
}

// EliminateDeadCode removes unreachable instructions after unconditional
// terminators (OpReturn, OpReturnValue, OpThrow, OpJump) and rewrites all jump
// offsets to match the compacted bytecode.
func EliminateDeadCode(instructions code.Instructions) code.Instructions {
	if len(instructions) == 0 {
		return instructions
	}

	// ── 1. Mark alive / dead ──────────────────────────────────────────────
	alive := make([]bool, len(instructions))
	jumpTargets := collectJumpTargets(instructions)

	ip := 0
	for ip < len(instructions) {
		def, err := code.Lookup(instructions[ip])
		if err != nil {
			alive[ip] = true
			ip++
			continue
		}

		totalWidth := 1
		for _, w := range def.OperandWidths {
			totalWidth += w
		}

		op := code.Opcode(instructions[ip])
		if op == code.OpSelect {
			numCases := int(instructions[ip+1])
			hasDefault := int(instructions[ip+2])
			totalWidth += numCases * 3
			if hasDefault != 0 {
				totalWidth += 2
			}
		}

		// Mark opcode and all operand bytes as alive
		for i := 0; i < totalWidth; i++ {
			if ip+i < len(instructions) {
				alive[ip+i] = true
			}
		}

		if isUnconditionalTerminator(op) {
			for next := ip + totalWidth; next < len(instructions); next++ {
				if jumpTargets[next] {
					break
				}
				alive[next] = false
			}
			nextLive := ip + totalWidth
			for nextLive < len(instructions) && !jumpTargets[nextLive] {
				nextLive++
			}
			ip = nextLive
			continue
		}

		ip += totalWidth
	}

	// ── 2. Build old-position → new-position map and compact ──────────────
	oldToNew := make(map[int]int)
	var result code.Instructions
	oldPos := 0
	for oldPos < len(instructions) {
		if !alive[oldPos] {
			// Skip this byte (part of dead instruction)
			oldPos++
			continue
		}
		def, err := code.Lookup(instructions[oldPos])
		if err != nil {
			oldToNew[oldPos] = len(result)
			result = append(result, instructions[oldPos])
			oldPos++
			continue
		}
		totalWidth := 1
		for _, w := range def.OperandWidths {
			totalWidth += w
		}
		op := code.Opcode(instructions[oldPos])
		if op == code.OpSelect {
			numCases := int(instructions[oldPos+1])
			hasDefault := int(instructions[oldPos+2])
			totalWidth += numCases * 3
			if hasDefault != 0 {
				totalWidth += 2
			}
		}
		oldToNew[oldPos] = len(result)
		for i := 0; i < totalWidth; i++ {
			result = append(result, instructions[oldPos+i])
		}
		oldPos += totalWidth
	}

	// ── 4. Relocate jump offsets ──────────────────────────────────────────
	newIP := 0
	for newIP < len(result) {
		def, err := code.Lookup(result[newIP])
		if err != nil {
			newIP++
			continue
		}

		totalWidth := 1
		for _, w := range def.OperandWidths {
			totalWidth += w
		}

		op := code.Opcode(result[newIP])
		switch op {
		case code.OpJump, code.OpJumpNotTruthy, code.OpJumpNotNull,
			code.OpJumpNotTruthyKeep, code.OpJumpTruthyKeep,
			code.OpJumpIfEqual, code.OpJumpIfTrue:
			oldTarget := int(code.ReadUint16(result[newIP+1:]))
			newTarget, ok := oldToNew[oldTarget]
			if !ok {
				// Target was in dead code — clamp to end
				newTarget = len(result)
			}
			code.WriteUint16(result, newIP+1, uint16(newTarget))
		case code.OpTry:
			oldTryBody := int(code.ReadUint16(result[newIP+1:]))
			oldCatchBody := int(code.ReadUint16(result[newIP+3:]))
			if newTryBody, ok := oldToNew[oldTryBody]; ok {
				code.WriteUint16(result, newIP+1, uint16(newTryBody))
			}
			if newCatchBody, ok := oldToNew[oldCatchBody]; ok {
				code.WriteUint16(result, newIP+3, uint16(newCatchBody))
			}
		case code.OpSelect:
			// OpSelect inline descriptors: flags(1) + offset(2) per case
			numCases := int(result[newIP+1])
			hasDefault := int(result[newIP+2])
			off := newIP + 3
			for i := 0; i < numCases; i++ {
				oldTarget := int(code.ReadUint16(result[off+1:]))
				if newTarget, ok := oldToNew[oldTarget]; ok {
					code.WriteUint16(result, off+1, uint16(newTarget))
				}
				off += 3 // flags(1) + offset(2)
			}
			if hasDefault != 0 {
				oldTarget := int(code.ReadUint16(result[off:]))
				if newTarget, ok := oldToNew[oldTarget]; ok {
					code.WriteUint16(result, off, uint16(newTarget))
				}
			}
			// totalWidth only covers opcode + operands; add inline descriptor size
			totalWidth += numCases*3
			if hasDefault != 0 {
				totalWidth += 2
			}
		}

		newIP += totalWidth
	}

	return result
}

func isUnconditionalTerminator(op code.Opcode) bool {
	return op == code.OpReturn || op == code.OpReturnValue || op == code.OpThrow || op == code.OpJump
}

func collectJumpTargets(instructions code.Instructions) map[int]bool {
	targets := make(map[int]bool)
	ip := 0
	for ip < len(instructions) {
		def, err := code.Lookup(instructions[ip])
		if err != nil {
			ip++
			continue
		}
		totalWidth := 1
		for _, w := range def.OperandWidths {
			totalWidth += w
		}

		op := code.Opcode(instructions[ip])
		switch op {
		case code.OpJump, code.OpJumpNotTruthy, code.OpJumpNotNull,
			code.OpJumpNotTruthyKeep, code.OpJumpTruthyKeep,
			code.OpJumpIfEqual, code.OpJumpIfTrue:
			if len(instructions) > ip+1 {
				offset := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				targets[offset] = true
			}
		case code.OpTry:
			if len(instructions) > ip+3 {
				tryBody := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				catchBody := int(instructions[ip+3])<<8 | int(instructions[ip+4])
				targets[tryBody] = true
				targets[catchBody] = true
			}
		case code.OpSelect:
			// OpSelect inline descriptors: flags(1) + offset(2) per case
			numCases := int(instructions[ip+1])
			hasDefault := int(instructions[ip+2])
			off := ip + 3
			for i := 0; i < numCases; i++ {
				offset := int(instructions[off+1])<<8 | int(instructions[off+2])
				targets[offset] = true
				off += 3
			}
			if hasDefault != 0 && len(instructions) > off+1 {
				offset := int(instructions[off])<<8 | int(instructions[off+1])
				targets[offset] = true
			}
			totalWidth += numCases*3
			if hasDefault != 0 {
				totalWidth += 2
			}
		}

		ip += totalWidth
	}
	return targets
}

// TryAlgebraicSimplify attempts to simplify infix expressions with constant
// operands using algebraic identities.
func TryAlgebraicSimplify(node *ast.InfixExpression) (ast.Node, bool) {
	leftConst, leftIsInt := isIntLiteral(node.Left)
	rightConst, rightIsInt := isIntLiteral(node.Right)
	_, leftIsFloat := isFloatLiteral(node.Left)
	_, rightIsFloat := isFloatLiteral(node.Right)

	switch node.Operator {
	case "+":
		if rightIsInt && rightConst == 0 {
			return node.Left, true
		}
		if leftIsInt && leftConst == 0 {
			return node.Right, true
		}
		if rightIsFloat && rightConst == 0.0 {
			return node.Left, true
		}
		if leftIsFloat && leftConst == 0.0 {
			return node.Right, true
		}

	case "-":
		if rightIsInt && rightConst == 0 {
			return node.Left, true
		}
		if rightIsFloat && rightConst == 0.0 {
			return node.Left, true
		}

	case "*":
		if rightIsInt && rightConst == 1 {
			return node.Left, true
		}
		if leftIsInt && leftConst == 1 {
			return node.Right, true
		}
		if (rightIsInt && rightConst == 0) || (leftIsInt && leftConst == 0) {
			return &ast.IntegerLiteral{Value: 0}, true
		}
		if rightIsFloat && rightConst == 1.0 {
			return node.Left, true
		}
		if leftIsFloat && leftConst == 1.0 {
			return node.Right, true
		}
		if rightIsFloat && rightConst == 0.0 {
			return &ast.FloatLiteral{Value: 0}, true
		}
		if leftIsFloat && leftConst == 0.0 {
			return &ast.FloatLiteral{Value: 0}, true
		}

	case "/":
		if rightIsInt && rightConst == 1 {
			return node.Left, true
		}
		if rightIsFloat && rightConst == 1.0 {
			return node.Left, true
		}
		if leftIsInt && leftConst == 0 && !(rightIsInt && rightConst == 0) {
			return &ast.IntegerLiteral{Value: 0}, true
		}
		if leftIsFloat && leftConst == 0.0 && !(rightIsFloat && rightConst == 0.0) {
			return &ast.FloatLiteral{Value: 0}, true
		}

	case "%":
		if rightIsInt && rightConst == 1 {
			return &ast.IntegerLiteral{Value: 0}, true
		}

	case "<<", ">>":
		if rightIsInt && rightConst == 0 {
			return node.Left, true
		}

	case "&":
		if (rightIsInt && rightConst == 0) || (leftIsInt && leftConst == 0) {
			return &ast.IntegerLiteral{Value: 0}, true
		}
		if rightIsInt && rightConst == -1 {
			return node.Left, true
		}
		if leftIsInt && leftConst == -1 {
			return node.Right, true
		}

	case "|":
		if rightIsInt && rightConst == 0 {
			return node.Left, true
		}
		if leftIsInt && leftConst == 0 {
			return node.Right, true
		}
		if (rightIsInt && rightConst == -1) || (leftIsInt && leftConst == -1) {
			return &ast.IntegerLiteral{Value: -1}, true
		}

	case "^":
		if rightIsInt && rightConst == 0 {
			return node.Left, true
		}
		if leftIsInt && leftConst == 0 {
			return node.Right, true
		}

	case "==":
		if node.Left.String() == node.Right.String() && !isMutable(node.Left) {
			return &ast.Boolean{Value: true}, true
		}

	case "!=":
		if node.Left.String() == node.Right.String() && !isMutable(node.Left) {
			return &ast.Boolean{Value: false}, true
		}
	}

	return nil, false
}

func isIntLiteral(n ast.Node) (int64, bool) {
	if lit, ok := n.(*ast.IntegerLiteral); ok {
		return lit.Value, true
	}
	return 0, false
}

func isFloatLiteral(n ast.Node) (float64, bool) {
	if lit, ok := n.(*ast.FloatLiteral); ok {
		return lit.Value, true
	}
	return 0, false
}

func isMutable(n ast.Node) bool {
	switch n.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.Boolean, *ast.Null:
		return false
	}
	return true
}
