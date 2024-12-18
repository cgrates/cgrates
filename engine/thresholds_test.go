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
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestThresholdsSort(t *testing.T) {
	ts := Thresholds{
		&Threshold{weight: 30.0, tPrfl: &ThresholdProfile{ID: "FIRST"}},
		&Threshold{weight: 40.0, tPrfl: &ThresholdProfile{ID: "SECOND"}},
		&Threshold{weight: 30.0, tPrfl: &ThresholdProfile{ID: "THIRD"}},
		&Threshold{weight: 35.0, tPrfl: &ThresholdProfile{ID: "FOURTH"}},
	}
	ts.Sort()
	eInst := Thresholds{
		&Threshold{weight: 40.0, tPrfl: &ThresholdProfile{ID: "SECOND"}},
		&Threshold{weight: 35.0, tPrfl: &ThresholdProfile{ID: "FOURTH"}},
		&Threshold{weight: 30.0, tPrfl: &ThresholdProfile{ID: "FIRST"}},
		&Threshold{weight: 30.0, tPrfl: &ThresholdProfile{ID: "THIRD"}},
	}
	if !reflect.DeepEqual(eInst, ts) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eInst), utils.ToJSON(ts))
	}
}

func TestThresholdsCache(t *testing.T) {
	var dmTH *DataManager
	tPrfls := []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
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

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil

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
		if err := dmTH.SetThresholdProfile(context.TODO(), th, true); err != nil {
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
		if err := dmTH.SetThreshold(context.TODO(), th); err != nil {
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
	var thServ *ThresholdS
	var tPrfls = []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
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
	argsGetThresholds := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "TH_1",
				"Weight":    "10.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "TH_2",
				"Weight":    "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "ThresholdPrefix123",
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ = NewThresholdService(dmTH, cfg, &FilterS{dm: dmTH, cfg: cfg}, nil)

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
		if err := dmTH.SetThresholdProfile(context.TODO(), th, true); err != nil {
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
		if err := dmTH.SetThreshold(context.TODO(), th); err != nil {
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
		argsGetThresholds := []*utils.CGREvent{
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "TH_1",
						"Weight":    "10.0",
					},
			},
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "TH_2",
						"Weight":    "20.0",
				},
			},
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "ThresholdPrefix123",
				},
			},
		}

		cfg := config.NewDefaultCGRConfig()
		data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
		dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
		cfg.ThresholdSCfg().StoreInterval = 0
		cfg.ThresholdSCfg().StringIndexedFields = nil
		cfg.ThresholdSCfg().PrefixIndexedFields = nil
		thServ = NewThresholdService(dmTH, cfg, &FilterS{dm: dmTH, cfg: cfg})
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
		argsGetThresholds := []*utils.CGREvent{
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "TH_1",
						"Weight":    "10.0",
				},
			},
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "TH_2",
						"Weight":    "20.0",
				},
			},
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "ThresholdPrefix123",
					},
			},
		}

		cfg := config.NewDefaultCGRConfig()
		data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
		dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
		cfg.ThresholdSCfg().StoreInterval = 0
		cfg.ThresholdSCfg().StringIndexedFields = nil
		cfg.ThresholdSCfg().PrefixIndexedFields = nil
		thServ = NewThresholdService(dmTH, cfg, &FilterS{dm: dmTH, cfg: cfg})
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
		argsGetThresholds := []*utils.CGREvent{
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "TH_1",
						"Weight":    "10.0",
					},
			},
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "TH_2",
						"Weight":    "20.0",
					},
			},
			{
					Tenant: "cgrates.org",
					ID:     "Ev1",
					Event: map[string]any{
						"Threshold": "ThresholdPrefix123",
				},
			},
		}

		cfg := config.NewDefaultCGRConfig()
		data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
		dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
		cfg.ThresholdSCfg().StoreInterval = 0
		cfg.ThresholdSCfg().StringIndexedFields = nil
		cfg.ThresholdSCfg().PrefixIndexedFields = nil
		thServ = NewThresholdService(dmTH, cfg, &FilterS{dm: dmTH, cfg: cfg})
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
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
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

