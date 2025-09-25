package object

import (
	"jabline/pkg/ast"
	"strings"
)

// Function representa una función definida por el usuario con soporte para closures
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
	// Nuevo: soporte para closures
	CapturedVars     map[string]Object // variables capturadas del entorno externo
	IsClosureCreated bool              // indica si esta función fue creada como closure
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

	// Agregar información de closure si aplica
	if f.IsClosureCreated && len(f.CapturedVars) > 0 {
		out += " [closure with " + string(rune(len(f.CapturedVars))) + " captured vars]"
	}

	return out
}

// CreateClosure crea un closure capturando variables del entorno
func (f *Function) CreateClosure(env *Environment, requiredVars []string) *Function {
	closure := &Function{
		Parameters:       f.Parameters,
		Body:             f.Body,
		Env:              f.Env,
		CapturedVars:     make(map[string]Object),
		IsClosureCreated: true,
	}

	// Capturar variables necesarias
	for _, varName := range requiredVars {
		if obj, ok := env.Get(varName); ok {
			closure.CapturedVars[varName] = obj
		}
	}

	return closure
}

// GetCapturedVar obtiene una variable capturada por el closure
func (f *Function) GetCapturedVar(name string) (Object, bool) {
	if f.CapturedVars == nil {
		return nil, false
	}
	obj, ok := f.CapturedVars[name]
	return obj, ok
}

// UpdateCapturedVar actualiza una variable capturada por el closure
func (f *Function) UpdateCapturedVar(name string, value Object) {
	if f.CapturedVars == nil {
		f.CapturedVars = make(map[string]Object)
	}
	f.CapturedVars[name] = value
}

// BuiltinFunction is the type for built-in function implementations
type BuiltinFunction func(args ...Object) Object

// Builtin represents a built-in function
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

// ArrowFunction representa una arrow function definida por el usuario con soporte para closures
type ArrowFunction struct {
	Parameters []*ast.Identifier
	Body       ast.Expression // Arrow functions tienen cuerpos de expresión
	Env        *Environment
	// Nuevo: soporte para closures
	CapturedVars     map[string]Object // variables capturadas del entorno externo
	IsClosureCreated bool              // indica si esta función fue creada como closure
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

	// Agregar información de closure si aplica
	if af.IsClosureCreated && len(af.CapturedVars) > 0 {
		out += " [closure]"
	}

	return out
}

// CreateClosure crea un closure capturando variables del entorno
func (af *ArrowFunction) CreateClosure(env *Environment, requiredVars []string) *ArrowFunction {
	closure := &ArrowFunction{
		Parameters:       af.Parameters,
		Body:             af.Body,
		Env:              af.Env,
		CapturedVars:     make(map[string]Object),
		IsClosureCreated: true,
	}

	// Capturar variables necesarias
	for _, varName := range requiredVars {
		if obj, ok := env.Get(varName); ok {
			closure.CapturedVars[varName] = obj
		}
	}

	return closure
}

// GetCapturedVar obtiene una variable capturada por el closure
func (af *ArrowFunction) GetCapturedVar(name string) (Object, bool) {
	if af.CapturedVars == nil {
		return nil, false
	}
	obj, ok := af.CapturedVars[name]
	return obj, ok
}

// UpdateCapturedVar actualiza una variable capturada por el closure
func (af *ArrowFunction) UpdateCapturedVar(name string, value Object) {
	if af.CapturedVars == nil {
		af.CapturedVars = make(map[string]Object)
	}
	af.CapturedVars[name] = value
}

// AsyncFunction representa una función async definida por el usuario con soporte para closures
type AsyncFunction struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
	// Nuevo: soporte para closures
	CapturedVars     map[string]Object // variables capturadas del entorno externo
	IsClosureCreated bool              // indica si esta función fue creada como closure
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

	// Agregar información de closure si aplica
	if af.IsClosureCreated && len(af.CapturedVars) > 0 {
		out += " [async closure]"
	}

	return out
}

// CreateClosure crea un closure capturando variables del entorno
func (af *AsyncFunction) CreateClosure(env *Environment, requiredVars []string) *AsyncFunction {
	closure := &AsyncFunction{
		Parameters:       af.Parameters,
		Body:             af.Body,
		Env:              af.Env,
		CapturedVars:     make(map[string]Object),
		IsClosureCreated: true,
	}

	// Capturar variables necesarias
	for _, varName := range requiredVars {
		if obj, ok := env.Get(varName); ok {
			closure.CapturedVars[varName] = obj
		}
	}

	return closure
}

// GetCapturedVar obtiene una variable capturada por el closure
func (af *AsyncFunction) GetCapturedVar(name string) (Object, bool) {
	if af.CapturedVars == nil {
		return nil, false
	}
	obj, ok := af.CapturedVars[name]
	return obj, ok
}

// UpdateCapturedVar actualiza una variable capturada por el closure
func (af *AsyncFunction) UpdateCapturedVar(name string, value Object) {
	if af.CapturedVars == nil {
		af.CapturedVars = make(map[string]Object)
	}
	af.CapturedVars[name] = value
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
