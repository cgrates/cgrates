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
package v1

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestCacheLoadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}

	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
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
	if err := cache.GetItemIDs(context.Background(), argsGetItem,
		&replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestCacheReloadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}

	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
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
	if err := cache.GetItemIDs(context.Background(), argsGetItem,
		&replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestCacheSetAndRemoveItems(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestCacheSetAndRemoveItems1", nil, nil, true, utils.NonTransactional)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestCacheSetAndRemoveItems2", nil, nil, true, utils.NonTransactional)
	ch.SetWithoutReplicate(utils.CacheAttributeProfiles, "cgrates.org:TestCacheSetAndRemoveItems3", nil, nil, true, utils.NonTransactional)
	cache := NewCacheSv1(ch)

	argsRemItm := &utils.AttrReloadCacheWithAPIOpts{
		AttributeProfileIDs: []string{"cgrates.org:TestCacheSetAndRemoveItems1",
			"cgrates.org:TestCacheSetAndRemoveItems2", "cgrates.org:TestCacheSetAndRemoveItems3"},
	}
	var reply string
	if err := cache.RemoveItems(context.Background(), argsRemItm, &reply); err != nil {
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
	if err := cache.HasItem(context.Background(), argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}
	argsHasItem.ArgsGetCacheItem.ItemID = "cgrates.org:TestCacheSetAndRemoveItems2"
	if err := cache.HasItem(context.Background(), argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}
	argsHasItem.ArgsGetCacheItem.ItemID = "cgrates.org:TestCacheSetAndRemoveItems3"
	if err := cache.HasItem(context.Background(), argsHasItem, &replyBool); err != nil {
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
	if err := cache.GetItemIDs(context.Background(), argsGetItem, &replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestCacheSv1GetItem(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cacheService := engine.NewCacheS(cfg, dm, nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	itemKey := "cgrates.org:TestCacheSv1_GetItem"
	cacheService.SetWithoutReplicate(utils.CacheAttributeProfiles, itemKey, "testValue", nil, true, utils.NonTransactional)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  itemKey,
		},
	}
	var reply any
	err := cache.GetItem(context.Background(), args, &reply)
	if err == nil {
		t.Errorf("NOT_FOUND")
	} else if reply == "testValue" {
		t.Errorf("expected reply to be 'testValue', got %v", reply)
	}
	args.ArgsGetCacheItem.ItemID = "nonexistentItem"
	err = cache.GetItem(context.Background(), args, &reply)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("expected error %v, got %v", utils.ErrNotFound, err)
	}
}

func TestCacheSv1GetItemWithRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cacheService := engine.NewCacheS(cfg, dm, nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	itemKey := "cgrates.org:TestCacheSv1_GetItemWithRemote"
	cacheService.SetWithoutReplicate(utils.CacheAttributeProfiles, itemKey, "testValue", nil, true, utils.NonTransactional)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  itemKey,
		},
	}
	var reply any
	err := cache.GetItemWithRemote(context.Background(), args, &reply)
	args.ArgsGetCacheItem.ItemID = "nonexistentItem"
	err = cache.GetItemWithRemote(context.Background(), args, &reply)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("expected error %v, got %v", utils.ErrNotFound, err)
	}
}

func TestCacheSv1GetItemExpiryTime(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	ctx := context.Background()
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "TestCache",
			ItemID:  "TestItem",
		},
	}
	var reply time.Time
	err := cache.GetItemExpiryTime(ctx, args, &reply)
	err = cache.GetItemExpiryTime(nil, args, &reply)
	if err == nil {
		t.Errorf("expected an error when context is nil")
	}
	args.ArgsGetCacheItem.CacheID = ""
	err = cache.GetItemExpiryTime(ctx, args, &reply)
	if err == nil {
		t.Errorf("expected an error when CacheID is empty")
	}
	args.ArgsGetCacheItem.CacheID = "TestCache"
	args.ArgsGetCacheItem.ItemID = ""
	err = cache.GetItemExpiryTime(ctx, args, &reply)
	if err == nil {
		t.Errorf("expected an error when ItemID is empty")
	}
}

