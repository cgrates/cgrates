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
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        []string{"FLTR_TH_1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:          12,
			Blocker:          false,
			Weight:           20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:           "cgrates.org",
			ID:               "TH_2",
			FilterIDs:        []string{"FLTR_TH_2", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:          12,
			MinSleep:         5 * time.Minute,
			Blocker:          false,
			Weight:           20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:           "cgrates.org",
			ID:               "TH_3",
			FilterIDs:        []string{"FLTR_TH_3", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:          12,
			MinSleep:         5 * time.Minute,
			Blocker:          false,
			Weight:           20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
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
	dmTH.SetFilter(context.Background(), fltrTh1, true)
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
	dmTH.SetFilter(context.Background(), fltrTh2, true)
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
	dmTH.SetFilter(context.Background(), fltrTh3, true)
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(context.TODO(), th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(context.TODO(), th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(context.TODO(), th); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(context.TODO(), th.Tenant,
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
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        []string{"FLTR_TH_1", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:          12,
			Blocker:          false,
			Weight:           20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:           "cgrates.org",
			ID:               "TH_2",
			FilterIDs:        []string{"FLTR_TH_2", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:          12,
			MinSleep:         5 * time.Minute,
			Blocker:          false,
			Weight:           20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:           "cgrates.org",
			ID:               "TH_3",
			FilterIDs:        []string{"FLTR_TH_3", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:          12,
			MinSleep:         5 * time.Minute,
			Blocker:          false,
			Weight:           20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
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
	thServ = NewThresholdService(dmTH, defaultCfg, &FilterS{dm: dmTH, cfg: defaultCfg}, nil)
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
	dmTH.SetFilter(context.Background(), fltrTh1, true)
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
	dmTH.SetFilter(context.TODO(), fltrTh2, true)
	dmTH.SetFilter(context.TODO(), fltrTh3, true)
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(context.TODO(), th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(context.TODO(), th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(context.TODO(), th); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(context.TODO(), th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	thMatched, err := thServ.matchingThresholdsForEvent(context.TODO(), argsGetThresholds[0].Tenant, argsGetThresholds[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[0].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Hits, thMatched[0].Hits)
	}
	thMatched, err = thServ.matchingThresholdsForEvent(context.TODO(), argsGetThresholds[1].Tenant, argsGetThresholds[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[1].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Hits, thMatched[0].Hits)
	}
	thMatched, err = thServ.matchingThresholdsForEvent(context.TODO(), argsGetThresholds[2].Tenant, argsGetThresholds[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].Tenant) {
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
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
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
	dmTH.SetFilter(context.TODO(), fltrTh1, true)
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
	dmTH.SetFilter(context.TODO(), fltrTh2, true)
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
	dmTH.SetFilter(context.TODO(), fltrTh3, true)
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(context.Background(),th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(context.Background(),th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(context.Background(),th); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(context.Background(),th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	thIDs := []string{"TH_1"}
	if thMatched, err := thServ.processEvent(context.Background(),argsGetThresholds[0].Tenant, argsGetThresholds[0]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_2"}
	if thMatched, err := thServ.processEvent(context.Background(),argsGetThresholds[1].Tenant, argsGetThresholds[1]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_3"}
	if thMatched, err := thServ.processEvent(context.Background(),argsGetThresholds[2].Tenant, argsGetThresholds[2]); err != utils.ErrPartiallyExecuted {
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
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
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
	dmTH.SetFilter(context.TODO(), fltrTh1, true)
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
	dmTH.SetFilter(context.TODO(), fltrTh2, true)
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
	dmTH.SetFilter(context.TODO(), fltrTh3, true)
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(context.Background(),th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(context.Background(),th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(context.Background(),th); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	thIDs := []string{"TH_1"}
	if thMatched, err := thServ.processEvent(context.Background(),argsGetThresholds[0].Tenant, argsGetThresholds[0]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_2"}
	if thMatched, err := thServ.processEvent(context.Background(),argsGetThresholds[1].Tenant, argsGetThresholds[1]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_3"}
	if thMatched, err := thServ.processEvent(context.Background(),argsGetThresholds[2].Tenant, argsGetThresholds[2]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}
	thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[0].Tenant, argsGetThresholds[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	thMatched, err = thServ.matchingThresholdsForEvent(argsGetThresholds[1].Tenant, argsGetThresholds[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	thMatched, err = thServ.matchingThresholdsForEvent(argsGetThresholds[2].Tenant, argsGetThresholds[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].Tenant) {
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
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
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
	dmTH.SetFilter(context.TODO(), fltrTh1, true)
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
	dmTH.SetFilter(context.TODO(), fltrTh2, true)
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
	dmTH.SetFilter(context.TODO(), fltrTh3, true)

	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(context.Background(),th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(context.Background(),th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(context.Background(),th); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(context.Background(),th.Tenant,
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
		MaxHits:   12,
		Blocker:   false,
		Weight:    20.0,
		ActionProfileIDs: []string{"ACT_1", "ACT_2"},
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
	if err = dmTH.SetThresholdProfile(context.Background(),thPrf, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if temptTh, err := dmTH.GetThresholdProfile(context.Background(),thPrf.Tenant,
		thPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thPrf, temptTh) {
		t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
	}
	if err = dmTH.SetThreshold(context.Background(),th); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if temptTh, err := dmTH.GetThreshold(context.Background(),th.Tenant,
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
	thMatched, err := thServ.matchingThresholdsForEvent(ev.Tenant, ev)
	if err != nil {
		t.Fatalf("Error: %+v", err)
	}
	thMatched.unlock()
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
*/
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

	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.RemoveThreshold(context.Background(), th.Tenant, th.ID); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "THUP1",
		MaxHits: 1,
	}

	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "THUP1",
		MaxHits: 1,
		MinHits: 1,
	}

	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:   "cgrates.org",
		ID:       "THUP1",
		MaxHits:  1,
		MinHits:  1,
		MinSleep: 1,
	}

	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}
	if err := dm.RemoveThresholdProfile(context.Background(), thp.Tenant, thp.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

// func TestThresholdsProcessEventAccountUpdateErrPartExec(t *testing.T) {
// 	utils.Logger.SetLogLevel(4)
// 	utils.Logger.SetSyslog(nil)

// 	var buf bytes.Buffer
// 	log.SetOutput(&buf)
// 	defer func() {
// 		log.SetOutput(os.Stderr)
// 	}()

// 	thPrf := &ThresholdProfile{
// 		Tenant:    "cgrates.org",
// 		ID:        "TH1",
// 		FilterIDs: []string{"*string:~*req.Account:1001"},
// 		MinHits:   2,
// 		MaxHits:   5,
// 		Weight:    10,
// 		ActionIDs: []string{"actPrf"},
// 	}
// 	th := &Threshold{
// 		Tenant: "cgrates.org",
// 		ID:     "TH1",
// 		Hits:   2,
// 		tPrfl:  thPrf,
// 	}

// 	args := &ThresholdsArgsProcessEvent{
// 		ThresholdIDs: []string{"TH1"},
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "ThresholdProcessEvent",
// 			Event: map[string]interface{}{
// 				utils.AccountField: "1001",
// 			},
// 			APIOpts: map[string]interface{}{
// 				utils.MetaEventType: utils.AccountUpdate,
// 			},
// 		},
// 	}
// 	expLog := `[WARNING] <ThresholdS> failed executing actions: actPrf, error: NOT_FOUND`
// 	if err := processEventWithThreshold(context.Background(), args, dm); err == nil ||
// 		err != utils.ErrPartiallyExecuted {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
// 	}

// 	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
// 		t.Errorf("expected log <%+v> \nto be included in: <%+v>", expLog, rcvLog)
// 	}
// 	utils.Logger.SetLogLevel(0)
// }

func TestThresholdsProcessEventAsyncExecErr(t *testing.T) {
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().ActionSConns = []string{"actPrfID"}
	cfg.RPCConns()["actPrfID"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	connMgr := NewConnManager(cfg, make(map[string]chan birpc.ClientConnector))

	thPrf := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "TH1",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		MinHits:          2,
		MaxHits:          5,
		Weight:           10,
		ActionProfileIDs: []string{"actPrfID"},
		Async:            true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		tPrfl:  thPrf,
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThresholdProcessEvent",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}
	expLog := `[WARNING] <ThresholdS> failed executing actions for threshold: cgrates.org:TH1, error: DISCONNECTED`
	if err := processEventWithThreshold(context.Background(), connMgr, thPrf.ActionProfileIDs,
		args, th); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
}

// func TestThresholdsProcessEvent3(t *testing.T) {
// 	thPrf := &ThresholdProfile{
// 		Tenant:    "cgrates.org",
// 		ID:        "TH1",
// 		FilterIDs: []string{"*string:~*req.Account:1001"},
// 		MinHits:   3,
// 		MaxHits:   5,
// 		Weight:    10,
// 		ActionIDs: []string{"actPrf"},
// 	}
// 	th := &Threshold{
// 		Tenant: "cgrates.org",
// 		ID:     "TH1",
// 		Hits:   2,
// 		tPrfl:  thPrf,
// 	}

// 	args := &ThresholdsArgsProcessEvent{
// 		ThresholdIDs: []string{"TH1"},
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "ThresholdProcessEvent",
// 			Event: map[string]interface{}{
// 				utils.AccountField: "1001",
// 			},
// 			APIOpts: map[string]interface{}{
// 				utils.MetaEventType: utils.AccountUpdate,
// 			},
// 		},
// 	}
// 	if err := processEventWithThreshold(args, dm); err != nil {
// 		t.Error(err)
// 	}
// }

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
	tS := NewThresholdService(dm, cfg, nil, nil)

	expLog1 := `[INFO] <ThresholdS> shutdown initialized`
	expLog2 := `[INFO] <ThresholdS> shutdown complete`
	tS.Shutdown(context.Background())

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
	tS := NewThresholdService(dm, cfg, nil, nil)

	exp := &Threshold{
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	Cache.SetWithoutReplicate(utils.CacheThresholds, "cgrates.org:TH1", exp, nil, true,
		utils.NonTransactional)
	tS.storedTdIDs.Add("cgrates.org:TH1")
	tS.storeThresholds(context.Background())

	if rcv, err := tS.dm.GetThreshold(context.Background(), "cgrates.org", "TH1", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	Cache.Remove(context.Background(), utils.CacheThresholds, "cgrates.org:TH1", true, utils.NonTransactional)
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
	tS := NewThresholdService(nil, cfg, nil, nil)

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
	tS.storeThresholds(context.Background())

	if !reflect.DeepEqual(tS.storedTdIDs, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, tS.storedTdIDs)
	}
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v>\n to be in included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	Cache.Remove(context.Background(), utils.CacheThresholds, "TH1", true, utils.NonTransactional)
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
	tS := NewThresholdService(dm, cfg, nil, nil)

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
	tS.storeThresholds(context.Background())

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	Cache.Remove(context.Background(), utils.CacheThresholds, "TH2", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdNilDirtyField(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil, nil)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	if err := tS.StoreThreshold(context.Background(), th); err != nil {
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

func TestThresholdsProcessEventStoreThOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
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
	if rcvIDs, err := tS.processEvent(context.Background(), args.Tenant, args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	} else if rcv, err := tS.dm.GetThreshold(context.Background(), "cgrates.org", "TH2", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else {
		rcv.tPrfl = nil
		rcv.dirty = nil
		rcv.Snooze = time.Time{}
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
	tmp := config.CgrConfig()
	tmpC := Cache
	tmpCMgr := connMgr

	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	connMgr = NewConnManager(cfg, make(map[string]chan birpc.ClientConnector))
	dm := NewDataManager(data, cfg.CacheCfg(), connMgr)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(nil, cfg, filterS, connMgr)
	Cache = NewCacheS(cfg, dm, nil)

	defer func() {
		connMgr = tmpCMgr
		Cache = tmpC
		config.SetCgrConfig(tmp)
		log.SetOutput(os.Stderr)
	}()
	thPrf := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "TH3",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		MinHits:          2,
		MaxHits:          5,
		Weight:           10,
		ActionProfileIDs: []string{utils.MetaNone},
		Blocker:          true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH3",
		Hits:   4,
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
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
	expLog2 := `[WARNING] <ThresholdService> failed removing from cache non-recurrent threshold: cgrates.org:TH3, error: DISCONNECTED`

	if _, err := tS.processEvent(context.Background(), args.Tenant, args); err == nil ||
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
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH5",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH5",
		Hits:   2,
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
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

	if _, err := tS.processEvent(context.Background(), args.Tenant, args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

}

func TestThresholdsV1ProcessEventOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "TH1",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{utils.MetaNone},
		MinHits:          2,
		MaxHits:          5,
		Weight:           10,
		Blocker:          true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "TH2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{utils.MetaNone},
		MinHits:          0,
		MaxHits:          7,
		Weight:           20,
		Blocker:          false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			ID: "V1ProcessEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	var reply []string
	exp := []string{"TH1", "TH2"}
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
		}
	}
}

func TestThresholdsV1ProcessEventPartExecErr(t *testing.T) {
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
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH4",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   0,
		MaxHits:   7,
		Weight:    20,
		Blocker:   false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			ID: "V1ProcessEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	expLog := `[WARNING] <ThresholdService> threshold: cgrates.org:TH4, ignoring event: cgrates.org:V1ProcessEventTest, error: MANDATORY_IE_MISSING: [connIDs]`
	var reply []string
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	} else {
		if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
			t.Errorf("expected log <%+v> to be included in: <%+v>",
				expLog, rcvLog)
		}
	}

	utils.Logger.SetLogLevel(0)
}

func TestThresholdsV1ProcessEventMissingArgs(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   0,
		MaxHits:   7,
		Weight:    20,
		Blocker:   false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			ID: "V1ProcessEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	args = &ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	var reply []string
	experr := `MANDATORY_IE_MISSING: [ID]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &ThresholdsArgsProcessEvent{
		CGREvent: nil,
	}
	experr = `MANDATORY_IE_MISSING: [CGREvent]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			ID:    "V1ProcessEventTest",
			Event: nil,
		},
	}
	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestThresholdsV1GetThresholdOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}

	expTh := Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   0,
	}
	var rplyTh Threshold
	if err := tS.V1GetThreshold(context.Background(), &utils.TenantID{
		ID: "TH1",
	}, &rplyTh); err != nil {
		t.Error(err)
	} else {
		var snooze time.Time
		rplyTh.dirty = nil
		rplyTh.Snooze = snooze
		if !reflect.DeepEqual(rplyTh, expTh) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expTh), utils.ToJSON(rplyTh))
		}
	}
}

func TestThresholdsV1GetThresholdNotFoundErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}

	var rplyTh Threshold
	if err := tS.V1GetThreshold(context.Background(), &utils.TenantID{
		ID: "TH2",
	}, &rplyTh); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdMatchingThresholdForEventLocks(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ThresholdProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      fmt.Sprintf("TH%d", i),
			Weight:  20.00,
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	dm.RemoveThreshold(context.Background(), "cgrates.org", "TH1")
	mth, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", &ThresholdsArgsProcessEvent{
		ThresholdIDs: ids.AsSlice(),
		CGREvent:     new(utils.CGREvent),
	})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mth.unlock()
	for _, rPrf := range prfs {
		if rPrf.ID == "TH1" {
			if rPrf.isLocked() {
				t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
			}
			continue
		}
		if !rPrf.isLocked() {
			t.Fatalf("Expected profile to be locked %q", rPrf.ID)
		}
		if r, err := dm.GetThreshold(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if !r.isLocked() {
			t.Fatalf("Expected Threshold to be locked %q", rPrf.ID)
		}
	}

}

func TestThresholdMatchingThresholdForEventLocks2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ThresholdProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      fmt.Sprintf("TH%d", i),
			Weight:  20.00,
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	rPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH20",
		FilterIDs: []string{"FLTR_RES_201"},
		Weight:    20.00,
		MaxHits:   5,
	}
	err = db.SetThresholdProfileDrv(context.Background(), rPrf)
	if err != nil {
		t.Fatal(err)
	}
	prfs = append(prfs, rPrf)
	ids.Add(rPrf.ID)
	_, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", &ThresholdsArgsProcessEvent{
		ThresholdIDs: ids.AsSlice(),
		CGREvent:     new(utils.CGREvent),
	})
	expErr := utils.ErrPrefixNotFound(rPrf.FilterIDs[0])
	if err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s ,received: %+v", expErr, err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
		if rPrf.ID == "TH20" {
			continue
		}
		if r, err := dm.GetThreshold(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected Threshold to not be locked %q", rPrf.ID)
		}
	}
}

func TestThresholdMatchingThresholdForEventLocksBlocker(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ThresholdProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      fmt.Sprintf("TH%d", i),
			Weight:  float64(10 - i),
			Blocker: i == 4,
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	mres, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", &ThresholdsArgsProcessEvent{
		ThresholdIDs: ids.AsSlice(),
		CGREvent:     new(utils.CGREvent),
	})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	if len(mres) != 5 {
		t.Fatal("Expected 6 Thresholds")
	}
	for _, rPrf := range prfs[5:] {
		if rPrf.isLocked() {
			t.Errorf("Expected profile to not be locked %q", rPrf.ID)
		}
		if r, err := dm.GetThreshold(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected Threshold to not be locked %q", rPrf.ID)
		}
	}
	for _, rPrf := range prfs[:5] {
		if !rPrf.isLocked() {
			t.Errorf("Expected profile to be locked %q", rPrf.ID)
		}
		if r, err := dm.GetThreshold(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if !r.isLocked() {
			t.Fatalf("Expected Threshold to be locked %q", rPrf.ID)
		}
	}
}

func TestThresholdMatchingThresholdForEventLocks3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	prfs := make([]*ThresholdProfile, 0)
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tnt, id string) (*ThresholdProfile, error) {
			if id == "TH1" {
				return nil, utils.ErrNotImplemented
			}
			rPrf := &ThresholdProfile{
				Tenant:  "cgrates.org",
				ID:      id,
				MaxHits: 5,
				Weight:  20.00,
			}
			Cache.Set(ctx, utils.CacheThresholds, rPrf.TenantID(), &Threshold{
				Tenant: rPrf.Tenant,
				ID:     rPrf.ID,
			}, nil, true, utils.NonTransactional)
			prfs = append(prfs, rPrf)
			return rPrf, nil
		},
	}
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		ids.Add(fmt.Sprintf("TH%d", i))
	}
	_, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", &ThresholdsArgsProcessEvent{
		ThresholdIDs: ids.AsSlice(),
		CGREvent:     new(utils.CGREvent),
	})
	if err != utils.ErrNotImplemented {
		t.Fatalf("Error: %+v", err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
	}

}

func TestThresholdMatchingThresholdForEventLocks4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ThresholdProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      fmt.Sprintf("TH%d", i),
			Weight:  20.00,
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	ids.Add("TH20")
	mres, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", &ThresholdsArgsProcessEvent{
		ThresholdIDs: ids.AsSlice(),
		CGREvent:     new(utils.CGREvent),
	})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	for _, rPrf := range prfs {
		if !rPrf.isLocked() {
			t.Errorf("Expected profile to be locked %q", rPrf.ID)
		}
		if r, err := dm.GetThreshold(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if !r.isLocked() {
			t.Fatalf("Expected Threshold to be locked %q", rPrf.ID)
		}
	}

}

func TestThresholdMatchingThresholdForEventLocks5(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), NewConnManager(cfg, make(map[string]chan birpc.ClientConnector)))
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	cfg.DataDbCfg().RmtConns = []string{"test"}
	cfg.DataDbCfg().Items[utils.CacheThresholds].Remote = true
	config.SetCgrConfig(cfg)
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ThresholdProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      fmt.Sprintf("TH%d", i),
			Weight:  20.00,
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	dm.RemoveThreshold(context.Background(), "cgrates.org", "TH1")
	_, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", &ThresholdsArgsProcessEvent{
		ThresholdIDs: ids.AsSlice(),
		CGREvent:     new(utils.CGREvent),
	})
	if err != utils.ErrDisconnected {
		t.Errorf("Error: %+v", err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
		if rPrf.ID == "TH1" {
			continue
		}
		if r, err := dm.GetThreshold(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected Threshold to not be locked %q", rPrf.ID)
		}
	}
}

func TestThresholdsRunBackupStoreIntervalLessThanZero(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := &ThresholdService{
		dm:          dm,
		storedTdIDs: make(utils.StringSet),
		cgrcfg:      cfg,
		filterS:     filterS,
		loopStopped: make(chan struct{}, 1),
		stopBackup:  make(chan struct{}),
	}

	tS.runBackup(context.Background())

	if len(tS.loopStopped) != 1 {
		t.Errorf("expected loopStopped field to have only one element, received: <%+v>", len(tS.loopStopped))
	}
}

func TestThresholdsRunBackupStop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := &ThresholdService{
		dm: dm,
		storedTdIDs: utils.StringSet{
			"Th1": struct{}{},
		},
		cgrcfg:      cfg,
		filterS:     filterS,
		loopStopped: make(chan struct{}, 1),
		stopBackup:  make(chan struct{}),
	}

	value := &Threshold{
		Hits:   0,
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "Th1",
	}

	Cache.SetWithoutReplicate(utils.CacheThresholds, "Th1", value, nil, true,
		utils.NonTransactional)

	exp := &Threshold{
		Hits:   0,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "Th1",
	}

	go func() {
		time.Sleep(9 * time.Millisecond)
		close(tS.stopBackup)
	}()
	tS.runBackup(context.Background())

	if rcv, err := tS.dm.GetThreshold(context.Background(), "cgrates.org", "Th1", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	if len(tS.loopStopped) != 1 {
		t.Errorf("expected loopStopped field to have only one element, received: <%+v>", len(tS.loopStopped))
	}
}

func TestThresholdsReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := &ThresholdService{
		dm:          dm,
		filterS:     filterS,
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cgrcfg:      cfg,
	}
	tS.loopStopped <- struct{}{}
	tS.Reload(context.Background())
	close(tS.stopBackup)
}

func TestThresholdsStartLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := &ThresholdService{
		dm:          dm,
		filterS:     filterS,
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cgrcfg:      cfg,
	}

	tS.StartLoop(context.Background())
	time.Sleep(10 * time.Millisecond)

	if len(tS.loopStopped) != 1 {
		t.Errorf("expected loopStopped field to have only one element, received: <%+v>", len(tS.loopStopped))
	}
}

func TestThresholdsV1GetThresholdsForEventOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{},
		CGREvent: &utils.CGREvent{
			ID: "TestGetThresholdsForEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	exp := Thresholds{
		{
			Tenant: "cgrates.org",
			Hits:   0,
			ID:     "TH1",
			tPrfl:  thPrf,
			dirty:  utils.BoolPointer(false),
		},
	}
	var reply Thresholds
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
	}
}

func TestThresholdsV1GetThresholdsForEventMissingArgs(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{},
		CGREvent:     nil,
	}

	experr := `MANDATORY_IE_MISSING: [CGREvent]`
	var reply Thresholds
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}

	args = &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}

	args = &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestGetThresholdsForEvent",
			Event:  nil,
		},
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}
}

