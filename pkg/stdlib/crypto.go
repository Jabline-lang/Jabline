package stdlib

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"jabline/pkg/object"
)

var CryptoBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"md5", &object.Builtin{Fn: cryptoMD5}},
	{"sha256", &object.Builtin{Fn: cryptoSHA256}},
	{"base64Encode", &object.Builtin{Fn: cryptoBase64Encode}},
	{"base64Decode", &object.Builtin{Fn: cryptoBase64Decode}},
}

func cryptoMD5(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `md5` must be STRING, got %s", args[0].Type())
	}
	hash := md5.Sum([]byte(s.Value))
	return &object.String{Value: fmt.Sprintf("%x", hash)}
}

func cryptoSHA256(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `sha256` must be STRING, got %s", args[0].Type())
	}
	hash := sha256.Sum256([]byte(s.Value))
	return &object.String{Value: fmt.Sprintf("%x", hash)}
}

func cryptoBase64Encode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `base64Encode` must be STRING, got %s", args[0].Type())
	}
	return &object.String{Value: base64.StdEncoding.EncodeToString([]byte(s.Value))}
}

func cryptoBase64Decode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `base64Decode` must be STRING, got %s", args[0].Type())
	}
	decoded, err := base64.StdEncoding.DecodeString(s.Value)
	if err != nil {
		return newError("failed to decode base64: %s", err)
	}
	return &object.String{Value: string(decoded)}
}
