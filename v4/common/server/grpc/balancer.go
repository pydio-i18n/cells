package grpc

import (
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
	"sync"
)

const name = "lb"

// newBuilder creates a new roundrobin balancer builder.
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(name, &rrPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type rrPickerBuilder struct{}

func (*rrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	scs := make(map[string]*rrPickerConns)
	for sc, sci := range info.ReadySCs {
		for _, s := range sci.Address.Attributes.Value("services").([]string) {
			v, ok := scs[s]
			if !ok {
				v = &rrPickerConns{}
				scs[s] = v
			}

			v.subConns = append(v.subConns, sc)
		}
	}
	for _, sc := range scs {
		sc.next = rand.Intn(len(sc.subConns))
	}
	return &rrPicker{
		subConns: scs,
	}
}

type rrPicker struct {
	subConns map[string]*rrPickerConns
}

type rrPickerConns struct {
	subConns []balancer.SubConn
	mu   sync.Mutex
	next int
}

func (p *rrPicker) Pick(i balancer.PickInfo) (balancer.PickResult, error) {
	serviceName := servicecontext.GetServiceName(i.Ctx)
	pc, ok := p.subConns[serviceName]
	if !ok {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	pc.mu.Lock()
	sc := pc.subConns[pc.next]
	pc.next = (pc.next + 1) % len(pc.subConns)
	pc.mu.Unlock()
	return balancer.PickResult{SubConn: sc}, nil
}