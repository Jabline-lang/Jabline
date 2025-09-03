package object

import (
	"fmt"
	"strings"
)

// Array representa un array/lista de objetos
type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}
	out := "["
	out += strings.Join(elements, ", ")
	out += "]"
	return out
}

// Hash representa un hash map/diccionario
type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out := "{"
	out += strings.Join(pairs, ", ")
	out += "}"
	return out
}

// HashKey representa una clave única para el hash map
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// HashPair representa un par clave-valor en el hash map
type HashPair struct {
	Key   Object
	Value Object
}
