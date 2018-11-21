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
	"strconv"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/sentinel"
)

type RedisStorage struct {
	dbPool        *pool.Pool
	maxConns      int
	ms            Marshaler
	cacheCfg      config.CacheCfg
	sentinelName  string
	sentinelInsts []*sentinelInst
	db            int    //database number used when recconect sentinel
	pass          string //password used when recconect sentinel
	sentinelMux   sync.RWMutex
}

type sentinelInst struct {
	addr string
	conn *sentinel.Client
}

func NewRedisStorage(address string, db int, pass, mrshlerStr string,
	maxConns int, cacheCfg config.CacheCfg, sentinelName string) (*RedisStorage, error) {

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

	var mrshler Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}

	if sentinelName != "" {
		var err error
		addrs := strings.Split(address, ";")
		sentinelInsts := make([]*sentinelInst, len(addrs))
		for i, addr := range addrs {
			sentinelInsts[i] = &sentinelInst{addr: addr}
			if sentinelInsts[i].conn, err = sentinel.NewClientCustom("tcp",
				addr, maxConns, df, sentinelName); err != nil {
				return nil, err
			}
		}
		return &RedisStorage{maxConns: maxConns, ms: mrshler,
			cacheCfg: cacheCfg, sentinelName: sentinelName,
			sentinelInsts: sentinelInsts, db: db, pass: pass}, nil
	} else {
		p, err := pool.NewCustom("tcp", address, maxConns, df)
		if err != nil {
			return nil, err
		}
		return &RedisStorage{dbPool: p, maxConns: maxConns,
			ms: mrshler, cacheCfg: cacheCfg}, nil
	}
}

func reconnectSentinel(addr, sentinelName string, db int, pass string, maxConns int) (*sentinel.Client, error) {
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
	return sentinel.NewClientCustom("tcp", addr, maxConns, df, sentinelName)
}

