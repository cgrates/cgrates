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

	"github.com/cgrates/cgrates/utils"
)

func TestASRGetStringValue(t *testing.T) {
	asr, _ := NewASR(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := asr.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev)
	if strVal := asr.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	asr.AddEvent(ev2)
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev3)
	if strVal := asr.GetStringValue(""); strVal != "33.33333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev3.ID)
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4)
	asr.AddEvent(ev5)
	asr.RemEvent(ev.ID)
	if strVal := asr.GetStringValue(""); strVal != "66.66667%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev2.ID)
	if strVal := asr.GetStringValue(""); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev4.ID)
	asr.RemEvent(ev5.ID)
	if strVal := asr.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestASRGetValue(t *testing.T) {
	asr, _ := NewASR(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev)
	if v := asr.GetValue(); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	asr.AddEvent(ev2)
	asr.AddEvent(ev3)
	if v := asr.GetValue(); v != 33.33333 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev3.ID)
	if v := asr.GetValue(); v != 50.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4)
	asr.AddEvent(ev5)
	asr.RemEvent(ev.ID)
	if v := asr.GetValue(); v != 66.666670 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev2.ID)
	if v := asr.GetValue(); v != 100.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev4.ID)
	if v := asr.GetValue(); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev5.ID)
	if v := asr.GetValue(); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}
}

func TestACDGetStringValue(t *testing.T) {
	acd, _ := NewACD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			utils.Usage:  time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev)
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acd.AddEvent(ev2)
	acd.AddEvent(ev3)
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev.ID)
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4)
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev5)
	if strVal := acd.GetStringValue(""); strVal != "1m15s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetStringValue(""); strVal != "1m15s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev5.ID)
	acd.RemEvent(ev4.ID)
	acd.RemEvent(ev5.ID)
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
}

func TestACDGetFloat64Value(t *testing.T) {
	acd, _ := NewACD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	acd.AddEvent(ev)
	if v := acd.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong acd value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	acd.AddEvent(ev2)
	if v := acd.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong acd value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4)
	if strVal := acd.GetFloat64Value(); strVal != 35.0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.AddEvent(ev5)
	if strVal := acd.GetFloat64Value(); strVal != 53.333333333 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetFloat64Value(); strVal != 53.333333333 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev4.ID)
	if strVal := acd.GetFloat64Value(); strVal != 50.0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev.ID)
	if strVal := acd.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev5.ID)
	if strVal := acd.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
}

func TestACDGetValue(t *testing.T) {
	acd, _ := NewACD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	acd.AddEvent(ev)
	if v := acd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(8 * time.Second)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acd.AddEvent(ev2)
	acd.AddEvent(ev3)
	if v := acd.GetValue(); v != time.Duration(9*time.Second) {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev.ID)
	if v := acd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev2.ID)
	if v := acd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Duration(4*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4)
	acd.AddEvent(ev5)
	if v := acd.GetValue(); v != time.Duration(2*time.Minute+45*time.Second) {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev5.ID)
	acd.RemEvent(ev4.ID)
	if v := acd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev3.ID)
	if v := acd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong acd value: %+v", v)
	}
}

func TestTCDGetStringValue(t *testing.T) {
	tcd, _ := NewTCD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Usage":      time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := tcd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev)
	if strVal := tcd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"Usage":      time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	tcd.AddEvent(ev2)
	tcd.AddEvent(ev3)
	if strVal := tcd.GetStringValue(""); strVal != "20s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev.ID)
	if strVal := tcd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4)
	tcd.AddEvent(ev5)
	if strVal := tcd.GetStringValue(""); strVal != "2m30s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev4.ID)
	if strVal := tcd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev5.ID)
	tcd.RemEvent(ev3.ID)
	if strVal := tcd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestTCDGetFloat64Value(t *testing.T) {
	tcd, _ := NewTCD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	tcd.AddEvent(ev)
	if v := tcd.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	tcd.AddEvent(ev2)
	if v := tcd.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4)
	if strVal := tcd.GetFloat64Value(); strVal != 70.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.AddEvent(ev5)
	if strVal := tcd.GetFloat64Value(); strVal != 160.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetFloat64Value(); strVal != 160.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev4.ID)
	if strVal := tcd.GetFloat64Value(); strVal != 100.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev.ID)
	if strVal := tcd.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev5.ID)
	if strVal := tcd.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
}

