package typechecker

import (
	"fmt"
	"jabline/pkg/ast"
)

// Checker represents the AOT type checker
type Checker struct {
	errors []string
	env    *Environment
}

// New creates a new Type Checker instance
func New() *Checker {
	return &Checker{
		errors: []string{},
		env:    NewEnvironment(),
	}
}

// Check is the main entry point to type check a program.
// Returns a list of error messages. If empty, the program is type-safe.
func (c *Checker) Check(program *ast.Program) []string {
	// First pass: register type aliases before checking statements
	for _, stmt := range program.Statements {
		if alias, ok := stmt.(*ast.TypeAliasStatement); ok {
			c.checkTypeAliasStatement(alias)
		}
	}
	// Second pass: check all statements
	for _, stmt := range program.Statements {
		c.checkStatement(stmt)
	}
	return c.errors
}

func (c *Checker) addError(node ast.Node, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if node != nil && node.GetToken().Line > 0 {
		msg = fmt.Sprintf("Line %d:%d - Type Error: %s", node.GetToken().Line, node.GetToken().Column, msg)
	} else {
		msg = "Type Error: " + msg
	}
	c.errors = append(c.errors, msg)
}

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

func (c *Checker) checkStatement(stmt ast.Statement) {
	switch node := stmt.(type) {
	case *ast.LetStatement:
		c.checkLetStatement(node)
	case *ast.ConstStatement:
		c.checkConstStatement(node)
	case *ast.ReturnStatement:
		c.checkReturnStatement(node)
	case *ast.EchoStatement:
		c.checkEchoStatement(node)
	case *ast.ExpressionStatement:
		c.checkExpression(node.Expression)
	case *ast.BlockStatement:
		c.checkBlockStatement(node)
	case *ast.FunctionStatement:
		c.checkFunctionStatement(node)
	case *ast.AsyncFunctionStatement:
		c.checkAsyncFunctionStatement(node)
	case *ast.DeferStatement:
		c.checkDeferStatement(node)
	case *ast.DoWhileStatement:
		c.checkDoWhileStatement(node)
	case *ast.WhileStatement:
		c.checkWhileStatement(node)
	case *ast.ForStatement:
		c.checkForStatement(node)
	case *ast.ForEachStatement:
		c.checkForEachStatement(node)
	case *ast.SwitchStatement:
		c.checkSwitchStatement(node)
	case *ast.MatchStatement:
		c.checkMatchStatement(node)
	case *ast.SelectStatement:
		c.checkSelectStatement(node)
	case *ast.TryStatement:
		c.checkTryStatement(node)
	case *ast.RetryStatement:
		c.checkRetryStatement(node)
	case *ast.ThrowStatement:
		c.checkThrowStatement(node)
	case *ast.BreakStatement:
		c.checkBreakStatement(node)
	case *ast.ContinueStatement:
		c.checkContinueStatement(node)
	case *ast.ImportStatement:
		c.checkImportStatement(node)
	case *ast.ExportStatement:
		c.checkExportStatement(node)
	case *ast.EnumStatement:
		c.checkEnumStatement(node)
	case *ast.StructStatement:
		c.checkStructStatement(node)
	case *ast.InterfaceStatement:
		c.checkInterfaceStatement(node)
	case *ast.ServiceStatement:
		c.checkServiceStatement(node)
	case *ast.MeterStatement:
		c.checkMeterStatement(node)
	case *ast.TraceStatement:
		c.checkTraceStatement(node)
	case *ast.AssignmentStatement:
		c.checkAssignmentStatement(node)
	case *ast.TypeAliasStatement:
		c.checkTypeAliasStatement(node)
	}
}

func (c *Checker) checkTypeAliasStatement(node *ast.TypeAliasStatement) {
	// Get the target type name from the expression (e.g. "int", "string", "MyType")
	typeName := fmt.Sprintf("%s", node.Type)
	targetType := ParseASTType(&ast.TypeExpression{Value: typeName}, c.env)
	// Check for alias cycles (A -> A) or (A -> B -> A)
	if string(targetType) == node.Name.Value {
		c.addError(node, "type alias cycle: %s refers to itself", node.Name.Value)
		return
	}
	// Check for name collisions with built-in types
	builtinCheck := ParseASTType(&ast.TypeExpression{Value: node.Name.Value}, nil)
	if builtinCheck != TypeType(node.Name.Value) {
		c.addError(node, "cannot redefine built-in type '%s' as alias", node.Name.Value)
		return
	}
	c.env.DefineAlias(node.Name.Value, targetType)
}

