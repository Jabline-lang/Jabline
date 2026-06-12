package compiler

import (
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object" // New import
	"jabline/pkg/symbol"
)

func (c *Compiler) compileLetStatement(node *ast.LetStatement) error {
	if node.Destructure != nil {
		return c.compileDestructuring(
			node.Destructure,
			node.Value,
			false, // not const
		)
	}

	if err := c.Compile(node.Value); err != nil {
		return err
	}

	var typeName string

	if node.Type != nil {
		typeName = node.Type.Value

		// Resolve type alias to underlying type
		if resolved, ok := c.typeAliases[typeName]; ok {
			typeName = resolved
		}

		typeIdx := c.addConstant(&object.String{Value: typeName})
		c.emit(code.OpCheckType, typeIdx)
	} else {
		// Basic inference
		typeName = c.inferType(node.Value)
	}

	sym := c.symbolTable.DefineWithType(node.Name.Value, typeName)

	if sym.Scope == symbol.GlobalScope { // Use sym.Scope
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
	}

	return nil
}

func (c *Compiler) compileDestructuring(pattern *ast.DestructuringPattern, value ast.Expression, isConst bool) error {
	if err := c.Compile(value); err != nil {
		return err
	}

	// Collect variables to extract with their indices in the source
	type varInfo struct {
		sym    symbol.Symbol
		field  ast.DestructuringField
		srcIdx int // index in array (or for rest)
		isRest bool
	}
	var vars []varInfo

	for i, field := range pattern.Fields {
		if field.Value == nil {
			continue
		}
		var sym symbol.Symbol
		if isConst {
			sym = c.symbolTable.DefineConstWithType(field.Value.Value, "")
		} else {
			sym = c.symbolTable.DefineWithType(field.Value.Value, "")
		}
		vars = append(vars, varInfo{sym: sym, field: field, srcIdx: i, isRest: field.Rest})
	}

	// Emit extraction code for each variable.
	// For all fields except the last, Dup the source so the next extraction can still use it.
	for j, vi := range vars {
		isLast := j == len(vars)-1
		if !isLast {
			c.emit(code.OpDup)
		}

		if pattern.IsHash {
			key := vi.field.Key.String()
			keyIdx := c.addConstant(&object.String{Value: key})
			c.emit(code.OpConstant, keyIdx)
			c.emit(code.OpIndex)
		} else if vi.isRest {
			// For ...rest, get arr[i:] — use builtin slice or manual copy
			// For now, fall back to just getting the element
			idx := c.addConstant(&object.Integer{Value: int64(vi.srcIdx)})
			c.emit(code.OpConstant, idx)
			c.emit(code.OpIndex)
		} else {
			idx := c.addConstant(&object.Integer{Value: int64(vi.srcIdx)})
			c.emit(code.OpConstant, idx)
			c.emit(code.OpIndex)
		}

		if vi.sym.Scope == symbol.GlobalScope {
			c.emit(code.OpSetGlobal, vi.sym.Index)
		} else {
			c.emit(code.OpSetLocal, vi.sym.Index)
		}
	}

	// If no variables were extracted, discard the source value
	if len(vars) == 0 {
		c.emit(code.OpPop)
	}

	return nil
}

func (c *Compiler) compileInterfaceStatement(node *ast.InterfaceStatement) error {
	interfaceDef := &object.Interface{
		Name:           node.Name.Value,
		TypeParameters: []string{},
		Methods:        make(map[string]*object.InterfaceMethod),
	}

	for _, tp := range node.TypeParameters {
		interfaceDef.TypeParameters = append(interfaceDef.TypeParameters, tp.Value)
	}

	for name, methodNode := range node.Methods {
		params := make([]string, len(methodNode.Parameters))
		for i, p := range methodNode.Parameters {
			params[i] = p.Value
		}

		retType := ""
		if methodNode.ReturnType != nil {
			retType = methodNode.ReturnType.String()
		}

		interfaceDef.Methods[name] = &object.InterfaceMethod{
			Name:       name,
			Parameters: params,
			ReturnType: retType,
		}
	}

	// Add interface definition to constants pool
	interfaceConstIdx := c.addConstant(interfaceDef)

	// Define the symbol for the interface name
	sym := c.symbolTable.Define(node.Name.Value)

	// Inject the object directly in the compiler symbol table to make it accessible during static type checking
	updatedSym := sym
	updatedSym.Value = interfaceDef
	c.symbolTable.GetStore()[node.Name.Value] = updatedSym

	// Also define it as an allowable static type string
	c.symbolTable.DefineType(node.Name.Value)

	// Emit instructions to push the interface definition (as a constant) onto the stack
	// and then assign it to the variable associated with the interface's name.
	c.emit(code.OpConstant, interfaceConstIdx)

	if sym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
	}

	return nil
}

