package abstraction

import (
	"context"
	"sync"
)

type memoryDiscovery struct {
	mu        sync.RWMutex
	instances map[string][]ServiceInstance
}

func newMemoryDiscovery() *memoryDiscovery {
	return &memoryDiscovery{instances: make(map[string][]ServiceInstance)}
}

func (m *memoryDiscovery) Register(ctx context.Context, instance ServiceInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[instance.ServiceName] = append(m.instances[instance.ServiceName], instance)
	return nil
}

func (m *memoryDiscovery) Deregister(ctx context.Context) error { return nil }

func (m *memoryDiscovery) Discover(ctx context.Context, serviceName string) ([]ServiceInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.instances[serviceName], nil
}

func (m *memoryDiscovery) Watch(ctx context.Context, serviceName string) (<-chan []ServiceInstance, error) {
	ch := make(chan []ServiceInstance, 1)
	return ch, nil
}

func (m *memoryDiscovery) Healthy(ctx context.Context) error { return nil }