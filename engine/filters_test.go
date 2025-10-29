/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package engine

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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
	if passes, err := rf.passString(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaString,
		Element: "~Category", Values: []string{"cal"}}
	if passes, err := rf.passString(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotString,
		Element: "~Category", Values: []string{"call"}}
	if passes, err := rf.Pass(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: utils.MetaNotString,
		Element: "~Category", Values: []string{"cal"}}
	if passes, err := rf.Pass(cd, []utils.DataProvider{cd}); err != nil {
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
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaEmpty, Element: "~ExtraFields", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	cd.ExtraFields = map[string]string{}
	rf = &FilterRule{Type: utils.MetaEmpty, Element: "~ExtraFields", Values: []string{}}
	if passes, err := rf.passEmpty(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotEmpty, Element: "~Category", Values: []string{}}
	if passes, err := rf.Pass(cd, []utils.DataProvider{}); err != nil {
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
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaExists, Element: "~ExtraFields1", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	cd.ExtraFields = map[string]string{}
	rf = &FilterRule{Type: utils.MetaExists, Element: "~ExtraFields", Values: []string{}}
	if passes, err := rf.passExists(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotExists, Element: "~Category1", Values: []string{}}
	if passes, err := rf.Pass(cd, []utils.DataProvider{}); err != nil {
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
	if passes, err := rf.passStringPrefix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~Category", Values: []string{"premium"}}
	if passes, err := rf.passStringPrefix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~Destination", Values: []string{"+49"}}
	if passes, err := rf.passStringPrefix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~Destination", Values: []string{"+499"}}
	if passes, err := rf.passStringPrefix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~navigation", Values: []string{"off"}}
	if passes, err := rf.passStringPrefix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaPrefix, Element: "~nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringPrefix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotPrefix, Element: "~Category", Values: []string{"premium"}}
	if passes, err := rf.Pass(cd, []utils.DataProvider{cd}); err != nil {
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
	if passes, err := rf.passStringSuffix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~Category", Values: []string{"premium"}}
	if passes, err := rf.passStringSuffix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~Destination", Values: []string{"963"}}
	if passes, err := rf.passStringSuffix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~Destination", Values: []string{"4966"}}
	if passes, err := rf.passStringSuffix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~navigation", Values: []string{"off"}}
	if passes, err := rf.passStringSuffix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaSuffix, Element: "~nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringSuffix(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotSuffix, Element: "~Destination", Values: []string{"963"}}
	if passes, err := rf.Pass(cd, []utils.DataProvider{cd}); err != nil {
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
	if passes, err := rf.passRSR([]utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(utils.MetaRSR, "", []string{"~navigation(on)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR([]utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	rf, err = NewFilterRule(utils.MetaRSR, "", []string{"~navigation(off)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSR([]utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	//not
	rf, err = NewFilterRule(utils.MetaNotRSR, "", []string{"~navigation(off)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd, []utils.DataProvider{cd}); err != nil {
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
	if passes, err := rf.passDestinations(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewFilterRule(utils.MetaDestinations, "~Destination", []string{"RO"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, []utils.DataProvider{cd}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	//not
	rf, err = NewFilterRule(utils.MetaNotDestinations, "~Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd, []utils.DataProvider{cd}); err != nil {
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
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 40)
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(utils.MetaLessOrEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ASR", []string{"35.5"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("should not pass")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ACD", []string{"1m50s"})
	if err != nil {
		t.Error(err)
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ACD"}, time.Duration(2*time.Minute))
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not pass")
	}
	// Second
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, time.Duration(20*time.Second))
	rf, err = NewFilterRule("*gte", "~ASR", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}

	rf, err = NewFilterRule("*gte", "~ASR", []string{"10"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
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
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
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
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
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
	if passes, err := rf.passGreaterThan(ev, []utils.DataProvider{ev}); err != nil {
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
	if passes, err := rf.passEqualTo(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 39)
	if passes, err := rf.passEqualTo(ev, []utils.DataProvider{ev}); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(utils.MetaNotEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(ev, []utils.DataProvider{ev}); err != nil {
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
	if passes, err := rf.passEqualTo(ev, []utils.DataProvider{ev}); err != nil {
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
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaEmpty, "~MetaEmpty", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaEmpty, Element: "~MetaEmpty", Values: []string{}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaExists, "~MetaExists", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaExists, Element: "~MetaExists", Values: []string{}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaPrefix, "~MetaPrefix", []string{"stringPrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaPrefix, Element: "~MetaPrefix", Values: []string{"stringPrefix"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaSuffix, "~MetaSuffix", []string{"stringSuffix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaSuffix, Element: "~MetaSuffix", Values: []string{"stringSuffix"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaTimings, "~MetaTimings", []string{""})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaTimings, Element: "~MetaTimings", Values: []string{""}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaDestinations, "~MetaDestinations", []string{""})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaDestinations, Element: "~MetaDestinations", Values: []string{""}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaLessThan, "~MetaLessThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaLessThan, Element: "~MetaLessThan", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaLessOrEqual, "~MetaLessOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaLessOrEqual, Element: "~MetaLessOrEqual", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaGreaterThan, "~MetaGreaterThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaGreaterThan, Element: "~MetaGreaterThan", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~MetaGreaterOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaGreaterOrEqual, Element: "~MetaGreaterOrEqual", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
}

func TestInlineFilterPassFiltersForEvent(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	failEvent := map[string]any{
		"Account": "1001",
	}
	passEvent := map[string]any{
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
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*req.Account:1007"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notstring:~*req.Account:1007"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	failEvent = map[string]any{
		"Account": "2001",
	}
	passEvent = map[string]any{
		"Account": "1007",
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*prefix:~*req.Account:10"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*prefix:~*req.Account:10"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*suffix:~*req.Account:07"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*suffix:~*req.Account:07"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notsuffix:~*req.Account:07"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	failEvent = map[string]any{
		"Tenant": "anotherTenant.org",
	}
	passEvent = map[string]any{
		"Tenant": "cgrates.org",
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Tenant(~^cgr.*\\.org$)"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Tenant(~^cgr.*\\.org$)"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*notrsr::~*req.Tenant(~^cgr.*\\.org$)"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, "")
	failEvent = map[string]any{
		utils.Destination: "+5086517174963",
	}
	passEvent = map[string]any{
		utils.Destination: "+4986517174963",
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:~*req.Destination:EU"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:~*req.Destination:EU_LANDLINE"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	failEvent = map[string]any{
		utils.Weight: 10,
	}
	passEvent = map[string]any{
		utils.Weight: 20,
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gte:~*req.Weight:20"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gte:~*req.Weight:10"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	failEvent = map[string]any{
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
	passEvent = map[string]any{
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
			t.Error(err)
		} else if pass {
			t.Errorf("For %s expecting: %+v, received: %+v", key, false, pass)
		}
		if pass, err := filterS.Pass("cgrates.org", []string{"*empty:~*req." + key + ":"},
			pEv); err != nil {
			t.Error(err)
		} else if !pass {
			t.Errorf("For %s expecting: %+v, received: %+v", key, true, pass)
		}
	}
	if pass, err := filterS.Pass("cgrates.org", []string{"*exists:~*req.NewKey:"},
		fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("For NewKey expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org", []string{"*notexists:~*req.NewKey:"},
		fEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("For NewKey expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassRsr(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent1 := map[string]any{
		"8": "0045664",
	}
	pEv1 := utils.MapStorage{}
	pEv1.Set([]string{utils.MetaReq}, passEvent1)

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.8(~^004)"}, pEv1); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	passEvent1 = map[string]any{
		"5": "0",
	}
	pEv1 = utils.MapStorage{}
	pEv1.Set([]string{utils.MetaReq}, passEvent1)

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.5(~^0$)"}, pEv1); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassRsr2(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	dmFilterPass := NewDataManager(NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items), config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	pEv1 := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"8": "00545664",
		},
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.8(~^004)"}, pEv1); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
}

func TestPassFiltersForEventWithEmptyFilter(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent1 := map[string]any{
		utils.Tenant:      "cgrates.org",
		utils.Account:     "1010",
		utils.Destination: "+49",
		utils.Weight:      10,
	}
	passEvent2 := map[string]any{
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
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("itsyscom.com",
		[]string{}, pEv2); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	ev := map[string]any{
		"Test": "MultipleCharacter",
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Test(~^\\w{30,})"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]any{
		"Test": "MultipleCharacter123456789MoreThan30Character",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Test(~^\\w{30,})"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	ev = map[string]any{
		"Test": map[string]any{
			"Test2": "MultipleCharacter",
		},
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Test.Test2(~^\\w{30,})"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]any{
		"Test": map[string]any{
			"Test2": "MultipleCharacter123456789MoreThan30Character",
		},
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Test.Test2(~^\\w{30,})"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}

	ev = map[string]any{
		utils.Account:     "1003",
		utils.Subject:     "1003",
		utils.Destination: "1002",
		utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
		utils.Usage:       "1m20s",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*req.Account:1003", "*prefix:~*req.Destination:10",
			"*suffix:~*req.Subject:03", "*rsr::~*req.Destination(1002)"},
		pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassFilterMaxCost(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	//check with max usage -1 should fail
	passEvent1 := map[string]any{
		"MaxUsage": time.Duration(-1),
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent1)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.MaxUsage{*duration_nanoseconds}(>0)"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
	//check with max usage 0 should fail
	passEvent2 := map[string]any{
		"MaxUsage": time.Duration(0),
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent2)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.MaxUsage{*duration_nanoseconds}(>0)"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: false, received: %+v", pass)
	}
	//check with max usage 123 should pass
	passEvent3 := map[string]any{
		"MaxUsage": time.Duration(123),
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent3)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.MaxUsage{*duration_nanoseconds}(>0)"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true, received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true, received: %+v", pass)
	}
}

func TestPassFilterMissingField(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}

	passEvent1 := map[string]any{
		"test": "call",
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent1)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Category(^$)"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent2 := map[string]any{
		"Category": "",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent2)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~Category(^$)"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent3 := map[string]any{
		"Category": "call",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent3)
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*rsr::~*req.Category(^$)"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
}

func TestEventCostFilter(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
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
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*ec.Charges[0].Increments[0].Accounting.AccountID:cgrates.org:1001"}, cgrDp); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*ec.Charges[0].Rating.Rates[0].Value:0.1574"}, cgrDp); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*ec.Charges[0].Increments[0].Accounting.Balance.ID:MONETARY_POSTPAID"}, cgrDp); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}
}

func TestComputeThresholdIndexes(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	thd1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FLTR_2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
	}
	if err := dm.SetThresholdProfile(thd1, true); err != nil {
		t.Error(err)
	}
	thd2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH_2",
		FilterIDs: []string{utils.META_NONE},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
	}
	if err := dm.SetThresholdProfile(thd2, true); err != nil {
		t.Error(err)
	}
	thIDs := []string{"TH_1", "TH_2"}
	if _, err := ComputeThresholdIndexes(dm, "cgrates.org", &thIDs, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringMap{
		"*prefix:~*req.Account:1001": {"TH_1": true},
		"*none:*any:*any":            {"TH_2": true},
	}

	if fltrIndexer, err := ComputeThresholdIndexes(dm, "cgrates.org", nil, "ID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fltrIndexer.indexes, expIndexes) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(expIndexes), utils.ToJSON(fltrIndexer.indexes))
	}
}

func TestComputeChargerIndexes(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter3",
		Rules: []*FilterRule{
			{
				Element: "~*req.Destination",
				Type:    utils.MetaString,
				Values:  []string{"10", "20"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	chP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CHRG_1",
		FilterIDs: []string{"Filter3", "*string:~*req.Account:1001", utils.META_NONE},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		AttributeIDs: []string{"ATTR_1"},
		Weight:       20,
	}

	if err := dm.SetChargerProfile(chP, true); err != nil {
		t.Error(err)
	}
	chIDs := []string{"CHRG_1"}
	if _, err := ComputeChargerIndexes(dm, "cgrates.org", &chIDs, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringMap{
		"*string:~*req.Account:1001":   {"CHRG_1": true},
		"*string:~*req.Destination:10": {"CHRG_1": true},
		"*string:~*req.Destination:20": {"CHRG_1": true},
		"*none:*any:*any":              {"CHRG_1": true},
	}
	if fltrIndexer, err := ComputeChargerIndexes(dm, "cgrates.org", nil, "ID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fltrIndexer.indexes, expIndexes) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(expIndexes), utils.ToJSON(fltrIndexer.indexes))
	}
}

func TestComputeResourceIndexes(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_RES_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Resources",
				Values:  []string{"ResourceProfile2"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	rs := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_GR_TEST",
		FilterIDs: []string{"FLTR_RES_1", utils.META_NONE},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(-1),
		Limit:             2,
		AllocationMessage: "Account1Channels",
		Weight:            20,
		ThresholdIDs:      []string{utils.META_NONE},
	}
	if err := dm.SetResourceProfile(rs, true); err != nil {
		t.Error(err)
	}
	chIDs := []string{"RES_GR_TEST"}
	if _, err := ComputeResourceIndexes(dm, "cgrates.org", &chIDs, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringMap{
		"*none:*any:*any":                          {"RES_GR_TEST": true},
		"*string:~*req.Resources:ResourceProfile2": {"RES_GR_TEST": true},
	}
	if fltrIndexer, err := ComputeResourceIndexes(dm, "cgrates.org", nil, "ID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fltrIndexer.indexes, expIndexes) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(fltrIndexer.indexes), utils.ToJSON(expIndexes))
	}
}
func TestComputeSupplierIndexes(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_SUPP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Supplier",
				Values:  []string{"SupplierProfile1"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	spp := &SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_2",
		Sorting:   utils.MetaLC,
		FilterIDs: []string{"FLTR_SUPP_1"},
		Suppliers: []*Supplier{
			{
				ID:         "SPL1",
				FilterIDs:  []string{"FLTR_1"},
				AccountIDs: []string{"accc"},
				Weight:     20,
				Blocker:    false,
			},
		},
		Weight: 10,
	}
	if err := dm.SetSupplierProfile(spp, true); err != nil {
		t.Error(err)
	}
	chIDs := []string{"SPL_2"}
	if _, err := ComputeSupplierIndexes(dm, "cgrates.org", &chIDs, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringMap{
		"*string:~*req.Supplier:SupplierProfile1": {"SPL_2": true},
	}
	if fltrIndexes, err := ComputeSupplierIndexes(dm, "cgrates.org", nil, "ID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIndexes, fltrIndexes.indexes) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(expIndexes), utils.ToJSON(fltrIndexes.indexes))
	}
}

func TestRemoveItemFromIndexRP(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	fltr.Compile()
	rs := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_GR_TEST",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(-1),
		Limit:             2,
		AllocationMessage: "Account1Channels",
		Weight:            20,
		ThresholdIDs:      []string{utils.META_NONE},
	}
	if err := dm.SetResourceProfile(rs, true); err != nil {
		t.Error(err)
	}
	fltInd := NewFilterIndexer(dm, utils.ResourceProfilesPrefix, rs.Tenant)

	if err := fltInd.RemoveItemFromIndex("cgrates.org", rs.ID, []string{}); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveResourceProfile("cgrates.org", rs.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestRemoveItemFromIndexCHP(t *testing.T) {
	Cache.Clear(nil)

	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		}}
	if err := fltr.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	chp := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Raw",
		FilterIDs:    []string{"FLTR_CP_2"},
		RunID:        utils.MetaRaw,
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       0,
	}
	if err := dm.SetChargerProfile(chp, false); err != nil {
		t.Error(err)
	}
	fltInd := NewFilterIndexer(dm, utils.ChargerProfilePrefix, chp.Tenant)
	if err := fltInd.RemoveItemFromIndex("cgrates.org", chp.ID, []string{}); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveFilter("cgrates.org", fltr.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := fltInd.RemoveItemFromIndex("cgrates.org", chp.ID, []string{}); err == nil {
		t.Error(err)
	}
	if err := dm.SetChargerProfile(chp, true); err == nil || !strings.HasPrefix(err.Error(), "broken reference to filter:") {
		t.Error(err)
	}
}

func TestComputeStatIndexes(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Stats",
				Values:  []string{"StatQueueProfile1"},
			},
		},
	}
	if err := fltr.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	sq := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE2",
		FilterIDs: []string{"FLTR_STATS_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum",
			},
			{
				MetricID: utils.MetaACD,
			},
		},
		ThresholdIDs: []string{"Val1", "Val2"},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     1,
	}

	if err := dm.SetStatQueueProfile(sq, true); err != nil {
		t.Error(err)
	}
	stIDs := []string{"TEST_PROFILE2"}
	if _, err := ComputeStatIndexes(dm, "cgrates.org", &stIDs, ""); err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringMap{
		"*string:~*req.Stats:StatQueueProfile1": {"TEST_PROFILE2": true},
	}
	if fltrIndexer, err := ComputeStatIndexes(dm, "cgrates.org", nil, "ID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIndexes, fltrIndexer.indexes) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(expIndexes), utils.ToJSON(fltrIndexer))
	}
}

func TestComputeAttributeIndexes(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltrAttr1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Attribute",
				Values:  []string{"AttributeProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req." + utils.Weight,
				Values:  []string{"9.0"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001"},
			},
		},
	}
	if err := fltrAttr1.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(fltrAttr1); err != nil {
		t.Error(err)
	}
	ap := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf1",
		FilterIDs: []string{"FLTR_ATTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts: []string{"con1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	if err := dm.SetAttributeProfile(ap, true); err != nil {
		t.Error(err)
	}
	ids := []string{"AttrPrf1"}
	if _, err := ComputeAttributeIndexes(dm, "cgrates.org", "con1", &ids, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringMap{
		"*string:~*req.Attribute:AttributeProfile1": {
			"AttrPrf1": true,
		},
		"*prefix:~*req.Account:1001": {
			"AttrPrf1": true,
		},
	}
	if fltrIndexer, err := ComputeAttributeIndexes(dm, "cgrates.org", "con1", nil, "ID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIndexes, fltrIndexer.indexes) {
		t.Errorf("Expected %v , Receveid %v", utils.ToJSON(expIndexes), utils.ToJSON(fltrIndexer.indexes))
	}

}

func TestComputeDispatcherIndexes(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "DSP_FLT",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"2009"},
			},
		}}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}

	dpP := &DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"DSP_FLT"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Strategy:   utils.MetaFirst,
		Weight:     20,
		Subsystems: []string{utils.MetaAttributes, utils.MetaSessionS},
	}
	if err := dm.SetDispatcherProfile(dpP, true); err != nil {
		t.Error(err)
	}
	ids := []string{"Dsp1"}
	if _, err := ComputeDispatcherIndexes(dm, "cgrates.org", utils.MetaAttributes, &ids, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expIndexes := map[string]utils.StringMap{
		"*string:~*req.Account:2009": {
			"Dsp1": true,
		},
	}
	if fltrIndexer, err := ComputeDispatcherIndexes(dm, "cgrates.org", utils.MetaSessionS, nil, "ID"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIndexes, fltrIndexer.indexes) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(expIndexes), utils.ToJSON(fltrIndexer.indexes))
	}
}

func TestRemoveItemFromIndexSQP(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	sqs := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "DistinctMetricProfile",
		QueueLength: 10,
		FilterIDs:   []string{"FLTR_1"},
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaDDC,
			},
		},
		ThresholdIDs: []string{utils.META_NONE},
		Stored:       true,
		Weight:       20,
	}
	if err := dm.SetStatQueueProfile(sqs, true); err != nil {
		t.Error(err)
	}
	fltrIndexer := NewFilterIndexer(dm, utils.StatQueueProfilePrefix, sqs.Tenant)
	if err := fltrIndexer.RemoveItemFromIndex(sqs.Tenant, sqs.ID, []string{}); err != nil {
		t.Error(err)
	}
}

func TestRemoveItemFromIndexSPP(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_SUPP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Supplier",
				Values:  []string{"SupplierProfile1"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	spp := &SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_ACNT_1001",
		FilterIDs: []string{"FLTR_SUPP_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:     "supplier1",
				Weight: 10,
			},
		},
		Weight: 20,
	}
	if err := dm.SetSupplierProfile(spp, true); err != nil {
		t.Error(err)
	}
	fltrIndexer := NewFilterIndexer(dm, utils.SupplierProfilePrefix, spp.Tenant)
	if err := fltrIndexer.RemoveItemFromIndex(spp.Tenant, spp.ID, []string{}); err != nil {
		t.Error(err)
	}
}

func TestRemoveItemFromIndexDP(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "DSP_FLT",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"2009"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	dpp := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DSP_Test1",
		FilterIDs:  []string{"DSP_FLT"},
		Strategy:   utils.MetaFirst,
		Subsystems: []string{utils.MetaAttributes, utils.MetaSessionS},
		Weight:     20,
	}
	if err := dm.SetDispatcherProfile(dpp, true); err != nil {
		t.Error(err)
	}
	fltrIndexer := NewFilterIndexer(dm, utils.DispatcherProfilePrefix, dpp.Tenant)
	if err := fltrIndexer.RemoveItemFromIndex(dpp.Tenant, dpp.ID, []string{}); err != nil {
		t.Error(err)
	}
}

func TestUpdateFilterIndexes(t *testing.T) {

	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001", "1002"},
			},
			{
				Type:    "*prefix",
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Values:  []string{"10", "20"},
			},
			{
				Type:    "*rsr",
				Element: "",
				Values:  []string{"~*req.Subject(~^1.*1$)", "~*req.Destination(1002)"},
			},
		},
	}
	if err := dm.SetFilter(oldFlt); err != nil {
		t.Error(err)
	}
	chg := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CPP_3",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weight:       20,
	}

	if err := dm.SetChargerProfile(chg, true); err != nil {
		t.Error(err)
	}
	thP := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_ACNT_1001",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		MaxHits:   1,
		MinHits:   1,
		MinSleep:  time.Duration(1 * time.Second),
		Weight:    10.0,
		ActionIDs: []string{"ACT_LOG_WARNING"},
		Async:     true,
	}
	if err := dm.SetThresholdProfile(thP, true); err != nil {
		t.Error(err)
	}
	rcf := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RCFG1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(10) * time.Microsecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
	}
	if err := dm.SetResourceProfile(rcf, true); err != nil {
		t.Error(err)
	}
	supp := &SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_DESTINATION",
		FilterIDs: []string{"FLTR_1"},
		Sorting:   utils.MetaLC,
		Suppliers: []*Supplier{
			{
				ID:            "local",
				RatingPlanIDs: []string{"RP_LOCAL"},
				Weight:        10,
			},
			{
				ID:            "mobile",
				RatingPlanIDs: []string{"RP_MOBILE"},
				FilterIDs:     []string{"*destinations:~*req.Destination:DST_MOBILE"},
				Weight:        10,
			},
		},
		Weight: 100,
	}
	if err := dm.SetSupplierProfile(supp, true); err != nil {
		t.Error(err)
	}
	stat := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaACD,
			},
		},
		ThresholdIDs: []string{"THD_ACNT_1001"},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     1,
	}
	if err := dm.SetStatQueueProfile(stat, true); err != nil {
		t.Error(err)
	}
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.META_ANY},
		FilterIDs: []string{"FLTR_1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
				Value: config.NewRSRParsersMustCompile("1011", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	if err := dm.SetAttributeProfile(attr, true); err != nil {
		t.Error(err)
	}
	dpp := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DSP_Test1",
		FilterIDs:  []string{"FLTR_1"},
		Strategy:   utils.MetaFirst,
		Subsystems: []string{utils.MetaAttributes, utils.MetaSessionS},
		Weight:     20,
	}
	if err := dm.SetDispatcherProfile(dpp, true); err != nil {
		t.Error(err)
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1001"},
			},
			{
				Type:    utils.MetaRSR,
				Element: utils.EmptyString,
				Values:  []string{"~*req.Tenant(~^cgr.*\\.org$)"},
			},
		},
	}
	if err := UpdateFilterIndexes(dm, "cgrates.org", oldFlt, newFlt); err != nil {
		t.Error(err)
	}

}

