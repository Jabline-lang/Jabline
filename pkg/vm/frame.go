package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/object"
)

type Frame struct {
	cl             *object.Closure
	ip             int
	basePointer    int
	savedGlobals   []object.Object
	savedConstants []object.Object
	TypeArgs       map[string]string
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          -1,
		basePointer: basePointer,
		TypeArgs:    make(map[string]string),
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
