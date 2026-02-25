package compiler

import (
	"fmt"
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

	c.enterScopeWithType(returnType) // Enter the function's new scope
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
	for _, p := range node.Parameters {
		paramType := ""
		if p.Type != nil {
			paramType = p.Type.Value
		}
		sym := c.symbolTable.DefineWithType(p.Value, paramType)

		// If the parameter has a type annotation, insert runtime check
		if p.Type != nil {
			typeIdx := c.addConstant(&object.String{Value: p.Type.Value})
			c.emit(code.OpGetLocal, sym.Index)
			c.emit(code.OpCheckType, typeIdx)
			c.emit(code.OpPop) // CheckType inspects the top of stack, Pop removes it
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
	instructions := c.leaveScope()              // Exit the function's scope

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

	typeParams := []string{}
	for _, tp := range node.TypeParameters {
		typeParams = append(typeParams, tp.Value)
	}

	compiledFn := &object.CompiledFunction{
		Instructions:   instructions,
		NumLocals:      numLocals,
		NumParameters:  numParams,
		Name:           fnName,
		TypeParameters: typeParams,
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

	c.enterScopeWithType(returnType)
	c.symbolTable.DefineFunctionName(node.Name.Value)

	for _, tp := range node.TypeParameters {
		c.symbolTable.DefineType(tp.Value)
	}

	for _, p := range node.Parameters {
		paramType := ""
		if p.Type != nil {
			paramType = p.Type.Value
		}
		sym := c.symbolTable.DefineWithType(p.Value, paramType)

		if p.Type != nil {
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
	instructions := c.leaveScope()

	for _, s := range freeSymbols {
		c.emit(code.OpGetFree, s.Index)
	}

	typeParams := []string{}
	for _, tp := range node.TypeParameters {
		typeParams = append(typeParams, tp.Value)
	}

	compiledFn := &object.CompiledFunction{
		Instructions:   instructions,
		NumLocals:      numLocals,
		NumParameters:  len(node.Parameters),
		IsAsync:        true,
		Name:           node.Name.Value,
		TypeParameters: typeParams,
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
		// Static type validation
		valType := c.inferType(node.ReturnValue)
		if err := c.checkTypeMatch(c.expectedReturnType, valType, node.ReturnValue); err != nil {
			return fmt.Errorf("compile error: return type mismatch - %s", err)
		}

		if err := c.Compile(node.ReturnValue); err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	} else {
		// If return type is expected but no value provided
		if c.expectedReturnType != "" && c.expectedReturnType != "any" {
			return fmt.Errorf("compile error: return type mismatch - expected %s, got void", c.expectedReturnType)
		}
		c.emit(code.OpReturn)
	}
	return nil
}
