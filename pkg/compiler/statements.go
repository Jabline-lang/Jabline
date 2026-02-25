package compiler

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object" // New import
	"jabline/pkg/symbol"
)

func (c *Compiler) compileLetStatement(node *ast.LetStatement) error {
	if err := c.Compile(node.Value); err != nil {
		return err
	}

	var typeName string
	if node.Type != nil {
		typeName = node.Type.Value
		valType := c.inferType(node.Value)
		if err := c.checkTypeMatch(typeName, valType, node.Value); err != nil {
			// Add file:line:col info to the error
			return fmt.Errorf("compile error: %s", err)
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

	// Compile the value to be assigned
	if err := c.Compile(node.Value); err != nil {
		return err
	}

	// We only support assignment to identifiers for now (e.g. x = 5)
	ident, ok := node.Left.(*ast.Identifier)
	if !ok {
		return fmt.Errorf("assignment target must be an identifier")
	}

	// Reject assignments to constants
	if c.symbolTable.IsConstant(ident.Value) {
		return fmt.Errorf("cannot assign to constant '%s'", ident.Value)
	}

	sym, ok := c.symbolTable.Resolve(ident.Value)
	if !ok {
		return fmt.Errorf("undefined variable %s", ident.Value)
	}

	// Static type validation for assignment
	valType := c.inferType(node.Value)
	if err := c.checkTypeMatch(sym.DataType, valType, node.Value); err != nil {
		return fmt.Errorf("compile error: assignment to '%s' failed - %s", ident.Value, err)
	}

	switch sym.Scope {
	case symbol.GlobalScope:
		c.emit(code.OpSetGlobal, sym.Index)
	case symbol.LocalScope:
		c.emit(code.OpSetLocal, sym.Index)
	case symbol.FreeScope:
		c.emit(code.OpSetFree, sym.Index)
	default:
		return fmt.Errorf("cannot assign to %s scope", sym.Scope)
	}

	return nil
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
		return fmt.Errorf("break statement outside of loop")
	}
	c.loops[c.loopIndex].BreakPos = append(c.loops[c.loopIndex].BreakPos, jumpPos)
	return nil
}

func (c *Compiler) compileContinueStatement(node *ast.ContinueStatement) error {
	if c.loopIndex < 0 {
		return fmt.Errorf("continue statement outside of loop")
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
	// Do NOT create a new CompilationScope (c.enterScope), because that resets instruction offsets.
	// We want OpTry/OpJump offsets to be relative to the current function's bytecode.
	// However, we DO want a new SymbolScope for the catch block variables.

	// Manually create a new symbol scope
	originalSymbolTable := c.symbolTable
	c.symbolTable = symbol.NewEnclosedSymbolTable(originalSymbolTable)

	// Restore symbol table on exit
	defer func() {
		c.symbolTable = originalSymbolTable
	}()

	// 1. Emit OpTry with a placeholder operand (points to catch block start)
	opTryPos := c.emit(code.OpTry, 9999) // Placeholder for CatchIP

	// 2. Compile TryBlock
	if err := c.Compile(node.TryBlock); err != nil {
		return err
	}

	// Remove potential OpPop after TryBlock if it's the last instruction
	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}

	// Emit OpEndTry immediately after TryBlock (path of success)
	c.emit(code.OpEndTry)

	// 3. Emit OpJump with a placeholder operand (points after catch block)
	jumpAfterCatchPos := c.emit(code.OpJump, 9999) // Placeholder to jump over catch block

	// 4. Mark catch block start and patch OpTry operand
	catchStartPos := len(c.currentInstructions())
	c.changeOperand(opTryPos, catchStartPos) // OpTry now points to catch block

	// Handle CatchBlock if present
	if node.CatchBlock != nil {
		// If CatchParam is defined, define it in the symbol table and store the exception object
		if node.CatchParam != nil {
			sym := c.symbolTable.Define(node.CatchParam.Value)
			// The VM pushes the exception object onto the stack before jumping to catchStartPos
			// So, we need to pop it and set it as a local variable.
			if sym.Scope == symbol.GlobalScope {
				c.emit(code.OpSetGlobal, sym.Index)
			} else {
				c.emit(code.OpSetLocal, sym.Index)
			}
		} else {
			// If no catch param, pop the exception object from the stack
			c.emit(code.OpPop)
		}

		// Compile CatchBlock
		if err := c.Compile(node.CatchBlock); err != nil {
			return err
		}
		// Remove potential OpPop after CatchBlock
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
	} else {
		c.emit(code.OpPop) // Pop the exception that OpTry pushed on failed try.
	}

	// 5. Mark end of try/catch and patch OpJump operand
	endTryPos := len(c.currentInstructions())
	c.changeOperand(jumpAfterCatchPos, endTryPos) // Jump after catch block

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
func (c *Compiler) compileConstStatement(node *ast.ConstStatement) error {
	if err := c.Compile(node.Value); err != nil {
		return err
	}

	var typeName string
	if node.Type != nil {
		typeName = node.Type.Value
		valType := c.inferType(node.Value)
		if err := c.checkTypeMatch(typeName, valType, node.Value); err != nil {
			return fmt.Errorf("compile error: constant '%s' type mismatch - %s", node.Name.Value, err)
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
		return fmt.Errorf("builtin 'echo' not found")
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

	// 3. Condition: while (index < len(iterable))
	c.enterLoop(-1) // Use -1 so continue statements are collected and patched to increment

	// Get length of iterable
	lenSym, ok := c.symbolTable.Resolve("len") // Assume 'len' builtin is available
	if !ok {
		return fmt.Errorf("builtin 'len' not found for ForEachStatement")
	}
	c.emit(code.OpGetBuiltin, lenSym.Index)

	if iterableSym.Scope == symbol.GlobalScope {
		c.emit(code.OpGetGlobal, iterableSym.Index)
	} else {
		c.emit(code.OpGetLocal, iterableSym.Index)
	}

	c.emit(code.OpCall, 1) // Call len(iterable)

	// Get current index
	if indexSym.Scope == symbol.GlobalScope {
		c.emit(code.OpGetGlobal, indexSym.Index)
	} else {
		c.emit(code.OpGetLocal, indexSym.Index)
	}

	// Compare: index < len(iterable)
	c.emit(code.OpGreaterThan) // Stack: [len, index]. OpGreaterThan -> [len > index] (true if elements remain)

	// Jump if not truthy (i.e., len <= index, loop finished)
	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999) // Placeholder to jump out of loop

	// 4. Get current item: let item = iterable[index];
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

	c.emit(code.OpIndex) // Stack: [item]

	// Store item in the user-defined variable for the loop body
	itemVarSym := c.symbolTable.Define(node.Variable.Value) // Define user's loop variable
	if itemVarSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, itemVarSym.Index)
	} else {
		c.emit(code.OpSetLocal, itemVarSym.Index)
	}

	// 5. Compile Body
	if err := c.Compile(node.Body); err != nil {
		return err
	}

	// Remove last OpPop if it exists
	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}

	// Handle continue statements
	loopScope := c.leaveLoop()
	// Update continue jumps to point to increment step
	continuePos := len(c.currentInstructions())

	// 6. Increment index: index = index + 1;
	if indexSym.Scope == symbol.GlobalScope {
		c.emit(code.OpGetGlobal, indexSym.Index)
	} else {
		c.emit(code.OpGetLocal, indexSym.Index)
	}

	c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 1}))
	c.emit(code.OpAdd)

	if indexSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, indexSym.Index)
	} else {
		c.emit(code.OpSetLocal, indexSym.Index)
	}

	// 7. Jump back to loop start
	c.emit(code.OpJump, loopStartPos)

	// Patch continue jumps to point to increment
	for _, pos := range loopScope.ContinueJumps {
		c.changeOperand(pos, continuePos)
	}

	// 8. Patch jump out of loop (after loop body)
	afterLoopPos := len(c.currentInstructions())
	c.changeOperand(jumpNotTruthyPos, afterLoopPos)

	// Handle break statements
	for _, breakPos := range loopScope.BreakPos {
		c.changeOperand(breakPos, afterLoopPos)
	}

	return nil
}
