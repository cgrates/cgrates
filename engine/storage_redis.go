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
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
	"github.com/mediocregopher/radix/v3"
)

type RedisStorage struct {
	client radix.Client
	ms     utils.Marshaler
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
	redisSCAN     = "SCAN"
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
	redisHSET     = "HSET"
	redisHSCAN    = "HSCAN"

	redisLoadError = "Redis is loading the dataset in memory"
	RedisLimit     = 524287 // https://github.com/StackExchange/StackExchange.Redis/issues/201#issuecomment-98639005
)

func NewRedisStorage(address string, db int, user, pass, mrshlerStr string,
	maxConns, attempts int, sentinelName string, isCluster bool, clusterSync,
	clusterOnDownDelay, connTimeout, readTimeout, writeTimeout time.Duration,
	pipelineWindow time.Duration, pipelineLimit int,
	tlsConn bool, tlsClientCert, tlsClientKey, tlsCACert string) (_ *RedisStorage, err error) {
	var ms utils.Marshaler
	if ms, err = utils.NewMarshaler(mrshlerStr); err != nil {
		return
	}

	dialOpts := make([]radix.DialOpt, 1, 6)
	dialOpts[0] = radix.DialSelectDB(db)
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
			if cert, err = tls.LoadX509KeyPair(tlsClientCert, tlsClientKey); err != nil {
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
			if !rootCAs.AppendCertsFromPEM(ca) {
				return
			}
		}
		dialOpts = append(dialOpts, radix.DialUseTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      rootCAs,
		}))
	}

	dialOpts = append(dialOpts,
		radix.DialReadTimeout(readTimeout),
		radix.DialWriteTimeout(writeTimeout),
		radix.DialConnectTimeout(connTimeout))

	var client radix.Client
	if client, err = newRedisClient(address, sentinelName,
		isCluster, clusterSync, clusterOnDownDelay,
		pipelineWindow, pipelineLimit,
		maxConns, attempts, dialOpts); err != nil {
		return
	}
	return &RedisStorage{
		ms:     ms,
		client: client,
	}, nil
}

func redisDial(network, addr string, attempts int, opts ...radix.DialOpt) (conn radix.Conn, err error) {
	fib := utils.FibDuration(time.Millisecond, 0)
	for i := 0; i < attempts; i++ {
		if conn, err = radix.Dial(network, addr, opts...); err == nil ||
			err != nil && !strings.Contains(err.Error(), redisLoadError) {
			break
		}
		time.Sleep(fib())
	}
	return
}

func newRedisClient(address, sentinelName string, isCluster bool, clusterSync, clusterOnDownDelay,
	pipelineWindow time.Duration, pipelineLimit, maxConns, attempts int, dialOpts []radix.DialOpt,
) (radix.Client, error) {
	dialFunc := func(network, addr string) (radix.Conn, error) {
		return redisDial(network, addr, attempts, dialOpts...)
	}
	dialFuncAuthOnly := func(network, addr string) (radix.Conn, error) {
		return redisDial(network, addr, attempts, dialOpts[1:]...)
	}

	// Configure common pool options.
	poolOpts := make([]radix.PoolOpt, 0, 2)
	poolOpts = append(poolOpts, radix.PoolPipelineWindow(pipelineWindow, pipelineLimit))

	switch {
	case isCluster:
		return radix.NewCluster(utils.InfieldSplit(address),
			radix.ClusterSyncEvery(clusterSync),
			radix.ClusterOnDownDelayActionsBy(clusterOnDownDelay),
			radix.ClusterPoolFunc(func(network, addr string) (radix.Client, error) {
				// in cluster enviorment do not select the DB as we expect to have only one DB
				return radix.NewPool(network, addr, maxConns, append(poolOpts, radix.PoolConnFunc(dialFuncAuthOnly))...)
			}))
	case sentinelName != utils.EmptyString:
		return radix.NewSentinel(sentinelName, utils.InfieldSplit(address),
			radix.SentinelConnFunc(dialFuncAuthOnly),
			radix.SentinelPoolFunc(func(network, addr string) (radix.Client, error) {
				return radix.NewPool(network, addr, maxConns, append(poolOpts, radix.PoolConnFunc(dialFunc))...)
			}))
	default:
		return radix.NewPool(utils.TCP, address, maxConns, append(poolOpts, radix.PoolConnFunc(dialFunc))...)
	}
}

// Cmd function get a connection from the pool.
// Handles automatic failover in case of network disconnects
func (rs *RedisStorage) Cmd(rcv any, cmd string, args ...string) error {
	return rs.client.Do(radix.Cmd(rcv, cmd, args...))
}

