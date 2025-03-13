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

package utils

import (
	"reflect"
	"testing"
)

func TestChargerProfileSet(t *testing.T) {
	cp := ChargerProfile{}
	exp := ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		RunID:        MetaDefault,
		AttributeIDs: []string{"attr1"},
	}
	if err := cp.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := cp.Set([]string{"NotAField"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := cp.Set([]string{"NotAField", "1"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := cp.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{RunID}, MetaDefault, false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{AttributeIDs}, "attr1", false); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, cp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(cp))
	}
}

func TestChargerProfilesSort(t *testing.T) {
	cp := ChargerProfiles{{}, {Weights: DynamicWeights{
		{
			Weight: 20,
		},
	},
		weight: 20}}
	exp := ChargerProfiles{{Weights: DynamicWeights{
		{
			Weight: 20,
		},
	},
		weight: 20}, {}}
	cp.Sort()
	if !reflect.DeepEqual(exp, cp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(cp))
	}
}

func TestChargerProfileAsInterface(t *testing.T) {
	cp := ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		RunID:        MetaDefault,
		AttributeIDs: []string{"attr1"},
	}
	if _, err := cp.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := cp.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := cp.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := cp.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := cp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := cp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := cp.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{RunID}); err != nil {
		t.Fatal(err)
	} else if exp := cp.RunID; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{AttributeIDs}); err != nil {
		t.Fatal(err)
	} else if exp := cp.AttributeIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{AttributeIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := cp.AttributeIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := cp.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := cp.FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := cp.String(), ToJSON(cp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

}

func TestChargerProfileMerge(t *testing.T) {
	dp := &ChargerProfile{}
	exp := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		RunID:        MetaDefault,
		AttributeIDs: []string{"attr1"},
	}
	if dp.Merge(&ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		RunID:        MetaDefault,
		AttributeIDs: []string{"attr1"},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(dp))
	}
}

func TestChargerProfileSetBlockers(t *testing.T) {
	cp := &ChargerProfile{}

	exp := &ChargerProfile{
		Blockers: DynamicBlockers{
			{Blocker: true},
		},
	}

	err := cp.Set([]string{Blockers}, ";true", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !reflect.DeepEqual(exp, cp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(cp))
	}

}

func TestChargerProfileFieldAsInterfaceBlockers(t *testing.T) {

	cp := &ChargerProfile{
		Blockers: DynamicBlockers{
			{Blocker: true},
		},
	}

	rcv, err := cp.FieldAsInterface([]string{Blockers})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !reflect.DeepEqual(rcv, cp.Blockers) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(cp), ToJSON(rcv))
	}

}
