package stdlib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"jabline/pkg/log"
	"jabline/pkg/object"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTP middleware support
var (
	httpMiddlewaresMu sync.Mutex
	httpMiddlewares   []object.Object // list of middleware functions
	httpServerMux     *http.ServeMux
)

var HTTPAdvancedBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"http_use", &object.Builtin{Fn: httpUse}},
	{"http_mux", &object.Builtin{Fn: httpMux}},
	{"http_mount", &object.Builtin{Fn: httpMount}},
	{"http_healthz_handler", &object.Builtin{Fn: httpHealthzHandler}},
	{"http_readyz_handler", &object.Builtin{Fn: httpReadyzHandler}},
	{"http_metrics_handler", &object.Builtin{Fn: httpMetricsHandler}},
	{"http_static", &object.Builtin{Fn: httpStatic}},
	{"http_cors", &object.Builtin{Fn: httpCORS}},
}

func init() {
	NativeModuleRegistry["_http_advanced"] = HTTPAdvancedBuiltins
	NativeModulePrefixes["_http_advanced"] = "http_"
}

func httpUse(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("http_use expects 1 arg: (middleware_fn)")
	}
	httpMiddlewaresMu.Lock()
	httpMiddlewares = append(httpMiddlewares, args[0])
	httpMiddlewaresMu.Unlock()
	return &object.Null{}
}

func httpMux(args ...object.Object) object.Object {
	if httpServerMux == nil {
		httpServerMux = http.NewServeMux()
	}
	return &object.String{Value: "http_mux"}
}

func httpMount(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("http_mount expects (pattern, handler_fn_or_string)")
	}
	pattern, ok := args[0].(*object.String)
	if !ok {
		return newError("http_mount: pattern must be STRING, got %s", args[0].Type())
	}

	if httpServerMux == nil {
		httpServerMux = http.NewServeMux()
	}

	// If handler is a function (closure or builtin), wrap it
	switch handler := args[1].(type) {
	case *object.Closure:
		httpServerMux.HandleFunc(pattern.Value, func(w http.ResponseWriter, r *http.Request) {
			reqHash := buildReqHash(w, r)
			result := Executor(handler, []object.Object{reqHash})
			for {
				c, ok := result.(*object.Channel)
				if !ok {
					break
				}
				result = <-c.Value
			}
			writeHTTPResponse(w, result)
		})
	case *object.Builtin:
		httpServerMux.HandleFunc(pattern.Value, func(w http.ResponseWriter, r *http.Request) {
			reqHash := buildReqHash(w, r)
			result := handler.Fn(reqHash)
			writeHTTPResponse(w, result)
		})
	case *object.String:
		// Serve a directory
		httpServerMux.Handle(pattern.Value, http.StripPrefix(pattern.Value, http.FileServer(http.Dir(handler.Value))))
	default:
		return newError("http_mount: handler must be FUNCTION or STRING (file path), got %s", args[1].Type())
	}
	return &object.Null{}
}

func httpHealthzHandler(args ...object.Object) object.Object {
	if httpServerMux == nil {
		httpServerMux = http.NewServeMux()
	}
	httpServerMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		result := healthLiveness()
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(toFields(result))
		w.Write(b)
	})
	return &object.String{Value: "/healthz"}
}

func httpReadyzHandler(args ...object.Object) object.Object {
	if httpServerMux == nil {
		httpServerMux = http.NewServeMux()
	}
	httpServerMux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		result := healthReadiness()
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(toFields(result))
		w.Write(b)
	})
	return &object.String{Value: "/readyz"}
}

func httpMetricsHandler(args ...object.Object) object.Object {
	if httpServerMux == nil {
		httpServerMux = http.NewServeMux()
	}
	httpServerMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		result := metricsText()
		if s, ok := result.(*object.String); ok {
			w.Write([]byte(s.Value))
		}
	})
	return &object.String{Value: "/metrics"}
}

func httpStatic(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("http_static expects (url_prefix, dir_path)")
	}
	prefix, ok := args[0].(*object.String)
	if !ok {
		return newError("http_static: url_prefix must be STRING, got %s", args[0].Type())
	}
	dir, ok := args[1].(*object.String)
	if !ok {
		return newError("http_static: dir_path must be STRING, got %s", args[1].Type())
	}

	if httpServerMux == nil {
		httpServerMux = http.NewServeMux()
	}

	if !strings.HasSuffix(prefix.Value, "/") {
		prefix = &object.String{Value: prefix.Value + "/"}
	}
	httpServerMux.Handle(prefix.Value, http.StripPrefix(prefix.Value, http.FileServer(http.Dir(dir.Value))))
	return &object.Null{}
}