func TestCacheSv1RemoveItem(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	ctx := context.Background()
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: "TestCache",
			ItemID:  "TestItem",
		},
	}
	var reply string
	err := cache.RemoveItem(ctx, args, &reply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	args.ArgsGetCacheItem.CacheID = "TestCache"
	args.ArgsGetCacheItem.ItemID = ""
	err = cache.RemoveItem(ctx, args, &reply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCacheSv1Clear(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	ctx := context.Background()
	args := &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: []string{"TestCache1", "TestCache2"},
	}
	var reply string
	err := cache.Clear(ctx, args, &reply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	args.CacheIDs = []string{}
	err = cache.Clear(ctx, args, &reply)
	if err != nil {
		t.Errorf("expected an error when CacheIDs is empty")
	}
}

func TestCacheSv1PrecacheStatus(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	ctx := context.Background()
	args := &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: []string{"TestCache1", "TestCache2"},
	}
	rply := make(map[string]string)
	err := cache.PrecacheStatus(ctx, args, &rply)
	if err == nil {
		t.Errorf("expected error, got %v", err)
	}
	args.CacheIDs = []string{}
	err = cache.PrecacheStatus(ctx, args, &rply)
	if err != nil {
		t.Errorf("expected no error")
	}
}

func TestCacheSv1HasGroup(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	ctx := context.Background()
	args := &utils.ArgsGetGroupWithAPIOpts{}
	var rply bool
	err := cache.HasGroup(ctx, args, &rply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	args.GroupID = "NonExistingGroup"
	err = cache.HasGroup(ctx, args, &rply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if rply {
		t.Errorf("expected group 'NonExistingGroup' to not exist, but HasGroup returned true")
	}
}

func TestCacheSv1RemoveGroup(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	ctx := context.Background()
	args := &utils.ArgsGetGroupWithAPIOpts{}
	var rply string
	err := cache.RemoveGroup(ctx, args, &rply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if rply == "" {
		t.Errorf("expected a non-empty reply after removing group, but got an empty string")
	}
	args.GroupID = "NonExistingGroup"
	err = cache.RemoveGroup(ctx, args, &rply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if rply == "" {
		t.Errorf("expected a non-empty reply for non-existing group, but got an empty string")
	}
}

func TestCacheSv1ReplicateSet(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	ctx := context.Background()
	args := &utils.ArgCacheReplicateSet{}
	var reply string
	err := cache.ReplicateSet(ctx, args, &reply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reply == "" {
		t.Errorf("expected a non-empty reply after replicating set, but got an empty string")
	}
	err = cache.ReplicateSet(nil, args, &reply)
	if err != nil {
		t.Errorf("expected no error with nil context, got %v", err)
	}
}

func TestCacheSv1ReplicateRemove(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	ctx := context.Background()
	args := &utils.ArgCacheReplicateRemove{}
	var reply string
	err := cache.ReplicateRemove(ctx, args, &reply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reply == "" {
		t.Errorf("expected a non-empty reply after removing cache key, but got an empty string")
	}
	err = cache.ReplicateRemove(nil, args, &reply)
	if err != nil {
		t.Errorf("expected no error with nil context, got %v", err)
	}
}

func TestCacheSv1GetCacheStats(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)

	ctx := context.Background()
	args := &utils.AttrCacheIDsWithAPIOpts{}
	var reply map[string]*ltcache.CacheStats

	err := cache.GetCacheStats(ctx, args, &reply)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reply == nil {
		t.Errorf("expected reply to be non-nil, got nil")
	}

}

func TestCacheSv1GetGroupItemIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)

	ctx := context.Background()
	args := &utils.ArgsGetGroupWithAPIOpts{}
	var reply []string

	err := cache.GetGroupItemIDs(ctx, args, &reply)

	if err == nil {
		t.Errorf("NOT_FOUND")
	}
	if reply != nil {
		t.Errorf("expected reply to be non-nil, got nil")
	}

}
