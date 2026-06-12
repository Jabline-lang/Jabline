package code

import (
	"testing"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		op   Opcode
		name string
		wid  []int
	}{
		{OpConstant, "OpConstant", []int{2}},
		{OpAdd, "OpAdd", []int{}},
		{OpSub, "OpSub", []int{}},
		{OpMul, "OpMul", []int{}},
		{OpDiv, "OpDiv", []int{}},
		{OpMod, "OpMod", []int{}},
		{OpPop, "OpPop", []int{}},
		{OpDup, "OpDup", []int{}},
		{OpTrue, "OpTrue", []int{}},
		{OpFalse, "OpFalse", []int{}},
		{OpNull, "OpNull", []int{}},
		{OpEqual, "OpEqual", []int{}},
		{OpNotEqual, "OpNotEqual", []int{}},
		{OpGreaterThan, "OpGreaterThan", []int{}},
		{OpMinus, "OpMinus", []int{}},
		{OpBang, "OpBang", []int{}},
		{OpJumpNotTruthy, "OpJumpNotTruthy", []int{2}},
		{OpJump, "OpJump", []int{2}},
		{OpJumpNotNull, "OpJumpNotNull", []int{2}},
		{OpJumpNotTruthyKeep, "OpJumpNotTruthyKeep", []int{2}},
		{OpJumpTruthyKeep, "OpJumpTruthyKeep", []int{2}},
		{OpJumpIfEqual, "OpJumpIfEqual", []int{2}},
		{OpJumpIfTrue, "OpJumpIfTrue", []int{2}},
		{OpGetGlobal, "OpGetGlobal", []int{2}},
		{OpSetGlobal, "OpSetGlobal", []int{2}},
		{OpGetLocal, "OpGetLocal", []int{1}},
		{OpSetLocal, "OpSetLocal", []int{1}},
		{OpGetFree, "OpGetFree", []int{1}},
		{OpSetFree, "OpSetFree", []int{1}},
		{OpArray, "OpArray", []int{2}},
		{OpHash, "OpHash", []int{2}},
		{OpIndex, "OpIndex", []int{}},
		{OpInstance, "OpInstance", []int{2}},
		{OpGetProperty, "OpGetProperty", []int{}},
		{OpSetProperty, "OpSetProperty", []int{}},
		{OpCall, "OpCall", []int{1}},
		{OpReturnValue, "OpReturnValue", []int{}},
		{OpReturn, "OpReturn", []int{}},
		{OpClosure, "OpClosure", []int{2, 1}},
		{OpGetBuiltin, "OpGetBuiltin", []int{1}},
		{OpImport, "OpImport", []int{}},
		{OpThrow, "OpThrow", []int{}},
		{OpTry, "OpTry", []int{2, 2}},
		{OpEndTry, "OpEndTry", []int{}},
		{OpFinally, "OpFinally", []int{}},
		{OpEndFinally, "OpEndFinally", []int{}},
		{OpBitAnd, "OpBitAnd", []int{}},
		{OpBitOr, "OpBitOr", []int{}},
		{OpBitXor, "OpBitXor", []int{}},
		{OpBitNot, "OpBitNot", []int{}},
		{OpShiftLeft, "OpShiftLeft", []int{}},
		{OpShiftRight, "OpShiftRight", []int{}},
		{OpRegisterMethod, "OpRegisterMethod", []int{2, 2}},
		{OpService, "OpService", []int{2, 2}},
		{OpCheckType, "OpCheckType", []int{2}},
		{OpSendChannel, "OpSendChannel", []int{}},
		{OpRecvChannel, "OpRecvChannel", []int{}},
		{OpCurrentClosure, "OpCurrentClosure", []int{}},
		{OpInstantiate, "OpInstantiate", []int{1}},
		{OpFloatAdd, "OpFloatAdd", []int{}},
		{OpFloatSub, "OpFloatSub", []int{}},
		{OpFloatMul, "OpFloatMul", []int{}},
		{OpFloatDiv, "OpFloatDiv", []int{}},
		{OpIntToFloat, "OpIntToFloat", []int{}},
		{OpIsType, "OpIsType", []int{2}},
		{OpConstant8, "OpConstant8", []int{1}},
		{OpIncLocal, "OpIncLocal", []int{1}},
		{OpDecLocal, "OpDecLocal", []int{1}},
		{OpIncGlobal, "OpIncGlobal", []int{2}},
		{OpDecGlobal, "OpDecGlobal", []int{2}},
		{OpCallMethodFast, "OpCallMethodFast", []int{2, 1}},
		{OpMetricInc, "OpMetricInc", []int{2}},
		{OpTraceStart, "OpTraceStart", []int{2}},
		{OpTraceEnd, "OpTraceEnd", []int{}},
		{OpDefer, "OpDefer", []int{1}},
		{OpSelect, "OpSelect", []int{1, 1}},
		{OpNextItem, "OpNextItem", []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := Lookup(byte(tt.op))
			if err != nil {
				t.Fatalf("Lookup(%d) error: %s", tt.op, err)
			}
			if def.Name != tt.name {
				t.Errorf("Name = %q, want %q", def.Name, tt.name)
			}
			if len(def.OperandWidths) != len(tt.wid) {
				t.Errorf("len(OperandWidths) = %d, want %d", len(def.OperandWidths), len(tt.wid))
			}
			for i, w := range def.OperandWidths {
				if w != tt.wid[i] {
					t.Errorf("OperandWidths[%d] = %d, want %d", i, w, tt.wid[i])
				}
			}
		})
	}
}