func (c *Checker) checkLetStatement(node *ast.LetStatement) {
	var valType TypeType = TypeAny
	if node.Value != nil {
		valType = c.checkExpression(node.Value)
	}

	// Handle destructuring: let [a, b] = arr
	if node.Destructure != nil {
		for _, field := range node.Destructure.Fields {
			if field.Value != nil {
				fieldType := TypeAny
				c.env.Set(field.Value.Value, fieldType)
			}
		}
		return
	}

	if node.Name == nil {
		return
	}

	if c.env.ExistsInCurrentScope(node.Name.Value) {
		c.addError(node, "variable '%s' shadows existing declaration in the same scope", node.Name.Value)
	}

	declaredType := TypeAny
	if node.Type != nil {
		declaredType = ParseASTType(node.Type, c.env)
	} else {
		declaredType = valType
	}

	if declaredType != TypeAny && valType != TypeAny && declaredType != valType {
		// Allow implicit numeric conversions: int/float literals to specific numeric types
		if !c.isNumericType(declaredType) || !c.isNumericType(valType) {
			c.addError(node, "type mismatch in let statement: expected %s, got %s", declaredType, valType)
		}
	}

	c.env.Set(node.Name.Value, declaredType)
}

func (c *Checker) checkConstStatement(node *ast.ConstStatement) {
	var valType TypeType = TypeAny
	if node.Value != nil {
		valType = c.checkExpression(node.Value)
	}

	// Handle destructuring: const [a, b] = arr
	if node.Destructure != nil {
		for _, field := range node.Destructure.Fields {
			if field.Value != nil {
				fieldType := TypeAny
				c.env.Set(field.Value.Value, fieldType)
			}
		}
		return
	}

	if node.Name == nil {
		return
	}

	if c.env.ExistsInCurrentScope(node.Name.Value) {
		c.addError(node, "variable '%s' shadows existing declaration in the same scope", node.Name.Value)
	}

	declaredType := TypeAny
	if node.Type != nil {
		declaredType = ParseASTType(node.Type, c.env)
	} else {
		declaredType = valType
	}

	if declaredType != TypeAny && valType != TypeAny && declaredType != valType {
		// Allow implicit numeric conversions: int/float literals to specific numeric types
		if !c.isNumericType(declaredType) || !c.isNumericType(valType) {
			c.addError(node, "type mismatch in const statement: expected %s, got %s", declaredType, valType)
		}
	}

	c.env.Set(node.Name.Value, declaredType)
	c.env.MarkConst(node.Name.Value)
}

func (c *Checker) checkReturnStatement(node *ast.ReturnStatement) {
	var valType TypeType = TypeVoid
	if node.ReturnValue != nil {
		valType = c.checkExpression(node.ReturnValue)
	}

	if c.env.expectedReturn != TypeAny && c.env.expectedReturn != valType {
		if valType != TypeAny {
			c.addError(node, "type mismatch in return: expected %s, got %s", c.env.expectedReturn, valType)
		}
	}
}

func (c *Checker) checkEchoStatement(node *ast.EchoStatement) {
	for _, val := range node.Values {
		c.checkExpression(val)
	}
}

func (c *Checker) checkBlockStatement(node *ast.BlockStatement) {
	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)

	for _, stmt := range node.Statements {
		c.checkStatement(stmt)
	}

	c.env = previousEnv
}

