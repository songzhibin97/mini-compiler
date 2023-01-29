package benchmark

import (
	"testing"

	"github.com/songzhibin97/mini-compiler/vm"

	"github.com/songzhibin97/mini-compiler/compiler"

	"github.com/songzhibin97/mini-interpreter/eval"
	"github.com/songzhibin97/mini-interpreter/object"

	"github.com/songzhibin97/mini-interpreter/lexer"
	"github.com/songzhibin97/mini-interpreter/parser"
)

var input = `func fibonacci(a) {if (a < 0) { return 0 } else { return fibonacci(a-1) + fibonacci(a-2) }} fibonacci(10)`

func BenchmarkInterpreter(b *testing.B) {
	env := object.NewEnv(nil)
	p := parser.NewParser(lexer.NewLexer(input))
	program := p.ParseProgram()
	for i := 0; i < b.N; i++ {
		eval.Eval(program, env)
	}
}

func BenchmarkCompiler(b *testing.B) {
	p := parser.NewParser(lexer.NewLexer(input))
	program := p.ParseProgram()
	comp := compiler.NewCompiler()
	err := comp.Compiler(program)
	if err != nil {
		b.Error(err)
	}
	for i := 0; i < b.N; i++ {
		v := vm.NewVM(comp.Bytecode())
		err = v.Run()
		if err != nil {
			b.Error(err)
		}
	}
}