func (c *Compiler) compileAssignmentStatement(node *ast.AssignmentStatement) error {
	// Handle Property/Index Assignment: obj.prop = val OR arr[i] = val
	if indexExpr, ok := node.Left.(*ast.IndexExpression); ok {
		// 1. Push Object
		if err := c.Compile(indexExpr.Left); err != nil {
			return err
		}
		// 2. Push Index/Key
		if err := c.Compile(indexExpr.Index); err != nil {
			return err
		}
		// 3. Push Value
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		c.emit(code.OpSetProperty)
		return nil
	}

	if arrayIndexExpr, ok := node.Left.(*ast.ArrayIndexExpression); ok {
		// 1. Push Array
		if err := c.Compile(arrayIndexExpr.Left); err != nil {
			return err
		}
		// 2. Push Index
		if err := c.Compile(arrayIndexExpr.Index); err != nil {
			return err
		}
		// 3. Push Value
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		c.emit(code.OpSetProperty)
		return nil
	}

	// We only support assignment to identifiers for now (e.g. x = 5)
	ident, ok := node.Left.(*ast.Identifier)
	if !ok {
		return c.errorPos("assignment target must be an identifier")
	}

	// Reject assignments to constants
	if c.symbolTable.IsConstant(ident.Value) {
		return c.errorPos("cannot assign to constant '%s'", ident.Value)
	}

	sym, ok := c.symbolTable.Resolve(ident.Value)
	if !ok {
		return c.errorPos("undefined variable %s", ident.Value)
	}

	// Optimization: Detect i = i + 1 or i = i - 1
	if op, ok := c.isIncrementPattern(ident, node.Value); ok {
		if sym.Scope == symbol.LocalScope {
			if op == "+" {
				c.emit(code.OpIncLocal, sym.Index)
			} else {
				c.emit(code.OpDecLocal, sym.Index)
			}
			return nil
		} else if sym.Scope == symbol.GlobalScope {
			if op == "+" {
				c.emit(code.OpIncGlobal, sym.Index)
			} else {
				c.emit(code.OpDecGlobal, sym.Index)
			}
			return nil
		}
	}

	// Non-optimized path: Compile the value to be assigned
	if err := c.Compile(node.Value); err != nil {
		return err
	}

	// Emit OpCheckType for assignments if we have a type
	if sym.DataType != "" {
		typeIdx := c.addConstant(&object.String{Value: sym.DataType})
		c.emit(code.OpCheckType, typeIdx)
	}

	switch sym.Scope {
	case symbol.GlobalScope:
		c.emit(code.OpSetGlobal, sym.Index)
	case symbol.LocalScope:
		c.emit(code.OpSetLocal, sym.Index)
	case symbol.FreeScope:
		c.emit(code.OpSetFree, sym.Index)
	default:
		return c.errorPos("cannot assign to %s scope", sym.Scope)
	}

	return nil
}

func (c *Compiler) isIncrementPattern(ident *ast.Identifier, expr ast.Expression) (string, bool) {
	infix, ok := expr.(*ast.InfixExpression)
	if !ok {
		return "", false
	}

	if infix.Operator != "+" && infix.Operator != "-" {
		return "", false
	}

	leftIdent, ok := infix.Left.(*ast.Identifier)
	if !ok || leftIdent.Value != ident.Value {
		return "", false
	}

	intLit, ok := infix.Right.(*ast.IntegerLiteral)
	if ok && intLit.Value == 1 {
		return infix.Operator, true
	}

	floatLit, ok := infix.Right.(*ast.FloatLiteral)
	if ok && floatLit.Value == 1.0 {
		return infix.Operator, true
	}

	return "", false
}

func (c *Compiler) compileExpressionStatement(node *ast.ExpressionStatement) error {
	if err := c.Compile(node.Expression); err != nil {
		return err
	}
	c.emit(code.OpPop) // Keep this line as before.
	return nil
}

