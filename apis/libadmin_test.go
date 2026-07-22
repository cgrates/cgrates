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
)

func TestCallCacheForFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	tnt := "cgrates.org"
	flt := &engine.Filter{
		Tenant: tnt,
		ID:     "FLTR1",
		Rules: []*engine.FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001"},
		}},
	}
	if err := flt.Compile(); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.TODO(), flt, true); err != nil {
		t.Fatal(err)
	}
	th := &utils.ThresholdProfile{
		Tenant:    tnt,
		ID:        "TH1",
		FilterIDs: []string{flt.ID},
	}
	if err := dm.SetThresholdProfile(context.TODO(), th, true); err != nil {
		t.Fatal(err)
	}

	exp := map[string][]string{
		utils.CacheFilters:                {"cgrates.org:FLTR1"},
		utils.CacheThresholdFilterIndexes: {"cgrates.org:*string:*req.Account:1001"},
	}
	rpl, err := composeCacheArgsForFilter(dm, context.TODO(), flt, tnt, flt.TenantID(), map[string][]string{utils.CacheFilters: {"cgrates.org:FLTR1"}})
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rpl, exp) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	flt = &engine.Filter{
		Tenant: tnt,
		ID:     "FLTR1",
		Rules: []*engine.FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1002"},
		}},
	}
	if err := flt.Compile(); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.TODO(), flt, true); err != nil {
		t.Fatal(err)
	}
	exp = map[string][]string{
		utils.CacheFilters:                {"cgrates.org:FLTR1"},
		utils.CacheThresholdFilterIndexes: {"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
	}
	rpl, err = composeCacheArgsForFilter(dm, context.TODO(), flt, tnt, flt.TenantID(), rpl)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rpl, exp) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
}

func TestCallCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	connMgr.SetCache(cacheS)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	dm.SetCache(cacheS)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	admS := NewAdminSv1(cfg, dm, connMgr, fltrs, locker)
	admS.cfg.AdminSCfg().Conns[utils.MetaCaches] = []*config.DynamicConns{
		{ConnIDs: []string{"*internal"}},
	}
	opts := map[string]any{
		utils.MetaCache: utils.MetaNone,
	}
	errExp := "UNSUPPORTED_SERVICE_METHOD"

	// Reload
	if err := admS.CallCache(context.Background(), utils.MetaReload, "cgrates.org", "", "", utils.EmptyString, nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Load
	if err := admS.CallCache(context.Background(), utils.MetaLoad, "cgrates.org", "", "", utils.EmptyString, nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Clear - Thresholds
	if err := admS.CallCache(context.Background(), utils.MetaClear, "cgrates.org", utils.CacheThresholdProfiles, "", utils.EmptyString, nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Clear - Resources
	if err := admS.CallCache(context.Background(), utils.MetaClear, "cgrates.org", utils.CacheResourceProfiles, "", utils.EmptyString, nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Clear - Stats
	if err := admS.CallCache(context.Background(), utils.MetaClear, "cgrates.org", utils.CacheStatQueueProfiles, "", utils.EmptyString, nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestCallCacheForRemoveIndexes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	connMgr.SetCache(cacheS)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	dm.SetCache(cacheS)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	admS := NewAdminSv1(cfg, dm, connMgr, fltrs, locker)
	admS.cfg.AdminSCfg().Conns[utils.MetaCaches] = []*config.DynamicConns{
		{ConnIDs: []string{"*internal"}},
	}
	opts := map[string]any{
		utils.MetaCache: utils.MetaNone,
	}
	errExp := "UNSUPPORTED_SERVICE_METHOD"

	// Reload
	if err := admS.callCacheForRemoveIndexes(context.Background(), utils.MetaReload, "cgrates.org", "", nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Load
	if err := admS.callCacheForRemoveIndexes(context.Background(), utils.MetaLoad, "cgrates.org", "", nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Remove
	if err := admS.callCacheForRemoveIndexes(context.Background(), utils.MetaRemove, "cgrates.org", "", nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	if err := admS.callCacheForRemoveIndexes(context.Background(), utils.MetaClear, "cgrates.org", "", nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestCallCacheForComputeIndexes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	connMgr.SetCache(cacheS)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	dm.SetCache(cacheS)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	admS := NewAdminSv1(cfg, dm, connMgr, fltrs, locker)
	admS.cfg.AdminSCfg().Conns[utils.MetaCaches] = []*config.DynamicConns{
		{ConnIDs: []string{"*internal"}},
	}
	opts := map[string]any{
		utils.MetaCache: utils.MetaNone,
	}
	errExp := "UNSUPPORTED_SERVICE_METHOD"

	// Reload
	if err := admS.callCacheForComputeIndexes(context.Background(), utils.MetaReload, "cgrates.org", nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Load
	if err := admS.callCacheForComputeIndexes(context.Background(), utils.MetaLoad, "cgrates.org", nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Remove
	if err := admS.callCacheForComputeIndexes(context.Background(), utils.MetaRemove, "cgrates.org", nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	if err := admS.callCacheForComputeIndexes(context.Background(), utils.MetaClear, "cgrates.org", nil, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestCallCacheMultiple(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	connMgr.SetCache(cacheS)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	dm.SetCache(cacheS)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	admS := NewAdminSv1(cfg, dm, connMgr, fltrs, locker)
	admS.cfg.AdminSCfg().Conns[utils.MetaCaches] = []*config.DynamicConns{
		{ConnIDs: []string{"*internal"}},
	}
	opts := map[string]any{
		utils.MetaCache: utils.MetaNone,
	}
	errExp := "UNSUPPORTED_SERVICE_METHOD"

	// Reload
	if err := admS.callCacheMultiple(context.Background(), utils.MetaReload, "cgrates.org", "", []string{"itemID"}, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Load
	if err := admS.callCacheMultiple(context.Background(), utils.MetaLoad, "cgrates.org", "", []string{"itemID"}, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	// Remove
	if err := admS.callCacheMultiple(context.Background(), utils.MetaRemove, "cgrates.org", "", []string{"itemID"}, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

	if err := admS.callCacheMultiple(context.Background(), utils.MetaClear, "cgrates.org", "", []string{"itemID"}, opts); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}
