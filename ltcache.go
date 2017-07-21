/*
ltcache.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM GmbH. All Rights Reserved.

A LRU cache with TTL capabilities.
Original ideas from golang groupcache/lru.go

*/

package ltcache

import (
	"container/list"
	"sync"
	"time"
)

// A key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type key interface{}

type cachedItem struct {
	key        key
	value      interface{}
	expiryTime time.Time
}

// LTCache is an LRU/TTL cache. It is safe for concurrent access.
type LTCache struct {
	// simple locking for now, ToDo: try locking per key
	sync.RWMutex
	// cache holds the items
	cache map[key]*cachedItem
	// onEvicted will execute specific function if defined when an item will be removed
	onEvicted func(k key, value interface{})
	// maxEntries represents maximum number of entries allowed by LRU cache mechanism
	maxEntries int
	// ttl represents the lifetime of an cachedItem
	ttl time.Duration
	// staticTTL prevents expiryTime to be modified on key get/set
	staticTTL bool

	lruIdx  *list.List
	lruRefs map[key]*list.Element // index the list element based on it's key in cache
	ttlIdx  *list.List
	ttlRefs map[key]*list.Element // index the list element based on it' key in cache
}

// NewLTCache initializes a new cache.
func NewLTCache(maxEntries int, ttl time.Duration, staticTTL bool,
	onEvicted func(k key, value interface{})) (ltCache *LTCache) {
	ltCache = &LTCache{
		cache:      make(map[key]*cachedItem),
		onEvicted:  onEvicted,
		maxEntries: maxEntries,
		ttl:        ttl,
		staticTTL:  staticTTL,
		lruIdx:     list.New(),
		lruRefs:    make(map[key]*list.Element),
		ttlIdx:     list.New(),
		ttlRefs:    make(map[key]*list.Element),
	}
	if ltCache.ttl != 0 {
		go ltCache.cleanExpired()
	}
	return
}

// Get looks up a key's value from the cache.
func (c *LTCache) Get(k key) (value interface{}, ok bool) {
	c.Lock()
	defer c.Unlock()
	ci, has := c.cache[k]
	if !has {
		return
	}
	value, ok = ci.value, true
	if c.maxEntries != 0 { // update lru indexes
		c.lruIdx.MoveToFront(c.lruRefs[k])
	}
	if c.ttl > 0 && !c.staticTTL { // update ttl indexes
		ci.expiryTime = time.Now().Add(c.ttl)
		c.ttlIdx.MoveToFront(c.ttlRefs[k])
	}
	return
}

// Set sets/adds a value to the cache.
func (c *LTCache) Set(k key, value interface{}) {
	c.Lock()
	defer c.Unlock()
	now := time.Now()
	if ci, ok := c.cache[k]; ok {
		ci.value = value
		if c.maxEntries != 0 { // update lru indexes
			c.lruIdx.MoveToFront(c.lruRefs[k])
		}
		if c.ttl > 0 && !c.staticTTL { // update ttl indexes
			ci.expiryTime = now.Add(c.ttl)
			c.ttlIdx.MoveToFront(c.ttlRefs[k])
		}
		return
	}
	ci := &cachedItem{key: k, value: value}
	c.cache[k] = ci
	if c.maxEntries != 0 {
		c.lruRefs[k] = c.lruIdx.PushFront(ci)
	}
	if c.ttl > 0 {
		ci.expiryTime = now.Add(c.ttl)
		c.ttlRefs[k] = c.ttlIdx.PushFront(ci)
	}
	if c.maxEntries != 0 {
		var lElm *list.Element
		if c.lruIdx.Len() > c.maxEntries {
			lElm = c.lruIdx.Back()
		}
		if lElm != nil {
			c.removeKey(lElm.Value.(*cachedItem).key)
		}
	}
}

// Remove removes the provided key from the cache.
func (c *LTCache) Remove(k key) {
	c.Lock()
	defer c.Unlock()
	c.removeKey(k)
}

// removeElement completely removes an Element from the cache
func (c *LTCache) removeKey(k key) {
	ci, has := c.cache[k]
	if !has {
		return
	}
	if c.maxEntries != 0 {
		c.lruIdx.Remove(c.lruRefs[k])
		delete(c.lruRefs, k)
	}
	if c.ttl != 0 {
		c.ttlIdx.Remove(c.ttlRefs[k])
		delete(c.ttlRefs, k)
	}
	delete(c.cache, ci.key)
	if c.onEvicted != nil {
		c.onEvicted(ci.key, ci.value)
	}
}

// cleanExpired checks items indexed for TTL and expires them when necessary
func (c *LTCache) cleanExpired() {
	for {
		c.Lock()
		if c.ttlIdx.Len() == 0 {
			c.Unlock()
			time.Sleep(c.ttl)
			continue
		}
		ci := c.ttlIdx.Back().Value.(*cachedItem)
		now := time.Now()
		if now.Before(ci.expiryTime) {
			c.Unlock()
			time.Sleep(ci.expiryTime.Sub(now))
			continue
		}
		c.removeKey(ci.key)
		c.Unlock()
	}
}

// Len returns the number of items in the cache.
func (c *LTCache) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.cache)
}

// Clear purges all stored items from the cache.
func (c *LTCache) Clear() {
	c.Lock()
	defer c.Unlock()
	if c.onEvicted != nil {
		for _, ci := range c.cache {
			c.onEvicted(ci.key, ci.value)
		}
	}
	c.cache = make(map[key]*cachedItem)
	c.lruIdx = c.lruIdx.Init()
	c.lruRefs = make(map[key]*list.Element)
	c.ttlIdx = c.ttlIdx.Init()
	c.ttlRefs = make(map[key]*list.Element)
}
