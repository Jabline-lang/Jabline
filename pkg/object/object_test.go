package object

import (
	"strings"
	"testing"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestInteger(t *testing.T) {
	obj := &Integer{Value: 42}
	if obj.Type() != INTEGER_OBJ {
		t.Errorf("Type() = %s; want INTEGER", obj.Type())
	}
	if obj.Inspect() != "42" {
		t.Errorf("Inspect() = %q; want 42", obj.Inspect())
	}
	if obj.HashKey().Type != INTEGER_OBJ {
		t.Errorf("HashKey().Type = %s; want INTEGER", obj.HashKey().Type)
	}
}

func TestFloat(t *testing.T) {
	obj := &Float{Value: 3.14}
	if obj.Type() != FLOAT_OBJ {
		t.Errorf("Type() = %s; want FLOAT", obj.Type())
	}
	if obj.Inspect() != "3.14" {
		t.Errorf("Inspect() = %q; want 3.14", obj.Inspect())
	}
}

func TestBoolean(t *testing.T) {
	trueObj := &Boolean{Value: true}
	falseObj := &Boolean{Value: false}
	if trueObj.Type() != BOOLEAN_OBJ {
		t.Errorf("Type() = %s; want BOOLEAN", trueObj.Type())
	}
	if trueObj.Inspect() != "true" {
		t.Errorf("true.Inspect() = %q; want true", trueObj.Inspect())
	}
	if falseObj.Inspect() != "false" {
		t.Errorf("false.Inspect() = %q; want false", falseObj.Inspect())
	}
	if trueObj.HashKey().Value != 1 || falseObj.HashKey().Value != 0 {
		t.Errorf("Boolean HashKey values wrong: true=%d false=%d", trueObj.HashKey().Value, falseObj.HashKey().Value)
	}
}

func TestString(t *testing.T) {
	obj := &String{Value: "hello"}
	if obj.Type() != STRING_OBJ {
		t.Errorf("Type() = %s; want STRING", obj.Type())
	}
	if obj.Inspect() != "hello" {
		t.Errorf("Inspect() = %q; want hello", obj.Inspect())
	}
	hk := obj.HashKey()
	if hk.Type != STRING_OBJ {
		t.Errorf("HashKey().Type = %s; want STRING", hk.Type)
	}
	// Same string should produce same hash key
	obj2 := &String{Value: "hello"}
	if obj.HashKey() != obj2.HashKey() {
		t.Errorf("same strings should have same HashKey")
	}
	// Different strings should (likely) produce different hash keys
	obj3 := &String{Value: "world"}
	if obj.HashKey() == obj3.HashKey() {
		t.Errorf("different strings should have different HashKeys (very unlikely collision)")
	}
}

func TestNull(t *testing.T) {
	obj := &Null{}
	if obj.Type() != NULL_OBJ {
		t.Errorf("Type() = %s; want NULL", obj.Type())
	}
	if obj.Inspect() != "null" {
		t.Errorf("Inspect() = %q; want null", obj.Inspect())
	}
}

func TestNumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		obj      Object
		wantType ObjectType
		inspect  string
	}{
		{"Int8", &Int8{Value: 42}, INT8_OBJ, "42"},
		{"Int16", &Int16{Value: 42}, INT16_OBJ, "42"},
		{"Int32", &Int32{Value: 42}, INT32_OBJ, "42"},
		{"Int64", &Int64{Value: 42}, INT64_OBJ, "42"},
		{"UInt8", &UInt8{Value: 42}, UINT8_OBJ, "42"},
		{"UInt16", &UInt16{Value: 42}, UINT16_OBJ, "42"},
		{"UInt32", &UInt32{Value: 42}, UINT32_OBJ, "42"},
		{"UInt64", &UInt64{Value: 42}, UINT64_OBJ, "42"},
		{"Float32", &Float32{Value: 3.5}, FLOAT32_OBJ, "3.5"},
		{"Float64", &Float64{Value: 3.5}, FLOAT64_OBJ, "3.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.obj.Type() != tt.wantType {
				t.Errorf("Type() = %s; want %s", tt.obj.Type(), tt.wantType)
			}
			if tt.obj.Inspect() != tt.inspect {
				t.Errorf("Inspect() = %q; want %q", tt.obj.Inspect(), tt.inspect)
			}
			// All numeric types should implement HashKey
			if h, ok := tt.obj.(Hashable); !ok {
				t.Errorf("%s does not implement Hashable", tt.wantType)
			} else {
				h.HashKey()
			}
		})
	}
}

