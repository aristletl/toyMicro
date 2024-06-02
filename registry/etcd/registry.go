package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ztruane/toy-micro/registry"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var typesMap = map[mvccpb.Event_EventType]registry.EventType{
	mvccpb.PUT:    registry.EventTypeAdd,
	mvccpb.DELETE: registry.EventTypeDelete,
}

type Registry struct {
	client *clientv3.Client
	sess   *concurrency.Session

	mutex       sync.RWMutex
	watchCancel []func()
}

func NewRegistry(client *clientv3.Client) (*Registry, error) {
	sess, err := concurrency.NewSession(client)
	if err != nil {
		return nil, err
	}
	return &Registry{
		client: client,
		sess:   sess,
		mutex:  sync.RWMutex{},
	}, nil
}

func (r *Registry) Register(ctx context.Context, service registry.ServiceInstance) error {
	instanceKey := fmt.Sprintf("micro/%s/%s", service.ServiceName, service.Addr)
	val, err := json.Marshal(service)
	if err != nil {
		return err
	}
	_, err = r.client.Put(ctx, instanceKey, string(val), clientv3.WithLease(r.sess.Lease()))
	return err
}

func (r *Registry) Unregister(ctx context.Context, serviceName string) error {
	serviceKey := fmt.Sprintf("micro/%s", serviceName)
	_, err := r.client.Delete(ctx, serviceKey)
	return err
}

func (r *Registry) ListerServer(ctx context.Context, serviceName string) ([]registry.ServiceInstance, error) {
	serviceKey := fmt.Sprintf("micro/%s", serviceName)

	kvs, err := r.client.Get(clientv3.WithRequireLeader(ctx), serviceKey, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	instances := make([]registry.ServiceInstance, 0, len(kvs.Kvs))
	for _, v := range kvs.Kvs {
		var ins registry.ServiceInstance
		err = json.Unmarshal(v.Value, &ins)
		if err != nil {
			fmt.Println(err)
			continue
		}

		instances = append(instances, ins)
	}

	return instances, nil
}

func (r *Registry) Subscribe(ctx context.Context, serviceName string) (<-chan registry.Event, error) {
	serviceKey := fmt.Sprintf("micro/%s", serviceName)
	ctxNew, cancel := context.WithCancel(ctx)
	r.mutex.Lock()
	r.watchCancel = append(r.watchCancel, cancel)
	r.mutex.Unlock()
	ch := r.client.Watch(ctxNew, serviceKey, clientv3.WithPrefix())
	res := make(chan registry.Event, 1)
	go func() {
		for eventCh := range ch {
			if eventCh.Canceled {
				return
			}
			if eventCh.Err() != nil {
				return
			}
			for _, event := range eventCh.Events {
				var ins registry.ServiceInstance
				err := json.Unmarshal(event.Kv.Value, &ins)
				if err != nil {
					continue
				}
				select {
				case res <- registry.Event{
					Type:     typesMap[event.Type],
					Instance: ins,
				}:
				case <-ctxNew.Done():
					return
				}

			}
		}
	}()

	return res, nil
}

func (r *Registry) Close() error {
	r.mutex.RLock()
	watchCancel := r.watchCancel
	r.mutex.RUnlock()
	for _, cancel := range watchCancel {
		cancel()
	}

	r.sess.Close()
	r.client.Close()

	return nil
}