func TestFilterSPass11(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		Cache.Clear(nil)
	}()
	cfg.FilterSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.FilterSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	acc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.VOICE: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("DST1"),
					Weight:         10},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("DST2"), Weight: 20},
			}},
	}
	rsr := &Resource{
		Tenant: "cgrates.org",
		ID:     "RL1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				ID:         "RU1",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
				Units:      2,
			},
		},
		TTLIdx: []string{"RU1"},
	}
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ_1",
		dirty:  utils.BoolPointer(true),
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 1,
				Count:    1,
				Events: map[string]*StatWithCompress{
					"cgrates.org:TestStatRemExpired_1": {Stat: 1, CompressFactor: 1},
				},
			},
		}}
	dm.SetResource(rsr)
	dm.SetAccount(acc)
	dm.SetStatQueue(sq)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, args, reply any) error {
		if serviceMethod == utils.ResourceSv1GetResource {
			tntId, concat := args.(*utils.TenantID)
			if !concat {
				return utils.ErrNotConvertible
			}
			rpl, err := dm.GetResource(tntId.Tenant, tntId.ID, false, false, utils.NonTransactional)
			if err != nil {
				return err
			}
			*reply.(**Resource) = rpl
			return nil
		} else if serviceMethod == utils.StatSv1GetQueueFloatMetrics {
			rpl := map[string]float64{
				utils.MetaACC: 100.0,
			}
			*reply.(*map[string]float64) = rpl
			return nil
		}
		return utils.ErrNotImplemented
	})

	fltrs := []*Filter{
		{
			Tenant: "cgrates.org",
			ID:     "FLTR_ACC",
			Rules: []*FilterRule{{
				Type:    utils.MetaString,
				Element: "~*accounts.1001.BalanceMap.*voice[0].Value",
				Values:  []string{"~*accounts.1001.BalanceMap.*voice[0].Value" + utils.IfaceAsString(20*float64(time.Second))},
			}},
		},
		{
			Tenant: "cgrates.org",
			ID:     "FLTR_RES",
			Rules: []*FilterRule{
				{
					Type:    "*lte",
					Element: "~*resources.RL1.Usage.RUI.Units",
					Values:  []string{"~*resources.RL1.Usage.RUI.Units.2"},
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "FLTR_STAT",
			Rules: []*FilterRule{
				{
					Type:    "*gt",
					Element: "~*stats.SQ_1.*asr",
					Values:  []string{"~*stats.SQ_1.*asr.10.0"},
				},
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS):     clientConn,
	})
	ev := &utils.MapStorage{}
	fS := NewFilterS(cfg, connMgr, dm)
	fltriDs := make([]string, len(fltrs))
	for i, fltr := range fltrs {
		dm.SetFilter(fltr)
		fltriDs[i] = fltr.ID
	}
	if _, err := fS.Pass("cgrates.org", fltriDs, ev); err != nil {
		t.Error(err)
	}
}

func TestValidateInlineFilters(t *testing.T) {
	cases := []struct {
		fltrs       []string
		expectedErr string
	}{
		{[]string{"FLTR", "*string:~*req.Account:1001"}, ""},
		{[]string{"FLTR", "*string:~*req,Acoount1001"}, "inline parse error for string: <*string:~*req,Acoount1001>"},
		{[]string{"*exists:~*req.Supplier:"}, ""},
		{[]string{"*exists:*req.Supplier"}, "inline parse error for string: <*exists:*req.Supplier>"},
		{[]string{"*rsr:~*req.Account:(10)"}, "invalid RSRFilter start rule in string: <(10)>"},
	}
	computeTestName := func(idx int, params []string) string {
		return fmt.Sprintf("Test No %d with parameters: %v", idx, params)
	}

	for i, c := range cases {
		t.Run(computeTestName(i, c.fltrs), func(t *testing.T) {
			err := validateInlineFilters(c.fltrs)
			if err != nil {
				if c.expectedErr == "" {
					t.Errorf("did not expect error, received: %v", err)
				}
			} else if c.expectedErr != "" {
				t.Errorf("expected error: %v", err)
			}
		})
	}
}
func TestComputeDispatcherIndexesErr(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	dm.SetDispatcherProfile(&DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DSP1",
		FilterIDs:  []string{"DSP_FLT"},
		Strategy:   utils.MetaFirst,
		Subsystems: []string{utils.MetaAttributes, utils.MetaSessionS},
		Weight:     20,
	}, true)

	if _, err := ComputeDispatcherIndexes(dm, "cgrates.org", utils.MetaSessionS, &[]string{"DSP1"}, ""); err == nil {
		t.Error(err)

	}
}

