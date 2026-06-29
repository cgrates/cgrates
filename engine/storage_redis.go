/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"bytes"
	"compress/zlib"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/redis/rueidis"
)

type RedisStorage struct {
	client rueidis.Client
	ms     Marshaler
}

// Redis commands
const (
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
	redis_RENAME   = "RENAME"
	redis_HSET     = "HMSET"

	redisLoadError = "Redis is loading the dataset in memory"
	RedisLimit     = 524287 // https://github.com/StackExchange/StackExchange.Redis/issues/201#issuecomment-98639005
)

func NewRedisStorage(address string, db int, user, pass, mrshlerStr string,
	maxConns, retryAttempts int, sentinelName string, isCluster bool, clusterSync,
	clusterOnDownDelay, connTimeout, pipelineWindow time.Duration, pipelineLimit int,
	tlsConn bool, tlsClientCert, tlsClientKey, tlsCACert string) (_ *RedisStorage, err error) {
	var ms Marshaler
	if ms, err = NewMarshaler(mrshlerStr); err != nil {
		return
	}
	opt := rueidis.ClientOption{
		InitAddress:      []string{address},
		ConnWriteTimeout: connTimeout,
		BlockingPoolSize: maxConns,
		ClusterOption: rueidis.ClusterOption{
			ShardsRefreshInterval: clusterSync,
		},
		DisableRetry: false,
	}
	if pass != utils.EmptyString {
		if user != utils.EmptyString {
			opt.Username = user
		}
		opt.Password = pass
	}
	if pipelineLimit > 0 {
		opt.PipelineMultiplex = pipelineLimit
	}
	if pipelineWindow > 0 {
		opt.MaxFlushDelay = pipelineWindow
	}

	opt.RetryDelay = func(attempts int, cmd rueidis.Completed, err error) time.Duration {
		if attempts == retryAttempts {
			return -1 // stop retrying
		}
		if clusterOnDownDelay > 0 {
			// if CLUSTERDOWN state, delay and retry
			if err != nil && strings.HasPrefix(err.Error(), "CLUSTERDOWN") {
				return clusterOnDownDelay
			}
		}
		// taken from rueidis defaultRetryDelayFn
		// delay the next retry exponentially without considering the error. From 1 microsecond, Max delay is 1 second
		base := 1 << min(retryAttempts, attempts)
		// randomise delay so not all clients restart at the same time to not overload the server
		return min(1*time.Second, time.Duration(base+rand.IntN(base))*time.Microsecond)
	}
	// Add TLS configuration
	if tlsConn {
		tlsConfig := &tls.Config{}
		if tlsClientCert != "" && tlsClientKey != "" {
			cert, err := tls.LoadX509KeyPair(tlsClientCert, tlsClientKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificates: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		if tlsCACert != "" {
			caCert, err := os.ReadFile(tlsCACert)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		} else {
			systemCerts, err := x509.SystemCertPool()
			if err == nil && systemCerts != nil {
				tlsConfig.RootCAs = systemCerts
			} else {
				tlsConfig.RootCAs = x509.NewCertPool()
			}
		}

		opt.TLSConfig = tlsConfig
	}
	if !isCluster { // DB is not selectable if its a cluster
		opt.SelectDB = db
	}
	if sentinelName != utils.EmptyString {
		opt.Sentinel = rueidis.SentinelOption{
			MasterSet: sentinelName,
		}
	}
	client, err := rueidis.NewClient(opt)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	return &RedisStorage{
		ms:     ms,
		client: client,
	}, nil
}

// Cmd executes a Redis command with the given receiver for the result.
func (rs *RedisStorage) Cmd(rcv any, cmd string, args ...string) error {
	ctx := context.Background()
	var cmdObj rueidis.Completed

	switch cmd {
	case redis_GET:
		cmdObj = rs.client.B().Get().Key(args[0]).Build()
	case redis_SET:
		cmdObj = rs.client.B().Set().Key(args[0]).Value(args[1]).Build()
	case redis_DEL:
		if config.CgrConfig().DataDbCfg().Opts.RedisClusterOndownDelay > 0 {
			for range config.CgrConfig().DataDbCfg().Opts.RedisConnectAttempts {
				if err := rs.client.Do(ctx, rs.client.B().Del().Key(args...).Build()).Error(); err != nil {
					if strings.HasPrefix(err.Error(), "CLUSTERDOWN") {
						time.Sleep(config.CgrConfig().DataDbCfg().Opts.RedisClusterOndownDelay)
						continue
					}
					return err
				}
				return nil
			}
		}
		return rs.client.Do(ctx, rs.client.B().Del().Key(args...).Build()).Error()
	case redis_EXISTS:
		cmdObj = rs.client.B().Exists().Key(args...).Build()
	case redis_KEYS:
		cmdObj = rs.client.B().Keys().Pattern(args[0]).Build()
	case redis_FLUSHDB:
		cmdObj = rs.client.B().Flushdb().Build()
	case redis_HGETALL:
		cmdObj = rs.client.B().Hgetall().Key(args[0]).Build()
	case redis_HDEL:
		cmdObj = rs.client.B().Hdel().Key(args[0]).Field(args[1:]...).Build()
	case redis_HMGET:
		cmdObj = rs.client.B().Hmget().Key(args[0]).Field(args[1:]...).Build()
	case redis_SADD:
		cmdObj = rs.client.B().Sadd().Key(args[0]).Member(args[1:]...).Build()
	case redis_SMEMBERS:
		cmdObj = rs.client.B().Smembers().Key(args[0]).Build()
	case redis_SREM:
		cmdObj = rs.client.B().Srem().Key(args[0]).Member(args[1:]...).Build()
	case redis_LRANGE:
		start, _ := strconv.ParseInt(args[1], 10, 64)
		stop, _ := strconv.ParseInt(args[2], 10, 64)
		cmdObj = rs.client.B().Lrange().Key(args[0]).Start(start).Stop(stop).Build()
	case redis_LPUSH:
		cmdObj = rs.client.B().Lpush().Key(args[0]).Element(args[1:]...).Build()
	case redis_RPUSH:
		cmdObj = rs.client.B().Rpush().Key(args[0]).Element(args[1:]...).Build()
	case redis_LPOP:
		cmdObj = rs.client.B().Lpop().Key(args[0]).Build()
	case redis_RPOP:
		cmdObj = rs.client.B().Rpop().Key(args[0]).Build()
	case redis_LLEN:
		cmdObj = rs.client.B().Llen().Key(args[0]).Build()
	case redis_RENAME:
		cmdObj = rs.client.B().Rename().Key(args[0]).Newkey(args[1]).Build()
	default:
		// Fallback to arbitrary command
		arbitrary := rs.client.B().Arbitrary(cmd)
		for _, arg := range args {
			arbitrary = arbitrary.Args(arg)
		}
		cmdObj = arbitrary.Build()
	}

	result := rs.client.Do(ctx, cmdObj)
	if result.Error() != nil {
		if rueidis.IsRedisNil(result.Error()) {
			return nil
		}
		if config.CgrConfig().DataDbCfg().Opts.RedisClusterOndownDelay > 0 {
			for range config.CgrConfig().DataDbCfg().Opts.RedisConnectAttempts {
				if err := result.Error(); err != nil {
					if strings.HasPrefix(err.Error(), "CLUSTERDOWN") {
						time.Sleep(config.CgrConfig().DataDbCfg().Opts.RedisClusterOndownDelay)
						continue
					}
					return err
				}
				return nil
			}
		}
		return result.Error()
	}

	return rs.parseResult(result, rcv)
}

// FlatCmd executes a Redis command with flat key-value arguments, optimized for hash operations
func (rs *RedisStorage) FlatCmd(rcv any, cmd, key string, args ...any) error {
	ctx := context.Background()
	var cmdObj rueidis.Completed

	switch cmd {
	case redis_HSET:
		b := rs.client.B().Hset().Key(key).FieldValue()
		for _, arg := range args {
			switch v := arg.(type) {
			case map[string]string:
				for k, val := range v {
					b.FieldValue(k, val)
				}
			case Versions:
				for k, val := range v {
					b.FieldValue(k, strconv.FormatInt(val, 10))
				}
			case map[string]int64:
				for k, val := range v {
					b.FieldValue(k, strconv.FormatInt(val, 10))
				}
			}
		}
		cmdObj = b.Build()

	case redis_HMGET:
		strFields := make([]string, 0, len(args))
		for _, arg := range args {
			strFields = append(strFields, fmt.Sprint(arg))
		}
		cmdObj = rs.client.B().Hmget().Key(key).Field(strFields...).Build()

	case redis_HDEL:
		strFields := make([]string, 0, len(args))
		for _, arg := range args {
			strFields = append(strFields, fmt.Sprint(arg))
		}
		cmdObj = rs.client.B().Hdel().Key(key).Field(strFields...).Build()

	case redis_SADD:
		strMembers := make([]string, 0, len(args))
		for _, arg := range args {
			strMembers = append(strMembers, fmt.Sprint(arg))
		}
		cmdObj = rs.client.B().Sadd().Key(key).Member(strMembers...).Build()

	case redis_SREM:
		strMembers := make([]string, 0, len(args))
		for _, arg := range args {
			strMembers = append(strMembers, fmt.Sprint(arg))
		}
		cmdObj = rs.client.B().Srem().Key(key).Member(strMembers...).Build()

	case redis_LPUSH:
		strElements := make([]string, 0, len(args))
		for _, arg := range args {
			strElements = append(strElements, fmt.Sprint(arg))
		}
		cmdObj = rs.client.B().Lpush().Key(key).Element(strElements...).Build()

	case redis_RPUSH:
		strElements := make([]string, 0, len(args))
		for _, arg := range args {
			strElements = append(strElements, fmt.Sprint(arg))
		}
		cmdObj = rs.client.B().Rpush().Key(key).Element(strElements...).Build()

	default:
		// Fallback to arbitrary command
		arbitrary := rs.client.B().Arbitrary(cmd).Args(key)
		for _, arg := range args {
			arbitrary = arbitrary.Args(fmt.Sprint(arg))
		}
		cmdObj = arbitrary.Build()
	}

	result := rs.client.Do(ctx, cmdObj)
	if result.Error() != nil {
		if rueidis.IsRedisNil(result.Error()) {
			return nil
		}
		if config.CgrConfig().DataDbCfg().Opts.RedisClusterOndownDelay > 0 {
			for range config.CgrConfig().DataDbCfg().Opts.RedisConnectAttempts {
				if err := result.Error(); err != nil {
					if strings.HasPrefix(err.Error(), "CLUSTERDOWN") {
						time.Sleep(config.CgrConfig().DataDbCfg().Opts.RedisClusterOndownDelay)
						continue
					}

					return err
				}
				return nil
			}
		}
		return result.Error()
	}

	return rs.parseResult(result, rcv)
}

// parseResult unmarshals a Redis result into the receiver based on its type
func (rs *RedisStorage) parseResult(result rueidis.RedisResult, rcv any) error {
	switch v := rcv.(type) {
	case nil:
		return nil
	case *[]byte:
		val, err := result.AsBytes()
		if err != nil {
			return err
		}
		*v = val
	case *string:
		val, err := result.AsBytes()
		if err != nil {
			return err
		}
		*v = string(val)
	case *[]string:
		vals, err := result.AsStrSlice()
		if err != nil {
			arr, err2 := result.ToArray()
			if err2 != nil {
				return err
			}
			for _, item := range arr {
				s, _ := item.AsBytes()
				*v = append(*v, string(s))
			}
			return nil
		}
		*v = vals
	case *map[string]string:
		vals, err := result.AsStrMap()
		if err != nil {
			return err
		}
		for k, val := range vals {
			(*v)[k] = val
		}
	case *int:
		val, err := result.AsInt64()
		if err != nil {
			return err
		}
		*v = int(val)
	case *int64:
		val, err := result.AsInt64()
		if err != nil {
			return err
		}
		*v = val
	case *bool:
		val, err := result.AsBool()
		if err != nil {
			return err
		}
		*v = val
	case *[][]byte:
		arr, err := result.ToArray()
		if err != nil {
			return err
		}
		for _, item := range arr {
			val, err := item.AsBytes()
			if err != nil {
				continue
			}
			*v = append(*v, val)
		}
	default:
		// Try to unmarshal as bytes using the marshaler
		val, err := result.AsBytes()
		if err != nil {
			return err
		}
		if len(val) == 0 {
			return nil
		}
		return rs.ms.Unmarshal(val, rcv)
	}

	return nil
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

func (rs *RedisStorage) IsDBEmpty() (resp bool, err error) {
	var keys []string
	keys, err = rs.GetKeysForPrefix(utils.EmptyString, utils.EmptyString)
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
	if keys, err = rs.GetKeysForPrefix(prefix, utils.EmptyString); err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd(nil, redis_DEL, key); err != nil {
			return
		}
	}
	return
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
	if err = rs.Cmd(&keys, redis_KEYS, utils.ActionPlanPrefix+"*"); err != nil {
		return
	}
	for _, key := range keys {
		if err = rs.Cmd(nil, redis_SADD, utils.ActionPlanIndexes, key); err != nil {
			return
		}
	}
	return
}

// GetKeysForPrefix returns the keys from redis that have the given prefix. If search is
// populated it will also match the string given in search parameter
func (rs *RedisStorage) GetKeysForPrefix(prefix, search string) ([]string, error) {
	var keys []string
	if prefix == utils.ActionPlanPrefix {
		if err := rs.Cmd(&keys, redis_SMEMBERS, utils.ActionPlanIndexes); err != nil {
			return nil, fmt.Errorf("failed to get ActionPlan from index: %v", err)
		}
	} else {
		// Use SCAN for other prefixes to avoid blocking Redis.
		var err error
		if search != utils.EmptyString {
			if keys, err = rs.scanKeysForPrefixContaining(prefix, search); err != nil {
				return nil, fmt.Errorf("failed to scan keys for matching <%s>: %v", search, err)
			}
		} else {
			if keys, err = rs.scanKeysForPrefix(prefix); err != nil {
				return nil, fmt.Errorf("failed to scan keys for prefix %s: %v", prefix, err)
			}
		}
	}
	if len(keys) == 0 {
		return nil, nil
	}
	if filterIndexesPrefixMap.Has(prefix) {
		return rs.getKeysForFilterIndexesKeys(keys)
	}
	return keys, nil
}

// scanKeysForPrefixContaining performs a non-blocking scan for keys matching the given prefix and search value.
func (rs *RedisStorage) scanKeysForPrefixContaining(prefix, search string) ([]string, error) {
	if search == utils.EmptyString {
		return nil, nil
	}

	ctx := context.Background()
	var keys []string
	pattern := prefix + utils.Meta + search + utils.Meta
	cursor := uint64(0)
	batchSize := int64(config.CgrConfig().DataDbCfg().Opts.RedisBatchSize)

	// Iterate through all keys using SCAN
	for {
		cmd := rs.client.B().Scan().Cursor(cursor).Match(pattern).Count(batchSize).Build()
		result := rs.client.Do(ctx, cmd)
		if result.Error() != nil && result.Error() != rueidis.Nil {
			return nil, result.Error()
		}

		scanEntry, err := result.AsScanEntry()
		if err != nil {
			return nil, err
		}

		keys = append(keys, scanEntry.Elements...)

		cursor = scanEntry.Cursor
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// scanKeysForPrefix performs a non-blocking scan for keys matching the given prefix.
func (rs *RedisStorage) scanKeysForPrefix(prefix string) ([]string, error) {
	ctx := context.Background()
	var keys []string
	cursor := uint64(0)
	batchSize := int64(config.CgrConfig().DataDbCfg().Opts.RedisBatchSize)

	// Iterate through all keys using SCAN
	for {
		cmd := rs.client.B().Scan().Cursor(cursor).Match(prefix + "*").Count(batchSize).Build()
		result := rs.client.Do(ctx, cmd)
		if result.Error() != nil && result.Error() != rueidis.Nil {
			return nil, result.Error()
		}

		scanEntry, err := result.AsScanEntry()
		if err != nil {
			return nil, err
		}

		keys = append(keys, scanEntry.Elements...)

		cursor = scanEntry.Cursor
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// Used to check if specific subject is stored using prefix key attached to entity
func (rs *RedisStorage) HasDataDrv(category, subject, tenant string) (exists bool, err error) {
	var i int
	switch category {
	case utils.DestinationPrefix, utils.RatingPlanPrefix, utils.RatingProfilePrefix,
		utils.ActionPrefix, utils.ActionPlanPrefix, utils.AccountPrefix:
		err = rs.Cmd(&i, redis_EXISTS, category+subject)
		return i == 1, err
	case utils.ResourcesPrefix, utils.ResourceProfilesPrefix, utils.IPAllocationsPrefix,
		utils.IPProfilesPrefix, utils.StatQueuePrefix, utils.StatQueueProfilePrefix,
		utils.ThresholdPrefix, utils.ThresholdProfilePrefix, utils.FilterPrefix,
		utils.RouteProfilePrefix, utils.AttributeProfilePrefix, utils.ChargerProfilePrefix,
		utils.DispatcherProfilePrefix, utils.DispatcherHostPrefix:
		err := rs.Cmd(&i, redis_EXISTS, category+utils.ConcatenatedKey(tenant, subject))
		return i == 1, err
	}
	return false, errors.New("unsupported HasData category")
}

func (rs *RedisStorage) GetRatingPlanDrv(key string) (rp *RatingPlan, err error) {
	key = utils.RatingPlanPrefix + key
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
	if out, err = io.ReadAll(r); err != nil {
		return
	}
	r.Close()
	err = rs.ms.Unmarshal(out, &rp)
	return
}

func (rs *RedisStorage) SetRatingPlanDrv(rp *RatingPlan) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rp); err != nil {
		return
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.Cmd(nil, redis_SET, utils.RatingPlanPrefix+rp.Id, b.String())
	return
}

func (rs *RedisStorage) RemoveRatingPlanDrv(key string) (err error) {
	var keys []string
	if err = rs.Cmd(&keys, redis_KEYS, utils.RatingPlanPrefix+key+"*"); err != nil {
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
	key = utils.RatingProfilePrefix + key
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
	key := utils.RatingProfilePrefix + rpf.Id
	err = rs.Cmd(nil, redis_SET, key, string(result))
	return
}

func (rs *RedisStorage) RemoveRatingProfileDrv(key string) (err error) {
	var keys []string
	if err = rs.Cmd(&keys, redis_KEYS, utils.RatingProfilePrefix+key+"*"); err != nil {
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
func (rs *RedisStorage) GetDestinationDrv(key, transactionID string) (dest *Destination, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.DestinationPrefix+key); err != nil {
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
	if out, err = io.ReadAll(r); err != nil {
		return
	}
	r.Close()
	err = rs.ms.Unmarshal(out, &dest)
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
	err = rs.Cmd(nil, redis_SET, utils.DestinationPrefix+dest.Id, b.String())
	return
}

func (rs *RedisStorage) GetReverseDestinationDrv(key, transactionID string) (ids []string, err error) {
	if err = rs.Cmd(&ids, redis_SMEMBERS, utils.ReverseDestinationPrefix+key); err != nil {
		return
	}
	if len(ids) == 0 {
		err = utils.ErrNotFound
	}
	return
}

func (rs *RedisStorage) SetReverseDestinationDrv(destID string, prefixes []string, transactionID string) (err error) {
	for _, p := range prefixes {
		if err = rs.Cmd(nil, redis_SADD, utils.ReverseDestinationPrefix+p, destID); err != nil {
			return
		}
	}
	return
}

func (rs *RedisStorage) RemoveDestinationDrv(destID, transactionID string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.DestinationPrefix+destID)
}

func (rs *RedisStorage) RemoveReverseDestinationDrv(dstID, prfx, transactionID string) (err error) {
	return rs.Cmd(nil, redis_SREM, utils.ReverseDestinationPrefix+prfx, dstID)
}

func (rs *RedisStorage) GetActionsDrv(key string) (as Actions, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ActionPrefix+key); err != nil {
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
	return rs.Cmd(nil, redis_SET, utils.ActionPrefix+key, string(result))
}

func (rs *RedisStorage) RemoveActionsDrv(key string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ActionPrefix+key)
}

func (rs *RedisStorage) GetSharedGroupDrv(key string) (sg *SharedGroup, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.SharedGroupPrefix+key); err != nil {
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
	return rs.Cmd(nil, redis_SET, utils.SharedGroupPrefix+sg.Id, string(result))
}

func (rs *RedisStorage) RemoveSharedGroupDrv(id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.SharedGroupPrefix+id)
}

func (rs *RedisStorage) GetAccountDrv(key string) (ub *Account, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.AccountPrefix+key); err != nil {
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
	return rs.Cmd(nil, redis_SET, utils.AccountPrefix+acc.ID, string(result))
}

func (rs *RedisStorage) RemoveAccountDrv(key string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.AccountPrefix+key)
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
	if err = rs.Cmd(&marshaleds, redis_LRANGE,
		utils.LoadInstKey, "0", strconv.Itoa(limit)); err != nil {
		if errCh := Cache.Set(utils.LoadInstKey, "", nil, nil,
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
	if err = Cache.Remove(utils.LoadInstKey, "", cCommit, transactionID); err != nil {
		return nil, err
	}
	if err := Cache.Set(utils.LoadInstKey, "", loadInsts, nil,
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
	err = guardian.Guardian.Guard(func() error { // Make sure we do it locked since other instance can modify history while we read it
		var histLen int
		if err := rs.Cmd(&histLen, redis_LLEN, utils.LoadInstKey); err != nil {
			return err
		}
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			if err = rs.Cmd(nil, redis_RPOP, utils.LoadInstKey); err != nil {
				return err
			}
		}
		return rs.Cmd(nil, redis_LPUSH, utils.LoadInstKey, string(marshaled))
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.LoadInstKey)

	if errCh := Cache.Remove(utils.LoadInstKey, "",
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	return
}

func (rs *RedisStorage) GetActionTriggersDrv(key string) (atrs ActionTriggers, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ActionTriggerPrefix+key); err != nil {
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
		return rs.Cmd(nil, redis_DEL, utils.ActionTriggerPrefix+key)
	}
	var result []byte
	if result, err = rs.ms.Marshal(atrs); err != nil {
		return
	}
	if err = rs.Cmd(nil, redis_SET, utils.ActionTriggerPrefix+key, string(result)); err != nil {
		return
	}
	return
}

func (rs *RedisStorage) RemoveActionTriggersDrv(key string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.ActionTriggerPrefix+key)
}

func (rs *RedisStorage) GetActionPlanDrv(key string) (ats *ActionPlan, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.ActionPlanPrefix+key); err != nil {
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
	if out, err = io.ReadAll(r); err != nil {
		return
	}
	r.Close()
	err = rs.ms.Unmarshal(out, &ats)
	return
}
func (rs *RedisStorage) RemoveActionPlanDrv(key string) (err error) {
	if err = rs.Cmd(nil, redis_SREM, utils.ActionPlanIndexes, utils.ActionPlanPrefix+key); err != nil {
		return
	}
	return rs.Cmd(nil, redis_DEL, utils.ActionPlanPrefix+key)
}

func (rs *RedisStorage) SetActionPlanDrv(key string, ats *ActionPlan) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(ats); err != nil {
		return
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	if err = rs.Cmd(nil, redis_SADD, utils.ActionPlanIndexes, utils.ActionPlanPrefix+key); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.ActionPlanPrefix+key, b.String())
}

func (rs *RedisStorage) GetAllActionPlansDrv() (ats map[string]*ActionPlan, err error) {
	var keys []string
	if keys, err = rs.GetKeysForPrefix(utils.ActionPlanPrefix, utils.EmptyString); err != nil {
		return
	}
	if len(keys) == 0 {
		err = utils.ErrNotFound
		return
	}
	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		if ats[key[len(utils.ActionPlanPrefix):]], err = rs.GetActionPlanDrv(key[len(utils.ActionPlanPrefix):]); err != nil {
			return nil, err
		}
	}
	return
}

func (rs *RedisStorage) GetAccountActionPlansDrv(acntID string) (aPlIDs []string, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET,
		utils.AccountActionPlansPrefix+acntID); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &aPlIDs)
	return
}

func (rs *RedisStorage) SetAccountActionPlansDrv(acntID string, aPlIDs []string) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(aPlIDs); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.AccountActionPlansPrefix+acntID, string(result))
}

