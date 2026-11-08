package abstraction

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 覆盖 redis_cache.go 的 Incr 错误分支: 对非整数值的 key 执行 INCR 会报错.

func TestRedisCache_Incr_NonIntegerValue(t *testing.T) {
	rc, mr := newTestRedisCache(t)
	ctx := context.Background()

	require.NoError(t, mr.Set("strkey", "abc"))
	_, err := rc.Incr(ctx, "strkey", time.Minute)
	require.Error(t, err)
}