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
package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var sMap = map[string]interface{}{
	"test1": nil,
	"test2": 42,
	"test3": 42.3,
	"test4": true,
	"test5": "test",
	"test6": time.Duration(10 * time.Second),
	"test7": "42s",
	"test8": time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
	"test9": "2009-11-10T23:00:00Z",
}
var safEv = &SafEvent{Me: NewMapEvent(sMap)}

func TestSafEventNewSafEvent(t *testing.T) {
	if rply := NewSafEvent(sMap); !reflect.DeepEqual(safEv, rply) {
		t.Errorf("Expecting %+v, received: %+v", safEv, rply)
	}
}

func TestSafEventMapEvent(t *testing.T) {
	expected := NewMapEvent(sMap)
	if rply := safEv.MapEvent(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventFieldAsInterface(t *testing.T) {
	data := config.DataProvider(safEv)
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

func TestSafEventFieldAsString(t *testing.T) {
	data := config.DataProvider(safEv)
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

func TestSafEventAsNavigableMap(t *testing.T) {
	data := config.DataProvider(safEv)
	if rply, err := data.AsNavigableMap(nil); err != nil {
		t.Error(err)
	} else if expected := config.NewNavigableMap(sMap); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventRemoteHost(t *testing.T) {
	data := config.DataProvider(safEv)
	if rply, expected := data.RemoteHost(), new(utils.LocalAddr); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventClone(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("clone", func(t *testing.T) {
			t.Parallel()
			safEv.Clone()
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test4", true)
		})
	}
	rply := safEv.Clone()
	if !reflect.DeepEqual(safEv, rply) {
		t.Errorf("Expecting %+v, received: %+v", safEv, rply)
	}
	rply.Set("test4", false)
	if reflect.DeepEqual(safEv, rply) {
		t.Errorf("Expecting %+v, received: %+v", safEv, rply)
	}
}

func TestSafEventString(t *testing.T) {
	expected := safEv.Me.String()
	if rply := safEv.String(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	se := safEv.Clone()
	for i := 0; i < 10; i++ {
		t.Run("string", func(t *testing.T) {
			t.Parallel()
			se.String()
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			se.Remove("test4")
		})
	}
	se.Remove("test5")
	expected = se.Me.String()
	if rply := se.String(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventHasField(t *testing.T) {
	if rply := safEv.HasField("test4"); !rply {
		t.Errorf("Expecting true, received: %+v", rply)
	}
	se := safEv.Clone()
	for i := 0; i < 10; i++ {
		t.Run("field", func(t *testing.T) {
			t.Parallel()
			se.HasField("test4")
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			se.Remove("test4")
		})
	}
	se.Remove("test5")
	if rply := se.HasField("test5"); rply {
		t.Errorf("Expecting false, received: %+v", rply)
	}
}

func TestSafEventGet(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("get", func(t *testing.T) {
			t.Parallel()
			safEv.Get("test4")
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test4", true)
		})
	}
	safEv.Remove("test4")
	if rply, has := safEv.Get("test4"); has {
		t.Errorf("Expecting 'test4' to not be a field, recived: %+v", rply)
	}
	safEv.Set("test4", false)
	if rply, has := safEv.Get("test4"); !has {
		t.Errorf("Expecting 'test4' to be a field")
	} else if rply != false {
		t.Errorf("Expecting false, received: %+v", rply)
	}
	safEv.Set("test4", true)
	if rply, has := safEv.Get("test4"); !has {
		t.Errorf("Expecting 'test4' to be a field")
	} else if rply != true {
		t.Errorf("Expecting true, received: %+v", rply)
	}
}

func TestSafEventGetIgnoreErrors(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getIgnore", func(t *testing.T) {
			t.Parallel()
			safEv.GetIgnoreErrors("test4")
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test4", true)
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test4")
		})
	}
	safEv.Remove("test4")
	if rply := safEv.GetIgnoreErrors("test4"); rply != nil {
		t.Errorf("Expecting: null, recived: %+v", rply)
	}
	safEv.Set("test4", false)
	if rply := safEv.GetIgnoreErrors("test4"); rply != false {
		t.Errorf("Expecting false, received: %+v", rply)
	}
	safEv.Set("test4", true)
	if rply := safEv.GetIgnoreErrors("test4"); rply != true {
		t.Errorf("Expecting true, received: %+v", rply)
	}
}

func TestSafEventGetString(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getString", func(t *testing.T) {
			t.Parallel()
			if _, err := safEv.GetString("test4"); err != nil && err != utils.ErrNotFound {
				t.Error(err)
			}
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test4", true)
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test4")
		})
	}
	safEv.Remove("test2")
	if _, err := safEv.GetString("test2"); err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v ,recived: %+v", utils.ErrNotFound, err)
	}
	safEv.Set("test2", 42.3)
	if rply, err := safEv.GetString("test2"); err != nil {
		t.Error(err)
	} else if rply != "42.3" {
		t.Errorf("Expecting 42.3, received: %+v", rply)
	}
	safEv.Set("test2", 42)
	if rply, err := safEv.GetString("test2"); err != nil {
		t.Error(err)
	} else if rply != "42" {
		t.Errorf("Expecting 42, received: %+v", rply)
	}
}

