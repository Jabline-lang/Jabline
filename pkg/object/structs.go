package object

type Struct struct {
	Name           string
	TypeParameters []string
	Fields         map[string]string
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

type InstantiatedStruct struct {
	Struct       *Struct
	FullTypeName string
}

func (is *InstantiatedStruct) Type() ObjectType { return INSTANTIATED_STRUCT_OBJ }
func (is *InstantiatedStruct) Inspect() string {
	return "instantiated struct " + is.FullTypeName
}

type Instance struct {
	StructName string // Esto ahora almacenará el nombre completo instanciado si aplica
	Fields     map[string]Object
}

func (i *Instance) Type() ObjectType { return INSTANCE_OBJ }
func (i *Instance) Inspect() string {
	out := i.StructName + " {\n"
	for name, value := range i.Fields {
		out += "  " + name + ": " + value.Inspect() + ",\n"
	}
	out += "}\n"
	return out
}

type InterfaceMethod struct {
	Name       string
	Parameters []string // We only keep param names for now
	ReturnType string   // Expected return type name
}

type Interface struct {
	Name           string
	TypeParameters []string
	Methods        map[string]*InterfaceMethod
}

func (i *Interface) Type() ObjectType { return INTERFACE_OBJ }
func (i *Interface) Inspect() string {
	out := "interface " + i.Name + " {\n"
	for name, method := range i.Methods {
		out += "  " + name + "("
		for j, p := range method.Parameters {
			out += p
			if j < len(method.Parameters)-1 {
				out += ", "
			}
		}
		out += "): " + method.ReturnType + "\n"
	}
	out += "}"
	return out
}
