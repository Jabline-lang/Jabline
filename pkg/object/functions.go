package object

import (
	"jabline/pkg/ast"
	"strings"
)

type Function struct {
	Parameters       []*ast.Identifier
	Body             *ast.BlockStatement
	Env              *Environment
	CapturedVars     map[string]Object
	IsClosureCreated bool
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	out := "fn("
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out += strings.Join(params, ", ")
	out += ") {\n"
	out += f.Body.String()
	out += "\n}"

	if f.IsClosureCreated && len(f.CapturedVars) > 0 {
		out += " [closure with " + string(rune(len(f.CapturedVars))) + " captured vars]"
	}

	return out
}

func (f *Function) CreateClosure(env *Environment, requiredVars []string) *Function {
	closure := &Function{
		Parameters:       f.Parameters,
		Body:             f.Body,
		Env:              f.Env,
		CapturedVars:     make(map[string]Object),
		IsClosureCreated: true,
	}

	for _, varName := range requiredVars {
		if obj, ok := env.Get(varName); ok {
			closure.CapturedVars[varName] = obj
		}
	}

	return closure
}

func (f *Function) GetCapturedVar(name string) (Object, bool) {
	if f.CapturedVars == nil {
		return nil, false
	}
	obj, ok := f.CapturedVars[name]
	return obj, ok
}

func (f *Function) UpdateCapturedVar(name string, value Object) {
	if f.CapturedVars == nil {
		f.CapturedVars = make(map[string]Object)
	}
	f.CapturedVars[name] = value
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

type ArrowFunction struct {
	Parameters       []*ast.Identifier
	Body             ast.Expression
	Env              *Environment
	CapturedVars     map[string]Object
	IsClosureCreated bool
}

func (af *ArrowFunction) Type() ObjectType { return ARROW_FUNCTION_OBJ }
func (af *ArrowFunction) Inspect() string {
	out := ""
	params := []string{}
	for _, p := range af.Parameters {
		params = append(params, p.String())
	}

	if len(af.Parameters) == 1 {
		out += af.Parameters[0].String()
	} else {
		out += "("
		out += strings.Join(params, ", ")
		out += ")"
	}

	out += " => "
	out += af.Body.String()

	if af.IsClosureCreated && len(af.CapturedVars) > 0 {
		out += " [closure]"
	}

	return out
}

func (af *ArrowFunction) CreateClosure(env *Environment, requiredVars []string) *ArrowFunction {
	closure := &ArrowFunction{
		Parameters:       af.Parameters,
		Body:             af.Body,
		Env:              af.Env,
		CapturedVars:     make(map[string]Object),
		IsClosureCreated: true,
	}

	for _, varName := range requiredVars {
		if obj, ok := env.Get(varName); ok {
			closure.CapturedVars[varName] = obj
		}
	}

	return closure
}

func (af *ArrowFunction) GetCapturedVar(name string) (Object, bool) {
	if af.CapturedVars == nil {
		return nil, false
	}
	obj, ok := af.CapturedVars[name]
	return obj, ok
}

func (af *ArrowFunction) UpdateCapturedVar(name string, value Object) {
	if af.CapturedVars == nil {
		af.CapturedVars = make(map[string]Object)
	}
	af.CapturedVars[name] = value
}

type AsyncFunction struct {
	Parameters       []*ast.Identifier
	Body             *ast.BlockStatement
	Env              *Environment
	CapturedVars     map[string]Object
	IsClosureCreated bool
}

func (af *AsyncFunction) Type() ObjectType { return ASYNC_FUNCTION_OBJ }
func (af *AsyncFunction) Inspect() string {
	out := "async fn("
	params := []string{}
	for _, p := range af.Parameters {
		params = append(params, p.String())
	}
	out += strings.Join(params, ", ")
	out += ") {\n"
	out += af.Body.String()
	out += "\n}"

	if af.IsClosureCreated && len(af.CapturedVars) > 0 {
		out += " [async closure]"
	}

	return out
}

func (af *AsyncFunction) CreateClosure(env *Environment, requiredVars []string) *AsyncFunction {
	closure := &AsyncFunction{
		Parameters:       af.Parameters,
		Body:             af.Body,
		Env:              af.Env,
		CapturedVars:     make(map[string]Object),
		IsClosureCreated: true,
	}

	for _, varName := range requiredVars {
		if obj, ok := env.Get(varName); ok {
			closure.CapturedVars[varName] = obj
		}
	}

	return closure
}

func (af *AsyncFunction) GetCapturedVar(name string) (Object, bool) {
	if af.CapturedVars == nil {
		return nil, false
	}
	obj, ok := af.CapturedVars[name]
	return obj, ok
}

func (af *AsyncFunction) UpdateCapturedVar(name string, value Object) {
	if af.CapturedVars == nil {
		af.CapturedVars = make(map[string]Object)
	}
	af.CapturedVars[name] = value
}

type PromiseState int

const (
	PENDING PromiseState = iota
	RESOLVED
	REJECTED
)

func (ps PromiseState) String() string {
	switch ps {
	case PENDING:
		return "pending"
	case RESOLVED:
		return "resolved"
	case REJECTED:
		return "rejected"
	default:
		return "unknown"
	}
}

type Promise struct {
	State      PromiseState
	Value      Object
	Reason     Object
	OnResolved []func(Object)
	OnRejected []func(Object)
}

func (p *Promise) Type() ObjectType { return PROMISE_OBJ }
func (p *Promise) Inspect() string {
	switch p.State {
	case PENDING:
		return "Promise { <pending> }"
	case RESOLVED:
		if p.Value != nil {
			return "Promise { <resolved>: " + p.Value.Inspect() + " }"
		}
		return "Promise { <resolved> }"
	case REJECTED:
		if p.Reason != nil {
			return "Promise { <rejected>: " + p.Reason.Inspect() + " }"
		}
		return "Promise { <rejected> }"
	default:
		return "Promise { <unknown> }"
	}
}

func (p *Promise) Resolve(value Object) {
	if p.State != PENDING {
		return
	}

	p.State = RESOLVED
	p.Value = value

	for _, callback := range p.OnResolved {
		callback(value)
	}

	p.OnResolved = nil
	p.OnRejected = nil
}

func (p *Promise) Reject(reason Object) {
	if p.State != PENDING {
		return
	}

	p.State = REJECTED
	p.Reason = reason

	for _, callback := range p.OnRejected {
		callback(reason)
	}

	p.OnResolved = nil
	p.OnRejected = nil
}

func (p *Promise) Then(onResolved func(Object), onRejected func(Object)) {
	switch p.State {
	case PENDING:
		if onResolved != nil {
			p.OnResolved = append(p.OnResolved, onResolved)
		}
		if onRejected != nil {
			p.OnRejected = append(p.OnRejected, onRejected)
		}
	case RESOLVED:
		if onResolved != nil {
			onResolved(p.Value)
		}
	case REJECTED:
		if onRejected != nil {
			onRejected(p.Reason)
		}
	}
}

func NewPromise() *Promise {
	return &Promise{
		State:      PENDING,
		Value:      nil,
		Reason:     nil,
		OnResolved: make([]func(Object), 0),
		OnRejected: make([]func(Object), 0),
	}
}

func NewResolvedPromise(value Object) *Promise {
	return &Promise{
		State:      RESOLVED,
		Value:      value,
		Reason:     nil,
		OnResolved: nil,
		OnRejected: nil,
	}
}

func NewRejectedPromise(reason Object) *Promise {
	return &Promise{
		State:      REJECTED,
		Value:      nil,
		Reason:     reason,
		OnResolved: nil,
		OnRejected: nil,
	}
}