// This CMD function get a connection from the pool.
// Handles automatic failover in case of network disconnects
func (rs *RedisStorage) Cmd(cmd string, args ...interface{}) *redis.Resp {
	if rs.sentinelName != "" {
		var err error
		for i := range rs.sentinelInsts {
			rs.sentinelMux.Lock()

			if rs.sentinelInsts[i].conn == nil {
				rs.sentinelInsts[i].conn, err = reconnectSentinel(rs.sentinelInsts[i].addr,
					rs.sentinelName, rs.db, rs.pass, rs.maxConns)
				if err != nil {
					if i == len(rs.sentinelInsts)-1 {
						rs.sentinelMux.Unlock()
						return redis.NewResp(fmt.Errorf("No sentinels active"))
					}
					rs.sentinelMux.Unlock()
					continue
				}
			}
			sConn := rs.sentinelInsts[i].conn
			rs.sentinelMux.Unlock()

			conn, err := sConn.GetMaster(rs.sentinelName)
			if err != nil {
				if i == len(rs.sentinelInsts)-1 {
					return redis.NewResp(fmt.Errorf("No sentinels active"))
				}
				rs.sentinelMux.Lock()
				rs.sentinelInsts[i].conn = nil
				rs.sentinelMux.Unlock()
				utils.Logger.Warning(fmt.Sprintf("<RedisStorage> sentinel at address: %s became nil error: %s ",
					rs.sentinelInsts[i].addr, err.Error()))
				continue
			}
			result := conn.Cmd(cmd, args...)
			sConn.PutMaster(rs.sentinelName, conn)
			return result
		}
	}

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
	if rs.dbPool != nil {
		rs.dbPool.Empty()
	}
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

func (rs *RedisStorage) IsDBEmpty() (resp bool, err error) {
	var keys []string
	keys, err = rs.GetKeysForPrefix("")
	if err != nil {
		return
	}
	if len(keys) != 0 {
		return false, nil
	}
	return true, nil
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

func (rs *RedisStorage) RemoveReverseForPrefix(prefix string) (err error) {
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
			if err := rs.RemoveDestination(dest.Id, utils.NonTransactional); err != nil {
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
			if err := rs.RemoveAlias(al.GetId(), utils.NonTransactional); err != nil {
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
				if err = rs.RemAccountActionPlans(acntID, []string{apl.Id}); err != nil {
					return err
				}
			}
		}
	}
	return nil
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
func (rs *RedisStorage) HasDataDrv(category, subject, tenant string) (bool, error) {
	switch category {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX,
		utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX, utils.DERIVEDCHARGERS_PREFIX:
		i, err := rs.Cmd("EXISTS", category+subject).Int()
		return i == 1, err
	case utils.ResourcesPrefix, utils.ResourceProfilesPrefix, utils.StatQueuePrefix,
		utils.StatQueueProfilePrefix, utils.ThresholdPrefix, utils.ThresholdProfilePrefix,
		utils.FilterPrefix, utils.SupplierProfilePrefix, utils.AttributeProfilePrefix, utils.ChargerProfilePrefix:
		i, err := rs.Cmd("EXISTS", category+utils.ConcatenatedKey(tenant, subject)).Int()
		return i == 1, err
	}
	return false, errors.New("unsupported HasData category")
}

func (rs *RedisStorage) GetRatingPlanDrv(key string) (rp *RatingPlan, err error) {
	key = utils.RATING_PLAN_PREFIX + key
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
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
	err = rs.ms.Unmarshal(out, &rp)
	if err != nil {
		return nil, err
	}
	return
}

func (rs *RedisStorage) SetRatingPlanDrv(rp *RatingPlan) (err error) {
	result, err := rs.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.Cmd("SET", utils.RATING_PLAN_PREFIX+rp.Id, b.Bytes()).Err
	return
}

func (rs *RedisStorage) RemoveRatingPlanDrv(key string) error {
	keys, err := rs.Cmd("KEYS", utils.RATING_PLAN_PREFIX+key+"*").List()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err = rs.Cmd("DEL", key).Err; err != nil {
			return err
		}
	}
	return nil
}

func (rs *RedisStorage) GetRatingProfileDrv(key string) (rpf *RatingProfile, err error) {
	key = utils.RATING_PROFILE_PREFIX + key
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &rpf); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetRatingProfileDrv(rpf *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rpf)
	if err != nil {
		return err
	}
	key := utils.RATING_PROFILE_PREFIX + rpf.Id
	if err = rs.Cmd("SET", key, result).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) RemoveRatingProfileDrv(key string) error {
	keys, err := rs.Cmd("KEYS", utils.RATING_PROFILE_PREFIX+key+"*").List()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err = rs.Cmd("DEL", key).Err; err != nil {
			return err
		}
	}
	return nil
}

// GetDestination retrieves a destination with id from  tp_db
func (rs *RedisStorage) GetDestination(key string, skipCache bool,
	transactionID string) (dest *Destination, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheDestinations, key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Destination), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", utils.DESTINATION_PREFIX+key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			Cache.Set(utils.CacheDestinations, key, nil, nil,
				cacheCommit(transactionID), transactionID)
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
	Cache.Set(utils.CacheDestinations, key, dest, nil,
		cacheCommit(transactionID), transactionID)
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
	return
}

