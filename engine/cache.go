//Simple caching library with expiration capabilities
package engine

import "sync"

const (
	PREFIX_LEN   = 4
	KIND_ADD     = "ADD"
	KIND_ADP     = "ADP"
	KIND_REM     = "REM"
	KIND_POP     = "POP"
	KIND_PRF     = "PRF"
	DOUBLE_CACHE = true
)

var (
	mux   sync.RWMutex
	cache cacheStore
	// transaction stuff
	transactionBuffer []*transactionItem
	transactionMux    sync.Mutex
	transactionON     = false
	transactionLock   = false
	dumper            *cacheDumper
)

type transactionItem struct {
	key   string
	value interface{}
	kind  string
}

func CacheSetDumperPath(path string) (err error) {
	if dumper == nil {
		dumper, err = newCacheDumper(path)
	}
	return
}

func init() {
	if DOUBLE_CACHE {
		cache = newDoubleStore()
	} else {
		cache = newSimpleStore()
	}
}

func CacheBeginTransaction() {
	transactionMux.Lock()
	transactionLock = true
	transactionON = true
}

func CacheRollbackTransaction() {
	transactionBuffer = nil
	transactionLock = false
	transactionON = false
	transactionMux.Unlock()
}

func CacheCommitTransaction() {
	transactionON = false
	// apply all transactioned items
	mux.Lock()
	for _, item := range transactionBuffer {
		switch item.kind {
		case KIND_REM:
			CacheRemKey(item.key)
		case KIND_PRF:
			CacheRemPrefixKey(item.key)
		case KIND_ADD:
			CacheSet(item.key, item.value)
		case KIND_ADP:
			CachePush(item.key, item.value.(string))
		case KIND_POP:
			CachePop(item.key, item.value.(string))
		}
	}
	mux.Unlock()
	transactionBuffer = nil
	transactionLock = false
	transactionMux.Unlock()
}

func CacheLoad(path string, keys []string) error {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		return cache.Load(path, keys)
	}
	return nil
}

// The function to be used to cache a key/value pair when expiration is not needed
func CacheSet(key string, value interface{}) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		cache.Put(key, value)
		//log.Println("ADD: ", key)
	} else {
		transactionBuffer = append(transactionBuffer, &transactionItem{key: key, value: value, kind: KIND_ADD})
	}
}

// The function to extract a value for a key that never expire
func CacheGet(key string) (v interface{}, err error) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.Get(key)
}

// Appends to an existing slice in the cache key
func CachePush(key string, values ...string) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		cache.Append(key, values...)
	} else {
		transactionBuffer = append(transactionBuffer, &transactionItem{key: key, value: values, kind: KIND_ADP})
	}
}

func CachePop(key string, value string) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		cache.Pop(key, value)
	} else {
		transactionBuffer = append(transactionBuffer, &transactionItem{key: key, value: value, kind: KIND_POP})
	}
}

func CacheRemKey(key string) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		cache.Delete(key)
	} else {
		transactionBuffer = append(transactionBuffer, &transactionItem{key: key, kind: KIND_REM})
	}
}

func CacheRemPrefixKey(prefix string) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		cache.DeletePrefix(prefix)
	} else {
		transactionBuffer = append(transactionBuffer, &transactionItem{key: prefix, kind: KIND_PRF})
	}
}

// Delete all keys from cache
func CacheFlush() {
	mux.Lock()
	defer mux.Unlock()
	if DOUBLE_CACHE {
		cache = newDoubleStore()
	} else {
		cache = newSimpleStore()
	}
}

func CacheCountEntries(prefix string) (result int) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.CountEntriesForPrefix(prefix)
}

func CacheGetAllEntries(prefix string) (map[string]interface{}, error) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.GetAllForPrefix(prefix)
}

func CacheGetEntriesKeys(prefix string) (keys []string) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.GetKeysForPrefix(prefix)
}
