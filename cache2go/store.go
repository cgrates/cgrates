//Simple caching library with expiration capabilities
package cache2go

import (
	"errors"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type cacheStore interface {
	Put(string, interface{})
	Append(string, interface{})
	Get(string) (interface{}, error)
	GetAge(string) (time.Duration, error)
	Delete(string)
	DeletePrefix(string)
	CountEntriesForPrefix(string) int64
	GetAllForPrefix(string) (map[string]timestampedValue, error)
	GetKeysForPrefix(string) []string
}

// easy to be counted exported by prefix
type cacheDoubleStore map[string]map[string]timestampedValue

func newDoubleStore() cacheDoubleStore {
	return make(cacheDoubleStore)
}

func (cs cacheDoubleStore) Put(key string, value interface{}) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if _, ok := cs[prefix]; !ok {
		cs[prefix] = make(map[string]timestampedValue)
	}
	cs[prefix][key] = timestampedValue{time.Now(), value}
}

func (cs cacheDoubleStore) Append(key string, value interface{}) {
	var elements map[interface{}]struct{}
	if v, err := cs.Get(key); err == nil {
		elements = v.(map[interface{}]struct{})
	} else {
		elements = make(map[interface{}]struct{})
	}
	// check if the val is already present
	if _, found := elements[value]; !found {
		elements[value] = struct{}{}
	}
	cache.Put(key, elements)
}

func (cs cacheDoubleStore) Get(key string) (interface{}, error) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap[key]; exists {
			return ti.value, nil
		}
	}
	return nil, errors.New(utils.ERR_NOT_FOUND)
}

func (cs cacheDoubleStore) GetAge(key string) (time.Duration, error) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap[key]; exists {
			return time.Since(ti.timestamp), nil
		}
	}
	return -1, errors.New(utils.ERR_NOT_FOUND)
}

func (cs cacheDoubleStore) Delete(key string) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		delete(keyMap, key)
	}
}

func (cs cacheDoubleStore) DeletePrefix(prefix string) {
	delete(cs, prefix)
}

func (cs cacheDoubleStore) CountEntriesForPrefix(prefix string) int64 {
	if _, ok := cs[prefix]; ok {
		return int64(len(cs[prefix]))
	}
	return 0
}

func (cs cacheDoubleStore) GetAllForPrefix(prefix string) (map[string]timestampedValue, error) {
	if keyMap, ok := cs[prefix]; ok {
		return keyMap, nil
	}
	return nil, errors.New(utils.ERR_NOT_FOUND)
}

func (cs cacheDoubleStore) GetKeysForPrefix(prefix string) (keys []string) {
	prefix, key := prefix[:PREFIX_LEN], prefix[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		for iterKey := range keyMap {
			if len(key) > 0 && strings.HasPrefix(iterKey, key) {
				keys = append(keys, prefix+iterKey)
			}
		}
	}
	return
}

// faster to access
type cacheSimpleStore struct {
	cache    map[string]timestampedValue
	counters map[string]int64
}

func newSimpleStore() cacheSimpleStore {
	return cacheSimpleStore{
		cache:    make(map[string]timestampedValue),
		counters: make(map[string]int64),
	}
}

func (cs cacheSimpleStore) Put(key string, value interface{}) {
	if _, ok := cs.cache[key]; !ok {
		// only count if the key is not already there
		cs.count(key)
	}
	cs.cache[key] = timestampedValue{time.Now(), value}
}

func (cs cacheSimpleStore) Append(key string, value interface{}) {
	var elements map[interface{}]struct{}
	if v, err := cs.Get(key); err == nil {
		elements = v.(map[interface{}]struct{})
	} else {
		elements = make(map[interface{}]struct{})
	}
	// check if the val is already present
	if _, found := elements[value]; !found {
		elements[value] = struct{}{}
	}
	cache.Put(key, elements)
}

func (cs cacheSimpleStore) Get(key string) (interface{}, error) {
	if ti, exists := cs.cache[key]; exists {
		return ti.value, nil
	}
	return nil, errors.New(utils.ERR_NOT_FOUND)
}

func (cs cacheSimpleStore) GetAge(key string) (time.Duration, error) {
	if ti, exists := cs.cache[key]; exists {
		return time.Since(ti.timestamp), nil
	}

	return -1, errors.New(utils.ERR_NOT_FOUND)
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

func (cs cacheSimpleStore) CountEntriesForPrefix(prefix string) int64 {
	if _, ok := cs.counters[prefix]; ok {
		return cs.counters[prefix]
	}
	return 0
}

func (cs cacheSimpleStore) GetAllForPrefix(prefix string) (map[string]timestampedValue, error) {
	result := make(map[string]timestampedValue)
	found := false
	for key, ti := range cs.cache {
		if strings.HasPrefix(key, prefix) {
			result[key[PREFIX_LEN:]] = ti
			found = true
		}
	}
	if !found {
		return nil, errors.New(utils.ERR_NOT_FOUND)
	}
	return result, nil
}

func (cs cacheSimpleStore) GetKeysForPrefix(prefix string) (keys []string) {
	for key, _ := range cs.cache {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return
}
