/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package utils

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
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
	expected := "invalid Weight <not_a_float64> in string: <fltr1&fltr2;not_a_float64>"
	if _, err := NewDynamicWeightsFromString(dwsStr, ";", "&"); err == nil || err.Error() != expected {
		t.Errorf("expecting: %+v, received: %+v", expected, err)
	}

	exp := DynamicWeights{{}}
	if rcv, err := NewDynamicWeightsFromString("", "", ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v, Received %v", ToJSON(exp), ToJSON(rcv))
	}

	exps := DynamicWeights{
		{
			FilterIDs: nil,
			Weight:    5,
		},
	}

	expErr := "invalid DynamicWeight format for string <;5;>"
	if rcv, err := NewDynamicWeightsFromString(";5", ";", "&"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exps, rcv) {
		t.Errorf("expecting: %+v, received: %+v", ToJSON(exps), ToJSON(rcv))
	} else if _, err := NewDynamicWeightsFromString(";5;", ";", "&"); err == nil || err.Error() != expErr {
		t.Errorf("Expected <%+v> Received <%+v>", expErr, err)
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

func TestDynamicWeightEquals(t *testing.T) {
	//Test the case where one of the fields is nil
	dW := &DynamicWeight{
		FilterIDs: nil,
		Weight:    10,
	}
	dnWg := &DynamicWeight{
		FilterIDs: []string{"fltr1", "fltr2"},
		Weight:    10,
	}
	if rcv := dW.Equals(dnWg); rcv {
		t.Error("FilterIDs should not match")
	}

	//Test the case where filters don't match
	dW.FilterIDs = []string{"fltr1", "fltr3"}
	if rcv := dW.Equals(dnWg); rcv {
		t.Error("FilterIDs should not match")
	}
}

func TestNewBalanceDynamicWeightsFromString(t *testing.T) {
	dwsStr := "fltr1&fltr2;20;fltr3;30"
	eDws := DynamicWeights{
		{FilterIDs: []string{"fltr1", "fltr2"}, Weight: 20.0},
		{FilterIDs: []string{"fltr3"}, Weight: 30.0},
	}
	if dws, err := NewBalanceDynamicWeightsFromString(dwsStr, ";", "&"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDws, dws) {
		t.Errorf("expecting: %+v, received: %+v", eDws, dws)
	}

	if dws, err := NewBalanceDynamicWeightsFromString("", "", ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(DynamicWeights{{}}, dws) {
		t.Errorf("expecting: %+v, received: %+v", DynamicWeights{{}}, dws)
	}

	wantAsc := float64(time.Now().Unix())
	if dws, err := NewBalanceDynamicWeightsFromString("fltr1;"+MetaTimeAsc, ";", "&"); err != nil {
		t.Error(err)
	} else if len(dws) != 1 || !reflect.DeepEqual([]string{"fltr1"}, dws[0].FilterIDs) {
		t.Errorf("unexpected DynamicWeights: %+v", dws)
	} else if dws[0].Weight-wantAsc > 2 {
		t.Errorf("Weight = %v, want ~%v", dws[0].Weight, wantAsc)
	}

	wantDesc := float64(time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC).Unix() - time.Now().Unix())
	if dws, err := NewBalanceDynamicWeightsFromString(";"+MetaTimeDesc, ";", "&"); err != nil {
		t.Error(err)
	} else if len(dws) != 1 || dws[0].FilterIDs != nil {
		t.Errorf("unexpected DynamicWeights: %+v", dws)
	} else if math.Abs(dws[0].Weight-wantDesc) > 2 {
		t.Errorf("Weight = %v, want ~%v", dws[0].Weight, wantDesc)
	}
	expErr := "invalid Weight "
	if _, err := NewBalanceDynamicWeightsFromString(";abc", ";", "&"); err == nil || !strings.HasPrefix(err.Error(), expErr) {
		t.Errorf("expected to receive %s, got %v", expErr, err)
	}
}
