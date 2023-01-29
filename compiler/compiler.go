package compiler

import (
	"fmt"
	"math"
	"sort"

	"github.com/songzhibin97/mini-compiler/code"
	"github.com/songzhibin97/mini-interpreter/ast"
	"github.com/songzhibin97/mini-interpreter/object"
)

const fakeAddress = math.MaxInt64

type (
	EmittedInstruction struct {
		OpCode code.Opcode // 发送的指令
		Pos    int         // 位置
	}

	CompilationScope struct {
		instructions code.Instructions // 指令

		lastInstruction EmittedInstruction
		preInstruction  EmittedInstruction
	}

	Compiler struct {
		constants []object.Object

		symbolTable *SymbolTable

		scopes     []CompilationScope
		scopeIndex int
	}

	Bytecode struct {
		Instructions code.Instructions // 指令
		Constants    []object.Object
	}
)

func init() {
	defaultCompiler = func(c *Compiler, node ast.Node) error {
		switch node := node.(type) {
		case *ast.Program:
			for _, stmt := range node.Stmts {
				err := c.Compiler(stmt, defaultCompiler)
				if err != nil {
					return err
				}
			}

		case *ast.Identifier:
			symbol, ok := c.symbolTable.GetDefine(node.Value)
			if !ok {
				return fmt.Errorf("undefined variable %s", node.Value)
			}

			c.loadSymbol(symbol)

		case *ast.String:
			c.emit(code.OpConstant, c.addConstant(&object.Stringer{Value: node.Value}))

		case *ast.Array:
			for _, element := range node.Elements {
				err := c.Compiler(element)
				if err != nil {
					return err
				}
			}
			c.emit(code.OpArray, len(node.Elements))

		case *ast.Map:
			// 单元测试有序性
			{
				keys := make([]ast.Expr, 0, len(node.Elements))
				for key := range node.Elements {
					keys = append(keys, key)
				}
				sort.Slice(keys, func(i, j int) bool {
					return keys[i].String() < keys[j].String()
				})
				for _, key := range keys {
					err := c.Compiler(key)
					if err != nil {
						return err
					}
					err = c.Compiler(node.Elements[key])
					if err != nil {
						return err
					}
				}

			}
			// 其他模式
			{

				//for k, v := range node.Elements {
				//	err := c.Compiler(k)
				//	if err != nil {
				//		return err
				//	}
				//	err = c.Compiler(v)
				//	if err != nil {
				//		return err
				//	}
				//}
			}

			c.emit(code.OpMap, len(node.Elements)*2)

		case *ast.Boolean:
			if node.Value {
				c.emit(code.OpTrue)
			} else {
				c.emit(code.OpFalse)
			}

		case *ast.VarStmt:
			symbol := c.symbolTable.Define(node.Name.Value)
			err := c.Compiler(node.Value)
			if err != nil {
				return err
			}
			if symbol.Scope == GlobalScope {
				c.emit(code.OpSetGlobal, symbol.Index)
			} else {
				c.emit(code.OpSetLocal, symbol.Index)
			}

		case *ast.BlockStmt:
			for _, stmt := range node.Stmts {
				err := c.Compiler(stmt)
				if err != nil {
					return err
				}
			}

		case *ast.ExprStmt:
			err := c.Compiler(node.Expr)
			if err != nil {
				return err
			}
			c.emit(code.OpPop)

		case *ast.IfExpr:
			err := c.Compiler(node.Condition)
			if err != nil {
				return err
			}
			// 插入假的地址
			jumpNotTrue := c.emit(code.OpJumpConditionNotTrue, fakeAddress)
			err = c.Compiler(node.Consequence)
			if err != nil {
				return err
			}

			// 判断最后一条指令是否是pop,如果是则进行删除
			if c.lastInstructionIs(code.OpPop) {
				// 进行删除, 把last指向pre last用于修改fake address
				c.removeLastPop()
			}
			// 在插入一条指令else指令
			jumpPos := c.emit(code.OpJump, fakeAddress)

			// 将jumpNotTrue fake 替换为else 位置
			curPos := len(c.curInstructions())
			c.changeOperand(jumpNotTrue, curPos)

			if node.Alternative == nil {
				c.emit(code.OpNil)
			} else {
				err = c.Compiler(node.Alternative)
				if err != nil {
					return err
				}

				if c.lastInstructionIs(code.OpPop) {
					c.removeLastPop()
				}
			}

			// 将 jump替换到if语句结束
			curPos = len(c.curInstructions())
			c.changeOperand(jumpPos, curPos)

		case *ast.ReturnStmt:
			err := c.Compiler(node.Value)
			if err != nil {
				return err
			}
			c.emit(code.OpReturnValue)

		case *ast.FuncExpr:
			// 进入新的作用域 函数作用域

			// 将函数放到符号表
			symbol := c.symbolTable.Define(node.Name.Value)

			c.enterScope()

			c.symbolTable.defineFunction(node.Name.Value)

			for _, p := range node.Params {
				c.symbolTable.Define(p.Value)
			}

			err := c.Compiler(node.Body)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.replaceLastPopWithReturn() // 删除最后一条pop指令,并且替换为return value
			}

			// 如果最后一条指令不是return value 插一条return指令进去
			if !c.lastInstructionIs(code.OpReturnValue) {
				c.emit(code.OpReturn)
			}

			ctx := c.symbolTable.Context
			numLocals := c.symbolTable.count
			instructions := c.leaveScope()
			for _, symbol := range ctx {
				c.loadSymbol(symbol)
			}

			compiledFn := &CompiledFunction{
				Instructions:  instructions,
				NumLocals:     numLocals,
				NumParameters: len(node.Params),
			}

			c.emit(code.OpClosure, c.addConstant(compiledFn), len(ctx))

			if symbol.Scope == GlobalScope {
				c.emit(code.OpSetGlobal, symbol.Index)
			} else {
				c.emit(code.OpSetLocal, symbol.Index)
			}

		case *ast.CallExpr:
			err := c.Compiler(node.Func)
			if err != nil {
				return err
			}

			for _, arg := range node.Args {
				err = c.Compiler(arg)
				if err != nil {
					return err
				}
			}
			c.emit(code.OpCall, len(node.Args))

		case *ast.Integer:
			c.emit(code.OpConstant, c.addConstant(&object.Integer{
				Value: node.Value,
			}))

		case *ast.InfixExpr:
			if node.Operator == "<" {
				// 指令重排
				err := c.Compiler(node.Right)
				if err != nil {
					return err
				}
				err = c.Compiler(node.Left)
				if err != nil {
					return err
				}
				c.emit(code.OpGTR)
				return nil
			}

			err := c.Compiler(node.Left)
			if err != nil {
				return err
			}
			err = c.Compiler(node.Right)
			if err != nil {
				return err
			}
			switch node.Operator {
			case "+":
				c.emit(code.OpAdd)
			case "-":
				c.emit(code.OpSub)
			case "*":
				c.emit(code.OpMul)
			case "/":
				c.emit(code.OpQuo)
			case ">":
				c.emit(code.OpGTR)
			case "==":
				c.emit(code.OpEQL)
			case "!=":
				c.emit(code.OpNEQ)

			default:
				return fmt.Errorf("unknown operator %s", node.Operator)
			}

		case *ast.PrefixExpr:
			err := c.Compiler(node.Right)
			if err != nil {
				return err
			}
			switch node.Operator {
			case "-":
				c.emit(code.OpMinus)
			case "!":
				c.emit(code.OpBang)
			default:
				return fmt.Errorf("unknown operator %s", node.Operator)
			}

		case *ast.IndexExpr:
			err := c.Compiler(node.Left)
			if err != nil {
				return err
			}
			err = c.Compiler(node.Index)
			if err != nil {
				return err
			}
			c.emit(code.OpIndex)
		}

		return nil
	}
}

