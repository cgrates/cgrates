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

var (
	dmTH   *DataManager
	thServ *ThresholdService
	tPrfls = []*ThresholdProfile{
		&ThresholdProfile{
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
		&ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		&ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
	}
	ths = Thresholds{
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
	argsGetThresholds = []*ArgsProcessEvent{
		&ArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_1",
					"Weight":    "10.0",
				},
			},
		},
		&ArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "TH_2",
					"Weight":    "20.0",
				},
			},
		},
		&ArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]interface{}{
					"Threshold": "ThresholdPrefix123",
				},
			},
		},
	}
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

func TestThresholdsPopulateThresholdService(t *testing.T) {
	data, _ := NewMapStorage()
	dmTH = NewDataManager(data)
	defaultCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	thServ, err = NewThresholdService(dmTH, nil, nil, 0,
		&FilterS{dm: dmTH, cfg: defaultCfg})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestThresholdsAddFilters(t *testing.T) {
	fltrTh1 := &Filter{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*FilterRule{
			&FilterRule{
				Type:      MetaString,
				FieldName: "Threshold",
				Values:    []string{"TH_1"},
			},
			&FilterRule{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.Weight,
				Values:    []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh1)
	fltrTh2 := &Filter{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			&FilterRule{
				Type:      MetaString,
				FieldName: "Threshold",
				Values:    []string{"TH_2"},
			},
			&FilterRule{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.Weight,
				Values:    []string{"15.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh2)
	fltrTh3 := &Filter{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			&FilterRule{
				Type:      MetaPrefix,
				FieldName: "Threshold",
				Values:    []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(fltrTh3)
}

func TestThresholdsCache(t *testing.T) {
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
	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[0].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Hits, thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[1].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Hits, thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[2]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[2].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[2].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Hits, thMatched[0].Hits)
	}
}

func TestThresholdsprocessEvent(t *testing.T) {
	thIDs := []string{"TH_1"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[0]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_2"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[1]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_3"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[2]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}
}

func TestThresholdsVerifyIfExecuted(t *testing.T) {
	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[2]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[2].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}
}
