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
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4)
	asr.AddEvent(ev5)
	asr.RemEvent(ev.TenantID())
	if strVal := asr.GetStringValue(""); strVal != "66.66667%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev2.TenantID())
	if strVal := asr.GetStringValue(""); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev4.TenantID())
	asr.RemEvent(ev5.TenantID())
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
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4)
	asr.AddEvent(ev5)
	asr.RemEvent(ev.TenantID())
	if v := asr.GetValue(); v != 66.666670 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev2.TenantID())
	if v := asr.GetValue(); v != 100.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev4.TenantID())
	if v := asr.GetValue(); v != 100.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev5.TenantID())
	if v := asr.GetValue(); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}
}

func TestACDGetStringValue(t *testing.T) {
	acd, _ := NewACD()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			utils.USAGE:  time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev)
	if strVal := acd.GetStringValue(""); strVal != "10s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acd.AddEvent(ev2)
	acd.AddEvent(ev3)
	if strVal := acd.GetStringValue(""); strVal != "3.333333333s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev3.TenantID())
	if strVal := acd.GetStringValue(""); strVal != "5s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev.TenantID())
	if strVal := acd.GetStringValue(""); strVal != "0s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4)
	if strVal := acd.GetStringValue(""); strVal != "30s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev5)
	if strVal := acd.GetStringValue(""); strVal != "50s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.TenantID())
	if strVal := acd.GetStringValue(""); strVal != "1m15s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev5.TenantID())
	acd.RemEvent(ev4.TenantID())
	acd.RemEvent(ev5.TenantID())
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
}

func TestACDGetFloat64Value(t *testing.T) {
	acd, _ := NewACD()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	acd.AddEvent(ev)
	if v := acd.GetFloat64Value(); v != 10.0 {
		t.Errorf("wrong acd value: %f", v)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	acd.AddEvent(ev2)
	if v := acd.GetFloat64Value(); v != 5.0 {
		t.Errorf("wrong acd value: %f", v)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4)
	if strVal := acd.GetFloat64Value(); strVal != 23.333333333 {
		t.Errorf("wrong acd value: %f", strVal)
	}
	acd.AddEvent(ev5)
	if strVal := acd.GetFloat64Value(); strVal != 40.0 {
		t.Errorf("wrong acd value: %f", strVal)
	}
	acd.RemEvent(ev2.TenantID())
	if strVal := acd.GetFloat64Value(); strVal != 53.333333333 {
		t.Errorf("wrong acd value: %f", strVal)
	}
	acd.RemEvent(ev4.TenantID())
	if strVal := acd.GetFloat64Value(); strVal != 50.0 {
		t.Errorf("wrong acd value: %f", strVal)
	}
	acd.RemEvent(ev.TenantID())
	if strVal := acd.GetFloat64Value(); strVal != 90.0 {
		t.Errorf("wrong acd value: %f", strVal)
	}
	acd.RemEvent(ev5.TenantID())
	if strVal := acd.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong acd value: %f", strVal)
	}
}

func TestACDGetValue(t *testing.T) {
	acd, _ := NewACD()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	acd.AddEvent(ev)
	if v := acd.GetValue(); v != time.Duration(10*time.Second) {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(8 * time.Second)}}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acd.AddEvent(ev2)
	acd.AddEvent(ev3)
	if v := acd.GetValue(); v != time.Duration(6*time.Second) {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev.TenantID())
	if v := acd.GetValue(); v != time.Duration(4*time.Second) {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev2.TenantID())
	if v := acd.GetValue(); v != time.Duration(0*time.Second) {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(4*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4)
	acd.AddEvent(ev5)
	if v := acd.GetValue(); v != time.Duration(1*time.Minute+50*time.Second) {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev5.TenantID())
	acd.RemEvent(ev4.TenantID())
	if v := acd.GetValue(); v != time.Duration(0*time.Second) {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev3.TenantID())
	if v := acd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong acd value: %+v", v)
	}
}

func TestTCDGetStringValue(t *testing.T) {
	tcd, _ := NewTCD()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := tcd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev)
	if strVal := tcd.GetStringValue(""); strVal != "10s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	tcd.AddEvent(ev2)
	tcd.AddEvent(ev3)
	if strVal := tcd.GetStringValue(""); strVal != "20s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev2.TenantID())
	if strVal := tcd.GetStringValue(""); strVal != "10s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev.TenantID())
	if strVal := tcd.GetStringValue(""); strVal != "0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4)
	tcd.AddEvent(ev5)
	if strVal := tcd.GetStringValue(""); strVal != "2m30s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev4.TenantID())
	if strVal := tcd.GetStringValue(""); strVal != "1m30s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev5.TenantID())
	tcd.RemEvent(ev3.TenantID())
	if strVal := tcd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestTCDGetFloat64Value(t *testing.T) {
	tcd, _ := NewTCD()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	tcd.AddEvent(ev)
	if v := tcd.GetFloat64Value(); v != 10.0 {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	tcd.AddEvent(ev2)
	if v := tcd.GetFloat64Value(); v != 10.0 {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
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
	tcd.RemEvent(ev2.TenantID())
	if strVal := tcd.GetFloat64Value(); strVal != 160.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev4.TenantID())
	if strVal := tcd.GetFloat64Value(); strVal != 100.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev.TenantID())
	if strVal := tcd.GetFloat64Value(); strVal != 90.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev5.TenantID())
	if strVal := tcd.GetFloat64Value(); strVal != -1.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
}

