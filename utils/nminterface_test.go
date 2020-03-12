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
	"testing"
)

func TestNewNMInterface(t *testing.T) {
	var nm NM = NewNMInterface("1001")
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
	if nm.String() != "1001" {
		t.Errorf("Expected \"1001\" ,received: %q", nm.String())
	}
	if nm.Interface() != "1001" {
		t.Errorf("Expected \"1001\" ,received: %q", nm.Interface())
	}
	if nm.Type() != NMInterfaceType {
		t.Errorf("Expected %v ,received: %v", NMInterfaceType, nm.Type())
	}

	if _, err := nm.Field(nil); err != ErrNotImplemented {
		t.Error(err)
	}
	if err := nm.Set(PathItems{{}}, nil); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(nil); err != ErrNotImplemented {
		t.Error(err)
	}
}
