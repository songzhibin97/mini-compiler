package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbolTable_Define(t *testing.T) {
	global := NewSymbolTable()
	a := global.Define("a")
	b := global.Define("b")
	local := NewEnclosedSymbolTable(global)
	c := local.Define("c")
	d := local.Define("d")
	next := NewEnclosedSymbolTable(local)
	e := next.Define("e")
	f := next.Define("f")
	assert.Equal(t, a, Symbol{
		Scope: GlobalScope,
		Index: 0,
		Name:  "a",
	})
	assert.Equal(t, b, Symbol{
		Scope: GlobalScope,
		Index: 1,
		Name:  "b",
	})
	assert.Equal(t, c, Symbol{
		Scope: LocalScope,
		Index: 0,
		Name:  "c",
	})
	assert.Equal(t, d, Symbol{
		Scope: LocalScope,
		Index: 1,
		Name:  "d",
	})
	assert.Equal(t, e, Symbol{
		Scope: LocalScope,
		Index: 0,
		Name:  "e",
	})
	assert.Equal(t, f, Symbol{
		Scope: LocalScope,
		Index: 1,
		Name:  "f",
	})
}

func TestSymbolTable_GetDefine(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")
	bs := []Symbol{
		{
			Scope: BuiltinScope,
			Index: 0,
			Name:  "ba",
		},
		{
			Scope: BuiltinScope,
			Index: 1,
			Name:  "bb",
		},
		{
			Scope: BuiltinScope,
			Index: 2,
			Name:  "bc",
		},
		{
			Scope: BuiltinScope,
			Index: 3,
			Name:  "bd",
		},
	}
	for _, expect := range bs {
		global.DefineBuiltin(expect.Index, expect.Name)
	}
	cs := []Symbol{
		{
			Scope: ContextScope,
			Index: 0,
			Name:  "ca",
		},
		{
			Scope: ContextScope,
			Index: 1,
			Name:  "cb",
		},
		{
			Scope: ContextScope,
			Index: 2,
			Name:  "cc",
		},
		{
			Scope: ContextScope,
			Index: 3,
			Name:  "cd",
		},
	}
	for _, expect := range cs {
		global.defineContext(expect)
	}
	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")
	tests := []Symbol{
		{
			Scope: GlobalScope,
			Index: 0,
			Name:  "a",
		},
		{
			Scope: GlobalScope,
			Index: 1,
			Name:  "b",
		},
		{
			Scope: LocalScope,
			Index: 0,
			Name:  "c",
		},
		{
			Scope: LocalScope,
			Index: 1,
			Name:  "d",
		},
	}
	for _, test := range tests {
		v, ok := local.GetDefine(test.Name)
		assert.Equal(t, ok, true)
		assert.Equal(t, v, test)
	}
	for _, table := range []*SymbolTable{global, local} {
		for _, expect := range bs {
			v, ok := table.GetDefine(expect.Name)
			assert.Equal(t, ok, true)
			assert.Equal(t, v, expect)
		}
	}
	for _, expect := range cs {
		v, ok := global.GetDefine(expect.Name)
		assert.Equal(t, ok, true)
		assert.Equal(t, v, expect)
	}
}
