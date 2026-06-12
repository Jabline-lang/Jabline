package stdlib

import (
	"encoding/csv"
	"os"
	"strings"

	"jabline/pkg/object"
)

func init() {
	NativeModuleRegistry["_csv"] = CSVBuiltins
	NativeModulePrefixes["_csv"] = "csv_"
}

var CSVBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"csv_parse", &object.Builtin{Fn: csvParse}},
	{"csv_encode", &object.Builtin{Fn: csvEncode}},
	{"csv_parse_from_file", &object.Builtin{Fn: csvParseFile}},
}

func csvParse(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("csv_parse expects 1 argument, got %d", len(args))
	}

	s, ok := args[0].(*object.String)
	if !ok {
		return newError("csv_parse expects string, got %s", args[0].Type())
	}

	reader := csv.NewReader(strings.NewReader(s.Value))
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return newError("csv_parse error: %s", err.Error())
	}

	rows := make([]object.Object, len(records))
	for i, record := range records {
		cols := make([]object.Object, len(record))
		for j, col := range record {
			cols[j] = &object.String{Value: col}
		}
		rows[i] = &object.Array{Elements: cols}
	}

	return &object.Array{Elements: rows}
}

func csvEncode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("csv_encode expects 1 argument, got %d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("csv_encode expects array of arrays, got %s", args[0].Type())
	}

	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	for _, row := range arr.Elements {
		rowArr, ok := row.(*object.Array)
		if !ok {
			return newError("csv_encode: each row must be an array, got %s", row.Type())
		}

		record := make([]string, len(rowArr.Elements))
		for i, col := range rowArr.Elements {
			strObj, ok := col.(*object.String)
			if !ok {
				return newError("csv_encode: each column must be a string, got %s", col.Type())
			}
			record[i] = strObj.Value
		}

		if err := writer.Write(record); err != nil {
			return newError("csv_encode error: %s", err.Error())
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return newError("csv_encode error: %s", err.Error())
	}

	return &object.String{Value: strings.TrimSpace(sb.String())}
}

func csvParseFile(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("csv_parse_from_file expects 1 argument, got %d", len(args))
	}

	path, ok := args[0].(*object.String)
	if !ok {
		return newError("csv_parse_from_file expects string path, got %s", args[0].Type())
	}

	// This uses the filesystem - sandbox check is done by the caller
	content, err := readFile(path.Value)
	if err != nil {
		return newError("csv_parse_from_file: %s", err.Error())
	}

	return csvParse(&object.String{Value: string(content)})
}

// readFile is a hook that can be set from the VM.
// It defaults to os.ReadFile but can be replaced for sandboxing.
var readFile func(path string) ([]byte, error)

func defaultReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func init() {
	readFile = defaultReadFile
}
