package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
	"jabline/pkg/token"
)

func evalAssignmentStatement(node *ast.AssignmentStatement, env *object.Environment) object.Object {
	var finalValue object.Object

	if isCompoundAssignment(node.Token.Type) {
		currentValue := getCurrentValue(node.Left, env)
		if isError(currentValue) {
			return currentValue
		}

		rightValue := Eval(node.Value, env)
		if isError(rightValue) {
			return rightValue
		}

		finalValue = performCompoundOperation(node.Token.Type, currentValue, rightValue)
		if isError(finalValue) {
			return finalValue
		}
	} else {
		finalValue = Eval(node.Value, env)
		if isError(finalValue) {
			return finalValue
		}
	}

	switch left := node.Left.(type) {
	case *ast.Identifier:
		return evalSimpleAssignment(left, finalValue, env)
	case *ast.ArrayIndexExpression:
		return evalArrayElementAssignment(left, finalValue, env)
	case *ast.IndexExpression:
		return evalFieldAssignment(left, finalValue, env)
	default:
		return newError("invalid assignment target: %T", node.Left)
	}
}

func evalSimpleAssignment(identifier *ast.Identifier, value object.Object, env *object.Environment) object.Object {
	if env.IsConstant(identifier.Value) {
		return newError("cannot reassign constant '%s'", identifier.Value)
	}

	result := env.Set(identifier.Value, value)
	if result == nil {
		return newError("cannot assign to constant '%s'", identifier.Value)
	}
	return value
}

func evalArrayElementAssignment(arrayIndex *ast.ArrayIndexExpression, value object.Object, env *object.Environment) object.Object {
	arrayObj := Eval(arrayIndex.Left, env)
	if isError(arrayObj) {
		return arrayObj
	}

	indexObj := Eval(arrayIndex.Index, env)
	if isError(indexObj) {
		return indexObj
	}

	switch container := arrayObj.(type) {
	case *object.Array:
		return evalArrayElementAssignmentDirect(container, indexObj, value, arrayIndex.Left, env)
	case *object.Hash:
		return evalHashElementAssignmentDirect(container, indexObj, value, arrayIndex.Left, env)
	default:
		return newError("index assignment not supported on %T", arrayObj)
	}
}

func evalArrayElementAssignmentDirect(array *object.Array, index, value object.Object, leftExpr ast.Expression, env *object.Environment) object.Object {
	if index.Type() != object.INTEGER_OBJ {
		return newError("array index must be integer: %s", index.Type())
	}

	idx := index.(*object.Integer).Value
	if idx < 0 || idx >= int64(len(array.Elements)) {
		return newError("array index out of bounds: %d", idx)
	}

	array.Elements[idx] = value

	if identifier, ok := leftExpr.(*ast.Identifier); ok {
		env.Set(identifier.Value, array)
	}

	return value
}

func evalHashElementAssignmentDirect(hash *object.Hash, index, value object.Object, leftExpr ast.Expression, env *object.Environment) object.Object {
	hashKey, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %T", index)
	}

	hash.Pairs[hashKey.HashKey()] = object.HashPair{Key: index, Value: value}

	if identifier, ok := leftExpr.(*ast.Identifier); ok {
		env.Set(identifier.Value, hash)
	}

	return value
}

func evalFieldAssignment(fieldAccess *ast.IndexExpression, value object.Object, env *object.Environment) object.Object {
	obj := Eval(fieldAccess.Left, env)
	if isError(obj) {
		return obj
	}

	if instance, ok := obj.(*object.Instance); ok {
		identifier, ok := fieldAccess.Index.(*ast.Identifier)
		if !ok {
			return newError("invalid field access")
		}

		fieldName := identifier.Value
		instance.Fields[fieldName] = value

		if identifier, ok := fieldAccess.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, instance)
		}

		return value
	}

	if hash, ok := obj.(*object.Hash); ok {
		key := Eval(fieldAccess.Index, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %T", key)
		}

		hash.Pairs[hashKey.HashKey()] = object.HashPair{Key: key, Value: value}

		if identifier, ok := fieldAccess.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, hash)
		}

		return value
	}

	return newError("field assignment not supported on %T", obj)
}

