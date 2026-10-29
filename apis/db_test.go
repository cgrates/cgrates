// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestDBHasGetItem(t *testing.T) {
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
}

func TestDBGetStats(t *testing.T) {
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
	var reply engine.CacheStatsWithMetadata

	args := &utils.AttrCacheIDsDB{
		AttrCacheIDsWithAPIOpts: utils.AttrCacheIDsWithAPIOpts{
			Tenant:   "cgrates.org",
			APIOpts:  map[string]any{},
			CacheIDs: []string{utils.CacheAccounts},
		},
	}
	if err := adms.GetStats(context.Background(), args, &reply); err != nil {
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
	if !reflect.DeepEqual(expcacheStats, reply) {
		t.Errorf("expected %s, received %s", utils.ToJSON(expcacheStats), utils.ToJSON(reply))
	}
}

func TestDBHasGroup(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	adms := NewAdminSv1(cfg, dm, nil, nil, locker)

	var rply bool
	args := &utils.ArgsGetGroupDB{
		ArgsGetGroupWithAPIOpts: utils.ArgsGetGroupWithAPIOpts{
			Tenant:  "cgrates.org",
			APIOpts: map[string]any{},
			ArgsGetGroup: utils.ArgsGetGroup{
				CacheID: utils.CacheAttributeFilterIndexes,
				GroupID: "*string",
			},
		},
	}
	if err := adms.HasGroup(context.Background(), args, &rply); err != nil {
		t.Error(err)
	}
	if rply {
		t.Error("expected to not exist")
	}
}

func TestDBGetGroupItemIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	adms := NewAdminSv1(cfg, dm, nil, nil, locker)

	if err := data.SetIndexesDrv(context.Background(), utils.CacheAttributeFilterIndexes, "*sessions",
		map[string]utils.StringSet{
			"*string:*req.Account:1001": utils.NewStringSet([]string{"ATTR1"}),
		}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	var reply []string

	args := &utils.ArgsGetGroupDB{
		ArgsGetGroupWithAPIOpts: utils.ArgsGetGroupWithAPIOpts{
			Tenant:  "cgrates.org",
			APIOpts: map[string]any{},
			ArgsGetGroup: utils.ArgsGetGroup{
				CacheID: utils.CacheAttributeFilterIndexes,
				GroupID: "*sessions",
			},
		},
	}

	if err := adms.GetGroupItemIDs(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	exp := []string{"*sessions:*string:*req.Account:1001"}
	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected %v\n but received %v", exp, reply)
	}
}

func TestDBGetItemIDs(t *testing.T) {
	args := &utils.ArgsGetCacheItemIDsDB{
		ArgsGetCacheItemIDsWithAPIOpts: utils.ArgsGetCacheItemIDsWithAPIOpts{
			ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
				CacheID:      utils.CacheAccounts,
				ItemIDPrefix: "cgrates.org:1001",
			},
		},
	}
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

	exp := []string{"cgrates.org:1001"}
	var reply []string
	if err := adms.GetItemIDs(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("expected %v, received %v", exp, reply)
	}
}

func TestDBGetItemIDsErr(t *testing.T) {
	args := &utils.ArgsGetCacheItemIDsDB{
		ArgsGetCacheItemIDsWithAPIOpts: utils.ArgsGetCacheItemIDsWithAPIOpts{
			ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
				CacheID:      utils.CacheAccounts,
				ItemIDPrefix: "cgrates.org:1001",
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	adms := NewAdminSv1(cfg, dm, nil, nil, locker)

	var reply []string
	if err := adms.GetItemIDs(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected error %v, teceived error %v", utils.ErrNotFound, err)
	}
}

func TestDBGetItem(t *testing.T) {
	args := &utils.ArgsGetCacheItemDB{
		ArgsGetCacheItemWithAPIOpts: utils.ArgsGetCacheItemWithAPIOpts{
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheAccounts,
				ItemID:  "cgrates.org:1001",
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	adms := NewAdminSv1(cfg, dm, nil, nil, locker)

	acc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1001",
	}
	if err := dm.SetAccount(context.Background(), acc, false); err != nil {
		t.Error(err)
	}

	var reply any
	if err := adms.GetItem(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, acc) {
		t.Errorf("expected %s, received %s", utils.ToJSON(acc), utils.ToJSON(reply))
	}
}

func TestDBGetItemWithRemoteFalse(t *testing.T) {
	args := &utils.ArgsGetCacheItemDB{
		ArgsGetCacheItemWithAPIOpts: utils.ArgsGetCacheItemWithAPIOpts{
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheAccounts,
				ItemID:  "cgrates.org:1001",
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	cfg.DbCfg().DBConns[utils.MetaDefault].RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.DbCfg().Items[args.CacheID].Remote = false
	connMgr := engine.NewConnManager(cfg)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	adms := NewAdminSv1(cfg, dm, connMgr, nil, locker)

	var reply any = utils.OK
	if err := adms.GetItemWithRemote(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected error %v, received error %v", utils.ErrNotFound, err)
	}
}

func TestDBGetItemWithRemoteLocal(t *testing.T) {
	args := &utils.ArgsGetCacheItemDB{
		ArgsGetCacheItemWithAPIOpts: utils.ArgsGetCacheItemWithAPIOpts{
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheAccounts,
				ItemID:  "cgrates.org:1001",
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	cfg.DbCfg().DBConns[utils.MetaDefault].RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.DbCfg().Items[args.CacheID].Remote = true

	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &mockClientConn{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.AdminSv1GetItem: func(_ *context.Context, args, reply any) error {
				*reply.(*any) = utils.ArgsGetCacheItem{
					CacheID: utils.CacheAccounts,
					ItemID:  "cgrates.org:1001",
				}
				return nil
			},
		},
	}
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.AdminSv1GetItem, clientconn)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	cacheS := engine.NewCacheS(cfg, dm, connMgr, nil, locker)
	dm.SetCache(cacheS)
	connMgr.SetCache(cacheS)
	adms := NewAdminSv1(cfg, dm, connMgr, nil, locker)

	exp := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1001",
	}
	if err := data.SetAccountDrv(context.Background(), exp); err != nil {
		t.Error(err)
	}
	var reply any
	if err := adms.GetItemWithRemote(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}