func TestTCDGetValue(t *testing.T) {
	tcd, _ := NewTCD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	tcd.AddEvent(ev)
	if v := tcd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(5 * time.Second)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	tcd.AddEvent(ev2)
	tcd.AddEvent(ev3)
	if v := tcd.GetValue(); v != time.Duration(15*time.Second) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev.ID)
	if v := tcd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev2.ID)
	if v := tcd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4)
	tcd.AddEvent(ev5)
	if v := tcd.GetValue(); v != time.Duration(2*time.Minute+30*time.Second) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev5.ID)
	tcd.RemEvent(ev4.ID)
	if v := tcd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev3.ID)
	if v := tcd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong tcd value: %+v", v)
	}
}

func TestACCGetStringValue(t *testing.T) {
	acc, _ := NewACC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	if strVal := acc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev)
	if strVal := acc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	acc.AddEvent(ev2)
	acc.AddEvent(ev3)
	if strVal := acc.GetStringValue(""); strVal != "12.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev3.ID)
	if strVal := acc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.6}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	acc.AddEvent(ev4)
	acc.AddEvent(ev5)
	acc.RemEvent(ev.ID)
	if strVal := acc.GetStringValue(""); strVal != "3.4" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetStringValue(""); strVal != "3.4" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev4.ID)
	acc.RemEvent(ev5.ID)
	if strVal := acc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acc value: %s", strVal)
	}
}

func TestACCGetValue(t *testing.T) {
	acc, _ := NewACC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	if strVal := acc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.AddEvent(ev)
	if strVal := acc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acc.AddEvent(ev2)
	acc.AddEvent(ev3)
	if strVal := acc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev3.ID)
	if strVal := acc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "5.6"}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "1.2"}}
	acc.AddEvent(ev4)
	acc.AddEvent(ev5)
	acc.RemEvent(ev.ID)
	if strVal := acc.GetValue(); strVal != 3.4 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetValue(); strVal != 3.4 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev4.ID)
	acc.RemEvent(ev5.ID)
	if strVal := acc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
}

func TestTCCGetStringValue(t *testing.T) {
	tcc, _ := NewTCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	if strVal := tcc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev)
	if strVal := tcc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.7}}
	tcc.AddEvent(ev2)
	tcc.AddEvent(ev3)
	if strVal := tcc.GetStringValue(""); strVal != "18" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev3.ID)
	if strVal := tcc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.6}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	tcc.AddEvent(ev4)
	tcc.AddEvent(ev5)
	tcc.RemEvent(ev.ID)
	if strVal := tcc.GetStringValue(""); strVal != "6.8" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetStringValue(""); strVal != "6.8" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev4.ID)
	tcc.RemEvent(ev5.ID)
	if strVal := tcc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcc value: %s", strVal)
	}
}

func TestTCCGetValue(t *testing.T) {
	tcc, _ := NewTCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	if strVal := tcc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.AddEvent(ev)
	if strVal := tcc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	tcc.AddEvent(ev2)
	tcc.AddEvent(ev3)
	if strVal := tcc.GetValue(); strVal != 13.5 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev3.ID)
	if strVal := tcc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "5.6"}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "1.2"}}
	tcc.AddEvent(ev4)
	tcc.AddEvent(ev5)
	tcc.RemEvent(ev.ID)
	if strVal := tcc.GetValue(); strVal != 6.8 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetValue(); strVal != 6.8 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev4.ID)
	tcc.RemEvent(ev5.ID)
	if strVal := tcc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
}

func TestPDDGetStringValue(t *testing.T) {
	pdd, _ := NewPDD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			utils.Usage:  time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    time.Duration(5 * time.Second),
		}}
	if strVal := pdd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev)
	if strVal := pdd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	pdd.AddEvent(ev2)
	pdd.AddEvent(ev3)
	if strVal := pdd.GetStringValue(""); strVal != "1.666666666s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev3.ID)
	if strVal := pdd.GetStringValue(""); strVal != "2.5s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev.ID)
	if strVal := pdd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    time.Duration(10 * time.Second),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			utils.PDD: time.Duration(10 * time.Second),
		},
	}
	pdd.AddEvent(ev4)
	if strVal := pdd.GetStringValue(""); strVal != "5s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev5)
	if strVal := pdd.GetStringValue(""); strVal != "3.333333333s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetStringValue(""); strVal != "5s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev5.ID)
	pdd.RemEvent(ev4.ID)
	pdd.RemEvent(ev5.ID)
	if strVal := pdd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestPDDGetFloat64Value(t *testing.T) {
	pdd, _ := NewPDD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second),
			utils.PDD:    time.Duration(5 * time.Second)}}
	pdd.AddEvent(ev)
	if v := pdd.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong pdd value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	pdd.AddEvent(ev2)
	if v := pdd.GetFloat64Value(); v != 2.5 {
		t.Errorf("wrong pdd value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    time.Duration(10 * time.Second),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	pdd.AddEvent(ev4)
	if strVal := pdd.GetFloat64Value(); strVal != 5 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.AddEvent(ev5)
	if strVal := pdd.GetFloat64Value(); strVal != 3.75 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetFloat64Value(); strVal != 5 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev4.ID)
	if strVal := pdd.GetFloat64Value(); strVal != 2.5 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev.ID)
	if strVal := pdd.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev5.ID)
	if strVal := pdd.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
}

