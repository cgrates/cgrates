/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either vemsion 3 of the License, or
(at your option) any later vemsion.

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
	"fmt"
	"io/ioutil"

	"strings"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
)

type MapStorage struct {
	dict map[string][]byte
	ms   Marshaler
}

func NewMapStorage() (*MapStorage, error) {
	return &MapStorage{dict: make(map[string][]byte), ms: NewCodecMsgpackMarshaler()}, nil
}

func NewMapStorageJson() (*MapStorage, error) {
	return &MapStorage{dict: make(map[string][]byte), ms: new(JSONBufMarshaler)}, nil
}

func (ms *MapStorage) Close() {}

func (ms *MapStorage) Flush(ignore string) error {
	ms.dict = make(map[string][]byte)
	return nil
}

func (ms *MapStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	keysForPrefix := make([]string, 0)
	for key := range ms.dict {
		if strings.HasPrefix(key, prefix) {
			keysForPrefix = append(keysForPrefix, key)
		}
	}
	return keysForPrefix, nil
}

func (ms *MapStorage) CacheRatingAll() error {
	return ms.cacheRating(nil, nil, nil, nil, nil, nil, nil, nil)
}

func (ms *MapStorage) CacheRatingPrefixes(prefixes ...string) error {
	pm := map[string][]string{
		utils.DESTINATION_PREFIX:     []string{},
		utils.RATING_PLAN_PREFIX:     []string{},
		utils.RATING_PROFILE_PREFIX:  []string{},
		utils.LCR_PREFIX:             []string{},
		utils.DERIVEDCHARGERS_PREFIX: []string{},
		utils.ACTION_PREFIX:          []string{},
		utils.ACTION_PLAN_PREFIX:     []string{},
		utils.SHARED_GROUP_PREFIX:    []string{},
	}
	for _, prefix := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = nil
	}
	return ms.cacheRating(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.ACTION_PLAN_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (ms *MapStorage) CacheRatingPrefixValues(prefixes map[string][]string) error {
	pm := map[string][]string{
		utils.DESTINATION_PREFIX:     []string{},
		utils.RATING_PLAN_PREFIX:     []string{},
		utils.RATING_PROFILE_PREFIX:  []string{},
		utils.LCR_PREFIX:             []string{},
		utils.DERIVEDCHARGERS_PREFIX: []string{},
		utils.ACTION_PREFIX:          []string{},
		utils.ACTION_PLAN_PREFIX:     []string{},
		utils.SHARED_GROUP_PREFIX:    []string{},
	}
	for prefix, ids := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = ids
	}
	return ms.cacheRating(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.ACTION_PLAN_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (ms *MapStorage) cacheRating(dKeys, rpKeys, rpfKeys, lcrKeys, dcsKeys, actKeys, aplKeys, shgKeys []string) error {
	cache2go.BeginTransaction()
	if dKeys == nil || (float64(cache2go.CountEntries(utils.DESTINATION_PREFIX))*utils.DESTINATIONS_LOAD_THRESHOLD < float64(len(dKeys))) {
		cache2go.RemPrefixKey(utils.DESTINATION_PREFIX)
	} else {
		CleanStalePrefixes(dKeys)
	}
	if rpKeys == nil {
		cache2go.RemPrefixKey(utils.RATING_PLAN_PREFIX)
	}
	if rpfKeys == nil {
		cache2go.RemPrefixKey(utils.RATING_PROFILE_PREFIX)
	}
	if lcrKeys == nil {
		cache2go.RemPrefixKey(utils.LCR_PREFIX)
	}
	if dcsKeys == nil {
		cache2go.RemPrefixKey(utils.DERIVEDCHARGERS_PREFIX)
	}
	if actKeys == nil {
		cache2go.RemPrefixKey(utils.ACTION_PREFIX) // Forced until we can fine tune it
	}
	if aplKeys == nil {
		cache2go.RemPrefixKey(utils.ACTION_PLAN_PREFIX)
	}
	if shgKeys == nil {
		cache2go.RemPrefixKey(utils.SHARED_GROUP_PREFIX) // Forced until we can fine tune it
	}
	for k, _ := range ms.dict {
		if strings.HasPrefix(k, utils.DESTINATION_PREFIX) {
			if _, err := ms.GetDestination(k[len(utils.DESTINATION_PREFIX):]); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
		if strings.HasPrefix(k, utils.RATING_PLAN_PREFIX) {
			cache2go.RemKey(k)
			if _, err := ms.GetRatingPlan(k[len(utils.RATING_PLAN_PREFIX):], true); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
		if strings.HasPrefix(k, utils.RATING_PROFILE_PREFIX) {
			cache2go.RemKey(k)
			if _, err := ms.GetRatingProfile(k[len(utils.RATING_PROFILE_PREFIX):], true); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
		if strings.HasPrefix(k, utils.LCR_PREFIX) {
			cache2go.RemKey(k)
			if _, err := ms.GetLCR(k[len(utils.LCR_PREFIX):], true); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
		if strings.HasPrefix(k, utils.DERIVEDCHARGERS_PREFIX) {
			cache2go.RemKey(k)
			if _, err := ms.GetDerivedChargers(k[len(utils.DERIVEDCHARGERS_PREFIX):], true); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
		if strings.HasPrefix(k, utils.ACTION_PREFIX) {
			cache2go.RemKey(k)
			if _, err := ms.GetActions(k[len(utils.ACTION_PREFIX):], true); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
		if strings.HasPrefix(k, utils.ACTION_PLAN_PREFIX) {
			cache2go.RemKey(k)
			if _, err := ms.GetActionPlans(k[len(utils.ACTION_PLAN_PREFIX):], true); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
		if strings.HasPrefix(k, utils.SHARED_GROUP_PREFIX) {
			cache2go.RemKey(k)
			if _, err := ms.GetSharedGroup(k[len(utils.SHARED_GROUP_PREFIX):], true); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
	}
	cache2go.CommitTransaction()
	return nil
}

func (ms *MapStorage) CacheAccountingAll() error {
	return ms.cacheAccounting(nil)
}

func (ms *MapStorage) CacheAccountingPrefixes(prefixes ...string) error {
	pm := map[string][]string{
		utils.ALIASES_PREFIX: []string{},
	}
	for _, prefix := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = nil
	}
	return ms.cacheAccounting(pm[utils.ALIASES_PREFIX])
}

func (ms *MapStorage) CacheAccountingPrefixValues(prefixes map[string][]string) error {
	pm := map[string][]string{
		utils.ALIASES_PREFIX: []string{},
	}
	for prefix, ids := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = ids
	}
	return ms.cacheAccounting(pm[utils.ALIASES_PREFIX])
}

func (ms *MapStorage) cacheAccounting(alsKeys []string) error {
	cache2go.BeginTransaction()
	if alsKeys == nil {
		cache2go.RemPrefixKey(utils.ALIASES_PREFIX) // Forced until we can fine tune it
	}
	for k, _ := range ms.dict {
		if strings.HasPrefix(k, utils.ALIASES_PREFIX) {
			cache2go.RemKey(k)
			if _, err := ms.GetAlias(k[len(utils.ALIASES_PREFIX):], true); err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
	}
	cache2go.CommitTransaction()
	return nil
}

// Used to check if specific subject is stored using prefix key attached to entity
func (ms *MapStorage) HasData(categ, subject string) (bool, error) {
	switch categ {
	case utils.DESTINATION_PREFIX:
		_, exists := ms.dict[utils.DESTINATION_PREFIX+subject]
		return exists, nil
	case utils.RATING_PLAN_PREFIX:
		_, exists := ms.dict[utils.RATING_PLAN_PREFIX+subject]
		return exists, nil
	}
	return false, errors.New("Unsupported category")
}

func (ms *MapStorage) GetRatingPlan(key string, skipCache bool) (rp *RatingPlan, err error) {
	key = utils.RATING_PLAN_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*RatingPlan), nil
		} else {
			return nil, err
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
		rp = new(RatingPlan)
		err = ms.ms.Unmarshal(out, rp)
		cache2go.Cache(key, rp)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetRatingPlan(rp *RatingPlan) (err error) {
	result, err := ms.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	ms.dict[utils.RATING_PLAN_PREFIX+rp.Id] = b.Bytes()
	response := 0
	if historyScribe != nil {
		go historyScribe.Record(rp.GetHistoryRecord(), &response)
	}
	return
}

func (ms *MapStorage) GetRatingProfile(key string, skipCache bool) (rpf *RatingProfile, err error) {
	key = utils.RATING_PROFILE_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*RatingProfile), nil
		} else {
			return nil, err
		}
	}
	if values, ok := ms.dict[key]; ok {
		rpf = new(RatingProfile)

		err = ms.ms.Unmarshal(values, rpf)
		cache2go.Cache(key, rpf)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetRatingProfile(rpf *RatingProfile) (err error) {
	result, err := ms.ms.Marshal(rpf)
	ms.dict[utils.RATING_PROFILE_PREFIX+rpf.Id] = result
	response := 0
	if historyScribe != nil {
		go historyScribe.Record(rpf.GetHistoryRecord(false), &response)
	}
	return
}

func (ms *MapStorage) RemoveRatingProfile(key string) (err error) {
	for k := range ms.dict {
		if strings.HasPrefix(k, key) {
			delete(ms.dict, key)
			cache2go.RemKey(k)
			response := 0
			rpf := &RatingProfile{Id: key}
			if historyScribe != nil {
				go historyScribe.Record(rpf.GetHistoryRecord(true), &response)
			}
		}
	}
	return
}

func (ms *MapStorage) GetLCR(key string, skipCache bool) (lcr *LCR, err error) {
	key = utils.LCR_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*LCR), nil
		} else {
			return nil, err
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &lcr)
		cache2go.Cache(key, lcr)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetLCR(lcr *LCR) (err error) {
	result, err := ms.ms.Marshal(lcr)
	ms.dict[utils.LCR_PREFIX+lcr.GetId()] = result
	return
}

func (ms *MapStorage) GetDestination(key string) (dest *Destination, err error) {
	key = utils.DESTINATION_PREFIX + key
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
		// create optimized structure
		for _, p := range dest.Prefixes {
			cache2go.CachePush(utils.DESTINATION_PREFIX+p, dest.Id)
		}
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetDestination(dest *Destination) (err error) {
	result, err := ms.ms.Marshal(dest)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	ms.dict[utils.DESTINATION_PREFIX+dest.Id] = b.Bytes()
	response := 0
	if historyScribe != nil {
		go historyScribe.Record(dest.GetHistoryRecord(), &response)
	}
	return
}

func (ms *MapStorage) GetActions(key string, skipCache bool) (as Actions, err error) {
	key = utils.ACTION_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(Actions), nil
		} else {
			return nil, err
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &as)
		cache2go.Cache(key, as)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetActions(key string, as Actions) (err error) {
	result, err := ms.ms.Marshal(&as)
	ms.dict[utils.ACTION_PREFIX+key] = result
	return
}

func (ms *MapStorage) GetSharedGroup(key string, skipCache bool) (sg *SharedGroup, err error) {
	key = utils.SHARED_GROUP_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*SharedGroup), nil
		} else {
			return nil, err
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &sg)
		if err == nil {
			cache2go.Cache(key, sg)
		}
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetSharedGroup(sg *SharedGroup) (err error) {
	result, err := ms.ms.Marshal(sg)
	ms.dict[utils.SHARED_GROUP_PREFIX+sg.Id] = result
	return
}

func (ms *MapStorage) GetAccount(key string) (ub *Account, err error) {
	if values, ok := ms.dict[utils.ACCOUNT_PREFIX+key]; ok {
		ub = &Account{Id: key}
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
		if ac, err := ms.GetAccount(ub.Id); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = ub.ActionTriggers
			ac.UnitCounters = ub.UnitCounters
			ac.AllowNegative = ub.AllowNegative
			ac.Disabled = ub.Disabled
			ub = ac
		}
	}
	result, err := ms.ms.Marshal(ub)
	ms.dict[utils.ACCOUNT_PREFIX+ub.Id] = result
	return
}

func (ms *MapStorage) RemoveAccount(key string) (err error) {
	delete(ms.dict, utils.ACCOUNT_PREFIX+key)
	return
}

func (ms *MapStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	if values, ok := ms.dict[utils.CDR_STATS_QUEUE_PREFIX+key]; ok {
		sq = &StatsQueue{}
		err = ms.ms.Unmarshal(values, sq)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	result, err := ms.ms.Marshal(sq)
	ms.dict[utils.CDR_STATS_QUEUE_PREFIX+sq.GetId()] = result
	return
}

func (ms *MapStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
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
	result, err := ms.ms.Marshal(sub)
	ms.dict[utils.PUBSUB_SUBSCRIBERS_PREFIX+key] = result
	return
}

func (ms *MapStorage) RemoveSubscriber(key string) (err error) {
	delete(ms.dict, utils.PUBSUB_SUBSCRIBERS_PREFIX+key)
	return
}

func (ms *MapStorage) SetUser(up *UserProfile) error {
	result, err := ms.ms.Marshal(up)
	if err != nil {
		return err
	}
	ms.dict[utils.USERS_PREFIX+up.GetId()] = result
	return nil
}
func (ms *MapStorage) GetUser(key string) (up *UserProfile, err error) {
	up = &UserProfile{}
	if values, ok := ms.dict[utils.USERS_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &up)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetUsers() (result []*UserProfile, err error) {
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
	delete(ms.dict, utils.USERS_PREFIX+key)
	return nil
}

func (ms *MapStorage) SetAlias(al *Alias) error {
	result, err := ms.ms.Marshal(al.Values)
	if err != nil {
		return err
	}
	ms.dict[utils.ALIASES_PREFIX+al.GetId()] = result
	return nil
}

func (ms *MapStorage) GetAlias(key string, skipCache bool) (al *Alias, err error) {
	origKey := key
	key = utils.ALIASES_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			al = &Alias{Values: x.(AliasValues)}
			al.SetId(key[len(utils.ALIASES_PREFIX):])
			return al, nil
		} else {
			return nil, err
		}
	}
	if values, ok := ms.dict[key]; ok {
		al = &Alias{Values: make(AliasValues, 0)}
		al.SetId(key[len(utils.ALIASES_PREFIX):])
		err = ms.ms.Unmarshal(values, &al.Values)
		if err == nil {
			cache2go.Cache(key, al.Values)
			for _, value := range al.Values {

				for target, pairs := range value.Pairs {
					for _, alias := range pairs {
						var existingKeys map[string]bool
						rKey := utils.REVERSE_ALIASES_PREFIX + alias + target + al.Context
						if x, err := cache2go.Get(rKey); err == nil {
							existingKeys = x.(map[string]bool)
						} else {
							existingKeys = make(map[string]bool)
						}
						existingKeys[utils.ConcatenatedKey(origKey, value.DestinationId)] = true
						cache2go.Cache(rKey, existingKeys)
					}
				}
			}
		}
	} else {
		return nil, utils.ErrNotFound
	}
	return al, nil
}

func (ms *MapStorage) RemoveAlias(key string) error {
	al := &Alias{}
	al.SetId(key)
	origKey := key
	key = utils.ALIASES_PREFIX + key
	aliasValues := make(AliasValues, 0)
	if values, ok := ms.dict[key]; ok {
		ms.ms.Unmarshal(values, &aliasValues)
	}
	delete(ms.dict, key)
	for _, value := range aliasValues {
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				var existingKeys map[string]bool
				rKey := utils.REVERSE_ALIASES_PREFIX + alias + target + al.Context
				if x, err := cache2go.Get(rKey); err == nil {
					existingKeys = x.(map[string]bool)
				}
				for eKey := range existingKeys {
					if strings.HasPrefix(eKey, origKey) {
						delete(existingKeys, eKey)
					}
				}
				if len(existingKeys) == 0 {
					cache2go.RemKey(rKey)
				} else {
					cache2go.Cache(rKey, existingKeys)
				}
			}
			cache2go.RemKey(key)
		}
	}
	return nil
}

func (ms *MapStorage) GetLoadHistory(limitItems int, skipCache bool) ([]*LoadInstance, error) {
	return nil, nil
}

func (ms *MapStorage) AddLoadHistory(*LoadInstance, int) error {
	return nil
}

func (ms *MapStorage) GetActionTriggers(key string) (atrs ActionTriggers, err error) {
	if values, ok := ms.dict[utils.ACTION_TRIGGER_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &atrs)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetActionTriggers(key string, atrs ActionTriggers) (err error) {
	if len(atrs) == 0 {
		// delete the key
		delete(ms.dict, utils.ACTION_TRIGGER_PREFIX+key)
		return
	}
	result, err := ms.ms.Marshal(&atrs)
	ms.dict[utils.ACTION_TRIGGER_PREFIX+key] = result
	return
}

func (ms *MapStorage) GetActionPlans(key string, skipCache bool) (ats ActionPlans, err error) {
	key = utils.ACTION_PLAN_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(ActionPlans), nil
		} else {
			return nil, err
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &ats)
		cache2go.Cache(key, ats)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetActionPlans(key string, ats ActionPlans) (err error) {
	if len(ats) == 0 {
		// delete the key
		delete(ms.dict, utils.ACTION_PLAN_PREFIX+key)
		cache2go.RemKey(utils.ACTION_PLAN_PREFIX + key)
		return
	}
	result, err := ms.ms.Marshal(&ats)
	ms.dict[utils.ACTION_PLAN_PREFIX+key] = result
	return
}

func (ms *MapStorage) GetAllActionPlans() (ats map[string]ActionPlans, err error) {
	apls, err := cache2go.GetAllEntries(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}

	ats = make(map[string]ActionPlans, len(apls))
	for key, value := range apls {
		apl := value.Value().(ActionPlans)
		ats[key] = apl
	}

	return
}

func (ms *MapStorage) GetDerivedChargers(key string, skipCache bool) (dcs *utils.DerivedChargers, err error) {
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*utils.DerivedChargers), nil
		} else {
			return nil, err
		}
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &dcs)
		cache2go.Cache(key, dcs)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers) error {
	if dcs == nil || len(dcs.Chargers) == 0 {
		delete(ms.dict, utils.DERIVEDCHARGERS_PREFIX+key)
		cache2go.RemKey(utils.DERIVEDCHARGERS_PREFIX + key)
		return nil
	}
	result, err := ms.ms.Marshal(dcs)
	ms.dict[utils.DERIVEDCHARGERS_PREFIX+key] = result
	return err
}

func (ms *MapStorage) SetCdrStats(cs *CdrStats) error {
	result, err := ms.ms.Marshal(cs)
	ms.dict[utils.CDR_STATS_PREFIX+cs.Id] = result
	return err
}

func (ms *MapStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	if values, ok := ms.dict[utils.CDR_STATS_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &cs)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetAllCdrStats() (css []*CdrStats, err error) {
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

func (ms *MapStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) error {
	result, err := ms.ms.Marshal(cc)
	ms.dict[utils.LOG_CALL_COST_PREFIX+source+runid+"_"+cgrid] = result
	return err
}

func (ms *MapStorage) GetCallCostLog(cgrid, source, runid string) (cc *CallCost, err error) {
	if values, ok := ms.dict[utils.LOG_CALL_COST_PREFIX+source+runid+"_"+cgrid]; ok {
		err = ms.ms.Unmarshal(values, &cc)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	mat, err := ms.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := ms.ms.Marshal(&as)
	if err != nil {
		return
	}
	ms.dict[utils.LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano)] = []byte(fmt.Sprintf("%s*%s*%s", ubId, string(mat), string(mas)))
	return
}

func (ms *MapStorage) LogActionPlan(source string, at *ActionPlan, as Actions) (err error) {
	mat, err := ms.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := ms.ms.Marshal(&as)
	if err != nil {
		return
	}
	ms.dict[utils.LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano)] = []byte(fmt.Sprintf("%s*%s", string(mat), string(mas)))
	return
}