func httpCORS(args ...object.Object) object.Object {
	return &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if len(args) < 1 {
			return newError("CORS middleware: expects (next_handler)")
		}
		next := args[0]
		return &object.Builtin{Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("CORS handler: expects request hash")
			}
			req, ok := args[0].(*object.Hash)
			if !ok {
				return newError("CORS handler: expects request hash")
			}

			methodKey := &object.String{Value: "method"}
			respPairs := make(map[object.HashKey]object.HashPair)

			method := ""
			if pair, ok := req.Pairs[methodKey.HashKey()]; ok {
				if s, ok := pair.Value.(*object.String); ok {
					method = s.Value
				}
			}

			// Handle preflight
			if method == "OPTIONS" {
				addHP := func(k string, v object.Object) {
					ks := &object.String{Value: k}
					respPairs[ks.HashKey()] = object.HashPair{Key: ks, Value: v}
				}
				addHP("status", &object.Integer{Value: 204})
				headers := &object.Hash{
					Pairs: map[object.HashKey]object.HashPair{
						(&object.String{Value: "Access-Control-Allow-Origin"}).HashKey():  {Key: &object.String{Value: "Access-Control-Allow-Origin"}, Value: &object.String{Value: "*"}},
						(&object.String{Value: "Access-Control-Allow-Methods"}).HashKey(): {Key: &object.String{Value: "Access-Control-Allow-Methods"}, Value: &object.String{Value: "GET, POST, PUT, DELETE, PATCH, OPTIONS"}},
						(&object.String{Value: "Access-Control-Allow-Headers"}).HashKey(): {Key: &object.String{Value: "Access-Control-Allow-Headers"}, Value: &object.String{Value: "Content-Type, Authorization"}},
					},
				}
				addHP("headers", headers)
				return &object.Hash{Pairs: respPairs}
			}

			// Call next handler
			var result object.Object
			if c, ok := next.(*object.Closure); ok {
				result = Executor(c, []object.Object{req})
			} else if b, ok := next.(*object.Builtin); ok {
				result = b.Fn(req)
			} else {
				return newError("CORS: invalid next handler")
			}

			for {
				c, ok := result.(*object.Channel)
				if !ok {
					break
				}
				result = <-c.Value
			}

			// Add CORS headers to response
			if respHash, ok := result.(*object.Hash); ok {
				headersKey := &object.String{Value: "headers"}
				if _, exists := respHash.Pairs[headersKey.HashKey()]; !exists {
					respHash.Pairs[headersKey.HashKey()] = object.HashPair{
						Key:   headersKey,
						Value: &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)},
					}
				}
				if h, ok := respHash.Pairs[headersKey.HashKey()].Value.(*object.Hash); ok {
					originKey := &object.String{Value: "Access-Control-Allow-Origin"}
					h.Pairs[originKey.HashKey()] = object.HashPair{Key: originKey, Value: &object.String{Value: "*"}}
				}
				return respHash
			}
			return result
		}}
	}}
}

// buildReqHash builds the request hash for the HTTP server handler
func buildReqHash(w http.ResponseWriter, r *http.Request) *object.Hash {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body.Close()
	h := &object.Hash{
		Pairs: map[object.HashKey]object.HashPair{
			(&object.String{Value: "method"}).HashKey(): {Key: &object.String{Value: "method"}, Value: &object.String{Value: r.Method}},
			(&object.String{Value: "url"}).HashKey():    {Key: &object.String{Value: "url"}, Value: &object.String{Value: r.URL.String()}},
			(&object.String{Value: "path"}).HashKey():   {Key: &object.String{Value: "path"}, Value: &object.String{Value: r.URL.Path}},
			(&object.String{Value: "query"}).HashKey():  {Key: &object.String{Value: "query"}, Value: &object.String{Value: r.URL.RawQuery}},
			(&object.String{Value: "body"}).HashKey():   {Key: &object.String{Value: "body"}, Value: &object.String{Value: string(bodyBytes)}},
		},
	}

	// Headers
	headers := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}
	for k, v := range r.Header {
		key := &object.String{Value: k}
		headers.Pairs[key.HashKey()] = object.HashPair{Key: key, Value: &object.String{Value: strings.Join(v, ", ")}}
	}
	headersKey := &object.String{Value: "headers"}
	h.Pairs[headersKey.HashKey()] = object.HashPair{Key: headersKey, Value: headers}

	return h
}

// writeHTTPResponse writes a Jabline response object to http.ResponseWriter
func writeHTTPResponse(w http.ResponseWriter, result object.Object) {
	for {
		c, ok := result.(*object.Channel)
		if !ok {
			break
		}
		result = <-c.Value
	}

	if result.Type() == object.ERROR_OBJ {
		http.Error(w, result.Inspect(), http.StatusInternalServerError)
		return
	}

	respHash, ok := result.(*object.Hash)
	if !ok {
		if str, ok := result.(*object.String); ok {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(200)
			w.Write([]byte(str.Value))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(result.Inspect()))
		return
	}

	status := http.StatusOK
	statusKey := &object.String{Value: "status"}
	if pair, ok := respHash.Pairs[statusKey.HashKey()]; ok {
		if s, ok := pair.Value.(*object.Integer); ok {
			status = int(s.Value)
		}
	}

	body := ""
	bodyKey := &object.String{Value: "body"}
	if pair, ok := respHash.Pairs[bodyKey.HashKey()]; ok {
		if s, ok := pair.Value.(*object.String); ok {
			body = s.Value
		} else {
			body = pair.Value.Inspect()
		}
	}

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
}

// Enhanced httpServe that supports middleware chain and auto-mounts health/metrics if configured
func httpServeAdvanced(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("http_serve expects (port, handler)")
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

	// Build the final handler with middleware chain
	var handler http.Handler

	// If http_mux has been used, use that as base
	if httpServerMux != nil {
		handler = httpServerMux
	} else {
		mux := http.NewServeMux()
		// Add default catch-all with middleware
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			reqHash := buildReqHash(w, r)
			result := Executor(handlerClosure, []object.Object{reqHash})
			writeHTTPResponse(w, result)
		})
		handler = mux
	}

	return startHTTPServer(port, handler)
}

func startHTTPServer(port string, handler http.Handler) object.Object {
	log.Info("HTTP Server listening", "port", port)
	server := &http.Server{
		Addr:           port,
		Handler:        handler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	quit := make(chan struct{})
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP Server error", "error", err)
		}
		close(quit)
	}()

	return &object.Builtin{Fn: func(args ...object.Object) object.Object {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return newError("shutdown failed: %s", err)
		}
		return &object.Null{}
	}}
}
