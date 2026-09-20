package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepository(db), mock
}

func TestCreate(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice", "pass", "Alice", "a@b.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	u := &User{Username: "alice", Password: "pass", Nickname: "Alice", Email: "a@b.com"}
	id, err := repo.Create(context.Background(), u)
	require.NoError(t, err)
	assert.Equal(t, int64(42), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`INSERT INTO users`).WillReturnError(errors.New("db down"))
	_, err := repo.Create(context.Background(), &User{})
	assert.Error(t, err)
}

func TestFindByUsername(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, username, password, nickname, COALESCE\(email`).WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "password", "nickname", "email", "avatar", "level", "experience",
			"signature", "bio", "follower_count", "following_count", "liked_count", "status",
			"coin_count", "gender", "created_at", "updated_at",
		}).AddRow(1, "alice", "hash", "Alice", "a@b.com", "avatar.png", 3, 100,
			"hello", "bio text", 10, 5, 8, 1, 20, 2, now, now))

	u, err := repo.FindByUsername(context.Background(), "alice")
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, "alice", u.Username)
	assert.Equal(t, "hash", u.Password)
	assert.Equal(t, "Alice", u.Nickname)
	assert.Equal(t, "a@b.com", u.Email)
	assert.Equal(t, "avatar.png", u.Avatar)
	assert.Equal(t, int32(3), u.Level)
	assert.Equal(t, int64(100), u.Experience)
	assert.Equal(t, "hello", u.Signature)
	assert.Equal(t, "bio text", u.Bio)
	assert.Equal(t, int32(10), u.FollowerCount)
	assert.Equal(t, int32(5), u.FollowingCount)
	assert.Equal(t, int32(8), u.LikedCount)
	assert.Equal(t, int32(1), u.Status)
	assert.Equal(t, int32(20), u.CoinCount)
	assert.Equal(t, int32(2), u.Gender)
	assert.Equal(t, now, u.CreatedAt)
	assert.Equal(t, now, u.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByUsername_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, username, password, nickname`).WillReturnError(sql.ErrNoRows)
	_, err := repo.FindByUsername(context.Background(), "nobody")
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestFindByNickname(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, username, password, nickname, COALESCE\(email`).WithArgs("Alice").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "password", "nickname", "email", "avatar", "level", "experience",
			"signature", "bio", "follower_count", "following_count", "liked_count", "status",
			"coin_count", "gender", "created_at", "updated_at",
		}).AddRow(1, "alice", "hash", "Alice", "a@b.com", "avatar.png", 3, 100,
			"hello", "bio text", 10, 5, 8, 1, 20, 2, now, now))

	u, err := repo.FindByNickname(context.Background(), "Alice")
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, "alice", u.Username)
	assert.Equal(t, "hash", u.Password)
	assert.Equal(t, "Alice", u.Nickname)
	assert.Equal(t, "a@b.com", u.Email)
	assert.Equal(t, "avatar.png", u.Avatar)
	assert.Equal(t, int32(3), u.Level)
	assert.Equal(t, int64(100), u.Experience)
	assert.Equal(t, "hello", u.Signature)
	assert.Equal(t, "bio text", u.Bio)
	assert.Equal(t, int32(10), u.FollowerCount)
	assert.Equal(t, int32(5), u.FollowingCount)
	assert.Equal(t, int32(8), u.LikedCount)
	assert.Equal(t, int32(1), u.Status)
	assert.Equal(t, int32(20), u.CoinCount)
	assert.Equal(t, int32(2), u.Gender)
	assert.Equal(t, now, u.CreatedAt)
	assert.Equal(t, now, u.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByNickname_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, username, password, nickname`).WillReturnError(sql.ErrNoRows)
	_, err := repo.FindByNickname(context.Background(), "nobody")
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestFindByID(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, username, password, nickname, COALESCE\(email`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "password", "nickname", "email", "avatar", "level", "experience",
			"signature", "bio", "follower_count", "following_count", "liked_count", "status",
			"coin_count", "gender", "created_at", "updated_at",
		}).AddRow(1, "alice", "hash", "Alice", "a@b.com", "avatar.png", 3, 100,
			"hello", "bio text", 10, 5, 8, 1, 20, 2, now, now))

	u, err := repo.FindByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, "alice", u.Username)
	assert.Equal(t, "hash", u.Password)
	assert.Equal(t, "Alice", u.Nickname)
	assert.Equal(t, "a@b.com", u.Email)
	assert.Equal(t, "avatar.png", u.Avatar)
	assert.Equal(t, int32(3), u.Level)
	assert.Equal(t, int64(100), u.Experience)
	assert.Equal(t, "hello", u.Signature)
	assert.Equal(t, "bio text", u.Bio)
	assert.Equal(t, int32(10), u.FollowerCount)
	assert.Equal(t, int32(5), u.FollowingCount)
	assert.Equal(t, int32(8), u.LikedCount)
	assert.Equal(t, int32(1), u.Status)
	assert.Equal(t, int32(20), u.CoinCount)
	assert.Equal(t, int32(2), u.Gender)
	assert.Equal(t, now, u.CreatedAt)
	assert.Equal(t, now, u.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, username, password, nickname`).WillReturnError(sql.ErrNoRows)
	_, err := repo.FindByID(context.Background(), 99)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestToPB(t *testing.T) {
	repo, _ := newMockRepo(t)
	now := time.Now()
	u := &User{
		ID: 1, Username: "alice", Nickname: "Alice", Email: "a@b.com",
		Avatar: "avatar.png", Level: 3, CreatedAt: now,
	}

	pb := repo.ToPB(u)
	assert.Equal(t, int64(1), pb.UserId)
	assert.Equal(t, "alice", pb.Username)
	assert.Equal(t, "Alice", pb.Nickname)
	assert.Equal(t, "a@b.com", pb.Email)
	assert.Equal(t, "avatar.png", pb.Avatar)
	assert.Equal(t, int32(3), pb.Level)
	assert.Equal(t, now.Format(time.RFC3339), pb.CreatedAt)
}
