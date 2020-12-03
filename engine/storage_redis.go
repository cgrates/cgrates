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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/mediocregopher/radix/v3"
)

type RedisStorage struct {
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

func NewRedisStorage(address string, db int, user, pass, mrshlerStr string,
	maxConns int, sentinelName string, isCluster bool, clusterSync,
	clusterOnDownDelay time.Duration, tlsConn bool,
	tlsClientCert, tlsClientKey, tlsCACert string) (rs *RedisStorage, err error) {
	rs = new(RedisStorage)
	if rs.ms, err = NewMarshaler(mrshlerStr); err != nil {
		rs = nil
		return
	}

	dialOpts := []radix.DialOpt{
		radix.DialSelectDB(db),
	}
	if pass != utils.EmptyString {
		if user == utils.EmptyString {
			dialOpts = append(dialOpts, radix.DialAuthPass(pass))
		} else {
			dialOpts = append(dialOpts, radix.DialAuthUser(user, pass))
		}
	}

	if tlsConn {
		var cert tls.Certificate
		if tlsClientCert != "" && tlsClientKey != "" {
			cert, err = tls.LoadX509KeyPair(tlsClientCert, tlsClientKey)
			if err != nil {
				return
			}
		}
		var rootCAs *x509.CertPool
		if rootCAs, err = x509.SystemCertPool(); err != nil {
			return
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		if tlsCACert != "" {
			var ca []byte
			if ca, err = ioutil.ReadFile(tlsCACert); err != nil {
				return
			}

			if ok := rootCAs.AppendCertsFromPEM(ca); !ok {
				return
			}
		}
		dialOpts = append(dialOpts, radix.DialUseTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      rootCAs,
		}))
	}

	dialFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr, dialOpts...)
	}
	dialFuncAuthOnly := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr, dialOpts[1:]...)
	}
	switch {
	case isCluster:
		if rs.client, err = radix.NewCluster(utils.InfieldSplit(address),
			radix.ClusterSyncEvery(clusterSync),
			radix.ClusterOnDownDelayActionsBy(clusterOnDownDelay),
			radix.ClusterPoolFunc(func(network, addr string) (radix.Client, error) {
				// in cluster enviorment do not select the DB as we expect to have only one DB
				return radix.NewPool(network, addr, maxConns, radix.PoolConnFunc(dialFuncAuthOnly))
			})); err != nil {
			rs = nil
			return
		}
	case sentinelName != utils.EmptyString:
		if rs.client, err = radix.NewSentinel(sentinelName, utils.InfieldSplit(address),
			radix.SentinelConnFunc(dialFuncAuthOnly),
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
func (rs *RedisStorage) Cmd(rcv interface{}, cmd string, args ...string) error {
	return rs.client.Do(radix.Cmd(rcv, cmd, args...))
}

// FlatCmd function get a connection from the pool.
// Handles automatic failover in case of network disconnects
func (rs *RedisStorage) FlatCmd(rcv interface{}, cmd, key string, args ...interface{}) error {
	return rs.client.Do(radix.FlatCmd(rcv, cmd, key, args...))
}

func (rs *RedisStorage) Close() {
	if rs.client != nil {
		rs.client.Close()
	}
}

func (rs *RedisStorage) Flush(ignore string) error {
	return rs.Cmd(nil, redis_FLUSHDB)
}

func (rs *RedisStorage) Marshaler() Marshaler {
	return rs.ms
}

func (rs *RedisStorage) SelectDatabase(dbName string) (err error) {
	return rs.Cmd(nil, redis_SELECT, dbName)
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

func (rs *RedisStorage) RemoveReverseForPrefix(prefix string) (err error) {
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

func (rs *RedisStorage) getKeysForFilterIndexesKeys(fkeys []string) (keys []string, err error) {
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

func (rs *RedisStorage) RebbuildActionPlanKeys() (err error) {
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

func (rs *RedisStorage) GetKeysForPrefix(prefix string) (keys []string, err error) {
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
func (rs *RedisStorage) HasDataDrv(category, subject, tenant string) (exists bool, err error) {
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

func (rs *RedisStorage) GetRatingPlanDrv(key string) (rp *RatingPlan, err error) {
	key = utils.RATING_PLAN_PREFIX + key
	var values []byte
	if err = rs.Cmd(&values, redis_GET, key); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
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

func (rs *RedisStorage) SetRatingPlanDrv(rp *RatingPlan) (err error) {
	result, err := rs.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.Cmd(nil, redis_SET, utils.RATING_PLAN_PREFIX+rp.Id, b.String())
	return
}

func (rs *RedisStorage) RemoveRatingPlanDrv(key string) (err error) {
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

func (rs *RedisStorage) GetRatingProfileDrv(key string) (rpf *RatingProfile, err error) {
	key = utils.RATING_PROFILE_PREFIX + key
	var values []byte
	if err = rs.Cmd(&values, redis_GET, key); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &rpf)
	return
}

func (rs *RedisStorage) SetRatingProfileDrv(rpf *RatingProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rpf); err != nil {
		return
	}
	key := utils.RATING_PROFILE_PREFIX + rpf.Id
	err = rs.Cmd(nil, redis_SET, key, string(result))
	return
}

func (rs *RedisStorage) RemoveRatingProfileDrv(key string) (err error) {
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
func (rs *RedisStorage) GetDestinationDrv(key string, skipCache bool,
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
		return
	} else if len(values) == 0 {
		if errCh := Cache.Set(utils.CacheDestinations, key, nil, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
		err = utils.ErrNotFound
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

func (rs *RedisStorage) SetDestinationDrv(dest *Destination, transactionID string) (err error) {
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

func (rs *RedisStorage) GetReverseDestinationDrv(key string,
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

func (rs *RedisStorage) SetReverseDestinationDrv(dest *Destination, transactionID string) (err error) {
	for _, p := range dest.Prefixes {
		if err = rs.Cmd(nil, redis_SADD, utils.REVERSE_DESTINATION_PREFIX+p, dest.Id); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStorage) RemoveDestinationDrv(destID, transactionID string) (err error) {
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

func (rs *RedisStorage) UpdateReverseDestinationDrv(oldDest, newDest *Destination, transactionID string) (err error) {
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

func (rs *RedisStorage) GetActionsDrv(key string) (as Actions, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ACTION_PREFIX+key); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &as)
	return
}

func (rs *RedisStorage) SetActionsDrv(key string, as Actions) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(&as); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ACTION_PREFIX+key, string(result))
}

func (rs *RedisStorage) RemoveActionsDrv(key string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ACTION_PREFIX+key)
}

func (rs *RedisStorage) GetSharedGroupDrv(key string) (sg *SharedGroup, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.SHARED_GROUP_PREFIX+key); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &sg)
	return
}

func (rs *RedisStorage) SetSharedGroupDrv(sg *SharedGroup) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sg); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.SHARED_GROUP_PREFIX+sg.Id, string(result))
}

func (rs *RedisStorage) RemoveSharedGroupDrv(id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.SHARED_GROUP_PREFIX+id)
}

func (rs *RedisStorage) GetAccountDrv(key string) (ub *Account, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ACCOUNT_PREFIX+key); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	ub = &Account{ID: key}
	if err = rs.ms.Unmarshal(values, ub); err != nil {
		return nil, err
	}
	return ub, nil
}

func (rs *RedisStorage) SetAccountDrv(acc *Account) (err error) {
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

func (rs *RedisStorage) RemoveAccountDrv(key string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ACCOUNT_PREFIX+key)
}

// Limit will only retrieve the last n items out of history, newest first
func (rs *RedisStorage) GetLoadHistory(limit int, skipCache bool,
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
func (rs *RedisStorage) AddLoadHistory(ldInst *utils.LoadInstance, loadHistSize int, transactionID string) (err error) {
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

func (rs *RedisStorage) GetActionTriggersDrv(key string) (atrs ActionTriggers, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ACTION_TRIGGER_PREFIX+key); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &atrs)
	return
}

func (rs *RedisStorage) SetActionTriggersDrv(key string, atrs ActionTriggers) (err error) {
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

func (rs *RedisStorage) RemoveActionTriggersDrv(key string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ACTION_TRIGGER_PREFIX+key)
}

func (rs *RedisStorage) GetActionPlanDrv(key string, skipCache bool,
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
		return
	} else if len(values) == 0 {
		if errCh := Cache.Set(utils.CacheActionPlans, key, nil, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
		err = utils.ErrNotFound
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
func (rs *RedisStorage) RemoveActionPlanDrv(key string,
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

func (rs *RedisStorage) SetActionPlanDrv(key string, ats *ActionPlan,
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

func (rs *RedisStorage) GetAllActionPlansDrv() (ats map[string]*ActionPlan, err error) {
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

func (rs *RedisStorage) GetAccountActionPlansDrv(acntID string, skipCache bool,
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
		return
	} else if len(values) == 0 {
		if errCh := Cache.Set(utils.CacheAccountActionPlans, acntID, nil, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
		err = utils.ErrNotFound
		return
	}
	if err = rs.ms.Unmarshal(values, &aPlIDs); err != nil {
		return
	}
	err = Cache.Set(utils.CacheAccountActionPlans, acntID, aPlIDs, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (rs *RedisStorage) SetAccountActionPlansDrv(acntID string, aPlIDs []string, overwrite bool) (err error) {
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

func (rs *RedisStorage) RemAccountActionPlansDrv(acntID string, aPlIDs []string) (err error) {
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

func (rs *RedisStorage) PushTask(t *Task) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(t); err != nil {
		return
	}
	return rs.Cmd(nil, redis_RPUSH, utils.TASKS_KEY, string(result))
}

func (rs *RedisStorage) PopTask() (t *Task, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_LPOP, utils.TASKS_KEY); err != nil {
		return
	}
	t = &Task{}
	err = rs.ms.Unmarshal(values, t)
	return
}

func (rs *RedisStorage) GetResourceProfileDrv(tenant, id string) (rsp *ResourceProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &rsp)
	return
}

func (rs *RedisStorage) SetResourceProfileDrv(rsp *ResourceProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rsp); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ResourceProfilesPrefix+rsp.TenantID(), string(result))
}

func (rs *RedisStorage) RemoveResourceProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ResourcesPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetResourceDrv(r *Resource) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ResourcesPrefix+r.TenantID(), string(result))
}

func (rs *RedisStorage) RemoveResourceDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ResourcesPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetTimingDrv(id string) (t *utils.TPTiming, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.TimingsPrefix+id); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &t)
	return
}

func (rs *RedisStorage) SetTimingDrv(t *utils.TPTiming) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(t); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.TimingsPrefix+t.ID, string(result))
}

func (rs *RedisStorage) RemoveTimingDrv(id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.TimingsPrefix+id)
}

func (rs *RedisStorage) GetVersions(itm string) (vrs Versions, err error) {
	if itm != "" {
		var fldVal int64
		mn := radix.MaybeNil{Rcv: &fldVal}
		if err = rs.Cmd(&mn, redis_HGET, utils.TBLVersions, itm); err != nil {
			return nil, err
		} else if mn.Nil {
			err = utils.ErrNotFound
			return
		}
		return Versions{itm: fldVal}, nil
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

func (rs *RedisStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	if overwrite {
		if err = rs.RemoveVersions(nil); err != nil {
			return
		}
	}
	return rs.FlatCmd(nil, redis_HMSET, utils.TBLVersions, vrs)
}

func (rs *RedisStorage) RemoveVersions(vrs Versions) (err error) {
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
func (rs *RedisStorage) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &sq)
	return
}

// SetStatQueueProfileDrv stores a StatsQueue into DataDB
func (rs *RedisStorage) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sq); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(sq.Tenant, sq.ID), string(result))
}