func (c *Compiler) compileBlockStatement(node *ast.BlockStatement) error {
	for _, s := range node.Statements {
		if err := c.Compile(s); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileDeferStatement(node *ast.DeferStatement) error {
	// Compile the call expression normally (pushes fn + args + OpCall)
	if err := c.Compile(node.Call); err != nil {
		return err
	}
	// Replace the last instruction (OpCall or OpCallSpread) with OpDefer.
	// OpDefer has the same operand layout (1 byte numArgs).
	lastIns := c.scopes[c.scopeIndex].lastInstruction
	if lastIns.Opcode == code.OpCall || lastIns.Opcode == code.OpCallSpread {
		numArgs := int(c.currentInstructions()[lastIns.Position+1])
		newIns := code.Make(code.OpDefer, numArgs)
		c.replaceInstruction(lastIns.Position, newIns)
	}
	return nil
}

func (c *Compiler) compileDoWhileStatement(node *ast.DoWhileStatement) error {
	startPos := len(c.currentInstructions())

	c.enterLoop(startPos)

	if err := c.Compile(node.Body); err != nil {
		return err
	}

	if err := c.Compile(node.Condition); err != nil {
		return err
	}

	c.emit(code.OpJumpIfTrue, startPos)

	loop := c.leaveLoop()
	afterPos := len(c.currentInstructions())

	for _, breakPos := range loop.BreakPos {
		c.changeOperand(breakPos, afterPos)
	}

	return nil
}

func (c *Compiler) compileWhileStatement(node *ast.WhileStatement) error {
	startPos := len(c.currentInstructions())

	if err := c.Compile(node.Condition); err != nil {
		return err
	}

	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

	c.enterLoop(startPos)

	if err := c.Compile(node.Body); err != nil {
		return err
	}

	c.emit(code.OpJump, startPos)

	loop := c.leaveLoop()
	afterPos := len(c.currentInstructions())

	c.changeOperand(jumpNotTruthyPos, afterPos)

	for _, breakPos := range loop.BreakPos {
		c.changeOperand(breakPos, afterPos)
	}

	return nil
}

func (c *Compiler) compileForStatement(node *ast.ForStatement) error {
	// c.enterScope() // Removed to avoid local variable issues in main

	if node.Init != nil {
		if err := c.Compile(node.Init); err != nil {
			return err
		}
	}

	startPos := len(c.currentInstructions())

	jumpNotTruthyPos := -1
	if node.Condition != nil {
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		jumpNotTruthyPos = c.emit(code.OpJumpNotTruthy, 9999)
	}

	c.enterLoop(-1) // -1 because update position is unknown yet

	if err := c.Compile(node.Body); err != nil {
		return err
	}

	loop := c.leaveLoop()
	updatePos := len(c.currentInstructions())

	for _, pos := range loop.ContinueJumps {
		c.changeOperand(pos, updatePos)
	}

	if node.Update != nil {
		if err := c.Compile(node.Update); err != nil {
			return err
		}
	}

	c.emit(code.OpJump, startPos)

	afterPos := len(c.currentInstructions())

	if jumpNotTruthyPos != -1 {
		c.changeOperand(jumpNotTruthyPos, afterPos)
	}

	for _, breakPos := range loop.BreakPos {
		c.changeOperand(breakPos, afterPos)
	}

	// c.leaveScope() // Removed

	return nil
}

func (c *Compiler) compileBreakStatement(node *ast.BreakStatement) error {
	jumpPos := c.emit(code.OpJump, 9999)
	if c.loopIndex < 0 {
		return c.errorPos("break statement outside of loop")
	}
	c.loops[c.loopIndex].BreakPos = append(c.loops[c.loopIndex].BreakPos, jumpPos)
	return nil
}

func (c *Compiler) compileContinueStatement(node *ast.ContinueStatement) error {
	if c.loopIndex < 0 {
		return c.errorPos("continue statement outside of loop")
	}
	pos := c.loops[c.loopIndex].ContinuePos
	if pos == -1 {
		jumpPos := c.emit(code.OpJump, 9999)
		c.loops[c.loopIndex].ContinueJumps = append(c.loops[c.loopIndex].ContinueJumps, jumpPos)
	} else {
		c.emit(code.OpJump, pos)
	}
	return nil
}
func (c *Compiler) compileStructStatement(node *ast.StructStatement) error {
	structDef := &object.Struct{
		Name:           node.Name.Value,
		TypeParameters: []string{},
		Fields:         make(map[string]string),
	}

	for _, tp := range node.TypeParameters {
		structDef.TypeParameters = append(structDef.TypeParameters, tp.Value)
	}

	for name, typeExpr := range node.Fields {
		structDef.Fields[name] = typeExpr.String()
	}

	// Add struct definition to constants pool
	structConstIdx := c.addConstant(structDef)

	// Define the symbol for the struct name
	sym := c.symbolTable.Define(node.Name.Value)

	// Update the symbol in the symbol table to store the actual struct definition object
	// This makes the struct definition directly accessible when resolving the symbol later.
	updatedSym := sym
	updatedSym.Value = structDef // Store the *object.Struct here
	c.symbolTable.GetStore()[node.Name.Value] = updatedSym

	// Emit instructions to push the struct definition (as a constant) onto the stack
	// and then assign it to the variable associated with the struct's name.
	c.emit(code.OpConstant, structConstIdx) // Push the constant index of the struct definition

	if sym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
	}

	return nil
}
func (c *Compiler) compileServiceStatement(node *ast.ServiceStatement) error {
	// 1. Compile Config Fields
	for name, expr := range node.Fields {
		c.emit(code.OpConstant, c.addConstant(&object.String{Value: name}))
		if err := c.Compile(expr); err != nil {
			return err
		}
	}

	// 2. Emit OpService
	nameIdx := c.addConstant(&object.String{Value: node.Name.Value})
	c.emit(code.OpService, nameIdx, len(node.Fields))

	// Define variable
	sym := c.symbolTable.Define(node.Name.Value)
	if sym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
	}

	// 3. Compile Methods
	for _, method := range node.Methods {
		// Inject implicit receiver info
		method.ReceiverType = &ast.Identifier{Value: node.Name.Value}
		method.ReceiverName = &ast.Identifier{Value: "this"}

		if err := c.compileFunctionStatement(method); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) compileRetryStatement(node *ast.RetryStatement) error {
	// 1. Compile Attempts Expression
	if err := c.Compile(node.Attempts); err != nil {
		return err
	}
	attemptsSym := c.symbolTable.Define("$$attempts$$")
	if attemptsSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, attemptsSym.Index)
	} else {
		c.emit(code.OpSetLocal, attemptsSym.Index)
	}

	// Mark loop start
	loopStartPos := len(c.currentInstructions())

	// 2. OpTry to wrap the block
	opTryPos := c.emit(code.OpTry, 9999)

	// 3. Compile the Retry Block
	if err := c.Compile(node.RetryBlock); err != nil {
		return err
	}
	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}

	c.emit(code.OpEndTry)

	// If successful, jump out of the retry loop completely
	successJumpPos := c.emit(code.OpJump, 9999)

	// 4. Internal Catch Handler (Retry Logic)
	catchStartPos := len(c.currentInstructions())
	c.changeOperand(opTryPos, catchStartPos)

	// Stack has the Exception. Save it temporarily.
	errSym := c.symbolTable.Define("$$retry_err$$")
	if errSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, errSym.Index)
	} else {
		c.emit(code.OpSetLocal, errSym.Index)
	}

	// Decrement attempts: attempts = attempts - 1
	if attemptsSym.Scope == symbol.GlobalScope {
		c.emit(code.OpGetGlobal, attemptsSym.Index)
	} else {
		c.emit(code.OpGetLocal, attemptsSym.Index)
	}
	c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 1}))
	c.emit(code.OpSub)

	// Update attempts variable (keeping value on stack for comparison via OpSet... wait OpSet pops)
	// We need to Dup if we want to use it.
	// Or just Set, then Get again.
	if attemptsSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, attemptsSym.Index)
	} else {
		c.emit(code.OpSetLocal, attemptsSym.Index)
	}

	// Check if attempts > 0
	if attemptsSym.Scope == symbol.GlobalScope {
		c.emit(code.OpGetGlobal, attemptsSym.Index)
	} else {
		c.emit(code.OpGetLocal, attemptsSym.Index)
	}
	c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 0}))
	c.emit(code.OpGreaterThan)

	// If attempts > 0, Jump back to loop start (Retry)
	// OpJumpIfTrue (or JumpTruthy)
	c.emit(code.OpJumpIfTrue, loopStartPos)

	// 5. Attempts Exhausted (attempts <= 0)
	// Load the saved exception
	if errSym.Scope == symbol.GlobalScope {
		c.emit(code.OpGetGlobal, errSym.Index)
	} else {
		c.emit(code.OpGetLocal, errSym.Index)
	}

	if node.CatchBlock != nil {
		// User provided catch block
		// Bind exception to catch param if provided
		if node.CatchParam != nil {
			catchParamSym := c.symbolTable.Define(node.CatchParam.Value)
			if catchParamSym.Scope == symbol.GlobalScope {
				c.emit(code.OpSetGlobal, catchParamSym.Index)
			} else {
				c.emit(code.OpSetLocal, catchParamSym.Index)
			}
		} else {
			c.emit(code.OpPop) // Discard exception if unnamed
		}

		if err := c.Compile(node.CatchBlock); err != nil {
			return err
		}
	} else {
		// No catch block, re-throw the exception
		c.emit(code.OpThrow)
	}

	// End of structure
	afterPos := len(c.currentInstructions())
	c.changeOperand(successJumpPos, afterPos)

	// retryJumpPos is already set to loopStartPos, no need to patch to end.
	// But wait, OpJumpIfTrue takes an operand. We set it to loopStartPos. Correct.

	return nil
}

