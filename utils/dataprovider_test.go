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

func TestDPDynamicInterface(t *testing.T) {
	nm := NavigableMap2{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	var expected interface{} = "Field5[1]"
	if rply, err := DPDynamicInterface("Field5[1]", nm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}

	expected = 101
	if rply, err := DPDynamicInterface("~Field5[1]", nm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %v ,received: %v", expected, rply)
	}

}

func TestDPDynamicString(t *testing.T) {
	nm := NavigableMap2{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	var expected interface{} = "Field5[1]"
	if rply, err := DPDynamicString("Field5[1]", nm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}

	expected = "101"
	if rply, err := DPDynamicString("~Field5[1]", nm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %v ,received: %v", expected, rply)
	}

}

func TestAppendNavMapVal(t *testing.T) {
	nm := NavigableMap2{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	expected := NavigableMap2{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101), NewNMData(18)},
	}
	if err := AppendNavMapVal(nm, PathItems{{Field: "Field5"}}, NewNMData(18)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, nm) {
		t.Errorf("Expected %v ,received: %v", expected, nm)
	}
	if err := AppendNavMapVal(nm, nil, NewNMData(18)); err != ErrWrongPath {
		t.Error(err)
	}
}

func TestComposeNavMapVal(t *testing.T) {
	nm := NavigableMap2{
		"Field4": &NMSlice{},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	if err := ComposeNavMapVal(nm, nil, NewNMData(18)); err != ErrWrongPath {
		t.Error(err)
	}
	if err := ComposeNavMapVal(nm, PathItems{{Field: "Field4"}}, NewNMData(18)); err != ErrWrongPath {
		t.Error(err)
	}
	expected := NavigableMap2{
		"Field4": &NMSlice{},
		"Field5": &NMSlice{NewNMData(10), NewNMData("10118")},
	}
	if err := ComposeNavMapVal(nm, PathItems{{Field: "Field5"}}, NewNMData(18)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, nm) {
		t.Errorf("Expected %v ,received: %v", expected, nm)
	}

	expected = NavigableMap2{
		"Field4": &NMSlice{},
		"Field5": &NMSlice{NewNMData(10), NewNMData("10118")},
		"Field6": &NMSlice{NewNMData(10)},
	}
	if err := ComposeNavMapVal(nm, PathItems{{Field: "Field6"}}, NewNMData(10)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, nm) {
		t.Errorf("Expected %v ,received: %v", expected, nm)
	}

	nm = NavigableMap2{
		"Field4": NewNMData(1),
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	if err := ComposeNavMapVal(nm, PathItems{{Field: "Field4"}}, NewNMData(10)); err != ErrNotImplemented {
		t.Error(err)
	}

	if err := ComposeNavMapVal(nm, PathItems{{Field: "Field5"}}, &mockNMInterface{data: 10}); err != ErrNotImplemented {
		t.Error(err)
	}
}

// mock NMInterface structure

type mockNMInterface struct{ data interface{} }

func (nmi *mockNMInterface) String() string {
	return IfaceAsString(nmi.data)
}

// Interface returns the wraped interface
func (nmi *mockNMInterface) Interface() interface{} {
	return nmi.data
}

// Field not implemented only used in order to implement the NM interface
func (nmi *mockNMInterface) Field(path PathItems) (val NMInterface, err error) {
	return nil, ErrNotImplemented
}

// Set not implemented
func (nmi *mockNMInterface) Set(path PathItems, val NMInterface) (err error) {
	return ErrNotImplemented
}

// Remove not implemented only used in order to implement the NM interface
func (nmi *mockNMInterface) Remove(path PathItems) (err error) {
	return ErrNotImplemented
}

// Type returns the type of the NM interface
func (nmi *mockNMInterface) Type() NMType {
	return NMDataType
}

// Empty returns true if the NM is empty(no data)
func (nmi *mockNMInterface) Empty() bool {
	return nmi == nil || nmi.data == nil
}

// GetField not implemented only used in order to implement the NM interface
func (nmi *mockNMInterface) GetField(path PathItem) (val NMInterface, err error) {
	return nil, ErrNotImplemented
}

// SetField not implemented
func (nmi *mockNMInterface) SetField(path PathItem, val NMInterface) (err error) {
	return ErrNotImplemented
}

// Len not implemented only used in order to implement the NM interface
func (nmi *mockNMInterface) Len() int {
	return 0
}
