package admin

import (
	"context"
	"database/sql"
	"os"
	"testing"
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

func seedBase(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `TRUNCATE audit_logs, manuscript_status_events, manuscript_edit_versions, video_process_events, manuscript_daily_metrics, videos, manuscripts, categories, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, username, password) VALUES (900001, 'evt_user', 'x') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO categories (id, name) VALUES (900001, 'evt_cat') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO manuscripts (id, title, user_id, category_id, status, review_status)
		 VALUES (900001, 'evt_ms', 900001, 900001, 0, 0) ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("seed manuscript: %v", err)
	}
}

func TestSQLWriterRecordStatusEvent(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()
	seedBase(t, db)

	w := NewManuscriptEventWriter(db)
	if err := w.RecordStatusEvent(ctx, 900001, 900001, 0, 3, "PUBLISH", "ADMIN", 1, "test publish"); err != nil {
		t.Fatalf("RecordStatusEvent: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM manuscript_status_events WHERE manuscript_id = 900001 AND action = 'PUBLISH' AND to_status = 3`).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 status event, got %d", n)
	}
}

func TestSQLWriterRecordVideoProcessEvent(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()
	seedBase(t, db)

	_, err := db.ExecContext(ctx,
		`INSERT INTO videos (id, manuscript_id, title, process_status)
		 VALUES (900001, 900001, 'evt_video', 0) ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("seed video: %v", err)
	}

	w := NewManuscriptEventWriter(db)
	if err := w.RecordVideoProcessEvent(ctx, 900001, 900001, 0, 1, "TRANSCODING", 100); err != nil {
		t.Fatalf("RecordVideoProcessEvent: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM video_process_events WHERE video_id = 900001 AND stage = 'TRANSCODING'`).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 process event, got %d", n)
	}
}

func TestSQLWriterRecordEditVersion(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()
	seedBase(t, db)

	w := NewManuscriptEventWriter(db)
	if err := w.RecordEditVersion(ctx, 900001, 900001, "", "approved", "review_status,status"); err != nil {
		t.Fatalf("RecordEditVersion: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM manuscript_edit_versions WHERE manuscript_id = 900001 AND status = 'PENDING'`).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 edit version, got %d", n)
	}
}

func TestSQLWriterUpsertDailyMetric(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()
	seedBase(t, db)

	db.ExecContext(ctx, `INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, view_count)
		VALUES (CURRENT_DATE, 900001, 900001, 5) ON CONFLICT (metric_date, manuscript_id) DO NOTHING`)

	if err := upsertDailyMetric(db, ctx, 900001, 900001, "view_count", 1); err != nil {
		t.Fatalf("upsertDailyMetric: %v", err)
	}

	var v int
	if err := db.QueryRowContext(ctx,
		`SELECT view_count FROM manuscript_daily_metrics WHERE manuscript_id = 900001 AND metric_date = CURRENT_DATE`).Scan(&v); err != nil {
		t.Fatalf("query: %v", err)
	}
	if v != 6 {
		t.Fatalf("want view_count 6, got %d", v)
	}
}

func upsertDailyMetric(db *sql.DB, ctx context.Context, manuscriptID, userID int64, field string, delta int) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, `+field+`)
		 VALUES (CURRENT_DATE, $1, $2, $3)
		 ON CONFLICT (metric_date, manuscript_id)
		 DO UPDATE SET `+field+` = manuscript_daily_metrics.`+field+` + $3, updated_at = NOW()`,
		manuscriptID, userID, delta)
	return err
}

func TestServiceRecordAuditWritesLog(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()
	seedBase(t, db)

	repo := NewRepository(db)
	svc := NewService(repo)
	if err := svc.RecordAudit(ctx, 900001, "tester", "role", "UPDATE_ROLE", "roles", "900001", 0, "更新角色", ""); err != nil {
		t.Fatalf("RecordAudit: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE action = 'UPDATE_ROLE' AND target_id = '900001'`).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 audit log, got %d", n)
	}
}
