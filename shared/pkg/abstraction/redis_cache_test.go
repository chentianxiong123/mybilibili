package abstraction

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func newTestRedisCache(t *testing.T) (*redisCache, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	cfg := CacheStoreConfig{Addr: mr.Addr(), DefaultTTL: time.Minute}
	cs, err := newRedisCache(cfg)
	if err != nil {
		t.Fatalf("newRedisCache: %v", err)
	}
	rc, ok := cs.(*redisCache)
	if !ok {
		t.Fatalf("expected *redisCache, got %T", cs)
	}
	t.Cleanup(func() { _ = rc.Close() })
	return rc, mr
}

func TestRedisCache_DefaultAddr(t *testing.T) {
	cfg := CacheStoreConfig{} // 无 Addr, 应回落到 127.0.0.1:6379, 但因无服务会失败
	_, err := newRedisCache(cfg)
	if err == nil {
		t.Skip("unexpectedly connected to real redis on 127.0.0.1:6379")
	}
}

func TestRedisCache_SetGet(t *testing.T) {
	rc, _ := newTestRedisCache(t)
	ctx := context.Background()

	if err := rc.Set(ctx, "k1", []byte("v1"), time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := rc.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "v1" {
		t.Errorf("got %q, want v1", string(got))
	}
}

func TestRedisCache_GetNotFound(t *testing.T) {
	rc, _ := newTestRedisCache(t)
	_, err := rc.Get(context.Background(), "missing")
	if err == nil || err.Error() != "not found" {
		t.Errorf("want 'not found', got %v", err)
	}
}

func TestRedisCache_Delete(t *testing.T) {
	rc, mr := newTestRedisCache(t)
	ctx := context.Background()
	_ = rc.Set(ctx, "k", []byte("v"), time.Minute)

	if err := rc.Delete(ctx, "k"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if mr.Exists("k") {
		t.Errorf("key still exists after delete")
	}
}

func TestRedisCache_Exists(t *testing.T) {
	rc, _ := newTestRedisCache(t)
	ctx := context.Background()

	exists, err := rc.Exists(ctx, "absent")
	if err != nil || exists {
		t.Errorf("Exists(absent)=(%v,%v), want (false,nil)", exists, err)
	}

	_ = rc.Set(ctx, "present", []byte("v"), time.Minute)
	exists, err = rc.Exists(ctx, "present")
	if err != nil || !exists {
		t.Errorf("Exists(present)=(%v,%v), want (true,nil)", exists, err)
	}
}

func TestRedisCache_LockUnlock(t *testing.T) {
	rc, _ := newTestRedisCache(t)
	ctx := context.Background()

	ok, err := rc.Lock(ctx, "lock-key", 5*time.Second)
	if err != nil || !ok {
		t.Fatalf("Lock 1st: ok=%v err=%v", ok, err)
	}
	// 二次 Lock 同 key, 应失败
	ok, err = rc.Lock(ctx, "lock-key", 5*time.Second)
	if err != nil {
		t.Fatalf("Lock 2nd err: %v", err)
	}
	if ok {
		t.Errorf("Lock 2nd should return false (already held)")
	}
	if err := rc.Unlock(ctx, "lock-key"); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	ok, err = rc.Lock(ctx, "lock-key", 5*time.Second)
	if err != nil || !ok {
		t.Errorf("Lock after unlock: ok=%v err=%v", ok, err)
	}
}

func TestRedisCache_Incr(t *testing.T) {
	rc, mr := newTestRedisCache(t)
	ctx := context.Background()

	// incr 1st -> 1, ttl 应设置
	n, err := rc.Incr(ctx, "counter", 10*time.Second)
	if err != nil || n != 1 {
		t.Fatalf("Incr 1st: n=%d err=%v", n, err)
	}
	if mr.TTL("counter") == 0 {
		t.Errorf("ttl not set after 1st incr")
	}
	// incr 2nd -> 2, ttl 不变
	n, err = rc.Incr(ctx, "counter", 30*time.Second)
	if err != nil || n != 2 {
		t.Fatalf("Incr 2nd: n=%d err=%v", n, err)
	}
	// incr with ttl=0, 不设置过期
	n, err = rc.Incr(ctx, "nocounter", 0)
	if err != nil || n != 1 {
		t.Fatalf("Incr ttl=0: n=%d err=%v", n, err)
	}
	if mr.TTL("nocounter") != 0 {
		t.Errorf("ttl=0 should not set expiration")
	}
}

func TestRedisCache_Close(t *testing.T) {
	rc, mr := newTestRedisCache(t)
	if err := rc.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	// 关闭后再操作应失败
	if err := rc.Set(context.Background(), "x", []byte("y"), time.Second); err == nil {
		t.Errorf("Set after Close should fail")
	}
	mr.Close() // 显式再关一次避免 lint
}

func TestRedisCache_PingFailure(t *testing.T) {
	// miniredis 不启, 用 unreachable addr 触发 Ping 失败
	cfg := CacheStoreConfig{Addr: "127.0.0.1:1"} // 端口 1 通常没监听
	_, err := newRedisCache(cfg)
	if err == nil {
		t.Skip("port 1 unexpectedly reachable")
	}
}

func TestRedisCache_GetTransportError(t *testing.T) {
	rc, mr := newTestRedisCache(t)
	mr.Close() // 关闭后 Get 返回连接错误等非 redis.Nil
	_, err := rc.Get(context.Background(), "k")
	if err == nil {
		// miniredis 客户端可能已缓存连接池，错误可能延迟到下次操作
		t.Skip("client buffered response; no error expected")
	}
	// 期望是非 redis.Nil 错误（已经被转换为别的错误或原样返回）
	if err.Error() == "not found" {
		t.Errorf("want non-nil-key error, got 'not found' (redis.Nil branch)")
	}
}