package stdlib

import (
	"encoding/xml"
	"strings"

	"jabline/pkg/object"
)

func init() {
	NativeModuleRegistry["_xml"] = XMLBuiltins
	NativeModulePrefixes["_xml"] = "xml_"
}

var XMLBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"xml_parse", &object.Builtin{Fn: xmlParse}},
	{"xml_encode", &object.Builtin{Fn: xmlEncode}},
	{"xml_parse_file", &object.Builtin{Fn: xmlParseFile}},
}

func xmlParse(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("xml_parse expects 1 argument, got %d", len(args))
	}

	s, ok := args[0].(*object.String)
	if !ok {
		return newError("xml_parse expects string, got %s", args[0].Type())
	}

	var result map[string]interface{}
	if err := xml.Unmarshal([]byte(s.Value), &result); err != nil {
		// XML into generic map may fail; try as raw bytes
		return xmlParseRaw(s.Value)
	}

	return yamlToObject(result)
}

func xmlParseRaw(xmlStr string) object.Object {
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	stack := make([]interface{}, 0)
	var root object.Object

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			elem := map[string]interface{}{
				"name": t.Name.Local,
			}
			attrs := make(map[string]string)
			for _, attr := range t.Attr {
				attrs[attr.Name.Local] = attr.Value
			}
			if len(attrs) > 0 {
				elem["attrs"] = attrs
			}
			stack = append(stack, elem)
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && len(stack) > 0 {
				if elem, ok := stack[len(stack)-1].(map[string]interface{}); ok {
					elem["text"] = text
				}
			}
		case xml.EndElement:
			if len(stack) == 1 {
				root = yamlToObject(stack[0])
			} else if len(stack) > 1 {
				child := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				parent := stack[len(stack)-1].(map[string]interface{})
				parent["child"] = child
			}
			stack = stack[:len(stack)-1]
		}
	}

	if root != nil {
		return root
	}
	return object.NullObj
}

func xmlEncode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("xml_encode expects 1 argument, got %d", len(args))
	}

	goVal := objectToGoValue(args[0])
	data, err := xml.MarshalIndent(goVal, "", "  ")
	if err != nil {
		return newError("xml_encode error: %s", err.Error())
	}

	return &object.String{Value: string(data)}
}

func xmlParseFile(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("xml_parse_file expects 1 argument, got %d", len(args))
	}

	path, ok := args[0].(*object.String)
	if !ok {
		return newError("xml_parse_file expects string, got %s", args[0].Type())
	}

	content, err := readFile(path.Value)
	if err != nil {
		return newError("xml_parse_file: %s", err.Error())
	}

	return xmlParse(&object.String{Value: string(content)})
}