func (c *Checker) checkFunctionStatement(node *ast.FunctionStatement) {
	fnType := TypeFunction
	c.env.Set(node.Name.Value, fnType)

	expectedReturn := TypeAny
	if node.ReturnType != nil {
		expectedReturn = ParseASTType(node.ReturnType, c.env)
	}

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.expectedReturn = expectedReturn

	for _, param := range node.Parameters {
		paramType := TypeAny
		if param.Type != nil {
			paramType = ParseASTType(param.Type, c.env)
		}
		c.env.Set(param.Value, paramType)
	}

	c.checkBlockStatement(node.Body)

	c.env = previousEnv
}

func (c *Checker) checkAsyncFunctionStatement(node *ast.AsyncFunctionStatement) {
	fnType := TypeFunction
	c.env.Set(node.Name.Value, fnType)

	expectedReturn := TypeAny
	if node.ReturnType != nil {
		expectedReturn = ParseASTType(node.ReturnType, c.env)
	}

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.expectedReturn = expectedReturn

	for _, param := range node.Parameters {
		paramType := TypeAny
		if param.Type != nil {
			paramType = ParseASTType(param.Type, c.env)
		}
		c.env.Set(param.Value, paramType)
	}

	c.checkBlockStatement(node.Body)

	c.env = previousEnv
}

func (c *Checker) checkDeferStatement(node *ast.DeferStatement) {
	c.checkExpression(node.Call)
}

func (c *Checker) checkDoWhileStatement(node *ast.DoWhileStatement) {
	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.loopDepth++

	c.checkBlockStatement(node.Body)

	condType := c.checkExpression(node.Condition)
	if condType != TypeBool && condType != TypeAny {
		c.addError(node, "do-while condition must be a boolean, got %s", condType)
	}

	c.env = previousEnv
}

func (c *Checker) checkWhileStatement(node *ast.WhileStatement) {
	condType := c.checkExpression(node.Condition)
	if condType != TypeBool && condType != TypeAny {
		c.addError(node, "while condition must be a boolean, got %s", condType)
	}

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.loopDepth++

	c.checkBlockStatement(node.Body)

	c.env = previousEnv
}

func (c *Checker) checkForStatement(node *ast.ForStatement) {
	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.loopDepth++

	if node.Init != nil {
		c.checkStatement(node.Init)
	}

	if node.Condition != nil {
		condType := c.checkExpression(node.Condition)
		if condType != TypeBool && condType != TypeAny {
			c.addError(node, "for condition must be a boolean, got %s", condType)
		}
	}

	if node.Update != nil {
		c.checkStatement(node.Update)
	}

	c.checkBlockStatement(node.Body)

	c.env = previousEnv
}

func (c *Checker) checkForEachStatement(node *ast.ForEachStatement) {
	iterType := c.checkExpression(node.Iterable)
	if iterType != TypeArray && iterType != TypeHash && iterType != TypeString && iterType != TypeAny {
		c.addError(node, "foreach iterable must be an array, hash, or string, got %s", iterType)
	}

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.loopDepth++

	// Infer element type from iterable
	elemType := TypeAny
	if iterType == TypeString {
		elemType = TypeString
	}
	c.env.Set(node.Variable.Value, elemType)

	c.checkBlockStatement(node.Body)

	c.env = previousEnv
}

func (c *Checker) checkSwitchStatement(node *ast.SwitchStatement) {
	subjectType := c.checkExpression(node.Expression)
	isStringOrInt := subjectType == TypeString || subjectType == TypeInt ||
		subjectType == TypeFloat || subjectType == TypeAny

	for _, caseClause := range node.Cases {
		caseType := c.checkExpression(caseClause.Value)
		if caseType != TypeAny && subjectType != TypeAny &&
			caseType != subjectType && !isStringOrInt {
			c.addError(caseClause, "case type %s does not match switch subject type %s", caseType, subjectType)
		}
		for _, s := range caseClause.Statements {
			c.checkStatement(s)
		}
	}

	if node.DefaultCase != nil {
		for _, s := range node.DefaultCase.Statements {
			c.checkStatement(s)
		}
	}
}

func (c *Checker) checkMatchStatement(node *ast.MatchStatement) {
	c.checkExpression(node.Expression)

	for _, matchCase := range node.Cases {
		if !matchCase.IsDefault {
			c.checkExpression(matchCase.Pattern)
		}
		for _, s := range matchCase.Statements {
			c.checkStatement(s)
		}
	}
}

