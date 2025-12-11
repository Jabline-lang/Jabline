package compiler

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"jabline/pkg/stdlib"
	"reflect"
)

type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable
	scopes      []CompilationScope
	scopeIndex  int
	loops       []LoopScope
	loopIndex   int
	
	currentNode ast.Node
	exports     map[string]int
}

type LoopScope struct {
	ContinuePos int
	BreakPos    []int
}

type CompilationScope struct {
	instructions        code.Instructions
	sourceMap           code.SourceMap
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		sourceMap:           make(code.SourceMap),
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()
	
	for i, v := range stdlib.Registry {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
		loops:       []LoopScope{},
		loopIndex:   -1,
		exports:     make(map[string]int),
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	c := New()
	c.symbolTable = s
	c.constants = constants
	return c
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	SymbolTable  *SymbolTable
	SourceMap    code.SourceMap
	Exports      map[string]int
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
		SymbolTable:  c.symbolTable,
		SourceMap:    c.scopes[c.scopeIndex].sourceMap,
		Exports:      c.scopes[c.scopeIndex].exports,
	}
}

func (c *Compiler) GetSymbolTable() *SymbolTable {
	return c.symbolTable
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) Compile(node ast.Node) error {

	c.currentNode = node

	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil { return err }
		}
		return nil


	case *ast.FunctionStatement: return c.compileFunctionStatement(node)
	case *ast.LetStatement: return c.compileLetStatement(node)
	case *ast.AssignmentStatement: return c.compileAssignmentStatement(node)
	case *ast.ExpressionStatement: return c.compileExpressionStatement(node)
	case *ast.BlockStatement: return c.compileBlockStatement(node)
	case *ast.ReturnStatement: return c.compileReturnStatement(node)
	case *ast.WhileStatement: return c.compileWhileStatement(node)
	case *ast.ForStatement: return c.compileForStatement(node)
	case *ast.BreakStatement: return c.compileBreakStatement(node)
	case *ast.ContinueStatement: return c.compileContinueStatement(node)
	case *ast.StructStatement: return c.compileStructStatement(node)
	case *ast.ThrowStatement: return c.compileThrowStatement(node)
	case *ast.TryStatement: return c.compileTryStatement(node)
	case *ast.SwitchStatement: return c.compileSwitchStatement(node)
	case *ast.ConstStatement: return c.compileConstStatement(node)
	case *ast.EchoStatement: return c.compileEchoStatement(node)
	case *ast.ImportStatement: return c.compileImportStatement(node)
	case *ast.ExportStatement: return c.compileExportStatement(node)


	case *ast.CallExpression: return c.compileCallExpression(node)
	case *ast.TemplateLiteral: return c.compileTemplateLiteral(node)
	case *ast.NullishCoalescingExpression: return c.compileNullishCoalescingExpression(node)
	case *ast.OptionalChainingExpression: return c.compileOptionalChainingExpression(node)
	case *ast.TernaryExpression: return c.compileTernaryExpression(node)
	case *ast.Identifier: return c.compileIdentifier(node)
	case *ast.IntegerLiteral: return c.compileIntegerLiteral(node)
	case *ast.StringLiteral: return c.compileStringLiteral(node)
	case *ast.Boolean: return c.compileBoolean(node)
	case *ast.Null: return c.compileNull(node)
	case *ast.ArrayLiteral: return c.compileArrayLiteral(node)
	case *ast.HashLiteral: return c.compileHashLiteral(node)
	case *ast.StructLiteral: return c.compileStructLiteral(node)
	case *ast.IndexExpression: return c.compileIndexExpression(node)
	case *ast.ArrayIndexExpression: return c.compileArrayIndexExpression(node)
	case *ast.PrefixExpression: return c.compilePrefixExpression(node)
	case *ast.InfixExpression: return c.compileInfixExpression(node)
	case *ast.IfExpression: return c.compileIfExpression(node)
	case *ast.FunctionLiteral: return c.compileFunctionLiteral(node)

	default:
		return fmt.Errorf("unknown node type: %T", node)
	}
}
