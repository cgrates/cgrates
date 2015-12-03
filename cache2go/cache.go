//Simple caching library with expiration capabilities
package cache2go

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

type transactionItem struct {
	key   string
	value interface{}
	kind  string
}

func init() {
	if DOUBLE_CACHE {
		cache = newDoubleStore()
	} else {
		cache = newSimpleStore()
	}
}

var (
	mux   sync.RWMutex
	cache cacheStore
	// transaction stuff
	transactionBuffer []*transactionItem
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
			Push(item.key, item.value)
		case KIND_POP:
			Pop(item.key, item.value)
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
		cache.Put(key, value)
		//log.Println("ADD: ", key)
	} else {
		transactionBuffer = append(transactionBuffer, &transactionItem{key: key, value: value, kind: KIND_ADD})
	}
}

// Appends to an existing slice in the cache key
func Push(key string, value interface{}) {
	if !transactionLock {
		mux.Lock()
		defer mux.Unlock()
	}
	if !transactionON {
		cache.Append(key, value)
	} else {
		transactionBuffer = append(transactionBuffer, &transactionItem{key: key, value: value, kind: KIND_ADP})
	}
}

// The function to extract a value for a key that never expire
func Get(key string) (v interface{}, err error) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.Get(key)
}

func Pop(key string, value interface{}) {
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

func RemKey(key string) {
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

func RemPrefixKey(prefix string) {
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
func Flush() {
	mux.Lock()
	defer mux.Unlock()
	if DOUBLE_CACHE {
		cache = newDoubleStore()
	} else {
		cache = newSimpleStore()
	}
}

func CountEntries(prefix string) (result int) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.CountEntriesForPrefix(prefix)
}

func GetAllEntries(prefix string) (map[string]interface{}, error) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.GetAllForPrefix(prefix)
}

func GetEntriesKeys(prefix string) (keys []string) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.GetKeysForPrefix(prefix)
}
