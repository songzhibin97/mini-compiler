package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/songzhibin97/mini-interpreter/object"

	"github.com/songzhibin97/mini-compiler/vm"

	"github.com/songzhibin97/mini-compiler/compiler"
	"github.com/songzhibin97/mini-interpreter/lexer"
	"github.com/songzhibin97/mini-interpreter/parser"
)

const PROMPT = ">>>"

func Start(in io.Reader, out io.Writer) {
	fmt.Println("Welcome to Mini-compiler")
	scanner := bufio.NewScanner(in)
	symbolTable := compiler.NewSymbolTable()
	for i, s := range compiler.IterBuiltin() {
		symbolTable.DefineBuiltin(i, s)
	}
	globals := make([]object.Object, vm.GlobalsSize)
	constants := []object.Object{}
	for {
		_, _ = fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		p := parser.NewParser(lexer.NewLexer(line))
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, s := range p.Errors() {
				_, _ = io.WriteString(out, "\t"+s+"\r\n")
			}
		}
		comp := compiler.NewCompilerWithSymbol(symbolTable, constants)
		err := comp.Compiler(program)
		if err != nil {
			_, _ = io.WriteString(out, "\t Compilation failed:"+err.Error()+"\r\n")
			continue
		}
		constants = comp.Bytecode().Constants
		v := vm.NewVMWithGlobals(comp.Bytecode(), globals)
		err = v.Run()
		if err != nil {
			_, _ = io.WriteString(out, "\t VM failed:"+err.Error()+"\r\n")
			continue
		}
		lastPopped := v.LastPoppedStackElem()
		_, _ = io.WriteString(out, lastPopped.Inspect())
		_, _ = io.WriteString(out, "\n")
	}
}
