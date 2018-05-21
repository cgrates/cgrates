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
	thServ ThresholdService
	dmTH   *DataManager
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
	var filters1 []*FilterRule
	var filters2 []*FilterRule
	var preffilter []*FilterRule
	var defaultf []*FilterRule
	second := 1 * time.Second
	thServ = ThresholdService{
		dm:      dmTH,
		filterS: &FilterS{dm: dmTH},
	}
	ref := NewFilterIndexer(dmTH, utils.ThresholdProfilePrefix, "cgrates.org")

	//filter1
	x, err := NewFilterRule(MetaString, "Threshold", []string{"ThresholdProfile1"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "UsageInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"9.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	filter5 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter5", Rules: filters1}
	dmTH.SetFilter(filter5)
	ref.IndexTPFilter(FilterToTPFilter(filter5), "TEST_PROFILE1")

	//filter2
	x, err = NewFilterRule(MetaString, "Threshold", []string{"ThresholdProfile2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "PddInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"15.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	filter6 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter6", Rules: filters2}
	dmTH.SetFilter(filter6)
	ref.IndexTPFilter(FilterToTPFilter(filter6), "TEST_PROFILE2")
	//prefix filter
	x, err = NewFilterRule(MetaPrefix, "Threshold", []string{"ThresholdProfilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	preffilter = append(preffilter, x)
	preffilter3 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "preffilter3", Rules: preffilter}
	dmTH.SetFilter(preffilter3)
	ref.IndexTPFilter(FilterToTPFilter(preffilter3), "TEST_PROFILE3")
	//default filter
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultf = append(defaultf, x)
	defaultf3 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "defaultf3", Rules: defaultf}
	dmTH.SetFilter(defaultf3)
	ref.IndexTPFilter(FilterToTPFilter(defaultf3), "TEST_PROFILE4")

	tPrfl := []*ThresholdProfile{
		&ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"filter5"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
		&ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE2",
			FilterIDs: []string{"filter6"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
		&ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE3",
			FilterIDs: []string{"preffilter3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
		&ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE4",
			FilterIDs: []string{"defaultf3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
	}
	for _, profile := range tPrfl {
		if err = dmTH.SetThresholdProfile(profile, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
		if err = dmTH.SetThreshold(&Threshold{Tenant: profile.Tenant, ID: profile.ID}); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	err = ref.StoreIndexes(true, utils.NonTransactional)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

// func TestThresholdsmatchingThresholdsForEvent(t *testing.T) {
// 	stringEv := utils.CGREvent{
// 		Tenant: "cgrates.org",
// 		ID:     "utils.CGREvent1",
// 		Event: map[string]interface{}{
// 			"Threshold":      "ThresholdProfile1",
// 			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
// 			"UsageInterval":  "1s",
// 			"PddInterval":    "1s",
// 			"Weight":         "20.0",
// 		},
// 	}
// 	argEv := &ArgsProcessEvent{CGREvent: stringEv}
// 	sprf, err := thServ.matchingThresholdsForEvent(argEv)
// 	if err != nil {
// 		t.Errorf("Error: %+v", err)
// 	}
// 	if !reflect.DeepEqual(0, sprf) {
// 		t.Errorf("Expecting: %+v, received: %+v", 0, sprf)
// 	} //should not pass atm but still return something other than empty string
// }

// func TestThresholdsprocessEvent(t *testing.T) {

// }
