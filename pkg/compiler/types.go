package compiler

import (
	"fmt"
	"jabline/pkg/ast"
)

// inferType tries to determine the type of an AST node at compile time.
func (c *Compiler) inferType(node ast.Node) string {
	res := ""
	switch n := node.(type) {
	case *ast.IntegerLiteral:
		res = "int"
	case *ast.FloatLiteral:
		res = "float"
	case *ast.StringLiteral:
		res = "string"
	case *ast.Boolean:
		res = "bool"
	case *ast.Identifier:
		sym, ok := c.symbolTable.Resolve(n.Value)
		if ok {
			res = sym.DataType
		}
	case *ast.InfixExpression:
		res = c.inferInfixType(n)
	case *ast.CallExpression:
		res = c.inferCallType(n)
	case *ast.FunctionLiteral:
		if n.ReturnType != nil {
			res = n.ReturnType.Value
		}
	case *ast.ArrowFunction:
		if n.ReturnType != nil {
			res = n.ReturnType.Value
		}
	}
	// Uncomment for noisy debug
	// fmt.Printf("DEBUG: inferType(%T) -> %q\n", node, res)
	return res
}

func (c *Compiler) inferInfixType(node *ast.InfixExpression) string {
	leftType := c.inferType(node.Left)
	rightType := c.inferType(node.Right)

	// Basic numeric promotion
	if leftType == "float" || rightType == "float" {
		if (leftType == "int" || leftType == "float") && (rightType == "int" || rightType == "float") {
			return "float"
		}
	}

	if leftType == "int" && rightType == "int" {
		return "int"
	}

	if leftType == "string" || rightType == "string" {
		if node.Operator == "+" {
			return "string"
		}
	}

	// Boolean results for comparisons
	switch node.Operator {
	case "==", "!=", "<", ">", "<=", ">=":
		return "bool"
	}

	return leftType // Fallback
}

func (c *Compiler) inferCallType(node *ast.CallExpression) string {
	if ident, ok := node.Function.(*ast.Identifier); ok {
		sym, ok := c.symbolTable.Resolve(ident.Value)
		if ok {
			// For now, we'd need to store function return types in the symbol table.
			// This will be added in the next step.
			return sym.DataType
		}
	}
	return ""
}

func (c *Compiler) checkTypeMatch(expected, actual string, node ast.Node) error {
	// fmt.Printf("DEBUG: checkTypeMatch(expected=%q, actual=%q)\n", expected, actual)
	if expected == "" || actual == "" || expected == "any" || actual == "any" {
		return nil
	}

	if expected != actual {
		// Allow int to float promotion implicitly in some cases?
		// For now, let's be strict but mindful of the user request.
		if expected == "float" && actual == "int" {
			return nil
		}
		return fmt.Errorf("type mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}