func (c *Checker) checkSelectStatement(node *ast.SelectStatement) {
	for _, cas := range node.Cases {
		chType := c.checkExpression(cas.Channel)
		if chType != TypeAny && chType != TypeChannel {
			c.addError(cas, "select case channel must be a channel type, got %s", chType)
		}
		if cas.IsSend {
			c.checkExpression(cas.Value)
		}
		for _, s := range cas.Statements {
			c.checkStatement(s)
		}
	}
	if node.DefaultCase != nil {
		for _, s := range node.DefaultCase.Statements {
			c.checkStatement(s)
		}
	}
}

func (c *Checker) checkTryStatement(node *ast.TryStatement) {
	c.checkBlockStatement(node.TryBlock)

	if node.CatchBlock != nil {
		previousEnv := c.env
		c.env = NewEnclosedEnvironment(previousEnv)

		if node.CatchParam != nil {
			c.env.Set(node.CatchParam.Value, TypeError)
		}

		c.checkBlockStatement(node.CatchBlock)

		c.env = previousEnv
	}
}

func (c *Checker) checkRetryStatement(node *ast.RetryStatement) {
	attemptsType := c.checkExpression(node.Attempts)
	if attemptsType != TypeInt && attemptsType != TypeAny {
		c.addError(node, "retry attempts must be an integer, got %s", attemptsType)
	}

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.loopDepth++

	c.checkBlockStatement(node.RetryBlock)

	c.env = previousEnv

	if node.CatchBlock != nil {
		catchEnv := NewEnclosedEnvironment(c.env)
		if node.CatchParam != nil {
			catchEnv.Set(node.CatchParam.Value, TypeError)
		}
		c.checkBlockStatement(node.CatchBlock)
	}
}

func (c *Checker) checkThrowStatement(node *ast.ThrowStatement) {
	c.checkExpression(node.Value)
}

func (c *Checker) checkBreakStatement(node *ast.BreakStatement) {
	if c.env.loopDepth <= 0 {
		c.addError(node, "break statement outside of loop")
	}
}

func (c *Checker) checkContinueStatement(node *ast.ContinueStatement) {
	if c.env.loopDepth <= 0 {
		c.addError(node, "continue statement outside of loop")
	}
}

func (c *Checker) checkImportStatement(node *ast.ImportStatement) {
	c.checkExpression(node.ModuleName)
}

func (c *Checker) checkExportStatement(node *ast.ExportStatement) {
	if node.Statement != nil {
		c.checkStatement(node.Statement)
	}
}

func (c *Checker) checkEnumStatement(node *ast.EnumStatement) {
	c.env.Set(node.Name.Value, TypeHash)
}

func (c *Checker) checkStructStatement(node *ast.StructStatement) {
	c.env.Set(node.Name.Value, TypeType(node.Name.Value))

	for _, fieldType := range node.Fields {
		ParseASTType(fieldType, c.env)
	}
}

func (c *Checker) checkInterfaceStatement(node *ast.InterfaceStatement) {
	c.env.Set(node.Name.Value, TypeType(node.Name.Value))
}

func (c *Checker) checkServiceStatement(node *ast.ServiceStatement) {
	c.env.Set(node.Name.Value, TypeType(node.Name.Value))

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.Set("this", TypeType(node.Name.Value))

	for _, fieldExpr := range node.Fields {
		c.checkExpression(fieldExpr)
	}

	for _, method := range node.Methods {
		c.checkFunctionStatement(method)
	}

	c.env = previousEnv
}

func (c *Checker) checkMeterStatement(node *ast.MeterStatement) {
	c.checkExpression(node.Name)
}

func (c *Checker) checkTraceStatement(node *ast.TraceStatement) {
	c.checkExpression(node.Name)
	if node.Body != nil {
		c.checkBlockStatement(node.Body)
	}
}

