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
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

func TestNewRankingS(t *testing.T) {
	dm := &DataManager{}
	connMgr := &ConnManager{}
	filterS := &FilterS{}
	cgrcfg := &config.CGRConfig{}

	rankingS := NewRankingS(dm, connMgr, filterS, cgrcfg)

	if rankingS == nil {
		t.Fatal("Expected RankingS to be non-nil, got nil")
	}

	if rankingS.dm != dm {
		t.Errorf("Expected dm to be '%v', got '%v'", dm, rankingS.dm)
	}
	if rankingS.connMgr != connMgr {
		t.Errorf("Expected connMgr to be '%v', got '%v'", connMgr, rankingS.connMgr)
	}
	if rankingS.filterS != filterS {
		t.Errorf("Expected filterS to be '%v', got '%v'", filterS, rankingS.filterS)
	}
	if rankingS.cgrcfg != cgrcfg {
		t.Errorf("Expected cgrcfg to be '%v', got '%v'", cgrcfg, rankingS.cgrcfg)
	}
	if rankingS.crn == nil {
		t.Error("Expected crn to be initialized, got nil")
	}
	if rankingS.crnTQsMux == nil {
		t.Error("Expected crnTQsMux to be initialized, got nil")
	}
	if len(rankingS.crnTQs) != 0 {
		t.Errorf("Expected crnTQs to be empty, got length %d", len(rankingS.crnTQs))
	}
	if rankingS.storedRankings == nil {
		t.Error("Expected storedRankings to be initialized, got nil")
	}
	if rankingS.storingStopped == nil {
		t.Error("Expected storingStopped channel to be initialized, got nil")
	}
	if rankingS.rankingStop == nil {
		t.Error("Expected rankingStop channel to be initialized, got nil")
	}
}
func TestAsRankingSummary(t *testing.T) {
	lastUpdateTime, err := time.Parse(time.RFC3339, "2024-10-21T00:00:00Z")
	if err != nil {
		t.Fatalf("Failed to parse time: %v", err)
	}

	ranking := &Ranking{
		Tenant:            "tenant1",
		ID:                "ranking1",
		LastUpdate:        lastUpdateTime,
		SortedStatIDs:     []string{"stat1", "stat2", "stat3"},
		Metrics:           make(map[string]map[string]float64),
		Sorting:           "asc",
		SortingParameters: []string{"param1", "param2"},
	}

	rkSm := ranking.asRankingSummary()

	if rkSm.Tenant != ranking.Tenant {
		t.Errorf("Expected Tenant to be '%s', got '%s'", ranking.Tenant, rkSm.Tenant)
	}
	if rkSm.ID != ranking.ID {
		t.Errorf("Expected ID to be '%s', got '%s'", ranking.ID, rkSm.ID)
	}
	if rkSm.LastUpdate != ranking.LastUpdate {
		t.Errorf("Expected LastUpdate to be '%v', got '%v'", ranking.LastUpdate, rkSm.LastUpdate)
	}

}

func TestRankingProcessEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.RankingSCfg().StoreInterval = 1

	data, dErr := NewInternalDB(nil, nil, true, false, config.CgrConfig().DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)

	rankingProfile := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "RankingProfile1",
		Schedule:          "0 0 * * *",
		StatIDs:           []string{"Stat1", "Stat2"},
		MetricIDs:         []string{"MetricA", "MetricB"},
		Sorting:           "*asc",
		SortingParameters: []string{"MetricA", "MetricB"},
		Stored:            true,
		ThresholdIDs:      []string{"*none"},
	}

	if err := dm.SetRankingProfile(rankingProfile); err != nil {
		t.Error(err)
	}

	if retrievedProfile, err := dm.GetRankingProfile(rankingProfile.Tenant, rankingProfile.ID, true, false, ""); err != nil {
		t.Errorf("Error retrieving ranking profile: %+v", err)
	} else if !reflect.DeepEqual(rankingProfile, retrievedProfile) {
		t.Errorf("Expecting: %+v, received: %+v", rankingProfile, retrievedProfile)
	}

}

