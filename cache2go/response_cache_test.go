package cache2go

import (
	"testing"
	"time"
)

func TestRCacheSetGet(t *testing.T) {
	rc := NewResponseCache(5 * time.Second)
	rc.Cache("test", &CacheItem{Value: "best"})
	v, err := rc.Get("test")
	if err != nil || v.Value.(string) != "best" {
		t.Error("Error retriving response cache: ", v, err)
	}
}

func TestRCacheExpire(t *testing.T) {
	rc := NewResponseCache(1 * time.Microsecond)
	rc.Cache("test", &CacheItem{Value: "best"})
	time.Sleep(2 * time.Millisecond)
	_, err := rc.Get("test")
	if err == nil {
		t.Error("Error expiring response cache: ", err)
	}
}
