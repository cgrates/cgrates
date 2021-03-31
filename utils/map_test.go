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
	"sort"
	"testing"
)

func TestMapStringToInt64(t *testing.T) {
	t1 := map[string]int64{"test": int64(21)}
	t2 := map[string]string{"test": "21"}
	t3, err := MapStringToInt64(t2)
	if err != nil {
		t.Error("Got Error: ", err)
	}
	if !reflect.DeepEqual(t1, t3) {
		t.Errorf("Expecting: %+v, received: %+v", t1, t3)
	}
}

func TestMapSubsystemIDsFromSlice(t *testing.T) {
	sls := []string{"*event", "*thresholds:*ids:ID1&ID2&ID3", "*thresholds:*derived_reply", "*attributes:*disabled", "*stats:*ids:ID"}
	eMp := FlagsWithParams{
		"*event":      map[string][]string{},
		"*thresholds": map[string][]string{MetaIDs: {"ID1", "ID2", "ID3"}, MetaDerivedReply: {}},
		"*attributes": map[string][]string{"*disabled": {}},
		"*stats":      map[string][]string{MetaIDs: {"ID"}},
	}
	if mp := FlagsWithParamsFromSlice(sls); !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, mp)
	}
}

func TestMapSubsystemIDsHasKey(t *testing.T) {
	sls := []string{"*event", "*thresholds:*ids:ID1&ID2&ID3", "*attributes", "*stats:*ids:ID"}
	eMp := FlagsWithParams{
		"*event":      map[string][]string{},
		"*thresholds": map[string][]string{MetaIDs: {"ID1", "ID2", "ID3"}},
		"*attributes": map[string][]string{},
		"*stats":      map[string][]string{MetaIDs: {"ID"}},
	}
	mp := FlagsWithParamsFromSlice(sls)
	if !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, mp)
	}
	if has := mp.Has("*event"); !has {
		t.Errorf("Expecting: true, received: %+v", has)
	}
	if has := mp.Has("*thresholds"); !has {
		t.Errorf("Expecting: true, received: %+v", has)
	}
	if has := mp.Has("*resources"); has {
		t.Errorf("Expecting: false, received: %+v", has)
	}
}

func TestMapSubsystemIDsGetIDs(t *testing.T) {
	sls := []string{"*event", "*thresholds:*ids:ID1&ID2&ID3", "*attributes", "*stats:*ids:ID"}
	eMp := FlagsWithParams{
		"*event":      map[string][]string{},
		"*thresholds": map[string][]string{MetaIDs: {"ID1", "ID2", "ID3"}},
		"*attributes": map[string][]string{},
		"*stats":      map[string][]string{MetaIDs: {"ID"}},
	}
	mp := FlagsWithParamsFromSlice(sls)
	if !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, mp)
	}
	eIDs := []string{"ID1", "ID2", "ID3"}
	if ids := mp.ParamsSlice("*thresholds", MetaIDs); !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eIDs, ids)
	}
	eIDs = nil
	if ids := mp.ParamsSlice("*event", MetaIDs); !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eIDs, ids)
	}
	if ids := mp.ParamsSlice("*test", MetaIDs); !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eIDs, ids)
	}
}

func TestFlagsToSlice(t *testing.T) {
	sls := []string{"*event", "*thresholds:*ids:ID1&ID2&ID3", "*attributes", "*stats:*ids:ID", "*routes:*derived_reply"}
	eMp := FlagsWithParams{
		"*event":      map[string][]string{},
		"*thresholds": map[string][]string{MetaIDs: {"ID1", "ID2", "ID3"}},
		"*attributes": map[string][]string{},
		"*stats":      map[string][]string{MetaIDs: {"ID"}},
		"*routes":     map[string][]string{MetaDerivedReply: {}},
	}
	mp := FlagsWithParamsFromSlice(sls)
	if !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, mp)
	}
	sort.Strings(sls)
	flgSls := mp.SliceFlags()
	sort.Strings(flgSls)
	if !reflect.DeepEqual(flgSls, sls) {
		t.Errorf("Expecting: %+v, received: %+v", sls, flgSls)
	}
}

func TestFlagsWithParamsGetBool(t *testing.T) {
	flagsWithParams := &FlagsWithParams{
		"test":  map[string][]string{"false": {}, "string2": {}},
		"test2": map[string][]string{"string2": {}},
		"test3": map[string][]string{"true": {}},
		"empty": map[string][]string{},
	}
	key := "notpresent"
	if rcv := flagsWithParams.GetBool(key); rcv != false {
		t.Errorf("Expecting: false, received: %+v", ToJSON(rcv))
	}
	key = "empty"
	if rcv := flagsWithParams.GetBool(key); rcv != true {
		t.Errorf("Expecting: false, received: %+v", ToJSON(rcv))
	}
	key = "test"
	if rcv := flagsWithParams.GetBool(key); rcv != false {
		t.Errorf("Expecting: false, received: %+v", ToJSON(rcv))
	}
	key = "test2"
	if rcv := flagsWithParams.GetBool(key); rcv != true {
		t.Errorf("Expecting: true, received: %+v", ToJSON(rcv))
	}
	key = "test3"
	if rcv := flagsWithParams.GetBool(key); rcv != true {
		t.Errorf("Expecting: true, received: %+v", ToJSON(rcv))
	}
}

