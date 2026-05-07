//go:build !runner || runner_websocket

package stdlib

import (
	"jabline/pkg/object"

	"github.com/gorilla/websocket"
)

var WebsocketBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"wsConnect", &object.Builtin{Fn: wsConnect}},
	{"wsRead", &object.Builtin{Fn: wsRead}},
	{"wsWriteText", &object.Builtin{Fn: wsWriteText}},
	{"wsClose", &object.Builtin{Fn: wsClose}},
}

func wsConnect(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for wsConnect. got=%d, want=1 (url)", len(args))
	}
	urlStr, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `wsConnect` must be STRING, got %s", args[0].Type())
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(urlStr.Value, nil)
	if err != nil {
		return newError("websocket connection error: %s", err)
	}

	return &object.WebSocket{Conn: conn}
}

func getConnInfo(arg object.Object) (*websocket.Conn, object.Object) {
	wsObj, ok := arg.(*object.WebSocket)
	if !ok {
		return nil, newError("argument must be WEBSOCKET_OBJ")
	}
	if wsObj.Conn == nil {
		return nil, newError("websocket connection is closed")
	}
	return wsObj.Conn, nil
}

func wsRead(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for wsRead. got=%d, want=1", len(args))
	}
	conn, errObj := getConnInfo(args[0])
	if errObj != nil {
		return errObj
	}

	msgType, msg, err := conn.ReadMessage()
	if err != nil {
		return newError("websocket read error: %s", err)
	}

	if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
		return &object.String{Value: string(msg)}
	}
	return &object.Null{}
}

func wsWriteText(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments for wsWriteText. got=%d, want=2", len(args))
	}
	conn, errObj := getConnInfo(args[0])
	if errObj != nil {
		return errObj
	}

	strObj, ok := args[1].(*object.String)
	if !ok {
		return newError("second argument to wsWriteText must be a string")
	}

	err := conn.WriteMessage(websocket.TextMessage, []byte(strObj.Value))
	if err != nil {
		return newError("websocket write error: %s", err)
	}

	return &object.Null{}
}

func wsClose(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for wsClose. got=%d, want=1", len(args))
	}
	wsObj, ok := args[0].(*object.WebSocket)
	if !ok {
		return newError("argument must be WEBSOCKET_OBJ")
	}
	if wsObj.Conn != nil {
		wsObj.Conn.Close()
		wsObj.Conn = nil
	}
	return &object.Null{}
}


