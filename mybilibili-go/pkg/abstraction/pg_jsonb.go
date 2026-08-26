package abstraction

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type pgJSONB struct {
	db *sql.DB
}

func newPGJSONB(cfg DocumentStoreConfig) (DocumentStore, error) {
	dsn := cfg.DSN
	if dsn == "" {
		dsn = cfg.Path
	}
	if dsn == "" {
		return nil, fmt.Errorf("pg-jsonb: DSN required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("pg-jsonb: open db: %w", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS jsonb_documents (
		id TEXT NOT NULL,
		collection_name TEXT NOT NULL,
		doc_json JSONB NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (collection_name, id)
	)`)
	if err != nil {
		return nil, fmt.Errorf("pg-jsonb: create table: %w", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_jsonb_documents_gin ON jsonb_documents USING GIN(doc_json)`)
	if err != nil {
		return nil, fmt.Errorf("pg-jsonb: create index: %w", err)
	}
	return &pgJSONB{db: db}, nil
}

func (p *pgJSONB) Insert(ctx context.Context, collection string, doc any) (string, error) {
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	var m map[string]any
	json.Unmarshal(jsonBytes, &m)
	id := ""
	if v, ok := m["id"]; ok {
		id = fmt.Sprintf("%v", v)
	}
	if id == "" {
		id = fmt.Sprintf("%d", time.Now().UnixNano())
		m["id"] = id
		jsonBytes, _ = json.Marshal(m)
	}
	_, err = p.db.ExecContext(ctx,
		`INSERT INTO jsonb_documents (id, collection_name, doc_json) VALUES ($1, $2, $3)
		 ON CONFLICT (collection_name, id) DO UPDATE SET doc_json = $3, updated_at = NOW()`,
		id, collection, string(jsonBytes))
	return id, err
}

func (p *pgJSONB) FindByID(ctx context.Context, collection, id string, result any) error {
	var jsonBytes string
	err := p.db.QueryRowContext(ctx,
		`SELECT doc_json FROM jsonb_documents WHERE collection_name = $1 AND id = $2`,
		collection, id).Scan(&jsonBytes)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonBytes), result)
}

func (p *pgJSONB) Update(ctx context.Context, collection, id string, doc any) error {
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx,
		`UPDATE jsonb_documents SET doc_json = $1, updated_at = NOW() WHERE collection_name = $2 AND id = $3`,
		string(jsonBytes), collection, id)
	return err
}

func (p *pgJSONB) Delete(ctx context.Context, collection, id string) error {
	_, err := p.db.ExecContext(ctx,
		`DELETE FROM jsonb_documents WHERE collection_name = $1 AND id = $2`, collection, id)
	return err
}

func (p *pgJSONB) Query(ctx context.Context, collection string, filter QueryFilter, result any) error {
	where := "WHERE collection_name = $1"
	args := []interface{}{collection}
	argIdx := 1
	for k, v := range filter.Filters {
		argIdx++
		where += fmt.Sprintf(" AND doc_json->>'%s' = $%d", escapeJSONKey(k), argIdx)
		args = append(args, fmt.Sprintf("%v", v))
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.PageSize
	orderBy := "created_at DESC"
	if filter.Sort != "" {
		orderBy = fmt.Sprintf("doc_json->>'%s' %s", escapeJSONKey(filter.Sort), filter.Order)
		if filter.Order == "" {
			orderBy = fmt.Sprintf("doc_json->>'%s' ASC", escapeJSONKey(filter.Sort))
		}
	}
	args = append(args, filter.PageSize, offset)
	query := fmt.Sprintf(
		`SELECT id, doc_json FROM jsonb_documents %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		where, orderBy, argIdx+1, argIdx+2)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	var list []any
	for rows.Next() {
		var rowID, jsonBytes string
		if err := rows.Scan(&rowID, &jsonBytes); err != nil {
			continue
		}
		var doc any
		json.Unmarshal([]byte(jsonBytes), &doc)
		if m, ok := doc.(map[string]any); ok {
			if v, has := m["id"]; !has || v == nil || fmt.Sprintf("%v", v) == "" {
				m["id"] = rowID
			}
		}
		list = append(list, doc)
	}
	outBytes, _ := json.Marshal(list)
	return json.Unmarshal(outBytes, result)
}

func escapeJSONKey(k string) string {
	return strings.ReplaceAll(k, "'", "''")
}