func (rs *RedisStorage) RemAccountActionPlansDrv(acntID string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.AccountActionPlansPrefix+acntID)
}

func (rs *RedisStorage) PushTask(t *Task) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(t); err != nil {
		return
	}
	return rs.Cmd(nil, redis_RPUSH, utils.TasksKey, string(result))
}

func (rs *RedisStorage) PopTask() (t *Task, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_LPOP, utils.TasksKey); err != nil {
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

func (rs *RedisStorage) GetIPProfileDrv(tenant, id string) (*IPProfile, error) {
	var values []byte
	if err := rs.Cmd(&values, redis_GET, utils.IPProfilesPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, utils.ErrNotFound
	}
	var ipp *IPProfile
	if err := rs.ms.Unmarshal(values, &ipp); err != nil {
		return nil, err
	}
	return ipp, nil
}

func (rs *RedisStorage) SetIPProfileDrv(ipp *IPProfile) error {
	result, err := rs.ms.Marshal(ipp)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.IPProfilesPrefix+ipp.TenantID(), string(result))
}

func (rs *RedisStorage) RemoveIPProfileDrv(tenant, id string) error {
	return rs.Cmd(nil, redis_DEL, utils.IPProfilesPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetIPAllocationsDrv(tenant, id string) (*IPAllocations, error) {
	var values []byte
	if err := rs.Cmd(&values, redis_GET, utils.IPAllocationsPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, utils.ErrNotFound
	}
	var ip *IPAllocations
	if err := rs.ms.Unmarshal(values, &ip); err != nil {
		return nil, err
	}
	return ip, nil
}

func (rs *RedisStorage) SetIPAllocationsDrv(ip *IPAllocations) error {
	result, err := rs.ms.Marshal(ip)
	if err != nil {
		return err
	}
	return rs.Cmd(nil, redis_SET, utils.IPAllocationsPrefix+ip.TenantID(), string(result))
}

func (rs *RedisStorage) RemoveIPAllocationsDrv(tenant, id string) error {
	return rs.Cmd(nil, redis_DEL, utils.IPAllocationsPrefix+utils.ConcatenatedKey(tenant, id))
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
		ctx := context.Background()
		cmd := rs.client.B().Hget().Key(utils.TBLVersions).Field(itm).Build()
		result := rs.client.Do(ctx, cmd)
		if result.Error() != nil {
			// Check if it's a nil response (key doesn't exist)
			if rueidis.IsRedisNil(result.Error()) {
				err = utils.ErrNotFound
				return
			}
			return nil, result.Error()
		}
		val, err := result.AsInt64()
		if err != nil {
			return nil, utils.ErrNotFound
		}
		return Versions{itm: val}, nil
	}

	ctx := context.Background()
	cmd := rs.client.B().Hgetall().Key(utils.TBLVersions).Build()
	result := rs.client.Do(ctx, cmd)
	if result.Error() != nil && result.Error() != rueidis.Nil {
		return nil, result.Error()
	}

	mp, err := result.AsStrMap()
	if err != nil {
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
	return rs.FlatCmd(nil, redis_HSET, utils.TBLVersions, vrs)
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
	if ssq == nil {
		if ssq, err = NewStoredStatQueue(sq, rs.ms); err != nil {
			return
		}
	}
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

func (rs *RedisStorage) SetTrendProfileDrv(sg *TrendProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sg); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.TrendsProfilePrefix+utils.ConcatenatedKey(sg.Tenant, sg.ID), string(result))
}

func (rs *RedisStorage) GetTrendProfileDrv(tenant string, id string) (sg *TrendProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.TrendsProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &sg)
	return
}

