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

package apis

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestCacheHasItemAndGetItem(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestGetAttributeProfile", nil, nil, true, utils.NonTransactional)
	cache := NewCacheSv1(ch)

	var replyBool bool
	argsHasItem := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TestGetAttributeProfile",
		},
	}
	if err := cache.HasItem(nil, argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if !replyBool {
		t.Errorf("Unexpected replyBool returned")
	}

	argsGetItem := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	var reply []string
	expectedRPly := []string{"cgrates.org:TestGetAttributeProfile"}
	if err := cache.GetItemIDs(nil, argsGetItem, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedRPly, reply) {
		t.Errorf("Expected %+v, received %+v", expectedRPly, reply)
	}
}

func TestCacheSetAndRemoveItem(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestCacheSetAndRemoveItem", nil, nil, true, utils.NonTransactional)
	cache := NewCacheSv1(ch)

	var replyBool bool
	argsHasItem := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TestCacheSetAndRemoveItem",
		},
	}
	if err := cache.HasItem(nil, argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if !replyBool {
		t.Errorf("Unexpected replyBool returned")
	}

	argsRemItm := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TestCacheSetAndRemoveItem",
		},
	}
	var reply string
	if err := cache.RemoveItem(nil, argsRemItm, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected replyBool returned")
	}

	argsHasItem = &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TestCacheSetAndRemoveItem",
		},
	}
	if err := cache.HasItem(nil, argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}

	argsGetItem := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	var replyStr []string
	if err := cache.GetItemIDs(nil, argsGetItem, &replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestCacheSetAndRemoveItems(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestCacheSetAndRemoveItems1", nil, nil, true, utils.NonTransactional)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestCacheSetAndRemoveItems2", nil, nil, true, utils.NonTransactional)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestCacheSetAndRemoveItems3", nil, nil, true, utils.NonTransactional)
	cache := NewCacheSv1(ch)

	argsRemItm := &utils.AttrReloadCacheWithAPIOpts{
		AttributeProfileIDs: []string{"cgrates.org:TestCacheSetAndRemoveItems1",
			"cgrates.org:TestCacheSetAndRemoveItems2", "cgrates.org:TestCacheSetAndRemoveItems3"},
	}
	var reply string
	if err := cache.RemoveItems(nil, argsRemItm, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected replyBool returned")
	}

	argsHasItem := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TestCacheSetAndRemoveItems1",
		},
	}
	var replyBool bool
	if err := cache.HasItem(nil, argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}
	argsHasItem.ArgsGetCacheItem.ItemID = "cgrates.org:TestCacheSetAndRemoveItems2"
	if err := cache.HasItem(nil, argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}
	argsHasItem.ArgsGetCacheItem.ItemID = "cgrates.org:TestCacheSetAndRemoveItems3"
	if err := cache.HasItem(nil, argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}

	argsGetItem := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	var replyStr []string
	if err := cache.GetItemIDs(nil, argsGetItem, &replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestCacheClear(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestCacheClearAttributes", nil, nil, true, utils.NonTransactional)
	ch.SetWithoutReplicate(utils.CacheRateProfiles, "cgrates.org:TestCacheClearRates", nil, nil, true, utils.NonTransactional)
	cache := NewCacheSv1(ch)

	argsHasItem := &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: nil,
	}
	var replyString string
	if err := cache.Clear(nil, argsHasItem, &replyString); err != nil {
		t.Error(err)
	} else if replyString != utils.OK {
		t.Errorf("Unexpected replyString returned")
	}

	argsGetItem := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	var replyStr []string
	if err := cache.GetItemIDs(nil, argsGetItem, &replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	argsGetItem = &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheRateProfiles,
		},
	}
	if err := cache.GetItemIDs(nil, argsGetItem, &replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestCacheLoadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	var reply string
	if err := cache.LoadCache(context.Background(), utils.NewAttrReloadCacheWithOpts(),
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexcpected rep[ly returned")
	}

	argsGetItem := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	var replyStr []string
	if err := cache.GetItemIDs(nil, argsGetItem,
		&replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestCacheReloadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	var reply string
	if err := cache.ReloadCache(context.Background(), utils.NewAttrReloadCacheWithOpts(),
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexcpected rep[ly returned")
	}

	argsGetItem := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	var replyStr []string
	if err := cache.GetItemIDs(nil, argsGetItem,
		&replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestGetCacheStats(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestGetCacheStats", nil, nil, true, utils.NonTransactional)
	var reply map[string]*ltcache.CacheStats

	args := &utils.AttrCacheIDsWithAPIOpts{
		Tenant:   "cgrates.org",
		APIOpts:  map[string]any{},
		CacheIDs: []string{utils.CacheAttributeProfiles},
	}
	if err := cache.GetCacheStats(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply[utils.CacheAttributeProfiles].Items != 1 {
		t.Errorf("Expected 1\n but received %v", reply[utils.CacheAttributeProfiles].Items)
	}
}

func TestPrecacheStatus(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	var reply map[string]string

	args := &utils.AttrCacheIDsWithAPIOpts{
		Tenant:   "cgrates.org",
		APIOpts:  map[string]any{},
		CacheIDs: []string{utils.CacheAttributeProfiles},
	}
	if err := cache.PrecacheStatus(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	exp := map[string]string{
		utils.CacheAttributeProfiles: utils.MetaPrecaching,
	}
	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected %v\n but received %v", exp, reply)
	}
}

func TestHasGroup(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	var reply bool

	args := &utils.ArgsGetGroupWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: utils.CacheAttributeProfiles,
			GroupID: "Group",
		},
	}
	if err := cache.HasGroup(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply {
		t.Error("Expected false")
	}
}

func TestGetGroupItemIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestGetCacheStats", nil, []string{"AttrGroup"}, true, utils.NonTransactional)

	var reply []string

	args := &utils.ArgsGetGroupWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: utils.CacheAttributeProfiles,
			GroupID: "AttrGroup",
		},
	}

	if err := cache.GetGroupItemIDs(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	exp := []string{"cgrates.org:TestGetCacheStats"}
	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected %v\n but received %v", exp, reply)
	}
}

func TestRemoveGroup(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	var reply string

	args := &utils.ArgsGetGroupWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: utils.CacheAttributeProfiles,
			GroupID: "AttrGroup",
		},
	}

	if err := cache.RemoveGroup(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected OK\n but received %v", reply)
	}
}

func TestReplicateSet(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	var reply string

	args := &utils.ArgCacheReplicateSet{
		CacheID: utils.CacheAttributeProfiles,
		Tenant:  "cgrates.org",
	}

	if err := cache.ReplicateSet(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected OK\n but received %v", reply)
	}
}

func TestReplicateRemove(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	var reply string

	args := &utils.ArgCacheReplicateRemove{
		CacheID: utils.CacheAttributeProfiles,
		Tenant:  "cgrates.org",
	}

	if err := cache.ReplicateRemove(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected OK\n but received %v", reply)
	}
}

func TestGetItemExpiryTime(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg, nil)
	connMgr := engine.NewConnManager(cfg)
	ch := engine.NewCacheS(cfg, dm, connMgr, nil)
	cache := NewCacheSv1(ch)

	var reply time.Time

	args := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
		},
	}

	if err := cache.GetItemExpiryTime(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}
