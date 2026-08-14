package core

import (
	"context"
	"database/sql"
	"os"
	"testing"
)

func TestCommentRepositoryBatchQueries(t *testing.T) {
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

	if _, err := db.ExecContext(ctx, `TRUNCATE comments, user_interactions, manuscripts, categories, users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id, username, password) VALUES (920001, 'batch_user', 'x') ON CONFLICT (id) DO NOTHING`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO categories (id, name) VALUES (920001, 'batch_cat') ON CONFLICT (id) DO NOTHING`); err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO manuscripts (id, title, user_id, category_id) VALUES (920001, 'batch_ms', 920001, 920001) ON CONFLICT (id) DO NOTHING`); err != nil {
		t.Fatalf("seed manuscript: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO comments (id, manuscript_id, user_id, content, like_count) VALUES (920001, 920001, 920001, 'a', 5), (920002, 920001, 920001, 'b', 0)`); err != nil {
		t.Fatalf("seed comments: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type) VALUES (920001, 'COMMENT', 920001, 'like')`); err != nil {
		t.Fatalf("seed interaction: %v", err)
	}

	repo := NewCommentRepository(db)
	counts, err := repo.BatchGetLikeCounts(ctx, "COMMENT", []int64{920001, 920002, 999999})
	if err != nil {
		t.Fatalf("BatchGetLikeCounts: %v", err)
	}
	if counts[920001] != 5 {
		t.Fatalf("expected like_count 5 for 920001, got %d", counts[920001])
	}
	if counts[920002] != 0 {
		t.Fatalf("expected like_count 0 for 920002, got %d", counts[920002])
	}
	if _, ok := counts[999999]; ok {
		t.Fatalf("nonexistent id should be absent")
	}

	liked, err := repo.BatchIsLiked(ctx, 920001, "COMMENT", []int64{920001, 920002})
	if err != nil {
		t.Fatalf("BatchIsLiked: %v", err)
	}
	if !liked[920001] {
		t.Fatalf("expected 920001 liked")
	}
	if liked[920002] {
		t.Fatalf("expected 920002 not liked")
	}
}