package code

type Opcode byte

const (
	OpConstant Opcode = iota

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod

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
	OpSetFree

	OpArray
	OpHash
	OpIndex
	OpInstance
	OpGetProperty
	OpSetProperty

	OpCall
	OpReturnValue
	OpReturn
	OpClosure

	OpSpawn
	OpAwait

	OpGetBuiltin
	OpImport

	OpThrow
	OpTry
	OpEndTry

	OpBitAnd
	OpBitOr
	OpBitXor
	OpBitNot
	OpShiftLeft
	OpShiftRight
	OpRegisterMethod
	OpService
	OpCheckType
	OpSendChannel
	OpRecvChannel
	OpCurrentClosure
	OpInstantiate
)
