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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/mediocregopher/radix/v3"
)

type RedisStorage struct {
	client radix.Client
	ms     Marshaler
}

// Redis commands
const (
	redisAUTH     = "AUTH"
	redisSELECT   = "SELECT"
	redisFLUSHDB  = "FLUSHDB"
	redisDEL      = "DEL"
	redisHGETALL  = "HGETALL"
	redisKEYS     = "KEYS"
	redisSADD     = "SADD"
	redisSMEMBERS = "SMEMBERS"
	redisSREM     = "SREM"
	redisEXISTS   = "EXISTS"
	redisGET      = "GET"
	redisSET      = "SET"
	redisLRANGE   = "LRANGE"
	redisLLEN     = "LLEN"
	redisRPOP     = "RPOP"
	redisLPUSH    = "LPUSH"
	redisRPUSH    = "RPUSH"
	redisLPOP     = "LPOP"
	redisHMGET    = "HMGET"
	redisHDEL     = "HDEL"
	redisHGET     = "HGET"
	redisRENAME   = "RENAME"
	redisHMSET    = "HMSET"
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
			if ca, err = os.ReadFile(tlsCACert); err != nil {
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
	return rs.Cmd(nil, redisFLUSHDB)
}

func (rs *RedisStorage) Marshaler() Marshaler {
	return rs.ms
}

func (rs *RedisStorage) SelectDatabase(dbName string) (err error) {
	return rs.Cmd(nil, redisSELECT, dbName)
}

func (rs *RedisStorage) IsDBEmpty() (resp bool, err error) {
	var keys []string
	keys, err = rs.GetKeysForPrefix(context.TODO(), "")
	if err != nil {
		return
	}
	if len(keys) != 0 {
		return false, nil
	}
	return true, nil
}

func (rs *RedisStorage) RemoveKeysForPrefix(prefix string) (err error) {
	var keys []string
	if keys, err = rs.GetKeysForPrefix(context.TODO(), prefix); err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd(nil, redisDEL, key); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStorage) getKeysForFilterIndexesKeys(fkeys []string) (keys []string, err error) {
	for _, itemIDPrefix := range fkeys {
		mp := make(map[string]string)
		if err = rs.Cmd(&mp, redisHGETALL, itemIDPrefix); err != nil {
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

func (rs *RedisStorage) GetKeysForPrefix(ctx *context.Context, prefix string) (keys []string, err error) {
	err = rs.Cmd(&keys, redisKEYS, prefix+"*")
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
func (rs *RedisStorage) HasDataDrv(ctx *context.Context, category, subject, tenant string) (exists bool, err error) {
	var i int
	switch category {
	case utils.ResourcesPrefix, utils.ResourceProfilesPrefix, utils.StatQueuePrefix,
		utils.StatQueueProfilePrefix, utils.ThresholdPrefix, utils.ThresholdProfilePrefix,
		utils.FilterPrefix, utils.RouteProfilePrefix, utils.AttributeProfilePrefix,
		utils.ChargerProfilePrefix, utils.DispatcherProfilePrefix, utils.DispatcherHostPrefix,
		utils.RateProfilePrefix:
		err := rs.Cmd(&i, redisEXISTS, category+utils.ConcatenatedKey(tenant, subject))
		return i == 1, err
	}
	return false, errors.New("unsupported HasData category")
}

// Limit will only retrieve the last n items out of history, newest first
func (rs *RedisStorage) GetLoadHistory(limit int, skipCache bool,
	transactionID string) (loadInsts []*utils.LoadInstance, err error) {
	if limit == 0 {
		return nil, nil
	}

	if !skipCache {
		if x, ok := Cache.Get(utils.LoadInstKey, ""); ok {
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
	if err = rs.Cmd(&marshaleds, redisLRANGE,
		utils.LoadInstKey, "0", strconv.Itoa(limit)); err != nil {
		if errCh := Cache.Set(context.TODO(), utils.LoadInstKey, "", nil, nil,
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
	if err = Cache.Remove(context.TODO(), utils.LoadInstKey, "", cCommit, transactionID); err != nil {
		return nil, err
	}
	if err := Cache.Set(context.TODO(), utils.LoadInstKey, "", loadInsts, nil,
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
	_, err = guardian.Guardian.Guard(context.TODO(), func(_ *context.Context) (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
		var histLen int
		if err := rs.Cmd(&histLen, redisLLEN, utils.LoadInstKey); err != nil {
			return nil, err
		}
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			if err = rs.Cmd(nil, redisRPOP, utils.LoadInstKey); err != nil {
				return nil, err
			}
		}
		return nil, rs.Cmd(nil, redisLPUSH, utils.LoadInstKey, string(marshaled))
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.LoadInstKey)

	if errCh := Cache.Remove(context.TODO(), utils.LoadInstKey, "",
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	return
}

func (rs *RedisStorage) GetResourceProfileDrv(tenant, id string) (rsp *ResourceProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
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
	return rs.Cmd(nil, redisSET, utils.ResourceProfilesPrefix+rsp.TenantID(), string(result))
}

func (rs *RedisStorage) RemoveResourceProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.ResourcesPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
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
	return rs.Cmd(nil, redisSET, utils.ResourcesPrefix+r.TenantID(), string(result))
}

func (rs *RedisStorage) RemoveResourceDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ResourcesPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetVersions(itm string) (vrs Versions, err error) {
	if itm != "" {
		var fldVal int64
		mn := radix.MaybeNil{Rcv: &fldVal}
		if err = rs.Cmd(&mn, redisHGET, utils.TBLVersions, itm); err != nil {
			return nil, err
		} else if mn.Nil {
			err = utils.ErrNotFound
			return
		}
		return Versions{itm: fldVal}, nil
	}
	var mp map[string]string
	if err = rs.Cmd(&mp, redisHGETALL, utils.TBLVersions); err != nil {
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
	return rs.FlatCmd(nil, redisHMSET, utils.TBLVersions, vrs)
}

func (rs *RedisStorage) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) != 0 {
		for key := range vrs {
			if err = rs.Cmd(nil, redisHDEL, utils.TBLVersions, key); err != nil {
				return
			}
		}
		return
	}
	return rs.Cmd(nil, redisDEL, utils.TBLVersions)
}

// GetStatQueueProfileDrv retrieves a StatQueueProfile from dataDB
func (rs *RedisStorage) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
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
	return rs.Cmd(nil, redisSET, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(sq.Tenant, sq.ID), string(result))
}

// RemStatQueueProfileDrv removes a StatsQueue from dataDB
func (rs *RedisStorage) RemStatQueueProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

// GetStatQueueDrv retrieves the stored metrics for a StatsQueue
func (rs *RedisStorage) GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
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
	if ssq == nil {
		if ssq, err = NewStoredStatQueue(sq, rs.ms); err != nil {
			return
		}
	}
	var result []byte
	if result, err = rs.ms.Marshal(ssq); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.StatQueuePrefix+ssq.SqID(), string(result))
}

// RemStatQueueDrv removes a StatsQueue
func (rs *RedisStorage) RemStatQueueDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id))
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (rs *RedisStorage) GetThresholdProfileDrv(tenant, ID string) (tp *ThresholdProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.ThresholdProfilePrefix+utils.ConcatenatedKey(tenant, ID)); err != nil {
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
	return rs.Cmd(nil, redisSET, utils.ThresholdProfilePrefix+tp.TenantID(), string(result))
}

// RemThresholdProfileDrv removes a ThresholdProfile from dataDB/cache
func (rs *RedisStorage) RemThresholdProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ThresholdProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetThresholdDrv(tenant, id string) (r *Threshold, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.ThresholdPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
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
	return rs.Cmd(nil, redisSET, utils.ThresholdPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveThresholdDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ThresholdPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetFilterDrv(ctx *context.Context, tenant, id string) (r *Filter, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.FilterPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetFilterDrv(ctx *context.Context, r *Filter) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.FilterPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveFilterDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.FilterPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetRouteProfileDrv(tenant, id string) (r *RouteProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.RouteProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
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
	return rs.Cmd(nil, redisSET, utils.RouteProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveRouteProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.RouteProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetAttributeProfileDrv(ctx *context.Context, tenant, id string) (r *AttributeProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.AttributeProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetAttributeProfileDrv(ctx *context.Context, r *AttributeProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.AttributeProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveAttributeProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.AttributeProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetChargerProfileDrv(tenant, id string) (r *ChargerProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.ChargerProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
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
	return rs.Cmd(nil, redisSET, utils.ChargerProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveChargerProfileDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ChargerProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetDispatcherProfileDrv(ctx *context.Context, tenant, id string) (r *DispatcherProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.DispatcherProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetDispatcherProfileDrv(ctx *context.Context, r *DispatcherProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.DispatcherProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveDispatcherProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.DispatcherProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetDispatcherHostDrv(tenant, id string) (r *DispatcherHost, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.DispatcherHostPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
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
	return rs.Cmd(nil, redisSET, utils.DispatcherHostPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveDispatcherHostDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.DispatcherHostPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetStorageType() string {
	return utils.Redis
}

func (rs *RedisStorage) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	if itemIDPrefix != "" {
		var fldVal int64
		mn := radix.MaybeNil{Rcv: &fldVal}
		if err = rs.Cmd(&mn, redisHGET, utils.LoadIDs, itemIDPrefix); err != nil {
			return
		} else if mn.Nil {
			err = utils.ErrNotFound
			return
		}
		return map[string]int64{itemIDPrefix: fldVal}, nil
	}
	mpLoadIDs := make(map[string]string)
	if err = rs.Cmd(&mpLoadIDs, redisHGETALL, utils.LoadIDs); err != nil {
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

func (rs *RedisStorage) SetLoadIDsDrv(ctx *context.Context, loadIDs map[string]int64) error {
	return rs.FlatCmd(nil, redisHMSET, utils.LoadIDs, loadIDs)
}

func (rs *RedisStorage) RemoveLoadIDsDrv() (err error) {
	return rs.Cmd(nil, redisDEL, utils.LoadIDs)
}

func (rs *RedisStorage) GetRateProfileDrv(ctx *context.Context, tenant, id string) (rpp *utils.RateProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.RateProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &rpp)
	return
}

func (rs *RedisStorage) SetRateProfileDrv(ctx *context.Context, rpp *utils.RateProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rpp); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.RateProfilePrefix+utils.ConcatenatedKey(rpp.Tenant, rpp.ID), string(result))
}

func (rs *RedisStorage) RemoveRateProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.RateProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetActionProfileDrv(ctx *context.Context, tenant, id string) (ap *ActionProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.ActionProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &ap)
	return
}

func (rs *RedisStorage) SetActionProfileDrv(ctx *context.Context, ap *ActionProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(ap); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.ActionProfilePrefix+utils.ConcatenatedKey(ap.Tenant, ap.ID), string(result))
}

func (rs *RedisStorage) RemoveActionProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ActionProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

// GetIndexesDrv retrieves Indexes from dataDB
func (rs *RedisStorage) GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
	mp := make(map[string]string)
	dbKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	if len(idxKey) == 0 {
		if err = rs.Cmd(&mp, redisHGETALL, dbKey); err != nil {
			return
		} else if len(mp) == 0 {
			return nil, utils.ErrNotFound
		}
	} else {
		var itmMpStrLst []string
		if err = rs.Cmd(&itmMpStrLst, redisHMGET, dbKey, idxKey); err != nil {
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
func (rs *RedisStorage) SetIndexesDrv(ctx *context.Context, idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	originKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	dbKey := originKey
	if transactionID != utils.EmptyString {
		dbKey = "tmp_" + utils.ConcatenatedKey(dbKey, transactionID)
	}
	if commit && transactionID != utils.EmptyString {
		return rs.Cmd(nil, redisRENAME, dbKey, originKey)
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
		if err = rs.Cmd(nil, redisHDEL, deleteArgs...); err != nil {
			return
		}
	}
	if len(mp) != 0 {
		return rs.FlatCmd(nil, redisHMSET, dbKey, mp)
	}
	return
}

func (rs *RedisStorage) RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error) {
	if idxKey == utils.EmptyString {
		return rs.Cmd(nil, redisDEL, utils.CacheInstanceToPrefix[idxItmType]+tntCtx)
	}
	return rs.Cmd(nil, redisHDEL, utils.CacheInstanceToPrefix[idxItmType]+tntCtx, idxKey)
}

func (rs *RedisStorage) GetAccountDrv(ctx *context.Context, tenant, id string) (ap *utils.Account, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.AccountPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &ap)
	return
}

func (rs *RedisStorage) SetAccountDrv(ctx *context.Context, ap *utils.Account) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(ap); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.AccountPrefix+utils.ConcatenatedKey(ap.Tenant, ap.ID), string(result))
}

func (rs *RedisStorage) RemoveAccountDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.AccountPrefix+utils.ConcatenatedKey(tenant, id))
}
