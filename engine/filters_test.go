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
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf := &FilterRule{Type: utils.MetaString,
		FieldName: "~Category", Values: []string{"call"}}
	if passes, err := rf.passString(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaString,
		FieldName: "~Category", Values: []string{"cal"}}
	if passes, err := rf.passString(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotString,
		FieldName: "~Category", Values: []string{"call"}}
	if passes, err := rf.Pass(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: utils.MetaNotString,
		FieldName: "~Category", Values: []string{"cal"}}
	if passes, err := rf.Pass(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
}

func TestFilterPassEmpty(t *testing.T) {
	cd := &CallDescriptor{
		Category:      "",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf := &FilterRule{Type: utils.MetaEmpty, FieldName: "~Category", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaEmpty, FieldName: "~ExtraFields", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	cd.ExtraFields = map[string]string{}
	rf = &FilterRule{Type: utils.MetaEmpty, FieldName: "~ExtraFields", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotEmpty, FieldName: "~Category", Values: []string{}}
	if passes, err := rf.Pass(cd, []config.DataProvider{}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
}

func TestFilterPassExists(t *testing.T) {
	cd := &CallDescriptor{
		Category:      "",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf := &FilterRule{Type: utils.MetaExists, FieldName: "~Category", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaExists, FieldName: "~ExtraFields1", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	cd.ExtraFields = map[string]string{}
	rf = &FilterRule{Type: utils.MetaExists, FieldName: "~ExtraFields", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotExists, FieldName: "~Category1", Values: []string{}}
	if passes, err := rf.Pass(cd, []config.DataProvider{}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
}

func TestFilterPassStringPrefix(t *testing.T) {
	cd := &CallDescriptor{
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf := &FilterRule{Type: utils.MetaPrefix, FieldName: "~Category", Values: []string{"call"}}
	if passes, err := rf.passStringPrefix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, FieldName: "~Category", Values: []string{"premium"}}
	if passes, err := rf.passStringPrefix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, FieldName: "~Destination", Values: []string{"+49"}}
	if passes, err := rf.passStringPrefix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, FieldName: "~Destination", Values: []string{"+499"}}
	if passes, err := rf.passStringPrefix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, FieldName: "~navigation", Values: []string{"off"}}
	if passes, err := rf.passStringPrefix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, FieldName: "~nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringPrefix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotPrefix, FieldName: "~Category", Values: []string{"premium"}}
	if passes, err := rf.Pass(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
}

func TestFilterPassStringSuffix(t *testing.T) {
	cd := &CallDescriptor{
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf := &FilterRule{Type: utils.MetaSuffix, FieldName: "~Category", Values: []string{"call"}}
	if passes, err := rf.passStringSuffix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, FieldName: "~Category", Values: []string{"premium"}}
	if passes, err := rf.passStringSuffix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, FieldName: "~Destination", Values: []string{"963"}}
	if passes, err := rf.passStringSuffix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, FieldName: "~Destination", Values: []string{"4966"}}
	if passes, err := rf.passStringSuffix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, FieldName: "~navigation", Values: []string{"off"}}
	if passes, err := rf.passStringSuffix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, FieldName: "~nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringSuffix(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotSuffix, FieldName: "~Destination", Values: []string{"963"}}
	if passes, err := rf.Pass(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
}

func TestFilterPassRSRFields(t *testing.T) {
	cd := &CallDescriptor{
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf, err := NewFilterRule(utils.MetaRSR, "", []string{"~Tenant(~^cgr.*\\.org$)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR([]config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(utils.MetaRSR, "", []string{"~navigation(on)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR([]config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	rf, err = NewFilterRule(utils.MetaRSR, "", []string{"~navigation(off)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR([]config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	//not
	rf, err = NewFilterRule(utils.MetaNotRSR, "", []string{"~navigation(off)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
}

func TestFilterPassDestinations(t *testing.T) {
	Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, "")
	cd := &CallDescriptor{
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf, err := NewFilterRule(utils.MetaDestinations, "~Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(utils.MetaDestinations, "~Destination", []string{"RO"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	//not
	rf, err = NewFilterRule(utils.MetaNotDestinations, "~Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd, []config.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
}

func TestFilterPassGreaterThan(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaLessThan, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	ev := config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 20, false, true)
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 40, false, true)
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(utils.MetaLessOrEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ASR", []string{"35.5"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 20, false, true)
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("should not pass")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ACD", []string{"1m50s"})
	if err != nil {
		t.Error(err)
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ACD"}, time.Duration(2*time.Minute), false, true)
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not pass")
	}
	// Second
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, time.Duration(20*time.Second), false, true)
	rf, err = NewFilterRule("*gte", "~ASR", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}

	rf, err = NewFilterRule("*gte", "~ASR", []string{"10"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, float64(20*time.Second), false, true)
	rf, err = NewFilterRule("*gte", "~ASR", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}
	//Here converter will be consider part of path and will get error : NOT_FOUND
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 20, false, true)
	rf, err = NewFilterRule("*gte", "~ASR{*duration_seconds}", []string{"10"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("passing")
	}
	//Here converter will be consider part of path and will get error : NOT_FOUND
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 20, false, true)
	rf, err = NewFilterRule("*gte", "~ASR{*duration_seconds}", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("passing")
	}
}

func TestFilterpassEqualTo(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	ev := config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 40.0, false, true)
	if passes, err := rf.passEqualTo(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, 39, false, true)
	if passes, err := rf.passEqualTo(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(utils.MetaNotEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing", passes)
	}
	ev = config.NewNavigableMap(nil)
	ev.Set([]string{"ASR"}, "string1", false, true)
	rf, err = NewFilterRule(utils.MetaEqual, "~ASR", []string{"string1"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passEqualTo(ev, []config.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
}

func TestFilterNewRequestFilter(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaString, "~MetaString", []string{"String"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf := &FilterRule{Type: utils.MetaString, FieldName: "~MetaString", Values: []string{"String"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaEmpty, "~MetaEmpty", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaEmpty, FieldName: "~MetaEmpty", Values: []string{}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaExists, "~MetaExists", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaExists, FieldName: "~MetaExists", Values: []string{}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaPrefix, "~MetaPrefix", []string{"stringPrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaPrefix, FieldName: "~MetaPrefix", Values: []string{"stringPrefix"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaSuffix, "~MetaSuffix", []string{"stringSuffix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaSuffix, FieldName: "~MetaSuffix", Values: []string{"stringSuffix"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaTimings, "~MetaTimings", []string{""})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaTimings, FieldName: "~MetaTimings", Values: []string{""}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaDestinations, "~MetaDestinations", []string{""})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaDestinations, FieldName: "~MetaDestinations", Values: []string{""}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaLessThan, "~MetaLessThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaLessThan, FieldName: "~MetaLessThan", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaLessOrEqual, "~MetaLessOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaLessOrEqual, FieldName: "~MetaLessOrEqual", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaGreaterThan, "~MetaGreaterThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaGreaterThan, FieldName: "~MetaGreaterThan", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~MetaGreaterOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaGreaterOrEqual, FieldName: "~MetaGreaterOrEqual", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
}

func TestInlineFilterPassFiltersForEvent(t *testing.T) {
	data := NewInternalDB(nil, nil)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil, nil)
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
	fEv := config.NewNavigableMap(nil)
	fEv.Set([]string{utils.MetaReq}, failEvent, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*req.Account:1007:error"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*req.Account:1007"}, fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	pEv := config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*req.Account:1007"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notstring:~*req.Account:1007"}, pEv); err != nil {
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
	fEv = config.NewNavigableMap(nil)
	fEv.Set([]string{utils.MetaReq}, failEvent, false, false)
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*prefix:~*req.Account:10"}, fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*prefix:~*req.Account:10"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*suffix:~*req.Account:07"}, fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*suffix:~*req.Account:07"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notsuffix:~*req.Account:07"}, pEv); err != nil {
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
	fEv = config.NewNavigableMap(nil)
	fEv.Set([]string{utils.MetaReq}, failEvent, false, false)
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Tenant(~^cgr.*\\.org$)"}, fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Tenant(~^cgr.*\\.org$)"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notrsr::~*req.Tenant(~^cgr.*\\.org$)"}, pEv); err != nil {
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
	fEv = config.NewNavigableMap(nil)
	fEv.Set([]string{utils.MetaReq}, failEvent, false, false)
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:~*req.Destination:EU"}, fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:~*req.Destination:EU_LANDLINE"}, pEv); err != nil {
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
	fEv = config.NewNavigableMap(nil)
	fEv.Set([]string{utils.MetaReq}, failEvent, false, false)
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gte:~*req.Weight:20"}, fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gte:~*req.Weight:10"}, pEv); err != nil {
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
	fEv = config.NewNavigableMap(nil)
	fEv.Set([]string{utils.MetaReq}, failEvent, false, false)
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent, false, false)
	for key := range failEvent {
		if pass, err := filterS.Pass("cgrates.org", []string{"*empty:~*req." + key + ":"},
			fEv); err != nil {
			t.Errorf(err.Error())
		} else if pass {
			t.Errorf("For %s expecting: %+v, received: %+v", key, false, pass)
		}
		if pass, err := filterS.Pass("cgrates.org", []string{"*empty:~*req." + key + ":"},
			pEv); err != nil {
			t.Errorf(err.Error())
		} else if !pass {
			t.Errorf("For %s expecting: %+v, received: %+v", key, true, pass)
		}
	}
	if pass, err := filterS.Pass("cgrates.org", []string{"*exists:~*req.NewKey:"},
		fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("For NewKey expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org", []string{"*notexists:~*req.NewKey:"},
		fEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("For NewKey expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassFiltersForEventWithEmptyFilter(t *testing.T) {
	data := NewInternalDB(nil, nil)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil, nil)
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
	pEv1 := config.NewNavigableMap(nil)
	pEv1.Set([]string{utils.MetaReq}, passEvent1, false, false)
	pEv2 := config.NewNavigableMap(nil)
	pEv2.Set([]string{utils.MetaReq}, passEvent2, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{}, pEv1); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("itsyscom.com",
		[]string{}, pEv2); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	ev := map[string]interface{}{
		"Test": "MultipleCharacter",
	}
	pEv := config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, ev, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Test(~^\\w{30,})"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]interface{}{
		"Test": "MultipleCharacter123456789MoreThan30Character",
	}
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, ev, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Test(~^\\w{30,})"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	ev = map[string]interface{}{
		"Test": map[string]interface{}{
			"Test2": "MultipleCharacter",
		},
	}
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, ev, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Test.Test2(~^\\w{30,})"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]interface{}{
		"Test": map[string]interface{}{
			"Test2": "MultipleCharacter123456789MoreThan30Character",
		},
	}
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, ev, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Test.Test2(~^\\w{30,})"}, pEv); err != nil {
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
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, ev, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*req.Account:1003", "*prefix:~*req.Destination:10",
			"*suffix:~*req.Subject:03", "*rsr::~*req.Destination(1002)"},
		pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassFilterMaxCost(t *testing.T) {
	data := NewInternalDB(nil, nil)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil, nil)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	//check with max usage -1 should fail
	passEvent1 := map[string]interface{}{
		"MaxUsage": time.Duration(-1),
	}
	pEv := config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent1, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.MaxUsage{*duration_nanoseconds}(>0)"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
	//check with max usage 0 should fail
	passEvent2 := map[string]interface{}{
		"MaxUsage": time.Duration(0),
	}
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent2, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.MaxUsage{*duration_nanoseconds}(>0)"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false, received: %+v", pass)
	}
	//check with max usage 123 should pass
	passEvent3 := map[string]interface{}{
		"MaxUsage": time.Duration(123),
	}
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent3, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.MaxUsage{*duration_nanoseconds}(>0)"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true, received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true, received: %+v", pass)
	}
}

func TestPassFilterMissingField(t *testing.T) {
	data := NewInternalDB(nil, nil)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil, nil)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}

	passEvent1 := map[string]interface{}{
		"test": "call",
	}
	pEv := config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent1, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Category(^$)"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent2 := map[string]interface{}{
		"Category": "",
	}
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent2, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Category(^$)"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent3 := map[string]interface{}{
		"Category": "call",
	}
	pEv = config.NewNavigableMap(nil)
	pEv.Set([]string{utils.MetaReq}, passEvent3, false, false)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Category(^$)"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
}