func (rs *RedisStorage) RemTrendProfileDrv(tenant string, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.TrendsProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetTrendDrv(tenant, id string) (tr *Trend, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.TrendPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &tr)
	return tr, err
}

func (rs *RedisStorage) SetTrendDrv(r *Trend) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(r); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.TrendPrefix+utils.ConcatenatedKey(r.Tenant, r.ID), string(result))
}

func (rs *RedisStorage) RemoveTrendDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.TrendPrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) SetRankingProfileDrv(sg *RankingProfile) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(sg); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.RankingsProfilePrefix+utils.ConcatenatedKey(sg.Tenant, sg.ID), string(result))
}

func (rs *RedisStorage) GetRankingProfileDrv(tenant string, id string) (sg *RankingProfile, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.RankingsProfilePrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &sg)
	return
}
func (rs *RedisStorage) RemRankingProfileDrv(tenant string, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.RankingsProfilePrefix+utils.ConcatenatedKey(tenant, id))
}

func (rs *RedisStorage) GetRankingDrv(tenant, id string) (rn *Ranking, err error) {
	var values []byte
	if err = rs.Cmd(&values, redis_GET, utils.RankingPrefix+utils.ConcatenatedKey(tenant, id)); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	err = rs.ms.Unmarshal(values, &rn)
	return rn, err
}

