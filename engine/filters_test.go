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
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/baningo"
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
		Element: "~Category", Values: []string{"call"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passString(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}

	rf = &FilterRule{Type: utils.MetaString,
		Element: "~Category", Values: []string{"cal"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passString(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotString,
		Element: "~Category", Values: []string{"call"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: utils.MetaNotString,
		Element: "~Category", Values: []string{"cal"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
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
	rf := &FilterRule{Type: utils.MetaEmpty, Element: "~Category", Values: []string{}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaEmpty, Element: "~ExtraFields", Values: []string{}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	cd.ExtraFields = map[string]string{}
	rf = &FilterRule{Type: utils.MetaEmpty, Element: "~ExtraFields", Values: []string{}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotEmpty, Element: "~Category", Values: []string{}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
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
	rf := &FilterRule{Type: utils.MetaExists, Element: "~Category", Values: []string{}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaExists, Element: "~ExtraFields1", Values: []string{}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	cd.ExtraFields = map[string]string{}
	rf = &FilterRule{Type: utils.MetaExists, Element: "~ExtraFields", Values: []string{}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotExists, Element: "~Category1", Values: []string{}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
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
	rf := &FilterRule{Type: utils.MetaPrefix, Element: "~Category", Values: []string{"call"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~Category", Values: []string{"premium"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~Destination", Values: []string{"+49"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~Destination", Values: []string{"+499"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~navigation", Values: []string{"off"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~nonexisting", Values: []string{"off"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passing, err := rf.passStringPrefix(cd); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotPrefix, Element: "~Category", Values: []string{"premium"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
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
	rf := &FilterRule{Type: utils.MetaSuffix, Element: "~Category", Values: []string{"call"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~Category", Values: []string{"premium"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~Destination", Values: []string{"963"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~Destination", Values: []string{"4966"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~navigation", Values: []string{"off"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~nonexisting", Values: []string{"off"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passing, err := rf.passStringSuffix(cd); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotSuffix, Element: "~Destination", Values: []string{"963"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
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
	rf, err := NewFilterRule(utils.MetaRSR, "~Tenant", []string{"~^cgr.*\\.org$"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(utils.MetaRSR, "~navigation", []string{"on"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	rf, err = NewFilterRule(utils.MetaRSR, "~navigation", []string{"off"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	//not
	rf, err = NewFilterRule(utils.MetaNotRSR, "~navigation", []string{"off"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
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
	ev := utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 40)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(utils.MetaLessOrEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ASR", []string{"35.5"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("should not pass")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ACD", []string{"1m50s"})
	if err != nil {
		t.Error(err)
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ACD"}, 2*time.Minute)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not pass")
	}
	// Second
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20*time.Second)
	rf, err = NewFilterRule("*gte", "~ASR", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}

	rf, err = NewFilterRule("*gte", "~ASR", []string{"10"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, float64(20*time.Second))
	rf, err = NewFilterRule("*gte", "~ASR", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}
	//Here converter will be consider part of path and will get error : NOT_FOUND
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	rf, err = NewFilterRule("*gte", "~ASR{*duration_seconds}", []string{"10"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("passing")
	}
	//Here converter will be consider part of path and will get error : NOT_FOUND
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	rf, err = NewFilterRule("*gte", "~ASR{*duration_seconds}", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
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
	ev := utils.MapStorage{}
	ev.Set([]string{"ASR"}, 40.0)
	if passes, err := rf.passEqualTo(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 39)
	if passes, err := rf.passEqualTo(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(utils.MetaNotEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing", passes)
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, "string1")
	rf, err = NewFilterRule(utils.MetaEqual, "~ASR", []string{"string1"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passEqualTo(ev); err != nil {
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
	erf := &FilterRule{Type: utils.MetaString, Element: "~MetaString", Values: []string{"String"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaEmpty, "~MetaEmpty", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaEmpty, Element: "~MetaEmpty", Values: []string{}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaExists, "~MetaExists", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaExists, Element: "~MetaExists", Values: []string{}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaPrefix, "~MetaPrefix", []string{"stringPrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaPrefix, Element: "~MetaPrefix", Values: []string{"stringPrefix"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaSuffix, "~MetaSuffix", []string{"stringSuffix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaSuffix, Element: "~MetaSuffix", Values: []string{"stringSuffix"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	// rf, err = NewFilterRule(utils.MetaTimings, "~MetaTimings", []string{""})
	// if err != nil {
	// 	t.Errorf("Error: %+v", err)
	// }
	// erf = &FilterRule{Type: utils.MetaTimings, Element: "~MetaTimings", Values: []string{""}, negative: utils.BoolPointer(false)}
	// if err = erf.CompileValues(); err != nil {
	// 	t.Fatal(err)
	// }
	// if !reflect.DeepEqual(erf, rf) {
	// 	t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	// }
	rf, err = NewFilterRule(utils.MetaDestinations, "~MetaDestinations", []string{"1001"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaDestinations, Element: "~MetaDestinations", Values: []string{"1001"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaLessThan, "~MetaLessThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaLessThan, Element: "~MetaLessThan", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaLessOrEqual, "~MetaLessOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaLessOrEqual, Element: "~MetaLessOrEqual", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaGreaterThan, "~MetaGreaterThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaGreaterThan, Element: "~MetaGreaterThan", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~MetaGreaterOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaGreaterOrEqual, Element: "~MetaGreaterOrEqual", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
}

func TestInlineFilterPassFiltersForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
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
	fEv := utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
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
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
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
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
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
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Tenant:~^cgr.*\\.org$"}, fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Tenant:~^cgr.*\\.org$"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notrsr:~*req.Tenant:~^cgr.*\\.org$"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	failEvent = map[string]interface{}{
		utils.Weight: 10,
	}
	passEvent = map[string]interface{}{
		utils.Weight: 20,
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
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
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
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
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
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
	pEv1 := utils.MapStorage{}
	pEv1.Set([]string{utils.MetaReq}, passEvent1)
	pEv2 := utils.MapStorage{}
	pEv2.Set([]string{utils.MetaReq}, passEvent2)
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
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Test:~^\\w{30,}"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]interface{}{
		"Test": "MultipleCharacter123456789MoreThan30Character",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Test:~^\\w{30,}"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	ev = map[string]interface{}{
		"Test": map[string]interface{}{
			"Test2": "MultipleCharacter",
		},
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Test.Test2:~^\\w{30,}"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]interface{}{
		"Test": map[string]interface{}{
			"Test2": "MultipleCharacter123456789MoreThan30Character",
		},
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Test.Test2:~^\\w{30,}"}, pEv); err != nil {
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
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{
			"*string:~*req.Account:1003",
			"*prefix:~*req.Destination:10",
			"*suffix:~*req.Subject:03",
			"*rsr:~*req.Destination:1002"},
		pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassFilterMaxCost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	//check with max usage -1 should fail
	passEvent1 := map[string]interface{}{
		"MaxUsage": -1,
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent1)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0s"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
	//check with max usage 0 should fail
	passEvent2 := map[string]interface{}{
		"MaxUsage": 0,
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent2)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0s"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false, received: %+v", pass)
	}
	//check with max usage 123 should pass
	passEvent3 := map[string]interface{}{
		"MaxUsage": 123,
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent3)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0"}, pEv); err != nil {
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
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}

	passEvent1 := map[string]interface{}{
		"test": "call",
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent1)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Category:^$"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent2 := map[string]interface{}{
		"Category": "",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent2)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Category:^$"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent3 := map[string]interface{}{
		"Category": "call",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent3)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr:~*req.Category:^$"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
}

func TestEventCostFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	cd := &EventCost{
		Cost:  utils.Float64Pointer(0.264933),
		CGRID: "d8534def2b7067f4f5ad4f7ec7bbcc94bb46111a",
		Rates: ChargedRates{
			"3db483c": RateGroups{
				{
					Value:              0.1574,
					RateUnit:           60000000000,
					RateIncrement:      30000000000,
					GroupIntervalStart: 0,
				},
				{
					Value:              0.1574,
					RateUnit:           60000000000,
					RateIncrement:      1000000000,
					GroupIntervalStart: 30000000000,
				},
			},
		},
		RunID: "*default",
		Usage: utils.DurationPointer(101 * time.Second),
		Rating: Rating{
			"7f3d423": &RatingUnit{
				MaxCost:          40,
				RatesID:          "3db483c",
				TimingID:         "128e970",
				ConnectFee:       0,
				RoundingMethod:   "*up",
				MaxCostStrategy:  "*disconnect",
				RatingFiltersID:  "f8e95f2",
				RoundingDecimals: 4,
			},
		},
		Charges: []*ChargingInterval{
			{
				RatingID: "7f3d423",
				Increments: []*ChargingIncrement{
					{
						Cost:           0.0787,
						Usage:          30000000000,
						AccountingID:   "fee8a3a",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "7f3d423",
				Increments: []*ChargingIncrement{
					{
						Cost:           0.002623,
						Usage:          1000000000,
						AccountingID:   "3463957",
						CompressFactor: 71,
					},
				},
				CompressFactor: 1,
			},
		},
		Timings: ChargedTimings{
			"128e970": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
		StartTime: time.Date(2019, 12, 06, 11, 57, 32, 0, time.UTC),
		Accounting: Accounting{
			"3463957": &BalanceCharge{
				Units:         0.002623,
				RatingID:      "",
				AccountID:     "cgrates.org:1001",
				BalanceUUID:   "154419f2-45e0-4629-a203-06034ccb493f",
				ExtraChargeID: "",
			},
			"fee8a3a": &BalanceCharge{
				Units:         0.0787,
				RatingID:      "",
				AccountID:     "cgrates.org:1001",
				BalanceUUID:   "154419f2-45e0-4629-a203-06034ccb493f",
				ExtraChargeID: "",
			},
		},
		RatingFilters: RatingFilters{
			"f8e95f2": RatingMatchedFilters{
				"Subject":           "*out:cgrates.org:mo_call_UK_Mobile_O2_GBRCN:*any",
				"RatingPlanID":      "RP_MO_CALL_44800",
				"DestinationID":     "DST_44800",
				"DestinationPrefix": "44800",
			},
		},
		AccountSummary: &AccountSummary{
			ID:            "234189200129930",
			Tenant:        "cgrates.org",
			Disabled:      false,
			AllowNegative: false,
			BalanceSummaries: BalanceSummaries{
				&BalanceSummary{
					ID:       "MOBILE_DATA",
					Type:     "*data",
					UUID:     "08a05723-5849-41b9-b6a9-8ee362539280",
					Value:    3221225472,
					Disabled: false,
				},
				&BalanceSummary{
					ID:       "MOBILE_SMS",
					Type:     "*sms",
					UUID:     "06a87f20-3774-4eeb-826e-a79c5f175fd3",
					Value:    247,
					Disabled: false,
				},
				&BalanceSummary{
					ID:       "MOBILE_VOICE",
					Type:     "*voice",
					UUID:     "4ad16621-6e22-4e35-958e-5e1ff93ad7b7",
					Value:    14270000000000,
					Disabled: false,
				},
				&BalanceSummary{
					ID:       "MONETARY_POSTPAID",
					Type:     "*monetary",
					UUID:     "154419f2-45e0-4629-a203-06034ccb493f",
					Value:    50,
					Disabled: false,
				},
			},
		},
	}
	cd.initCache()
	cgrDp := utils.MapStorage{utils.MetaEC: cd}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*ec.Charges[0].Increments[0].Accounting.Balance.Value:50"}, cgrDp); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*ec.Charges[0].Increments[0].Accounting.AccountID:cgrates.org:1001"}, cgrDp); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*ec.Charges[0].Rating.Rates[0].Value:0.1574"}, cgrDp); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*ec.Charges[0].Increments[0].Accounting.Balance.ID:MONETARY_POSTPAID"}, cgrDp); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*lt:~*ec.AccountSummary.BalanceSummaries.MONETARY_POSTPAID.Value:60"}, cgrDp); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
}

func TestVerifyPrefixes(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaString, "~*req.Account", []string{"1001"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixes := []string{utils.DynamicDataPrefix + utils.MetaReq}
	if check := verifyPrefixes(rf, prefixes); !check {
		t.Errorf("Expecting: true , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*req.Account", []string{"~*req.Field1"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if check := verifyPrefixes(rf, prefixes); !check {
		t.Errorf("Expecting: true , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*req.Account", []string{"~*req.Field1", "~*req.Field2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if check := verifyPrefixes(rf, prefixes); !check {
		t.Errorf("Expecting: true , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*vars.Account", []string{"1001"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if check := verifyPrefixes(rf, prefixes); check {
		t.Errorf("Expecting: false , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*req.Account", []string{"~*req.Field1", "~*vars.Field2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if check := verifyPrefixes(rf, prefixes); check {
		t.Errorf("Expecting: false , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*req.Account", []string{"~*req.Field1", "~*vars.Field2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixes = []string{utils.DynamicDataPrefix + utils.MetaReq, utils.DynamicDataPrefix + utils.MetaVars}
	if check := verifyPrefixes(rf, prefixes); !check {
		t.Errorf("Expecting: true , received: %+v", check)
	}
}

func TestPassPartial(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent := map[string]interface{}{
		"Account": "1007",
	}
	fEv := utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, passEvent)
	prefixes := []string{utils.DynamicDataPrefix + utils.MetaReq}
	if pass, ruleList, err := filterS.LazyPass("cgrates.org",
		[]string{"*string:~*req.Account:1007"}, fEv, prefixes); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	} else if len(ruleList) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(ruleList))
	}
	// in PartialPass we verify the filters matching the prefixes
	if pass, ruleList, err := filterS.LazyPass("cgrates.org",
		[]string{"*string:~*req.Account:1007", "*string:~*vars.Field1:Val1"}, fEv, prefixes); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	} else if len(ruleList) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(ruleList))
	}
	if pass, ruleList, err := filterS.LazyPass("cgrates.org",
		[]string{"*string:~*req.Account:1010", "*string:~*vars.Field1:Val1"}, fEv, prefixes); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	} else if len(ruleList) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(ruleList))
	}
}

func TestNewFilterFromInline(t *testing.T) {
	exp := &Filter{
		Tenant: "cgrates.org",
		ID:     "*string:~*req.Account:~*uhc.<~*req.CGRID;-Account>;1001",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"~*uhc.<~*req.CGRID;-Account>", "1001"},
			},
		},
	}
	if err := exp.Compile(); err != nil {
		t.Fatal(err)
	}
	if rcv, err := NewFilterFromInline("cgrates.org", "*string:~*req.Account:~*uhc.<~*req.CGRID;-Account>;1001"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	if _, err := NewFilterFromInline("cgrates.org", "*string:~*req.Account"); err == nil {
		t.Error("Expected error received nil")
	}

	if _, err := NewFilterFromInline("cgrates.org", "*string:~*req.Account:~*req.CGRID{*;1001"); err == nil {
		t.Error("Expected error received nil")
	}
}

func TestVerifyInlineFilterS(t *testing.T) {
	if err := verifyInlineFilterS([]string{"ATTR", "*string:~*req,Acoount:1001"}); err != nil {
		t.Error(err)
	}
	if err := verifyInlineFilterS([]string{"ATTR", "*string:~*req,Acoount1001"}); err == nil {
		t.Errorf("Expected error received nil")
	}
}

func TestActivationIntervalPass(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent := map[string]interface{}{
		"CustomTime": time.Date(2013, time.July, 1, 0, 0, 0, 0, time.UTC),
	}
	fEv := utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*ai:~*req.CustomTime:2013-06-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*ai:~*req.CustomTime:2013-09-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*ai:~*req.CustomTime:;2013-09-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*ai:~*req.CustomTime:;2013-06-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*ai:~*req.CustomTime:2013-06-01T00:00:00Z;2013-09-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*ai:~*req.CustomTime:2013-08-01T00:00:00Z;2013-09-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
}

func TestFilterPassIPNet(t *testing.T) {
	cd := utils.MapStorage{
		"IP":      "192.0.2.0",
		"WrongIP": "192.0.3.",
	}
	rf := &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP", Values: []string{"192.0.2.1/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP", Values: []string{"~IP2", "192.0.3.0/30"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotIPNet,
		Element: "~IP", Values: []string{"192.0.2.0/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: utils.MetaNotIPNet,
		Element: "~IP", Values: []string{"192.0.3.0/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}

	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP2", Values: []string{"192.0.2.0/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}

	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP", Values: []string{"192.0.2.0"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}

	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~WrongIP", Values: []string{"192.0.2.0"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP{*duration}", Values: []string{"192.0.2.0/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if _, err := rf.Pass(cd); err == nil {
		t.Error(err)
	}
}

func TestAPIBan(t *testing.T) {
	var counter int
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responses := map[string]struct {
			code int
			body []byte
		}{
			"/testKey/check/1.2.3.251": {code: http.StatusOK, body: []byte(`{"ipaddress":["1.2.3.251"], "ID":"987654321"}`)},
			"/testKey/check/1.2.3.254": {code: http.StatusBadRequest, body: []byte(`{"ipaddress":["not blocked"], "ID":"none"}`)},
		}
		if val, has := responses[r.URL.EscapedPath()]; has {
			w.WriteHeader(val.code)
			if val.body != nil {
				w.Write(val.body)
			}
			return
		}
		counter++
		w.WriteHeader(http.StatusOK)
		if counter < 2 {
			_, _ = w.Write([]byte(`{"ipaddress": ["1.2.3.251", "1.2.3.252"], "ID": "100"}`))
		} else {
			_, _ = w.Write([]byte(`{"ID": "none"}`))
			counter = 0
		}
	}))
	defer testServer.Close()
	baningo.RootURL = testServer.URL + "/"

	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"bannedIP":  "1.2.3.251",
			"bannedIP2": "1.2.3.252",
			"IP":        "1.2.3.253",
			"IP2":       "1.2.3.254",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	config.CgrConfig().APIBanCfg().Keys = []string{"testKey"}
	if pass, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.IP:*all"}, dp); err != nil {
		t.Fatal(err)
	} else if pass {
		t.Error("Expected not to pass")
	}
	// from cache
	if pass, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.IP:*all"}, dp); err != nil {
		t.Fatal(err)
	} else if pass {
		t.Error("Expected not to pass")
	}
	if pass, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.IP2:*single"}, dp); err != nil {
		t.Fatal(err)
	} else if pass {
		t.Error("Expected not to pass")
	}
	Cache.Clear([]string{utils.MetaAPIBan})
	if pass, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.bannedIP:*single"}, dp); err != nil {
		t.Fatal(err)
	} else if !pass {
		t.Error("Expected to pass")
	}
	if pass, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.bannedIP2:*all"}, dp); err != nil {
		t.Fatal(err)
	} else if !pass {
		t.Error("Expected to pass")
	}

	if pass, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.notFound:*all"}, dp); err != nil {
		t.Fatal(err)
	} else if pass {
		t.Error("Expected not to pass")
	}
	expErr := "invalid value for apiban filter: <*any>"
	if _, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.IP:*any"}, dp); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}
	baningo.RootURL = "http://127.0.0.1:12345/"

	expErr = `Get "http://127.0.0.1:12345/testKey/banned/100": dial tcp 127.0.0.1:12345: connect: connection refused`
	if _, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.IP:*all"}, dp); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}
	expErr = `Get "http://127.0.0.1:12345/testKey/check/1.2.3.253": dial tcp 127.0.0.1:12345: connect: connection refused`
	if _, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.IP:*single"}, dp); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

	expErr = `invalid converter value in string: <*>, err: unsupported converter definition: <*>`
	if _, err := filterS.Pass("cgrates.org", []string{"*apiban:~*req.<~*req.IP>{*}:*all"}, dp); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}
}