func TestThresholdsProcessEventAsyncExecErr(t *testing.T) {
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	tmpLogger := utils.Logger
	defer func() {
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().ActionSConns = []string{"actPrfID"}
	cfg.RPCConns()["actPrfID"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	connMgr := NewConnManager(cfg)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
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
		Event: map[string]any{
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
// 			Event: map[string]any{
// 				utils.AccountField: "1001",
// 			},
// 			APIOpts: map[string]any{
// 				utils.MetaEventType: utils.AccountUpdate,
// 			},
// 		},
// 	}
// 	if err := processEventWithThreshold(args, dm); err != nil {
// 		t.Error(err)
// 	}
// }

func TestThresholdsStoreThresholdsOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
	tmpLogger := utils.Logger
	defer func() {
		Cache = tmp
		utils.Logger = tmpLogger
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

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

	Cache.Remove(context.Background(), utils.CacheThresholds, "TH1", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdsCacheGetErr(t *testing.T) {
	tmp := Cache
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
		Cache = tmp
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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

	Cache.Remove(context.Background(), utils.CacheThresholds, "TH2", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdNilDirtyField(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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

func TestThresholdsRPCClone(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventTest",
		Event:  make(map[string]any),
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"THD_ID"},
		},
	}
	args.SetCloneable(true)

	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventTest",
		Event:  make(map[string]any),
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"THD_ID"},
		},
	}

	if out, err := args.RPCClone(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, exp) {
		t.Errorf("expected: <%T>, \nreceived: <%T>",
			args, exp)
	}
}

func TestThresholdsProcessEventStoreThOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH2"},
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
		rcv.weight = 0
		rcv.dirty = nil
		rcv.Snooze = time.Time{}
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
		}
	}

}

func TestThresholdsProcessEventMaxHitsDMErr(t *testing.T) {
	tmpLogger := utils.Logger
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)
	tmp := config.CgrConfig()
	tmpC := Cache
	tmpCMgr := connMgr

	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMgr = NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), connMgr)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(nil, cfg, filterS, connMgr)
	Cache = NewCacheS(cfg, dm, nil, nil)

	defer func() {
		connMgr = tmpCMgr
		Cache = tmpC
		config.SetCgrConfig(tmp)
		utils.Logger = tmpLogger
	}()
	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
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
	Cache.SetWithoutReplicate(utils.CacheThresholdProfiles, thPrf.TenantID(), thPrf, nil, true, utils.NonTransactional)
	Cache.SetWithoutReplicate(utils.CacheThresholds, th.TenantID(), th, nil, true, utils.NonTransactional)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH3"},
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
}

func TestThresholdsProcessEventNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH5",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
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

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH6"},
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "TH1",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{utils.MetaNone},
		MinHits:          2,
		MaxHits:          5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		Blocker: true,
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
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blocker: false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
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
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
		Cache = tmp
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
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
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blocker: false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
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
}

