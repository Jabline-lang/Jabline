package code

type Opcode byte

const (
	OpConstant Opcode = iota

	OpAdd
	OpSub
	OpMul
	OpDiv

	OpPop
	OpDup

	OpTrue
	OpFalse
	OpNull

	OpEqual
	OpNotEqual
	OpGreaterThan

	OpMinus
	OpBang

	OpJumpNotTruthy
	OpJump
	OpJumpNotNull
	OpJumpNotTruthyKeep
	OpJumpTruthyKeep
	OpJumpIfEqual
	OpJumpIfTrue

	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpGetFree

	OpArray
	OpHash
	OpIndex
	OpInstance

	OpCall
	OpReturnValue
	OpReturn
	OpClosure

	OpGetBuiltin
	OpImport

	OpThrow
	OpTry
	OpEndTry
)
