package compiler

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"jabline/pkg/stdlib"
	"jabline/pkg/symbol" // New import
)

type Compiler struct {
	constants   []object.Object
	symbolTable *symbol.SymbolTable // Updated type
	scopes      []CompilationScope
	scopeIndex  int
	loops       []LoopScope
	loopIndex   int

	currentNode        ast.Node
	exports            map[string]int
	expectedReturnType string
}

type LoopScope struct {
	ContinuePos   int
	BreakPos      []int
	ContinueJumps []int // Jumps to patch for continue statements
}

type CompilationScope struct {
	instructions        code.Instructions
	sourceMap           code.SourceMap
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	expectedReturnType  string
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

	symbolTable := symbol.NewSymbolTable() // Corrected

	for i, v := range stdlib.Registry {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	c := &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
		loops:       []LoopScope{},
		loopIndex:   -1,
		exports:     make(map[string]int),
	}

	return c
}

func NewWithState(s *symbol.SymbolTable, constants []object.Object) *Compiler { // Corrected
	c := New()
	c.symbolTable = s
	c.constants = constants
	return c
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	SymbolTable  *symbol.SymbolTable // Corrected
	SourceMap    code.SourceMap
	Exports      map[string]int
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) setInstructions(ins code.Instructions) {
	c.scopes[c.scopeIndex].instructions = ins
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
		SymbolTable:  c.symbolTable,
		SourceMap:    c.scopes[c.scopeIndex].sourceMap,
		Exports:      c.exports,
	}
}

func (c *Compiler) GetSymbolTable() *symbol.SymbolTable {
	return c.symbolTable
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.setInstructions(updatedInstructions)
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := c.currentInstructions()[opPos]
	newInstruction := code.Make(code.Opcode(op), operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	newIns := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = newIns
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) enterScope() {
	c.enterScopeWithType("")
}

func (c *Compiler) enterScopeWithType(expectedReturn string) {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		sourceMap:           make(code.SourceMap),
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		expectedReturnType:  expectedReturn,
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.expectedReturnType = expectedReturn
	c.symbolTable = symbol.NewEnclosedSymbolTable(c.symbolTable) // Corrected
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer

	if c.scopeIndex >= 0 {
		c.expectedReturnType = c.scopes[c.scopeIndex].expectedReturnType
	} else {
		c.expectedReturnType = ""
	}

	return instructions
}

func (c *Compiler) enterLoop(continuePos int) {
	loop := LoopScope{
		ContinuePos:   continuePos,
		BreakPos:      []int{},
		ContinueJumps: []int{},
	}
	c.loops = append(c.loops, loop)
	c.loopIndex++
}

func (c *Compiler) leaveLoop() LoopScope {
	loop := c.loops[c.loopIndex]
	c.loops = c.loops[:c.loopIndex]
	c.loopIndex--
	return loop
}

func (c *Compiler) Compile(node ast.Node) error {

	c.currentNode = node

	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
		// Remove the last OpPop instruction if it exists for the program
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
		return nil

	case *ast.FunctionStatement:
		return c.compileFunctionStatement(node)
	case *ast.LetStatement:
		return c.compileLetStatement(node)
	case *ast.AssignmentStatement:
		return c.compileAssignmentStatement(node)
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(node)
	case *ast.BlockStatement:
		return c.compileBlockStatement(node)
	case *ast.ReturnStatement:
		return c.compileReturnStatement(node)
	case *ast.WhileStatement:
		return c.compileWhileStatement(node)
	case *ast.ForStatement:
		return c.compileForStatement(node)
	case *ast.ForEachStatement:
		return c.compileForEachStatement(node)
	case *ast.BreakStatement:
		return c.compileBreakStatement(node)
	case *ast.ContinueStatement:
		return c.compileContinueStatement(node)
	case *ast.StructStatement:
		return c.compileStructStatement(node)
	case *ast.ThrowStatement:
		return c.compileThrowStatement(node)
	case *ast.TryStatement:
		return c.compileTryStatement(node)
	case *ast.RetryStatement:
		return c.compileRetryStatement(node)
	case *ast.ServiceStatement:
		return c.compileServiceStatement(node)
	case *ast.SwitchStatement:
		return c.compileSwitchStatement(node)
	case *ast.EnumStatement:
		return c.compileEnumStatement(node)
	case *ast.ConstStatement:
		return c.compileConstStatement(node)
	case *ast.EchoStatement:
		return c.compileEchoStatement(node)
	case *ast.ImportStatement:
		return c.compileImportStatement(node)
	case *ast.ExportStatement:
		return c.compileExportStatement(node)

	case *ast.CallExpression:
		return c.compileCallExpression(node)
	case *ast.TemplateLiteral:
		return c.compileTemplateLiteral(node)
	case *ast.NullishCoalescingExpression:
		return c.compileNullishCoalescingExpression(node)
	case *ast.OptionalChainingExpression:
		return c.compileOptionalChainingExpression(node)
	case *ast.TernaryExpression:
		return c.compileTernaryExpression(node)
	case *ast.Identifier:
		return c.compileIdentifier(node)
	case *ast.IntegerLiteral:
		return c.compileIntegerLiteral(node)
	case *ast.FloatLiteral:
		return c.compileFloatLiteral(node)
	case *ast.StringLiteral:
		return c.compileStringLiteral(node)
	case *ast.Boolean:
		return c.compileBoolean(node)
	case *ast.Null:
		return c.compileNull(node)
	case *ast.ArrayLiteral:
		return c.compileArrayLiteral(node)
	case *ast.HashLiteral:
		return c.compileHashLiteral(node)
	case *ast.StructLiteral:
		return c.compileStructLiteral(node)
	case *ast.IndexExpression:
		return c.compileIndexExpression(node)
	case *ast.ArrayIndexExpression:
		return c.compileArrayIndexExpression(node)
	case *ast.PrefixExpression:
		return c.compilePrefixExpression(node)
	case *ast.InfixExpression:
		return c.compileInfixExpression(node)
	case *ast.InstantiatedExpression:
		return c.compileInstantiatedExpression(node)
	case *ast.IfExpression:
		return c.compileIfExpression(node)
	case *ast.FunctionLiteral:
		return c.compileFunctionLiteral(node)
	case *ast.ArrowFunction:
		return c.compileArrowFunction(node)
	case *ast.AsyncFunctionLiteral:
		return c.compileAsyncFunctionLiteral(node)
	case *ast.AsyncFunctionStatement:
		return c.compileAsyncFunctionStatement(node)
	case *ast.AwaitExpression:
		return c.compileAwaitExpression(node)
	case *ast.SpawnExpression:
		return c.compileSpawnExpression(node)

	default:
		return fmt.Errorf("unknown node type: %T", node)
	}
}