func TestLookupInvalid(t *testing.T) {
	_, err := Lookup(255)
	if err == nil {
		t.Errorf("expected error for invalid opcode")
	}
}

func TestMake(t *testing.T) {
	tests := []struct {
		name     string
		op       Opcode
		operands []int
		want     []byte
	}{
		{"OpConstant 65535", OpConstant, []int{65535}, []byte{byte(OpConstant), 255, 255}},
		{"OpConstant 0", OpConstant, []int{0}, []byte{byte(OpConstant), 0, 0}},
		{"OpGetLocal 255", OpGetLocal, []int{255}, []byte{byte(OpGetLocal), 255}},
		{"OpGetLocal 0", OpGetLocal, []int{0}, []byte{byte(OpGetLocal), 0}},
		{"OpClosure 65535 255", OpClosure, []int{65535, 255}, []byte{byte(OpClosure), 255, 255, 255}},
		{"OpAdd", OpAdd, []int{}, []byte{byte(OpAdd)}},
		{"OpCall 1", OpCall, []int{1}, []byte{byte(OpCall), 1}},
		{"OpRegisterMethod", OpRegisterMethod, []int{1, 2}, []byte{byte(OpRegisterMethod), 0, 1, 0, 2}},
		{"OpConstant8 255", OpConstant8, []int{255}, []byte{byte(OpConstant8), 255}},
		{"OpJumpIfTrue 500", OpJumpIfTrue, []int{500}, []byte{byte(OpJumpIfTrue), 1, 244}},
		{"OpInvalid", Opcode(255), []int{}, []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Make(tt.op, tt.operands...)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i, b := range got {
				if b != tt.want[i] {
					t.Errorf("byte[%d] = %d, want %d", i, b, tt.want[i])
				}
			}
		})
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		name     string
		op       Opcode
		ins      []byte
		want     []int
		wantRead int
	}{
		{"OpConstant", OpConstant, []byte{255, 255}, []int{65535}, 2},
		{"OpGetLocal", OpGetLocal, []byte{255}, []int{255}, 1},
		{"OpClosure", OpClosure, []byte{255, 255, 255}, []int{65535, 255}, 3},
		{"OpCall", OpCall, []byte{5}, []int{5}, 1},
		{"OpJump", OpJump, []byte{0, 16}, []int{16}, 2},
		{"NoOperands", OpAdd, []byte{}, []int{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := Lookup(byte(tt.op))
			if err != nil {
				t.Fatal(err)
			}
			operands, read := ReadOperands(def, tt.ins)
			if read != tt.wantRead {
				t.Errorf("read = %d, want %d", read, tt.wantRead)
			}
			if len(operands) != len(tt.want) {
				t.Fatalf("len(operands) = %d, want %d", len(operands), len(tt.want))
			}
			for i, o := range operands {
				if o != tt.want[i] {
					t.Errorf("operand[%d] = %d, want %d", i, o, tt.want[i])
				}
			}
		})
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpAdd),
	}
	concat := Instructions{}
	for _, ins := range instructions {
		concat = append(concat, ins...)
	}
	out := concat.String()
	if out == "" {
		t.Errorf("String() returned empty")
	}
	if len(out) == 0 {
		t.Fatal("empty output")
	}
}

func TestInstructionRoundTrip(t *testing.T) {
	op := OpConstant
	val := 42
	ins := Make(op, val)
	def, _ := Lookup(byte(op))
	operands, read := ReadOperands(def, ins[1:])
	if read != 2 {
		t.Errorf("read = %d, want 2", read)
	}
	if len(operands) != 1 || operands[0] != val {
		t.Errorf("operand = %d, want %d", operands[0], val)
	}
}

func TestMakeLookupRoundTripAllOpcodes(t *testing.T) {
	for op := Opcode(0); op <= OpNextItem; op++ {
		def, err := Lookup(byte(op))
		if err != nil {
			t.Errorf("Lookup(%d) failed: %s", op, err)
			continue
		}

		operands := make([]int, len(def.OperandWidths))
		for i, w := range def.OperandWidths {
			switch w {
			case 1:
				operands[i] = 127
			case 2:
				operands[i] = 300
			}
		}

		ins := Make(op, operands...)
		if len(ins) == 0 && op != 255 {
			t.Errorf("Make(%s) returned empty", def.Name)
		}
		if len(ins) > 0 && byte(ins[0]) != byte(op) {
			t.Errorf("first byte = %d, want %d", ins[0], op)
		}
	}
}

func TestReadUint16(t *testing.T) {
	tests := []struct {
		name string
		ins  []byte
		want uint16
	}{
		{"zero", []byte{0, 0}, 0},
		{"one", []byte{0, 1}, 1},
		{"max", []byte{255, 255}, 65535},
		{"256", []byte{1, 0}, 256},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadUint16(tt.ins)
			if got != tt.want {
				t.Errorf("ReadUint16 = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSourcePos(t *testing.T) {
	sp := SourcePos{Line: 5, Column: 10}
	if sp.Line != 5 {
		t.Errorf("Line = %d, want 5", sp.Line)
	}
	if sp.Column != 10 {
		t.Errorf("Column = %d, want 10", sp.Column)
	}
}

func TestSourceMap(t *testing.T) {
	sm := SourceMap{0: {Line: 1, Column: 1}, 3: {Line: 2, Column: 5}}
	if len(sm) != 2 {
		t.Errorf("len = %d, want 2", len(sm))
	}
}
