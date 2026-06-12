package vm

import (
	"sync"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

// DeferredCall holds a function and its arguments for deferred execution.
type DeferredCall struct {
	Fn   object.Object
	Args []object.Object
}

type Frame struct {
	cl             *object.Closure
	ip             int
	basePointer    int
	savedGlobals   *GlobalStore
	savedConstants []object.Object
	TypeArgs       map[string]string

	// Deferred execution support
	deferred    []DeferredCall
	deferIndex  int           // -1 = not deferring, 0..N = next call to execute (from end)
	savedReturn object.Object // return value saved while executing deferred calls
}

var framePool = sync.Pool{
	New: func() interface{} {
		return &Frame{
			TypeArgs: make(map[string]string),
		}
	},
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	f := framePool.Get().(*Frame)
	f.cl = cl
	f.ip = -1
	f.basePointer = basePointer
	f.savedGlobals = nil
	f.savedConstants = nil
	f.deferred = nil
	f.deferIndex = -1
	f.savedReturn = nil
	f.TypeArgs = make(map[string]string)
	return f
}

func ReleaseFrame(f *Frame) {
	f.cl = nil
	f.savedGlobals = nil
	f.savedConstants = nil
	f.deferred = nil
	f.savedReturn = nil
	f.TypeArgs = nil
	framePool.Put(f)
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}

// Exported accessors for DAP
func (f *Frame) IP() int {
	return f.ip
}

func (f *Frame) BasePointer() int {
	return f.basePointer
}

func (f *Frame) Cl() *object.Closure {
	return f.cl
}
