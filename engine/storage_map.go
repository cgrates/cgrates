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
	cacheCfg config.CacheConfig
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
	return &MapStorage{dict: make(map[string][]byte), ms: NewCodecMsgpackMarshaler(),
		cacheCfg: config.CgrConfig().CacheCfg()}, nil
}

func NewMapStorageJson() (mpStorage *MapStorage, err error) {
	mpStorage, err = NewMapStorage()
	mpStorage.ms = new(JSONBufMarshaler)
	return
}

func (ms *MapStorage) Close() {}

func (ms *MapStorage) Flush(ignore string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.dict = make(map[string][]byte)
	return nil
}

func (ms *MapStorage) Marshaler() Marshaler {
	return ms.ms
}

func (ms *MapStorage) SelectDatabase(dbName string) (err error) {
	return
}

func (ms *MapStorage) RebuildReverseForPrefix(prefix string) error {
	// ToDo: should do transaction
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
	case utils.AccountActionPlansPrefix:
		return nil
	default:
		return utils.ErrInvalidKey
	}
	return nil
}

func (ms *MapStorage) RemoveReverseForPrefix(prefix string) error {
	// ToDo: should do transaction
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
			if err := ms.RemoveDestination(dest.Id, utils.NonTransactional); err != nil {
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
			if err := ms.RemoveAlias(al.GetId(), utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.AccountActionPlansPrefix:
		return nil
	default:
		return utils.ErrInvalidKey
	}
	return nil
}

func (ms *MapStorage) IsDBEmpty() (resp bool, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.dict) == 0, nil
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
func (ms *MapStorage) HasDataDrv(categ, subject string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	switch categ {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX,
		utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX, utils.DERIVEDCHARGERS_PREFIX,
		utils.ResourcesPrefix, utils.StatQueuePrefix, utils.ThresholdPrefix,
		utils.FilterPrefix, utils.SupplierProfilePrefix, utils.AttributeProfilePrefix:
		_, exists := ms.dict[categ+subject]
		return exists, nil
	}
	return false, errors.New("Unsupported HasData category")
}

func (ms *MapStorage) GetRatingPlanDrv(key string) (rp *RatingPlan, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.RATING_PLAN_PREFIX + key
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
		err = ms.ms.Unmarshal(out, &rp)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetRatingPlanDrv(rp *RatingPlan) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	ms.dict[utils.RATING_PLAN_PREFIX+rp.Id] = b.Bytes()
	return
}

func (ms *MapStorage) RemoveRatingPlanDrv(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for k := range ms.dict {
		if strings.HasPrefix(k, key) {
			delete(ms.dict, key)
		}
	}
	return
}

func (ms *MapStorage) GetRatingProfileDrv(key string) (rpf *RatingProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.RATING_PROFILE_PREFIX + key
	if values, ok := ms.dict[key]; ok {
		if err = ms.ms.Unmarshal(values, &rpf); err != nil {
			return nil, err
		}
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetRatingProfileDrv(rpf *RatingProfile) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(rpf)
	ms.dict[utils.RATING_PROFILE_PREFIX+rpf.Id] = result
	return
}

func (ms *MapStorage) RemoveRatingProfileDrv(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for k := range ms.dict {
		if strings.HasPrefix(k, key) {
			delete(ms.dict, key)
		}
	}
	return
}

func (ms *MapStorage) GetLCRDrv(id string) (lcr *LCR, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key := utils.LCR_PREFIX + id
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &lcr)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetLCRDrv(lcr *LCR) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(lcr)
	ms.dict[utils.LCR_PREFIX+lcr.GetId()] = result
	return
}

func (ms *MapStorage) RemoveLCRDrv(id, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.LCR_PREFIX+id)
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
		err = ms.ms.Unmarshal(out, &dest)
		if err != nil {
			cache.Set(key, nil, cCommit, transactionID)
			return nil, utils.ErrNotFound
		}
	}
	if dest == nil {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, dest, cCommit, transactionID)

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
	if idMap, ok := ms.dict.smembers(prefix, ms.ms); !ok {
		cache.Set(prefix, nil, cacheCommit(transactionID), transactionID)
		return nil, utils.ErrNotFound
	} else {
		ids = idMap.Slice()
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

func (ms *MapStorage) GetActionsDrv(key string) (as Actions, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	cachekey := utils.ACTION_PREFIX + key
	if values, ok := ms.dict[cachekey]; ok {
		err = ms.ms.Unmarshal(values, &as)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetActionsDrv(key string, as Actions) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	cachekey := utils.ACTION_PREFIX + key
	result, err := ms.ms.Marshal(&as)
	ms.dict[cachekey] = result
	return
}

func (ms *MapStorage) RemoveActionsDrv(key string) (err error) {
	cachekey := utils.ACTION_PREFIX + key
	ms.mu.Lock()
	delete(ms.dict, cachekey)
	ms.mu.Unlock()
	return
}

func (ms *MapStorage) GetSharedGroupDrv(key string) (sg *SharedGroup, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	cachekey := utils.SHARED_GROUP_PREFIX + key
	if values, ok := ms.dict[cachekey]; ok {
		err = ms.ms.Unmarshal(values, &sg)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (ms *MapStorage) SetSharedGroupDrv(sg *SharedGroup) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sg)
	ms.dict[utils.SHARED_GROUP_PREFIX+sg.Id] = result
	return
}

func (ms *MapStorage) RemoveSharedGroupDrv(id, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.SHARED_GROUP_PREFIX+id)
	return
}

func (ms *MapStorage) GetAccount(key string) (ub *Account, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[utils.ACCOUNT_PREFIX+key]
	if !ok {
		return nil, utils.ErrNotFound
	}
	ub = &Account{ID: key}
	err = ms.ms.Unmarshal(values, &ub)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
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

func (ms *MapStorage) GetCdrStatsQueueDrv(key string) (sq *CDRStatsQueue, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.CDR_STATS_QUEUE_PREFIX+key]; ok {
		sq = &CDRStatsQueue{}
		err = ms.ms.Unmarshal(values, sq)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetCdrStatsQueueDrv(sq *CDRStatsQueue) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sq)
	ms.dict[utils.CDR_STATS_QUEUE_PREFIX+sq.GetId()] = result
	return
}

func (ms *MapStorage) RemoveCdrStatsQueueDrv(id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.dict, id)
	return
}

