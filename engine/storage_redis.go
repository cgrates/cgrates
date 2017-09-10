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
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

type RedisStorage struct {
	dbPool          *pool.Pool
	maxConns        int
	ms              Marshaler
	cacheCfg        config.CacheConfig
	loadHistorySize int
}

func NewRedisStorage(address string, db int, pass, mrshlerStr string, maxConns int, cacheCfg config.CacheConfig, loadHistorySize int) (*RedisStorage, error) {
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
	return &RedisStorage{dbPool: p, maxConns: maxConns, ms: mrshler,
		cacheCfg: cacheCfg, loadHistorySize: loadHistorySize}, nil
}

// This CMD function get a connection from the pool.
// Handles automatic failover in case of network disconnects
func (rs *RedisStorage) Cmd(cmd string, args ...interface{}) *redis.Resp {
	c1, err := rs.dbPool.Get()
	if err != nil {
		return redis.NewResp(err)
	}
	result := c1.Cmd(cmd, args...)
	if result.IsType(redis.IOErr) { // Failover mecanism
		utils.Logger.Warning(fmt.Sprintf("<RedisStorage> error <%s>, attempting failover.", result.Err.Error()))
		for i := 0; i < rs.maxConns; i++ { // Two attempts, one on connection of original pool, one on new pool
			c2, err := rs.dbPool.Get()
			if err == nil {
				if result2 := c2.Cmd(cmd, args...); !result2.IsType(redis.IOErr) {
					rs.dbPool.Put(c2)
					return result2
				}
			}
		}
	} else {
		rs.dbPool.Put(c1)
	}
	return result
}

func (rs *RedisStorage) Close() {
	rs.dbPool.Empty()
}

func (rs *RedisStorage) Flush(ignore string) error {
	return rs.Cmd("FLUSHDB").Err
}

func (rs *RedisStorage) Marshaler() Marshaler {
	return rs.ms
}

func (rs *RedisStorage) SelectDatabase(dbName string) (err error) {
	return rs.Cmd("SELECT", dbName).Err
}

