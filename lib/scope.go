package lib

type Scope struct {
	vars map[string]any
	Name string
}

func NewScope(name string) *Scope {
	return &Scope{Name: name, vars: make(map[string]any)}
}

func (s *Scope) AddVar(name string, t any) {
	s.vars[name] = t
}

func (s *Scope) GetVar(name string) any {
	return s.vars[name]
}
