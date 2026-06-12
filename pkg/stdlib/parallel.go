package stdlib

import (
	"jabline/pkg/object"
	"sync"
)

var ParallelBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"parallel_map", &object.Builtin{Fn: parallelMap}},
	{"parallel_filter", &object.Builtin{Fn: parallelFilter}},
	{"parallel_reduce", &object.Builtin{Fn: parallelReduce}},
	{"parallel_each", &object.Builtin{Fn: parallelEach}},
	{"parallel_pool", &object.Builtin{Fn: parallelPool}},
	{"parallel_pool_submit", &object.Builtin{Fn: parallelPoolSubmit}},
	{"parallel_pool_await", &object.Builtin{Fn: parallelPoolAwait}},
}

func init() {
	NativeModuleRegistry["_parallel"] = ParallelBuiltins
	NativeModulePrefixes["_parallel"] = "parallel_"
}

func parallelMap(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("parallel_map expects (array, fn)")
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("parallel_map: first arg must be ARRAY, got %s", args[0].Type())
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("parallel_map: second arg must be FUNCTION, got %s", args[1].Type())
	}

	numWorkers := len(arr.Elements)
	if len(args) >= 3 {
		if n, ok := args[2].(*object.Integer); ok && n.Value > 0 {
			numWorkers = int(n.Value)
		}
	}
	if numWorkers > len(arr.Elements) {
		numWorkers = len(arr.Elements)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	type result struct {
		idx int
		val object.Object
	}

	results := make([]object.Object, len(arr.Elements))
	ch := make(chan result, len(arr.Elements))
	var wg sync.WaitGroup

	work := make(chan int, len(arr.Elements))
	for i := range arr.Elements {
		work <- i
	}
	close(work)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range work {
				res := Executor(fn, []object.Object{arr.Elements[idx], &object.Integer{Value: int64(idx)}})
				ch <- result{idx, res}
			}
		}()
	}

	wg.Wait()
	close(ch)

	for r := range ch {
		results[r.idx] = r.val
	}

	return &object.Array{Elements: results}
}

func parallelFilter(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("parallel_filter expects (array, fn)")
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("parallel_filter: first arg must be ARRAY, got %s", args[0].Type())
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("parallel_filter: second arg must be FUNCTION, got %s", args[1].Type())
	}

	numWorkers := len(arr.Elements)
	if len(args) >= 3 {
		if n, ok := args[2].(*object.Integer); ok && n.Value > 0 {
			numWorkers = int(n.Value)
		}
	}
	if numWorkers > len(arr.Elements) {
		numWorkers = len(arr.Elements)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	type filterResult struct {
		idx   int
		pass  bool
	}

	ch := make(chan filterResult, len(arr.Elements))
	var wg sync.WaitGroup

	work := make(chan int, len(arr.Elements))
	for i := range arr.Elements {
		work <- i
	}
	close(work)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range work {
				res := Executor(fn, []object.Object{arr.Elements[idx], &object.Integer{Value: int64(idx)}})
				// Unwrap channel
				for {
					c, ok := res.(*object.Channel)
					if !ok {
						break
					}
					res = <-c.Value
				}
				pass := false
				if b, ok := res.(*object.Boolean); ok {
					pass = b.Value
				}
				ch <- filterResult{idx, pass}
			}
		}()
	}

	wg.Wait()
	close(ch)

	filtered := make([]int, 0)
	for r := range ch {
		if r.pass {
			filtered = append(filtered, r.idx)
		}
	}

	elements := make([]object.Object, len(filtered))
	for i, idx := range filtered {
		elements[i] = arr.Elements[idx]
	}

	return &object.Array{Elements: elements}
}