func TestSafEventGetStringIgnoreErrors(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getStringIgn", func(t *testing.T) {
			t.Parallel()
			safEv.GetStringIgnoreErrors("test4")
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test4", true)
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test4")
		})
	}
	safEv.Remove("test2")
	if rply := safEv.GetStringIgnoreErrors("test2"); rply != "" {
		t.Errorf("Expecting: ,recived: %+v", err)
	}
	safEv.Set("test2", 42.3)
	if rply := safEv.GetStringIgnoreErrors("test2"); rply != "42.3" {
		t.Errorf("Expecting 42.3, received: %+v", rply)
	}
	safEv.Set("test2", 42)
	if rply := safEv.GetStringIgnoreErrors("test2"); rply != "42" {
		t.Errorf("Expecting 42, received: %+v", rply)
	}
}

func TestSafEventGetDuration(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getDuration", func(t *testing.T) {
			t.Parallel()
			if _, err := safEv.GetDuration("test6"); err != nil && err != utils.ErrNotFound {
				t.Error(err)
			}
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test6", time.Duration(10*time.Second))
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test6")
		})
	}
	safEv.Remove("test7")
	if _, err := safEv.GetDuration("test7"); err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v ,recived: %+v", utils.ErrNotFound, err)
	}
	if rply, err := safEv.GetDuration("test5"); err == nil {
		t.Errorf("Expecting: error,recived: %+v", rply)
	}
	safEv.Set("test7", "42s")
	expected := time.Duration(42 * time.Second)
	if rply, err := safEv.GetDuration("test7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = time.Duration(10 * time.Second)
	safEv.Set("test6", expected)
	if rply, err := safEv.GetDuration("test6"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventGetDurationIgnoreErrors(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getDurationIgn", func(t *testing.T) {
			t.Parallel()
			safEv.GetDurationIgnoreErrors("test6")
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test6", time.Duration(10*time.Second))
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test6")
		})
	}
	safEv.Remove("test7")
	if rply := safEv.GetDurationIgnoreErrors("test7"); rply != time.Duration(0) {
		t.Errorf("Expecting: %+v ,recived: %+v", time.Duration(0), rply)
	}
	safEv.Set("test7", "42s")
	expected := time.Duration(42 * time.Second)
	if rply := safEv.GetDurationIgnoreErrors("test7"); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = time.Duration(10 * time.Second)
	safEv.Set("test6", expected)
	if rply := safEv.GetDurationIgnoreErrors("test6"); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventGetDurationPtr(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getDurationPtr", func(t *testing.T) {
			t.Parallel()
			if _, err := safEv.GetDurationPtr("test6"); err != nil && err != utils.ErrNotFound {
				t.Error(err)
			}
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test6", time.Duration(10*time.Second))
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test6")
		})
	}
	safEv.Remove("test7")
	if _, err := safEv.GetDurationPtr("test7"); err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v ,recived: %+v", utils.ErrNotFound, err)
	}
	if rply, err := safEv.GetDurationPtr("test5"); err == nil {
		t.Errorf("Expecting: error,recived: %+v", rply)
	}
	safEv.Set("test7", "42s")
	expected := time.Duration(42 * time.Second)
	if rply, err := safEv.GetDurationPtr("test7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(&expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = time.Duration(10 * time.Second)
	safEv.Set("test6", expected)
	if rply, err := safEv.GetDurationPtr("test6"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(&expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventGetDurationPtrOrDefault(t *testing.T) {
	def := time.Duration(450)
	for i := 0; i < 10; i++ {
		t.Run("getDurationPtrDef", func(t *testing.T) {
			t.Parallel()
			if _, err := safEv.GetDurationPtrOrDefault("test6", &def); err != nil {
				t.Error(err)
			}
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test6", time.Duration(10*time.Second))
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test6")
		})
	}
	safEv.Remove("test7")
	if rply, err := safEv.GetDurationPtrOrDefault("test7", &def); err != nil {
		t.Errorf("Expecting: %+v ,recived: %+v", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(&def, rply) {
		t.Errorf("Expecting %+v, received: %+v", def, rply)
	}
	if rply, err := safEv.GetDurationPtrOrDefault("test5", &def); err == nil {
		t.Errorf("Expecting: error,recived: %+v", rply)
	}
	safEv.Set("test7", "42s")
	expected := time.Duration(42 * time.Second)
	if rply, err := safEv.GetDurationPtrOrDefault("test7", &def); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(&expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = time.Duration(10 * time.Second)
	safEv.Set("test6", expected)
	if rply, err := safEv.GetDurationPtrOrDefault("test6", &def); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(&expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventGetTime(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getTime", func(t *testing.T) {
			t.Parallel()
			if _, err := safEv.GetTime("test8", ""); err != nil && err != utils.ErrNotFound {
				t.Error(err)
			}
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test8", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test8")
		})
	}
	safEv.Remove("test9")
	if _, err := safEv.GetTime("test9", ""); err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v ,recived: %+v", utils.ErrNotFound, err)
	}
	if rply, err := safEv.GetTime("test5", ""); err == nil {
		t.Errorf("Expecting: error,recived: %+v", rply)
	}
	safEv.Set("test9", "2010-11-10T23:00:00Z")
	expected := time.Date(2010, 11, 10, 23, 0, 0, 0, time.UTC)
	if rply, err := safEv.GetTime("test9", ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	safEv.Set("test8", expected)
	if rply, err := safEv.GetTime("test8", ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventGetTimeIgnoreErrors(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getTimeIgn", func(t *testing.T) {
			t.Parallel()
			safEv.GetTimeIgnoreErrors("test8", "")
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test8", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test8")
		})
	}
	safEv.Remove("test9")
	if rply := safEv.GetTimeIgnoreErrors("test9", ""); !rply.IsZero() {
		t.Errorf("Expecting January 1, year 1, 00:00:00.000000000 UTC, received: %+v", rply)
	}
	if rply := safEv.GetTimeIgnoreErrors("test5", ""); !rply.IsZero() {
		t.Errorf("Expecting January 1, year 1, 00:00:00.000000000 UTC, received: %+v", rply)
	}
	safEv.Set("test9", "2010-11-10T23:00:00Z")
	expected := time.Date(2010, 11, 10, 23, 0, 0, 0, time.UTC)
	if rply := safEv.GetTimeIgnoreErrors("test9", ""); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	safEv.Set("test8", expected)
	if rply := safEv.GetTimeIgnoreErrors("test8", ""); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventGetSetString(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getSetString", func(t *testing.T) {
			t.Parallel()
			safEv.GetSetString("test4", "true")
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test4")
		})
	}
	safEv.Remove("test2")
	expected := "test2"
	if rply, err := safEv.GetSetString("test2", expected); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expecting: %+v ,recived: %+v", expected, rply)
	}
	safEv.Set("test2", 42.3)
	if rply, err := safEv.GetSetString("test2", ""); err != nil {
		t.Error(err)
	} else if rply != "42.3" {
		t.Errorf("Expecting 42.3, received: %+v", rply)
	}
	safEv.Set("test2", 42)
	if rply, err := safEv.GetSetString("test2", "test"); err != nil {
		t.Error(err)
	} else if rply != "42" {
		t.Errorf("Expecting 42, received: %+v", rply)
	}
}

