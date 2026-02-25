package object

import (
	"encoding/json"
	"fmt"
	"net"
)

type RemoteChannel struct {
	Conn net.Conn
	Encoder *json.Encoder
	Decoder *json.Decoder
}

func (rc *RemoteChannel) Type() ObjectType { return "REMOTE_CHANNEL" }
func (rc *RemoteChannel) Inspect() string {
	return fmt.Sprintf("RemoteChannel(%s)", rc.Conn.RemoteAddr())
}

func (rc *RemoteChannel) Send(obj Object) error {
	// Protocol: JSON Line
	// Serialize object to native map/type
	native := ObjectToNative(obj)
	return rc.Encoder.Encode(native)
}

func (rc *RemoteChannel) Receive() (Object, error) {
	var native interface{}
	err := rc.Decoder.Decode(&native)
	if err != nil {
		return nil, err
	}
	return NativeToObject(native), nil
}

// Helpers for serialization (Should act as bridge between Object and Go types)

func ObjectToNative(obj Object) interface{} {
	if obj == nil { return nil }
	switch obj := obj.(type) {
	case *Integer: return obj.Value
	case *String: return obj.Value
	case *Boolean: return obj.Value
	case *Float: return obj.Value
	case *Array:
		list := make([]interface{}, len(obj.Elements))
		for i, el := range obj.Elements {
			list[i] = ObjectToNative(el)
		}
		return list
	case *Hash:
		m := make(map[string]interface{})
		for _, pair := range obj.Pairs {
			key, ok := pair.Key.(*String)
			if ok {
				m[key.Value] = ObjectToNative(pair.Value)
			}
		}
		return m
	}
	return nil
}

func NativeToObject(val interface{}) Object {
	switch v := val.(type) {
	case float64: return &Integer{Value: int64(v)} // JSON unmarshals numbers as floats
	case string: return &String{Value: v}
	case bool: return &Boolean{Value: v}
	case nil: return &Null{}
	case []interface{}:
		elements := make([]Object, len(v))
		for i, el := range v {
			elements[i] = NativeToObject(el)
		}
		return &Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[HashKey]HashPair)
		for k, val := range v {
			keyObj := &String{Value: k}
			valObj := NativeToObject(val)
			pairs[keyObj.HashKey()] = HashPair{Key: keyObj, Value: valObj}
		}
		return &Hash{Pairs: pairs}
	}
	return &Null{}
}