func TestFIRemoveItemFromIndexErr(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		return utils.ErrNotImplemented
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	testCases := []struct {
		name       string
		itemType   string
		cachePart  string
		itemID     string
		oldFilters []string
	}{
		{
			name:      "Threshold Not Found",
			itemType:  utils.ThresholdProfilePrefix,
			itemID:    "TH1",
			cachePart: utils.MetaThresholdProfiles,
		},
		{
			name:      "Attributes Not Found",
			itemType:  utils.AttributeProfilePrefix,
			itemID:    "ATTR_1",
			cachePart: utils.MetaAttributeProfiles,
		},
		{
			name:      "Resource Not Found",
			itemType:  utils.ResourceProfilesPrefix,
			itemID:    "RSC_1",
			cachePart: utils.MetaResourceProfile,
		},
		{
			name:      "Stat Queue Not Found",
			itemType:  utils.StatQueueProfilePrefix,
			itemID:    "SQ_1",
			cachePart: utils.MetaStatQueueProfiles,
		},
		{
			name:      "Supplier Not Found",
			itemType:  utils.SupplierProfilePrefix,
			itemID:    "SPP_1",
			cachePart: utils.MetaSupplierProfiles,
		},
		{
			name:      "ChargerProfile  Not Found",
			itemType:  utils.ChargerProfilePrefix,
			itemID:    "SQ_1",
			cachePart: utils.MetaChargerProfiles,
		},
		{
			name:      "DispatcherProfile   Not Found",
			itemType:  utils.DispatcherProfilePrefix,
			itemID:    "SPP_1",
			cachePart: utils.MetaDispatcherProfiles,
		},
	}
	config.SetCgrConfig(cfg)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg.DataDbCfg().Items[tc.cachePart].Remote = true
			if err := NewFilterIndexer(dm, tc.itemType, "cgrates.org").RemoveItemFromIndex("cgrates.org", tc.itemID, tc.oldFilters); err == nil {
				t.Error("expected error, received nil")
			}
		})
	}
}

