package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
	"strings"
)

// evalWhileStatement evaluates a while loop
func evalWhileStatement(ws *ast.WhileStatement, env *object.Environment) object.Object {
	if ws == nil || ws.Condition == nil {
		return newError("invalid while statement")
	}

	var result object.Object = NULL

	for {
		condition := Eval(ws.Condition, env)
		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		result = Eval(ws.Body, env)

		// Handle control flow statements
		if result != nil {
			switch result.Type() {
			case object.BREAK_OBJ:
				return NULL // Break out of loop
			case object.CONTINUE_OBJ:
				continue // Continue to next iteration
			case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
				return result // Propagate return/error/exception
			}
		}
	}

	return result
}

// evalForStatement evaluates a for loop
func evalForStatement(fs *ast.ForStatement, env *object.Environment) object.Object {
	if fs == nil {
		return newError("invalid for statement")
	}

	// Create a new environment for the for loop
	forEnv := object.NewEnclosedEnvironment(env)

	// Evaluate initialization
	if fs.Init != nil {
		initResult := Eval(fs.Init, forEnv)
		if isError(initResult) {
			return initResult
		}
	}

	var result object.Object = NULL

	for {
		// Check condition
		if fs.Condition != nil {
			condition := Eval(fs.Condition, forEnv)
			if isError(condition) {
				return condition
			}
			if !isTruthy(condition) {
				break
			}
		}

		// Execute body
		result = Eval(fs.Body, forEnv)

		// Handle control flow statements
		if result != nil {
			switch result.Type() {
			case object.BREAK_OBJ:
				return NULL // Break out of loop
			case object.CONTINUE_OBJ:
				// Continue to update step, don't return here
			case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
				return result // Propagate return/error/exception
			}
		}

		// Execute update
		if fs.Update != nil {
			updateResult := Eval(fs.Update, forEnv)
			if isError(updateResult) {
				return updateResult
			}
		}
	}

	return result
}

