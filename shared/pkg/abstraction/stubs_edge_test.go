package abstraction

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 覆盖 stubs.go 的错误/边界分支:
// - memoryDocStore.FindByID / Update 的 collection 不存在分支
// - memoryQueue.Publish 在订阅者缓冲满时的 default 分支
// - memoryStorage.List 前缀不匹配时 continue

func TestMemoryDocStore_FindByID_MissingCollection(t *testing.T) {
	d := newMemoryDocStore()
	var out map[string]any
	err := d.FindByID(context.Background(), "no-such-col", "1", &out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryDocStore_Update_MissingCollection(t *testing.T) {
	d := newMemoryDocStore()
	err := d.Update(context.Background(), "no-such-col", "1", map[string]any{"a": 1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryDocStore_Delete_MissingCollection(t *testing.T) {
	d := newMemoryDocStore()
	require.NoError(t, d.Delete(context.Background(), "no-such-col", "1"))
}

func TestMemoryQueue_Publish_SubscriberBufferFull(t *testing.T) {
	q := newMemoryQueue()
	ctx := context.Background()
	_, err := q.Subscribe(ctx, "hot", "g1")
	require.NoError(t, err)

	// 缓冲容量 100; 多发几条让 select 走 default 分支(非阻塞丢弃)
	for i := 0; i < 110; i++ {
		require.NoError(t, q.Publish(ctx, "hot", Message{ID: fmt.Sprintf("m%d", i)}))
	}
	assert.NoError(t, q.Close())
}

func TestMemoryStorage_List_SkipsOtherBuckets(t *testing.T) {
	st := newMemoryStorage()
	ctx := context.Background()
	require.NoError(t, st.Put(ctx, "bucketA", "a.txt", strings.NewReader("a"), "text/plain"))
	require.NoError(t, st.Put(ctx, "bucketB", "b.txt", strings.NewReader("bb"), "text/plain"))

	// bucketA 下只有 a.txt; bucketB 的文件因前缀不匹配被 continue 跳过
	list, err := st.List(ctx, "bucketA", "")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "a.txt", list[0].Key)
}

func TestMemoryStorage_List_PrefixNoMatch(t *testing.T) {
	st := newMemoryStorage()
	ctx := context.Background()
	require.NoError(t, st.Put(ctx, "bucket", "x.txt", strings.NewReader("data"), "text/plain"))

	// 前缀完全不匹配 → 循环内 continue, 最后返回空切片
	list, err := st.List(ctx, "bucket", "nomatch/")
	require.NoError(t, err)
	assert.Len(t, list, 0)
}