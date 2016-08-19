//Simple caching library with expiration capabilities
package cache2go

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	lru "github.com/hashicorp/golang-lru"
)

type cacheStore interface {
	Put(string, interface{})
	Get(string) (interface{}, bool)
	Delete(string)
	DeletePrefix(string)
	CountEntriesForPrefix(string) int
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

// easy to be counted exported by prefix
type lrustore map[string]*lru.Cache

func newLruStore() lrustore {
	c := make(lrustore)
	if cfg != nil && cfg.Destinations != nil {
		c[utils.DESTINATION_PREFIX], _ = lru.New(cfg.Destinations.Limit)
	} else {
		c[utils.DESTINATION_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.ReverseDestinations != nil {
		c[utils.REVERSE_DESTINATION_PREFIX], _ = lru.New(cfg.ReverseDestinations.Limit)
	} else {
		c[utils.REVERSE_DESTINATION_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.RatingPlans != nil {
		c[utils.RATING_PLAN_PREFIX], _ = lru.New(cfg.RatingPlans.Limit)
	} else {
		c[utils.RATING_PLAN_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.RatingProfiles != nil {
		c[utils.RATING_PROFILE_PREFIX], _ = lru.New(cfg.RatingProfiles.Limit)
	} else {
		c[utils.RATING_PROFILE_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.Lcr != nil {
		c[utils.LCR_PREFIX], _ = lru.New(cfg.Lcr.Limit)
	} else {
		c[utils.LCR_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.CdrStats != nil {
		c[utils.CDR_STATS_PREFIX], _ = lru.New(cfg.CdrStats.Limit)
	} else {
		c[utils.CDR_STATS_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.Actions != nil {
		c[utils.ACTION_PREFIX], _ = lru.New(cfg.Actions.Limit)
	} else {
		c[utils.ACTION_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.ActionPlans != nil {
		c[utils.ACTION_PLAN_PREFIX], _ = lru.New(cfg.ActionPlans.Limit)
	} else {
		c[utils.ACTION_PLAN_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.ActionTriggers != nil {
		c[utils.ACTION_TRIGGER_PREFIX], _ = lru.New(cfg.ActionTriggers.Limit)
	} else {
		c[utils.ACTION_TRIGGER_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.SharedGroups != nil {
		c[utils.SHARED_GROUP_PREFIX], _ = lru.New(cfg.SharedGroups.Limit)
	} else {
		c[utils.SHARED_GROUP_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.Aliases != nil {
		c[utils.ALIASES_PREFIX], _ = lru.New(cfg.Aliases.Limit)
	} else {
		c[utils.ALIASES_PREFIX], _ = lru.New(10000)
	}
	if cfg != nil && cfg.ReverseAliases != nil {
		c[utils.REVERSE_ALIASES_PREFIX], _ = lru.New(cfg.ReverseAliases.Limit)
	} else {
		c[utils.REVERSE_ALIASES_PREFIX], _ = lru.New(10000)
	}

	return c
}

func (cs lrustore) Put(key string, value interface{}) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	mp, ok := cs[prefix]
	if !ok {
		var err error
		mp, err = lru.New(10000)
		if err != nil {
			utils.Logger.Debug(fmt.Sprintf("<cache>: error at init: %v", err))
		}
		cs[prefix] = mp
	}
	mp.Add(key, value)
}

func (cs lrustore) Get(key string) (interface{}, bool) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		if ti, exists := keyMap.Get(key); exists {
			return ti, true
		}
	}
	return nil, false
}

func (cs lrustore) Delete(key string) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		keyMap.Remove(key)
	}
}

func (cs lrustore) DeletePrefix(prefix string) {
	delete(cs, prefix)
}

func (cs lrustore) CountEntriesForPrefix(prefix string) int {
	if m, ok := cs[prefix]; ok {
		return m.Len()
	}
	return 0
}

func (cs lrustore) GetKeysForPrefix(prefix string) (keys []string) {
	prefix, key := prefix[:PREFIX_LEN], prefix[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		for _, iterKey := range keyMap.Keys() {
			iterKeyString := iterKey.(string)
			if len(key) == 0 || strings.HasPrefix(iterKeyString, key) {
				keys = append(keys, prefix+iterKeyString)
			}
		}
	}
	return
}

type cacheLRUTTL map[string]*Cache

func newLRUTTL(cfg *config.CacheConfig) cacheLRUTTL {
	c := make(cacheLRUTTL)
	if cfg != nil && cfg.Destinations != nil {
		c[utils.DESTINATION_PREFIX] = NewLRUTTL(cfg.Destinations.Limit, cfg.Destinations.TTL)
	} else {
		c[utils.DESTINATION_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.ReverseDestinations != nil {
		c[utils.REVERSE_DESTINATION_PREFIX] = NewLRUTTL(cfg.ReverseDestinations.Limit, cfg.ReverseDestinations.TTL)
	} else {
		c[utils.REVERSE_DESTINATION_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.RatingPlans != nil {
		c[utils.RATING_PLAN_PREFIX] = NewLRUTTL(cfg.RatingPlans.Limit, cfg.RatingPlans.TTL)
	} else {
		c[utils.RATING_PLAN_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.RatingProfiles != nil {
		c[utils.RATING_PROFILE_PREFIX] = NewLRUTTL(cfg.RatingProfiles.Limit, cfg.RatingProfiles.TTL)
	} else {
		c[utils.RATING_PROFILE_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.Lcr != nil {
		c[utils.LCR_PREFIX] = NewLRUTTL(cfg.Lcr.Limit, cfg.Lcr.TTL)
	} else {
		c[utils.LCR_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.CdrStats != nil {
		c[utils.CDR_STATS_PREFIX] = NewLRUTTL(cfg.CdrStats.Limit, cfg.CdrStats.TTL)
	} else {
		c[utils.CDR_STATS_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.Actions != nil {
		c[utils.ACTION_PREFIX] = NewLRUTTL(cfg.Actions.Limit, cfg.Actions.TTL)
	} else {
		c[utils.ACTION_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.ActionPlans != nil {
		c[utils.ACTION_PLAN_PREFIX] = NewLRUTTL(cfg.ActionPlans.Limit, cfg.ActionPlans.TTL)
	} else {
		c[utils.ACTION_PLAN_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.ActionTriggers != nil {
		c[utils.ACTION_TRIGGER_PREFIX] = NewLRUTTL(cfg.ActionTriggers.Limit, cfg.ActionTriggers.TTL)
	} else {
		c[utils.ACTION_TRIGGER_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.SharedGroups != nil {
		c[utils.SHARED_GROUP_PREFIX] = NewLRUTTL(cfg.SharedGroups.Limit, cfg.SharedGroups.TTL)
	} else {
		c[utils.SHARED_GROUP_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.Aliases != nil {
		c[utils.ALIASES_PREFIX] = NewLRUTTL(cfg.Aliases.Limit, cfg.Aliases.TTL)
	} else {
		c[utils.ALIASES_PREFIX] = NewLRUTTL(10000, 0)
	}
	if cfg != nil && cfg.ReverseAliases != nil {
		c[utils.REVERSE_ALIASES_PREFIX] = NewLRUTTL(cfg.ReverseAliases.Limit, cfg.ReverseAliases.TTL)
	} else {
		c[utils.REVERSE_ALIASES_PREFIX] = NewLRUTTL(10000, 0)
	}

	return c
}

func (cs cacheLRUTTL) Put(key string, value interface{}) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	mp, ok := cs[prefix]
	if !ok {
		mp = NewLRUTTL(10000, 0)
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

func (cs cacheLRUTTL) GetKeysForPrefix(prefix string) (keys []string) {
	prefix, key := prefix[:PREFIX_LEN], prefix[PREFIX_LEN:]
	if keyMap, ok := cs[prefix]; ok {
		for iterKey := range keyMap.cache {
			if len(key) == 0 || strings.HasPrefix(iterKey, key) {
				keys = append(keys, prefix+iterKey)
			}
		}
	}
	return
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

func (cs cacheSimpleStore) GetKeysForPrefix(prefix string) (keys []string) {
	for key, _ := range cs.cache {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return
}
