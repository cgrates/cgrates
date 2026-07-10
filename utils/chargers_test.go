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

func TestChargerProfileClone(t *testing.T) {
	tests := []struct {
		name string
		cp   *ChargerProfile
	}{
		{
			name: "Complete ChargerProfile",
			cp: &ChargerProfile{
				Tenant:    "cgrates.org",
				ID:        "profileID",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Weights: DynamicWeights{
					{
						Weight: 5,
					},
				},
				RunID:        MetaDefault,
				AttributeIDs: []string{"attr2"},
				Blockers:     DynamicBlockers{},
			},
		},
		{
			name: "Nil fields",
			cp: &ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "profileID",
				FilterIDs:    nil,
				Weights:      nil,
				RunID:        "",
				AttributeIDs: nil,
				Blockers:     nil,
			},
		},
		{
			name: "Nil ChargerProfile",
			cp:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cp.Clone()
			if !reflect.DeepEqual(got, tt.cp) {
				t.Errorf("Expected %v, recieved %v", tt.cp, got)
			}

			if tt.cp != nil && tt.cp == got {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cp.CacheClone()
			if !reflect.DeepEqual(got, tt.cp) {
				t.Errorf("Expected %v, recieved %v", tt.cp, got)
			}
		})
	}
}

func TestChargerProfileAsMapStringInterface(t *testing.T) {
	tests := []struct {
		name string
		cp   *ChargerProfile
		want map[string]any
	}{
		{
			name: "ChargerProfile with values",
			cp: &ChargerProfile{
				Tenant:    "cgrates.org",
				ID:        "id1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: DynamicWeights{
					{
						Weight: 10,
					},
				},
				RunID:        MetaDefault,
				AttributeIDs: []string{"attr1"},
				Blockers:     DynamicBlockers{},
			},
			want: map[string]any{
				Tenant:    "cgrates.org",
				ID:        "id1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: DynamicWeights{
					{
						Weight: 10,
					},
				},
				RunID:        MetaDefault,
				AttributeIDs: []string{"attr1"},
				Blockers:     DynamicBlockers{},
			},
		},
		{
			name: "Nil ChargerProfile",
			cp:   nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cp.AsMapStringInterface()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %#+v, recieved %#+v", tt.want, got)
			}
		})
	}
}

func TestMapStringInterfaceToChargerProfile(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]any
		want *ChargerProfile
	}{
		{
			m: map[string]any{
				Tenant:    "cgrates.org",
				ID:        "id1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: DynamicWeights{
					{
						Weight: 10,
					},
				},
				RunID:        MetaDefault,
				AttributeIDs: []string{"attr1"},
				Blockers:     DynamicBlockers{},
			},
			want: &ChargerProfile{
				Tenant:    "cgrates.org",
				ID:        "id1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: DynamicWeights{
					{
						Weight: 10,
					},
				},
				RunID:        MetaDefault,
				AttributeIDs: []string{"attr1"},
				Blockers:     DynamicBlockers{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapStringInterfaceToChargerProfile(tt.m)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %#+v, recieved %#+v", tt.want, got)
			}
		})
	}
}

func TestChargerProfileTenantID(t *testing.T) {
	cp := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "id1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		RunID:        MetaDefault,
		AttributeIDs: []string{"attr1"},
		Blockers:     DynamicBlockers{},
	}
	want := "cgrates.org:id1"

	got := cp.TenantID()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected %v, recieved %v", want, got)
	}
}
