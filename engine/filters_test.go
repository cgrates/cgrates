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
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
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
	//not
	rf = &FilterRule{Type: "*notstring", FieldName: "Category", Values: []string{"call"}}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: "*notstring", FieldName: "Category", Values: []string{"cal"}}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
}

func TestFilterPassEmpty(t *testing.T) {
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf := &FilterRule{Type: MetaEmpty, FieldName: "Category", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaEmpty, FieldName: "Direction", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: MetaEmpty, FieldName: "ExtraFields", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	cd.ExtraFields = map[string]string{}
	rf = &FilterRule{Type: MetaEmpty, FieldName: "ExtraFields", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	//not
	rf = &FilterRule{Type: "*notempty", FieldName: "Direction", Values: []string{}}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: "*notempty", FieldName: "Category", Values: []string{}}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
}

func TestFilterPassExists(t *testing.T) {
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf := &FilterRule{Type: MetaExists, FieldName: "Category", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaExists, FieldName: "Direction", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaExists, FieldName: "ExtraFields1", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	cd.ExtraFields = map[string]string{}
	rf = &FilterRule{Type: MetaExists, FieldName: "ExtraFields", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	//not
	rf = &FilterRule{Type: "*notexists", FieldName: "Direction", Values: []string{}}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: "*notexists", FieldName: "Category1", Values: []string{}}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
}

func TestFilterPassStringPrefix(t *testing.T) {
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
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
	//not
	rf = &FilterRule{Type: "*notprefix", FieldName: "Category", Values: []string{"premium"}}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
}

func TestFilterPassStringSuffix(t *testing.T) {
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf := &FilterRule{Type: MetaSuffix, FieldName: "Category", Values: []string{"call"}}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaSuffix, FieldName: "Category", Values: []string{"premium"}}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: MetaSuffix, FieldName: "Destination", Values: []string{"963"}}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaSuffix, FieldName: "Destination", Values: []string{"4966"}}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: MetaSuffix, FieldName: "navigation", Values: []string{"off"}}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: MetaSuffix, FieldName: "nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
	//not
	rf = &FilterRule{Type: "*notsuffix", FieldName: "Destination", Values: []string{"963"}}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
}

func TestFilterPassRSRFields(t *testing.T) {
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
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
	//not
	rf, err = NewFilterRule("*notrsr", "", []string{"~navigation(off)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd, nil); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
}

func TestFilterPassDestinations(t *testing.T) {
	Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, "")
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
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
	//not
	rf, err = NewFilterRule("*notdestinations", "Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd, nil); err != nil {
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
	ev.Set([]string{"ASR"}, 20, false, true)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 40, false, true)
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
	ev.Set([]string{"ASR"}, 20, false, true)
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
	ev.Set([]string{"ACD"}, time.Duration(2*time.Minute), false, true)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not pass")
	}
	//not
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 20, false, true)
	rf, err = NewFilterRule("*notgt", "ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("passing")
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
	rf, err = NewFilterRule(MetaEmpty, "MetaEmpty", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaEmpty, FieldName: "MetaEmpty", Values: []string{}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(MetaExists, "MetaExists", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaExists, FieldName: "MetaExists", Values: []string{}}
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
	rf, err = NewFilterRule(MetaSuffix, "MetaSuffix", []string{"stringSuffix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: MetaSuffix, FieldName: "MetaSuffix", Values: []string{"stringSuffix"}}
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
	//not
	rf, err = NewFilterRule("*notgt", "NotMetaGreaterOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: "*gt", FieldName: "NotMetaGreaterOrEqual", Values: []string{"20"}, negative: true}
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
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notstring:Account:1007"}, config.NewNavigableMap(passEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
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
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*suffix:Account:07"}, config.NewNavigableMap(failEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*suffix:Account:07"}, config.NewNavigableMap(passEvent)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notsuffix:Account:07"}, config.NewNavigableMap(passEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
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
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notrsr::~Tenant(~^cgr.*\\.org$)"}, config.NewNavigableMap(passEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
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

	failEvent = map[string]interface{}{
		"EmptyString":   "nonEmpty",
		"EmptySlice":    []string{""},
		"EmptyMap":      map[string]string{"": ""},
		"EmptyPtr":      &struct{}{},
		"EmptyPtr2":     &struct{}{},
		"EmptyPtrSlice": &[]string{""},
		"EmptyPtrMap":   &map[string]string{"": ""},
	}
	var testnil *struct{}
	testnil = nil
	passEvent = map[string]interface{}{
		"EmptyString":   "",
		"EmptySlice":    []string{},
		"EmptyMap":      map[string]string{},
		"EmptyPtr":      testnil,
		"EmptyPtr2":     nil,
		"EmptyPtrSlice": &[]string{},
		"EmptyPtrMap":   &map[string]string{},
	}
	for key := range failEvent {
		if pass, err := filterS.Pass("cgrates.org", []string{"*empty:" + key + ":"},
			config.NewNavigableMap(failEvent)); err != nil {
			t.Errorf(err.Error())
		} else if pass {
			t.Errorf("For %s expecting: %+v, received: %+v", key, false, pass)
		}
		if pass, err := filterS.Pass("cgrates.org", []string{"*empty:" + key + ":"},
			config.NewNavigableMap(passEvent)); err != nil {
			t.Errorf(err.Error())
		} else if !pass {
			t.Errorf("For %s expecting: %+v, received: %+v", key, true, pass)
		}
	}
	if pass, err := filterS.Pass("cgrates.org", []string{"*exists:" + "NewKey" + ":"},
		config.NewNavigableMap(failEvent)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("For NewKey expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org", []string{"*notexists:" + "NewKey" + ":"},
		config.NewNavigableMap(failEvent)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("For NewKey expecting: %+v, received: %+v", true, pass)
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
		[]string{"*string:Account:1003", "*prefix:Destination:10", "*suffix:Subject:03", "*rsr::~Destination(1002)"},
		config.NewNavigableMap(ev)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassFilterMaxCost(t *testing.T) {
	data, _ := NewMapStorage()
	dmFilterPass := NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	//check with max usage -1 should fail
	passEvent1 := map[string]interface{}{
		"MaxUsage": time.Duration(-1),
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~MaxUsage{*duration_nanoseconds}(>0)"}, config.NewNavigableMap(passEvent1)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
	//check with max usage 0 should fail
	passEvent2 := map[string]interface{}{
		"MaxUsage": time.Duration(0),
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~MaxUsage{*duration_nanoseconds}(>0)"}, config.NewNavigableMap(passEvent2)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false, received: %+v", pass)
	}
	//check with max usage 123 should pass
	passEvent3 := map[string]interface{}{
		"MaxUsage": time.Duration(123),
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~MaxUsage{*duration_nanoseconds}(>0)"}, config.NewNavigableMap(passEvent3)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true, received: %+v", pass)
	}
}

func TestPassFilterMissingField(t *testing.T) {
	data, _ := NewMapStorage()
	dmFilterPass := NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}

	passEvent1 := map[string]interface{}{
		"test": "call",
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Category(^$)"}, config.NewNavigableMap(passEvent1)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent2 := map[string]interface{}{
		"Category": "",
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Category(^$)"}, config.NewNavigableMap(passEvent2)); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent3 := map[string]interface{}{
		"Category": "call",
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Category(^$)"}, config.NewNavigableMap(passEvent3)); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
}
