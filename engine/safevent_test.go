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

func TestSafEventClone(t *testing.T) {
	var rply *SafEvent
	safEv = safEv.Clone()
	t.Run("clone1", func(t *testing.T) {
		t.Parallel()
		rply = safEv.Clone()
		if !reflect.DeepEqual(safEv, rply) {
			t.Errorf("Expecting %+v, received: %+v", safEv, rply)
		}
	})
	t.Run("clone2", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test4", false)
		if reflect.DeepEqual(safEv, rply) {
			t.Errorf("Expecting %+v, received: %+v", safEv, rply)
		}
	})
}

func TestSafEventString(t *testing.T) {
	var rply string
	expected := safEv.Me.String()
	t.Run("string1", func(t *testing.T) {
		t.Parallel()
		rply = safEv.String()
		if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("string2", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test4", false)
		rply = safEv.String()
		safEv.RLock()
		expected = safEv.Me.String()
		safEv.RUnlock()
		if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", safEv, rply)
		}
	})
	t.Run("string3", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test4", true)
		rply = safEv.String()
		safEv.RLock()
		expected = safEv.Me.String()
		safEv.RUnlock()
		if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", safEv, rply)
		}
	})
}

func TestSafEventHasField(t *testing.T) {
	t.Run("field1", func(t *testing.T) {
		t.Parallel()
		if rply := safEv.HasField("test4"); !rply {
			t.Errorf("Expecting true, received: %+v", rply)
		}
	})
	t.Run("field2", func(t *testing.T) {
		t.Parallel()
		safEv.Remove("test4")
		if rply := safEv.HasField("test4"); rply {
			t.Errorf("Expecting false, received: %+v", rply)
		}
	})
	t.Run("field3", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test4", true)
		if rply := safEv.HasField("test4"); !rply {
			t.Errorf("Expecting true, received: %+v", rply)
		}
	})
}

func TestSafEventGet(t *testing.T) {
	t.Run("get1", func(t *testing.T) {
		t.Parallel()
		if rply, has := safEv.Get("test5"); !has {
			t.Errorf("Expecting 'test5' to be a field")
		} else if rply != "test" {
			t.Errorf("Expecting test, received: %+v", rply)
		}
	})
	t.Run("get2", func(t *testing.T) {
		t.Parallel()
		safEv.Remove("test5")
		if rply, has := safEv.Get("test5"); has {
			t.Errorf("Expecting 'test5' to not be a field, recived %+v", rply)
		}
	})
	t.Run("get3", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test5", false)
		if rply, has := safEv.Get("test5"); !has {
			t.Errorf("Expecting 'test5' to be a field")
		} else if rply != false {
			t.Errorf("Expecting false, received: %+v", rply)
		}
	})
}

func TestSafEventGetIgnoreErrors(t *testing.T) {
	t.Run("getIgnore1", func(t *testing.T) {
		t.Parallel()
		if rply := safEv.GetIgnoreErrors("test2"); rply != 42 {
			t.Errorf("Expecting 42, received: %+v", rply)
		}
	})
	t.Run("getIgnore2", func(t *testing.T) {
		t.Parallel()
		safEv.Remove("test2")
		if rply := safEv.GetIgnoreErrors("test2"); rply != nil {
			t.Errorf("Expecting null, received: %+v", rply)
		}
	})
	t.Run("getIgnore3", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test2", 43)
		if rply := safEv.GetIgnoreErrors("test2"); rply != 43 {
			t.Errorf("Expecting 43, received: %+v", rply)
		}
	})
}

func TestSafEventGetString(t *testing.T) {
	t.Run("getString1", func(t *testing.T) {
		t.Parallel()
		if rply, err := safEv.GetString("test6"); err != nil {
			t.Error(err)
		} else if rply != "10s" {
			t.Errorf("Expecting 10s, received: %+v", rply)
		}
	})
	t.Run("getString2", func(t *testing.T) {
		t.Parallel()
		if rply, err := safEv.GetString("test"); err == nil {
			t.Errorf("Expecting 'test' to not be a field, recived %+v", rply)
		}
	})
	t.Run("getString3", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test7", "43s")
		if rply, err := safEv.GetString("test7"); err != nil {
			t.Error(err)
		} else if rply != "43s" {
			t.Errorf("Expecting 43s, received: %+v", rply)
		}
	})
}

func TestSafEventGetStringIgnoreErrors(t *testing.T) {
	t.Run("getStringIgnore1", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test1", nil)
		if rply := safEv.GetStringIgnoreErrors("test1"); rply != "" {
			t.Errorf("Expecting , received: %+v", rply)
		}
	})
	t.Run("getStringIgnore2", func(t *testing.T) {
		t.Parallel()
		safEv.Remove("test2")
		if rply := safEv.GetStringIgnoreErrors("test2"); rply != "" {
			t.Errorf("Expecting null, received: %+v", rply)
		}
	})
	t.Run("getStringIgnore3", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test1", 43)
		if rply := safEv.GetStringIgnoreErrors("test1"); rply != "43" {
			t.Errorf("Expecting 43, received: %+v", rply)
		}
	})
}

