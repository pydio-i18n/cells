package registry

type Node interface{
	Name() string
	Address() []string
	Endpoints() []string
	Metadata() map[string]string

	As(interface{}) bool
}

type Endpoint interface {
	Name()     string
	Metadata() map[string]string
}

type NodeRegistry interface {
	RegisterNode(Node) error
	DeregisterNode(Node) error
	GetNode(string) (Node, error)
	ListNodes() ([]Node, error)
	As(interface{}) bool
}