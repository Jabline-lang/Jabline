package object

type Service struct {
	Name   string
	Config map[string]Object
}

func (s *Service) Type() ObjectType { return "SERVICE" }
func (s *Service) Inspect() string {
	out := "service " + s.Name + " {\n"
	for k, v := range s.Config {
		out += "  " + k + ": " + v.Inspect() + "\n"
	}
	out += "}"
	return out
}