func TestComputeSupplierIndexesErrs(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	testCases := []struct {
		name   string
		dm     *DataManager
		tenant string
		sppIDs *[]string
		expErr bool
	}{
		{
			name:   "Supplier Not Found",
			dm:     dm,
			expErr: true,
			sppIDs: &[]string{"SPP_1"},
		},
		{
			name:   "Without Filter",
			dm:     dm,
			expErr: false,
			sppIDs: &[]string{"SPP_2"},
		},
		{
			name:   "Filter doesn't exist",
			dm:     dm,
			expErr: true,
			sppIDs: &[]string{"SPP_3"},
		},
	}
	dm.SetSupplierProfile(&SupplierProfile{
		Tenant:  "cgrates.org",
		ID:      "SPP_2",
		Sorting: utils.MetaQOS,
	}, true)

	dm.SetSupplierProfile(&SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPP_3",
		Sorting:   utils.MetaQOS,
		FilterIDs: []string{"FLT_1"},
	}, true)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ComputeSupplierIndexes(tc.dm, "cgrates.org", tc.sppIDs, ""); (err != nil) != tc.expErr {
				t.Errorf("expected error: %v, received: %v", tc.expErr, err)
			}
		})
	}
}

func TestComputeStatIndexesErrs(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	testCases := []struct {
		name   string
		dm     *DataManager
		tenant string
		stIDs  *[]string
		expErr bool
	}{
		{
			name:   "StatQueueProfile Not Found",
			dm:     dm,
			expErr: true,
			stIDs:  &[]string{"SQ_1"},
		},
		{
			name:   "Without Filter",
			dm:     dm,
			expErr: false,
			stIDs:  &[]string{"TEST_PROFILE1"},
		},
		{
			name:   "Filter dosnt exist ",
			dm:     dm,
			expErr: true,
		},
	}
	dm.SetStatQueueProfile(&StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "TEST_PROFILE1",
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum:~Val",
			},
		},
		ThresholdIDs: []string{"*none"},
	}, true)
	dm.SetStatQueueProfile(&StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "TestStats",
		FilterIDs:   []string{"FLTR_STATS"},
		QueueLength: 100,
		TTL:         time.Duration(1 * time.Second),
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum:~Value",
			},
			{
				MetricID: "*average:~Value",
			},
			{
				MetricID: "*sum:~Usage",
			},
		},
	}, true)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ComputeStatIndexes(tc.dm, "cgrates.org", tc.stIDs, ""); (err != nil) != tc.expErr {
				t.Errorf("expected error: %v, received: %v", tc.expErr, err)
			}
		})
	}
}

