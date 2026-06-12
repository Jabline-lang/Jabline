package stdlib

import (
	"encoding/json"
	"jabline/pkg/object"
	"sync"
	"time"
)

var (
	healthMu          sync.RWMutex
	livenessCheck     object.Object
	readinessCheck    object.Object
	healthServerAddr  string
)

var HealthBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"health_liveness", &object.Builtin{Fn: healthLiveness}},
	{"health_readiness", &object.Builtin{Fn: healthReadiness}},
	{"health_set_liveness", &object.Builtin{Fn: healthSetLiveness}},
	{"health_set_readiness", &object.Builtin{Fn: healthSetReadiness}},
	{"health_get_liveness", &object.Builtin{Fn: healthGetLiveness}},
	{"health_get_readiness", &object.Builtin{Fn: healthGetReadiness}},
	{"health_handler", &object.Builtin{Fn: healthHandler}},
}

func init() {
	NativeModuleRegistry["_health"] = HealthBuiltins
	NativeModulePrefixes["_health"] = "health_"
}

func healthLiveness(args ...object.Object) object.Object {
	healthMu.RLock()
	fn := livenessCheck
	healthMu.RUnlock()
	if fn == nil {
		return &object.Boolean{Value: true}
	}
	return executeCheck(fn)
}

func healthReadiness(args ...object.Object) object.Object {
	healthMu.RLock()
	fn := readinessCheck
	healthMu.RUnlock()
	if fn == nil {
		return &object.Boolean{Value: true}
	}
	return executeCheck(fn)
}

func executeCheck(fn object.Object) object.Object {
	closure, ok := fn.(*object.Closure)
	if !ok {
		return &object.Boolean{Value: true}
	}
	result := Executor(closure, []object.Object{})
	for {
		c, ok := result.(*object.Channel)
		if !ok {
			break
		}
		result = <-c.Value
	}
	if result.Type() == object.ERROR_OBJ {
		return &object.Hash{
			Pairs: map[object.HashKey]object.HashPair{
				(&object.String{Value: "healthy"}).HashKey():   {Key: &object.String{Value: "healthy"}, Value: &object.Boolean{Value: false}},
				(&object.String{Value: "error"}).HashKey():     {Key: &object.String{Value: "error"}, Value: result},
			},
		}
	}
	if b, ok := result.(*object.Boolean); ok {
		return &object.Hash{
			Pairs: map[object.HashKey]object.HashPair{
				(&object.String{Value: "healthy"}).HashKey(): {Key: &object.String{Value: "healthy"}, Value: &object.Boolean{Value: b.Value}},
			},
		}
	}
	return result
}

func healthSetLiveness(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("health_set_liveness expects 1 arg: (function)")
	}
	healthMu.Lock()
	livenessCheck = args[0]
	healthMu.Unlock()
	return &object.Null{}
}

func healthSetReadiness(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("health_set_readiness expects 1 arg: (function)")
	}
	healthMu.Lock()
	readinessCheck = args[0]
	healthMu.Unlock()
	return &object.Null{}
}

func healthGetLiveness(args ...object.Object) object.Object {
	healthMu.RLock()
	defer healthMu.RUnlock()
	if livenessCheck == nil {
		return &object.Null{}
	}
	return livenessCheck
}

func healthGetReadiness(args ...object.Object) object.Object {
	healthMu.RLock()
	defer healthMu.RUnlock()
	if readinessCheck == nil {
		return &object.Null{}
	}
	return readinessCheck
}

func healthHandler(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("health_handler expects 1 arg: (type_string)")
	}
	t, ok := args[0].(*object.String)
	if !ok {
		return newError("health_handler: arg must be STRING, got %s", args[0].Type())
	}
	var fn func() object.Object
	switch t.Value {
	case "liveness", "live", "healthz":
		fn = func() object.Object {
			result := healthLiveness()
			data, _ := json.Marshal(toFields(result))
			return &object.String{Value: string(data)}
		}
	case "readiness", "ready", "readyz":
		fn = func() object.Object {
			result := healthReadiness()
			data, _ := json.Marshal(toFields(result))
			return &object.String{Value: string(data)}
		}
	default:
		return newError("health_handler: unknown type '%s', use: liveness or readiness", t.Value)
	}
	return &object.Builtin{Fn: func(args ...object.Object) object.Object {
		return fn()
	}}
}

// HealthCheckMiddleware returns a Hash suitable as an HTTP handler response
func HealthCheckHandlerFunc(checkType string) func() object.Object {
	switch checkType {
	case "liveness":
		return func() object.Object {
			result := healthLiveness()
			b, _ := json.Marshal(toFields(result))
			return &object.String{
				Value: string(b),
			}
		}
	case "readiness":
		return func() object.Object {
			result := healthReadiness()
			b, _ := json.Marshal(toFields(result))
			return &object.String{
				Value: string(b),
			}
		}
	}
	return nil
}

var startTime = time.Now()
