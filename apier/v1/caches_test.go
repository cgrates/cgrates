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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCacheLoadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, nil)
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}

	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)

	var reply string
	if err := cache.LoadCache(utils.NewAttrReloadCacheWithOpts(),
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
	if err := cache.GetItemIDs(argsGetItem,
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
	if err := cache.ReloadCache(utils.NewAttrReloadCacheWithOpts(),
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
	if err := cache.GetItemIDs(argsGetItem,
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
	if err := cache.RemoveItems(argsRemItm, &reply); err != nil {
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
	if err := cache.HasItem(argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}
	argsHasItem.ArgsGetCacheItem.ItemID = "cgrates.org:TestCacheSetAndRemoveItems2"
	if err := cache.HasItem(argsHasItem, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}
	argsHasItem.ArgsGetCacheItem.ItemID = "cgrates.org:TestCacheSetAndRemoveItems3"
	if err := cache.HasItem(argsHasItem, &replyBool); err != nil {
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
	if err := cache.GetItemIDs(argsGetItem, &replyStr); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}
