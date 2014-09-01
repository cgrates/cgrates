//Simple caching library with expiration capabilities
package cache2go

import (
	"errors"
	"strings"
	"sync"
	"time"
)

const (
	PREFIX_LEN = 4
)

type timestampedValue struct {
	timestamp time.Time
	value     interface{}
}

var (
	cache    = make(map[string]timestampedValue)
	mux      sync.RWMutex
	counters = make(map[string]int64)
)

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
	defer mux.RUnlock()
	if r, ok := cache[key]; ok {
		return time.Since(r.timestamp), nil
	}
	return 0, errors.New("not found")
}

func RemKey(key string) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := cache[key]; ok {
		delete(cache, key)
		descount(key)
	}
}

func RemPrefixKey(prefix string) {
	mux.Lock()
	defer mux.Unlock()
	for key, _ := range cache {
		if strings.HasPrefix(key, prefix) {
			delete(cache, key)
			descount(key)
		}
	}
}

func GetAllEntries(prefix string) map[string]interface{} {
	mux.Lock()
	defer mux.Unlock()
	result := make(map[string]interface{})
	for key, timestampedValue := range cache {
		if strings.HasPrefix(key, prefix) {
			result[key] = timestampedValue.value
		}
	}
	return result
}

// Delete all keys from cache
func Flush() {
	mux.Lock()
	defer mux.Unlock()
	cache = make(map[string]timestampedValue)
	counters = make(map[string]int64)
}

func CountEntries(prefix string) (result int64) {
	mux.RLock()
	defer mux.RUnlock()
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
