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
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix/v3"
)

type RedisStoragev3 struct {
	client radix.Client
	ms     Marshaler
}

// Redis commands
const (
	redis_AUTH     = "AUTH"
	redis_SELECT   = "SELECT"
	redis_FLUSHDB  = "FLUSHDB"
	redis_DEL      = "DEL"
	redis_HGETALL  = "HGETALL"
	redis_KEYS     = "KEYS"
	redis_SADD     = "SADD"
	redis_SMEMBERS = "SMEMBERS"
	redis_SREM     = "SREM"
	redis_EXISTS   = "EXISTS"
	redis_GET      = "GET"
	redis_SET      = "SET"
	redis_LRANGE   = "LRANGE"
	redis_LLEN     = "LLEN"
	redis_RPOP     = "RPOP"
	redis_LPUSH    = "LPUSH"
	redis_RPUSH    = "RPUSH"
	redis_LPOP     = "LPOP"
	redis_HMGET    = "HMGET"
	redis_HDEL     = "HDEL"
	redis_HGET     = "HGET"
	redis_RENAME   = "RENAME"
	redis_HMSET    = "HMSET"
)

func NewRedisStoragev3(address string, db int, user, pass, mrshlerStr string,
	maxConns int, sentinelName string) (rs *RedisStoragev3, err error) {

	rs = new(RedisStoragev3)

	if rs.ms, err = NewMarshaler(mrshlerStr); err != nil {
		rs = nil
		return
	}

	dialOpts := []radix.DialOpt{
		radix.DialSelectDB(db),
	}
	if user == utils.EmptyString {
		dialOpts = append(dialOpts, radix.DialAuthPass(pass))
	} else {
		dialOpts = append(dialOpts, radix.DialAuthUser(user, pass))
	}

	dialFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr, dialOpts...)
	}

	switch {
	case sentinelName != utils.EmptyString:
		if rs.client, err = radix.NewSentinel(sentinelName, utils.InfieldSplit(address),
			radix.SentinelConnFunc(dialFunc),
			radix.SentinelPoolFunc(func(network, addr string) (radix.Client, error) {
				return radix.NewPool(network, addr, maxConns, radix.PoolConnFunc(dialFunc))
			})); err != nil {
			rs = nil
			return
		}
	default:
		if rs.client, err = radix.NewPool(utils.TCP, address, maxConns, radix.PoolConnFunc(dialFunc)); err != nil {
			rs = nil
			return
		}
	}

	return
}

// Cmd function get a connection from the pool.
// Handles automatic failover in case of network disconnects
func (rs *RedisStoragev3) Cmd(rcv interface{}, cmd string, args ...string) error {
	return rs.client.Do(radix.Cmd(rcv, cmd, args...))
}

// Cmd function get a connection from the pool.
// Handles automatic failover in case of network disconnects
func (rs *RedisStoragev3) FlatCmd(rcv interface{}, cmd, key string, args ...interface{}) error {
	return rs.client.Do(radix.FlatCmd(rcv, cmd, key, args...))
}

func (rs *RedisStoragev3) Close() {
	if rs.client != nil {
		rs.client.Close()
	}
}

func (rs *RedisStoragev3) Flush(ignore string) error {
	return rs.Cmd(nil, redis_FLUSHDB)
}

func (rs *RedisStoragev3) Marshaler() Marshaler {
	return rs.ms
}

func (rs *RedisStoragev3) SelectDatabase(dbName string) (err error) {
	return rs.Cmd(nil, redis_SELECT, dbName)
}

