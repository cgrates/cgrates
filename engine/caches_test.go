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
	"regexp"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestCachesReplicateRemove(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		utils.ReplicatorSv1GetDispatcherHost: {
			Replicate: true,
		},
	}
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateRemove: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg): clientconn,
	})
	chS := CacheS{
		cfg: cfg,
		dm:  dm,
	}
	SetConnManager(connMgr)
	if err := chS.ReplicateRemove(utils.ReplicatorSv1GetDispatcherHost, "itm_id"); err != nil {
		t.Error(err)
	}
}

func TestCacheSSetWithReplicate(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgCacheReplicateSet{
		CacheID:  utils.ReplicatorSv1GetActions,
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
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg): clientconn,
	})
	ltcache := ltcache.NewTransCache(map[string]*ltcache.CacheConfig{
		args.CacheID: {
			MaxItems: 2,
		},
	})
	cacheS := &CacheS{
		dm:     dm,
		cfg:    cfg,
		tCache: ltcache,
	}
	SetConnManager(connMgr)
	if err := cacheS.SetWithReplicate(args); err != nil {
		t.Error(err)
	}

}

func TestCacheSV1GetItemIDs(t *testing.T) {
	args := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		APIOpts: map[string]any{},
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
				OnEvicted: []func(itmID string, value interface{}){
					func(itmID string, value any) {

					},
				},
			},
		})
	tscache.Set("cacheID", "itemID", "", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := &[]string{}
	exp := &[]string{"itemID"}
	if err := chS.V1GetItemIDs(context.Background(), args, reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
	tscache.Remove("cacheID", "itemID", true, utils.NonTransactional)
	if err := chS.V1GetItemIDs(context.Background(), args, reply); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestCacheSV1HasItem(t *testing.T) {
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		APIOpts: map[string]any{},
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
				OnEvicted: []func(itmID string, value interface{}){
					func(itmID string, value any) {

					},
				},
			},
		})
	tscache.Set("cacheID", "itemID", "", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := utils.BoolPointer(false)
	if err := chS.V1HasItem(context.Background(), args, reply); err != nil {
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
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1GetItem: func(ctx *context.Context, args, reply any) error {

				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg): clientconn,
	},
	)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Second * 1,
				StaticTTL: true,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {

					},
				},
			},
		})
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	SetConnManager(connMgr)
	var reply any = "str"
	if err := chS.V1GetItemWithRemote(context.Background(), args, &reply); err != nil {
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
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			},
		})
	tscache.Set("cacheID", "itemID", "value", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	var reply any
	if err := chS.V1GetItem(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if val, cancast := reply.(string); cancast {
		if val != "value" {
			t.Errorf("expected value,received %v", val)
		}
	}
	tscache.Remove("cacheID", "itemID", true, utils.NonTransactional)
	if err := chS.V1GetItem(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestCacheSV1GetItemExpiryTime(t *testing.T) {
	cacheID := "testCache"
	itemID := "testItem"
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			cacheID: {
				MaxItems: -1,
				TTL:      30 * time.Minute,
			},
		})
	chS := &CacheS{
		tCache: tscache,
	}
	chS.tCache.Set(cacheID, itemID, "value", []string{}, true, "")
	want := time.Now().Add(30 * time.Minute)

	var got time.Time
	if err := chS.V1GetItemExpiryTime(context.Background(),
		&utils.ArgsGetCacheItemWithAPIOpts{
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: cacheID,
				ItemID:  itemID,
			},
		}, &got); err != nil {
		t.Fatalf("V1GetItemExpiryTime(%q,%q): got unexpected err=%v", cacheID, itemID, err)
	}
	if diff := want.Sub(got); diff < 0 || diff > time.Millisecond {
		t.Errorf("V1GetItemExpiryTime(%q,%q) = %v, want %v (diff %v, margin 1ms)",
			cacheID, itemID, got.Format(time.StampMicro), want.Format(time.StampMicro), diff)
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
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			},
		})
	tscache.Set("cacheID", "itemID", "value", []string{}, true, "tId")

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	var reply string
	if err := chS.V1RemoveItem(context.Background(), args, &reply); err != nil {
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
			OnEvicted: []func(itmID string, value any){
				func(itmID string, value any) {},
			},
		},
		utils.CacheResources: {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: []func(itmID string, value any){
				func(itmID string, value any) {},
			},
		},
		utils.CacheFilters: {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: []func(itmID string, value any){
				func(itmID string, value any) {},
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
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}

	reply := "error"

	if err := chS.V1RemoveItems(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %v,received %v", utils.OK, reply)
	}
}

