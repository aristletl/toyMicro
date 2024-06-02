package roundrobin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync"
)

type WeightPicker struct {
	conns []*conn
	mutex sync.Mutex
}

func (w *WeightPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()
	var totalWeight uint32 = 0
	var maxWeightConn *conn
	for _, cc := range w.conns {
		totalWeight += cc.weight
		cc.currentWeight += cc.efficientWeight
		if maxWeightConn == nil || maxWeightConn.currentWeight < cc.currentWeight {
			maxWeightConn = cc
		}
	}

	maxWeightConn.currentWeight -= totalWeight
	return balancer.PickResult{
		SubConn: maxWeightConn.SubConn,
		Done: func(info balancer.DoneInfo) {
			if info.Err != nil {
				maxWeightConn.efficientWeight--
			} else {
				maxWeightConn.efficientWeight++
			}
		},
	}, nil
}

type WeightPickerBuilder struct {
}

func (w *WeightPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	for subConn, subConnInfo := range info.ReadySCs {
		weight := uint32(subConnInfo.Address.Attributes.Value("weight").(int))
		conns = append(conns, &conn{
			SubConn:         subConn,
			weight:          weight,
			efficientWeight: weight,
		})
	}
	return &WeightPicker{conns: conns}
}

type conn struct {
	balancer.SubConn
	weight          uint32
	currentWeight   uint32
	efficientWeight uint32
}
