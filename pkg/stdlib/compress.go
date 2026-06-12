package stdlib

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"

	"jabline/pkg/object"
)

func init() {
	NativeModuleRegistry["_compress"] = CompressBuiltins
	NativeModulePrefixes["_compress"] = "cmp_"
}

var CompressBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"cmp_gzip", &object.Builtin{Fn: compressGzip}},
	{"cmp_gunzip", &object.Builtin{Fn: compressGunzip}},
	{"cmp_zlib", &object.Builtin{Fn: compressZlib}},
	{"cmp_unzlib", &object.Builtin{Fn: compressUnzlib}},
	{"cmp_deflate", &object.Builtin{Fn: compressDeflate}},
}

func compressGzip(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("cmp_gzip expects 1 argument, got %d", len(args))
	}

	data, ok := getBytes(args[0])
	if !ok {
		return newError("cmp_gzip expects string or bytes, got %s", args[0].Type())
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return newError("cmp_gzip error: %s", err.Error())
	}
	if err := w.Close(); err != nil {
		return newError("cmp_gzip error: %s", err.Error())
	}

	return &object.String{Value: string(buf.Bytes())}
}

func compressGunzip(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("cmp_gunzip expects 1 argument, got %d", len(args))
	}

	data, ok := getBytes(args[0])
	if !ok {
		return newError("cmp_gunzip expects string or bytes, got %s", args[0].Type())
	}

	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return newError("cmp_gunzip error: %s", err.Error())
	}
	defer r.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		return newError("cmp_gunzip error: %s", err.Error())
	}

	return &object.String{Value: string(out)}
}

func compressZlib(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("cmp_zlib expects 1 argument, got %d", len(args))
	}

	data, ok := getBytes(args[0])
	if !ok {
		return newError("cmp_zlib expects string or bytes, got %s", args[0].Type())
	}

	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return newError("cmp_zlib error: %s", err.Error())
	}
	if err := w.Close(); err != nil {
		return newError("cmp_zlib error: %s", err.Error())
	}

	return &object.String{Value: string(buf.Bytes())}
}

func compressUnzlib(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("cmp_unzlib expects 1 argument, got %d", len(args))
	}

	data, ok := getBytes(args[0])
	if !ok {
		return newError("cmp_unzlib expects string or bytes, got %s", args[0].Type())
	}

	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return newError("cmp_unzlib error: %s", err.Error())
	}
	defer r.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		return newError("cmp_unzlib error: %s", err.Error())
	}

	return &object.String{Value: string(out)}
}

func compressDeflate(args ...object.Object) object.Object {
	return compressZlib(args...)
}

func getBytes(obj object.Object) ([]byte, bool) {
	switch o := obj.(type) {
	case *object.String:
		return []byte(o.Value), true
	default:
		return nil, false
	}
}