func (c *Checker) checkAssignmentStatement(node *ast.AssignmentStatement) {
	if indexExpr, ok := node.Left.(*ast.IndexExpression); ok {
		leftType := c.checkExpression(indexExpr.Left)
		c.checkExpression(indexExpr.Index)

		valType := c.checkExpression(node.Value)
		if leftType != TypeAny && valType != TypeAny && leftType != valType {
			c.addError(node, "type mismatch in assignment: expected %s, got %s", leftType, valType)
		}
		return
	}

	if arrayIndexExpr, ok := node.Left.(*ast.ArrayIndexExpression); ok {
		leftType := c.checkExpression(arrayIndexExpr.Left)
		c.checkExpression(arrayIndexExpr.Index)

		valType := c.checkExpression(node.Value)
		if leftType != TypeAny && valType != TypeAny {
			elemType := c.inferElementType(leftType)
			if elemType != TypeAny && elemType != valType {
				c.addError(node, "type mismatch in assignment: expected %s, got %s", elemType, valType)
			}
		}
		return
	}

	if ident, ok := node.Left.(*ast.Identifier); ok {
		targetType, exists := c.env.Get(ident.Value)
		if !exists {
			c.addError(node, "undefined variable '%s' in assignment", ident.Value)
			return
		}

		if c.env.IsConst(ident.Value) {
			c.addError(node, "cannot assign to const variable '%s'", ident.Value)
			return
		}

		valType := c.checkExpression(node.Value)
		if targetType != TypeAny && valType != TypeAny && targetType != valType {
			// Allow implicit numeric conversions
			if !c.isNumericType(targetType) || !c.isNumericType(valType) {
				c.addError(node, "type mismatch in assignment: expected %s, got %s", targetType, valType)
			}
		}
		return
	}

	c.addError(node, "invalid assignment target")
}

// ---------------------------------------------------------------------------
// Expressions
// ---------------------------------------------------------------------------

func (c *Checker) checkExpression(expr ast.Expression) TypeType {
	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		return TypeInt
	case *ast.FloatLiteral:
		return TypeFloat
	case *ast.StringLiteral:
		return TypeString
	case *ast.TemplateLiteral:
		return c.checkTemplateLiteral(node)
	case *ast.Boolean:
		return TypeBool
	case *ast.Null:
		return TypeNull
	case *ast.Identifier:
		return c.checkIdentifier(node)
	case *ast.InfixExpression:
		return c.checkInfixExpression(node)
	case *ast.PrefixExpression:
		return c.checkPrefixExpression(node)
	case *ast.PostfixExpression:
		return c.checkPostfixExpression(node)
	case *ast.IfExpression:
		return c.checkIfExpression(node)
	case *ast.CallExpression:
		return c.checkCallExpression(node)
	case *ast.FunctionLiteral:
		return c.checkFunctionLiteral(node)
	case *ast.AsyncFunctionLiteral:
		return c.checkAsyncFunctionLiteral(node)
	case *ast.ArrowFunction:
		return c.checkArrowFunction(node)
	case *ast.ArrayLiteral:
		return TypeArray
	case *ast.HashLiteral:
		return TypeHash
	case *ast.StructLiteral:
		return c.checkStructLiteral(node)
	case *ast.IndexExpression:
		return c.checkIndexExpression(node)
	case *ast.ArrayIndexExpression:
		return c.checkArrayIndexExpression(node)
	case *ast.SliceExpression:
		return TypeArray
	case *ast.TernaryExpression:
		return c.checkTernaryExpression(node)
	case *ast.NullishCoalescingExpression:
		return c.checkNullishCoalescingExpression(node)
	case *ast.OptionalChainingExpression:
		return c.checkOptionalChainingExpression(node)
	case *ast.SpawnExpression:
		return c.checkSpawnExpression(node)
	case *ast.AwaitExpression:
		return c.checkAwaitExpression(node)
	case *ast.InstantiatedExpression:
		return c.checkInstantiatedExpression(node)
	case *ast.SpreadExpr:
		return c.checkExpression(node.Right)
	}
	return TypeAny
}

func (c *Checker) checkIdentifier(node *ast.Identifier) TypeType {
	if t, ok := c.env.Get(node.Value); ok {
		return t
	}
	// Undefined variables are treated as Any (runtime will catch it)
	return TypeAny
}

