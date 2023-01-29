package vm

import (
	"errors"
	"fmt"

	"github.com/songzhibin97/mini-compiler/code"
	"github.com/songzhibin97/mini-compiler/compiler"
	"github.com/songzhibin97/mini-interpreter/object"
)

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Nil = &object.Nil{}

const (
	StackSize   = 1 << 11
	GlobalsSize = 1 << 16
	FrameSize   = 1 << 11
)

var ErrStackOverflow = errors.New("stack overflow")

type VM struct {
	constants []object.Object

	stack []object.Object
	sp    int // Stack Pointer 始终指向下一个位置, 栈顶应该是 stack[sp-1]

	globals []object.Object

	frames      []*Frame
	framesIndex int
}

func (v *VM) curFrame() *Frame {
	return v.frames[v.framesIndex-1]
}

func (v *VM) pushFrame(frame *Frame) {
	v.frames[v.framesIndex] = frame
	v.framesIndex++
}

func (v *VM) popFrame() *Frame {
	v.framesIndex--
	return v.frames[v.framesIndex]
}

func (v *VM) push(o object.Object) error {
	if v.sp >= StackSize {
		return ErrStackOverflow
	}
	v.stack[v.sp] = o
	v.sp++
	return nil
}

func (v *VM) pop() object.Object {
	if v.sp <= 0 {
		return nil
	}
	v.sp--
	return v.stack[v.sp]
}

func (v *VM) executeArithmeticOperation(op code.Opcode) error {
	right := v.pop()
	left := v.pop()
	if left.Type() == object.INT && right.Type() == object.INT {
		return v.executeArithmeticIntegerOperation(op, left, right)
	}
	if left.Type() == object.String && right.Type() == object.String {
		return v.executeArithmeticStringOperation(op, left, right)
	}
	return fmt.Errorf("unsupported types for operation %s %s", left.Type(), right.Type())
}

func (v *VM) executeArithmeticIntegerOperation(op code.Opcode, left, right object.Object) error {
	lv, rv := left.(*object.Integer).Value, right.(*object.Integer).Value
	var result int64
	switch op {
	case code.OpAdd:
		result = lv + rv
	case code.OpSub:
		result = lv - rv
	case code.OpMul:
		result = lv * rv
	case code.OpQuo:
		result = lv / rv
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return v.push(&object.Integer{Value: result})
}

func (v *VM) executeArithmeticStringOperation(op code.Opcode, left, right object.Object) error {
	lv, rv := left.(*object.Stringer).Value, right.(*object.Stringer).Value
	var result string
	switch op {
	case code.OpAdd:
		result = lv + rv
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return v.push(&object.Stringer{Value: result})
}

func (v *VM) executeIndexOperation(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INT:
		return v.executeArrayIndex(left, index)
	case left.Type() == object.MAP:
		return v.executeMapIndex(left, index)
	default:
		return fmt.Errorf("unsupported types for operation %s", left.Type())
	}
}

func (v *VM) executeArrayIndex(array, index object.Object) error {
	arr := array.(*object.Array)
	idx := index.(*object.Integer).Value

	if idx < 0 || idx > int64(len(arr.Elements)-1) {
		return v.push(Nil)
	}
	return v.push(arr.Elements[idx])
}

func (v *VM) executeMapIndex(mp, index object.Object) error {
	hashMap := mp.(*object.Map)
	idx, ok := index.(object.HashAble)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}
	val, ok := hashMap.Elements[idx.MapKey()]
	if !ok {
		return v.push(Nil)
	} else {
		return v.push(val.Value)
	}
}

func (v *VM) executeComparisonOperation(op code.Opcode) error {
	right := v.pop()
	left := v.pop()
	if left.Type() == object.INT && right.Type() == object.INT {
		return v.executeComparisonIntegerOperation(op, left, right)
	}
	switch op {
	case code.OpEQL:
		return v.push(translationBooleanObject(left == right))
	case code.OpNEQ:
		return v.push(translationBooleanObject(left != right))
	default:
		return fmt.Errorf("unknown operator %s %s", left.Type(), right.Type())
	}
}

