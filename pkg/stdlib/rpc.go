package stdlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"jabline/pkg/object"
	"net/http"
	"sync"
)

// JSON-RPC module — lightweight RPC over HTTP/JSON (gRPC alternative)

var (
	rpcHandlers   = make(map[string]*object.Closure)
	rpcHandlersMu sync.RWMutex
)

var RPCBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"rpc_serve", &object.Builtin{Fn: rpcServe}},
	{"rpc_call", &object.Builtin{Fn: rpcCall}},
	{"rpc_register", &object.Builtin{Fn: rpcRegister}},
}

func init() {
	NativeModuleRegistry["_rpc"] = RPCBuiltins
	NativeModulePrefixes["_rpc"] = "rpc_"

	Registry = append(Registry, []struct {
		Name   string
		Object object.Object
	}{
		{"rpc_serve", &object.Builtin{Fn: rpcServe}},
		{"rpc_call", &object.Builtin{Fn: rpcCall}},
		{"rpc_register", &object.Builtin{Fn: rpcRegister}},
	}...)
}

func rpcRegister(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("rpc_register expects 2 args: (name, fn)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("name must be STRING, got %s", args[0].Type())
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("fn must be a FUNCTION/CLOSURE, got %s", args[1].Type())
	}

	rpcHandlersMu.Lock()
	rpcHandlers[name.Value] = fn
	rpcHandlersMu.Unlock()

	return &object.Boolean{Value: true}
}

// JSON-RPC request/response
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func rpcServe(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("rpc_serve expects at least 2 args: (port, handler_fn)")
	}
	portObj, ok := args[0].(*object.Integer)
	if !ok {
		return newError("port must be INTEGER, got %s", args[0].Type())
	}

	// Optional: register individual handlers from a hash
	if len(args) >= 3 {
		if handlers, ok := args[2].(*object.Hash); ok {
			for _, pair := range handlers.Pairs {
				if name, ok := pair.Key.(*object.String); ok {
					if fn, ok := pair.Value.(*object.Closure); ok {
						rpcHandlersMu.Lock()
						rpcHandlers[name.Value] = fn
						rpcHandlersMu.Unlock()
					}
				}
			}
		}
	}

	port := fmt.Sprintf(":%d", portObj.Value)
	mux := http.NewServeMux()

	mux.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			sendRPCError(w, nil, -32700, "Parse error", err.Error())
			return
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			sendRPCError(w, nil, -32700, "Parse error", err.Error())
			return
		}

		if req.JSONRPC != "2.0" {
			sendRPCError(w, req.ID, -32600, "Invalid Request", "jsonrpc must be 2.0")
			return
		}

		rpcHandlersMu.RLock()
		fn, exists := rpcHandlers[req.Method]
		rpcHandlersMu.RUnlock()

		if !exists {
			sendRPCError(w, req.ID, -32601, "Method not found", fmt.Sprintf("'%s' is not registered", req.Method))
			return
		}

		// Parse params into Jabline objects
		var args []object.Object
		if req.Params != nil {
			var rawParams []json.RawMessage
			if err := json.Unmarshal(req.Params, &rawParams); err != nil {
				// Try as single object
				var single map[string]interface{}
				if err2 := json.Unmarshal(req.Params, &single); err2 != nil {
					sendRPCError(w, req.ID, -32602, "Invalid params", err.Error())
					return
				}
				args = append(args, rpcGoToJabline(single))
			} else {
				for _, p := range rawParams {
					var val interface{}
					json.Unmarshal(p, &val)
					args = append(args, rpcGoToJabline(val))
				}
			}
		}

		if Executor == nil {
			sendRPCError(w, req.ID, -32603, "Internal error", "VM Executor not initialized")
			return
		}

		result := Executor(fn, args)
		if errObj, ok := result.(*object.Error); ok {
			sendRPCError(w, req.ID, -32603, "Internal error", errObj.Message)
			return
		}

		sendRPCSuccess(w, req.ID, rpcJablineToGo(result))
	})

	fmt.Printf("JSON-RPC server listening on %s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		return newError("rpc_serve failed: %s", err)
	}

	return &object.Null{}
}

func sendRPCError(w http.ResponseWriter, id interface{}, code int, message string, data string) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendRPCSuccess(w http.ResponseWriter, id interface{}, result interface{}) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func rpcCall(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("rpc_call expects at least 3 args: (url, method, params...)")
	}
	url, ok := args[0].(*object.String)
	if !ok {
		return newError("url must be STRING, got %s", args[0].Type())
	}
	method, ok := args[1].(*object.String)
	if !ok {
		return newError("method must be STRING, got %s", args[1].Type())
	}

	params := make([]interface{}, 0)
	for i := 2; i < len(args); i++ {
		params = append(params, rpcJablineToGo(args[i]))
	}

	reqBody := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method.Value,
		ID:      1,
	}
	if len(params) > 0 {
		raw, _ := json.Marshal(params)
		reqBody.Params = raw
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(url.Value, "application/json", bytes.NewReader(body))
	if err != nil {
		return newError("rpc_call failed: %s", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return newError("rpc_call: invalid response: %s", err)
	}

	if rpcResp.Error != nil {
		return newError("rpc_call error (%d): %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcGoToJabline(rpcResp.Result)
}

// rpcJablineToGo converts a Jabline object to a Go value for JSON serialization.
func rpcJablineToGo(obj object.Object) interface{} {
	switch v := obj.(type) {
	case *object.String:
		return v.Value
	case *object.Integer:
		return v.Value
	case *object.Float:
		return v.Value
	case *object.Boolean:
		return v.Value
	case *object.Null:
		return nil
	case *object.Array:
		arr := make([]interface{}, len(v.Elements))
		for i, e := range v.Elements {
			arr[i] = rpcJablineToGo(e)
		}
		return arr
	case *object.Hash:
		m := make(map[string]interface{})
		for _, pair := range v.Pairs {
			if key, ok := pair.Key.(*object.String); ok {
				m[key.Value] = rpcJablineToGo(pair.Value)
			}
		}
		return m
	default:
		return v.Inspect()
	}
}

// rpcGoToJabline converts a Go value to a Jabline object.
func rpcGoToJabline(val interface{}) object.Object {
	if val == nil {
		return &object.Null{}
	}
	switch v := val.(type) {
	case string:
		return &object.String{Value: v}
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
			return &object.Integer{Value: int64(v)}
		}
		return &object.Float{Value: v}
	case bool:
		return &object.Boolean{Value: v}
	case map[string]interface{}:
		pairs := make(map[object.HashKey]object.HashPair)
		for k, val := range v {
			ks := &object.String{Value: k}
			pairs[ks.HashKey()] = object.HashPair{Key: ks, Value: rpcGoToJabline(val)}
		}
		return &object.Hash{Pairs: pairs}
	case []interface{}:
		elements := make([]object.Object, len(v))
		for i, e := range v {
			elements[i] = rpcGoToJabline(e)
		}
		return &object.Array{Elements: elements}
	default:
		return &object.String{Value: fmt.Sprint(v)}
	}
}
