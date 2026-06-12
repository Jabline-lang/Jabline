package stdlib

import (
	"jabline/pkg/object"
	"sync"
	"sync/atomic"
	"time"
)

type circuitStateType int32

const (
	circuitClosed   = 0
	circuitOpen     = 1
	circuitHalfOpen = 2
)

type circuitBreaker struct {
	name         string
	state        int32 // 0=closed, 1=open, 2=half_open
	failureCount int64
	threshold    int64
	timeout      time.Duration
	lastFailure  time.Time
	mu           sync.Mutex
}

var (
	circuitsMu sync.Mutex
	circuits   = make(map[string]*circuitBreaker)
)

var ResilienceBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"retry", &object.Builtin{Fn: retryFunc}},
	{"retry_with_backoff", &object.Builtin{Fn: retryWithBackoff}},
	{"circuit_new", &object.Builtin{Fn: circuitNew}},
	{"circuit_call", &object.Builtin{Fn: circuitCall}},
	{"circuit_state", &object.Builtin{Fn: circuitGetState}},
	{"circuit_reset", &object.Builtin{Fn: circuitReset}},
	{"timeout", &object.Builtin{Fn: timeoutFunc}},
	{"backpressure", &object.Builtin{Fn: backpressureFunc}},
	{"backpressure_acquire", &object.Builtin{Fn: backpressureAcquire}},
	{"backpressure_release", &object.Builtin{Fn: backpressureRelease}},
}

func init() {
	NativeModuleRegistry["_resilience"] = ResilienceBuiltins
	NativeModulePrefixes["_resilience"] = ""
}

func retryFunc(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("retry expects (fn, max_attempts)")
	}
	fn, ok := args[0].(*object.Closure)
	if !ok {
		return newError("retry: first arg must be FUNCTION, got %s", args[0].Type())
	}
	maxAttempts, ok := args[1].(*object.Integer)
	if !ok || maxAttempts.Value < 1 {
		return newError("retry: max_attempts must be positive INTEGER")
	}

	max := int(maxAttempts.Value)
	for i := 0; i < max; i++ {
		result := Executor(fn, []object.Object{})
		// Unwrap channel
		for {
			c, ok := result.(*object.Channel)
			if !ok {
				break
			}
			result = <-c.Value
		}
		if result.Type() != object.ERROR_OBJ {
			return result
		}
		if i < max-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return newError("retry: all %d attempts failed", max)
}

func retryWithBackoff(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("retry_with_backoff expects (fn, max_attempts, [base_delay_ms])")
	}
	fn, ok := args[0].(*object.Closure)
	if !ok {
		return newError("retry_with_backoff: first arg must be FUNCTION, got %s", args[0].Type())
	}
	maxAttempts, ok := args[1].(*object.Integer)
	if !ok || maxAttempts.Value < 1 {
		return newError("retry_with_backoff: max_attempts must be positive INTEGER")
	}

	baseDelay := 100
	if len(args) >= 3 {
		if d, ok := args[2].(*object.Integer); ok && d.Value > 0 {
			baseDelay = int(d.Value)
		}
	}

	max := int(maxAttempts.Value)
	for i := 0; i < max; i++ {
		result := Executor(fn, []object.Object{})
		for {
			c, ok := result.(*object.Channel)
			if !ok {
				break
			}
			result = <-c.Value
		}
		if result.Type() != object.ERROR_OBJ {
			return result
		}
		if i < max-1 {
			// Exponential backoff with jitter
			delay := time.Duration(baseDelay*(1<<uint(i))) * time.Millisecond
			if delay > 10*time.Second {
				delay = 10 * time.Second
			}
			time.Sleep(delay)
		}
	}
	return newError("retry: all %d attempts failed", max)
}

func circuitNew(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("circuit_new expects (name, threshold, timeout_ms)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("circuit_new: name must be STRING, got %s", args[0].Type())
	}
	threshold, ok := args[1].(*object.Integer)
	if !ok || threshold.Value < 1 {
		return newError("circuit_new: threshold must be positive INTEGER")
	}
	timeoutMs, ok := args[2].(*object.Integer)
	if !ok || timeoutMs.Value < 1 {
		return newError("circuit_new: timeout_ms must be positive INTEGER")
	}

	circuitsMu.Lock()
	defer circuitsMu.Unlock()

	if _, exists := circuits[name.Value]; exists {
		return &object.String{Value: "circuit:" + name.Value}
	}

	circuits[name.Value] = &circuitBreaker{
		name:      name.Value,
		state:     int32(circuitClosed),
		threshold: threshold.Value,
		timeout:   time.Duration(timeoutMs.Value) * time.Millisecond,
	}
	return &object.String{Value: "circuit:" + name.Value}
}

