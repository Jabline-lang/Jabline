package compiler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

type SerializableBytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	SourceMap    code.SourceMap
	Sandbox      string
}

// BytecodeMagic is the 4-byte magic header for Jabline bytecode.
var BytecodeMagic = []byte{'J', 'A', 'B', 'L'}

// BytecodeVersion is the current bytecode format version.
const BytecodeVersion uint16 = 1

func Serialize(b *Bytecode) ([]byte, error) {
	registerTypes()

	sb := SerializableBytecode{
		Instructions: b.Instructions,
		Constants:    b.Constants,
		SourceMap:    b.SourceMap,
		Sandbox:      b.Sandbox,
	}

	var gobBuf bytes.Buffer
	enc := gob.NewEncoder(&gobBuf)
	if err := enc.Encode(sb); err != nil {
		return nil, err
	}

	// Prepend magic header + version
	var out bytes.Buffer
	out.Write(BytecodeMagic)
	_ = out.WriteByte(byte(BytecodeVersion >> 8))
	_ = out.WriteByte(byte(BytecodeVersion))
	out.Write(gobBuf.Bytes())
	return out.Bytes(), nil
}

func Deserialize(data []byte) (*Bytecode, error) {
	registerTypes()

	if len(data) < 6 {
		return nil, fmt.Errorf("bytecode too short: %d bytes (need at least 6)", len(data))
	}

	// Verify magic header
	for i, b := range BytecodeMagic {
		if data[i] != b {
			return nil, fmt.Errorf("invalid bytecode magic: expected %c, got %c", b, rune(data[i]))
		}
	}

	// Verify version
	version := uint16(data[4])<<8 | uint16(data[5])
	if version != BytecodeVersion {
		return nil, fmt.Errorf("unsupported bytecode version: got %d, expected %d", version, BytecodeVersion)
	}

	var sb SerializableBytecode
	buf := bytes.NewBuffer(data[6:])
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&sb); err != nil {
		return nil, err
	}

	return &Bytecode{
		Instructions: sb.Instructions,
		Constants:    sb.Constants,
		SourceMap:    sb.SourceMap,
		Sandbox:      sb.Sandbox,
	}, nil
}

func registerTypes() {
	gob.Register(&object.Integer{})
	gob.Register(&object.Float{})
	gob.Register(&object.String{})
	gob.Register(&object.Boolean{})
	gob.Register(&object.Null{})
	gob.Register(&object.Int8{})
	gob.Register(&object.Int16{})
	gob.Register(&object.Int32{})
	gob.Register(&object.Int64{})
	gob.Register(&object.UInt8{})
	gob.Register(&object.UInt16{})
	gob.Register(&object.UInt32{})
	gob.Register(&object.UInt64{})
	gob.Register(&object.Float32{})
	gob.Register(&object.Float64{})
	gob.Register(&object.CompiledFunction{})
	gob.Register(&object.Closure{})
	gob.Register(&object.Channel{})
	gob.Register(&object.Array{})
	gob.Register(&object.Hash{})
	gob.Register(&object.Struct{})
	gob.Register(&object.Instance{})
	gob.Register(&object.InstantiatedStruct{})
	gob.Register(&object.Service{})
	gob.Register(&object.Interface{})
	gob.Register(&object.InterfaceMethod{})
	gob.Register(&object.BoundMethod{})
	gob.Register(&object.InstantiatedFunction{})
	gob.Register(&object.Promise{})
	gob.Register(&object.Error{})
	gob.Register(&object.ReturnValue{})
	gob.Register(&object.Break{})
	gob.Register(&object.Continue{})
	gob.Register(&object.Exception{})
	gob.Register(&object.Panic{})
	gob.Register(&object.DateTime{})
	gob.Register(&object.Regex{})
	gob.Register(&object.Image{})
	gob.Register(&object.WebSocket{})
	gob.Register(&object.Database{})
	gob.Register(&object.RemoteChannel{})
	gob.Register(&object.NetChannel{})
	gob.Register(&object.Function{})
	gob.Register(&object.ArrowFunction{})
	gob.Register(&object.AsyncFunction{})
	gob.Register(&object.Builtin{})
	gob.Register(&object.Environment{})
}