func (rs *RedisStoragev3) IsDBEmpty() (resp bool, err error) {
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

func (rs *RedisStoragev3) RebuildReverseForPrefix(prefix string) (err error) {
	if !utils.SliceHasMember([]string{utils.AccountActionPlansPrefix, utils.REVERSE_DESTINATION_PREFIX}, prefix) {
		return utils.ErrInvalidKey
	}
	var keys []string
	keys, err = rs.GetKeysForPrefix(prefix)
	if err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd(nil, redis_DEL, key); err != nil {
			return
		}
	}
	switch prefix {
	case utils.REVERSE_DESTINATION_PREFIX:
		if keys, err = rs.GetKeysForPrefix(utils.DESTINATION_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			dest, err := rs.GetDestinationDrv(key[len(utils.DESTINATION_PREFIX):], true, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err = rs.SetReverseDestinationDrv(dest, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.AccountActionPlansPrefix:
		if keys, err = rs.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			apl, err := rs.GetActionPlanDrv(key[len(utils.ACTION_PLAN_PREFIX):], true, utils.NonTransactional) // skipCache on get since loader checks and caches empty data for loaded objects
			if err != nil {
				return err
			}
			for acntID := range apl.AccountIDs {
				if err = rs.SetAccountActionPlansDrv(acntID, []string{apl.Id}, false); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (rs *RedisStoragev3) RemoveReverseForPrefix(prefix string) (err error) {
	if !utils.SliceHasMember([]string{utils.AccountActionPlansPrefix, utils.REVERSE_DESTINATION_PREFIX}, prefix) {
		return utils.ErrInvalidKey
	}
	var keys []string
	keys, err = rs.GetKeysForPrefix(prefix)
	if err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd(nil, redis_DEL, key); err != nil {
			return
		}
	}
	switch prefix {
	case utils.REVERSE_DESTINATION_PREFIX:
		if keys, err = rs.GetKeysForPrefix(utils.DESTINATION_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			dest, err := rs.GetDestinationDrv(key[len(utils.DESTINATION_PREFIX):], true, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err := rs.RemoveDestinationDrv(dest.Id, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.AccountActionPlansPrefix:
		if keys, err = rs.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			apl, err := rs.GetActionPlanDrv(key[len(utils.ACTION_PLAN_PREFIX):], true, utils.NonTransactional) // skipCache on get since loader checks and caches empty data for loaded objects
			if err != nil {
				return err
			}
			for acntID := range apl.AccountIDs {
				if err = rs.RemAccountActionPlansDrv(acntID, []string{apl.Id}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (rs *RedisStoragev3) getKeysForFilterIndexesKeys(fkeys []string) (keys []string, err error) {
	for _, itemIDPrefix := range fkeys {
		mp := make(map[string]string)
		if err = rs.Cmd(&mp, redis_HGETALL, itemIDPrefix); err != nil {
			return
		} else if len(mp) == 0 {
			return nil, utils.ErrNotFound
		}
		for k := range mp {
			keys = append(keys, utils.ConcatenatedKey(itemIDPrefix, k))
		}
	}
	return
}

func (rs *RedisStoragev3) RebbuildActionPlanKeys() (err error) {
	var keys []string
	if err = rs.Cmd(&keys, redis_KEYS, utils.ACTION_PLAN_PREFIX+"*"); err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd(nil, redis_SADD, utils.ActionPlanIndexes, key); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStoragev3) GetKeysForPrefix(prefix string) (keys []string, err error) {
	if prefix == utils.ACTION_PLAN_PREFIX { // so we can avoid the full scan on scheduler reloads
		err = rs.Cmd(&keys, redis_SMEMBERS, utils.ActionPlanIndexes)
	} else {
		err = rs.Cmd(&keys, redis_KEYS, prefix+"*")
	}
	if err != nil {
		return
	}
	if len(keys) != 0 {
		if filterIndexesPrefixMap.Has(prefix) {
			return rs.getKeysForFilterIndexesKeys(keys)
		}
		return
	}
	return nil, nil
}

// Used to check if specific subject is stored using prefix key attached to entity
func (rs *RedisStoragev3) HasDataDrv(category, subject, tenant string) (exists bool, err error) {
	var i int
	switch category {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX,
		utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX:
		err = rs.Cmd(&i, redis_EXISTS, category+subject)
		return i == 1, err
	case utils.ResourcesPrefix, utils.ResourceProfilesPrefix, utils.StatQueuePrefix,
		utils.StatQueueProfilePrefix, utils.ThresholdPrefix, utils.ThresholdProfilePrefix,
		utils.FilterPrefix, utils.RouteProfilePrefix, utils.AttributeProfilePrefix,
		utils.ChargerProfilePrefix, utils.DispatcherProfilePrefix, utils.DispatcherHostPrefix,
		utils.RateProfilePrefix:
		err := rs.Cmd(&i, redis_EXISTS, category+utils.ConcatenatedKey(tenant, subject))
		return i == 1, err
	}
	return false, errors.New("unsupported HasData category")
}

func (rs *RedisStoragev3) GetRatingPlanDrv(key string) (rp *RatingPlan, err error) {
	key = utils.RATING_PLAN_PREFIX + key
	var values []byte
	if err = rs.Cmd(&values, redis_GET, key); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	b := bytes.NewBuffer(values)
	var r io.ReadCloser
	if r, err = zlib.NewReader(b); err != nil {
		return
	}
	var out []byte
	if out, err = ioutil.ReadAll(r); err != nil {
		return
	}
	r.Close()
	err = rs.ms.Unmarshal(out, &rp)
	return
}

func (rs *RedisStoragev3) SetRatingPlanDrv(rp *RatingPlan) (err error) {
	result, err := rs.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.Cmd(nil, redis_SET, utils.RATING_PLAN_PREFIX+rp.Id, b.String())
	return
}

func (rs *RedisStoragev3) RemoveRatingPlanDrv(key string) (err error) {
	var keys []string
	if err = rs.Cmd(&keys, redis_KEYS, utils.RATING_PLAN_PREFIX+key+"*"); err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd(nil, redis_DEL, key); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStoragev3) GetRatingProfileDrv(key string) (rpf *RatingProfile, err error) {
	key = utils.RATING_PROFILE_PREFIX + key
	var values []byte
	if err = rs.Cmd(&values, redis_GET, key); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	err = rs.ms.Unmarshal(values, &rpf)
	return
}

func (rs *RedisStoragev3) SetRatingProfileDrv(rpf *RatingProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rpf); err != nil {
		return
	}
	key := utils.RATING_PROFILE_PREFIX + rpf.Id
	err = rs.Cmd(nil, redis_SET, key, string(result))
	return
}

func (rs *RedisStoragev3) RemoveRatingProfileDrv(key string) (err error) {
	var keys []string
	if err = rs.Cmd(&keys, redis_KEYS, utils.RATING_PROFILE_PREFIX+key+"*"); err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd(nil, redis_DEL, key); err != nil {
			return
		}
	}
	return
}

// GetDestination retrieves a destination with id from  tp_db
func (rs *RedisStoragev3) GetDestinationDrv(key string, skipCache bool,
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
	if err = rs.Cmd(&values, redis_GET, utils.DESTINATION_PREFIX+key); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			if errCh := Cache.Set(utils.CacheDestinations, key, nil, nil,
				cacheCommit(transactionID), transactionID); errCh != nil {
				return nil, errCh
			}
			err = utils.ErrNotFound
		}
		return
	}
	b := bytes.NewBuffer(values)
	var r io.ReadCloser
	if r, err = zlib.NewReader(b); err != nil {
		return
	}
	var out []byte
	if out, err = ioutil.ReadAll(r); err != nil {
		return
	}
	r.Close()
	if err = rs.ms.Unmarshal(out, &dest); err != nil {
		return
	}
	err = Cache.Set(utils.CacheDestinations, key, dest, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStoragev3) SetDestinationDrv(dest *Destination, transactionID string) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(dest); err != nil {
		return
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.Cmd(nil, redis_SET, utils.DESTINATION_PREFIX+dest.Id, b.String())
	return
}

func (rs *RedisStoragev3) GetReverseDestinationDrv(key string,
	skipCache bool, transactionID string) (ids []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheReverseDestinations, key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	if err = rs.Cmd(&ids, redis_SMEMBERS, utils.REVERSE_DESTINATION_PREFIX+key); err != nil {
		return
	}
	if len(ids) == 0 {
		if err = Cache.Set(utils.CacheReverseDestinations, key, nil, nil,
			cacheCommit(transactionID), transactionID); err != nil {
			return
		}
		err = utils.ErrNotFound
		return
	}
	err = Cache.Set(utils.CacheReverseDestinations, key, ids, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStoragev3) SetReverseDestinationDrv(dest *Destination, transactionID string) (err error) {
	for _, p := range dest.Prefixes {
		if err = rs.Cmd(nil, redis_SADD, utils.REVERSE_DESTINATION_PREFIX+p, dest.Id); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStoragev3) RemoveDestinationDrv(destID, transactionID string) (err error) {
	// get destination for prefix list
	var d *Destination
	if d, err = rs.GetDestinationDrv(destID, false, transactionID); err != nil {
		return
	}
	if err = rs.Cmd(nil, redis_DEL, utils.DESTINATION_PREFIX+destID); err != nil {
		return
	}
	if err = Cache.Remove(utils.CacheDestinations, destID,
		cacheCommit(transactionID), transactionID); err != nil {
		return
	}
	if d == nil {
		return utils.ErrNotFound
	}
	for _, prefix := range d.Prefixes {
		if err = rs.Cmd(nil, redis_SREM, utils.REVERSE_DESTINATION_PREFIX+prefix, destID); err != nil {
			return
		}
		rs.GetReverseDestinationDrv(prefix, true, transactionID) // it will recache the destination
	}
	return
}

func (rs *RedisStoragev3) UpdateReverseDestinationDrv(oldDest, newDest *Destination, transactionID string) (err error) {
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
	for _, obsoletePrefix := range obsoletePrefixes {
		if err = rs.Cmd(nil, redis_SREM,
			utils.REVERSE_DESTINATION_PREFIX+obsoletePrefix, oldDest.Id); err != nil {
			return
		}
		if err = Cache.Remove(utils.CacheReverseDestinations, obsoletePrefix,
			cCommit, transactionID); err != nil {
			return
		}
	}

	// add the id to all new prefixes
	for _, addedPrefix := range addedPrefixes {
		if err = rs.Cmd(nil, redis_SADD, utils.REVERSE_DESTINATION_PREFIX+addedPrefix, newDest.Id); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStoragev3) GetActionsDrv(key string) (as Actions, err error) {
	key = utils.ACTION_PREFIX + key
	var values []byte
	if err = rs.Cmd(&values, redis_GET, key); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	err = rs.ms.Unmarshal(values, &as)
	return
}

func (rs *RedisStoragev3) SetActionsDrv(key string, as Actions) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(&as); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ACTION_PREFIX+key, string(result))
}

func (rs *RedisStoragev3) RemoveActionsDrv(key string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ACTION_PREFIX+key)
}

func (rs *RedisStoragev3) GetSharedGroupDrv(key string) (sg *SharedGroup, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.SHARED_GROUP_PREFIX+key); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	err = rs.ms.Unmarshal(values, &sg)
	return
}

func (rs *RedisStoragev3) SetSharedGroupDrv(sg *SharedGroup) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sg); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.SHARED_GROUP_PREFIX+sg.Id, string(result))
}

func (rs *RedisStoragev3) RemoveSharedGroupDrv(id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.SHARED_GROUP_PREFIX+id)
}

func (rs *RedisStoragev3) GetAccountDrv(key string) (ub *Account, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ACCOUNT_PREFIX+key); err != nil {
		return
		// } else if rpl.IsType(redis.Nil) {
		// 	return nil, utils.ErrNotFound
	}
	ub = &Account{ID: key}
	if err = rs.ms.Unmarshal(values, ub); err != nil {
		return nil, err
	}
	return ub, nil
}

func (rs *RedisStoragev3) SetAccountDrv(acc *Account) (err error) {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(acc.BalanceMap) == 0 {
		var ac *Account
		if ac, err = rs.GetAccountDrv(acc.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = acc.ActionTriggers
			ac.UnitCounters = acc.UnitCounters
			ac.AllowNegative = acc.AllowNegative
			ac.Disabled = acc.Disabled
			acc = ac
		}
	}
	acc.UpdateTime = time.Now()
	var result []byte
	if result, err = rs.ms.Marshal(acc); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ACCOUNT_PREFIX+acc.ID, string(result))
}

func (rs *RedisStoragev3) RemoveAccountDrv(key string) (err error) {
	if err = rs.Cmd(nil, redis_DEL, utils.ACCOUNT_PREFIX+key); err == redis.ErrRespNil {
		err = utils.ErrNotFound
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (rs *RedisStoragev3) GetLoadHistory(limit int, skipCache bool,
	transactionID string) (loadInsts []*utils.LoadInstance, err error) {
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
	cCommit := cacheCommit(transactionID)
	var marshaleds [][]byte
	if err = rs.Cmd(&marshaleds, redis_LRANGE,
		utils.LOADINST_KEY, "0", strconv.Itoa(limit)); err != nil {
		if errCh := Cache.Set(utils.LOADINST_KEY, "", nil, nil,
			cCommit, transactionID); errCh != nil {
			return nil, errCh
		}
		return
	}
	loadInsts = make([]*utils.LoadInstance, len(marshaleds))
	for idx, marshaled := range marshaleds {
		if err = rs.ms.Unmarshal(marshaled, loadInsts[idx]); err != nil {
			return nil, err
		}
	}
	if err = Cache.Remove(utils.LOADINST_KEY, "", cCommit, transactionID); err != nil {
		return nil, err
	}
	if err := Cache.Set(utils.LOADINST_KEY, "", loadInsts, nil,
		cCommit, transactionID); err != nil {
		return nil, err
	}
	if len(loadInsts) < limit || limit == -1 {
		return loadInsts, nil
	}
	return loadInsts[:limit], nil
}

// Adds a single load instance to load history
func (rs *RedisStoragev3) AddLoadHistory(ldInst *utils.LoadInstance, loadHistSize int, transactionID string) (err error) {
	if loadHistSize == 0 { // Load history disabled
		return
	}
	var marshaled []byte
	if marshaled, err = rs.ms.Marshal(&ldInst); err != nil {
		return
	}
	_, err = guardian.Guardian.Guard(func() (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
		var histLen int
		if err := rs.Cmd(&histLen, redis_LLEN, utils.LOADINST_KEY); err != nil {
			return nil, err
		}
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			if err = rs.Cmd(nil, redis_RPOP, utils.LOADINST_KEY); err != nil {
				return nil, err
			}
		}
		return nil, rs.Cmd(nil, redis_LPUSH, utils.LOADINST_KEY, string(marshaled))
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.LOADINST_KEY)

	if errCh := Cache.Remove(utils.LOADINST_KEY, "",
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	return
}

func (rs *RedisStoragev3) GetActionTriggersDrv(key string) (atrs ActionTriggers, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ACTION_TRIGGER_PREFIX+key); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	err = rs.ms.Unmarshal(values, &atrs)
	return
}

func (rs *RedisStoragev3) SetActionTriggersDrv(key string, atrs ActionTriggers) (err error) {
	if len(atrs) == 0 {
		// delete the key
		return rs.Cmd(nil, redis_DEL, utils.ACTION_TRIGGER_PREFIX+key)
	}
	var result []byte
	if result, err = rs.ms.Marshal(atrs); err != nil {
		return
	}
	if err = rs.Cmd(nil, redis_SET, utils.ACTION_TRIGGER_PREFIX+key, string(result)); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) RemoveActionTriggersDrv(key string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ACTION_TRIGGER_PREFIX+key)
}

func (rs *RedisStoragev3) GetActionPlanDrv(key string, skipCache bool,
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
	if err = rs.Cmd(&values, redis_GET, utils.ACTION_PLAN_PREFIX+key); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			if errCh := Cache.Set(utils.CacheActionPlans, key, nil, nil,
				cacheCommit(transactionID), transactionID); errCh != nil {
				return nil, errCh
			}
			err = utils.ErrNotFound
		}
		return
	}
	b := bytes.NewBuffer(values)
	var r io.ReadCloser
	if r, err = zlib.NewReader(b); err != nil {
		return
	}
	var out []byte
	if out, err = ioutil.ReadAll(r); err != nil {
		return
	}
	r.Close()
	if err = rs.ms.Unmarshal(out, &ats); err != nil {
		return
	}
	err = Cache.Set(utils.CacheActionPlans, key, ats, nil,
		cacheCommit(transactionID), transactionID)
	return
}
func (rs *RedisStoragev3) RemoveActionPlanDrv(key string,
	transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if err = rs.Cmd(nil, redis_SREM, utils.ActionPlanIndexes, utils.ACTION_PLAN_PREFIX+key); err != nil {
		return
	}
	err = rs.Cmd(nil, redis_DEL, utils.ACTION_PLAN_PREFIX+key)
	if errCh := Cache.Remove(utils.CacheActionPlans, key,
		cCommit, transactionID); errCh != nil {
		return errCh
	}
	return
}

func (rs *RedisStoragev3) SetActionPlanDrv(key string, ats *ActionPlan,
	overwrite bool, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if len(ats.ActionTimings) == 0 {
		// delete the key
		if err = rs.Cmd(nil, redis_SREM, utils.ActionPlanIndexes, utils.ACTION_PLAN_PREFIX+key); err != nil {
			return
		}
		err = rs.Cmd(nil, redis_DEL, utils.ACTION_PLAN_PREFIX+key)
		if errCh := Cache.Remove(utils.CacheActionPlans, key,
			cCommit, transactionID); errCh != nil {
			return errCh
		}
		return
	}
	if !overwrite {
		// get existing action plan to merge the account ids
		if existingAts, _ := rs.GetActionPlanDrv(key, true, transactionID); existingAts != nil {
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
	if err = rs.Cmd(nil, redis_SADD, utils.ActionPlanIndexes, utils.ACTION_PLAN_PREFIX+key); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ACTION_PLAN_PREFIX+key, b.String())
}

func (rs *RedisStoragev3) GetAllActionPlansDrv() (ats map[string]*ActionPlan, err error) {
	var keys []string
	if keys, err = rs.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
		return
	}
	if len(keys) == 0 {
		err = utils.ErrNotFound
		return
	}
	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		if ats[key[len(utils.ACTION_PLAN_PREFIX):]], err = rs.GetActionPlanDrv(key[len(utils.ACTION_PLAN_PREFIX):],
			false, utils.NonTransactional); err != nil {
			return nil, err
		}
	}
	return
}