func (ms *MapStorage) GetSubscribersDrv() (result map[string]*SubscriberData, err error) {
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
func (ms *MapStorage) SetSubscriberDrv(key string, sub *SubscriberData) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sub)
	ms.dict[utils.PUBSUB_SUBSCRIBERS_PREFIX+key] = result
	return
}

func (ms *MapStorage) RemoveSubscriberDrv(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.PUBSUB_SUBSCRIBERS_PREFIX+key)
	return
}

func (ms *MapStorage) SetUserDrv(up *UserProfile) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(up)
	if err != nil {
		return err
	}
	ms.dict[utils.USERS_PREFIX+up.GetId()] = result
	return nil
}
func (ms *MapStorage) GetUserDrv(key string) (up *UserProfile, err error) {
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

func (ms *MapStorage) GetUsersDrv() (result []*UserProfile, err error) {
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

func (ms *MapStorage) RemoveUserDrv(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.USERS_PREFIX+key)
	return nil
}

func (ms *MapStorage) GetAlias(key string, skipCache bool, transactionID string) (al *Alias, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	cacheKey := utils.ALIASES_PREFIX + key
	cCommit := cacheCommit(transactionID)
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Alias), nil
		}
	}
	if values, ok := ms.dict[cacheKey]; ok {
		al = &Alias{Values: make(AliasValues, 0)}
		al.SetId(key)
		if err = ms.ms.Unmarshal(values, &al.Values); err != nil {
			return nil, err
		}
	} else {
		cache.Set(cacheKey, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(cacheKey, al, cCommit, transactionID)
	return

}

func (ms *MapStorage) SetAlias(al *Alias, transactionID string) error {

	result, err := ms.ms.Marshal(al.Values)
	if err != nil {
		return err
	}
	key := utils.ALIASES_PREFIX + al.GetId()
	ms.mu.Lock()
	defer ms.mu.Unlock()
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

	cCommit := cacheCommit(transactionID)
	if idMap, ok := ms.dict.smembers(key, ms.ms); len(idMap) > 0 && ok {
		ids = idMap.Slice()
	} else {
		cache.Set(key, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	cache.Set(key, ids, cCommit, transactionID)
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
	al, err := ms.GetAlias(key, false, utils.NonTransactional)
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

func (ms *MapStorage) GetActionTriggersDrv(key string) (atrs ActionTriggers, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.ACTION_TRIGGER_PREFIX + key
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &atrs)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetActionTriggersDrv(key string, atrs ActionTriggers) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(atrs) == 0 {
		// delete the key
		delete(ms.dict, utils.ACTION_TRIGGER_PREFIX+key)
		return
	}
	result, err := ms.ms.Marshal(&atrs)
	ms.dict[utils.ACTION_TRIGGER_PREFIX+key] = result
	return
}

func (ms *MapStorage) RemoveActionTriggersDrv(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.ACTION_TRIGGER_PREFIX+key)
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

func (ms *MapStorage) RemoveActionPlan(key string, transactionID string) error {
	cCommit := cacheCommit(transactionID)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.ACTION_PLAN_PREFIX+key)
	cache.RemKey(utils.ACTION_PLAN_PREFIX+key, cCommit, transactionID)
	return nil
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

func (ms *MapStorage) GetAccountActionPlans(acntID string, skipCache bool, transactionID string) (apIDs []string, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key := utils.AccountActionPlansPrefix + acntID
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	values, ok := ms.dict[key]
	if !ok {
		cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		err = utils.ErrNotFound
		return nil, err
	}
	if err = ms.ms.Unmarshal(values, &apIDs); err != nil {
		return nil, err
	}
	cache.Set(key, apIDs, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MapStorage) SetAccountActionPlans(acntID string, apIDs []string, overwrite bool) (err error) {
	if !overwrite {
		if oldaPlIDs, err := ms.GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return err
		} else {
			for _, oldAPid := range oldaPlIDs {
				if !utils.IsSliceMember(apIDs, oldAPid) {
					apIDs = append(apIDs, oldAPid)
				}
			}
		}
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(apIDs)
	if err != nil {
		return err
	}
	ms.dict[utils.AccountActionPlansPrefix+acntID] = result

	return
}

func (ms *MapStorage) RemAccountActionPlans(acntID string, apIDs []string) (err error) {
	key := utils.AccountActionPlansPrefix + acntID
	if len(apIDs) == 0 {
		delete(ms.dict, key)
		return
	}
	oldaPlIDs, err := ms.GetAccountActionPlans(acntID, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	for i := 0; i < len(oldaPlIDs); {
		if utils.IsSliceMember(apIDs, oldaPlIDs[i]) {
			oldaPlIDs = append(oldaPlIDs[:i], oldaPlIDs[i+1:]...)
			continue
		}
		i++
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(oldaPlIDs) == 0 {
		delete(ms.dict, key)
		return
	}
	var result []byte
	if result, err = ms.ms.Marshal(oldaPlIDs); err != nil {
		return err
	}
	ms.dict[key] = result
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

func (ms *MapStorage) GetDerivedChargersDrv(key string) (dcs *utils.DerivedChargers, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &dcs)
	} else {
		return nil, utils.ErrNotFound
	}
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

func (ms *MapStorage) RemoveDerivedChargersDrv(id, transactionID string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	cCommit := cacheCommit(transactionID)
	delete(ms.dict, id)
	cache.RemKey(id, cCommit, transactionID)
	return
}

func (ms *MapStorage) SetCdrStatsDrv(cs *CdrStats) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(cs)
	ms.dict[utils.CDR_STATS_PREFIX+cs.Id] = result
	return err
}

func (ms *MapStorage) GetCdrStatsDrv(key string) (cs *CdrStats, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.CDR_STATS_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &cs)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetAllCdrStatsDrv() (css []*CdrStats, err error) {
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

func (ms *MapStorage) GetResourceProfileDrv(tenant, id string) (rsp *ResourceProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key := utils.ResourceProfilesPrefix + utils.ConcatenatedKey(tenant, id)
	values, ok := ms.dict[key]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &rsp)
	if err != nil {
		return nil, err
	}
	return
}

func (ms *MapStorage) SetResourceProfileDrv(r *ResourceProfile) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(r)
	if err != nil {
		return err
	}
	ms.dict[utils.ResourceProfilesPrefix+r.TenantID()] = result
	return nil
}

func (ms *MapStorage) RemoveResourceProfileDrv(tenant, id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.ResourceProfilesPrefix + utils.ConcatenatedKey(tenant, id)
	delete(ms.dict, key)
	return nil
}

func (ms *MapStorage) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key := utils.ResourcesPrefix + utils.ConcatenatedKey(tenant, id)
	values, ok := ms.dict[key]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &r)
	if err != nil {
		return nil, err
	}
	return
}

func (ms *MapStorage) SetResourceDrv(r *Resource) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(r)
	if err != nil {
		return err
	}
	ms.dict[utils.ResourcesPrefix+r.TenantID()] = result
	return
}

func (ms *MapStorage) RemoveResourceDrv(tenant, id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.ResourcesPrefix + utils.ConcatenatedKey(tenant, id)
	delete(ms.dict, key)
	return
}

func (ms *MapStorage) GetTimingDrv(id string) (t *utils.TPTiming, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key := utils.TimingsPrefix + id
	if values, ok := ms.dict[key]; ok {
		if err = ms.ms.Unmarshal(values, &t); err != nil {
			return nil, err
		}
	} else {
		return nil, utils.ErrNotFound
	}
	return

}

func (ms *MapStorage) SetTimingDrv(t *utils.TPTiming) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(t)
	if err != nil {
		return err
	}
	key := utils.TimingsPrefix + t.ID
	ms.dict[key] = result
	return nil
}

