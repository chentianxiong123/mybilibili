package support

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type Ticket struct {
	ID              int64      `json:"id"`
	TicketNo        string     `json:"ticket_no"`
	UserID          int64      `json:"user_id"`
	SessionID       int64      `json:"session_id"`
	Source          string     `json:"source"`
	Category        string     `json:"category"`
	Priority        string     `json:"priority"`
	Status          string     `json:"status"`
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	EntryReply      string     `json:"entry_reply"`
	AdminReply      string     `json:"admin_reply"`
	AssigneeAdminID int64      `json:"assignee_admin_id"`
	ProcessedAt     *time.Time `json:"processed_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func genTicketNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("TK%s", hex.EncodeToString(b))
}

func (r *Repository) Create(ctx context.Context, userID int64, source, category, priority, title, content string) (*Ticket, error) {
	t := &Ticket{
		TicketNo: genTicketNo(), UserID: userID, Source: source,
		Category: category, Priority: priority, Status: "PENDING",
		Title: title, Content: content,
	}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO support_tickets (ticket_no, user_id, source, category, priority, title, content)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		t.TicketNo, t.UserID, t.Source, t.Category, t.Priority, t.Title, t.Content,
	).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*Ticket, error) {
	t := &Ticket{}
	var st sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT id, ticket_no, COALESCE(user_id,0), COALESCE(session_id,0), COALESCE(source,''),
		        COALESCE(category,''), COALESCE(priority,''), status, title, COALESCE(content,''),
		        COALESCE(entry_reply,''), COALESCE(admin_reply,''), COALESCE(assignee_admin_id,0),
		        processed_at, created_at
		 FROM support_tickets WHERE id = $1`, id,
	).Scan(&t.ID, &t.TicketNo, &t.UserID, &t.SessionID, &t.Source,
		&t.Category, &t.Priority, &t.Status, &t.Title, &t.Content,
		&t.EntryReply, &t.AdminReply, &t.AssigneeAdminID, &st, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	if st.Valid {
		t.ProcessedAt = &st.Time
	}
	return t, nil
}

func (r *Repository) List(ctx context.Context, status string, page, size int32) ([]*Ticket, int64, error) {
	offset := (page - 1) * size
	query := `SELECT id, ticket_no, COALESCE(user_id,0), COALESCE(session_id,0), COALESCE(source,''), COALESCE(category,''), COALESCE(priority,''), status, title, COALESCE(content,''), COALESCE(entry_reply,''), COALESCE(admin_reply,''), COALESCE(assignee_admin_id,0), created_at, processed_at FROM support_tickets`
	countQuery := `SELECT COUNT(*) FROM support_tickets`
	args := []interface{}{size, offset}
	countArgs := []interface{}{}
	if status != "" {
		query += ` WHERE status = $1`
		countQuery += ` WHERE status = $1`
		args = append([]interface{}{status}, args...)
		countArgs = []interface{}{status}
	}
	query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	var total int64
	r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*Ticket, 0)
	for rows.Next() {
		t := &Ticket{}
		rows.Scan(&t.ID, &t.TicketNo, &t.UserID, &t.SessionID, &t.Source, &t.Category, &t.Priority, &t.Status, &t.Title, &t.Content, &t.EntryReply, &t.AdminReply, &t.AssigneeAdminID, &t.CreatedAt, &t.ProcessedAt)
		list = append(list, t)
	}
	return list, total, nil
}

func (r *Repository) Process(ctx context.Context, id int64, adminID int64, reply string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE support_tickets SET status = 'PROCESSED', admin_reply = $1, assignee_admin_id = $2, processed_at = NOW() WHERE id = $3`,
		reply, adminID, id)
	return err
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int64, title, content string) (*Ticket, error) {
	return s.repo.Create(ctx, userID, "USER_FEEDBACK", "GENERAL", "NORMAL", title, content)
}

func (s *Service) List(ctx context.Context, status string, page, size int32) ([]*Ticket, int64, error) {
	return s.repo.List(ctx, status, page, size)
}

func (s *Service) Process(ctx context.Context, id, adminID int64, reply string) error {
	return s.repo.Process(ctx, id, adminID, reply)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Ticket, error) {
	return s.repo.GetByID(ctx, id)
}