func (rs *RedisStoragev3) GetAccountActionPlansDrv(acntID string, skipCache bool,
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
	if err = rs.Cmd(&values, redis_GET,
		utils.AccountActionPlansPrefix+acntID); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			if errCh := Cache.Set(utils.CacheAccountActionPlans, acntID, nil, nil,
				cacheCommit(transactionID), transactionID); errCh != nil {
				return nil, errCh
			}
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &aPlIDs); err != nil {
		return
	}
	err = Cache.Set(utils.CacheAccountActionPlans, acntID, aPlIDs, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStoragev3) SetAccountActionPlansDrv(acntID string, aPlIDs []string, overwrite bool) (err error) {
	if !overwrite {
		var oldaPlIDs []string
		if oldaPlIDs, err = rs.GetAccountActionPlansDrv(acntID, true, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return
		}
		for _, oldAPid := range oldaPlIDs {
			if !utils.IsSliceMember(aPlIDs, oldAPid) {
				aPlIDs = append(aPlIDs, oldAPid)
			}
		}
	}
	var result []byte
	if result, err = rs.ms.Marshal(aPlIDs); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.AccountActionPlansPrefix+acntID, string(result))
}

func (rs *RedisStoragev3) RemAccountActionPlansDrv(acntID string, aPlIDs []string) (err error) {
	key := utils.AccountActionPlansPrefix + acntID
	if len(aPlIDs) == 0 {
		return rs.Cmd(nil, redis_DEL, key)
	}
	var oldaPlIDs []string
	if oldaPlIDs, err = rs.GetAccountActionPlansDrv(acntID, true, utils.NonTransactional); err != nil {
		return
	}
	for i := 0; i < len(oldaPlIDs); {
		if utils.IsSliceMember(aPlIDs, oldaPlIDs[i]) {
			oldaPlIDs = append(oldaPlIDs[:i], oldaPlIDs[i+1:]...)
			continue // if we have stripped, don't increase index so we can check next element by next run
		}
		i++
	}
	if len(oldaPlIDs) == 0 { // no more elements, remove the reference
		return rs.Cmd(nil, redis_DEL, key)
	}
	var result []byte
	if result, err = rs.ms.Marshal(oldaPlIDs); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, key, string(result))
}

