package abstraction

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type pgFTS struct {
	db *sql.DB
}

func newPGFTS(cfg SearchEngineConfig) (SearchEngine, error) {
	dsn := cfg.DSN
	if dsn == "" && len(cfg.Addresses) > 0 {
		dsn = cfg.Addresses[0]
	}
	if dsn == "" {
		return nil, fmt.Errorf("pg-fts: DSN required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("pg-fts: open db: %w", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS search_documents (
		id TEXT NOT NULL,
		index_name TEXT NOT NULL,
		doc_json JSONB NOT NULL,
		search_vector TSVECTOR,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (index_name, id)
	)`)
	if err != nil {
		return nil, fmt.Errorf("pg-fts: create table: %w", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_search_documents_gin ON search_documents USING GIN(search_vector)`)
	if err != nil {
		return nil, fmt.Errorf("pg-fts: create index: %w", err)
	}
	return &pgFTS{db: db}, nil
}

func (p *pgFTS) Index(ctx context.Context, index string, docID string, doc any) error {
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	vals := []string{}
	m := doc.(map[string]interface{})
	for _, v := range m {
		s, ok := v.(string)
		if ok {
			vals = append(vals, s)
		}
	}
	searchText := strings.Join(vals, " ")
	_, err = p.db.ExecContext(ctx,
		`INSERT INTO search_documents (id, index_name, doc_json, search_vector)
		 VALUES ($1, $2, $3, to_tsvector('simple', $4))
		 ON CONFLICT (index_name, id) DO UPDATE SET doc_json = $3, search_vector = to_tsvector('simple', $4)`,
		docID, index, string(jsonBytes), searchText)
	return err
}

func (p *pgFTS) Delete(ctx context.Context, index string, docID string) error {
	_, err := p.db.ExecContext(ctx,
		`DELETE FROM search_documents WHERE index_name = $1 AND id = $2`, index, docID)
	return err
}

func (p *pgFTS) Search(ctx context.Context, index string, query string, opts SearchOptions) (*SearchResult, error) {
	if opts.PageSize < 1 {
		opts.PageSize = 20
	}
	if opts.Page < 1 {
		opts.Page = 1
	}
	offset := (opts.Page - 1) * opts.PageSize

	tsQuery := strings.Join(strings.Fields(query), " & ")

	var total int64
	err := p.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM search_documents
		 WHERE index_name = $1 AND search_vector @@ to_tsquery('simple', $2)`,
		index, tsQuery).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := p.db.QueryContext(ctx,
		`SELECT id, doc_json, ts_rank(search_vector, to_tsquery('simple', $2)) as score
		 FROM search_documents
		 WHERE index_name = $1 AND search_vector @@ to_tsquery('simple', $2)
		 ORDER BY score DESC LIMIT $3 OFFSET $4`,
		index, tsQuery, opts.PageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hits []SearchHit
	for rows.Next() {
		var id string
		var jsonBytes string
		var score float64
		if err := rows.Scan(&id, &jsonBytes, &score); err != nil {
			continue
		}
		var source any
		json.Unmarshal([]byte(jsonBytes), &source)
		hits = append(hits, SearchHit{ID: id, Score: score, Source: source})
	}
	return &SearchResult{Total: total, Hits: hits}, nil
}

func (p *pgFTS) BulkIndex(ctx context.Context, index string, docs map[string]any) error {
	for id, doc := range docs {
		if err := p.Index(ctx, index, id, doc); err != nil {
			return err
		}
	}
	return nil
}