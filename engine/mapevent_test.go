/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var mapEv = MapEvent(map[string]interface{}{
	"test1": nil,
	"test2": 42,
	"test3": 42.3,
	"test4": true,
	"test5": "test",
	"test6": time.Duration(10 * time.Second),
	"test7": "42s",
	"test8": time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
	"test9": "2009-11-10T23:00:00Z",
})

func TestMapEventNewMapEvent(t *testing.T) {
	if rply, expected := NewMapEvent(nil), MapEvent(make(map[string]interface{})); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	mp := map[string]interface{}{
		"test1": nil,
		"test2": 42,
		"test3": 42.3,
		"test4": true,
		"test5": "test",
	}
	if rply, expected := NewMapEvent(mp), MapEvent(mp); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventFieldAsInterface(t *testing.T) {
	data := config.DataProvider(mapEv)
	if _, err := data.FieldAsInterface([]string{"first", "second"}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := data.FieldAsInterface([]string{"first"}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if rply, err := data.FieldAsInterface([]string{"test1"}); err != nil {
		t.Error(err)
	} else if rply != nil {
		t.Errorf("Expecting %+v, received: %+v", nil, rply)
	}
	if rply, err := data.FieldAsInterface([]string{"test4"}); err != nil {
		t.Error(err)
	} else if expected := true; rply != expected {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventFieldAsString(t *testing.T) {
	data := config.DataProvider(mapEv)
	if _, err := data.FieldAsString([]string{"first", "second"}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := data.FieldAsString([]string{"first"}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if rply, err := data.FieldAsString([]string{"test1"}); err != nil {
		t.Error(err)
	} else if rply != "" {
		t.Errorf("Expecting %+v, received: %+v", "", rply)
	}
	if rply, err := data.FieldAsString([]string{"test4"}); err != nil {
		t.Error(err)
	} else if expected := "true"; rply != expected {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventAsNavigableMap(t *testing.T) {
	data := config.DataProvider(mapEv)
	if rply, err := data.AsNavigableMap(nil); err != nil {
		t.Error(err)
	} else if expected := config.NewNavigableMap(mapEv); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventRemoteHost(t *testing.T) {
	data := config.DataProvider(mapEv)
	if rply, expected := data.RemoteHost(), new(utils.LocalAddr); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventString(t *testing.T) {
	me := NewMapEvent(nil)
	if rply, expected := me.String(), utils.ToJSON(me); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	if rply, expected := mapEv.String(), utils.ToJSON(mapEv); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventHasField(t *testing.T) {
	me := NewMapEvent(nil)
	if rply := me.HasField("test1"); rply {
		t.Errorf("Expecting false, received: %+v", rply)
	}
	if rply := mapEv.HasField("test2"); !rply {
		t.Errorf("Expecting true, received: %+v", rply)
	}
	if rply := mapEv.HasField("test"); rply {
		t.Errorf("Expecting false, received: %+v", rply)
	}
}

func TestMapEventGetString(t *testing.T) {
	if rply, err := mapEv.GetString("test"); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rply != utils.EmptyString {
		t.Errorf("Expected error: %+v , received string: %+v", utils.ErrNotFound, rply)
	}
	if rply, err := mapEv.GetString("test2"); err != nil {
		t.Error(err)
	} else if rply != "42" {
		t.Errorf("Expecting %+v, received: %+v", "42", rply)
	}
	if rply, err := mapEv.GetString("test1"); err != nil {
		t.Error(err)
	} else if rply != utils.EmptyString {
		t.Errorf("Expecting , received: %+v", rply)
	}
}

func TestMapEventGetStringIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetStringIgnoreErrors("test"); rply != utils.EmptyString {
		t.Errorf("Expected: , received: %+v", rply)
	}
	if rply := mapEv.GetStringIgnoreErrors("test2"); rply != "42" {
		t.Errorf("Expecting 42, received: %+v", rply)
	}
	if rply := mapEv.GetStringIgnoreErrors("test1"); rply != utils.EmptyString {
		t.Errorf("Expecting , received: %+v", rply)
	}
}

func TestMapEventGetDuration(t *testing.T) {
	if rply, err := mapEv.GetDuration("test"); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rply != time.Duration(0) {
		t.Errorf("Expected: %+v , received duration: %+v", time.Duration(0), rply)
	}
	expected := time.Duration(10 * time.Second)
	if rply, err := mapEv.GetDuration("test6"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = time.Duration(42 * time.Second)
	if rply, err := mapEv.GetDuration("test7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = time.Duration(42)
	if rply, err := mapEv.GetDuration("test2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetDurationIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetDurationIgnoreErrors("test"); rply != time.Duration(0) {
		t.Errorf("Expected: %+v, received: %+v", time.Duration(0), rply)
	}
	expected := time.Duration(10 * time.Second)
	if rply := mapEv.GetDurationIgnoreErrors("test6"); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = time.Duration(42 * time.Second)
	if rply := mapEv.GetDurationIgnoreErrors("test7"); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = time.Duration(42)
	if rply := mapEv.GetDurationIgnoreErrors("test2"); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetTime(t *testing.T) {
	if rply, err := mapEv.GetTime("test", utils.EmptyString); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if !rply.IsZero() {
		t.Errorf("Expected: January 1, year 1, 00:00:00.000000000 UTC, received: %+v", rply)
	}
	expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	if rply, err := mapEv.GetTime("test8", utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	if rply, err := mapEv.GetTime("test9", utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetTimeIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetTimeIgnoreErrors("test", utils.EmptyString); !rply.IsZero() {
		t.Errorf("Expected: January 1, year 1, 00:00:00.000000000 UTC, received: %+v", rply)
	}
	expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	if rply := mapEv.GetTimeIgnoreErrors("test8", utils.EmptyString); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	if rply := mapEv.GetTimeIgnoreErrors("test9", utils.EmptyString); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
}

func TestMapEventClone(t *testing.T) {
	rply := mapEv.Clone()
	if !reflect.DeepEqual(mapEv, rply) {
		t.Errorf("Expecting %+v, received: %+v", mapEv, rply)
	}
	rply["test1"] = "testTest"
	if reflect.DeepEqual(mapEv, rply) {
		t.Errorf("Expecting different from: %+v, received: %+v", mapEv, rply)
	}
}

func TestMapEventAsMapString(t *testing.T) {
	expected := map[string]string{
		"test1": utils.EmptyString,
		"test2": "42",
		"test3": "42.3",
		"test4": "true",
		"test5": "test",
	}
	mpIgnore := utils.StringMap{
		"test6": true,
		"test7": false,
		"test8": true,
		"test9": false,
	}
	if rply, err := mapEv.AsMapString(mpIgnore); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	var mp MapEvent
	mp = nil
	if rply, err := mp.AsMapString(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(map[string]string{}, rply) {
		t.Errorf("Expecting %+v, received: %+v", map[string]string{}, rply)
	}
	if rply, err := mp.AsMapString(mpIgnore); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(map[string]string{}, rply) {
		t.Errorf("Expecting %+v, received: %+v", map[string]string{}, rply)
	}
}

func TestMapEventAsMapStringIgnoreErrors(t *testing.T) {
	expected := map[string]string{
		"test1": utils.EmptyString,
		"test2": "42",
		"test3": "42.3",
		"test4": "true",
		"test5": "test",
	}
	mpIgnore := utils.StringMap{
		"test6": true,
		"test7": true,
		"test8": true,
		"test9": true,
	}
	if rply := mapEv.AsMapStringIgnoreErrors(mpIgnore); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	var mp MapEvent
	mp = nil
	if rply := mp.AsMapStringIgnoreErrors(nil); !reflect.DeepEqual(map[string]string{}, rply) {
		t.Errorf("Expecting %+v, received: %+v", map[string]string{}, rply)
	}
	if rply := mp.AsMapStringIgnoreErrors(mpIgnore); !reflect.DeepEqual(map[string]string{}, rply) {
		t.Errorf("Expecting %+v, received: %+v", map[string]string{}, rply)
	}
}

func TestMapEventAsCDR(t *testing.T) {
	me := NewMapEvent(nil)
	expected := &CDR{Cost: -1.0, ExtraFields: make(map[string]string)}
	if rply, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	expected = &CDR{
		CGRID:       "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Cost:        -1.0,
		RunID:       utils.MetaRaw,
		ToR:         utils.VOICE,
		RequestType: cfg.GeneralCfg().DefaultReqType,
		Tenant:      cfg.GeneralCfg().DefaultTenant,
		Category:    cfg.GeneralCfg().DefaultCategory,
		ExtraFields: make(map[string]string),
	}
	if rply, err := me.AsCDR(cfg, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	me = MapEvent{"SetupTime": "clearly not time string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"AnswerTime": "clearly not time string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"Usage": "clearly not duration string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"Partial": "clearly not bool string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"PreRated": "clearly not bool string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"Cost": "clearly not float64 string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"ExtraField1": 5, "ExtraField2": "extra"}
	expected = &CDR{
		Cost: -1.0,
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		}}
	if rply, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	me = MapEvent{
		"ExtraField1": 5,
		"Source":      1001,
		"CostSource":  "1002",
		"ExtraField2": "extra",
	}
	expected = &CDR{
		CGRID:      "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Cost:       -1.0,
		Source:     "1001",
		CostSource: "1002",
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		},
		RunID:       utils.MetaRaw,
		ToR:         utils.VOICE,
		RequestType: cfg.GeneralCfg().DefaultReqType,
		Tenant:      cfg.GeneralCfg().DefaultTenant,
		Category:    cfg.GeneralCfg().DefaultCategory,
	}
	if rply, err := me.AsCDR(cfg, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	me = MapEvent{
		"ExtraField1": 5,
		"Source":      1001,
		"CostSource":  "1002",
		"ExtraField2": "extra",
		"SetupTime":   "2009-11-10T23:00:00Z",
		"Usage":       "42s",
		"PreRated":    "True",
		"Cost":        "42.3",
	}
	expected = &CDR{
		CGRID:      "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Tenant:     "itsyscom.com",
		Cost:       42.3,
		Source:     "1001",
		CostSource: "1002",
		PreRated:   true,
		Usage:      time.Duration(42 * time.Second),
		SetupTime:  time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		},
		RunID:       utils.MetaRaw,
		ToR:         utils.VOICE,
		RequestType: cfg.GeneralCfg().DefaultReqType,
		Category:    cfg.GeneralCfg().DefaultCategory,
	}
	if rply, err := me.AsCDR(cfg, "itsyscom.com", utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}
