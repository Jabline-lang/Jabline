package stdlib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"jabline/pkg/object"
)

var CryptoBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"__native_md5", &object.Builtin{Fn: cryptoMD5}},
	{"__native_sha256", &object.Builtin{Fn: cryptoSHA256}},
	{"__native_base64Encode", &object.Builtin{Fn: cryptoBase64Encode}},
	{"__native_base64Decode", &object.Builtin{Fn: cryptoBase64Decode}},
	{"__native_aesEncrypt", &object.Builtin{Fn: cryptoAESEncrypt}},
	{"__native_aesDecrypt", &object.Builtin{Fn: cryptoAESDecrypt}},
	{"__native_randomBytes", &object.Builtin{Fn: cryptoRandomBytes}},
	{"__native_uuid", &object.Builtin{Fn: cryptoUUID}},
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

func cryptoRandomBytes(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	n, ok := args[0].(*object.Integer)
	if !ok {
		return newError("argument to `randomBytes` must be INTEGER, got %s", args[0].Type())
	}

	bytes := make([]byte, n.Value)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return newError("failed to generate random bytes: %s", err)
	}

	return &object.String{Value: string(bytes)}
}

func cryptoUUID(args ...object.Object) object.Object {
	uuid := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, uuid); err != nil {
		return newError("failed to generate UUID: %s", err)
	}

	// Variant and version bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant 10

	return &object.String{Value: fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])}
}

func cryptoAESEncrypt(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	keyStr, ok1 := args[0].(*object.String) // key
	plain, ok2 := args[1].(*object.String)  // plaintext

	if !ok1 || !ok2 {
		return newError("arguments to `aesEncrypt` must be STRING")
	}

	key := []byte(keyStr.Value)
	block, err := aes.NewCipher(key)
	if err != nil {
		return newError("failed to create AES cipher: %s", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return newError("failed to create GCM: %s", err)
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return newError("failed to generate nonce: %s", err)
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plain.Value), nil)
	return &object.String{Value: base64.StdEncoding.EncodeToString(ciphertext)}
}

func cryptoAESDecrypt(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	keyStr, ok1 := args[0].(*object.String) // key
	cipherStr, ok2 := args[1].(*object.String) // base64 ciphertext

	if !ok1 || !ok2 {
		return newError("arguments to `aesDecrypt` must be STRING")
	}

	data, err := base64.StdEncoding.DecodeString(cipherStr.Value)
	if err != nil {
		return newError("failed to decode base64 ciphertext: %s", err)
	}

	key := []byte(keyStr.Value)
	block, err := aes.NewCipher(key)
	if err != nil {
		return newError("failed to create AES cipher: %s", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return newError("failed to create GCM: %s", err)
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return newError("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return newError("failed to decrypt: %s (check if your key is correct)", err)
	}

	return &object.String{Value: string(plaintext)}
}

