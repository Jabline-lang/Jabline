package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
)

// evalArrayElements evaluates all elements in an array literal
func evalArrayElements(elems []ast.Expression, env *object.Environment) []object.Object {
	result := []object.Object{}

	for _, e := range elems {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// evalArrayIndexExpression evaluates array index access
func evalArrayIndexExpression(array, index object.Object) object.Object {
	switch {
	case array.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndex(array, index)
	case array.Type() == object.HASH_OBJ:
		return evalHashIndex(array, index)
	default:
		return newError("index operator not supported: %s[%s]", array.Type(), index.Type())
	}
}

// evalArrayIndex evaluates array indexing with integer index
func evalArrayIndex(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

// evalHashLiteral evaluates a hash literal
func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %T", key)
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

// evalHashIndex evaluates hash indexing
func evalHashIndex(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %T", index)
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

// evalArrayAssignment handles assignment to array elements
func evalArrayAssignment(array, index, value object.Object) object.Object {
	switch array.Type() {
	case object.ARRAY_OBJ:
		if index.Type() != object.INTEGER_OBJ {
			return newError("array index must be integer: %s", index.Type())
		}

		arrayObj := array.(*object.Array)
		idx := index.(*object.Integer).Value

		if idx < 0 || idx >= int64(len(arrayObj.Elements)) {
			return newError("array index out of bounds: %d", idx)
		}

		// Create a new array with the updated element
		newElements := make([]object.Object, len(arrayObj.Elements))
		copy(newElements, arrayObj.Elements)
		newElements[idx] = value

		// Return the new array (immutable semantics)
		return &object.Array{Elements: newElements}

	case object.HASH_OBJ:
		hashObj := array.(*object.Hash)
		hashKey, ok := index.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %T", index)
		}

		// Create a new hash with the updated pair
		newPairs := make(map[object.HashKey]object.HashPair)
		for k, v := range hashObj.Pairs {
			newPairs[k] = v
		}
		newPairs[hashKey.HashKey()] = object.HashPair{Key: index, Value: value}

		return &object.Hash{Pairs: newPairs}

	default:
		return newError("index assignment not supported: %s", array.Type())
	}
}
