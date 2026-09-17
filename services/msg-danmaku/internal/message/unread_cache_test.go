package message

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUnreadCache(t *testing.T) (*miniredis.Miniredis, *UnreadCache) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	db, _ := newMessageDB(t)
	repo := NewMessageRepository(db)
	cache := NewUnreadCache(client, repo, time.Minute)
	return mr, cache
}

func unreadCountsRows(vals ...int32) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"private", "reply", "at", "like", "system"}).AddRow(vals[0], vals[1], vals[2], vals[3], vals[4])
}

func TestInt64Str(t *testing.T) {
	assert.Equal(t, "123", int64Str(123))
	assert.Equal(t, "0", int64Str(0))
	assert.Equal(t, "-42", int64Str(-42))
}

func TestParseInt32(t *testing.T) {
	assert.Equal(t, int32(5), parseInt32("5"))
	assert.Equal(t, int32(-1), parseInt32("-1"))
	assert.Equal(t, int32(0), parseInt32(""))
	assert.Equal(t, int32(0), parseInt32("notanumber"))
	assert.Equal(t, int32(0), parseInt32("99999999999999"))
}

func TestUnreadKey(t *testing.T) {
	assert.Equal(t, "unread:1001", unreadKey(1001))
	assert.Equal(t, "unread:0", unreadKey(0))
}

func TestHashToCounts(t *testing.T) {
	hash := map[string]string{
		"private": "2",
		"reply":   "3",
		"like":    "-1",
		"unknown": "9",
	}
	counts := hashToCounts(hash)
	assert.Equal(t, int32(2), counts["private"])
	assert.Equal(t, int32(3), counts["reply"])
	assert.Equal(t, int32(0), counts["like"], "negative value should not overwrite zero")
	assert.Equal(t, int32(0), counts["system"])
	assert.Equal(t, int32(0), counts["dynamic"])
}

func TestNewUnreadCache_DefaultTTL(t *testing.T) {
	cache := NewUnreadCache(nil, nil, 0)
	assert.Equal(t, 5*time.Minute, cache.ttl)
	cache2 := NewUnreadCache(nil, nil, 10*time.Minute)
	assert.Equal(t, 10*time.Minute, cache2.ttl)
}

func TestUnreadCache_Counts_Miss_Backfill(t *testing.T) {
	mr, cache := newTestUnreadCache(t)
	ctx := context.Background()

	realDB, mock := newMessageDB(t)
	cache.repo = NewMessageRepository(realDB)
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(1001)).
		WillReturnRows(unreadCountsRows(1, 2, 3, 4, 5))

	counts, err := cache.Counts(ctx, 1001)
	require.NoError(t, err)
	assert.Equal(t, int32(1), counts["private"])
	assert.Equal(t, int32(2), counts["reply"])
	assert.Equal(t, int32(14), counts["dynamic"])

	// 已回填到 redis
	assert.True(t, mr.Exists("unread:1001"))
	assert.Equal(t, "1", mr.HGet("unread:1001", "private"))
	assert.Equal(t, "5", mr.HGet("unread:1001", "system"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnreadCache_Counts_Hit(t *testing.T) {
	mr, cache := newTestUnreadCache(t)
	ctx := context.Background()
	mr.HSet("unread:1001", "private", "7", "reply", "8")

	counts, err := cache.Counts(ctx, 1001)
	require.NoError(t, err)
	assert.Equal(t, int32(7), counts["private"])
	assert.Equal(t, int32(8), counts["reply"])
	assert.Equal(t, int32(0), counts["at"], "absent keys stay zero")
	// 命中缓存不应触发回源写缓存
	assert.Equal(t, "7", mr.HGet("unread:1001", "private"))
}

func TestUnreadCache_Counts_RedisError(t *testing.T) {
	db, _ := newMessageDB(t)
	defer db.Close()
	repo := NewMessageRepository(db)

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: 0})
	cache := NewUnreadCache(client, repo, time.Minute)
	mr.Close()

	_, err := cache.Counts(context.Background(), 1)
	assert.Error(t, err)
}

func TestUnreadCache_Invalidate(t *testing.T) {
	mr, cache := newTestUnreadCache(t)
	mr.HSet("unread:1001", "private", "1")
	assert.True(t, mr.Exists("unread:1001"))

	cache.Invalidate(context.Background(), 1001)
	assert.False(t, mr.Exists("unread:1001"))
}

func TestUnreadCache_Counts_Backfill_Zero(t *testing.T) {
	db, mock := newMessageDB(t)
	defer db.Close()
	repo := NewMessageRepository(db)
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(9)).
		WillReturnRows(unreadCountsRows(0, 0, 0, 0, 0))

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cache := NewUnreadCache(client, repo, time.Minute)

	counts, err := cache.Counts(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, int32(0), counts["private"])
	assert.Equal(t, int32(0), counts["dynamic"])
	// 全零也回填，避免每次 miss
	assert.True(t, mr.Exists("unread:9"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnreadCache_Counts_BackfillError(t *testing.T) {
	db, mock := newMessageDB(t)
	defer db.Close()
	repo := NewMessageRepository(db)
	// HGetAll miss，回源 SQL 出错，Counts 返回 DB 零值但不报错（忽略 err）
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(10)).
		WillReturnError(errDB)

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cache := NewUnreadCache(client, repo, time.Minute)

	counts, err := cache.Counts(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, int32(0), counts["private"])
	assert.Equal(t, int32(0), counts["dynamic"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnreadCache_Invalidate_Nil(t *testing.T) {
	// Invalidate 对不存在 key 的 Del 不应 panic/报错
	mr, cache := newTestUnreadCache(t)
	cache.Invalidate(context.Background(), 404)
	assert.False(t, mr.Exists("unread:404"))
}
