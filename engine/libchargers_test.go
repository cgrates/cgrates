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

	"github.com/cgrates/cgrates/utils"
)

func TestChargerProfileSet(t *testing.T) {
	cp := ChargerProfile{}
	exp := ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "ID",
		FilterIDs:    []string{"fltr1", "*string:~*req.Account:1001"},
		Weight:       10,
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"attr1"},
	}
	if err := cp.Set([]string{}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := cp.Set([]string{"NotAField"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := cp.Set([]string{"NotAField", "1"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := cp.Set([]string{utils.Tenant}, "cgrates.org", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{utils.ID}, "ID", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{utils.FilterIDs}, "fltr1;*string:~*req.Account:1001", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{utils.Weight}, 10, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{utils.RunID}, utils.MetaDefault, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{utils.AttributeIDs}, "attr1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, cp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(cp))
	}
}

func TestChargerProfilesSort(t *testing.T) {
	cp := ChargerProfiles{{}, {Weight: 10}}
	exp := ChargerProfiles{{Weight: 10}, {}}
	cp.Sort()
	if !reflect.DeepEqual(exp, cp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(cp))
	}
}

func TestChargerProfileAsInterface(t *testing.T) {
	cp := ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "ID",
		FilterIDs:    []string{"fltr1", "*string:~*req.Account:1001"},
		Weight:       10,
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"attr1"},
	}
	if _, err := cp.FieldAsInterface(nil); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := cp.FieldAsInterface([]string{"field"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := cp.FieldAsInterface([]string{"field", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := cp.FieldAsInterface([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := cp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := cp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{utils.Weight}); err != nil {
		t.Fatal(err)
	} else if exp := cp.Weight; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{utils.RunID}); err != nil {
		t.Fatal(err)
	} else if exp := cp.RunID; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{utils.AttributeIDs}); err != nil {
		t.Fatal(err)
	} else if exp := cp.AttributeIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := cp.FieldAsInterface([]string{utils.AttributeIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := cp.AttributeIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := cp.FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := cp.FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := cp.String(), utils.ToJSON(cp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

}

func TestChargerProfileMerge(t *testing.T) {
	dp := &ChargerProfile{}
	exp := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "ID",
		FilterIDs:    []string{"fltr1", "*string:~*req.Account:1001"},
		Weight:       10,
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"attr1"},
	}
	if dp.Merge(&ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "ID",
		FilterIDs:    []string{"fltr1", "*string:~*req.Account:1001"},
		Weight:       10,
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"attr1"},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}