func TestPDDGetValue(t *testing.T) {
	pdd, _ := NewPDD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second),
			utils.PDD:    time.Duration(9 * time.Second)}}
	pdd.AddEvent(ev)
	if v := pdd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(8 * time.Second),
			utils.PDD:    time.Duration(10 * time.Second)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	pdd.AddEvent(ev2)
	pdd.AddEvent(ev3)
	if v := pdd.GetValue(); v != time.Duration(6333333333*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	pdd.RemEvent(ev.ID)
	if v := pdd.GetValue(); v != time.Duration(5*time.Second) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	pdd.RemEvent(ev2.ID)
	if v := pdd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    time.Duration(8 * time.Second),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Duration(4*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	pdd.AddEvent(ev4)
	pdd.AddEvent(ev5)
	if v := pdd.GetValue(); v != time.Duration(2666666666*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	pdd.RemEvent(ev5.ID)
	pdd.RemEvent(ev4.ID)
	if v := pdd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	pdd.RemEvent(ev3.ID)
	if v := pdd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
}

func TestDDCGetStringValue(t *testing.T) {
	ddc, _ := NewDCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := ddc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	ddc.AddEvent(ev)
	if strVal := ddc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}

	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1001"}}
	ddc.AddEvent(ev2)
	ddc.AddEvent(ev3)
	if strVal := ddc.GetStringValue(""); strVal != "2" {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev.ID)
	if strVal := ddc.GetStringValue(""); strVal != "2" {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev2.ID)
	if strVal := ddc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev3.ID)
	if strVal := ddc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong ddc value: %s", strVal)
	}
}

func TestDDCGetFloat64Value(t *testing.T) {
	ddc, _ := NewDCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	ddc.AddEvent(ev)
	if v := ddc.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong ddc value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ddc.AddEvent(ev2)
	if v := ddc.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong ddc value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":           time.Duration(1 * time.Minute),
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         time.Duration(10 * time.Second),
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":           time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	ddc.AddEvent(ev4)
	if strVal := ddc.GetFloat64Value(); strVal != 2 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.AddEvent(ev5)
	if strVal := ddc.GetFloat64Value(); strVal != 3 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev2.ID)
	if strVal := ddc.GetFloat64Value(); strVal != 3 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	ddc.RemEvent(ev4.ID)
	if strVal := ddc.GetFloat64Value(); strVal != 2 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev.ID)
	if strVal := ddc.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev5.ID)
	if strVal := ddc.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
}

func TestStatSumGetFloat64Value(t *testing.T) {
	statSum, _ := NewStatSum(2, "Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	statSum.AddEvent(ev)
	if v := statSum.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong statSum value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	statSum.AddEvent(ev2)
	if v := statSum.GetFloat64Value(); v != 20.0 {
		t.Errorf("wrong statSum value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Cost":            "20",
			"Usage":           time.Duration(1 * time.Minute),
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         time.Duration(10 * time.Second),
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Cost":            "20",
			"Usage":           time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	statSum.AddEvent(ev4)
	if strVal := statSum.GetFloat64Value(); strVal != 40 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.AddEvent(ev5)
	if strVal := statSum.GetFloat64Value(); strVal != 60 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetFloat64Value(); strVal != 60 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev4.ID)
	if strVal := statSum.GetFloat64Value(); strVal != 40 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev.ID)
	if strVal := statSum.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev5.ID)
	if strVal := statSum.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
}

func TestStatSumGetStringValue(t *testing.T) {
	statSum, _ := NewStatSum(2, "Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := statSum.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	statSum.AddEvent(ev)
	if strVal := statSum.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}

	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1001"}}
	statSum.AddEvent(ev2)
	statSum.AddEvent(ev3)
	if strVal := statSum.GetStringValue(""); strVal != "60" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev.ID)
	if strVal := statSum.GetStringValue(""); strVal != "40" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev3.ID)
	if strVal := statSum.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statSum value: %s", strVal)
	}
}

func TestStatAverageGetFloat64Value(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	statAvg.AddEvent(ev)
	if v := statAvg.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong statAvg value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	statAvg.AddEvent(ev2)
	if v := statAvg.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong statAvg value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Cost":            "30",
			"Usage":           time.Duration(1 * time.Minute),
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         time.Duration(10 * time.Second),
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Cost":            "20",
			"Usage":           time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	statAvg.AddEvent(ev4)
	if strVal := statAvg.GetFloat64Value(); strVal != 25 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.AddEvent(ev5)
	if strVal := statAvg.GetFloat64Value(); strVal != 23.33333 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetFloat64Value(); strVal != 23.33333 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev4.ID)
	if strVal := statAvg.GetFloat64Value(); strVal != 20 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev.ID)
	if strVal := statAvg.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev5.ID)
	if strVal := statAvg.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
}

