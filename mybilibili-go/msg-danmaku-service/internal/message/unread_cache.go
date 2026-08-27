package message

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func int64Str(n int64) string { return strconv.FormatInt(n, 10) }

func parseInt32(s string) int32 {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(n)
}

// UnreadCache 未读红点的 Redis 缓存层。
// DB 是源真相(conversations.unread_count + messages.is_read)，Redis Hash 为纯缓存：
// 读优先查 Redis，miss 则从 DB 重算回填；任何影响未读的写操作 invalidate 对应 key，
// 下次读取自动重算，保证一致性与零漂移。
type UnreadCache struct {
	client *redis.Client
	repo   *MessageRepository
	ttl    time.Duration
}

func NewUnreadCache(client *redis.Client, repo *MessageRepository, ttl time.Duration) *UnreadCache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &UnreadCache{client: client, repo: repo, ttl: ttl}
}

func unreadKey(uid int64) string {
	return "unread:" + int64Str(uid)
}

// Counts 返回某用户的六类未读数，优先走缓存，miss 回源 DB 重算回填。
func (c *UnreadCache) Counts(ctx context.Context, uid int64) (map[string]int32, error) {
	key := unreadKey(uid)
	hash, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	// 命中缓存
	if len(hash) > 0 {
		return hashToCounts(hash), nil
	}
	// 回源重算并回填
	counts := c.repo.GetUnreadCountsByType(ctx, uid)
	fields := make(map[string]any, len(counts))
	for k, v := range counts {
		fields[k] = v
	}
	if len(fields) > 0 {
		_ = c.client.HSet(ctx, key, fields).Err()
		_ = c.client.Expire(ctx, key, c.ttl).Err()
	}
	return counts, nil
}

// Invalidate 使某用户未读缓存失效（写入影响未读时调用）。
func (c *UnreadCache) Invalidate(ctx context.Context, uid int64) {
	_ = c.client.Del(ctx, unreadKey(uid)).Err()
}

func hashToCounts(hash map[string]string) map[string]int32 {
	counts := map[string]int32{
		"private": 0, "reply": 0, "at": 0, "like": 0, "system": 0, "dynamic": 0,
	}
	for k, v := range hash {
		n := parseInt32(v)
		if n > 0 {
			counts[k] = n
		}
	}
	return counts
}
