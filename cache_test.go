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
	&cachedItem{key: "1", value: "one"},
	&cachedItem{key: "2", value: "two"},
	&cachedItem{key: "3", value: "three"},
	&cachedItem{key: "4", value: "four"},
	&cachedItem{key: "5", value: "five"},
}
var lastEvicted string

func TestSetGetRemNoIndexes(t *testing.T) {
	cache := New(UnlimitedCaching, 0, false,
		func(k key, v interface{}) { lastEvicted = k.(string) })
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
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
	if cache.ttlIdx.Len() != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	if val, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "two" {
		t.Errorf("wrong item value: %v", val)
	}
	if ks := cache.Keys(); len(ks) != 5 {
		t.Errorf("wrong keys: %+v", ks)
	}
	cache.Set("2", "twice")
	if val, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "twice" {
		t.Errorf("wrong item value: %v", val)
	}
	if lastEvicted != "" {
		t.Error("lastEvicted var should be empty")
	}
	cache.Remove("2")
	if lastEvicted != "2" { // onEvicted should populate this var
		t.Error("lastEvicted var should be 2")
	}
	if _, has := cache.Get("2"); has {
		t.Error("item still in cache")
	}
	if len(cache.cache) != 4 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.Len() != 4 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
}

func TestSetGetRemLRU(t *testing.T) {
	cache := New(3, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
	}
	if len(cache.cache) != 3 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).key.(string) != "5" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).key.(string) != "3" {
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
	cache.Set("3", "third")
	if val, has := cache.Get("3"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "third" {
		t.Errorf("wrong item value: %v", val)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).key.(string) != "3" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).key.(string) != "4" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	cache.Set("2", "second")
	if val, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "second" {
		t.Errorf("wrong item value: %v", val)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).key.(string) != "2" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).key.(string) != "5" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	// 4 should have been removed
	if _, has := cache.Get("4"); has {
		t.Error("item still in cache")
	}
	cache.Remove("2")
	if _, has := cache.Get("2"); has {
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
	if cache.lruIdx.Front().Value.(*cachedItem).key.(string) != "3" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).key.(string) != "5" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
}

func TestSetGetRemTTLDynamic(t *testing.T) {
	cache := New(UnlimitedCaching, time.Duration(10*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
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
	if _, has := cache.Get("2"); !has {
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
	cache := New(UnlimitedCaching, time.Duration(10*time.Millisecond), true, nil)
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
	}
	if cache.Len() != 5 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if _, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if cache.Len() != 0 {
		t.Errorf("Wrong items in cache: %+v", cache.cache)
	}
}

func TestSetGetRemLRUttl(t *testing.T) {
	nrItems := 3
	cache := New(nrItems, time.Duration(10*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
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
	cache.Remove("4")
	cache.Set("3", "third")
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

// BenchmarkSetSimpleCache 	10000000	       180 ns/op
func BenchmarkSetSimpleCache(b *testing.B) {
	cache := New(UnlimitedCaching, 0, false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.key, ci.value)
	}
}

// BenchmarkGetSimpleCache 	10000000	       120 ns/op
func BenchmarkGetSimpleCache(b *testing.B) {
	cache := New(UnlimitedCaching, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.key)
	}
}

// BenchmarkSetLRU         	 5000000	       303 ns/op
func BenchmarkSetLRU(b *testing.B) {
	cache := New(3, 0, false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.key, ci.value)
	}
}

// BenchmarkGetLRU         	10000000	       140 ns/op
func BenchmarkGetLRU(b *testing.B) {
	cache := New(3, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.key)
	}
}

// BenchmarkSetTTL         	10000000	       225 ns/op
func BenchmarkSetTTL(b *testing.B) {
	cache := New(0, time.Duration(time.Millisecond), false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.key, ci.value)
	}
}

// BenchmarkGetTTL         	10000000	       221 ns/op
func BenchmarkGetTTL(b *testing.B) {
	cache := New(0, time.Duration(5*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.key)
	}
}

// BenchmarkSetLRUttl      	 5000000	       381 ns/op
func BenchmarkSetLRUttl(b *testing.B) {
	cache := New(3, time.Duration(time.Millisecond), false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.key, ci.value)
	}
}

// BenchmarkGetLRUttl      	10000000	       182 ns/op
func BenchmarkGetLRUttl(b *testing.B) {
	cache := New(3, time.Duration(5*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.key, ci.value)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.key)
	}
}
