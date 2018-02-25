/*
ltcache.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM GmbH. All Rights Reserved.

A LRU cache with TTL capabilities.

*/

package ltcache

import (
	"math/rand"
	"testing"
	"time"
)

var testCIs = []*cachedItem{
	&cachedItem{itemID: "_1_", value: "one"},
	&cachedItem{itemID: "_2_", value: "two", groupIDs: []string{"grp1"}},
	&cachedItem{itemID: "_3_", value: "three", groupIDs: []string{"grp1", "grp2"}},
	&cachedItem{itemID: "_4_", value: "four", groupIDs: []string{"grp1", "grp2", "grp3"}},
	&cachedItem{itemID: "_5_", value: "five", groupIDs: []string{"grp4"}},
}
var lastEvicted string

func TestSetGetRemNoIndexes(t *testing.T) {
	cache := NewCache(UnlimitedCaching, 0, false,
		func(itmID string, v interface{}) { lastEvicted = itmID })
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, ci.groupIDs)
	}
	if len(cache.cache) != 5 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if len(cache.groups) != 4 {
		t.Errorf("Wrong intems in groups: %+v", cache.groups)
	} else if len(cache.groups["grp1"]) != 3 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp1"])
	} else if len(cache.groups["grp2"]) != 2 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp2"])
	} else if len(cache.groups["grp3"]) != 1 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp3"])
	} else if len(cache.groups["grp4"]) != 1 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp4"])
	}
	if cache.lruIdx.Len() != 0 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 0 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	if itmIDs := cache.GetItemIDs(""); len(itmIDs) != 5 {
		t.Errorf("received: %+v", itmIDs)
	}
	if itmIDs := cache.GetItemIDs("_"); len(itmIDs) != 5 {
		t.Errorf("received: %+v", itmIDs)
	}
	if itmIDs := cache.GetItemIDs("_1"); len(itmIDs) != 1 {
		t.Errorf("received: %+v", itmIDs)
	}
	if itmIDs := cache.GetItemIDs("_1_"); len(itmIDs) != 1 {
		t.Errorf("received: %+v", itmIDs)
	}
	if itmIDs := cache.GetItemIDs("_1__"); len(itmIDs) != 0 {
		t.Errorf("received: %+v", itmIDs)
	}
	if val, has := cache.Get("_2_"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "two" {
		t.Errorf("wrong item value: %v", val)
	}
	if len(cache.cache) != 5 {
		t.Errorf("wrong keys: %+v", cache.cache)
	}
	cache.Set("_2_", "twice", []string{"grp21"})
	if val, has := cache.Get("_2_"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "twice" {
		t.Errorf("wrong item value: %v", val)
	}
	if len(cache.groups) != 5 {
		t.Errorf("Wrong intems in groups: %+v", cache.groups)
	} else if len(cache.groups["grp1"]) != 2 { // one gone through set
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp1"])
	} else if len(cache.groups["grp21"]) != 1 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp21"])
	}
	if lastEvicted != "" {
		t.Error("lastEvicted var should be empty")
	}
	cache.Remove("_2_")
	if len(cache.groups) != 4 {
		t.Errorf("Wrong intems in groups: %+v", cache.groups)
	} else if len(cache.groups["grp1"]) != 2 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp1"])
	} else if len(cache.groups["grp21"]) != 0 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp21"])
	}
	if lastEvicted != "_2_" { // onEvicted should populate this var
		t.Error("lastEvicted var should be 2")
	}
	if _, has := cache.Get("_2_"); has {
		t.Error("item still in cache")
	}
	if len(cache.cache) != 4 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.Len() != 4 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	cache.Remove("_3_")
	if len(cache.groups) != 4 {
		t.Errorf("Wrong intems in groups: %+v", cache.groups)
	} else if len(cache.groups["grp1"]) != 1 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp1"])
	} else if len(cache.groups["grp2"]) != 1 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp2"])
	} else if len(cache.groups["grp3"]) != 1 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp3"])
	} else if len(cache.groups["grp4"]) != 1 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp4"])
	}
	cache.RemoveGroup("nonexistent")
	cache.RemoveGroup("grp1")
	if len(cache.cache) != 2 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if len(cache.groups) != 1 {
		t.Errorf("Wrong intems in groups: %+v", cache.groups)
	} else if len(cache.groups["grp4"]) != 1 {
		t.Errorf("Wrong intems in group: %+v", cache.groups["grp4"])
	}
	cache.RemoveGroup("grp1")
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}

}

