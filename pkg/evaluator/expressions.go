package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
)

// evalIdentifier evaluates an identifier by looking it up in the environment or built-ins
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	// First check if it's a built-in function
	if builtin := getBuiltin(node.Value); builtin != nil {
		return builtin
	}

	// Then check the environment
	val, ok := env.Get(node.Value)
	if !ok {
		return newNameError(node.Value, nil)
	}
	return val
}

// evalIfExpression evaluates an if expression
func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	if ie == nil || ie.Condition == nil {
		return newRuntimeError("invalid if expression", nil)
	}

	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

// evalIndexExpression evaluates index expressions (for field access)
func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.INSTANCE_OBJ:
		return evalInstanceIndexExpression(left, index)
	default:
		return newTypeError("INSTANCE", getTypeName(left), nil)
	}
}

// evalInstanceIndexExpression evaluates index expressions on instances (struct field access)
func evalInstanceIndexExpression(instance, index object.Object) object.Object {
	instanceObject := instance.(*object.Instance)

	var key string
	switch idx := index.(type) {
	case *object.String:
		key = idx.Value
	default:
		// For field access, the index should be treated as a field name
		// We get the literal value from the AST node
		return newRuntimeError("invalid field access", nil)
	}

	value, ok := instanceObject.Fields[key]
	if !ok {
		return newKeyError(key, nil)
	}
	return value
}

// evalFieldAccess evaluates field access on objects
func evalFieldAccess(left object.Object, fieldName string) object.Object {
	switch obj := left.(type) {
	case *object.Instance:
		value, ok := obj.Fields[fieldName]
		if !ok {
			return newKeyError(fieldName, nil)
		}
		return value
	default:
		return newTypeError("INSTANCE", getTypeName(left), nil)
	}
}

// evalTernaryExpression evaluates ternary conditional expressions
func evalTernaryExpression(te *ast.TernaryExpression, env *object.Environment) object.Object {
	if te == nil || te.Condition == nil {
		return newRuntimeError("invalid ternary expression", nil)
	}

	condition := Eval(te.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(te.TrueValue, env)
	} else {
		return Eval(te.FalseValue, env)
	}
}

// evalNullishCoalescingExpression evaluates nullish coalescing expressions like a ?? b
func evalNullishCoalescingExpression(nce *ast.NullishCoalescingExpression, env *object.Environment) object.Object {
	if nce == nil || nce.Left == nil {
		return newRuntimeError("invalid nullish coalescing expression", nil)
	}

	left := Eval(nce.Left, env)
	if isError(left) {
		return left
	}

	// Only evaluate right side if left is null
	if isNullish(left) {
		return Eval(nce.Right, env)
	}

	return left
}

// evalOptionalChainingExpression evaluates optional chaining expressions like obj?.prop
func evalOptionalChainingExpression(oce *ast.OptionalChainingExpression, env *object.Environment) object.Object {
	if oce == nil || oce.Left == nil {
		return newRuntimeError("invalid optional chaining expression", nil)
	}

	left := Eval(oce.Left, env)
	if isError(left) {
		return left
	}

	// If left is null, return null without trying to access the property
	if isNullish(left) {
		return NULL
	}

	// Extract field name from the right side (should be an identifier)
	var fieldName string
	if ident, ok := oce.Right.(*ast.Identifier); ok {
		fieldName = ident.Value
	} else {
		return newRuntimeError("invalid property access in optional chaining", nil)
	}

	// Evaluate field access on the non-null object
	return evalFieldAccess(left, fieldName)
}

// evalAwaitExpression evaluates await expressions like await somePromise
func evalAwaitExpression(ae *ast.AwaitExpression, env *object.Environment) object.Object {
	if ae == nil || ae.Value == nil {
		return newRuntimeError("invalid await expression", nil)
	}

	// Evaluate the expression to await
	value := Eval(ae.Value, env)
	if isError(value) {
		return value
	}

	// If it's a Promise, await it
	if promise, ok := value.(*object.Promise); ok {
		// Use the global event loop to await the promise
		result, err := GlobalEventLoop.Await(promise)
		if err != nil {
			return err // Return the rejection reason as an error
		}
		return result
	}

	// If it's not a Promise, return it immediately
	return value
}

// evalTemplateLiteral evaluates template literals with interpolation
func evalTemplateLiteral(tl *ast.TemplateLiteral, env *object.Environment) object.Object {
	var result string

	for i, part := range tl.Parts {
		result += part

		// Add interpolated expression if exists
		if i < len(tl.Expressions) {
			value := Eval(tl.Expressions[i], env)
			if isError(value) {
				return value
			}
			result += value.Inspect()
		}
	}

	return &object.String{Value: result}
}