func TestThresholdsV1ProcessEventMissingArgs(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
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
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blocker: false,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	args = &utils.CGREvent{
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var reply []string
	experr := `MANDATORY_IE_MISSING: [ID]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = nil
	experr = `MANDATORY_IE_MISSING: [CGREvent]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		ID:    "V1ProcessEventTest",
		Event: nil,
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
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
	if err := tS.V1GetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}

	var rplyTh Threshold
	if err := tS.V1GetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH2",
		},
	}, &rplyTh); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdMatchingThresholdForEventLocks(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TH%d", i),
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				},
			},
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	dm.RemoveThreshold(context.Background(), "cgrates.org", "TH1")
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	mth, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
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
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TH%d", i),
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				},
			},
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
		Weights: utils.DynamicWeights{
			{
				Weight: 20.00,
			},
		},
		MaxHits: 5,
	}
	err := db.SetThresholdProfileDrv(context.Background(), rPrf)
	if err != nil {
		t.Fatal(err)
	}
	prfs = append(prfs, rPrf)
	ids.Add(rPrf.ID)
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	_, err = rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
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
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TH%d", i),
			Weights: utils.DynamicWeights{
				{
					Weight: float64(10 - i),
				},
			},
			Blocker: i == 4,
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	mres, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
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
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tnt, id string) (*ThresholdProfile, error) {
			if id == "TH1" {
				return nil, utils.ErrNotImplemented
			}
			rPrf := &ThresholdProfile{
				Tenant:  "cgrates.org",
				ID:      id,
				MaxHits: 5,
				Weights: utils.DynamicWeights{
					{
						Weight: 20.00,
					},
				},
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
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	_, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
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
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TH%d", i),
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				},
			},
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	ids.Add("TH20")
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	mres, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
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
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), NewConnManager(cfg))
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
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TH%d", i),
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				},
			},
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	dm.RemoveThreshold(context.Background(), "cgrates.org", "TH1")
	_, err := rS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
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
	tS := &ThresholdS{
		cfg:         cfg,
		loopStopped: make(chan struct{}, 1),
	}
	tS.runBackup(context.Background())
	select {
	case <-tS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestThresholdsRunBackupStop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tnt := "cgrates.org"
	thID := "Th1"
	tS := &ThresholdS{
		dm: dm,
		storedTdIDs: utils.StringSet{
			thID: struct{}{},
		},
		cfg:         cfg,
		loopStopped: make(chan struct{}, 1),
		stopBackup:  make(chan struct{}),
	}
	value := &Threshold{
		dirty:  utils.BoolPointer(true),
		Tenant: tnt,
		ID:     thID,
	}
	Cache.SetWithoutReplicate(utils.CacheThresholds, thID, value, nil, true, "")

	// Backup loop checks for the state of the stopBackup
	// channel after storing the threshold. Channel can be
	// safely closed beforehand.
	close(tS.stopBackup)
	tS.runBackup(context.Background())

	want := &Threshold{
		dirty:  utils.BoolPointer(false),
		Tenant: tnt,
		ID:     thID,
	}
	if got, err := tS.dm.GetThreshold(context.Background(), tnt, thID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("dm.GetThreshold(%q,%q): got unexpected err=%v", tnt, thID, err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("dm.GetThreshold(%q,%q) = %v, want %v", tnt, thID, got, want)
	}

	select {
	case <-tS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestThresholdsReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	tS := &ThresholdS{
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cfg:         cfg,
	}
	tS.loopStopped <- struct{}{}
	tS.Reload(context.Background())
	close(tS.stopBackup)
	select {
	case <-tS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestThresholdsStartLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	tS := &ThresholdS{
		loopStopped: make(chan struct{}, 1),
		cfg:         cfg,
	}
	tS.StartLoop(context.Background())
	select {
	case <-tS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	args := &utils.CGREvent{
		ID: "TestGetThresholdsForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	exp := Thresholds{
		{
			Tenant: "cgrates.org",
			Hits:   0,
			ID:     "TH1",
			tPrfl:  thPrf,
			dirty:  utils.BoolPointer(false),
			weight: 10,
		},
	}
	var reply Thresholds
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp[0], reply[0])
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	var args *utils.CGREvent

	experr := `MANDATORY_IE_MISSING: [CGREvent]`
	var reply Thresholds
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestGetThresholdsForEvent",
		Event:  nil,
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "TH2",
		MinHits: 0,
		MaxHits: 7,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf2, true); err != nil {
		t.Error(err)
	}

	expIDs := []string{"TH1", "TH2"}
	var reply []string
	if err := tS.V1GetThresholdIDs(context.Background(), &utils.TenantWithAPIOpts{}, &reply); err != nil {
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
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	var reply []string
	if err := tS.V1GetThresholdIDs(context.Background(), &utils.TenantWithAPIOpts{}, &reply); err == nil ||
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
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
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
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
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
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
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(nil, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
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
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
	}, &reply); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestThresholdsLockUnlockThresholdProfiles(t *testing.T) {
	thPrf := &ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
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
		ID:     "TH1",
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event:  map[string]any{},
	}

	if _, err := tS.matchingThresholdsForEvent(context.Background(), "cgrates.org", args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsStoreThresholdCacheSetErr(t *testing.T) {
	tmpLogger := utils.Logger
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
		utils.Logger = tmpLogger
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr = NewConnManager(cfg)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, connMgr)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		dirty:  utils.BoolPointer(true),
	}
	Cache.SetWithoutReplicate(utils.CacheThresholds, th.TenantID(), th, nil, true, utils.NonTransactional)
	expLog := `[WARNING] <ThresholdService> failed caching Threshold with ID: cgrates.org:TH1, error: DISCONNECTED`
	if err := tS.StoreThreshold(context.Background(), th); err == nil ||
		err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrDisconnected, err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcv)
	}
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH1"},
		},
	}
	if _, err := tS.matchingThresholdsForEvent(context.Background(), "cgrates.org",
		args); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsProcessEventIgnoreFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)
	cfg.ThresholdSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		{
			Value: true,
		},
	}
	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH",
		FilterIDs: []string{"*string:~*req.Threshold:testThresholdValue"},
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH",
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	// testing if the profile matches wtih profile ignore filters on false
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			"Threshold": "testThresholdValue",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH"},
			utils.MetaProfileIgnoreFilters: false,
		},
	}
	exp := []string{"TH"}
	if rcv, err := tS.processEvent(context.Background(), args.Tenant, args); err != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
	// testing if the profile matches with wtih profile ignore filters on true
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			"Threshold": "testThresholdValue2",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	exp2 := []string{"TH"}
	if rcv2, err := tS.processEvent(context.Background(), args2.Tenant, args2); err != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv2, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp2, rcv2)
	}
}

