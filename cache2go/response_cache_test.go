package cache2go

import (
	"testing"
	"time"
)

func TestRCacheSetGet(t *testing.T) {
	rc := NewResponseCache(5 * time.Second)
	rc.Cache("test", "best")
	v, err := rc.Get("test")
	if err != nil || v.(string) != "best" {
		t.Error("Error retriving response cache: ", v, err)
	}
}

func TestRCacheExpire(t *testing.T) {
	rc := NewResponseCache(1 * time.Microsecond)
	rc.Cache("test", "best")
	time.Sleep(1 * time.Millisecond)
	_, err := rc.Get("test")
	if err == nil {
		t.Error("Error expiring response cache: ", err)
	}
}