func (ms *MapStorage) RemoveTimingDrv(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.TimingsPrefix + id
	delete(ms.dict, key)
	return nil
}

//GetFilterIndexesDrv retrieves Indexes from dataDB
func (ms *MapStorage) GetFilterIndexesDrv(dbKey, filterType string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[dbKey]
	if !ok {
		return nil, utils.ErrNotFound
	}
	if len(fldNameVal) != 0 {
		rcvidx := make(map[string]utils.StringMap)
		err = ms.ms.Unmarshal(values, &rcvidx)
		if err != nil {
			return nil, err
		}
		indexes = make(map[string]utils.StringMap)
		for fldName, fldVal := range fldNameVal {
			if _, has := indexes[utils.ConcatenatedKey(filterType, fldName, fldVal)]; !has {
				indexes[utils.ConcatenatedKey(filterType, fldName, fldVal)] = make(utils.StringMap)
			}
			if len(rcvidx[utils.ConcatenatedKey(filterType, fldName, fldVal)]) != 0 {
				indexes[utils.ConcatenatedKey(filterType, fldName, fldVal)] = rcvidx[utils.ConcatenatedKey(fldName, fldVal)]
			}
		}

		return
	} else {
		err = ms.ms.Unmarshal(values, &indexes)
		if err != nil {
			return nil, err
		}
	}
	return
}

