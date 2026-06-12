//go:build !runner || runner_http

package stdlib

import (
	"context"
	"fmt"
	"io"
	"jabline/pkg/log"
	"jabline/pkg/object"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// httpClient is a shared HTTP client with a 30-second timeout to prevent
// dangling connections from hanging the Jabline process indefinitely.
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

var HTTPBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"http_get", &object.Builtin{Fn: httpGet}},
	{"http_post", &object.Builtin{Fn: httpPost}},
	{"http_serve", &object.Builtin{Fn: httpServe}},
}

func init() {
	NativeModuleRegistry["_http"] = HTTPBuiltins
	
	Registry = append(Registry, []struct {
		Name   string
		Object object.Object
	}{
		{"http_serve", &object.Builtin{Fn: httpServe}},
		{"path_segments", &object.Builtin{Fn: pathSegmentsFunc}},
		{"path_param", &object.Builtin{Fn: pathParamFunc}},
		{"query_params", &object.Builtin{Fn: queryParamsFunc}},
		{"query_param", &object.Builtin{Fn: queryParamFunc}},
		{"render_template", &object.Builtin{Fn: renderTemplateFunc}},
	}...)
}

func httpGet(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	url, ok := args[0].(*object.String)
	if !ok {
		return newError("arg must be string")
	}

	resp, err := httpClient.Get(url.Value)
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

	resp, err := httpClient.Post(url.Value, contentType, strings.NewReader(bodyInput.Value))
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

	mux := http.NewServeMux()
	httpLimiter := make(chan struct{}, 10000)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpLimiter <- struct{}{}
		defer func() { <-httpLimiter }()

		// 1. Build Request Object (Hash)
		reqHash := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}

		// Method
		methodKey := &object.String{Value: "method"}
		reqHash.Pairs[methodKey.HashKey()] = object.HashPair{Key: methodKey, Value: &object.String{Value: r.Method}}

		// URL (full)
		urlKey := &object.String{Value: "url"}
		reqHash.Pairs[urlKey.HashKey()] = object.HashPair{Key: urlKey, Value: &object.String{Value: r.URL.String()}}

		// Path (without query)
		pathKey := &object.String{Value: "path"}
		reqHash.Pairs[pathKey.HashKey()] = object.HashPair{Key: pathKey, Value: &object.String{Value: r.URL.Path}}

		// Query parameters as Hash
		queryPairs := make(map[object.HashKey]object.HashPair)
		for qk, qvs := range r.URL.Query() {
			qkObj := &object.String{Value: qk}
			if len(qvs) == 1 {
				queryPairs[qkObj.HashKey()] = object.HashPair{Key: qkObj, Value: &object.String{Value: qvs[0]}}
			} else {
				var elems []object.Object
				for _, v := range qvs {
					elems = append(elems, &object.String{Value: v})
				}
				queryPairs[qkObj.HashKey()] = object.HashPair{Key: qkObj, Value: &object.Array{Elements: elems}}
			}
		}
		queryKey := &object.String{Value: "query"}
		reqHash.Pairs[queryKey.HashKey()] = object.HashPair{Key: queryKey, Value: &object.Hash{Pairs: queryPairs}}

		// Host
		hostKey := &object.String{Value: "host"}
		reqHash.Pairs[hostKey.HashKey()] = object.HashPair{Key: hostKey, Value: &object.String{Value: r.Host}}

		// Headers as Hash
		headerPairs := make(map[object.HashKey]object.HashPair)
		for hk, hv := range r.Header {
			hKey := &object.String{Value: strings.ToLower(hk)}
			if len(hv) == 1 {
				headerPairs[hKey.HashKey()] = object.HashPair{Key: hKey, Value: &object.String{Value: hv[0]}}
			} else {
				var elems []object.Object
				for _, v := range hv {
					elems = append(elems, &object.String{Value: v})
				}
				headerPairs[hKey.HashKey()] = object.HashPair{Key: hKey, Value: &object.Array{Elements: elems}}
			}
		}
		reqHeadersKey := &object.String{Value: "headers"}
		reqHash.Pairs[reqHeadersKey.HashKey()] = object.HashPair{Key: reqHeadersKey, Value: &object.Hash{Pairs: headerPairs}}

		// Body (limited to 10MB to prevent OOM)
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyKey := &object.String{Value: "body"}
		reqHash.Pairs[bodyKey.HashKey()] = object.HashPair{Key: bodyKey, Value: &object.String{Value: string(bodyBytes)}}

		// 2. Execute Jabline Handler
		// We expect the handler to return a Hash: { status: 200, body: "...", headers: {...} }
		if os.Getenv("JABLINE_DEBUG_HTTP") == "1" {
			log.Debug("HTTP dispatch", "method", r.Method, "url", r.URL.String())
		}
		result := Executor(handlerClosure, []object.Object{reqHash})
		if os.Getenv("JABLINE_DEBUG_HTTP") == "1" {
			log.Debug("HTTP handler returned", "result", result.Inspect())
		}

		// 3. Process Response — resolve any nested async channels
		for {
			chanObj, isChan := result.(*object.Channel)
			if !isChan {
				break
			}
			result = <-chanObj.Value
		}

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

		// Headers
		headersKey := &object.String{Value: "headers"}
		if pair, ok := respHash.Pairs[headersKey.HashKey()]; ok {
			if headersHash, ok := pair.Value.(*object.Hash); ok {
				for _, hpair := range headersHash.Pairs {
					if k, ok := hpair.Key.(*object.String); ok {
						if v, ok := hpair.Value.(*object.String); ok {
							w.Header().Set(k.Value, v.Value)
						}
					}
				}
			}
		}

		w.WriteHeader(status)
		w.Write([]byte(body))
	})

	log.Info("HTTP Server listening", "port", port)
	server := &http.Server{
		Addr:           port,
		Handler:        mux,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP Server error", "error", err)
		}
	}()

	<-quit
	log.Info("HTTP shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return newError("graceful shutdown failed: %s", err)
	}

	signal.Stop(quit)
	log.Info("HTTP server stopped")
	return &object.Null{}
}

