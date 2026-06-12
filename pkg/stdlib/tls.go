package stdlib

import (
	"crypto/tls"
	"net"
	"time"

	"jabline/pkg/object"
)

func init() {
	NativeModuleRegistry["_tls"] = TLSBuiltins
	NativeModulePrefixes["_tls"] = "tls_"
}

var TLSBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"tls_connect", &object.Builtin{Fn: tlsConnect}},
	{"tls_dial", &object.Builtin{Fn: tlsDial}},
	{"tls_verify", &object.Builtin{Fn: tlsVerify}},
}

func tlsConnect(args ...object.Object) object.Object {
	if len(args) != 1 && len(args) != 2 {
		return newError("tls_connect expects 1-2 arguments, got %d", len(args))
	}

	addr, ok := args[0].(*object.String)
	if !ok {
		return newError("tls_connect expects string address, got %s", args[0].Type())
	}

	insecure := false
	if len(args) == 2 {
		b, ok := args[1].(*object.Boolean)
		if ok {
			insecure = b.Value
		}
	}

	conn, err := tls.Dial("tcp", addr.Value, &tls.Config{
		InsecureSkipVerify: insecure,
	})
	if err != nil {
		return newError("tls_connect error: %s", err.Error())
	}

	return &object.NetChannel{
		Conn:   conn,
		Buffer: make([]byte, 4096),
	}
}

func tlsDial(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("tls_dial expects 2-3 arguments, got %d", len(args))
	}

	network, ok := args[0].(*object.String)
	if !ok {
		return newError("tls_dial expects string network, got %s", args[0].Type())
	}

	addr, ok := args[1].(*object.String)
	if !ok {
		return newError("tls_dial expects string address, got %s", args[1].Type())
	}

	timeout := 30 * time.Second
	if len(args) == 3 {
		if secObj, ok := args[2].(*object.Integer); ok {
			timeout = time.Duration(secObj.Value) * time.Second
		}
	}

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: timeout},
		Config:    &tls.Config{},
	}

	conn, err := dialer.Dial(network.Value, addr.Value)
	if err != nil {
		return newError("tls_dial error: %s", err.Error())
	}

	return &object.NetChannel{
		Conn:   conn,
		Buffer: make([]byte, 4096),
	}
}

func tlsVerify(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("tls_verify expects 1 argument, got %d", len(args))
	}

	addr, ok := args[0].(*object.String)
	if !ok {
		return newError("tls_verify expects string, got %s", args[0].Type())
	}

	conn, err := tls.Dial("tcp", addr.Value, &tls.Config{})
	if err != nil {
		return &object.Boolean{Value: false}
	}
	defer conn.Close()

	// Verify connection state
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return &object.Boolean{Value: false}
	}

	return &object.Boolean{Value: true}
}
