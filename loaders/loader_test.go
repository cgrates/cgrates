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

package loaders

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

func TestRemoveFromDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	for _, lType := range []string{utils.MetaAttributes, utils.MetaResources, utils.MetaFilters, utils.MetaStats,
		utils.MetaThresholds, utils.MetaRoutes, utils.MetaChargers,
		utils.MetaRateProfiles, utils.MetaActionProfiles, utils.MetaAccounts} {
		if err := removeFromDB(context.Background(), dm, lType, true, false, profileTest{utils.Tenant: "cgrates.org", utils.ID: "ID"}); err != utils.ErrNotFound &&
			err != utils.ErrDSPProfileNotFound && err != utils.ErrDSPHostNotFound {
			t.Error(err)
		}
	}
	if err := removeFromDB(context.Background(), dm, utils.MetaRateProfiles, true, true, &utils.RateProfile{Tenant: "cgrates.org", ID: "ID", Rates: map[string]*utils.Rate{"RT1": {}}}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := removeFromDB(context.Background(), dm, utils.EmptyString, true, false, profileTest{utils.Tenant: "cgrates.org", utils.ID: "ID"}); err != nil {
		t.Error(err)
	}
}

func testDryRunWithData(lType string, data profile) (string, error) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	err := dryRun(context.Background(), lType, "test", data)
	return buf.String(), err
}

func testDryRun(t *testing.T, lType string) string {
	buf, err := testDryRunWithData(lType, profileTest{
		utils.Tenant: "cgrates.org",
		utils.ID:     "ID",
	})
	if err != nil {
		t.Fatal(lType, err)
	}
	return buf
}

func newOrderNavMap(mp utils.MapStorage) (o *utils.OrderedNavigableMap) {
	o = utils.NewOrderedNavigableMap()
	for k, v := range mp {
		o.SetAsSlice(utils.NewFullPath(k), []*utils.DataNode{utils.NewLeafNode(v)})
	}
	return
}
func TestDryRun(t *testing.T) {
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: AttributeProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaAttributes); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: ResourceProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaResources); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	if rplyLog, err := testDryRunWithData(utils.MetaFilters, &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "ID",
	}); err != nil {
		t.Fatal(err)
	} else if expLog := "[INFO] <LoaderS-test> DRY_RUN: Filter: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"Rules\":[]}"; !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: StatsQueueProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaStats); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: ThresholdProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaThresholds); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: RouteProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaRoutes); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: ChargerProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaChargers); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: RateProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaRateProfiles); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: ActionProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaActionProfiles); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: Accounts: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		testDryRun(t, utils.MetaAccounts); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	expErrMsg := `empty RSRParser in rule: <>`
	if _, err := testDryRunWithData(utils.MetaFilters, &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "ID",
		Rules:  []*engine.FilterRule{{}},
	}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}
}

