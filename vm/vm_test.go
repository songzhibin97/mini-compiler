package vm

import (
	"testing"

	"github.com/songzhibin97/mini-interpreter/object"

	"github.com/songzhibin97/mini-compiler/compiler"
	"github.com/stretchr/testify/assert"

	"github.com/songzhibin97/mini-interpreter/ast"
	"github.com/songzhibin97/mini-interpreter/lexer"
	"github.com/songzhibin97/mini-interpreter/parser"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func parse(input string) *ast.Program {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	return p.ParseProgram()
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		testIntegerObject(t, int64(expected), actual)
	case bool:
		testBooleanObject(t, expected, actual)
	case string:
		testStringObject(t, expected, actual)
	case []int:
		testArrayObject(t, expected, actual)
	case map[object.MapKey]int64:
		testMapObject(t, expected, actual)
	case *object.Nil:
		assert.Equal(t, actual, Nil)
	}
}

func testIntegerObject(t *testing.T, expected int64, actual object.Object) {
	result, ok := actual.(*object.Integer)
	assert.Equal(t, ok, true)
	assert.Equal(t, result.Value, expected)
}

func testBooleanObject(t *testing.T, expected bool, actual object.Object) {
	result, ok := actual.(*object.Boolean)
	assert.Equal(t, ok, true)
	assert.Equal(t, result.Value, expected)
}

func testStringObject(t *testing.T, expected string, actual object.Object) {
	result, ok := actual.(*object.Stringer)
	assert.Equal(t, ok, true)
	assert.Equal(t, result.Value, expected)
}

func testArrayObject(t *testing.T, expected []int, actual object.Object) {
	result, ok := actual.(*object.Array)
	assert.Equal(t, ok, true)
	assert.Equal(t, len(result.Elements), len(expected))
	for i, elem := range expected {
		testIntegerObject(t, int64(elem), result.Elements[i])
	}
}

func testMapObject(t *testing.T, expected map[object.MapKey]int64, actual object.Object) {
	result, ok := actual.(*object.Map)
	assert.Equal(t, ok, true)
	assert.Equal(t, len(result.Elements), len(expected))
	for k, v := range expected {
		v2, ok := result.Elements[k]
		assert.Equal(t, ok, true)
		assert.Equal(t, v, v2)
	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "1",
			expected: 1,
		},
		{
			input:    "2",
			expected: 2,
		},
		{
			input:    "1+2",
			expected: 3,
		},
		{
			input:    "1-2",
			expected: -1,
		},
		{
			input:    "2*2",
			expected: 4,
		},
		{
			input:    "10/2",
			expected: 5,
		},
		{
			input:    "1 + 2 + 3 + 4 - 5",
			expected: 5,
		},
		{
			input:    "(1 * 2 + 3 - 4) * 10 / 2",
			expected: 5,
		},
		{
			input:    "1 + 2 * (3 + 4)",
			expected: 15,
		},
		{
			input:    "-1",
			expected: -1,
		},
		{
			input:    "-2",
			expected: -2,
		},
		{
			input:    "-10+100-10",
			expected: 80,
		},
		{
			input:    "(-5 + 10) * 2 + -10",
			expected: 0,
		},
	}
	runVmTests(t, tests)
}

func TestBooleanExpr(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "true",
			expected: true,
		},
		{
			input:    "false",
			expected: false,
		},
		{
			input:    "1 < 2",
			expected: true,
		},
		{
			input:    "1 > 2",
			expected: false,
		},
		{
			input:    "1 < 1",
			expected: false,
		},
		{
			input:    "1 > 1",
			expected: false,
		},
		{
			input:    "1 == 1",
			expected: true,
		},
		{
			input:    "1 != 1",
			expected: false,
		},
		{
			input:    "1 == 2",
			expected: false,
		},
		{
			input:    "1 != 2",
			expected: true,
		},
		{
			input:    "true == true",
			expected: true,
		},
		{
			input:    "false == false",
			expected: true,
		},
		{
			input:    "true == false",
			expected: false,
		},
		{
			input:    "true != false",
			expected: true,
		},
		{
			input:    "false != true",
			expected: true,
		},
		{
			input:    "(1 < 2) == true",
			expected: true,
		},
		{
			input:    "(1 < 2) == false",
			expected: false,
		},
		{
			input:    "(1 > 2) == true",
			expected: false,
		},
		{
			input:    "(1 > 2) == false",
			expected: true,
		},
		{
			input:    "!true",
			expected: false,
		},
		{
			input:    "!false",
			expected: true,
		},
		{
			input:    "!1",
			expected: false,
		},
		{
			input:    "!!1",
			expected: true,
		},
		{
			input:    "!!true",
			expected: true,
		},
		{
			input:    "!!false",
			expected: false,
		},
		{
			input:    "!(if (false) { 5 })",
			expected: true,
		},
	}
	runVmTests(t, tests)
}

func TestIfExpr(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "if (true) {1}",
			expected: 1,
		},
		{
			input:    "if (true) {1} else {2}",
			expected: 1,
		},
		{
			input:    "if (false) {1} else {2}",
			expected: 2,
		},
		{
			input:    "if (1) {1} else {2}",
			expected: 1,
		},
		{
			input:    "if (1) {1}",
			expected: 1,
		},
		{
			input:    "if (1 > 2) {1} else {2}",
			expected: 2,
		},
		{
			input:    "if (1 < 2) {1} else {2}",
			expected: 1,
		},
		{
			input:    "if (1 > 2) { 1 }",
			expected: Nil,
		},
		{
			input:    "if (false) { 1 }",
			expected: Nil,
		},
		{
			input:    "if ((if (false) { 1 })) { 1 } else { 2 }",
			expected: 2,
		},
	}
	runVmTests(t, tests)
}

