package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	case *ast.Program:
		return evalProgram(node.Statements, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return evalExpressionStatement(node, env)

	case *ast.ReturnStatement:
		return evalReturnStatement(node, env)

	case *ast.LetStatement:
		return evalLetStatement(node, env)

	case *ast.ConstStatement:
		return evalConstStatement(node, env)

	case *ast.EchoStatement:
		return evalEchoStatement(node, env)

	case *ast.WhileStatement:
		return evalWhileStatement(node, env)

	case *ast.ForStatement:
		return evalForStatement(node, env)

	case *ast.ForEachStatement:
		return evalForEachStatement(node, env)

	case *ast.AssignmentStatement:
		return evalAssignmentStatement(node, env)

	case *ast.FunctionStatement:
		function := &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
		closureFunction := createClosureIfNeeded(function, env)
		env.Set(node.Name.Value, closureFunction)
		return closureFunction

	case *ast.AsyncFunctionStatement:
		asyncFunction := &object.AsyncFunction{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
		closureAsyncFunction := createClosureIfNeeded(asyncFunction, env)
		env.Set(node.Name.Value, closureAsyncFunction)
		return closureAsyncFunction

	case *ast.StructStatement:
		fields := make(map[string]string)
		for name, typeExpr := range node.Fields {
			fields[name] = typeExpr.Value
		}
		structDef := &object.Struct{
			Name:   node.Name.Value,
			Fields: fields,
		}
		env.Set(node.Name.Value, structDef)
		return structDef

	case *ast.BreakStatement:
		return &object.Break{}

	case *ast.ContinueStatement:
		return &object.Continue{}

	case *ast.TryStatement:
		return evalTryStatement(node, env)

	case *ast.ThrowStatement:
		return evalThrowStatement(node, env)

	case *ast.SwitchStatement:
		return evalSwitchStatement(node, env)

	case *ast.ImportStatement:
		return evalImportStatement(node, env)

	case *ast.ExportStatement:
		return evalExportStatement(node, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToPyBoolean(node.Value)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.TemplateLiteral:
		return evalTemplateLiteral(node, env)

	case *ast.Null:
		return NULL

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		if node.Operator == "&&" || node.Operator == "||" {
			return evalLogicalInfixExpression(node, env)
		}
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.PostfixExpression:
		return evalPostfixExpression(node, env)

	case *ast.TernaryExpression:
		return evalTernaryExpression(node, env)

	case *ast.NullishCoalescingExpression:
		return evalNullishCoalescingExpression(node, env)

	case *ast.OptionalChainingExpression:
		return evalOptionalChainingExpression(node, env)

	case *ast.AsyncFunctionLiteral:
		return evaluateNestedAsyncFunction(node, env)

	case *ast.AwaitExpression:
		return evalAwaitExpression(node, env)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		return evaluateNestedFunction(node, env)

	case *ast.ArrowFunction:
		return evaluateNestedArrowFunction(node, env)

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)

	case *ast.ArrayLiteral:
		elements := evalArrayElements(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.ArrayIndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalArrayIndexExpression(left, index)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		if identifier, ok := node.Index.(*ast.Identifier); ok {
			return evalFieldAccess(left, identifier.Value)
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.StructLiteral:
		structObj, ok := env.Get(node.Name.Value)
		if !ok {
			return newError("undefined struct: %s", node.Name.Value)
		}
		structDef, ok := structObj.(*object.Struct)
		if !ok {
			return newError("identifier is not a struct: %s", node.Name.Value)
		}
		fields := make(map[string]object.Object)
		for name, valueExpr := range node.Fields {
			if _, exists := structDef.Fields[name]; !exists {
				return newError("unknown field '%s' in struct %s", name, node.Name.Value)
			}
			value := Eval(valueExpr, env)
			if isError(value) {
				return value
			}
			fields[name] = value
		}
		for fieldName := range structDef.Fields {
			if _, provided := fields[fieldName]; !provided {
				return newError("missing field '%s' in struct %s", fieldName, node.Name.Value)
			}
		}
		return &object.Instance{
			StructName: node.Name.Value,
			Fields:     fields,
		}

	default:
		return newError("unknown node type: %T", node)
	}
}
