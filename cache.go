/*
Cache.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM GmbH. All Rights Reserved.

A LRU cache with TTL capabilities.
Original ideas from golang groupcache/lru.go

*/

package ltcache

import (
	"container/list"
	"strings"
	"sync"
	"time"
)

const (
	UnlimitedCaching = -1
	DisabledCaching  = 0
)

type cachedItem struct {
	itemID     string
	value      interface{}
	expiryTime time.Time
	groupIDs   []string // list of group this item belongs to
}

// Cache is an LRU/TTL cache. It is safe for concurrent access.
type Cache struct {
	sync.RWMutex
	// cache holds the items
	cache  map[string]*cachedItem
	groups map[string]map[string]struct{} // map[groupID]map[itemKey]struct{}
	// onEvicted will execute specific function if defined when an item will be removed
	onEvicted func(itmID string, value interface{})
	// maxEntries represents maximum number of entries allowed by LRU cache mechanism
	// -1 for unlimited caching, 0 for disabling caching
	maxEntries int
	// ttl represents the lifetime of an cachedItem
	ttl time.Duration
	// staticTTL prevents expiryTime to be modified on key get/set
	staticTTL bool

	lruIdx  *list.List
	lruRefs map[string]*list.Element // index the list element based on it's key in cache
	ttlIdx  *list.List
	ttlRefs map[string]*list.Element // index the list element based on it' key in cache
}

// New initializes a new cache.
func NewCache(maxEntries int, ttl time.Duration, staticTTL bool,
	onEvicted func(itmID string, value interface{})) (c *Cache) {
	c = &Cache{
		cache:      make(map[string]*cachedItem),
		groups:     make(map[string]map[string]struct{}),
		onEvicted:  onEvicted,
		maxEntries: maxEntries,
		ttl:        ttl,
		staticTTL:  staticTTL,
		lruIdx:     list.New(),
		lruRefs:    make(map[string]*list.Element),
		ttlIdx:     list.New(),
		ttlRefs:    make(map[string]*list.Element),
	}
	if c.ttl > 0 {
		go c.cleanExpired()
	}
	return
}

// Get looks up a key's value from the cache
func (c *Cache) Get(itmID string) (value interface{}, ok bool) {
	c.Lock()
	defer c.Unlock()
	ci, has := c.cache[itmID]
	if !has {
		return
	}
	value, ok = ci.value, true
	if c.maxEntries != UnlimitedCaching { // update lru indexes
		c.lruIdx.MoveToFront(c.lruRefs[itmID])
	}
	if c.ttl > 0 && !c.staticTTL { // update ttl indexes
		ci.expiryTime = time.Now().Add(c.ttl)
		c.ttlIdx.MoveToFront(c.ttlRefs[itmID])
	}
	return
}

func (c *Cache) GetItemExpiryTime(itmID string) (exp time.Time, ok bool) {
	c.RLock()
	defer c.RUnlock()
	var ci *cachedItem
	ci, ok = c.cache[itmID]
	if !ok {
		return
	}
	exp = ci.expiryTime
	return
}

func (c *Cache) HasItem(itmID string) (has bool) {
	c.RLock()
	_, has = c.cache[itmID]
	c.RUnlock()
	return
}

