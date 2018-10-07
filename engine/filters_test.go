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
	if passes, err := rf.passString(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaString, FieldName: "Category", Values: []string{"cal"}}
	if passes, err := rf.passString(cd); err != nil {
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
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "Category", Values: []string{"premium"}}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "Destination", Values: []string{"+49"}}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "Destination", Values: []string{"+499"}}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "navigation", Values: []string{"off"}}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaPrefix, FieldName: "nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
}

func TestFilterPassRSRFields(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call",
		Tenant: "cgrates.org", Subject: "dan", Destination: "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"}}
	rf, err := NewFilterRule(MetaRSR, "", []string{"~Tenant(~^cgr.*\\.org$)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(MetaRSR, "", []string{"~navigation(on)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	rf, err = NewFilterRule(MetaRSR, "", []string{"~navigation(off)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd); err != nil {
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
	if passes, err := rf.passDestinations(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(MetaDestinations, "Destination", []string{"RO"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd); err != nil {
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
	ev := config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 20, true)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 40, true)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(MetaLessOrEqual, "ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(MetaGreaterOrEqual, "ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 20, true)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("should not pass")
	}
	rf, err = NewFilterRule(MetaGreaterOrEqual, "ACD", []string{"1m50s"})
	if err != nil {
		t.Error(err)
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ACD"}, time.Duration(2*time.Minute), true)
	if passes, err := rf.passGreaterThan(ev); err != nil {
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
	if _, err := filterS.Pass("cgrates.org",
		[]string{"*string:Account:1007:error"}, nil); err == nil {
		t.Errorf(err.Error())
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:Account:1007"}, config.NewNavigableMap(failEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:Account:1007"}, config.NewNavigableMap(passEvent)); err != nil {
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
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*prefix:Account:10"}, config.NewNavigableMap(failEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*prefix:Account:10"}, config.NewNavigableMap(passEvent)); err != nil {
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
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Tenant(~^cgr.*\\.org$)"}, config.NewNavigableMap(failEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Tenant(~^cgr.*\\.org$)"}, config.NewNavigableMap(passEvent)); err != nil {
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
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:Destination:EU"}, config.NewNavigableMap(failEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:Destination:EU_LANDLINE"}, config.NewNavigableMap(passEvent)); err != nil {
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
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gte:Weight:20"}, config.NewNavigableMap(failEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gte:Weight:10"}, config.NewNavigableMap(passEvent)); err != nil {
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
	if pass, err := filterS.Pass("cgrates.org",
		[]string{}, config.NewNavigableMap(passEvent1)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("itsyscom.com",
		[]string{}, config.NewNavigableMap(passEvent2)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	ev := map[string]interface{}{
		"Test": "MultipleCharacter",
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Test(~^\\w{30,})"}, config.NewNavigableMap(ev)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]interface{}{
		"Test": "MultipleCharacter123456789MoreThan30Character",
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Test(~^\\w{30,})"}, config.NewNavigableMap(ev)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	ev = map[string]interface{}{
		"Test": map[string]interface{}{
			"Test2": "MultipleCharacter",
		},
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Test.Test2(~^\\w{30,})"}, config.NewNavigableMap(ev)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]interface{}{
		"Test": map[string]interface{}{
			"Test2": "MultipleCharacter123456789MoreThan30Character",
		},
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Test.Test2(~^\\w{30,})"}, config.NewNavigableMap(ev)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}

	ev = map[string]interface{}{
		utils.Account:     "1003",
		utils.Subject:     "1003",
		utils.Destination: "1002",
		utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
		utils.Usage:       "1m20s",
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:Account:1003", "*prefix:Destination:10", "*rsr::~Destination(1002)"}, config.NewNavigableMap(ev)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}
