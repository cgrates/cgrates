package cache2go

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type myStruct struct {
	data string
}

func TestCache(t *testing.T) {
	cache := NewTTL(time.Second)
	a := &myStruct{data: "mama are mere"}
	cache.Set("mama", a)
	b, ok := cache.Get("mama")
	if !ok || b == nil || b != a {
		t.Error("Error retriving data from cache", b)
	}
}

func TestCacheExpire(t *testing.T) {
	cache := NewTTL(5 * time.Millisecond)
	a := &myStruct{data: "mama are mere"}
	cache.Set("mama", a)
	b, ok := cache.Get("mama")
	if !ok || b == nil || b != a {
		t.Error("Error retriving data from cache", b)
	}
	time.Sleep(5 * time.Millisecond)
	b, ok = cache.Get("mama")
	if ok || b != nil {
		t.Error("Error expiring data from cache", b)
	}
}

func TestLRU(t *testing.T) {
	cache := NewLRU(32)
	for i := 0; i < 40; i++ {
		cache.Set(fmt.Sprintf("%d", i), i)
	}
	if cache.Len() != 32 {
		t.Error("error dicarding least recently used entry: ", cache.Len())
	}
	last := cache.ll.Back().Value.(entry).Value().(int)
	if last != 8 {
		t.Error("error dicarding least recently used entry: ", last)
	}
}

func TestLRUParallel(t *testing.T) {
	cache := NewLRU(32)
	wg := sync.WaitGroup{}
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			cache.Set(fmt.Sprintf("%d", x), x)
		}(i)
	}
	wg.Wait()
	if cache.Len() != 32 {
		t.Error("error dicarding least recently used entry: ", cache.Len())
	}
}

func TestFlush(t *testing.T) {
	cache := NewTTL(5 * time.Millisecond)
	a := &myStruct{data: "mama are mere"}
	cache.Set("mama", a)
	time.Sleep(5 * time.Millisecond)
	cache.Flush()
	b, ok := cache.Get("mama")
	if ok || b != nil {
		t.Error("Error expiring data")
	}
}

func TestFlushNoTimeout(t *testing.T) {
	cache := NewTTL(5 * time.Millisecond)
	a := &myStruct{data: "mama are mere"}
	cache.Set("mama", a)
	cache.Flush()
	b, ok := cache.Get("mama")
	if ok || b != nil {
		t.Error("Error expiring data")
	}
}

func TestRemKey(t *testing.T) {
	cache := NewLRU(10)
	cache.Set("t11_mm", "test")
	if t1, ok := cache.Get("t11_mm"); !ok || t1 != "test" {
		t.Error("Error setting cache")
	}
	cache.Remove("t11_mm")
	if t1, ok := cache.Get("t11_mm"); ok || t1 == "test" {
		t.Error("Error removing cached key")
	}
}

func TestCount(t *testing.T) {
	cache := NewTTL(10 * time.Millisecond)
	cache.Set("dst_A1", "1")
	cache.Set("dst_A2", "2")
	cache.Set("rpf_A3", "3")
	cache.Set("dst_A4", "4")
	cache.Set("dst_A5", "5")
	if cache.Len() != 5 {
		t.Error("Error countiong entries: ", cache.Len())
	}
}
