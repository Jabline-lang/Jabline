package lsp

import (
	"jabline/pkg/ast"
	"jabline/pkg/token"
)

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

	case *ast.CallExpression:
		if childPath = FindPathToNode(n.Function, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		for _, arg := range n.Arguments {
			if childPath = FindPathToNode(arg, line, col); childPath != nil {
				return append([]ast.Node{node}, childPath...)
			}
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

	case *ast.InfixExpression:
		if childPath = FindPathToNode(n.Left, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}
		if isTokenAt(n.Token, line, col) { // Operator
			return []ast.Node{node}
		}
		if childPath = FindPathToNode(n.Right, line, col); childPath != nil {
			return append([]ast.Node{node}, childPath...)
		}

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