func (c *Compiler) compileThrowStatement(node *ast.ThrowStatement) error {
	if err := c.Compile(node.Value); err != nil {
		return err
	}
	c.emit(code.OpThrow)
	return nil
}
func (c *Compiler) compileTryStatement(node *ast.TryStatement) error {
	// Use the enclosing scope directly — no enclosed scope needed.
	// Variables defined inside try/catch blocks use the parent's indices,
	// avoiding free-variable confusion since try is NOT a closure.

	// OpTry: placeholder for CatchIP and FinallyIP
	opTryPos := c.emit(code.OpTry, 9999, 9999)

	// Compile TryBlock
	if err := c.Compile(node.TryBlock); err != nil {
		return err
	}
	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}
	c.emit(code.OpEndTry)

	// Jump to finally (or past catch) on success
	jumpToFinally := c.emit(code.OpJump, 9999)

	// === Finally Block (compiled BEFORE catch so catch follows it) ===
	finallyStartPos := len(c.currentInstructions())
	var afterFinallyJump int
	if node.Finally != nil {
		c.emit(code.OpFinally)
		if err := c.Compile(node.Finally); err != nil {
			return err
		}
		c.emit(code.OpEndFinally)
		// On success path (no exception), skip over the catch block
		afterFinallyJump = c.emit(code.OpJump, 9999)
	}

	// === Catch Block ===
	catchStartPos := len(c.currentInstructions())
	if node.CatchBlock != nil {
		if node.CatchType != nil {
			typeIdx := c.addConstant(&object.String{Value: node.CatchType.Value})
			c.emit(code.OpIsType, typeIdx)
			skipCatch := c.emit(code.OpJumpNotTruthy, 9999)

			if node.CatchParam != nil {
				sym, existed := c.symbolTable.Resolve(node.CatchParam.Value)
				if !existed {
					sym = c.symbolTable.Define(node.CatchParam.Value)
				}
				if sym.Scope == symbol.GlobalScope {
					c.emit(code.OpSetGlobal, sym.Index)
				} else {
					c.emit(code.OpSetLocal, sym.Index)
				}
			} else {
				c.emit(code.OpPop)
			}

			if err := c.Compile(node.CatchBlock); err != nil {
				return err
			}
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}

			c.emit(code.OpJump, 9999)

			rethrowPos := len(c.currentInstructions())
			c.changeOperand(skipCatch, rethrowPos)
			c.emit(code.OpThrow)
		} else {
			if node.CatchParam != nil {
				sym, existed := c.symbolTable.Resolve(node.CatchParam.Value)
				if !existed {
					sym = c.symbolTable.Define(node.CatchParam.Value)
				}
				if sym.Scope == symbol.GlobalScope {
					c.emit(code.OpSetGlobal, sym.Index)
				} else {
					c.emit(code.OpSetLocal, sym.Index)
				}
			} else {
				c.emit(code.OpPop)
			}
			if err := c.Compile(node.CatchBlock); err != nil {
				return err
			}
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}
	} else {
		c.emit(code.OpPop)
	}

	// === Patching ===
	afterCatchPos := len(c.currentInstructions())
	if node.Finally != nil {
		c.changeOperand(opTryPos, catchStartPos, finallyStartPos)
		c.changeOperand(jumpToFinally, finallyStartPos)
		c.changeOperand(afterFinallyJump, afterCatchPos)
	} else {
		c.changeOperand(opTryPos, catchStartPos, 0)
		c.changeOperand(jumpToFinally, afterCatchPos)
	}

	return nil
}
func (c *Compiler) compileSwitchStatement(node *ast.SwitchStatement) error {
	if err := c.Compile(node.Expression); err != nil {
		return err
	}

	var jumpToEnds []int

	for _, caseClause := range node.Cases {
		c.emit(code.OpDup) // Duplicate subject for comparison

		if err := c.Compile(caseClause.Value); err != nil {
			return err
		}

		c.emit(code.OpEqual)

		jumpNotMatch := c.emit(code.OpJumpNotTruthy, 9999)

		// Case matched
		c.emit(code.OpPop) // Pop the original subject

		for _, s := range caseClause.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

		jumpToEnd := c.emit(code.OpJump, 9999)
		jumpToEnds = append(jumpToEnds, jumpToEnd)

		// Case not matched
		afterPos := len(c.currentInstructions())
		c.changeOperand(jumpNotMatch, afterPos)
	}

	if node.DefaultCase != nil {
		c.emit(code.OpPop) // Pop original subject
		for _, s := range node.DefaultCase.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
	} else {
		c.emit(code.OpPop) // Pop original subject if no match and no default
	}

	afterSwitchPos := len(c.currentInstructions())
	for _, pos := range jumpToEnds {
		c.changeOperand(pos, afterSwitchPos)
	}

	return nil
}
func (c *Compiler) compileSelectStatement(node *ast.SelectStatement) error {
	numCases := len(node.Cases)
	hasDefault := 0
	if node.DefaultCase != nil {
		hasDefault = 1
	}

	// 1. Compile all channel/value expressions onto the stack
	for _, cas := range node.Cases {
		if err := c.Compile(cas.Channel); err != nil {
			return err
		}
		if cas.IsSend {
			if err := c.Compile(cas.Value); err != nil {
				return err
			}
		}
	}

	// 2. Emit OpSelect opcode + operands (3 bytes total)
	c.emit(code.OpSelect, numCases, hasDefault)

	// 3. Write case descriptors (flags + 2-byte offset) inline
	descStart := len(c.currentInstructions())
	for _, cas := range node.Cases {
		flags := byte(0)
		if cas.IsSend {
			flags |= 1
		}
		c.setInstructions(append(c.currentInstructions(), flags, 0, 0))
	}
	if hasDefault != 0 {
		c.setInstructions(append(c.currentInstructions(), 0, 0))
	}

	// 4. Compile each case body with backpatched offsets
	var jumpToEnds []int
	for i, cas := range node.Cases {
		bodyStart := len(c.currentInstructions())
		// Write body offset into descriptor
		off := descStart + i*3 + 1
		code.WriteUint16(c.currentInstructions(), off, uint16(bodyStart))

		for _, s := range cas.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
		jumpToEnds = append(jumpToEnds, c.emit(code.OpJump, 9999))
	}

	// 5. Compile default body
	if node.DefaultCase != nil {
		bodyStart := len(c.currentInstructions())
		defaultOff := descStart + numCases*3
		code.WriteUint16(c.currentInstructions(), defaultOff, uint16(bodyStart))
		for _, s := range node.DefaultCase.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
		jumpToEnds = append(jumpToEnds, c.emit(code.OpJump, 9999))
	}

	// 6. Backpatch all jumps to end
	endPos := len(c.currentInstructions())
	for _, pos := range jumpToEnds {
		c.changeOperand(pos, endPos)
	}

	return nil
}

