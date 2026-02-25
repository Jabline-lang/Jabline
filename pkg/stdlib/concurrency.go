package stdlib

import (
	"encoding/json"
	"fmt"
	"jabline/pkg/object"
	"net"
	"strings"
)

var ConcurrencyBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"make_chan", &object.Builtin{Fn: makeChan}},
	{"send", &object.Builtin{Fn: sendChan}},
	{"recv", &object.Builtin{Fn: recvChan}},
	{"connect", &object.Builtin{Fn: connectFunc}},
	{"listen", &object.Builtin{Fn: listenFunc}},
}

func makeChan(args ...object.Object) object.Object {
	ch := make(chan object.Object, 10) // Increased buffer size to 10
	return &object.Channel{Value: ch}
}

func sendChan(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	val := args[1]

	switch ch := args[0].(type) {
	case *object.Channel:
		ch.Value <- val
		return val
	case *object.RemoteChannel:
		if err := ch.Send(val); err != nil {
			return newError("remote send failed: %s", err)
		}
		return val
	default:
		return newError("arg must be channel")
	}
}

func recvChan(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}

	switch ch := args[0].(type) {
	case *object.Channel:
		val, ok := <-ch.Value
		if !ok {
			return &object.Null{}
		}
		return val
	case *object.RemoteChannel:
		val, err := ch.Receive()
		if err != nil {
			return newError("remote recv failed: %s", err)
		}
		return val
	default:
		return newError("arg must be channel")
	}
}

func connectFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("connect expects (url)")
	}
	urlStr, ok := args[0].(*object.String)
	if !ok {
		return newError("url must be string")
	}

	parts := strings.Split(urlStr.Value, "://")
	if len(parts) != 2 {
		return newError("invalid url format, expected protocol://addr")
	}
	proto := parts[0]
	addr := parts[1]

	conn, err := net.Dial(proto, addr)
	if err != nil {
		return newError("connect failed: %s", err)
	}

	return &object.RemoteChannel{
		Conn:    conn,
		Encoder: json.NewEncoder(conn),
		Decoder: json.NewDecoder(conn),
	}
}

func listenFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("listen expects (port)")
	}
	portObj, ok := args[0].(*object.Integer)
	if !ok {
		return newError("port must be integer")
	}

	addr := fmt.Sprintf(":%d", portObj.Value)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return newError("listen failed: %s", err)
	}

	clientChan := make(chan object.Object)

	go func() {
		defer close(clientChan)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			clientChan <- &object.RemoteChannel{
				Conn:    conn,
				Encoder: json.NewEncoder(conn),
				Decoder: json.NewDecoder(conn),
			}
		}
	}()

	return &object.Channel{Value: clientChan}
}