func TestFlagsWithParamsValue(t *testing.T) {
	flagsWithParams := &FlagsWithParams{
		"test":  map[string][]string{"string2": {}},
		"empty": map[string][]string{},
	}
	key := "notpresent"
	if rcv := flagsWithParams.ParamValue(key); rcv != EmptyString {
		t.Errorf("Expecting: %q, received: %+v", EmptyString, rcv)
	}
	key = "empty"
	if rcv := flagsWithParams.ParamValue(key); rcv != EmptyString {
		t.Errorf("Expecting: %q, received: %+v", EmptyString, rcv)
	}
	key = "test"
	if rcv := flagsWithParams.ParamValue(key); rcv != "string2" {
		t.Errorf("Expecting: string2, received: %+v", rcv)
	}
}

func TestFlagParamsValue(t *testing.T) {
	flagsWithParams := &FlagParams{
		"test":  []string{"string2"},
		"empty": []string{},
	}
	key := "notpresent"
	if rcv := flagsWithParams.ParamValue(key); rcv != EmptyString {
		t.Errorf("Expecting: %q, received: %+v", EmptyString, rcv)
	}
	key = "empty"
	if rcv := flagsWithParams.ParamValue(key); rcv != EmptyString {
		t.Errorf("Expecting: %q, received: %+v", EmptyString, rcv)
	}
	key = "test"
	if rcv := flagsWithParams.ParamValue(key); rcv != "string2" {
		t.Errorf("Expecting: string2, received: %+v", rcv)
	}
}

func TestFlagParamsAdd(t *testing.T) {
	flgs := make(FlagParams)
	exp := FlagParams{
		MetaIDs: []string{"id1", "id2"},
	}
	flgs.Add([]string{MetaIDs, "id1&id2", "ignored"})
	if !reflect.DeepEqual(flgs, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, flgs)
	}
}

func TestFlagsToSlice2(t *testing.T) {
	sls := []string{"*event", "*thresholds:*ids:ID1&ID2&ID3", "*attributes", "*stats:*ids:ID", "*routes:*derived_reply", "*cdrs:*attributes", "*cdrs:*stats:ID"}
	eMp := FlagsWithParams{
		"*event":      map[string][]string{},
		"*thresholds": map[string][]string{MetaIDs: {"ID1", "ID2", "ID3"}},
		"*attributes": map[string][]string{},
		"*stats":      map[string][]string{MetaIDs: {"ID"}},
		"*routes":     map[string][]string{MetaDerivedReply: {}},
		"*cdrs":       map[string][]string{MetaAttributes: {}, MetaStats: {"ID"}},
	}
	mp := FlagsWithParamsFromSlice(sls)
	if !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(eMp), ToJSON(mp))
	}
	sort.Strings(sls)
	flgSls := mp.SliceFlags()
	sort.Strings(flgSls)
	if !reflect.DeepEqual(flgSls, sls) {
		t.Errorf("Expecting: %+v, received: %+v", sls, flgSls)
	}

	sls = []string{"*attributes", "*stats:ID"}
	flgSls = mp["*cdrs"].SliceFlags()
	sort.Strings(flgSls)
	if !reflect.DeepEqual(flgSls, sls) {
		t.Errorf("Expecting: %+v, received: %+v", sls, flgSls)
	}
}

func TestMapStringToInt64Err(t *testing.T) {
	t2 := map[string]string{"test": "a"}
	_, err := MapStringToInt64(t2)
	if err == nil {
		t.Error("Got Error: ", err)
	}
}

func TestFlagsWithParamsClone(t *testing.T) {
	fWp := FlagsWithParams{
		MetaEvent:      {},
		MetaRoutes:     nil,
		MetaThresholds: {MetaIDs: {"ID1", "ID2", "ID3"}, MetaDerivedReply: {}},
		MetaAttributes: {"*disabled": {}},
		MetaStats:      {MetaIDs: {"ID"}},
	}

	cln := fWp.Clone()
	if !reflect.DeepEqual(cln, fWp) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(fWp), ToJSON(cln))
	}
	cln[MetaDispatchers] = FlagParams{}
	if _, has := fWp[MetaDispatchers]; has {
		t.Errorf("Expected clone to not modify the cloned")
	}
	cln[MetaThresholds][MetaIDs][0] = ""
	if fWp[MetaThresholds][MetaIDs][0] != "ID1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	fWp = nil
	cln = fWp.Clone()
	if !reflect.DeepEqual(cln, fWp) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(fWp), ToJSON(cln))
	}
}