func (c *Compiler) compileConstStatement(node *ast.ConstStatement) error {
	if node.Destructure != nil {
		return c.compileDestructuring(node.Destructure, node.Value, true)
	}

	if err := c.Compile(node.Value); err != nil {
		return err
	}

	var typeName string
	if node.Type != nil {
		typeName = node.Type.Value

		// Resolve type alias to underlying type
		if resolved, ok := c.typeAliases[typeName]; ok {
			typeName = resolved
		}

		typeIdx := c.addConstant(&object.String{Value: typeName})
		c.emit(code.OpCheckType, typeIdx)
	} else {
		typeName = c.inferType(node.Value)
	}

	sym := c.symbolTable.DefineConstWithType(node.Name.Value, typeName)

	if sym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
	}

	return nil
}

func (c *Compiler) compileEnumStatement(node *ast.EnumStatement) error {
	pairs := make(map[object.HashKey]object.HashPair)

	for i, variant := range node.Values {
		keyObj := &object.String{Value: variant.Value}
		valObj := &object.Integer{Value: int64(i)}
		pairs[keyObj.HashKey()] = object.HashPair{Key: keyObj, Value: valObj}
	}

	enumHash := &object.Hash{Pairs: pairs}
	constIdx := c.addConstant(enumHash)
	c.emit(code.OpConstant, constIdx)

	sym := c.symbolTable.DefineConst(node.Name.Value)

	if sym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
	}

	return nil
}