func TestSafEventGetMapInterface(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("getMapInt", func(t *testing.T) {
			t.Parallel()
			safEv.GetMapInterface()
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test8", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test8")
		})
	}
	safEv.Remove("test8")
	if rply := safEv.GetMapInterface(); !reflect.DeepEqual(sMap, rply) {
		t.Errorf("Expecting %+v, received: %+v", sMap, rply)
	}
	safEv.Set("test8", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
	if rply := safEv.GetMapInterface(); !reflect.DeepEqual(sMap, rply) {
		t.Errorf("Expecting %+v, received: %+v", sMap, rply)
	}
}

func TestSafEventAsMapInterface(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("asMapInt", func(t *testing.T) {
			t.Parallel()
			safEv.AsMapInterface()
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test8", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test8")
		})
	}
	safEv.Set("test8", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
	rply := safEv.AsMapInterface()
	if !reflect.DeepEqual(sMap, rply) {
		t.Errorf("Expecting %+v, received: %+v", sMap, rply)
	}
	safEv.Remove("test8")
	if reflect.DeepEqual(sMap, rply) {
		t.Errorf("Expecting not be %+v, received: %+v", sMap, rply)
	}
	safEv.Set("test8", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
}

func TestSafEventAsMapString(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("asMapStr", func(t *testing.T) {
			t.Parallel()
			safEv.AsMapString(nil)
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test9", true)
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test9")
		})
	}
	var expected map[string]string
	if expected, err = safEv.Me.AsMapString(nil); err != nil {
		t.Error(err)
	}
	if rply, err := safEv.AsMapString(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	delete(expected, "test1")
	if rply, err := safEv.AsMapString(utils.StringMap{"test1": true}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventAsMapStringIgnoreErrors(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("asMapStr", func(t *testing.T) {
			t.Parallel()
			safEv.AsMapStringIgnoreErrors(nil)
		})
		t.Run("set", func(t *testing.T) {
			t.Parallel()
			safEv.Set("test9", true)
		})
		t.Run("remove", func(t *testing.T) {
			t.Parallel()
			safEv.Remove("test9")
		})
	}
	var expected map[string]string
	if expected, err = safEv.Me.AsMapString(nil); err != nil {
		t.Error(err)
	}
	if rply := safEv.AsMapStringIgnoreErrors(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected["test8"], rply["test8"])
	}
	delete(expected, "test1")
	if rply := safEv.AsMapStringIgnoreErrors(utils.StringMap{"test1": true}); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSafEventAsCDR(t *testing.T) {
	se := SafEvent{Me: NewMapEvent(nil)}
	expected := &CDR{Cost: -1.0, ExtraFields: make(map[string]string)}
	if rply, err := se.AsCDR(nil, utils.EmptyString, utils.EmptyString); err != nil {
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
	if rply, err := se.AsCDR(cfg, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	se = SafEvent{Me: MapEvent{"SetupTime": "clearly not time string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"AnswerTime": "clearly not time string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"Usage": "clearly not duration string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"Partial": "clearly not bool string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"PreRated": "clearly not bool string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"Cost": "clearly not float64 string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"ExtraField1": 5, "ExtraField2": "extra"}}
	expected = &CDR{
		Cost: -1.0,
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		}}
	if rply, err := se.AsCDR(nil, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	se = SafEvent{Me: MapEvent{
		"ExtraField1": 5,
		"Source":      1001,
		"CostSource":  "1002",
		"ExtraField2": "extra",
	}}
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
	if rply, err := se.AsCDR(cfg, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	se = SafEvent{Me: MapEvent{
		"ExtraField1": 5,
		"Source":      1001,
		"CostSource":  "1002",
		"ExtraField2": "extra",
		"SetupTime":   "2009-11-10T23:00:00Z",
		"Usage":       "42s",
		"PreRated":    "True",
		"Cost":        "42.3",
	}}
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
	if rply, err := se.AsCDR(cfg, "itsyscom.com", utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}
