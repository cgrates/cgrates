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

package engine

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

func TestStartTrendS(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.TrendSCfg().Enabled = true
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	tS := NewTrendS(dm, nil, nil, cfg)
	dm.SetTrendProfile(&TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "ID1",
		Schedule:        "@every 1s",
		StatID:          "stat1",
		Metrics:         []string{"metric1", "metric2"},
		TTL:             time.Minute,
		QueueLength:     10,
		MinItems:        5,
		CorrelationType: "*last",
		Tolerance:       0.1,
		Stored:          true,
		ThresholdIDs:    []string{"threshold1"},
	})
	if err := tS.StartTrendS(); err != nil {
		t.Error(err)
	}
}

func TestStoreTrend(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	tS := &TrendS{
		cgrcfg:       cfg,
		dm:           dm,
		storedTrends: make(utils.StringSet),
	}

	trendProfile := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "trendID1",
		Schedule:        "@every 1s",
		StatID:          "stat1",
		Metrics:         []string{"metric1", "metric2"},
		TTL:             time.Minute,
		QueueLength:     100,
		MinItems:        10,
		CorrelationType: "average",
		Tolerance:       0.05,
		Stored:          true,
		ThresholdIDs:    []string{"threshold1"},
	}

	trend := &Trend{
		Tenant:   "cgrates.org",
		ID:       "trendID1",
		tPrfl:    trendProfile,
		Metrics:  make(map[time.Time]map[string]*MetricWithTrend),
		RunTimes: []time.Time{time.Now().Add(-time.Second)},
		mLast:    map[string]time.Time{"metric1": time.Now().Add(-time.Minute), "metric2": time.Now().Add(-2 * time.Minute)},
		mCounts:  map[string]int{"metric1": 5, "metric2": 3},
		mTotals:  map[string]float64{"metric1": 100.5, "metric2": 60.3},
	}

	cfg.TrendSCfg().StoreInterval = 0
	if err := tS.storeTrend(trend); err != nil {
		t.Errorf("Expected no error when StoreInterval is 0, but got: %v", err)
	}
	if len(tS.storedTrends) != 0 {
		t.Error("Expected storedTrends to be empty when StoreInterval is 0")
	}

	cfg.TrendSCfg().StoreInterval = -1
	if err := tS.storeTrend(trend); err != nil {
		t.Errorf("Expected no error when StoreInterval is -1, but got: %v", err)
	}

	cfg.TrendSCfg().StoreInterval = time.Second
	if err := tS.storeTrend(trend); err != nil {
		t.Errorf("Expected no error when StoreInterval is positive, but got: %v", err)
	}
	if _, exists := tS.storedTrends[trend.TenantID()]; !exists {
		t.Errorf("Expected trendID %v to be stored in storedTrends for positive StoreInterval", trend.ID)
	}

	retrievedTrend, err := dm.GetTrend(trend.Tenant, trend.ID, true, true, "")
	if err != nil {
		t.Errorf("Error retrieving trend from data manager: %v", err)
	}
	if retrievedTrend == nil {
		t.Error("Expected a stored trend to be found in data manager, but got nil")
	} else {
		if retrievedTrend.Tenant != trend.Tenant {
			t.Errorf("Expected retrieved trend Tenant to be %s, but got %s", trend.Tenant, retrievedTrend.Tenant)
		}
		if retrievedTrend.ID != trend.ID {
			t.Errorf("Expected retrieved trend ID to be %s, but got %s", trend.ID, retrievedTrend.ID)
		}

	}
}
func TestV1GetTrendSummary(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	tS := &TrendS{
		dm: dm,
	}

	trendID := "trendID1"
	tenantID := "cgrates.org"
	trend := &Trend{
		Tenant: tenantID,
		ID:     trendID,
	}
	if err := dm.SetTrend(trend); err != nil {
		t.Fatalf("Failed to set trend in data manager: %v", err)
	}

	arg := utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: tenantID,
			ID:     trendID,
		},
		APIOpts: make(map[string]any),
	}
	var reply TrendSummary

	err := tS.V1GetTrendSummary(nil, arg, &reply)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if reply.Tenant != tenantID || reply.ID != trendID {
		t.Errorf("Expected reply to have Tenant: %s and ID: %s, but got Tenant: %s and ID: %s",
			tenantID, trendID, reply.Tenant, reply.ID)
	}
}