func TestArray(t *testing.T) {
	obj := &Array{Elements: []Object{&Integer{Value: 1}, &Integer{Value: 2}}}
	if obj.Type() != ARRAY_OBJ {
		t.Errorf("Type() = %s; want ARRAY", obj.Type())
	}
	if obj.Inspect() != "[1, 2]" {
		t.Errorf("Inspect() = %q; want [1, 2]", obj.Inspect())
	}
	// Empty array
	empty := &Array{Elements: []Object{}}
	if empty.Inspect() != "[]" {
		t.Errorf("empty.Inspec() = %q; want []", empty.Inspect())
	}
}

func TestHash(t *testing.T) {
	pairs := map[HashKey]HashPair{
		(&String{Value: "a"}).HashKey(): {Key: &String{Value: "a"}, Value: &Integer{Value: 1}},
	}
	obj := &Hash{Pairs: pairs}
	if obj.Type() != HASH_OBJ {
		t.Errorf("Type() = %s; want HASH", obj.Type())
	}
	// Should contain "a: 1"
	ins := obj.Inspect()
	if ins != "{a: 1}" {
		t.Errorf("Inspect() = %q; want {a: 1}", ins)
	}
}

func TestHashCycleDetection(t *testing.T) {
	self := &Hash{Pairs: make(map[HashKey]HashPair)}
	key := (&String{Value: "self"}).HashKey()
	self.Pairs[key] = HashPair{Key: &String{Value: "self"}, Value: self}
	ins := self.Inspect()
	if ins != "{self: { <cycle> }}" {
		t.Errorf("cycle Inspect() = %q; want {self: { <cycle> }}", ins)
	}
}

func TestPanic(t *testing.T) {
	obj := &Panic{Message: "something went wrong"}
	if obj.Type() != PANIC_OBJ {
		t.Errorf("Type() = %s; want PANIC", obj.Type())
	}
	if obj.Inspect() != "panic: something went wrong" {
		t.Errorf("Inspect() = %q; want panic: something went wrong", obj.Inspect())
	}
}

func TestError(t *testing.T) {
	obj := &Error{Message: "fail"}
	if obj.Type() != ERROR_OBJ {
		t.Errorf("Type() = %s; want ERROR", obj.Type())
	}
	if obj.Inspect() != "ERROR: fail" {
		t.Errorf("Inspect() = %q; want ERROR: fail", obj.Inspect())
	}
}

func TestReturnValue(t *testing.T) {
	obj := &ReturnValue{Value: &Integer{Value: 42}}
	if obj.Type() != RETURN_OBJ {
		t.Errorf("Type() = %s; want RETURN", obj.Type())
	}
	if obj.Inspect() != "42" {
		t.Errorf("Inspect() = %q; want 42", obj.Inspect())
	}
}

func TestBreakContinue(t *testing.T) {
	b := &Break{}
	if b.Type() != BREAK_OBJ {
		t.Errorf("Break.Type() = %s; want BREAK", b.Type())
	}
	if b.Inspect() != "break" {
		t.Errorf("Break.Inspect() = %q; want break", b.Inspect())
	}
	c := &Continue{}
	if c.Type() != CONTINUE_OBJ {
		t.Errorf("Continue.Type() = %s; want CONTINUE", c.Type())
	}
	if c.Inspect() != "continue" {
		t.Errorf("Continue.Inspect() = %q; want continue", c.Inspect())
	}
}

func TestException(t *testing.T) {
	obj := &Exception{Message: "oops", Value: &String{Value: "detail"}}
	if obj.Type() != EXCEPTION_OBJ {
		t.Errorf("Type() = %s; want EXCEPTION", obj.Type())
	}
	_ = obj.Inspect() // just ensure no panic

	// Exception with nil value should use Message
	obj2 := &Exception{Message: "just message"}
	ins := obj2.Inspect()
	if ins != "EXCEPTION: just message" {
		t.Errorf("Inspect() = %q; want EXCEPTION: just message", ins)
	}
}

