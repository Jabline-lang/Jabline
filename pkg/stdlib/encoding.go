package stdlib

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"jabline/pkg/object"
	"math"
)

var EncodingBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"base64Encode", &object.Builtin{Fn: encodingBase64Encode}},
	{"base64Decode", &object.Builtin{Fn: encodingBase64Decode}},
	{"hexEncode", &object.Builtin{Fn: encodingHexEncode}},
	{"hexDecode", &object.Builtin{Fn: encodingHexDecode}},
	{"pack", &object.Builtin{Fn: encodingPack}},
	{"unpack", &object.Builtin{Fn: encodingUnpack}},
}

// ─── Base64 / Hex (existing) ──────────────────────────────────────────────

func encodingBase64Encode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	input, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `base64Encode` must be STRING, got %s", args[0].Type())
	}
	return &object.String{Value: base64.StdEncoding.EncodeToString([]byte(input.Value))}
}

func encodingBase64Decode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	input, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `base64Decode` must be STRING, got %s", args[0].Type())
	}
	decoded, err := base64.StdEncoding.DecodeString(input.Value)
	if err != nil {
		return newError("failed to decode base64: %s", err)
	}
	return &object.String{Value: string(decoded)}
}

func encodingHexEncode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	input, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `hexEncode` must be STRING, got %s", args[0].Type())
	}
	return &object.String{Value: hex.EncodeToString([]byte(input.Value))}
}

func encodingHexDecode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	input, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `hexDecode` must be STRING, got %s", args[0].Type())
	}
	decoded, err := hex.DecodeString(input.Value)
	if err != nil {
		return newError("failed to decode hex: %s", err)
	}
	return &object.String{Value: string(decoded)}
}

// ─── Binary Serialization (pack / unpack) ────────────────────────────────
//
// Format tags (1 byte):
//   0x01 = null
//   0x02 = bool false
//   0x03 = bool true
//   0x04 = int64  (8 bytes, big-endian)
//   0x05 = float64 (8 bytes, big-endian IEEE 754)
//   0x06 = string  (4-byte length + UTF-8 bytes)
//   0x07 = array   (4-byte count + elements)
//   0x08 = hash    (4-byte count + key-value pairs)

const (
	tagNull   byte = 0x01
	tagFalse  byte = 0x02
	tagTrue   byte = 0x03
	tagInt    byte = 0x04
	tagFloat  byte = 0x05
	tagString byte = 0x06
	tagArray  byte = 0x07
	tagHash   byte = 0x08
)

func encodingPack(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("pack: wrong number of arguments. got=%d, want=1", len(args))
	}
	var buf bytes.Buffer
	if err := packObject(&buf, args[0]); err != nil {
		return newError("pack error: %s", err)
	}
	// Return as string (binary string) since Jabline has no Bytes type yet
	return &object.String{Value: buf.String()}
}

func encodingUnpack(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("unpack: wrong number of arguments. got=%d, want=1", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("unpack: argument must be a packed STRING")
	}
	r := bytes.NewReader([]byte(s.Value))
	obj, err := unpackObject(r)
	if err != nil {
		return newError("unpack error: %s", err)
	}
	return obj
}

func packObject(buf *bytes.Buffer, obj object.Object) error {
	switch v := obj.(type) {
	case *object.Null:
		buf.WriteByte(tagNull)

	case *object.Boolean:
		if v.Value {
			buf.WriteByte(tagTrue)
		} else {
			buf.WriteByte(tagFalse)
		}

	case *object.Integer:
		buf.WriteByte(tagInt)
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(v.Value))
		buf.Write(b)

	case *object.Float:
		buf.WriteByte(tagFloat)
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, math.Float64bits(v.Value))
		buf.Write(b)

	case *object.String:
		buf.WriteByte(tagString)
		strBytes := []byte(v.Value)
		lb := make([]byte, 4)
		binary.BigEndian.PutUint32(lb, uint32(len(strBytes)))
		buf.Write(lb)
		buf.Write(strBytes)

	case *object.Array:
		buf.WriteByte(tagArray)
		lb := make([]byte, 4)
		binary.BigEndian.PutUint32(lb, uint32(len(v.Elements)))
		buf.Write(lb)
		for _, el := range v.Elements {
			if err := packObject(buf, el); err != nil {
				return err
			}
		}

	case *object.Hash:
		buf.WriteByte(tagHash)
		lb := make([]byte, 4)
		binary.BigEndian.PutUint32(lb, uint32(len(v.Pairs)))
		buf.Write(lb)
		for _, pair := range v.Pairs {
			if err := packObject(buf, pair.Key); err != nil {
				return err
			}
			if err := packObject(buf, pair.Value); err != nil {
				return err
			}
		}

	default:
		// Fallback: serialize Inspect() as a string
		buf.WriteByte(tagString)
		strBytes := []byte(obj.Inspect())
		lb := make([]byte, 4)
		binary.BigEndian.PutUint32(lb, uint32(len(strBytes)))
		buf.Write(lb)
		buf.Write(strBytes)
	}
	return nil
}

func unpackObject(r *bytes.Reader) (object.Object, error) {
	tag, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch tag {
	case tagNull:
		return &object.Null{}, nil

	case tagFalse:
		return &object.Boolean{Value: false}, nil

	case tagTrue:
		return &object.Boolean{Value: true}, nil

	case tagInt:
		b := make([]byte, 8)
		if _, err := r.Read(b); err != nil {
			return nil, err
		}
		return &object.Integer{Value: int64(binary.BigEndian.Uint64(b))}, nil

	case tagFloat:
		b := make([]byte, 8)
		if _, err := r.Read(b); err != nil {
			return nil, err
		}
		return &object.Float{Value: math.Float64frombits(binary.BigEndian.Uint64(b))}, nil

	case tagString:
		lb := make([]byte, 4)
		if _, err := r.Read(lb); err != nil {
			return nil, err
		}
		strLen := binary.BigEndian.Uint32(lb)
		strBytes := make([]byte, strLen)
		if _, err := r.Read(strBytes); err != nil {
			return nil, err
		}
		return &object.String{Value: string(strBytes)}, nil

	case tagArray:
		lb := make([]byte, 4)
		if _, err := r.Read(lb); err != nil {
			return nil, err
		}
		count := binary.BigEndian.Uint32(lb)
		elements := make([]object.Object, count)
		for i := uint32(0); i < count; i++ {
			el, err := unpackObject(r)
			if err != nil {
				return nil, err
			}
			elements[i] = el
		}
		return &object.Array{Elements: elements}, nil

	case tagHash:
		lb := make([]byte, 4)
		if _, err := r.Read(lb); err != nil {
			return nil, err
		}
		count := binary.BigEndian.Uint32(lb)
		pairs := make(map[object.HashKey]object.HashPair)
		for i := uint32(0); i < count; i++ {
			key, err := unpackObject(r)
			if err != nil {
				return nil, err
			}
			val, err := unpackObject(r)
			if err != nil {
				return nil, err
			}
			h, ok := key.(object.Hashable)
			if !ok {
				return nil, fmt.Errorf("key is not hashable")
			}
			pairs[h.HashKey()] = object.HashPair{Key: key, Value: val}
		}
		return &object.Hash{Pairs: pairs}, nil

	default:
		return nil, fmt.Errorf("unknown tag: 0x%02x", tag)
	}
}
