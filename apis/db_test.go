// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestDMDBItems(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	adms := NewAdminSv1(cfg, dm, nil, nil, locker)
	if err := dm.SetAccount(context.Background(), &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1001",
	}, false); err != nil {
		t.Error(err)
	}
	var replyB bool
	args := &utils.ArgsGetCacheItemDB{
		ArgsGetCacheItemWithAPIOpts: utils.ArgsGetCacheItemWithAPIOpts{
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheAccounts,
				ItemID:  "cgrates.org:1001",
			},
		},
	}
	if err := adms.HasItem(context.Background(), args, &replyB); err != nil {
		t.Error(err)
	} else if !replyB {
		t.Errorf("expected item to exist db")
	}
	argsGetItem := &utils.ArgsGetCacheItemIDsDB{
		ArgsGetCacheItemIDsWithAPIOpts: utils.ArgsGetCacheItemIDsWithAPIOpts{
			ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
				CacheID: utils.CacheAccounts,
			},
		},
	}
	var reply []string
	erply := []string{"cgrates.org:1001"}
	if err := adms.GetItemIDs(context.Background(), argsGetItem, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(erply, reply) {
		t.Errorf("expected %+v, received %+v", erply, reply)
	}
	argsDB := &utils.AttrDBIDs{
		AttrCacheIDsWithAPIOpts: utils.AttrCacheIDsWithAPIOpts{
			Tenant:   "cgrates.org",
			APIOpts:  map[string]any{},
			CacheIDs: []string{utils.CacheAccounts},
		},
	}
	var replyMD engine.CacheStatsWithMetadata
	if err := adms.GetStats(context.Background(), argsDB, &replyMD); err != nil {
		t.Error(err)
	}
	expcacheStats := engine.CacheStatsWithMetadata{
		CacheStatistics: map[string]*ltcache.CacheStats{
			utils.CacheAccounts: {Items: 1},
		},
		Metadata: map[string]any{
			utils.NodeID: cfg.GeneralCfg().NodeID,
		},
	}
	if !reflect.DeepEqual(expcacheStats, replyMD) {
		t.Errorf("expected %s, received %s", utils.ToJSON(expcacheStats), utils.ToJSON(replyMD))
	}
	expAcc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1001",
	}
	var replyItm any
	argsCacheItm := &utils.ArgsGetCacheItemDB{
		ArgsGetCacheItemWithAPIOpts: utils.ArgsGetCacheItemWithAPIOpts{
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheAccounts,
				ItemID:  "cgrates.org:1001",
			},
		},
	}
	if err := adms.GetItem(context.Background(), argsCacheItm, &replyItm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyItm, expAcc) {
		t.Errorf("expected %s, received %s", utils.ToJSON(expAcc), utils.ToJSON(replyItm))
	}
	argsItmIds := &utils.ArgsGetCacheItemIDsDB{
		ArgsGetCacheItemIDsWithAPIOpts: utils.ArgsGetCacheItemIDsWithAPIOpts{
			ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
				CacheID:      utils.CacheAccounts,
				ItemIDPrefix: "cgrates.org:1001",
			},
		},
	}
	exp := []string{"cgrates.org:1001"}
	var replyItms []string
	if err := adms.GetItemIDs(context.Background(), argsItmIds, &replyItms); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("expected %v, received %v", exp, reply)
	}
}