// evalForEachStatement evaluates a foreach loop
func evalForEachStatement(node *ast.ForEachStatement, env *object.Environment) object.Object {
	iterable := Eval(node.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	// Create new environment for the loop
	forEnv := object.NewEnclosedEnvironment(env)
	var result object.Object = NULL

	switch iter := iterable.(type) {
	case *object.Array:
		for _, elem := range iter.Elements {
			// Set the loop variable
			forEnv.Set(node.Variable.Value, elem)

			// Execute body
			result = Eval(node.Body, forEnv)

			// Handle control flow statements
			if result != nil {
				switch result.Type() {
				case object.BREAK_OBJ:
					return NULL // Break out of loop
				case object.CONTINUE_OBJ:
					continue // Continue to next iteration
				case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
					return result // Propagate return/error/exception
				}
			}
		}
	case *object.Hash:
		for _, pair := range iter.Pairs {
			// Set the loop variable to the value
			forEnv.Set(node.Variable.Value, pair.Value)

			// Execute body
			result = Eval(node.Body, forEnv)

			// Handle control flow statements
			if result != nil {
				switch result.Type() {
				case object.BREAK_OBJ:
					return NULL // Break out of loop
				case object.CONTINUE_OBJ:
					continue // Continue to next iteration
				case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
					return result // Propagate return/error/exception
				}
			}
		}
	case *object.String:
		// Iterate over string characters
		str := iter.Value
		for _, char := range str {
			// Set the loop variable to the character
			forEnv.Set(node.Variable.Value, &object.String{Value: string(char)})

			// Execute body
			result = Eval(node.Body, forEnv)

			// Handle control flow statements
			if result != nil {
				switch result.Type() {
				case object.BREAK_OBJ:
					return NULL // Break out of loop
				case object.CONTINUE_OBJ:
					continue // Continue to next iteration
				case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
					return result // Propagate return/error/exception
				}
			}
		}
	default:
		return newError("object is not iterable: %T", iterable)
	}

	return result
}

// evalTryStatement evaluates a try-catch statement
func evalTryStatement(ts *ast.TryStatement, env *object.Environment) object.Object {
	if ts == nil || ts.TryBlock == nil {
		return newError("invalid try statement")
	}

	// Execute the try block
	tryResult := Eval(ts.TryBlock, env)

	// Check if an exception was thrown
	if exception, ok := tryResult.(*object.Exception); ok {
		// If there's a catch block, handle the exception
		if ts.CatchBlock != nil {
			// Create new environment for catch block
			catchEnv := object.NewEnclosedEnvironment(env)

			// If there's a catch parameter, bind the exception value to it
			if ts.CatchParam != nil {
				// Pass the actual exception value, not the exception object
				if exception.Value != nil {
					catchEnv.Set(ts.CatchParam.Value, exception.Value)
				} else {
					catchEnv.Set(ts.CatchParam.Value, &object.String{Value: exception.Message})
				}
			}

			// Execute catch block
			return Eval(ts.CatchBlock, catchEnv)
		}
		// No catch block, propagate the exception
		return exception
	}

	// If it's any other type of error, handle it in catch block
	if isError(tryResult) {
		if ts.CatchBlock != nil {
			// Create new environment for catch block
			catchEnv := object.NewEnclosedEnvironment(env)

			// If there's a catch parameter, bind the error message to it
			if ts.CatchParam != nil {
				// Extract and simplify the error message for the catch parameter
				errorMessage := tryResult.(*object.Error).Message

				// Try to extract just the main error message, skip formatting
				if strings.Contains(errorMessage, "] ") {
					parts := strings.Split(errorMessage, "] ")
					if len(parts) > 1 {
						// Get the main message part, before suggestions
						mainMessage := parts[1]
						if strings.Contains(mainMessage, "\n\n") {
							mainMessage = strings.Split(mainMessage, "\n\n")[0]
						}
						errorMessage = mainMessage
					}
				}

				catchEnv.Set(ts.CatchParam.Value, &object.String{Value: errorMessage})
			}

			// Execute catch block
			return Eval(ts.CatchBlock, catchEnv)
		}
		return tryResult
	}

	// No exception, return the result
	return tryResult
}

// evalThrowStatement evaluates a throw statement
func evalThrowStatement(ts *ast.ThrowStatement, env *object.Environment) object.Object {
	if ts == nil || ts.Value == nil {
		return &object.Exception{Message: "null exception thrown", Value: NULL}
	}

	// Evaluate the value to be thrown
	value := Eval(ts.Value, env)
	if isError(value) {
		return value
	}

	// Create and return an exception object
	return &object.Exception{
		Message: "exception thrown",
		Value:   value,
	}
}

// evalSwitchStatement evaluates a switch statement
func evalSwitchStatement(ss *ast.SwitchStatement, env *object.Environment) object.Object {
	if ss == nil || ss.Expression == nil {
		return newError("invalid switch statement")
	}

	// Evaluate the switch expression
	switchValue := Eval(ss.Expression, env)
	if isError(switchValue) {
		return switchValue
	}

	var matchedCase *ast.CaseClause
	var executeDefault bool = true

	// Find matching case
	for _, caseClause := range ss.Cases {
		caseValue := Eval(caseClause.Value, env)
		if isError(caseValue) {
			return caseValue
		}

		// Compare values for equality
		if isEqualValues(switchValue, caseValue) {
			matchedCase = caseClause
			executeDefault = false
			break
		}
	}

	var result object.Object = NULL

	// Execute matched case and handle fall-through
	if matchedCase != nil {
		var executeRemainingCases bool = false

		// Execute from the matched case onwards (fall-through behavior)
		for _, caseClause := range ss.Cases {
			if caseClause == matchedCase {
				executeRemainingCases = true
			}

			if executeRemainingCases {
				for _, stmt := range caseClause.Statements {
					result = Eval(stmt, env)

					// Handle control flow
					if result != nil {
						switch result.Type() {
						case object.BREAK_OBJ:
							return NULL // Break out of switch
						case object.CONTINUE_OBJ:
							// Continue is not valid in switch, treat as error
							return newError("continue statement not allowed in switch")
						case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
							return result // Propagate return/error/exception
						}
					}
				}
			}
		}
	}

	// Execute default case if no case matched
	if executeDefault && ss.DefaultCase != nil {
		for _, stmt := range ss.DefaultCase.Statements {
			result = Eval(stmt, env)

			// Handle control flow
			if result != nil {
				switch result.Type() {
				case object.BREAK_OBJ:
					return NULL // Break out of switch
				case object.CONTINUE_OBJ:
					// Continue is not valid in switch, treat as error
					return newError("continue statement not allowed in switch")
				case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
					return result // Propagate return/error/exception
				}
			}
		}
	}

	return result
}

// isEqualValues compares two objects for equality (used in switch cases)
func isEqualValues(left, right object.Object) bool {
	if left.Type() != right.Type() {
		return false
	}

	switch left.Type() {
	case object.INTEGER_OBJ:
		return left.(*object.Integer).Value == right.(*object.Integer).Value
	case object.FLOAT_OBJ:
		return left.(*object.Float).Value == right.(*object.Float).Value
	case object.STRING_OBJ:
		return left.(*object.String).Value == right.(*object.String).Value
	case object.BOOLEAN_OBJ:
		return left.(*object.Boolean).Value == right.(*object.Boolean).Value
	case object.NULL_OBJ:
		return true // both are null
	default:
		// For other types, use pointer comparison (same object instance)
		return left == right
	}
}
