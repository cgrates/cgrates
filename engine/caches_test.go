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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestCacheSSetWithReplicateTrue(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgCacheReplicateSet{
		CacheID: utils.CacheAccounts,
		ItemID:  "itemID",
		Value: &utils.CachedRPCResponse{
			Result: "reply",
			Error:  nil},
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
				argCache, canCast := args.(*utils.ArgCacheReplicateSet)
				if !canCast {
					return errors.New("cannot cast")
				}
				Cache.Set(nil, argCache.CacheID, argCache.ItemID, argCache.Value, nil, true, utils.EmptyString)
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg), utils.CacheSv1, clientconn)

	stopchan := make(chan struct{}, 1)
	close(stopchan)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.SetWithReplicate(context.Background(), args); err != nil {
		t.Error(err)
	}

	expectedVal := &utils.CachedRPCResponse{
		Result: "reply",
		Error:  nil,
	}
	if val, ok := Cache.Get(utils.CacheAccounts, "itemID"); !ok {
		t.Errorf("Expected value")
	} else {
		valConverted, canCast := val.(*utils.CachedRPCResponse)
		if !canCast {
			t.Error("Should cast")
		}
		if valConverted.Error != nil {
			t.Errorf("Expected error <%v>, Received error <%v>", expectedVal.Error, valConverted.Error)
		}
		if !reflect.DeepEqual(expectedVal.Result, valConverted.Result) {
			t.Errorf("Expected %v, received %v", utils.ToJSON(expectedVal), utils.ToJSON(valConverted))
		}
	}
}

func TestCacheSSetWithReplicateFalse(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgCacheReplicateSet{
		CacheID:  utils.CacheAccounts,
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

	connMgr := NewConnManager(cfg)

	stopchan := make(chan struct{}, 1)
	close(stopchan)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.SetWithReplicate(context.Background(), args); err != nil {
		t.Error(err)
	}
}

func TestCacheSGetWithRemote(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
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
				var valBack string = "test_value_was_set"
				*reply.(*interface{}) = valBack
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.CacheSv1GetItem, clientconn)

	stopchan := make(chan struct{}, 1)
	close(stopchan)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	// first we have to set the value in order to get it from our mock
	cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "test_value_was_set", []string{}, true, utils.NonTransactional)
	var reply interface{}
	expected := "test_value_was_set"
	if err := cacheS.V1GetItemWithRemote(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		strVal, canCast := reply.(string)
		if !canCast {
			t.Error("must be a string")
		}
		if strVal != expected {
			t.Errorf("Expected %v, received %v", expected, strVal)
		}
	}
}

func TestCacheSGetWithRemoteFalse(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{

		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
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

	connMgr := NewConnManager(cfg)

	stopchan := make(chan struct{}, 1)
	close(stopchan)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply interface{} = utils.OK
	if err := cacheS.V1GetItemWithRemote(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}
func TestRemoveWithoutReplicate(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	connMgr := NewConnManager(cfg)
	chS := NewCacheS(cfg, dm, connMgr, nil)

	chS.tCache.Set(utils.CacheAccounts, "itemId", "value", nil, true, utils.NonTransactional)

	chS.RemoveWithoutReplicate(utils.CacheAccounts, "itemId", true, utils.NonTransactional)
	if _, has := chS.tCache.Get(utils.CacheAccounts, "itemId"); has {
		t.Error("This itemId shouldn't exist")
	}

}

func TestV1GetItemExpiryTimeFromCacheErr(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply time.Time
	if err := cacheS.V1GetItemExpiryTime(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestV1GetItemErr(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply interface{}
	if err := cacheS.V1GetItem(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}
func TestV1GetItemIDsErr(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID:      utils.CacheAccounts,
			ItemIDPrefix: "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply []string
	if err := cacheS.V1GetItemIDs(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}