func (rs *RedisStoragev3) PushTask(t *Task) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(t); err != nil {
		return
	}
	return rs.Cmd(nil, redis_RPUSH, utils.TASKS_KEY, string(result))
}

func (rs *RedisStoragev3) PopTask() (t *Task, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_LPOP, utils.TASKS_KEY); err != nil {
		return
	}
	t = &Task{}
	err = rs.ms.Unmarshal(values, t)
	return
}

func (rs *RedisStoragev3) GetResourceProfileDrv(tenant, id string) (rsp *ResourceProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	err = rs.ms.Unmarshal(values, &rsp)
	return
}

func (rs *RedisStoragev3) SetResourceProfileDrv(rsp *ResourceProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rsp); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ResourceProfilesPrefix+rsp.TenantID(), string(result))
}

func (rs *RedisStoragev3) RemoveResourceProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStoragev3) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ResourcesPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStoragev3) SetResourceDrv(r *Resource) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ResourcesPrefix+r.TenantID(), string(result))
}

func (rs *RedisStoragev3) RemoveResourceDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ResourcesPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStoragev3) GetTimingDrv(id string) (t *utils.TPTiming, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.TimingsPrefix+id); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	err = rs.ms.Unmarshal(values, &t)
	return
}

func (rs *RedisStoragev3) SetTimingDrv(t *utils.TPTiming) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(t); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.TimingsPrefix+t.ID, string(result))
}

