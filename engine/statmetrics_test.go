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
	"github.com/cgrates/cgrates/utils"
	"testing"
	"time"
)

func TestASRGetStringValue(t *testing.T) {
	asr, _ := NewASR()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := asr.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev)
	if strVal := asr.GetStringValue(""); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	asr.AddEvent(ev2)
	asr.AddEvent(ev3)
	if strVal := asr.GetStringValue(""); strVal != "33.33333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev3.TenantID())
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev.TenantID())
	if strVal := asr.GetStringValue(""); strVal != "0%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev2.TenantID())
	if strVal := asr.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong asr value: %s", strVal)
	}

}

func TestASRGetValue(t *testing.T) {
	asr, _ := NewASR()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev)
	if v := asr.GetValue(); v != 100.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	asr.AddEvent(ev2)
	asr.AddEvent(ev3)
	if v := asr.GetValue(); v != 33.33333 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev3.TenantID())
	if v := asr.GetValue(); v != 50.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev.TenantID())
	if v := asr.GetValue(); v != 0.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev2.TenantID())
	if v := asr.GetValue(); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}
}

func TestACDGetStringValue(t *testing.T) {
	acd, _ := NewACD()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev)
	if strVal := acd.GetStringValue(""); strVal != "10" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acd.AddEvent(ev2)
	acd.AddEvent(ev3)
	if strVal := acd.GetStringValue(""); strVal != "3.33333" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev3.TenantID())
	if strVal := acd.GetStringValue(""); strVal != "5" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev.TenantID())
	if strVal := acd.GetStringValue(""); strVal != "0" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.TenantID())
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}

}

func TestACDGetValue(t *testing.T) {
	acd, _ := NewACD()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	acd.AddEvent(ev)
	if v := acd.GetValue(); v != 10.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acd.AddEvent(ev2)
	acd.AddEvent(ev3)
	if v := acd.GetValue(); v != 3.33333 {
		t.Errorf("wrong asr value: %f", v)
	}
	acd.RemEvent(ev3.TenantID())
	if v := acd.GetValue(); v != 5.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	acd.RemEvent(ev.TenantID())
	if v := acd.GetValue(); v != 0.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	acd.RemEvent(ev2.TenantID())
	if v := acd.GetValue(); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}

}
