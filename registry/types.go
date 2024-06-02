package registry

import (
	"context"
	"io"
)

//go:generate mockgen -package=mocks -destination=mocks/registry.mock.go -source=types.go Registry
type Registry interface {
	// Register 注册一个服务实例
	Register(ctx context.Context, service ServiceInstance) error
	// Unregister 注销
	Unregister(ctx context.Context, serviceName string) error
	// ListerServer 客户端的方法，用于获取某个服务的实例列表
	ListerServer(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	// Subscribe 订阅服务状态，该服务相关的事件会通过该接口告知(比如，新节点加入，老节点退出等)
	Subscribe(ctx context.Context, serviceName string) (<-chan Event, error)
	io.Closer
}

// ServiceInstance 代表一个服务实例
type ServiceInstance struct {
	ServiceName string
	Addr        string
	Weight      uint32

	Group string // 分组信息
}

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeAdd               // 新增
	EventTypeDelete            // 删除
	EventTypeUpdate            // 更新
)

type Event struct {
	Type     EventType
	Instance ServiceInstance
}
