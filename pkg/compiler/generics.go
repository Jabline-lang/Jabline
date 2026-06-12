package compiler

import (
	"strings"

	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

type MonomorphizeCache struct {
	instantiatedStructs   map[string]*instantiatedStructEntry
	instantiatedFunctions map[string]*instantiatedFuncEntry
}

type instantiatedStructEntry struct {
	structObj *object.Struct
}

type instantiatedFuncEntry struct {
	closure   *object.Closure
	constants []object.Object
}

func NewMonomorphizeCache() *MonomorphizeCache {
	return &MonomorphizeCache{
		instantiatedStructs:   make(map[string]*instantiatedStructEntry),
		instantiatedFunctions: make(map[string]*instantiatedFuncEntry),
	}
}

func GetInstantiatedStructKey(name string, typeArgs []string) string {
	return name + "[" + strings.Join(typeArgs, ", ") + "]"
}

func GetInstantiatedFuncKey(name string, typeArgs []string) string {
	return name + "[" + strings.Join(typeArgs, ", ") + "]"
}

func GenerateTypeArgsMap(typeParams []string, typeArgs []string) map[string]string {
	m := make(map[string]string, len(typeParams))
	for i, tp := range typeParams {
		if i < len(typeArgs) {
			m[tp] = typeArgs[i]
		}
	}
	return m
}

func SubstituteTypeInExpr(expr ast.Expression, typeArgs map[string]string) ast.Expression {
	switch e := expr.(type) {
	case *ast.Identifier:
		if concrete, ok := typeArgs[e.Value]; ok {
			return &ast.Identifier{Value: concrete}
		}
		return e
	case *ast.InfixExpression:
		return &ast.InfixExpression{
			Left:     SubstituteTypeInExpr(e.Left, typeArgs),
			Operator: e.Operator,
			Right:    SubstituteTypeInExpr(e.Right, typeArgs),
		}
	case *ast.PrefixExpression:
		return &ast.PrefixExpression{
			Operator: e.Operator,
			Right:    SubstituteTypeInExpr(e.Right, typeArgs),
		}
	case *ast.FunctionLiteral:
		return &ast.FunctionLiteral{
			Parameters:     e.Parameters,
			ReturnType:     substituteTypeExpr(e.ReturnType, typeArgs),
			Body:           substituteBlock(e.Body, typeArgs),
			TypeParameters: nil,
		}
	case *ast.CallExpression:
		subArgs := make([]ast.Expression, len(e.Arguments))
		for i, a := range e.Arguments {
			subArgs[i] = SubstituteTypeInExpr(a, typeArgs)
		}
		return &ast.CallExpression{
			Function:  SubstituteTypeInExpr(e.Function, typeArgs),
			Arguments: subArgs,
		}
	case *ast.IndexExpression:
		return &ast.IndexExpression{
			Left:  SubstituteTypeInExpr(e.Left, typeArgs),
			Index: SubstituteTypeInExpr(e.Index, typeArgs),
		}
	case *ast.ArrayLiteral:
		sub := make([]ast.Expression, len(e.Elements))
		for i, el := range e.Elements {
			sub[i] = SubstituteTypeInExpr(el, typeArgs)
		}
		return &ast.ArrayLiteral{Elements: sub}
	case *ast.HashLiteral:
		sub := make(map[ast.Expression]ast.Expression)
		for k, v := range e.Pairs {
			sub[SubstituteTypeInExpr(k, typeArgs)] = SubstituteTypeInExpr(v, typeArgs)
		}
		return &ast.HashLiteral{Pairs: sub}
	case *ast.IfExpression:
		return &ast.IfExpression{
			Condition:   SubstituteTypeInExpr(e.Condition, typeArgs),
			Consequence: substituteBlock(e.Consequence, typeArgs),
			Alternative: substituteBlock(e.Alternative, typeArgs),
		}
	case *ast.InstantiatedExpression:
		return SubstituteTypeInExpr(e.Left, typeArgs)
	case *ast.IntegerLiteral:
	case *ast.StringLiteral:
	case *ast.FloatLiteral:
	case *ast.Boolean:
	case *ast.Null:
	case *ast.PostfixExpression:
		return &ast.PostfixExpression{
			Left:     SubstituteTypeInExpr(e.Left, typeArgs),
			Operator: e.Operator,
		}
	case *ast.ArrayIndexExpression:
		return &ast.ArrayIndexExpression{
			Left:  SubstituteTypeInExpr(e.Left, typeArgs),
			Index: SubstituteTypeInExpr(e.Index, typeArgs),
		}
	case *ast.TernaryExpression:
		return &ast.TernaryExpression{
			Condition:   SubstituteTypeInExpr(e.Condition, typeArgs),
			TrueValue:   SubstituteTypeInExpr(e.TrueValue, typeArgs),
			FalseValue:  SubstituteTypeInExpr(e.FalseValue, typeArgs),
		}
	case *ast.NullishCoalescingExpression:
		return &ast.NullishCoalescingExpression{
			Left:  SubstituteTypeInExpr(e.Left, typeArgs),
			Right: SubstituteTypeInExpr(e.Right, typeArgs),
		}
	case *ast.OptionalChainingExpression:
		return &ast.OptionalChainingExpression{
			Left:  SubstituteTypeInExpr(e.Left, typeArgs),
			Right: SubstituteTypeInExpr(e.Right, typeArgs),
		}
	case *ast.SpawnExpression:
		return &ast.SpawnExpression{
			Call: SubstituteTypeInExpr(e.Call, typeArgs).(*ast.CallExpression),
		}
	case *ast.TypeExpression:
		return substituteTypeExpr(e, typeArgs)
	case *ast.StructLiteral:
		subFields := make(map[string]ast.Expression, len(e.Fields))
		for k, v := range e.Fields {
			subFields[k] = SubstituteTypeInExpr(v, typeArgs)
		}
		return &ast.StructLiteral{
			Name:   SubstituteTypeInExpr(e.Name, typeArgs),
			Fields: subFields,
		}
	case *ast.TemplateLiteral:
		subExprs := make([]ast.Expression, len(e.Expressions))
		for i, exp := range e.Expressions {
			subExprs[i] = SubstituteTypeInExpr(exp, typeArgs)
		}
		return &ast.TemplateLiteral{
			Parts:       e.Parts,
			Expressions: subExprs,
		}
	case *ast.ArrowFunction:
		return &ast.ArrowFunction{
			Parameters:     e.Parameters,
			ReturnType:     substituteTypeExpr(e.ReturnType, typeArgs),
			Body:           SubstituteTypeInExpr(e.Body, typeArgs),
		}
	case *ast.AsyncFunctionLiteral:
		return &ast.AsyncFunctionLiteral{
			Parameters:     e.Parameters,
			ReturnType:     substituteTypeExpr(e.ReturnType, typeArgs),
			Body:           substituteBlock(e.Body, typeArgs),
		}
	case *ast.AwaitExpression:
		return &ast.AwaitExpression{
			Value: SubstituteTypeInExpr(e.Value, typeArgs),
		}
	default:
	}
	return expr
}

func substituteBlock(block *ast.BlockStatement, typeArgs map[string]string) *ast.BlockStatement {
	if block == nil {
		return nil
	}
	sub := &ast.BlockStatement{}
	for _, stmt := range block.Statements {
		sub.Statements = append(sub.Statements, substituteStatement(stmt, typeArgs))
	}
	return sub
}

func substituteStatement(stmt ast.Statement, typeArgs map[string]string) ast.Statement {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return &ast.ExpressionStatement{
			Expression: SubstituteTypeInExpr(s.Expression, typeArgs),
		}
	case *ast.ReturnStatement:
		return &ast.ReturnStatement{
			ReturnValue: SubstituteTypeInExpr(s.ReturnValue, typeArgs),
		}
	case *ast.LetStatement:
		return &ast.LetStatement{
			Name:        s.Name,
			Destructure: s.Destructure,
			Type:        substituteTypeExpr(s.Type, typeArgs),
			Value:       SubstituteTypeInExpr(s.Value, typeArgs),
		}
	case *ast.ConstStatement:
		return &ast.ConstStatement{
			Name:        s.Name,
			Destructure: s.Destructure,
			Type:        substituteTypeExpr(s.Type, typeArgs),
			Value:       SubstituteTypeInExpr(s.Value, typeArgs),
		}
	case *ast.AssignmentStatement:
		return &ast.AssignmentStatement{
			Left:  SubstituteTypeInExpr(s.Left, typeArgs),
			Value: SubstituteTypeInExpr(s.Value, typeArgs),
		}
	case *ast.BlockStatement:
		return substituteBlock(s, typeArgs)
	case *ast.DoWhileStatement:
		return &ast.DoWhileStatement{
			Body:      substituteBlock(s.Body, typeArgs),
			Condition: SubstituteTypeInExpr(s.Condition, typeArgs),
		}
	case *ast.DeferStatement:
		return &ast.DeferStatement{
			Call: SubstituteTypeInExpr(s.Call, typeArgs),
		}
	case *ast.WhileStatement:
		return &ast.WhileStatement{
			Condition: SubstituteTypeInExpr(s.Condition, typeArgs),
			Body:      substituteBlock(s.Body, typeArgs),
		}
	case *ast.ForStatement:
		return &ast.ForStatement{
			Init:      substituteStatement(s.Init, typeArgs),
			Condition: SubstituteTypeInExpr(s.Condition, typeArgs),
			Update:    substituteStatement(s.Update, typeArgs),
			Body:      substituteBlock(s.Body, typeArgs),
		}
	case *ast.ForEachStatement:
		return &ast.ForEachStatement{
			Variable: s.Variable,
			Iterable: SubstituteTypeInExpr(s.Iterable, typeArgs),
			Body:     substituteBlock(s.Body, typeArgs),
		}
	case *ast.TryStatement:
		return &ast.TryStatement{
			TryBlock:   substituteBlock(s.TryBlock, typeArgs),
			CatchParam: s.CatchParam,
			CatchType:  substituteTypeExpr(s.CatchType, typeArgs),
			CatchBlock: substituteBlock(s.CatchBlock, typeArgs),
			Finally:    substituteBlock(s.Finally, typeArgs),
		}
	case *ast.ThrowStatement:
		return &ast.ThrowStatement{
			Value: SubstituteTypeInExpr(s.Value, typeArgs),
		}
	case *ast.RetryStatement:
		return &ast.RetryStatement{
			Attempts:   SubstituteTypeInExpr(s.Attempts, typeArgs),
			RetryBlock: substituteBlock(s.RetryBlock, typeArgs),
			CatchBlock: substituteBlock(s.CatchBlock, typeArgs),
			CatchParam: s.CatchParam,
		}
	case *ast.MeterStatement:
		return &ast.MeterStatement{
			Name:     SubstituteTypeInExpr(s.Name, typeArgs),
			Operator: s.Operator,
		}
	case *ast.TraceStatement:
		return &ast.TraceStatement{
			Name: SubstituteTypeInExpr(s.Name, typeArgs),
			Body: substituteBlock(s.Body, typeArgs),
		}
	case *ast.StructStatement:
		subFields := make(map[string]*ast.TypeExpression, len(s.Fields))
		for k, v := range s.Fields {
			subFields[k] = substituteTypeExpr(v, typeArgs)
		}
		return &ast.StructStatement{
			Name:   s.Name,
			TypeParameters: s.TypeParameters,
			Fields: subFields,
		}
	case *ast.InterfaceStatement:
		subMethods := make(map[string]*ast.FunctionSignature, len(s.Methods))
		for k, sig := range s.Methods {
			subParams := make([]*ast.Identifier, len(sig.Parameters))
			for i, p := range sig.Parameters {
				subParams[i] = p
			}
			subMethods[k] = &ast.FunctionSignature{
				Name:       sig.Name,
				Parameters: subParams,
				ReturnType: substituteTypeExpr(sig.ReturnType, typeArgs),
			}
		}
		return &ast.InterfaceStatement{
			Name:   s.Name,
			TypeParameters: s.TypeParameters,
			Methods: subMethods,
		}
	case *ast.ServiceStatement:
		subFields := make(map[string]ast.Expression, len(s.Fields))
		for k, v := range s.Fields {
			subFields[k] = SubstituteTypeInExpr(v, typeArgs)
		}
		subMethods := make([]*ast.FunctionStatement, len(s.Methods))
		for i, m := range s.Methods {
			subMethods[i] = substituteStatement(m, typeArgs).(*ast.FunctionStatement)
		}
		return &ast.ServiceStatement{
			Name:    s.Name,
			Fields:  subFields,
			Methods: subMethods,
		}
	case *ast.ImportStatement:
		return s
	case *ast.ReExportStatement:
		return s
	case *ast.ExportStatement:
		if s.Statement != nil {
			return &ast.ExportStatement{
				ExportType:     s.ExportType,
				Statement:      substituteStatement(s.Statement, typeArgs),
				ExportList:     s.ExportList,
				ModuleName:     s.ModuleName,
				NamespaceAlias: s.NamespaceAlias,
				IsDefault:      s.IsDefault,
			}
		}
		return s
	case *ast.FunctionStatement:
		subParams := make([]*ast.Identifier, len(s.Parameters))
		for i, p := range s.Parameters {
			subParams[i] = p
		}
		return &ast.FunctionStatement{
			ReceiverName:   s.ReceiverName,
			ReceiverType:   s.ReceiverType,
			Name:           s.Name,
			TypeParameters: nil,
			Parameters:     subParams,
			ReturnType:     substituteTypeExpr(s.ReturnType, typeArgs),
			Body:           substituteBlock(s.Body, typeArgs),
		}
	case *ast.AsyncFunctionStatement:
		return &ast.AsyncFunctionStatement{
			Name:           s.Name,
			TypeParameters: nil,
			Parameters:     s.Parameters,
			ReturnType:     substituteTypeExpr(s.ReturnType, typeArgs),
			Body:           substituteBlock(s.Body, typeArgs),
		}
	case *ast.EnumStatement:
		return s
	case *ast.BreakStatement:
	case *ast.ContinueStatement:
	case *ast.EchoStatement:
		subVals := make([]ast.Expression, len(s.Values))
		for i, v := range s.Values {
			subVals[i] = SubstituteTypeInExpr(v, typeArgs)
		}
		return &ast.EchoStatement{
			Values: subVals,
		}
	case *ast.MatchStatement:
		subCases := make([]*ast.MatchCase, len(s.Cases))
		for i, mc := range s.Cases {
			subStmts := make([]ast.Statement, len(mc.Statements))
			for j, st := range mc.Statements {
				subStmts[j] = substituteStatement(st, typeArgs)
			}
			subCases[i] = &ast.MatchCase{
				Pattern:    SubstituteTypeInExpr(mc.Pattern, typeArgs),
				IsDefault:  mc.IsDefault,
				Statements: subStmts,
			}
		}
		return &ast.MatchStatement{
			Expression: SubstituteTypeInExpr(s.Expression, typeArgs),
			Cases:      subCases,
		}
	case *ast.SwitchStatement:
		subCases := make([]*ast.CaseClause, len(s.Cases))
		for i, cc := range s.Cases {
			subStmts := make([]ast.Statement, len(cc.Statements))
			for j, st := range cc.Statements {
				subStmts[j] = substituteStatement(st, typeArgs)
			}
			subCases[i] = &ast.CaseClause{
				Value:      SubstituteTypeInExpr(cc.Value, typeArgs),
				Statements: subStmts,
			}
		}
		var subDefault *ast.DefaultClause
		if s.DefaultCase != nil {
			subStmts := make([]ast.Statement, len(s.DefaultCase.Statements))
			for j, st := range s.DefaultCase.Statements {
				subStmts[j] = substituteStatement(st, typeArgs)
			}
			subDefault = &ast.DefaultClause{
				Statements: subStmts,
			}
		}
		return &ast.SwitchStatement{
			Expression:  SubstituteTypeInExpr(s.Expression, typeArgs),
			Cases:       subCases,
			DefaultCase: subDefault,
		}
	case *ast.SelectStatement:
		subCases := make([]*ast.SelectCase, len(s.Cases))
		for i, cas := range s.Cases {
			subStmts := make([]ast.Statement, len(cas.Statements))
			for j, st := range cas.Statements {
				subStmts[j] = substituteStatement(st, typeArgs)
			}
			subCases[i] = &ast.SelectCase{
				Channel:      SubstituteTypeInExpr(cas.Channel, typeArgs),
				Value:        SubstituteTypeInExpr(cas.Value, typeArgs),
				BindingName:  cas.BindingName,
				BindingIsNew: cas.BindingIsNew,
				IsSend:       cas.IsSend,
				Statements:   subStmts,
			}
		}
		var subDefault *ast.DefaultClause
		if s.DefaultCase != nil {
			subStmts := make([]ast.Statement, len(s.DefaultCase.Statements))
			for j, st := range s.DefaultCase.Statements {
				subStmts[j] = substituteStatement(st, typeArgs)
			}
			subDefault = &ast.DefaultClause{
				Statements: subStmts,
			}
		}
		return &ast.SelectStatement{
			Cases:       subCases,
			DefaultCase: subDefault,
		}
	default:
	}
	return stmt
}

