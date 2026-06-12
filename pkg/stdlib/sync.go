package stdlib

import (
	"sync"
	"sync/atomic"

	"jabline/pkg/object"
)

var (
	mutexes   = make(map[string]*sync.Mutex)
	rwmutexes = make(map[string]*sync.RWMutex)
	waitgroups = make(map[string]*sync.WaitGroup)
	condvars  = make(map[string]*sync.Cond)
	mu        sync.Mutex
	onceVals  sync.Map
)

func init() {
	NativeModuleRegistry["_sync"] = SyncBuiltins
	NativeModulePrefixes["_sync"] = "sync_"
}

var SyncBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"sync_mutex_new", &object.Builtin{Fn: syncMutexNew}},
	{"sync_mutex_lock", &object.Builtin{Fn: syncMutexLock}},
	{"sync_mutex_unlock", &object.Builtin{Fn: syncMutexUnlock}},
	{"sync_rwmutex_new", &object.Builtin{Fn: syncRWMutexNew}},
	{"sync_rwlock", &object.Builtin{Fn: syncRWLock}},
	{"sync_rwunlock", &object.Builtin{Fn: syncRWUnlock}},
	{"sync_wg_new", &object.Builtin{Fn: syncWGNew}},
	{"sync_wg_add", &object.Builtin{Fn: syncWGAdd}},
	{"sync_wg_done", &object.Builtin{Fn: syncWGDone}},
	{"sync_wg_wait", &object.Builtin{Fn: syncWGWait}},
	{"sync_once", &object.Builtin{Fn: syncOnce}},
	{"sync_atomic_load", &object.Builtin{Fn: syncAtomicLoad}},
	{"sync_atomic_store", &object.Builtin{Fn: syncAtomicStore}},
	{"sync_atomic_add", &object.Builtin{Fn: syncAtomicAdd}},
	{"sync_atomic_swap", &object.Builtin{Fn: syncAtomicSwap}},
	{"sync_cond_new", &object.Builtin{Fn: syncCondNew}},
	{"sync_cond_wait", &object.Builtin{Fn: syncCondWait}},
	{"sync_cond_signal", &object.Builtin{Fn: syncCondSignal}},
	{"sync_cond_broadcast", &object.Builtin{Fn: syncCondBroadcast}},
}

func syncMutexNew(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_mutex_new expects a string name argument")
	}
	mu.Lock()
	defer mu.Unlock()
	_, exists := mutexes[id]
	if exists {
		return newError("sync_mutex_new: mutex '%s' already exists", id)
	}
	mutexes[id] = &sync.Mutex{}
	return &object.String{Value: "mutex:" + id}
}

func syncMutexLock(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_mutex_lock expects a string ID")
	}
	mu.Lock()
	m, ok := mutexes[id]
	mu.Unlock()
	if !ok {
		return newError("sync_mutex_lock: mutex '%s' not found", id)
	}
	m.Lock()
	return object.NullObj
}

func syncMutexUnlock(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_mutex_unlock expects a string ID")
	}
	mu.Lock()
	m, ok := mutexes[id]
	mu.Unlock()
	if !ok {
		return newError("sync_mutex_unlock: mutex '%s' not found", id)
	}
	m.Unlock()
	return object.NullObj
}

func syncRWMutexNew(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_rwmutex_new expects a string name argument")
	}
	mu.Lock()
	defer mu.Unlock()
	_, exists := rwmutexes[id]
	if exists {
		return newError("sync_rwmutex_new: rwmutex '%s' already exists", id)
	}
	rwmutexes[id] = &sync.RWMutex{}
	return &object.String{Value: "rwmutex:" + id}
}

func syncRWLock(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_rwlock expects a string ID")
	}
	mu.Lock()
	m, ok := rwmutexes[id]
	mu.Unlock()
	if !ok {
		return newError("sync_rwlock: rwmutex '%s' not found", id)
	}

	if len(args) > 1 {
		if b, ok := args[1].(*object.Boolean); ok && b.Value {
			m.Lock() // Write lock
			return object.NullObj
		}
	}

	m.RLock() // Read lock
	return object.NullObj
}

func syncRWUnlock(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_rwunlock expects a string ID")
	}
	mu.Lock()
	m, ok := rwmutexes[id]
	mu.Unlock()
	if !ok {
		return newError("sync_rwunlock: rwmutex '%s' not found", id)
	}

	if len(args) > 1 {
		if b, ok := args[1].(*object.Boolean); ok && b.Value {
			m.Unlock()
			return object.NullObj
		}
	}

	m.RUnlock()
	return object.NullObj
}

func syncWGNew(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_wg_new expects a string name argument")
	}
	mu.Lock()
	defer mu.Unlock()
	_, exists := waitgroups[id]
	if exists {
		return newError("sync_wg_new: waitgroup '%s' already exists", id)
	}
	waitgroups[id] = &sync.WaitGroup{}
	return &object.String{Value: "wg:" + id}
}

