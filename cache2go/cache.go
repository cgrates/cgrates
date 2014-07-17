//Simple caching library with expiration capabilities
package cache2go

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type expiringCacheEntry interface {
	XCache(key string, expire time.Duration, value expiringCacheEntry)
	timer() *time.Timer
	age() time.Duration
	KeepAlive()
}

// Structure that must be embeded in the objectst that must be cached with expiration.
// If the expiration is not needed this can be ignored
type XEntry struct {
	sync.Mutex
	key            string
	keepAlive      bool
	expireDuration time.Duration
	timestamp      time.Time
	t              *time.Timer
}

type timestampedValue struct {
	timestamp time.Time
	value     interface{}
}

const (
	PREFIX_LEN = 4
)

var (
	xcache   = make(map[string]expiringCacheEntry)
	xMux     sync.RWMutex
	cache    = make(map[string]timestampedValue)
	mux      sync.RWMutex
	cMux     sync.Mutex
	counters = make(map[string]int64)
)

// The main function to cache with expiration
func (xe *XEntry) XCache(key string, expire time.Duration, value expiringCacheEntry) {
	xe.keepAlive = true
	xe.key = key
	xe.expireDuration = expire
	xe.timestamp = time.Now()
	xMux.Lock()
	if _, ok := xcache[key]; !ok {
		// only count if the key is not already there
		count(key)
	}
	xcache[key] = value
	xMux.Unlock()
	go xe.expire()
}

// The internal mechanism for expiartion
func (xe *XEntry) expire() {
	for xe.keepAlive {
		xe.Lock()
		xe.keepAlive = false
		xe.Unlock()
		xe.t = time.NewTimer(xe.expireDuration)
		<-xe.t.C
		if !xe.keepAlive {
			xMux.Lock()
			if _, ok := xcache[xe.key]; ok {
				delete(xcache, xe.key)
				descount(xe.key)
			}
			xMux.Unlock()
		}
	}
}

// Getter for the timer
func (xe *XEntry) timer() *time.Timer {
	return xe.t
}

func (xe *XEntry) age() time.Duration {
	return time.Since(xe.timestamp)

}

// Mark entry to be kept another expirationDuration period
func (xe *XEntry) KeepAlive() {
	xe.Lock()
	defer xe.Unlock()
	xe.keepAlive = true
}

// Get an entry from the expiration cache and mark it for keeping alive
func GetXCached(key string) (ece expiringCacheEntry, err error) {
	xMux.RLock()
	defer xMux.RUnlock()
	if r, ok := xcache[key]; ok {
		r.KeepAlive()
		return r, nil
	}
	return nil, errors.New("not found")
}

// The function to be used to cache a key/value pair when expiration is not needed
func Cache(key string, value interface{}) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := cache[key]; !ok {
		// only count if the key is not already there
		count(key)
	}
	cache[key] = timestampedValue{time.Now(), value}
}

// The function to extract a value for a key that never expire
func GetCached(key string) (v interface{}, err error) {
	mux.RLock()
	defer mux.RUnlock()
	if r, ok := cache[key]; ok {
		return r.value, nil
	}
	return nil, errors.New("not found")
}

func GetKeyAge(key string) (time.Duration, error) {
	mux.RLock()
	if r, ok := cache[key]; ok {
		return time.Since(r.timestamp), nil
	}
	mux.RUnlock()
	xMux.RLock()
	if r, ok := xcache[key]; ok {
		return r.age(), nil
	}
	xMux.RUnlock()
	return 0, errors.New("not found")
}

func RemKey(key string) {
	mux.Lock()
	if _, ok := cache[key]; ok {
		delete(cache, key)
		descount(key)
	}
	mux.Unlock()
	xMux.Lock()
	if r, ok := xcache[key]; ok {
		if r.timer() != nil {
			r.timer().Stop()
		}
	}
	if _, ok := xcache[key]; ok {
		delete(xcache, key)
		descount(key)
	}
	xMux.Unlock()
}

func RemPrefixKey(prefix string) {
	mux.Lock()
	for key, _ := range cache {
		if strings.HasPrefix(key, prefix) {
			delete(cache, key)
			descount(key)
		}
	}
	mux.Unlock()
	xMux.Lock()
	for key, _ := range xcache {
		if strings.HasPrefix(key, prefix) {
			if r, ok := xcache[key]; ok {
				if r.timer() != nil {
					r.timer().Stop()
				}
			}
			delete(xcache, key)
			descount(key)
		}
	}
	xMux.Unlock()
}

func GetAllEntries(prefix string) map[string]interface{} {
	mux.Lock()
	result := make(map[string]interface{})
	for key, timestampedValue := range cache {
		if strings.HasPrefix(key, prefix) {
			result[key] = timestampedValue.value
		}
	}
	mux.Unlock()
	xMux.Lock()
	for key, value := range xcache {
		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	xMux.Unlock()
	return result
}

// Delete all keys from cache
func Flush() {
	mux.Lock()
	cache = make(map[string]timestampedValue)
	xMux.Lock()
	mux.Unlock()
	for _, v := range xcache {
		if v.timer() != nil {
			v.timer().Stop()
		}
	}
	xcache = make(map[string]expiringCacheEntry)
	xMux.Unlock()
	cMux.Lock()
	counters = make(map[string]int64)
	defer cMux.Unlock()
}

func CountEntries(prefix string) (result int64) {
	if _, ok := counters[prefix]; ok {
		return counters[prefix]
	}
	return 0
}

// increments the counter for the specified key prefix
func count(key string) {
	if len(key) < PREFIX_LEN {
		return
	}
	cMux.Lock()
	defer cMux.Unlock()
	prefix := key[:PREFIX_LEN]
	if _, ok := counters[prefix]; ok {
		// increase the value
		counters[prefix] += 1
	} else {
		counters[prefix] = 1
	}
}

// decrements the counter for the specified key prefix
func descount(key string) {
	if len(key) < PREFIX_LEN {
		return
	}
	cMux.Lock()
	defer cMux.Unlock()
	prefix := key[:PREFIX_LEN]
	if value, ok := counters[prefix]; ok && value > 0 {
		counters[prefix] -= 1
	}
}

func GetEntriesKeys(prefix string) (keys []string) {
	mux.RLock()
	defer mux.RUnlock()
	for key, _ := range cache {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return
}

func XGetEntriesKeys(prefix string) (keys []string) {
	xMux.RLock()
	defer xMux.RUnlock()
	for key, _ := range xcache {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return
}
