package core

import (
	"context"
	"database/sql"
	"time"

	pb "mybilibili/internal/core/pb"
)

type User struct {
	ID        int64
	Username  string
	Password  string
	Nickname  string
	Email     string
	Avatar    string
	Level     int32
	Status    int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, u *User) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (username, password, nickname, email) VALUES ($1, $2, $3, $4) RETURNING id`,
		u.Username, u.Password, u.Nickname, u.Email,
	).Scan(&id)
	return id, err
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	u := &User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password, nickname, email, avatar, level, status, created_at, updated_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Nickname, &u.Email, &u.Avatar, &u.Level, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*User, error) {
	u := &User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password, nickname, email, avatar, level, status, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Nickname, &u.Email, &u.Avatar, &u.Level, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) ToPB(u *User) *pb.GetUserResponse {
	return &pb.GetUserResponse{
		UserId:    u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Level:     u.Level,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}