func TestIndexExpr(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "[][0]",
			expected: Nil,
		},
		{
			input:    "[1,2,3][0]",
			expected: 1,
		},
		{
			input:    "[1,2,3][1]",
			expected: 2,
		},
		{
			input:    "[1,2,3][2]",
			expected: 3,
		},
		{
			input:    "[1,2,3][4]",
			expected: Nil,
		},
		{
			input:    "[1,2,3][-1]",
			expected: Nil,
		},
		{
			input:    "{}[0]",
			expected: Nil,
		},
		{
			input:    "{1:2,3:4}[0]",
			expected: Nil,
		},
		{
			input:    "{1:2,3:4}[1]",
			expected: 2,
		},
		{
			input:    "{1:2,3:4}[3]",
			expected: 4,
		},
	}

	runVmTests(t, tests)
}

func TestVarStmt(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "var one = 1 one",
			expected: 1,
		},
		{
			input:    "var one = 1 var two = 2 one + two",
			expected: 3,
		},
		{
			input:    "var one = 1 var two = one + one  one + two",
			expected: 3,
		},
	}

	runVmTests(t, tests)
}

func TestCallFuncExpr(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "func test(){} test()",
			expected: Nil,
		},
		{
			input:    "func test(){1+2} test()",
			expected: 3,
		},
		{
			input: `func one(){1}
			func two(){2}
			func three(){one()+two()}
			three()
			`,
			expected: 3,
		},
		{
			input: `
			var a = 1
			func test() {
				var b = 2
			return a + b
			}
			test()
			`,
			expected: 3,
		},
		{
			input: `
			func test() {
				var a = 1
				var b = 2
				return a + b
			}
			test()
			`,
			expected: 3,
		},
		{
			input: `
			func a() {
				var a = 1
				return a
			}
			func b() {
				var b = 2
				return b
			}
			func c(){
				return a() + b() + 3
			}
			c()
			`,
			expected: 6,
		},
		{
			input: `
			func test(a) {a}
			test(2)
			`,
			expected: 2,
		},
		{
			input: `
			func test(a,b) {a+b}
			test(1,2)
			`,
			expected: 3,
		},
		{
			input: `
			func test(a,b) {
				var c = a+b
				return c
			}
			test(1,2)
			`,
			expected: 3,
		},
		{
			input: `
			var g = 10
			func test(a,b) {
				var c = a+b
				return c + g
			}
			test(1,2)
			`,
			expected: 13,
		},
	}

	runVmTests(t, tests)
}

func TestBuiltins(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "len([])",
			expected: 0,
		},
		{
			input:    "len([1,2,3])",
			expected: 3,
		},
		{
			input:    "len({})",
			expected: 0,
		},
		{
			input:    "len({1:1,2:2,3:3})",
			expected: 3,
		},
		{
			input:    `len("")`,
			expected: 0,
		},
		{
			input:    `len("123")`,
			expected: 3,
		},
	}

	runVmTests(t, tests)
}

func TestStringer(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `"mini-compiler"`,
			expected: "mini-compiler",
		},
		{
			input:    `"mini" + "-" + "compiler"`,
			expected: "mini-compiler",
		},
	}

	runVmTests(t, tests)
}

func TestArray(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "[]",
			expected: []int{},
		},
		{
			input:    "[1,2,3]",
			expected: []int{1, 2, 3},
		},
		{
			input:    "[1+2, 3-4, 5*6]",
			expected: []int{3, -1, 30},
		},
	}

	runVmTests(t, tests)
}

func TestMap(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "{}",
			expected: map[object.HashAble]int64{},
		},
		{
			input: "{1:2, 3:4, 5:6}",
			expected: map[object.HashAble]int64{
				&object.Integer{Value: 1}: 2,
				&object.Integer{Value: 3}: 4,
				&object.Integer{Value: 5}: 6,
			},
		},
		{
			input: "{1:2+3, 4:5*6}",
			expected: map[object.HashAble]int64{
				&object.Integer{Value: 1}: 5,
				&object.Integer{Value: 4}: 30,
			},
		},
	}

	runVmTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `func test1(a) {func test2(b) { a + b} return test2 } test1(1)(2)`,
			expected: 3,
		},
		{
			input:    `func test1(a) {func test2(b) { func test3(c) {return a + b + c} return test3} return test2} test1(1)(2)(3)`,
			expected: 6,
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
					}()
				}()
			}
			test()
			`,

			expected: 10,
		},
		{
			input: `
			func test (a) {
				if (a == 0 ) {
					return 0
				} else {
					return test(a-1)
				}
			}
			test(10)
			`,

			expected: 0,
		},
		{
			input: `
			func test () {
				func test2(a) {
					if (a == 0) {
						return 0
					} else {
						return test2(a-1)
					}
				}
				test2(1)
			}
			test()
			`,

			expected: 0,
		},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, test := range tests {
		program := parse(test.input)

		comp := compiler.NewCompiler()
		err := comp.Compiler(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		vm := NewVM(comp.Bytecode())
		assert.NoError(t, vm.Run())
		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, test.expected, stackElem)
	}
}
