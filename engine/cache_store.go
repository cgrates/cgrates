//Simple caching library with expiration capabilities
package engine

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/syndtr/goleveldb/leveldb"
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
	Save(string, []string, *utils.CacheFileInfo) error
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

func (cs cacheDoubleStore) Save(path string, prefixes []string, cfi *utils.CacheFileInfo) error {
	//log.Printf("path: %s prefixes: %v", path, prefixes)
	if path == "" || len(prefixes) == 0 {
		return nil
	}
	//log.Print("saving cache prefixes: ", prefixes)
	// create a the path
	if err := os.MkdirAll(path, 0766); err != nil {
		utils.Logger.Info("<cache encoder>:" + err.Error())
		return err
	}

	var wg sync.WaitGroup
	for _, prefix := range prefixes {
		prefix = prefix[:PREFIX_LEN]
		mapValue, found := cs[prefix]
		if !found {
			continue
		}
		wg.Add(1)
		go func(key string, data map[string]interface{}) {
			defer wg.Done()

			dataEncoder := NewCodecMsgpackMarshaler()
			db, err := leveldb.OpenFile(filepath.Join(path, key+".cache"), nil)
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

			for k, v := range data {
				if encData, err := dataEncoder.Marshal(v); err == nil {
					if len(encData) > 1000 {
						var buf bytes.Buffer
						w := zlib.NewWriter(&buf)
						w.Write(encData)
						w.Close()
						encData = buf.Bytes()
					}
					db.Put([]byte(k), encData, nil)
				} else {
					utils.Logger.Info("<cache encoder>:" + err.Error())
					break
				}
			}
		}(prefix, mapValue)
	}
	wg.Wait()
	return utils.SaveCacheFileInfo(path, cfi)
}

func (cs cacheDoubleStore) Load(path string, prefixes []string) error {
	if path == "" || len(prefixes) == 0 {
		return nil
	}
	start := time.Now()
	var wg sync.WaitGroup
	var mux sync.Mutex
	for _, prefix := range prefixes {
		prefix = prefix[:PREFIX_LEN] // make sure it's only limited to prefix length'
		p := filepath.Join(path, prefix+".cache")
		if _, err := os.Stat(p); os.IsNotExist(err) { // no cache file for this prefix
			continue
		}
		wg.Add(1)
		go func(dirPath, key string) {
			defer wg.Done()
			db, err := leveldb.OpenFile(dirPath, nil)
			if err != nil {
				utils.Logger.Info("<cache decoder>: " + err.Error())
				return
			}
			defer db.Close()
			dataDecoder := NewCodecMsgpackMarshaler()
			val := make(map[string]interface{})
			iter := db.NewIterator(nil, nil)
			for iter.Next() {
				// Remember that the contents of the returned slice should not be modified, and
				// only valid until the next call to Next.
				k := iter.Key()
				data := iter.Value()
				var encData []byte
				if data[0] == 120 && data[1] == 156 { //zip header
					x := bytes.NewBuffer(data)
					r, err := zlib.NewReader(x)
					if err != nil {
						//log.Printf("%s err3", key)
						utils.Logger.Info("<cache decoder>: " + err.Error())
						break
					}
					out, err := ioutil.ReadAll(r)
					if err != nil {
						//log.Printf("%s err4", key)
						utils.Logger.Info("<cache decoder>: " + err.Error())
						break
					}
					r.Close()
					encData = out
				} else {
					encData = data
				}
				kv := CacheTypeFactory(key, "", nil)
				v := kv.Value()
				if err := dataDecoder.Unmarshal(encData, &v); err != nil {
					//log.Printf("%s err5", key)
					utils.Logger.Info("<cache decoder>: " + err.Error())
					break
				}
				val[string(k)] = v
			}
			iter.Release()
			mux.Lock()
			cs[key] = val
			mux.Unlock()
		}(p, prefix)
	}
	wg.Wait()
	utils.Logger.Info(fmt.Sprintf("Cache %v load time: %v", prefixes, time.Since(start)))
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

func (cs cacheSimpleStore) Save(path string, keys []string, cfi *utils.CacheFileInfo) error {
	utils.Logger.Info("simplestore save")
	return nil
}

func (cs cacheSimpleStore) Load(path string, keys []string) error {
	utils.Logger.Info("simplestore load")
	return nil
}