func (c *Checker) checkTemplateLiteral(node *ast.TemplateLiteral) TypeType {
	for _, expr := range node.Expressions {
		c.checkExpression(expr)
	}
	return TypeString
}

func (c *Checker) checkInfixExpression(node *ast.InfixExpression) TypeType {
	leftType := c.checkExpression(node.Left)
	rightType := c.checkExpression(node.Right)

	switch node.Operator {
	case "+", "-", "*", "/":
		return c.checkArithmeticOp(node, leftType, rightType)
	case "%":
		return c.checkModOp(node, leftType, rightType)
	case "==", "!=":
		return c.checkEqualityOp(node, leftType, rightType)
	case "<", ">", "<=", ">=":
		return c.checkComparisonOp(node, leftType, rightType)
	case "&&", "||":
		return c.checkLogicalOp(node, leftType, rightType)
	case "&", "|", "^", "<<", ">>":
		return c.checkBitwiseOp(node, leftType, rightType)
	case "<-":
		return c.checkChannelSend(node, leftType, rightType)
	case "??":
		return leftType
	}

	return TypeAny
}

func (c *Checker) checkArithmeticOp(node ast.Node, leftType, rightType TypeType) TypeType {
	if leftType == TypeInt && rightType == TypeInt {
		return TypeInt
	}
	if (leftType == TypeFloat || leftType == TypeInt) && (rightType == TypeFloat || rightType == TypeInt) {
		return TypeFloat
	}
	if node.(*ast.InfixExpression).Operator == "+" && leftType == TypeString && rightType == TypeString {
		return TypeString
	}
	if leftType != TypeAny && rightType != TypeAny {
		c.addError(node, "invalid arithmetic operation: %s %s %s", leftType, node.(*ast.InfixExpression).Operator, rightType)
	}
	return TypeAny
}

func (c *Checker) checkModOp(node ast.Node, leftType, rightType TypeType) TypeType {
	if leftType == TypeInt && rightType == TypeInt {
		return TypeInt
	}
	if leftType != TypeAny && rightType != TypeAny {
		c.addError(node, "invalid operation: %% on %s and %s", leftType, rightType)
	}
	return TypeAny
}

func (c *Checker) checkEqualityOp(node ast.Node, leftType, rightType TypeType) TypeType {
	if leftType != TypeAny && rightType != TypeAny && leftType != rightType {
		// Allow numeric type mixing
		if !c.areNumericTypes(leftType, rightType) {
			c.addError(node, "invalid comparison: %s %s %s", leftType, node.(*ast.InfixExpression).Operator, rightType)
		}
	}
	return TypeBool
}

func (c *Checker) checkComparisonOp(node ast.Node, leftType, rightType TypeType) TypeType {
	if leftType != TypeAny && rightType != TypeAny && leftType != rightType {
		if !c.areNumericTypes(leftType, rightType) {
			c.addError(node, "invalid comparison: %s %s %s", leftType, node.(*ast.InfixExpression).Operator, rightType)
		}
	}
	return TypeBool
}

func (c *Checker) checkLogicalOp(node ast.Node, leftType, rightType TypeType) TypeType {
	if leftType != TypeBool && leftType != TypeAny {
		c.addError(node, "left operand of %s must be boolean, got %s", node.(*ast.InfixExpression).Operator, leftType)
	}
	if rightType != TypeBool && rightType != TypeAny {
		c.addError(node, "right operand of %s must be boolean, got %s", node.(*ast.InfixExpression).Operator, rightType)
	}
	return TypeBool
}

func (c *Checker) checkBitwiseOp(node ast.Node, leftType, rightType TypeType) TypeType {
	if leftType == TypeInt && rightType == TypeInt {
		return TypeInt
	}
	if leftType != TypeAny && rightType != TypeAny {
		c.addError(node, "invalid bitwise operation: %s %s %s", leftType, node.(*ast.InfixExpression).Operator, rightType)
	}
	return TypeAny
}

func (c *Checker) checkChannelSend(node ast.Node, leftType, rightType TypeType) TypeType {
	return TypeChannel
}

