package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type (
	Instructions []byte // 指令
	Opcode       byte   // 操作码

	Definition struct {
		Name          string // 名称
		OperandWidths []int  // 操作数宽度
	}
)

const (
	OpConstant Opcode = iota // 运算常量

	OpPop // 清理弹栈

	OpArray // 数组

	OpMap // hash map

	OpIndex

	OpAdd // +
	OpSub // -
	OpMul // *
	OpQuo // /

	OpTrue  // true
	OpFalse // false

	OpEQL // ==
	OpNEQ // !=
	OpGTR // >  (通过指令重排满足 < 运算符)

	OpMinus // -
	OpBang  // !

	OpCall        // call func
	OpReturnValue // return value
	OpReturn      // return nil(隐式返回)

	OpGetBuiltin

	OpGetGlobal // 获取全局变量值
	OpSetGlobal // 设置全局变量值

	OpGetLocal
	OpSetLocal

	OpClosure
	OpContext
	OpCurrClosure

	OpJump                 // 无条件跳转
	OpJumpConditionNotTrue // 条件不为真跳转

	OpNil // nil
)

func (ins Instructions) String() string {
	out := bytes.Buffer{}
	for i := 0; i < len(ins); {
		def, err := FindDefinitionByOp(ins[i])
		if err != nil {
			_, _ = fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		_, _ = fmt.Fprintf(&out, "%04d %s\n", i, ins.format(def, operands))
		i += read + 1
	}
	return out.String()
}

func (ins Instructions) format(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])

	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// definitions 默认定义
var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}}, // constant: address index

	OpPop: {"OpPop", []int{}},

	OpArray: {"OpArray", []int{2}}, // 弹栈数量 array len

	OpMap: {"OpMap", []int{2}}, // 弹栈数量  kv * 2

	OpIndex: {"OpIndex", []int{}},

	// 算数运算符
	OpAdd: {"OpAdd", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpQuo: {"OpQuo", []int{}},

	// bool
	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},

	// 比较运算符
	OpEQL: {"OpEQL", []int{}},
	OpNEQ: {"OpNEQ", []int{}},
	OpGTR: {"OpGTR", []int{}},

	// 前缀表达式
	OpMinus: {"OpMinus", []int{}},
	OpBang:  {"OpBang", []int{}},

	// func
	OpCall:        {"OpCall", []int{1}}, // call arg len
	OpReturnValue: {"OpReturnValue", []int{}},
	OpReturn:      {"OpReturn", []int{}},

	// 调用内部方法
	OpGetBuiltin: {"OpGetBuiltin", []int{1}}, // call builtin index

	// 值绑定
	OpGetGlobal: {"OpGetGlobal", []int{2}}, // get: address
	OpSetGlobal: {"OpSetGlobal", []int{2}}, // set: address

	OpGetLocal: {"OpGetLocal", []int{1}},
	OpSetLocal: {"OpSetLocal", []int{1}},

	OpClosure:     {"OpClosure", []int{2, 1}}, // OpConstant index, count closure
	OpContext:     {"OpContext", []int{1}},
	OpCurrClosure: {"OpCurrClosure", []int{}},

	// 指令跳转
	OpJump:                 {"OpJump", []int{2}},                 // jump: address
	OpJumpConditionNotTrue: {"OpJumpConditionNotTrue", []int{2}}, // jump: address

	OpNil: {"OpNil", []int{}},
}

// FindDefinitionByOp 根据op code 获取定义的操作结构
func FindDefinitionByOp(op byte) (*Definition, error) {
	v, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return v, nil
}

// ReadOperands 读取运算常量
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

func Make(op Opcode, operands ...int) []byte {
	v, ok := definitions[op]
	if !ok {
		return []byte{}
	}
	instructionLen := 1
	for _, w := range v.OperandWidths {
		instructionLen += w
	}
	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, operand := range operands {
		width := v.OperandWidths[i]
		switch width {
		case 1:
			instruction[offset] = byte(operand)

		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(operand))
		}
		offset += width
	}
	return instruction
}
