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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

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
