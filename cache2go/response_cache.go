package cache2go

import (
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("NOT_FOUND")

type ResponseCache struct {
	ttl   time.Duration
	cache map[string]interface{}
	mu    sync.RWMutex
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		ttl:   ttl,
		cache: make(map[string]interface{}),
	}
}

func (rc *ResponseCache) Cache(key string, item interface{}) {
	rc.mu.Lock()
	rc.cache[key] = item
	rc.mu.Unlock()
	go func() {
		time.Sleep(rc.ttl)
		rc.mu.Lock()
		delete(rc.cache, key)
		rc.mu.Unlock()
	}()
}

func (rc *ResponseCache) Get(key string) (interface{}, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	item, ok := rc.cache[key]
	if !ok {
		return nil, ErrNotFound
	}
	return item, nil
}
