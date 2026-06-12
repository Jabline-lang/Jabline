package vm

import (
	"context"
	"jabline/pkg/object"
)

// Fork crea un VM ligero para ejecutar un closure en una goroutine separada.
// Comparte constants, globals, methods y Types del VM padre como solo-lectura.
// Cada fork tiene su propio stack y frames independientes.
func (parent *VM) Fork() *VM {
	ctx, cancel := context.WithCancel(context.Background())

	// Deep copy methods to avoid data races between parent and child VMs
	methodsCopy := make(map[string]map[string]*object.Closure, len(parent.methods))
	for k, v := range parent.methods {
		innerCopy := make(map[string]*object.Closure, len(v))
		for mk, mv := range v {
			innerCopy[mk] = mv
		}
		methodsCopy[k] = innerCopy
	}
	typesCopy := make(map[string]object.Object, len(parent.Types))
	for k, v := range parent.Types {
		typesCopy[k] = v
	}

	child := &VM{
		constants: parent.constants,
		globals:   GlobalStoreFromSlice(parent.globals.Snapshot()),
		methods:   methodsCopy,
		Types:     typesCopy,
		loader:    parent.loader,
		filename:  parent.filename,
		Telemetry: parent.Telemetry,

		// Privado del fork — empieza pequeño y crece bajo demanda
		stack:           make([]object.Object, InitialStackSize),
		sp:              0,
		frames:          make([]*Frame, InitialFrames),
		framesIndex:     0,
		handlers:        []ExceptionHandler{},
		finallyHandlers: []FinallyHandler{},
		Ctx:             ctx,
		Cancel:          cancel,

		// Sandbox: hereda la política del padre
		Sandbox: parent.Sandbox,
	}
	return child
}

// RunClosure ejecuta un closure con los argumentos dados en este VM fork.
// Crea un frame inicial para el closure y ejecuta hasta completar.
func (vm *VM) RunClosure(cl *object.Closure, args []object.Object) object.Object {
	// 1. Colocar un slot para el function_obj para OpReturnValue (basePointer - 1)
	vm.stack[0] = Null
	vm.sp = 1

	// 2. Coloca los argumentos en el stack
	for _, arg := range args {
		vm.stack[vm.sp] = arg
		vm.sp++
	}

	// 3. Crea el frame para el closure (basePointer apunta al primer argumento)
	frame := NewFrame(cl, vm.sp-len(args))
	if cl.Globals != nil {
		frame.savedGlobals = vm.globals
		vm.globals = GlobalStoreFromSlice(cl.Globals)
	}
	if cl.Constants != nil {
		frame.savedConstants = vm.constants
		vm.constants = cl.Constants
	}
	vm.frames[0] = frame
	vm.framesIndex = 1
	vm.sp = frame.basePointer + cl.Fn.NumLocals

	if err := vm.Run(); err != nil {
		if vm.framesIndex > 0 {
			vm.frames[0] = nil
			vm.framesIndex = 0
			ReleaseFrame(frame)
		}
		return &object.Error{Message: err.Error()}
	}

	// El valor de retorno queda en vm.stack[vm.sp-1]
	if vm.sp > 0 {
		result := vm.stack[vm.sp-1]
		if result == nil {
			return Null
		}
		return result
	}
	return Null
}