func TestProcessThresholdsEmptySortedStatIDs(t *testing.T) {
	rankingService := &RankingS{
		connMgr: &ConnManager{},
		cgrcfg:  &config.CGRConfig{},
	}

	ranking := &Ranking{
		Tenant: "cgrates.org",
		ID:     "ID",
		rkPrfl: &RankingProfile{
			ThresholdIDs: []string{"threshold1"},
		},
		SortedStatIDs: []string{},
	}

	err := rankingService.processThresholds(ranking)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestProcessEEsHandlesEmptySortedStatIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RankingSCfg().StoreInterval = 1
	data, dErr := NewInternalDB(nil, nil, true, false, config.CgrConfig().DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)

	rankingService := &RankingS{
		cgrcfg:  cfg,
		connMgr: &ConnManager{},
		dm:      dm,
	}

	rk := &Ranking{
		Tenant:            "cgrates.org",
		ID:                "ID",
		LastUpdate:        time.Now(),
		SortedStatIDs:     []string{},
		Metrics:           make(map[string]map[string]float64),
		Sorting:           "",
		SortingParameters: []string{},
		rkPrfl:            nil,
		metricIDs:         utils.StringSet{},
	}

	err := rankingService.processEEs(rk)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestProcessEEsHandlesEmptyEEsConns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RankingSCfg().StoreInterval = 1

	data, dErr := NewInternalDB(nil, nil, true, false, config.CgrConfig().DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)

	rankingService := &RankingS{
		cgrcfg:  cfg,
		connMgr: &ConnManager{},
		dm:      dm,
	}

	rk := &Ranking{
		Tenant:            "cgrates.org",
		ID:                "ID",
		LastUpdate:        time.Now(),
		SortedStatIDs:     []string{"stat_id_1", "stat_id_2"},
		Metrics:           make(map[string]map[string]float64),
		Sorting:           "",
		SortingParameters: []string{},
		rkPrfl:            nil,
		metricIDs:         utils.StringSet{},
	}

	cfg.RankingSCfg().EEsConns = []string{}

	err := rankingService.processEEs(rk)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestV1ScheduleQueriesInvalidRankingID(t *testing.T) {

	ctx := context.Background()

	tS := &RankingS{
		crn:       cron.New(),
		crnTQs:    make(map[string]map[string]cron.EntryID),
		crnTQsMux: &sync.RWMutex{},
	}

	args := &utils.ArgScheduleRankingQueries{
		TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ID",
			},
			APIOpts: make(map[string]any),
		},
		RankingIDs: []string{"invalidID"},
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

func TestStoreRanking(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, err := NewInternalDB(nil, nil, true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	rkg := NewRankingS(dm, nil, nil, cfg)

	ranking := &Ranking{
		rkPrfl: &RankingProfile{
			Tenant:            "cgrates.org",
			ID:                "ID1",
			Schedule:          "@every 1s",
			StatIDs:           []string{"stat1", "stat2"},
			MetricIDs:         []string{"metric1", "metric2"},
			Sorting:           "asc",
			SortingParameters: []string{"metric1:true"},
			Stored:            true,
			ThresholdIDs:      []string{"threshold1"},
		},
	}

	cfg.RankingSCfg().StoreInterval = 0
	if err := rkg.storeRanking(ranking); err != nil {
		t.Errorf("Expected no error when StoreInterval is 0, but got: %v", err)
	}
	if len(rkg.storedRankings) != 0 {
		t.Error("Expected storedRankings to be empty when StoreInterval is 0")
	}

	cfg.RankingSCfg().StoreInterval = -1
	if err := rkg.storeRanking(ranking); err != nil {
		t.Errorf("Expected no error when StoreInterval is -1, but got: %v", err)
	}

	cfg.RankingSCfg().StoreInterval = time.Second
	if err := rkg.storeRanking(ranking); err != nil {
		t.Errorf("Expected no error when StoreInterval is positive, but got: %v", err)
	}

}

func TestRankingsStoreRankings(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RankingSCfg().Enabled = true
	cfg.RankingSCfg().StoreInterval = time.Millisecond * 1300
	cfg.RankingSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	dataDB, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
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
	rkg := NewRankingS(dm, connMgr, nil, cfg)

	rankingProfile := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "ID1",
		Schedule:          "@every 1s",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{},
		Sorting:           "*desc",
		SortingParameters: []string{utils.MetaTCD, utils.MetaACC},
		Stored:            true,
	}
	dm.SetRankingProfile(rankingProfile)

	if err := rkg.StartRankingS(); err != nil {
		t.Fatalf("Unexpected error when starting rankings: %v", err)
	}
	t.Cleanup(func() { rkg.StopRankingS() })

	time.Sleep(1200 * time.Millisecond)

	profile, err := dm.GetRanking("cgrates.org", "ID1", false, false, "")
	if err != nil {
		t.Errorf("Error retrieving ranking profile: %v", err)
	}
	if profile == nil {
		t.Fatal("Expected ranking profile to be present, but it was not found")
	}

	if profile.ID != "ID1" {
		t.Errorf("Expected profile ID to be 'ID1', but got %v", profile.ID)
	}
	if profile.Tenant != "cgrates.org" {
		t.Errorf("Expected tenant to be 'cgrates.org', but got %v", profile.Tenant)
	}

	if profile.Sorting != "*desc" {
		t.Errorf("Expected sorting to be 'desc', but got %v", profile.Sorting)
	}
	if !slices.Equal(profile.SortingParameters, profile.SortingParameters) {
		t.Errorf("Expected SortingParameters to be ['metric1:true'], but got %v", profile.SortingParameters)
	}
	if !reflect.DeepEqual(profile.Metrics, map[string]map[string]float64{"stat1": {"*acc": 22.2, "*tcd": 20000000000}, "stat2": {"*acc": 22.2, "*tcd": 23000000000}}) {
		t.Errorf("Expected sorting to be 'asc', but got %v", profile.Metrics)
	}

}

func TestV1GetRankingSummary(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RankingSCfg().Enabled = true

	dataDB, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	connMgr := NewConnManager(config.NewDefaultCGRConfig(), nil)
	rkg := NewRankingS(dm, connMgr, nil, cfg)

	rankingProfile := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "TestRanking",
		Schedule:          "@every 1s",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{},
		Sorting:           "*desc",
		SortingParameters: []string{utils.MetaTCD, utils.MetaACC},
		Stored:            true,
	}
	dm.SetRankingProfile(rankingProfile)

	if err := rkg.StartRankingS(); err != nil {
		t.Fatalf("Unexpected error when starting rankings: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "TestRanking",
		},
		APIOpts: make(map[string]any),
	}
	reply := &RankingSummary{}

	err := rkg.V1GetRankingSummary(nil, arg, reply)
	if err != nil {
		t.Fatalf("V1GetRankingSummary failed with error: %v", err)
	}

	expectedTenant := "cgrates.org"
	if reply.Tenant != expectedTenant {
		t.Errorf("Expected Tenant to be %q, but got %q", expectedTenant, reply.Tenant)
	}

	expectedID := "TestRanking"
	if reply.ID != expectedID {
		t.Errorf("Expected ID to be %q, but got %q", expectedID, reply.ID)
	}
}

