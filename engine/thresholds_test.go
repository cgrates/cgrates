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
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
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
		if err = dmTH.SetThreshold(th); err != nil {
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
		if err = dmTH.SetThreshold(th); err != nil {
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
		if err = dmTH.SetThreshold(th); err != nil {
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
		if err = dmTH.SetThreshold(th); err != nil {
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
		if err = dmTH.SetThreshold(th); err != nil {
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
	if err = dmTH.SetThreshold(th); err != nil {
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

func TestThresholdsUpdateThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, true), cfg.CacheCfg(), nil)
	thp := &ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "THUP1",
	}
	th := &Threshold{
		Tenant: thp.Tenant,
		ID:     thp.ID,
		Hits:   5,
		Snooze: time.Now(),
	}
	expTh := &Threshold{
		Tenant: thp.Tenant,
		ID:     thp.ID,
	}

	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.RemoveThreshold(th.Tenant, th.ID); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "THUP1",
		MaxHits: 1,
	}

	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "THUP1",
		MaxHits: 1,
		MinHits: 1,
	}

	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:   "cgrates.org",
		ID:       "THUP1",
		MaxHits:  1,
		MinHits:  1,
		MinSleep: 1,
	}

	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}
	if err := dm.RemoveThresholdProfile(thp.Tenant, thp.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

func TestThresholdsProcessEventAccountUpdateErrPartExec(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		ActionIDs: []string{"actPrf"},
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		tPrfl:  thPrf,
	}

	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]interface{}{
				utils.MetaEventType: utils.AccountUpdate,
			},
		},
	}
	expLog := `[WARNING] <ThresholdS> failed executing actions: actPrf, error: NOT_FOUND`
	if err := th.ProcessEvent(args, dm); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	}

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}
	utils.Logger.SetLogLevel(0)
}

func TestThresholdsProcessEventAsyncExecErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		ActionIDs: []string{"actPrf"},
		Async:     true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		tPrfl:  thPrf,
	}

	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	expLog := `[WARNING] <ThresholdS> failed executing actions: actPrf, error: NOT_FOUND`
	if err := th.ProcessEvent(args, dm); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestThresholdsProcessEvent3(t *testing.T) {
	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   3,
		MaxHits:   5,
		Weight:    10,
		ActionIDs: []string{"actPrf"},
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		tPrfl:  thPrf,
	}

	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]interface{}{
				utils.MetaEventType: utils.AccountUpdate,
			},
		},
	}
	if err := th.ProcessEvent(args, dm); err != nil {
		t.Error(err)
	}
}

func TestThresholdsShutdown(t *testing.T) {
	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil)

	expLog1 := `[INFO] <ThresholdS> shutdown initialized`
	expLog2 := `[INFO] <ThresholdS> shutdown complete`
	tS.Shutdown()

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog1) ||
		!strings.Contains(rcvLog, expLog2) {
		t.Errorf("expected logs <%+v> and <%+v> \n to be included in <%+v>",
			expLog1, expLog2, rcvLog)
	}
	utils.Logger.SetLogLevel(0)
}

func TestThresholdsStoreThresholdsOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil)

	value := &Threshold{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	Cache.SetWithoutReplicate(utils.CacheThresholds, "TH1", value, nil, true,
		utils.NonTransactional)
	tS.storedTdIDs.Add("TH1")
	exp := &Threshold{
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	tS.storeThresholds()

	if rcv, err := tS.dm.GetThreshold("cgrates.org", "TH1", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	Cache.Remove(utils.CacheThresholds, "TH1", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdsStoreThErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	tS := NewThresholdService(nil, cfg, nil)

	value := &Threshold{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	Cache.SetWithoutReplicate(utils.CacheThresholds, "TH1", value, nil, true,
		utils.NonTransactional)
	tS.storedTdIDs.Add("TH1")
	exp := utils.StringSet{
		"TH1": struct{}{},
	}
	expLog := `[WARNING] <ThresholdS> failed saving Threshold with tenant: cgrates.org and ID: TH1, error: NO_DATABASE_CONNECTION`
	tS.storeThresholds()

	if !reflect.DeepEqual(tS.storedTdIDs, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, tS.storedTdIDs)
	}
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v>\n to be in included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	Cache.Remove(utils.CacheThresholds, "TH1", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdsCacheGetErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil)

	value := &Threshold{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	Cache.SetWithoutReplicate(utils.CacheThresholds, "TH2", value, nil, true,
		utils.NonTransactional)
	tS.storedTdIDs.Add("TH1")
	expLog := `[WARNING] <ThresholdS> failed retrieving from cache treshold with ID: TH1`
	tS.storeThresholds()

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	Cache.Remove(utils.CacheThresholds, "TH2", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdNilDirtyField(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	if err := tS.StoreThreshold(th); err != nil {
		t.Error(err)
	}
}

func TestThresholdsSetCloneable(t *testing.T) {
	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_ID"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event:  map[string]interface{}{},
		},
		clnb: false,
	}

	exp := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_ID"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event:  map[string]interface{}{},
		},
		clnb: true,
	}
	args.SetCloneable(true)

	if !reflect.DeepEqual(args, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, args)
	}
}

func TestThresholdsRPCClone(t *testing.T) {
	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_ID"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "EventTest",
			Event:   make(map[string]interface{}),
			APIOpts: make(map[string]interface{}),
		},
		clnb: true,
	}

	exp := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_ID"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "EventTest",
			Event:   make(map[string]interface{}),
			APIOpts: make(map[string]interface{}),
		},
	}

	if out, err := args.RPCClone(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out.(*ThresholdsArgsProcessEvent), exp) {
		t.Errorf("expected: <%T>, \nreceived: <%T>",
			args, exp)
	}
}

func TestThresholdsProcessEventOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		MinHits: 2,
		MaxHits: 5,
		Weight:  10,
		Blocker: true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThdProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	expIDs := []string{"TH1"}
	expStored := utils.StringSet{
		"cgrates.org:TH1": struct{}{},
	}
	if rcvIDs, err := tS.processEvent(args.Tenant, args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	} else if !reflect.DeepEqual(tS.storedTdIDs, expStored) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expStored, tS.storedTdIDs)
	}

}

func TestThresholdsProcessEventStoreThOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		MinHits: 2,
		MaxHits: 5,
		Weight:  10,
		Blocker: true,
	}

	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThdProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	exp := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH2",
		Hits:   1,
	}
	expIDs := []string{"TH2"}
	if rcvIDs, err := tS.processEvent(args.Tenant, args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	} else if rcv, err := tS.dm.GetThreshold("cgrates.org", "TH2", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else {
		rcv.tPrfl = nil
		rcv.dirty = nil
		var snooze time.Time
		rcv.Snooze = snooze
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
		}
	}

}

func TestThresholdsProcessEventMaxHitsDMErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(nil, cfg, filterS)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		MinHits: 2,
		MaxHits: 5,
		Weight:  10,
		Blocker: true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH3",
		Hits:   4,
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH3"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThdProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	expLog1 := `[WARNING] <ThresholdService> failed removing from database non-recurrent threshold: cgrates.org:TH3, error: NO_DATABASE_CONNECTION`
	expLog2 := `[WARNING] <ThresholdService> failed removing from cache non-recurrent threshold: cgrates.org:TH3, error: NO_DATABASE_CONNECTION`

	if _, err := tS.processEvent(args.Tenant, args); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ErrPartiallyExecuted, err)
	}

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog1) ||
		!strings.Contains(rcvLog, expLog2) {
		t.Errorf("expected logs <%+v> and <%+v> to be included in: <%+v>",
			expLog1, expLog2, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestThresholdsProcessEventNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH5",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		MinHits: 2,
		MaxHits: 5,
		Weight:  10,
		Blocker: true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH5",
		Hits:   2,
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH6"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThdProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	if _, err := tS.processEvent(args.Tenant, args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

}
