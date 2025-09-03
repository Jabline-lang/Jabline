package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
	"jabline/pkg/token"
)

// evalAssignmentStatement evaluates assignment statements (=, +=, -=, etc.)
func evalAssignmentStatement(node *ast.AssignmentStatement, env *object.Environment) object.Object {
	// For compound assignments, we need to get the current value first
	var finalValue object.Object

	if isCompoundAssignment(node.Token.Type) {
		// Get current value for compound assignment
		currentValue := getCurrentValue(node.Left, env)
		if isError(currentValue) {
			return currentValue
		}

		// Evaluate the right side value
		rightValue := Eval(node.Value, env)
		if isError(rightValue) {
			return rightValue
		}

		// Perform the compound operation
		finalValue = performCompoundOperation(node.Token.Type, currentValue, rightValue)
		if isError(finalValue) {
			return finalValue
		}
	} else {
		// Simple assignment - evaluate the right side value
		finalValue = Eval(node.Value, env)
		if isError(finalValue) {
			return finalValue
		}
	}

	// Check what kind of assignment target this is
	switch left := node.Left.(type) {
	case *ast.Identifier:
		// Simple variable assignment: x = value
		return evalSimpleAssignment(left, finalValue, env)
	case *ast.ArrayIndexExpression:
		// Array element assignment: arr[index] = value
		return evalArrayElementAssignment(left, finalValue, env)
	case *ast.IndexExpression:
		// Field assignment: obj.field = value or hash[key] = value
		return evalFieldAssignment(left, finalValue, env)
	default:
		return newError("invalid assignment target: %T", node.Left)
	}
}

// evalSimpleAssignment handles simple variable assignments
func evalSimpleAssignment(identifier *ast.Identifier, value object.Object, env *object.Environment) object.Object {
	// Check if it's a constant (cannot be reassigned)
	if env.IsConstant(identifier.Value) {
		return newError("cannot reassign constant '%s'", identifier.Value)
	}

	result := env.Set(identifier.Value, value)
	if result == nil {
		return newError("cannot assign to constant '%s'", identifier.Value)
	}
	return value
}

// evalArrayElementAssignment handles array element assignments
func evalArrayElementAssignment(arrayIndex *ast.ArrayIndexExpression, value object.Object, env *object.Environment) object.Object {
	// Get the array/hash object
	arrayObj := Eval(arrayIndex.Left, env)
	if isError(arrayObj) {
		return arrayObj
	}

	// Get the index
	indexObj := Eval(arrayIndex.Index, env)
	if isError(indexObj) {
		return indexObj
	}

	// Perform the assignment based on the container type
	switch container := arrayObj.(type) {
	case *object.Array:
		return evalArrayElementAssignmentDirect(container, indexObj, value, arrayIndex.Left, env)
	case *object.Hash:
		return evalHashElementAssignmentDirect(container, indexObj, value, arrayIndex.Left, env)
	default:
		return newError("index assignment not supported on %T", arrayObj)
	}
}

// evalArrayElementAssignmentDirect handles direct array element assignment
func evalArrayElementAssignmentDirect(array *object.Array, index, value object.Object, leftExpr ast.Expression, env *object.Environment) object.Object {
	if index.Type() != object.INTEGER_OBJ {
		return newError("array index must be integer: %s", index.Type())
	}

	idx := index.(*object.Integer).Value
	if idx < 0 || idx >= int64(len(array.Elements)) {
		return newError("array index out of bounds: %d", idx)
	}

	// Modify the array in place
	array.Elements[idx] = value

	// Update the variable in the environment if it's an identifier
	if identifier, ok := leftExpr.(*ast.Identifier); ok {
		env.Set(identifier.Value, array)
	}

	return value
}

