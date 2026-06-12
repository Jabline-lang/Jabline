package stdlib

import (
	"bufio"
	"fmt"
	"jabline/pkg/object"
	"os"
	"strings"
)

var IOBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"readFile", &object.Builtin{Fn: ioReadFile}},
	{"writeFile", &object.Builtin{Fn: ioWriteFile}},
	{"appendFile", &object.Builtin{Fn: ioAppendFile}},
	{"copyFile", &object.Builtin{Fn: ioCopyFile}},
	{"fileExists", &object.Builtin{Fn: ioFileExists}},
	{"readDir", &object.Builtin{Fn: ioReadDir}},
	{"readLines", &object.Builtin{Fn: ioReadLines}},
	{"echoUser", &object.Builtin{Fn: ioEchoUser}},
}

func ioReadFile(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	filename, ok := args[0].(*object.String)
	if !ok {
		return newError("arg must be string")
	}

	content, err := os.ReadFile(filename.Value)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	return &object.String{Value: string(content)}
}

func ioReadDir(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("readDir expects 1 argument, got %d", len(args))
	}
	dirname, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to readDir must be STRING, got %s", args[0].Type())
	}

	entries, err := os.ReadDir(dirname.Value)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	var names []object.Object
	for _, entry := range entries {
		names = append(names, &object.String{Value: entry.Name()})
	}
	return &object.Array{Elements: names}
}

func ioReadLines(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	filename, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `readLines` must be STRING, got %s", args[0].Type())
	}

	file, err := os.Open(filename.Value)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	defer file.Close()

	var lines []object.Object
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, &object.String{Value: scanner.Text()})
	}

	if err := scanner.Err(); err != nil {
		return &object.Error{Message: err.Error()}
	}

	return &object.Array{Elements: lines}
}

func ioWriteFile(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	filename, ok1 := args[0].(*object.String)
	content, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("args must be strings")
	}

	err := os.WriteFile(filename.Value, []byte(content.Value), 0644)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	return &object.Boolean{Value: true}
}

func ioAppendFile(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("appendFile expects 2 arguments (filename, content), got %d", len(args))
	}
	filename, ok1 := args[0].(*object.String)
	content, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to appendFile must be STRING, got %s and %s", args[0].Type(), args[1].Type())
	}

	f, err := os.OpenFile(filename.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	defer f.Close()

	if _, err := f.WriteString(content.Value); err != nil {
		return &object.Error{Message: err.Error()}
	}
	return &object.Boolean{Value: true}
}

func ioCopyFile(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("copyFile expects 2 arguments (src, dst), got %d", len(args))
	}
	src, ok1 := args[0].(*object.String)
	dst, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to copyFile must be STRING, got %s and %s", args[0].Type(), args[1].Type())
	}

	data, err := os.ReadFile(src.Value)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	if err := os.WriteFile(dst.Value, data, 0644); err != nil {
		return &object.Error{Message: err.Error()}
	}
	return &object.Boolean{Value: true}
}

func ioFileExists(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("fileExists expects 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to fileExists must be STRING, got %s", args[0].Type())
	}

	_, err := os.Stat(path.Value)
	if err == nil {
		return &object.Boolean{Value: true}
	}
	return &object.Boolean{Value: false}
}

func ioEchoUser(args ...object.Object) object.Object {

	if len(args) > 0 {
		fmt.Print(args[0].Inspect())
	}

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return &object.Null{}
	}

	text = strings.TrimRight(text, "\r\n")

	return &object.String{Value: text}
}
