package stdlib

import (
	"jabline/pkg/object"
	"runtime"
)

var RuntimeBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"gc", &object.Builtin{Fn: runtimeGC}},
	{"memoryStats", &object.Builtin{Fn: runtimeMemoryStats}},
}

func runtimeGC(args ...object.Object) object.Object {
	if len(args) != 0 {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}
	runtime.GC()
	return &object.Null{}
}

func runtimeMemoryStats(args ...object.Object) object.Object {
	if len(args) != 0 {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pairs := make(map[object.HashKey]object.HashPair)

	addMetric := func(key string, value uint64) {
		k := &object.String{Value: key}
		v := &object.Integer{Value: int64(value)}
		pairs[k.HashKey()] = object.HashPair{Key: k, Value: v}
	}

	addMetric("alloc", m.Alloc)
	addMetric("totalAlloc", m.TotalAlloc)
	addMetric("sys", m.Sys)
	addMetric("numGC", uint64(m.NumGC))
	addMetric("mallocs", m.Mallocs)
	addMetric("frees", m.Frees)
	addMetric("heapAlloc", m.HeapAlloc)
	addMetric("heapSys", m.HeapSys)

	return &object.Hash{Pairs: pairs}
}
