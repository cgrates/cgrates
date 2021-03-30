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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestThresholdsSort(t *testing.T) {
	ts := Thresholds{
		&Threshold{tPrfl: &ThresholdProfile{ID: "FIRST", Weight: 30.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "SECOND", Weight: 40.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "THIRD", Weight: 30.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "FOURTH", Weight: 35.0}},
	}
	ts.Sort()
	eInst := Thresholds{
		&Threshold{tPrfl: &ThresholdProfile{ID: "SECOND", Weight: 40.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "FOURTH", Weight: 35.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "FIRST", Weight: 30.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "THIRD", Weight: 30.0}},
	}
	if !reflect.DeepEqual(eInst, ts) {
		t.Errorf("expecting: %+v, received: %+v", eInst, ts)
	}
}

func TestThresholdsCache(t *testing.T) {
	var dmTH *DataManager
	tPrfls := []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
	}
	ths := Thresholds{
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_1",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_2",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_3",
			Hits:   0,
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ThresholdSCfg().StoreInterval = 0
	defaultCfg.ThresholdSCfg().StringIndexedFields = nil
	defaultCfg.ThresholdSCfg().PrefixIndexedFields = nil

	fltrTh1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh1, true)
	fltrTh2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh2, true)
	fltrTh3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(fltrTh3, true)
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(th, 0, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}

}

func TestThresholdsmatchingThresholdsForEvent(t *testing.T) {
	var dmTH *DataManager
	var thServ *ThresholdService
	var tPrfls = []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
	}
	ths := Thresholds{
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_1",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_2",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_3",
			Hits:   0,
		},
	}
	argsGetThresholds := []*ThresholdsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_1",
					"Weight":    "10.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_2",
					"Weight":    "20.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "ThresholdPrefix123",
				},
			},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ThresholdSCfg().StoreInterval = 0
	defaultCfg.ThresholdSCfg().StringIndexedFields = nil
	defaultCfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ = NewThresholdService(dmTH, defaultCfg, &FilterS{dm: dmTH, cfg: defaultCfg})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	fltrTh1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh1, true)
	fltrTh2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh2, true)
	fltrTh3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(fltrTh3, true)
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(th, 0, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[0].Tenant, argsGetThresholds[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[0].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Hits, thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[1].Tenant, argsGetThresholds[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[1].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Hits, thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[2].Tenant, argsGetThresholds[2]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[2].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[2].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Hits, thMatched[0].Hits)
	}
}

/*
func TestThresholdsProcessEvent(t *testing.T) {
	var dmTH *DataManager
	var thServ *ThresholdService
	var tPrfls = []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
	}
	ths := Thresholds{
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_1",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_2",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_3",
			Hits:   0,
		},
	}
	argsGetThresholds := []*ThresholdsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_1",
					"Weight":    "10.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_2",
					"Weight":    "20.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "ThresholdPrefix123",
				},
			},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ThresholdSCfg().StoreInterval = 0
	defaultCfg.ThresholdSCfg().StringIndexedFields = nil
	defaultCfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ = NewThresholdService(dmTH, defaultCfg, &FilterS{dm: dmTH, cfg: defaultCfg})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	fltrTh1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh1, true)
	fltrTh2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh2, true)
	fltrTh3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(fltrTh3, true)
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(th, 0, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	thIDs := []string{"TH_1"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[0].Tenant, argsGetThresholds[0]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_2"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[1].Tenant, argsGetThresholds[1]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_3"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[2].Tenant, argsGetThresholds[2]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}
}

func TestThresholdsVerifyIfExecuted(t *testing.T) {
	var dmTH *DataManager
	var thServ *ThresholdService
	var tPrfls = []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
	}
	ths := Thresholds{
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_1",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_2",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_3",
			Hits:   0,
		},
	}
	argsGetThresholds := []*ThresholdsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_1",
					"Weight":    "10.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_2",
					"Weight":    "20.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "ThresholdPrefix123",
				},
			},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ThresholdSCfg().StoreInterval = 0
	defaultCfg.ThresholdSCfg().StringIndexedFields = nil
	defaultCfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ = NewThresholdService(dmTH, defaultCfg, &FilterS{dm: dmTH, cfg: defaultCfg})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	fltrTh1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh1, true)
	fltrTh2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh2, true)
	fltrTh3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(fltrTh3, true)
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(th, 0, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	thIDs := []string{"TH_1"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[0].Tenant, argsGetThresholds[0]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_2"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[1].Tenant, argsGetThresholds[1]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_3"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[2].Tenant, argsGetThresholds[2]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}
	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[0].Tenant, argsGetThresholds[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[1].Tenant, argsGetThresholds[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[2].Tenant, argsGetThresholds[2]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[2].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}
}
func TestThresholdsProcessEvent2(t *testing.T) {
	var dmTH *DataManager
	var thServ *ThresholdService
	tPrfls := []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
	}
	ths := Thresholds{
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_1",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_2",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_3",
			Hits:   0,
		},
	}
	argsGetThresholds := []*ThresholdsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_1",
					"Weight":    "10.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_2",
					"Weight":    "20.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "ThresholdPrefix123",
				},
			},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ThresholdSCfg().StoreInterval = 0
	defaultCfg.ThresholdSCfg().StringIndexedFields = nil
	defaultCfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ = NewThresholdService(dmTH, defaultCfg, &FilterS{dm: dmTH, cfg: defaultCfg})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	fltrTh1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh1, true)
	fltrTh2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh2, true)
	fltrTh3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(fltrTh3, true)

	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(th, 0, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH_4",
		FilterIDs: []string{"FLTR_TH_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:   12,
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1", "ACT_2"},
		Async:     false,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_4",
		Hits:   0,
	}
	ev := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH_1", "TH_2", "TH_3", "TH_4"},
		CGREvent:     argsGetThresholds[0].CGREvent,
	}
	if err = dmTH.SetThresholdProfile(thPrf, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if temptTh, err := dmTH.GetThresholdProfile(thPrf.Tenant,
		thPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thPrf, temptTh) {
		t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
	}
	if err = dmTH.SetThreshold(th, 0, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if temptTh, err := dmTH.GetThreshold(th.Tenant,
		th.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(th, temptTh) {
		t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
	}
	thIDs := []string{"TH_1", "TH_4"}
	thIDsRev := []string{"TH_4", "TH_1"}
	if thMatched, err := thServ.processEvent(ev.Tenant, ev); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) && !reflect.DeepEqual(thIDsRev, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(ev.Tenant, ev); err != nil {
		t.Errorf("Error: %+v", err)
	} else {
		for _, thM := range thMatched {
			if !reflect.DeepEqual(thPrf.Tenant, thM.Tenant) {
				t.Errorf("Expecting: %+v, received: %+v", thPrf.Tenant, thM.Tenant)
			} else if reflect.DeepEqual(thIDs[0], thM.ID) && thM.Hits != 1 {
				t.Errorf("Expecting: 1 for %+v, received: %+v", thM.ID, thM.Hits)
			} else if reflect.DeepEqual(thIDs[1], thM.ID) && thM.Hits != 1 {
				t.Errorf("Expecting: 1 for %+v, received: %+v", thM.ID, thM.Hits)
			}
		}
	}
}
*/
