package stdlib

import (
	"bytes"
	"jabline/pkg/object"
	"text/template"
)

var TemplateBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"render_template", &object.Builtin{Fn: renderTemplateFunc}},
}

// Convert Jabline object to Go interface{} for text/template
func jablineToGoValue(obj object.Object) interface{} {
	switch o := obj.(type) {
	case *object.String:
		return o.Value
	case *object.Integer:
		return o.Value
	case *object.Boolean:
		return o.Value
	case *object.Array:
		var arr []interface{}
		for _, el := range o.Elements {
			arr = append(arr, jablineToGoValue(el))
		}
		return arr
	case *object.Hash:
		m := make(map[string]interface{})
		for _, pair := range o.Pairs {
			if strKey, ok := pair.Key.(*object.String); ok {
				m[strKey.Value] = jablineToGoValue(pair.Value)
			}
		}
		return m
	}
	return nil
}

// render_template(tmplString, dataHash)
func renderTemplateFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("render_template expects 2 arguments: (template_string, data_hash)")
	}

	tmplStr, ok := args[0].(*object.String)
	if !ok {
		return newError("first argument must be a string")
	}

	dataHash, ok := args[1].(*object.Hash)
	if !ok {
		return newError("second argument must be a hash")
	}

	goData := jablineToGoValue(dataHash)

	tmpl, err := template.New("jabline").Parse(tmplStr.Value)
	if err != nil {
		return newError("template parse error: %s", err.Error())
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, goData); err != nil {
		return newError("template execute error: %s", err.Error())
	}

	return &object.String{Value: buf.String()}
}
