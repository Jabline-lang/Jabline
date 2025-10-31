package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
)

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if builtin := getBuiltin(node.Value); builtin != nil {
		return builtin
	}

	val, ok := env.Get(node.Value)
	if !ok {
		return newNameError(node.Value, nil)
	}
	return val
}

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

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.INSTANCE_OBJ:
		return evalInstanceIndexExpression(left, index)
	default:
		return newTypeError("INSTANCE", getTypeName(left), nil)
	}
}

func evalInstanceIndexExpression(instance, index object.Object) object.Object {
	instanceObject := instance.(*object.Instance)

	var key string
	switch idx := index.(type) {
	case *object.String:
		key = idx.Value
	default:
		return newRuntimeError("invalid field access", nil)
	}

	value, ok := instanceObject.Fields[key]
	if !ok {
		return newKeyError(key, nil)
	}
	return value
}

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

func evalNullishCoalescingExpression(nce *ast.NullishCoalescingExpression, env *object.Environment) object.Object {
	if nce == nil || nce.Left == nil {
		return newRuntimeError("invalid nullish coalescing expression", nil)
	}

	left := Eval(nce.Left, env)
	if isError(left) {
		return left
	}

	if isNullish(left) {
		return Eval(nce.Right, env)
	}

	return left
}

func evalOptionalChainingExpression(oce *ast.OptionalChainingExpression, env *object.Environment) object.Object {
	if oce == nil || oce.Left == nil {
		return newRuntimeError("invalid optional chaining expression", nil)
	}

	left := Eval(oce.Left, env)
	if isError(left) {
		return left
	}

	if isNullish(left) {
		return NULL
	}

	var fieldName string
	if ident, ok := oce.Right.(*ast.Identifier); ok {
		fieldName = ident.Value
	} else {
		return newRuntimeError("invalid property access in optional chaining", nil)
	}

	return evalFieldAccess(left, fieldName)
}

func evalAwaitExpression(ae *ast.AwaitExpression, env *object.Environment) object.Object {
	if ae == nil || ae.Value == nil {
		return newRuntimeError("invalid await expression", nil)
	}

	value := Eval(ae.Value, env)
	if isError(value) {
		return value
	}

	if promise, ok := value.(*object.Promise); ok {
		result, err := GlobalEventLoop.Await(promise)
		if err != nil {
			return err
		}
		return result
	}

	return value
}

func evalTemplateLiteral(tl *ast.TemplateLiteral, env *object.Environment) object.Object {
	var result string

	for i, part := range tl.Parts {
		result += part

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
