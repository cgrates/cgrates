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

	"github.com/cgrates/cgrates/utils"
)

var mapEv = MapEvent(map[string]interface{}{
	"test1": nil,
	"test2": 42,
	"test3": 42.3,
	"test4": true,
	"test5": "test",
	"test6": 10 * time.Second,
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
	data := utils.DataProvider(mapEv)
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
	data := utils.DataProvider(mapEv)
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

func TestMapEventRemoteHost(t *testing.T) {
	data := utils.DataProvider(mapEv)
	if rply, expected := data.RemoteHost(), utils.LocalAddr(); !reflect.DeepEqual(expected, rply) {
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
	} else if rply != 0 {
		t.Errorf("Expected: %+v , received duration: %+v", 0, rply)
	}
	expected := 10 * time.Second
	if rply, err := mapEv.GetDuration("test6"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = 42 * time.Second
	if rply, err := mapEv.GetDuration("test7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = 42
	if rply, err := mapEv.GetDuration("test2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetDurationIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetDurationIgnoreErrors("test"); rply != 0 {
		t.Errorf("Expected: %+v, received: %+v", 0, rply)
	}
	expected := 10 * time.Second
	if rply := mapEv.GetDurationIgnoreErrors("test6"); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = 42 * time.Second
	if rply := mapEv.GetDurationIgnoreErrors("test7"); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = 42
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

func TestGetTimePtr(t *testing.T) {
	rcv1, err := mapEv.GetTimePtr("test", utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv1 != nil {
		t.Errorf("Expected: nil, received: %+v", rcv1)
	}
	expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	rcv2, err := mapEv.GetTimePtr("test8", utils.EmptyString)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, *rcv2) {
		t.Errorf("Expecting %+v, received: %+v", expected, rcv2)
	}
	rcv3, err := mapEv.GetTimePtr("test9", utils.EmptyString)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, *rcv3) {
		t.Errorf("Expecting %+v, received: %+v", expected, rcv3)
	}
	if rcv1 == rcv2 || rcv2 == rcv3 || rcv1 == rcv3 {
		t.Errorf("Expecting to be different adresses")
	}
}

func TestGetTimePtrIgnoreErrors(t *testing.T) {
	rcv1 := mapEv.GetTimePtrIgnoreErrors("test", utils.EmptyString)
	if rcv1 != nil {
		t.Errorf("Expected: nil, received: %+v", rcv1)
	}
	expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	rcv2 := mapEv.GetTimePtrIgnoreErrors("test8", utils.EmptyString)
	if rcv2 != nil && !reflect.DeepEqual(expected, *rcv2) {
		t.Errorf("Expecting %+v, received: %+v", expected, rcv2)
	}
	rcv3 := mapEv.GetTimePtrIgnoreErrors("test9", utils.EmptyString)
	if rcv3 != nil && !reflect.DeepEqual(expected, *rcv3) {
		t.Errorf("Expecting %+v, received: %+v", expected, rcv3)
	}
	if rcv1 == rcv2 || rcv2 == rcv3 || rcv1 == rcv3 {
		t.Errorf("Expecting to be different adresses")
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
	mpIgnore := utils.NewStringSet([]string{"test6", "test7", "test8", "test9"})

	if rply := mapEv.AsMapString(mpIgnore); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	var mp MapEvent = nil
	if rply := mp.AsMapString(nil); !reflect.DeepEqual(map[string]string{}, rply) {
		t.Errorf("Expecting %+v, received: %+v", map[string]string{}, rply)
	}
	if rply := mp.AsMapString(mpIgnore); !reflect.DeepEqual(map[string]string{}, rply) {
		t.Errorf("Expecting %+v, received: %+v", map[string]string{}, rply)
	}
}

func TestMapEventGetTInt64(t *testing.T) {
	if rply, err := mapEv.GetTInt64("test2"); err != nil {
		t.Error(err)
	} else if rply != int64(42) {
		t.Errorf("Expecting %+v, received: %+v", int64(42), rply)
	}

	if rply, err := mapEv.GetTInt64("test3"); err != nil {
		t.Error(err)
	} else if rply != int64(42) {
		t.Errorf("Expecting %+v, received: %+v", int64(42), rply)
	}

	if rply, err := mapEv.GetTInt64("test4"); err == nil {
		t.Errorf("Expecting error, received: %+v with error %v", rply, err)
	}

	if rply, err := mapEv.GetTInt64("0test"); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting error: %v, received: %+v with error %v", utils.ErrNotFound, rply, err)
	}
}

func TestMapEventGetFloat64(t *testing.T) {
	if rply, err := mapEv.GetFloat64("test2"); err != nil {
		t.Error(err)
	} else if rply != float64(42) {
		t.Errorf("Expecting %+v, received: %+v", float64(42), rply)
	}

	if rply, err := mapEv.GetFloat64("test3"); err != nil {
		t.Error(err)
	} else if rply != float64(42.3) {
		t.Errorf("Expecting %+v, received: %+v", float64(42.3), rply)
	}

	if rply, err := mapEv.GetFloat64("test4"); err == nil {
		t.Errorf("Expecting error, received: %+v with error %v", rply, err)
	}

	if rply, err := mapEv.GetFloat64("0test"); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting error: %v, received: %+v with error %v", utils.ErrNotFound, rply, err)
	}
}

func TestMapEventGetDurationPtr(t *testing.T) {
	if rply, err := mapEv.GetDurationPtr("test4"); err == nil {
		t.Errorf("Expecting error, received: %+v with error %v", rply, err)
	}
	if rply, err := mapEv.GetDurationPtr("test"); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rply != nil {
		t.Errorf("Expected: %+v , received duration: %+v", nil, rply)
	}
	expected := utils.DurationPointer(10 * time.Second)
	if rply, err := mapEv.GetDurationPtr("test6"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = utils.DurationPointer(42 * time.Second)
	if rply, err := mapEv.GetDurationPtr("test7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = utils.DurationPointer(42)
	if rply, err := mapEv.GetDurationPtr("test2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetDurationPtrIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetDurationPtrIgnoreErrors("test"); rply != nil {
		t.Errorf("Expected: %+v, received: %+v", nil, rply)
	}
	expected := utils.DurationPointer(10 * time.Second)
	if rply := mapEv.GetDurationPtrIgnoreErrors("test6"); *rply != *expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = utils.DurationPointer(42 * time.Second)
	if rply := mapEv.GetDurationPtrIgnoreErrors("test7"); *rply != *expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = utils.DurationPointer(42)
	if rply := mapEv.GetDurationPtrIgnoreErrors("test2"); *rply != *expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetDurationPtrOrDefault(t *testing.T) {
	mapEv := NewMapEvent(nil)
	dflt := time.Nanosecond
	if ptr, _ := mapEv.GetDurationPtrOrDefault("test7", &dflt); dflt.String() != ptr.String() {
		t.Errorf("Expected: %+v, received: %+v", dflt, ptr)
	}
	newVal := 2 * time.Nanosecond
	mapEv["test7"] = newVal
	if ptr, _ := mapEv.GetDurationPtrOrDefault("test7", &dflt); newVal.String() != ptr.String() {
		t.Errorf("Expected: %+v, received: %+v", newVal, ptr)
	}
}

func TestMapEventCloneError(t *testing.T) {
	var testStruct MapEvent = nil
	var exp MapEvent = nil
	result := testStruct.Clone()
	if !reflect.DeepEqual(result, exp) {
		t.Errorf("Expected: %+v, received: %+v", exp, result)
	}
}

func TestMapEventData(t *testing.T) {
	testStruct := MapEvent{
		"key1": "val1",
	}
	expStruct := map[string]interface{}{
		"key1": "val1",
	}
	result := testStruct.Data()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("Expected: %+v, received: %+v", expStruct, result)
	}
}
