package leastactive

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Picker struct {
	conns []*conn
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	res := p.conns[0]
	for i := 1; i < len(p.conns); i++ {
		if res.numReq > p.conns[i].numReq {
			res = p.conns[i]
		}
	}
	res.numReq++
	return balancer.PickResult{SubConn: res.sub, Done: func(info balancer.DoneInfo) {
		atomic.AddUint64(&res.numReq, -1)
	}}, nil
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	for subConn, _ := range info.ReadySCs {
		conns = append(conns, &conn{numReq: 0, sub: subConn})
	}

	return &Picker{conns: conns}
}

type conn struct {
	numReq uint64
	sub    balancer.SubConn
}
