package abstraction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

type memoryQueue struct {
	mu   sync.Mutex
	subs map[string][]chan Message
}

func newMemoryQueue() *memoryQueue {
	return &memoryQueue{subs: make(map[string][]chan Message)}
}

func (m *memoryQueue) Publish(ctx context.Context, topic string, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if chans, ok := m.subs[topic]; ok {
		for _, ch := range chans {
			select {
			case ch <- msg:
			default:
			}
		}
	}
	return nil
}

func (m *memoryQueue) Subscribe(ctx context.Context, topic, group string) (<-chan Message, error) {
	ch := make(chan Message, 100)
	m.mu.Lock()
	m.subs[topic] = append(m.subs[topic], ch)
	m.mu.Unlock()
	return ch, nil
}

func (m *memoryQueue) Ack(ctx context.Context, topic string, msg Message) error  { return nil }
func (m *memoryQueue) Nack(ctx context.Context, topic string, msg Message) error { return nil }
func (m *memoryQueue) Enqueue(ctx context.Context, queue string, msg Message, delay time.Duration) error {
	return m.Publish(ctx, queue, msg)
}
func (m *memoryQueue) Close() error { return nil }

type memoryCache struct {
	mu       sync.RWMutex
	data     map[string]cacheEntry
	maxItems int
}

type cacheEntry struct {
	value []byte
	exp   time.Time
}

func newMemoryCache(cfg CacheStoreConfig) (*memoryCache, error) {
	max := cfg.MaxItems
	if max == 0 {
		max = 10000
	}
	return &memoryCache{data: make(map[string]cacheEntry), maxItems: max}, nil
}

func (m *memoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	if !e.exp.IsZero() && time.Now().After(e.exp) {
		return nil, errors.New("expired")
	}
	return e.value, nil
}

func (m *memoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = cacheEntry{value: value, exp: time.Now().Add(ttl)}
	return nil
}

func (m *memoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *memoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}

func (m *memoryCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	_, err := m.Get(ctx, key)
	if err == nil {
		return false, nil
	}
	m.Set(ctx, key, []byte("locked"), ttl)
	return true, nil
}

func (m *memoryCache) Unlock(ctx context.Context, key string) error {
	return m.Delete(ctx, key)
}

func (m *memoryCache) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	if e, ok := m.data[key]; ok {
		n = int64(len(e.value))
	}
	n++
	m.data[key] = cacheEntry{value: []byte{byte(n)}, exp: time.Now().Add(ttl)}
	return n, nil
}

func (m *memoryCache) Close() error { return nil }

type memoryCaller struct{}

func newMemoryCaller() *memoryCaller { return &memoryCaller{} }

func (m *memoryCaller) Call(ctx context.Context, target, method string, req, resp any) error {
	return errors.New("memory caller not implemented")
}

func (m *memoryCaller) CallStream(ctx context.Context, target, method string, req any) (<-chan []byte, error) {
	return nil, errors.New("memory caller not implemented")
}

func (m *memoryCaller) Close() error { return nil }

type memoryDocStore struct {
	mu          sync.RWMutex
	collections map[string]map[string]any
	counter     int64
}

func newMemoryDocStore() *memoryDocStore {
	return &memoryDocStore{
		collections: make(map[string]map[string]any),
	}
}

func (m *memoryDocStore) Insert(ctx context.Context, collection string, doc any) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.collections[collection] == nil {
		m.collections[collection] = make(map[string]any)
	}
	m.counter++
	id := fmt.Sprintf("%d", m.counter)
	m.collections[collection][id] = doc
	return id, nil
}

func (m *memoryDocStore) FindByID(ctx context.Context, collection, id string, result any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.collections[collection] == nil {
		return errors.New("not found")
	}
	doc, ok := m.collections[collection][id]
	if !ok {
		return errors.New("not found")
	}
	data, _ := json.Marshal(doc)
	json.Unmarshal(data, result)
	return nil
}

func (m *memoryDocStore) Update(ctx context.Context, collection, id string, doc any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.collections[collection] == nil {
		return errors.New("not found")
	}
	m.collections[collection][id] = doc
	return nil
}

func (m *memoryDocStore) Delete(ctx context.Context, collection, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.collections[collection] != nil {
		delete(m.collections[collection], id)
	}
	return nil
}

func (m *memoryDocStore) Query(ctx context.Context, collection string, filter QueryFilter, result any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.collections[collection] == nil {
		data, _ := json.Marshal([]any{})
		json.Unmarshal(data, result)
		return nil
	}
	var list []any
	for _, doc := range m.collections[collection] {
		if filter.Filters != nil {
			docMap, ok := doc.(map[string]any)
			if ok {
				match := true
				for k, v := range filter.Filters {
					if docMap[k] != v {
						match = false
						break
					}
				}
				if !match {
					continue
				}
			}
		}
		list = append(list, doc)
	}
	data, _ := json.Marshal(list)
	json.Unmarshal(data, result)
	return nil
}

type memorySearch struct {
	mu      sync.RWMutex
	indexes map[string]map[string]any
}

func newMemorySearch() *memorySearch {
	return &memorySearch{indexes: make(map[string]map[string]any)}
}

func (m *memorySearch) Index(ctx context.Context, index string, docID string, doc any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.indexes[index] == nil {
		m.indexes[index] = make(map[string]any)
	}
	m.indexes[index][docID] = doc
	return nil
}

func (m *memorySearch) Delete(ctx context.Context, index string, docID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.indexes[index] != nil {
		delete(m.indexes[index], docID)
	}
	return nil
}

func (m *memorySearch) Search(ctx context.Context, index string, query string, opts SearchOptions) (*SearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := &SearchResult{}
	if m.indexes[index] == nil {
		return result, nil
	}
	for _, doc := range m.indexes[index] {
		result.Hits = append(result.Hits, SearchHit{ID: "", Score: 1.0, Source: doc})
		result.Total++
	}
	return result, nil
}

func (m *memorySearch) BulkIndex(ctx context.Context, index string, docs map[string]any) error {
	for id, doc := range docs {
		m.Index(ctx, index, id, doc)
	}
	return nil
}

type memoryStorage struct {
	mu    sync.RWMutex
	files map[string][]byte
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{files: make(map[string][]byte)}
}

func (m *memoryStorage) Put(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	data, _ := io.ReadAll(body)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[bucket+"/"+key] = data
	return nil
}

func (m *memoryStorage) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.files[bucket+"/"+key]
	if !ok {
		return nil, errors.New("not found")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *memoryStorage) Delete(ctx context.Context, bucket, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, bucket+"/"+key)
	return nil
}

func (m *memoryStorage) Head(ctx context.Context, bucket, key string) (*FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.files[bucket+"/"+key]
	if !ok {
		return nil, errors.New("not found")
	}
	return &FileInfo{Key: key, Size: int64(len(data))}, nil
}

func (m *memoryStorage) List(ctx context.Context, bucket, prefix string) ([]FileInfo, error) {
	return nil, nil
}

func (m *memoryStorage) SignedURL(ctx context.Context, bucket, key string, expire time.Duration) (string, error) {
	return "", errors.New("signed url not available in memory mode")
}

func init() {
	_ = fmt.Sprintf
	_ = json.Marshal
	_ = bytes.NewReader
	_ = io.ReadAll
}
