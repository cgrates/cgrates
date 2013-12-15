package cache2go

import (
	"testing"
	"time"
)

type myStruct struct {
	XEntry
	data string
}

func TestCache(t *testing.T) {
	a := &myStruct{data: "mama are mere"}
	a.XCache("mama", 1*time.Second, a)
	b, err := GetXCached("mama")
	if err != nil || b == nil || b != a {
		t.Error("Error retriving data from cache", err)
	}
}

func TestCacheExpire(t *testing.T) {
	a := &myStruct{data: "mama are mere"}
	a.XCache("mama", 1*time.Second, a)
	b, err := GetXCached("mama")
	if err != nil || b == nil || b.(*myStruct).data != "mama are mere" {
		t.Error("Error retriving data from cache", err)
	}
	time.Sleep(1001 * time.Millisecond)
	b, err = GetXCached("mama")
	if err == nil || b != nil {
		t.Error("Error expiring data")
	}
}

func TestCacheKeepAlive(t *testing.T) {
	a := &myStruct{data: "mama are mere"}
	a.XCache("mama", 1*time.Second, a)
	b, err := GetXCached("mama")
	if err != nil || b == nil || b.(*myStruct).data != "mama are mere" {
		t.Error("Error retriving data from cache", err)
	}
	time.Sleep(500 * time.Millisecond)
	b.KeepAlive()
	time.Sleep(501 * time.Millisecond)
	if err != nil {
		t.Error("Error keeping cached data alive", err)
	}
	time.Sleep(1000 * time.Millisecond)
	b, err = GetXCached("mama")
	if err == nil || b != nil {
		t.Error("Error expiring data")
	}
}

func TestFlush(t *testing.T) {
	a := &myStruct{data: "mama are mere"}
	a.XCache("mama", 10*time.Second, a)
	time.Sleep(1000 * time.Millisecond)
	XFlush()
	b, err := GetXCached("mama")
	if err == nil || b != nil {
		t.Error("Error expiring data")
	}
}

func TestFlushNoTimout(t *testing.T) {
	a := &myStruct{data: "mama are mere"}
	a.XCache("mama", 10*time.Second, a)
	XFlush()
	b, err := GetXCached("mama")
	if err == nil || b != nil {
		t.Error("Error expiring data")
	}
}

func TestRemKey(t *testing.T) {
	Cache("t1", "test")
	if t1, err := GetCached("t1"); err != nil || t1 != "test" {
		t.Error("Error setting cache")
	}
	RemKey("t1")
	if t1, err := GetCached("t1"); err == nil || t1 == "test" {
		t.Error("Error removing cached key")
	}
}

func TestXRemKey(t *testing.T) {
	a := &myStruct{data: "mama are mere"}
	a.XCache("mama", 10*time.Second, a)
	if t1, err := GetXCached("mama"); err != nil || t1 != a {
		t.Error("Error setting xcache")
	}
	XRemKey("mama")
	if t1, err := GetXCached("mama"); err == nil || t1 == a {
		t.Error("Error removing xcached key: ", err, t1)
	}
}

func TestGetKeyAge(t *testing.T) {
	Cache("t1", "test")
	d, err := GetKeyAge("t1")
	if err != nil || d > time.Millisecond || d < time.Nanosecond {
		t.Error("Error getting cache key age: ", d)
	}
}

func TestXGetKeyAge(t *testing.T) {
	a := &myStruct{data: "mama are mere"}
	a.XCache("t1", 10*time.Second, a)
	d, err := GetXKeyAge("t1")
	if err != nil || d > time.Millisecond || d < time.Nanosecond {
		t.Error("Error getting cache key age: ", d)
	}
}
