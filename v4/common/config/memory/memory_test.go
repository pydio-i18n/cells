package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/pydio/cells/v4/common/config"
)

func TestInit(t *testing.T) {
	ctx := context.Background()
	conf, err := config.OpenStore(ctx, "memory:///")

	conf.Set("")
	fmt.Println(conf)
}