func TestCacheSV1Clear(t *testing.T) {
	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts:  map[string]any{},
		Tenant:   "cgrates.org",
		CacheIDs: []string{"cacheID", "cacheID2", "cacheID3"},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cfgCache := map[string]*ltcache.CacheConfig{
		"cacheID": {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: []func(itmID string, value any){
				func(itmID string, value any) {},
			},
		},
		"cacheID2": {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: []func(itmID string, value any){
				func(itmID string, value any) {},
			},
		},
		"cacheID3": {
			MaxItems:  3,
			TTL:       time.Minute * 30,
			StaticTTL: false,
			OnEvicted: []func(itmID string, value any){
				func(itmID string, value any) {},
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
	if err := chS.V1Clear(context.Background(), args, &reply); err != nil {
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
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			}},
	)
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	reply := "reply"
	if err := chS.V1ReplicateSet(context.Background(), args, &reply); err != nil {
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
		APIOpts:  map[string]any{},
		Tenant:   "cgrates.org",
		CacheIDs: []string{"cacheID", "cacheID2", "cacheID3"},
	}
	reply := map[string]*ltcache.CacheStats{}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheID": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			},
			"cacheID2": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			},
			"cacheID3": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
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
	if err := chS.V1GetCacheStats(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(reply), utils.ToJSON(exp))
	}

}

func TestCachesPrecache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		utils.CacheDestinations: {
			Limit:    1,
			Precache: true,
			TTL:      time.Minute * 2,
			Remote:   true,
		},
	}
	pcI := map[string]chan struct{}{
		utils.CacheDestinations: make(chan struct{})}
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:     cfg,
		dm:      dm,
		pcItems: pcI,
	}
	if err := chS.Precache(); err != nil {
		t.Error(err)
	}

}

func TestV1PrecacheStatus(t *testing.T) {
	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts:  map[string]any{},
		Tenant:   "cgrates.org",
		CacheIDs: []string{utils.CacheFilters},
	}
	cfg := config.NewDefaultCGRConfig()

	pcI := map[string]chan struct{}{
		utils.CacheFilters: make(chan struct{}),
	}
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := &CacheS{
		cfg:     cfg,
		dm:      dm,
		pcItems: pcI,
	}

	reply := map[string]string{}
	exp := map[string]string{
		utils.CacheFilters: utils.MetaPrecaching,
	}
	if err := chS.V1PrecacheStatus(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("expected %+v,received %+v", exp, reply)
	}
	args.CacheIDs = []string{}
	if err := chS.V1PrecacheStatus(context.Background(), args, &reply); err == nil {
		t.Error(err)
	}
}

func TestCacheSV1HasGroup(t *testing.T) {
	args := &utils.ArgsGetGroupWithAPIOpts{
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: "cacheId",
			GroupID: "groupId",
		},
		APIOpts: map[string]any{},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheId": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			}},
	)
	tscache.Set("cacheId", "itemId", "value", []string{"groupId"}, true, "tId")
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}

	var reply bool
	if err := chS.V1HasGroup(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Error("expected true,received false")
	}

}

func TestCacheSV1HasGroupItemIDs(t *testing.T) {
	args := &utils.ArgsGetGroupWithAPIOpts{
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: "cacheId",
			GroupID: "groupId",
		},
		APIOpts: map[string]any{},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheId": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			}},
	)
	tscache.Set("cacheId", "itemId", "value", []string{"groupId"}, true, "tId")
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	var reply []string
	exp := []string{"itemId"}
	if err := chS.V1GetGroupItemIDs(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("expected %+v,received %+v", exp, reply)
	}

}