func TestThresholdsProcessEventIgnoreFiltersErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)
	cfg.ThresholdSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		{
			Value: true,
		},
	}
	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH",
		FilterIDs: []string{"*string:~*req.Threshold:testThresholdValue"},
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH",
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			"Threshold": "testThresholdValue",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH"},
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}

	if _, err := tS.processEvent(context.Background(), args.Tenant, args); err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, err)
	}

}

func TestThresholdProfileSet(t *testing.T) {
	th := ThresholdProfile{}
	exp := ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		MaxHits:          10,
		MinHits:          10,
		MinSleep:         10,
		Blocker:          true,
		Async:            true,
		ActionProfileIDs: []string{"acc1"},
	}
	if err := th.Set([]string{}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := th.Set([]string{"NotAField"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := th.Set([]string{"NotAField", "1"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := th.Set([]string{utils.Tenant}, "cgrates.org", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{utils.ID}, "ID", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{utils.FilterIDs}, "fltr1;*string:~*req.Account:1001", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{utils.Weights}, ";10", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{utils.MaxHits}, 10, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{utils.MinHits}, 10, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{utils.MinSleep}, 10, false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if err := th.Set([]string{utils.Blocker}, true, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{utils.Async}, true, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{utils.ActionProfileIDs}, "acc1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, th) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(th))
	}
}

func TestThresholdProfileAsInterface(t *testing.T) {
	tp := ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		MaxHits:          10,
		MinHits:          10,
		MinSleep:         10,
		Blocker:          true,
		Async:            true,
		ActionProfileIDs: []string{"acc1"},
	}
	if _, err := tp.FieldAsInterface(nil); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := tp.FieldAsInterface([]string{"field"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := tp.FieldAsInterface([]string{"field", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := tp.FieldAsInterface([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := tp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := tp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.Weights}); err != nil {
		t.Fatal(err)
	} else if exp := tp.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.ActionProfileIDs}); err != nil {
		t.Fatal(err)
	} else if exp := tp.ActionProfileIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.ActionProfileIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := tp.ActionProfileIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if val, err := tp.FieldAsInterface([]string{utils.MaxHits}); err != nil {
		t.Fatal(err)
	} else if exp := tp.MaxHits; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.MinHits}); err != nil {
		t.Fatal(err)
	} else if exp := tp.MinHits; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.MinSleep}); err != nil {
		t.Fatal(err)
	} else if exp := tp.MinSleep; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.Blocker}); err != nil {
		t.Fatal(err)
	} else if exp := tp.Blocker; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{utils.Async}); err != nil {
		t.Fatal(err)
	} else if exp := tp.Async; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := tp.FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := tp.FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := tp.String(), utils.ToJSON(tp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestThresholdProfileMerge(t *testing.T) {
	dp := &ThresholdProfile{}
	exp := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		MaxHits:          10,
		MinHits:          10,
		MinSleep:         10,
		Blocker:          true,
		Async:            true,
		ActionProfileIDs: []string{"acc1"},
	}
	if dp.Merge(&ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		MaxHits:          10,
		MinHits:          10,
		MinSleep:         10,
		Blocker:          true,
		Async:            true,
		ActionProfileIDs: []string{"acc1"},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}

func TestThresholdSmatchingThresholdsForEventGetOptsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().Opts.ProfileIDs = []*config.DynamicStringSliceOpt{
		{
			FilterIDs: []string{"*string"},
			Tenant:    "cgrates.org",
			Values:    []string{"ProfIdVal"},
		},
	}

	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), cM)
	filterS := NewFilterS(cfg, cM, dm)
	tS := &ThresholdS{
		dm: dm,
		storedTdIDs: utils.StringSet{
			"Th1": struct{}{},
		},
		cfg:         cfg,
		fltrS:       filterS,
		loopStopped: make(chan struct{}, 1),
		stopBackup:  make(chan struct{}),
	}

	args := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "3m",
		},
	}

	expErr := "inline parse error for string: <*string>"
	if _, err := tS.matchingThresholdsForEvent(context.Background(), args.Tenant, args); err.Error() != expErr || err == nil {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}

}

