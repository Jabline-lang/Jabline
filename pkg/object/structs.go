package object

// Struct representa una definición de estructura
type Struct struct {
	Name   string
	Fields map[string]string // tipos de los campos
}

func (s *Struct) Type() ObjectType { return STRUCT_OBJ }
func (s *Struct) Inspect() string {
	out := "struct " + s.Name + " {\n"
	for name, fieldType := range s.Fields {
		out += "  " + name + ": " + fieldType + ",\n"
	}
	out += "}"
	return out
}

// Instance representa una instancia de una estructura
type Instance struct {
	StructName string
	Fields     map[string]Object
}

func (i *Instance) Type() ObjectType { return INSTANCE_OBJ }
func (i *Instance) Inspect() string {
	out := i.StructName + " {\n"
	for name, value := range i.Fields {
		out += "  " + name + ": " + value.Inspect() + ",\n"
	}
	out += "}"
	return out
}
