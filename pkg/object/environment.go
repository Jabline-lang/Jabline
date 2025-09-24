package object

// Environment representa un entorno de ejecución que almacena variables y constantes
type Environment struct {
	store     map[string]Object
	constants map[string]Object // mapa para constantes inmutables
	outer     *Environment
}

// NewEnvironment crea un nuevo entorno vacío
func NewEnvironment() *Environment {
	return &Environment{
		store:     make(map[string]Object),
		constants: make(map[string]Object),
		outer:     nil,
	}
}

// NewEnclosedEnvironment crea un nuevo entorno que extiende otro entorno
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get obtiene el valor de una variable o constante por nombre
// Busca primero en el entorno actual, luego en entornos externos
func (e *Environment) Get(name string) (Object, bool) {
	// Buscar primero en constantes
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

	e.constants[name] = val
	return val
}

// IsConstant verifica si un nombre es una constante
func (e *Environment) IsConstant(name string) bool {
	_, isConst := e.constants[name]
	if !isConst && e.outer != nil {
		return e.outer.IsConstant(name)
	}
	return isConst
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

	return result
}
