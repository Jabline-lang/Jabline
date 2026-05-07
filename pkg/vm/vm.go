package vm

import (
	"context"
	"fmt"
	"io/fs"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"jabline/pkg/stdlib"
)

// EmbeddedModules can be set from main to bundle the standard library into the binary.
// It is checked before the OS filesystem when loading modules.
var EmbeddedModules fs.FS

var GlobalLoader *ModuleLoader

// GlobalVM apunta al VM principal en ejecución.
// Los forks (HTTP, async, spawn) heredan de él.
var GlobalVM *VM

func init() {
	GlobalLoader = NewModuleLoaderWithEmbed(EmbeddedModules)
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
	Types   map[string]object.Object

	LastError object.Object // Store last encountered error for recover()

	Ctx    context.Context
	Cancel context.CancelFunc

	Telemetry *Telemetry
	Debug     *DebugSession
}

type ExceptionHandler struct {
	CatchIP    int
	StackSP    int
	FrameIndex int
}

func New(instructions code.Instructions, constants []object.Object, filename string) *VM {
	return NewWithLoader(instructions, constants, filename, NewModuleLoaderWithEmbed(EmbeddedModules))
}

func NewWithLoader(instructions code.Instructions, constants []object.Object, filename string, loader *ModuleLoader) *VM {
	mainFn := &object.CompiledFunction{Instructions: instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)
	ctx, cancel := context.WithCancel(context.Background())

	vm := &VM{
		constants:   constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      make([]*Frame, MaxFrames),
		framesIndex: 1,
		handlers:    []ExceptionHandler{},
		filename:    filename,
		loader:      loader,
		methods:     make(map[string]map[string]*object.Closure),
		Types:       make(map[string]object.Object),
		Ctx:         ctx,
		Cancel:      cancel,
	}
	vm.frames[0] = mainFrame

	// Populate the Types registry with Structs and Interfaces from constants
	for _, c := range constants {
		if s, ok := c.(*object.Struct); ok {
			vm.Types[s.Name] = s
		} else if i, ok := c.(*object.Interface); ok {
			vm.Types[i.Name] = i
		}
	}
	return vm
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
				// Fuzzy lookup: Find the nearest position <= current IP
				bestIP := -1
				for ip := range frm.cl.Fn.SourceMap {
					if ip <= frm.ip && ip > bestIP {
						bestIP = ip
					}
				}
				if bestIP != -1 {
					pos = frm.cl.Fn.SourceMap[bestIP]
				}
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
	errObj := &object.Error{Message: msg}
	vm.LastError = errObj
	vm.stack[vm.sp] = errObj
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
		// Debug hook
		if vm.Debug != nil {
			vm.Debug.OnInstruction(vm)
		}

		// Context cancellation check
		if err := vm.Ctx.Err(); err != nil {
			return vm.newRuntimeError("process explicitly cancelled by context")
		}

		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			if err := vm.opConstant(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpConstant8:
			if err := vm.opConstant8(ins, &ip); err != nil {
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
		case code.OpFloatAdd, code.OpFloatSub, code.OpFloatMul, code.OpFloatDiv:
			if err := vm.opFloatBinary(op); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpIntToFloat:
			if err := vm.opIntToFloat(); err != nil {
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
		case code.OpIncGlobal:
			if err := vm.opIncGlobal(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpDecGlobal:
			if err := vm.opDecGlobal(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpSetLocal:
			vm.opSetLocal(ins, &ip)
		case code.OpGetLocal:
			if err := vm.opGetLocal(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpIncLocal:
			if err := vm.opIncLocal(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpDecLocal:
			if err := vm.opDecLocal(ins, &ip); err != nil {
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

		case code.OpCallMethodFast:
			// methodNameIdx (2 bytes), numArgs (1 byte)
			methodNameIdx := int(code.ReadUint16(ins[ip+1:]))
			numArgs := int(ins[ip+3])
			ip += 3                   // Advance IP past operands
			vm.currentFrame().ip = ip // Save the updated IP

			methodNameObj := vm.constants[methodNameIdx]
			receiver := vm.stack[vm.sp-numArgs-1]

			callable, err := vm.getProperty(receiver, methodNameObj)
			if err != nil {
				return vm.handleNativeError(err.Error())
			}

			// Extract closure from BoundMethod or use as is
			var fn *object.Closure
			var builtin *object.Builtin
			switch c := callable.(type) {
			case *object.BoundMethod:
				fn = c.Function
			case *object.Closure:
				fn = c
			case *object.Builtin:
				builtin = c
			default:
				var tType string
				if callable != nil {
					tType = string(callable.Type())
				} else {
					tType = "nil"
				}
				return vm.handleNativeError(fmt.Sprintf("method %s is not a function (got %s)", methodNameObj.Inspect(), tType))
			}

			// Stack Shuffle:
			// We have [Receiver, Arg1, ..., ArgN]
			// We need [Callable, Receiver, Arg1, ..., ArgN]
			// Shift everything up by 1 position

			for len(vm.stack) <= vm.sp {
				vm.stack = append(vm.stack, nil)
			}

			for i := vm.sp; i > vm.sp-numArgs-1; i-- {
				vm.stack[i] = vm.stack[i-1]
			}
			
			if builtin != nil {
				vm.stack[vm.sp-numArgs-1] = builtin
			} else {
				vm.stack[vm.sp-numArgs-1] = fn
			}
			vm.sp++

			if builtin != nil {
				// Execute the builtin directly (numArgs + 1 incorporates self)
				if err := vm.executeCallBuiltin(builtin, numArgs+1); err != nil {
					return vm.handleNativeError(err.Error())
				}
				continue
			}

			// Now call executeCallClosure
			if err := vm.executeCallClosure(fn, numArgs+1, nil); err != nil {
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
		case code.OpIsType:
			if err := vm.opIsType(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpCheckType:
			if err := vm.opCheckType(ins, &ip); err != nil {
				if err := vm.handleNativeError(err.Error()); err != nil {
					return err
				}
				continue
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
		case code.OpMetricInc:
			if err := vm.opMetricInc(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpTraceStart:
			if err := vm.opTraceStart(ins, &ip); err != nil {
				return vm.handleNativeError(err.Error())
			}
		case code.OpTraceEnd:
			if err := vm.opTraceEnd(); err != nil {
				return vm.handleNativeError(err.Error())
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