//SetFilterIndexesDrv stores Indexes into DataDB
func (ms *MapStorage) SetFilterIndexesDrv(dbKey string, indexes map[string]utils.StringMap) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(indexes)
	if err != nil {
		return err
	}
	ms.dict[dbKey] = result
	return
}

func (ms *MapStorage) RemoveFilterIndexesDrv(id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, id)
	return
}

//GetFilterReverseIndexesDrv retrieves ReverseIndexes from dataDB
func (ms *MapStorage) GetFilterReverseIndexesDrv(dbKey string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[dbKey]
	if !ok {
		return nil, utils.ErrNotFound
	}
	if len(fldNameVal) != 0 {
		rcvidx := make(map[string]utils.StringMap)
		err = ms.ms.Unmarshal(values, &rcvidx)
		if err != nil {
			return nil, err
		}
		indexes = make(map[string]utils.StringMap)
		for fldName, fldVal := range fldNameVal {
			if _, has := indexes[utils.ConcatenatedKey(fldName, fldVal)]; !has {
				indexes[utils.ConcatenatedKey(fldName, fldVal)] = make(utils.StringMap)
			}
			indexes[utils.ConcatenatedKey(fldName, fldVal)] = rcvidx[utils.ConcatenatedKey(fldName, fldVal)]
		}
		return
	} else {
		err = ms.ms.Unmarshal(values, &indexes)
		if err != nil {
			return nil, err
		}
	}
	return
}