func (rs *RedisStorage) SetRankingDrv(rn *Ranking) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(rn); err != nil {
		return
	}
	return rs.Cmd(nil, redis_SET, utils.RankingPrefix+utils.ConcatenatedKey(rn.Tenant, rn.ID), string(result))
}

func (rs *RedisStorage) RemoveRankingDrv(tenant, id string) (err error) {
	return rs.Cmd(nil, redis_DEL, utils.RankingPrefix+utils.ConcatenatedKey(tenant, id))
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
		err = utils.ErrDSPProfileNotFound
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
		err = utils.ErrDSPHostNotFound
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
	return utils.MetaRedis
}

func (rs *RedisStorage) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	if itemIDPrefix != "" {
		ctx := context.Background()
		cmd := rs.client.B().Hget().Key(utils.LoadIDs).Field(itemIDPrefix).Build()
		result := rs.client.Do(ctx, cmd)
		if result.Error() != nil {
			if rueidis.IsRedisNil(result.Error()) {
				err = utils.ErrNotFound
			} else {
				err = result.Error()
			}
			return
		}
		val, err := result.AsInt64()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				return nil, utils.ErrNotFound
			}
			return nil, err
		}
		return map[string]int64{itemIDPrefix: val}, nil
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
	return rs.FlatCmd(nil, redis_HSET, utils.LoadIDs, loadIDs)
}