// RemStatQueueProfileDrv removes a StatsQueue from dataDB
func (rs *RedisStorage) RemStatQueueProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

// GetStatQueueDrv retrieves the stored metrics for a StatsQueue
func (rs *RedisStorage) GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	var ssq StoredStatQueue
	if err = rs.ms.Unmarshal(values, &ssq); err != nil {
		return
	}
	sq, err = ssq.AsStatQueue(rs.ms)
	return
}

// SetStatQueueDrv stores the metrics for a StatsQueue
func (rs *RedisStorage) SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(ssq); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.StatQueuePrefix+ssq.SqID(), string(result))
}

// RemStatQueueDrv removes a StatsQueue
func (rs *RedisStorage) RemStatQueueDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id))
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (rs *RedisStorage) GetThresholdProfileDrv(tenant, ID string) (tp *ThresholdProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ThresholdProfilePrefix+utils.ConcatenatedKey(tenant, ID)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &tp)
	return
}

// SetThresholdProfileDrv stores a ThresholdProfile into DataDB
func (rs *RedisStorage) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(tp); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ThresholdProfilePrefix+tp.TenantID(), string(result))
}

// RemThresholdProfileDrv removes a ThresholdProfile from dataDB/cache
func (rs *RedisStorage) RemThresholdProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ThresholdProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetThresholdDrv(tenant, id string) (r *Threshold, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ThresholdPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetThresholdDrv(r *Threshold) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ThresholdPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveThresholdDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ThresholdPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetFilterDrv(tenant, id string) (r *Filter, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.FilterPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetFilterDrv(r *Filter) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.FilterPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveFilterDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.FilterPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetRouteProfileDrv(tenant, id string) (r *RouteProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.RouteProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetRouteProfileDrv(r *RouteProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.RouteProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveRouteProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.RouteProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetAttributeProfileDrv(tenant, id string) (r *AttributeProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.AttributeProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetAttributeProfileDrv(r *AttributeProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.AttributeProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.AttributeProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetChargerProfileDrv(tenant, id string) (r *ChargerProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ChargerProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetChargerProfileDrv(r *ChargerProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ChargerProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveChargerProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ChargerProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetDispatcherProfileDrv(tenant, id string) (r *DispatcherProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.DispatcherProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetDispatcherProfileDrv(r *DispatcherProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.DispatcherProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveDispatcherProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.DispatcherProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetDispatcherHostDrv(tenant, id string) (r *DispatcherHost, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.DispatcherHostPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetDispatcherHostDrv(r *DispatcherHost) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.DispatcherHostPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveDispatcherHostDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.DispatcherHostPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetStorageType() string {
	return utils.REDIS
}

func (rs *RedisStorage) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	if itemIDPrefix != "" {
		var fldVal int64
		mn := radix.MaybeNil{Rcv: &fldVal}
		if err = rs.Cmd(&mn, redis_HGET, utils.LoadIDs, itemIDPrefix); err != nil {
			return
		} else if mn.Nil {
			err = utils.ErrNotFound
			return
		}
		return map[string]int64{itemIDPrefix: fldVal}, nil
	}
	mpLoadIDs := make(map[string]string)
	if err = rs.Cmd(&mpLoadIDs, redis_HGETALL, utils.LoadIDs); err != nil {
		return
	}
	if len(mpLoadIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	loadIDs = make(map[string]int64)
	for key, val := range mpLoadIDs {
		if loadIDs[key], err = strconv.ParseInt(val, 10, 64); err != nil {
			return nil, err
		}
	}
	return
}

func (rs *RedisStorage) SetLoadIDsDrv(loadIDs map[string]int64) error {
	return rs.FlatCmd(nil, redis_HMSET, utils.LoadIDs, loadIDs)
}

func (rs *RedisStorage) RemoveLoadIDsDrv() (err error) {
	return rs.Cmd(nil, redis_DEL, utils.LoadIDs)
}

func (rs *RedisStorage) GetRateProfileDrv(tenant, id string) (rpp *RateProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.RateProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &rpp)
	return
}

func (rs *RedisStorage) SetRateProfileDrv(rpp *RateProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rpp); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.RateProfilePrefix+utils.ConcatenatedKey(rpp.Tenant, rpp.ID), string(result))
}

func (rs *RedisStorage) RemoveRateProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.RateProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetActionProfileDrv(tenant, id string) (ap *ActionProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ActionProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &ap)
	return
}

func (rs *RedisStorage) SetActionProfileDrv(ap *ActionProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(ap); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ActionProfilePrefix+utils.ConcatenatedKey(ap.Tenant, ap.ID), string(result))
}

func (rs *RedisStorage) RemoveActionProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ActionProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

// GetIndexesDrv retrieves Indexes from dataDB
func (rs *RedisStorage) GetIndexesDrv(idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
	mp := make(map[string]string)
	dbKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	if len(idxKey) == 0 {
		if err = rs.Cmd(&mp, redis_HGETALL, dbKey); err != nil {
			return
		} else if len(mp) == 0 {
			return nil, utils.ErrNotFound
		}
	} else {
		var itmMpStrLst []string
		if err = rs.Cmd(&itmMpStrLst, redis_HMGET, dbKey, idxKey); err != nil {
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
func (rs *RedisStorage) SetIndexesDrv(idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	originKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	dbKey := originKey
	if transactionID != utils.EmptyString {
		dbKey = "tmp_" + utils.ConcatenatedKey(dbKey, transactionID)
	}
	if commit && transactionID != utils.EmptyString {
		return rs.Cmd(nil, redis_RENAME, dbKey, originKey)
	}
	mp := make(map[string]string)
	deleteArgs := []string{dbKey} // the dbkey is necesary for the HDEL command
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
		return rs.FlatCmd(nil, redis_HMSET, dbKey, mp)
	}
	return
}

func (rs *RedisStorage) RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error) {
	if idxKey == utils.EmptyString {
		return rs.Cmd(nil, redis_DEL, utils.CacheInstanceToPrefix[idxItmType]+tntCtx)
	}
	return rs.Cmd(nil, redis_HDEL, utils.CacheInstanceToPrefix[idxItmType]+tntCtx, idxKey)
}