func syncWGAdd(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("sync_wg_add expects ID and delta")
	}
	id, ok := args[0].(*object.String)
	if !ok {
		return newError("sync_wg_add expects string ID, got %s", args[0].Type())
	}
	delta, ok := args[1].(*object.Integer)
	if !ok {
		return newError("sync_wg_add expects integer delta, got %s", args[1].Type())
	}
	mu.Lock()
	wg, ok := waitgroups[stripPrefix(id.Value)]
	mu.Unlock()
	if !ok {
		return newError("sync_wg_add: waitgroup '%s' not found", id.Value)
	}
	wg.Add(int(delta.Value))
	return object.NullObj
}

func syncWGDone(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_wg_done expects a string ID")
	}
	mu.Lock()
	wg, ok := waitgroups[id]
	mu.Unlock()
	if !ok {
		return newError("sync_wg_done: waitgroup '%s' not found", id)
	}
	wg.Done()
	return object.NullObj
}

func syncWGWait(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_wg_wait expects a string ID")
	}
	mu.Lock()
	wg, ok := waitgroups[id]
	mu.Unlock()
	if !ok {
		return newError("sync_wg_wait: waitgroup '%s' not found", id)
	}
	wg.Wait()
	return object.NullObj
}

func syncOnce(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("sync_once expects 2 arguments, got %d", len(args))
	}

	key, ok := args[0].(*object.String)
	if !ok {
		return newError("sync_once expects string key, got %s", args[0].Type())
	}

	fn, ok := args[1].(*object.Closure)
	if !ok && args[1].Type() == "BUILTIN" {
		// Execute immediately for builtins
		builtin := args[1].(*object.Builtin)
		return builtin.Fn()
	}

	_, loaded := onceVals.LoadOrStore(key.Value, true)
	if loaded {
		return object.NullObj
	}

	if ok {
		return Executor(fn, []object.Object{})
	}

	return newError("sync_once expects a function, got %s", args[1].Type())
}

var atomicStore = sync.Map{}

func syncAtomicLoad(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("sync_atomic_load expects 1 argument, got %d", len(args))
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("sync_atomic_load expects string key, got %s", args[0].Type())
	}

	val, ok := atomicStore.Load(key.Value)
	if !ok {
		return object.NullObj
	}
	return val.(object.Object)
}

func syncAtomicStore(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("sync_atomic_store expects 2 arguments, got %d", len(args))
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("sync_atomic_store expects string key, got %s", args[0].Type())
	}
	atomicStore.Store(key.Value, args[1])
	return args[1]
}

func syncAtomicAdd(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("sync_atomic_add expects 2 arguments, got %d", len(args))
	}
	keyStr, ok := args[0].(*object.String)
	if !ok {
		return newError("sync_atomic_add expects string key, got %s", args[0].Type())
	}
	delta, ok := args[1].(*object.Integer)
	if !ok {
		return newError("sync_atomic_add expects integer delta, got %s", args[1].Type())
	}

	// Load or create a per-key counter
	val, _ := atomicCounters.LoadOrStore(keyStr.Value, new(int64))
	counter := val.(*int64)
	actual := atomic.AddInt64(counter, delta.Value)
	return &object.Integer{Value: actual}
}

var atomicCounters sync.Map

func syncAtomicSwap(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("sync_atomic_swap expects 2 arguments, got %d", len(args))
	}
	keyStr, ok := args[0].(*object.String)
	if !ok {
		return newError("sync_atomic_swap expects string key, got %s", args[0].Type())
	}
	atomicStore.Store(keyStr.Value, args[1])
	return args[1]
}

func syncCondNew(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_cond_new expects a string name")
	}
	mu.Lock()
	defer mu.Unlock()
	_, exists := condvars[id]
	if exists {
		return newError("sync_cond_new: cond '%s' already exists", id)
	}
	condvars[id] = sync.NewCond(&sync.Mutex{})
	return &object.String{Value: "cond:" + id}
}

func syncCondWait(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_cond_wait expects a string ID")
	}
	mu.Lock()
	c, ok := condvars[id]
	mu.Unlock()
	if !ok {
		return newError("sync_cond_wait: cond '%s' not found", id)
	}
	c.L.Lock()
	c.Wait()
	c.L.Unlock()
	return object.NullObj
}

func syncCondSignal(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_cond_signal expects a string ID")
	}
	mu.Lock()
	c, ok := condvars[id]
	mu.Unlock()
	if !ok {
		return newError("sync_cond_signal: cond '%s' not found", id)
	}
	c.Signal()
	return object.NullObj
}

func syncCondBroadcast(args ...object.Object) object.Object {
	id := getID(args)
	if id == "" {
		return newError("sync_cond_broadcast expects a string ID")
	}
	mu.Lock()
	c, ok := condvars[id]
	mu.Unlock()
	if !ok {
		return newError("sync_cond_broadcast: cond '%s' not found", id)
	}
	c.Broadcast()
	return object.NullObj
}

// Helper: get ID from args, stripping prefix like "mutex:", "wg:", etc.
func getID(args []object.Object) string {
	if len(args) == 0 {
		return ""
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return ""
	}
	return stripPrefix(s.Value)
}

func stripPrefix(s string) string {
	for _, prefix := range []string{"mutex:", "rwmutex:", "wg:", "cond:"} {
		if len(s) > len(prefix) && s[:len(prefix)] == prefix {
			return s[len(prefix):]
		}
	}
	return s
}
