package code

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:          {"OpConstant", []int{2}},
	OpAdd:               {"OpAdd", []int{}},
	OpSub:               {"OpSub", []int{}},
	OpMul:               {"OpMul", []int{}},
	OpDiv:               {"OpDiv", []int{}},
	OpPop:               {"OpPop", []int{}},
	OpDup:               {"OpDup", []int{}},
	OpTrue:              {"OpTrue", []int{}},
	OpFalse:             {"OpFalse", []int{}},
	OpNull:              {"OpNull", []int{}},
	OpEqual:             {"OpEqual", []int{}},
	OpNotEqual:          {"OpNotEqual", []int{}},
	OpGreaterThan:       {"OpGreaterThan", []int{}},
	OpMinus:             {"OpMinus", []int{}},
	OpBang:              {"OpBang", []int{}},
	OpJumpNotTruthy:     {"OpJumpNotTruthy", []int{2}},
	OpJump:              {"OpJump", []int{2}},
	OpJumpNotNull:       {"OpJumpNotNull", []int{2}},
	OpJumpNotTruthyKeep: {"OpJumpNotTruthyKeep", []int{2}},
	OpJumpTruthyKeep:    {"OpJumpTruthyKeep", []int{2}},
	OpJumpIfEqual:       {"OpJumpIfEqual", []int{2}},
	OpJumpIfTrue:        {"OpJumpIfTrue", []int{2}},
	OpGetGlobal:         {"OpGetGlobal", []int{2}},
	OpSetGlobal:         {"OpSetGlobal", []int{2}},
	OpGetLocal:          {"OpGetLocal", []int{1}},
	OpSetLocal:          {"OpSetLocal", []int{1}},
	OpGetFree:           {"OpGetFree", []int{1}},
	OpArray:             {"OpArray", []int{2}},
	OpHash:              {"OpHash", []int{2}},
	OpIndex:             {"OpIndex", []int{}},
	OpInstance:          {"OpInstance", []int{2}},
	OpCall:              {"OpCall", []int{1}},
	OpReturnValue:       {"OpReturnValue", []int{}},
	OpReturn:            {"OpReturn", []int{}},
	OpClosure:           {"OpClosure", []int{2, 1}},
	OpGetBuiltin:        {"OpGetBuiltin", []int{1}},
	OpImport:            {"OpImport", []int{}},
	OpThrow:             {"OpThrow", []int{}},
	OpTry:               {"OpTry", []int{2}},
	OpEndTry:            {"OpEndTry", []int{}},
}
