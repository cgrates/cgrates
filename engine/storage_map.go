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
package engine

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type MapStorage struct {
	dict     storage
	tasks    [][]byte
	ms       Marshaler
	mu       sync.RWMutex
	cacheCfg *config.CacheConfig
}

type storage map[string][]byte

func (s storage) sadd(key, value string, ms Marshaler) {
	idMap := utils.StringMap{}
	if values, ok := s[key]; ok {
		ms.Unmarshal(values, &idMap)
	}
	idMap[value] = true
	values, _ := ms.Marshal(idMap)
	s[key] = values
}

func (s storage) srem(key, value string, ms Marshaler) {
	idMap := utils.StringMap{}
	if values, ok := s[key]; ok {
		ms.Unmarshal(values, &idMap)
	}
	delete(idMap, value)
	values, _ := ms.Marshal(idMap)
	s[key] = values
}

func (s storage) smembers(key string, ms Marshaler) (idMap utils.StringMap, ok bool) {
	var values []byte
	values, ok = s[key]
	if ok {
		ms.Unmarshal(values, &idMap)
	}
	return
}

func NewMapStorage() (*MapStorage, error) {
	return &MapStorage{dict: make(map[string][]byte), ms: NewCodecMsgpackMarshaler(), cacheCfg: &config.CacheConfig{RatingPlans: &config.CacheParamConfig{Precache: true}}}, nil
}

func NewMapStorageJson() (*MapStorage, error) {
	return &MapStorage{dict: make(map[string][]byte), ms: new(JSONBufMarshaler), cacheCfg: &config.CacheConfig{RatingPlans: &config.CacheParamConfig{Precache: true}}}, nil
}

func (ms *MapStorage) Close() {}

func (ms *MapStorage) Flush(ignore string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.dict = make(map[string][]byte)
	return nil
}

