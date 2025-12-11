package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VM struct {
	constants    []object.Object
	stack        []object.Object
	sp           int
	globals      []object.Object
	
	frames      []*Frame
	framesIndex int
	
	handlers    []ExceptionHandler
	filename    string
	loader      *ModuleLoader
}

type ExceptionHandler struct {
	CatchIP    int
	StackSP    int
	FrameIndex int
}

func New(instructions code.Instructions, constants []object.Object, filename string) *VM {
	return NewWithLoader(instructions, constants, filename, NewModuleLoader())
}

func NewWithLoader(instructions code.Instructions, constants []object.Object, filename string, loader *ModuleLoader) *VM {
	mainFn := &object.CompiledFunction{Instructions: instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)
	
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:    constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		globals:      make([]object.Object, GlobalsSize),
		frames:       frames,
		framesIndex:  1,
		handlers:     []ExceptionHandler{},
		filename:     filename,
		loader:       loader,
	}
}

func NewWithGlobalsStore(instructions code.Instructions, constants []object.Object, globals []object.Object, filename string) *VM {
	vm := New(instructions, constants, filename)
	vm.globals = globals
	return vm
}

func (vm *VM) newRuntimeError(format string, a ...interface{}) *RuntimeError {
	currentIP := vm.currentFrame().ip
	var pos code.SourcePos
	
	if vm.currentFrame().cl != nil && vm.currentFrame().cl.Fn != nil {
		if vm.currentFrame().cl.Fn.SourceMap != nil {
			pos = vm.currentFrame().cl.Fn.SourceMap[currentIP]
		}
	}

	return &RuntimeError{
		Message: fmt.Sprintf(format, a...),
		Line:    pos.Line,
		Column:  pos.Column,
		File:    vm.filename,
	}
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			if err := vm.opConstant(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpPop: vm.opPop()
		case code.OpDup:
			if err := vm.opDup(); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			if err := vm.opBinary(op); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpTrue: if err := vm.opTrue(); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpFalse: if err := vm.opFalse(); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpNull: if err := vm.opNull(); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			if err := vm.opComparison(op); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpBang, code.OpMinus:
			if err := vm.opPrefix(op); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpJump: vm.opJump(ins, &ip)
		case code.OpJumpNotTruthy: vm.opJumpNotTruthy(ins, &ip)
		case code.OpJumpNotTruthyKeep: vm.opJumpNotTruthyKeep(ins, &ip)
		case code.OpJumpTruthyKeep: vm.opJumpTruthyKeep(ins, &ip)
		case code.OpJumpNotNull: vm.opJumpNotNull(ins, &ip)
		case code.OpJumpIfEqual: vm.opJumpIfEqual(ins, &ip)
		case code.OpJumpIfTrue: vm.opJumpIfTrue(ins, &ip)
		case code.OpSetGlobal: vm.opSetGlobal(ins, &ip)
		case code.OpGetGlobal: if err := vm.opGetGlobal(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpSetLocal: vm.opSetLocal(ins, &ip)
		case code.OpGetLocal: if err := vm.opGetLocal(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpGetBuiltin: if err := vm.opGetBuiltin(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpGetFree: if err := vm.opGetFree(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpArray: if err := vm.opArray(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpHash: if err := vm.opHash(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpIndex: if err := vm.opIndex(); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpCall: if err := vm.opCall(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpReturnValue: if err := vm.opReturnValue(); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpReturn: if err := vm.opReturn(); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpClosure: if err := vm.opClosure(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpInstance: if err := vm.opInstance(ins, &ip); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpImport: if err := vm.opImport(); err != nil { return vm.newRuntimeError(err.Error()) }
		case code.OpTry: vm.opTry(ins, &ip)
		case code.OpEndTry: vm.opEndTry()
		case code.OpThrow: if err := vm.opThrow(); err != nil { return vm.newRuntimeError(err.Error()) }
		}
		
		vm.currentFrame().ip = ip
	}
	return nil
}
