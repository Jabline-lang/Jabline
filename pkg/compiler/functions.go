package compiler

import (
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"jabline/pkg/symbol"
)

func (c *Compiler) compileFunctionStatement(node *ast.FunctionStatement) error {
	var fnName string
	if node.ReceiverName != nil {
		// Mangle method name: StructName.MethodName
		fnName = node.ReceiverType.Value + "." + node.Name.Value
	} else {
		fnName = node.Name.Value
	}

	returnType := ""
	if node.ReturnType != nil {
		returnType = node.ReturnType.Value
	}

	outerSym := c.symbolTable.DefineWithType(fnName, returnType) // Define the function name in the outer scope.

	c.enterScopeWithType(returnType, true) // Enter the function's new scope
	c.symbolTable.DefineFunctionName(fnName)

	// Define type parameters in the scope so they are recognized as types
	for _, tp := range node.TypeParameters {
		c.symbolTable.DefineType(tp.Value)
	}

	// If method, define receiver as the first parameter (local 0)
	if node.ReceiverName != nil {
		c.symbolTable.Define(node.ReceiverName.Value)
	}

	// Compile the parameters as local variables within the function's scope.
	var paramSymbols []symbol.Symbol
	for _, p := range node.Parameters {
		paramType := ""
		if p.Type != nil {
			paramType = p.Type.Value
		}
		sym := c.symbolTable.DefineWithType(p.Value, paramType)
		paramSymbols = append(paramSymbols, sym)
	}

	// Emit default value initialization prologue for parameters with defaults
	// This must run BEFORE type checks so defaults are applied first
	for i, p := range node.Parameters {
		if p.DefaultValue != nil {
			sym := paramSymbols[i]
			// get_local idx; jump_not_null skip; pop; <default>; set_local idx; jump end; skip: pop
			c.emit(code.OpGetLocal, sym.Index)
			jumpNotNullPos := c.emit(code.OpJumpNotNull, 9999)
			c.emit(code.OpPop)
			if err := c.Compile(p.DefaultValue); err != nil {
				return err
			}
			c.emit(code.OpSetLocal, sym.Index)
			jumpEndPos := c.emit(code.OpJump, 9999)
			skipPos := len(c.currentInstructions())
			c.changeOperand(jumpNotNullPos, skipPos)
			c.emit(code.OpPop)
			endPos := len(c.currentInstructions())
			c.changeOperand(jumpEndPos, endPos)
		}
	}

	// Runtime type checks for parameters with type annotations (after defaults applied)
	for i, p := range node.Parameters {
		if p.Type != nil {
			sym := paramSymbols[i]
			typeIdx := c.addConstant(&object.String{Value: p.Type.Value})
			c.emit(code.OpGetLocal, sym.Index)
			c.emit(code.OpCheckType, typeIdx)
			c.emit(code.OpPop)
		}
	}

	if err := c.Compile(node.Body); err != nil {
		return err
	}

	if c.lastInstructionIs(code.OpPop) {
		c.replaceLastPopWithReturn()
	}
	if !c.lastInstructionIs(code.OpReturnValue) {
		c.emit(code.OpReturn)
	}

	freeSymbols := c.symbolTable.FreeSymbols
	numLocals := c.symbolTable.NumDefinitions() // Access via getter
	fnSymTable := c.symbolTable                 // Save before leaveScope
	instructions, sourceMap := c.leaveScope()   // Exit the function's scope

	for _, s := range freeSymbols {
		switch s.Scope {
		case symbol.GlobalScope:
			c.emit(code.OpGetGlobal, s.Index)
		case symbol.LocalScope:
			c.emit(code.OpGetLocal, s.Index)
		case symbol.FreeScope:
			c.emit(code.OpGetFree, s.Index)
		case symbol.FunctionScope:
			c.emit(code.OpCurrentClosure)
		}
	}

	numParams := len(node.Parameters)
	if node.ReceiverName != nil {
		numParams++
	}

	isVariadic := false
	if len(node.Parameters) > 0 && node.Parameters[len(node.Parameters)-1].Variadic {
		isVariadic = true
	}

	typeParams := []string{}
	for _, tp := range node.TypeParameters {
		typeParams = append(typeParams, tp.Value)
	}

	compiledFn := &object.CompiledFunction{
		Instructions:   instructions,
		NumLocals:      numLocals,
		NumParameters:  numParams,
		SourceMap:      sourceMap,
		Name:           fnName,
		IsVariadic:     isVariadic,
		TypeParameters: typeParams,
		SymTable:       fnSymTable,
	}
	// Emits the closure onto the stack
	c.emit(code.OpClosure, c.addConstant(compiledFn), len(freeSymbols))

	if node.ReceiverName != nil {
		structNameIdx := c.addConstant(&object.String{Value: node.ReceiverType.Value})
		methodNameIdx := c.addConstant(&object.String{Value: node.Name.Value})
		c.emit(code.OpRegisterMethod, structNameIdx, methodNameIdx)
	} else {
		// Now assign the closure (which is on top of the stack) to the outer symbol.
		if outerSym.Scope == symbol.GlobalScope { // Use outerSym.Scope
			c.emit(code.OpSetGlobal, outerSym.Index)
		} else {
			c.emit(code.OpSetLocal, outerSym.Index)
		}
	}

	return nil
}

