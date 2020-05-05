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

func TestNewNMInterface(t *testing.T) {
	nm2 := NewNMData("1001")
	expectednm := &NMData{data: "1001"}
	if !reflect.DeepEqual(expectednm, nm2) {
		t.Errorf("Expected %v ,received: %v", ToJSON(expectednm), ToJSON(nm2))
	}
	var nm NMInterface = nm2
	expected := "1001"
	if rply := nm.Interface(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
	if nm2.data != expected {
		t.Errorf("Expected %q ,received: %q", expected, nm2.data)
	}
}

func TestNMDataLen(t *testing.T) {
	var nm NMInterface = NewNMData("1001")
	if rply := nm.Len(); rply != 0 {
		t.Errorf("Expected 0 ,received: %v", rply)
	}
}

func TestNMDataString(t *testing.T) {
	var nm NMInterface = NewNMData("1001")
	expected := "1001"
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNMDataInterface(t *testing.T) {
	var nm NMInterface = NewNMData("1001")
	expected := "1001"
	if rply := nm.Interface(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNMDataField(t *testing.T) {
	var nm NMInterface = NewNMData("1001")
	if _, err := nm.Field(nil); err != ErrNotImplemented {
		t.Error(err)
	}
}

func TestNMDataRemove(t *testing.T) {
	var nm NMInterface = NewNMData("1001")
	if err := nm.Remove(nil); err != ErrNotImplemented {
		t.Error(err)
	}
}

func TestNMDataEmpty(t *testing.T) {
	var nm NMInterface = NewNMData("1001")
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
	nm = NewNMData(nil)
	if !nm.Empty() {
		t.Error("Expected empty type")
	}
}

func TestNMDataType(t *testing.T) {
	var nm NMInterface = NewNMData("1001")
	if nm.Type() != NMDataType {
		t.Errorf("Expected %v ,received: %v", NMDataType, nm.Type())
	}
}

func TestNMDataSet(t *testing.T) {
	var nm NMInterface = NewNMData("1001")
	if _, err := nm.Set(PathItems{{}}, nil); err != ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Set(nil, NewNMData("1002")); err != nil {
		t.Error(err)
	}
	expected := "1002"
	if rply := nm.Interface(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}
