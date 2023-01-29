package vm

import (
	"github.com/songzhibin97/mini-compiler/code"
	"github.com/songzhibin97/mini-compiler/compiler"
)

type Frame struct {
	cl          *compiler.Closure
	ip          int
	basePointer int
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}

func NewFrame(cl *compiler.Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          -1,
		basePointer: basePointer,
	}
}
