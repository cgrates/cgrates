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
