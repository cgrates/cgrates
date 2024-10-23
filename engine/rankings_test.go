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

	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
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
	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
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

	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
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
