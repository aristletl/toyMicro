package roundrobin

import (
	"github.com/ztruane/toy-micro/loadbalance"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync/atomic"
)

type Picker struct {
	//conns []balancer.SubConn
	ins []instance
	cnt uint64

	filter loadbalance.Filter
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	//if len(p.conns) == 0 {
	//	return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	//}
	//cnt := atomic.AddUint64(&p.cnt, 1)
	//idx := cnt % uint64(len(p.conns))
	//return balancer.PickResult{
	//	SubConn: p.conns[idx],
	//}, nil

	if len(p.ins) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	candidate := make([]instance, 0, len(p.ins))
	for _, v := range p.ins {
		if p.filter(info, v.addr) {
			candidate = append(candidate, v)
		}
	}
	cnt := atomic.AddUint64(&p.cnt, 1)
	idx := cnt % uint64(len(p.ins))
	return balancer.PickResult{
		SubConn: candidate[idx].sub,
	}, nil
}

type PickerBuilder struct {
	Filter loadbalance.Filter
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	//conns := make([]balancer.SubConn, 0, len(info.ReadySCs))
	//for conn, _ := range info.ReadySCs {
	//	conns = append(conns, conn)
	//}
	//return &Picker{conns: conns}

	ins := make([]instance, 0, len(info.ReadySCs))
	for sub, subInfo := range info.ReadySCs {
		ins = append(ins, instance{
			sub:  sub,
			addr: subInfo.Address,
		})
	}
	return &Picker{
		ins:    ins,
		filter: p.Filter,
	}
}

func (p *PickerBuilder) Name() string {
	return "ROUND_ROBIN"
}

type instance struct {
	sub  balancer.SubConn
	addr resolver.Address
}
