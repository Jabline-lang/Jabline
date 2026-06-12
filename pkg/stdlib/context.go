package stdlib

import (
	"context"
	"jabline/pkg/object"
	"sync"
	"time"
)

type cancelContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	contextsMu sync.Mutex
	contexts   = make(map[string]*cancelContext)
	ctxCounter int64
)

var ContextBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"with_timeout", &object.Builtin{Fn: withTimeoutFunc}},
	{"with_cancel", &object.Builtin{Fn: withCancelFunc}},
	{"with_value", &object.Builtin{Fn: withValueFunc}},
	{"ctx_done", &object.Builtin{Fn: ctxDone}},
	{"ctx_cancel", &object.Builtin{Fn: ctxCancel}},
	{"ctx_deadline", &object.Builtin{Fn: ctxDeadline}},
	{"ctx_err", &object.Builtin{Fn: ctxErr}},
}

func init() {
	NativeModuleRegistry["_context"] = ContextBuiltins
	NativeModulePrefixes["_context"] = "ctx_"
}

func newCtxID() string {
	contextsMu.Lock()
	ctxCounter++
	contextsMu.Unlock()
	return "ctx"
}

// ctxIDFromArgs attempts to extract a context ID from args, defaulting to "ctx"
func ctxIDFromArgs(args []object.Object) string {
	for _, a := range args {
		if s, ok := a.(*object.String); ok {
			return stripPrefixCtx(s.Value)
		}
	}
	return "ctx"
}

func stripPrefixCtx(s string) string {
	for _, p := range []string{"ctx:", "cancel:", "timeout:"} {
		if len(s) > len(p) && s[:len(p)] == p {
			return s[len(p):]
		}
	}
	return s
}

func withTimeoutFunc(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("with_timeout expects (ms, fn)")
	}
	ms, ok := args[0].(*object.Integer)
	if !ok {
		return newError("with_timeout: first arg must be INTEGER (ms), got %s", args[0].Type())
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("with_timeout: second arg must be FUNCTION, got %s", args[1].Type())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ms.Value)*time.Millisecond)
	defer cancel()

	id := "with_timeout"
	contextsMu.Lock()
	contexts[id] = &cancelContext{ctx: ctx, cancel: cancel}
	contextsMu.Unlock()
	defer func() {
		contextsMu.Lock()
		delete(contexts, id)
		contextsMu.Unlock()
	}()

	done := make(chan struct{})
	var result object.Object

	go func() {
		result = Executor(fn, []object.Object{&object.String{Value: "ctx:" + id}})
		close(done)
	}()

	select {
	case <-done:
		// Unwrap channels
		for {
			c, ok := result.(*object.Channel)
			if !ok {
				break
			}
			result = <-c.Value
		}
		if ctx.Err() != nil {
			return newError("with_timeout: context cancelled: %s", ctx.Err())
		}
		return result
	case <-ctx.Done():
		return newError("with_timeout: deadline exceeded (%dms)", ms.Value)
	}
}

func withCancelFunc(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("with_cancel expects (fn)")
	}
	fn, ok := args[0].(*object.Closure)
	if !ok {
		return newError("with_cancel: first arg must be FUNCTION, got %s", args[0].Type())
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctxID := "with_cancel"

	contextsMu.Lock()
	contexts[ctxID] = &cancelContext{ctx: ctx, cancel: cancel}
	contextsMu.Unlock()

	defer func() {
		contextsMu.Lock()
		delete(contexts, ctxID)
		contextsMu.Unlock()
	}()

	done := make(chan struct{})
	var result object.Object

	go func() {
		result = Executor(fn, []object.Object{&object.String{Value: "ctx:" + ctxID}})
		close(done)
	}()

	select {
	case <-done:
		if ctx.Err() != nil {
			return &object.Hash{
				Pairs: map[object.HashKey]object.HashPair{
					(&object.String{Value: "cancelled"}).HashKey(): {Key: &object.String{Value: "cancelled"}, Value: &object.Boolean{Value: true}},
					(&object.String{Value: "error"}).HashKey():     {Key: &object.String{Value: "error"}, Value: &object.String{Value: ctx.Err().Error()}},
				},
			}
		}
		return result
	case <-ctx.Done():
		return &object.Hash{
			Pairs: map[object.HashKey]object.HashPair{
				(&object.String{Value: "cancelled"}).HashKey(): {Key: &object.String{Value: "cancelled"}, Value: &object.Boolean{Value: true}},
				(&object.String{Value: "error"}).HashKey():     {Key: &object.String{Value: "error"}, Value: &object.String{Value: "context cancelled"}},
			},
		}
	}
}

func withValueFunc(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("with_value expects (key, value, fn)")
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("with_value: key must be STRING, got %s", args[0].Type())
	}
	var val string
	if s, ok := args[1].(*object.String); ok {
		val = s.Value
	} else {
		val = args[1].Inspect()
	}
	fn, ok := args[2].(*object.Closure)
	if len(args) >= 3 && !ok {
		// Try with 2 args (key, value) -> return context, or with 3 args (key, value, fn)
	}
	
	if !ok && len(args) >= 3 {
		return newError("with_value: third arg must be FUNCTION, got %s", args[2].Type())
	}
	
	if ok {
		ctx := context.WithValue(context.Background(), key.Value, val)
		ctxID := key.Value
		contextsMu.Lock()
		contexts[ctxID] = &cancelContext{ctx: ctx, cancel: func() {}}
		contextsMu.Unlock()
		defer func() {
			contextsMu.Lock()
			delete(contexts, ctxID)
			contextsMu.Unlock()
		}()
		return Executor(fn, []object.Object{&object.String{Value: "ctx:" + ctxID}})
	}

	// Just return a context
	ctx := context.WithValue(context.Background(), key.Value, val)
	ctxID := key.Value
	contextsMu.Lock()
	contexts[ctxID] = &cancelContext{ctx: ctx, cancel: func() {}}
	contextsMu.Unlock()
	return &object.String{Value: "ctx:" + ctxID}
}

func ctxDone(args ...object.Object) object.Object {
	id := ctxIDFromArgs(args)
	contextsMu.Lock()
	cc, exists := contexts[id]
	contextsMu.Unlock()
	if !exists {
		return &object.Boolean{Value: false}
	}
	select {
	case <-cc.ctx.Done():
		return &object.Boolean{Value: true}
	default:
		return &object.Boolean{Value: false}
	}
}

func ctxCancel(args ...object.Object) object.Object {
	id := ctxIDFromArgs(args)
	contextsMu.Lock()
	cc, exists := contexts[id]
	contextsMu.Unlock()
	if !exists {
		return newError("ctx_cancel: context '%s' not found", id)
	}
	cc.cancel()
	return &object.Null{}
}

func ctxDeadline(args ...object.Object) object.Object {
	id := ctxIDFromArgs(args)
	contextsMu.Lock()
	cc, exists := contexts[id]
	contextsMu.Unlock()
	if !exists {
		return &object.Null{}
	}
	deadline, ok := cc.ctx.Deadline()
	if !ok {
		return &object.Null{}
	}
	return &object.String{Value: deadline.Format(time.RFC3339)}
}

func ctxErr(args ...object.Object) object.Object {
	id := ctxIDFromArgs(args)
	contextsMu.Lock()
	cc, exists := contexts[id]
	contextsMu.Unlock()
	if !exists {
		return &object.Null{}
	}
	err := cc.ctx.Err()
	if err == nil {
		return &object.Null{}
	}
	return &object.String{Value: err.Error()}
}
