// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func TestDMDBItemsMethods(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.DbCfg().Items[utils.CacheAccounts].TTL = time.Hour
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
	if err := adms.HasDBItem(context.Background(), args, &replyB); err != nil {
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
	if err := adms.GetDBItemIDs(context.Background(), argsGetItem, &reply); err != nil {
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
	var replyMD map[string]*ltcache.CacheStats
	if err := adms.GetDBStats(context.Background(), argsDB, &replyMD); err != nil {
		t.Error(err)
	}
	expcacheStats := map[string]*ltcache.CacheStats{
		utils.CacheAccounts: {Items: 1},
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
	if err := adms.GetDBItem(context.Background(), argsCacheItm, &replyItm); err != nil {
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
	if err := adms.GetDBItemIDs(context.Background(), argsItmIds, &replyItms); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, replyItms) {
		t.Errorf("expected %v, received %v", exp, replyItms)
	}
	var replyExp time.Time
	if err := adms.GetDBItemExpiryTime(context.Background(), argsCacheItm, &replyExp); err != nil {
		t.Error(err)
	} else if now := time.Now(); !replyExp.After(now) || replyExp.After(now.Add(time.Hour)) {
		t.Errorf("expected expiry time within the next hour, received %v", replyExp)
	}
	if err := dm.SetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org",
		map[string]utils.StringSet{
			"*string:*req.Account:1001": utils.NewStringSet([]string{"ATTR_1001"}),
		}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	argsGroup := &utils.ArgsGetGroupDB{
		ArgsGetGroupWithAPIOpts: utils.ArgsGetGroupWithAPIOpts{
			Tenant:  "cgrates.org",
			APIOpts: map[string]any{},
			ArgsGetGroup: utils.ArgsGetGroup{
				CacheID: utils.CacheAttributeFilterIndexes,
				GroupID: "cgrates.org",
			},
		},
	}
	if err := adms.HasDBGroup(context.Background(), argsGroup, &replyB); err != nil {
		t.Error(err)
	} else if !replyB {
		t.Errorf("expected group  in db")
	}
	var replyGrpIDs []string
	expGrpIDs := []string{"cgrates.org:*string:*req.Account:1001"}
	if err := adms.GetDBGroupItemIDs(context.Background(), argsGroup, &replyGrpIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expGrpIDs, replyGrpIDs) {
		t.Errorf("expected %v, received %v", expGrpIDs, replyGrpIDs)
	}
}
