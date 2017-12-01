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

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/utils"
)

func TestReqFilterPassString(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf := &RequestFilter{Type: MetaString, FieldName: "Category", Values: []string{"call"}}
	if passes, err := rf.passString(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &RequestFilter{Type: MetaString, FieldName: "Category", Values: []string{"cal"}}
	if passes, err := rf.passString(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
}

func TestReqFilterPassStringPrefix(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf := &RequestFilter{Type: MetaStringPrefix, FieldName: "Category", Values: []string{"call"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "Category", Values: []string{"premium"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"+49"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"+499"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "navigation", Values: []string{"off"}}
	if passes, err := rf.passStringPrefix(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringPrefix(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
}

func TestReqFilterPassRSRFields(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf, err := NewRequestFilter(MetaRSRFields, "", []string{"Tenant(~^cgr.*\\.org$)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSRFields(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewRequestFilter(MetaRSRFields, "", []string{"navigation(on)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSRFields(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	rf, err = NewRequestFilter(MetaRSRFields, "", []string{"navigation(off)"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passRSRFields(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
}

func TestReqFilterPassDestinations(t *testing.T) {
	cache.Set(utils.REVERSE_DESTINATION_PREFIX+"+49", []string{"DE", "EU_LANDLINE"}, true, "")
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf, err := NewRequestFilter(MetaDestinations, "Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = NewRequestFilter(MetaDestinations, "Destination", []string{"RO"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
}

func TestReqFilterPassGreaterThan(t *testing.T) {
	rf, err := NewRequestFilter(MetaLessThan, "ASR", []string{"40"})
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
	rf, err = NewRequestFilter(MetaLessOrEqual, "ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewRequestFilter(MetaGreaterOrEqual, "ASR", []string{"40"})
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
	rf, err = NewRequestFilter(MetaGreaterOrEqual, "ACD", []string{"1m50s"})
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

func TestReqFilterNewRequestFilter(t *testing.T) {
	rf, err := NewRequestFilter(MetaString, "MetaString", []string{"String"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf := &RequestFilter{Type: MetaString, FieldName: "MetaString", Values: []string{"String"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewRequestFilter(MetaStringPrefix, "MetaStringPrefix", []string{"stringPrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &RequestFilter{Type: MetaStringPrefix, FieldName: "MetaStringPrefix", Values: []string{"stringPrefix"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewRequestFilter(MetaTimings, "MetaTimings", []string{""})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &RequestFilter{Type: MetaTimings, FieldName: "MetaTimings", Values: []string{""}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewRequestFilter(MetaDestinations, "MetaDestinations", []string{""})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &RequestFilter{Type: MetaDestinations, FieldName: "MetaDestinations", Values: []string{""}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewRequestFilter(MetaLessThan, "MetaLessThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &RequestFilter{Type: MetaLessThan, FieldName: "MetaLessThan", Values: []string{"20"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewRequestFilter(MetaLessOrEqual, "MetaLessOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &RequestFilter{Type: MetaLessOrEqual, FieldName: "MetaLessOrEqual", Values: []string{"20"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewRequestFilter(MetaGreaterThan, "MetaGreaterThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &RequestFilter{Type: MetaGreaterThan, FieldName: "MetaGreaterThan", Values: []string{"20"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewRequestFilter(MetaGreaterOrEqual, "MetaGreaterOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &RequestFilter{Type: MetaGreaterOrEqual, FieldName: "MetaGreaterOrEqual", Values: []string{"20"}}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
}