// FlatCmd function get a connection from the pool.
// Handles automatic failover in case of network disconnects
func (rs *RedisStorage) FlatCmd(rcv any, cmd, key string, args ...any) error {
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

func (rs *RedisStorage) Marshaler() utils.Marshaler {
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

// func (rs *RedisStorage) GetKeysForPrefix(ctx *context.Context, prefix string) (keys []string, err error) {
// 	err = rs.Cmd(&keys, redisSCAN, "0", "MATCH", prefix+"*")
// 	if err != nil {
// 		return
// 	}
// 	if len(keys) != 0 {
// 		if filterIndexesPrefixMap.Has(prefix) {
// 			return rs.getKeysForFilterIndexesKeys(keys)
// 		}
// 		return
// 	}
// 	return nil, nil
// }

func (rs *RedisStorage) GetKeysForPrefix(ctx *context.Context, prefix string) (keys []string, err error) {
	scan := radix.NewScanner(rs.client, radix.ScanOpts{
		Command: redisSCAN,
		Pattern: prefix + utils.Meta,
	})
	var key string
	for scan.Next(&key) {
		keys = append(keys, key)
	}
	if err = scan.Close(); err != nil {
		return nil, err
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
		utils.ChargerProfilePrefix,
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
	err = guardian.Guardian.Guard(context.TODO(), func(_ *context.Context) error { // Make sure we do it locked since other instance can modify history while we read it
		var histLen int
		if err := rs.Cmd(&histLen, redisLLEN, utils.LoadInstKey); err != nil {
			return err
		}
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			if err = rs.Cmd(nil, redisRPOP, utils.LoadInstKey); err != nil {
				return err
			}
		}
		return rs.Cmd(nil, redisLPUSH, utils.LoadInstKey, string(marshaled))
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.LoadInstKey)

	if errCh := Cache.Remove(context.TODO(), utils.LoadInstKey, "",
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	return
}

func (rs *RedisStorage) GetResourceProfileDrv(ctx *context.Context, tenant, id string) (rsp *utils.ResourceProfile, err error) {
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

func (rs *RedisStorage) SetResourceProfileDrv(ctx *context.Context, rsp *utils.ResourceProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rsp); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.ResourceProfilesPrefix+rsp.TenantID(), string(result))
}

func (rs *RedisStorage) RemoveResourceProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetResourceDrv(ctx *context.Context, tenant, id string) (r *utils.Resource, err error) {
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

func (rs *RedisStorage) SetResourceDrv(ctx *context.Context, r *utils.Resource) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.ResourcesPrefix+r.TenantID(), string(result))
}

func (rs *RedisStorage) RemoveResourceDrv(ctx *context.Context, tenant, id string) (err error) {
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
func (rs *RedisStorage) GetStatQueueProfileDrv(ctx *context.Context, tenant string, id string) (sq *StatQueueProfile, err error) {
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
func (rs *RedisStorage) SetStatQueueProfileDrv(ctx *context.Context, sq *StatQueueProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sq); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(sq.Tenant, sq.ID), string(result))
}

// RemStatQueueProfileDrv removes a StatsQueue from dataDB
func (rs *RedisStorage) RemStatQueueProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

// GetStatQueueDrv retrieves the stored metrics for a StatsQueue
func (rs *RedisStorage) GetStatQueueDrv(ctx *context.Context, tenant, id string) (sq *StatQueue, err error) {
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
func (rs *RedisStorage) SetStatQueueDrv(ctx *context.Context, ssq *StoredStatQueue, sq *StatQueue) (err error) {
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
func (rs *RedisStorage) RemStatQueueDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) SetTrendProfileDrv(ctx *context.Context, sg *utils.TrendProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sg); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.TrendProfilePrefix+utils.ConcatenatedKey(sg.Tenant, sg.ID), string(result))
}

func (rs *RedisStorage) GetTrendProfileDrv(ctx *context.Context, tenant string, id string) (sg *utils.TrendProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.TrendProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &sg)
	return
}

func (rs *RedisStorage) RemTrendProfileDrv(ctx *context.Context, tenant string, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.TrendProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetTrendDrv(ctx *context.Context, tenant, id string) (r *utils.Trend, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.TrendPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetTrendDrv(ctx *context.Context, r *utils.Trend) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.TrendPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveTrendDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.TrendPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) SetRankingProfileDrv(ctx *context.Context, sg *utils.RankingProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sg); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.RankingProfilePrefix+utils.ConcatenatedKey(sg.Tenant, sg.ID), string(result))
}

