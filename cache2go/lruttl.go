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

	lruIndex   *list.List
	ttlIndex   []*list.Element
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
func NewLRUTTL(maxEntries int, expire time.Duration) *Cache {
	c := &Cache{
		maxEntries: maxEntries,
		expiration: expire,
		lruIndex:   list.New(),
		ttlIndex:   make([]*list.Element, 0),
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
		if len(c.ttlIndex) == 0 {
			time.Sleep(c.expiration)
			continue
		}
		c.mu.RLock()
		e := c.ttlIndex[0]
		c.mu.RUnlock()

		en := e.Value.(*entry)
		if time.Now().After(en.timestamp.Add(c.expiration)) {
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
		c.lruIndex = list.New()
		c.ttlIndex = make([]*list.Element, 0)
	}

	if e, ok := c.cache[key]; ok {
		c.lruIndex.MoveToFront(e)

		en := e.Value.(*entry)
		en.value = value
		en.timestamp = time.Now()

		c.mu.Unlock()
		return
	}
	e := c.lruIndex.PushFront(&entry{key: key, value: value, timestamp: time.Now()})
	c.ttlIndex = append(c.ttlIndex, e)
	c.cache[key] = e
	c.mu.Unlock()

	if c.maxEntries != 0 && c.lruIndex.Len() > c.maxEntries {
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
		c.lruIndex.MoveToFront(e)
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
	e := c.lruIndex.Back()
	if e != nil {
		c.removeElement(e)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.lruIndex.Remove(e)
	if c.expiration > 0 {
		for i, se := range c.ttlIndex {
			if se == e {
				//delete
				copy(c.ttlIndex[i:], c.ttlIndex[i+1:])
				c.ttlIndex[len(c.ttlIndex)-1] = nil
				c.ttlIndex = c.ttlIndex[:len(c.ttlIndex)-1]
				break
			}
		}
	}
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
	return c.lruIndex.Len()
}

// empties the whole cache
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lruIndex = list.New()
	c.ttlIndex = make([]*list.Element, 0)
	c.cache = make(map[interface{}]*list.Element)
}
