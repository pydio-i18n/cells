package registry

type Service interface{
	Name() string
	Version() string
	Nodes() []Node
	Tags() []string

	Start() error
	Stop() error

	IsGeneric() bool
	IsGRPC() bool
	IsREST() bool
}