func (c *Compiler) compileEchoStatement(node *ast.EchoStatement) error {
	sym, ok := c.symbolTable.Resolve("echo") // Renamed variable
	if !ok {
		return c.errorPos("builtin 'echo' not found")
	}

	c.emit(code.OpGetBuiltin, sym.Index) // Use sym.Index

	for _, val := range node.Values {
		if err := c.Compile(val); err != nil {
			return err
		}
	}

	c.emit(code.OpCall, len(node.Values))
	c.emit(code.OpPop)

	return nil
}

func (c *Compiler) compileImportStatement(node *ast.ImportStatement) error {
	if err := c.Compile(node.ModuleName); err != nil {
		return err
	}

	c.emit(code.OpImport) // Pushes module hash

	switch node.ImportType {
	case ast.IMPORT_NAMESPACE, ast.IMPORT_ALIAS: // import * as name, import "mod" as name
		sym := c.symbolTable.Define(node.NamespaceAlias.Value)
		if sym.Scope == symbol.GlobalScope {
			c.emit(code.OpSetGlobal, sym.Index)
		} else {
			c.emit(code.OpSetLocal, sym.Index)
		}

	case ast.IMPORT_NAMED: // import { a, b }
		for _, item := range node.NamedImports {
			c.emit(code.OpDup)

			keyStr := &object.String{Value: item.Name.Value}
			c.emit(code.OpConstant, c.addConstant(keyStr))
			c.emit(code.OpIndex)

			varName := item.Name.Value
			if item.Alias != nil {
				varName = item.Alias.Value
			}

			sym := c.symbolTable.Define(varName)
			if sym.Scope == symbol.GlobalScope {
				c.emit(code.OpSetGlobal, sym.Index)
			} else {
				c.emit(code.OpSetLocal, sym.Index)
			}
			// OpSet consumes the value (property)
		}
		c.emit(code.OpPop) // Pop the module hash

	case ast.IMPORT_DEFAULT: // import d from "mod"
		c.emit(code.OpDup)

		keyStr := &object.String{Value: "default"}
		c.emit(code.OpConstant, c.addConstant(keyStr))
		c.emit(code.OpIndex)

		sym := c.symbolTable.Define(node.DefaultImport.Value)
		if sym.Scope == symbol.GlobalScope {
			c.emit(code.OpSetGlobal, sym.Index)
		} else {
			c.emit(code.OpSetLocal, sym.Index)
		}
		c.emit(code.OpPop) // Pop module hash

	case ast.IMPORT_SIDE_EFFECT:
		c.emit(code.OpPop)
	}

	return nil
}

