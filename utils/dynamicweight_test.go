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

func TestNewDynamicWeightsFromString(t *testing.T) {
	eDws := []*DynamicWeight{
		{
			FilterIDs: []string{"fltr1", "fltr2"},
			Weight:    20.0,
		},
		{
			FilterIDs: []string{"fltr3"},
			Weight:    30.0,
		},
		{
			FilterIDs: []string{"fltr4", "fltr5"},
			Weight:    50.0,
		},
	}
	dwsStr := "fltr1&fltr2;20;fltr3;30;fltr4&fltr5;50"
	if dws, err := NewDynamicWeightsFromString(dwsStr, ";", "&"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDws, dws) {
		t.Errorf("expecting: %+v, received: %+v", eDws, dws)
	}

}
