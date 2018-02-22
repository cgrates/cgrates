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
)

type cacheStore interface {
	Put(string, interface{}, []string)
	Get(string) (interface{}, bool)
	Delete(string)
	DeletePrefix(string)
	CountEntriesForPrefix(string) int
	GetKeysForPrefix(string) []string
	Clear()
}

type cacheLRUTTL map[string]*ltcache.Cache

func newLRUTTL(cfg config.CacheConfig) (c cacheLRUTTL) {
	c = map[string]*ltcache.Cache{
		utils.ANY: ltcache.New(ltcache.UnlimitedCaching, ltcache.UnlimitedCaching, false, nil), // no limits for default cache instance
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

func (cs cacheLRUTTL) Put(key string, value interface{}, grpIDs []string) {
	cs.cacheInstance(key[:PREFIX_LEN]).Set(key[PREFIX_LEN:], value, grpIDs)
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
		for _, iterKey := range c.Items() {
			if len(key) == 0 || strings.HasPrefix(iterKey, key) {
				keys = append(keys, prefix+iterKey)
			}
		}
	}
	return
}

func (cs cacheLRUTTL) Clear() {
	for _, cInst := range cs {
		cInst.Clear()
	}
}