//SetFilterReverseIndexesDrv stores ReverseIndexes into DataDB
func (ms *MapStorage) SetFilterReverseIndexesDrv(dbKey string, indexes map[string]utils.StringMap) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(indexes)
	if err != nil {
		return err
	}
	ms.dict[dbKey] = result
	return
}

//RemoveFilterReverseIndexesDrv removes ReverseIndexes for a specific itemID
func (ms *MapStorage) RemoveFilterReverseIndexesDrv(dbKey string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, dbKey)
	return
}

func (ms *MapStorage) MatchFilterIndexDrv(dbKey, filterType, fldName, fldVal string) (itemIDs utils.StringMap, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[dbKey]
	if !ok {
		return nil, utils.ErrNotFound
	}
	var indexes map[string]utils.StringMap
	if err = ms.ms.Unmarshal(values, &indexes); err != nil {
		return nil, err
	}
	if _, hasIt := indexes[utils.ConcatenatedKey(filterType, fldName, fldVal)]; hasIt {
		itemIDs = indexes[utils.ConcatenatedKey(filterType, fldName, fldVal)]
	}
	if len(itemIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

// GetStatQueueProfile retrieves a StatQueueProfile from dataDB
func (ms *MapStorage) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	key := utils.StatQueueProfilePrefix + utils.ConcatenatedKey(tenant, id)
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[key]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &sq)
	if err != nil {
		return nil, err
	}
	return
}

// SetStatsQueueDrv stores a StatsQueue into DataDB
func (ms *MapStorage) SetStatQueueProfileDrv(sqp *StatQueueProfile) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sqp)
	if err != nil {
		return err
	}
	ms.dict[utils.StatQueueProfilePrefix+utils.ConcatenatedKey(sqp.Tenant, sqp.ID)] = result
	return
}

// RemStatsQueueDrv removes a StatsQueue from dataDB
func (ms *MapStorage) RemStatQueueProfileDrv(tenant, id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.StatQueueProfilePrefix + utils.ConcatenatedKey(tenant, id)
	delete(ms.dict, key)
	return
}

// GetStatQueue retrieves the stored metrics for a StatsQueue
func (ms *MapStorage) GetStoredStatQueueDrv(tenant, id string) (sq *StoredStatQueue, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id)]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &sq)
	return
}

// SetStatQueue stores the metrics for a StatsQueue
func (ms *MapStorage) SetStoredStatQueueDrv(sq *StoredStatQueue) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	var result []byte
	result, err = ms.ms.Marshal(sq)
	if err != nil {
		return err
	}
	ms.dict[utils.StatQueuePrefix+sq.SqID()] = result
	return
}

// RemStatQueue removes a StatsQueue
func (ms *MapStorage) RemStoredStatQueueDrv(tenant, id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id))
	return
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (ms *MapStorage) GetThresholdProfileDrv(tenant, ID string) (tp *ThresholdProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, ID)
	values, ok := ms.dict[key]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &tp)
	if err != nil {
		return nil, err
	}
	return
}

// SetThresholdProfileDrv stores a ThresholdProfile into DataDB
func (ms *MapStorage) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(tp)
	if err != nil {
		return err
	}
	ms.dict[utils.ThresholdProfilePrefix+tp.TenantID()] = result
	return
}

