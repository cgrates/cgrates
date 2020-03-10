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

func TestOrderedNavigableMap(t *testing.T) {
	// var err error
	onm := NewOrderedNavigableMap()

	onm.Set(PathItems{{Field: "Field1"}}, NewNMInterface(10))
	onm.Set(PathItems{{Field: "Field2", Index: IntPointer(0)}}, NewNMInterface("1001"))
	onm.Set(PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(0)}}, NewNMInterface(10))
	onm.Set(PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(1)}}, NewNMInterface(11))
	onm.Set(PathItems{{Field: "Field2", Index: IntPointer(2)}}, NewNMInterface(111))
	onm.Set(PathItems{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5"}}, NewNMInterface(5))
	var expnm NM = NavigableMap2{
		"Field1": NewNMInterface(10),
		"Field2": &NMSlice{
			NewNMInterface("1001"),
			NavigableMap2{
				"Account": &NMSlice{NewNMInterface(10), NewNMInterface(11)},
			},
			NewNMInterface(111),
		},
		"Field3": NavigableMap2{
			"Field4": NavigableMap2{
				"Field5": NewNMInterface(5),
			},
		},
	}
	if onm.Empty() {
		t.Error("Expected not empty type")
	}
	if onm.Type() != NMMapType {
		t.Errorf("Expected %v ,received: %v", NMInterfaceType, onm.Type())
	}
	if !reflect.DeepEqual(expnm, onm.nm) {
		t.Errorf("Expected %s ,received: %s", expnm, onm.nm)
	}
	expOrder := []PathItems{
		PathItems{{Field: "Field1"}},
		PathItems{{Field: "Field2", Index: IntPointer(0)}},
		PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(0)}},
		PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(1)}},
		PathItems{{Field: "Field2", Index: IntPointer(2)}},
		PathItems{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5"}},
	}
	if !reflect.DeepEqual(expOrder, onm.order) {
		t.Errorf("Expected %s ,received: %s", expOrder, onm.order)
	}

	path := PathItems{{Field: "Field2"}}
	//
	// sliceDeNM
	exp := &NMSlice{NewNMInterface("500"), NewNMInterface("502")}
	if err := onm.Set(path, exp); err != nil {
		t.Error(err)
	}
	path = PathItems{{Field: "Field2"}}
	if val, err := onm.Field(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, exp) {
		t.Errorf("Expected %q ,received: %q", exp, val.Interface())
	}

	expOrder = []PathItems{
		PathItems{{Field: "Field1"}},
		PathItems{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5"}},
		PathItems{{Field: "Field2", Index: IntPointer(0)}},
		PathItems{{Field: "Field2", Index: IntPointer(1)}},
	}
	if !reflect.DeepEqual(expOrder, onm.order) {
		t.Errorf("Expected %s ,received: %s", expOrder, onm.order)
	}

	path = PathItems{{Field: "Field2", Index: IntPointer(0)}}
	if val, err := onm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Interface())
	}

	// path = []string{"Field2[0]"}
	// if err := onm.Set(path, NewNMInterface("1002")); err != nil {
	// 	t.Error(err)
	// }
	// if val, err := onm.Field(path); err != nil {
	// 	t.Error(err)
	// } else if val.Interface() != "1002" {
	// 	t.Errorf("Expected %q ,received: %q", "1002", val.Interface())
	// }
	// fmt.Println(onm.order)
	/*
		path = []string{"Field1"}
		if val, err := onm.Field(path); err != nil {
			t.Error(err)
		} else if val.Interface() != 10 {
			t.Errorf("Expected %q ,received: %q", 10, val.Interface())
		}

		path = []string{"Field3", "Field4", "Field5"}
		if val, err := onm.Field(path); err != nil {
			t.Error(err)
		} else if val.Interface() != 5 {
			t.Errorf("Expected %q ,received: %q", 5, val.Interface())
		}
		/*
			path = []string{"Field2[2]"}
			if err := nm.Set(path, NewNMInterface("500")); err != nil {
				t.Error(err)
			}
			if val, err := nm.Field(path); err != nil {
				t.Error(err)
			} else if val.Interface() != "500" {
				t.Errorf("Expected %q ,received: %q", "500", val.Interface())
			}

			path = []string{"Field2[1]", "Account"}
			if err := nm.Set(path, NewNMInterface("5")); err != nil {
				t.Error(err)
			}
			path = []string{"Field2[1]", "Account[0]"}
			if val, err := nm.Field(path); err != nil {
				t.Error(err)
			} else if val.Interface() != "5" {
				t.Errorf("Expected %q ,received: %q", "5", val.Interface())
			}
			path = []string{"Field2[1]", "Account[1]"}
			if _, err := nm.Field(path); err != ErrNotFound {
				t.Error(err)
			}

			fmt.Println(nm)*/
}