func TestStatAverageGetStringValue(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := statAvg.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	statAvg.AddEvent(ev)
	if strVal := statAvg.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}

	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1001"}}
	statAvg.AddEvent(ev2)
	statAvg.AddEvent(ev3)
	if strVal := statAvg.GetStringValue(""); strVal != "20" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev.ID)
	if strVal := statAvg.GetStringValue(""); strVal != "20" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev3.ID)
	if strVal := statAvg.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
}

var jMarshaler JSONMarshaler

func TestASRMarshal(t *testing.T) {
	asr, _ := NewASR(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev)
	var nasr StatASR
	expected := []byte(`{"Answered":1,"Count":1,"Events":{"EVENT_1":true},"MinItems":2}`)
	if b, err := asr.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := nasr.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(asr, nasr) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(asr), utils.ToJSON(nasr))
	}
}

func TestACDMarshal(t *testing.T) {
	acd, _ := NewACD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	acd.AddEvent(ev)
	var nacd StatACD
	expected := []byte(`{"Sum":10000000000,"Count":1,"Events":{"EVENT_1":10000000000},"MinItems":2}`)
	if b, err := acd.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := nacd.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(acd, nacd) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(acd), utils.ToJSON(nacd))
	}
}

func TestTCDMarshal(t *testing.T) {
	tcd, _ := NewTCD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	tcd.AddEvent(ev)
	var ntcd StatTCD
	expected := []byte(`{"Sum":10000000000,"Count":1,"Events":{"EVENT_1":10000000000},"MinItems":2}`)
	if b, err := tcd.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := ntcd.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(tcd, ntcd) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(tcd), utils.ToJSON(ntcd))
	}
}

func TestACCMarshal(t *testing.T) {
	acc, _ := NewACC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	acc.AddEvent(ev)
	var nacc StatACC
	expected := []byte(`{"Sum":12.3,"Count":1,"Events":{"EVENT_1":12.3},"MinItems":2}`)
	if b, err := acc.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := nacc.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(acc, nacc) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(acc), utils.ToJSON(nacc))
	}
}

func TestTCCMarshal(t *testing.T) {
	tcc, _ := NewTCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	tcc.AddEvent(ev)
	var ntcc StatTCC
	expected := []byte(`{"Sum":12.3,"Count":1,"Events":{"EVENT_1":12.3},"MinItems":2}`)
	if b, err := tcc.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := ntcc.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(tcc, ntcc) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(tcc), utils.ToJSON(ntcc))
	}
}

func TestPDDMarshal(t *testing.T) {
	pdd, _ := NewPDD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second),
			utils.PDD:    time.Duration(5 * time.Second)}}
	pdd.AddEvent(ev)
	var ntdd StatPDD
	expected := []byte(`{"Sum":5000000000,"Count":1,"Events":{"EVENT_1":5000000000},"MinItems":2}`)
	if b, err := pdd.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := ntdd.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(pdd, ntdd) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(pdd), utils.ToJSON(ntdd))
	}
}

func TestDCCMarshal(t *testing.T) {
	ddc, _ := NewDCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	ddc.AddEvent(ev)
	var nddc StatDDC
	expected := []byte(`{"Destinations":{"1002":{"EVENT_1":true}},"Events":{"EVENT_1":"1002"},"MinItems":2}`)
	if b, err := ddc.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := nddc.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(ddc, nddc) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(ddc), utils.ToJSON(nddc))
	}
}

func TestStatSumMarshal(t *testing.T) {
	statSum, _ := NewStatSum(2, "Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	statSum.AddEvent(ev)
	var nstatSum StatSum
	expected := []byte(`{"Sum":20,"Events":{"EVENT_1":20},"MinItems":2,"FieldName":"Cost"}`)
	if b, err := statSum.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := nstatSum.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(statSum, nstatSum) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(statSum), utils.ToJSON(nstatSum))
	}
}

func TestStatAverageMarshal(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	statAvg.AddEvent(ev)
	var nstatAvg StatAverage
	expected := []byte(`{"Sum":20,"Count":1,"Events":{"EVENT_1":20},"MinItems":2,"FieldName":"Cost"}`)
	if b, err := statAvg.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := nstatAvg.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(statAvg, nstatAvg) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(statAvg), utils.ToJSON(nstatAvg))
	}
}
