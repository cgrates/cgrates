package cache2go

import (
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("NOT_FOUND")

type CacheItem struct {
	Value interface{}
	Err   error
}

type ResponseCache struct {
	ttl   time.Duration
	cache map[string]*CacheItem
	mu    sync.RWMutex
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		ttl:   ttl,
		cache: make(map[string]*CacheItem),
	}
}

func (rc *ResponseCache) Cache(key string, item *CacheItem) {
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

func (rc *ResponseCache) Get(key string) (*CacheItem, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	item, ok := rc.cache[key]
	if !ok {
		return nil, ErrNotFound
	}
	return item, nil
}
