package vm

import (
	"encoding/json"
	"fmt"
	"jabline/pkg/object"
	"net/http"
)

func (vm *VM) StartService(service *object.Service) object.Object {
	portVal, ok := service.Config["port"]
	if !ok {
		return &object.Error{Message: "Service missing 'port' configuration"}
	}
	port := fmt.Sprintf("%d", portVal.(*object.Integer).Value)

	fmt.Printf("🚀 Service '%s' listening on port %s...\n", service.Name, port)

	// Register Handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:] // remove leading /
		if path == "" { return }

		methods, ok := vm.methods[service.Name]
		if !ok {
			http.Error(w, "Service not found", 404)
			return
		}
		closure, ok := methods[path]
		if !ok {
			http.Error(w, "Method not found", 404)
			return
		}

		// Create a fresh VM for this request
		// Sharing constants and globals (READ-ONLY ideally)
		// Note: passing closure.Fn.Instructions assumes it's self-contained or refers to globals/constants correctly.
		reqVM := NewWithGlobalsStore(closure.Fn.Instructions, vm.constants, vm.globals, "service")
		
		// If closure has captured vars (free variables), we need to inject them?
		// OpGetFree depends on closure.Free. 
		// But here we are running instructions directly, bypassing OpCall logic.
		// If the function uses free variables, this simple runner will FAIL because OpGetFree expects a closure frame.
		// Fix: Wrap execution in a frame.
		
		// Setup Frame manually
		frame := NewFrame(closure, 0)
		reqVM.frames[0] = frame
		reqVM.sp = closure.Fn.NumLocals // Reserve space for locals (args are locals 0..N)
		
		// TODO: Parse args from Request and push to stack (into locals slots)
		// For now, 0 args.
		
		err := reqVM.Run()
		if err != nil {
			fmt.Println("Runtime Error:", err)
			http.Error(w, err.Error(), 500)
			return
		}
		
		// Result is on top of stack (pushed by OpReturnValue)
		// But OpReturnValue pops the frame.
		// If we run the function body directly, OpReturnValue is the last instruction.
		// It pushes to stack[sp-1] (overwriting func?).
		// In bare Run(), there is no caller frame. OpReturnValue checks if framesIndex==0.
		
		result := reqVM.StackTop() 
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(objectToNative(result))
	}

	err := http.ListenAndServe(":"+port, http.HandlerFunc(handler))
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	
	return Null
}

func objectToNative(obj object.Object) interface{} {
	if obj == nil { return nil }
	switch obj := obj.(type) {
	case *object.Integer:
		return obj.Value
	case *object.String:
		return obj.Value
	case *object.Boolean:
		return obj.Value
	case *object.Hash:
		m := make(map[string]interface{})
		for _, pair := range obj.Pairs {
			key, ok := pair.Key.(*object.String)
			if ok {
				m[key.Value] = objectToNative(pair.Value)
			}
		}
		return m
	}
	return obj.Inspect()
}
