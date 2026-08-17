package studio

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type ExportTask struct {
	ID           string
	TaskKey      string
	TaskName     string
	UserID       int64
	ProjectID    string
	Status       string
	Progress     int32
	Message      string
	OutputURL    string
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTask(ctx context.Context, userID int64, projectID string) (*ExportTask, error) {
	t := &ExportTask{UserID: userID, ProjectID: projectID, Status: "PENDING", Progress: 0}
	// Uses a simple UUID-like ID from the task number
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO operation_tasks (task_key, task_type, task_name, target_type, target_id, status, operator_id)
		 VALUES ($1, 'studio_export', 'Studio Export', 'project', $2, 'PENDING', $3) RETURNING id`,
		"studio_"+projectID+"_"+itoa(time.Now().UnixNano()), projectID, userID).Scan(&id)
	if err != nil {
		return nil, err
	}
	t.ID = itoa(id)
	return t, nil
}

func (r *Repository) GetTask(ctx context.Context, taskID string) (*ExportTask, error) {
	t := &ExportTask{}
	var id int64
	err := r.db.QueryRowContext(ctx,
		`SELECT id, task_key, task_name, status, progress, message, error_message, created_at, updated_at
		 FROM operation_tasks WHERE id = $1`, taskID).Scan(&id, &t.TaskKey, &t.TaskName, &t.Status, &t.Progress, &t.Message, &t.ErrorMessage, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	t.ID = itoa(id)
	return t, nil
}

func (r *Repository) CancelTask(ctx context.Context, taskID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE operation_tasks SET status = 'CANCELLED' WHERE id = $1 AND status = 'PENDING'`, taskID)
	return err
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTask(ctx context.Context, userID int64, projectID string) (*ExportTask, error) {
	return s.repo.CreateTask(ctx, userID, projectID)
}

func (s *Service) GetTask(ctx context.Context, taskID string) (*ExportTask, error) {
	return s.repo.GetTask(ctx, taskID)
}

func (s *Service) CancelTask(ctx context.Context, taskID string) error {
	return s.repo.CancelTask(ctx, taskID)
}

func (s *Service) UploadAsset(ctx context.Context, userID, taskID int64, assetType, filename string, reader io.Reader) (string, error) {
	base := os.Getenv("STUDIO_DATA_DIR")
	if base == "" {
		base = "/tmp/studio-data"
	}
	dir := filepath.Join(base, "assets", fmt.Sprintf("u%d", userID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	safeName := filepath.Base(filename)
	dst := filepath.Join(dir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName))
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/assets/studio/u%d/%s", userID, filepath.Base(dst))
	_ = taskID
	_ = assetType
	return url, nil
}

var _ = context.Background
