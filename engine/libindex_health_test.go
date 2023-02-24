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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/ericlagergren/decimal"
)

func TestGetFltrIdxHealthForRateRates(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	rt := &utils.RateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "TEST_RATE_TEST",
		FilterIDs:       []string{"*string:~*req.Account:dan"},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rt, false, true); err != nil {
		t.Error(err)
	}
	rply, err := GetFltrIdxHealthForRateRates(context.Background(), dm, ltcache.NewCache(50, 60*time.Second, true, nil),
		ltcache.NewCache(40, 30*time.Second, false, nil),
		ltcache.NewCache(20, 20*time.Second, true, nil))
	if err != nil {
		t.Error(err)
	}
	exp := &FilterIHReply{
		MissingObjects: nil,
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  make(map[string][]string),
		MissingFilters: make(map[string][]string),
	}
	if !reflect.DeepEqual(rply, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rply)
	}
}

func TestGetFiltersRateProfilesErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	if _, err := getFilters(context.Background(), dm, utils.CacheRateProfilesFilterIndexes, utils.CGRateSorg, "fltrID"); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetFiltersActionProfilesOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	ap := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "fltrID",
		FilterIDs: []string{"fltr_test"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*APAction{{}},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Usage",
				Values:  []string{"10s"},
			},
		},
	}
	if err := newFlt.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(context.Background(), newFlt, false); err != nil {
		t.Error(err)
	}

	if err := dm.SetActionProfile(context.Background(), ap, true); err != nil {
		t.Error(err)
	}

	if rcv, err := getFilters(context.Background(), dm, utils.CacheActionProfilesFilterIndexes, utils.CGRateSorg, "fltrID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, ap.FilterIDs) {
		t.Errorf("Expected %v\n but received %v", ap.FilterIDs, rcv)
	}
}

func TestGetFiltersActionProfilesErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	if _, err := getFilters(context.Background(), dm, utils.CacheActionProfilesFilterIndexes, utils.CGRateSorg, "fltrID"); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetFiltersAccountsOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	ap := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "fltrID",
		FilterIDs: []string{"fltr_test"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": {
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{Big: decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{Big: decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
					},
				},
			},
		},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Usage",
				Values:  []string{"10s"},
			},
		},
	}
	if err := newFlt.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(context.Background(), newFlt, false); err != nil {
		t.Error(err)
	}

	if err := dm.SetAccount(context.Background(), ap, true); err != nil {
		t.Error(err)
	}

	if rcv, err := getFilters(context.Background(), dm, utils.CacheAccountsFilterIndexes, utils.CGRateSorg, "fltrID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, ap.FilterIDs) {
		t.Errorf("Expected %v\n but received %v", ap.FilterIDs, rcv)
	}
}

func TestGetFiltersAccountsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	if _, err := getFilters(context.Background(), dm, utils.CacheAccountsFilterIndexes, utils.CGRateSorg, "fltrID"); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetFiltersDefault(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	expErr := `unsupported index type:<"inexistent">`
	if _, err := getFilters(context.Background(), dm, "inexistent", utils.CGRateSorg, "fltrID"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}