func TestV1RemoveGroup(t *testing.T) {
	args := &utils.ArgsGetGroupWithAPIOpts{
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: "cacheId",
			GroupID: "groupId",
		},
		APIOpts: map[string]any{},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheId": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			}},
	)
	tscache.Set("cacheId", "itemId", "value", []string{"groupId"}, true, "tId")
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	var reply string

	if err := chS.V1RemoveGroup(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %v,received %v", utils.OK, reply)
	}

	if has := tscache.HasGroup(args.CacheID, args.GroupID); has {
		t.Errorf("expected false,received %+v", has)
	}

}

func TestCacheSV1ReplicateRemove(t *testing.T) {
	args := &utils.ArgCacheReplicateRemove{
		CacheID: "cacheID",
		ItemID:  "itemID",
		APIOpts: map[string]any{},
		Tenant:  "cgrates.org",
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheId": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {},
				},
			}},
	)
	tscache.Set(args.CacheID, args.ItemID, "value", []string{"groupId"}, true, "tId")
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	var reply string

	if err := chS.V1ReplicateRemove(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %v,received %v", utils.OK, reply)
	}

	if _, has := tscache.Get(args.CacheID, args.ItemID); has {
		t.Errorf("expected false,received %+v", has)
	}
}

func TestNewCacheS(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		"cacheID": {
			Limit:     3,
			TTL:       2 * time.Minute,
			StaticTTL: true,
			Replicate: true,
		},
	}
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateRemove: func(ctx *context.Context, args, reply any) error {

				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg): clientconn,
	})
	expCacheS := &CacheS{}
	SetConnManager(connMgr)
	if c := NewCacheS(cfg, dm, &CapsStats{}); reflect.DeepEqual(expCacheS, c) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(expCacheS), utils.ToJSON(c))
	}

}

func TestCacheRemoveWithoutReplicate(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheId": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {
					},
				},
			}},
	)
	tscache.Set("cacheID", "itemId", "value", []string{"groupId"}, true, "tId")
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	chS.RemoveWithoutReplicate("cacheID", "itemId", true, "tId")

	if _, has := tscache.Get("cacheID", "itemId"); has {
		t.Error("shouldn't exist")
	}
}
func TestCacheRemoveGroup(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			"cacheId": {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {
					},
				},
			}},
	)
	tscache.Set("cacheID", "itemId", "value", []string{"groupId"}, true, "tId")
	chS := &CacheS{
		cfg:    cfg,
		dm:     dm,
		tCache: tscache,
	}
	chS.RemoveGroup("cacheID", "groupId")
	if _, has := tscache.Get("cacheID", "itemId"); has {
		t.Error("shouldn't exist")
	}

}

func TestUpdateReplicationFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := *Cache
	defer func() {
		*Cache = tmp
	}()
	Cache.Clear(nil)
	Cache = NewCacheS(cfg, nil, nil)
	Cache.tCache = ltcache.NewTransCache(map[string]*ltcache.CacheConfig{
		utils.CacheReplicationHosts: {
			MaxItems: 3,
		},
	})
	objType, objID, connID := "obj", "id", "conn"
	UpdateReplicationFilters("obj", "id", "conn")
	if val, has := Cache.Get(utils.CacheReplicationHosts, objType+objID+utils.ConcatenatedKeySep+connID); !has {
		t.Error("has no value")
	} else if val.(string) != connID {
		t.Errorf("expected %v,received %v", connID, val)
	}
}

