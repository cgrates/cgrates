/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package cache

import (
	"container/list"
	"sync"
	"time"
)

// Cache is an LRU cache.
type Cache struct {
	mu sync.RWMutex
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	maxEntries int

	lruIndex   *list.List
	ttlIndex   []*list.Element
	cache      map[string]*list.Element
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
		cache:      make(map[string]*list.Element),
	}
	if c.expiration > 0 {
		c.ttlIndex = make([]*list.Element, 0)
		go c.cleanExpired()
	}
	return c
}

// cleans expired entries performing minimal checks
func (c *Cache) cleanExpired() {
	for {
		c.mu.RLock()
		if len(c.ttlIndex) == 0 {
			c.mu.RUnlock()
			time.Sleep(c.expiration)
			continue
		}
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
		c.cache = make(map[string]*list.Element)
		c.lruIndex = list.New()
		if c.expiration > 0 {
			c.ttlIndex = make([]*list.Element, 0)
		}
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
	if c.expiration > 0 {
		c.ttlIndex = append(c.ttlIndex, e)
	}
	c.cache[key] = e

	if c.maxEntries != 0 && c.lruIndex.Len() > c.maxEntries {
		c.removeOldest()
	}
	c.mu.Unlock()
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
		//e.Value.(*entry).timestamp = time.Now() don't update the timestamp on get'
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
func (c *Cache) removeOldest() {
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
	if e.Value != nil {
		kv := e.Value.(*entry)
		delete(c.cache, kv.key)
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
	if c.expiration > 0 {
		c.ttlIndex = make([]*list.Element, 0)
	}
	c.cache = make(map[string]*list.Element)
}
