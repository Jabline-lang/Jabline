package object

import (
	"jabline/pkg/ast"
	"strings"
)

// Function representa una función definida por el usuario
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
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
	return out
}

// BuiltinFunction is the type for built-in function implementations
type BuiltinFunction func(args ...Object) Object

// Builtin represents a built-in function
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

// ArrowFunction representa una arrow function definida por el usuario
type ArrowFunction struct {
	Parameters []*ast.Identifier
	Body       ast.Expression // Arrow functions tienen cuerpos de expresión
	Env        *Environment
}

func (af *ArrowFunction) Type() ObjectType { return ARROW_FUNCTION_OBJ }
func (af *ArrowFunction) Inspect() string {
	out := ""
	params := []string{}
	for _, p := range af.Parameters {
		params = append(params, p.String())
	}

	// Single parameter without parentheses
	if len(af.Parameters) == 1 {
		out += af.Parameters[0].String()
	} else {
		// Multiple or zero parameters with parentheses
		out += "("
		out += strings.Join(params, ", ")
		out += ")"
	}

	out += " => "
	out += af.Body.String()
	return out
}

// AsyncFunction representa una función async definida por el usuario
type AsyncFunction struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
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
	return out
}

// PromiseState representa el estado de una Promise
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

// Promise representa una promesa para operaciones asíncronas
type Promise struct {
	State  PromiseState
	Value  Object // valor cuando está resolved
	Reason Object // razón cuando está rejected

	// Callbacks para when resolved/rejected
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

// Resolve resuelve la Promise con un valor
func (p *Promise) Resolve(value Object) {
	if p.State != PENDING {
		return // Ya está resuelta o rechazada
	}

	p.State = RESOLVED
	p.Value = value

	// Ejecutar callbacks
	for _, callback := range p.OnResolved {
		callback(value)
	}

	// Limpiar callbacks
	p.OnResolved = nil
	p.OnRejected = nil
}

// Reject rechaza la Promise con una razón
func (p *Promise) Reject(reason Object) {
	if p.State != PENDING {
		return // Ya está resuelta o rechazada
	}

	p.State = REJECTED
	p.Reason = reason

	// Ejecutar callbacks
	for _, callback := range p.OnRejected {
		callback(reason)
	}

	// Limpiar callbacks
	p.OnResolved = nil
	p.OnRejected = nil
}

// Then agrega callbacks para cuando se resuelve o rechaza
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

// NewPromise crea una nueva Promise en estado pending
func NewPromise() *Promise {
	return &Promise{
		State:      PENDING,
		Value:      nil,
		Reason:     nil,
		OnResolved: make([]func(Object), 0),
		OnRejected: make([]func(Object), 0),
	}
}

// NewResolvedPromise crea una Promise ya resuelta
func NewResolvedPromise(value Object) *Promise {
	return &Promise{
		State:      RESOLVED,
		Value:      value,
		Reason:     nil,
		OnResolved: nil,
		OnRejected: nil,
	}
}

// NewRejectedPromise crea una Promise ya rechazada
func NewRejectedPromise(reason Object) *Promise {
	return &Promise{
		State:      REJECTED,
		Value:      nil,
		Reason:     reason,
		OnResolved: nil,
		OnRejected: nil,
	}
}
