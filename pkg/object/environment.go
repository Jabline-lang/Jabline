package object

type Environment struct {
	store       map[string]Object
	constants   map[string]Object
	outer       *Environment
	ModulePath  string
	closureVars map[string]Object
	isClosure   bool
}

func NewEnvironment() *Environment {
	return &Environment{
		store:       make(map[string]Object),
		constants:   make(map[string]Object),
		closureVars: make(map[string]Object),
		outer:       nil,
		isClosure:   false,
	}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	if outer != nil {
		env.ModulePath = outer.ModulePath
	}
	return env
}

func NewClosureEnvironment(outer *Environment) *Environment {
	env := NewEnclosedEnvironment(outer)
	env.isClosure = true
	return env
}

func (e *Environment) Get(name string) (Object, bool) {

	if obj, ok := e.closureVars[name]; ok {
		return obj, true
	}

	obj, ok := e.constants[name]
	if ok {
		return obj, true
	}

	obj, ok = e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {

	if _, isConst := e.constants[name]; isConst {
		return nil
	}

	if _, isCaptured := e.closureVars[name]; isCaptured {
		e.closureVars[name] = val
		return val
	}

	if e.outer != nil {
		if _, existsInOuter := e.outer.Get(name); existsInOuter {
			return e.outer.Set(name, val)
		}
	}

	e.store[name] = val
	return val
}

func (e *Environment) SetConstant(name string, val Object) Object {

	if _, exists := e.store[name]; exists {
		return nil
	}
	if _, exists := e.constants[name]; exists {
		return nil
	}
	if _, exists := e.closureVars[name]; exists {
		return nil
	}

	e.constants[name] = val
	return val
}

func (e *Environment) CaptureVariable(name string, value Object) {
	if e.closureVars == nil {
		e.closureVars = make(map[string]Object)
	}
	e.closureVars[name] = value
}

func (e *Environment) GetCapturedVariables() map[string]Object {
	captured := make(map[string]Object)
	for name, obj := range e.closureVars {
		captured[name] = obj
	}
	return captured
}

func (e *Environment) IsConstant(name string) bool {
	_, isConst := e.constants[name]
	if !isConst && e.outer != nil {
		return e.outer.IsConstant(name)
	}
	return isConst
}

func (e *Environment) IsCaptured(name string) bool {
	_, isCaptured := e.closureVars[name]
	return isCaptured
}

func (e *Environment) GetStore() map[string]Object {
	return e.store
}

func (e *Environment) GetAll() map[string]Object {
	result := make(map[string]Object)

	for name, obj := range e.store {
		result[name] = obj
	}

	for name, obj := range e.constants {
		result[name] = obj
	}

	for name, obj := range e.closureVars {
		result[name] = obj
	}

	return result
}

func (e *Environment) FindVariableEnvironment(name string) *Environment {

	if _, ok := e.store[name]; ok {
		return e
	}
	if _, ok := e.constants[name]; ok {
		return e
	}
	if _, ok := e.closureVars[name]; ok {
		return e
	}

	if e.outer != nil {
		return e.outer.FindVariableEnvironment(name)
	}

	return nil
}

func (e *Environment) CreateClosureCapture(requiredVars []string) map[string]Object {
	captured := make(map[string]Object)

	for _, varName := range requiredVars {
		if obj, ok := e.Get(varName); ok {
			captured[varName] = obj
		}
	}

	return captured
}

func (e *Environment) ApplyClosureCapture(capturedVars map[string]Object) {
	if e.closureVars == nil {
		e.closureVars = make(map[string]Object)
	}

	for name, obj := range capturedVars {
		e.closureVars[name] = obj
	}
}

func (e *Environment) GetDepth() int {
	depth := 0
	current := e
	for current.outer != nil {
		depth++
		current = current.outer
	}
	return depth
}

func (e *Environment) IsNestedIn(other *Environment) bool {
	current := e.outer
	for current != nil {
		if current == other {
			return true
		}
		current = current.outer
	}
	return false
}

func (e *Environment) Clone() *Environment {
	clone := &Environment{
		store:       make(map[string]Object),
		constants:   make(map[string]Object),
		closureVars: make(map[string]Object),
		isClosure:   e.isClosure,
		outer:       e.outer,
	}

	for name, obj := range e.store {
		clone.store[name] = obj
	}

	for name, obj := range e.constants {
		clone.constants[name] = obj
	}

	for name, obj := range e.closureVars {
		clone.closureVars[name] = obj
	}

	return clone
}
