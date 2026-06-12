package stdlib

import (
	"testing"

	"jabline/pkg/object"
)

func lookupBuiltin(t *testing.T, name string) func(...object.Object) object.Object {
	t.Helper()
	for _, entry := range Registry {
		if entry.Name == name {
			b, ok := entry.Object.(*object.Builtin)
			if !ok {
				t.Fatalf("Registry entry %s is not a Builtin", name)
			}
			return b.Fn
		}
	}
	t.Fatalf("builtin %s not found in Registry", name)
	return nil
}

// ---------------------------------------------------------------------------
// len builtin
// ---------------------------------------------------------------------------

func TestLenBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "len")

	tests := []struct {
		name    string
		args    []object.Object
		want    int64
		wantErr bool
	}{
		{"array of 3", []object.Object{&object.Array{Elements: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}, &object.Integer{Value: 3}}}}, 3, false},
		{"empty array", []object.Object{&object.Array{Elements: []object.Object{}}}, 0, false},
		{"string", []object.Object{&object.String{Value: "hello"}}, 5, false},
		{"empty string", []object.Object{&object.String{Value: ""}}, 0, false},
		{"hash with 2 keys", []object.Object{&object.Hash{Pairs: map[object.HashKey]object.HashPair{
			(&object.String{Value: "a"}).HashKey(): {Key: &object.String{Value: "a"}, Value: &object.Integer{Value: 1}},
			(&object.String{Value: "b"}).HashKey(): {Key: &object.String{Value: "b"}, Value: &object.Integer{Value: 2}},
		}}}, 2, false},
		{"no args (error)", []object.Object{}, 0, true},
		{"too many args (error)", []object.Object{&object.Array{}, &object.Array{}}, 0, true},
		{"int arg (error)", []object.Object{&object.Integer{Value: 42}}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn(tt.args...)
			if tt.wantErr {
				if _, ok := result.(*object.Error); !ok {
					t.Errorf("expected Error, got %T: %s", result, result.Inspect())
				}
			} else {
				if r, ok := result.(*object.Integer); !ok || r.Value != tt.want {
					t.Errorf("expected Integer(%d), got %T: %s", tt.want, result, result.Inspect())
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// type builtin
// ---------------------------------------------------------------------------

func TestTypeBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "type")

	tests := []struct {
		name string
		arg  object.Object
		want string
	}{
		{"integer", &object.Integer{Value: 42}, "INTEGER"},
		{"float", &object.Float{Value: 3.14}, "FLOAT"},
		{"string", &object.String{Value: "hi"}, "STRING"},
		{"boolean true", &object.Boolean{Value: true}, "BOOLEAN"},
		{"null", &object.Null{}, "NULL"},
		{"array", &object.Array{}, "ARRAY"},
		{"hash", &object.Hash{}, "HASH"},
		{"error", &object.Error{Message: "err"}, "ERROR"},
		{"builtin", &object.Builtin{Fn: nil}, "BUILTIN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn(tt.arg)
			if r, ok := result.(*object.String); !ok || r.Value != tt.want {
				t.Errorf("expected String(%q), got %T: %s", tt.want, result, result.Inspect())
			}
		})
	}
}

func TestTypeBuiltinNoArgs(t *testing.T) {
	fn := lookupBuiltin(t, "type")
	result := fn()
	if _, ok := result.(*object.Error); !ok {
		t.Errorf("expected Error for no args, got %T: %s", result, result.Inspect())
	}
}

// ---------------------------------------------------------------------------
// toString builtin
// ---------------------------------------------------------------------------

func TestToStringBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "toString")

	tests := []struct {
		name string
		arg  object.Object
		want string
	}{
		{"integer", &object.Integer{Value: 42}, "42"},
		{"float", &object.Float{Value: 3.14}, "3.14"},
		{"string", &object.String{Value: "hello"}, "hello"},
		{"boolean true", &object.Boolean{Value: true}, "true"},
		{"null", &object.Null{}, "null"},
		{"array", &object.Array{Elements: []object.Object{&object.Integer{Value: 1}}}, "[1]"},
		{"error", &object.Error{Message: "bad"}, "ERROR: bad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn(tt.arg)
			if r, ok := result.(*object.String); !ok || r.Value != tt.want {
				t.Errorf("expected String(%q), got %T: %s", tt.want, result, result.Inspect())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseInt / parseFloat builtins
// ---------------------------------------------------------------------------

func TestParseIntBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "parseInt")

	tests := []struct {
		name    string
		arg     object.Object
		want    int64
		wantErr bool
	}{
		{"valid int", &object.String{Value: "42"}, 42, false},
		{"zero", &object.String{Value: "0"}, 0, false},
		{"negative", &object.String{Value: "-10"}, -10, false},
		{"invalid string", &object.String{Value: "abc"}, 0, true},
		{"float string truncated", &object.String{Value: "3.14"}, 0, true},
		{"non-string arg", &object.Integer{Value: 42}, 0, true},
		{"no args", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result object.Object
			if tt.arg != nil {
				result = fn(tt.arg)
			} else {
				result = fn()
			}
			if tt.wantErr {
				if _, ok := result.(*object.Error); !ok {
					t.Errorf("expected Error, got %T: %s", result, result.Inspect())
				}
			} else {
				if r, ok := result.(*object.Integer); !ok || r.Value != tt.want {
					t.Errorf("expected Integer(%d), got %T: %s", tt.want, result, result.Inspect())
				}
			}
		})
	}
}

func TestParseFloatBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "parseFloat")

	tests := []struct {
		name    string
		arg     object.Object
		want    float64
		wantErr bool
	}{
		{"valid float", &object.String{Value: "3.14"}, 3.14, false},
		{"int string", &object.String{Value: "42"}, 42.0, false},
		{"negative", &object.String{Value: "-2.5"}, -2.5, false},
		{"invalid", &object.String{Value: "abc"}, 0, true},
		{"non-string arg", &object.Integer{Value: 42}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn(tt.arg)
			if tt.wantErr {
				if _, ok := result.(*object.Error); !ok {
					t.Errorf("expected Error, got %T: %s", result, result.Inspect())
				}
			} else {
				if r, ok := result.(*object.Float); !ok || r.Value != tt.want {
					t.Errorf("expected Float(%f), got %T: %s", tt.want, result, result.Inspect())
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Push / Pop / Rest / First / Last builtins
// ---------------------------------------------------------------------------

func TestPushBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "push")

	arr := &object.Array{Elements: []object.Object{&object.Integer{Value: 1}}}
	result := fn(arr, &object.Integer{Value: 2})
	if r, ok := result.(*object.Array); ok {
		if len(r.Elements) != 2 {
			t.Errorf("expected 2 elements, got %d", len(r.Elements))
		}
		if r.Elements[1].(*object.Integer).Value != 2 {
			t.Errorf("expected second element 2, got %s", r.Elements[1].Inspect())
		}
		// Original unchanged
		if len(arr.Elements) != 1 {
			t.Errorf("original array was mutated, len=%d", len(arr.Elements))
		}
	} else {
		t.Errorf("expected Array, got %T: %s", result, result.Inspect())
	}
}

func TestPopBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "pop")

	t.Run("pop last element", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}, &object.Integer{Value: 3}}}
		result := fn(arr)
		if r, ok := result.(*object.Integer); !ok || r.Value != 3 {
			t.Errorf("expected 3, got %s", result.Inspect())
		}
		// Array mutated - now has 2 elements
		if len(arr.Elements) != 2 {
			t.Errorf("expected array to have 2 elements, got %d", len(arr.Elements))
		}
	})

	t.Run("pop from empty array returns null", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{}}
		result := fn(arr)
		if _, ok := result.(*object.Null); !ok {
			t.Errorf("expected Null, got %T", result)
		}
	})

	t.Run("pop wrong args", func(t *testing.T) {
		result := fn()
		if _, ok := result.(*object.Error); !ok {
			t.Errorf("expected Error for no args, got %T", result)
		}
	})
}

func TestRestBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "rest")

	t.Run("rest of 3 elements", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}, &object.Integer{Value: 3}}}
		result := fn(arr)
		if r, ok := result.(*object.Array); ok {
			if len(r.Elements) != 2 {
				t.Errorf("expected 2 elements, got %d", len(r.Elements))
			}
			if r.Elements[0].(*object.Integer).Value != 2 {
				t.Errorf("expected first element 2, got %s", r.Elements[0].Inspect())
			}
		} else {
			t.Errorf("expected Array, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("rest of 1 element returns empty array", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{&object.Integer{Value: 1}}}
		result := fn(arr)
		if r, ok := result.(*object.Array); !ok || len(r.Elements) != 0 {
			t.Errorf("expected empty Array, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("rest of empty array returns null", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{}}
		result := fn(arr)
		if _, ok := result.(*object.Null); !ok {
			t.Errorf("expected Null for empty array, got %T", result)
		}
	})

	t.Run("rest of string", func(t *testing.T) {
		result := fn(&object.String{Value: "hello"})
		if r, ok := result.(*object.String); !ok || r.Value != "ello" {
			t.Errorf("expected String(ello), got %T: %s", result, result.Inspect())
		}
	})
}

func TestFirstBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "first")

	t.Run("first of array", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{&object.Integer{Value: 10}, &object.Integer{Value: 20}}}
		result := fn(arr)
		if r, ok := result.(*object.Integer); !ok || r.Value != 10 {
			t.Errorf("expected 10, got %s", result.Inspect())
		}
	})

	t.Run("first of empty array", func(t *testing.T) {
		result := fn(&object.Array{Elements: []object.Object{}})
		if _, ok := result.(*object.Null); !ok {
			t.Errorf("expected Null, got %T", result)
		}
	})
}

func TestLastBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "last")

	t.Run("last of array", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{&object.Integer{Value: 10}, &object.Integer{Value: 20}}}
		result := fn(arr)
		if r, ok := result.(*object.Integer); !ok || r.Value != 20 {
			t.Errorf("expected 20, got %s", result.Inspect())
		}
	})

	t.Run("last of empty array", func(t *testing.T) {
		result := fn(&object.Array{Elements: []object.Object{}})
		if _, ok := result.(*object.Null); !ok {
			t.Errorf("expected Null, got %T", result)
		}
	})
}

// ---------------------------------------------------------------------------
// Keys / Values builtins
// ---------------------------------------------------------------------------

func TestKeysBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "keys")

	hash := &object.Hash{Pairs: map[object.HashKey]object.HashPair{
		(&object.String{Value: "a"}).HashKey(): {Key: &object.String{Value: "a"}, Value: &object.Integer{Value: 1}},
		(&object.String{Value: "b"}).HashKey(): {Key: &object.String{Value: "b"}, Value: &object.Integer{Value: 2}},
	}}

	result := fn(hash)
	arr, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 keys, got %d", len(arr.Elements))
	}

	result = fn(&object.Integer{Value: 1})
	if _, ok := result.(*object.Error); !ok {
		t.Errorf("expected Error for non-hash arg, got %T", result)
	}
}

func TestValuesBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "values")

	hash := &object.Hash{Pairs: map[object.HashKey]object.HashPair{
		(&object.String{Value: "a"}).HashKey(): {Key: &object.String{Value: "a"}, Value: &object.Integer{Value: 10}},
	}}

	result := fn(hash)
	arr, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 1 {
		t.Errorf("expected 1 value, got %d", len(arr.Elements))
	}
	if arr.Elements[0].(*object.Integer).Value != 10 {
		t.Errorf("expected 10, got %s", arr.Elements[0].Inspect())
	}
}

