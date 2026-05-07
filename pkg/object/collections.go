package object

import (
	"fmt"
	"strings"
)

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

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) InspectWithVisited(visited map[*Hash]bool) string {
	if visited[h] {
		return "{ <cycle> }"
	}
	visited[h] = true
	defer delete(visited, h)

	pairs := []string{}
	for _, pair := range h.Pairs {
		keyStr := pair.Key.Inspect()
		valStr := ""
		if valHash, ok := pair.Value.(*Hash); ok {
			valStr = valHash.InspectWithVisited(visited)
		} else {
			valStr = pair.Value.Inspect()
		}
		pairs = append(pairs, fmt.Sprintf("%s: %s", keyStr, valStr))
	}
	out := "{"
	out += strings.Join(pairs, ", ")
	out += "}"
	return out
}

func (h *Hash) Inspect() string {
	return h.InspectWithVisited(make(map[*Hash]bool))
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type HashPair struct {
	Key   Object
	Value Object
}
