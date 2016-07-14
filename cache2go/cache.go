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
	MaxEntries int

	// OnEvicted optionally specificies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key string, value interface{})

	ll         *list.List
	cache      map[interface{}]*list.Element
	expiration time.Duration
	isTTL      bool
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewLRU(maxEntries int) *Cache {
	c := &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
	return c
}

func NewTTL(expire time.Duration) *Cache {
	c := &Cache{
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
		expiration: expire,
		isTTL:      true,
	}
	go c.cleanExpired()
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
		en := e.Value.(*entryTTL)
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

		en := e.Value.(entry)
		en.SetValue(value)
		en.SetTimestamp(time.Now())

		c.mu.Unlock()
		return
	}
	var e *list.Element
	if c.isTTL {
		e = c.ll.PushFront(&entryTTL{key: key, value: value, timestamp: time.Now()})
	} else {
		e = c.ll.PushFront(&entryLRU{key: key, value: value})
	}
	c.cache[key] = e
	c.mu.Unlock()

	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
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
		e.Value.(entry).SetTimestamp(time.Now())
		return e.Value.(entry).Value(), true
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
	kv := e.Value.(entry)
	delete(c.cache, kv.Key())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.Key(), kv.Value())
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
