package evaluator

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jabline/pkg/object"
)

var IOBuiltins = map[string]*object.Builtin{
	"readFile": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `readFile` must be STRING, got %T", args[0])
			}

			content, err := ioutil.ReadFile(filename.Value)
			if err != nil {
				return newError("error reading file: %s", err.Error())
			}

			return &object.String{Value: string(content)}
		},
	},

	"writeFile": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `writeFile` must be STRING, got %T", args[0])
			}

			content, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `writeFile` must be STRING, got %T", args[1])
			}

			err := ioutil.WriteFile(filename.Value, []byte(content.Value), 0644)
			if err != nil {
				return newError("error writing file: %s", err.Error())
			}

			return TRUE
		},
	},

	"fileExists": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `fileExists` must be STRING, got %T", args[0])
			}

			_, err := os.Stat(filename.Value)
			if os.IsNotExist(err) {
				return FALSE
			}

			return TRUE
		},
	},

	"deleteFile": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			filename, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `deleteFile` must be STRING, got %T", args[0])
			}

			err := os.Remove(filename.Value)
			if err != nil {
				return newError("error deleting file: %s", err.Error())
			}

			return TRUE
		},
	},

	"createDir": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			dirname, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `createDir` must be STRING, got %T", args[0])
			}

			err := os.MkdirAll(dirname.Value, 0755)
			if err != nil {
				return newError("error creating directory: %s", err.Error())
			}

			return TRUE
		},
	},

	"listDir": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			dirname, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `listDir` must be STRING, got %T", args[0])
			}

			files, err := ioutil.ReadDir(dirname.Value)
			if err != nil {
				return newError("error reading directory: %s", err.Error())
			}

			elements := make([]object.Object, len(files))
			for i, file := range files {
				fileInfo := make(map[object.HashKey]object.HashPair)

				nameKey := (&object.String{Value: "name"}).HashKey()
				fileInfo[nameKey] = object.HashPair{
					Key:   &object.String{Value: "name"},
					Value: &object.String{Value: file.Name()},
				}

				isFileKey := (&object.String{Value: "isFile"}).HashKey()
				fileInfo[isFileKey] = object.HashPair{
					Key:   &object.String{Value: "isFile"},
					Value: nativeBoolToPyObject(!file.IsDir()),
				}

				sizeKey := (&object.String{Value: "size"}).HashKey()
				fileInfo[sizeKey] = object.HashPair{
					Key:   &object.String{Value: "size"},
					Value: &object.Integer{Value: file.Size()},
				}

				elements[i] = &object.Hash{Pairs: fileInfo}
			}

			return &object.Array{Elements: elements}
		},
	},

	"getWorkingDir": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}

			wd, err := os.Getwd()
			if err != nil {
				return newError("error getting working directory: %s", err.Error())
			}

			return &object.String{Value: wd}
		},
	},

	"changeDir": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			dirname, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `changeDir` must be STRING, got %T", args[0])
			}

			err := os.Chdir(dirname.Value)
			if err != nil {
				return newError("error changing directory: %s", err.Error())
			}

			return TRUE
		},
	},

	"pathJoin": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) == 0 {
				return newError("wrong number of arguments. got=%d, want at least 1", len(args))
			}

			paths := make([]string, len(args))
			for i, arg := range args {
				str, ok := arg.(*object.String)
				if !ok {
					return newError("argument %d to `pathJoin` must be STRING, got %T", i, arg)
				}
				paths[i] = str.Value
			}

			result := filepath.Join(paths...)
			return &object.String{Value: result}
		},
	},

	"pathBase": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `pathBase` must be STRING, got %T", args[0])
			}

			result := filepath.Base(path.Value)
			return &object.String{Value: result}
		},
	},

	"pathDir": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `pathDir` must be STRING, got %T", args[0])
			}

			result := filepath.Dir(path.Value)
			return &object.String{Value: result}
		},
	},

	"httpGet": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			url, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `httpGet` must be STRING, got %T", args[0])
			}

			resp, err := http.Get(url.Value)
			if err != nil {
				return newError("HTTP GET error: %s", err.Error())
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return newError("error reading response body: %s", err.Error())
			}

			result := make(map[object.HashKey]object.HashPair)

			statusKey := (&object.String{Value: "status"}).HashKey()
			result[statusKey] = object.HashPair{
				Key:   &object.String{Value: "status"},
				Value: &object.Integer{Value: int64(resp.StatusCode)},
			}

			bodyKey := (&object.String{Value: "body"}).HashKey()
			result[bodyKey] = object.HashPair{
				Key:   &object.String{Value: "body"},
				Value: &object.String{Value: string(body)},
			}

			return &object.Hash{Pairs: result}
		},
	},

	"httpPost": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			url, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `httpPost` must be STRING, got %T", args[0])
			}

			data, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `httpPost` must be STRING, got %T", args[1])
			}

			resp, err := http.Post(url.Value, "application/json", strings.NewReader(data.Value))
			if err != nil {
				return newError("HTTP POST error: %s", err.Error())
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return newError("error reading response body: %s", err.Error())
			}

			result := make(map[object.HashKey]object.HashPair)

			statusKey := (&object.String{Value: "status"}).HashKey()
			result[statusKey] = object.HashPair{
				Key:   &object.String{Value: "status"},
				Value: &object.Integer{Value: int64(resp.StatusCode)},
			}

			bodyKey := (&object.String{Value: "body"}).HashKey()
			result[bodyKey] = object.HashPair{
				Key:   &object.String{Value: "body"},
				Value: &object.String{Value: string(body)},
			}

			return &object.Hash{Pairs: result}
		},
	},

	"input": {
		Fn: func(args ...object.Object) object.Object {
			var prompt string
			if len(args) == 1 {
				promptObj, ok := args[0].(*object.String)
				if !ok {
					return newError("argument to `input` must be STRING, got %T", args[0])
				}
				prompt = promptObj.Value
			} else if len(args) > 1 {
				return newError("wrong number of arguments. got=%d, want 0 or 1", len(args))
			}

			if prompt != "" {
				fmt.Print(prompt)
			}

			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				return newError("error reading input: %s", err.Error())
			}

			text = strings.TrimSuffix(text, "\n")
			text = strings.TrimSuffix(text, "\r")

			return &object.String{Value: text}
		},
	},

	"print": {
		Fn: func(args ...object.Object) object.Object {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(arg.Inspect())
			}
			return NULL
		},
	},

	"println": {
		Fn: func(args ...object.Object) object.Object {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(arg.Inspect())
			}
			fmt.Println()
			return NULL
		},
	},

	"getEnv": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			key, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `getEnv` must be STRING, got %T", args[0])
			}

			value := os.Getenv(key.Value)
			if value == "" {
				return NULL
			}

			return &object.String{Value: value}
		},
	},

	"setEnv": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			key, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `setEnv` must be STRING, got %T", args[0])
			}

			value, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `setEnv` must be STRING, got %T", args[1])
			}

			err := os.Setenv(key.Value, value.Value)
			if err != nil {
				return newError("error setting environment variable: %s", err.Error())
			}

			return TRUE
		},
	},

	"now": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}

			timestamp := time.Now().Unix()
			return &object.Integer{Value: timestamp}
		},
	},

	"sleep": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			duration, ok := args[0].(*object.Integer)
			if !ok {
				return newError("argument to `sleep` must be INTEGER, got %T", args[0])
			}

			time.Sleep(time.Duration(duration.Value) * time.Millisecond)
			return NULL
		},
	},

	"formatTime": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			timestamp, ok := args[0].(*object.Integer)
			if !ok {
				return newError("first argument to `formatTime` must be INTEGER, got %T", args[0])
			}

			format, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `formatTime` must be STRING, got %T", args[1])
			}

			t := time.Unix(timestamp.Value, 0)

			goFormat := format.Value
			goFormat = strings.ReplaceAll(goFormat, "YYYY", "2006")
			goFormat = strings.ReplaceAll(goFormat, "MM", "01")
			goFormat = strings.ReplaceAll(goFormat, "DD", "02")
			goFormat = strings.ReplaceAll(goFormat, "HH", "15")
			goFormat = strings.ReplaceAll(goFormat, "mm", "04")
			goFormat = strings.ReplaceAll(goFormat, "ss", "05")

			result := t.Format(goFormat)
			return &object.String{Value: result}
		},
	},
}

func nativeBoolToPyObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}