func (rs *RedisStorage) GetRankingProfileDrv(ctx *context.Context, tenant string, id string) (sg *utils.RankingProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.RankingProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &sg)
	return
}

func (rs *RedisStorage) RemRankingProfileDrv(ctx *context.Context, tenant string, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.RankingProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetRankingDrv(ctx *context.Context, tenant, id string) (rn *utils.Ranking, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.RankingPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &rn)
	return rn, err
}

func (rs *RedisStorage) SetRankingDrv(_ *context.Context, rn *utils.Ranking) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rn); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.RankingPrefix+utils.ConcatenatedKey(rn.Tenant, rn.ID), string(result))
}

func (rs *RedisStorage) RemoveRankingDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.RankingPrefix+utils.ConcatenatedKey(tenant, id))
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (rs *RedisStorage) GetThresholdProfileDrv(ctx *context.Context, tenant, ID string) (tp *ThresholdProfile, err error) {
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
func (rs *RedisStorage) SetThresholdProfileDrv(ctx *context.Context, tp *ThresholdProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(tp); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.ThresholdProfilePrefix+tp.TenantID(), string(result))
}

// RemThresholdProfileDrv removes a ThresholdProfile from dataDB/cache
func (rs *RedisStorage) RemThresholdProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ThresholdProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetThresholdDrv(ctx *context.Context, tenant, id string) (r *Threshold, err error) {
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

func (rs *RedisStorage) SetThresholdDrv(ctx *context.Context, r *Threshold) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.ThresholdPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveThresholdDrv(ctx *context.Context, tenant, id string) (err error) {
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

func (rs *RedisStorage) RemoveFilterDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.FilterPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetRouteProfileDrv(ctx *context.Context, tenant, id string) (r *utils.RouteProfile, err error) {
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

func (rs *RedisStorage) SetRouteProfileDrv(ctx *context.Context, r *utils.RouteProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.RouteProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveRouteProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.RouteProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetAttributeProfileDrv(ctx *context.Context, tenant, id string) (r *utils.AttributeProfile, err error) {
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

func (rs *RedisStorage) SetAttributeProfileDrv(ctx *context.Context, r *utils.AttributeProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.AttributeProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveAttributeProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.AttributeProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetChargerProfileDrv(_ *context.Context, tenant, id string) (r *utils.ChargerProfile, err error) {
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

func (rs *RedisStorage) SetChargerProfileDrv(_ *context.Context, r *utils.ChargerProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.ChargerProfilePrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveChargerProfileDrv(_ *context.Context, tenant, id string) (err error) {
	return rs.Cmd(nil, redisDEL, utils.ChargerProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetStorageType() string {
	return utils.MetaRedis
}

func (rs *RedisStorage) GetItemLoadIDsDrv(ctx *context.Context, itemIDPrefix string) (loadIDs map[string]int64, err error) {
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

func (rs *RedisStorage) SetRateProfileDrv(ctx *context.Context, rpp *utils.RateProfile, optOverwrite bool) (err error) {
	rpMap, err := rpp.AsDataDBMap(rs.ms)
	if err != nil {
		return
	}
	if optOverwrite {
		rs.Cmd(nil, redisDEL, utils.RateProfilePrefix+utils.ConcatenatedKey(rpp.Tenant, rpp.ID))
	}
	return rs.FlatCmd(nil, redisHSET, utils.RateProfilePrefix+utils.ConcatenatedKey(rpp.Tenant, rpp.ID), rpMap)
}

func (rs *RedisStorage) GetRateProfileDrv(ctx *context.Context, tenant, id string) (rpp *utils.RateProfile, err error) {
	mapRP := make(map[string]any)
	if err = rs.Cmd(&mapRP, redisHGETALL, utils.RateProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(mapRP) == 0 {
		err = utils.ErrNotFound
		return
	}
	return utils.NewRateProfileFromMapDataDBMap(tenant, id, mapRP, rs.ms)
}

// GetRateProfileRateIDsDrv will return back all the rate IDs from a profile
func (rs *RedisStorage) GetRateProfileRatesDrv(ctx *context.Context, tnt, profileID, rtPrfx string, needIDs bool) (rateIDs []string, rates []*utils.Rate, err error) {
	key := utils.RateProfilePrefix + utils.ConcatenatedKey(tnt, profileID)
	prefix := utils.Rates + utils.ConcatenatedKeySep
	if rtPrfx != utils.EmptyString {
		prefix = utils.ConcatenatedKey(utils.Rates, rtPrfx)
	}
	var rateField string
	scan := radix.NewScanner(rs.client, radix.ScanOpts{
		Command: redisHSCAN,
		Key:     key,
		Pattern: prefix + utils.Meta,
	})
	idx := 0
	for scan.Next(&rateField) {
		idx++
		if idx%2 != 0 {
			if needIDs {
				rateIDs = append(rateIDs, strings.TrimPrefix(rateField, utils.Rates+utils.ConcatenatedKeySep))
			}
			continue
		}
		if needIDs {
			continue // we don't deserialize values for needIDs
		}
		rtToAppend := new(utils.Rate)
		if err = rs.ms.Unmarshal([]byte(rateField), rtToAppend); err != nil {
			return nil, nil, err
		}
		rates = append(rates, rtToAppend)
	}
	if err = scan.Close(); err != nil {
		return nil, nil, err
	}
	return
}

func (rs *RedisStorage) RemoveRateProfileDrv(ctx *context.Context, tenant, id string, rateIDs *[]string) (err error) {
	// if we want to remove just some rates from our profile, we will remove by their key Rates:rateID
	if rateIDs != nil {
		tntID := utils.ConcatenatedKey(tenant, id)
		for _, rateID := range *rateIDs {
			err = rs.Cmd(nil, redisHDEL, utils.RateProfilePrefix+tntID, utils.Rates+utils.InInFieldSep+rateID)
			if err != nil {
				return
			}
		}
		return
	}
	return rs.Cmd(nil, redisDEL, utils.RateProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetActionProfileDrv(ctx *context.Context, tenant, id string) (ap *utils.ActionProfile, err error) {
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

func (rs *RedisStorage) SetActionProfileDrv(ctx *context.Context, ap *utils.ActionProfile) (err error) {
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
func (rs *RedisStorage) GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
	mp := make(map[string]string)
	dbKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	if transactionID != utils.NonTransactional {
		dbKey = "tmp_" + utils.ConcatenatedKey(dbKey, transactionID)
	}
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
			if len(deleteArgs) == RedisLimit+1 { // minus dbkey
				if err = rs.Cmd(nil, redisHDEL, deleteArgs...); err != nil {
					return
				}
				deleteArgs = []string{dbKey} // the dbkey is necesary for the HDEL command
			}
			continue
		}
		var encodedMp []byte
		if encodedMp, err = rs.ms.Marshal(strMp); err != nil {
			return
		}
		mp[key] = string(encodedMp)
		if len(mp) == RedisLimit {
			if err = rs.FlatCmd(nil, redisHMSET, dbKey, mp); err != nil {
				return
			}
			mp = make(map[string]string)
		}
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

func (rs *RedisStorage) RemoveIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey string) (err error) {
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

func (rs *RedisStorage) GetConfigSectionsDrv(ctx *context.Context, nodeID string, sectionIDs []string) (sectionMap map[string][]byte, err error) {
	sectionMap = make(map[string][]byte)
	if len(sectionIDs) == 0 {
		if err = rs.Cmd(&sectionMap, redisHGETALL, utils.ConfigPrefix+nodeID); err != nil {
			return
		}
		return
	}
	sections := [][]byte{}
	if err = rs.FlatCmd(&sections, redisHMGET, utils.ConfigPrefix+nodeID, sectionIDs); err != nil {
		return
	}
	for i, sectionBytes := range sections {
		if len(sectionBytes) != 0 {
			sectionMap[sectionIDs[i]] = sectionBytes
		}
	}
	if len(sectionMap) == 0 {
		err = utils.ErrNotFound
		return
	}
	return
}

func (rs *RedisStorage) SetConfigSectionsDrv(ctx *context.Context, nodeID string, sectionsData map[string][]byte) (err error) {
	if err = rs.FlatCmd(nil, redisHSET, utils.ConfigPrefix+nodeID, sectionsData); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) RemoveConfigSectionsDrv(ctx *context.Context, nodeID string, sectionIDs []string) (err error) {
	if err = rs.FlatCmd(nil, redisHDEL, utils.ConfigPrefix+nodeID, sectionIDs); err != nil {
		return
	}
	return
}

// DumpDataDB will dump all of datadb from memory to a file, only for InternalDB
func (rs *RedisStorage) DumpDataDB() error {
	return utils.ErrNotImplemented
}

// Will rewrite every dump file of DataDB,  only for InternalDB
func (rs *RedisStorage) RewriteDataDB() (err error) {
	return utils.ErrNotImplemented
}

// BackupDataDB will momentarely stop any dumping and rewriting until all dump folder is backed up in folder path backupFolderPath, making zip true will create a zip file in the path instead, only for InternalDB
func (rs *RedisStorage) BackupDataDB(backupFolderPath string, zip bool) (err error) {
	return utils.ErrNotImplemented
}
