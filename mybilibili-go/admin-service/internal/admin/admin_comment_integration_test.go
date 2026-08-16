package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleCommentAdminListAndManage(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()
	seedBase(t, db)

	if _, err := db.ExecContext(ctx, `DELETE FROM comments`); err != nil {
		t.Fatalf("clean comments: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO comments (id, manuscript_id, user_id, content, status) VALUES (910001, 900001, 900001, 'hello spam', 0), (910002, 900001, 900001, 'normal one', 0)`); err != nil {
		t.Fatalf("seed comments: %v", err)
	}

	h := NewAdminDataHandler(db)
	// list
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/comment/list?keyword=spam", nil)
	h.handleCommentAdmin(rec, req)
	if rec.Code != 200 {
		t.Fatalf("list status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "hello spam") {
		t.Fatalf("list body missing 'hello spam': %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "normal one") {
		t.Fatalf("list should not contain 'normal one' with keyword=spam: %s", rec.Body.String())
	}
	var total int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments`).Scan(&total); err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 comments, got %d", total)
	}

	// status update
	rec = httptest.NewRecorder()
	body := `{"status":1}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/comment/910001/status", strings.NewReader(body))
	h.handleCommentAdmin(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status update code %d", rec.Code)
	}
	var st int
	if err := db.QueryRowContext(ctx, `SELECT status FROM comments WHERE id=910001`).Scan(&st); err != nil {
		t.Fatalf("query status: %v", err)
	}
	if st != 1 {
		t.Fatalf("expected status 1, got %d", st)
	}

	// get by id
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/comment/910002", nil)
	h.handleCommentAdmin(rec, req)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "normal one") {
		t.Fatalf("get by id failed: code %d body %s", rec.Code, rec.Body.String())
	}

	// delete
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/admin/comment/910002/delete", nil)
	h.handleCommentAdmin(rec, req)
	if rec.Code != 200 {
		t.Fatalf("delete code %d", rec.Code)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments WHERE id=910002`).Scan(&n); err != nil {
		t.Fatalf("query after delete: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected comment deleted, count %d", n)
	}
}
