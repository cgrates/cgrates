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
	"io/ioutil"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

var (
	ErrRedisNotFound = errors.New("RedisNotFound")
)

type RedisStorage struct {
	db *pool.Pool
	ms Marshaler
}

func NewRedisStorage(address string, db int, pass, mrshlerStr string, maxConns int) (*RedisStorage, error) {
	df := func(network, addr string) (*redis.Client, error) {
		client, err := redis.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		if len(pass) != 0 {
			if err = client.Cmd("AUTH", pass).Err; err != nil {
				client.Close()
				return nil, err
			}
		}
		if db != 0 {
			if err = client.Cmd("SELECT", db).Err; err != nil {
				client.Close()
				return nil, err
			}
		}
		return client, nil
	}
	p, err := pool.NewCustom("tcp", address, maxConns, df)
	if err != nil {
		return nil, err
	}
	var mrshler Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return &RedisStorage{db: p, ms: mrshler}, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Empty()
}

func (rs *RedisStorage) Flush(ignore string) error {
	return rs.db.Cmd("FLUSHDB").Err
}

func (rs *RedisStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	r := rs.db.Cmd("KEYS", prefix+"*")
	if r.Err != nil {
		return nil, r.Err
	}
	return r.List()
}

func (rs *RedisStorage) CacheRatingAll() error {
	return rs.cacheRating(nil, nil, nil, nil, nil, nil, nil, nil)
}

