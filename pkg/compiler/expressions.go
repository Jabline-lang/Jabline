package compiler

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"jabline/pkg/symbol" // New import
)

func (c *Compiler) compileTemplateLiteral(node *ast.TemplateLiteral) error { return nil }
func (c *Compiler) compileNullishCoalescingExpression(node *ast.NullishCoalescingExpression) error {
	if err := c.Compile(node.Left); err != nil {
		return err
	}

	// OpJumpNotNull peeks the stack. If not null, it jumps.
	jumpNotNullPos := c.emit(code.OpJumpNotNull, 9999)

	c.emit(code.OpPop) // Pop the null
	if err := c.Compile(node.Right); err != nil {
		return err
	}

	afterRightPos := len(c.currentInstructions())
	c.changeOperand(jumpNotNullPos, afterRightPos)

	return nil
}
func (c *Compiler) compileOptionalChainingExpression(node *ast.OptionalChainingExpression) error {
	if err := c.Compile(node.Left); err != nil {
		return err
	}

	// a?.b
	// If a is null, return null.
	jumpNotNullPos := c.emit(code.OpJumpNotNull, 9999)

	// Here a is null. Just leave it or replace with explicit null if needed.
	// Actually OpJumpNotNull peeks. So Null is still on stack.
	jumpEndPos := c.emit(code.OpJump, 9999)

	afterNullPos := len(c.currentInstructions())
	c.changeOperand(jumpNotNullPos, afterNullPos)

	// Here a is NOT null.
	if err := c.Compile(node.Right); err != nil { // Right is usually Identifier or Index
		return err
	}

	afterRightPos := len(c.currentInstructions())
	c.changeOperand(jumpEndPos, afterRightPos)

	return nil
}
func (c *Compiler) compileTernaryExpression(node *ast.TernaryExpression) error {
	if err := c.Compile(node.Condition); err != nil {
		return err
	}

	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

	if err := c.Compile(node.TrueValue); err != nil {
		return err
	}

	jumpAfterPos := c.emit(code.OpJump, 9999)

	afterConsequencePos := len(c.currentInstructions())
	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

	if err := c.Compile(node.FalseValue); err != nil {
		return err
	}

	afterAlternativePos := len(c.currentInstructions())
	c.changeOperand(jumpAfterPos, afterAlternativePos)

	return nil
}

func (c *Compiler) compileIdentifier(node *ast.Identifier) error {
	sym, ok := c.symbolTable.Resolve(node.Value) // Renamed variable
	if !ok {
		return fmt.Errorf("undefined variable %s", node.Value)
	}

	switch sym.Scope { // Use sym.Scope
	case symbol.GlobalScope:
		c.emit(code.OpGetGlobal, sym.Index)
	case symbol.LocalScope:
		c.emit(code.OpGetLocal, sym.Index)
	case symbol.BuiltinScope:
		c.emit(code.OpGetBuiltin, sym.Index)
	case symbol.FreeScope:
		c.emit(code.OpGetFree, sym.Index)
	case symbol.FunctionScope:
		c.emit(code.OpCurrentClosure)
	}
	return nil
}

func (c *Compiler) compileInstantiatedExpression(node *ast.InstantiatedExpression) error {
	// 1. Compile the base expression (Function or Struct)
	if err := c.Compile(node.Left); err != nil {
		return err
	}

	// 2. Compile type arguments as constants
	for _, arg := range node.TypeArguments {
		argIdx := c.addConstant(&object.String{Value: arg.String()})
		c.emit(code.OpConstant, argIdx)
	}

	// 3. Emit OpInstantiate with the number of type arguments
	c.emit(code.OpInstantiate, len(node.TypeArguments))

	return nil
}

func (c *Compiler) compileIndexExpression(node *ast.IndexExpression) error {
	if err := c.Compile(node.Left); err != nil {
		return err
	}
	if err := c.Compile(node.Index); err != nil {
		return err
	}
	c.emit(code.OpIndex)
	return nil
}

func (c *Compiler) compileArrayIndexExpression(node *ast.ArrayIndexExpression) error {
	if err := c.Compile(node.Left); err != nil {
		return err
	}
	if err := c.Compile(node.Index); err != nil {
		return err
	}
	c.emit(code.OpIndex) // OpIndex is generic for both array and hash indexing
	return nil
}

