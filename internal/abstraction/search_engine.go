package abstraction

import "context"

type SearchEngine interface {
	Index(ctx context.Context, index string, docID string, doc any) error
	Delete(ctx context.Context, index string, docID string) error
	Search(ctx context.Context, index string, query string, opts SearchOptions) (*SearchResult, error)
	BulkIndex(ctx context.Context, index string, docs map[string]any) error
}

type SearchOptions struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
	Fields   []string
}

type SearchResult struct {
	Total int64
	Hits  []SearchHit
}

type SearchHit struct {
	ID      string
	Score   float64
	Source  any
}