func (rs *RedisStoragev3) RemoveTimingDrv(id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.TimingsPrefix+id)
}

func (rs *RedisStoragev3) GetVersions(itm string) (vrs Versions, err error) {
	if itm != "" {
		var fldVal string
		if err = rs.Cmd(&fldVal, redis_HGET, utils.TBLVersions, itm); err != nil {
			if err == redis.ErrRespNil {
				err = utils.ErrNotFound
			}
			return nil, err
		}
		var intVal int64
		if intVal, err = strconv.ParseInt(fldVal, 10, 64); err != nil {
			return nil, err
		}
		return Versions{itm: intVal}, nil
	}
	var mp map[string]string
	if err = rs.Cmd(&mp, redis_HGETALL, utils.TBLVersions); err != nil {
		return nil, err
	}
	if len(mp) == 0 {
		return nil, utils.ErrNotFound
	}
	if vrs, err = utils.MapStringToInt64(mp); err != nil {
		return nil, err
	}
	return
}

func (rs *RedisStoragev3) SetVersions(vrs Versions, overwrite bool) (err error) {
	if overwrite {
		if err = rs.RemoveVersions(nil); err != nil {
			return
		}
	}
	return rs.FlatCmd(nil, redis_HMSET, utils.TBLVersions, vrs)
}

