package abstraction

import (
	"context"
	"errors"
	"sync"
	"time"
)

type memoryQueue struct {
	mu       sync.Mutex
	subs     map[string][]chan Message
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

func (m *memoryQueue) Ack(ctx context.Context, topic string, msg Message) error { return nil }
func (m *memoryQueue) Nack(ctx context.Context, topic string, msg Message) error { return nil }
func (m *memoryQueue) Enqueue(ctx context.Context, queue string, msg Message, delay time.Duration) error { return m.Publish(ctx, queue, msg) }
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