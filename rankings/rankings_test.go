// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package rankings

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
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
		{"cgrates.org", "01", "*rankingProfiles:cgrates.org:01"},
		{"cgrates.org", "02", "*rankingProfiles:cgrates.org:02"},
		{"cgrates.org", "03", "*rankingProfiles:cgrates.org:03"},
	}

	for _, test := range tests {
		result := utils.RankingProfileLockKey(test.tenant, test.id)

		if result != test.expected {
			t.Errorf("rankingProfileLockKey(%q, %q) = %v; want %v", test.tenant, test.id, result, test.expected)
		}
	}
}

func TestNewRankingService(t *testing.T) {
	cgrcfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cgrcfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cgrcfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cgrcfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cgrcfg, nil, locker)
	cache := engine.NewCacheS(cgrcfg, nil, nil, nil, locker)
	filterS := &engine.FilterS{}
	connMgr := &engine.ConnManager{}

	rankingService := NewRankingS(dm, cache, connMgr, filterS, cgrcfg)

	if rankingService == nil {
		t.Fatal("NewRankingService() returned nil")
	}

	if rankingService.dm != dm {
		t.Errorf("Expected dm to be %v, got %v", dm, rankingService.dm)
	}

	if rankingService.cache != cache {
		t.Errorf("Expected cache to be %v, got %v", cache, rankingService.cache)
	}

	if rankingService.cgrcfg != cgrcfg {
		t.Errorf("Expected cfg to be %v, got %v", cgrcfg, rankingService.cgrcfg)
	}

	if rankingService.fltrS != filterS {
		t.Errorf("Expected fltrS to be %v, got %v", filterS, rankingService.fltrS)
	}

	if rankingService.connMgr != connMgr {
		t.Errorf("Expected connMgr to be %v, got %v", connMgr, rankingService.connMgr)
	}
}

func TestStoreRanking(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	dataDB, _ := engine.NewInternalDB([]string{}, []string{}, &ltcache.TransCacheOpts{}, map[string]*config.ItemOpts{})
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil, locker)
	dm.SetCache(cacheS)
	rkg := NewRankingS(dm, cacheS, nil, nil, cfg)
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
