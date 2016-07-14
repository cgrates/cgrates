package cache2go

import "time"

type entry interface {
	Key() string
	SetKey(string)
	Value() interface{}
	SetValue(interface{})
	Timestamp() time.Time
	SetTimestamp(time.Time)
}

type entryLRU struct {
	key   string
	value interface{}
}

func (lru *entryLRU) Key() string {
	return lru.key
}
func (lru *entryLRU) SetKey(k string) {
	lru.key = k
}
func (lru *entryLRU) Value() interface{} {
	return lru.value
}
func (lru *entryLRU) SetValue(v interface{}) {
	lru.value = v
}
func (lru *entryLRU) Timestamp() time.Time {
	return time.Time{}
}
func (lru *entryLRU) SetTimestamp(time.Time) {}

type entryTTL struct {
	key       string
	value     interface{}
	timestamp time.Time
}

func (ttl *entryTTL) Key() string {
	return ttl.key
}
func (ttl *entryTTL) SetKey(k string) {
	ttl.key = k
}
func (ttl *entryTTL) Value() interface{} {
	return ttl.value
}
func (ttl *entryTTL) SetValue(v interface{}) {
	ttl.value = v
}
func (ttl *entryTTL) Timestamp() time.Time {
	return ttl.timestamp
}
func (ttl *entryTTL) SetTimestamp(t time.Time) {
	ttl.timestamp = t
}
