package cache2go

import (
	"errors"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var ErrNotFound = errors.New("NOT_FOUND")

type CacheItem struct {
	Value interface{}
	Err   error
}

type ResponseCache struct {
	ttl       time.Duration
	cache     map[string]*CacheItem
	semaphore map[string]chan bool // used for waiting till the first goroutine processes the response
	mu        sync.RWMutex
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		ttl:       ttl,
		cache:     make(map[string]*CacheItem),
		semaphore: make(map[string]chan bool),
		mu:        sync.RWMutex{},
	}
}

func (rc *ResponseCache) Cache(key string, item *CacheItem) {
	if rc.ttl == 0 {
		return
	}
	rc.mu.Lock()
	rc.cache[key] = item
	if _, found := rc.semaphore[key]; found {
		close(rc.semaphore[key])  // send release signal
		delete(rc.semaphore, key) // delete key
	}
	rc.mu.Unlock()
	go func() {
		time.Sleep(rc.ttl)
		rc.mu.Lock()
		delete(rc.cache, key)
		rc.mu.Unlock()
	}()
}

func (rc *ResponseCache) Get(key string) (*CacheItem, error) {
	if rc.ttl == 0 {
		return nil, utils.ErrNotImplemented
	}
	rc.wait(key) // wait for other goroutine processsing this key
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	item, ok := rc.cache[key]
	if !ok {
		return nil, ErrNotFound
	}
	return item, nil
}

func (rc *ResponseCache) wait(key string) {
	rc.mu.RLock()
	lockChan, found := rc.semaphore[key]
	rc.mu.RUnlock()
	if found {
		<-lockChan
	} else {
		rc.mu.Lock()
		rc.semaphore[key] = make(chan bool)
		rc.mu.Unlock()
	}
}