func (rs *RedisStorage) GetReverseDestination(key string,
	skipCache bool, transactionID string) (ids []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheReverseDestinations, key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	if ids, err = rs.Cmd("SMEMBERS", utils.REVERSE_DESTINATION_PREFIX+key).List(); err != nil {
		return
	} else if len(ids) == 0 {
		Cache.Set(utils.CacheReverseDestinations, key, nil, nil,
			cacheCommit(transactionID), transactionID)
		err = utils.ErrNotFound
		return
	}
	Cache.Set(utils.CacheReverseDestinations, key, ids, nil,
		cacheCommit(transactionID), transactionID)
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
	// get destination for prefix list
	d, err := rs.GetDestination(destID, false, transactionID)
	if err != nil {
		return
	}
	err = rs.Cmd("DEL", utils.DESTINATION_PREFIX+destID).Err
	if err != nil {
		return err
	}
	Cache.Remove(utils.CacheDestinations, destID,
		cacheCommit(transactionID), transactionID)
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
	// remove id for all obsolete prefixes
	cCommit := cacheCommit(transactionID)
	var err error
	for _, obsoletePrefix := range obsoletePrefixes {
		err = rs.Cmd("SREM",
			utils.REVERSE_DESTINATION_PREFIX+obsoletePrefix, oldDest.Id).Err
		if err != nil {
			return err
		}
		Cache.Remove(utils.CacheReverseDestinations, obsoletePrefix,
			cCommit, transactionID)
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

func (rs *RedisStorage) GetActionsDrv(key string) (as Actions, err error) {
	key = utils.ACTION_PREFIX + key
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &as); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetActionsDrv(key string, as Actions) (err error) {
	result, err := rs.ms.Marshal(&as)
	err = rs.Cmd("SET", utils.ACTION_PREFIX+key, result).Err
	return
}

func (rs *RedisStorage) RemoveActionsDrv(key string) (err error) {
	err = rs.Cmd("DEL", utils.ACTION_PREFIX+key).Err
	return
}

func (rs *RedisStorage) GetSharedGroupDrv(key string) (sg *SharedGroup, err error) {
	key = utils.SHARED_GROUP_PREFIX + key
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &sg); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetSharedGroupDrv(sg *SharedGroup) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sg); err != nil {
		return
	}
	err = rs.Cmd("SET", utils.SHARED_GROUP_PREFIX+sg.Id, result).Err
	return
}

func (rs *RedisStorage) RemoveSharedGroupDrv(id,
	transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	err = rs.Cmd("DEL", utils.SHARED_GROUP_PREFIX+id).Err
	Cache.Remove(utils.CacheSharedGroups, id, cCommit,
		transactionID)
	return err
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
	err = rs.Cmd("DEL", utils.ACCOUNT_PREFIX+key).Err
	if err == redis.ErrRespNil {
		err = utils.ErrNotFound
	}
	return
}

func (rs *RedisStorage) GetSubscribersDrv() (result map[string]*SubscriberData, err error) {
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

func (rs *RedisStorage) SetSubscriberDrv(key string, sub *SubscriberData) (err error) {
	result, err := rs.ms.Marshal(sub)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.PUBSUB_SUBSCRIBERS_PREFIX+key, result).Err
}

func (rs *RedisStorage) RemoveSubscriberDrv(key string) (err error) {
	err = rs.Cmd("DEL", utils.PUBSUB_SUBSCRIBERS_PREFIX+key).Err
	return
}

func (rs *RedisStorage) SetUserDrv(up *UserProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(up); err != nil {
		return
	}
	return rs.Cmd("SET", utils.USERS_PREFIX+up.GetId(), result).Err
}