func (c *Compiler) compileAsyncFunctionStatement(node *ast.AsyncFunctionStatement) error {
	returnType := ""
	if node.ReturnType != nil {
		returnType = node.ReturnType.Value
	}

	outerSym := c.symbolTable.DefineWithType(node.Name.Value, returnType)

	c.enterScopeWithType(returnType, true)
	c.symbolTable.DefineFunctionName(node.Name.Value)

	for _, tp := range node.TypeParameters {
		c.symbolTable.DefineType(tp.Value)
	}

	var paramSymbols []symbol.Symbol
	for _, p := range node.Parameters {
		paramType := ""
		if p.Type != nil {
			paramType = p.Type.Value
		}
		sym := c.symbolTable.DefineWithType(p.Value, paramType)
		paramSymbols = append(paramSymbols, sym)
	}

	// Emit default value initialization prologue for parameters with defaults
	for i, p := range node.Parameters {
		if p.DefaultValue != nil {
			sym := paramSymbols[i]
			c.emit(code.OpGetLocal, sym.Index)
			jumpNotNullPos := c.emit(code.OpJumpNotNull, 9999)
			c.emit(code.OpPop)
			if err := c.Compile(p.DefaultValue); err != nil {
				return err
			}
			c.emit(code.OpSetLocal, sym.Index)
			jumpEndPos := c.emit(code.OpJump, 9999)
			skipPos := len(c.currentInstructions())
			c.changeOperand(jumpNotNullPos, skipPos)
			c.emit(code.OpPop)
			endPos := len(c.currentInstructions())
			c.changeOperand(jumpEndPos, endPos)
		}
	}

	// Runtime type checks for parameters with type annotations (after defaults applied)
	for i, p := range node.Parameters {
		if p.Type != nil {
			sym := paramSymbols[i]
			typeIdx := c.addConstant(&object.String{Value: p.Type.Value})
			c.emit(code.OpGetLocal, sym.Index)
			c.emit(code.OpCheckType, typeIdx)
			c.emit(code.OpPop)
		}
	}

	if err := c.Compile(node.Body); err != nil {
		return err
	}

	if c.lastInstructionIs(code.OpPop) {
		c.replaceLastPopWithReturn()
	}
	if !c.lastInstructionIs(code.OpReturnValue) {
		c.emit(code.OpReturn)
	}

	freeSymbols := c.symbolTable.FreeSymbols
	numLocals := c.symbolTable.NumDefinitions()
	fnSymTable := c.symbolTable
	instructions, sourceMap := c.leaveScope()

	for _, s := range freeSymbols {
		switch s.Scope {
		case symbol.GlobalScope:
			c.emit(code.OpGetGlobal, s.Index)
		case symbol.LocalScope:
			c.emit(code.OpGetLocal, s.Index)
		case symbol.FreeScope:
			c.emit(code.OpGetFree, s.Index)
		case symbol.FunctionScope:
			c.emit(code.OpCurrentClosure)
		}
	}

	typeParams := []string{}
	for _, tp := range node.TypeParameters {
		typeParams = append(typeParams, tp.Value)
	}

	isVariadic := false
	if len(node.Parameters) > 0 && node.Parameters[len(node.Parameters)-1].Variadic {
		isVariadic = true
	}

	compiledFn := &object.CompiledFunction{
		Instructions:   instructions,
		NumLocals:      numLocals,
		NumParameters:  len(node.Parameters),
		SourceMap:      sourceMap,
		IsAsync:        true,
		IsVariadic:     isVariadic,
		Name:           node.Name.Value,
		TypeParameters: typeParams,
		SymTable:       fnSymTable,
	}
	c.emit(code.OpClosure, c.addConstant(compiledFn), len(freeSymbols))

	if outerSym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, outerSym.Index)
	} else {
		c.emit(code.OpSetLocal, outerSym.Index)
	}

	return nil
}

func (c *Compiler) compileReturnStatement(node *ast.ReturnStatement) error {
	if node.ReturnValue != nil {
		valType := c.inferType(node.ReturnValue)
		if err := c.checkTypeMatch(c.expectedReturnType, valType, node.ReturnValue); err != nil {
			return c.errorPos("compile error: return type mismatch - %s", err)
		}

		if _, isCall := node.ReturnValue.(*ast.CallExpression); isCall {
			c.tailCallReturn = true
			if err := c.Compile(node.ReturnValue); err != nil {
				c.tailCallReturn = false
				return err
			}
			c.tailCallReturn = false
			lastOp := c.scopes[c.scopeIndex].lastInstruction
			if lastOp.Opcode == code.OpCall {
				argCount := int(c.currentInstructions()[lastOp.Position+1])
				c.replaceInstruction(lastOp.Position, code.Make(code.OpTailCall, argCount))
				c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpTailCall
				return nil
			}
			c.emit(code.OpReturnValue)
			return nil
		}

		if err := c.Compile(node.ReturnValue); err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	} else {
		if c.expectedReturnType != "" && c.expectedReturnType != "any" {
			return c.errorPos("compile error: return type mismatch - expected %s, got void", c.expectedReturnType)
		}
		c.emit(code.OpReturn)
	}
	return nil
}
