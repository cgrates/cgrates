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

	"github.com/cgrates/cgrates/utils"
)

func TestExternalAttributeProfileAsAttributeProfile(t *testing.T) {
	extAttr := &ExternalAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Context:   "con1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Substitute: []*AttributeSubstitute{
			&AttributeSubstitute{
				FieldName: "FL1",
				Initial:   "In1",
				Alias:     "Al1",
				Append:    true,
			},
		},
		Weight: 20,
	}
	attrMap := make(map[string]map[string]*AttributeSubstitute)
	attrMap["FL1"] = make(map[string]*AttributeSubstitute)
	attrMap["FL1"]["In1"] = &AttributeSubstitute{
		FieldName: "FL1",
		Initial:   "In1",
		Alias:     "Al1",
		Append:    true,
	}
	expected := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Context:   "con1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Substitutes: attrMap,
		Weight:      20,
	}

	rcv := extAttr.AsAttributeProfile()
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestNewExternalAttributeProfileFromAttributeProfile(t *testing.T) {
	attrMap := make(map[string]map[string]*AttributeSubstitute)
	attrMap["FL1"] = make(map[string]*AttributeSubstitute)
	attrMap["FL1"]["In1"] = &AttributeSubstitute{
		FieldName: "FL1",
		Initial:   "In1",
		Alias:     "Al1",
		Append:    true,
	}
	attrPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Context:   "con1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Substitutes: attrMap,
		Weight:      20,
	}

	expected := &ExternalAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Context:   "con1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Substitute: []*AttributeSubstitute{
			&AttributeSubstitute{
				FieldName: "FL1",
				Initial:   "In1",
				Alias:     "Al1",
				Append:    true,
			},
		},
		Weight: 20,
	}

	rcv := NewExternalAttributeProfileFromAttributeProfile(attrPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}
