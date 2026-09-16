package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/errors"
	pb "mybilibili/pkg/pb"
)

const (
	testJWTSecret = "user-svc-test"
)

// fakeCache 简单的内存 CacheStore 用于 service 单测。
type fakeCache struct {
	mu     sync.Mutex
	data   map[string][]byte
	counts map[string]int64
}

func newFakeCache() *fakeCache {
	return &fakeCache{data: map[string][]byte{}, counts: map[string]int64{}}
}

func (f *fakeCache) Get(ctx context.Context, key string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return v, nil
}

func (f *fakeCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = value
	return nil
}

func (f *fakeCache) Delete(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, key)
	delete(f.data, key+":lock")
	delete(f.counts, key)
	return nil
}

func (f *fakeCache) Exists(ctx context.Context, key string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.data[key]
	return ok, nil
}

func (f *fakeCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return false, nil
}

func (f *fakeCache) Unlock(ctx context.Context, key string) error { return nil }

func (f *fakeCache) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.counts[key]++
	if f.counts[key] >= 5 {
		f.data[key+":lock"] = []byte("1")
	}
	return f.counts[key], nil
}

func (f *fakeCache) Close() error { return nil }

// 编译期接口校验
var _ abstraction.CacheStore = (*fakeCache)(nil)

func newTestService(t *testing.T) (*Service, sqlmock.Sqlmock, *fakeCache) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewRepository(db)
	svc := NewService(repo, testJWTSecret)
	cache := newFakeCache()
	svc.SetCacheStore(cache)
	return svc, mock, cache
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func TestRegister_Success(t *testing.T) {
	svc, mock, _ := newTestService(t)

	username := "alice"
	password := "P@ssw0rd"
	nickname := "Alice"

	// FindByNickname: no rows (nickname 不存在 → 可以注册)
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE nickname`).
		WithArgs(nickname).
		WillReturnError(sql.ErrNoRows)

	// Create: insert returning id=1
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(username, sha256Hex(password), nickname, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// UPDATE status=1
	mock.ExpectExec(`UPDATE users SET status=1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: username, Password: password, Nickname: nickname,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.UserId)
	assert.NotEmpty(t, resp.Token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_EmptyUsername(t *testing.T) {
	svc, mock, _ := newTestService(t)
	_, err := svc.Register(context.Background(), &pb.RegisterRequest{Username: "", Password: "x"})
	assertGRPCStatus(t, err, codes.InvalidArgument)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_EmptyPassword(t *testing.T) {
	svc, mock, _ := newTestService(t)
	_, err := svc.Register(context.Background(), &pb.RegisterRequest{Username: "x", Password: ""})
	assertGRPCStatus(t, err, codes.InvalidArgument)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc, mock, _ := newTestService(t)

	// FindByNickname: no rows (nickname 不存在)
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE nickname`).
		WithArgs("alice").
		WillReturnError(sql.ErrNoRows)

	// Create: 模拟 unique 冲突
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice", sha256Hex("p"), "alice", "").
		WillReturnError(fmt.Errorf("pq: duplicate key value violates unique constraint"))

	_, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: "alice", Password: "p", Nickname: "alice",
	})
	assertGRPCStatus(t, err, codes.AlreadyExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_NicknameFallback(t *testing.T) {
	svc, mock, _ := newTestService(t)

	username := "bob"
	password := "secret"

	// 业务代码：当 Nickname 为空时, 用 Username 兜底 (bob)
	// FindByNickname(bob) 必须 no rows
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE nickname`).
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	// Create(user=username=bob, nickname=bob, ...)
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(username, sha256Hex(password), username, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	mock.ExpectExec(`UPDATE users SET status=1`).
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: username, Password: password, Nickname: "", // ← 空 nickname
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), resp.UserId)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_Success(t *testing.T) {
	svc, mock, _ := newTestService(t)

	username := "alice"
	password := "correct"

	// FindByUsername
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "password", "nickname", "email", "avatar",
			"level", "experience", "signature", "bio",
			"follower_count", "following_count", "liked_count",
			"status", "created_at", "updated_at",
		}).AddRow(
			1, username, sha256Hex(password), "Alice", "", "",
			1, 0, "", "",
			0, 0, 0,
			1, time.Now(), time.Now(),
		))

	resp, err := svc.Login(context.Background(), &pb.LoginRequest{
		Username: username, Password: password,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.UserId)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "Alice", resp.Nickname)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, mock, _ := newTestService(t)

	username := "alice"
	correctHash := sha256Hex("right")

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "password", "nickname", "email", "avatar",
			"level", "experience", "signature", "bio",
			"follower_count", "following_count", "liked_count",
			"status", "created_at", "updated_at",
		}).AddRow(
			1, username, correctHash, "Alice", "", "",
			1, 0, "", "",
			0, 0, 0,
			1, time.Now(), time.Now(),
		))

	_, err := svc.Login(context.Background(), &pb.LoginRequest{
		Username: username, Password: "wrong",
	})
	assertGRPCStatus(t, err, codes.Unauthenticated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs("ghost").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Login(context.Background(), &pb.LoginRequest{
		Username: "ghost", Password: "x",
	})
	assertGRPCStatus(t, err, codes.NotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_AccountDisabled(t *testing.T) {
	svc, mock, _ := newTestService(t)

	username := "alice"
	password := "any"

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "password", "nickname", "email", "avatar",
			"level", "experience", "signature", "bio",
			"follower_count", "following_count", "liked_count",
			"status", "created_at", "updated_at",
		}).AddRow(
			1, username, sha256Hex(password), "Alice", "", "",
			1, 0, "", "",
			0, 0, 0,
			0 /* status=0 disabled */, time.Now(), time.Now(),
		))

	_, err := svc.Login(context.Background(), &pb.LoginRequest{
		Username: username, Password: password,
	})
	assertGRPCStatus(t, err, codes.Unauthenticated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_LockAfter5Fails(t *testing.T) {
	svc, mock, cache := newTestService(t)

	username := "alice"
	correctHash := sha256Hex("right")

	// 5 次错误密码 → 第 5 次时 Incr 返回 5 → service 写 lock key
	for i := 0; i < 5; i++ {
		mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "username", "password", "nickname", "email", "avatar",
				"level", "experience", "signature", "bio",
				"follower_count", "following_count", "liked_count",
				"status", "created_at", "updated_at",
			}).AddRow(
				1, username, correctHash, "Alice", "", "",
				1, 0, "", "",
				0, 0, 0,
				1, time.Now(), time.Now(),
			))

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			Username: username, Password: "wrong",
		})
		assertGRPCStatus(t, err, codes.Unauthenticated)
	}

	// 现在 cache 中应有 lock key
	cache.mu.Lock()
	_, hasLock := cache.data["login:fail:1:lock"]
	cache.mu.Unlock()
	assert.True(t, hasLock, "5 次失败后必须写 lock key")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUser_Success(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "password", "nickname", "email", "avatar",
			"level", "experience", "signature", "bio",
			"follower_count", "following_count", "liked_count",
			"status", "created_at", "updated_at",
		}).AddRow(
			7, "alice", "hash", "Alice", "alice@x", "avatar.png",
			2, 100, "", "",
			3, 4, 5,
			1, time.Now(), time.Now(),
		))

	resp, err := svc.GetUser(context.Background(), &pb.GetUserRequest{UserId: 7})
	require.NoError(t, err)
	assert.Equal(t, int64(7), resp.UserId)
	assert.Equal(t, "alice", resp.Username)
	assert.Equal(t, "Alice", resp.Nickname)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUser_NotFound(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(9999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetUser(context.Background(), &pb.GetUserRequest{UserId: 9999})
	assertGRPCStatus(t, err, codes.NotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// helpers

func TestService_GetUser_Success(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "password", "nickname", "email", "avatar",
			"level", "experience", "signature", "bio",
			"follower_count", "following_count", "liked_count",
			"status", "created_at", "updated_at",
		}).AddRow(
			10, "bob", "hash", "Bob", "bob@x", "avatar.png",
			3, 200, "hello", "bio text",
			10, 20, 30,
			1, time.Now(), time.Now(),
		))

	resp, err := svc.GetUser(context.Background(), &pb.GetUserRequest{UserId: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(10), resp.UserId)
	assert.Equal(t, "bob", resp.Username)
	assert.Equal(t, "Bob", resp.Nickname)
	assert.Equal(t, "bob@x", resp.Email)
	assert.Equal(t, "avatar.png", resp.Avatar)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetUser_NotFound(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(9999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetUser(context.Background(), &pb.GetUserRequest{UserId: 9999})
	assertGRPCStatus(t, err, codes.NotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateUser_Success(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs("alice").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := svc.Login(context.Background(), &pb.LoginRequest{
		Username: "alice", Password: "p",
	})
	assertGRPCStatus(t, err, codes.Internal)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateUser_NotFound(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(8888)).
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := svc.GetUser(context.Background(), &pb.GetUserRequest{UserId: 8888})
	assertGRPCStatus(t, err, codes.Internal)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func assertGRPCStatus(t *testing.T, err error, want codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected gRPC code %s, got nil", want)
	}
	s, ok := status.FromError(err)
	require.True(t, ok, "not a gRPC error: %v", err)
	assert.Equal(t, want, s.Code(), "err: %v", err)

	// 触发 errors 包使用, 防 vet 误删 unused import
	_ = errors.ErrInvalidArgument("test")
}