func TestBuiltin(t *testing.T) {
	fn := func(args ...Object) Object { return &Null{} }
	obj := &Builtin{Fn: fn, Name: "test"}
	if obj.Type() != BUILTIN_OBJ {
		t.Errorf("Type() = %s; want BUILTIN", obj.Type())
	}
	if obj.Inspect() != "builtin function" {
		t.Errorf("Inspect() = %q; want builtin function", obj.Inspect())
	}
}

func TestClosure(t *testing.T) {
	obj := &Closure{Fn: &CompiledFunction{Instructions: []byte{0x00}}}
	if obj.Type() != CLOSURE_OBJ {
		t.Errorf("Type() = %s; want CLOSURE", obj.Type())
	}
	ins := obj.Inspect()
	if ins[:7] != "Closure" {
		t.Errorf("Inspect() = %q; want Closure[0x...]", ins)
	}
}

func TestCompiledFunction(t *testing.T) {
	cf := &CompiledFunction{
		Instructions:  []byte{0x00, 0x01},
		NumLocals:     2,
		NumParameters: 1,
		IsAsync:       false,
		Name:          "testFn",
	}
	if cf.Type() != COMPILED_FUNCTION_OBJ {
		t.Errorf("Type() = %s; want COMPILED_FUNCTION_OBJ", cf.Type())
	}
	ins := cf.Inspect()
	if ins != "CompiledFunction[0x" && !contains(ins, "testFn") {
		t.Errorf("Inspect() = %q; want CompiledFunction[0x..., testFn]", ins)
	}
}

func TestPromise(t *testing.T) {
	t.Run("new promise is pending", func(t *testing.T) {
		p := NewPromise()
		if p.State != PENDING {
			t.Errorf("State = %d; want PENDING", p.State)
		}
		if p.Inspect() != "Promise { <pending> }" {
			t.Errorf("Inspect() = %q; want Promise { <pending> }", p.Inspect())
		}
	})

	t.Run("resolve promise", func(t *testing.T) {
		p := NewPromise()
		var resolved Object
		p.Then(func(v Object) { resolved = v }, nil)
		p.Resolve(&Integer{Value: 42})
		if p.State != RESOLVED {
			t.Errorf("State = %d; want RESOLVED", p.State)
		}
		if resolved == nil || resolved.(*Integer).Value != 42 {
			t.Errorf("resolved value = %v; want 42", resolved)
		}
	})

	t.Run("reject promise", func(t *testing.T) {
		p := NewPromise()
		p.Then(nil, func(r Object) {
			if r.(*String).Value != "error" {
				t.Errorf("rejected value = %v; want error", r)
			}
		})
		p.Reject(&String{Value: "error"})
		if p.State != REJECTED {
			t.Errorf("State = %d; want REJECTED", p.State)
		}
	})

	t.Run("then on already resolved promise", func(t *testing.T) {
		p := NewResolvedPromise(&Integer{Value: 99})
		var resolved Object
		p.Then(func(v Object) { resolved = v }, nil)
		if resolved == nil || resolved.(*Integer).Value != 99 {
			t.Errorf("resolved value = %v; want 99", resolved)
		}
	})

	t.Run("new resolved promise", func(t *testing.T) {
		p := NewResolvedPromise(&Integer{Value: 5})
		if p.State != RESOLVED {
			t.Errorf("State = %d; want RESOLVED", p.State)
		}
	})

	t.Run("new rejected promise", func(t *testing.T) {
		p := NewRejectedPromise(&String{Value: "fail"})
		if p.State != REJECTED {
			t.Errorf("State = %d; want REJECTED", p.State)
		}
	})
}

func TestBoundMethod(t *testing.T) {
	receiver := &Integer{Value: 1}
	closure := &Closure{Fn: &CompiledFunction{}}
	obj := &BoundMethod{Receiver: receiver, Function: closure}
	if obj.Type() != BOUND_METHOD_OBJ {
		t.Errorf("Type() = %s; want BOUND_METHOD", obj.Type())
	}
}