func TestThresholdSmatchingThresholdsForEventWeightErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().Opts.ProfileIDs = []*config.DynamicStringSliceOpt{
		{
			Tenant: "cgrates.org",
			Values: []string{"ACC1"},
		},
	}
	cfg.ThresholdSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		{
			Tenant: "cgrates.org",
			Value:  true,
		},
	}
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return &ThresholdProfile{
				Tenant:           "cgrates.org",
				ID:               "THD_2",
				FilterIDs:        []string{"*string:~*req.Account:1001"},
				ActionProfileIDs: []string{"actPrfID"},
				MaxHits:          7,
				MinHits:          0,
				Weights: []*utils.DynamicWeight{
					{
						FilterIDs: []string{"*stirng:~*req.Account:1001"},
						Weight:    64,
					},
				},
				Async: true,
			}, nil
		},
		GetThresholdDrvF: func(ctx *context.Context, tenant, id string) (*Threshold, error) { return &Threshold{}, nil },
	}
	filterS := NewFilterS(cfg, cM, dm)
	tS := &ThresholdS{
		dm: dm,
		storedTdIDs: utils.StringSet{
			"Th1": struct{}{},
		},
		cfg:         cfg,
		fltrS:       filterS,
		loopStopped: make(chan struct{}, 1),
		stopBackup:  make(chan struct{}),
	}

	args := &utils.CGREvent{
		ID:     "EvID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "3m",
		},
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	if _, err := tS.matchingThresholdsForEvent(context.Background(), args.Tenant, args); err.Error() != expErr || err == nil {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestThresholdsV1ResetThresholdStoreErr(t *testing.T) {

	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM

	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr = NewConnManager(cfg)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, connMgr)
	th := &Threshold{
		Hits:   2,
		Tenant: "cgrates.org",
		ID:     "TH1",
		dirty:  utils.BoolPointer(true),
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	Cache.SetWithoutReplicate(utils.CacheThresholds, th.TenantID(), th, nil, true, utils.NonTransactional)

	var reply string
	if err := tS.V1ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TH1",
		},
	}, &reply); err == nil || err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("Expected error <%+v>, Received error <%+v>", utils.ErrDisconnected, err)
	}
}