func (ms *MapStorage) RebuildReverseForPrefix(prefix string) error {
	// FIXME: should do transaction
	keys, err := ms.GetKeysForPrefix(prefix)
	if err != nil {
		return err
	}
	for _, key := range keys {
		ms.mu.Lock()
		delete(ms.dict, key)
		ms.mu.Unlock()
	}
	switch prefix {
	case utils.REVERSE_DESTINATION_PREFIX:
		keys, err = ms.GetKeysForPrefix(utils.DESTINATION_PREFIX)
		if err != nil {
			return err
		}
		for _, key := range keys {
			dest, err := ms.GetDestination(key[len(utils.DESTINATION_PREFIX):], false, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err := ms.SetReverseDestination(dest, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.REVERSE_ALIASES_PREFIX:
		keys, err = ms.GetKeysForPrefix(utils.ALIASES_PREFIX)
		if err != nil {
			return err
		}
		for _, key := range keys {
			al, err := ms.GetAlias(key[len(utils.ALIASES_PREFIX):], false, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err := ms.SetReverseAlias(al, utils.NonTransactional); err != nil {
				return err
			}
		}
	default:
		return utils.ErrInvalidKey
	}
	return nil
}

func (ms *MapStorage) PreloadRatingCache() error {
	if ms.cacheCfg == nil {
		return nil
	}
	if ms.cacheCfg.Destinations != nil && ms.cacheCfg.Destinations.Precache {
		if err := ms.PreloadCacheForPrefix(utils.DESTINATION_PREFIX); err != nil {
			return err
		}
	}

	if ms.cacheCfg.ReverseDestinations != nil && ms.cacheCfg.ReverseDestinations.Precache {
		if err := ms.PreloadCacheForPrefix(utils.REVERSE_DESTINATION_PREFIX); err != nil {
			return err
		}
	}

	if ms.cacheCfg.RatingPlans != nil && ms.cacheCfg.RatingPlans.Precache {
		if err := ms.PreloadCacheForPrefix(utils.RATING_PLAN_PREFIX); err != nil {
			return err
		}
	}

	if ms.cacheCfg.RatingProfiles != nil && ms.cacheCfg.RatingProfiles.Precache {
		if err := ms.PreloadCacheForPrefix(utils.RATING_PROFILE_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.Lcr != nil && ms.cacheCfg.Lcr.Precache {
		if err := ms.PreloadCacheForPrefix(utils.LCR_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.CdrStats != nil && ms.cacheCfg.CdrStats.Precache {
		if err := ms.PreloadCacheForPrefix(utils.CDR_STATS_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.Actions != nil && ms.cacheCfg.Actions.Precache {
		if err := ms.PreloadCacheForPrefix(utils.ACTION_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.ActionPlans != nil && ms.cacheCfg.ActionPlans.Precache {
		if err := ms.PreloadCacheForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.ActionTriggers != nil && ms.cacheCfg.ActionTriggers.Precache {
		if err := ms.PreloadCacheForPrefix(utils.ACTION_TRIGGER_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.SharedGroups != nil && ms.cacheCfg.SharedGroups.Precache {
		if err := ms.PreloadCacheForPrefix(utils.SHARED_GROUP_PREFIX); err != nil {
			return err
		}
	}
	// add more prefixes if needed
	return nil
}

func (ms *MapStorage) PreloadAccountingCache() error {
	if ms.cacheCfg == nil {
		return nil
	}
	if ms.cacheCfg.Aliases != nil && ms.cacheCfg.Aliases.Precache {
		if err := ms.PreloadCacheForPrefix(utils.ALIASES_PREFIX); err != nil {
			return err
		}
	}

	if ms.cacheCfg.ReverseAliases != nil && ms.cacheCfg.ReverseAliases.Precache {
		if err := ms.PreloadCacheForPrefix(utils.REVERSE_ALIASES_PREFIX); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MapStorage) PreloadCacheForPrefix(prefix string) error {
	transID := cache.BeginTransaction()
	cache.RemPrefixKey(prefix, false, transID)
	keyList, err := ms.GetKeysForPrefix(prefix)
	if err != nil {
		cache.RollbackTransaction(transID)
		return err
	}
	switch prefix {
	case utils.RATING_PLAN_PREFIX:
		for _, key := range keyList {
			_, err := ms.GetRatingPlan(key[len(utils.RATING_PLAN_PREFIX):], true, transID)
			if err != nil {
				cache.RollbackTransaction(transID)
				return err
			}
		}
	default:
		cache.RollbackTransaction(transID)
		return utils.ErrInvalidKey
	}
	cache.CommitTransaction(transID)
	return nil
}

func (ms *MapStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	keysForPrefix := make([]string, 0)
	for key := range ms.dict {
		if strings.HasPrefix(key, prefix) {
			keysForPrefix = append(keysForPrefix, key)
		}
	}
	return keysForPrefix, nil
}

// Used to check if specific subject is stored using prefix key attached to entity
func (ms *MapStorage) HasData(categ, subject string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	switch categ {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX, utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX, utils.DERIVEDCHARGERS_PREFIX:
		_, exists := ms.dict[categ+subject]
		return exists, nil
	}
	return false, errors.New("Unsupported HasData category")
}

func (ms *MapStorage) GetRatingPlan(key string, skipCache bool, transactionID string) (rp *RatingPlan, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.RATING_PLAN_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(*RatingPlan), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	cCommit := cacheCommit(transactionID)
	if values, ok := ms.dict[key]; ok {
		b := bytes.NewBuffer(values)
		r, err := zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r.Close()
		rp = new(RatingPlan)
		err = ms.ms.Unmarshal(out, rp)
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, rp, cCommit, transactionID)
	return
}

func (ms *MapStorage) SetRatingPlan(rp *RatingPlan, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	ms.dict[utils.RATING_PLAN_PREFIX+rp.Id] = b.Bytes()
	response := 0
	if historyScribe != nil {
		go historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(), &response)
	}
	cache.RemKey(utils.RATING_PLAN_PREFIX+rp.Id, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) GetRatingProfile(key string, skipCache bool, transactionID string) (rpf *RatingProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.RATING_PROFILE_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(*RatingProfile), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	cCommit := cacheCommit(transactionID)
	if values, ok := ms.dict[key]; ok {
		rpf = new(RatingProfile)
		err = ms.ms.Unmarshal(values, rpf)
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, rpf, cCommit, transactionID)
	return
}

func (ms *MapStorage) SetRatingProfile(rpf *RatingProfile, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(rpf)
	ms.dict[utils.RATING_PROFILE_PREFIX+rpf.Id] = result
	response := 0
	if historyScribe != nil {
		go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(false), &response)
	}
	cache.RemKey(utils.RATING_PROFILE_PREFIX+rpf.Id, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) RemoveRatingProfile(key string, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for k := range ms.dict {
		if strings.HasPrefix(k, key) {
			delete(ms.dict, key)
			cache.RemKey(k, cacheCommit(transactionID), transactionID)
			response := 0
			rpf := &RatingProfile{Id: key}
			if historyScribe != nil {
				go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(true), &response)
			}
		}
	}
	return
}

func (ms *MapStorage) GetLCR(key string, skipCache bool, transactionID string) (lcr *LCR, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.LCR_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(*LCR), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	cCommit := cacheCommit(transactionID)
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &lcr)
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, lcr, cCommit, transactionID)
	return
}

func (ms *MapStorage) SetLCR(lcr *LCR, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(lcr)
	ms.dict[utils.LCR_PREFIX+lcr.GetId()] = result
	cache.RemKey(utils.LCR_PREFIX+lcr.GetId(), cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) GetDestination(key string, skipCache bool, transactionID string) (dest *Destination, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	cCommit := cacheCommit(transactionID)
	key = utils.DESTINATION_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(*Destination), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	if values, ok := ms.dict[key]; ok {
		b := bytes.NewBuffer(values)
		r, err := zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r.Close()
		dest = new(Destination)
		err = ms.ms.Unmarshal(out, dest)
		if err != nil {
			cache.Set(key, dest, cCommit, transactionID)
		}
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetDestination(dest *Destination, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(dest)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	key := utils.DESTINATION_PREFIX + dest.Id
	ms.dict[key] = b.Bytes()
	response := 0
	if historyScribe != nil {
		go historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) GetReverseDestination(prefix string, skipCache bool, transactionID string) (ids []string, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	prefix = utils.REVERSE_DESTINATION_PREFIX + prefix
	if !skipCache {
		if x, ok := cache.Get(prefix); ok {
			if x != nil {
				return x.([]string), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	if idMap, ok := ms.dict.smembers(prefix, ms.ms); ok {
		ids = idMap.Slice()
	} else {
		cache.Set(prefix, nil, cacheCommit(transactionID), transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(prefix, ids, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) SetReverseDestination(dest *Destination, transactionID string) (err error) {
	for _, p := range dest.Prefixes {
		key := utils.REVERSE_DESTINATION_PREFIX + p

		ms.mu.Lock()
		ms.dict.sadd(key, dest.Id, ms.ms)
		ms.mu.Unlock()
		cache.RemKey(key, cacheCommit(transactionID), transactionID)
	}
	return
}

func (ms *MapStorage) RemoveDestination(destID string, transactionID string) (err error) {
	key := utils.DESTINATION_PREFIX + destID
	// get destination for prefix list
	d, err := ms.GetDestination(destID, false, transactionID)
	if err != nil {
		return
	}
	ms.mu.Lock()
	delete(ms.dict, key)
	ms.mu.Unlock()
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	for _, prefix := range d.Prefixes {
		ms.mu.Lock()
		ms.dict.srem(utils.REVERSE_DESTINATION_PREFIX+prefix, destID, ms.ms)
		ms.mu.Unlock()
		ms.GetReverseDestination(prefix, true, transactionID) // it will recache the destination
	}
	return
}

func (ms *MapStorage) UpdateReverseDestination(oldDest, newDest *Destination, transactionID string) error {
	//log.Printf("Old: %+v, New: %+v", oldDest, newDest)
	var obsoletePrefixes []string
	var addedPrefixes []string
	var found bool
	for _, oldPrefix := range oldDest.Prefixes {
		found = false
		for _, newPrefix := range newDest.Prefixes {
			if oldPrefix == newPrefix {
				found = true
				break
			}
		}
		if !found {
			obsoletePrefixes = append(obsoletePrefixes, oldPrefix)
		}
	}

	for _, newPrefix := range newDest.Prefixes {
		found = false
		for _, oldPrefix := range oldDest.Prefixes {
			if newPrefix == oldPrefix {
				found = true
				break
			}
		}
		if !found {
			addedPrefixes = append(addedPrefixes, newPrefix)
		}
	}
	//log.Print("Obsolete prefixes: ", obsoletePrefixes)
	//log.Print("Added prefixes: ", addedPrefixes)
	// remove id for all obsolete prefixes
	cCommit := cacheCommit(transactionID)
	var err error
	for _, obsoletePrefix := range obsoletePrefixes {
		ms.mu.Lock()
		ms.dict.srem(utils.REVERSE_DESTINATION_PREFIX+obsoletePrefix, oldDest.Id, ms.ms)
		ms.mu.Unlock()
		cache.RemKey(utils.REVERSE_DESTINATION_PREFIX+obsoletePrefix, cCommit, transactionID)
	}

	// add the id to all new prefixes
	for _, addedPrefix := range addedPrefixes {
		ms.mu.Lock()
		ms.dict.sadd(utils.REVERSE_DESTINATION_PREFIX+addedPrefix, newDest.Id, ms.ms)
		ms.mu.Unlock()
		cache.RemKey(utils.REVERSE_DESTINATION_PREFIX+addedPrefix, cCommit, transactionID)
	}
	return err
}

func (ms *MapStorage) GetActions(key string, skipCache bool, transactionID string) (as Actions, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	cCommit := cacheCommit(transactionID)
	key = utils.ACTION_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(Actions), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &as)
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, as, cCommit, transactionID)
	return
}

func (ms *MapStorage) SetActions(key string, as Actions, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	cCommit := cacheCommit(transactionID)
	result, err := ms.ms.Marshal(&as)
	ms.dict[utils.ACTION_PREFIX+key] = result
	cache.RemKey(utils.ACTION_PREFIX+key, cCommit, transactionID)
	return
}

func (ms *MapStorage) RemoveActions(key string, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.ACTION_PREFIX+key)
	cache.RemKey(utils.ACTION_PREFIX+key, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) GetSharedGroup(key string, skipCache bool, transactionID string) (sg *SharedGroup, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.SHARED_GROUP_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(*SharedGroup), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	cCommit := cacheCommit(transactionID)
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &sg)
		if err == nil {
			cache.Set(key, sg, cCommit, transactionID)
		}
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetSharedGroup(sg *SharedGroup, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sg)
	ms.dict[utils.SHARED_GROUP_PREFIX+sg.Id] = result
	cache.RemKey(utils.SHARED_GROUP_PREFIX+sg.Id, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) GetAccount(key string) (ub *Account, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.ACCOUNT_PREFIX+key]; ok {
		ub = &Account{ID: key}
		err = ms.ms.Unmarshal(values, ub)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetAccount(ub *Account) (err error) {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(ub.BalanceMap) == 0 {
		if ac, err := ms.GetAccount(ub.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = ub.ActionTriggers
			ac.UnitCounters = ub.UnitCounters
			ac.AllowNegative = ub.AllowNegative
			ac.Disabled = ub.Disabled
			ub = ac
		}
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(ub)
	ms.dict[utils.ACCOUNT_PREFIX+ub.ID] = result
	return
}

func (ms *MapStorage) RemoveAccount(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.ACCOUNT_PREFIX+key)
	return
}

func (ms *MapStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.CDR_STATS_QUEUE_PREFIX+key]; ok {
		sq = &StatsQueue{}
		err = ms.ms.Unmarshal(values, sq)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sq)
	ms.dict[utils.CDR_STATS_QUEUE_PREFIX+sq.GetId()] = result
	return
}

func (ms *MapStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result = make(map[string]*SubscriberData)
	for key, value := range ms.dict {
		if strings.HasPrefix(key, utils.PUBSUB_SUBSCRIBERS_PREFIX) {
			sub := &SubscriberData{}
			if err = ms.ms.Unmarshal(value, sub); err == nil {
				result[key[len(utils.PUBSUB_SUBSCRIBERS_PREFIX):]] = sub
			}
		}
	}
	return
}
func (ms *MapStorage) SetSubscriber(key string, sub *SubscriberData) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sub)
	ms.dict[utils.PUBSUB_SUBSCRIBERS_PREFIX+key] = result
	return
}

func (ms *MapStorage) RemoveSubscriber(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.PUBSUB_SUBSCRIBERS_PREFIX+key)
	return
}

func (ms *MapStorage) SetUser(up *UserProfile) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(up)
	if err != nil {
		return err
	}
	ms.dict[utils.USERS_PREFIX+up.GetId()] = result
	return nil
}
func (ms *MapStorage) GetUser(key string) (up *UserProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	up = &UserProfile{}
	if values, ok := ms.dict[utils.USERS_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &up)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetUsers() (result []*UserProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for key, value := range ms.dict {
		if strings.HasPrefix(key, utils.USERS_PREFIX) {
			up := &UserProfile{}
			if err = ms.ms.Unmarshal(value, up); err == nil {
				result = append(result, up)
			}
		}
	}
	return
}

func (ms *MapStorage) RemoveUser(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.USERS_PREFIX+key)
	return nil
}

func (ms *MapStorage) GetAlias(key string, skipCache bool, transactionID string) (al *Alias, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	origKey := key
	key = utils.ALIASES_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				al = &Alias{Values: x.(AliasValues)}
				al.SetId(origKey)
				return al, nil
			}
			return nil, utils.ErrNotFound
		}
	}
	cCommit := cacheCommit(transactionID)
	if values, ok := ms.dict[key]; ok {
		al = &Alias{Values: make(AliasValues, 0)}
		al.SetId(key[len(utils.ALIASES_PREFIX):])
		err = ms.ms.Unmarshal(values, &al.Values)
		if err == nil {
			cache.Set(key, al.Values, cCommit, transactionID)
		}
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	return al, nil
}

func (ms *MapStorage) SetAlias(al *Alias, transactionID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(al.Values)
	if err != nil {
		return err
	}
	key := utils.ALIASES_PREFIX + al.GetId()
	ms.dict[key] = result
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	return nil
}

func (ms *MapStorage) GetReverseAlias(reverseID string, skipCache bool, transactionID string) (ids []string, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.REVERSE_ALIASES_PREFIX + reverseID
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.([]string), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	var values []string
	cCommit := cacheCommit(transactionID)
	if idMap, ok := ms.dict.smembers(key, ms.ms); len(idMap) > 0 && ok {
		values = idMap.Slice()
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, values, cCommit, transactionID)
	return
}

func (ms *MapStorage) SetReverseAlias(al *Alias, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	for _, value := range al.Values {
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := strings.Join([]string{utils.REVERSE_ALIASES_PREFIX, alias, target, al.Context}, "")
				id := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
				ms.mu.Lock()
				ms.dict.sadd(rKey, id, ms.ms)
				ms.mu.Unlock()

				cache.RemKey(rKey, cCommit, transactionID)
			}
		}
	}
	return
}

func (ms *MapStorage) RemoveAlias(key string, transactionID string) error {
	// get alias for values list
	al, err := ms.GetAlias(key, false, transactionID)
	if err != nil {
		return err
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	key = utils.ALIASES_PREFIX + key

	aliasValues := make(AliasValues, 0)
	if values, ok := ms.dict[key]; ok {
		ms.ms.Unmarshal(values, &aliasValues)
	}
	delete(ms.dict, key)
	cCommit := cacheCommit(transactionID)
	cache.RemKey(key, cCommit, transactionID)
	for _, value := range al.Values {
		tmpKey := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := utils.REVERSE_ALIASES_PREFIX + alias + target + al.Context
				ms.dict.srem(rKey, tmpKey, ms.ms)
				cache.RemKey(rKey, cCommit, transactionID)
				/*_, err = ms.GetReverseAlias(rKey, true) // recache
				if err != nil {
					return err
				}*/
			}
		}
	}
	return nil
}

func (ms *MapStorage) UpdateReverseAlias(oldAl, newAl *Alias, transactionID string) error {
	return nil
}

func (ms *MapStorage) GetLoadHistory(limitItems int, skipCache bool, transactionID string) ([]*utils.LoadInstance, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return nil, nil
}

func (ms *MapStorage) AddLoadHistory(*utils.LoadInstance, int, string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return nil
}

func (ms *MapStorage) GetActionTriggers(key string, skipCache bool, transactionID string) (atrs ActionTriggers, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	cCommit := cacheCommit(transactionID)
	key = utils.ACTION_TRIGGER_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(ActionTriggers), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &atrs)
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, atrs, cCommit, transactionID)
	return
}

func (ms *MapStorage) SetActionTriggers(key string, atrs ActionTriggers, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(atrs) == 0 {
		// delete the key
		delete(ms.dict, utils.ACTION_TRIGGER_PREFIX+key)
		return
	}
	result, err := ms.ms.Marshal(&atrs)
	ms.dict[utils.ACTION_TRIGGER_PREFIX+key] = result
	cache.RemKey(utils.ACTION_TRIGGER_PREFIX+key, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) RemoveActionTriggers(key string, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.ACTION_TRIGGER_PREFIX+key)
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) GetActionPlan(key string, skipCache bool, transactionID string) (ats *ActionPlan, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.ACTION_PLAN_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(*ActionPlan), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &ats)
	} else {
		cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, ats, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) SetActionPlan(key string, ats *ActionPlan, overwrite bool, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if len(ats.ActionTimings) == 0 {
		ms.mu.Lock()
		defer ms.mu.Unlock()
		// delete the key
		delete(ms.dict, utils.ACTION_PLAN_PREFIX+key)
		cache.RemKey(utils.ACTION_PLAN_PREFIX+key, cCommit, transactionID)
		return
	}
	if !overwrite {
		// get existing action plan to merge the account ids
		if existingAts, _ := ms.GetActionPlan(key, true, transactionID); existingAts != nil {
			if ats.AccountIDs == nil && len(existingAts.AccountIDs) > 0 {
				ats.AccountIDs = make(utils.StringMap)
			}
			for accID := range existingAts.AccountIDs {
				ats.AccountIDs[accID] = true
			}
		}
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(&ats)
	ms.dict[utils.ACTION_PLAN_PREFIX+key] = result
	cache.RemKey(utils.ACTION_PLAN_PREFIX+key, cCommit, transactionID)
	return
}

func (ms *MapStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	keys, err := ms.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}

	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		ap, err := ms.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		ats[key[len(utils.ACTION_PLAN_PREFIX):]] = ap
	}

	return
}

func (ms *MapStorage) PushTask(t *Task) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(t)
	if err != nil {
		return err
	}
	ms.tasks = append(ms.tasks, result)
	return nil
}

func (ms *MapStorage) PopTask() (t *Task, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(ms.tasks) > 0 {
		var values []byte
		values, ms.tasks = ms.tasks[0], ms.tasks[1:]
		t = &Task{}
		err = ms.ms.Unmarshal(values, t)
	} else {
		err = utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetDerivedChargers(key string, skipCache bool, transactionID string) (dcs *utils.DerivedChargers, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	cCommit := cacheCommit(transactionID)
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(*utils.DerivedChargers), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &dcs)
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, dcs, cCommit, transactionID)
	return
}

func (ms *MapStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers, transactionID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	cCommit := cacheCommit(transactionID)
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if dcs == nil || len(dcs.Chargers) == 0 {
		delete(ms.dict, key)
		cache.RemKey(key, cCommit, transactionID)
		return nil
	}
	result, err := ms.ms.Marshal(dcs)
	ms.dict[key] = result
	cache.RemKey(key, cCommit, transactionID)
	return err
}

func (ms *MapStorage) SetCdrStats(cs *CdrStats) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(cs)
	ms.dict[utils.CDR_STATS_PREFIX+cs.Id] = result
	return err
}

func (ms *MapStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.CDR_STATS_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &cs)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetAllCdrStats() (css []*CdrStats, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for key, value := range ms.dict {
		if !strings.HasPrefix(key, utils.CDR_STATS_PREFIX) {
			continue
		}
		cs := &CdrStats{}
		err = ms.ms.Unmarshal(value, cs)
		css = append(css, cs)
	}
	return
}

func (ms *MapStorage) SetSMCost(smCost *SMCost) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(smCost)
	ms.dict[utils.LOG_CALL_COST_PREFIX+smCost.CostSource+smCost.RunID+"_"+smCost.CGRID] = result
	return err
}

func (ms *MapStorage) GetSMCost(cgrid, source, runid, originHost, originID string) (smCost *SMCost, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.LOG_CALL_COST_PREFIX+source+runid+"_"+cgrid]; ok {
		err = ms.ms.Unmarshal(values, &smCost)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetStructVersion(v *StructVersion) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	var result []byte
	result, err = ms.ms.Marshal(v)
	if err != nil {
		return
	}
	ms.dict[utils.VERSION_PREFIX+"struct"] = result
	return
}

func (ms *MapStorage) GetStructVersion() (rsv *StructVersion, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rsv = &StructVersion{}
	if values, ok := ms.dict[utils.VERSION_PREFIX+"struct"]; ok {
		err = ms.ms.Unmarshal(values, &rsv)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetResourceLimit(id string, skipCache bool, transactionID string) (*ResourceLimit, error) {
	return nil, nil
}
func (ms *MapStorage) SetResourceLimit(rl *ResourceLimit, transactionID string) error {
	return nil
}

func (ms *MapStorage) RemoveResourceLimit(id string, transactionID string) error {
	return nil
}

func (ms *MapStorage) GetReqFilterIndexes(dbKey string) (indexes map[string]map[string]utils.StringMap, err error) {
	return
}
func (ms *MapStorage) SetReqFilterIndexes(dbKey string, indexes map[string]map[string]utils.StringMap) (err error) {
	return
}
func (ms *MapStorage) MatchReqFilterIndex(dbKey, fieldValKey string) (itemIDs utils.StringMap, err error) {
	return
}
