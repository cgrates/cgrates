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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestSplitFilterIndexes(t *testing.T) {
	tntGrpIdxKey := "tntCtx:*prefix:~*accounts:1001"
	tntGrp, idxKey, err := splitFilterIndex(tntGrpIdxKey)
	if err != nil {
		t.Error(err)
	}
	expTntGrp := "tntCtx"
	expIdxKey := "*prefix:~*accounts:1001"
	if expTntGrp != tntGrp && expIdxKey != idxKey {
		t.Errorf("Expected %v and %v\n but received %v and %v", expTntGrp, expIdxKey, tntGrp, idxKey)
	}
}

func TestSplitFilterIndexesWrongFormat(t *testing.T) {
	tntGrpIdxKey := "tntCtx:*prefix:~*accounts"
	_, _, err := splitFilterIndex(tntGrpIdxKey)
	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestComputeIndexes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	thd := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	dm.SetThresholdProfile(context.Background(), thd, false)
	transactionID := utils.GenUUID()
	indexes, err := ComputeIndexes(context.Background(), dm, "cgrates.org", utils.EmptyString, utils.CacheThresholdFilterIndexes,
		nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
			th, e := dm.GetThresholdProfile(context.Background(), tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(th.FilterIDs)), nil
		}, nil)
	if err != nil {
		t.Error(err)
	}
	exp := make(utils.StringSet)
	exp.Add(utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1001"))
	if !reflect.DeepEqual(exp, indexes) {
		t.Errorf("Expected %v\n but received %v", exp, indexes)
	}
}

func TestComputeIndexesIDsNotNil(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	thd := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	dm.SetThresholdProfile(context.Background(), thd, false)
	transactionID := utils.GenUUID()
	_, err := ComputeIndexes(context.Background(), dm, "cgrates.org", utils.EmptyString, utils.CacheThresholdFilterIndexes,
		&[]string{utils.CacheThresholdFilterIndexes, utils.CacheAccountsFilterIndexes}, transactionID, func(tnt, id, grp string) (*[]string, error) {
			th, e := dm.GetThresholdProfile(context.Background(), tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(th.FilterIDs)), nil
		}, nil)
	if err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestRemoveIndexFiltersItem(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	thd := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	dm.SetThresholdProfile(context.Background(), thd, false)
	// dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	if err := removeIndexFiltersItem(context.Background(), dm, utils.CacheThresholdFilterIndexes, "cgrates.org", "", []string{"account"}); err != nil {
		t.Error(err)
	}
}

func TestRemoveFilterIndexesForFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	thd := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	dm.SetThresholdProfile(context.Background(), thd, false)
	exp := make(utils.StringSet)
	exp.Add(utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1001"))
	// dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	if err := removeFilterIndexesForFilter(context.Background(), dm, utils.CacheThresholdFilterIndexes, "cgrates.org", []string{""}, exp); err != nil {
		t.Error(err)
	}
}

func TestLibIndexSetUpdateRemAttributeProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	// Set an AttributeProfile without filterIDs
	attrPrf := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*Attribute{
			{
				Type:  utils.MetaConstant,
				Path:  "~*req.Account",
				Value: config.NewRSRParsersMustCompile("1002", cfg.GeneralCfg().RSRSep),
			},
		},
	}
	err := dm.SetAttributeProfile(context.Background(), attrPrf, true)
	if err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringSet{
		"*none:*any:*any": {
			"ATTR_TEST": {},
		},
	}
	rcvIndexes, err := dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, attrPrf.Tenant,
		utils.EmptyString, utils.NonTransactional, false, false)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIndexes, expIndexes) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expIndexes), utils.ToJSON(rcvIndexes))
	}

	// Add a non-indexed filter type
	attrPrf = &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_TEST",
		FilterIDs: []string{"*gt:~*req.Element:10"},
		Attributes: []*Attribute{
			{
				Type:  utils.MetaConstant,
				Path:  "~*req.Account",
				Value: config.NewRSRParsersMustCompile("1002", cfg.GeneralCfg().RSRSep),
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attrPrf, true)
	if err != nil {
		t.Error(err)
	}
	rcvIndexes, err = dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, attrPrf.Tenant,
		utils.EmptyString, utils.NonTransactional, false, false)
	// if err != nil {
	// 	t.Error(err)
	// } else if !reflect.DeepEqual(rcvIndexes, expIndexes) {
	// 	t.Errorf("expected: <%+v>, \nreceived: <%+v>",
	// 		utils.ToJSON(expIndexes), utils.ToJSON(rcvIndexes))
	// }

	// Add an indexed filter type
	attrPrf = &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_TEST",
		FilterIDs: []string{"*gt:~*req.Element:10", "*prefix:~*req.Account:10"},
		Attributes: []*Attribute{
			{
				Type:  utils.MetaConstant,
				Path:  "~*req.Account",
				Value: config.NewRSRParsersMustCompile("1002", cfg.GeneralCfg().RSRSep),
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attrPrf, true)
	if err != nil {
		t.Error(err)
	}
	expIndexes = map[string]utils.StringSet{
		"*prefix:*req.Account:10": {
			"ATTR_TEST": {},
		},
	}
	rcvIndexes, err = dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, attrPrf.Tenant,
		utils.EmptyString, utils.NonTransactional, false, false)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIndexes, expIndexes) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expIndexes), utils.ToJSON(rcvIndexes))
	}

	// Add another indexed filter type
	attrPrf = &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_TEST",
		FilterIDs: []string{"*gt:~*req.Element:10", "*prefix:~*req.Account:10", "*string:~*req.Category:call"},
		Attributes: []*Attribute{
			{
				Type:  utils.MetaConstant,
				Path:  "~*req.Account",
				Value: config.NewRSRParsersMustCompile("1002", cfg.GeneralCfg().RSRSep),
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attrPrf, true)
	if err != nil {
		t.Error(err)
	}
	expIndexes = map[string]utils.StringSet{
		"*prefix:*req.Account:10": {
			"ATTR_TEST": {},
		},
		"*string:*req.Category:call": {
			"ATTR_TEST": {},
		},
	}
	rcvIndexes, err = dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, attrPrf.Tenant,
		utils.EmptyString, utils.NonTransactional, false, false)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIndexes, expIndexes) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expIndexes), utils.ToJSON(rcvIndexes))
	}

	// Remove an indexed filter type
	attrPrf = &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_TEST",
		FilterIDs: []string{"*gt:~*req.Element:10", "*prefix:~*req.Account:10"},
		Attributes: []*Attribute{
			{
				Type:  utils.MetaConstant,
				Path:  "~*req.Account",
				Value: config.NewRSRParsersMustCompile("1002", cfg.GeneralCfg().RSRSep),
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attrPrf, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIndexes, expIndexes) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expIndexes), utils.ToJSON(rcvIndexes))
	}
	expIndexes = map[string]utils.StringSet{
		"*prefix:*req.Account:10": {
			"ATTR_TEST": {},
		},
	}
	rcvIndexes, err = dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, attrPrf.Tenant,
		utils.EmptyString, utils.NonTransactional, false, false)
	if err != nil {
		t.Error(err)
	}

	// Remove the attribute profile
	err = dm.RemoveAttributeProfile(context.Background(), attrPrf.Tenant, attrPrf.ID, true)
	if err != nil {
		t.Error(err)
	}
	rcvIndexes, err = dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, attrPrf.Tenant,
		utils.EmptyString, utils.NonTransactional, false, false)
	if err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestLibIndexModifyAttrPrfFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	// Set a non-indexed type filter
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterThan,
				Element: "~*req.Cost",
				Values:  []string{"10"},
			},
		},
	}
	err := dm.SetFilter(context.Background(), fltr, true)
	if err != nil {
		t.Error(err)
	}

	// Create an AttributeProfile using the previously created filter
	attrPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_TEST",
		FilterIDs: []string{"fltr_test"},
		Attributes: []*Attribute{
			{
				Type:  utils.MetaConstant,
				Path:  "~*req.Account",
				Value: config.NewRSRParsersMustCompile("1002", cfg.GeneralCfg().RSRSep),
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attrPrf, true)
	if err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringSet{
		"*none:*any:*any": {
			"ATTR_TEST": {},
		},
	}
	rcvIndexes, err := dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, attrPrf.Tenant,
		utils.EmptyString, utils.NonTransactional, false, false)
	// if err != nil {
	// 	t.Error(err)
	// } else if !reflect.DeepEqual(rcvIndexes, expIndexes) {
	// 	t.Errorf("expected: <%+v>, \nreceived: <%+v>",
	// 		utils.ToJSON(expIndexes), utils.ToJSON(rcvIndexes))
	// }

	// Make the filter indexable
	fltr = &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Account",
				Values:  []string{"10"},
			},
		},
	}
	err = dm.SetFilter(context.Background(), fltr, true)
	if err != nil {
		t.Error(err)
	}

	expIndexes = map[string]utils.StringSet{
		"*prefix:*req.Account:10": {
			"ATTR_TEST": {},
		},
	}
	rcvIndexes, err = dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, attrPrf.Tenant,
		utils.EmptyString, utils.NonTransactional, false, false)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIndexes, expIndexes) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expIndexes), utils.ToJSON(rcvIndexes))
	}
}

