package abstraction

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache_SetGetDelete(t *testing.T) {
	c, err := newMemoryCache(CacheStoreConfig{})
	require.NoError(t, err)
	ctx := context.Background()

	err = c.Set(ctx, "k", []byte("v"), time.Minute)
	require.NoError(t, err)

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, "v", string(got))

	ok, err := c.Exists(ctx, "k")
	require.NoError(t, err)
	assert.True(t, ok)

	err = c.Delete(ctx, "k")
	require.NoError(t, err)
	ok, err = c.Exists(ctx, "k")
	require.NoError(t, err)
	assert.False(t, ok)

	_, err = c.Get(ctx, "k")
	assert.Error(t, err)
}

func TestMemoryCache_Expired(t *testing.T) {
	c, err := newMemoryCache(CacheStoreConfig{})
	require.NoError(t, err)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), -time.Second))
	_, err = c.Get(ctx, "k")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestMemoryCache_LockUnlock(t *testing.T) {
	c, err := newMemoryCache(CacheStoreConfig{})
	require.NoError(t, err)
	ctx := context.Background()

	locked, err := c.Lock(ctx, "lock1", time.Minute)
	require.NoError(t, err)
	assert.True(t, locked)

	locked, err = c.Lock(ctx, "lock1", time.Minute)
	require.NoError(t, err)
	assert.False(t, locked)

	require.NoError(t, c.Unlock(ctx, "lock1"))
	locked, err = c.Lock(ctx, "lock1", time.Minute)
	require.NoError(t, err)
	assert.True(t, locked)
}

func TestMemoryCache_Incr(t *testing.T) {
	c, err := newMemoryCache(CacheStoreConfig{})
	require.NoError(t, err)
	ctx := context.Background()

	n, err := c.Incr(ctx, "counter", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	n, err = c.Incr(ctx, "counter", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	assert.NoError(t, c.Close())
}

func TestMemoryQueue_PublishSubscribe(t *testing.T) {
	q := newMemoryQueue()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := q.Subscribe(ctx, "topic", "group")
	require.NoError(t, err)

	require.NoError(t, q.Publish(ctx, "topic", Message{ID: "1", Payload: []byte("hello")}))
	require.NoError(t, q.Enqueue(ctx, "topic", Message{ID: "2", Payload: []byte("world")}, 0))

	select {
	case m := <-ch:
		assert.Equal(t, "hello", string(m.Payload))
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
	select {
	case m := <-ch:
		assert.Equal(t, "world", string(m.Payload))
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for second message")
	}

	assert.NoError(t, q.Ack(ctx, "topic", Message{}))
	assert.NoError(t, q.Nack(ctx, "topic", Message{}))
	assert.NoError(t, q.Close())
}

func TestMemoryQueue_PublishNoSubscriber(t *testing.T) {
	q := newMemoryQueue()
	assert.NoError(t, q.Publish(context.Background(), "orphan", Message{ID: "1"}))
}

func TestMemoryCaller_NotImplemented(t *testing.T) {
	c := newMemoryCaller()
	err := c.Call(context.Background(), "t", "m", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")

	_, err = c.CallStream(context.Background(), "t", "m", nil)
	assert.Error(t, err)
	assert.NoError(t, c.Close())
}

func TestMemoryDocStore_CRUD(t *testing.T) {
	d := newMemoryDocStore()
	ctx := context.Background()

	id, err := d.Insert(ctx, "col", map[string]any{"name": "a"})
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	var out map[string]any
	require.NoError(t, d.FindByID(ctx, "col", id, &out))
	assert.Equal(t, "a", out["name"])

	require.NoError(t, d.Update(ctx, "col", id, map[string]any{"name": "b"}))
	out = nil
	require.NoError(t, d.FindByID(ctx, "col", id, &out))
	assert.Equal(t, "b", out["name"])

	require.NoError(t, d.Delete(ctx, "col", id))
	err = d.FindByID(ctx, "col", id, &out)
	assert.Error(t, err)
}

func TestMemoryDocStore_QueryWithFilter(t *testing.T) {
	d := newMemoryDocStore()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := d.Insert(ctx, "col", map[string]any{"user_id": int64(7), "i": i})
		require.NoError(t, err)
	}
	_, err := d.Insert(ctx, "col", map[string]any{"user_id": int64(99), "i": 9})
	require.NoError(t, err)

	var list []map[string]any
	require.NoError(t, d.Query(ctx, "col", QueryFilter{
		Filters: map[string]any{"user_id": int64(7)},
	}, &list))
	assert.Len(t, list, 3)
	for _, doc := range list {
		assert.Equal(t, float64(7), doc["user_id"])
	}
}

func TestMemoryDocStore_QueryEmptyCollection(t *testing.T) {
	d := newMemoryDocStore()
	var list []map[string]any
	require.NoError(t, d.Query(context.Background(), "ghost", QueryFilter{}, &list))
	assert.Len(t, list, 0)
}

func TestMemorySearch_IndexSearchDelete(t *testing.T) {
	s := newMemorySearch()
	ctx := context.Background()

	require.NoError(t, s.Index(ctx, "videos", "1", map[string]any{"title": "a"}))
	require.NoError(t, s.Index(ctx, "videos", "2", map[string]any{"title": "b"}))

	res, err := s.Search(ctx, "videos", "a", SearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)
	assert.Len(t, res.Hits, 2)

	res, err = s.Search(ctx, "empty", "a", SearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), res.Total)

	require.NoError(t, s.Delete(ctx, "videos", "1"))
	res, err = s.Search(ctx, "videos", "a", SearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), res.Total)
}

func TestMemorySearch_BulkIndex(t *testing.T) {
	s := newMemorySearch()
	ctx := context.Background()
	require.NoError(t, s.BulkIndex(ctx, "videos", map[string]any{"1": map[string]any{"t": "a"}, "2": map[string]any{"t": "b"}}))

	res, err := s.Search(ctx, "videos", "", SearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)
}

func TestMemoryStorage_PutGetHeadListDelete(t *testing.T) {
	st := newMemoryStorage()
	ctx := context.Background()

	require.NoError(t, st.Put(ctx, "bucket", "a/b.txt", strings.NewReader("hello"), "text/plain"))
	require.NoError(t, st.Put(ctx, "bucket", "a/c.txt", strings.NewReader("world"), "text/plain"))
	require.NoError(t, st.Put(ctx, "bucket", "other.txt", strings.NewReader("zzz"), "text/plain"))

	rc, err := st.Get(ctx, "bucket", "a/b.txt")
	require.NoError(t, err)
	data := make([]byte, 5)
	_, _ = rc.Read(data)
	rc.Close()
	assert.Equal(t, "hello", string(data))

	fi, err := st.Head(ctx, "bucket", "a/b.txt")
	require.NoError(t, err)
	assert.Equal(t, int64(5), fi.Size)
	assert.Equal(t, "a/b.txt", fi.Key)

	list, err := st.List(ctx, "bucket", "a/")
	require.NoError(t, err)
	assert.Len(t, list, 2)

	require.NoError(t, st.Delete(ctx, "bucket", "a/b.txt"))
	_, err = st.Get(ctx, "bucket", "a/b.txt")
	assert.Error(t, err)

	_, err = st.Head(ctx, "bucket", "missing")
	assert.Error(t, err)

	_, err = st.SignedURL(ctx, "bucket", "k", time.Minute)
	assert.Error(t, err)
}

func TestMemoryStorage_ListEmpty(t *testing.T) {
	st := newMemoryStorage()
	list, err := st.List(context.Background(), "bucket", "")
	require.NoError(t, err)
	assert.Len(t, list, 0)
}