func TestChannel(t *testing.T) {
	ch := make(chan Object, 1)
	obj := &Channel{Value: ch}
	if obj.Type() != CHANNEL_OBJ {
		t.Errorf("Type() = %s; want CHANNEL", obj.Type())
	}
	ins := obj.Inspect()
	if ins[:7] != "Channel" {
		t.Errorf("Inspect() = %q; want Channel[0x...]", ins)
	}
	// Send/Receive should work
	obj.Value <- &Integer{Value: 42}
	result := <-obj.Value
	if result.(*Integer).Value != 42 {
		t.Errorf("channel receive = %d; want 42", result.(*Integer).Value)
	}
}

// ---------------------------------------------------------------------------
// Environment tests
// ---------------------------------------------------------------------------

func TestEnvironment(t *testing.T) {
	env := NewEnvironment()
	if env == nil {
		t.Fatal("NewEnvironment returned nil")
	}

	// Set and Get
	env.Set("x", &Integer{Value: 1})
	if obj, ok := env.Get("x"); !ok || obj.(*Integer).Value != 1 {
		t.Errorf("Get(x) = %v; want 1", obj)
	}
	if _, ok := env.Get("undefined"); ok {
		t.Errorf("Get(undefined) should be false")
	}
}

func TestEnclosedEnvironment(t *testing.T) {
	outer := NewEnvironment()
	outer.Set("x", &Integer{Value: 1})

	inner := NewEnclosedEnvironment(outer)
	inner.Set("y", &Integer{Value: 2})

	// Inner sees both
	if obj, ok := inner.Get("x"); !ok || obj.(*Integer).Value != 1 {
		t.Errorf("inner.Get(x) = %v; want 1", obj)
	}
	if obj, ok := inner.Get("y"); !ok || obj.(*Integer).Value != 2 {
		t.Errorf("inner.Get(y) = %v; want 2", obj)
	}
	// Outer does NOT see inner
	if _, ok := outer.Get("y"); ok {
		t.Errorf("outer.Get(y) should be false (scope)")
	}

	// Shadowing - Environment.Set() delegates to outer if name exists there
	// (mutable closure semantics)
	inner.Set("x", &Integer{Value: 99})
	if obj, _ := inner.Get("x"); obj.(*Integer).Value != 99 {
		t.Errorf("shadowed inner.Get(x) = %v; want 99", obj)
	}
	// The outer IS updated because Set() delegates to the enclosing scope
	// when the name is found there (closure mutation semantics)
	if obj, _ := outer.Get("x"); obj.(*Integer).Value != 99 {
		t.Errorf("outer.Get(x) after delegated set = %v; want 99", obj)
	}
}

func TestEnvironmentSetConstant(t *testing.T) {
	env := NewEnvironment()
	env.SetConstant("PI", &Float{Value: 3.14})

	if obj, ok := env.Get("PI"); !ok || obj.(*Float).Value != 3.14 {
		t.Errorf("Get(PI) = %v; want 3.14", obj)
	}

	// Cannot reassign constant
	prev := env.Set("PI", &Float{Value: 2.0})
	if prev != nil {
		t.Errorf("Set on constant should return nil, got %v", prev)
	}
	// Value should remain unchanged
	if obj, _ := env.Get("PI"); obj.(*Float).Value != 3.14 {
		t.Errorf("constant value changed to %v", obj)
	}
}

func TestEnvironmentClosureCapture(t *testing.T) {
	outer := NewEnvironment()
	outer.Set("x", &Integer{Value: 10})

	captured := outer.CreateClosureCapture([]string{"x"})
	if v, ok := captured["x"]; !ok || v.(*Integer).Value != 10 {
		t.Errorf("captured[x] = %v; want 10", v)
	}

	// Closure environment
	closureEnv := NewClosureEnvironment(outer)
	closureEnv.ApplyClosureCapture(captured)

	if obj, ok := closureEnv.Get("x"); !ok || obj.(*Integer).Value != 10 {
		t.Errorf("closureEnv.Get(x) = %v; want 10", obj)
	}
}

