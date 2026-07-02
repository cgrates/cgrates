/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package thresholds

import (
	"bytes"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

func TestThresholdsV1ProcessEventOK(t *testing.T) {
	defer func() {
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf1 := &utils.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "TH1",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{utils.MetaNone},
		MinHits:          2,
		MaxHits:          5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &utils.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "TH2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{utils.MetaNone},
		MinHits:          0,
		MaxHits:          7,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blocker: false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var reply []string
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Error(err)
	}
}

func TestThresholdsV1ProcessEventPartExecErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
		guardian.Guardian = guardian.New()
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf1 := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH4",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   0,
		MaxHits:   7,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blocker: false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	var reply []string
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ErrMandatoryIeMissing, err)
	}
}

func TestThresholdsV1ProcessEventMissingArgs(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf1 := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   0,
		MaxHits:   7,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blocker: false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	args = &utils.CGREvent{
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var reply []string
	experr := `MANDATORY_IE_MISSING: [ID]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = nil
	experr = `MANDATORY_IE_MISSING: [CGREvent]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		ID:    "V1ProcessEventTest",
		Event: nil,
	}
	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestThresholdsV1GetThresholdOK(t *testing.T) {
	tmpLogger := utils.Logger

	defer func() {
		utils.Logger = tmpLogger
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}

	expTh := utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   0,
	}
	var rplyTh utils.Threshold
	if err := tS.V1GetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
	}, &rplyTh); err != nil {
		t.Error(err)
	} else {
		var snooze time.Time
		rplyTh.Snooze = snooze
		if !reflect.DeepEqual(rplyTh, expTh) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expTh), utils.ToJSON(rplyTh))
		}
	}
}

func TestThresholdsV1GetThresholdNotFoundErr(t *testing.T) {
	tmpLogger := utils.Logger

	defer func() {
		utils.Logger = tmpLogger
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}

	var rplyTh utils.Threshold
	if err := tS.V1GetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH2",
		},
	}, &rplyTh); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsV1GetThresholdsForEventOK(t *testing.T) {
	defer func() {
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	args := &utils.CGREvent{
		ID: "TestGetThresholdsForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	exp := []*utils.Threshold{
		{
			Tenant: "cgrates.org",
			Hits:   0,
			ID:     "TH1",
		},
	}
	var reply []*utils.Threshold
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp[0], reply[0])
	}
}

func TestThresholdsV1GetThresholdsForEventMissingArgs(t *testing.T) {
	defer func() {
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	var args *utils.CGREvent

	experr := `MANDATORY_IE_MISSING: [CGREvent]`
	var reply []*utils.Threshold
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestGetThresholdsForEvent",
		Event:  nil,
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}
}

func TestThresholdsV1GetThresholdIDsOK(t *testing.T) {
	defer func() {
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf1 := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &utils.ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "TH2",
		MinHits: 0,
		MaxHits: 7,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	expIDs := []string{"TH1", "TH2"}
	var reply []string
	if err := tS.V1GetThresholdIDs(context.Background(), &utils.TenantWithAPIOpts{}, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, reply)
		}
	}
}

func TestThresholdsV1GetThresholdIDsGetKeysForPrefixErr(t *testing.T) {
	defer func() {
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	var reply []string
	if err := tS.V1GetThresholdIDs(context.Background(), &utils.TenantWithAPIOpts{}, &reply); err == nil ||
		err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestThresholdsV1ResetThresholdOK(t *testing.T) {
	defer func() {
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	expStored := utils.StringSet{
		"cgrates.org:TH1": {},
	}
	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: <%q>", reply)
	}
	if x, ok := cacheS.Get(utils.CacheThresholds, "cgrates.org:TH1"); !ok {
		t.Errorf("not ok")
	} else if x.(*utils.Threshold).Hits != 0 {
		t.Errorf("expected nr. of hits to be 0, received: <%+v>", x.(*utils.Threshold).Hits)
	} else if !reflect.DeepEqual(tS.storedThresholds, expStored) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expStored, tS.storedThresholds)
	}
}

func TestThresholdsV1ResetThresholdErrNotFound(t *testing.T) {
	defer func() {
		guardian.Guardian = guardian.New()
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH2",
		Hits:   2,
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}

	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
	}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsV1ResetThresholdNegativeStoreIntervalOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	cacheS.Clear(nil)
	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: <%q>", reply)
	}
	if gTH, err := dm.GetThreshold(context.Background(), th.Tenant, th.ID, false, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if gTH.Hits != 0 {
		t.Errorf("expected nr. of hits to be 0, received: <%+v>", th.Hits)
	}
}

func TestThresholdsV1ResetThresholdNegativeStoreIntervalErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, nil, cacheS, filterS, nil)

	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}

	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
	}, &reply); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestThresholdsV1ResetThresholdStoreErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(dbCM, cfg, cM)
	cacheS := engine.NewCacheS(cfg, dm, cM, nil)
	dm.SetCache(cacheS)
	cM.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, cM)
	th := &utils.Threshold{
		Hits:   2,
		Tenant: "cgrates.org",
		ID:     "TH1",
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	cacheS.SetWithoutReplicate(utils.CacheThresholds, th.TenantID(), th, nil, true, utils.NonTransactional)

	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
	}, &reply); err == nil || err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("Expected error <%+v>, Received error <%+v>", utils.ErrDisconnected, err)
	}
}