func (c *Checker) checkPrefixExpression(node *ast.PrefixExpression) TypeType {
	rightType := c.checkExpression(node.Right)

	switch node.Operator {
	case "!":
		if rightType != TypeBool && rightType != TypeAny {
			c.addError(node, "invalid operation: !%s", rightType)
		}
		return TypeBool
	case "-":
		if rightType != TypeInt && rightType != TypeFloat && rightType != TypeAny {
			c.addError(node, "invalid operation: -%s", rightType)
		}
		return rightType
	case "~":
		if rightType != TypeInt && rightType != TypeAny {
			c.addError(node, "invalid operation: ~%s", rightType)
		}
		return TypeInt
	case "<-":
		return TypeAny
	}

	return TypeAny
}

func (c *Checker) checkPostfixExpression(node *ast.PostfixExpression) TypeType {
	if ident, ok := node.Left.(*ast.Identifier); ok {
		targetType, exists := c.env.Get(ident.Value)
		if !exists {
			c.addError(node, "undefined variable '%s' in postfix expression", ident.Value)
			return TypeAny
		}
		if targetType != TypeInt && targetType != TypeFloat && targetType != TypeAny {
			c.addError(node, "invalid postfix operation on type %s", targetType)
		}
		return targetType
	}
	c.addError(node, "postfix operator only supported for identifiers")
	return TypeAny
}

func (c *Checker) checkIfExpression(node *ast.IfExpression) TypeType {
	condType := c.checkExpression(node.Condition)
	if condType != TypeBool && condType != TypeAny {
		c.addError(node, "if condition must be a boolean, got %s", condType)
	}

	c.checkBlockStatement(node.Consequence)
	if node.Alternative != nil {
		c.checkBlockStatement(node.Alternative)
	}

	return TypeAny
}

func (c *Checker) checkCallExpression(node *ast.CallExpression) TypeType {
	c.checkExpression(node.Function)

	for _, arg := range node.Arguments {
		c.checkExpression(arg)
	}

	return TypeAny
}

func (c *Checker) checkFunctionLiteral(node *ast.FunctionLiteral) TypeType {
	expectedReturn := TypeAny
	if node.ReturnType != nil {
		expectedReturn = ParseASTType(node.ReturnType, c.env)
	}

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.expectedReturn = expectedReturn

	for _, param := range node.Parameters {
		paramType := TypeAny
		if param.Type != nil {
			paramType = ParseASTType(param.Type, c.env)
		}
		c.env.Set(param.Value, paramType)
	}

	c.checkBlockStatement(node.Body)
	c.env = previousEnv

	return TypeFunction
}

func (c *Checker) checkAsyncFunctionLiteral(node *ast.AsyncFunctionLiteral) TypeType {
	expectedReturn := TypeAny
	if node.ReturnType != nil {
		expectedReturn = ParseASTType(node.ReturnType, c.env)
	}

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.expectedReturn = expectedReturn

	for _, param := range node.Parameters {
		paramType := TypeAny
		if param.Type != nil {
			paramType = ParseASTType(param.Type, c.env)
		}
		c.env.Set(param.Value, paramType)
	}

	c.checkBlockStatement(node.Body)
	c.env = previousEnv

	return TypeFunction
}

func (c *Checker) checkArrowFunction(node *ast.ArrowFunction) TypeType {
	expectedReturn := TypeAny
	if node.ReturnType != nil {
		expectedReturn = ParseASTType(node.ReturnType, c.env)
	}

	previousEnv := c.env
	c.env = NewEnclosedEnvironment(previousEnv)
	c.env.expectedReturn = expectedReturn

	for _, param := range node.Parameters {
		paramType := TypeAny
		if param.Type != nil {
			paramType = ParseASTType(param.Type, c.env)
		}
		c.env.Set(param.Value, paramType)
	}

	bodyType := c.checkExpression(node.Body)
	if expectedReturn != TypeAny && bodyType != TypeAny && expectedReturn != bodyType {
		c.addError(node, "arrow function return type mismatch: expected %s, got %s", expectedReturn, bodyType)
	}

	c.env = previousEnv

	return TypeFunction
}

