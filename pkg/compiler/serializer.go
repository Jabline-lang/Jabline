package compiler

import (
	"bytes"
	"encoding/gob"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

type SerializableBytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	SourceMap    code.SourceMap
}

func Serialize(b *Bytecode) ([]byte, error) {

	registerTypes()

	sb := SerializableBytecode{
		Instructions: b.Instructions,
		Constants:    b.Constants,
		SourceMap:    b.SourceMap,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(sb)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Deserialize(data []byte) (*Bytecode, error) {
	registerTypes()

	var sb SerializableBytecode
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&sb)
	if err != nil {
		return nil, err
	}

	return &Bytecode{
		Instructions: sb.Instructions,
		Constants:    sb.Constants,
		SourceMap:    sb.SourceMap,
	}, nil
}

func registerTypes() {

	gob.Register(&object.Integer{})
	gob.Register(&object.String{})
	gob.Register(&object.Boolean{})
	gob.Register(&object.Null{})
	gob.Register(&object.CompiledFunction{})
	gob.Register(&object.Closure{})
	gob.Register(&object.Channel{})

}