func (c *Compiler) compilePrefixExpression(node *ast.PrefixExpression) error {
	if err := c.Compile(node.Right); err != nil {
		return err
	}

	switch node.Operator {
	case "!":
		c.emit(code.OpBang)
	case "-":
		c.emit(code.OpMinus)
	case "~":
		c.emit(code.OpBitNot)
	case "<-":
		c.emit(code.OpRecvChannel)
	default:
		return fmt.Errorf("unknown operator %s", node.Operator)
	}

	return nil
}

func (c *Compiler) compileInfixExpression(node *ast.InfixExpression) error {
	if node.Operator == "<" {
		// Reorder operands for < because we only have OpGreaterThan
		if err := c.Compile(node.Right); err != nil {
			return err
		}
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		c.emit(code.OpGreaterThan)
		return nil
	}

	if node.Operator == ">=" {
		// a >= b is !(a < b) -> !(b > a)
		if err := c.Compile(node.Right); err != nil {
			return err
		}
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		c.emit(code.OpGreaterThan)
		c.emit(code.OpBang)
		return nil
	}

	if node.Operator == "<=" {
		// a <= b is !(a > b)
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Right); err != nil {
			return err
		}
		c.emit(code.OpGreaterThan)
		c.emit(code.OpBang)
		return nil
	}

	if node.Operator == "&&" {
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		jumpPos := c.emit(code.OpJumpNotTruthyKeep, 9999)
		c.emit(code.OpPop)
		if err := c.Compile(node.Right); err != nil {
			return err
		}
		afterRightPos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterRightPos)
		return nil
	}

	if node.Operator == "||" {
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		jumpPos := c.emit(code.OpJumpTruthyKeep, 9999)
		c.emit(code.OpPop)
		if err := c.Compile(node.Right); err != nil {
			return err
		}
		afterRightPos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterRightPos)
		return nil
	}

	if err := c.Compile(node.Left); err != nil {
		return err
	}
	if err := c.Compile(node.Right); err != nil {
		return err
	}

	switch node.Operator {
	case "+":
		c.emit(code.OpAdd)
	case "-":
		c.emit(code.OpSub)
	case "*":
		c.emit(code.OpMul)
	case "/":
		c.emit(code.OpDiv)
	case "%":
		c.emit(code.OpMod)
	case ">":
		c.emit(code.OpGreaterThan)
	case "==":
		c.emit(code.OpEqual)
	case "!=":
		c.emit(code.OpNotEqual)
	case "&":
		c.emit(code.OpBitAnd)
	case "|":
		c.emit(code.OpBitOr)
	case "^":
		c.emit(code.OpBitXor)
	case "<<":
		c.emit(code.OpShiftLeft)
	case ">>":
		c.emit(code.OpShiftRight)
	case "<-":
		c.emit(code.OpSendChannel)
	default:
		return fmt.Errorf("unknown operator %s", node.Operator)
	}

	return nil
}

func (c *Compiler) compileIfExpression(node *ast.IfExpression) error {
	if err := c.Compile(node.Condition); err != nil {
		return err
	}

	// Emit an `OpJumpNotTruthy` with a bogus value
	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

	if err := c.Compile(node.Consequence); err != nil {
		return err
	}

	// If there is an `else`, we need to jump over it if the `if` was true
	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}

	// Emit an `OpJump` with a bogus value
	jumpPos := c.emit(code.OpJump, 9999)

	afterConsequencePos := len(c.currentInstructions())
	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

	if node.Alternative == nil {
		// If no else, and condition false, we return Null (expression)
		c.emit(code.OpNull)
	} else {
		if err := c.Compile(node.Alternative); err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
	}

	afterAlternativePos := len(c.currentInstructions())
	c.changeOperand(jumpPos, afterAlternativePos)

	return nil
}

func (c *Compiler) compileCallExpression(node *ast.CallExpression) error {
	if err := c.Compile(node.Function); err != nil {
		return err
	}

	for _, arg := range node.Arguments {
		if err := c.Compile(arg); err != nil {
			return err
		}
	}

	c.emit(code.OpCall, len(node.Arguments))
	// c.emit(code.OpPop) // Removed to prevent stack underflow if needed, but standard is to keep it if it's expression statement.
	// Actually, CallExpression is an expression. It pushes a value.
	// If it is used as a statement, compileExpressionStatement emits OpPop.
	// So compileCallExpression should NOT emit OpPop.
	return nil
}