func TestComputeAttributeIndexesErr(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	testCases := []struct {
		name    string
		dm      *DataManager
		tenant  string
		attrIDs *[]string
		expErr  bool
	}{
		{
			name:    "AttributeProfile Not Found",
			dm:      dm,
			expErr:  true,
			attrIDs: &[]string{"ATTR_1"},
		},
		{
			name:    "Without Filter",
			dm:      dm,
			expErr:  false,
			attrIDs: &[]string{"ATTR_1002_SIMPLEAUTH"},
		},
		{
			name:   "Filter dosnt exist ",
			dm:     dm,
			expErr: true,
		},
	}
	dm.SetAttributeProfile(&AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR_1002_SIMPLEAUTH",
		Contexts: []string{"simpleauth"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Password",
				Type:  utils.META_CONSTANT,
				Value: config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20.0,
	}, true)
	dm.SetAttributeProfile(&AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1001_AUTH",
		Contexts:  []string{"simpleauth"},
		FilterIDs: []string{"FLTR_Attr"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Password",
				Type:  utils.META_CONSTANT,
				Value: config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20.0,
	}, true)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ComputeAttributeIndexes(tc.dm, "cgrates.org", "simpleauth", tc.attrIDs, ""); (err != nil) != tc.expErr {
				t.Errorf("expected error: %v, received: %v", tc.expErr, err)
			}
		})
	}
}

