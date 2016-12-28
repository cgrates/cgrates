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

func TestNewDTCSFromRPKey(t *testing.T) {
	rpKey := "*out:tenant12:call:dan12"
	if dtcs, err := NewDTCSFromRPKey(rpKey); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dtcs, &DirectionTenantCategorySubject{"*out", "tenant12", "call", "dan12"}) {
		t.Error("Received: ", dtcs)
	}
}

func TestPaginatorPaginateStringSlice(t *testing.T) {
	eOut := []string{"1", "2", "3", "4"}
	pgnt := new(Paginator)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"1", "2", "3"}
	pgnt.Limit = IntPointer(3)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"2", "3", "4"}
	pgnt.Offset = IntPointer(1)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{}
	pgnt.Offset = IntPointer(4)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"3"}
	pgnt.Offset = IntPointer(2)
	pgnt.Limit = IntPointer(1)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}
