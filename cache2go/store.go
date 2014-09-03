//Simple caching library with expiration capabilities
package cache2go

import (
	"errors"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type cacheStore map[string]map[string]timestampedValue

func (cs cacheStore) Put(key string, value interface{}) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if _, ok := cs[prefix]; !ok {
		cs[prefix] = make(map[string]timestampedValue)
	}
	cs[prefix][key] = timestampedValue{time.Now(), value}
}

func (cs cacheStore) Append(key string, value interface{}) {
	var elements []interface{}
	v, err := cs.Get(key)
	if err == nil {
		elements = v.([]interface{})
	}
	// check if the val is already present
	found := false
	for _, v := range elements {
		if value == v {
			found = true
			break
		}
	}
	if !found {
		elements = append(elements, value)
	}
	cache.Put(key, elements)
}

func (cs cacheStore) Get(key string) (interface{}, error) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap[key]; exists {
			return ti.value, nil
		}
	}
	return nil, errors.New(utils.ERR_NOT_FOUND)
}

func (cs cacheStore) GetAge(key string) (time.Duration, error) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap[key]; exists {
			return time.Since(ti.timestamp), nil
		}
	}
	return -1, errors.New(utils.ERR_NOT_FOUND)
}

func (cs cacheStore) Delete(key string) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if _, exists := keyMap[key]; exists {
			delete(keyMap, key)
		}
	}
}

func (cs cacheStore) DeletePrefix(prefix string) {
	if _, ok := cs[prefix]; ok {
		delete(cs, prefix)
	}
}
