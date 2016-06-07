//Simple caching library with expiration capabilities
package engine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/utils"
)

type cacheStore interface {
	Put(string, interface{})
	Get(string) (interface{}, error)
	Append(string, string)
	Pop(string, string)
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

func (cs cacheDoubleStore) Get(key string) (interface{}, error) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap[key]; exists {
			return ti, nil
		}
	}
	return nil, utils.ErrNotFound
}

func (cs cacheDoubleStore) Append(key string, value string) {
	var elements map[string]struct{} // using map for faster check if element is present
	if v, err := cs.Get(key); err == nil {
		elements = v.(map[string]struct{})
	} else {
		elements = make(map[string]struct{})
	}
	elements[value] = struct{}{}
	cache.Put(key, elements)
}

func (cs cacheDoubleStore) Pop(key string, value string) {
	if v, err := cs.Get(key); err == nil {
		elements, ok := v.(map[string]struct{})
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

func (cs cacheDoubleStore) Save(path string, prefixes []string) error {
	// create a the path
	if err := os.MkdirAll(path, 0766); err != nil {
		utils.Logger.Info("<cache encoder>:" + err.Error())
		return err
	}

	var wg sync.WaitGroup
	for _, prefix := range prefixes {
		prefix = prefix[:PREFIX_LEN]
		value, found := cs[prefix]
		if !found {
			continue
		}
		wg.Add(1)
		go func(key string, data map[string]interface{}) {
			defer wg.Done()
			dataFile, err := os.Create(filepath.Join(path, key) + ".cache")
			defer dataFile.Close()
			if err != nil {
				utils.Logger.Info("<cache encoder>:" + err.Error())
			}

			// serialize the data
			w := bufio.NewWriter(dataFile)
			dataEncoder := json.NewEncoder(w)
			log.Print("start: ", key)
			for k, v := range data {
				if err := dataEncoder.Encode(CacheTypeFactory(key, k, v)); err != nil {
					log.Print("err: ", key, err)
					utils.Logger.Info("<cache encoder>:" + err.Error())
					break
				}
			}
			log.Print("end: ", key, err)
			w.Flush()
		}(prefix, value)
	}
	wg.Wait()
	return nil
}

func (cs cacheDoubleStore) Load(path string, prefixes []string) error {
	var wg sync.WaitGroup
	for _, prefix := range prefixes {
		prefix = prefix[:PREFIX_LEN] // make sure it's only limited to prefix length'
		file := filepath.Join(path, prefix+".cache")
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
			scanner := bufio.NewScanner(dataFile)
			for scanner.Scan() {
				dataDecoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
				kv := CacheTypeFactory(key, "", nil)
				err = dataDecoder.Decode(&kv)
				if err != nil {
					log.Printf("err: %v", err)
					utils.Logger.Info("<cache decoder1>: " + err.Error())
					break
				}
				val[kv.Key()] = kv.Value()
			}
			if err := scanner.Err(); err != nil {
				utils.Logger.Info("<cache decoder2>: " + err.Error())
			}
			cs[key] = val
		}(file, prefix)
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

func (cs cacheSimpleStore) Append(key string, value string) {
	var elements map[string]struct{}
	if v, err := cs.Get(key); err == nil {
		elements = v.(map[string]struct{})
	} else {
		elements = make(map[string]struct{})
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

func (cs cacheSimpleStore) Pop(key string, value string) {
	if v, err := cs.Get(key); err == nil {
		elements, ok := v.(map[string]struct{})
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

type cacheKeyValue interface {
	Key() string
	Value() interface{}
}

type mapKeyValue struct {
	K string
	V map[string]struct{}
}

func (mkv *mapKeyValue) Key() string {
	return mkv.K
}

func (mkv *mapKeyValue) Value() interface{} {
	return mkv.V
}

type rpKeyValue struct {
	K string
	V *RatingPlan
}

func (mkv *rpKeyValue) Key() string {
	return mkv.K
}

func (mkv *rpKeyValue) Value() interface{} {
	return mkv.V
}

type rpfKeyValue struct {
	K string
	V *RatingProfile
}

func (mkv *rpfKeyValue) Key() string {
	return mkv.K
}

func (mkv *rpfKeyValue) Value() interface{} {
	return mkv.V
}

type lcrKeyValue struct {
	K string
	V *LCR
}

func (mkv *lcrKeyValue) Key() string {
	return mkv.K
}

func (mkv *lcrKeyValue) Value() interface{} {
	return mkv.V
}

type dcKeyValue struct {
	K string
	V *utils.DerivedChargers
}

func (mkv *dcKeyValue) Key() string {
	return mkv.K
}

func (mkv *dcKeyValue) Value() interface{} {
	return mkv.V
}

type acsKeyValue struct {
	K string
	V Actions
}

func (mkv *acsKeyValue) Key() string {
	return mkv.K
}

func (mkv *acsKeyValue) Value() interface{} {
	return mkv.V
}

type aplKeyValue struct {
	K string
	V *ActionPlan
}

func (mkv *aplKeyValue) Key() string {
	return mkv.K
}

func (mkv *aplKeyValue) Value() interface{} {
	return mkv.V
}

type sgKeyValue struct {
	K string
	V *SharedGroup
}

func (mkv *sgKeyValue) Key() string {
	return mkv.K
}

func (mkv *sgKeyValue) Value() interface{} {
	return mkv.V
}

type alsKeyValue struct {
	K string
	V AliasValues
}

func (mkv *alsKeyValue) Key() string {
	return mkv.K
}

func (mkv *alsKeyValue) Value() interface{} {
	return mkv.V
}

type loadKeyValue struct {
	K string
	V []*utils.LoadInstance
}

func (mkv *loadKeyValue) Key() string {
	return mkv.K
}

func (mkv *loadKeyValue) Value() interface{} {
	return mkv.V
}

func CacheTypeFactory(prefix string, key string, value interface{}) cacheKeyValue {
	switch prefix {
	case utils.DESTINATION_PREFIX:
		if value != nil {
			return &mapKeyValue{key, value.(map[string]struct{})}
		}
		return &mapKeyValue{"", make(map[string]struct{})}
	case utils.RATING_PLAN_PREFIX:
		if value != nil {
			return &rpKeyValue{key, value.(*RatingPlan)}
		}
		return &rpfKeyValue{"", &RatingProfile{}}
	case utils.RATING_PROFILE_PREFIX:
		if value != nil {
			return &rpfKeyValue{key, value.(*RatingProfile)}
		}
		return &rpfKeyValue{"", &RatingProfile{}}
	case utils.LCR_PREFIX:
		if value != nil {
			return &lcrKeyValue{key, value.(*LCR)}
		}
		return &lcrKeyValue{"", &LCR{}}
	case utils.DERIVEDCHARGERS_PREFIX:
		if value != nil {
			return &dcKeyValue{key, value.(*utils.DerivedChargers)}
		}
		return &dcKeyValue{"", &utils.DerivedChargers{}}
	case utils.ACTION_PREFIX:
		if value != nil {
			return &acsKeyValue{key, value.(Actions)}
		}
		return &acsKeyValue{"", Actions{}}
	case utils.ACTION_PLAN_PREFIX:
		if value != nil {
			return &aplKeyValue{key, value.(*ActionPlan)}
		}
		return &aplKeyValue{"", &ActionPlan{}}
	case utils.SHARED_GROUP_PREFIX:
		if value != nil {
			return &sgKeyValue{key, value.(*SharedGroup)}
		}
		return &sgKeyValue{"", &SharedGroup{}}
	case utils.ALIASES_PREFIX:
		if value != nil {
			return &alsKeyValue{key, value.(AliasValues)}
		}
		return &alsKeyValue{"", AliasValues{}}
	case utils.LOADINST_KEY[:PREFIX_LEN]:
		if value != nil {
			return &loadKeyValue{key, value.([]*utils.LoadInstance)}
		}
		return &loadKeyValue{"", make([]*utils.LoadInstance, 0)}
	}
	return nil
}