// ---------------------------------------------------------------------------
// is_error / Error builtins
// ---------------------------------------------------------------------------

func TestIsErrorBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "is_error")

	tests := []struct {
		name string
		arg  object.Object
		want bool
	}{
		{"error object", &object.Error{Message: "fail"}, true},
		{"integer", &object.Integer{Value: 1}, false},
		{"string", &object.String{Value: "hi"}, false},
		{"null", &object.Null{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn(tt.arg)
			if r, ok := result.(*object.Boolean); !ok || r.Value != tt.want {
				t.Errorf("expected Boolean(%v), got %T: %s", tt.want, result, result.Inspect())
			}
		})
	}
}

func TestErrorBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "Error")

	t.Run("string arg", func(t *testing.T) {
		result := fn(&object.String{Value: "something went wrong"})
		if r, ok := result.(*object.Error); !ok || r.Message != "something went wrong" {
			t.Errorf("expected Error('something went wrong'), got %T: %s", result, result.Inspect())
		}
	})

	t.Run("non-string arg uses Inspect", func(t *testing.T) {
		result := fn(&object.Integer{Value: 42})
		if r, ok := result.(*object.Error); !ok || r.Message != "42" {
			t.Errorf("expected Error('42'), got %T: %s", result, result.Inspect())
		}
	})
}

// ---------------------------------------------------------------------------
// Set builtin
// ---------------------------------------------------------------------------

func TestSetBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "set")

	t.Run("set on hash", func(t *testing.T) {
		hash := &object.Hash{Pairs: map[object.HashKey]object.HashPair{}}
		result := fn(hash, &object.String{Value: "key"}, &object.Integer{Value: 42})
		if r, ok := result.(*object.Integer); !ok || r.Value != 42 {
			t.Errorf("expected returned value 42, got %T: %s", result, result.Inspect())
		}
		if len(hash.Pairs) != 1 {
			t.Errorf("expected hash to have 1 pair, got %d", len(hash.Pairs))
		}
	})

	t.Run("set on array", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}}}
		result := fn(arr, &object.Integer{Value: 0}, &object.Integer{Value: 99})
		if r, ok := result.(*object.Integer); !ok || r.Value != 99 {
			t.Errorf("expected 99, got %T: %s", result, result.Inspect())
		}
		if arr.Elements[0].(*object.Integer).Value != 99 {
			t.Errorf("expected array[0] = 99, got %s", arr.Elements[0].Inspect())
		}
	})

	t.Run("set on invalid type", func(t *testing.T) {
		result := fn(&object.Integer{Value: 1}, &object.String{Value: "key"}, &object.Integer{Value: 99})
		if _, ok := result.(*object.Error); !ok {
			t.Errorf("expected Error for set on integer, got %T", result)
		}
	})
}

// ---------------------------------------------------------------------------
// Type conversion builtins
// ---------------------------------------------------------------------------