func substituteTypeExpr(te *ast.TypeExpression, typeArgs map[string]string) *ast.TypeExpression {
	if te == nil {
		return nil
	}
	if concrete, ok := typeArgs[te.Value]; ok {
		return &ast.TypeExpression{Value: concrete}
	}
	subArgs := make([]*ast.TypeExpression, len(te.Arguments))
	for i, arg := range te.Arguments {
		subArgs[i] = substituteTypeExpr(arg, typeArgs)
	}
	return &ast.TypeExpression{
		Value:     te.Value,
		Arguments: subArgs,
	}
}

func substituteTypeName(name string, typeArgs map[string]string) string {
	if name == "" {
		return ""
	}
	if concrete, ok := typeArgs[name]; ok {
		return concrete
	}
	return name
}

func (c *Compiler) CompileMonomorphizedStruct(structObj *object.Struct, typeArgs []string) error {
	key := GetInstantiatedStructKey(structObj.Name, typeArgs)

	for idx, existing := range c.constants {
		if is, ok := existing.(*object.InstantiatedStruct); ok {
			if is.FullTypeName == key {
				c.emit(code.OpConstant, idx)
				return nil
			}
		}
	}

	typeArgsMap := GenerateTypeArgsMap(structObj.TypeParameters, typeArgs)
	fullTypeName := key

	subFields := make(map[string]string)
	for name, fieldType := range structObj.Fields {
		subFields[name] = substituteTypeName(fieldType, typeArgsMap)
	}

	monoStruct := &object.Struct{
		Name:           fullTypeName,
		Fields:         subFields,
		TypeParameters: nil,
	}
	instStruct := &object.InstantiatedStruct{
		Struct:       monoStruct,
		FullTypeName: fullTypeName,
	}

	constIdx := c.addConstant(instStruct)
	c.emit(code.OpConstant, constIdx)
	return nil
}