func TestTrendProfileTenantID(t *testing.T) {
	profile := &TrendProfile{
		Tenant: "cgrates.org",
		ID:     "1",
	}
	result := profile.TenantID()
	expected := "cgrates.org:1"
	if result != expected {
		t.Errorf("TenantID() = %v; want %v", result, expected)
	}
}

func TestTrendTenantID(t *testing.T) {
	trend := &Trend{
		Tenant: "cgrates.org",
		ID:     "1",
	}
	result := trend.TenantID()
	expected := "cgrates.org:1"
	if result != expected {
		t.Errorf("TenantID() = %v; want %v", result, expected)
	}
}

func TestNewTrendS(t *testing.T) {
	dm := &DataManager{}
	connMgr := &ConnManager{}
	filterS := &FilterS{}
	cgrcfg := &config.CGRConfig{}

	trendS := NewTrendS(dm, connMgr, filterS, cgrcfg)

	if trendS == nil {
		t.Errorf("Expected NewTrendS to return a non-nil instance")
	}
	if trendS.dm != dm {
		t.Errorf("Expected DataManager to be set correctly, got %v, want %v", trendS.dm, dm)
	}
	if trendS.connMgr != connMgr {
		t.Errorf("Expected ConnManager to be set correctly, got %v, want %v", trendS.connMgr, connMgr)
	}
	if trendS.filterS != filterS {
		t.Errorf("Expected FilterS to be set correctly, got %v, want %v", trendS.filterS, filterS)
	}
	if trendS.cgrcfg != cgrcfg {
		t.Errorf("Expected CGRConfig to be set correctly, got %v, want %v", trendS.cgrcfg, cgrcfg)
	}

	if trendS.trendStop == nil {
		t.Errorf("Expected loopStopped to be initialized, but got nil")
	}
	if trendS.crnTQsMux == nil {
		t.Errorf("Expected crnTQsMux to be initialized, but got nil")
	}
	if trendS.crnTQs == nil {
		t.Errorf("Expected crnTQs to be initialized, but got nil")
	} else if len(trendS.crnTQs) != 0 {
		t.Errorf("Expected crnTQs to be empty, but got length %d", len(trendS.crnTQs))
	}

}

func TestProcessEEsWithError(t *testing.T) {

	trend := &Trend{
		ID:     "ID",
		Tenant: "cgrates.org",
	}

	mockConnMgr := &ConnManager{}
	trendService := &TrendS{
		cgrcfg:  &config.CGRConfig{},
		connMgr: mockConnMgr,
	}

	err := trendService.processEEs(trend)
	if err != nil || errors.Is(err, utils.ErrPartiallyExecuted) {
		t.Errorf("Expected error %v, got %v", utils.ErrPartiallyExecuted, err)
	}

}

func TestV1ScheduleQueriesInvalidTrendID(t *testing.T) {

	ctx := context.Background()

	tS := &TrendS{
		crn:       cron.New(),
		crnTQs:    make(map[string]map[string]cron.EntryID),
		crnTQsMux: &sync.RWMutex{},
	}

	args := &utils.ArgScheduleTrendQueries{
		TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ID",
			},
			APIOpts: make(map[string]any),
		},
		TrendIDs: []string{"invalidID"},
	}

	var scheduled int
	err := tS.V1ScheduleQueries(ctx, args, &scheduled)

	if err == nil {
		t.Errorf("expected an error but got none")
	}

	if scheduled != 0 {
		t.Errorf("expected scheduled to be 0 but got %d", scheduled)
	}
}

func TestProcessThresholds_OptsInitialization(t *testing.T) {
	tS := &TrendS{}

	trnd := &Trend{
		Tenant: "cgrates.org",
		ID:     "ID",
	}

	err := tS.processThresholds(trnd)

	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

}

