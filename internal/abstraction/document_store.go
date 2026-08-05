package abstraction

import "context"

type DocumentStore interface {
	Insert(ctx context.Context, collection string, doc any) (string, error)
	FindByID(ctx context.Context, collection, id string, result any) error
	Update(ctx context.Context, collection, id string, doc any) error
	Delete(ctx context.Context, collection, id string) error
	Query(ctx context.Context, collection string, filter QueryFilter, result any) error
}

type QueryFilter struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
	Filters  map[string]any
}