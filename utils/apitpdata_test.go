/*
Real-time Charging System for Telecom & ISP environments
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

func TestNewDTCSFromRPKey(t *testing.T) {
	rpKey := "*out:tenant12:call:dan12"
	if dtcs, err := NewDTCSFromRPKey(rpKey); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dtcs, &DirectionTenantCategorySubject{"*out", "tenant12", "call", "dan12"}) {
		t.Error("Received: ", dtcs)
	}
}