func TestSetToDBWithDBError(t *testing.T) {
	if err := setToDB(context.Background(), nil, utils.MetaAttributes, newProfileFunc(utils.MetaAttributes)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaResources, newProfileFunc(utils.MetaResources)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaStats, newProfileFunc(utils.MetaStats)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaThresholds, newProfileFunc(utils.MetaThresholds)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaChargers, newProfileFunc(utils.MetaChargers)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaActionProfiles, newProfileFunc(utils.MetaActionProfiles)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaFilters, newProfileFunc(utils.MetaFilters)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaRoutes, newProfileFunc(utils.MetaRoutes)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaRateProfiles, newProfileFunc(utils.MetaRateProfiles)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaAccounts, newProfileFunc(utils.MetaAccounts)(), true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	expErrMsg := `empty RSRParser in rule: <>`
	if err := setToDB(context.Background(), nil, utils.MetaFilters, &engine.Filter{Rules: []*engine.FilterRule{{Type: "*"}}}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}
}

func TestSetToDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	v1 := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
	if err := setToDB(context.Background(), dm, utils.MetaAttributes, v1, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v1, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v1), utils.ToJSON(prf))
	}

	v2 := &engine.ResourceProfile{Tenant: "cgrates.org", ID: "ID"}
	if err := setToDB(context.Background(), dm, utils.MetaResources, v2, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetResourceProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v2, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v2), utils.ToJSON(prf))
	}

	v3 := &engine.StatQueueProfile{Tenant: "cgrates.org", ID: "ID"}
	if err := setToDB(context.Background(), dm, utils.MetaStats, v3, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetStatQueueProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v3, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v3), utils.ToJSON(prf))
	}

	v4 := &engine.ThresholdProfile{Tenant: "cgrates.org", ID: "ID"}
	if err := setToDB(context.Background(), dm, utils.MetaThresholds, v4, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetThresholdProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v4, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v4), utils.ToJSON(prf))
	}

	v5 := &engine.ChargerProfile{Tenant: "cgrates.org", ID: "ID"}
	if err := setToDB(context.Background(), dm, utils.MetaChargers, v5, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetChargerProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v5, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v5), utils.ToJSON(prf))
	}

	v7 := &engine.ActionProfile{Tenant: "cgrates.org", ID: "ID", Targets: map[string]utils.StringSet{}}
	if err := setToDB(context.Background(), dm, utils.MetaActionProfiles, v7, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetActionProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v7, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v7), utils.ToJSON(prf))
	}

	v8 := &engine.Filter{Tenant: "cgrates.org", ID: "ID", Rules: make([]*engine.FilterRule, 0)}
	v8.Compile()
	if err := setToDB(context.Background(), dm, utils.MetaFilters, v8, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetFilter(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v8, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v8), utils.ToJSON(prf))
	}

	v9 := &engine.RouteProfile{Tenant: "cgrates.org", ID: "ID"}
	if err := setToDB(context.Background(), dm, utils.MetaRoutes, v9, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetRouteProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v9, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v9), utils.ToJSON(prf))
	}

	v11 := &utils.RateProfile{Tenant: "cgrates.org", ID: "ID", Rates: map[string]*utils.Rate{}, MinCost: utils.NewDecimal(0, 0), MaxCost: utils.NewDecimal(0, 0)}
	if err := setToDB(context.Background(), dm, utils.MetaRateProfiles, v11, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetRateProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v11, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v11), utils.ToJSON(prf))
	}

	v12 := &utils.Account{Tenant: "cgrates.org", ID: "ID", Balances: map[string]*utils.Balance{}, Opts: make(map[string]any)}
	if err := setToDB(context.Background(), dm, utils.MetaAccounts, v12, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetAccount(context.Background(), "cgrates.org", "ID"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v12, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v12), utils.ToJSON(prf))
	}

	v13 := &utils.RateProfile{Tenant: "cgrates.org", ID: "ID", Rates: map[string]*utils.Rate{}, MinCost: utils.NewDecimal(0, 0), MaxCost: utils.NewDecimal(0, 0)}
	if err := setToDB(context.Background(), dm, utils.MetaRateProfiles, v13, true, true); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetRateProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v13, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v13), utils.ToJSON(prf))
	}
}

