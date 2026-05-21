package vm

import (
	"sync"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

type Frame struct {
	cl             *object.Closure
	ip             int
	basePointer    int
	savedGlobals   *GlobalStore
	savedConstants []object.Object
	TypeArgs       map[string]string
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
	// keep f.TypeArgs allocated but clear it if we need to, though right now we just re-make it or leave it
	f.TypeArgs = make(map[string]string)
	return f
}

func ReleaseFrame(f *Frame) {
	f.cl = nil
	f.savedGlobals = nil
	f.savedConstants = nil
	framePool.Put(f)
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