func TestComputeDispatcherIndexesErrs(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	testCases := []struct {
		name   string
		dm     *DataManager
		tenant string
		dspIDs *[]string
		expErr bool
	}{
		{
			name:   "DispatcherProfile  Not Found",
			dm:     dm,
			tenant: "cgrates.org",
			dspIDs: &[]string{"DISP_1"},
			expErr: true,
		},
		{
			name:   "Without Filter",
			dm:     dm,
			tenant: "cgrates.org",
			dspIDs: &[]string{"DISP_2"},
			expErr: false,
		},
	}
	dm.SetDispatcherProfile(&DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DISP_2",
		Subsystems: []string{utils.MetaSessionS},
	}, true)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ComputeDispatcherIndexes(tc.dm, "cgrates.org", utils.MetaSessionS, tc.dspIDs, ""); (err != nil) != tc.expErr {
				t.Errorf("expected error: %v, received: %v", tc.expErr, err)
			}
		})
	}
}

func TestComputeResourceIndexesErrs(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	testCases := []struct {
		name   string
		dm     *DataManager
		tenant string
		resIDs *[]string
		expErr bool
	}{
		{
			name:   "ResourceProfile  Not Found",
			dm:     dm,
			tenant: "cgrates.org",
			resIDs: &[]string{"RES_1"},
			expErr: true,
		},
		{
			name:   "Without Filter",
			dm:     dm,
			tenant: "cgrates.org",
			resIDs: &[]string{"RES_2"},
			expErr: true,
		},
	}
	dm.SetResourceProfile(&ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES_2",
		FilterIDs:         []string{"FLTR_RES"},
		UsageTTL:          time.Duration(-1),
		Limit:             2,
		AllocationMessage: "Account1Channels",
		Weight:            20,
		ThresholdIDs:      []string{utils.META_NONE},
	}, true)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ComputeResourceIndexes(tc.dm, "cgrates.org", tc.resIDs, ""); (err != nil) != tc.expErr {
				t.Errorf("expected error: %v, received: %v", tc.expErr, err)
			}
		})
	}
}

func TestFiltersPassErrors(t *testing.T) {

	type args struct {
		rf   string
		fn   string
		vals []string
	}

	type exp struct {
		rcv *FilterRule
		err string
	}

	tests := []struct {
		name string
		args args
		exp  exp
	}{
		{
			name: "unsupported filters error",
			args: args{rf: "test", fn: "test", vals: []string{"val1"}},
			exp:  exp{rcv: nil, err: "Unsupported filter Type: test"},
		},
		{
			name: "element is mandatory error",
			args: args{rf: "*string", fn: "", vals: []string{"val1"}},
			exp:  exp{rcv: nil, err: "Element is mandatory for Type: *string"},
		},
		{
			name: "values is mandatory error",
			args: args{rf: "*string", fn: "test", vals: []string{}},
			exp:  exp{rcv: nil, err: "Values is mandatory for Type: *string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := NewFilterRule(tt.args.rf, tt.args.fn, tt.args.vals)

			if err != nil {
				if err.Error() != tt.exp.err {
					t.Errorf("expected error: %v, received: %v", tt.exp.err, err)
				}
			}

			if rcv != nil {
				t.Error(rcv)
			}
		})
	}
}

func TestFiltersPass(t *testing.T) {
	type args struct {
		fln *utils.MapStorage
		flv []utils.DataProvider
	}
	type exp struct {
		r   bool
		err string
	}

	f := FilterRule{
		Type: "*timings",
	}
	f2 := FilterRule{
		Type: "*test",
	}

	tests := []struct {
		name string
		args args
		exp  exp
	}{
		{
			name: "meta timing error",
			args: args{nil, nil},
			exp:  exp{false, "NOT_IMPLEMENTED"},
		},
		{
			name: "default error",
			args: args{nil, nil},
			exp:  exp{false, "NOT_IMPLEMENTED:*test"},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rcv bool
			var err error
			if i == 0 {
				rcv, err = f.Pass(tt.args.fln, tt.args.flv)
			} else {
				rcv, err = f2.Pass(tt.args.fln, tt.args.flv)
			}

			if err != nil {
				if err.Error() != tt.exp.err {
					t.Errorf("expected error: %v, received: %v", tt.exp.err, err)
				}
			}

			if rcv != tt.exp.r {
				t.Error(rcv)
			}
		})
	}
}

func TestFilterspassString(t *testing.T) {
	f := FilterRule{
		Element: "~test.test[0` ",
	}

	rcv, err := f.passString(utils.MapStorage{}, []utils.DataProvider{})

	if err != nil {
		if err.Error() != "" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f2 := FilterRule{
		Element: "~test.test",
	}

	rcv, err = f2.passString(utils.MapStorage{"test": "test"}, []utils.DataProvider{})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}
}

