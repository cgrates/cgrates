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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)
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
	rply, err := GetFltrIdxHealthForRateRates(context.Background(), dm, ltcache.NewCache(50, 60*time.Second, true, false, nil),
		ltcache.NewCache(40, 30*time.Second, false, false, nil),
		ltcache.NewCache(20, 20*time.Second, true, false, nil))
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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if _, err := getFilters(context.Background(), dm, utils.CacheRateProfilesFilterIndexes, utils.CGRateSorg, "fltrID"); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetFiltersActionProfilesOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{
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
		Actions:  []*utils.APAction{{}},
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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if _, err := getFilters(context.Background(), dm, utils.CacheActionProfilesFilterIndexes, utils.CGRateSorg, "fltrID"); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetFiltersAccountsOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if _, err := getFilters(context.Background(), dm, utils.CacheAccountsFilterIndexes, utils.CGRateSorg, "fltrID"); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetFiltersDefault(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	expErr := `unsupported index type:<"inexistent">`
	if _, err := getFilters(context.Background(), dm, "inexistent", utils.CGRateSorg, "fltrID"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestGetFilterAsIndexSetDynamicVal(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FL1",
		Rules: []*FilterRule{
			{
				Type:    "*string",
				Element: "*req.Account",
				Values:  []string{"~*accounts"},
			},
		},
	}

	exp := map[string]utils.StringSet{}
	if rcv, err := getFilterAsIndexSet(context.Background(), dm, ltcache.NewCache(50, 60*time.Second, true, false, nil), utils.CacheRateFilterIndexes, "cgrates.org:RPID", fltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

}

func TestGetFilterAsIndexSetElementNotDynamic(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FL1",
		Rules: []*FilterRule{
			{
				Type:    "*string",
				Element: "*req.Account",
				Values:  []string{"~*req.Account", "1001"},
			},
		},
	}

	exp := map[string]utils.StringSet{
		"*string:*req.Account:*req.Account": {},
	}
	if rcv, err := getFilterAsIndexSet(context.Background(), dm, ltcache.NewCache(50, 60*time.Second, true, false, nil), utils.CacheRateFilterIndexes, "cgrates.org:RPID", fltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v\n but received %+v", exp, rcv)
	}

}

func TestGetFilterAsIndexSetGetIHFltrIdxFromCacheErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := NewDataManager(data, cfg, nil)

	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FL1",
		Rules: []*FilterRule{
			{
				Type:    "*string",
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}

	if _, err := getFilterAsIndexSet(context.Background(), dm, ltcache.NewCache(50, 60*time.Second, true, false, nil), utils.CacheRateFilterIndexes, "cgrates.org:RPID", fltr); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestUpdateFilterIHMisingIndxGetIHFltrIdxFromCache1Err(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := NewDataManager(data, cfg, nil)

	rply := &FilterIHReply{
		MissingObjects: nil,
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  make(map[string][]string),
		MissingFilters: make(map[string][]string),
	}

	if _, err := updateFilterIHMisingIndx(context.Background(), dm, ltcache.NewCache(0, 0, false, false, nil), ltcache.NewCache(0, 0, false, false, nil), []string{}, utils.CacheRateFilterIndexes, "cgrates.org", "cgrates.org:RP", "RP", rply); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}
func TestUpdateFilterIHMisingIndxGetIHFltrIdxFromCache2Err(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := NewDataManager(data, cfg, nil)

	rply := &FilterIHReply{
		MissingObjects: nil,
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  make(map[string][]string),
		MissingFilters: make(map[string][]string),
	}

	if _, err := updateFilterIHMisingIndx(context.Background(), dm, ltcache.NewCache(0, 0, false, false, nil), ltcache.NewCache(0, 0, false, false, nil), []string{"fltr"}, utils.CacheRateFilterIndexes, "cgrates.org", "cgrates.org:RP", "RP", rply); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestUpdateFilterIHMisingIndxReplyIndexes(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	rply := &FilterIHReply{
		MissingObjects: []string{"cgrates.org:ATTR2"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*req.Account:1001": {"ATTR1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*string:*req.Account:1002": {"ATTR1"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:Fltr1": {"ATTR1"},
		},
	}

	if rcv, err := updateFilterIHMisingIndx(context.Background(), dm, ltcache.NewCache(0, 0, false, false, nil), ltcache.NewCache(0, 0, false, false, nil), []string{}, utils.CacheRateFilterIndexes, "cgrates.org", "cgrates.org:RP", "RP", rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, rply) {
		t.Errorf("Expected %+v\n but received %+v", rply, rcv)
	}
}

func TestUpdateFilterIHMisingIndxHasNotReplyIndexes(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	rply := &FilterIHReply{
		MissingObjects: []string{"cgrates.org:ATTR2"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*req.Account:1001": {"ATTR1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*string:*req.Account:1002": {"ATTR1"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:Fltr1": {"ATTR1"},
		},
	}

	ssRt := utils.StringSet{
		utils.CGRateSorg: {},
	}

	if err := Cache.Set(context.Background(), utils.CacheRateFilterIndexes, "cgrates.org:RP:*none:*any:*any", ssRt, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if rcv, err := updateFilterIHMisingIndx(context.Background(), dm, ltcache.NewCache(0, 0, false, false, nil), ltcache.NewCache(0, 0, false, false, nil), []string{}, utils.CacheRateFilterIndexes, "cgrates.org", "cgrates.org:RP", "NewRP", rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, rply) {
		t.Errorf("Expected %+v\n but received %+v", rply, rcv)
	}
}

func TestUpdateFilterIHMisingIndxGetFilterAsIndexSetErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := NewDataManager(data, cfg, nil)

	rply := &FilterIHReply{
		MissingObjects: nil,
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  make(map[string][]string),
		MissingFilters: make(map[string][]string),
	}

	flt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FL1",
		Rules: []*FilterRule{
			{
				Type:    "*string",
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}

	if err := Cache.Set(context.Background(), utils.CacheFilters, "cgrates.org:fltr", flt, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, err := updateFilterIHMisingIndx(context.Background(), dm, ltcache.NewCache(0, 0, false, false, nil), ltcache.NewCache(0, 0, false, false, nil), []string{"fltr"}, utils.CacheRateFilterIndexes, "cgrates.org", "cgrates.org:RP", "RP", rply); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestGetFltrIdxHealthGetKeysForPrefixErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(0, 0, false, false, nil)

	expErr := "unsupported prefix in GetKeysForPrefix: "
	if _, err := GetFltrIdxHealth(context.Background(), dm, useLtcache, useLtcache, useLtcache, ""); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestGetFltrIdxHealthgetIHObjFromCacheErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(0, 0, false, false, nil)
	if err := dm.SetAttributeProfile(context.Background(), &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := Cache.Set(context.Background(), utils.CacheAttributeProfiles, "cgrates.org:ATTR1", nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, err := GetFltrIdxHealth(context.Background(), dm, useLtcache, useLtcache, useLtcache, utils.CacheAttributeFilterIndexes); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetFltrIdxHealthIdxKeyFormatErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(0, 0, false, false, nil)
	if err := dm.SetAttributeProfile(context.Background(), &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}, false); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org",
		map[string]utils.StringSet{"*string:*req.Account": {"ATTR1": {}, "ATTR2": {}}},
		true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	expErr := "WRONG_IDX_KEY_FORMAT<cgrates.org:*string:*req.Account>"
	if _, err := GetFltrIdxHealth(context.Background(), dm, useLtcache, useLtcache, useLtcache, utils.CacheAttributeFilterIndexes); err == nil || expErr != err.Error() {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestGetRevFltrIdxHealthFromObjGetKeysForPrefixErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(0, 0, false, false, nil)

	expErr := "unsupported prefix in GetKeysForPrefix: "
	if _, err := getRevFltrIdxHealthFromObj(context.Background(), dm, useLtcache, useLtcache, useLtcache, ""); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestGetRevFltrIdxHealthFromObjIHObjFromCacheErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(0, 0, false, false, nil)
	if err := dm.SetAttributeProfile(context.Background(), &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := Cache.Set(context.Background(), utils.CacheAttributeProfiles, "cgrates.org:ATTR1", nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, err := getRevFltrIdxHealthFromObj(context.Background(), dm, useLtcache, useLtcache, useLtcache, utils.CacheAttributeFilterIndexes); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetRevFltrIdxHealthFromReverseGetKeysForPrefixErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(0, 0, false, false, nil)
	objCaches := make(map[string]*ltcache.Cache)

	rply := make(map[string]*ReverseFilterIHReply)

	expErr := "unsupported prefix in GetKeysForPrefix: "
	if _, err := getRevFltrIdxHealthFromReverse(context.Background(), dm, useLtcache, useLtcache, objCaches, rply); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestGetRatesFromCacheGetRateProfileErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(0, 0, false, false, nil)

	if _, err := getRatesFromCache(context.Background(), dm, useLtcache, "", ""); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetRatesFromCacheObjValNil(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	useLtcache := ltcache.NewCache(20, 20*time.Second, true, false, nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	useLtcache.Set("cgrates.org:ATTR1", nil, []string{})

	if _, err := getRatesFromCache(context.Background(), dm, useLtcache, "cgrates.org", "ATTR1"); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetRevFltrIdxHealthFromRateRatesGetKeysForPrefixErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(0, 0, false, false, nil)

	expErr := "unsupported prefix in GetKeysForPrefix: "
	if _, err := getRevFltrIdxHealthFromRateRates(context.Background(), dm, useLtcache, useLtcache, useLtcache); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestGetRevFltrIdxHealthFromRateRatesGetRatesFromCacheErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	db, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg, nil)

	if err := dm.SetAttributeProfile(context.Background(), &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*string:~*req.Account:1001", "Fltr1", "Fltr3"},
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: "cgrates.org",
		ID:     "Fltr3",
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetIndexes(context.Background(), utils.CacheReverseFilterIndexes, "cgrates.org:Fltr2",
		map[string]utils.StringSet{utils.CacheAttributeFilterIndexes: {"ATTR1": {}, "ATTR2": {}}},
		true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetRateProfile(context.Background(), &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID:        "RT1",
				FilterIDs: []string{"Fltr3"},
			},
		},
	}, false, false); err != nil {
		t.Fatal(err)
	}

	useLtcache := ltcache.NewCache(20, 20*time.Second, true, false, nil)
	useLtcache.Set("cgrates.org:RP1", nil, []string{})

	if _, err := getRevFltrIdxHealthFromRateRates(context.Background(), dm, useLtcache, useLtcache, useLtcache); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestGetFltrIdxHealthForRateRatesGetKeysForPrefixErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := NewDataManager(data, cfg, nil)

	useLtcache := ltcache.NewCache(-1, 0, false, false, nil)

	expErr := "unsupported prefix in GetKeysForPrefix: "
	if _, err := GetFltrIdxHealthForRateRates(context.Background(), dm, useLtcache, useLtcache, useLtcache); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}
