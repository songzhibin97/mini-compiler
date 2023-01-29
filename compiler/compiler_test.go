package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/songzhibin97/mini-compiler/code"
	"github.com/songzhibin97/mini-interpreter/ast"
	"github.com/songzhibin97/mini-interpreter/lexer"
	"github.com/songzhibin97/mini-interpreter/object"
	"github.com/songzhibin97/mini-interpreter/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func parse(input string) *ast.Program {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	return p.ParseProgram()
}

func mergeInstruction(s []code.Instructions) code.Instructions {
	res := code.Instructions{}
	for _, instructions := range s {
		res = append(res, instructions...)
	}
	return res
}

func testInstructions(t *testing.T, expected []code.Instructions, actual code.Instructions) {
	instructions := mergeInstruction(expected)

	assert.Equal(t, len(actual), len(instructions))
	assert.Equal(t, actual, instructions)
}

func testConstants(t *testing.T, expected []interface{}, actual []object.Object) {
	assert.Equal(t, len(expected), len(actual))
	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			testIntegerObject(t, int64(constant), actual[i])
		case string:
			testStringObject(t, constant, actual[i])
		case []code.Instructions:
			fn, ok := actual[i].(*CompiledFunction)
			assert.Equal(t, ok, true)
			testInstructions(t, constant, fn.Instructions)
		}
	}
}

func testIntegerObject(t *testing.T, expected int64, actual object.Object) {
	result, ok := actual.(*object.Integer)
	assert.Equal(t, ok, true)
	assert.Equal(t, result.Value, expected)
}

func testStringObject(t *testing.T, expected string, actual object.Object) {
	result, ok := actual.(*object.Stringer)
	assert.Equal(t, ok, true)
	assert.Equal(t, result.Value, expected)
}

func TestCompilationScope(t *testing.T) {
	compiler := NewCompiler()
	assert.Equal(t, compiler.scopeIndex, 0)

	globalSymbolTable := compiler.symbolTable

	compiler.emit(code.OpAdd)
	compiler.enterScope()
	assert.Equal(t, compiler.scopeIndex, 1)

	compiler.emit(code.OpSub)
	assert.Equal(t, len(compiler.scopes[compiler.scopeIndex].instructions), 1)
	assert.Equal(t, compiler.scopes[compiler.scopeIndex].lastInstruction.OpCode, code.OpSub)

	assert.Equal(t, compiler.symbolTable.External, globalSymbolTable)

	compiler.leaveScope()
	assert.Equal(t, compiler.scopeIndex, 0)

	assert.Equal(t, compiler.symbolTable, globalSymbolTable)
	assert.Equal(t, compiler.symbolTable.External == nil, true)

	compiler.emit(code.OpMul)
	assert.Equal(t, len(compiler.scopes[compiler.scopeIndex].instructions), 2)
	assert.Equal(t, compiler.scopes[compiler.scopeIndex].lastInstruction.OpCode, code.OpMul)
	assert.Equal(t, compiler.scopes[compiler.scopeIndex].preInstruction.OpCode, code.OpAdd)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1+2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1-2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1*2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "10/5",
			expectedConstants: []interface{}{10, 5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpQuo),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestBooleanExpr(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGTR),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGTR),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEQL),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNEQ),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEQL),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNEQ),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestIfExpr(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if (true) {1} 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue), // 1
				// 0001
				code.Make(code.OpJumpConditionNotTrue, 10), // 3
				// 0004
				code.Make(code.OpConstant, 0), // 3
				// 0007
				code.Make(code.OpJump, 11), // 3
				// 0010
				code.Make(code.OpNil), // 1
				// 0011
				code.Make(code.OpPop), // 1
				// 0012
				code.Make(code.OpConstant, 1), // 3
				// 0015
				code.Make(code.OpPop), // 1
			},
		},
		{
			input:             "if (true) {1} else {2} 3",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue), // 1
				// 0001
				code.Make(code.OpJumpConditionNotTrue, 10), // 3
				// 0004
				code.Make(code.OpConstant, 0), // 3
				// 0007
				code.Make(code.OpJump, 13), // 3
				// 0010
				code.Make(code.OpConstant, 1), // 3
				// 0013
				code.Make(code.OpPop), // 1
				// 0014
				code.Make(code.OpConstant, 2), // 3
				// 0017
				code.Make(code.OpPop), // 1
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestIndexExpr(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1,2,3][0]",
			expectedConstants: []interface{}{1, 2, 3, 0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1,2,3][1+1]",
			expectedConstants: []interface{}{1, 2, 3, 1, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1:2,3:4}[1]",
			expectedConstants: []interface{}{1, 2, 3, 4, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpMap, 4),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1:2,3:4}[2-1]",
			expectedConstants: []interface{}{1, 2, 3, 4, 2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpMap, 4),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpSub),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestVarStmt(t *testing.T) {
	tests := []compilerTestCase{
		{
			`var one = 1`,
			[]interface{}{1},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			`var one = 1
			var two = 2`,
			[]interface{}{1, 2},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			`var one = 1
			one`,
			[]interface{}{1},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			`var one = 1
			var two = one
			two `,
			[]interface{}{1},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestVarStmtScope(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `var a = 1
			func test(){a}`,
			expectedConstants: []interface{}{
				1,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `func test() {
				var a = 1
				a
			}`,
			expectedConstants: []interface{}{
				1,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `func test() {
				var a = 1
				var b = 2
				a + b
			}`,
			expectedConstants: []interface{}{
				1,
				2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFuncExpr(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "func test(){}",
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "func test (){ return 1 + 2}",
			expectedConstants: []interface{}{
				1, 2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "func test (){ 1 + 2 }",
			expectedConstants: []interface{}{
				1, 2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "func test (){ 1 2 }",
			expectedConstants: []interface{}{
				1, 2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCallFuncExpr(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "func test(){1} test()",
			expectedConstants: []interface{}{
				1,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "func test(a){a} test(1)",
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
				1,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "func test(a,b,c){a+b+c} test(1,2,3)",
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				1,
				2,
				3,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBuiltins(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `len([])`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 0),
				code.Make(code.OpArray, 0),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `len([1,2,3])`,
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)

}
func TestStringer(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"mini-compiler"`,
			expectedConstants: []interface{}{"mini-compiler"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"mini" + "-" + "compiler"`,
			expectedConstants: []interface{}{"mini", "-", "compiler"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestArray(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1+2, 3-4, 5*6]",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestMap(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpMap, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1:2, 3:4, 5:6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMap, 6),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1:2+3, 4:5*6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpMap, 4),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `func test1(a) {func test2(b) { a + b} }`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpContext, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `func test1(a) {func test2(b) { func test3(c) {return a + b + c}}}`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpContext, 0),
					code.Make(code.OpContext, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpContext, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 2),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 1, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			var global = 1
			func test(){
				var a = 2
				func test1() {
					var b = 3
					func test2() {
						var c = 4
						return global + a + b + c
					}
				}
			}
			`,
			expectedConstants: []interface{}{
				1, 2, 3, 4,
				[]code.Instructions{
					code.Make(code.OpConstant, 3),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpContext, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpContext, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 2),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpContext, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 4, 2),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 5, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 6, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, test := range tests {
		program := parse(test.input)

		compiler := NewCompiler()
		err := compiler.Compiler(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		bytecode := compiler.Bytecode()
		testInstructions(t, test.expectedInstructions, bytecode.Instructions)
		testConstants(t, test.expectedConstants, bytecode.Constants)
	}
}
