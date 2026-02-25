package stdlib

import (
	"encoding/base64"
	"encoding/hex"
	"jabline/pkg/object"
)

var EncodingBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"base64Encode", &object.Builtin{Fn: encodingBase64Encode}},
	{"base64Decode", &object.Builtin{Fn: encodingBase64Decode}},
	{"hexEncode", &object.Builtin{Fn: encodingHexEncode}},
	{"hexDecode", &object.Builtin{Fn: encodingHexDecode}},
}

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