func TestSetGetRemLRU(t *testing.T) {
	cache := NewCache(3, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if len(cache.cache) != 3 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).itemID != "_5_" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).itemID != "_3_" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 3 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	if _, has := cache.Get("2"); has {
		t.Error("item still in cache")
	}
	// rewrite and reposition 3
	cache.Set("_3_", "third", nil)
	if val, has := cache.Get("_3_"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "third" {
		t.Errorf("wrong item value: %v", val)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).itemID != "_3_" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).itemID != "_4_" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	cache.Set("_2_", "second", nil)
	if val, has := cache.Get("_2_"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "second" {
		t.Errorf("wrong item value: %v", val)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).itemID != "_2_" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).itemID != "_5_" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	// 4 should have been removed
	if _, has := cache.Get("_4_"); has {
		t.Error("item still in cache")
	}
	cache.Remove("_2_")
	if _, has := cache.Get("_2_"); has {
		t.Error("item still in cache")
	}
	if len(cache.cache) != 2 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 2 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 2 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).itemID != "_3_" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).itemID != "_5_" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
}

func TestSetGetRemTTLDynamic(t *testing.T) {
	cache := NewCache(UnlimitedCaching, time.Duration(10*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if len(cache.cache) != 5 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 0 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 0 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != 5 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 5 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if _, has := cache.Get("_2_"); !has {
		t.Error("item not in cache")
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if cache.Len() != 1 {
		t.Errorf("Wrong items in cache: %+v", cache.cache)
	}
	if cache.ttlIdx.Len() != 1 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 1 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
}

func TestSetGetRemTTLStatic(t *testing.T) {
	cache := NewCache(UnlimitedCaching, time.Duration(10*time.Millisecond), true, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if cache.Len() != 5 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if _, has := cache.Get("_2_"); !has {
		t.Error("item not in cache")
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if cache.Len() != 0 {
		t.Errorf("Wrong items in cache: %+v", cache.cache)
	}
}

func TestSetGetRemLRUttl(t *testing.T) {
	nrItems := 3
	cache := NewCache(nrItems, time.Duration(10*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if cache.Len() != nrItems {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != nrItems {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != nrItems {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	cache.Remove("_4_")
	cache.Set("_3_", "third", nil)
	nrItems = 2
	if cache.Len() != nrItems {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != nrItems {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != nrItems {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	time.Sleep(time.Duration(6 * time.Millisecond)) // timeout items which were not modified
	nrItems = 1
	if cache.Len() != nrItems {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != nrItems {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != nrItems {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
}

func TestCacheDisabled(t *testing.T) {
	cache := NewCache(DisabledCaching, time.Duration(10*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
		if _, has := cache.Get(ci.itemID); has {
			t.Errorf("Wrong intems in cache: %+v", cache.cache)
		}
	}
	if cache.Len() != 0 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 0 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 0 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	cache.Remove("4")
}

// BenchmarkSetSimpleCache 	10000000	       228 ns/op
func BenchmarkSetSimpleCache(b *testing.B) {
	cache := NewCache(UnlimitedCaching, 0, false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.itemID, ci.value, nil)
	}
}

// BenchmarkGetSimpleCache 	20000000	        99.7 ns/op
func BenchmarkGetSimpleCache(b *testing.B) {
	cache := NewCache(UnlimitedCaching, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.itemID)
	}
}

// BenchmarkSetLRU         	 5000000	       316 ns/op
func BenchmarkSetLRU(b *testing.B) {
	cache := NewCache(3, 0, false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.itemID, ci.value, nil)
	}
}

// BenchmarkGetLRU         	20000000	       114 ns/op
func BenchmarkGetLRU(b *testing.B) {
	cache := NewCache(3, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.itemID)
	}
}

// BenchmarkSetTTL         	50000000	        30.4 ns/op
func BenchmarkSetTTL(b *testing.B) {
	cache := NewCache(0, time.Duration(time.Millisecond), false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.itemID, ci.value, nil)
	}
}

// BenchmarkGetTTL         	20000000	        88.4 ns/op
func BenchmarkGetTTL(b *testing.B) {
	cache := NewCache(0, time.Duration(5*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.itemID)
	}
}

// BenchmarkSetLRUttl      	 5000000	       373 ns/op
func BenchmarkSetLRUttl(b *testing.B) {
	cache := NewCache(3, time.Duration(time.Millisecond), false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.itemID, ci.value, nil)
	}
}

// BenchmarkGetLRUttl      	10000000	       187 ns/op
func BenchmarkGetLRUttl(b *testing.B) {
	cache := NewCache(3, time.Duration(5*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.itemID)
	}
}
