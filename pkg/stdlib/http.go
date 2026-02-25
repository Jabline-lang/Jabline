package stdlib

import (
	"fmt"
	"io"
	"jabline/pkg/object"
	"net/http"
	"strings"
)

// Executor is injected by the VM to allow running closures from stdlib
var Executor object.VMExecutor

var HTTPBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"http_get", &object.Builtin{Fn: httpGet}},
	{"http_post", &object.Builtin{Fn: httpPost}},
	{"http_serve", &object.Builtin{Fn: httpServe}},
}

func httpGet(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	url, ok := args[0].(*object.String)
	if !ok {
		return newError("arg must be string")
	}

	resp, err := http.Get(url.Value)
	if err != nil {
		return newError("http error: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return newError("read error: %s", err)
	}

	return &object.String{Value: string(body)}
}

func httpPost(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("wrong number of arguments. got=%d, want=at least 2 (url, body)", len(args))
	}
	url, ok1 := args[0].(*object.String)
	bodyInput, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to `http_post` must be STRING")
	}

	contentType := "text/plain"
	if len(args) == 3 {
		ct, ok := args[2].(*object.String)
		if ok {
			contentType = ct.Value
		}
	}

	resp, err := http.Post(url.Value, contentType, strings.NewReader(bodyInput.Value))
	if err != nil {
		return newError("http post error: %s", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return newError("read error: %s", err)
	}

	return &object.String{Value: string(respBody)}
}

func httpServe(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args. usage: http_serve(port, handler)")
	}

	portObj, ok := args[0].(*object.Integer)
	if !ok {
		return newError("port must be integer")
	}
	port := fmt.Sprintf(":%d", portObj.Value)

	handlerClosure, ok := args[1].(*object.Closure)
	if !ok {
		return newError("handler must be a function/closure")
	}

	if Executor == nil {
		return newError("VM Executor not initialized")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1. Build Request Object (Hash)
		reqHash := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}

		// Method
		methodKey := &object.String{Value: "method"}
		reqHash.Pairs[methodKey.HashKey()] = object.HashPair{Key: methodKey, Value: &object.String{Value: r.Method}}

		// URL
		urlKey := &object.String{Value: "url"}
		reqHash.Pairs[urlKey.HashKey()] = object.HashPair{Key: urlKey, Value: &object.String{Value: r.URL.String()}}

		// Body
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyKey := &object.String{Value: "body"}
		reqHash.Pairs[bodyKey.HashKey()] = object.HashPair{Key: bodyKey, Value: &object.String{Value: string(bodyBytes)}}

		// 2. Execute Jabline Handler
		// We expect the handler to return a Hash: { status: 200, body: "...", headers: {...} }
		result := Executor(handlerClosure, []object.Object{reqHash})

		// 3. Process Response
		if result.Type() == object.ERROR_OBJ {
			http.Error(w, result.Inspect(), http.StatusInternalServerError)
			return
		}

		respHash, ok := result.(*object.Hash)
		if !ok {
			// If handler returns string, treat as body 200 OK
			if str, ok := result.(*object.String); ok {
				w.WriteHeader(200)
				w.Write([]byte(str.Value))
				return
			}
			http.Error(w, "Handler must return a Hash or String", http.StatusInternalServerError)
			return
		}

		// Status
		status := http.StatusOK
		statusKey := &object.String{Value: "status"}
		if pair, ok := respHash.Pairs[statusKey.HashKey()]; ok {
			if s, ok := pair.Value.(*object.Integer); ok {
				status = int(s.Value)
			}
		}

		// Body
		body := ""
		bodyRespKey := &object.String{Value: "body"}
		if pair, ok := respHash.Pairs[bodyRespKey.HashKey()]; ok {
			if s, ok := pair.Value.(*object.String); ok {
				body = s.Value
			}
		}

		// Headers (Optional implementation later)

		w.WriteHeader(status)
		w.Write([]byte(body))
	})

	fmt.Printf("Jabline HTTP Server listening on %s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		return newError("server error: %s", err)
	}

	return &object.Null{}
}
