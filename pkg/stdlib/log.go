package stdlib

import (
	"encoding/json"
	"fmt"
	"jabline/pkg/object"
	"os"
	"sync"
	"time"
)

var logMu sync.Mutex

var LogBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"log_info", &object.Builtin{Fn: logInfo}},
	{"log_warn", &object.Builtin{Fn: logWarn}},
	{"log_error", &object.Builtin{Fn: logError}},
	{"log_debug", &object.Builtin{Fn: logDebug}},
	{"log_set_level", &object.Builtin{Fn: logSetLevel}},
}

type logEntry struct {
	Level   string      `json:"level"`
	Time    string      `json:"time"`
	Message string      `json:"message"`
	Fields  interface{} `json:"fields,omitempty"`
}

var logLevel = 0 // 0=debug,1=info,2=warn,3=error

var logLevelNames = map[int]string{0: "DEBUG", 1: "INFO", 2: "WARN", 3: "ERROR"}

func init() {
	NativeModuleRegistry["_log"] = LogBuiltins
	NativeModulePrefixes["_log"] = "log_"
}

func logOutput(level int, msg string, fields interface{}) {
	logMu.Lock()
	defer logMu.Unlock()
	if level < logLevel {
		return
	}
	entry := logEntry{
		Level:   logLevelNames[level],
		Time:    time.Now().UTC().Format(time.RFC3339Nano),
		Message: msg,
		Fields:  fields,
	}
	var out []byte
	var err error
	if fields != nil {
		out, err = json.Marshal(entry)
	} else {
		out, err = json.Marshal(struct {
			Level   string `json:"level"`
			Time    string `json:"time"`
			Message string `json:"message"`
		}{
			Level:   logLevelNames[level],
			Time:    time.Now().UTC().Format(time.RFC3339Nano),
			Message: msg,
		})
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"level":"ERROR","time":"%s","message":"log marshal error: %s"}`+"\n", time.Now().UTC().Format(time.RFC3339Nano), err)
		return
	}
	var w *os.File
	if level >= 2 { // WARN, ERROR go to stderr
		w = os.Stderr
	} else {
		w = os.Stdout
	}
	w.Write(out)
	w.Write([]byte("\n"))
}

func toFields(obj object.Object) interface{} {
	if obj == nil {
		return nil
	}
	switch v := obj.(type) {
	case *object.Hash:
		m := make(map[string]interface{})
		for _, pair := range v.Pairs {
			k := ""
			if ks, ok := pair.Key.(*object.String); ok {
				k = ks.Value
			} else {
				continue
			}
			m[k] = toFields(pair.Value)
		}
		return m
	case *object.String:
		return v.Value
	case *object.Integer:
		return v.Value
	case *object.Float:
		return v.Value
	case *object.Boolean:
		return v.Value
	case *object.Array:
		arr := make([]interface{}, len(v.Elements))
		for i, el := range v.Elements {
			arr[i] = toFields(el)
		}
		return arr
	case *object.Null:
		return nil
	default:
		return v.Inspect()
	}
}

func logInfo(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("log_info expects at least 1 arg: (msg, [fields_hash])")
	}
	msg, ok := args[0].(*object.String)
	if !ok {
		return newError("log_info: message must be STRING, got %s", args[0].Type())
	}
	var fields interface{}
	if len(args) >= 2 {
		fields = toFields(args[1])
	}
	logOutput(1, msg.Value, fields)
	return &object.Null{}
}

func logWarn(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("log_warn expects at least 1 arg: (msg, [fields_hash])")
	}
	msg, ok := args[0].(*object.String)
	if !ok {
		return newError("log_warn: message must be STRING, got %s", args[0].Type())
	}
	var fields interface{}
	if len(args) >= 2 {
		fields = toFields(args[1])
	}
	logOutput(2, msg.Value, fields)
	return &object.Null{}
}

func logError(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("log_error expects at least 1 arg: (msg, [fields_hash])")
	}
	msg, ok := args[0].(*object.String)
	if !ok {
		return newError("log_error: message must be STRING, got %s", args[0].Type())
	}
	var fields interface{}
	if len(args) >= 2 {
		fields = toFields(args[1])
	}
	logOutput(3, msg.Value, fields)
	return &object.Null{}
}

func logDebug(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("log_debug expects at least 1 arg: (msg, [fields_hash])")
	}
	msg, ok := args[0].(*object.String)
	if !ok {
		return newError("log_debug: message must be STRING, got %s", args[0].Type())
	}
	var fields interface{}
	if len(args) >= 2 {
		fields = toFields(args[1])
	}
	logOutput(0, msg.Value, fields)
	return &object.Null{}
}

func logSetLevel(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("log_set_level expects 1 arg: (level_string)")
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("log_set_level: level must be STRING, got %s", args[0].Type())
	}
	switch s.Value {
	case "debug":
		logLevel = 0
	case "info":
		logLevel = 1
	case "warn":
		logLevel = 2
	case "error":
		logLevel = 3
	default:
		return newError("log_set_level: unknown level '%s', use: debug/info/warn/error", s.Value)
	}
	return &object.Null{}
}
