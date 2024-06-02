package loadbalance

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

type Filter func(info balancer.PickInfo, address resolver.Address) bool

func GroupFilter(info balancer.PickInfo, address resolver.Address) bool {
	group := info.Ctx.Value("group")
	if group == nil {
		return true
	}

	return group == address.Attributes.Value("group")
}