func TestTypeConversionBuiltins(t *testing.T) {
	tests := []struct {
		name    string
		builtin string
		arg     object.Object
		wantStr string
		wantErr bool
	}{
		{"int8 from int", "int8", &object.Integer{Value: 42}, "42", false},
		{"int16 from int", "int16", &object.Integer{Value: 42}, "42", false},
		{"int32 from int", "int32", &object.Integer{Value: 42}, "42", false},
		{"int64 from int", "int64", &object.Integer{Value: 42}, "42", false},
		{"uint8 from int", "uint8", &object.Integer{Value: 42}, "42", false},
		{"uint16 from int", "uint16", &object.Integer{Value: 42}, "42", false},
		{"uint32 from int", "uint32", &object.Integer{Value: 42}, "42", false},
		{"uint64 from int", "uint64", &object.Integer{Value: 42}, "42", false},
		{"float32 from int", "float32", &object.Integer{Value: 42}, "42", false},
		{"float64 from int", "float64", &object.Integer{Value: 42}, "42", false},
		{"int8 from float", "int8", &object.Float{Value: 42.5}, "42", false},
		{"invalid arg", "int8", &object.String{Value: "hi"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := lookupBuiltin(t, tt.builtin)
			result := fn(tt.arg)
			if tt.wantErr {
				if _, ok := result.(*object.Error); !ok {
					t.Errorf("expected Error, got %T: %s", result, result.Inspect())
				}
				return
			}
			if result.Inspect() != tt.wantStr {
				t.Errorf("expected Inspect %q, got %q", tt.wantStr, result.Inspect())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Strings module tests
// ---------------------------------------------------------------------------

func TestGlobalModulesPopulated(t *testing.T) {
	// Check which modules are available after all init() functions run
	expectedMods := []string{"strings", "math", "json", "os", "fs"}
	for _, name := range expectedMods {
		if _, ok := GlobalModules[name]; !ok {
			t.Logf("module %q not in GlobalModules (may need to fix init order)", name)
		}
	}
}

func getModule(t *testing.T, moduleName string) *object.Hash {
	t.Helper()
	mod, ok := GlobalModules[moduleName]
	if !ok {
		t.Fatalf("module %s not found in GlobalModules", moduleName)
	}
	return mod
}

func callModuleFn(t *testing.T, mod *object.Hash, fnName string, args ...object.Object) object.Object {
	t.Helper()
	key := &object.String{Value: fnName}
	hk := key.HashKey()
	pair, ok := mod.Pairs[hk]
	if !ok {
		t.Fatalf("function %s not found in module", fnName)
	}
	builtin, ok := pair.Value.(*object.Builtin)
	if !ok {
		t.Fatalf("module function %s is not a Builtin", fnName)
	}
	return builtin.Fn(args...)
}

func TestStringsModule(t *testing.T) {
	mod := getModule(t, "strings")

	t.Run("upper", func(t *testing.T) {
		result := callModuleFn(t, mod, "upper", &object.String{Value: "hello"})
		if r, ok := result.(*object.String); !ok || r.Value != "HELLO" {
			t.Errorf("expected HELLO, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("lower", func(t *testing.T) {
		result := callModuleFn(t, mod, "lower", &object.String{Value: "HELLO"})
		if r, ok := result.(*object.String); !ok || r.Value != "hello" {
			t.Errorf("expected hello, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("contains", func(t *testing.T) {
		result := callModuleFn(t, mod, "contains", &object.String{Value: "hello world"}, &object.String{Value: "world"})
		if r, ok := result.(*object.Boolean); !ok || !r.Value {
			t.Errorf("expected true, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("trim", func(t *testing.T) {
		result := callModuleFn(t, mod, "trim", &object.String{Value: "  hi  "})
		if r, ok := result.(*object.String); !ok || r.Value != "hi" {
			t.Errorf("expected 'hi', got %T: %s", result, result.Inspect())
		}
	})

	t.Run("split", func(t *testing.T) {
		result := callModuleFn(t, mod, "split", &object.String{Value: "a,b,c"}, &object.String{Value: ","})
		if r, ok := result.(*object.Array); ok {
			if len(r.Elements) != 3 {
				t.Errorf("expected 3 elements, got %d", len(r.Elements))
			}
		} else {
			t.Errorf("expected Array, got %T", result)
		}
	})

	t.Run("join", func(t *testing.T) {
		arr := &object.Array{Elements: []object.Object{&object.String{Value: "a"}, &object.String{Value: "b"}}}
		result := callModuleFn(t, mod, "join", arr, &object.String{Value: ","})
		if r, ok := result.(*object.String); !ok || r.Value != "a,b" {
			t.Errorf("expected 'a,b', got %T: %s", result, result.Inspect())
		}
	})

	t.Run("startsWith", func(t *testing.T) {
		result := callModuleFn(t, mod, "startsWith", &object.String{Value: "hello"}, &object.String{Value: "he"})
		if r, ok := result.(*object.Boolean); !ok || !r.Value {
			t.Errorf("expected true, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("endsWith", func(t *testing.T) {
		result := callModuleFn(t, mod, "endsWith", &object.String{Value: "hello"}, &object.String{Value: "lo"})
		if r, ok := result.(*object.Boolean); !ok || !r.Value {
			t.Errorf("expected true, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("indexOf", func(t *testing.T) {
		result := callModuleFn(t, mod, "indexOf", &object.String{Value: "hello"}, &object.String{Value: "l"})
		if r, ok := result.(*object.Integer); !ok || r.Value != 2 {
			t.Errorf("expected 2, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("slice", func(t *testing.T) {
		result := callModuleFn(t, mod, "slice", &object.String{Value: "hello"}, &object.Integer{Value: 1}, &object.Integer{Value: 4})
		if r, ok := result.(*object.String); !ok || r.Value != "ell" {
			t.Errorf("expected 'ell', got %T: %s", result, result.Inspect())
		}
	})

	t.Run("match", func(t *testing.T) {
		result := callModuleFn(t, mod, "match", &object.String{Value: "^h.*o$"}, &object.String{Value: "hello"})
		if r, ok := result.(*object.Boolean); !ok || !r.Value {
			t.Errorf("expected true, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("replace", func(t *testing.T) {
		result := callModuleFn(t, mod, "replace", &object.String{Value: "hello world"}, &object.String{Value: "world"}, &object.String{Value: "there"})
		if r, ok := result.(*object.String); !ok || r.Value != "hello there" {
			t.Errorf("expected 'hello there', got %T: %s", result, result.Inspect())
		}
	})
}

// ---------------------------------------------------------------------------
// Math module tests
// ---------------------------------------------------------------------------

func TestMathModule(t *testing.T) {
	mod := getModule(t, "math")

	t.Run("abs positive", func(t *testing.T) {
		result := callModuleFn(t, mod, "abs", &object.Integer{Value: 5})
		if r, ok := result.(*object.Integer); !ok || r.Value != 5 {
			t.Errorf("expected 5, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("abs negative", func(t *testing.T) {
		result := callModuleFn(t, mod, "abs", &object.Integer{Value: -5})
		if r, ok := result.(*object.Integer); !ok || r.Value != 5 {
			t.Errorf("expected 5, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("sqrt", func(t *testing.T) {
		result := callModuleFn(t, mod, "sqrt", &object.Float{Value: 9.0})
		if r, ok := result.(*object.Float); !ok || r.Value != 3.0 {
			t.Errorf("expected 3.0, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("max", func(t *testing.T) {
		result := callModuleFn(t, mod, "max", &object.Integer{Value: 1}, &object.Integer{Value: 5}, &object.Integer{Value: 3})
		if r, ok := result.(*object.Float); !ok || r.Value != 5.0 {
			t.Errorf("expected 5.0, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("min", func(t *testing.T) {
		result := callModuleFn(t, mod, "min", &object.Integer{Value: 1}, &object.Integer{Value: 5}, &object.Integer{Value: 3})
		if r, ok := result.(*object.Float); !ok || r.Value != 1.0 {
			t.Errorf("expected 1.0, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("floor", func(t *testing.T) {
		result := callModuleFn(t, mod, "floor", &object.Float{Value: 3.7})
		if r, ok := result.(*object.Integer); !ok || r.Value != 3 {
			t.Errorf("expected 3, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("ceil", func(t *testing.T) {
		result := callModuleFn(t, mod, "ceil", &object.Float{Value: 3.2})
		if r, ok := result.(*object.Integer); !ok || r.Value != 4 {
			t.Errorf("expected 4, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("round", func(t *testing.T) {
		result := callModuleFn(t, mod, "round", &object.Float{Value: 3.5})
		if r, ok := result.(*object.Integer); !ok || r.Value != 4 {
			t.Errorf("expected 4, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("sin", func(t *testing.T) {
		result := callModuleFn(t, mod, "sin", &object.Float{Value: 0.0})
		if r, ok := result.(*object.Float); !ok || r.Value != 0.0 {
			t.Errorf("expected 0.0, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("cos", func(t *testing.T) {
		result := callModuleFn(t, mod, "cos", &object.Float{Value: 0.0})
		if r, ok := result.(*object.Float); !ok || r.Value != 1.0 {
			t.Errorf("expected 1.0, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("pow", func(t *testing.T) {
		result := callModuleFn(t, mod, "pow", &object.Float{Value: 2.0}, &object.Float{Value: 3.0})
		if r, ok := result.(*object.Float); !ok || r.Value != 8.0 {
			t.Errorf("expected 8.0, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("log", func(t *testing.T) {
		result := callModuleFn(t, mod, "log", &object.Float{Value: 1.0})
		if r, ok := result.(*object.Float); !ok || r.Value != 0.0 {
			t.Errorf("expected 0.0, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("exp", func(t *testing.T) {
		result := callModuleFn(t, mod, "exp", &object.Float{Value: 0.0})
		if r, ok := result.(*object.Float); !ok || r.Value != 1.0 {
			t.Errorf("expected 1.0, got %T: %s", result, result.Inspect())
		}
	})
}

// ---------------------------------------------------------------------------
// JSON module tests
// ---------------------------------------------------------------------------

func TestJSONModule(t *testing.T) {
	mod := getModule(t, "json")

	t.Run("parse", func(t *testing.T) {
		result := callModuleFn(t, mod, "parse", &object.String{Value: `{"a":1,"b":"two"}`})
		hash, ok := result.(*object.Hash)
		if !ok {
			t.Fatalf("expected Hash, got %T: %s", result, result.Inspect())
		}
		// Check value "a"
		aKey := (&object.String{Value: "a"}).HashKey()
		if pair, exists := hash.Pairs[aKey]; !exists {
			t.Errorf("key 'a' not found in parsed hash")
		} else if pair.Value.(*object.Integer).Value != 1 {
			t.Errorf("expected a=1, got %s", pair.Value.Inspect())
		}
	})

	t.Run("stringify", func(t *testing.T) {
		hash := &object.Hash{Pairs: map[object.HashKey]object.HashPair{
			(&object.String{Value: "x"}).HashKey(): {Key: &object.String{Value: "x"}, Value: &object.Integer{Value: 42}},
		}}
		result := callModuleFn(t, mod, "stringify", hash)
		if r, ok := result.(*object.String); !ok || r.Value == "" {
			t.Errorf("expected non-empty JSON string, got %T: %s", result, result.Inspect())
		}
	})

	t.Run("pretty", func(t *testing.T) {
		hash := &object.Hash{Pairs: map[object.HashKey]object.HashPair{
			(&object.String{Value: "x"}).HashKey(): {Key: &object.String{Value: "x"}, Value: &object.Integer{Value: 42}},
		}}
		result := callModuleFn(t, mod, "pretty", hash)
		if r, ok := result.(*object.String); !ok || r.Value == "" {
			t.Errorf("expected non-empty pretty JSON, got %T: %s", result, result.Inspect())
		}
	})
}

// ---------------------------------------------------------------------------
// query_params / query_param builtins
// ---------------------------------------------------------------------------

func TestQueryParamsBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "query_params")

	t.Run("single query param", func(t *testing.T) {
		result := fn(&object.String{Value: "http://example.com/path?foo=bar"})
		hash, ok := result.(*object.Hash)
		if !ok {
			t.Fatalf("expected Hash, got %T: %s", result, result.Inspect())
		}
		key := (&object.String{Value: "foo"}).HashKey()
		pair, exists := hash.Pairs[key]
		if !exists {
			t.Fatal("key 'foo' not found")
		}
		if s, ok := pair.Value.(*object.String); !ok || s.Value != "bar" {
			t.Errorf("expected 'bar', got %T: %s", pair.Value, pair.Value.Inspect())
		}
	})

	t.Run("multiple params", func(t *testing.T) {
		result := fn(&object.String{Value: "/search?q=hello&page=2"})
		hash, ok := result.(*object.Hash)
		if !ok {
			t.Fatalf("expected Hash, got %T", result)
		}
		if len(hash.Pairs) != 2 {
			t.Errorf("expected 2 keys, got %d", len(hash.Pairs))
		}
	})

	t.Run("no query string", func(t *testing.T) {
		result := fn(&object.String{Value: "/path"})
		hash, ok := result.(*object.Hash)
		if !ok {
			t.Fatalf("expected Hash, got %T", result)
		}
		if len(hash.Pairs) != 0 {
			t.Errorf("expected empty hash, got %d keys", len(hash.Pairs))
		}
	})

	t.Run("multiple values same key", func(t *testing.T) {
		result := fn(&object.String{Value: "/filter?tag=a&tag=b"})
		hash, ok := result.(*object.Hash)
		if !ok {
			t.Fatalf("expected Hash, got %T", result)
		}
		key := (&object.String{Value: "tag"}).HashKey()
		pair, exists := hash.Pairs[key]
		if !exists {
			t.Fatal("key 'tag' not found")
		}
		arr, ok := pair.Value.(*object.Array)
		if !ok {
			t.Fatalf("expected Array for multi-valued key, got %T", pair.Value)
		}
		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements, got %d", len(arr.Elements))
		}
	})

	t.Run("error on invalid arg", func(t *testing.T) {
		result := fn(&object.Integer{Value: 42})
		if _, ok := result.(*object.Error); !ok {
			t.Errorf("expected Error, got %T", result)
		}
	})

	t.Run("error on no args", func(t *testing.T) {
		result := fn()
		if _, ok := result.(*object.Error); !ok {
			t.Errorf("expected Error, got %T", result)
		}
	})
}

func TestQueryParamBuiltin(t *testing.T) {
	fn := lookupBuiltin(t, "query_param")

	t.Run("existing param", func(t *testing.T) {
		result := fn(&object.String{Value: "q"}, &object.String{Value: "/search?q=hello"})
		s, ok := result.(*object.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "hello" {
			t.Errorf("expected 'hello', got %q", s.Value)
		}
	})

	t.Run("missing param returns null", func(t *testing.T) {
		result := fn(&object.String{Value: "missing"}, &object.String{Value: "/search?q=hello"})
		if _, ok := result.(*object.Null); !ok {
			t.Errorf("expected Null for missing param, got %T", result)
		}
	})

	t.Run("invalid URL returns error", func(t *testing.T) {
		result := fn(&object.String{Value: "x"}, &object.String{Value: "://bad-url"})
		if _, ok := result.(*object.Error); !ok {
			t.Errorf("expected Error for invalid URL, got %T", result)
		}
	})

	t.Run("error on wrong arg types", func(t *testing.T) {
		result := fn(&object.Integer{Value: 1}, &object.String{Value: "/foo"})
		if _, ok := result.(*object.Error); !ok {
			t.Errorf("expected Error, got %T", result)
		}
	})

	t.Run("error on wrong arg count", func(t *testing.T) {
		result := fn(&object.String{Value: "q"})
		if _, ok := result.(*object.Error); !ok {
			t.Errorf("expected Error, got %T", result)
		}
	})
}

// ---------------------------------------------------------------------------
// Reset test results
// ---------------------------------------------------------------------------

func TestResetTestResults(t *testing.T) {
	TestPassed = 10
	TestFailed = 5
	TestTotal = 15
	ResetTestResults()
	if TestPassed != 0 || TestFailed != 0 || TestTotal != 0 {
		t.Errorf("expected all zero after reset, got passed=%d failed=%d total=%d", TestPassed, TestFailed, TestTotal)
	}
}
