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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCallCacheForFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg, nil)
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
	th := &engine.ThresholdProfile{
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
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	admS := NewAdminSv1(cfg, dm, connMgr, fltrs, nil)
	admS.cfg.AdminSCfg().CachesConns = []string{"*internal"}
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
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	admS := NewAdminSv1(cfg, dm, connMgr, fltrs, nil)
	admS.cfg.AdminSCfg().CachesConns = []string{"*internal"}
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
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	admS := NewAdminSv1(cfg, dm, connMgr, fltrs, nil)
	admS.cfg.AdminSCfg().CachesConns = []string{"*internal"}
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
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	admS := NewAdminSv1(cfg, dm, connMgr, fltrs, nil)
	admS.cfg.AdminSCfg().CachesConns = []string{"*internal"}
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