func (rs *RedisStoragev3) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) != 0 {
		for key := range vrs {
			if err = rs.Cmd(nil, redis_HDEL, utils.TBLVersions, key); err != nil {
				return
			}
		}
		return
	}
	return rs.Cmd(nil, redis_DEL, utils.TBLVersions)
}

// GetStatQueueProfileDrv retrieves a StatQueueProfile from dataDB
func (rs *RedisStoragev3) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	key := utils.StatQueueProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
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
func (rs *RedisStoragev3) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	result, err := rs.ms.Marshal(sq)
	if err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(sq.Tenant, sq.ID), result)
}

// RemStatsQueueDrv removes a StatsQueue from dataDB
func (rs *RedisStoragev3) RemStatQueueProfileDrv(tenant, id string) (err error) {
	key := utils.StatQueueProfilePrefix + utils.ConcatenatedKey(tenant, id)
	err = rs.Cmd(nil, redis_DEL, key)
	return
}

// GetStoredStatQueue retrieves the stored metrics for a StatsQueue
func (rs *RedisStoragev3) GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error) {
	key := utils.StatQueuePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
		if err == redis.ErrRespNil {
			err = utils.ErrNotFound
		}
		return
	}
	var ssq StoredStatQueue
	if err = rs.ms.Unmarshal(values, &ssq); err != nil {
		return
	}
	sq, err = ssq.AsStatQueue(rs.ms)
	return
}

