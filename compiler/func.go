package compiler

import (
	"fmt"

	"github.com/songzhibin97/mini-interpreter/object"

	"github.com/songzhibin97/mini-compiler/code"
)

type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() object.Type { return "COMPILED_FUNCTION" }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

type Closure struct {
	Fn  *CompiledFunction
	Ctx []object.Object // 上下文环境
}

func (c *Closure) Type() object.Type { return "CLOSURE" }
func (c *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", c)
}

type Builtin struct {
	Fn   *object.Builtin
	Name string
}

var builtins = []*Builtin{
	{
		Fn: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Error{Error: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}
			switch arg := args[0].(type) {
			case *object.Map:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.Stringer:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return &object.Error{Error: fmt.Sprintf("argument to `len` not supported, got %s", args[0].Type())}
			}
		}},
		Name: "len",
	},
	{
		Fn: &object.Builtin{Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return &object.Nil{}
		}},
		Name: "print",
	},
}

func IterBuiltin() []string {
	res := []string{}
	for _, builtin := range builtins {
		res = append(res, builtin.Name)
	}
	return res
}

func GetBuiltinByIndex(idx int) *object.Builtin {
	if idx < 0 || idx > len(builtins) {
		return nil
	}
	return builtins[idx].Fn
}