func TestLoaderProcess(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, nil)
	expLd := &loader{
		cfg:       cfg,
		ldrCfg:    cfg.LoaderCfg()[0],
		dm:        dm,
		filterS:   fS,
		connMgr:   cM,
		dataCache: cache,
		Locker:    newLocker(cfg.LoaderCfg()[0].GetLockFilePath(), cfg.LoaderCfg()[0].ID),
	}
	if !reflect.DeepEqual(expLd, ld) {
		t.Errorf("Expected: %+v, received: %+v", expLd, ld)
	}

	expErrMsg := `unsupported loader action: <"notSupported">`
	if err := ld.process(context.Background(), nil, utils.MetaAttributes, "notSupported",
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}

	if err := ld.process(context.Background(), nil, utils.MetaAttributes, utils.MetaParse,
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err != nil {
		t.Error(err)
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	if err := ld.process(context.Background(), profileTest{utils.Tenant: "cgrates.org", utils.ID: "ID"}, utils.MetaAttributes, utils.MetaDryRun,
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err != nil {
		t.Error(err)
	}

	if expLog, rplyLog := "[INFO] <LoaderS-*default> DRY_RUN: AttributeProfile: {\"ID\":\"ID\",\"Tenant\":\"cgrates.org\"}",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	v1 := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
	if err := ld.process(context.Background(), v1, utils.MetaAttributes, utils.MetaStore,
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v1, prf) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(v1), utils.ToJSON(prf))
	}
	if err := ld.process(context.Background(), v1, utils.MetaAttributes, utils.MetaRemove,
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

type ccMock map[string]func(ctx *context.Context, args any, reply any) error

func (ccM ccMock) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM[serviceMethod]; has {
		return call(ctx, args, reply)
	}
	return rpcclient.ErrUnsupporteServiceMethod
}

func TestLoaderProcessCallCahe(t *testing.T) {
	var reloadCache, clearCache any
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCache)
	tntID := "cgrates.org:ID"
	iCh := make(chan birpc.ClientConnector, 1)
	iCh <- ccMock{
		utils.CacheSv1ReloadCache: func(_ *context.Context, args, _ any) error { reloadCache = args; return nil },
		utils.CacheSv1Clear:       func(_ *context.Context, args, _ any) error { clearCache = args; return nil },
	}
	cM.AddInternalConn(connID, utils.CacheSv1, iCh)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, []string{connID})
	{
		v := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
		if err := ld.process(context.Background(), v, utils.MetaAttributes, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			AttributeProfileIDs: []string{tntID}}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			CacheIDs: []string{utils.CacheAttributeFilterIndexes}}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}
	{
		v := &engine.ResourceProfile{Tenant: "cgrates.org", ID: "ID"}
		if err := ld.process(context.Background(), v, utils.MetaResources, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetResourceProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			ResourceProfileIDs: []string{tntID}, ResourceIDs: []string{tntID}}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			CacheIDs: []string{utils.CacheResourceFilterIndexes}}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}
	{
		v := &engine.StatQueueProfile{Tenant: "cgrates.org", ID: "ID"}
		if err := ld.process(context.Background(), v, utils.MetaStats, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetStatQueueProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			StatsQueueProfileIDs: []string{tntID}, StatsQueueIDs: []string{tntID}}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			CacheIDs: []string{utils.CacheStatFilterIndexes}}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}

	{
		v := &engine.ThresholdProfile{Tenant: "cgrates.org", ID: "ID"}
		if err := ld.process(context.Background(), v, utils.MetaThresholds, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetThresholdProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			ThresholdProfileIDs: []string{tntID}, ThresholdIDs: []string{tntID}}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			CacheIDs: []string{utils.CacheThresholdFilterIndexes}}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}

	{
		v := &engine.RouteProfile{Tenant: "cgrates.org", ID: "ID"}
		if err := ld.process(context.Background(), v, utils.MetaRoutes, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetRouteProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			RouteProfileIDs: []string{tntID}}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			CacheIDs: []string{utils.CacheRouteFilterIndexes}}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}

	{
		v := &engine.ChargerProfile{Tenant: "cgrates.org", ID: "ID"}
		if err := ld.process(context.Background(), v, utils.MetaChargers, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetChargerProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			ChargerProfileIDs: []string{tntID}}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			CacheIDs: []string{utils.CacheChargerFilterIndexes}}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}

	{
		v := &utils.RateProfile{Tenant: "cgrates.org", ID: "ID", Rates: map[string]*utils.Rate{}, MinCost: utils.NewDecimal(0, 0), MaxCost: utils.NewDecimal(0, 0)}
		if err := ld.process(context.Background(), v, utils.MetaRateProfiles, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetRateProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			RateProfileIDs: []string{tntID}}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			CacheIDs: []string{utils.CacheRateProfilesFilterIndexes, utils.CacheRateFilterIndexes}}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}

	{
		v := &engine.ActionProfile{Tenant: "cgrates.org", ID: "ID", Targets: map[string]utils.StringSet{}}
		if err := ld.process(context.Background(), v, utils.MetaActionProfiles, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetActionProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			ActionProfileIDs: []string{tntID}}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			CacheIDs: []string{utils.CacheActionProfiles, utils.CacheActionProfilesFilterIndexes}}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}

	reloadCache, clearCache = nil, nil

	{
		v := &engine.Filter{Tenant: "cgrates.org", ID: "ID", Rules: make([]*engine.FilterRule, 0)}
		if err := ld.process(context.Background(), v, utils.MetaFilters, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetFilter(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		exp := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
			FilterIDs: []string{tntID}}
		if !reflect.DeepEqual(exp, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
		}
		if !reflect.DeepEqual(nil, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(nil), utils.ToJSON(clearCache))
		}
	}

	reloadCache, clearCache = nil, nil

	{
		v := &utils.Account{Tenant: "cgrates.org", ID: "ID", Balances: map[string]*utils.Balance{}, Opts: make(map[string]any)}
		if err := ld.process(context.Background(), v, utils.MetaAccounts, utils.MetaStore,
			map[string]any{utils.MetaCache: utils.MetaReload}, true, false); err != nil {
			t.Error(err)
		}
		if prf, err := dm.GetAccount(context.Background(), "cgrates.org", "ID"); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
		expReload := &utils.AttrReloadCacheWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
		}
		if !reflect.DeepEqual(expReload, reloadCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expReload), utils.ToJSON(reloadCache))
		}
		expClear := &utils.AttrCacheIDsWithAPIOpts{
			CacheIDs: []string{utils.CacheAccounts, utils.CacheAccountsFilterIndexes},
			APIOpts: map[string]any{
				utils.MetaCache: utils.MetaReload,
			},
		}
		if !reflect.DeepEqual(expClear, clearCache) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(expClear), utils.ToJSON(clearCache))
		}
	}
}

