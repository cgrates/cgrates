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

func TestNewBlockersFromString(t *testing.T) {
	blkrs := DynamicBlockers{
		{
			FilterIDs: []string{"*string:~*opts.*cost:0"},
			Blocker:   false,
		},
		{
			FilterIDs: []string{"*suffix:~*req.Destination:+4432", "eq:~*opts.*usage:10s"},
			Blocker:   false,
		},
		{
			FilterIDs: []string{"*notstring:~*req.RequestType:*prepaid"},
			Blocker:   true,
		},
		{
			Blocker: false,
		},
	}
	blkrsStr := "*string:~*opts.*cost:0;false;*suffix:~*req.Destination:+4432&eq:~*opts.*usage:10s;false;*notstring:~*req.RequestType:*prepaid;true;;false"
	if rcv, err := NewDynamicBlockersFromString(blkrsStr, InfieldSep, ANDSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, blkrs) {
		t.Errorf("Expected %v \n received %v", ToJSON(blkrs), ToJSON(rcv))
	}

}

func TestNewBlockersFromString2(t *testing.T) {
	blkrs := DynamicBlockers{
		{
			FilterIDs: []string{"*string:~*opts.*cost:0"},
			Blocker:   false,
		},
		{},
	}
	blkrsStr := "*string:~*opts.*cost:0;false;;"
	if rcv, err := NewDynamicBlockersFromString(blkrsStr, InfieldSep, ANDSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, blkrs) {
		t.Errorf("Expected %+v \n ,received %+v", ToJSON(blkrs), ToJSON(rcv))
	}
}

func TestNewBlockersFromStringErrSeparator(t *testing.T) {
	blkrsStr := "*string:~*opts.*cost:0;false;;;"
	exp := "invalid DynamicBlocker format for string <*string:~*opts.*cost:0;false;;;>"
	if _, err := NewDynamicBlockersFromString(blkrsStr, InfieldSep, ANDSep); err.Error() != exp {
		t.Errorf("Expected %s \n received %s", exp, err.Error())
	}
}

func TestNewBlockersFromStringFormatBool(t *testing.T) {
	blkrsStr := "*string:~*opts.*cost:0;tttrrruuue"
	exp := "cannot convert bool with value: <tttrrruuue> into Blocker"
	if _, err := NewDynamicBlockersFromString(blkrsStr, InfieldSep, ANDSep); err.Error() != exp {
		t.Errorf("Expected %s \n received %s", exp, err.Error())
	}
}