func evalArrayIndexPostfix(arrayIndex *ast.ArrayIndexExpression, operator string, env *object.Environment) object.Object {
	left := Eval(arrayIndex.Left, env)
	if isError(left) {
		return left
	}

	index := Eval(arrayIndex.Index, env)
	if isError(index) {
		return index
	}

	if array, ok := left.(*object.Array); ok {
		idx, ok := index.(*object.Integer)
		if !ok {
			return newError("array index must be integer: %T", index)
		}

		if idx.Value < 0 || idx.Value >= int64(len(array.Elements)) {
			return newError("array index out of bounds: %d", idx.Value)
		}

		element := array.Elements[idx.Value]
		if element.Type() != object.INTEGER_OBJ {
			return newError("postfix operator only supported on integers")
		}

		integer := element.(*object.Integer)
		currentValue := integer.Value

		var newValue int64
		switch operator {
		case "++":
			newValue = currentValue + 1
		case "--":
			newValue = currentValue - 1
		default:
			return newError("unknown postfix operator: %s", operator)
		}

		array.Elements[idx.Value] = &object.Integer{Value: newValue}

		if identifier, ok := arrayIndex.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, array)
		}

		return integer
	}

	if hash, ok := left.(*object.Hash); ok {
		hashKey, ok := index.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %T", index)
		}

		pair, exists := hash.Pairs[hashKey.HashKey()]
		if !exists {
			return newError("hash key not found")
		}

		if pair.Value.Type() != object.INTEGER_OBJ {
			return newError("postfix operator only supported on integers")
		}

		integer := pair.Value.(*object.Integer)
		originalValue := integer.Value

		var newValue int64
		switch operator {
		case "++":
			newValue = originalValue + 1
		case "--":
			newValue = originalValue - 1
		default:
			return newError("unknown postfix operator: %s", operator)
		}

		hash.Pairs[hashKey.HashKey()] = object.HashPair{
			Key:   index,
			Value: &object.Integer{Value: newValue},
		}

		if identifier, ok := arrayIndex.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, hash)
		}

		return integer
	}

	return newError("postfix operator not supported on %T", left)
}

func evalFieldAccessPostfix(fieldAccess *ast.IndexExpression, operator string, env *object.Environment) object.Object {
	left := Eval(fieldAccess.Left, env)
	if isError(left) {
		return left
	}

	if instance, ok := left.(*object.Instance); ok {
		identifier, ok := fieldAccess.Index.(*ast.Identifier)
		if !ok {
			return newError("invalid field access")
		}

		fieldName := identifier.Value
		currentValue, exists := instance.Fields[fieldName]
		if !exists {
			return newError("field not found: %s", fieldName)
		}

		if currentValue.Type() != object.INTEGER_OBJ {
			return newError("postfix operator only supported on integers")
		}

		integer := currentValue.(*object.Integer)
		originalValue := integer.Value

		var newValue int64
		switch operator {
		case "++":
			newValue = originalValue + 1
		case "--":
			newValue = originalValue - 1
		default:
			return newError("unknown postfix operator: %s", operator)
		}

		instance.Fields[fieldName] = &object.Integer{Value: newValue}

		if identifier, ok := fieldAccess.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, instance)
		}

		return integer
	}

	if hash, ok := left.(*object.Hash); ok {
		keyExpr := fieldAccess.Index
		key := Eval(keyExpr, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %T", key)
		}

		pair, exists := hash.Pairs[hashKey.HashKey()]
		if !exists {
			return newError("hash key not found")
		}

		if pair.Value.Type() != object.INTEGER_OBJ {
			return newError("postfix operator only supported on integers")
		}

		integer := pair.Value.(*object.Integer)
		originalValue := integer.Value

		var newValue int64
		switch operator {
		case "++":
			newValue = originalValue + 1
		case "--":
			newValue = originalValue - 1
		default:
			return newError("unknown postfix operator: %s", operator)
		}

		hash.Pairs[hashKey.HashKey()] = object.HashPair{
			Key:   key,
			Value: &object.Integer{Value: newValue},
		}

		if identifier, ok := fieldAccess.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, hash)
		}

		return integer
	}

	return newError("postfix operator not supported on %T", left)
}

func isCompoundAssignment(tokenType token.TokenType) bool {
	switch tokenType {
	case token.PLUS_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.DIV_ASSIGN:
		return true
	default:
		return false
	}
}

func getCurrentValue(expr ast.Expression, env *object.Environment) object.Object {
	switch target := expr.(type) {
	case *ast.Identifier:
		val, ok := env.Get(target.Value)
		if !ok {
			return newError("identifier not found: %s", target.Value)
		}
		return val
	case *ast.ArrayIndexExpression:
		array := Eval(target.Left, env)
		if isError(array) {
			return array
		}
		index := Eval(target.Index, env)
		if isError(index) {
			return index
		}
		return evalArrayIndexExpression(array, index)
	case *ast.IndexExpression:
		obj := Eval(target.Left, env)
		if isError(obj) {
			return obj
		}
		if identifier, ok := target.Index.(*ast.Identifier); ok {
			return evalFieldAccess(obj, identifier.Value)
		}
		index := Eval(target.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(obj, index)
	default:
		return newError("invalid assignment target for compound operation: %T", expr)
	}
}

func performCompoundOperation(operator token.TokenType, left, right object.Object) object.Object {
	switch operator {
	case token.PLUS_ASSIGN:
		return evalInfixExpression("+", left, right)
	case token.SUB_ASSIGN:
		return evalInfixExpression("-", left, right)
	case token.MUL_ASSIGN:
		return evalInfixExpression("*", left, right)
	case token.DIV_ASSIGN:
		return evalInfixExpression("/", left, right)
	default:
		return newError("unknown compound assignment operator: %s", operator)
	}
}