func (rs *RedisStorage) CacheRatingPrefixes(prefixes ...string) error {
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
	return rs.cacheRating(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.ACTION_PLAN_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (rs *RedisStorage) CacheRatingPrefixValues(prefixes map[string][]string) error {
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
	return rs.cacheRating(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.ACTION_PLAN_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (rs *RedisStorage) cacheRating(dKeys, rpKeys, rpfKeys, lcrKeys, dcsKeys, actKeys, aplKeys, shgKeys []string) (err error) {
	cache2go.BeginTransaction()
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	if dKeys == nil || (float64(cache2go.CountEntries(utils.DESTINATION_PREFIX))*utils.DESTINATIONS_LOAD_THRESHOLD < float64(len(dKeys))) {
		// if need to load more than a half of exiting keys load them all
		utils.Logger.Info("Caching all destinations")
		if dKeys, err = conn.Cmd("KEYS", utils.DESTINATION_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.DESTINATION_PREFIX)
	} else if len(dKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching destinations: %v", dKeys))
		CleanStalePrefixes(dKeys)
	}
	for _, key := range dKeys {
		if len(key) <= len(utils.DESTINATION_PREFIX) {
			utils.Logger.Warning(fmt.Sprintf("Got malformed destination id: %s", key))
			continue
		}
		if _, err = rs.GetDestination(key[len(utils.DESTINATION_PREFIX):]); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(dKeys) != 0 {
		utils.Logger.Info("Finished destinations caching.")
	}
	if rpKeys == nil {
		utils.Logger.Info("Caching all rating plans")
		if rpKeys, err = conn.Cmd("KEYS", utils.RATING_PLAN_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.RATING_PLAN_PREFIX)
	} else if len(rpKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching rating plans: %v", rpKeys))
	}
	for _, key := range rpKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetRatingPlan(key[len(utils.RATING_PLAN_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(rpKeys) != 0 {
		utils.Logger.Info("Finished rating plans caching.")
	}
	if rpfKeys == nil {
		utils.Logger.Info("Caching all rating profiles")
		if rpfKeys, err = conn.Cmd("KEYS", utils.RATING_PROFILE_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.RATING_PROFILE_PREFIX)
	} else if len(rpfKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching rating profile: %v", rpfKeys))
	}
	for _, key := range rpfKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetRatingProfile(key[len(utils.RATING_PROFILE_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(rpfKeys) != 0 {
		utils.Logger.Info("Finished rating profile caching.")
	}
	if lcrKeys == nil {
		utils.Logger.Info("Caching LCR rules.")
		if lcrKeys, err = conn.Cmd("KEYS", utils.LCR_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.LCR_PREFIX)
	} else if len(lcrKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching LCR rules: %v", lcrKeys))
	}
	for _, key := range lcrKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetLCR(key[len(utils.LCR_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(lcrKeys) != 0 {
		utils.Logger.Info("Finished LCR rules caching.")
	}
	// DerivedChargers caching
	if dcsKeys == nil {
		utils.Logger.Info("Caching all derived chargers")
		if dcsKeys, err = conn.Cmd("KEYS", utils.DERIVEDCHARGERS_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.DERIVEDCHARGERS_PREFIX)
	} else if len(dcsKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching derived chargers: %v", dcsKeys))
	}
	for _, key := range dcsKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetDerivedChargers(key[len(utils.DERIVEDCHARGERS_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(dcsKeys) != 0 {
		utils.Logger.Info("Finished derived chargers caching.")
	}
	if actKeys == nil {
		utils.Logger.Info("Caching all actions")
		if actKeys, err = conn.Cmd("KEYS", utils.ACTION_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.ACTION_PREFIX)
	} else if len(actKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching actions: %v", actKeys))
	}
	for _, key := range actKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetActions(key[len(utils.ACTION_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(actKeys) != 0 {
		utils.Logger.Info("Finished actions caching.")
	}

	if aplKeys == nil {
		utils.Logger.Info("Caching all action plans")
		if aplKeys, err = rs.db.Cmd("KEYS", utils.ACTION_PLAN_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.ACTION_PLAN_PREFIX)
	} else if len(aplKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching action plan: %v", aplKeys))
	}
	for _, key := range aplKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(aplKeys) != 0 {
		utils.Logger.Info("Finished action plans caching.")
	}

	if shgKeys == nil {
		utils.Logger.Info("Caching all shared groups")
		if shgKeys, err = conn.Cmd("KEYS", utils.SHARED_GROUP_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.SHARED_GROUP_PREFIX)
	} else if len(shgKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching shared groups: %v", shgKeys))
	}
	for _, key := range shgKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetSharedGroup(key[len(utils.SHARED_GROUP_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(shgKeys) != 0 {
		utils.Logger.Info("Finished shared groups caching.")
	}

	cache2go.CommitTransaction()
	return nil
}

func (rs *RedisStorage) CacheAccountingAll() error {
	return rs.cacheAccounting(nil)
}

func (rs *RedisStorage) CacheAccountingPrefixes(prefixes ...string) error {
	pm := map[string][]string{
		utils.ALIASES_PREFIX: []string{},
	}
	for _, prefix := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = nil
	}
	return rs.cacheAccounting(pm[utils.ALIASES_PREFIX])
}

func (rs *RedisStorage) CacheAccountingPrefixValues(prefixes map[string][]string) error {
	pm := map[string][]string{
		utils.ALIASES_PREFIX: []string{},
	}
	for prefix, ids := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = ids
	}
	return rs.cacheAccounting(pm[utils.ALIASES_PREFIX])
}

func (rs *RedisStorage) cacheAccounting(alsKeys []string) (err error) {
	cache2go.BeginTransaction()
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	if alsKeys == nil {
		utils.Logger.Info("Caching all aliases")
		if alsKeys, err = conn.Cmd("KEYS", utils.ALIASES_PREFIX+"*").List(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.ALIASES_PREFIX)
		cache2go.RemPrefixKey(utils.REVERSE_ALIASES_PREFIX)
	} else if len(alsKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching aliases: %v", alsKeys))
	}
	al := &Alias{}
	for _, key := range alsKeys {
		// check if it already exists
		// to remove reverse cache keys
		if avs, err := cache2go.Get(key); err == nil && avs != nil {
			al.Values = avs.(AliasValues)
			al.SetId(key[len(utils.ALIASES_PREFIX):])
			al.RemoveReverseCache()
		}
		cache2go.RemKey(key)
		if _, err = rs.GetAlias(key[len(utils.ALIASES_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(alsKeys) != 0 {
		utils.Logger.Info("Finished aliases caching.")
	}
	utils.Logger.Info("Caching load history")
	if _, err = rs.GetLoadHistory(1, true); err != nil {
		cache2go.RollbackTransaction()
		return err
	}
	utils.Logger.Info("Finished load history caching.")
	cache2go.CommitTransaction()
	return nil
}

// Used to check if specific subject is stored using prefix key attached to entity
func (rs *RedisStorage) HasData(category, subject string) (bool, error) {
	switch category {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX, utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX, utils.DERIVEDCHARGERS_PREFIX:
		i, err := rs.db.Cmd("EXISTS", category+subject).Int()
		return i == 1, err
	}
	return false, errors.New("unsupported HasData category")
}

func (rs *RedisStorage) GetRatingPlan(key string, skipCache bool) (rp *RatingPlan, err error) {
	key = utils.RATING_PLAN_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*RatingPlan), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
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
	err = rs.db.Cmd("SET", utils.RATING_PLAN_PREFIX+rp.Id, b.Bytes()).Err
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(rp.GetHistoryRecord(), &response)
	}
	return
}

func (rs *RedisStorage) GetRatingProfile(key string, skipCache bool) (rpf *RatingProfile, err error) {

	key = utils.RATING_PROFILE_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*RatingProfile), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		rpf = new(RatingProfile)
		err = rs.ms.Unmarshal(values, rpf)
		cache2go.Cache(key, rpf)
	}
	return
}

func (rs *RedisStorage) SetRatingProfile(rpf *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rpf)
	err = rs.db.Cmd("SET", utils.RATING_PROFILE_PREFIX+rpf.Id, result).Err
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(rpf.GetHistoryRecord(false), &response)
	}
	return
}

func (rs *RedisStorage) RemoveRatingProfile(key string) error {
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	keys, err := conn.Cmd("KEYS", utils.RATING_PROFILE_PREFIX+key+"*").List()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err = conn.Cmd("DEL", key).Err; err != nil {
			return err
		}
		cache2go.RemKey(key)
		rpf := &RatingProfile{Id: key}
		if historyScribe != nil {
			response := 0
			go historyScribe.Record(rpf.GetHistoryRecord(true), &response)
		}
	}
	return nil
}

func (rs *RedisStorage) GetLCR(key string, skipCache bool) (lcr *LCR, err error) {
	key = utils.LCR_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*LCR), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &lcr)
		cache2go.Cache(key, lcr)
	}
	return
}

func (rs *RedisStorage) SetLCR(lcr *LCR) (err error) {
	result, err := rs.ms.Marshal(lcr)
	err = rs.db.Cmd("SET", utils.LCR_PREFIX+lcr.GetId(), result).Err
	cache2go.Cache(utils.LCR_PREFIX+lcr.GetId(), lcr)
	return
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	key = utils.DESTINATION_PREFIX + key
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); len(values) > 0 && err == nil {
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
			cache2go.Push(utils.DESTINATION_PREFIX+p, dest.Id)
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
	err = rs.db.Cmd("SET", utils.DESTINATION_PREFIX+dest.Id, b.Bytes()).Err
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(dest.GetHistoryRecord(), &response)
	}
	return
}

func (rs *RedisStorage) GetActions(key string, skipCache bool) (as Actions, err error) {
	key = utils.ACTION_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(Actions), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &as)
		cache2go.Cache(key, as)
	}
	return
}

func (rs *RedisStorage) SetActions(key string, as Actions) (err error) {
	result, err := rs.ms.Marshal(&as)
	err = rs.db.Cmd("SET", utils.ACTION_PREFIX+key, result).Err
	return
}

func (rs *RedisStorage) GetSharedGroup(key string, skipCache bool) (sg *SharedGroup, err error) {
	key = utils.SHARED_GROUP_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*SharedGroup), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &sg)
		cache2go.Cache(key, sg)
	}
	return
}

func (rs *RedisStorage) SetSharedGroup(sg *SharedGroup) (err error) {
	result, err := rs.ms.Marshal(sg)
	err = rs.db.Cmd("SET", utils.SHARED_GROUP_PREFIX+sg.Id, result).Err
	return
}

func (rs *RedisStorage) GetAccount(key string) (*Account, error) {
	rpl := rs.db.Cmd("GET", utils.ACCOUNT_PREFIX+key)
	if rpl.Err != nil {
		return nil, rpl.Err
	} else if rpl.IsType(redis.Nil) {
		return nil, ErrRedisNotFound
	}
	values, err := rpl.Bytes()
	if err != nil {
		return nil, err
	}
	ub := &Account{Id: key}
	if err = rs.ms.Unmarshal(values, ub); err != nil {
		return nil, err
	}
	return ub, nil
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
	err = rs.db.Cmd("SET", utils.ACCOUNT_PREFIX+ub.Id, result).Err
	return
}

func (rs *RedisStorage) RemoveAccount(key string) (err error) {
	return rs.db.Cmd("DEL", utils.ACCOUNT_PREFIX+key).Err

}

func (rs *RedisStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	var values []byte
	if values, err = rs.db.Cmd("GET", utils.CDR_STATS_QUEUE_PREFIX+key).Bytes(); err == nil {
		sq = &StatsQueue{}
		err = rs.ms.Unmarshal(values, &sq)
	}
	return
}

func (rs *RedisStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	result, err := rs.ms.Marshal(sq)
	err = rs.db.Cmd("SET", utils.CDR_STATS_QUEUE_PREFIX+sq.GetId(), result).Err
	return
}

func (rs *RedisStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return nil, err
	}
	defer rs.db.Put(conn)
	keys, err := conn.Cmd("KEYS", utils.PUBSUB_SUBSCRIBERS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	result = make(map[string]*SubscriberData)
	for _, key := range keys {
		if values, err := conn.Cmd("GET", key).Bytes(); err == nil {
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
	if err != nil {
		return err
	}
	return rs.db.Cmd("SET", utils.PUBSUB_SUBSCRIBERS_PREFIX+key, result).Err
}

func (rs *RedisStorage) RemoveSubscriber(key string) (err error) {
	err = rs.db.Cmd("DEL", utils.PUBSUB_SUBSCRIBERS_PREFIX+key).Err
	return
}

func (rs *RedisStorage) SetUser(up *UserProfile) (err error) {
	result, err := rs.ms.Marshal(up)
	if err != nil {
		return err
	}
	return rs.db.Cmd("SET", utils.USERS_PREFIX+up.GetId(), result).Err
}

func (rs *RedisStorage) GetUser(key string) (up *UserProfile, err error) {
	var values []byte
	if values, err = rs.db.Cmd("GET", utils.USERS_PREFIX+key).Bytes(); err == nil {
		up = &UserProfile{}
		err = rs.ms.Unmarshal(values, &up)
	}
	return
}

func (rs *RedisStorage) GetUsers() (result []*UserProfile, err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return nil, err
	}
	defer rs.db.Put(conn)
	keys, err := conn.Cmd("KEYS", utils.USERS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if values, err := conn.Cmd("GET", key).Bytes(); err == nil {
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
	return rs.db.Cmd("DEL", utils.USERS_PREFIX+key).Err
}

func (rs *RedisStorage) SetAlias(al *Alias) (err error) {
	result, err := rs.ms.Marshal(al.Values)
	if err != nil {
		return err
	}
	return rs.db.Cmd("SET", utils.ALIASES_PREFIX+al.GetId(), result).Err
}

func (rs *RedisStorage) GetAlias(key string, skipCache bool) (al *Alias, err error) {
	origKey := key
	key = utils.ALIASES_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			al = &Alias{Values: x.(AliasValues)}
			al.SetId(origKey)
			return al, nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		al = &Alias{Values: make(AliasValues, 0)}
		al.SetId(origKey)
		err = rs.ms.Unmarshal(values, &al.Values)
		if err == nil {
			cache2go.Cache(key, al.Values)
			// cache reverse alias
			al.SetReverseCache()
		}
	}
	return
}

func (rs *RedisStorage) RemoveAlias(key string) (err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	al := &Alias{}
	al.SetId(key)
	key = utils.ALIASES_PREFIX + key
	aliasValues := make(AliasValues, 0)
	if values, err := conn.Cmd("GET", key).Bytes(); err == nil {
		rs.ms.Unmarshal(values, &aliasValues)
	}
	al.Values = aliasValues
	err = conn.Cmd("DEL", key).Err
	if err == nil {
		al.RemoveReverseCache()
		cache2go.RemKey(key)
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (rs *RedisStorage) GetLoadHistory(limit int, skipCache bool) ([]*LoadInstance, error) {
	if limit == 0 {
		return nil, nil
	}
	if !skipCache {
		if x, err := cache2go.Get(utils.LOADINST_KEY); err != nil {
			return nil, err
		} else {
			items := x.([]*LoadInstance)
			if len(items) < limit || limit == -1 {
				return items, nil
			}
			return items[:limit], nil
		}
	}
	if limit != -1 {
		limit -= -1 // Decrease limit to match redis approach on lrange
	}
	marshaleds, err := rs.db.Cmd("LRANGE", utils.LOADINST_KEY, 0, limit).ListBytes()
	if err != nil {
		return nil, err
	}
	loadInsts := make([]*LoadInstance, len(marshaleds))
	for idx, marshaled := range marshaleds {
		var lInst LoadInstance
		err = rs.ms.Unmarshal(marshaled, &lInst)
		if err != nil {
			return nil, err
		}
		loadInsts[idx] = &lInst
	}
	cache2go.RemKey(utils.LOADINST_KEY)
	cache2go.Cache(utils.LOADINST_KEY, loadInsts)
	return loadInsts, nil
}

// Adds a single load instance to load history
func (rs *RedisStorage) AddLoadHistory(ldInst *LoadInstance, loadHistSize int) error {
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	if loadHistSize == 0 { // Load history disabled
		return nil
	}
	marshaled, err := rs.ms.Marshal(&ldInst)
	if err != nil {
		return err
	}
	_, err = Guardian.Guard(func() (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
		histLen, err := conn.Cmd("LLEN", utils.LOADINST_KEY).Int()
		if err != nil {
			return nil, err
		}
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			if err := conn.Cmd("RPOP", utils.LOADINST_KEY).Err; err != nil {
				return nil, err
			}
		}
		err = conn.Cmd("LPUSH", utils.LOADINST_KEY, marshaled).Err
		return nil, err
	}, 0, utils.LOADINST_KEY)
	return err
}

func (rs *RedisStorage) GetActionTriggers(key string) (atrs ActionTriggers, err error) {
	var values []byte
	if values, err = rs.db.Cmd("GET", utils.ACTION_TRIGGER_PREFIX+key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &atrs)
	}
	return
}

func (rs *RedisStorage) SetActionTriggers(key string, atrs ActionTriggers) (err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	if len(atrs) == 0 {
		// delete the key
		return conn.Cmd("DEL", utils.ACTION_TRIGGER_PREFIX+key).Err
	}
	result, err := rs.ms.Marshal(atrs)
	if err != nil {
		return err
	}
	return conn.Cmd("SET", utils.ACTION_TRIGGER_PREFIX+key, result).Err
}

func (rs *RedisStorage) GetActionPlan(key string, skipCache bool) (ats *ActionPlan, err error) {
	key = utils.ACTION_PLAN_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*ActionPlan), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
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
		err = rs.ms.Unmarshal(out, &ats)
		cache2go.Cache(key, ats)
	}
	return
}

func (rs *RedisStorage) SetActionPlan(key string, ats *ActionPlan) (err error) {
	if len(ats.ActionTimings) == 0 {
		// delete the key
		err = rs.db.Cmd("DEL", utils.ACTION_PLAN_PREFIX+key).Err
		cache2go.RemKey(utils.ACTION_PLAN_PREFIX + key)
		return err
	}
	// get existing action plan to merge the account ids
	if existingAts, _ := rs.GetActionPlan(utils.ACTION_PLAN_PREFIX, true); existingAts != nil {
		for accID := range existingAts.AccountIDs {
			ats.AccountIDs[accID] = true
		}
	}

	result, err := rs.ms.Marshal(ats)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	return rs.db.Cmd("SET", utils.ACTION_PLAN_PREFIX+key, b.Bytes()).Err
}

func (rs *RedisStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	apls, err := cache2go.GetAllEntries(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}

	ats = make(map[string]*ActionPlan, len(apls))
	for key, value := range apls {
		apl := value.(*ActionPlan)
		ats[key] = apl
	}

	return
}

func (rs *RedisStorage) PushTask(t *Task) error {
	result, err := rs.ms.Marshal(t)
	if err != nil {
		return err
	}
	return rs.db.Cmd("RPUSH", utils.TASKS_KEY, result).Err
}

func (rs *RedisStorage) PopTask() (t *Task, err error) {
	var values []byte
	if values, err = rs.db.Cmd("LPOP", utils.TASKS_KEY).Bytes(); err == nil {
		t = &Task{}
		err = rs.ms.Unmarshal(values, t)
	}
	return
}

func (rs *RedisStorage) GetDerivedChargers(key string, skipCache bool) (dcs *utils.DerivedChargers, err error) {
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			return x.(*utils.DerivedChargers), nil
		} else {
			return nil, err
		}
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &dcs)
		cache2go.Cache(key, dcs)
	}
	return dcs, err
}

func (rs *RedisStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers) (err error) {
	if dcs == nil || len(dcs.Chargers) == 0 {
		err = rs.db.Cmd("DEL", utils.DERIVEDCHARGERS_PREFIX+key).Err
		cache2go.RemKey(utils.DERIVEDCHARGERS_PREFIX + key)
		return err
	}
	marshaled, err := rs.ms.Marshal(dcs)
	if err != nil {
		return err
	}
	return rs.db.Cmd("SET", utils.DERIVEDCHARGERS_PREFIX+key, marshaled).Err
}

func (rs *RedisStorage) SetCdrStats(cs *CdrStats) error {
	marshaled, err := rs.ms.Marshal(cs)
	if err != nil {
		return err
	}
	return rs.db.Cmd("SET", utils.CDR_STATS_PREFIX+cs.Id, marshaled).Err
}

func (rs *RedisStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	var values []byte
	if values, err = rs.db.Cmd("GET", utils.CDR_STATS_PREFIX+key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &cs)
	}
	return
}

func (rs *RedisStorage) GetAllCdrStats() (css []*CdrStats, err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return nil, err
	}
	defer rs.db.Put(conn)
	keys, err := conn.Cmd("KEYS", utils.CDR_STATS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		value, err := conn.Cmd("GET", key).Bytes()
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
	return rs.db.Cmd("SET", utils.LOG_CALL_COST_PREFIX+source+runid+"_"+cgrid, result).Err
}

func (rs *RedisStorage) GetCallCostLog(cgrid, source, runid string) (cc *CallCost, err error) {
	var values []byte
	if values, err = rs.db.Cmd("GET", utils.LOG_CALL_COST_PREFIX+source+runid+"_"+cgrid).Bytes(); err == nil {
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
	return rs.db.Cmd("SET", utils.LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%v*%v*%v", ubId, string(mat), string(mas)))).Err
}

func (rs *RedisStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	return rs.db.Cmd("SET", utils.LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%v*%v", string(mat), string(mas)))).Err
}