func parallelReduce(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("parallel_reduce expects (array, fn, initial)")
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("parallel_reduce: first arg must be ARRAY, got %s", args[0].Type())
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("parallel_reduce: second arg must be FUNCTION, got %s", args[1].Type())
	}

	// Sequential reduce is safer (parallel reduce requires associativity guarantees)
	acc := args[2]
	for i, el := range arr.Elements {
		res := Executor(fn, []object.Object{acc, el, &object.Integer{Value: int64(i)}})
		// Unwrap channel
		for {
			c, ok := res.(*object.Channel)
			if !ok {
				break
			}
			res = <-c.Value
		}
		if res.Type() == object.ERROR_OBJ {
			return res
		}
		acc = res
	}
	return acc
}

func parallelEach(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("parallel_each expects (array, fn)")
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("parallel_each: first arg must be ARRAY, got %s", args[0].Type())
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("parallel_each: second arg must be FUNCTION, got %s", args[1].Type())
	}

	var wg sync.WaitGroup
	for i, el := range arr.Elements {
		wg.Add(1)
		go func(idx int, elem object.Object) {
			defer wg.Done()
			Executor(fn, []object.Object{elem, &object.Integer{Value: int64(idx)}})
		}(i, el)
	}
	wg.Wait()
	return &object.Null{}
}

type workerPool struct {
	mu     sync.Mutex
	jobs   chan poolJob
	results []object.Object
	wg     sync.WaitGroup
}

type poolJob struct {
	idx int
	fn  *object.Closure
	args []object.Object
}

var (
	poolsMu sync.Mutex
	pools   = make(map[string]*workerPool)
)

func parallelPool(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("parallel_pool expects (name, size)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("parallel_pool: name must be STRING, got %s", args[0].Type())
	}
	size, ok := args[1].(*object.Integer)
	if !ok || size.Value < 1 {
		return newError("parallel_pool: size must be positive INTEGER")
	}

	poolsMu.Lock()
	defer poolsMu.Unlock()

	if _, exists := pools[name.Value]; exists {
		return newError("parallel_pool: pool '%s' already exists", name.Value)
	}

	pool := &workerPool{
		jobs: make(chan poolJob, 1000),
	}
	pools[name.Value] = pool

	for w := 0; w < int(size.Value); w++ {
		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()
			for job := range pool.jobs {
				_ = Executor(job.fn, job.args)
			}
		}()
	}

	return &object.String{Value: "pool:" + name.Value}
}

func parallelPoolSubmit(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("parallel_pool_submit expects (pool_name, fn, [args...])")
	}
	poolName, ok := args[0].(*object.String)
	if !ok {
		return newError("parallel_pool_submit: pool_name must be STRING")
	}
	fn, ok := args[1].(*object.Closure)
	if !ok {
		return newError("parallel_pool_submit: fn must be FUNCTION")
	}

	poolsMu.Lock()
	pool, exists := pools[stripPrefix(poolName.Value)]
	poolsMu.Unlock()
	if !exists {
		return newError("parallel_pool_submit: pool '%s' not found", poolName.Value)
	}

	fnArgs := make([]object.Object, 0)
	if len(args) > 2 {
		fnArgs = args[2:]
	}

	pool.mu.Lock()
	idx := len(pool.results)
	pool.results = append(pool.results, &object.Null{})
	pool.mu.Unlock()

	pool.jobs <- poolJob{idx: idx, fn: fn, args: fnArgs}
	return &object.Integer{Value: int64(idx)}
}

func parallelPoolAwait(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("parallel_pool_await expects (pool_name)")
	}
	poolName, ok := args[0].(*object.String)
	if !ok {
		return newError("parallel_pool_await: pool_name must be STRING")
	}

	poolsMu.Lock()
	pool, exists := pools[stripPrefix(poolName.Value)]
	poolsMu.Unlock()
	if !exists {
		return newError("parallel_pool_await: pool '%s' not found", poolName.Value)
	}

	close(pool.jobs)
	pool.wg.Wait()

	pool.mu.Lock()
	results := make([]object.Object, len(pool.results))
	copy(results, pool.results)
	pool.mu.Unlock()

	delete(pools, stripPrefix(poolName.Value))
	return &object.Array{Elements: results}
}