func (rs *RedisStorage) LoadDataDBCache(dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs,
	aplIDs, aaPlIDs, atrgIDs, sgIDs, lcrIDs, dcIDs, alsIDs, rvAlsIDs, rpIDs, resIDs []string) (err error) {
	for key, ids := range map[string][]string{
		utils.DESTINATION_PREFIX:         dstIDs,
		utils.REVERSE_DESTINATION_PREFIX: rvDstIDs,
		utils.RATING_PLAN_PREFIX:         rplIDs,
		utils.RATING_PROFILE_PREFIX:      rpfIDs,
		utils.ACTION_PREFIX:              actIDs,
		utils.ACTION_PLAN_PREFIX:         aplIDs,
		utils.AccountActionPlansPrefix:   aaPlIDs,
		utils.ACTION_TRIGGER_PREFIX:      atrgIDs,
		utils.SHARED_GROUP_PREFIX:        sgIDs,
		utils.LCR_PREFIX:                 lcrIDs,
		utils.DERIVEDCHARGERS_PREFIX:     dcIDs,
		utils.ALIASES_PREFIX:             alsIDs,
		utils.REVERSE_ALIASES_PREFIX:     rvAlsIDs,
		utils.ResourceProfilesPrefix:     rpIDs,
		utils.ResourcesPrefix:            resIDs,
	} {
		if err = rs.CacheDataFromDB(key, ids, false); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStorage) RebuildReverseForPrefix(prefix string) (err error) {
	if !utils.IsSliceMember([]string{utils.REVERSE_DESTINATION_PREFIX, utils.REVERSE_ALIASES_PREFIX, utils.AccountActionPlansPrefix}, prefix) {
		return utils.ErrInvalidKey
	}
	var keys []string
	keys, err = rs.GetKeysForPrefix(prefix)
	if err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd("DEL", key).Err; err != nil {
			return
		}
	}
	switch prefix {
	case utils.REVERSE_DESTINATION_PREFIX:
		if keys, err = rs.GetKeysForPrefix(utils.DESTINATION_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			dest, err := rs.GetDestination(key[len(utils.DESTINATION_PREFIX):], true, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err = rs.SetReverseDestination(dest, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.REVERSE_ALIASES_PREFIX:
		if keys, err = rs.GetKeysForPrefix(utils.ALIASES_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			al, err := rs.GetAlias(key[len(utils.ALIASES_PREFIX):], true, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err = rs.SetReverseAlias(al, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.AccountActionPlansPrefix:
		if keys, err = rs.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			apl, err := rs.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], true, utils.NonTransactional) // skipCache on get since loader checks and caches empty data for loaded objects
			if err != nil {
				return err
			}
			for acntID := range apl.AccountIDs {
				if err = rs.SetAccountActionPlans(acntID, []string{apl.Id}, false); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CacheDataFromDB loads data to cache
// prfx represents the cache prefix, ids should be nil if all available data should be loaded
// mustBeCached specifies that data needs to be cached in order to be retrieved from db
func (rs *RedisStorage) CacheDataFromDB(prfx string, ids []string, mustBeCached bool) (err error) {
	if !utils.IsSliceMember([]string{utils.DESTINATION_PREFIX,
		utils.REVERSE_DESTINATION_PREFIX,
		utils.RATING_PLAN_PREFIX,
		utils.RATING_PROFILE_PREFIX,
		utils.ACTION_PREFIX,
		utils.ACTION_PLAN_PREFIX,
		utils.AccountActionPlansPrefix,
		utils.ACTION_TRIGGER_PREFIX,
		utils.SHARED_GROUP_PREFIX,
		utils.DERIVEDCHARGERS_PREFIX,
		utils.LCR_PREFIX,
		utils.ALIASES_PREFIX,
		utils.REVERSE_ALIASES_PREFIX,
		utils.ResourceProfilesPrefix,
		utils.ResourcesPrefix,
		utils.TimingsPrefix}, prfx) {
		return utils.NewCGRError(utils.REDIS,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedCachePrefix,
			fmt.Sprintf("prefix <%s> is not a supported cache prefix", prfx))
	}
	if ids == nil {
		keyIDs, err := rs.GetKeysForPrefix(prfx)
		if err != nil {
			return utils.NewCGRError(utils.REDIS,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("redis error <%s> querying keys for prefix: <%s>", prfx))
		}
		for _, keyID := range keyIDs {
			if mustBeCached { // Only consider loading ids which are already in cache
				if _, hasIt := cache.Get(keyID); !hasIt {
					continue
				}
			}
			ids = append(ids, keyID[len(prfx):])
		}
		var nrItems int
		if cCfg, has := rs.cacheCfg[utils.CachePrefixToInstance[prfx]]; has {
			nrItems = cCfg.Limit
		}
		if nrItems > 0 && nrItems < len(ids) {
			ids = ids[:nrItems]
		}
	}
	for _, dataID := range ids {
		if mustBeCached {
			if _, hasIt := cache.Get(prfx + dataID); !hasIt { // only cache if previously there
				continue
			}
		}
		switch prfx {
		case utils.DESTINATION_PREFIX:
			_, err = rs.GetDestination(dataID, true, utils.NonTransactional)
		case utils.REVERSE_DESTINATION_PREFIX:
			_, err = rs.GetReverseDestination(dataID, true, utils.NonTransactional)
		case utils.RATING_PLAN_PREFIX:
			_, err = rs.GetRatingPlan(dataID, true, utils.NonTransactional)
		case utils.RATING_PROFILE_PREFIX:
			_, err = rs.GetRatingProfile(dataID, true, utils.NonTransactional)
		case utils.ACTION_PREFIX:
			_, err = rs.GetActions(dataID, true, utils.NonTransactional)
		case utils.ACTION_PLAN_PREFIX:
			_, err = rs.GetActionPlan(dataID, true, utils.NonTransactional)
		case utils.AccountActionPlansPrefix:
			_, err = rs.GetAccountActionPlans(dataID, true, utils.NonTransactional)
		case utils.ACTION_TRIGGER_PREFIX:
			_, err = rs.GetActionTriggers(dataID, true, utils.NonTransactional)
		case utils.SHARED_GROUP_PREFIX:
			_, err = rs.GetSharedGroup(dataID, true, utils.NonTransactional)
		case utils.DERIVEDCHARGERS_PREFIX:
			_, err = rs.GetDerivedChargers(dataID, true, utils.NonTransactional)
		case utils.LCR_PREFIX:
			_, err = rs.GetLCR(dataID, true, utils.NonTransactional)
		case utils.ALIASES_PREFIX:
			_, err = rs.GetAlias(dataID, true, utils.NonTransactional)
		case utils.REVERSE_ALIASES_PREFIX:
			_, err = rs.GetReverseAlias(dataID, true, utils.NonTransactional)
		case utils.ResourceProfilesPrefix:
			_, err = rs.GetResourceProfile(dataID, true, utils.NonTransactional)
		case utils.ResourcesPrefix:
			_, err = rs.GetResource(dataID, true, utils.NonTransactional)
		case utils.TimingsPrefix:
			_, err = rs.GetTiming(dataID, true, utils.NonTransactional)
		}
		if err != nil {
			return utils.NewCGRError(utils.REDIS,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error <%s> querying redis for category: <%s>, dataID: <%s>", prfx, dataID))
		}
	}
	return
}

func (rs *RedisStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	r := rs.Cmd("KEYS", prefix+"*")
	if r.Err != nil {
		return nil, r.Err
	}
	if keys, _ := r.List(); len(keys) != 0 {
		return keys, nil
	}
	return nil, nil

}

// Used to check if specific subject is stored using prefix key attached to entity
func (rs *RedisStorage) HasData(category, subject string) (bool, error) {
	switch category {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX, utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX, utils.DERIVEDCHARGERS_PREFIX, utils.ResourcesPrefix:
		i, err := rs.Cmd("EXISTS", category+subject).Int()
		return i == 1, err
	}
	return false, errors.New("unsupported HasData category")
}

func (rs *RedisStorage) GetRatingPlan(key string, skipCache bool, transactionID string) (rp *RatingPlan, err error) {
	key = utils.RATING_PLAN_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*RatingPlan), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return nil, err
	}
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
	if err != nil {
		return nil, err
	}
	cache.Set(key, rp, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetRatingPlan(rp *RatingPlan, transactionID string) (err error) {
	result, err := rs.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.Cmd("SET", utils.RATING_PLAN_PREFIX+rp.Id, b.Bytes()).Err
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(), &response)
	}
	return
}

func (rs *RedisStorage) GetRatingProfile(key string, skipCache bool, transactionID string) (rpf *RatingProfile, err error) {
	key = utils.RATING_PROFILE_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*RatingProfile), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &rpf); err != nil {
		return
	}
	cache.Set(key, rpf, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetRatingProfile(rpf *RatingProfile, transactionID string) (err error) {
	result, err := rs.ms.Marshal(rpf)
	if err != nil {
		return err
	}
	key := utils.RATING_PROFILE_PREFIX + rpf.Id
	if err = rs.Cmd("SET", key, result).Err; err != nil {
		return
	}
	if historyScribe != nil {
		response := 0
		go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(false), &response)
	}
	return
}

func (rs *RedisStorage) RemoveRatingProfile(key string, transactionID string) error {
	keys, err := rs.Cmd("KEYS", utils.RATING_PROFILE_PREFIX+key+"*").List()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err = rs.Cmd("DEL", key).Err; err != nil {
			return err
		}
		cache.RemKey(key, cacheCommit(transactionID), transactionID)
		rpf := &RatingProfile{Id: key}
		if historyScribe != nil {
			response := 0
			go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(true), &response)
		}
	}
	return nil
}

func (rs *RedisStorage) GetLCR(key string, skipCache bool, transactionID string) (lcr *LCR, err error) {
	key = utils.LCR_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*LCR), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &lcr); err != nil {
		return
	}
	cache.Set(key, lcr, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetLCR(lcr *LCR, transactionID string) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(lcr); err != nil {
		return
	}
	key := utils.LCR_PREFIX + lcr.GetId()
	if err = rs.Cmd("SET", key, result).Err; err != nil {
		return
	}
	return
}

// GetDestination retrieves a destination with id from  tp_db
func (rs *RedisStorage) GetDestination(key string, skipCache bool, transactionID string) (dest *Destination, err error) {
	key = utils.DESTINATION_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Destination), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
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
	err = rs.ms.Unmarshal(out, &dest)
	if err != nil {
		return nil, err
	}
	cache.Set(key, dest, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination, transactionID string) (err error) {
	result, err := rs.ms.Marshal(dest)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	key := utils.DESTINATION_PREFIX + dest.Id
	if err = rs.Cmd("SET", key, b.Bytes()).Err; err != nil {
		return err
	}
	if historyScribe != nil {
		response := 0
		go historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	return
}

func (rs *RedisStorage) GetReverseDestination(key string, skipCache bool, transactionID string) (ids []string, err error) {
	key = utils.REVERSE_DESTINATION_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	if ids, err = rs.Cmd("SMEMBERS", key).List(); err != nil {
		return
	} else if len(ids) == 0 {
		cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		err = utils.ErrNotFound
		return
	}
	cache.Set(key, ids, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetReverseDestination(dest *Destination, transactionID string) (err error) {
	for _, p := range dest.Prefixes {
		key := utils.REVERSE_DESTINATION_PREFIX + p
		if err = rs.Cmd("SADD", key, dest.Id).Err; err != nil {
			break
		}
	}
	return
}

func (rs *RedisStorage) RemoveDestination(destID, transactionID string) (err error) {
	key := utils.DESTINATION_PREFIX + destID
	// get destination for prefix list
	d, err := rs.GetDestination(destID, false, transactionID)
	if err != nil {
		return
	}
	err = rs.Cmd("DEL", key).Err
	if err != nil {
		return err
	}
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	for _, prefix := range d.Prefixes {
		err = rs.Cmd("SREM", utils.REVERSE_DESTINATION_PREFIX+prefix, destID).Err
		if err != nil {
			return err
		}
		rs.GetReverseDestination(prefix, true, transactionID) // it will recache the destination
	}
	return
}

func (rs *RedisStorage) UpdateReverseDestination(oldDest, newDest *Destination, transactionID string) error {
	//log.Printf("Old: %+v, New: %+v", oldDest, newDest)
	var obsoletePrefixes []string
	var addedPrefixes []string
	var found bool
	if oldDest == nil {
		oldDest = new(Destination) // so we can process prefixes
	}
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
		err = rs.Cmd("SREM", utils.REVERSE_DESTINATION_PREFIX+obsoletePrefix, oldDest.Id).Err
		if err != nil {
			return err
		}
		cache.RemKey(utils.REVERSE_DESTINATION_PREFIX+obsoletePrefix, cCommit, transactionID)
	}

	// add the id to all new prefixes
	for _, addedPrefix := range addedPrefixes {
		err = rs.Cmd("SADD", utils.REVERSE_DESTINATION_PREFIX+addedPrefix, newDest.Id).Err
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *RedisStorage) GetActions(key string, skipCache bool, transactionID string) (as Actions, err error) {
	key = utils.ACTION_PREFIX + key
	if !skipCache {
		if x, err := cache.GetCloned(key); err != nil {
			if err.Error() != utils.ItemNotFound {
				return nil, err
			}
		} else if x == nil {
			return nil, utils.ErrNotFound
		} else {
			return x.(Actions), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &as); err != nil {
		return
	}
	cache.Set(key, as, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetActions(key string, as Actions, transactionID string) (err error) {
	result, err := rs.ms.Marshal(&as)
	err = rs.Cmd("SET", utils.ACTION_PREFIX+key, result).Err
	return
}

func (rs *RedisStorage) RemoveActions(key string, transactionID string) (err error) {
	err = rs.Cmd("DEL", utils.ACTION_PREFIX+key).Err
	cache.RemKey(utils.ACTION_PREFIX+key, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) GetSharedGroup(key string, skipCache bool, transactionID string) (sg *SharedGroup, err error) {
	key = utils.SHARED_GROUP_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*SharedGroup), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &sg); err != nil {
		return
	}
	cache.Set(key, sg, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetSharedGroup(sg *SharedGroup, transactionID string) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sg); err != nil {
		return
	}
	err = rs.Cmd("SET", utils.SHARED_GROUP_PREFIX+sg.Id, result).Err
	return
}

func (rs *RedisStorage) GetAccount(key string) (*Account, error) {
	rpl := rs.Cmd("GET", utils.ACCOUNT_PREFIX+key)
	if rpl.Err != nil {
		return nil, rpl.Err
	} else if rpl.IsType(redis.Nil) {
		return nil, utils.ErrNotFound
	}
	values, err := rpl.Bytes()
	if err != nil {
		return nil, err
	}
	ub := &Account{ID: key}
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
		if ac, err := rs.GetAccount(ub.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = ub.ActionTriggers
			ac.UnitCounters = ub.UnitCounters
			ac.AllowNegative = ub.AllowNegative
			ac.Disabled = ub.Disabled
			ub = ac
		}
	}
	result, err := rs.ms.Marshal(ub)
	err = rs.Cmd("SET", utils.ACCOUNT_PREFIX+ub.ID, result).Err
	return
}

func (rs *RedisStorage) RemoveAccount(key string) (err error) {
	return rs.Cmd("DEL", utils.ACCOUNT_PREFIX+key).Err

}

func (rs *RedisStorage) GetCdrStatsQueue(key string) (sq *CDRStatsQueue, err error) {
	var values []byte
	if values, err = rs.Cmd("GET", utils.CDR_STATS_QUEUE_PREFIX+key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	sq = new(CDRStatsQueue)
	if err = rs.ms.Unmarshal(values, &sq); err != nil {
		return nil, err
	}
	return
}

func (rs *RedisStorage) SetCdrStatsQueue(sq *CDRStatsQueue) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sq); err != nil {
		return
	}
	return rs.Cmd("SET", utils.CDR_STATS_QUEUE_PREFIX+sq.GetId(), result).Err
}

func (rs *RedisStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	keys, err := rs.Cmd("KEYS", utils.PUBSUB_SUBSCRIBERS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	result = make(map[string]*SubscriberData)
	for _, key := range keys {
		var values []byte
		if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
			if err == redis.ErrRespNil { // did not find the destination
				err = utils.ErrNotFound
			}
			return
		}
		sub := new(SubscriberData)
		if err = rs.ms.Unmarshal(values, sub); err != nil {
			return nil, err
		}
		result[key[len(utils.PUBSUB_SUBSCRIBERS_PREFIX):]] = sub
	}
	return
}

func (rs *RedisStorage) SetSubscriber(key string, sub *SubscriberData) (err error) {
	result, err := rs.ms.Marshal(sub)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.PUBSUB_SUBSCRIBERS_PREFIX+key, result).Err
}

func (rs *RedisStorage) RemoveSubscriber(key string) (err error) {
	err = rs.Cmd("DEL", utils.PUBSUB_SUBSCRIBERS_PREFIX+key).Err
	return
}

func (rs *RedisStorage) SetUser(up *UserProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(up); err != nil {
		return
	}
	return rs.Cmd("SET", utils.USERS_PREFIX+up.GetId(), result).Err
}

func (rs *RedisStorage) GetUser(key string) (up *UserProfile, err error) {
	var values []byte
	if values, err = rs.Cmd("GET", utils.USERS_PREFIX+key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	up = new(UserProfile)
	if err = rs.ms.Unmarshal(values, &up); err != nil {
		return nil, err
	}
	return
}

func (rs *RedisStorage) GetUsers() (result []*UserProfile, err error) {
	keys, err := rs.Cmd("KEYS", utils.USERS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if values, err := rs.Cmd("GET", key).Bytes(); err == nil {
			up := &UserProfile{}
			err = rs.ms.Unmarshal(values, up)
			result = append(result, up)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	return
}

func (rs *RedisStorage) RemoveUser(key string) error {
	return rs.Cmd("DEL", utils.USERS_PREFIX+key).Err
}

func (rs *RedisStorage) GetAlias(key string, skipCache bool, transactionID string) (al *Alias, err error) {
	cacheKey := utils.ALIASES_PREFIX + key
	cCommit := cacheCommit(transactionID)
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			al = x.(*Alias)
			return
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", cacheKey).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cCommit, transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	al = &Alias{Values: make(AliasValues, 0)}
	al.SetId(key)
	if err = rs.ms.Unmarshal(values, &al.Values); err != nil {
		return nil, err
	}
	cache.Set(cacheKey, al, cCommit, transactionID)
	return
}

func (rs *RedisStorage) SetAlias(al *Alias, transactionID string) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(al.Values)
	if err != nil {
		return
	}
	key := utils.ALIASES_PREFIX + al.GetId()
	if err = rs.Cmd("SET", key, result).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetReverseAlias(reverseID string, skipCache bool, transactionID string) (ids []string, err error) {
	key := utils.REVERSE_ALIASES_PREFIX + reverseID
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	if ids, err = rs.Cmd("SMEMBERS", key).List(); err != nil {
		return
	} else if len(ids) == 0 {
		cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		err = utils.ErrNotFound
		return
	}
	cache.Set(key, ids, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetReverseAlias(al *Alias, transactionID string) (err error) {
	for _, value := range al.Values {
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := strings.Join([]string{utils.REVERSE_ALIASES_PREFIX, alias, target, al.Context}, "")
				id := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
				if err = rs.Cmd("SADD", rKey, id).Err; err != nil {
					break
				}
			}
		}
	}
	return
}

func (rs *RedisStorage) RemoveAlias(id string, transactionID string) (err error) {
	key := utils.ALIASES_PREFIX + id
	// get alias for values list
	al, err := rs.GetAlias(id, false, transactionID)
	if err != nil {
		return
	}
	err = rs.Cmd("DEL", key).Err
	if err != nil {
		return err
	}
	cCommit := cacheCommit(transactionID)
	cache.RemKey(key, cCommit, transactionID)

	for _, value := range al.Values {
		tmpKey := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := utils.REVERSE_ALIASES_PREFIX + alias + target + al.Context
				err = rs.Cmd("SREM", rKey, tmpKey).Err
				if err != nil {
					return err
				}
				cache.RemKey(rKey, cCommit, transactionID)
				/*_, err = rs.GetReverseAlias(rKey, true) // recache
				if err != nil {
					return err
				}*/
			}
		}
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (rs *RedisStorage) GetLoadHistory(limit int, skipCache bool, transactionID string) ([]*utils.LoadInstance, error) {
	if limit == 0 {
		return nil, nil
	}

	if !skipCache {
		if x, ok := cache.Get(utils.LOADINST_KEY); ok {
			if x != nil {
				items := x.([]*utils.LoadInstance)
				if len(items) < limit || limit == -1 {
					return items, nil
				}
				return items[:limit], nil
			}
			return nil, utils.ErrNotFound
		}
	}
	if limit != -1 {
		limit -= -1 // Decrease limit to match redis approach on lrange
	}
	marshaleds, err := rs.Cmd("LRANGE", utils.LOADINST_KEY, 0, limit).ListBytes()
	cCommit := cacheCommit(transactionID)
	if err != nil {
		cache.Set(utils.LOADINST_KEY, nil, cCommit, transactionID)
		return nil, err
	}
	loadInsts := make([]*utils.LoadInstance, len(marshaleds))
	for idx, marshaled := range marshaleds {
		var lInst utils.LoadInstance
		err = rs.ms.Unmarshal(marshaled, &lInst)
		if err != nil {
			return nil, err
		}
		loadInsts[idx] = &lInst
	}
	cache.RemKey(utils.LOADINST_KEY, cCommit, transactionID)
	cache.Set(utils.LOADINST_KEY, loadInsts, cCommit, transactionID)
	if len(loadInsts) < limit || limit == -1 {
		return loadInsts, nil
	}
	return loadInsts[:limit], nil
}

// Adds a single load instance to load history
func (rs *RedisStorage) AddLoadHistory(ldInst *utils.LoadInstance, loadHistSize int, transactionID string) error {
	if loadHistSize == 0 { // Load history disabled
		return nil
	}
	marshaled, err := rs.ms.Marshal(&ldInst)
	if err != nil {
		return err
	}
	_, err = guardian.Guardian.Guard(func() (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
		histLen, err := rs.Cmd("LLEN", utils.LOADINST_KEY).Int()
		if err != nil {
			return nil, err
		}
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			if err := rs.Cmd("RPOP", utils.LOADINST_KEY).Err; err != nil {
				return nil, err
			}
		}
		err = rs.Cmd("LPUSH", utils.LOADINST_KEY, marshaled).Err
		return nil, err
	}, 0, utils.LOADINST_KEY)

	cache.RemKey(utils.LOADINST_KEY, cacheCommit(transactionID), transactionID)
	return err
}

func (rs *RedisStorage) GetActionTriggers(key string, skipCache bool, transactionID string) (atrs ActionTriggers, err error) {
	key = utils.ACTION_TRIGGER_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(ActionTriggers), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &atrs); err != nil {
		return
	}
	cache.Set(key, atrs, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetActionTriggers(key string, atrs ActionTriggers, transactionID string) (err error) {
	if len(atrs) == 0 {
		// delete the key
		return rs.Cmd("DEL", utils.ACTION_TRIGGER_PREFIX+key).Err
	}
	var result []byte
	if result, err = rs.ms.Marshal(atrs); err != nil {
		return err
	}
	if err = rs.Cmd("SET", utils.ACTION_TRIGGER_PREFIX+key, result).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) RemoveActionTriggers(key string, transactionID string) (err error) {
	key = utils.ACTION_TRIGGER_PREFIX + key
	err = rs.Cmd("DEL", key).Err
	cache.RemKey(key, cacheCommit(transactionID), transactionID)

	return
}

func (rs *RedisStorage) GetActionPlan(key string, skipCache bool, transactionID string) (ats *ActionPlan, err error) {
	key = utils.ACTION_PLAN_PREFIX + key
	if !skipCache {
		if x, err := cache.GetCloned(key); err != nil {
			if err.Error() != utils.ItemNotFound { // Only consider cache if item was found
				return nil, err
			}
		} else if x == nil { // item was placed nil in cache
			return nil, utils.ErrNotFound
		} else {
			return x.(*ActionPlan), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
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
	if err = rs.ms.Unmarshal(out, &ats); err != nil {
		return
	}
	cache.Set(key, ats, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetActionPlan(key string, ats *ActionPlan, overwrite bool, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	dbKey := utils.ACTION_PLAN_PREFIX + key
	if len(ats.ActionTimings) == 0 {
		// delete the key
		err = rs.Cmd("DEL", dbKey).Err
		cache.RemKey(dbKey, cCommit, transactionID)
		return
	}
	if !overwrite {
		// get existing action plan to merge the account ids
		if existingAts, _ := rs.GetActionPlan(key, true, transactionID); existingAts != nil {
			if ats.AccountIDs == nil && len(existingAts.AccountIDs) > 0 {
				ats.AccountIDs = make(utils.StringMap)
			}
			for accID := range existingAts.AccountIDs {
				ats.AccountIDs[accID] = true
			}
		}
	}
	var result []byte
	if result, err = rs.ms.Marshal(ats); err != nil {
		return
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	return rs.Cmd("SET", dbKey, b.Bytes()).Err
}

func (rs *RedisStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {

	keys, err := rs.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}

	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		ap, err := rs.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		ats[key[len(utils.ACTION_PLAN_PREFIX):]] = ap
	}

	return
}

func (rs *RedisStorage) GetAccountActionPlans(acntID string, skipCache bool, transactionID string) (aPlIDs []string, err error) {
	key := utils.AccountActionPlansPrefix + acntID
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &aPlIDs); err != nil {
		return
	}
	cache.Set(key, aPlIDs, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetAccountActionPlans(acntID string, aPlIDs []string, overwrite bool) (err error) {
	if !overwrite {
		if oldaPlIDs, err := rs.GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return err
		} else {
			for _, oldAPid := range oldaPlIDs {
				if !utils.IsSliceMember(aPlIDs, oldAPid) {
					aPlIDs = append(aPlIDs, oldAPid)
				}
			}
		}
	}
	var result []byte
	if result, err = rs.ms.Marshal(aPlIDs); err != nil {
		return err
	}
	return rs.Cmd("SET", utils.AccountActionPlansPrefix+acntID, result).Err
}

func (rs *RedisStorage) RemAccountActionPlans(acntID string, aPlIDs []string) (err error) {
	key := utils.AccountActionPlansPrefix + acntID
	if len(aPlIDs) == 0 {
		return rs.Cmd("DEL", key).Err
	}
	oldaPlIDs, err := rs.GetAccountActionPlans(acntID, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	for i := 0; i < len(oldaPlIDs); {
		if utils.IsSliceMember(aPlIDs, oldaPlIDs[i]) {
			oldaPlIDs = append(oldaPlIDs[:i], oldaPlIDs[i+1:]...)
			continue // if we have stripped, don't increase index so we can check next element by next run
		}
		i++
	}
	if len(oldaPlIDs) == 0 { // no more elements, remove the reference
		return rs.Cmd("DEL", key).Err
	}
	var result []byte
	if result, err = rs.ms.Marshal(oldaPlIDs); err != nil {
		return err
	}
	return rs.Cmd("SET", key, result).Err
}

func (rs *RedisStorage) PushTask(t *Task) error {
	result, err := rs.ms.Marshal(t)
	if err != nil {
		return err
	}
	return rs.Cmd("RPUSH", utils.TASKS_KEY, result).Err
}

func (rs *RedisStorage) PopTask() (t *Task, err error) {
	var values []byte
	if values, err = rs.Cmd("LPOP", utils.TASKS_KEY).Bytes(); err == nil {
		t = &Task{}
		err = rs.ms.Unmarshal(values, t)
	}
	return
}

func (rs *RedisStorage) GetDerivedChargers(key string, skipCache bool, transactionID string) (dcs *utils.DerivedChargers, err error) {
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.DerivedChargers), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &dcs); err != nil {
		return
	}
	cache.Set(key, dcs, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers, transactionID string) (err error) {
	key = utils.DERIVEDCHARGERS_PREFIX + key
	cCommit := cacheCommit(transactionID)
	if dcs == nil || len(dcs.Chargers) == 0 {
		if err = rs.Cmd("DEL", key).Err; err != nil {
			return
		}
		cache.RemKey(key, cCommit, transactionID)
		return
	}
	var marshaled []byte
	if marshaled, err = rs.ms.Marshal(dcs); err != nil {
		return
	}
	if err = rs.Cmd("SET", key, marshaled).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetCdrStats(cs *CdrStats) error {
	marshaled, err := rs.ms.Marshal(cs)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.CDR_STATS_PREFIX+cs.Id, marshaled).Err
}

func (rs *RedisStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	var values []byte
	if values, err = rs.Cmd("GET", utils.CDR_STATS_PREFIX+key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &cs); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetAllCdrStats() (css []*CdrStats, err error) {
	keys, err := rs.Cmd("KEYS", utils.CDR_STATS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		value, err := rs.Cmd("GET", key).Bytes()
		if err != nil {
			continue
		}
		cs := &CdrStats{}
		err = rs.ms.Unmarshal(value, cs)
		css = append(css, cs)
	}
	return
}

func (rs *RedisStorage) GetResourceProfile(id string, skipCache bool, transactionID string) (rsp *ResourceProfile, err error) {
	key := utils.ResourceProfilesPrefix + id
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ResourceProfile), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &rsp); err != nil {
		return
	}
	for _, fltr := range rsp.Filters {
		if err = fltr.CompileValues(); err != nil {
			return
		}
	}
	cache.Set(key, rsp, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetResourceProfile(rsp *ResourceProfile, transactionID string) error {
	result, err := rs.ms.Marshal(rsp)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.ResourceProfilesPrefix+rsp.ID, result).Err
}

func (rs *RedisStorage) RemoveResourceProfile(id string, transactionID string) (err error) {
	key := utils.ResourceProfilesPrefix + id
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) GetResource(id string, skipCache bool, transactionID string) (r *Resource, err error) {
	key := utils.ResourcesPrefix + id
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Resource), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &r); err != nil {
		return
	}
	cache.Set(key, r, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetResource(r *Resource) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.ResourcesPrefix+r.ID, result).Err
}

func (rs *RedisStorage) RemoveResource(id string, transactionID string) (err error) {
	key := utils.ResourcesPrefix + id
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) GetTiming(id string, skipCache bool, transactionID string) (t *utils.TPTiming, err error) {
	key := utils.TimingsPrefix + id
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.TPTiming), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &t); err != nil {
		return
	}
	cache.Set(key, t, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetTiming(t *utils.TPTiming, transactionID string) error {
	result, err := rs.ms.Marshal(t)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.TimingsPrefix+t.ID, result).Err
}

func (rs *RedisStorage) RemoveTiming(id string, transactionID string) (err error) {
	key := utils.TimingsPrefix + id
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) GetReqFilterIndexes(dbKey string) (indexes map[string]map[string]utils.StringMap, err error) {
	mp, err := rs.Cmd("HGETALL", dbKey).Map()
	if err != nil {
		return
	} else if len(mp) == 0 {
		return nil, utils.ErrNotFound
	}
	indexes = make(map[string]map[string]utils.StringMap)
	for k, v := range mp {
		var sm utils.StringMap
		if err = rs.ms.Unmarshal([]byte(v), &sm); err != nil {
			return
		}
		kSplt := strings.Split(k, utils.CONCATENATED_KEY_SEP)
		if len(kSplt) != 2 {
			return nil, fmt.Errorf("Malformed key in db: %s", k)
		}
		if _, hasKey := indexes[kSplt[0]]; !hasKey {
			indexes[kSplt[0]] = make(map[string]utils.StringMap)
		}
		if _, hasKey := indexes[kSplt[0]][kSplt[1]]; !hasKey {
			indexes[kSplt[0]][kSplt[1]] = make(utils.StringMap)
		}
		indexes[kSplt[0]][kSplt[1]] = sm
	}
	return
}

func (rs *RedisStorage) SetReqFilterIndexes(dbKey string, indexes map[string]map[string]utils.StringMap) (err error) {
	if err = rs.Cmd("DEL", dbKey).Err; err != nil { // DELETE before set
		return
	}
	mp := make(map[string]string)
	for fldName, fldValMp := range indexes {
		for fldVal, strMp := range fldValMp {
			if encodedMp, err := rs.ms.Marshal(strMp); err != nil {
				return err
			} else {
				mp[utils.ConcatenatedKey(fldName, fldVal)] = string(encodedMp)
			}
		}
	}
	return rs.Cmd("HMSET", dbKey, mp).Err
}

func (rs *RedisStorage) MatchReqFilterIndex(dbKey, fldName, fldVal string) (itemIDs utils.StringMap, err error) {
	fieldValKey := utils.ConcatenatedKey(fldName, fldVal)
	cacheKey := dbKey + fieldValKey
	if x, ok := cache.Get(cacheKey); ok { // Attempt to find in cache first
		if x == nil {
			return nil, utils.ErrNotFound
		}
		return x.(utils.StringMap), nil
	}
	// Not found in cache, check in DB
	fldValBytes, err := rs.Cmd("HGET", dbKey, fieldValKey).Bytes()
	if err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			cache.Set(cacheKey, nil, true, utils.NonTransactional)
			err = utils.ErrNotFound
		}
		return nil, err
	} else if err = rs.ms.Unmarshal(fldValBytes, &itemIDs); err != nil {
		return
	}
	cache.Set(cacheKey, itemIDs, true, utils.NonTransactional)
	return
}

func (rs *RedisStorage) GetVersions(itm string) (vrs Versions, err error) {
	x, err := rs.Cmd("HGETALL", itm).Map()
	if err != nil {
		return nil, err
	}
	vrs, err = utils.MapStringToInt64(x)
	if err != nil {
		return nil, err
	}
	if len(vrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (rs *RedisStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	if overwrite {
		if err = rs.RemoveVersions(vrs); err != nil {
			return
		}
	}
	return rs.Cmd("HMSET", utils.TBLVersions, vrs).Err
}

func (rs *RedisStorage) RemoveVersions(vrs Versions) (err error) {
	for key, _ := range vrs {
		err = rs.Cmd("HDEL", utils.TBLVersions, key).Err
		if err != nil {
			return err
		}
	}

	return
}

// GetStatsConfig retrieves a StatsConfig from dataDB
func (rs *RedisStorage) GetStatQueueProfile(sqID string) (sq *StatQueueProfile, err error) {
	key := utils.StatQueueProfilePrefix + sqID
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil {
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &sq); err != nil {
		return
	}
	for _, fltr := range sq.Filters {
		if err = fltr.CompileValues(); err != nil {
			return
		}
	}
	return
}

// SetStatsQueue stores a StatsQueue into DataDB
func (rs *RedisStorage) SetStatQueueProfile(sq *StatQueueProfile) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(sq)
	if err != nil {
		return
	}
	return rs.Cmd("SET", utils.StatQueueProfilePrefix+sq.ID, result).Err
}

// RemStatsQueue removes a StatsQueue from dataDB
func (rs *RedisStorage) RemStatQueueProfile(sqID string) (err error) {
	key := utils.StatQueueProfilePrefix + sqID
	err = rs.Cmd("DEL", key).Err
	return
}

// GetStoredStatQueue retrieves the stored metrics for a StatsQueue
func (rs *RedisStorage) GetStoredStatQueue(tenant, id string) (sq *StoredStatQueue, err error) {
	key := utils.StatQueuePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil {
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &sq); err != nil {
		return
	}
	return
}

// SetStoredStatQueue stores the metrics for a StatsQueue
func (rs *RedisStorage) SetStoredStatQueue(sq *StoredStatQueue) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(sq)
	if err != nil {
		return
	}
	return rs.Cmd("SET", utils.StatQueuePrefix+sq.SqID(), result).Err
}

// RemStatQueue removes a StatsQueue
func (rs *RedisStorage) RemStoredStatQueue(tenant, id string) (err error) {
	key := utils.StatQueuePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

// GetThresholdCfg retrieves a ThresholdCfg from dataDB/cache
func (rs *RedisStorage) GetThresholdCfg(ID string, skipCache bool, transactionID string) (th *ThresholdCfg, err error) {
	key := utils.ThresholdCfgPrefix + ID
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ThresholdCfg), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &th); err != nil {
		return
	}
	for _, fltr := range th.Filters {
		if err = fltr.CompileValues(); err != nil {
			return
		}
	}
	cache.Set(key, th, cacheCommit(transactionID), transactionID)
	return
}

// SetThresholdCfg stores a ThresholdCfg into DataDB
func (rs *RedisStorage) SetThresholdCfg(th *ThresholdCfg) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(th)
	if err != nil {
		return
	}
	return rs.Cmd("SET", utils.ThresholdCfgPrefix+th.ID, result).Err
}

// RemThresholdCfg removes a ThresholdCfg from dataDB/cache
func (rs *RedisStorage) RemThresholdCfg(ID string, transactionID string) (err error) {
	key := utils.ThresholdCfgPrefix + ID
	err = rs.Cmd("DEL", key).Err
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) GetStorageType() string {
	return utils.REDIS
}
