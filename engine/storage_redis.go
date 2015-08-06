/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"fmt"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
	"github.com/hoisie/redis"

	"io/ioutil"
	"time"
)

type RedisStorage struct {
	dbNb int
	db   *redis.Client
	ms   Marshaler
}

func NewRedisStorage(address string, db int, pass, mrshlerStr string) (*RedisStorage, error) {
	ndb := &redis.Client{Addr: address, Db: db, Password: pass}

	var mrshler Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return &RedisStorage{db: ndb, dbNb: db, ms: mrshler}, nil
}

func (rs *RedisStorage) Close() {
	// no close for me
	//rs.db.Quit()
}

func (rs *RedisStorage) Flush(ignore string) (err error) {
	err = rs.db.Flush(false)
	return
}

func (rs *RedisStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	return rs.db.Keys(prefix + "*")
}

func (rs *RedisStorage) CacheAll() error {
	return rs.Cache(nil, nil, nil, nil, nil, nil, nil)
}

func (rs *RedisStorage) CachePrefixes(prefixes ...string) error {
	pm := map[string][]string{
		utils.DESTINATION_PREFIX:     []string{},
		utils.RATING_PLAN_PREFIX:     []string{},
		utils.RATING_PROFILE_PREFIX:  []string{},
		utils.LCR_PREFIX:             []string{},
		utils.DERIVEDCHARGERS_PREFIX: []string{},
		utils.ACTION_PREFIX:          []string{},
		utils.SHARED_GROUP_PREFIX:    []string{},
	}
	for _, prefix := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = nil
	}
	return rs.Cache(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (rs *RedisStorage) CachePrefixValues(prefixes map[string][]string) error {
	pm := map[string][]string{
		utils.DESTINATION_PREFIX:     []string{},
		utils.RATING_PLAN_PREFIX:     []string{},
		utils.RATING_PROFILE_PREFIX:  []string{},
		utils.LCR_PREFIX:             []string{},
		utils.DERIVEDCHARGERS_PREFIX: []string{},
		utils.ACTION_PREFIX:          []string{},
		utils.SHARED_GROUP_PREFIX:    []string{},
	}
	for prefix, ids := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = ids
	}
	return rs.Cache(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (rs *RedisStorage) Cache(dKeys, rpKeys, rpfKeys, lcrKeys, dcsKeys, actKeys, shgKeys []string) (err error) {
	cache2go.BeginTransaction()
	if dKeys == nil || (float64(cache2go.CountEntries(utils.DESTINATION_PREFIX))*utils.DESTINATIONS_LOAD_THRESHOLD < float64(len(dKeys))) {
		// if need to load more than a half of exiting keys load them all
		Logger.Info("Caching all destinations")
		if dKeys, err = rs.db.Keys(utils.DESTINATION_PREFIX + "*"); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.DESTINATION_PREFIX)
	} else if len(dKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching destinations: %v", dKeys))
		CleanStalePrefixes(dKeys)
	}
	for _, key := range dKeys {
		if len(key) <= len(utils.DESTINATION_PREFIX) {
			Logger.Warning(fmt.Sprintf("Got malformed destination id: %s", key))
			continue
		}
		if _, err = rs.GetDestination(key[len(utils.DESTINATION_PREFIX):]); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(dKeys) != 0 {
		Logger.Info("Finished destinations caching.")
	}
	if rpKeys == nil {
		Logger.Info("Caching all rating plans")
		if rpKeys, err = rs.db.Keys(utils.RATING_PLAN_PREFIX + "*"); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.RATING_PLAN_PREFIX)
	} else if len(rpKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching rating plans: %v", rpKeys))
	}
	for _, key := range rpKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetRatingPlan(key[len(utils.RATING_PLAN_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(rpKeys) != 0 {
		Logger.Info("Finished rating plans caching.")
	}
	if rpfKeys == nil {
		Logger.Info("Caching all rating profiles")
		if rpfKeys, err = rs.db.Keys(utils.RATING_PROFILE_PREFIX + "*"); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.RATING_PROFILE_PREFIX)
	} else if len(rpfKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching rating profile: %v", rpfKeys))
	}
	for _, key := range rpfKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetRatingProfile(key[len(utils.RATING_PROFILE_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(rpfKeys) != 0 {
		Logger.Info("Finished rating profile caching.")
	}
	if lcrKeys == nil {
		Logger.Info("Caching LCR rules.")
		if lcrKeys, err = rs.db.Keys(utils.LCR_PREFIX + "*"); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.LCR_PREFIX)
	} else if len(lcrKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching LCR rules: %v", lcrKeys))
	}
	for _, key := range lcrKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetLCR(key[len(utils.LCR_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(lcrKeys) != 0 {
		Logger.Info("Finished LCR rules caching.")
	}
	// DerivedChargers caching
	if dcsKeys == nil {
		Logger.Info("Caching all derived chargers")
		if dcsKeys, err = rs.db.Keys(utils.DERIVEDCHARGERS_PREFIX + "*"); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.DERIVEDCHARGERS_PREFIX)
	} else if len(dcsKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching derived chargers: %v", dcsKeys))
	}
	for _, key := range dcsKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetDerivedChargers(key[len(utils.DERIVEDCHARGERS_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(dcsKeys) != 0 {
		Logger.Info("Finished derived chargers caching.")
	}
	if actKeys == nil {
		cache2go.RemPrefixKey(utils.ACTION_PREFIX)
	}
	if actKeys == nil {
		Logger.Info("Caching all actions")
		if actKeys, err = rs.db.Keys(utils.ACTION_PREFIX + "*"); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	} else if len(actKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching actions: %v", actKeys))
	}
	for _, key := range actKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetActions(key[len(utils.ACTION_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(actKeys) != 0 {
		Logger.Info("Finished actions caching.")
	}
	if shgKeys == nil {
		cache2go.RemPrefixKey(utils.SHARED_GROUP_PREFIX)
	}
	if shgKeys == nil {
		Logger.Info("Caching all shared groups")
		if shgKeys, err = rs.db.Keys(utils.SHARED_GROUP_PREFIX + "*"); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	} else if len(shgKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching shared groups: %v", shgKeys))
	}
	for _, key := range shgKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetSharedGroup(key[len(utils.SHARED_GROUP_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(shgKeys) != 0 {
		Logger.Info("Finished shared groups caching.")
	}
	cache2go.CommitTransaction()
	return nil
}

// Used to check if specific subject is stored using prefix key attached to entity
func (rs *RedisStorage) HasData(category, subject string) (bool, error) {
	switch category {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX, utils.ACTION_PREFIX, utils.ACTION_TIMING_PREFIX, utils.ACCOUNT_PREFIX:
		return rs.db.Exists(category + subject)
	}
	return false, errors.New("Unsupported category in HasData")
}

func (rs *RedisStorage) GetRatingPlan(key string, skipCache bool) (rp *RatingPlan, err error) {
	key = utils.RATING_PLAN_PREFIX + key
	if !skipCache {
		if x, err := cache2go.GetCached(key); err == nil {
			return x.(*RatingPlan), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
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
		err = rs.ms.Unmarshal(out, rp)
		cache2go.Cache(key, rp)
	}
	return
}

func (rs *RedisStorage) SetRatingPlan(rp *RatingPlan) (err error) {
	result, err := rs.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.db.Set(utils.RATING_PLAN_PREFIX+rp.Id, b.Bytes())
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(rp.GetHistoryRecord(), &response)
	}
	return
}

func (rs *RedisStorage) GetRatingProfile(key string, skipCache bool) (rpf *RatingProfile, err error) {

	key = utils.RATING_PROFILE_PREFIX + key
	if !skipCache {
		if x, err := cache2go.GetCached(key); err == nil {
			return x.(*RatingProfile), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
		rpf = new(RatingProfile)
		err = rs.ms.Unmarshal(values, rpf)
		cache2go.Cache(key, rpf)
	}
	return
}

func (rs *RedisStorage) SetRatingProfile(rpf *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rpf)
	err = rs.db.Set(utils.RATING_PROFILE_PREFIX+rpf.Id, result)
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(rpf.GetHistoryRecord(), &response)
	}
	return
}

func (rs *RedisStorage) GetLCR(key string, skipCache bool) (lcr *LCR, err error) {
	key = utils.LCR_PREFIX + key
	if !skipCache {
		if x, err := cache2go.GetCached(key); err == nil {
			return x.(*LCR), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
		err = rs.ms.Unmarshal(values, &lcr)
		cache2go.Cache(key, lcr)
	}
	return
}

func (rs *RedisStorage) SetLCR(lcr *LCR) (err error) {
	result, err := rs.ms.Marshal(lcr)
	err = rs.db.Set(utils.LCR_PREFIX+lcr.GetId(), result)
	cache2go.Cache(utils.LCR_PREFIX+lcr.GetId(), lcr)
	return
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	key = utils.DESTINATION_PREFIX + key
	var values []byte
	if values, err = rs.db.Get(key); len(values) > 0 && err == nil {
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
		err = rs.ms.Unmarshal(out, dest)
		// create optimized structure
		for _, p := range dest.Prefixes {
			cache2go.CachePush(utils.DESTINATION_PREFIX+p, dest.Id)
		}
	} else {
		return nil, errors.New("not found")
	}
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination) (err error) {
	result, err := rs.ms.Marshal(dest)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.db.Set(utils.DESTINATION_PREFIX+dest.Id, b.Bytes())
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(dest.GetHistoryRecord(), &response)
	}
	return
}

func (rs *RedisStorage) GetActions(key string, skipCache bool) (as Actions, err error) {
	key = utils.ACTION_PREFIX + key
	if !skipCache {
		if x, err := cache2go.GetCached(key); err == nil {
			return x.(Actions), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
		err = rs.ms.Unmarshal(values, &as)
		cache2go.Cache(key, as)
	}
	return
}

func (rs *RedisStorage) SetActions(key string, as Actions) (err error) {
	result, err := rs.ms.Marshal(&as)
	err = rs.db.Set(utils.ACTION_PREFIX+key, result)
	return
}

func (rs *RedisStorage) GetSharedGroup(key string, skipCache bool) (sg *SharedGroup, err error) {
	key = utils.SHARED_GROUP_PREFIX + key
	if !skipCache {
		if x, err := cache2go.GetCached(key); err == nil {
			return x.(*SharedGroup), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
		err = rs.ms.Unmarshal(values, &sg)
		cache2go.Cache(key, sg)
	}
	return
}

func (rs *RedisStorage) SetSharedGroup(sg *SharedGroup) (err error) {
	result, err := rs.ms.Marshal(sg)
	err = rs.db.Set(utils.SHARED_GROUP_PREFIX+sg.Id, result)
	return
}

func (rs *RedisStorage) GetAccount(key string) (ub *Account, err error) {
	var values []byte
	if values, err = rs.db.Get(utils.ACCOUNT_PREFIX + key); err == nil {
		ub = &Account{Id: key}
		err = rs.ms.Unmarshal(values, ub)
	}

	return
}

func (rs *RedisStorage) SetAccount(ub *Account) (err error) {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(ub.BalanceMap) == 0 {
		if ac, err := rs.GetAccount(ub.Id); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = ub.ActionTriggers
			ac.UnitCounters = ub.UnitCounters
			ac.AllowNegative = ub.AllowNegative
			ac.Disabled = ub.Disabled
			ub = ac
		}
	}
	result, err := rs.ms.Marshal(ub)
	err = rs.db.Set(utils.ACCOUNT_PREFIX+ub.Id, result)
	return
}

func (rs *RedisStorage) RemoveAccount(key string) (err error) {
	_, err = rs.db.Del(utils.ACCOUNT_PREFIX + key)
	return

}

func (rs *RedisStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	var values []byte
	if values, err = rs.db.Get(utils.CDR_STATS_QUEUE_PREFIX + key); err == nil {
		sq = &StatsQueue{}
		err = rs.ms.Unmarshal(values, &sq)
	}
	return
}

func (rs *RedisStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	result, err := rs.ms.Marshal(sq)
	err = rs.db.Set(utils.CDR_STATS_QUEUE_PREFIX+sq.GetId(), result)
	return
}

func (rs *RedisStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	keys, err := rs.db.Keys(utils.PUBSUB_SUBSCRIBERS_PREFIX + "*")
	if err != nil {
		return nil, err
	}
	result = make(map[string]*SubscriberData)
	for _, key := range keys {
		if values, err := rs.db.Get(key); err == nil {
			sub := &SubscriberData{}
			err = rs.ms.Unmarshal(values, sub)
			result[key[len(utils.PUBSUB_SUBSCRIBERS_PREFIX):]] = sub
		} else {
			return nil, utils.ErrNotFound
		}
	}
	return
}

func (rs *RedisStorage) SetSubscriber(key string, sub *SubscriberData) (err error) {
	result, err := rs.ms.Marshal(sub)
	rs.db.Set(utils.PUBSUB_SUBSCRIBERS_PREFIX+key, result)
	return
}

func (rs *RedisStorage) RemoveSubscriber(key string) (err error) {
	_, err = rs.db.Del(utils.PUBSUB_SUBSCRIBERS_PREFIX + key)
	return
}

func (rs *RedisStorage) SetUser(up *UserProfile) (err error) {
	result, err := rs.ms.Marshal(up)
	rs.db.Set(utils.USERS_PREFIX+up.GetId(), result)
	return
}

func (rs *RedisStorage) GetUser(key string) (up *UserProfile, err error) {
	var values []byte
	if values, err = rs.db.Get(utils.USERS_PREFIX + key); err == nil {
		up = &UserProfile{}
		err = rs.ms.Unmarshal(values, &up)
	}
	return
}

func (rs *RedisStorage) GetUsers() (result []*UserProfile, err error) {
	keys, err := rs.db.Keys(utils.USERS_PREFIX + "*")
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if values, err := rs.db.Get(key); err == nil {
			up := &UserProfile{}
			err = rs.ms.Unmarshal(values, up)
			result = append(result, up)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	return
}

func (rs *RedisStorage) RemoveUser(key string) (err error) {
	_, err = rs.db.Del(utils.USERS_PREFIX + key)
	return
}

func (rs *RedisStorage) SetAlias(al *Alias) (err error) {
	result, err := rs.ms.Marshal(al.Values)
	rs.db.Set(utils.ALIASES_PREFIX+al.GetId(), result)
	return
}

func (rs *RedisStorage) GetAlias(key string) (al *Alias, err error) {
	var values []byte
	if values, err = rs.db.Get(utils.ALIASES_PREFIX + key); err == nil {
		al = &Alias{Values: make(AliasValues, 0)}
		al.SetId(key)
		err = rs.ms.Unmarshal(values, al.Values)
	}
	return
}

func (rs *RedisStorage) GetAliases() (result []*Alias, err error) {
	keys, err := rs.db.Keys(utils.ALIASES_PREFIX + "*")
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if values, err := rs.db.Get(key); err == nil {
			al := &Alias{Values: make(AliasValues, 0)}
			err = rs.ms.Unmarshal(values, al.Values)
			result = append(result, al)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	return
}

func (rs *RedisStorage) RemoveAlias(key string) (err error) {
	_, err = rs.db.Del(utils.ALIASES_PREFIX + key)
	return
}

func (rs *RedisStorage) GetActionPlans(key string) (ats ActionPlans, err error) {
	var values []byte
	if values, err = rs.db.Get(utils.ACTION_TIMING_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal(values, &ats)
	}
	return
}

func (rs *RedisStorage) SetActionPlans(key string, ats ActionPlans) (err error) {
	if len(ats) == 0 {
		// delete the key
		_, err = rs.db.Del(utils.ACTION_TIMING_PREFIX + key)
		return err
	}
	result, err := rs.ms.Marshal(&ats)
	err = rs.db.Set(utils.ACTION_TIMING_PREFIX+key, result)
	return
}

func (rs *RedisStorage) GetAllActionPlans() (ats map[string]ActionPlans, err error) {
	keys, err := rs.db.Keys(utils.ACTION_TIMING_PREFIX + "*")
	if err != nil {
		return nil, err
	}
	ats = make(map[string]ActionPlans, len(keys))
	for _, key := range keys {
		values, err := rs.db.Get(key)
		if err != nil {
			continue
		}
		var tempAts ActionPlans
		err = rs.ms.Unmarshal(values, &tempAts)
		ats[key[len(utils.ACTION_TIMING_PREFIX):]] = tempAts
	}

	return
}

func (rs *RedisStorage) GetDerivedChargers(key string, skipCache bool) (dcs utils.DerivedChargers, err error) {
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if !skipCache {
		if x, err := cache2go.GetCached(key); err == nil {
			return x.(utils.DerivedChargers), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
		err = rs.ms.Unmarshal(values, &dcs)
		cache2go.Cache(key, dcs)
	}
	return dcs, err
}

func (rs *RedisStorage) SetDerivedChargers(key string, dcs utils.DerivedChargers) (err error) {
	if len(dcs) == 0 {
		_, err = rs.db.Del(utils.DERIVEDCHARGERS_PREFIX + key)
		// FIXME: Does cache need cleanup too?
		return err
	}
	marshaled, err := rs.ms.Marshal(dcs)
	err = rs.db.Set(utils.DERIVEDCHARGERS_PREFIX+key, marshaled)
	return err
}

func (rs *RedisStorage) SetCdrStats(cs *CdrStats) error {
	marshaled, err := rs.ms.Marshal(cs)
	err = rs.db.Set(utils.CDR_STATS_PREFIX+cs.Id, marshaled)
	return err
}

func (rs *RedisStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	var values []byte
	if values, err = rs.db.Get(utils.CDR_STATS_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal(values, &cs)
	}
	return
}

func (rs *RedisStorage) GetAllCdrStats() (css []*CdrStats, err error) {
	keys, err := rs.db.Keys(utils.CDR_STATS_PREFIX + "*")
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		value, err := rs.db.Get(key)
		if err != nil {
			continue
		}
		cs := &CdrStats{}
		err = rs.ms.Unmarshal(value, cs)
		css = append(css, cs)
	}
	return
}

func (rs *RedisStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(cc)
	if err != nil {
		return
	}
	err = rs.db.Set(utils.LOG_CALL_COST_PREFIX+source+runid+"_"+cgrid, result)
	return
}

func (rs *RedisStorage) GetCallCostLog(cgrid, source, runid string) (cc *CallCost, err error) {
	var values []byte
	if values, err = rs.db.Get(utils.LOG_CALL_COST_PREFIX + source + runid + "_" + cgrid); err == nil {
		err = rs.ms.Unmarshal(values, cc)
	}
	return
}

func (rs *RedisStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	rs.db.Set(utils.LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%v*%v*%v", ubId, string(mat), string(mas))))
	return
}

func (rs *RedisStorage) LogActionPlan(source string, at *ActionPlan, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	err = rs.db.Set(utils.LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%v*%v", string(mat), string(mas))))
	return
}

func (rs *RedisStorage) LogError(uuid, source, runid, errstr string) (err error) {
	err = rs.db.Set(utils.
		LOG_ERR+source+runid+"_"+uuid, []byte(errstr))
	return
}