func (c *Compiler) compileFunctionLiteral(node *ast.FunctionLiteral) error {
	returnType := ""
	if node.ReturnType != nil {
		returnType = node.ReturnType.Value
	}

	c.enterScopeWithType(returnType)

	for _, tp := range node.TypeParameters {
		c.symbolTable.DefineType(tp.Value)
	}

	for _, p := range node.Parameters {
		paramType := ""
		if p.Type != nil {
			paramType = p.Type.Value
		}
		c.symbolTable.DefineWithType(p.Value, paramType)
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
	numLocals := c.symbolTable.NumDefinitions() // Access via getter // Corrected
	instructions := c.leaveScope()

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

	compiledFn := &object.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		NumParameters: len(node.Parameters),
	}
	c.emit(code.OpClosure, c.addConstant(compiledFn), len(freeSymbols))

	return nil
}

func (c *Compiler) compileAsyncFunctionLiteral(node *ast.AsyncFunctionLiteral) error {
	returnType := ""
	if node.ReturnType != nil {
		returnType = node.ReturnType.Value
	}

	c.enterScopeWithType(returnType)

	for _, tp := range node.TypeParameters {
		c.symbolTable.DefineType(tp.Value)
	}

	for _, p := range node.Parameters {
		paramType := ""
		if p.Type != nil {
			paramType = p.Type.Value
		}
		c.symbolTable.DefineWithType(p.Value, paramType)
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

	compiledFn := &object.CompiledFunction{
		Instructions:   instructions,
		NumLocals:      numLocals,
		NumParameters:  len(node.Parameters),
		IsAsync:        true,
		TypeParameters: typeParams,
	}
	c.emit(code.OpClosure, c.addConstant(compiledFn), len(freeSymbols))

	return nil
}

func (c *Compiler) compileArrowFunction(node *ast.ArrowFunction) error {
	returnType := ""
	if node.ReturnType != nil {
		returnType = node.ReturnType.Value
	}

	c.enterScopeWithType(returnType)

	for _, p := range node.Parameters {
		paramType := ""
		if p.Type != nil {
			paramType = p.Type.Value
		}
		c.symbolTable.DefineWithType(p.Value, paramType)
	}

	// For arrow functions, the body is an expression
	if err := c.Compile(node.Body); err != nil {
		return err
	}

	// Validate expression type against return type if specified
	bodyType := c.inferType(node.Body)
	if err := c.checkTypeMatch(returnType, bodyType, node.Body); err != nil {
		return fmt.Errorf("compile error: arrow function return type mismatch - %s", err)
	}

	c.emit(code.OpReturnValue)

	freeSymbols := c.symbolTable.FreeSymbols
	numLocals := c.symbolTable.NumDefinitions()
	instructions := c.leaveScope()

	for _, s := range freeSymbols {
		switch s.Scope {
		case symbol.GlobalScope:
			c.emit(code.OpGetGlobal, s.Index)
		case symbol.LocalScope:
			c.emit(code.OpGetLocal, s.Index)
		case symbol.FreeScope:
			c.emit(code.OpGetFree, s.Index)
		case symbol.FunctionScope:
			c.emit(code.OpGetLocal, s.Index)
		}
	}

	compiledFn := &object.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		NumParameters: len(node.Parameters),
	}
	c.emit(code.OpClosure, c.addConstant(compiledFn), len(freeSymbols))

	return nil
}

func (c *Compiler) compileAwaitExpression(node *ast.AwaitExpression) error {
	if err := c.Compile(node.Value); err != nil {
		return err
	}
	c.emit(code.OpAwait)
	return nil
}

func (c *Compiler) compileSpawnExpression(node *ast.SpawnExpression) error {
	// Compile the callee, pushing to stack.
	if err := c.Compile(node.Call.Function); err != nil {
		return err
	}

	// Compile arguments, pushing to stack.
	for _, a := range node.Call.Arguments {
		if err := c.Compile(a); err != nil {
			return err
		}
	}

	// Emit spawn, defining number of arguments.
	c.emit(code.OpSpawn, len(node.Call.Arguments))
	return nil
}

func (c *Compiler) compileStructLiteral(node *ast.StructLiteral) error {
	// 1. Compile the Name expression (which should push a STRUCT object onto the stack)
	if err := c.Compile(node.Name); err != nil {
		return err
	}

	// 2. Compile and push field keys and values
	// Sort field names to ensure deterministic bytecode (optional but good)
	for fieldName, fieldValue := range node.Fields {
		// Push the field name as a string constant
		fieldNameObj := &object.String{Value: fieldName}
		fieldNameIdx := c.addConstant(fieldNameObj)
		c.emit(code.OpConstant, fieldNameIdx)

		// Compile the field value
		if err := c.Compile(fieldValue); err != nil {
			return err
		}
	}

	// Emit OpInstance with the number of fields as the operand
	c.emit(code.OpInstance, len(node.Fields))

	return nil
}
