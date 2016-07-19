package cache2go

import (
	"container/list"
	"sync"
	"time"
)

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	mu sync.RWMutex
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	maxEntries int

	// OnEvicted optionally specificies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key string, value interface{})

	ll         *list.List
	cache      map[interface{}]*list.Element
	expiration time.Duration
}

type entry struct {
	key       string
	value     interface{}
	timestamp time.Time
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New(maxEntries int, expire time.Duration) *Cache {
	c := &Cache{
		maxEntries: maxEntries,
		expiration: expire,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
	if c.expiration > 0 {
		go c.cleanExpired()
	}
	return c
}

// cleans expired entries performing minimal checks
func (c *Cache) cleanExpired() {
	for {
		c.mu.RLock()
		e := c.ll.Back()
		c.mu.RUnlock()
		if e == nil {
			time.Sleep(c.expiration)
			continue
		}
		en := e.Value.(*entry)
		if en.timestamp.Add(c.expiration).After(time.Now()) {
			c.mu.Lock()
			c.removeElement(e)
			c.mu.Unlock()
		} else {
			time.Sleep(time.Now().Sub(en.timestamp.Add(c.expiration)))
		}
	}
}

// Add adds a value to the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}

	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e)

		en := e.Value.(*entry)
		en.value = value
		en.timestamp = time.Now()

		c.mu.Unlock()
		return
	}
	e := c.ll.PushFront(&entry{key: key, value: value, timestamp: time.Now()})
	c.cache[key] = e
	c.mu.Unlock()

	if c.maxEntries != 0 && c.ll.Len() > c.maxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cache == nil {
		return
	}
	if e, hit := c.cache[key]; hit {
		c.ll.MoveToFront(e)
		e.Value.(*entry).timestamp = time.Now()
		return e.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}
	if e, hit := c.cache[key]; hit {
		c.removeElement(e)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}
	e := c.ll.Back()
	if e != nil {
		c.removeElement(e)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// empties the whole cache
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ll = list.New()
	c.cache = make(map[interface{}]*list.Element)
}