func TestEnvironmentClone(t *testing.T) {
	env := NewEnvironment()
	env.Set("a", &Integer{Value: 1})
	env.Set("b", &String{Value: "hello"})

	clone := env.Clone()
	if clone == env {
		t.Error("Clone returned same pointer")
	}

	// Clone should have same values
	if obj, ok := clone.Get("a"); !ok || obj.(*Integer).Value != 1 {
		t.Errorf("clone.Get(a) = %v; want 1", obj)
	}
	if obj, ok := clone.Get("b"); !ok || obj.(*String).Value != "hello" {
		t.Errorf("clone.Get(b) = %v; want hello", obj)
	}

	// Mutating clone should not affect original
	clone.Set("a", &Integer{Value: 99})
	if obj, _ := env.Get("a"); obj.(*Integer).Value != 1 {
		t.Errorf("original a changed to %v after clone mutation", obj)
	}
}

func TestEnvironmentGetDepth(t *testing.T) {
	root := NewEnvironment()
	if root.GetDepth() != 0 {
		t.Errorf("root depth = %d; want 0", root.GetDepth())
	}
	l1 := NewEnclosedEnvironment(root)
	if l1.GetDepth() != 1 {
		t.Errorf("l1 depth = %d; want 1", l1.GetDepth())
	}
	l2 := NewEnclosedEnvironment(l1)
	if l2.GetDepth() != 2 {
		t.Errorf("l2 depth = %d; want 2", l2.GetDepth())
	}
}

func TestEnvironmentIsNestedIn(t *testing.T) {
	root := NewEnvironment()
	l1 := NewEnclosedEnvironment(root)
	l2 := NewEnclosedEnvironment(l1)

	if !l2.IsNestedIn(root) {
		t.Errorf("l2 should be nested in root")
	}
	if !l2.IsNestedIn(l1) {
		t.Errorf("l2 should be nested in l1")
	}
	if l2.IsNestedIn(l2) {
		t.Errorf("l2 should not be nested in itself")
	}
}

func TestEnvironmentFindVariable(t *testing.T) {
	root := NewEnvironment()
	root.Set("x", &Integer{Value: 1})
	l1 := NewEnclosedEnvironment(root)
	l1.Set("y", &Integer{Value: 2})
	l2 := NewEnclosedEnvironment(l1)

	found := l2.FindVariableEnvironment("x")
	if found != root {
		t.Errorf("FindVariableEnvironment(x) should return root")
	}
	found = l2.FindVariableEnvironment("y")
	if found != l1 {
		t.Errorf("FindVariableEnvironment(y) should return l1")
	}
	found = l2.FindVariableEnvironment("z")
	if found != nil {
		t.Errorf("FindVariableEnvironment(z) should return nil")
	}
}

func TestEnvironmentIsConstant(t *testing.T) {
	env := NewEnvironment()
	env.SetConstant("C", &Integer{Value: 42})
	env.Set("v", &Integer{Value: 1})

	if !env.IsConstant("C") {
		t.Errorf("IsConstant(C) should be true")
	}
	if env.IsConstant("v") {
		t.Errorf("IsConstant(v) should be false")
	}
}

func TestEnvironmentGetAll(t *testing.T) {
	env := NewEnvironment()
	env.Set("x", &Integer{Value: 1})
	env.SetConstant("C", &Integer{Value: 42})

	all := env.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() should have 2 entries, got %d", len(all))
	}
}

func TestEnvironmentModulePath(t *testing.T) {
	root := NewEnvironment()
	root.ModulePath = "/foo"
	inner := NewEnclosedEnvironment(root)

	if inner.ModulePath != "/foo" {
		t.Errorf("inner.ModulePath = %q; want /foo", inner.ModulePath)
	}
}

// ---------------------------------------------------------------------------
// Pool tests
// ---------------------------------------------------------------------------

func TestPoolNewInteger(t *testing.T) {
	obj := NewInteger(42)
	if obj.Value != 42 {
		t.Errorf("Value = %d; want 42", obj.Value)
	}
	ReleaseInteger(obj)
}

