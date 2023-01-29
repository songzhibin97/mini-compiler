package code

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMake(t *testing.T) {
	type args struct {
		op       Opcode
		operands []int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{

		{
			name: "65535",
			args: args{
				op:       OpConstant,
				operands: []int{65535},
			},
			want: []byte{byte(OpConstant), 255, 255},
		},
		{
			name: "65534",
			args: args{
				op:       OpConstant,
				operands: []int{65534},
			},
			want: []byte{byte(OpConstant), 255, 254},
		},
		{
			name: "0",
			args: args{
				op:       OpConstant,
				operands: []int{0},
			},
			want: []byte{byte(OpConstant), 0, 0},
		},
		{
			name: "1",
			args: args{
				op:       OpConstant,
				operands: []int{1},
			},
			want: []byte{byte(OpConstant), 0, 1},
		},
		{
			name: "255",
			args: args{
				op:       OpGetLocal,
				operands: []int{255},
			},
			want: []byte{byte(OpGetLocal), 255},
		},
		{
			name: "",
			args: args{
				op:       OpClosure,
				operands: []int{65534, 255},
			},
			want: []byte{byte(OpClosure), 255, 254, 255},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Make(tt.args.op, tt.args.operands...)
			assert.Equal(t, len(tt.want), len(got))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInstructions_String(t *testing.T) {
	tests := []struct {
		name string
		ins  Instructions
		want string
	}{
		{
			name: "",
			ins:  Make(OpAdd),
			want: "0000 OpAdd\n",
		},
		{
			name: "",
			ins:  Make(OpConstant, 1),
			want: "0000 OpConstant 1\n",
		},
		{
			name: "",
			ins:  Make(OpConstant, 65535),
			want: "0000 OpConstant 65535\n",
		},
		{
			name: "",
			ins:  append([]byte(nil), append(Make(OpAdd), append(Make(OpConstant, 1), Make(OpConstant, 65535)...)...)...),
			want: "0000 OpAdd\n0001 OpConstant 1\n0004 OpConstant 65535\n",
		},
		{
			name: "",
			ins:  append([]byte(nil), append(Make(OpAdd), append(Make(OpGetLocal, 1), Make(OpConstant, 65535)...)...)...),
			want: "0000 OpAdd\n0001 OpGetLocal 1\n0003 OpConstant 65535\n",
		},
		{
			name: "",
			ins:  append([]byte(nil), append(Make(OpAdd), append(Make(OpGetLocal, 1), append(Make(OpConstant, 65535), Make(OpClosure, 65534, 255)...)...)...)...),
			want: "0000 OpAdd\n0001 OpGetLocal 1\n0003 OpConstant 65535\n0006 OpClosure 65534 255\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.ins.String(), "String()")
		})
	}
}

func TestReadOperands(t *testing.T) {

	tests := []struct {
		name      string
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{
			name:      "",
			op:        OpConstant,
			operands:  []int{1},
			bytesRead: 2,
		},
		{
			name:      "",
			op:        OpConstant,
			operands:  []int{65535},
			bytesRead: 2,
		},
		{
			name:      "",
			op:        OpGetLocal,
			operands:  []int{255},
			bytesRead: 1,
		},
		{
			name:      "",
			op:        OpClosure,
			operands:  []int{65534, 255},
			bytesRead: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instruction := Make(tt.op, tt.operands...)
			def, err := FindDefinitionByOp(byte(tt.op))
			assert.NoError(t, err)
			operandsRead, n := ReadOperands(def, instruction[1:])
			assert.Equal(t, n, tt.bytesRead)
			assert.Equal(t, operandsRead, tt.operands)
		})
	}
}