func (c *Checker) checkStructLiteral(node *ast.StructLiteral) TypeType {
	if ident, ok := node.Name.(*ast.Identifier); ok {
		structType, exists := c.env.Get(ident.Value)
		if !exists {
			c.addError(node, "undefined struct type '%s'", ident.Value)
			return TypeAny
		}
		return structType
	}
	c.checkExpression(node.Name)
	for _, fieldValue := range node.Fields {
		c.checkExpression(fieldValue)
	}
	return TypeAny
}

func (c *Checker) checkIndexExpression(node *ast.IndexExpression) TypeType {
	leftType := c.checkExpression(node.Left)
	c.checkExpression(node.Index)

	if leftType == TypeHash || leftType == TypeAny {
		return TypeAny
	}
	// If it's a struct type, we could look up the field type
	return TypeAny
}

func (c *Checker) checkArrayIndexExpression(node *ast.ArrayIndexExpression) TypeType {
	leftType := c.checkExpression(node.Left)
	indexType := c.checkExpression(node.Index)

	if leftType == TypeHash {
		if indexType != TypeString && indexType != TypeAny {
			c.addError(node, "hash index must be a string, got %s", indexType)
		}
		return TypeAny
	}

	if indexType != TypeInt && indexType != TypeAny {
		c.addError(node, "array index must be an integer, got %s", indexType)
	}

	if leftType != TypeArray && leftType != TypeString && leftType != TypeAny {
		c.addError(node, "cannot index non-array type %s", leftType)
	}

	return c.inferElementType(leftType)
}

func (c *Checker) checkTernaryExpression(node *ast.TernaryExpression) TypeType {
	condType := c.checkExpression(node.Condition)
	if condType != TypeBool && condType != TypeAny {
		c.addError(node, "ternary condition must be a boolean, got %s", condType)
	}

	trueType := c.checkExpression(node.TrueValue)
	falseType := c.checkExpression(node.FalseValue)

	if trueType != TypeAny && falseType != TypeAny && trueType != falseType {
		c.addError(node, "ternary branches must have the same type, got %s and %s", trueType, falseType)
	}

	// Return the more specific type
	if trueType != TypeAny {
		return trueType
	}
	return falseType
}

func (c *Checker) checkNullishCoalescingExpression(node *ast.NullishCoalescingExpression) TypeType {
	leftType := c.checkExpression(node.Left)
	rightType := c.checkExpression(node.Right)

	if leftType != TypeAny && rightType != TypeAny && leftType != rightType {
		// Allow null ?? anything
		if leftType != TypeNull && rightType != TypeNull {
			c.addError(node, "nullish coalescing operands must have the same type, got %s and %s", leftType, rightType)
		}
	}

	if leftType != TypeNull && leftType != TypeAny {
		return leftType
	}
	return rightType
}

func (c *Checker) checkOptionalChainingExpression(node *ast.OptionalChainingExpression) TypeType {
	c.checkExpression(node.Left)
	c.checkExpression(node.Right)
	return TypeAny
}

func (c *Checker) checkSpawnExpression(node *ast.SpawnExpression) TypeType {
	return c.checkCallExpression(node.Call)
}

func (c *Checker) checkAwaitExpression(node *ast.AwaitExpression) TypeType {
	return c.checkExpression(node.Value)
}

func (c *Checker) checkInstantiatedExpression(node *ast.InstantiatedExpression) TypeType {
	c.checkExpression(node.Left)
	for _, arg := range node.TypeArguments {
		ParseASTType(arg, c.env)
	}
	return TypeFunction
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (c *Checker) areNumericTypes(a, b TypeType) bool {
	return c.isNumericType(a) && c.isNumericType(b)
}

func (c *Checker) isNumericType(t TypeType) bool {
	switch t {
	case TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeUint8, TypeUint16, TypeUint32, TypeUint64,
		TypeFloat, TypeFloat32, TypeFloat64:
		return true
	}
	return false
}

func (c *Checker) inferElementType(containerType TypeType) TypeType {
	switch containerType {
	case TypeString:
		return TypeString
	case TypeArray:
		return TypeAny
	default:
		return TypeAny
	}
}
