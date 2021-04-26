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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCacheHasItemAndGetItem(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ch := engine.NewCacheS(cfg, dm, nil)
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