func (c *Compiler) compileExportStatement(node *ast.ExportStatement) error {
	if node.Statement != nil {
		if err := c.Compile(node.Statement); err != nil {
			return err
		}

		switch stmt := node.Statement.(type) {
		case *ast.LetStatement:
			c.symbolTable.MarkExported(stmt.Name.Value)
		case *ast.ConstStatement:
			c.symbolTable.MarkExported(stmt.Name.Value)
		case *ast.FunctionStatement:
			c.symbolTable.MarkExported(stmt.Name.Value)
		case *ast.StructStatement:
			c.symbolTable.MarkExported(stmt.Name.Value)
		case *ast.EnumStatement:
			c.symbolTable.MarkExported(stmt.Name.Value)
		case *ast.ServiceStatement:
			c.symbolTable.MarkExported(stmt.Name.Value)
		}
	}
	return nil
}
func (c *Compiler) compileForEachStatement(node *ast.ForEachStatement) error {
	// Do NOT enter a new scope. Use the current function's scope for locals.

	// 1. Compile the Iterable expression and store it in a temporary local variable
	if err := c.Compile(node.Iterable); err != nil {
		return err
	}
	iterableSym := c.symbolTable.Define("$$iterable$$") // Define a temporary symbol
	if iterableSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, iterableSym.Index)
	} else {
		c.emit(code.OpSetLocal, iterableSym.Index)
	}

	// 2. Initialize a temporary index variable to 0
	indexSym := c.symbolTable.Define("$$index$$") // Define a temporary symbol
	c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 0}))
	if indexSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, indexSym.Index)
	} else {
		c.emit(code.OpSetLocal, indexSym.Index)
	}

	loopStartPos := len(c.currentInstructions()) // Mark the start of the loop

	c.enterLoop(-1) // Use -1 so continue statements are collected and patched to increment

	// 3. Get next item from iterable: OpNextItem pops (index, iterable), pushes (value, nextIndex, hasNext)
	if iterableSym.Scope == symbol.GlobalScope {
		c.emit(code.OpGetGlobal, iterableSym.Index)
	} else {
		c.emit(code.OpGetLocal, iterableSym.Index)
	}

	if indexSym.Scope == symbol.GlobalScope {
		c.emit(code.OpGetGlobal, indexSym.Index)
	} else {
		c.emit(code.OpGetLocal, indexSym.Index)
	}

	c.emit(code.OpNextItem)

	// hasNext is on top — jump if falsy (exhausted iterable)
	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

	// 4. Store nextIndex back into the index slot (it's now on top)
	if indexSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, indexSym.Index)
	} else {
		c.emit(code.OpSetLocal, indexSym.Index)
	}

	// 5. Store value (now on top) in the user-defined loop variable
	itemVarSym := c.symbolTable.Define(node.Variable.Value)
	if itemVarSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, itemVarSym.Index)
	} else {
		c.emit(code.OpSetLocal, itemVarSym.Index)
	}

	// 6. Compile Body
	if err := c.Compile(node.Body); err != nil {
		return err
	}

	// Remove last OpPop if it exists
	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}

	// Handle continue statements
	loopScope := c.leaveLoop()
	continuePos := len(c.currentInstructions())

	// 7. Jump back to loop start
	c.emit(code.OpJump, loopStartPos)

	// Patch continue jumps to point to after-loop (same as jump back)
	for _, pos := range loopScope.ContinueJumps {
		c.changeOperand(pos, continuePos)
	}

	// 8. Exit point: pop the Null sentinels left by OpNextItem on exhaustion
	afterLoopPos := len(c.currentInstructions())
	c.changeOperand(jumpNotTruthyPos, afterLoopPos)

	c.emit(code.OpPop) // pop nextIndex (Null)
	c.emit(code.OpPop) // pop value (Null)

	// Handle break statements
	for _, breakPos := range loopScope.BreakPos {
		c.changeOperand(breakPos, afterLoopPos)
	}

	return nil
}