// evalHashElementAssignmentDirect handles direct hash element assignment
func evalHashElementAssignmentDirect(hash *object.Hash, index, value object.Object, leftExpr ast.Expression, env *object.Environment) object.Object {
	hashKey, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %T", index)
	}

	// Modify the hash in place
	hash.Pairs[hashKey.HashKey()] = object.HashPair{Key: index, Value: value}

	// Update the variable in the environment if it's an identifier
	if identifier, ok := leftExpr.(*ast.Identifier); ok {
		env.Set(identifier.Value, hash)
	}

	return value
}

// evalFieldAssignment handles field assignments (obj.field = value)
func evalFieldAssignment(fieldAccess *ast.IndexExpression, value object.Object, env *object.Environment) object.Object {
	// Get the object
	obj := Eval(fieldAccess.Left, env)
	if isError(obj) {
		return obj
	}

	// Handle struct field assignment
	if instance, ok := obj.(*object.Instance); ok {
		identifier, ok := fieldAccess.Index.(*ast.Identifier)
		if !ok {
			return newError("invalid field access")
		}

		fieldName := identifier.Value
		instance.Fields[fieldName] = value

		// Update the variable in the environment if the left side is an identifier
		if identifier, ok := fieldAccess.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, instance)
		}

		return value
	}

	// Handle hash field assignment
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

		// Update the variable in the environment if the left side is an identifier
		if identifier, ok := fieldAccess.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, hash)
		}

		return value
	}

	return newError("field assignment not supported on %T", obj)
}

// evalArrayIndexPostfix handles postfix operations on array elements
func evalArrayIndexPostfix(arrayIndex *ast.ArrayIndexExpression, operator string, env *object.Environment) object.Object {
	left := Eval(arrayIndex.Left, env)
	if isError(left) {
		return left
	}

	index := Eval(arrayIndex.Index, env)
	if isError(index) {
		return index
	}

	// Handle arrays
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

		// Update the array element
		array.Elements[idx.Value] = &object.Integer{Value: newValue}

		// Update the variable in the environment if it's an identifier
		if identifier, ok := arrayIndex.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, array)
		}

		return integer // Return original value (postfix behavior)
	}

	// Handle hash maps
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

		// Update the hash value
		hash.Pairs[hashKey.HashKey()] = object.HashPair{
			Key:   index,
			Value: &object.Integer{Value: newValue},
		}

		// Update the variable in the environment if it's an identifier
		if identifier, ok := arrayIndex.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, hash)
		}

		return integer // Return original value
	}

	return newError("postfix operator not supported on %T", left)
}

// evalFieldAccessPostfix handles postfix operations on field access
func evalFieldAccessPostfix(fieldAccess *ast.IndexExpression, operator string, env *object.Environment) object.Object {
	left := Eval(fieldAccess.Left, env)
	if isError(left) {
		return left
	}

	// Handle struct field access
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

		// Update the field
		instance.Fields[fieldName] = &object.Integer{Value: newValue}

		// Update the variable in the environment if it's an identifier
		if identifier, ok := fieldAccess.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, instance)
		}

		return integer // Return original value
	}

	// Handle hash map access
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

		// Update the hash value
		hash.Pairs[hashKey.HashKey()] = object.HashPair{
			Key:   key,
			Value: &object.Integer{Value: newValue},
		}

		// Update the variable in the environment if it's an identifier
		if identifier, ok := fieldAccess.Left.(*ast.Identifier); ok {
			env.Set(identifier.Value, hash)
		}

		return integer // Return original value
	}

	return newError("postfix operator not supported on %T", left)
}

// isCompoundAssignment checks if the token represents a compound assignment operator
func isCompoundAssignment(tokenType token.TokenType) bool {
	switch tokenType {
	case token.PLUS_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.DIV_ASSIGN:
		return true
	default:
		return false
	}
}

// getCurrentValue gets the current value of an assignment target
func getCurrentValue(expr ast.Expression, env *object.Environment) object.Object {
	switch target := expr.(type) {
	case *ast.Identifier:
		val, ok := env.Get(target.Value)
		if !ok {
			return newError("identifier not found: " + target.Value)
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

		// Handle field access
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

// performCompoundOperation performs the arithmetic operation for compound assignments
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
