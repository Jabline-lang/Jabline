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
	OpTailCall
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
	OpFinally
	OpEndFinally

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

	// Float specific opcodes
	OpFloatAdd
	OpFloatSub
	OpFloatMul
	OpFloatDiv
	OpIntToFloat
	OpIsType

	// Fast-Paths and Compression (Priority 15)
	OpConstant8
	OpIncLocal
	OpDecLocal
	OpIncGlobal
	OpDecGlobal
	OpCallMethodFast
	OpMetricInc
	OpTraceStart
	OpTraceEnd

	// Slice operator support
	OpSlice

	// Spread operator support
	OpBuildArrayWithSpread
	OpCallSpread

	// Defer support
	OpDefer

	// Select support
	OpSelect

	// For-each next item (handles arrays, strings, channels)
	OpNextItem
)
