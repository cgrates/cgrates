// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
