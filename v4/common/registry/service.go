package registry

type Service interface{
	Name() string
	Version() string
	Nodes() []Node
	Tags() []string

	IsGeneric() bool
	IsGRPC() bool
	IsREST() bool
}