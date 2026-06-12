package stdlib

import (
	"fmt"
	"jabline/pkg/object"
	"net/smtp"
)

var SMTPBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"smtp_send", &object.Builtin{Fn: smtpSend}},
}

func init() {
	NativeModuleRegistry["_smtp"] = SMTPBuiltins
	NativeModulePrefixes["_smtp"] = "smtp_"
}

func smtpSend(args ...object.Object) object.Object {
	if len(args) < 5 {
		return newError("smtp_send expects at least 5 args: (host, port, from, to, subject, body...)")
	}

	host, ok := args[0].(*object.String)
	if !ok {
		return newError("host must be STRING, got %s", args[0].Type())
	}
	port, ok := args[1].(*object.Integer)
	if !ok {
		return newError("port must be INTEGER, got %s", args[1].Type())
	}
	from, ok := args[2].(*object.String)
	if !ok {
		return newError("from must be STRING, got %s", args[2].Type())
	}
	to, ok := args[3].(*object.String)
	if !ok {
		return newError("to must be STRING, got %s", args[3].Type())
	}
	subject, ok := args[4].(*object.String)
	if !ok {
		return newError("subject must be STRING, got %s", args[4].Type())
	}

	body := ""
	if len(args) >= 6 {
		if b, ok := args[5].(*object.String); ok {
			body = b.Value
		}
	}

	username := ""
	password := ""
	if len(args) >= 8 {
		if u, ok := args[6].(*object.String); ok {
			username = u.Value
		}
		if p, ok := args[7].(*object.String); ok {
			password = p.Value
		}
	}

	addr := host.Value + ":" + fmt.Sprintf("%d", port.Value)

	msg := []byte("From: " + from.Value + "\r\n" +
		"To: " + to.Value + "\r\n" +
		"Subject: " + subject.Value + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" + body + "\r\n")

	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host.Value)
	}

	err := smtp.SendMail(addr, auth, from.Value, []string{to.Value}, msg)
	if err != nil {
		return newError("smtp_send failed: %s", err)
	}

	return &object.Boolean{Value: true}
}
