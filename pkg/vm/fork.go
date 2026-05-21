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
	child := &VM{
		// Compartido (solo lectura — el fork no modifica globals del padre
		// durante la ejecución de un request)
		constants: parent.constants,
		globals:   parent.globals,
		methods:   parent.methods,
		Types:     parent.Types,
		loader:    parent.loader,
		filename:  parent.filename,
		Telemetry: parent.Telemetry,

		// Privado del fork — empieza pequeño y crece bajo demanda
		stack:       make([]object.Object, InitialStackSize),
		sp:          0,
		frames:      make([]*Frame, InitialFrames),
		framesIndex: 0,
		handlers:    []ExceptionHandler{},
		Ctx:         ctx,
		Cancel:      cancel,
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