func TestPoolNewString(t *testing.T) {
	obj := NewString("hello")
	if obj.Value != "hello" {
		t.Errorf("Value = %q; want hello", obj.Value)
	}
	ReleaseString(obj)
}

func TestPoolNewBoolean(t *testing.T) {
	trueObj := NewBoolean(true)
	if trueObj != TrueObj {
		t.Errorf("NewBoolean(true) should return singleton TrueObj")
	}
	falseObj := NewBoolean(false)
	if falseObj != FalseObj {
		t.Errorf("NewBoolean(false) should return singleton FalseObj")
	}
}

func TestPoolReleaseObject(t *testing.T) {
	// Should not panic for any type
	ReleaseObject(&Integer{Value: 1})
	ReleaseObject(&String{Value: "hi"})
	ReleaseObject(&Null{})
	ReleaseObject(&Boolean{Value: true})
	ReleaseObject(&Array{})
	ReleaseObject(&Hash{})
}

// ---------------------------------------------------------------------------
// NativeToObject / ObjectToNative tests
// ---------------------------------------------------------------------------

func TestObjectToNative(t *testing.T) {
	tests := []struct {
		name string
		obj  Object
		want interface{}
	}{
		{"integer", &Integer{Value: 42}, int64(42)},
		{"string", &String{Value: "hi"}, "hi"},
		{"boolean true", &Boolean{Value: true}, true},
		{"boolean false", &Boolean{Value: false}, false},
		{"float", &Float{Value: 3.14}, float64(3.14)},
		{"null", &Null{}, nil},
		{"array", &Array{Elements: []Object{&Integer{Value: 1}}}, []interface{}{int64(1)}},
		{"hash", &Hash{Pairs: map[HashKey]HashPair{
			(&String{Value: "k"}).HashKey(): {Key: &String{Value: "k"}, Value: &Integer{Value: 2}},
		}}, map[string]interface{}{"k": int64(2)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ObjectToNative(tt.obj)
			if !equalNative(got, tt.want) {
				t.Errorf("ObjectToNative(%s) = %v (%T); want %v (%T)", tt.obj.Inspect(), got, got, tt.want, tt.want)
			}
		})
	}
}

func TestNativeToObject(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
		want Object
	}{
		{"float64 to integer", float64(42), &Integer{Value: 42}},
		{"string", "hello", &String{Value: "hello"}},
		{"bool true", true, &Boolean{Value: true}},
		{"nil", nil, &Null{}},
		{"[]interface{}", []interface{}{float64(1)}, &Array{Elements: []Object{&Integer{Value: 1}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NativeToObject(tt.val)
			if got.Inspect() != tt.want.Inspect() || got.Type() != tt.want.Type() {
				t.Errorf("NativeToObject(%v) = %s (%s); want %s (%s)", tt.val, got.Inspect(), got.Type(), tt.want.Inspect(), tt.want.Type())
			}
		})
	}
}

