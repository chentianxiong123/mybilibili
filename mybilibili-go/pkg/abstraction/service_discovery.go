package abstraction

import "context"

type ServiceInstance struct {
	ServiceName string
	InstanceID  string
	Host        string
	Port        int
	Metadata    map[string]string
	Healthy     bool
}

type ServiceDiscovery interface {
	Register(ctx context.Context, instance ServiceInstance) error
	Deregister(ctx context.Context) error
	Discover(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Watch(ctx context.Context, serviceName string) (<-chan []ServiceInstance, error)
	Healthy(ctx context.Context) error
}