func (v *VM) executeComparisonIntegerOperation(op code.Opcode, left, right object.Object) error {
	lv, rv := left.(*object.Integer).Value, right.(*object.Integer).Value
	switch op {
	case code.OpEQL:
		return v.push(translationBooleanObject(lv == rv))
	case code.OpNEQ:
		return v.push(translationBooleanObject(lv != rv))
	case code.OpGTR:
		return v.push(translationBooleanObject(lv > rv))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (v *VM) executeBangOperation() error {
	op := v.pop()
	switch op {
	case True:
		return v.push(False)
	case False:
		return v.push(True)
	case Nil:
		return v.push(True)
	default:
		return v.push(False)
	}
}

func (v *VM) executeMinusOperation() error {
	op := v.pop()
	if op.Type() != object.INT {
		return fmt.Errorf("unsupported type for minus %s", op.Type())
	}
	return v.push(&object.Integer{Value: -op.(*object.Integer).Value})
}

func (v *VM) isTrue(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Nil:
		return false
	default:
		return true
	}
}

func (v *VM) executeCall(args int) error {
	call := v.stack[v.sp-1-args]
	switch caller := call.(type) {
	case *compiler.Closure:
		return v.callClosure(caller, args)
	case *object.Builtin:
		return v.callBuiltin(caller, args)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (v *VM) callClosure(cl *compiler.Closure, args int) error {
	if args != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			cl.Fn.NumParameters, args)
	}
	frame := NewFrame(cl, v.sp-args)
	v.pushFrame(frame)

	v.sp = frame.basePointer + cl.Fn.NumLocals

	return nil
}

func (v *VM) callBuiltin(fn *object.Builtin, numArgs int) error {
	args := v.stack[v.sp-numArgs : v.sp]
	result := fn.Fn(args...)
	v.sp = v.sp - numArgs - 1

	if result != nil {
		return v.push(result)
	} else {
		return v.push(Nil)
	}
}

func (v *VM) newArray(start, end int) object.Object {
	elems := make([]object.Object, end-start)
	for i := start; i < end; i++ {
		elems[i-start] = v.stack[i]
	}
	return &object.Array{Elements: elems}
}

func (v *VM) newMap(start, end int) (object.Object, error) {
	mp := make(map[object.MapKey]object.HashValue, end-start)
	for i := start; i < end; i += 2 {
		key := v.stack[i]
		val := v.stack[i+1]
		hashAble, ok := key.(object.HashAble)
		if !ok {
			return Nil, fmt.Errorf("unable to hash key:%s", key.Type())
		}
		mp[hashAble.MapKey()] = object.HashValue{
			Key:   key,
			Value: val,
		}
	}
	return &object.Map{Elements: mp}, nil
}

func (v *VM) pushClosure(idx int, countCtx int) error {
	constant := v.constants[idx]
	fn, ok := constant.(*compiler.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}
	ctxs := make([]object.Object, countCtx)
	for i := 0; i < countCtx; i++ {
		ctxs[i] = v.stack[v.sp-countCtx+i]
	}
	v.sp -= countCtx
	closure := &compiler.Closure{Fn: fn, Ctx: ctxs}
	return v.push(closure)
}

func (v *VM) LastPoppedStackElem() object.Object {
	return v.stack[v.sp]
}

// Top 获取栈顶元素
func (v *VM) Top() object.Object {
	if v.sp == 0 {
		return nil
	}
	return v.stack[v.sp-1]
}

func defaultVmHandler(v *VM) error {
	var (
		instructions code.Instructions
		op           code.Opcode
	)
	for v.curFrame().ip < len(v.curFrame().Instructions())-1 {
		v.curFrame().ip++

		instructions = v.curFrame().Instructions()
		op = code.Opcode(instructions[v.curFrame().ip]) // 获取指令

		// 处理指令
		switch op {

		case code.OpConstant: // 常数
			index := code.ReadUint16(instructions[v.curFrame().ip+1:])
			v.curFrame().ip += 2

			err := v.push(v.constants[index])
			if err != nil {
				return err
			}

		case code.OpArray:
			ln := int(code.ReadUint16(instructions[v.curFrame().ip+1:]))
			v.curFrame().ip += 2

			array := v.newArray(v.sp-ln, v.sp)
			v.sp -= ln

			err := v.push(array)

			if err != nil {
				return err
			}

		case code.OpMap:
			ln := int(code.ReadUint16(instructions[v.curFrame().ip+1:]))
			v.curFrame().ip += 2

			mp, err := v.newMap(v.sp-ln, v.sp)
			if err != nil {
				return err
			}
			v.sp -= ln

			err = v.push(mp)
			if err != nil {
				return err
			}

		case code.OpIndex:
			index := v.pop()
			left := v.pop()
			err := v.executeIndexOperation(left, index)
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpQuo:
			err := v.executeArithmeticOperation(op)
			if err != nil {
				return err
			}

		case code.OpTrue:
			err := v.push(True)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := v.push(False)
			if err != nil {
				return err
			}

		case code.OpCall:
			args := code.ReadUint8(instructions[v.curFrame().ip+1:])
			v.curFrame().ip += 1

			err := v.executeCall(int(args))
			if err != nil {
				return err
			}

		case code.OpReturnValue:
			val := v.pop()
			frame := v.popFrame()
			v.sp = frame.basePointer - 1
			err := v.push(val)
			if err != nil {
				return err
			}

		case code.OpReturn:
			frame := v.popFrame()
			v.sp = frame.basePointer - 1
			err := v.push(Nil)
			if err != nil {
				return err
			}

		case code.OpGetBuiltin:
			idx := code.ReadUint8(instructions[v.curFrame().ip+1:])
			v.curFrame().ip += 1

			builtin := compiler.GetBuiltinByIndex(int(idx))
			if builtin == nil {
				return errors.New("invalid built-in function index")
			}
			err := v.push(builtin)
			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			// 设置变量
			index := code.ReadUint16(instructions[v.curFrame().ip+1:])
			v.curFrame().ip += 2

			v.globals[index] = v.pop()

		case code.OpGetGlobal:
			// 获取变量
			index := code.ReadUint16(instructions[v.curFrame().ip+1:])
			v.curFrame().ip += 2

			err := v.push(v.globals[index])
			if err != nil {
				return err
			}

		case code.OpSetLocal:
			idx := code.ReadUint8(instructions[v.curFrame().ip+1:])
			v.curFrame().ip += 1

			v.stack[v.curFrame().basePointer+int(idx)] = v.pop()

		case code.OpGetLocal:
			idx := code.ReadUint8(instructions[v.curFrame().ip+1:])
			v.curFrame().ip += 1

			err := v.push(v.stack[v.curFrame().basePointer+int(idx)])
			if err != nil {
				return err
			}

		case code.OpEQL, code.OpNEQ, code.OpGTR:
			err := v.executeComparisonOperation(op)
			if err != nil {
				return err
			}

		case code.OpMinus:
			err := v.executeMinusOperation()
			if err != nil {
				return err
			}

		case code.OpBang:
			err := v.executeBangOperation()
			if err != nil {
				return err
			}

		case code.OpJump:
			// 读取位置, 将i指向下一个要执行指令的位置
			pos := int(code.ReadUint16(instructions[v.curFrame().ip+1:]))
			v.curFrame().ip = pos - 1

		case code.OpClosure:
			idx := int(code.ReadUint16(instructions[v.curFrame().ip+1:]))
			countCtx := code.ReadUint8(instructions[v.curFrame().ip+3:])
			v.curFrame().ip += 3

			err := v.pushClosure(idx, int(countCtx))
			if err != nil {
				return err
			}

		case code.OpContext:
			idx := code.ReadUint8(instructions[v.curFrame().ip+1:])
			v.curFrame().ip += 1
			err := v.push(v.curFrame().cl.Ctx[idx])
			if err != nil {
				return err
			}

		case code.OpCurrClosure:
			err := v.push(v.curFrame().cl)
			if err != nil {
				return err
			}

		case code.OpJumpConditionNotTrue:
			// 判断条件是否为true, 如果不为true跳转到else分支否则正常执行下一条语句
			pos := int(code.ReadUint16(instructions[v.curFrame().ip+1:]))
			v.curFrame().ip += 2
			c := v.pop()
			if !v.isTrue(c) {
				v.curFrame().ip = pos - 1
			}

		case code.OpNil:
			err := v.push(Nil)
			if err != nil {
				return err
			}

		case code.OpPop: // 清理(弹栈)
			v.pop()
		}
	}
	return nil
}

func (v *VM) Run(handler ...func(vm *VM) error) error {
	handler = append(handler, defaultVmHandler)
	return handler[0](v)
}

func translationBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func NewVM(bytecode *compiler.Bytecode) *VM {
	mainFrame := &compiler.CompiledFunction{Instructions: bytecode.Instructions}
	frame := make([]*Frame, FrameSize)
	closure := &compiler.Closure{Fn: mainFrame}
	frame[0] = NewFrame(closure, 0)
	return &VM{
		constants: bytecode.Constants,
		stack:     make([]object.Object, StackSize), // 初始化栈大小
		sp:        0,

		globals:     make([]object.Object, GlobalsSize),
		frames:      frame,
		framesIndex: 1,
	}
}

func NewVMWithGlobals(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	mainFrame := &compiler.CompiledFunction{Instructions: bytecode.Instructions}
	frame := make([]*Frame, FrameSize)
	closure := &compiler.Closure{Fn: mainFrame}
	frame[0] = NewFrame(closure, 0)
	return &VM{
		constants: bytecode.Constants,
		stack:     make([]object.Object, StackSize), // 初始化栈大小
		sp:        0,

		globals:     globals,
		frames:      frame,
		framesIndex: 1,
	}
}