func TestTCDGetValue(t *testing.T) {
	tcd, _ := NewTCD()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	tcd.AddEvent(ev)
	if v := tcd.GetValue(); v != time.Duration(10*time.Second) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(5 * time.Second)}}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	tcd.AddEvent(ev2)
	tcd.AddEvent(ev3)
	if v := tcd.GetValue(); v != time.Duration(15*time.Second) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev.TenantID())
	if v := tcd.GetValue(); v != time.Duration(5*time.Second) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev2.TenantID())
	if v := tcd.GetValue(); v != time.Duration(0*time.Second) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1 * time.Minute),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"Usage":      time.Duration(1*time.Minute + 30*time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4)
	tcd.AddEvent(ev5)
	if v := tcd.GetValue(); v != time.Duration(2*time.Minute+30*time.Second) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev5.TenantID())
	tcd.RemEvent(ev4.TenantID())
	if v := tcd.GetValue(); v != time.Duration(0*time.Second) {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev3.TenantID())
	if v := tcd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong tcd value: %+v", v)
	}
}

func TestACCGetStringValue(t *testing.T) {
	acc, _ := NewACC()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	if strVal := acc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev)
	if strVal := acc.GetStringValue(""); strVal != "12.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acc.AddEvent(ev2)
	acc.AddEvent(ev3)
	if strVal := acc.GetStringValue(""); strVal != "4.1" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev3.TenantID())
	if strVal := acc.GetStringValue(""); strVal != "6.15" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.6}}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	acc.AddEvent(ev4)
	acc.AddEvent(ev5)
	acc.RemEvent(ev.TenantID())
	if strVal := acc.GetStringValue(""); strVal != "2.26667" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev2.TenantID())
	if strVal := acc.GetStringValue(""); strVal != "3.4" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev4.TenantID())
	acc.RemEvent(ev5.TenantID())
	if strVal := acc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acc value: %s", strVal)
	}
}

func TestACCGetValue(t *testing.T) {
	acc, _ := NewACC()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	if strVal := acc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.AddEvent(ev)
	if strVal := acc.GetValue(); strVal != 12.3 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acc.AddEvent(ev2)
	acc.AddEvent(ev3)
	if strVal := acc.GetValue(); strVal != 4.1 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev3.TenantID())
	if strVal := acc.GetValue(); strVal != 6.15 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "5.6"}}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "1.2"}}
	acc.AddEvent(ev4)
	acc.AddEvent(ev5)
	acc.RemEvent(ev.TenantID())
	if strVal := acc.GetValue(); strVal != 2.26667 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev2.TenantID())
	if strVal := acc.GetValue(); strVal != 3.4 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev4.TenantID())
	acc.RemEvent(ev5.TenantID())
	if strVal := acc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
}

func TestTCCGetStringValue(t *testing.T) {
	tcc, _ := NewTCC()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	if strVal := tcc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev)
	if strVal := tcc.GetStringValue(""); strVal != "12.3" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.7}}
	tcc.AddEvent(ev2)
	tcc.AddEvent(ev3)
	if strVal := tcc.GetStringValue(""); strVal != "18" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev3.TenantID())
	if strVal := tcc.GetStringValue(""); strVal != "12.3" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.6}}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	tcc.AddEvent(ev4)
	tcc.AddEvent(ev5)
	tcc.RemEvent(ev.TenantID())
	if strVal := tcc.GetStringValue(""); strVal != "6.8" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev2.TenantID())
	if strVal := tcc.GetStringValue(""); strVal != "6.8" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev4.TenantID())
	tcc.RemEvent(ev5.TenantID())
	if strVal := tcc.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong tcc value: %s", strVal)
	}
}

func TestTCCGetValue(t *testing.T) {
	tcc, _ := NewTCC()
	ev := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	if strVal := tcc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.AddEvent(ev)
	if strVal := tcc.GetValue(); strVal != 12.3 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev2 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	tcc.AddEvent(ev2)
	tcc.AddEvent(ev3)
	if strVal := tcc.GetValue(); strVal != 13.5 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev3.TenantID())
	if strVal := tcc.GetValue(); strVal != 12.3 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev4 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "5.6"}}
	ev5 := &StatEvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Fields: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "1.2"}}
	tcc.AddEvent(ev4)
	tcc.AddEvent(ev5)
	tcc.RemEvent(ev.TenantID())
	if strVal := tcc.GetValue(); strVal != 6.8 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev2.TenantID())
	if strVal := tcc.GetValue(); strVal != 6.8 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev4.TenantID())
	tcc.RemEvent(ev5.TenantID())
	if strVal := tcc.GetValue(); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
}
