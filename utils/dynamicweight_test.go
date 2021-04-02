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
	eDws := DynamicWeights{
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
		t.Errorf("expecting: %+v, received: %+v", ToJSON(eDws), ToJSON(dws))
	}

	///
	dwsStr = "fltr1&fltr2;20;30;fltr3;30;fltr4&fltr5;50"
	expected := "invalid DynamicWeight format for string <fltr1&fltr2;20;30;fltr3;30;fltr4&fltr5;50>"
	if _, err := NewDynamicWeightsFromString(dwsStr, ";", "&"); err == nil || err.Error() != expected {
		t.Errorf("expecting: %+v, received: %+v", expected, err)
	}

	eDws = DynamicWeights{
		{
			FilterIDs: []string{"fltr1", "fltr2"},
			Weight:    20.0,
		},
		{
			Weight: 30.0,
		},
		{
			FilterIDs: []string{"fltr4", "fltr5"},
			Weight:    50.0,
		},
	}
	dwsStr = "fltr1&fltr2;20;;30;fltr4&fltr5;50"
	if dws, err := NewDynamicWeightsFromString(dwsStr, ";", "&"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDws[1], dws[1]) {
		t.Errorf("expecting: %+v, received: %+v", eDws[1], dws[1])
	}
	eDws = DynamicWeights{
		{
			Weight: 20.0,
		},
	}
	dwsStr = ";20"
	if dws, err := NewDynamicWeightsFromString(dwsStr, ";", "&"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDws, dws) {
		t.Errorf("expecting: %+v, received: %+v", eDws, dws)
	}
	eDws = DynamicWeights{
		{
			Weight: 0.0,
		},
	}
	dwsStr = ";"
	if dws, err := NewDynamicWeightsFromString(dwsStr, ";", "&"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDws, dws) {
		t.Errorf("expecting: %+v, received: %+v", eDws, dws)
	}

	dwsStr = "fltr1&fltr2;not_a_float64"
	expected = "invalid Weight <not_a_float64> in string: <fltr1&fltr2;not_a_float64>"
	if _, err := NewDynamicWeightsFromString(dwsStr, ";", "&"); err == nil || err.Error() != expected {
		t.Errorf("expecting: %+v, received: %+v", expected, err)
	}
}

func TestDynamicWeightString(t *testing.T) {
	dynWeigh := DynamicWeights{}
	if rcv := dynWeigh.String(";", "&"); len(rcv) != 0 {
		t.Errorf("Expected empty slice")
	}

	expected := "fltr1&fltr2;10"
	dynWeigh = DynamicWeights{
		{
			FilterIDs: []string{"fltr1", "fltr2"},
			Weight:    10,
		},
	}
	if rcv := dynWeigh.String(";", "&"); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestCloneDynamicWeights(t *testing.T) {
	var dynWeigh DynamicWeights
	if rcv := dynWeigh.Clone(); len(rcv) != 0 {
		t.Errorf("Expected empty slice")
	}

	dynWeigh = DynamicWeights{
		{
			FilterIDs: []string{"fltr1", "fltr2"},
			Weight:    10,
		},
	}
	if rcv := dynWeigh.Clone(); !reflect.DeepEqual(dynWeigh, rcv) {
		t.Errorf("Expected %+v, received %+v", dynWeigh, rcv)
	}
}