func TestV1GetSchedule(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	dataDB, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	connMgr := NewConnManager(config.NewDefaultCGRConfig(), nil)
	rkg := NewRankingS(dm, connMgr, nil, cfg)

	rankingProfile := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "TestRanking",
		Schedule:          "@every 1s",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{},
		Sorting:           "*desc",
		SortingParameters: []string{utils.MetaTCD, utils.MetaACC},
		Stored:            true,
	}
	dm.SetRankingProfile(rankingProfile)

	if err := rkg.StartRankingS(); err != nil {
		t.Fatalf("Unexpected error when starting rankings: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	arg := &utils.ArgScheduledRankings{
		TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestRanking",
			},
			APIOpts: make(map[string]any),
		},
		RankingIDPrefixes: []string{"TestRanking"},
	}

	var schedRankings []utils.ScheduledRanking

	err := rkg.V1GetSchedule(nil, arg, &schedRankings)
	if err != nil {
		t.Fatalf("V1GetSchedule failed with error: %v", err)
	}

	if len(schedRankings) == 0 {
		t.Errorf("Expected at least one scheduled ranking, but got none")
	}

	if schedRankings[0].RankingID != "TestRanking" {
		t.Errorf("Expected RankingID to be %q, but got %q", "TestRanking", schedRankings[0].RankingID)
	}

	if schedRankings[0].Next.IsZero() {
		t.Errorf("Expected Next to be non-zero, but got zero time")
	}

	if schedRankings[0].Previous.IsZero() {
		t.Errorf("Expected Previous to be non-zero, but got zero time")
	}
}