func TestThresholdsV1GetThresholdIDsOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "TH2",
		MinHits: 0,
		MaxHits: 7,
		Weight:  20,
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	expIDs := []string{"TH1", "TH2"}
	var reply []string
	if err := tS.V1GetThresholdIDs(context.Background(), "", &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, reply)
		}
	}
}

func TestThresholdsV1GetThresholdIDsGetKeysForPrefixErr(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	var reply []string
	if err := tS.V1GetThresholdIDs(context.Background(), "", &reply); err == nil ||
		err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestThresholdsV1ResetThresholdOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	th := &Threshold{
		tPrfl:  thPrf,
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		dirty:  utils.BoolPointer(false),
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}

	expStored := utils.StringSet{
		"cgrates.org:TH1": {},
	}
	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantID{
		ID: "TH1",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: <%q>", reply)
	} else if th.Hits != 0 {
		t.Errorf("expected nr. of hits to be 0, received: <%+v>", th.Hits)
	} else if !reflect.DeepEqual(tS.storedTdIDs, expStored) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expStored, tS.storedTdIDs)
	}
}

func TestThresholdsV1ResetThresholdErrNotFound(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	th := &Threshold{
		tPrfl:  thPrf,
		Tenant: "cgrates.org",
		ID:     "TH2",
		Hits:   2,
		dirty:  utils.BoolPointer(false),
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}

	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantID{
		ID: "TH1",
	}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsV1ResetThresholdNegativeStoreIntervalOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	th := &Threshold{
		tPrfl:  thPrf,
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		dirty:  utils.BoolPointer(false),
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}

	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantID{
		ID: "TH1",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: <%q>", reply)
	} else if th.Hits != 0 {
		t.Errorf("expected nr. of hits to be 0, received: <%+v>", th.Hits)
	}
}

