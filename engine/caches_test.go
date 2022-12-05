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
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestCacheSSetWithReplicateTrue(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgCacheReplicateSet{
		CacheID:  "chID",
		ItemID:   "itemID",
		Value:    &utils.CachedRPCResponse{Result: "reply", Error: nil},
		GroupIDs: []string{"groupId", "groupId"},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Replicate: true,
		},
	}

	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1ReplicateSet: func(_ *context.Context, args, reply interface{}) error {
				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg), utils.CacheSv1, clientconn)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			ltcache.DefaultCacheInstance: {
				MaxItems:  3,
				TTL:       time.Second * 1,
				StaticTTL: true,
				OnEvicted: func(itmID string, value interface{}) {

				},
			},
		},
	)
	casheS := &CacheS{
		cfg:     cfg,
		dm:      dm,
		tCache:  tscache,
		connMgr: connMgr,
	}
	if err := casheS.SetWithReplicate(context.Background(), args); err != nil {
		t.Error(err)
	}
}

func TestCacheSSetWithReplicateFalse(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgCacheReplicateSet{
		CacheID:  "chID",
		ItemID:   "itemID",
		Value:    &utils.CachedRPCResponse{Result: "reply", Error: nil},
		GroupIDs: []string{"groupId", "groupId"},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Replicate: false,
		},
	}

	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1ReplicateSet: func(_ *context.Context, args, reply interface{}) error {
				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg), utils.CacheSv1, clientconn)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			ltcache.DefaultCacheInstance: {
				MaxItems:  3,
				TTL:       time.Second * 1,
				StaticTTL: true,
				OnEvicted: func(itmID string, value interface{}) {

				},
			},
		},
	)
	casheS := &CacheS{
		cfg:     cfg,
		dm:      dm,
		tCache:  tscache,
		connMgr: connMgr,
	}
	if err := casheS.SetWithReplicate(context.Background(), args); err != nil {
		t.Error(err)
	}
}

func TestCacheSGetWithRemote(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{

		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "cacheID",
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Remote: true,
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1GetItem: func(_ *context.Context, args, reply interface{}) error {

				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.CacheSv1GetItem, clientconn)

	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Second * 1,
				StaticTTL: true,
				OnEvicted: func(itmID string, value interface{}) {

				},
			},
		})
	chS := &CacheS{
		cfg:     cfg,
		dm:      dm,
		tCache:  tscache,
		connMgr: connMgr,
	}
	var reply interface{} = "str"
	if err := chS.V1GetItemWithRemote(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestCacheSGetWithRemoteFalse(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{

		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "cacheID",
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Remote: false,
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1GetItem: func(_ *context.Context, args, reply interface{}) error {

				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.CacheSv1GetItem, clientconn)

	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Second * 1,
				StaticTTL: true,
				OnEvicted: func(itmID string, value interface{}) {

				},
			},
		})
	chS := &CacheS{
		cfg:     cfg,
		dm:      dm,
		tCache:  tscache,
		connMgr: connMgr,
	}

	var reply interface{} = "str"
	if err := chS.V1GetItemWithRemote(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}
