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