func circuitCall(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("circuit_call expects (circuit_name, fn)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("circuit_call: name must be STRING, got %s", args[0].Type())
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("circuit_call: fn must be FUNCTION, got %s", args[1].Type())
	}

	circuitsMu.Lock()
	cb, exists := circuits[stripPrefixCtx(name.Value)]
	circuitsMu.Unlock()
	if !exists {
		return newError("circuit_call: circuit '%s' not found", name.Value)
	}

	cb.mu.Lock()
	currentState := int(atomic.LoadInt32(&cb.state))
	if currentState == circuitOpen {
		if time.Since(cb.lastFailure) > cb.timeout {
			atomic.StoreInt32(&cb.state, int32(circuitHalfOpen))
		} else {
			cb.mu.Unlock()
			return newError("circuit '%s' is OPEN, request rejected", cb.name)
		}
	}
	cb.mu.Unlock()

	result := Executor(fn, []object.Object{})
	for {
		c, ok := result.(*object.Channel)
		if !ok {
			break
		}
		result = <-c.Value
	}

	if result.Type() == object.ERROR_OBJ {
		atomic.AddInt64(&cb.failureCount, 1)
		cb.mu.Lock()
		cb.lastFailure = time.Now()
		if atomic.LoadInt64(&cb.failureCount) >= cb.threshold {
			atomic.StoreInt32(&cb.state, int32(circuitOpen))
		}
		cb.mu.Unlock()
	} else {
		// Success
		atomic.StoreInt64(&cb.failureCount, 0)
	atomic.StoreInt32(&cb.state, int32(circuitClosed))
	}

	return result
}

func circuitGetState(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("circuit_state expects (name)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("circuit_state: name must be STRING, got %s", args[0].Type())
	}
	circuitsMu.Lock()
	cb, exists := circuits[stripPrefixCtx(name.Value)]
	circuitsMu.Unlock()
	if !exists {
		return newError("circuit_state: circuit '%s' not found", name.Value)
	}

	st := int(atomic.LoadInt32(&cb.state))
	var stateStr string
	switch st {
	case circuitClosed:
		stateStr = "closed"
	case circuitOpen:
		stateStr = "open"
	case circuitHalfOpen:
		stateStr = "half_open"
	}

	pairs := make(map[object.HashKey]object.HashPair)
	add := func(k string, v object.Object) {
		ks := &object.String{Value: k}
		pairs[ks.HashKey()] = object.HashPair{Key: ks, Value: v}
	}
	add("name", &object.String{Value: cb.name})
	add("state", &object.String{Value: stateStr})
	add("failures", &object.Integer{Value: atomic.LoadInt64(&cb.failureCount)})
	add("threshold", &object.Integer{Value: cb.threshold})

	return &object.Hash{Pairs: pairs}
}

func circuitReset(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("circuit_reset expects (name)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("circuit_reset: name must be STRING, got %s", args[0].Type())
	}
	circuitsMu.Lock()
	cb, exists := circuits[stripPrefixCtx(name.Value)]
	circuitsMu.Unlock()
	if !exists {
		return newError("circuit_reset: circuit '%s' not found", name.Value)
	}
	atomic.StoreInt64(&cb.failureCount, 0)
	atomic.StoreInt32(&cb.state, int32(circuitClosed))
	return &object.Null{}
}

func timeoutFunc(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("timeout expects (ms, fn)")
	}
	ms, ok := args[0].(*object.Integer)
	if !ok {
		return newError("timeout: first arg must be INTEGER (ms), got %s", args[0].Type())
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("timeout: second arg must be FUNCTION, got %s", args[0].Type())
	}

	done := make(chan object.Object, 1)
	go func() {
		result := Executor(fn, []object.Object{})
		done <- result
	}()

	select {
	case result := <-done:
		return result
	case <-time.After(time.Duration(ms.Value) * time.Millisecond):
		return newError("timeout: function did not complete within %dms", ms.Value)
	}
}

// Backpressure

type bpLimit struct {
	mu      sync.Mutex
	cond    *sync.Cond
	max     int64
	current int64
}

var (
	bpMu    sync.Mutex
	bpLimits = make(map[string]*bpLimit)
)

func backpressureFunc(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("backpressure expects (name, max_concurrent)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("backpressure: name must be STRING, got %s", args[0].Type())
	}
	maxVal, ok := args[1].(*object.Integer)
	if !ok || maxVal.Value < 1 {
		return newError("backpressure: max_concurrent must be positive INTEGER")
	}

	bpMu.Lock()
	defer bpMu.Unlock()

	if _, exists := bpLimits[name.Value]; exists {
		return &object.String{Value: "bp:" + name.Value}
	}

	lim := &bpLimit{max: maxVal.Value}
	lim.cond = sync.NewCond(&lim.mu)
	bpLimits[name.Value] = lim
	return &object.String{Value: "bp:" + name.Value}
}

func backpressureAcquire(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("backpressure_acquire expects (name)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("backpressure_acquire: name must be STRING, got %s", args[0].Type())
	}

	bpMu.Lock()
	lim, exists := bpLimits[stripPrefixCtx(name.Value)]
	bpMu.Unlock()
	if !exists {
		return newError("backpressure_acquire: backpressure '%s' not found", name.Value)
	}

	lim.mu.Lock()
	for lim.current >= lim.max {
		lim.cond.Wait()
	}
	lim.current++
	lim.mu.Unlock()
	return &object.Null{}
}

func backpressureRelease(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("backpressure_release expects (name)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("backpressure_release: name must be STRING, got %s", args[0].Type())
	}

	bpMu.Lock()
	lim, exists := bpLimits[stripPrefixCtx(name.Value)]
	bpMu.Unlock()
	if !exists {
		return newError("backpressure_release: backpressure '%s' not found", name.Value)
	}

	lim.mu.Lock()
	lim.current--
	lim.cond.Signal()
	lim.mu.Unlock()
	return &object.Null{}
}