func TestReplicateMultipleIDs(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.AttributeSCfg().Enabled = true
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		utils.CacheReplicationHosts: {
			Limit: 3,
		},
	}

	Cache = NewCacheS(cfg, nil, nil)
	connClient := make(chan birpc.ClientConnector, 1)
	connClient <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReloadCache: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): connClient,
	})
	objType := "obj"
	objIds := []string{"rpl1", "rpl2"}
	method := utils.CacheSv1ReloadCache
	args := &utils.AttrReloadCacheWithAPIOpts{
		AccountActionPlanIDs: []string{"accID"},
	}
	if err := replicateMultipleIDs(connMgr, cfg.ApierCfg().CachesConns, false, objType, objIds, method, args); err != nil {
		t.Error(err)
	}
	if err := replicateMultipleIDs(connMgr, cfg.ApierCfg().CachesConns, true, objType, objIds, method, args); err != nil {
		t.Error(err)
	}
}
func TestCachesGetWithRemote(t *testing.T) {
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheTBLTPActionPlans,
			ItemID:  "cacheItem",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheTBLTPActionPlans: {
			Limit:  3,
			Remote: false,
		},
	}
	db := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := NewCacheS(cfg, dm, nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1GetItem: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = utils.OK
				return utils.ErrNotFound
			},
		},
	}

	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg): clientconn,
	})
	SetConnManager(connMgr)
	tpAc := &utils.TPActionTrigger{
		BalanceId:             "id",
		Id:                    "STANDARD_TRIGGERS",
		ThresholdType:         "*min_balance",
		ThresholdValue:        2,
		Recurrent:             false,
		MinSleep:              "0",
		ExpirationDate:        "date",
		BalanceType:           "*monetary",
		BalanceDestinationIds: "FS_USERS",
		ActionsId:             "LOG_WARNING",
		Weight:                10,
	}
	chS.tCache.Set(args.CacheID, "cacheItem", tpAc, []string{"cacheItem"}, true, utils.NonTransactional)
	if val, err := chS.GetWithRemote(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, tpAc) {
		t.Errorf("expected %v,received %v", utils.ToJSON(tpAc), utils.ToJSON(val))
	}
	chS.tCache.Remove(utils.CacheTBLTPActionPlans, "cacheItem", true, utils.NonTransactional)
	if _, err := chS.GetWithRemote(args); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected %v,received %v", utils.ErrNotFound, err)
	}
	cfg.DataDbCfg().Items[utils.CacheTBLTPActionPlans].Remote = true
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	/*
		if _, err = chS.GetWithRemote(args); err == nil || err != utils.ErrNotFound {
			t.Errorf("expected %v,received %v", utils.ErrNotFound, err)
		}
	*/
}

func TestV1LoadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	defer func() {
		dm = tmpDm
	}()
	attr := &utils.AttrReloadCacheWithAPIOpts{
		Tenant:                  "cgrates.org",
		FilterIDs:               []string{"cgrates.org:FLTR_ID"},
		AttributeFilterIndexIDs: []string{"cgrates.org:*any:*string:*req.Account:1001", "cgrates.org:*any:*string:*req.Account:1002"},
	}
	chS := NewCacheS(cfg, dm, nil)
	var reply string
	if err := chS.V1LoadCache(context.Background(), attr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("reply should  be %v", utils.OK)
	}

}

func TestCacheSBeginTransaction(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, false, false, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, nil)

	expFormat := `........-....-....-....-............`
	rcv := cacheS.BeginTransaction()
	if matched, err := regexp.Match(expFormat, []byte(rcv)); err != nil {
		t.Error(err)
	} else if !matched {
		t.Errorf("Unexpected transaction format, Received <%v>", rcv)
	}

}

func TestCachesV1ReLoadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	defer func() {
		dm = tmpDm
	}()
	attr := &utils.AttrReloadCacheWithAPIOpts{
		Tenant:                  "cgrates.org",
		FilterIDs:               []string{"cgrates.org:FLTR_ID"},
		AttributeFilterIndexIDs: []string{"cgrates.org:*any:*string:*req.Account:1001", "cgrates.org:*any:*string:*req.Account:1002"},
	}
	chS := NewCacheS(cfg, dm, nil)
	var reply string
	if err := chS.V1ReloadCache(context.Background(), attr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("reply should  be %v", utils.OK)
	}

}
