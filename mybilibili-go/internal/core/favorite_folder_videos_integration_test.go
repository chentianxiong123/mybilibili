package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleFolderVideos(t *testing.T) {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		t.Skip("PG_DSN not set; skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	_, _ = db.ExecContext(ctx, `TRUNCATE favorite_folder_videos, favorite_folders, manuscripts, categories, users RESTART IDENTITY CASCADE`)
	_, _ = db.ExecContext(ctx, `INSERT INTO users (id, username, password) VALUES (930001, 'fav_video', 'x') ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO categories (id, name) VALUES (930001, 'fav_cat') ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO manuscripts (id, title, user_id, category_id) VALUES (930001, 'ms1', 930001, 930001), (930002, 'ms2', 930001, 930001) ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO favorite_folders (id, user_id, name) VALUES (930001, 930001, 'test_folder') ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO favorite_folder_videos (folder_id, manuscript_id) VALUES (930001, 930001), (930001, 930002) ON CONFLICT DO NOTHING`)

	h := NewFavoriteHandler(db)
	// GET folder videos
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/favorites/930001/videos?page=1&page_size=10", nil)
	req.Header.Set("X-User-Id", "930001")
	h.handleByID(rec, req)
	if rec.Code != 200 {
		t.Fatalf("GET videos status %d body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		List []map[string]interface{} `json:"list"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.List) != 2 {
		t.Fatalf("expected 2 videos, got %d: %s", len(resp.List), rec.Body.String())
	}

	// PUT update folder videos (replace with only ms1)
	rec = httptest.NewRecorder()
	body := `{"manuscript_ids":[930001]}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/favorites/930001/videos", strings.NewReader(body))
	req.Header.Set("X-User-Id", "930001")
	h.handleByID(rec, req)
	if rec.Code != 200 {
		t.Fatalf("PUT videos status %d", rec.Code)
	}
	var cnt int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM favorite_folder_videos WHERE folder_id=930001`).Scan(&cnt)
	if cnt != 1 {
		t.Fatalf("expected 1 video after update, got %d", cnt)
	}
}