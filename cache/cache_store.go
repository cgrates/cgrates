/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package cache

import (
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
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
type lrustore map[string]*lru.Cache

func newLruStore() (c lrustore) {
	c = make(map[string]*lru.Cache)
	c[utils.ANY], _ = lru.New(0)
	if cfg == nil {
		return
	}
	// dynamically configure cache instances based on CacheConfig
	for cfgKey := range cfg {
		cacheInstanceID := cfgKey
		if prefixKey, has := utils.CacheInstanceToPrefix[cfgKey]; has {
			cacheInstanceID = prefixKey // old aliases, backwards compatibility purpose
		}
		c[cacheInstanceID], _ = lru.New(cfg[cfgKey].Limit)
	}
	return
}

func (cs lrustore) Put(key string, value interface{}) {
	prefix, key := key[:PREFIX_LEN], key[PREFIX_LEN:]
	mp, ok := cs[prefix]
	if !ok {
		var err error
		mp, err = lru.New(10000)
		if err != nil {
			return
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

type cacheLRUTTL map[string]*ltcache.Cache

func newLRUTTL(cfg config.CacheConfig) (c cacheLRUTTL) {
	c = map[string]*ltcache.Cache{
		utils.ANY: ltcache.New(0, 0, false, nil), // no limits for default cache instance
	}
	if cfg == nil {
		return
	}
	// dynamically configure cache instances based on CacheConfig
	for cfgKey := range cfg {
		cacheInstanceID := cfgKey
		if prefixKey, has := utils.CacheInstanceToPrefix[cfgKey]; has {
			cacheInstanceID = prefixKey // old aliases, backwards compatibility purpose
		}
		c[cacheInstanceID] = ltcache.New(cfg[cfgKey].Limit, cfg[cfgKey].TTL, cfg[cfgKey].StaticTTL, nil)
	}
	return
}

func (cs cacheLRUTTL) cacheInstance(instID string) (c *ltcache.Cache) {
	var ok bool
	if c, ok = cs[instID]; !ok {
		c = cs[utils.ANY]
	}
	return
}

func (cs cacheLRUTTL) Put(key string, value interface{}) {
	cs.cacheInstance(key[:PREFIX_LEN]).Set(key[PREFIX_LEN:], value)
}

func (cs cacheLRUTTL) Get(key string) (interface{}, bool) {
	return cs.cacheInstance(key[:PREFIX_LEN]).Get(key[PREFIX_LEN:])
}

func (cs cacheLRUTTL) Delete(key string) {
	cs.cacheInstance(key[:PREFIX_LEN]).Remove(key[PREFIX_LEN:])
}

func (cs cacheLRUTTL) DeletePrefix(prefix string) {
	if c, hasInst := cs[prefix]; hasInst {
		c.Clear()
	}
}

func (cs cacheLRUTTL) CountEntriesForPrefix(prefix string) (cnt int) {
	if c, ok := cs[prefix]; ok {
		cnt = c.Len()
	}
	return
}

func (cs cacheLRUTTL) GetKeysForPrefix(prefix string) (keys []string) {
	prefix, key := prefix[:PREFIX_LEN], prefix[PREFIX_LEN:]
	if c, ok := cs[prefix]; ok {
		for _, ifKey := range c.Keys() {
			iterKey := ifKey.(string)
			if len(key) == 0 || strings.HasPrefix(iterKey, key) {
				keys = append(keys, prefix+iterKey)
			}
		}
	}
	return
}