func (rs *RedisStorage) GetUserDrv(key string) (up *UserProfile, err error) {
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

func (rs *RedisStorage) GetUsersDrv() (result []*UserProfile, err error) {
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

func (rs *RedisStorage) RemoveUserDrv(key string) error {
	return rs.Cmd("DEL", utils.USERS_PREFIX+key).Err
}

func (rs *RedisStorage) GetAlias(key string, skipCache bool,
	transactionID string) (al *Alias, err error) {
	cCommit := cacheCommit(transactionID)
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheAliases, key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			al = x.(*Alias)
			return
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET",
		utils.ALIASES_PREFIX+key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			Cache.Set(utils.CacheAliases, key, nil, nil,
				cCommit, transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	al = &Alias{Values: make(AliasValues, 0)}
	al.SetId(key)
	if err = rs.ms.Unmarshal(values, &al.Values); err != nil {
		return nil, err
	}
	Cache.Set(utils.CacheAliases, key, al, nil,
		cCommit, transactionID)
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

func (rs *RedisStorage) GetReverseAlias(reverseID string, skipCache bool,
	transactionID string) (ids []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheReverseAliases, reverseID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	if ids, err = rs.Cmd("SMEMBERS",
		utils.REVERSE_ALIASES_PREFIX+reverseID).List(); err != nil {
		return
	} else if len(ids) == 0 {
		Cache.Set(utils.CacheReverseAliases, reverseID, nil, nil,
			cacheCommit(transactionID), transactionID)
		err = utils.ErrNotFound
		return
	}
	Cache.Set(utils.CacheReverseAliases, reverseID, ids, nil,
		cacheCommit(transactionID), transactionID)
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
	// get alias for values list
	al, err := rs.GetAlias(id, false, transactionID)
	if err != nil {
		return
	}
	err = rs.Cmd("DEL", utils.ALIASES_PREFIX+id).Err
	if err != nil {
		return err
	}
	cCommit := cacheCommit(transactionID)
	Cache.Remove(utils.CacheAliases, id, cCommit, transactionID)
	for _, value := range al.Values {
		tmpKey := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				revID := alias + target + al.Context
				err = rs.Cmd("SREM", utils.REVERSE_ALIASES_PREFIX+revID, tmpKey).Err
				if err != nil {
					return err
				}
				Cache.Remove(utils.CacheReverseAliases, revID,
					cCommit, transactionID)
			}
		}
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (rs *RedisStorage) GetLoadHistory(limit int, skipCache bool,
	transactionID string) ([]*utils.LoadInstance, error) {
	if limit == 0 {
		return nil, nil
	}

	if !skipCache {
		if x, ok := Cache.Get(utils.LOADINST_KEY, ""); ok {
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
	marshaleds, err := rs.Cmd("LRANGE",
		utils.LOADINST_KEY, 0, limit).ListBytes()
	cCommit := cacheCommit(transactionID)
	if err != nil {
		Cache.Set(utils.LOADINST_KEY, "", nil, nil,
			cCommit, transactionID)
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
	Cache.Remove(utils.LOADINST_KEY, "", cCommit, transactionID)
	Cache.Set(utils.LOADINST_KEY, "", loadInsts, nil,
		cCommit, transactionID)
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

	Cache.Remove(utils.LOADINST_KEY, "",
		cacheCommit(transactionID), transactionID)
	return err
}

func (rs *RedisStorage) GetActionTriggersDrv(key string) (atrs ActionTriggers, err error) {
	key = utils.ACTION_TRIGGER_PREFIX + key
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &atrs); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetActionTriggersDrv(key string, atrs ActionTriggers) (err error) {
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

func (rs *RedisStorage) RemoveActionTriggersDrv(key string) (err error) {
	key = utils.ACTION_TRIGGER_PREFIX + key
	err = rs.Cmd("DEL", key).Err
	return
}

func (rs *RedisStorage) GetActionPlan(key string, skipCache bool,
	transactionID string) (ats *ActionPlan, err error) {
	if !skipCache {
		if x, err := Cache.GetCloned(utils.CacheActionPlans, key); err != nil {
			if err != ltcache.ErrNotFound { // Only consider cache if item was found
				return nil, err
			}
		} else if x == nil { // item was placed nil in cache
			return nil, utils.ErrNotFound
		} else {
			return x.(*ActionPlan), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET", utils.ACTION_PLAN_PREFIX+key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			Cache.Set(utils.CacheActionPlans, key, nil, nil,
				cacheCommit(transactionID), transactionID)
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
	Cache.Set(utils.CacheActionPlans, key, ats, nil,
		cacheCommit(transactionID), transactionID)
	return
}
func (rs *RedisStorage) RemoveActionPlan(key string,
	transactionID string) error {
	cCommit := cacheCommit(transactionID)
	err := rs.Cmd("DEL", utils.ACTION_PLAN_PREFIX+key).Err
	Cache.Remove(utils.CacheActionPlans, key,
		cCommit, transactionID)
	return err
}

func (rs *RedisStorage) SetActionPlan(key string, ats *ActionPlan,
	overwrite bool, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if len(ats.ActionTimings) == 0 {
		// delete the key
		err = rs.Cmd("DEL", utils.ACTION_PLAN_PREFIX+key).Err
		Cache.Remove(utils.CacheActionPlans, key,
			cCommit, transactionID)
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
	return rs.Cmd("SET", utils.ACTION_PLAN_PREFIX+key, b.Bytes()).Err
}

func (rs *RedisStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	keys, err := rs.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}
	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		ap, err := rs.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):],
			false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		ats[key[len(utils.ACTION_PLAN_PREFIX):]] = ap
	}
	return
}

func (rs *RedisStorage) GetAccountActionPlans(acntID string, skipCache bool,
	transactionID string) (aPlIDs []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheAccountActionPlans, acntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	var values []byte
	if values, err = rs.Cmd("GET",
		utils.AccountActionPlansPrefix+acntID).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			Cache.Set(utils.CacheAccountActionPlans, acntID, nil, nil,
				cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &aPlIDs); err != nil {
		return
	}
	Cache.Set(utils.CacheAccountActionPlans, acntID, aPlIDs, nil,
		cacheCommit(transactionID), transactionID)
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

func (rs *RedisStorage) GetDerivedChargersDrv(key string) (dcs *utils.DerivedChargers, err error) {
	key = utils.DERIVEDCHARGERS_PREFIX + key
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &dcs); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetDerivedChargers(key string,
	dcs *utils.DerivedChargers, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if dcs == nil || len(dcs.Chargers) == 0 {
		if err = rs.Cmd("DEL", utils.DERIVEDCHARGERS_PREFIX+key).Err; err != nil {
			return
		}
		Cache.Remove(utils.CacheDerivedChargers, key,
			cCommit, transactionID)
		return
	}
	var marshaled []byte
	if marshaled, err = rs.ms.Marshal(dcs); err != nil {
		return
	}
	if err = rs.Cmd("SET", utils.DERIVEDCHARGERS_PREFIX+key, marshaled).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) RemoveDerivedChargersDrv(id, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if err = rs.Cmd("DEL", utils.DERIVEDCHARGERS_PREFIX+id).Err; err != nil {
		return err
	}
	Cache.Remove(utils.CacheDerivedChargers, id, cCommit, transactionID)
	return
}

func (rs *RedisStorage) GetResourceProfileDrv(tenant, id string) (rsp *ResourceProfile, err error) {
	key := utils.ResourceProfilesPrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &rsp); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetResourceProfileDrv(rsp *ResourceProfile) error {
	result, err := rs.ms.Marshal(rsp)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.ResourceProfilesPrefix+rsp.TenantID(), result).Err
}

func (rs *RedisStorage) RemoveResourceProfileDrv(tenant, id string) (err error) {
	key := utils.ResourceProfilesPrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	key := utils.ResourcesPrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &r); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetResourceDrv(r *Resource) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.ResourcesPrefix+r.TenantID(), result).Err
}

func (rs *RedisStorage) RemoveResourceDrv(tenant, id string) (err error) {
	key := utils.ResourcesPrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetTimingDrv(id string) (t *utils.TPTiming, err error) {
	key := utils.TimingsPrefix + id
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &t); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetTimingDrv(t *utils.TPTiming) error {
	result, err := rs.ms.Marshal(t)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.TimingsPrefix+t.ID, result).Err
}

func (rs *RedisStorage) RemoveTimingDrv(id string) (err error) {
	key := utils.TimingsPrefix + id
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

//GetFilterIndexesDrv retrieves Indexes from dataDB
//filterType is used togheter with fieldName:Val
func (rs *RedisStorage) GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	mp := make(map[string]string)
	dbKey := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	if len(fldNameVal) == 0 {
		mp, err = rs.Cmd("HGETALL", dbKey).Map()
		if err != nil {
			return
		} else if len(mp) == 0 {
			return nil, utils.ErrNotFound
		}
	} else {
		var itmMpStrLst []string
		for fldName, fldVal := range fldNameVal {
			concatTypeNameVal := utils.ConcatenatedKey(filterType, fldName, fldVal)
			itmMpStrLst, err = rs.Cmd("HMGET", dbKey, concatTypeNameVal).List()
			if err != nil {
				return
			} else if itmMpStrLst[0] == "" {
				return nil, utils.ErrNotFound
			}
			mp[concatTypeNameVal] = itmMpStrLst[0]
		}
	}
	indexes = make(map[string]utils.StringMap)
	for k, v := range mp {
		var sm utils.StringMap
		if err = rs.ms.Unmarshal([]byte(v), &sm); err != nil {
			return
		}
		if _, hasKey := indexes[k]; !hasKey {
			indexes[k] = make(utils.StringMap)
		}
		indexes[k] = sm
	}
	return
}

//SetFilterIndexesDrv stores Indexes into DataDB
func (rs *RedisStorage) SetFilterIndexesDrv(cacheID, itemIDPrefix string,
	indexes map[string]utils.StringMap, commit bool, transactionID string) (err error) {
	originKey := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	dbKey := originKey
	if transactionID != "" {
		dbKey = "tmp_" + utils.ConcatenatedKey(dbKey, transactionID)
	}
	if commit && transactionID != "" {
		return rs.Cmd("RENAME", dbKey, originKey).Err
	} else {
		mp := make(map[string]string)
		nameValSls := []interface{}{dbKey}
		for key, strMp := range indexes {
			if len(strMp) == 0 { // remove with no more elements inside
				nameValSls = append(nameValSls, key)
				continue
			}
			if encodedMp, err := rs.ms.Marshal(strMp); err != nil {
				return err
			} else {
				mp[key] = string(encodedMp)
			}
		}
		if len(nameValSls) != 1 {
			if err = rs.Cmd("HDEL", nameValSls...).Err; err != nil {
				return err
			}
		}
		if len(mp) != 0 {
			return rs.Cmd("HMSET", dbKey, mp).Err
		}
		return
	}
}

func (rs *RedisStorage) RemoveFilterIndexesDrv(cacheID, itemIDPrefix string) (err error) {
	return rs.Cmd("DEL", utils.CacheInstanceToPrefix[cacheID]+itemIDPrefix).Err
}

func (rs *RedisStorage) MatchFilterIndexDrv(cacheID, itemIDPrefix,
	filterType, fldName, fldVal string) (itemIDs utils.StringMap, err error) {
	fieldValKey := utils.ConcatenatedKey(filterType, fldName, fldVal)
	fldValBytes, err := rs.Cmd("HGET",
		utils.CacheInstanceToPrefix[cacheID]+itemIDPrefix, fieldValKey).Bytes()
	if err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return nil, err
	} else if err = rs.ms.Unmarshal(fldValBytes, &itemIDs); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetVersions(itm string) (vrs Versions, err error) {
	if itm != "" {
		fldVal, err := rs.Cmd("HGET", utils.TBLVersions, itm).Str()
		if err != nil {
			if err == redis.ErrRespNil {
				err = utils.ErrNotFound
			}
			return nil, err
		}
		intVal, err := strconv.ParseInt(fldVal, 10, 64)
		if err != nil {
			return nil, err
		}
		return Versions{itm: intVal}, nil
	}
	mp, err := rs.Cmd("HGETALL", utils.TBLVersions).Map()
	if err != nil {
		return nil, err
	}
	vrs, err = utils.MapStringToInt64(mp)
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
		if err = rs.RemoveVersions(nil); err != nil {
			return
		}
	}
	return rs.Cmd("HMSET", utils.TBLVersions, vrs).Err
}

func (rs *RedisStorage) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) != 0 {
		for key := range vrs {
			err = rs.Cmd("HDEL", utils.TBLVersions, key).Err
			if err != nil {
				return err
			}
		}
		return
	}
	return rs.Cmd("DEL", utils.TBLVersions).Err
}

// GetStatQueueProfileDrv retrieves a StatQueueProfile from dataDB
func (rs *RedisStorage) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	key := utils.StatQueueProfilePrefix + utils.ConcatenatedKey(tenant, id)
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

// SetStatsQueueDrv stores a StatsQueue into DataDB
func (rs *RedisStorage) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	result, err := rs.ms.Marshal(sq)
	if err != nil {
		return
	}
	return rs.Cmd("SET", utils.StatQueueProfilePrefix+utils.ConcatenatedKey(sq.Tenant, sq.ID), result).Err
}