// SetStoredStatQueue stores the metrics for a StatsQueue
func (rs *RedisStoragev3) SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(ssq)
	if err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.StatQueuePrefix+ssq.SqID(), result)
}

// RemoveStatQueue removes a StatsQueue
func (rs *RedisStoragev3) RemStatQueueDrv(tenant, id string) (err error) {
	key := utils.StatQueuePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (rs *RedisStoragev3) GetThresholdProfileDrv(tenant, ID string) (tp *ThresholdProfile, err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, ID)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
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
func (rs *RedisStoragev3) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(tp)
	if err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ThresholdProfilePrefix+tp.TenantID(), result)
}

// RemoveThresholdProfile removes a ThresholdProfile from dataDB/cache
func (rs *RedisStoragev3) RemThresholdProfileDrv(tenant, id string) (err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, id)
	err = rs.Cmd(nil, redis_DEL, key)
	return
}

func (rs *RedisStoragev3) GetThresholdDrv(tenant, id string) (r *Threshold, err error) {
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
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

func (rs *RedisStoragev3) SetThresholdDrv(r *Threshold) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.ThresholdPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result)
}

func (rs *RedisStoragev3) RemoveThresholdDrv(tenant, id string) (err error) {
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) GetFilterDrv(tenant, id string) (r *Filter, err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &r); err != nil {
		return nil, err
	}
	return
}

func (rs *RedisStoragev3) SetFilterDrv(r *Filter) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.FilterPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result)
}

func (rs *RedisStoragev3) RemoveFilterDrv(tenant, id string) (err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) GetRouteProfileDrv(tenant, id string) (r *RouteProfile, err error) {
	key := utils.RouteProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
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

func (rs *RedisStoragev3) SetRouteProfileDrv(r *RouteProfile) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.RouteProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result)
}

func (rs *RedisStoragev3) RemoveRouteProfileDrv(tenant, id string) (err error) {
	key := utils.RouteProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) GetAttributeProfileDrv(tenant, id string) (r *AttributeProfile, err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
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

func (rs *RedisStoragev3) SetAttributeProfileDrv(r *AttributeProfile) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.AttributeProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result)
}

func (rs *RedisStoragev3) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) GetChargerProfileDrv(tenant, id string) (r *ChargerProfile, err error) {
	key := utils.ChargerProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
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

func (rs *RedisStoragev3) SetChargerProfileDrv(r *ChargerProfile) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.ChargerProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result)
}

func (rs *RedisStoragev3) RemoveChargerProfileDrv(tenant, id string) (err error) {
	key := utils.ChargerProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) GetDispatcherProfileDrv(tenant, id string) (r *DispatcherProfile, err error) {
	key := utils.DispatcherProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
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

func (rs *RedisStoragev3) SetDispatcherProfileDrv(r *DispatcherProfile) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.DispatcherProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result)
}

func (rs *RedisStoragev3) RemoveDispatcherProfileDrv(tenant, id string) (err error) {
	key := utils.DispatcherProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) GetDispatcherHostDrv(tenant, id string) (r *DispatcherHost, err error) {
	key := utils.DispatcherHostPrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
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

func (rs *RedisStoragev3) SetDispatcherHostDrv(r *DispatcherHost) (err error) {
	result, err := rs.ms.Marshal(r)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.DispatcherHostPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), result)
}

func (rs *RedisStoragev3) RemoveDispatcherHostDrv(tenant, id string) (err error) {
	key := utils.DispatcherHostPrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) GetStorageType() string {
	return utils.REDIS
}