func equalNative(a, b interface{}) bool {
	switch a := a.(type) {
	case int64:
		if b, ok := b.(int64); ok {
			return a == b
		}
	case string:
		if b, ok := b.(string); ok {
			return a == b
		}
	case bool:
		if b, ok := b.(bool); ok {
			return a == b
		}
	case float64:
		if b, ok := b.(float64); ok {
			return a == b
		}
	case nil:
		return b == nil
	case []interface{}:
		if b, ok := b.([]interface{}); ok {
			if len(a) != len(b) {
				return false
			}
			for i := range a {
				if !equalNative(a[i], b[i]) {
					return false
				}
			}
			return true
		}
	case map[string]interface{}:
		if b, ok := b.(map[string]interface{}); ok {
			if len(a) != len(b) {
				return false
			}
			for k, v := range a {
				if !equalNative(v, b[k]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Service tests
// ---------------------------------------------------------------------------

func TestService(t *testing.T) {
	obj := &Service{
		Name:   "MyService",
		Config: map[string]Object{"port": &Integer{Value: 8080}},
	}
	if obj.Type() != SERVICE_OBJ {
		t.Errorf("Type() = %s; want SERVICE", obj.Type())
	}
	ins := obj.Inspect()
	if ins[:17] != "service MyService" {
		t.Errorf("Inspect() = %q; want service MyService {...}", ins)
	}
}

// ---------------------------------------------------------------------------
// Database tests
// ---------------------------------------------------------------------------

func TestDatabase(t *testing.T) {
	obj := &Database{DB: nil}
	if obj.Type() != DATABASE_OBJ {
		t.Errorf("Type() = %s; want DATABASE", obj.Type())
	}
	if obj.Inspect() != "<DatabaseConnection (Closed)>" {
		t.Errorf("Inspect() = %q; want <DatabaseConnection (Closed)>", obj.Inspect())
	}
}

// ---------------------------------------------------------------------------
// RemoteChannel tests
// ---------------------------------------------------------------------------

func TestRemoteChannelType(t *testing.T) {
// Can't easily test RemoteChannel without net.Conn, but test type
if _, ok := interface{}(&RemoteChannel{}).(Object); !ok {
	t.Errorf("RemoteChannel does not implement Object")
}
}

// ---------------------------------------------------------------------------
// Struct / Instance tests
// ---------------------------------------------------------------------------

func TestStruct(t *testing.T) {
	obj := &Struct{
		Name:   "Point",
		Fields: map[string]string{"x": "int", "y": "int"},
	}
	if obj.Type() != STRUCT_OBJ {
		t.Errorf("Type() = %s; want STRUCT", obj.Type())
	}
	ins := obj.Inspect()
	if !contains(ins, "struct Point") {
		t.Errorf("Inspect() = %q; want struct Point {\\n  ...}", ins)
	}
}

func TestInstance(t *testing.T) {
	obj := &Instance{
		StructName: "Point",
		Fields:     map[string]Object{"x": &Integer{Value: 10}},
	}
	if obj.Type() != INSTANCE_OBJ {
		t.Errorf("Type() = %s; want INSTANCE", obj.Type())
	}
	ins := obj.Inspect()
	if !contains(ins, "Point") || !contains(ins, "10") {
		t.Errorf("Inspect() = %q; want Point {\\n  x: 10...}", ins)
	}
}

func TestInterface(t *testing.T) {
	obj := &Interface{
		Name: "Drawable",
		Methods: map[string]*InterfaceMethod{
			"draw": {Name: "draw", Parameters: []string{}, ReturnType: "void"},
		},
	}
	if obj.Type() != INTERFACE_OBJ {
		t.Errorf("Type() = %s; want INTERFACE", obj.Type())
	}
}

func TestInstantiatedStruct(t *testing.T) {
	obj := &InstantiatedStruct{
		Struct:       &Struct{Name: "Array"},
		FullTypeName: "Array<int>",
	}
	if obj.Type() != INSTANTIATED_STRUCT_OBJ {
		t.Errorf("Type() = %s; want INSTANTIATED_STRUCT", obj.Type())
	}
	if obj.Inspect() != "instantiated struct Array<int>" {
		t.Errorf("Inspect() = %q; want instantiated struct Array<int>", obj.Inspect())
	}
}

// ---------------------------------------------------------------------------
// WebSocket tests (basic, no real connection)
// ---------------------------------------------------------------------------

func TestWebSocket(t *testing.T) {
	obj := &WebSocket{Conn: nil}
	if obj.Type() != WEBSOCKET_OBJ {
		t.Errorf("Type() = %s; want WEBSOCKET", obj.Type())
	}
	if obj.Inspect() != "WebSocket(closed)" {
		t.Errorf("Inspect() = %q; want WebSocket(closed)", obj.Inspect())
	}
}

func TestInstantiatedFunction(t *testing.T) {
	obj := &InstantiatedFunction{
		Closure:      &Closure{Fn: &CompiledFunction{}},
		TypeArgs:     map[string]string{"T": "int"},
		FullTypeName: "Array<int>",
	}
	if obj.Type() != INSTANTIATED_FUNCTION_OBJ {
		t.Errorf("Type() = %s; want INSTANTIATED_FUNCTION", obj.Type())
	}
	if obj.Inspect() != "instantiated fn Array<int>" {
		t.Errorf("Inspect() = %q; want instantiated fn Array<int>", obj.Inspect())
	}
}