// RemStatsQueueDrv removes a StatsQueue from dataDB
func (rs *RedisStorage) RemStatQueueProfileDrv(tenant, id string) (err error) {
	key := utils.StatQueueProfilePrefix + utils.ConcatenatedKey(tenant, id)
	err = rs.Cmd("DEL", key).Err
	return
}

// GetStoredStatQueue retrieves the stored metrics for a StatsQueue
func (rs *RedisStorage) GetStoredStatQueueDrv(tenant, id string) (sq *StoredStatQueue, err error) {
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
func (rs *RedisStorage) SetStoredStatQueueDrv(sq *StoredStatQueue) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(sq)
	if err != nil {
		return
	}
	return rs.Cmd("SET", utils.StatQueuePrefix+sq.SqID(), result).Err
}

// RemoveStatQueue removes a StatsQueue
func (rs *RedisStorage) RemStoredStatQueueDrv(tenant, id string) (err error) {
	key := utils.StatQueuePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (rs *RedisStorage) GetThresholdProfileDrv(tenant, ID string) (tp *ThresholdProfile, err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, ID)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil {
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &tp); err != nil {
		return
	}
	return
}

// SetThresholdProfileDrv stores a ThresholdProfile into DataDB
func (rs *RedisStorage) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(tp)
	if err != nil {
		return
	}
	return rs.Cmd("SET", utils.ThresholdProfilePrefix+tp.TenantID(), result).Err
}