func TestFilterspassExists(t *testing.T) {
	f := FilterRule{
		Element: "~test.test",
	}

	rcv, err := f.passExists(utils.MapStorage{"test": "test"})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}
}

func TestFilterspassEmpty(t *testing.T) {
	f := FilterRule{
		Element: "~test.test[0` ",
	}

	rcv, err := f.passEmpty(utils.MapStorage{})

	if err != nil {
		if err.Error() != "" {
			t.Error(err)
		}
	}

	if rcv != true {
		t.Error("Recived:", rcv)
	}

	f2 := FilterRule{
		Element: "~test.test",
	}

	rcv, err = f2.passEmpty(utils.MapStorage{"test": "test"})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}
}

func TestFilterspassStringPrefix(t *testing.T) {
	f := FilterRule{
		Element: "~test.test",
	}

	rcv, err := f.passStringPrefix(utils.MapStorage{"test": "test"}, []utils.DataProvider{})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f2 := FilterRule{
		Values: []string{"~test.test"},
	}

	rcv, err = f2.passStringPrefix(utils.MapStorage{"test": "test"}, []utils.DataProvider{utils.MapStorage{"test": "test"}})

	if err != nil {
		if err.Error() != "" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}
}

func TestFilterspassStringSuffix(t *testing.T) {
	f := FilterRule{
		Element: "~test.test",
	}

	rcv, err := f.passStringSuffix(utils.MapStorage{"test": "test"}, []utils.DataProvider{})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f2 := FilterRule{
		Values: []string{"~test.test"},
	}

	rcv, err = f2.passStringSuffix(utils.MapStorage{"test": "test"}, []utils.DataProvider{utils.MapStorage{"test": "test"}})

	if err != nil {
		if err.Error() != "" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}
}

func TestFilterspassDestination(t *testing.T) {
	f := FilterRule{
		Element: "~test.test",
	}

	rcv, err := f.passDestinations(utils.MapStorage{"test": "test"}, []utils.DataProvider{})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f2 := FilterRule{
		Element: "~test.test[0` ",
	}

	rcv, err = f2.passDestinations(utils.MapStorage{}, []utils.DataProvider{})

	if err != nil {
		if err.Error() != "" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}
}

func TestFilterspassGreaterThan(t *testing.T) {
	f := FilterRule{
		Element: "~test.test",
	}

	rcv, err := f.passGreaterThan(utils.MapStorage{"test": "test"}, []utils.DataProvider{})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f2 := FilterRule{
		Values: []string{"~test.test"},
	}

	rcv, err = f2.passGreaterThan(utils.MapStorage{"test": "test"}, []utils.DataProvider{utils.MapStorage{"test": "test"}})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f4 := FilterRule{
		Values: []string{"val1", "val2"},
	}

	rcv, err = f4.passGreaterThan(utils.MapStorage{}, []utils.DataProvider{utils.MapStorage{"test": "test"}})

	if err != nil {
		if err.Error() != "incomparable: <0001-01-01 00:00:00 +0000 UTC> with <<nil>>" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}
}

func TestFilterspassEqualTo(t *testing.T) {
	f := FilterRule{
		Element: "~test.test",
	}

	rcv, err := f.passEqualTo(utils.MapStorage{"test": "test"}, []utils.DataProvider{})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f2 := FilterRule{
		Values: []string{"~test.test"},
	}

	rcv, err = f2.passEqualTo(utils.MapStorage{"test": "test"}, []utils.DataProvider{utils.MapStorage{"test": "test"}})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f3 := FilterRule{
		Element: "~test.test",
	}

	rcv, err = f3.passEqualTo(utils.MapStorage{}, []utils.DataProvider{utils.MapStorage{"test": "test"}})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}

	f4 := FilterRule{
		Values: []string{"val1", "val2"},
	}

	rcv, err = f4.passEqualTo(utils.MapStorage{}, []utils.DataProvider{utils.MapStorage{"test": "test"}})

	if err != nil {
		if err.Error() != "incomparable: <0001-01-01 00:00:00 +0000 UTC> with <<nil>>" {
			t.Error(err)
		}
	}

	if rcv != false {
		t.Error("Recived:", rcv)
	}
}

func TestFiltersNewFilterRule(t *testing.T) {
	_, err := NewFilterRule("*rsr", "test", []string{"test)"})

	if err != nil {
		if err.Error() != "invalid RSRFilter start rule in string: <test)>" {
			t.Error(err)
		}
	}
}

func TestFiltersgetFieldNameDataProvider(t *testing.T) {
	fS := FilterS{}

	type args struct {
		initialDP utils.DataProvider
		fieldName string
		tenant    string
	}

	tests := []struct {
		name string
		args args
		exp  utils.DataProvider
		err  string
	}{
		{
			name: "",
			args: args{
				initialDP: utils.MapStorage{},
				fieldName: "~*accounts",
				tenant:    "test",
			},
			exp: nil,
			err: "invalid fieldname <~*accounts>",
		},
		{
			name: "",
			args: args{
				initialDP: utils.MapStorage{},
				fieldName: "~*resources",
				tenant:    "test",
			},
			exp: nil,
			err: "invalid fieldname <~*resources>",
		},
		{
			name: "",
			args: args{
				initialDP: utils.MapStorage{},
				fieldName: "~*stats",
				tenant:    "test",
			},
			exp: nil,
			err: "invalid fieldname <~*stats>",
		},
		{
			name: "",
			args: args{
				initialDP: utils.MapStorage{},
				fieldName: "test",
				tenant:    "test",
			},
			exp: nil,
			err: "filter path: <test> doesn't have a valid prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := fS.getFieldNameDataProvider(tt.args.initialDP, tt.args.fieldName, tt.args.tenant)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %v, received %v", tt.exp, rcv)
			}
		})
	}
}

func TestFiltersgetFieldValueDataProvider(t *testing.T) {
	fS := FilterS{}

	type args struct {
		initialDP utils.DataProvider
		fieldName string
		tenant    string
	}

	tests := []struct {
		name string
		args args
		exp  utils.DataProvider
		err  string
	}{
		{
			name: "",
			args: args{
				initialDP: utils.MapStorage{},
				fieldName: "~*accounts",
				tenant:    "test",
			},
			exp: nil,
			err: "invalid fieldname <~*accounts>",
		},
		{
			name: "",
			args: args{
				initialDP: utils.MapStorage{},
				fieldName: "~*resources",
				tenant:    "test",
			},
			exp: nil,
			err: "invalid fieldname <~*resources>",
		},
		{
			name: "",
			args: args{
				initialDP: utils.MapStorage{},
				fieldName: "~*stats",
				tenant:    "test",
			},
			exp: nil,
			err: "invalid fieldname <~*stats>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := fS.getFieldValueDataProvider(tt.args.initialDP, tt.args.fieldName, tt.args.tenant)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %v, received %v", tt.exp, rcv)
			}
		})
	}
}

