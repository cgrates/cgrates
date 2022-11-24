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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

func TestCachesReplicateRemove(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		"id": {
			Replicate: true,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReplicateRemove: func(args, reply interface{}) error {
				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg): clientconn,
	})
	chS := CacheS{
		cfg: cfg,
		dm:  dm,
	}
	SetConnManager(connMgr)
	if err := chS.ReplicateRemove("id", "itm_id"); err != nil {

		t.Error(err)

	}
}

func TestCacheSSetWithReplicate(t *testing.T) {
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

	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReplicateSet: func(args, reply interface{}) error {
				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg): clientconn,
	})
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
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	SetConnManager(connMgr)

	if err := casheS.SetWithReplicate(args); err != nil {
		t.Error(err)
	}
}

func TestCacheSV1GetItemIDs(t *testing.T) {
	args := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		APIOpts: map[string]interface{}{},
		Tenant:  "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID:      "cacheID",
			ItemIDPrefix: "itemID",
		},
	}
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
	tscache.Set("cacheID", "itemID", "", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := &[]string{}
	exp := &[]string{"itemID"}
	if err := chS.V1GetItemIDs(args, reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestCacheSV1HasItem(t *testing.T) {
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		APIOpts: map[string]interface{}{},
		Tenant:  "cgrates.org",
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "cacheID",
			ItemID:  "itemID",
		},
	}
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
	tscache.Set("cacheID", "itemID", "", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := utils.BoolPointer(false)
	if err := chS.V1HasItem(args, reply); err != nil {
		t.Error(err)
	}
}

func TestCacheSV1GetItemWithRemote(t *testing.T) {
	args := &utils.ArgsGetCacheItemWithAPIOpts{

		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "cacheID",
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Remote: true,
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1GetItem: func(args, reply interface{}) error {

				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg): clientconn,
	},
	)
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
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	SetConnManager(connMgr)
	var reply interface{} = "str"
	if err := chS.V1GetItemWithRemote(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestCacheSV1GetItem(t *testing.T) {
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "cacheID",
			ItemID:  "itemID",
		},
	}
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
	tscache.Set("cacheID", "itemID", "value", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	var reply interface{} = ""
	if err := chS.V1GetItem(args, &reply); err != nil {
		t.Error(err)
	}

}

func TestCacheSV1GetItemExpiryTime(t *testing.T) {
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "cacheID",
			ItemID:  "itemID",
		},
	}
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: func(itmID string, value interface{}) {

				},
			},
		})
	tscache.Set("cacheID", "itemID", "value", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := now
	loc, _ := time.LoadLocation("EST")
	exp := now.Add(30 * time.Minute).In(loc).Minute()
	if err := chS.V1GetItemExpiryTime(args, &reply); err != nil {
		t.Error(err)
	} else if reply.Minute() != exp {
		t.Errorf("expected %+v,received %+v", exp, reply)
	}

}

func TestCacheSV1RemoveItem(t *testing.T) {
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "cacheID",
			ItemID:  "itemID",
		},
	}
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: func(itmID string, value interface{}) {

				},
			},
		})
	tscache.Set("cacheID", "itemID", "value", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := "error"
	if err := chS.V1RemoveItem(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %v,received %v", utils.OK, reply)
	}
}

func TestCacheSV1RemoveItems(t *testing.T) {
	args := utils.NewAttrReloadCacheWithOpts()
	args.DestinationIDs = []string{"dest1", "dest2", "dest3"}
	args.ResourceIDs = []string{"res1", "res2", "res3"}
	args.FilterIDs = []string{"fltr1", "filtr2", "filtr3"}
	cfgCache := map[string]*ltcache.CacheConfig{
		utils.CacheDestinations: {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: func(itmID string, value interface{}) {
			},
		},
		utils.CacheResources: {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: func(itmID string, value interface{}) {

			},
		},
		utils.CacheFilters: {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: func(itmID string, value interface{}) {

			},
		},
	}
	args2 := map[string][]string{
		utils.CacheDestinations: {"dest1", "dest2", "dest3"},
		utils.CacheResources:    {"res1", "res2", "res3"},
		utils.CacheFilters:      {"fltr1", "filtr2", "filtr3"},
	}
	tscache := ltcache.NewTransCache(cfgCache)

	for keyId := range cfgCache {
		if itemids, has := args2[keyId]; has {
			for _, itemid := range itemids {
				tscache.Set(keyId, itemid, "value", []string{}, true, "tId")
			}
		}
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}

	reply := "error"

	if err := chS.V1RemoveItems(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %v,received %v", utils.OK, reply)
	}
}

func TestCacheSV1Clear(t *testing.T) {
	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts:  map[string]interface{}{},
		Tenant:   "cgrates.org",
		CacheIDs: []string{"cacheID", "cacheID2", "cacheID3"},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cfgCache := map[string]*ltcache.CacheConfig{
		"cacheID": {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: func(itmID string, value interface{}) {
			},
		},
		"cacheID2": {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: func(itmID string, value interface{}) {

			},
		},
		"cacheID3": {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: func(itmID string, value interface{}) {

			},
		},
	}
	tscache := ltcache.NewTransCache(cfgCache)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := "error"
	if err := chS.V1Clear(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestCacheSV1ReplicateSet(t *testing.T) {
	fltr := &Filter{
		Tenant: "cgrates",
		ID:     "filterID",
	}
	args := &utils.ArgCacheReplicateSet{
		CacheID: "cacheID",
		ItemID:  "itemID",
		Value:   fltr,
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: func(itmID string, value interface{}) {
				},
			}},
	)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := "reply"
	if err := chS.V1ReplicateSet(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %+v,received %+v", utils.OK, reply)
	}
	val, has := chS.tCache.Get(args.CacheID, args.ItemID)

	if !has {
		t.Error("has no value")
	}

	if !reflect.DeepEqual(val, fltr) {
		t.Errorf("expected %+v,received %+v", "", utils.ToJSON(val))
	}
}

func TestCacheSV1GetCacheStats(t *testing.T) {
	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts:  map[string]interface{}{},
		Tenant:   "cgrates.org",
		CacheIDs: []string{"cacheID", "cacheID2", "cacheID3"},
	}
	reply := map[string]*ltcache.CacheStats{}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: func(itmID string, value interface{}) {
				},
			},
			"cacheID2": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: func(itmID string, value interface{}) {
				},
			},
			"cacheID3": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: func(itmID string, value interface{}) {
				},
			},
		},
	)

	for i, id := range args.CacheIDs {

		tscache.Set(id, fmt.Sprintf("%s%d", "item", i), "value", []string{}, true, "tId")
	}
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	exp := map[string]*ltcache.CacheStats{
		"cacheID":  {Items: 1, Groups: 0},
		"cacheID2": {Items: 1, Groups: 0},
		"cacheID3": {Items: 1, Groups: 0},
	}
	if err := chS.V1GetCacheStats(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(reply), utils.ToJSON(exp))
	}

}
