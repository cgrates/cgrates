//Simple caching library with expiration capabilities
package cache2go

import (
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type cacheStore interface {
	Put(string, interface{})
	Get(string) (interface{}, bool)
	Delete(string)
	DeletePrefix(string)
	CountEntriesForPrefix(string) int
	GetAllForPrefix(string) map[string]interface{}
	GetKeysForPrefix(string) []string
}

// easy to be counted exported by prefix
type cacheDoubleStore map[string]map[string]interface{}

func newDoubleStore() cacheDoubleStore {
	return make(cacheDoubleStore)
}

func (cs cacheDoubleStore) Put(key string, value interface{}) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	mp, ok := cs[prefix]
	if !ok {
		mp = make(map[string]interface{})
		cs[prefix] = mp
	}
	mp[key] = value
	/*if err := dumper.put(prefix, key, value); err != nil {
		utils.Logger.Info("<cache dumper> put error: " + err.Error())
	}*/
}

func (cs cacheDoubleStore) Get(key string) (interface{}, bool) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap[key]; exists {
			return ti, true
		}
	}
	return nil, false
}

func (cs cacheDoubleStore) Delete(key string) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		delete(keyMap, key)
		/*if err := dumper.delete(prefix, key); err != nil {
			utils.Logger.Info("<cache dumper> delete error: " + err.Error())
		}*/
	}
}

func (cs cacheDoubleStore) DeletePrefix(prefix string) {
	delete(cs, prefix)

	/*if err := dumper.deleteAll(prefix); err != nil {
		utils.Logger.Info("<cache dumper> delete all error: " + err.Error())
	}*/
}

func (cs cacheDoubleStore) CountEntriesForPrefix(prefix string) int {
	if m, ok := cs[prefix]; ok {
		return len(m)
	}
	return 0
}

func (cs cacheDoubleStore) GetAllForPrefix(prefix string) map[string]interface{} {
	if keyMap, ok := cs[prefix]; ok {
		return keyMap
	}
	return nil
}

func (cs cacheDoubleStore) GetKeysForPrefix(prefix string) (keys []string) {
	prefix, key := prefix[:PREFIX_LEN], prefix[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		for iterKey := range keyMap {
			if len(key) == 0 || strings.HasPrefix(iterKey, key) {
				keys = append(keys, prefix+iterKey)
			}
		}
	}
	return
}

type cacheParam struct {
	limit      int
	expiration time.Duration
}

func (ct *cacheParam) createCache() *Cache {
	return NewLRUTTL(ct.limit, ct.expiration)
}

type cacheLRUTTL map[string]*Cache

func newLRUTTL(types map[string]*cacheParam) cacheLRUTTL {
	c := make(map[string]*Cache, len(types))
	for prefix, param := range types {
		c[prefix] = param.createCache()
	}

	return c
}

func (cs cacheLRUTTL) Put(key string, value interface{}) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	mp, ok := cs[prefix]
	if !ok {
		mp = NewLRUTTL(1000, 0) //FixME
		cs[prefix] = mp
	}
	mp.Set(key, value)
}

func (cs cacheLRUTTL) Get(key string) (interface{}, bool) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap.Get(key); exists {
			return ti, true
		}
	}
	return nil, false
}

func (cs cacheLRUTTL) Delete(key string) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		keyMap.Remove(key)
	}
}

func (cs cacheLRUTTL) DeletePrefix(prefix string) {
	delete(cs, prefix)
}

func (cs cacheLRUTTL) CountEntriesForPrefix(prefix string) int {
	if m, ok := cs[prefix]; ok {
		return m.Len()
	}
	return 0
}

func (cs cacheLRUTTL) GetAllForPrefix(prefix string) map[string]interface{} {
	return nil
}

func (cs cacheLRUTTL) GetKeysForPrefix(prefix string) (keys []string) {
	return nil
}

func (cs cacheLRUTTL) Load(path string, prefixes []string) error {
	return utils.ErrNotImplemented
}

// faster to access
type cacheSimpleStore struct {
	cache    map[string]interface{}
	counters map[string]int
}

func newSimpleStore() cacheSimpleStore {
	return cacheSimpleStore{
		cache:    make(map[string]interface{}),
		counters: make(map[string]int),
	}
}

func (cs cacheSimpleStore) Put(key string, value interface{}) {
	if _, ok := cs.cache[key]; !ok {
		// only count if the key is not already there
		cs.count(key)
	}
	cs.cache[key] = value
}

func (cs cacheSimpleStore) Get(key string) (interface{}, bool) {
	if value, exists := cs.cache[key]; exists {
		return value, true
	}
	return nil, false
}

func (cs cacheSimpleStore) Delete(key string) {
	if _, ok := cs.cache[key]; ok {
		delete(cs.cache, key)
		cs.descount(key)
	}
}

func (cs cacheSimpleStore) DeletePrefix(prefix string) {
	for key, _ := range cs.cache {
		if strings.HasPrefix(key, prefix) {
			delete(cs.cache, key)
			cs.descount(key)
		}
	}
}

// increments the counter for the specified key prefix
func (cs cacheSimpleStore) count(key string) {
	if len(key) < PREFIX_LEN {
		return
	}
	prefix := key[:PREFIX_LEN]
	if _, ok := cs.counters[prefix]; ok {
		// increase the value
		cs.counters[prefix] += 1
	} else {
		cs.counters[prefix] = 1
	}
}

// decrements the counter for the specified key prefix
func (cs cacheSimpleStore) descount(key string) {
	if len(key) < PREFIX_LEN {
		return
	}
	prefix := key[:PREFIX_LEN]
	if value, ok := cs.counters[prefix]; ok && value > 0 {
		cs.counters[prefix] -= 1
	}
}

func (cs cacheSimpleStore) CountEntriesForPrefix(prefix string) int {
	if _, ok := cs.counters[prefix]; ok {
		return cs.counters[prefix]
	}
	return 0
}

func (cs cacheSimpleStore) GetAllForPrefix(prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	found := false
	for key, ti := range cs.cache {
		if strings.HasPrefix(key, prefix) {
			result[key[PREFIX_LEN:]] = ti
			found = true
		}
	}
	if !found {
		return nil
	}
	return result
}

func (cs cacheSimpleStore) GetKeysForPrefix(prefix string) (keys []string) {
	for key, _ := range cs.cache {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return
}

func (cs cacheSimpleStore) Load(path string, keys []string) error {
	return utils.ErrNotImplemented
}
