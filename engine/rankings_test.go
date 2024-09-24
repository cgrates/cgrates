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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
)

func TestTenantID(t *testing.T) {
	rp := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "01",
		QueryInterval:     5 * time.Minute,
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
	rp := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		QueryInterval:     5 * time.Minute,
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{"metric1"},
		Sorting:           "asc",
		SortingParameters: []string{"param1"},
		ThresholdIDs:      []string{"threshold1"},
	}

	rpo := RankingProfileWithAPIOpts{
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
		result := rankingProfileLockKey(test.tenant, test.id)

		if result != test.expected {
			t.Errorf("rankingProfileLockKey(%q, %q) = %v; want %v", test.tenant, test.id, result, test.expected)
		}
	}
}

func TestNewRankingService(t *testing.T) {
	dm := &DataManager{}
	cgrcfg := &config.CGRConfig{}
	filterS := &FilterS{}
	connMgr := &ConnManager{}

	rankingService := NewRankingService(dm, cgrcfg, filterS, connMgr)

	if rankingService == nil {
		t.Fatal("NewRankingService() returned nil")
	}

	if rankingService.dm != dm {
		t.Errorf("Expected dm to be %v, got %v", dm, rankingService.dm)
	}

	if rankingService.cfg != cgrcfg {
		t.Errorf("Expected cfg to be %v, got %v", cgrcfg, rankingService.cfg)
	}

	if rankingService.fltrS != filterS {
		t.Errorf("Expected fltrS to be %v, got %v", filterS, rankingService.fltrS)
	}

	if rankingService.connMgr != connMgr {
		t.Errorf("Expected connMgr to be %v, got %v", connMgr, rankingService.connMgr)
	}
}
