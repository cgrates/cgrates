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

func TestExternalAliasProfileAsAliasProfile(t *testing.T) {
	extAls := &ExternalAliasProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Aliases: []*AliasEntry{
			&AliasEntry{
				FieldName: "FL1",
				Initial:   "In1",
				Alias:     "Al1",
			},
		},
		Weight: 20,
	}
	alsMap := make(map[string]map[string]string)
	alsMap["FL1"] = make(map[string]string)
	alsMap["FL1"]["In1"] = "Al1"
	expected := &AliasProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Aliases: alsMap,
		Weight:  20,
	}

	rcv := extAls.AsAliasProfile()
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestNewExternalAliasProfileFromAliasProfile(t *testing.T) {
	alsMap := make(map[string]map[string]string)
	alsMap["FL1"] = make(map[string]string)
	alsMap["FL1"]["In1"] = "Al1"
	alsPrf := &AliasProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Aliases: alsMap,
		Weight:  20,
	}

	expected := &ExternalAliasProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Aliases: []*AliasEntry{
			&AliasEntry{
				FieldName: "FL1",
				Initial:   "In1",
				Alias:     "Al1",
			},
		},
		Weight: 20,
	}

	rcv := NewExternalAliasProfileFromAliasProfile(alsPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}