func TestV1GetRankingMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	connMgr := NewConnManager(config.NewDefaultCGRConfig(), nil)
	rkg := NewRankingS(dm, connMgr, nil, cfg)

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "",
		},
		APIOpts: make(map[string]any),
	}

	retRanking := &Ranking{}
	err := rkg.V1GetRanking(nil, arg, retRanking)

	if err == nil {
		t.Fatalf("Expected error when ID is missing, but got none")
	}

	if !strings.Contains(err.Error(), "MANDATORY_IE_MISSING: [ID") {
		t.Errorf("Expected error related to missing fields, but got: %v", err)
	}
}

func TestV1GetRankingSortedStatIDs(t *testing.T) {

	rk := &Ranking{
		SortedStatIDs: []string{"stat1", "stat2", "stat3"},
	}
	retRanking := &Ranking{}
	retRanking.SortedStatIDs = make([]string, len(rk.SortedStatIDs))
	copy(retRanking.SortedStatIDs, rk.SortedStatIDs)

	if len(retRanking.SortedStatIDs) != len(rk.SortedStatIDs) {
		t.Errorf("Expected SortedStatIDs length: %d, but got: %d", len(rk.SortedStatIDs), len(retRanking.SortedStatIDs))
	}
	for i, v := range rk.SortedStatIDs {
		if retRanking.SortedStatIDs[i] != v {
			t.Errorf("Expected SortedStatIDs[%d]: %s, but got: %s", i, v, retRanking.SortedStatIDs[i])
		}
	}

	rk.SortedStatIDs = []string{}
	retRanking = &Ranking{}
	retRanking.SortedStatIDs = make([]string, len(rk.SortedStatIDs))
	copy(retRanking.SortedStatIDs, rk.SortedStatIDs)

	if len(retRanking.SortedStatIDs) != 0 {
		t.Errorf("Expected empty SortedStatIDs, but got length: %d", len(retRanking.SortedStatIDs))
	}

	rk.SortedStatIDs = nil
	retRanking = &Ranking{}
	retRanking.SortedStatIDs = make([]string, len(rk.SortedStatIDs))
	copy(retRanking.SortedStatIDs, rk.SortedStatIDs)

	if len(retRanking.SortedStatIDs) != 0 {
		t.Errorf("Expected empty SortedStatIDs for nil input, but got length: %d", len(retRanking.SortedStatIDs))
	}
}

func TestV1ScheduleQueries(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	connMgr := NewConnManager(config.NewDefaultCGRConfig(), nil)
	rkg := NewRankingS(dm, connMgr, nil, cfg)

	rankingProfile := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "TestRanking",
		Schedule:          "@every 1s",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{},
		Sorting:           "*desc",
		SortingParameters: []string{utils.MetaTCD, utils.MetaACC},
		Stored:            true,
	}
	dm.SetRankingProfile(rankingProfile)

	if err := rkg.StartRankingS(); err != nil {
		t.Fatalf("Unexpected error when starting rankings: %v", err)
	}

	arg := &utils.ArgScheduleRankingQueries{
		TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestRanking",
			},
			APIOpts: make(map[string]any),
		},
		RankingIDs: []string{"TestRanking"},
	}

	var scheduled int
	err := rkg.V1ScheduleQueries(context.Background(), arg, &scheduled)

	if err != nil {
		t.Fatalf("V1ScheduleQueries failed with error: %v", err)
	}

	if scheduled == 0 {
		t.Errorf("Expected at least one scheduled ranking query, but got none")
	}

	if len(rkg.crnTQs["cgrates.org"]) == 0 {
		t.Errorf("Expected cron entries for tenant 'cgrates.org', but got none")
	}

	if _, exists := rkg.crnTQs["cgrates.org"]["TestRanking"]; !exists {
		t.Errorf("Expected cron entry for ranking ID 'TestRanking', but it was not found")
	}

}