func (c *Compiler) compileMatchStatement(node *ast.MatchStatement) error {
	if err := c.Compile(node.Expression); err != nil {
		return err
	}

	var jumpToEnds []int

	for _, matchCase := range node.Cases {
		if matchCase.IsDefault {
			continue // Handle default at the end
		}

		c.emit(code.OpDup) // Duplicate subject for each case comparison

		switch pattern := matchCase.Pattern.(type) {

		case *ast.Identifier:
			// Type match: case Int / case String / etc.
			typeIdx := c.addConstant(&object.String{Value: pattern.Value})
			c.emit(code.OpIsType, typeIdx)

		case *ast.ArrayLiteral:
			// Structural match: case [a, b, c]
			// Stack: [arr_dup]
			// Store arr_dup into a temp variable so we can call len() and index it.
			tmpArrSym := c.symbolTable.Define("$$match_arr$$")
			if tmpArrSym.Scope == symbol.GlobalScope {
				c.emit(code.OpSetGlobal, tmpArrSym.Index)
			} else {
				c.emit(code.OpSetLocal, tmpArrSym.Index)
			}

			// Call len($$match_arr$$) and compare with pattern element count.
			lenSym, ok := c.symbolTable.Resolve("len")
			if !ok {
				return c.errorPos("builtin 'len' not found for structural match")
			}
			c.emit(code.OpGetBuiltin, lenSym.Index)
			if tmpArrSym.Scope == symbol.GlobalScope {
				c.emit(code.OpGetGlobal, tmpArrSym.Index)
			} else {
				c.emit(code.OpGetLocal, tmpArrSym.Index)
			}
			c.emit(code.OpCall, 1) // result: actual length

			expectedLen := int64(len(pattern.Elements))
			expectedLenIdx := c.addConstant(&object.Integer{Value: expectedLen})
			c.emit(code.OpConstant, expectedLenIdx)
			c.emit(code.OpEqual) // true if lengths match

			// After jumpNotMatch we bind elements. Handled below.

		default:
			// Equality match: case 42 / case "hello" / etc.
			if err := c.Compile(matchCase.Pattern); err != nil {
				return err
			}
			c.emit(code.OpEqual)
		}

		jumpNotMatch := c.emit(code.OpJumpNotTruthy, 9999)

		// Case matched — pop the original subject from stack.
		c.emit(code.OpPop)

		// For structural array match: bind each identifier element to the array slot.
		if pattern, ok := matchCase.Pattern.(*ast.ArrayLiteral); ok {
			tmpArrSym, _ := c.symbolTable.Resolve("$$match_arr$$")
			for i, elem := range pattern.Elements {
				ident, ok := elem.(*ast.Identifier)
				if !ok {
					continue // skip non-identifier (wildcard '_' or literal) elements
				}
				// Load arr[i]
				if tmpArrSym.Scope == symbol.GlobalScope {
					c.emit(code.OpGetGlobal, tmpArrSym.Index)
				} else {
					c.emit(code.OpGetLocal, tmpArrSym.Index)
				}
				idxConst := c.addConstant(&object.Integer{Value: int64(i)})
				c.emit(code.OpConstant, idxConst)
				c.emit(code.OpIndex) // arr[i]

				// Bind to the variable name
				varSym := c.symbolTable.Define(ident.Value)
				if varSym.Scope == symbol.GlobalScope {
					c.emit(code.OpSetGlobal, varSym.Index)
				} else {
					c.emit(code.OpSetLocal, varSym.Index)
				}
			}
		}

		for _, s := range matchCase.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

		jumpToEnd := c.emit(code.OpJump, 9999)
		jumpToEnds = append(jumpToEnds, jumpToEnd)

		afterPos := len(c.currentInstructions())
		c.changeOperand(jumpNotMatch, afterPos)
	}

	// Handle default case
	foundDefault := false
	for _, matchCase := range node.Cases {
		if matchCase.IsDefault {
			c.emit(code.OpPop) // Pop subject
			for _, s := range matchCase.Statements {
				if err := c.Compile(s); err != nil {
					return err
				}
			}
			foundDefault = true
			break
		}
	}

	if !foundDefault {
		c.emit(code.OpPop) // Pop subject if no match
	}

	afterMatchPos := len(c.currentInstructions())
	for _, pos := range jumpToEnds {
		c.changeOperand(pos, afterMatchPos)
	}

	return nil
}