// pathSegmentsFunc splits a URL path into non-empty segments.
// path_segments("/users/:id") → ["users", ":id"]
// path_segments("/users/1?foo=bar") → ["users", "1"]
func pathSegmentsFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("path_segments expects 1 argument (path)")
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return newError("path_segments: argument must be a string")
	}

	// Strip query string
	rawPath := pathStr.Value
	if idx := strings.Index(rawPath, "?"); idx != -1 {
		rawPath = rawPath[:idx]
	}

	parts := strings.Split(rawPath, "/")
	var elements []object.Object
	for _, p := range parts {
		if p != "" {
			elements = append(elements, &object.String{Value: p})
		}
	}
	if elements == nil {
		elements = []object.Object{}
	}
	return &object.Array{Elements: elements}
}

// pathParamFunc extracts a named param from a pattern segment.
// path_param(":id") → "id"   (strips leading ":")
// path_param("users") → ""   (not a param → empty string)
// queryParamsFunc parses query parameters from a URL string and returns a Hash.
func queryParamsFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("query_params expects 1 argument (url string)")
	}
	urlStr, ok := args[0].(*object.String)
	if !ok {
		return newError("query_params: argument must be a string")
	}

	u, err := url.Parse(urlStr.Value)
	if err != nil {
		return newError("query_params: invalid URL: %s", err)
	}

	pairs := make(map[object.HashKey]object.HashPair)
	for key, values := range u.Query() {
		k := &object.String{Value: key}
		if len(values) == 1 {
			pairs[k.HashKey()] = object.HashPair{Key: k, Value: &object.String{Value: values[0]}}
		} else {
			var elems []object.Object
			for _, v := range values {
				elems = append(elems, &object.String{Value: v})
			}
			pairs[k.HashKey()] = object.HashPair{Key: k, Value: &object.Array{Elements: elems}}
		}
	}
	return &object.Hash{Pairs: pairs}
}

// queryParamFunc extracts a single query parameter by name from a URL string.
func queryParamFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("query_param expects 2 arguments (name, url)")
	}
	name, ok1 := args[0].(*object.String)
	urlStr, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("query_param: arguments must be strings")
	}

	u, err := url.Parse(urlStr.Value)
	if err != nil {
		return newError("query_param: invalid URL: %s", err)
	}

	val := u.Query().Get(name.Value)
	if val == "" {
		return &object.Null{}
	}
	return &object.String{Value: val}
}

func pathParamFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("path_param expects 1 argument")
	}
	seg, ok := args[0].(*object.String)
	if !ok {
		return newError("path_param: argument must be a string")
	}
	if strings.HasPrefix(seg.Value, ":") {
		return &object.String{Value: seg.Value[1:]}
	}
	return &object.String{Value: ""}
}


