//Simple caching library with expiration capabilities
package cache2go

import (
	"encoding/gob"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/utils"
)

type cacheStore interface {
	Put(string, interface{})
	Append(string, interface{})
	Get(string) (interface{}, error)
	Pop(string, interface{})
	Delete(string)
	DeletePrefix(string)
	CountEntriesForPrefix(string) int
	GetAllForPrefix(string) (map[string]interface{}, error)
	GetKeysForPrefix(string) []string
	Save(string, []string) error
	Load(string, []string) error
}

// easy to be counted exported by prefix
type cacheDoubleStore map[string]map[string]interface{}

type KeyValue struct {
	Key   string
	Value interface{}
}

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
}

func (cs cacheDoubleStore) Append(key string, value interface{}) {
	var elements map[interface{}]struct{} // using map for faster check if element is present
	if v, err := cs.Get(key); err == nil {
		elements = v.(map[interface{}]struct{})
	} else {
		elements = make(map[interface{}]struct{})
	}
	elements[value] = struct{}{}
	cache.Put(key, elements)
}

func (cs cacheDoubleStore) Get(key string) (interface{}, error) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap[key]; exists {
			return ti, nil
		}
	}
	return nil, utils.ErrNotFound
}

func (cs cacheDoubleStore) Pop(key string, value interface{}) {
	if v, err := cs.Get(key); err == nil {
		elements, ok := v.(map[interface{}]struct{})
		if ok {
			delete(elements, value)
			if len(elements) > 0 {
				cache.Put(key, elements)
			} else {
				cache.Delete(key)
			}
		}
	}
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

func (cs cacheDoubleStore) CountEntriesForPrefix(prefix string) int {
	if m, ok := cs[prefix]; ok {
		return len(m)
	}
	return 0
}

func (cs cacheDoubleStore) GetAllForPrefix(prefix string) (map[string]interface{}, error) {
	if keyMap, ok := cs[prefix]; ok {
		return keyMap, nil
	}
	return nil, utils.ErrNotFound
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

func (cs cacheDoubleStore) Save(path string, keys []string) error {
	// create a the path
	if err := os.MkdirAll(path, 0766); err != nil {
		utils.Logger.Info("<cache encoder>:" + err.Error())
		return err
	}

	var wg sync.WaitGroup
	for _, key := range keys {
		key = key[:PREFIX_LEN]
		value, found := cs[key]
		if !found {
			continue
		}
		wg.Add(1)
		go func(fileName string, data map[string]interface{}) {
			defer wg.Done()
			dataFile, err := os.Create(filepath.Join(path, fileName) + ".cache")
			defer dataFile.Close()
			if err != nil {
				utils.Logger.Info("<cache encoder>:" + err.Error())
			}

			// serialize the data
			dataEncoder := gob.NewEncoder(dataFile)
			log.Print("start: ", fileName)
			if err := dataEncoder.Encode(data); err != nil {
				log.Print("err: ", fileName, err)
				utils.Logger.Info("<cache encoder>:" + err.Error())
			} else {
				log.Print("end: ", fileName, err)
			}
		}(key, value)
	}
	wg.Wait()
	return nil
}

func (cs cacheDoubleStore) Load(path string, keys []string) error {
	var wg sync.WaitGroup
	for _, key := range keys {
		key = key[:PREFIX_LEN] // make sure it's only limited to prefix length'
		file := filepath.Join(path, key+".cache")
		wg.Add(1)
		go func(fileName, key string) {
			defer wg.Done()

			// open data file
			dataFile, err := os.Open(fileName)
			defer dataFile.Close()
			if err != nil {
				utils.Logger.Info("<cache decoder>: " + err.Error())
			}

			val := make(map[string]interface{})
			dataDecoder := gob.NewDecoder(dataFile)
			err = dataDecoder.Decode(&val)
			if err != nil {
				utils.Logger.Info("<cache decoder>: " + err.Error())
			}
			cs[key] = val
		}(file, key)
	}
	wg.Wait()
	return nil
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

func (cs cacheSimpleStore) Append(key string, value interface{}) {
	var elements map[interface{}]struct{}
	if v, err := cs.Get(key); err == nil {
		elements = v.(map[interface{}]struct{})
	} else {
		elements = make(map[interface{}]struct{})
	}
	elements[value] = struct{}{}
	cache.Put(key, elements)
}

func (cs cacheSimpleStore) Get(key string) (interface{}, error) {
	if value, exists := cs.cache[key]; exists {
		return value, nil
	}
	return nil, utils.ErrNotFound
}

func (cs cacheSimpleStore) Pop(key string, value interface{}) {
	if v, err := cs.Get(key); err == nil {
		elements, ok := v.(map[interface{}]struct{})
		if ok {
			delete(elements, value)
			if len(elements) > 0 {
				cache.Put(key, elements)
			} else {
				cache.Delete(key)
			}
		}
	}
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

func (cs cacheSimpleStore) GetAllForPrefix(prefix string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	found := false
	for key, ti := range cs.cache {
		if strings.HasPrefix(key, prefix) {
			result[key[PREFIX_LEN:]] = ti
			found = true
		}
	}
	if !found {
		return nil, utils.ErrNotFound
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

func (cs cacheSimpleStore) Save(path string, keys []string) error {
	utils.Logger.Info("simplestore save")
	return nil
}

func (cs cacheSimpleStore) Load(path string, keys []string) error {
	utils.Logger.Info("simplestore load")
	return nil
}
