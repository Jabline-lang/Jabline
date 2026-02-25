package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"jabline/pkg/stdlib"
)

func init() {
	stdlib.Executor = ExecuteClosureBridge
}

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VM struct {
	constants []object.Object
	stack     []object.Object
	sp        int
	globals   []object.Object

	frames      []*Frame
	framesIndex int

	handlers []ExceptionHandler
	filename string
	loader   *ModuleLoader

	methods map[string]map[string]*object.Closure
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
		constants:   constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
		handlers:    []ExceptionHandler{},
		filename:    filename,
		loader:      loader,
		methods:     make(map[string]map[string]*object.Closure),
	}
}

func NewWithGlobalsStore(instructions code.Instructions, constants []object.Object, globals []object.Object, filename string) *VM {
	vm := New(instructions, constants, filename)
	vm.globals = globals
	return vm
}

func (vm *VM) newRuntimeError(format string, a ...interface{}) *RuntimeError {
	msg := fmt.Sprintf(format, a...)

	var trace []CallFrame
	for i := 0; i < vm.framesIndex; i++ {
		frm := vm.frames[i]
		var pos code.SourcePos
		fnName := "<main>"

		if frm.cl != nil && frm.cl.Fn != nil {
			if frm.cl.Fn.SourceMap != nil {
				pos = frm.cl.Fn.SourceMap[frm.ip]
			}
			if frm.cl.Fn.Name != "" {
				fnName = frm.cl.Fn.Name
			} else if i > 0 {
				fnName = "<anonymous>"
			}
		}

		trace = append(trace, CallFrame{
			Function: fnName,
			File:     vm.filename,
			Line:     pos.Line,
			Column:   pos.Column,
		})
	}

	return &RuntimeError{
		Message:    msg,
		StackTrace: trace,
	}
}

func (vm *VM) handleNativeError(msg string) error {
	if len(vm.handlers) == 0 {
		return vm.newRuntimeError("%s", msg)
	}

	handler := vm.handlers[len(vm.handlers)-1]
	vm.handlers = vm.handlers[:len(vm.handlers)-1]

	// Unwind stack
	vm.sp = handler.StackSP
	vm.framesIndex = handler.FrameIndex

	// Convert msg to Error object and push to stack for catch
	vm.stack[vm.sp] = &object.Error{Message: msg}
	vm.sp++

	// Jump to catch block (offset by -1 because Run loop increments it)
	vm.currentFrame().ip = handler.CatchIP - 1
	return nil
}

func (vm *VM) Run() (err error) {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	ins = vm.currentFrame().Instructions()

	defer func() {
		if r := recover(); r != nil {
			errStr := fmt.Sprintf("panic: %v", r)
			// Handle as native error if possible
			err = vm.handleNativeError(errStr)
			// If we handled it (err == nil), it means we jumped to catch.
			// We need to resume execution.
			if err == nil {
				err = vm.Run()
			}
		}
	}()

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			if err := vm.opConstant(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpPop:
			vm.opPop()
		case code.OpDup:
			if err := vm.opDup(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpMod, code.OpBitAnd, code.OpBitOr, code.OpBitXor, code.OpShiftLeft, code.OpShiftRight:
			if err := vm.opBinary(op); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpTrue:
			if err := vm.opTrue(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpFalse:
			if err := vm.opFalse(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpNull:
			if err := vm.opNull(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			if err := vm.opComparison(op); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpBang, code.OpMinus, code.OpBitNot:
			if err := vm.opPrefix(op); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpJump:
			vm.opJump(ins, &ip)
		case code.OpJumpNotTruthy:
			vm.opJumpNotTruthy(ins, &ip)
		case code.OpJumpNotTruthyKeep:
			vm.opJumpNotTruthyKeep(ins, &ip)
		case code.OpJumpTruthyKeep:
			vm.opJumpTruthyKeep(ins, &ip)
		case code.OpJumpNotNull:
			vm.opJumpNotNull(ins, &ip)
		case code.OpJumpIfEqual:
			vm.opJumpIfEqual(ins, &ip)
		case code.OpJumpIfTrue:
			vm.opJumpIfTrue(ins, &ip)
		case code.OpSetGlobal:
			vm.opSetGlobal(ins, &ip)
		case code.OpGetGlobal:
			if err := vm.opGetGlobal(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpSetLocal:
			vm.opSetLocal(ins, &ip)
		case code.OpGetLocal:
			if err := vm.opGetLocal(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpGetBuiltin:
			if err := vm.opGetBuiltin(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpGetFree:
			if err := vm.opGetFree(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpSetFree:
			vm.opSetFree(ins, &ip)
		case code.OpArray:
			if err := vm.opArray(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpHash:
			if err := vm.opHash(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpIndex:
			if err := vm.opIndex(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpSetProperty:
			if err := vm.opSetProperty(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpCall:
			// Handle OpCall manually to manage IP updates correctly before frame switch
			numArgs := int(ins[ip+1])
			ip += 1                   // Advance IP past operand
			vm.currentFrame().ip = ip // Save the updated IP to the current frame (caller)

			if err := vm.executeCall(numArgs); err != nil {
				return vm.handleNativeError(err.Error())
			}
			continue // Continue loop with the new frame (callee)

		case code.OpReturnValue:
			if err := vm.opReturnValue(); err != nil {
				return vm.handleNativeError(err.Error())
			}
			if vm.framesIndex == 0 {
				return nil
			}
			continue // Frame popped, refresh.
		case code.OpReturn:
			if err := vm.opReturn(); err != nil {
				return vm.handleNativeError(err.Error())
			}
			if vm.framesIndex == 0 {
				return nil
			}
			continue // Frame popped, refresh.

		case code.OpAwait:
			if err := vm.opAwait(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpGetProperty:
			if err := vm.opIndex(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpCheckType:
			if err := vm.opCheckType(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpSendChannel:
			if err := vm.opSendChannel(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpRecvChannel:
			if err := vm.opRecvChannel(); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpCurrentClosure:
			vm.opCurrentClosure()
		case code.OpInstantiate:
			if err := vm.opInstantiate(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpClosure:
			if err := vm.opClosure(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpInstance:
			if err := vm.opInstance(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpImport:
			if err := vm.opImport(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpSpawn:
			if err := vm.opSpawn(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpTry:
			vm.opTry(ins, &ip)
		case code.OpEndTry:
			vm.opEndTry()
		case code.OpThrow:
			if err := vm.opThrow(); err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
			continue
		case code.OpRegisterMethod:
			if err := vm.opRegisterMethod(ins, &ip); err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
		case code.OpService:
			if err := vm.opService(ins, &ip); err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
		}

		vm.currentFrame().ip = ip
	}
	return nil
}

// LastPoppedStackElem returns the element that was last popped from the stack.
// Used primarily for testing purposes to assert final variable values.
func (vm *VM) LastPoppedStackElem() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}
