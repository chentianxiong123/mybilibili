package social

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		t.Skip("PG_DSN not set; skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func TestShareStatisticsAndWatchHistory(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()

	_, _ = db.ExecContext(ctx, `TRUNCATE shares, watch_history, manuscripts, categories, users RESTART IDENTITY CASCADE`)
	_, _ = db.ExecContext(ctx, `INSERT INTO users (id, username, password) VALUES (940001, 'share_u', 'x') ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO categories (id, name) VALUES (940001, 'share_c') ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO manuscripts (id, title, user_id, category_id) VALUES (940001, 'share_ms', 940001, 940001) ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO shares (user_id, manuscript_id, channel) VALUES (940001, 940001, 'wechat'), (940001, 940001, 'weibo'), (940001, 940001, 'wechat')`)

	shareRepo := NewShareRepository(db)
	_ = shareRepo
	h := &SocialHandler{shareRepo: shareRepo}

	// GET /api/v1/share/statistics?manuscript_id=940001
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/share/statistics?manuscript_id=940001", nil)
	h.handleShare(rec, req)
	if rec.Code != 200 {
		t.Fatalf("share statistics status %d body %s", rec.Code, rec.Body.String())
	}
	var stats map[string]int64
	json.Unmarshal(rec.Body.Bytes(), &stats)
	if stats["wechat"] != 2 {
		t.Fatalf("expected wechat=2, got %v", stats)
	}
	if stats["weibo"] != 1 {
		t.Fatalf("expected weibo=1, got %v", stats)
	}

	// watch history: POST then single DELETE
	_, _ = db.ExecContext(ctx, `INSERT INTO watch_history (user_id, manuscript_id, progress_seconds) VALUES (940001, 940001, 30) ON CONFLICT DO NOTHING`)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/watch-history/940001", nil)
	req.Header.Set("X-User-Id", "940001")
	h.handleWatchHistory(rec, req)
	if rec.Code != 200 {
		t.Fatalf("single delete watch history status %d", rec.Code)
	}

	var cnt int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM watch_history WHERE user_id=940001`).Scan(&cnt)
	if cnt != 0 {
		t.Fatalf("expected 0 watch history after single delete, got %d", cnt)
	}

	// re-insert and test clear-all
	_, _ = db.ExecContext(ctx, `INSERT INTO watch_history (user_id, manuscript_id, progress_seconds) VALUES (940001, 940001, 30) ON CONFLICT DO NOTHING`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/watch-history/", nil)
	req.Header.Set("X-User-Id", "940001")
	h.handleWatchHistory(rec, req)
	if rec.Code != 200 {
		t.Fatalf("clear all watch history status %d", rec.Code)
	}
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM watch_history WHERE user_id=940001`).Scan(&cnt)
	if cnt != 0 {
		t.Fatalf("expected 0 watch history after clear, got %d", cnt)
	}
}