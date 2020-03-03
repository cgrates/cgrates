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

import "testing"

func TestNavigableMap(t *testing.T) {
	var nm NM = NavigableMap2{
		"Field1": NewNMInterface(10),
		"Field2": &NMSlice{
			NewNMInterface("1001"),
			NavigableMap2{
				"Account": &NMSlice{NewNMInterface(10), NewNMInterface(11)},
			},
		},
		"Field3": NavigableMap2{
			"Field4": NavigableMap2{
				"Field5": NewNMInterface(5),
			},
		},
	}
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
	// expStr := `{"Field1":10,"Field2":["1001",{"Account":[10,11]}],"Field3":{"Field4":{"Field5":5}}}`
	// if nm.String() != expStr {
	// 	t.Errorf("Expected %q ,received: %q", expStr, nm.String())
	// }
	if nm.Type() != NMMapType {
		t.Errorf("Expected %v ,received: %v", NMInterfaceType, nm.Type())
	}
	path := []*PathItem{{Field: "Field1"}}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val.Interface())
	}

	path = []*PathItem{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5"}}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != 5 {
		t.Errorf("Expected %q ,received: %q", 5, val.Interface())
	}

	path = []*PathItem{{Field: "Field2", Index: IntPointer(2)}}
	if err := nm.Set(path, NewNMInterface("500")); err != nil {
		t.Error(err)
	}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Interface())
	}

	path = []*PathItem{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account"}}
	if err := nm.Set(path, NewNMInterface("5")); err != nil {
		t.Error(err)
	}
	path = []*PathItem{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account"}}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "5" {
		t.Errorf("Expected %q ,received: %q", "5", val.Interface())
	}
	path = []*PathItem{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(0)}}
	if _, err := nm.Field(path); err != ErrNotFound {
		t.Error(err)
	}
}
