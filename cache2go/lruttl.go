package cache2go

import (
	"container/list"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	ll    *list.List
	cache map[string]*list.Element
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators

type entry struct {
	key   string
	value interface{}
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewLRUTTL(maxEntries int, exp time.Duration) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[string]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *Cache) Set(key string, value interface{}) {
	if c.cache == nil {
		c.cache = make(map[string]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		if ele.Value == nil {
			utils.Logger.Debug(fmt.Sprintf("<lruttl_cache>: nil val element %+v", ele))
		}
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key string) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	utils.Logger.Debug("<lruttl_cache>: REMOVE OLDEST!")
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// empties the whole cache
func (c *Cache) Flush() {
	utils.Logger.Debug("<lruttl_cache>: FLUSH!")
	c.ll = list.New()
	c.cache = make(map[string]*list.Element)
}
