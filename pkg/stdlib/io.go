package stdlib

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"jabline/pkg/object"
	"os"
	"strings"
)

var IOBuiltins = []struct {
	Name    string
	Builtin *object.Builtin
}{
	{"readFile", &object.Builtin{Fn: ioReadFile}},
	{"writeFile", &object.Builtin{Fn: ioWriteFile}},
	{"echoUser", &object.Builtin{Fn: ioEchoUser}},
}

func ioReadFile(args ...object.Object) object.Object {
	if len(args) != 1 { return newError("wrong args") }
	filename, ok := args[0].(*object.String)
	if !ok { return newError("arg must be string") }
	
	content, err := ioutil.ReadFile(filename.Value)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	return &object.String{Value: string(content)}
}

func ioWriteFile(args ...object.Object) object.Object {
	if len(args) != 2 { return newError("wrong args") }
	filename, ok1 := args[0].(*object.String)
	content, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 { return newError("args must be strings") }
	
	err := ioutil.WriteFile(filename.Value, []byte(content.Value), 0644)
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
