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

func TestFilterPassString(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf := &FilterRule{Type: MetaString, FieldName: "Category", Values: []string{"call"}}
	if passes, err := rf.passString(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaString, FieldName: "Category", Values: []string{"cal"}}
	if passes, err := rf.passString(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
}

func TestFilterPassStringPrefix(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf := &FilterRule{Type: MetaPrefix, FieldName: "Category", Values: []string{"call"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "Category", Values: []string{"premium"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "Destination", Values: []string{"+49"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "Destination", Values: []string{"+499"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "navigation", Values: []string{"off"}}
	if passes, err := rf.passStringPrefix(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringPrefix(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
}

func TestFilterPassRSRFields(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf, err := NewFilterRule(MetaRSR, "", []string{"Tenant(~^cgr.*\\.org$)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(MetaRSR, "", []string{"navigation(on)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	rf, err = NewFilterRule(MetaRSR, "", []string{"navigation(off)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
}

func TestFilterPassDestinations(t *testing.T) {
	Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, "")
	cd := &CallDescriptor{Direction: "*out",
		Category: "call", Tenant: "cgrates.org",
		Subject: "dan", Destination: "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"}}
	rf, err := NewFilterRule(MetaDestinations, "Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(MetaDestinations, "Destination", []string{"RO"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
}

func TestFilterPassGreaterThan(t *testing.T) {
	rf, err := NewFilterRule(MetaLessThan, "ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	ev := map[string]interface{}{
		"ASR": 20,
	}
	if passes, err := rf.passGreaterThan(ev, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = map[string]interface{}{
		"ASR": 40,
	}
	if passes, err := rf.passGreaterThan(ev, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(MetaLessOrEqual, "ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(MetaGreaterOrEqual, "ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = map[string]interface{}{
		"ASR": 20,
	}
	if passes, err := rf.passGreaterThan(ev, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("should not pass")
	}
	rf, err = NewFilterRule(MetaGreaterOrEqual, "ACD", []string{"1m50s"})
	if err != nil {
		t.Error(err)
	}
	ev = map[string]interface{}{
		"ACD": time.Duration(2 * time.Minute),
	}
	if passes, err := rf.passGreaterThan(ev, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not pass")
	}
}

func TestFilterNewRequestFilter(t *testing.T) {
	rf, err := NewFilterRule(MetaString, "MetaString", []string{"String"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf := &FilterRule{Type: MetaString, FieldName: "MetaString", Values: []string{"String"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(MetaPrefix, "MetaPrefix", []string{"stringPrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaPrefix, FieldName: "MetaPrefix", Values: []string{"stringPrefix"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(MetaTimings, "MetaTimings", []string{""})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaTimings, FieldName: "MetaTimings", Values: []string{""}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(MetaDestinations, "MetaDestinations", []string{""})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaDestinations, FieldName: "MetaDestinations", Values: []string{""}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(MetaLessThan, "MetaLessThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaLessThan, FieldName: "MetaLessThan", Values: []string{"20"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(MetaLessOrEqual, "MetaLessOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaLessOrEqual, FieldName: "MetaLessOrEqual", Values: []string{"20"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(MetaGreaterThan, "MetaGreaterThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaGreaterThan, FieldName: "MetaGreaterThan", Values: []string{"20"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(MetaGreaterOrEqual, "MetaGreaterOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaGreaterOrEqual, FieldName: "MetaGreaterOrEqual", Values: []string{"20"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
}

func TestInlineFilterPassFiltersForEvent(t *testing.T) {
	data, _ := NewMapStorage()
	dmFilterPass := NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	failEvent := map[string]interface{}{
		"Account": "1001",
	}
	passEvent := map[string]interface{}{
		"Account": "1007",
	}
	if _, err := filterS.PassFiltersForEvent("cgrates.org",
		nil, []string{"*string:Account:1007:error"}); err == nil {
		t.Errorf(err.Error())
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		failEvent, []string{"*string:Account:1007"}); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		passEvent, []string{"*string:Account:1007"}); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	failEvent = map[string]interface{}{
		"Account": "2001",
	}
	passEvent = map[string]interface{}{
		"Account": "1007",
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		failEvent, []string{"*prefix:Account:10"}); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		passEvent, []string{"*prefix:Account:10"}); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	failEvent = map[string]interface{}{
		"Tenant": "anotherTenant.org",
	}
	passEvent = map[string]interface{}{
		"Tenant": "cgrates.org",
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		failEvent, []string{"*rsr::Tenant(~^cgr.*\\.org$)"}); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		passEvent, []string{"*rsr::Tenant(~^cgr.*\\.org$)"}); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, "")
	failEvent = map[string]interface{}{
		utils.Destination: "+5086517174963",
	}
	passEvent = map[string]interface{}{
		utils.Destination: "+4986517174963",
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		failEvent, []string{"*destinations:Destination:EU"}); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		passEvent, []string{"*destinations:Destination:EU_LANDLINE"}); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	failEvent = map[string]interface{}{
		utils.Weight: 10,
	}
	passEvent = map[string]interface{}{
		utils.Weight: 20,
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		failEvent, []string{"*gte:Weight:20"}); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		passEvent, []string{"*gte:Weight:10"}); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassFiltersForEventWithEmptyFilter(t *testing.T) {
	data, _ := NewMapStorage()
	dmFilterPass := NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent1 := map[string]interface{}{
		utils.Tenant:      "cgrates.org",
		utils.Account:     "1010",
		utils.Destination: "+49",
		utils.Weight:      10,
	}
	passEvent2 := map[string]interface{}{
		utils.Tenant:      "itsyscom.com",
		utils.Account:     "dan",
		utils.Destination: "+4986517174963",
		utils.Weight:      20,
	}
	if pass, err := filterS.PassFiltersForEvent("cgrates.org",
		passEvent1, []string{}); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.PassFiltersForEvent("itsyscom.com",
		passEvent2, []string{}); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}