func (rs *RedisStoragev3) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	if itemIDPrefix != "" {
		fldVal, err := rs.Cmd(redis_HGET, utils.LoadIDs, itemIDPrefix).Int64()
		if err != nil {
			if err == redis.ErrRespNil {
				err = utils.ErrNotFound
			}
			return nil, err
		}
		return map[string]int64{itemIDPrefix: fldVal}, nil
	}
	mpLoadIDs, err := rs.Cmd(redis_HGETALL, utils.LoadIDs).Map()
	if err != nil {
		return nil, err
	}
	loadIDs = make(map[string]int64)
	for key, val := range mpLoadIDs {
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		loadIDs[key] = intVal
	}
	if len(loadIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (rs *RedisStoragev3) SetLoadIDsDrv(loadIDs map[string]int64) error {
	return rs.Cmd(nil, redis_HMSET, utils.LoadIDs, loadIDs)
}

func (rs *RedisStoragev3) RemoveLoadIDsDrv() (err error) {
	return rs.Cmd(nil, redis_DEL, utils.LoadIDs)
}

func (rs *RedisStoragev3) GetRateProfileDrv(tenant, id string) (rpp *RateProfile, err error) {
	key := utils.RateProfilePrefix + utils.ConcatenatedKey(tenant, id)
	var values []byte
	if values, err = rs.Cmd(redis_GET, key).Bytes(); err != nil {
		if err == redis.ErrRespNil { // did not find the destination
			err = utils.ErrNotFound
		}
		return
	}
	if err = rs.ms.Unmarshal(values, &rpp); err != nil {
		return
	}
	return
}

func (rs *RedisStoragev3) SetRateProfileDrv(rpp *RateProfile) (err error) {
	result, err := rs.ms.Marshal(rpp)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.RateProfilePrefix+utils.ConcatenatedKey(rpp.Tenant, rpp.ID), result)
}

func (rs *RedisStoragev3) RemoveRateProfileDrv(tenant, id string) (err error) {
	key := utils.RateProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if err = rs.Cmd(nil, redis_DEL, key); err != nil {
		return
	}
	return
}

// GetIndexesDrv retrieves Indexes from dataDB
func (rs *RedisStoragev3) GetIndexesDrv(idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
	mp := make(map[string]string)
	dbKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	if len(idxKey) == 0 {
		mp, err = rs.Cmd(redis_HGETALL, dbKey).Map()
		if err != nil {
			return
		} else if len(mp) == 0 {
			return nil, utils.ErrNotFound
		}
	} else {
		var itmMpStrLst []string
		itmMpStrLst, err = rs.Cmd(redis_HMGET, dbKey, idxKey).List()
		if err != nil {
			return
		} else if itmMpStrLst[0] == utils.EmptyString {
			return nil, utils.ErrNotFound
		}
		mp[idxKey] = itmMpStrLst[0]
	}
	indexes = make(map[string]utils.StringSet)
	for k, v := range mp {
		var sm utils.StringSet
		if err = rs.ms.Unmarshal([]byte(v), &sm); err != nil {
			return
		}
		indexes[k] = sm
	}
	return
}

// SetIndexesDrv stores Indexes into DataDB
func (rs *RedisStoragev3) SetIndexesDrv(idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	originKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	dbKey := originKey
	if transactionID != utils.EmptyString {
		dbKey = "tmp_" + utils.ConcatenatedKey(dbKey, transactionID)
	}
	if commit && transactionID != utils.EmptyString {
		return rs.Cmd(redis_RENAME, dbKey, originKey)
	}
	mp := make(map[string]string)
	deleteArgs := []interface{}{dbKey} // the dbkey is necesary for the HDEL command
	for key, strMp := range indexes {
		if len(strMp) == 0 { // remove with no more elements inside
			deleteArgs = append(deleteArgs, key)
			continue
		}
		var encodedMp []byte
		if encodedMp, err = rs.ms.Marshal(strMp); err != nil {
			return
		}
		mp[key] = string(encodedMp)
	}
	if len(deleteArgs) != 1 {
		if err = rs.Cmd(nil, redis_HDEL, deleteArgs...); err != nil {
			return
		}
	}
	if len(mp) != 0 {
		return rs.Cmd(nil, redis_HMSET, dbKey, mp)
	}
	return
}

func (rs *RedisStoragev3) RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error) {
	if idxKey == utils.EmptyString {
		return rs.Cmd(nil, redis_DEL, utils.CacheInstanceToPrefix[idxItmType]+tntCtx)
	}
	return rs.Cmd(nil, redis_HDEL, utils.CacheInstanceToPrefix[idxItmType]+tntCtx, idxKey)
}
