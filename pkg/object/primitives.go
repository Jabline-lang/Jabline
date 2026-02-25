package object

import (
	"fmt"
	"hash/fnv"
	"math"
)

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }
func (f *Float) HashKey() HashKey {
	return HashKey{Type: f.Type(), Value: uint64(f.Value)}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// --- Signed Integers ---

type Int8 struct {
	Value int8
}
func (i *Int8) Type() ObjectType { return INT8_OBJ }
func (i *Int8) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Int8) HashKey() HashKey { return HashKey{Type: i.Type(), Value: uint64(i.Value)} }

type Int16 struct {
	Value int16
}
func (i *Int16) Type() ObjectType { return INT16_OBJ }
func (i *Int16) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Int16) HashKey() HashKey { return HashKey{Type: i.Type(), Value: uint64(i.Value)} }

type Int32 struct {
	Value int32
}
func (i *Int32) Type() ObjectType { return INT32_OBJ }
func (i *Int32) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Int32) HashKey() HashKey { return HashKey{Type: i.Type(), Value: uint64(i.Value)} }

type Int64 struct {
	Value int64
}
func (i *Int64) Type() ObjectType { return INT64_OBJ }
func (i *Int64) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Int64) HashKey() HashKey { return HashKey{Type: i.Type(), Value: uint64(i.Value)} }

// --- Unsigned Integers ---

type UInt8 struct {
	Value uint8
}
func (i *UInt8) Type() ObjectType { return UINT8_OBJ }
func (i *UInt8) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *UInt8) HashKey() HashKey { return HashKey{Type: i.Type(), Value: uint64(i.Value)} }

type UInt16 struct {
	Value uint16
}
func (i *UInt16) Type() ObjectType { return UINT16_OBJ }
func (i *UInt16) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *UInt16) HashKey() HashKey { return HashKey{Type: i.Type(), Value: uint64(i.Value)} }

type UInt32 struct {
	Value uint32
}
func (i *UInt32) Type() ObjectType { return UINT32_OBJ }
func (i *UInt32) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *UInt32) HashKey() HashKey { return HashKey{Type: i.Type(), Value: uint64(i.Value)} }

type UInt64 struct {
	Value uint64
}
func (i *UInt64) Type() ObjectType { return UINT64_OBJ }
func (i *UInt64) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *UInt64) HashKey() HashKey { return HashKey{Type: i.Type(), Value: i.Value} }

// --- Floating Point ---

type Float32 struct {
	Value float32
}
func (f *Float32) Type() ObjectType { return FLOAT32_OBJ }
func (f *Float32) Inspect() string  { return fmt.Sprintf("%g", f.Value) }
func (f *Float32) HashKey() HashKey {
	return HashKey{Type: f.Type(), Value: uint64(math.Float32bits(f.Value))}
}

type Float64 struct {
	Value float64
}
func (f *Float64) Type() ObjectType { return FLOAT64_OBJ }
func (f *Float64) Inspect() string  { return fmt.Sprintf("%g", f.Value) }
func (f *Float64) HashKey() HashKey {
	return HashKey{Type: f.Type(), Value: math.Float64bits(f.Value)}
}
