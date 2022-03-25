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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
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
	_, err = ComputeIndexes(context.Background(), dm, "cgrates.org", utils.EmptyString, utils.CacheThresholdFilterIndexes,
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
