package object

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	Conn *websocket.Conn
}

func (ws *WebSocket) Type() ObjectType {
	return WEBSOCKET_OBJ
}

func (ws *WebSocket) Inspect() string {
	if ws.Conn == nil {
		return "WebSocket(closed)"
	}
	return fmt.Sprintf("WebSocket(%p)", ws.Conn)
}

func (ws *WebSocket) CallMethod(method string, args ...Object) Object {
	switch method {
	case "read":
		if len(args) != 0 {
			return &Error{Message: fmt.Sprintf("wrong number of arguments for read. got=%d, want=0", len(args))}
		}
		if ws.Conn == nil {
			return &Error{Message: "websocket connection is closed"}
		}
		msgType, msg, err := ws.Conn.ReadMessage()
		if err != nil {
			return &Error{Message: fmt.Sprintf("websocket read error: %s", err)}
		}
		// For now we just return string representation, handling binary messages vs text
		if msgType == websocket.TextMessage {
			return &String{Value: string(msg)}
		} else {
			// Binary message
			return &String{Value: string(msg)}
		}

	case "writeText":
		if len(args) != 1 {
			return &Error{Message: fmt.Sprintf("wrong number of arguments for writeText. got=%d, want=1", len(args))}
		}
		if ws.Conn == nil {
			return &Error{Message: "websocket connection is closed"}
		}
		strObj, ok := args[0].(*String)
		if !ok {
			return &Error{Message: "argument to writeText must be a string"}
		}
		err := ws.Conn.WriteMessage(websocket.TextMessage, []byte(strObj.Value))
		if err != nil {
			return &Error{Message: fmt.Sprintf("websocket write error: %s", err)}
		}
		return &Null{}

	case "writeBinary":
		if len(args) != 1 {
			return &Error{Message: fmt.Sprintf("wrong number of arguments for writeBinary. got=%d, want=1", len(args))}
		}
		if ws.Conn == nil {
			return &Error{Message: "websocket connection is closed"}
		}
		strObj, ok := args[0].(*String)
		if !ok {
			return &Error{Message: "argument to writeBinary must be a string containing bytes"}
		}
		err := ws.Conn.WriteMessage(websocket.BinaryMessage, []byte(strObj.Value))
		if err != nil {
			return &Error{Message: fmt.Sprintf("websocket write error: %s", err)}
		}
		return &Null{}

	case "close":
		if len(args) != 0 {
			return &Error{Message: fmt.Sprintf("wrong number of arguments for close. got=%d, want=0", len(args))}
		}
		if ws.Conn != nil {
			ws.Conn.Close()
			ws.Conn = nil
		}
		return &Null{}

	default:
		return &Error{Message: fmt.Sprintf("unknown method '%s' for WebSocket object", method)}
	}
}