func TestUpdateFilterIndexThreshold(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
			{
				Type:    utils.MetaNotExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values: []string{utils.DynamicDataPrefix + utils.MetaAccounts,
					utils.DynamicDataPrefix + utils.MetaStats,
					utils.DynamicDataPrefix + utils.MetaResources,
					utils.DynamicDataPrefix + utils.MetaLibPhoneNumber},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{utils.DynamicDataPrefix},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal"},
			},
			{
				Type:    utils.MetaExists,
				Element: "*req.Cost",
				Values:  []string{utils.DynamicDataPrefix},
			},
			{
				Type:    utils.MetaString,
				Element: "*req.Account",
				Values:  []string{"1001"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}

	thP := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "ThP1",
		FilterIDs:        []string{"fltr_test"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}

	if err := dm.SetThresholdProfile(context.Background(), thP, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*exists:*req.Cost:*any":     {"ThP1": {}},
		"*exists:*req.Cost:unRegVal": {"ThP1": {}},
		"*exists::*req.Cost":         {"ThP1": {}},
		"*notexists:*req.Cost:*none": {"ThP1": {}},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheThresholdFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
	}

	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaNotExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values: []string{utils.DynamicDataPrefix + utils.MetaAccounts,
					utils.DynamicDataPrefix + utils.MetaStats,
					utils.DynamicDataPrefix + utils.MetaResources,
					utils.DynamicDataPrefix + utils.MetaLibPhoneNumber},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{utils.DynamicDataPrefix},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal"},
			},
			{
				Type:    utils.MetaExists,
				Element: "*req.Cost",
				Values:  []string{utils.DynamicDataPrefix},
			},
			{
				Type:    utils.MetaExists,
				Element: "*req.Cost",
				Values:  []string{"unRegVal"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "*req.Usage",
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*exists:*req.Cost:unRegVal": {"ThP1": {}},
		"*exists::*req.Cost":         {"ThP1": {}},
		"*notexists:*req.Cost:*none": {"ThP1": {}},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheThresholdFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterIndexGetIndexErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
			{
				Type:    utils.MetaNotExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values: []string{utils.DynamicDataPrefix + utils.MetaAccounts,
					utils.DynamicDataPrefix + utils.MetaStats,
					utils.DynamicDataPrefix + utils.MetaResources,
					utils.DynamicDataPrefix + utils.MetaLibPhoneNumber},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{utils.DynamicDataPrefix},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal"},
			},
			{
				Type:    utils.MetaExists,
				Element: "*req.Cost",
				Values:  []string{utils.DynamicDataPrefix},
			},
			{
				Type:    utils.MetaExists,
				Element: "*req.Cost",
				Values:  []string{"unRegVal"},
			},
		},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
		},
	}

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestUpdateFilterIndexGetIndexErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, utils.ErrNotFound
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
			{
				Type:    utils.MetaNotExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values: []string{utils.DynamicDataPrefix + utils.MetaAccounts,
					utils.DynamicDataPrefix + utils.MetaStats,
					utils.DynamicDataPrefix + utils.MetaResources,
					utils.DynamicDataPrefix + utils.MetaLibPhoneNumber},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{utils.DynamicDataPrefix},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal"},
			},
			{
				Type:    utils.MetaExists,
				Element: "*req.Cost",
				Values:  []string{utils.DynamicDataPrefix},
			},
			{
				Type:    utils.MetaExists,
				Element: "*req.Cost",
				Values:  []string{"unRegVal"},
			},
		},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{},
			},
		},
	}

	// no index for this filter so  no update needed
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Errorf("Expected error <%v>, Received error <%v>", nil, err)
	}

}

