package object

// Environment representa un entorno de ejecución que almacena variables y constantes
// Ahora con soporte mejorado para closures y scoping avanzado
type Environment struct {
	store     map[string]Object
	constants map[string]Object // mapa para constantes inmutables
	outer     *Environment
	// Nuevo: información para closures
	closureVars map[string]Object // variables capturadas por closures
	isClosure   bool              // indica si este ambiente es para un closure
}

// NewEnvironment crea un nuevo entorno vacío
func NewEnvironment() *Environment {
	return &Environment{
		store:       make(map[string]Object),
		constants:   make(map[string]Object),
		closureVars: make(map[string]Object),
		outer:       nil,
		isClosure:   false,
	}
}

// NewEnclosedEnvironment crea un nuevo entorno que extiende otro entorno
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// NewClosureEnvironment crea un entorno específico para closures
func NewClosureEnvironment(outer *Environment) *Environment {
	env := NewEnclosedEnvironment(outer)
	env.isClosure = true
	return env
}

// Get obtiene el valor de una variable o constante por nombre
// Busca primero en el entorno actual, luego en entornos externos
func (e *Environment) Get(name string) (Object, bool) {
	// Buscar primero en variables capturadas por closures
	if obj, ok := e.closureVars[name]; ok {
		return obj, true
	}

	// Buscar en constantes
	obj, ok := e.constants[name]
	if ok {
		return obj, true
	}

	// Luego buscar en variables
	obj, ok = e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set establece el valor de una variable en el entorno actual
func (e *Environment) Set(name string, val Object) Object {
	// Verificar si es una constante (no se puede modificar)
	if _, isConst := e.constants[name]; isConst {
		return nil // Error será manejado por el evaluador
	}

	// Si es una variable capturada por un closure, actualizarla
	if _, isCaptured := e.closureVars[name]; isCaptured {
		e.closureVars[name] = val
		return val
	}

	// Si la variable existe en el entorno padre y este es un closure,
	// actualizar en el entorno padre
	if e.outer != nil {
		if _, existsInOuter := e.outer.Get(name); existsInOuter {
			return e.outer.Set(name, val)
		}
	}

	e.store[name] = val
	return val
}

// SetConstant establece el valor de una constante en el entorno actual
func (e *Environment) SetConstant(name string, val Object) Object {
	// Verificar si ya existe como variable o constante
	if _, exists := e.store[name]; exists {
		return nil // Error será manejado por el evaluador
	}
	if _, exists := e.constants[name]; exists {
		return nil // Error será manejado por el evaluador
	}
	if _, exists := e.closureVars[name]; exists {
		return nil // Error será manejado por el evaluador
	}

	e.constants[name] = val
	return val
}

// CaptureVariable captura una variable del entorno externo para uso en closures
func (e *Environment) CaptureVariable(name string, value Object) {
	if e.closureVars == nil {
		e.closureVars = make(map[string]Object)
	}
	e.closureVars[name] = value
}

// GetCapturedVariables obtiene todas las variables capturadas por este entorno
func (e *Environment) GetCapturedVariables() map[string]Object {
	captured := make(map[string]Object)
	for name, obj := range e.closureVars {
		captured[name] = obj
	}
	return captured
}

// IsConstant verifica si un nombre es una constante
func (e *Environment) IsConstant(name string) bool {
	_, isConst := e.constants[name]
	if !isConst && e.outer != nil {
		return e.outer.IsConstant(name)
	}
	return isConst
}

// IsCaptured verifica si una variable está capturada por un closure
func (e *Environment) IsCaptured(name string) bool {
	_, isCaptured := e.closureVars[name]
	return isCaptured
}

// GetStore devuelve el store del entorno para acceso directo
func (e *Environment) GetStore() map[string]Object {
	return e.store
}

// GetAll devuelve todas las variables y constantes del entorno actual
func (e *Environment) GetAll() map[string]Object {
	result := make(map[string]Object)

	// Agregar variables
	for name, obj := range e.store {
		result[name] = obj
	}

	// Agregar constantes
	for name, obj := range e.constants {
		result[name] = obj
	}

	// Agregar variables capturadas
	for name, obj := range e.closureVars {
		result[name] = obj
	}

	return result
}

// FindVariableEnvironment encuentra el entorno donde está definida una variable
// Útil para determinar si una variable debe ser capturada por un closure
func (e *Environment) FindVariableEnvironment(name string) *Environment {
	// Verificar en el entorno actual
	if _, ok := e.store[name]; ok {
		return e
	}
	if _, ok := e.constants[name]; ok {
		return e
	}
	if _, ok := e.closureVars[name]; ok {
		return e
	}

	// Buscar en entornos externos
	if e.outer != nil {
		return e.outer.FindVariableEnvironment(name)
	}

	return nil
}

// CreateClosureCapture crea una captura de variables para un closure
// analizando qué variables del entorno externo son necesarias
func (e *Environment) CreateClosureCapture(requiredVars []string) map[string]Object {
	captured := make(map[string]Object)

	for _, varName := range requiredVars {
		if obj, ok := e.Get(varName); ok {
			captured[varName] = obj
		}
	}

	return captured
}

// ApplyClosureCapture aplica las variables capturadas a un entorno de closure
func (e *Environment) ApplyClosureCapture(capturedVars map[string]Object) {
	if e.closureVars == nil {
		e.closureVars = make(map[string]Object)
	}

	for name, obj := range capturedVars {
		e.closureVars[name] = obj
	}
}

// GetDepth obtiene la profundidad del entorno (cuántos niveles de anidamiento)
func (e *Environment) GetDepth() int {
	depth := 0
	current := e
	for current.outer != nil {
		depth++
		current = current.outer
	}
	return depth
}

// IsNestedIn verifica si este entorno está anidado dentro de otro
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

// Clone crea una copia profunda del entorno para uso en closures
func (e *Environment) Clone() *Environment {
	clone := &Environment{
		store:       make(map[string]Object),
		constants:   make(map[string]Object),
		closureVars: make(map[string]Object),
		isClosure:   e.isClosure,
		outer:       e.outer, // Mantener referencia al entorno exterior
	}

	// Copiar variables
	for name, obj := range e.store {
		clone.store[name] = obj
	}

	// Copiar constantes
	for name, obj := range e.constants {
		clone.constants[name] = obj
	}

	// Copiar variables capturadas
	for name, obj := range e.closureVars {
		clone.closureVars[name] = obj
	}

	return clone
}