func TestLoaderProcessData(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, nil)

	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	if err := ld.processData(context.Background(), NewStringCSV(`cgrates.org,ID
cgrates.org,ID2`, utils.CSVSep, -1), fc, utils.MetaAttributes, utils.MetaStore,
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else {
		v := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
		if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID2", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else {
		v := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID2"}
		if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
	}
}

type mockReader struct{}

func (mockReader) Read([]byte) (int, error) { return 0, utils.ErrNotFound }

func TestLoaderProcessDataErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, nil)

	fc := []*config.FCTemplate{
		{Filters: []string{"*string"}},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	expErrMsg := "inline parse error for string: <*string>"
	if err := ld.processData(context.Background(), NewStringCSV(`cgrates.org,ID
cgrates.org,ID2`, utils.CSVSep, -1), fc, utils.MetaAttributes, utils.MetaStore,
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %q, received: %v", expErrMsg, err)
	}

	fc = []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	expErrMsg = `unsupported loader action: <"notSupported">`
	if err := ld.processData(context.Background(), NewStringCSV(`cgrates.org,ID
cgrates.org,ID2`, utils.CSVSep, -1), fc, utils.MetaAttributes, "notSupported",
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %q, received: %v", expErrMsg, err)
	}

	if err := ld.processData(context.Background(), &CSVFile{csvRdr: csv.NewReader(mockReader{})}, fc, utils.MetaAttributes, "notSupported",
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err != utils.ErrNotFound {
		t.Errorf("Expected: %q, received: %v", utils.ErrNotFound, err)
	}
}

func TestLoaderProcessFileURL(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, nil)

	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}

	mux := http.NewServeMux()
	mux.Handle("/ok/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte(`cgrates.org,ID`)) }))
	s := httptest.NewServer(mux)
	defer s.Close()
	runtime.Gosched()

	if err := ld.processFile(context.Background(), &config.LoaderDataType{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}, s.URL+"/ok", utils.EmptyString, utils.MetaStore,
		map[string]any{utils.MetaCache: utils.MetaNone}, true, urlProvider{}); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else {
		v := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
		if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
	}

	if err := ld.processFile(context.Background(), &config.LoaderDataType{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}, s.URL+"/notFound", utils.EmptyString, utils.MetaStore,
		map[string]any{utils.MetaCache: utils.MetaNone}, true, urlProvider{}); err != utils.ErrNotFound {
		t.Errorf("Expected: %v, received: %v", utils.ErrNotFound, err)
	}

}

type mockLock struct{}

// lockFolder will attempt to lock the folder by creating the lock file
func (mockLock) Lock() error                { return utils.ErrExists }
func (mockLock) Unlock() (_ error)          { return }
func (mockLock) Locked() (_ bool, _ error)  { return true, utils.ErrExists }
func (mockLock) IsLockFile(string) (_ bool) { return }