func TestFiltersPassNIString(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.Destination: "1002",
			utils.Category:    "call",
		},
	}

	tests := []struct {
		name string
		fltr *FilterRule
		want bool
	}{
		{
			name: "NIStringMatch",
			fltr: &FilterRule{
				Type:    utils.MetaNIString,
				Element: "~*req.Category",
				Values:  []string{"call"},
			},
			want: true,
		},
		{
			name: "NIStringNotMatch",
			fltr: &FilterRule{
				Type:    utils.MetaNIString,
				Element: "~*req.Category",
				Values:  []string{"premium"},
			},
			want: false,
		},
		{
			name: "NIStringMultipleValues",
			fltr: &FilterRule{
				Type:    utils.MetaNIString,
				Element: "~*req.Destination",
				Values:  []string{"1001", "1002", "1003"},
			},
			want: true,
		},
	}
	var fS *FilterS
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fltr.CompileValues(); err != nil {
				t.Fatal(err)
			}
			fieldNameDP, err := fS.getFieldNameDataProvider(ev, tt.fltr.Element, utils.EmptyString)
			if err != nil {
				t.Fatal(err)
			}
			fieldValuesDP, err := fS.getFieldValuesDataProviders(ev, tt.fltr.Values, utils.EmptyString)
			if err != nil {
				t.Fatal(err)
			}
			got, gotErr := tt.fltr.Pass(fieldNameDP, fieldValuesDP)
			if gotErr != nil {
				t.Errorf("Pass() failed: %v", gotErr)
				return
			}
			if got != tt.want {
				t.Errorf("Pass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFiltersPassNISuffix(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.Account:     "1001",
			utils.Destination: "+4985837291",
			utils.Category:    "call",
		},
	}

	tests := []struct {
		name string
		fltr *FilterRule
		want bool
	}{
		{
			name: "MatchingNISuffix",
			fltr: &FilterRule{
				Type:    utils.MetaNISuffix,
				Element: "~*req.Destination",
				Values:  []string{"91"},
			},
			want: true,
		},
		{
			name: "NonMatchingSuffix",
			fltr: &FilterRule{
				Type:    utils.MetaNISuffix,
				Element: "~*req.Account",
				Values:  []string{"02"},
			},
			want: false,
		},
		{
			name: "NISuffixPassMultipleValues",
			fltr: &FilterRule{
				Type:    utils.MetaNISuffix,
				Element: "~*req.Destination",
				Values:  []string{"3", "91", "22"},
			},
			want: true,
		},
	}

	var fS *FilterS
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fltr.CompileValues(); err != nil {
				t.Fatal(err)
			}
			fieldNameDP, err := fS.getFieldNameDataProvider(ev, tt.fltr.Element, utils.EmptyString)
			if err != nil {
				t.Fatal(err)
			}
			fieldValuesDP, err := fS.getFieldValuesDataProviders(ev, tt.fltr.Values, utils.EmptyString)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tt.fltr.Pass(fieldNameDP, fieldValuesDP)
			if err != nil {
				t.Errorf("Pass() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Pass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFiltersPassNIExists(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.Account:     "1001",
			utils.Destination: "1002",
			utils.Category:    "call",
		},
	}

	tests := []struct {
		name string
		fltr *FilterRule
		want bool
	}{
		{
			name: "NIExistsField",
			fltr: &FilterRule{
				Type:    utils.MetaNIExists,
				Element: "~*req.Category",
				Values:  []string{},
			},
			want: true,
		},
		{
			name: "NonNIExistentField",
			fltr: &FilterRule{
				Type:    utils.MetaNIExists,
				Element: "~*req.NonExistentField",
				Values:  []string{},
			},
			want: false,
		},
		{
			name: "NIExistsAccount",
			fltr: &FilterRule{
				Type:    utils.MetaNIExists,
				Element: "~*req.Account",
				Values:  []string{},
			},
			want: true,
		},
		{
			name: "NIExistsDestination",
			fltr: &FilterRule{
				Type:    utils.MetaNIExists,
				Element: "~*req.Destination",
				Values:  []string{},
			},
			want: true,
		},
	}

	var fS *FilterS
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fltr.CompileValues(); err != nil {
				t.Fatal(err)
			}

			fieldNameDP, err := fS.getFieldNameDataProvider(ev, tt.fltr.Element, utils.EmptyString)
			if err != nil {
				t.Fatal(err)
			}
			fieldValuesDP, err := fS.getFieldValuesDataProviders(ev, tt.fltr.Values, utils.EmptyString)
			if err != nil {
				t.Fatal(err)
			}

			got, err := tt.fltr.Pass(fieldNameDP, fieldValuesDP)
			if err != nil {
				t.Errorf("Pass() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Pass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFiltersPassNIPrefix(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.Account:     "1001",
			utils.Destination: "+4985837291",
			utils.Category:    "call",
		},
	}

	tests := []struct {
		name string
		fltr *FilterRule
		want bool
	}{
		{
			name: "MatchingNIPrefix",
			fltr: &FilterRule{
				Type:    utils.MetaNIPrefix,
				Element: "~*req.Destination",
				Values:  []string{"+49"},
			},
			want: true,
		},
		{
			name: "NonMatchingNIPrefix",
			fltr: &FilterRule{
				Type:    utils.MetaNIPrefix,
				Element: "~*req.Account",
				Values:  []string{"20"},
			},
			want: false,
		},
		{
			name: "NIPrefixPassMultipleValues",
			fltr: &FilterRule{
				Type:    utils.MetaNIPrefix,
				Element: "~*req.Destination",
				Values:  []string{"+49", "+21", "+35"},
			},
			want: true,
		},
	}

	var fS *FilterS
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fltr.CompileValues(); err != nil {
				t.Fatal(err)
			}
			fieldNameDP, err := fS.getFieldNameDataProvider(ev, tt.fltr.Element, utils.EmptyString)
			if err != nil {
				t.Fatal(err)
			}
			fieldValuesDP, err := fS.getFieldValuesDataProviders(ev, tt.fltr.Values, utils.EmptyString)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tt.fltr.Pass(fieldNameDP, fieldValuesDP)
			if err != nil {
				t.Errorf("Pass() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Pass() = %v, want %v", got, tt.want)
			}
		})
	}
}