func TestTrendsStoreTrends(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.TrendSCfg().Enabled = true
	cfg.TrendSCfg().StoreInterval = time.Millisecond * 1500
	cfg.TrendSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	conn := make(chan context.ClientConnector, 1)
	conn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueFloatMetrics: func(ctx *context.Context, args, reply any) error {
				if args.(*utils.TenantIDWithAPIOpts).ID == "stat1" {
					*reply.(*map[string]float64) = map[string]float64{
						utils.MetaTCD: float64(20 * time.Second),
						utils.MetaACC: 22.2,
					}
				} else if args.(*utils.TenantIDWithAPIOpts).ID == "stat2" {
					*reply.(*map[string]float64) = map[string]float64{
						utils.MetaTCD: float64(23 * time.Second),
						utils.MetaACC: 22.2,
					}
				}
				return nil
			},
		},
	}
	connMgr = NewConnManager(config.NewDefaultCGRConfig(), map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats): conn,
	})
	trnd := NewTrendS(dm, connMgr, nil, cfg)
	trendProfile := &TrendProfile{
		Tenant:   "cgrates.org",
		StatID:   "stat1",
		ID:       "ID1",
		Schedule: "@every 1s",
		Stored:   true,
	}

	dm.SetTrendProfile(trendProfile)
	if err := trnd.StartTrendS(); err != nil {
		t.Fatalf("Unexpected error when starting trends: %v", err)
	}
	profile, err := dm.GetTrend("cgrates.org", "ID1", false, false, "")
	if err != nil {
		t.Errorf("Error retrieving trend profile: %v", err)
	}
	if profile == nil {
		t.Fatal("Expected trend profile to be present, but it was not found")
	}
	if profile.ID != "ID1" {
		t.Errorf("Expected profile ID to be 'ID1', but got %v", profile.ID)
	}
	if profile.Tenant != "cgrates.org" {
		t.Errorf("Expected tenant to be 'cgrates.org', but got %v", profile.Tenant)
	}

}

func TestTrendReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.TrendSCfg().Enabled = true
	cfg.TrendSCfg().StoreInterval = 0
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	conn := make(chan context.ClientConnector, 1)
	conn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueFloatMetrics: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]float64) = map[string]float64{
					utils.MetaTCD: float64(20 * time.Second),
					utils.MetaACC: 22.2,
				}
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats): conn,
	})
	trnd := NewTrendS(dm, connMgr, nil, cfg)
	trendProfile := &TrendProfile{
		Tenant:   "cgrates.org",
		StatID:   "stat1",
		ID:       "ID1",
		Schedule: "@every 1s",
		Stored:   true,
	}
	dm.SetTrendProfile(trendProfile)
	go trnd.asyncStoreTrends()
	trnd.Reload()
	select {
	case <-trnd.trendStop:
		t.Fatal("Expected trendStop channel to be closed after Reload")
	default:
	}
	if trnd.trendStop == nil {
		t.Fatal("Expected trendStop channel to be re-initialized after Reload")
	}
	if trnd.storingStopped == nil {
		t.Fatal("Expected storingStopped channel to be re-initialized after Reload")
	}
}

func TestV1GetTrendStoreIntervalZero(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.TrendSCfg().Enabled = true
	cfg.TrendSCfg().StoreInterval = 0
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	conn := make(chan context.ClientConnector, 1)
	conn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueFloatMetrics: func(ctx *context.Context, args, reply any) error {
				if args.(*utils.TenantIDWithAPIOpts).ID == "stat1" {
					*reply.(*map[string]float64) = map[string]float64{
						utils.MetaTCD: float64(20 * time.Second),
						utils.MetaACC: 22.2,
					}
				}
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats): conn,
	})
	trnd := NewTrendS(dm, connMgr, nil, cfg)
	trendProfile := &TrendProfile{
		Tenant:   "cgrates.org",
		StatID:   "stat1",
		ID:       "ID1",
		Schedule: "@every 1s",
		Stored:   true,
	}
	dm.SetTrendProfile(trendProfile)
	if err := trnd.StartTrendS(); err != nil {
		t.Fatalf("Unexpected error when starting trends: %v", err)
	}
	time.Sleep(1 * time.Second)
	ctx := context.Background()
	arg := &utils.ArgGetTrend{ID: "nonexistent"}
	var retTrend Trend

	err := trnd.V1GetTrend(ctx, arg, &retTrend)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.Is(err, utils.ErrNotFound) {
		t.Fatalf("Expected error type 'ErrNotFound', got: %v", err)
	}
}
