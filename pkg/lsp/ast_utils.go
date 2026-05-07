package lsp

import (
	"jabline/pkg/ast"
	"jabline/pkg/token"
)

// FindPathToNode performs a depth-first search through the AST and returns a path
// of nodes from the root to the node that contains the given (line, col) position.
// Line and col are 1-indexed as reported by the lexer.
func FindPathToNode(node ast.Node, line, col int) []ast.Node {
	if node == nil {
		return nil
	}

	var childPath []ast.Node

	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			if childPath = FindPathToNode(stmt, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			if childPath = FindPathToNode(stmt, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.ExpressionStatement:
		if childPath = FindPathToNode(n.Expression, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	// ── Statements ──────────────────────────────────────────────────────────

	case *ast.LetStatement:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Name, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if n.Value != nil {
			if childPath = FindPathToNode(n.Value, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.ConstStatement:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Name, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if n.Value != nil {
			if childPath = FindPathToNode(n.Value, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.ReturnStatement:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if n.ReturnValue != nil {
			if childPath = FindPathToNode(n.ReturnValue, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.AssignmentStatement:
		if childPath = FindPathToNode(n.Left, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Value, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.EchoStatement:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		for _, v := range n.Values {
			if childPath = FindPathToNode(v, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	// ── Functions ────────────────────────────────────────────────────────────

	case *ast.FunctionStatement:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Name, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		for _, p := range n.Parameters {
			if childPath = FindPathToNode(p, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}
		if childPath = FindPathToNode(n.Body, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.AsyncFunctionStatement:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Name, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		for _, p := range n.Parameters {
			if childPath = FindPathToNode(p, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}
		if childPath = FindPathToNode(n.Body, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.FunctionLiteral:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		for _, p := range n.Parameters {
			if childPath = FindPathToNode(p, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}
		if childPath = FindPathToNode(n.Body, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.AsyncFunctionLiteral:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		for _, p := range n.Parameters {
			if childPath = FindPathToNode(p, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}
		if childPath = FindPathToNode(n.Body, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.ArrowFunction:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		for _, p := range n.Parameters {
			if childPath = FindPathToNode(p, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}
		if childPath = FindPathToNode(n.Body, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	// ── Calls & Expressions ──────────────────────────────────────────────────

	case *ast.CallExpression:
		if childPath = FindPathToNode(n.Function, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		for _, arg := range n.Arguments {
			if childPath = FindPathToNode(arg, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.AwaitExpression:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Value, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.InfixExpression:
		if childPath = FindPathToNode(n.Left, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Right, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.PrefixExpression:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Right, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.PostfixExpression:
		if childPath = FindPathToNode(n.Left, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.TernaryExpression:
		if childPath = FindPathToNode(n.Condition, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.TrueValue, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.FalseValue, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.NullishCoalescingExpression:
		if childPath = FindPathToNode(n.Left, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Right, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.OptionalChainingExpression:
		if childPath = FindPathToNode(n.Left, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Right, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	// ── Control Flow ─────────────────────────────────────────────────────────

	case *ast.IfExpression:
		if childPath = FindPathToNode(n.Condition, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Consequence, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if n.Alternative != nil {
			if childPath = FindPathToNode(n.Alternative, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.WhileStatement:
		if childPath = FindPathToNode(n.Condition, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Body, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.ForStatement:
		if n.Init != nil {
			if childPath = FindPathToNode(n.Init, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}
		if n.Condition != nil {
			if childPath = FindPathToNode(n.Condition, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}
		if n.Update != nil {
			if childPath = FindPathToNode(n.Update, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}
		if childPath = FindPathToNode(n.Body, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.ForEachStatement:
		if childPath = FindPathToNode(n.Iterable, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Variable, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Body, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	// ── Data structures ──────────────────────────────────────────────────────

	case *ast.ArrayLiteral:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		for _, el := range n.Elements {
			if childPath = FindPathToNode(el, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.HashLiteral:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		for k, v := range n.Pairs {
			if childPath = FindPathToNode(k, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
			if childPath = FindPathToNode(v, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.ArrayIndexExpression:
		if childPath = FindPathToNode(n.Left, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Index, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.IndexExpression:
		if childPath = FindPathToNode(n.Left, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if childPath = FindPathToNode(n.Index, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.SpawnExpression:
		if childPath = FindPathToNode(n.Call, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	// ── Structs ───────────────────────────────────────────────────────────────

	case *ast.StructStatement:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Name, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.InterfaceStatement:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Name, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

	case *ast.StructLiteral:
		if childPath = FindPathToNode(n.Name, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		for _, v := range n.Fields {
			if childPath = FindPathToNode(v, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	// ── Terminals ────────────────────────────────────────────────────────────

	case *ast.Identifier:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}

	case *ast.IntegerLiteral:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}

	case *ast.FloatLiteral:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}

	case *ast.Boolean:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}

	case *ast.StringLiteral:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}

	case *ast.TemplateLiteral:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
		for _, expr := range n.Expressions {
			if childPath = FindPathToNode(expr, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
		}

	case *ast.Null:
		if isTokenAt(n.Token, line, col) {
			return []ast.Node{node}
		}
	}

	return nil
}

func isTokenAt(tok token.Token, line, col int) bool {
	if tok.Line != line {
		return false
	}
	start := tok.Column
	end := start + len(tok.Literal)
	return col >= start && col < end
}

// FindCurrentScope returns the deepest lexical scope associated with the current cursor AST path
func FindCurrentScope(docInfo *DocumentSemanticInfo, path []ast.Node) *Scope {
	if docInfo == nil || docInfo.SymbolTable == nil || docInfo.SymbolTable.RootScope == nil {
		return nil
	}

	var currentScope *Scope

	for i := len(path) - 1; i >= 0; i-- {
		node := path[i]

		var foundScope *Scope
		var findScopeByNode func(s *Scope, target ast.Node) *Scope
		findScopeByNode = func(s *Scope, target ast.Node) *Scope {
			if s.Node == target {
				return s
			}
			for _, child := range s.Children {
				if found := findScopeByNode(child, target); found != nil {
					return found
				}
			}
			return nil
		}

		if docInfo.SymbolTable.RootScope.Node == node {
			foundScope = docInfo.SymbolTable.RootScope
		} else {
			foundScope = findScopeByNode(docInfo.SymbolTable.RootScope, node)
		}

		if foundScope != nil {
			currentScope = foundScope
			break
		}
	}

	if currentScope == nil {
		currentScope = docInfo.SymbolTable.RootScope
	}

	return currentScope
}