func TestThresholdsV1ResetThresholdNegativeStoreIntervalErr(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(nil, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	th := &Threshold{
		tPrfl:  thPrf,
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		dirty:  utils.BoolPointer(false),
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}

	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantID{
		ID: "TH1",
	}, &reply); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestThresholdsLockUnlockThresholdProfiles(t *testing.T) {
	thPrf := &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "thPrf",
		Weight:  10,
		MaxHits: 5,
		MinHits: 2,
	}

	//lock profile with empty lkID parameter
	thPrf.lock(utils.EmptyString)

	if !thPrf.isLocked() {
		t.Fatal("expected profile to be locked")
	} else if thPrf.lkID == utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be non-empty")
	}

	//unlock previously locked profile
	thPrf.unlock()

	if thPrf.isLocked() {
		t.Fatal("expected profile to be unlocked")
	} else if thPrf.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}

	//unlock an already unlocked profile - nothing happens
	thPrf.unlock()

	if thPrf.isLocked() {
		t.Fatal("expected profile to be unlocked")
	} else if thPrf.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}
}

func TestThresholdsLockUnlockThresholds(t *testing.T) {
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "thPrf",
	}

	//lock resource with empty lkID parameter
	th.lock(utils.EmptyString)

	if !th.isLocked() {
		t.Fatal("expected resource to be locked")
	} else if th.lkID == utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be non-empty")
	}

	//unlock previously locked resource
	th.unlock()

	if th.isLocked() {
		t.Fatal("expected resource to be unlocked")
	} else if th.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}

	//unlock an already unlocked resource - nothing happens
	th.unlock()

	if th.isLocked() {
		t.Fatal("expected resource to be unlocked")
	} else if th.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}
}

func TestThresholdsMatchingThresholdsForEventNotFoundErr(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEvent",
			Event:  map[string]interface{}{},
		},
	}

	if _, err := tS.matchingThresholdsForEvent(context.Background(), "cgrates.org", args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsStoreThresholdCacheSetErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr = NewConnManager(cfg, make(map[string]chan birpc.ClientConnector))
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, connMgr)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		dirty:  utils.BoolPointer(true),
	}

	expLog := `[WARNING] <ThresholdService> failed caching Threshold with ID: cgrates.org:TH1, error: DISCONNECTED`
	if err := tS.StoreThreshold(context.Background(), th); err == nil ||
		err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrDisconnected, err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcv)
	}

	utils.Logger.SetLogLevel(0)
}

func TestThreholdsMatchingThresholdsForEventDoesNotPass(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	args := &ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
	}
	if _, err := tS.matchingThresholdsForEvent(context.Background(), "cgrates.org",
		args); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}
}
