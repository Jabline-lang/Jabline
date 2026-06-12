package vm

import (
	"encoding/json"
	"fmt"
	"io"
	"jabline/pkg/log"
	"jabline/pkg/object"
	"net/http"
	"strings"
)

func (vm *VM) StartService(service *object.Service) object.Object {
	portVal, ok := service.Config["port"]
	if !ok {
		return &object.Error{Message: "Service missing 'port' configuration"}
	}
	port := fmt.Sprintf("%d", portVal.(*object.Integer).Value)

	log.Info("Service listening", "name", service.Name, "port", port)

	httpLimiter := make(chan struct{}, 10000)
	handler := func(w http.ResponseWriter, r *http.Request) {
		httpLimiter <- struct{}{}
		defer func() { <-httpLimiter }()

		path := r.URL.Path[1:]
		if path == "" {
			return
		}

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

		reqVM := &VM{
			constants:   vm.constants,
			stack:       make([]object.Object, InitialStackSize),
			sp:          0,
			globals:     GlobalStoreFromSlice(vm.globals.Snapshot()),
			frames:      make([]*Frame, InitialFrames),
			framesIndex: 0,
			handlers:    []ExceptionHandler{},
			filename:    "service",
			methods:     make(map[string]map[string]*object.Closure),
			Types:       vm.Types,
			Ctx:         vm.Ctx,
			Cancel:      vm.Cancel,
		}

		// Push args onto stack: receiver (this), then request
		reqVM.push(service)
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		reqVM.push(requestToObject(r))

		// Set up a frame pointing to the closure's instructions, with basePointer at the start of args
		frame := NewFrame(closure, 0)
		reqVM.frames[0] = frame
		reqVM.framesIndex = 1

		// Reserve space for all locals
		reqVM.sp = closure.Fn.NumLocals

		err := reqVM.Run()
		if err != nil {
			if reqVM.framesIndex > 0 && reqVM.frames[0] == frame {
				reqVM.frames[0] = nil
				reqVM.framesIndex = 0
				ReleaseFrame(frame)
			}
			log.Error("Service handler error", "error", err)
			http.Error(w, err.Error(), 500)
			return
		}

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

// requestToObject converts an *http.Request into a Jabline Hash object.
func requestToObject(r *http.Request) *object.Hash {
	pairs := make(map[object.HashKey]object.HashPair)

	methodKey := &object.String{Value: "method"}
	pairs[methodKey.HashKey()] = object.HashPair{
		Key:   methodKey,
		Value: &object.String{Value: r.Method},
	}

	pathKey := &object.String{Value: "path"}
	pairs[pathKey.HashKey()] = object.HashPair{
		Key:   pathKey,
		Value: &object.String{Value: r.URL.Path},
	}

	queryPairs := make(map[object.HashKey]object.HashPair)
	for k, vals := range r.URL.Query() {
		key := &object.String{Value: k}
		queryPairs[key.HashKey()] = object.HashPair{
			Key:   key,
			Value: &object.String{Value: strings.Join(vals, ", ")},
		}
	}
	queryKey := &object.String{Value: "query"}
	pairs[queryKey.HashKey()] = object.HashPair{
		Key:   queryKey,
		Value: &object.Hash{Pairs: queryPairs},
	}

	headerPairs := make(map[object.HashKey]object.HashPair)
	for k, vals := range r.Header {
		key := &object.String{Value: k}
		headerPairs[key.HashKey()] = object.HashPair{
			Key:   key,
			Value: &object.String{Value: strings.Join(vals, ", ")},
		}
	}
	headersKey := &object.String{Value: "headers"}
	pairs[headersKey.HashKey()] = object.HashPair{
		Key:   headersKey,
		Value: &object.Hash{Pairs: headerPairs},
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err == nil && len(bodyBytes) > 0 {
		bodyKey := &object.String{Value: "body"}
		pairs[bodyKey.HashKey()] = object.HashPair{
			Key:   bodyKey,
			Value: &object.String{Value: string(bodyBytes)},
		}
	}

	remoteKey := &object.String{Value: "remote_addr"}
	pairs[remoteKey.HashKey()] = object.HashPair{
		Key:   remoteKey,
		Value: &object.String{Value: r.RemoteAddr},
	}

	return &object.Hash{Pairs: pairs}
}

func objectToNative(obj object.Object) interface{} {
	if obj == nil {
		return nil
	}
	switch obj := obj.(type) {
	case *object.Integer:
		return obj.Value
	case *object.String:
		return obj.Value
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return nil
	case *object.Array:
		arr := make([]interface{}, len(obj.Elements))
		for i, elem := range obj.Elements {
			arr[i] = objectToNative(elem)
		}
		return arr
	case *object.Hash:
		m := make(map[string]interface{})
		for _, pair := range obj.Pairs {
			key, ok := pair.Key.(*object.String)
			if ok {
				m[key.Value] = objectToNative(pair.Value)
			}
		}
		return m
	case *object.Error:
		return obj.Inspect()
	}
	return obj.Inspect()
}
