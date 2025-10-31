package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
	"strings"
)

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

		if result != nil {
			switch result.Type() {
			case object.BREAK_OBJ:
				return NULL
			case object.CONTINUE_OBJ:
				continue
			case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
				return result
			}
		}
	}

	return result
}

func evalForStatement(fs *ast.ForStatement, env *object.Environment) object.Object {
	if fs == nil {
		return newError("invalid for statement")
	}

	forEnv := object.NewEnclosedEnvironment(env)

	if fs.Init != nil {
		initResult := Eval(fs.Init, forEnv)
		if isError(initResult) {
			return initResult
		}
	}

	var result object.Object = NULL

	for {
		if fs.Condition != nil {
			condition := Eval(fs.Condition, forEnv)
			if isError(condition) {
				return condition
			}
			if !isTruthy(condition) {
				break
			}
		}

		result = Eval(fs.Body, forEnv)

		if result != nil {
			switch result.Type() {
			case object.BREAK_OBJ:
				return NULL
			case object.CONTINUE_OBJ:
			case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
				return result
			}
		}

		if fs.Update != nil {
			updateResult := Eval(fs.Update, forEnv)
			if isError(updateResult) {
				return updateResult
			}
		}
	}

	return result
}

func evalForEachStatement(node *ast.ForEachStatement, env *object.Environment) object.Object {
	iterable := Eval(node.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	forEnv := object.NewEnclosedEnvironment(env)
	var result object.Object = NULL

	switch iter := iterable.(type) {
	case *object.Array:
		for _, elem := range iter.Elements {
			forEnv.Set(node.Variable.Value, elem)
			result = Eval(node.Body, forEnv)
			if result != nil {
				switch result.Type() {
				case object.BREAK_OBJ:
					return NULL
				case object.CONTINUE_OBJ:
					continue
				case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
					return result
				}
			}
		}
	case *object.Hash:
		for _, pair := range iter.Pairs {
			forEnv.Set(node.Variable.Value, pair.Value)
			result = Eval(node.Body, forEnv)
			if result != nil {
				switch result.Type() {
				case object.BREAK_OBJ:
					return NULL
				case object.CONTINUE_OBJ:
					continue
				case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
					return result
				}
			}
		}
	case *object.String:
		str := iter.Value
		for _, char := range str {
			forEnv.Set(node.Variable.Value, &object.String{Value: string(char)})
			result = Eval(node.Body, forEnv)
			if result != nil {
				switch result.Type() {
				case object.BREAK_OBJ:
					return NULL
				case object.CONTINUE_OBJ:
					continue
				case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
					return result
				}
			}
		}
	default:
		return newError("object is not iterable: %T", iterable)
	}

	return result
}

func evalTryStatement(ts *ast.TryStatement, env *object.Environment) object.Object {
	if ts == nil || ts.TryBlock == nil {
		return newError("invalid try statement")
	}

	tryResult := Eval(ts.TryBlock, env)

	if exception, ok := tryResult.(*object.Exception); ok {
		if ts.CatchBlock != nil {
			catchEnv := object.NewEnclosedEnvironment(env)
			if ts.CatchParam != nil {
				if exception.Value != nil {
					catchEnv.Set(ts.CatchParam.Value, exception.Value)
				} else {
					catchEnv.Set(ts.CatchParam.Value, &object.String{Value: exception.Message})
				}
			}
			return Eval(ts.CatchBlock, catchEnv)
		}
		return exception
	}

	if isError(tryResult) {
		if ts.CatchBlock != nil {
			catchEnv := object.NewEnclosedEnvironment(env)
			if ts.CatchParam != nil {
				errorMessage := tryResult.(*object.Error).Message
				if strings.Contains(errorMessage, "] ") {
					parts := strings.Split(errorMessage, "] ")
					if len(parts) > 1 {
						mainMessage := parts[1]
						if strings.Contains(mainMessage, "\n\n") {
							mainMessage = strings.Split(mainMessage, "\n\n")[0]
						}
						errorMessage = mainMessage
					}
				}
				catchEnv.Set(ts.CatchParam.Value, &object.String{Value: errorMessage})
			}
			return Eval(ts.CatchBlock, catchEnv)
		}
		return tryResult
	}

	return tryResult
}

func evalThrowStatement(ts *ast.ThrowStatement, env *object.Environment) object.Object {
	if ts == nil || ts.Value == nil {
		return &object.Exception{Message: "null exception thrown", Value: NULL}
	}

	value := Eval(ts.Value, env)
	if isError(value) {
		return value
	}

	return &object.Exception{
		Message: "exception thrown",
		Value:   value,
	}
}

func evalSwitchStatement(ss *ast.SwitchStatement, env *object.Environment) object.Object {
	if ss == nil || ss.Expression == nil {
		return newError("invalid switch statement")
	}

	switchValue := Eval(ss.Expression, env)
	if isError(switchValue) {
		return switchValue
	}

	var matchedCase *ast.CaseClause
	var executeDefault bool = true

	for _, caseClause := range ss.Cases {
		caseValue := Eval(caseClause.Value, env)
		if isError(caseValue) {
			return caseValue
		}

		if isEqualValues(switchValue, caseValue) {
			matchedCase = caseClause
			executeDefault = false
			break
		}
	}

	var result object.Object = NULL

	if matchedCase != nil {
		var executeRemainingCases bool = false

		for _, caseClause := range ss.Cases {
			if caseClause == matchedCase {
				executeRemainingCases = true
			}
			if executeRemainingCases {
				for _, stmt := range caseClause.Statements {
					result = Eval(stmt, env)
					if result != nil {
						switch result.Type() {
						case object.BREAK_OBJ:
							return NULL
						case object.CONTINUE_OBJ:
							return newError("continue statement not allowed in switch")
						case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
							return result
						}
					}
				}
			}
		}
	}

	if executeDefault && ss.DefaultCase != nil {
		for _, stmt := range ss.DefaultCase.Statements {
			result = Eval(stmt, env)
			if result != nil {
				switch result.Type() {
				case object.BREAK_OBJ:
					return NULL
				case object.CONTINUE_OBJ:
					return newError("continue statement not allowed in switch")
				case object.RETURN_OBJ, object.ERROR_OBJ, object.EXCEPTION_OBJ:
					return result
				}
			}
		}
	}

	return result
}

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
		return true
	default:
		return left == right
	}
}
