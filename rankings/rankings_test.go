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

package rankings

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTenantID(t *testing.T) {
	rp := &utils.RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "01",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{"metric1"},
		Sorting:           "asc",
		SortingParameters: []string{"param1"},
		ThresholdIDs:      []string{"threshold1"},
	}

	tenantID := rp.TenantID()

	expectedTenantID := "cgrates.org:01"

	if tenantID != expectedTenantID {
		t.Errorf("TenantID() = %v; want %v", tenantID, expectedTenantID)
	}
}

func TestRankingProfileWithAPIOpts(t *testing.T) {
	rp := &utils.RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{"metric1"},
		Sorting:           "asc",
		SortingParameters: []string{"param1"},
		ThresholdIDs:      []string{"threshold1"},
	}

	rpo := utils.RankingProfileWithAPIOpts{
		RankingProfile: rp,
		APIOpts:        map[string]any{"option1": "value1"},
	}

	if rpo.APIOpts["option1"] != "value1" {
		t.Errorf("APIOpts[option1] = %v; want %v", rpo.APIOpts["option1"], "value1")
	}

	if rpo.Tenant != rp.Tenant {
		t.Errorf("RankingProfile Tenant = %v; want %v", rpo.Tenant, rp.Tenant)
	}

	if rpo.ID != rp.ID {
		t.Errorf("RankingProfile ID = %v; want %v", rpo.ID, rp.ID)
	}
}

func TestRankingProfileLockKey(t *testing.T) {
	tests := []struct {
		tenant   string
		id       string
		expected string
	}{
		{"cgrates.org", "01", "*ranking_profiles:cgrates.org:01"},
		{"cgrates.org", "02", "*ranking_profiles:cgrates.org:02"},
		{"cgrates.org", "03", "*ranking_profiles:cgrates.org:03"},
	}

	for _, test := range tests {
		result := utils.RankingProfileLockKey(test.tenant, test.id)

		if result != test.expected {
			t.Errorf("rankingProfileLockKey(%q, %q) = %v; want %v", test.tenant, test.id, result, test.expected)
		}
	}
}

func TestNewRankingService(t *testing.T) {
	dm := &engine.DataManager{}
	cgrcfg := &config.CGRConfig{}
	filterS := &engine.FilterS{}
	connMgr := &engine.ConnManager{}

	rankingService := NewRankingS(dm, connMgr, filterS, cgrcfg)

	if rankingService == nil {
		t.Fatal("NewRankingService() returned nil")
	}

	if rankingService.dm != dm {
		t.Errorf("Expected dm to be %v, got %v", dm, rankingService.dm)
	}

	if rankingService.cgrcfg != cgrcfg {
		t.Errorf("Expected cfg to be %v, got %v", cgrcfg, rankingService.cgrcfg)
	}

	if rankingService.filterS != filterS {
		t.Errorf("Expected fltrS to be %v, got %v", filterS, rankingService.filterS)
	}

	if rankingService.connMgr != connMgr {
		t.Errorf("Expected connMgr to be %v, got %v", connMgr, rankingService.connMgr)
	}
}

func TestStoreRanking(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB([]string{}, []string{}, map[string]*config.ItemOpts{})
	dm := engine.NewDataManager(dataDB, cfg, nil)
	rkg := NewRankingS(dm, nil, nil, cfg)
	ranking := &utils.Ranking{}
	ranking.SetConfig(&utils.RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "ID1",
		Schedule:          "@every 1s",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{"metric1", "metric2"},
		Sorting:           "asc",
		SortingParameters: []string{"metric1:true"},
		Stored:            true,
		ThresholdIDs:      []string{"threshold1"},
	})
	ctx := context.Background()
	cfg.RankingSCfg().StoreInterval = 0
	if err := rkg.storeRanking(ctx, ranking); err != nil {
		t.Errorf("Expected no error when StoreInterval is 0, but got: %v", err)
	}
	if len(rkg.storedRankings) != 0 {
		t.Error("Expected storedRankings to be empty when StoreInterval is 0")
	}
	cfg.RankingSCfg().StoreInterval = -1
	if err := rkg.storeRanking(ctx, ranking); err != nil {
		t.Errorf("Expected no error when StoreInterval is -1, but got: %v", err)
	}
	cfg.RankingSCfg().StoreInterval = time.Second
	if err := rkg.storeRanking(ctx, ranking); err != nil {
		t.Errorf("Expected no error when StoreInterval is positive, but got: %v", err)
	}
}