func TestSafEventGetDuration(t *testing.T) {
	t.Run("getDuration1", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(43 * time.Second)
		if rply, err := safEv.GetDuration("test7"); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getDuration2", func(t *testing.T) {
		t.Parallel()
		if rply, err := safEv.GetDuration("test"); err == nil {
			t.Errorf("Expecting 'test' to not be a field, recived %+v", rply)
		}
	})
	t.Run("getDuration3", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(10 * time.Second)
		if rply, err := safEv.GetDuration("test6"); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventGetDurationIgnoreErrors(t *testing.T) {
	t.Run("getDurationIgnore1", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(43 * time.Second)
		if rply := safEv.GetDurationIgnoreErrors("test7"); !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getDurationIgnore2", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(0)
		if rply := safEv.GetDurationIgnoreErrors("test"); !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getDurationIgnore3", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(10 * time.Second)
		if rply := safEv.GetDurationIgnoreErrors("test6"); !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventGetDurationPtr(t *testing.T) {
	t.Run("getDurationPtr1", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(43 * time.Second)
		if rply, err := safEv.GetDurationPtr("test7"); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(&expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getDurationPtr2", func(t *testing.T) {
		t.Parallel()
		if _, err := safEv.GetDurationPtr("test"); err != utils.ErrNotFound {
			t.Errorf("Expecting %+v, recived %+v", utils.ErrNotFound, err)
		}
	})
	t.Run("getDurationPtr3", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(11 * time.Second)
		safEv.Set("test6", expected)
		if rply, err := safEv.GetDurationPtr("test6"); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(&expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventGetDurationPtrOrDefault(t *testing.T) {
	def := time.Duration(450)
	safEv = NewSafEvent(sMap)
	t.Run("getDurationPtrDef1", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(42 * time.Second)
		if rply, err := safEv.GetDurationPtrOrDefault("test7", &def); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(&expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getDurationPtrDef2", func(t *testing.T) {
		t.Parallel()
		expected := time.Duration(12 * time.Second)
		if rply, err := safEv.GetDurationPtrOrDefault("othertest", &expected); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(&expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getDurationPtrDef3", func(t *testing.T) {
		t.Parallel()
		if rply, err := safEv.GetDurationPtrOrDefault("test5", &def); err == nil {
			t.Errorf("Expecting error, recived %+v", rply)
		}
	})
}

func TestSafEventGetGetTime(t *testing.T) {
	t.Run("getTime1", func(t *testing.T) {
		t.Parallel()
		expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
		if rply, err := safEv.GetTime("test8", ""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getTime2", func(t *testing.T) {
		t.Parallel()
		if rply, err := safEv.GetTime("test", ""); err == nil {
			t.Errorf("Expecting 'test' to not be a field, recived %+v", rply)
		}
		if rply, err := safEv.GetTime("test4", ""); err == nil {
			t.Errorf("Expecting 'test' to not be a field, recived %+v", rply)
		}
	})
	t.Run("getTime3", func(t *testing.T) {
		t.Parallel()
		expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
		if rply, err := safEv.GetTime("test9", ""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventGetTimeIgnoreErrors(t *testing.T) {
	t.Run("getTimeIgnore1", func(t *testing.T) {
		t.Parallel()
		expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
		if rply := safEv.GetTimeIgnoreErrors("test8", ""); !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getTimeIgnore2", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test8", nil)
		if rply := safEv.GetTimeIgnoreErrors("test", ""); !rply.IsZero() {
			t.Errorf("Expecting January 1, year 1, 00:00:00.000000000 UTC, received: %+v", rply)
		}
	})
	t.Run("getTimeIgnore3", func(t *testing.T) {
		t.Parallel()
		expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
		if rply := safEv.GetTimeIgnoreErrors("test9", ""); !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventGetSetString(t *testing.T) {
	t.Run("getSetString1", func(t *testing.T) {
		t.Parallel()
		var expected string
		if expected, err = safEv.GetString("test1"); err != nil {
			t.Error(err)
		}
		if rply, err := safEv.GetSetString("test1", "test"); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
		safEv.Remove("test1")
		safEv.Remove("test2")
		expected = "test"
		if rply, err := safEv.GetSetString("test1", expected); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
		if rply, err := safEv.GetString("test1"); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("getSetString2", func(t *testing.T) {
		t.Parallel()
		expected := "test"
		if rply, err := safEv.GetSetString("test2", "test"); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventGetMapInterface(t *testing.T) {
	t.Run("getMapInt1", func(t *testing.T) {
		t.Parallel()
		if rply := safEv.GetMapInterface(); !reflect.DeepEqual(sMap, rply) {
			t.Errorf("Expecting %+v, received: %+v", sMap, rply)
		}
	})
	t.Run("getMapInt2", func(t *testing.T) {
		t.Parallel()
		sMap["test10"] = true
		safEv.Set("test10", true)
	})
	t.Run("getMapInt3", func(t *testing.T) {
		t.Parallel()
		sMap["test12"] = "time"
		safEv.Set("test12", "time")
		if rply := safEv.GetMapInterface(); !reflect.DeepEqual(sMap, rply) {
			t.Errorf("Expecting %+v, received: %+v", sMap, rply)
		} else if rply["test12"] = 12; !reflect.DeepEqual(sMap, rply) {
			t.Errorf("Expecting %+v, received: %+v", sMap, rply)
		}
	})
}
func TestSafEventAsMapInterface(t *testing.T) {
	t.Run("asMapInt1", func(t *testing.T) {
		t.Parallel()
		if rply := safEv.AsMapInterface(); !reflect.DeepEqual(sMap, rply) {
			t.Errorf("Expecting %+v, received: %+v", sMap, rply)
		}
	})
	t.Run("asMapInt2", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test10", true)
	})
	t.Run("asMapInt3", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test12", "time")
		expected := safEv.GetMapInterface()
		if rply := safEv.AsMapInterface(); !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		} else if rply["test12"] = 12; reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventAsMapString(t *testing.T) {
	t.Run("asMapStr1", func(t *testing.T) {
		t.Parallel()
		var expected map[string]string
		if expected, err = safEv.Me.AsMapString(nil); err != nil {
			t.Error(err)
		}
		if rply, err := safEv.AsMapString(nil); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("asMapStr2", func(t *testing.T) {
		t.Parallel()
		safEv.Remove("test10")
	})
	t.Run("asMapStr3", func(t *testing.T) {
		t.Parallel()
		safEv.Remove("test12")
		var expected map[string]string
		if expected, err = safEv.Me.AsMapString(nil); err != nil {
			t.Error(err)
		}
		delete(expected, "test2")
		if rply, err := safEv.AsMapString(utils.StringMap{"test2": true}); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventAsMapStringIgnoreErrors(t *testing.T) {
	t.Run("asMapStrIgn1", func(t *testing.T) {
		t.Parallel()
		var expected map[string]string
		safEv.RLock()
		if expected, err = safEv.Me.AsMapString(nil); err != nil {
			t.Error(err)
		}
		safEv.RUnlock()
		if rply := safEv.AsMapStringIgnoreErrors(nil); !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
	t.Run("asMapStrIgn2", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test10", "test3")
	})
	t.Run("asMapStrIgn3", func(t *testing.T) {
		t.Parallel()
		safEv.Set("test12", 42)
		var expected map[string]string
		if expected, err = safEv.Me.AsMapString(utils.StringMap{"test12": true}); err != nil {
			t.Error(err)
		}
		if rply := safEv.AsMapStringIgnoreErrors(utils.StringMap{"test12": true}); !reflect.DeepEqual(expected, rply) {
			t.Errorf("Expecting %+v, received: %+v", expected, rply)
		}
	})
}

func TestSafEventAsCDR(t *testing.T) {
	se := SafEvent{Me: NewMapEvent(nil)}
	expected := &CDR{Cost: -1.0, ExtraFields: make(map[string]string)}
	if rply, err := se.AsCDR(nil, utils.EmptyString); err != nil {
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
		RequestType: cfg.DefaultReqType,
		Tenant:      cfg.DefaultTenant,
		Category:    cfg.DefaultCategory,
		ExtraFields: make(map[string]string),
	}
	if rply, err := se.AsCDR(cfg, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	se = SafEvent{Me: MapEvent{"SetupTime": "clearly not time string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"AnswerTime": "clearly not time string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"Usage": "clearly not duration string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"Partial": "clearly not bool string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"PreRated": "clearly not bool string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"Cost": "clearly not float64 string"}}
	if _, err := se.AsCDR(nil, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	se = SafEvent{Me: MapEvent{"ExtraField1": 5, "ExtraField2": "extra"}}
	expected = &CDR{
		Cost: -1.0,
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		}}
	if rply, err := se.AsCDR(nil, utils.EmptyString); err != nil {
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
		RequestType: cfg.DefaultReqType,
		Tenant:      cfg.DefaultTenant,
		Category:    cfg.DefaultCategory,
	}
	if rply, err := se.AsCDR(cfg, utils.EmptyString); err != nil {
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
		RequestType: cfg.DefaultReqType,
		Tenant:      cfg.DefaultTenant,
		Category:    cfg.DefaultCategory,
	}
	if rply, err := se.AsCDR(cfg, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}