var defaultCompiler func(c *Compiler, node ast.Node) error

func (c *Compiler) Compiler(node ast.Node, handler ...func(c *Compiler, node ast.Node) error) error {
	handler = append(handler, defaultCompiler)
	return handler[0](c, node)
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.curInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) curInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.curInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.curInstructions(), ins...)
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	c.scopes[c.scopeIndex].preInstruction = c.scopes[c.scopeIndex].lastInstruction
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{
		OpCode: op,
		Pos:    pos,
	}
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.curInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.OpCode == op
}

func (c *Compiler) removeLastPop() {
	c.scopes[c.scopeIndex].instructions = c.scopes[c.scopeIndex].instructions[:c.scopes[c.scopeIndex].lastInstruction.Pos]
	c.scopes[c.scopeIndex].lastInstruction = c.scopes[c.scopeIndex].preInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.scopes[c.scopeIndex].instructions[i+pos] = newInstruction[i]
	}
}

func (c *Compiler) replaceLastPopWithReturn() {
	pos := c.scopes[c.scopeIndex].lastInstruction.Pos
	c.replaceInstruction(pos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.OpCode = code.OpReturnValue
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.scopes[c.scopeIndex].instructions[opPos])
	c.replaceInstruction(opPos, code.Make(op, operand))
}

func (c *Compiler) leaveScope() code.Instructions {
	ins := c.curInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.External
	return ins
}

func (c *Compiler) enterScope() {
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
	c.scopes = append(c.scopes, CompilationScope{
		instructions:    code.Instructions{},
		lastInstruction: EmittedInstruction{},
		preInstruction:  EmittedInstruction{},
	})
	c.scopeIndex++
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case ContextScope:
		c.emit(code.OpContext, s.Index)
	case FunctionScope:
		c.emit(code.OpCurrClosure)
	}
}

func NewCompiler() *Compiler {
	symbolTable := NewSymbolTable()
	for idx, builtin := range builtins {
		symbolTable.DefineBuiltin(idx, builtin.Name)
	}
	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes: []CompilationScope{
			{
				instructions:    code.Instructions{},
				lastInstruction: EmittedInstruction{},
				preInstruction:  EmittedInstruction{},
			},
		},
	}
}

func NewCompilerWithSymbol(symbolTable *SymbolTable, constants []object.Object) *Compiler {
	return &Compiler{
		constants:   constants,
		symbolTable: symbolTable,
		scopes: []CompilationScope{
			{
				instructions:    code.Instructions{},
				lastInstruction: EmittedInstruction{},
				preInstruction:  EmittedInstruction{},
			},
		},
	}
}