func (rs *RedisStorage) RemoveLoadIDsDrv() (err error) {
	return rs.Cmd(nil, redis_DEL, utils.LoadIDs)
}

// GetIndexesDrv retrieves Indexes from dataDB
func (rs *RedisStorage) GetIndexesDrv(idxItmType, tntCtx string, idxKeys ...string) (indexes map[string]utils.StringSet, err error) {
	mp := make(map[string]string)
	dbKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	if len(idxKeys) == 0 {
		if err = rs.Cmd(&mp, redis_HGETALL, dbKey); err != nil {
			return
		} else if len(mp) == 0 {
			return nil, utils.ErrNotFound
		}
	} else {
		var itmMpStrLst []string
		args := make([]string, len(idxKeys)+1)
		args[0] = dbKey
		for i, key := range idxKeys {
			args[i+1] = key
		}
		if err = rs.Cmd(&itmMpStrLst, redis_HMGET, args...); err != nil {
			return
		}
		for i, val := range itmMpStrLst {
			if val != utils.EmptyString {
				mp[idxKeys[i]] = val
			}
		}
		if len(mp) == 0 {
			return nil, utils.ErrNotFound
		}
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
			if len(deleteArgs) == RedisLimit+1 { // minus dbkey
				if err = rs.Cmd(nil, redis_HDEL, deleteArgs...); err != nil {
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
			if err = rs.FlatCmd(nil, redis_HSET, dbKey, mp); err != nil {
				return
			}
			mp = make(map[string]string)
		}
	}
	if len(deleteArgs) != 1 {
		if err = rs.Cmd(nil, redis_HDEL, deleteArgs...); err != nil {
			return
		}
	}
	if len(mp) != 0 {
		return rs.FlatCmd(nil, redis_HSET, dbKey, mp)
	}
	return
}

func (rs *RedisStorage) RemoveIndexesDrv(idxItmType, tntCtx string, idxKeys ...string) (err error) {
	if len(idxKeys) == 0 {
		return rs.Cmd(nil, redis_DEL, utils.CacheInstanceToPrefix[idxItmType]+tntCtx)
	}
	args := make([]string, len(idxKeys)+1)
	args[0] = utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	for i, key := range idxKeys {
		args[i+1] = key
	}
	return rs.Cmd(nil, redis_HDEL, args...)
}

// Will backup active sessions in DataDB
func (rs *RedisStorage) SetBackupSessionsDrv(nodeID string,
	tnt string, storedSessions []*StoredSession) (err error) {
	mp := make(map[string]string)
	for _, sess := range storedSessions {
		// Convert time.Time values inside EventStart and SRuns Events, to string type values
		utils.MapIfaceTimeAsString(sess.EventStart)
		for i := range sess.SRuns {
			utils.MapIfaceTimeAsString(sess.SRuns[i].Event)
		}
		var sessByte []byte
		if sessByte, err = rs.ms.Marshal(sess); err != nil {
			return
		}
		mp[sess.CGRID] = string(sessByte)
		if len(mp) == RedisLimit {
			if err = rs.FlatCmd(nil, redis_HSET, utils.SessionsBackupPrefix+utils.ConcatenatedKey(tnt,
				nodeID), mp); err != nil {
				return
			}
			mp = make(map[string]string)
		}
	}
	return rs.FlatCmd(nil, redis_HSET, utils.SessionsBackupPrefix+utils.ConcatenatedKey(tnt, nodeID), mp)
}

// Will restore sessions that were active from dataDB backup
func (rs *RedisStorage) GetSessionsBackupDrv(nodeID, tnt string) (r []*StoredSession, err error) {
	mp := make(map[string]string)
	if err = rs.Cmd(&mp, redis_HGETALL, utils.SessionsBackupPrefix+utils.ConcatenatedKey(tnt,
		nodeID)); err != nil {
		return
	} else if len(mp) == 0 {
		return nil, utils.ErrNoBackupFound
	}
	for _, v := range mp {
		var ss *StoredSession
		if err = rs.ms.Unmarshal([]byte(v), &ss); err != nil {
			return
		}
		r = append(r, ss)
	}
	return
}

// Will remove one or all sessions from dataDB backup
func (rs *RedisStorage) RemoveSessionsBackupDrv(nodeID, tnt, cgrid string) error {
	if cgrid == utils.EmptyString {
		return rs.Cmd(nil, redis_DEL, utils.SessionsBackupPrefix+utils.ConcatenatedKey(tnt, nodeID))
	}
	return rs.Cmd(nil, redis_HDEL, utils.SessionsBackupPrefix+utils.ConcatenatedKey(tnt, nodeID), cgrid)
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

// RestoreDataDB, only for InternalDB
func (rs *RedisStorage) RestoreDataDB(backupFolderPath string) (err error) {
	return utils.ErrNotImplemented
}

// SnapshotDataDB, only for InternalDB
func (rs *RedisStorage) SnapshotDataDB(backupFolderPath string, zip bool) (err error) {
	return utils.ErrNotImplemented
}
