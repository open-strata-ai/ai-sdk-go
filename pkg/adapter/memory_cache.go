package adapter

import (
	"context"
	"sync"
	"time"
)

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// InMemoryCache is the default Cache adapter: an in-process TTL map (offline/test-friendly).
type InMemoryCache struct {
	mu    sync.Mutex
	store map[string]cacheEntry
}

// NewInMemoryCache builds an InMemoryCache.
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{store: map[string]cacheEntry{}}
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.store[key]
	if !ok {
		return nil, false, nil
	}
	if time.Now().After(e.expiresAt) {
		delete(c.store, key)
		return nil, false, nil
	}
	cp := make([]byte, len(e.value))
	copy(cp, e.value)
	return cp, true, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]byte, len(val))
	copy(cp, val)
	c.store[key] = cacheEntry{value: cp, expiresAt: time.Now().Add(ttl)}
	return nil
}