// RemThresholdProfile removes a ThresholdProfile from dataDB/cache
func (ms *MapStorage) RemThresholdProfileDrv(tenant, id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, id)
	delete(ms.dict, key)
	return
}

func (ms *MapStorage) GetThresholdDrv(tenant, id string) (r *Threshold, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	values, ok := ms.dict[key]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, r)
	if err != nil {
		return nil, err
	}
	return
}

func (ms *MapStorage) SetThresholdDrv(r *Threshold) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(r)
	if err != nil {
		return err
	}
	ms.dict[utils.ThresholdPrefix+utils.ConcatenatedKey(r.Tenant, r.ID)] = result
	return
}

func (ms *MapStorage) RemoveThresholdDrv(tenant, id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	delete(ms.dict, key)
	return
}

func (ms *MapStorage) GetFilterDrv(tenant, id string) (r *Filter, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[utils.FilterPrefix+utils.ConcatenatedKey(tenant, id)]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &r)
	if err != nil {
		return nil, err
	}
	for _, fltr := range r.RequestFilters {
		if err := fltr.CompileValues(); err != nil {
			return nil, err
		}
	}
	return
}

func (ms *MapStorage) SetFilterDrv(r *Filter) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(r)
	if err != nil {
		return err
	}
	ms.dict[utils.FilterPrefix+utils.ConcatenatedKey(r.Tenant, r.ID)] = result
	return
}

func (ms *MapStorage) RemoveFilterDrv(tenant, id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	delete(ms.dict, key)
	return
}

func (ms *MapStorage) GetSupplierProfileDrv(tenant, id string) (r *SupplierProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[utils.SupplierProfilePrefix+utils.ConcatenatedKey(tenant, id)]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &r)
	if err != nil {
		return nil, err
	}
	return
}

func (ms *MapStorage) SetSupplierProfileDrv(r *SupplierProfile) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(r)
	if err != nil {
		return err
	}
	ms.dict[utils.SupplierProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID)] = result
	return
}

func (ms *MapStorage) RemoveSupplierProfileDrv(tenant, id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.SupplierProfilePrefix + utils.ConcatenatedKey(tenant, id)
	delete(ms.dict, key)
	return
}

func (ms *MapStorage) GetAttributeProfileDrv(tenant, id string) (r *AttributeProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	values, ok := ms.dict[utils.AttributeProfilePrefix+utils.ConcatenatedKey(tenant, id)]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &r)
	if err != nil {
		return nil, err
	}
	return
}

func (ms *MapStorage) SetAttributeProfileDrv(r *AttributeProfile) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(r)
	if err != nil {
		return err
	}
	ms.dict[utils.AttributeProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID)] = result
	return
}

func (ms *MapStorage) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(tenant, id)
	delete(ms.dict, key)
	return
}

func (ms *MapStorage) GetVersions(itm string) (vrs Versions, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	values, ok := ms.dict[itm]
	if !ok {
		return nil, utils.ErrNotFound
	}
	err = ms.ms.Unmarshal(values, &vrs)
	if err != nil {
		return nil, err
	}
	return
}

func (ms *MapStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	var result []byte
	var x Versions
	if !overwrite {
		x, err = ms.GetVersions(utils.TBLVersions)
		if err != nil {
			return err
		}
		for key, _ := range vrs {
			if x[key] != vrs[key] {
				x[key] = vrs[key]
			}
		}
		result, err = ms.ms.Marshal(x)
		if err != nil {
			return err
		}
		ms.dict[utils.TBLVersions] = result
		return
	} else {
		result, err = ms.ms.Marshal(vrs)
		if err != nil {
			return err
		}
		if ms.RemoveVersions(vrs); err != nil {
			return err
		}
		ms.dict[utils.TBLVersions] = result
		return
	}
}

func (ms *MapStorage) RemoveVersions(vrs Versions) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.TBLVersions)
	return
}

func (ms *MapStorage) GetStorageType() string {
	return utils.MAPSTOR
}
