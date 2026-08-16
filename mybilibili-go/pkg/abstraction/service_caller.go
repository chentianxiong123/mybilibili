package abstraction

import "context"

type ServiceCaller interface {
	Call(ctx context.Context, target, method string, req, resp any) error
	CallStream(ctx context.Context, target, method string, req any) (<-chan []byte, error)
	Close() error
}
