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
	KIND_ADD   = "ADD"
	KIND_ADP   = "ADP"
	KIND_REM   = "REM"
	KIND_PRF   = "PRF"
)

type timestampedValue struct {
	timestamp time.Time
	value     interface{}
}

type transactionItem struct {
	key   string
	value interface{}
	kind  string
}

var (
	cache    = make(map[string]timestampedValue)
	mux      sync.RWMutex
	counters = make(map[string]int64)

	// transaction stuff
	transactionBuffer []transactionItem
	transactionMux    sync.Mutex
	transactionON     = false
	transactionLock   = false
)

func BeginTransaction() {
	transactionMux.Lock()
	transactionLock = true
	transactionON = true
}

func RollbackTransaction() {
	transactionBuffer = nil
	transactionLock = false
	transactionON = false
	transactionMux.Unlock()
}

func CommitTransaction() {
	transactionON = false
	// apply all transactioned items
	mux.Lock()
	for _, item := range transactionBuffer {
		switch item.kind {
		case KIND_REM:
			RemKey(item.key)
		case KIND_PRF:
			RemPrefixKey(item.key)
		case KIND_ADD:
			Cache(item.key, item.value)
		case KIND_ADP:
			CachePush(item.key, item.value)
		}
	}
	mux.Unlock()
	transactionBuffer = nil
	transactionLock = false
	transactionMux.Unlock()
}

// The function to be used to cache a key/value pair when expiration is not needed
func Cache(key string, value interface{}) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		if _, ok := cache[key]; !ok {
			// only count if the key is not already there
			count(key)
		}
		cache[key] = timestampedValue{time.Now(), value}
		//fmt.Println("ADD: ", key)
	} else {
		transactionBuffer = append(transactionBuffer, transactionItem{key: key, value: value, kind: KIND_ADD})
	}
}

// Appends to an existing slice in the cache key
func CachePush(key string, val interface{}) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		var elements []interface{}
		if ti, exists := cache[key]; exists {
			elements = ti.value.([]interface{})
		}
		// check if the val is already present
		found := false
		for _, v := range elements {
			if val == v {
				found = true
				break
			}
		}
		if !found {
			elements = append(elements, val)
		}
		cache[key] = timestampedValue{time.Now(), elements}
	} else {
		transactionBuffer = append(transactionBuffer, transactionItem{key: key, value: val, kind: KIND_ADP})
	}
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
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		if _, ok := cache[key]; ok {
			//fmt.Println("REM: ", key)
			delete(cache, key)
			descount(key)
		}
	} else {
		transactionBuffer = append(transactionBuffer, transactionItem{key: key, kind: KIND_REM})
	}
}

func RemPrefixKey(prefix string) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		for key, _ := range cache {
			if strings.HasPrefix(key, prefix) {
				//fmt.Println("PRF: ", key)
				delete(cache, key)
				descount(key)
			}
		}
	} else {

		transactionBuffer = append(transactionBuffer, transactionItem{key: prefix, kind: KIND_PRF})
	}
}

func GetAllEntries(prefix string) map[string]interface{} {
	mux.RLock()
	defer mux.RUnlock()
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
