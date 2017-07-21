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

func newLruStore() lrustore {
	c := make(map[string]*lru.Cache)
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
	if cfg != nil && cfg.AccountActionPlans != nil {
		c[utils.AccountActionPlansPrefix], _ = lru.New(cfg.AccountActionPlans.Limit)
	} else {
		c[utils.AccountActionPlansPrefix], _ = lru.New(10000)
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

func newLRUTTL(cfg *config.CacheConfig) (c cacheLRUTTL) {
	c = map[string]*ltcache.Cache{
		utils.ANY: ltcache.New(0, 0, false, nil), // no limits for default cache instance
	}
	if cfg == nil {
		return
	}
	if cfg.Destinations != nil {
		c[utils.DESTINATION_PREFIX] = ltcache.New(cfg.Destinations.Limit, cfg.Destinations.TTL, false, nil)
	}
	if cfg.ReverseDestinations != nil {
		c[utils.REVERSE_DESTINATION_PREFIX] = ltcache.New(cfg.ReverseDestinations.Limit, cfg.ReverseDestinations.TTL, false, nil)
	}
	if cfg.RatingPlans != nil {
		c[utils.RATING_PLAN_PREFIX] = ltcache.New(cfg.RatingPlans.Limit, cfg.RatingPlans.TTL, false, nil)
	}
	if cfg.RatingProfiles != nil {
		c[utils.RATING_PROFILE_PREFIX] = ltcache.New(cfg.RatingProfiles.Limit, cfg.RatingProfiles.TTL, false, nil)
	}
	if cfg.Lcr != nil {
		c[utils.LCR_PREFIX] = ltcache.New(cfg.Lcr.Limit, cfg.Lcr.TTL, false, nil)
	}
	if cfg.CdrStats != nil {
		c[utils.CDR_STATS_PREFIX] = ltcache.New(cfg.CdrStats.Limit, cfg.CdrStats.TTL, false, nil)
	}
	if cfg.Actions != nil {
		c[utils.ACTION_PREFIX] = ltcache.New(cfg.Actions.Limit, cfg.Actions.TTL, false, nil)
	}
	if cfg.ActionPlans != nil {
		c[utils.ACTION_PLAN_PREFIX] = ltcache.New(cfg.ActionPlans.Limit, cfg.ActionPlans.TTL, false, nil)
	}
	if cfg.ActionTriggers != nil {
		c[utils.ACTION_TRIGGER_PREFIX] = ltcache.New(cfg.ActionTriggers.Limit, cfg.ActionTriggers.TTL, false, nil)
	}
	if cfg.SharedGroups != nil {
		c[utils.SHARED_GROUP_PREFIX] = ltcache.New(cfg.SharedGroups.Limit, cfg.SharedGroups.TTL, false, nil)
	}
	if cfg.Aliases != nil {
		c[utils.ALIASES_PREFIX] = ltcache.New(cfg.Aliases.Limit, cfg.Aliases.TTL, false, nil)
	}
	if cfg != nil && cfg.ReverseAliases != nil {
		c[utils.REVERSE_ALIASES_PREFIX] = ltcache.New(cfg.ReverseAliases.Limit, cfg.ReverseAliases.TTL, false, nil)
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