// Set sets/adds a value to the cache.
func (c *Cache) Set(itmID string, value interface{}, grpIDs []string) {
	if c.maxEntries == DisabledCaching {
		return
	}
	c.Lock()
	defer c.Unlock()
	now := time.Now()
	if ci, ok := c.cache[itmID]; ok {
		ci.value = value
		c.remItemFromGroups(itmID, ci.groupIDs)
		ci.groupIDs = grpIDs
		c.addItemToGroups(itmID, grpIDs)
		if c.maxEntries != UnlimitedCaching { // update lru indexes
			c.lruIdx.MoveToFront(c.lruRefs[itmID])
		}
		if c.ttl > 0 && !c.staticTTL { // update ttl indexes
			ci.expiryTime = now.Add(c.ttl)
			c.ttlIdx.MoveToFront(c.ttlRefs[itmID])
		}
		return
	}
	ci := &cachedItem{itemID: itmID, value: value, groupIDs: grpIDs}
	c.cache[itmID] = ci
	c.addItemToGroups(itmID, grpIDs)
	if c.maxEntries != UnlimitedCaching {
		c.lruRefs[itmID] = c.lruIdx.PushFront(ci)
	}
	if c.ttl > 0 {
		ci.expiryTime = now.Add(c.ttl)
		c.ttlRefs[itmID] = c.ttlIdx.PushFront(ci)
	}
	if c.maxEntries != UnlimitedCaching {
		var lElm *list.Element
		if c.lruIdx.Len() > c.maxEntries {
			lElm = c.lruIdx.Back()
		}
		if lElm != nil {
			c.remove(lElm.Value.(*cachedItem).itemID)
		}
	}
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(itmID string) {
	c.Lock()
	c.remove(itmID)
	c.Unlock()
}

// GetItemIDs returns a list of items matching prefix
func (c *Cache) GetItemIDs(prfx string) (itmIDs []string) {
	c.RLock()
	for itmID := range c.cache {
		if strings.HasPrefix(itmID, prfx) {
			itmIDs = append(itmIDs, itmID)
		}
	}
	c.RUnlock()
	return
}

// GroupLength returns the length of a group
func (c *Cache) GroupLength(grpID string) int {
	c.RLock()
	defer c.RUnlock()
	return len(c.groups[grpID])
}

func (c *Cache) getGroupItemIDs(grpID string) (itmIDs []string) {
	for itmID := range c.groups[grpID] {
		itmIDs = append(itmIDs, itmID)
	}
	return
}

func (c *Cache) GetGroupItemIDs(grpID string) (itmIDs []string) {
	c.RLock()
	itmIDs = c.getGroupItemIDs(grpID)
	c.RUnlock()
	return
}

func (c *Cache) GetGroupItems(grpID string) (itms []interface{}) {
	for _, itmID := range c.GetGroupItemIDs(grpID) {
		itm, _ := c.Get(itmID)
		itms = append(itms, itm)
	}
	return
}

func (c *Cache) RemoveGroup(grpID string) {
	c.Lock()
	for itmID := range c.groups[grpID] {
		c.remove(itmID)
	}
	c.Unlock()
}

// remove completely removes an Element from the cache
func (c *Cache) remove(itmID string) {
	ci, has := c.cache[itmID]
	if !has {
		return
	}
	if c.maxEntries != UnlimitedCaching {
		c.lruIdx.Remove(c.lruRefs[itmID])
		delete(c.lruRefs, itmID)
	}
	if c.ttl > 0 {
		c.ttlIdx.Remove(c.ttlRefs[itmID])
		delete(c.ttlRefs, itmID)
	}
	c.remItemFromGroups(ci.itemID, ci.groupIDs)
	delete(c.cache, ci.itemID)
	if c.onEvicted != nil {
		c.onEvicted(ci.itemID, ci.value)
	}
}

// cleanExpired checks items indexed for TTL and expires them when necessary
func (c *Cache) cleanExpired() {
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
		c.remove(ci.itemID)
		c.Unlock()
	}
}

// addItemToGroups adds and item to a group
func (c *Cache) addItemToGroups(itmKey string, groupIDs []string) {
	for _, grpID := range groupIDs {
		if _, has := c.groups[grpID]; !has {
			c.groups[grpID] = make(map[string]struct{})
		}
		c.groups[grpID][itmKey] = struct{}{}
	}
}

// remItemFromGroups removes an item with itemKey from groups
func (c *Cache) remItemFromGroups(itmKey string, groupIDs []string) {
	for _, grpID := range groupIDs {
		delete(c.groups[grpID], itmKey)
		if len(c.groups[grpID]) == 0 {
			delete(c.groups, grpID)
		}
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.cache)
}

// Clear purges all stored items from the cache.
func (c *Cache) Clear() {
	c.Lock()
	defer c.Unlock()
	if c.onEvicted != nil {
		for _, ci := range c.cache {
			c.onEvicted(ci.itemID, ci.value)
		}
	}
	c.cache = make(map[string]*cachedItem)
	c.groups = make(map[string]map[string]struct{})
	c.lruIdx = c.lruIdx.Init()
	c.lruRefs = make(map[string]*list.Element)
	c.ttlIdx = c.ttlIdx.Init()
	c.ttlRefs = make(map[string]*list.Element)
}

type CacheStats struct {
	Items  int
	Groups int
}

// GetStats will return the CacheStats for this instance
func (c *Cache) GetCacheStats() (cs *CacheStats) {
	c.RLock()
	cs = &CacheStats{Items: len(c.cache), Groups: len(c.groups)}
	c.RUnlock()
	return
}
