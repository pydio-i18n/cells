package memory

import (
	"fmt"
	"github.com/pydio/cells/v4/common/registry"
	"runtime"
	"testing"
)

func TestMemory(t *testing.T) {
	m := &memory{}
	s := registry.NewService("test", "0.0.0", map[string]string{})

	m.RegisterService(s)

	fmt.Println(m.services[0])

	s = nil

	runtime.GC()

	fmt.Println(m.services[0].Name())
}