func TestUpdateFilterIndexRemoveIndexesFromThresholdErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheThresholdFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{"val"},
			},
		},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{

			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values: []string{utils.DynamicDataPrefix + utils.MetaAccounts,
					utils.DynamicDataPrefix + utils.MetaStats,
					utils.DynamicDataPrefix + utils.MetaResources,
					utils.DynamicDataPrefix + utils.MetaLibPhoneNumber},
			},
		},
	}

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexRemoveIndexesFromThresholdErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheThresholdFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values:  []string{"val"},
			},
		},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{

			{
				Type:    utils.MetaExists,
				Element: "~*req.Cost",
				Values: []string{utils.DynamicDataPrefix + utils.MetaAccounts,
					utils.DynamicDataPrefix + utils.MetaStats,
					utils.DynamicDataPrefix + utils.MetaResources,
					utils.DynamicDataPrefix + utils.MetaLibPhoneNumber},
			},
		},
	}

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexStatIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}
	statQProfl := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "StatQueueProfile3",
		FilterIDs:   []string{"fltr_test"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	if err := dm.SetStatQueueProfile(context.Background(), statQProfl, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {
			"StatQueueProfile3": {},
		},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheStatFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {
			"StatQueueProfile3": {},
		},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheStatFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterIndexStatErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheStatFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexStatErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheStatFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexResourceIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}
	resProf := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RL1",
		FilterIDs: []string{"fltr_test"},
		Weights: utils.DynamicWeights{
			{
				Weight: 100,
			}},
		Limit:        2,
		ThresholdIDs: []string{"TEST_ACTIONS"},

		UsageTTL:          time.Millisecond,
		AllocationMessage: "ALLOC",
	}

	if err := dm.SetResourceProfile(context.Background(), resProf, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {"RL1": {}},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheResourceFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"RL1": {}},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheResourceFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterIndexResourcetErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheResourceFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexResourceErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheResourceFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexRouteIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}
	routeProf := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr_test"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	if err := dm.SetRouteProfile(context.Background(), routeProf, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {"ID": {}},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheRouteFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"ID": {}},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheRouteFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterIndexRouteErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheRouteFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexRouteErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheRouteFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexChargerIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}
	ChProf := &ChargerProfile{

		Tenant:       "cgrates.org",
		ID:           "CPP_3",
		FilterIDs:    []string{"fltr_test"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		weight: 20,
	}

	if err := dm.SetChargerProfile(context.Background(), ChProf, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {"CPP_3": {}},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheChargerFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"CPP_3": {}},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheChargerFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterIndexChargerErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheChargerFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexChargerErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheChargerFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexAccountsIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}
	accPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "1004",
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

	if err := dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {"1004": {}},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheAccountsFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"1004": {}},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheAccountsFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterIndexAccountsErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheAccountsFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexAccountsErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheAccountsFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexAttributeErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheAttributeFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexAttributeErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheAttributeFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexActionProfilesIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}
	ap := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr_test"},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
		},
		Schedule: "* * * * 1-5",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1002": {}}},
		Actions: []*APAction{
			{
				ID:        "APAct1",
				FilterIDs: []string{"FLTR1", "FLTR2", "FLTR3"},
				TTL:       time.Minute,
				Type:      "type2",
				Opts: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
				Diktats: []*APDiktat{},
			},
		},
	}

	if err := dm.SetActionProfile(context.Background(), ap, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {"ID": {}},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheActionProfilesFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"ID": {}},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheActionProfilesFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterActionProfilesIndexErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheActionProfilesFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexActionProfilesErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheActionProfilesFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexRateProfilesIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}
	rpp := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr_test"},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"FIRST_GI": {
				ID:        "FIRST_GI",
				FilterIDs: []string{"*string:~*req.Category:call"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Blocker: false,
			},
		},
		MinCost: utils.DecimalNaN,
		MaxCost: utils.DecimalNaN,
	}

	if err := dm.SetRateProfile(context.Background(), rpp, false, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {"ID": {}},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheRateProfilesFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"ID": {}},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheRateProfilesFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterRateProfilesIndexErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheRateProfilesFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexRateProfilesErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheRateProfilesFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexDispatcherIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(context.Background(), oldFlt, true); err != nil {
		t.Error(err)
	}
	dpp := &DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "ID",
		FilterIDs:      []string{"fltr_test"},
		Weight:         65,
		Strategy:       utils.MetaLoad,
		StrategyParams: map[string]interface{}{"k": "v"},
		Hosts: DispatcherHostProfiles{
			{
				ID:        "C2",
				FilterIDs: []string{"fltr3"},
				Weight:    10,
				Params: map[string]interface{}{
					"param3": "value3",
				},
				Blocker: false,
			},
		},
	}

	if err := dm.SetDispatcherProfile(context.Background(), dpp, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {"ID": {}},
	}

	getindx, err := dm.GetIndexes(context.Background(), utils.CacheDispatcherFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindx), utils.ToJSON(getindx))
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

	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"ID": {}},
	}
	getindxNew, err := dm.GetIndexes(context.Background(), utils.CacheDispatcherFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected \n<%v>, \nReceived \n<%v>", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}

}

func TestUpdateFilterDispatcherIndexErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheDispatcherFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := utils.ErrNotImplemented
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestUpdateFilterIndexDispatcherErr2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheDispatcherFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
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

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := UpdateFilterIndex(context.Background(), dm, oldFlt, newFlt); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestRemoveFilterIndexesForFilterErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return make(map[string]utils.StringSet), utils.ErrNotImplemented
		},
	}
	exp := make(utils.StringSet)
	exp.Add(utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1001"))
	if err := removeFilterIndexesForFilter(context.Background(), dm, utils.CacheThresholdFilterIndexes, "cgrates.org", []string{""}, exp); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestRemoveItemFromFilterIndexErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)

	expErr := `broken reference to filter: stringFilter for itemType: *attribute_filter_indexes and ID: stringFilterID`
	if err := removeItemFromFilterIndex(context.Background(), dm, utils.CacheAttributeFilterIndexes, utils.CGRateSorg, utils.MetaRating, "stringFilterID", []string{"stringFilter"}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestRemoveIndexFiltersItemCacheRemoveErr(t *testing.T) {

	tmpc := Cache
	defer func() {
		Cache = tmpc
		Cache = NewCacheS(config.CgrConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheReverseFilterIndexes].Replicate = true
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1ReplicateRemove: func(ctx *context.Context, args, reply interface{}) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	indexes := map[string]utils.StringSet{utils.CacheRateFilterIndexes: {"Rate1": {}, "Rate2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheReverseFilterIndexes, utils.ConcatenatedKey("cgrates.org", "fltrID1"), indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if err := removeIndexFiltersItem(context.Background(), dm, utils.CacheRateFilterIndexes, "cgrates.org", "RPP_1", []string{"fltrID1"}); err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestRemoveIndexFiltersItemSetIndexesErr(t *testing.T) {

	tmpc := Cache
	defer func() {
		Cache = tmpc
		Cache = NewCacheS(config.CgrConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	tntGrp := utils.ConcatenatedKey("cgrates.org", "fltrID1")
	tntxKey := utils.ConcatenatedKey(tntGrp, utils.CacheRateFilterIndexes)
	indexes := utils.StringSet{"Rate1": {}, "Rate2": {}}
	if err := Cache.Set(context.Background(), utils.CacheReverseFilterIndexes, tntxKey, indexes, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if err := removeIndexFiltersItem(context.Background(), dm, utils.CacheRateFilterIndexes, "cgrates.org", "RPP_1", []string{"fltrID1"}); err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}