func TestLoaderProcessIFile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessIFileIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	tmpOut, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessIFileOut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpOut)
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  tmpIn,
		TpOutDir: tmpOut,
		Data: []*config.LoaderDataType{
			{
				Type:     utils.MetaAttributes,
				Filename: utils.AttributesCsv,
				Fields:   fc,
			},
		},
		FieldSeparator: utils.FieldsSep,
		Action:         utils.MetaStore,
		Opts: &config.LoaderSOptsCfg{
			WithIndex: true,
			Cache:     utils.MetaNone,
		},
	}, dm, cache, fS, cM, nil)
	expErrMsg := fmt.Sprintf(`rename %s/Chargers.csv %s/Chargers.csv: no such file or directory`, tmpIn, tmpOut)
	if err := ld.processIFile(utils.ChargersCsv); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}

	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ld.processIFile(utils.AttributesCsv); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else {
		v := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
		if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
	}

	if _, err := os.Stat(path.Join(tmpIn, utils.AttributesCsv)); err == nil {
		t.Errorf("Expected file to be moved")
	} else if !os.IsNotExist(err) {
		t.Error(err)
	}
	if _, err := os.Stat(path.Join(tmpOut, utils.AttributesCsv)); err != nil {
		t.Errorf("Expected file to be moved")
	}

	ld.Locker = mockLock{}
	if err := ld.processIFile(utils.AttributesCsv); err != utils.ErrExists {
		t.Fatal(err)
	}
}

func TestLoaderProcessFolder(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	tmpOut, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderOut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpOut)
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  tmpIn,
		TpOutDir: tmpOut,
		Data: []*config.LoaderDataType{
			{
				Type:     utils.MetaAttributes,
				Filename: utils.AttributesCsv,
				Fields:   fc,
			},
		},
		FieldSeparator: utils.FieldsSep,
		Action:         utils.MetaStore,
		Opts: &config.LoaderSOptsCfg{
			WithIndex: true,
			Cache:     utils.MetaNone,
		},
	}, dm, cache, fS, cM, nil)

	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	f, err = os.Create(path.Join(tmpIn, utils.ChargersCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ld.processFolder(context.Background(),
		map[string]any{utils.MetaCache: utils.MetaNone}, true, true); err != nil {
		t.Fatal(err)
	}

	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else {
		v := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
		if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expected: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
	}

	if _, err := os.Stat(path.Join(tmpIn, utils.AttributesCsv)); err == nil {
		t.Errorf("Expected file to be moved")
	} else if !os.IsNotExist(err) {
		t.Error(err)
	}
	if _, err := os.Stat(path.Join(tmpOut, utils.AttributesCsv)); err != nil {
		t.Errorf("Expected file to be moved")
	}

	if _, err := os.Stat(path.Join(tmpIn, utils.ChargersCsv)); err == nil {
		t.Errorf("Expected file to be moved")
	} else if !os.IsNotExist(err) {
		t.Error(err)
	}
	if _, err := os.Stat(path.Join(tmpOut, utils.ChargersCsv)); err != nil {
		t.Errorf("Expected file to be moved")
	}

	ld.Locker = mockLock{}
	if err := ld.processFolder(context.Background(),
		map[string]any{utils.MetaCache: utils.MetaNone}, true, true); err != utils.ErrExists {
		t.Fatal(err)
	}

	ld.Locker = nopLock{}
	ld.ldrCfg.TpInDir = "http://localhost:0"
	expErrMsg := `path:"http://localhost:0/Attributes.csv" is not reachable`
	if err := ld.processFolder(context.Background(),
		map[string]any{utils.MetaCache: utils.MetaNone}, true, true); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}
}

func TestLoaderProcessFolderErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	fc := []*config.FCTemplate{
		{Filters: []string{"*string"}},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderErrorsIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	tmpOut, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderErrorsOut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpOut)
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  tmpIn,
		TpOutDir: tmpOut,
		Data: []*config.LoaderDataType{
			{
				Type:     utils.MetaAttributes,
				Filename: utils.AttributesCsv,
				Fields:   fc,
			},
		},
		FieldSeparator: utils.FieldsSep,
		Action:         utils.MetaStore,
		Opts: &config.LoaderSOptsCfg{
			WithIndex: true,
			Cache:     utils.MetaNone,
		},
	}, dm, cache, fS, cM, nil)

	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	expErrMsg := "inline parse error for string: <*string>"
	if err := ld.processFolder(context.Background(),
		map[string]any{utils.MetaCache: utils.MetaNone}, true, true); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}

	if _, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	if _, err := os.Stat(path.Join(tmpIn, utils.AttributesCsv)); err != nil {
		t.Errorf("Expected file to not be moved because of template error: %v", err)
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	if err := ld.processFolder(context.Background(),
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false); err != nil {
		t.Fatal(err)
	}

	if _, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	if expLog, rplyLog := "<LoaderS-test> loaderType: <*attributes> cannot open files, err: inline parse error for string: <*string>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
}

func TestLoaderMoveUnprocessedFilesErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:      "test",
		Enabled: true,
		TpInDir: "notAFolder",
	}, nil, nil, nil, nil, nil)

	expErrMsg := "open notAFolder: no such file or directory"
	if err := ld.moveUnprocessedFiles(); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderMoveUnprocessedFilesErrorsIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	ld.ldrCfg.TpInDir = tmpIn
	ld.ldrCfg.TpOutDir = "notAFolder"
	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}

	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	expErrMsg = fmt.Sprintf("rename %s/Attributes.csv notAFolder/Attributes.csv: no such file or directory", tmpIn)
	if err := ld.moveUnprocessedFiles(); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}
}

func TestLoaderHandleFolder(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		RunDelay: time.Nanosecond,
		TpInDir:  "/tmp/TestLoaderHandleFolder",
		Opts:     &config.LoaderSOptsCfg{},
	}, nil, nil, nil, nil, nil)
	ld.Locker = mockLock{}
	stop := make(chan struct{})
	close(stop)

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	ld.handleFolder(stop)

	if expLog, rplyLog := "[INFO] <LoaderS-test> stop monitoring path </tmp/TestLoaderHandleFolder>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
}

func TestLoaderListenAndServe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		RunDelay: time.Nanosecond,
		TpInDir:  "/tmp/TestLoaderListenAndServe",
		Opts:     &config.LoaderSOptsCfg{},
	}, nil, nil, nil, nil, nil)
	ld.Locker = mockLock{}
	stop := make(chan struct{})
	close(stop)

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	ld.ListenAndServe(stop)
	runtime.Gosched()
	time.Sleep(time.Nanosecond)
	if expLog, rplyLog := "[INFO] Starting <LoaderS-test>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
}

func TestLoaderListenAndServeI(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  "/tmp/TestLoaderListenAndServeI",
		RunDelay: -1,
		Opts:     &config.LoaderSOptsCfg{},
	}, nil, nil, nil, nil, nil)
	ld.Locker = mockLock{}
	stop := make(chan struct{})
	close(stop)

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	ld.ListenAndServe(stop)
	runtime.Gosched()
	time.Sleep(time.Nanosecond)
	if expLog, rplyLog := "[INFO] Starting <LoaderS-test>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
}

func TestLoaderProcessZipErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	fc := []*config.FCTemplate{
		{Filters: []string{"*string"}},
	}
	for _, f := range fc {
		f.ComputePath()
	}

	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  utils.EmptyString,
		TpOutDir: utils.EmptyString,
		Data: []*config.LoaderDataType{
			{
				Type:     utils.MetaAttributes,
				Filename: utils.AttributesCsv,
				Fields:   fc,
			},
		},
		FieldSeparator: utils.FieldsSep,
		Action:         utils.MetaStore,
		Opts: &config.LoaderSOptsCfg{
			WithIndex: true,
			Cache:     utils.MetaNone,
		},
	}, dm, cache, fS, cM, nil)
	bufz := new(bytes.Buffer)
	w := zip.NewWriter(bufz)
	f, err := w.Create(utils.AttributesCsv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(`cgrates.org,ID`)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := zip.NewReader(bytes.NewReader(bufz.Bytes()), int64(bufz.Len()))
	if err != nil {
		t.Fatal(err)
	}

	expErrMsg := "inline parse error for string: <*string>"
	if err := ld.processZip(context.Background(),
		map[string]any{utils.MetaCache: utils.MetaNone}, true, true, r); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expected: %v, received: %v", expErrMsg, err)
	}

	if _, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)
	if err := ld.processZip(context.Background(),
		map[string]any{utils.MetaCache: utils.MetaNone}, true, false, r); err != nil {
		t.Fatal(err)
	}

	if _, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	if expLog, rplyLog := "<LoaderS-test> loaderType: <*attributes> cannot open files, err: inline parse error for string: <*string>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	ld.Locker = mockLock{}
	if err := ld.processZip(context.Background(),
		map[string]any{utils.MetaCache: utils.MetaNone}, true, true, r); err != utils.ErrExists {
		t.Fatal(err)
	}

}