// RemoveThresholdProfile removes a ThresholdProfile from dataDB/cache
func (rs *RedisStorage) RemThresholdProfileDrv(tenant, id string) (err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, id)
	err = rs.Cmd("DEL", key).Err
	return
}

func (rs *RedisStorage) GetThresholdDrv(tenant, id string) (r *Threshold, err error) {
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &r); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetThresholdDrv(r *Threshold) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.ThresholdPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result).Err
}

func (rs *RedisStorage) RemoveThresholdDrv(tenant, id string) (err error) {
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetFilterDrv(tenant, id string) (r *Filter, err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &r); err != nil {
		return
	}
	for _, fltr := range r.Rules {
		if err = fltr.CompileValues(); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStorage) SetFilterDrv(r *Filter) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.FilterPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result).Err
}

func (rs *RedisStorage) RemoveFilterDrv(tenant, id string) (err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetSupplierProfileDrv(tenant, id string) (r *SupplierProfile, err error) {
	key := utils.SupplierProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &r); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetSupplierProfileDrv(r *SupplierProfile) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.SupplierProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result).Err
}

func (rs *RedisStorage) RemoveSupplierProfileDrv(tenant, id string) (err error) {
	key := utils.SupplierProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetAttributeProfileDrv(tenant, id string) (r *AttributeProfile, err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &r); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetAttributeProfileDrv(r *AttributeProfile) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.AttributeProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result).Err
}

func (rs *RedisStorage) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetChargerProfileDrv(tenant, id string) (r *ChargerProfile, err error) {
	key := utils.ChargerProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd("GET", key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &r); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) SetChargerProfileDrv(r *ChargerProfile) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd("SET", utils.ChargerProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result).Err
}

func (rs *RedisStorage) RemoveChargerProfileDrv(tenant, id string) (err error) {
	key := utils.ChargerProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd("DEL", key).Err; err != nil {
		return
	}
	return
}

func (rs *RedisStorage) GetStorageType() string {
	return utils.REDIS
}
