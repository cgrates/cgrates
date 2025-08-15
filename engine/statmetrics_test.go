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
	"encoding/json"
	"math"
	"net"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
)

func TestASRGetStringValue(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "33.33333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev3.ID)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	asr.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	asr.RemEvent(ev.ID)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "66.66667%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev2.ID)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev4.ID)
	asr.RemEvent(ev5.ID)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestASRGetStringValue2(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "33.33333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "25%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev4.ID)
	asr.RemEvent(ev2.ID)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev2.ID)
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestASRGetStringValue3(t *testing.T) {
	asr := &StatASR{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatASR{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 1, CompressFactor: 1},
			"EVENT_2": {Stat: 0, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Answered:  1,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_1"]
	v.Stat = 0.5
	v.CompressFactor = 2
	v = expected.Events["EVENT_2"]
	v.Stat = 0
	v.CompressFactor = 2
	expected.Count = 4
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "25%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.RemEvent(ev4.ID)
	asr.RemEvent(ev2.ID)
	v = expected.Events["EVENT_1"]
	v.Stat = 1
	v.CompressFactor = 1
	v = expected.Events["EVENT_2"]
	v.Stat = 0
	v.CompressFactor = 1
	expected.Count = 2
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
}

func TestASRGetValue(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := asr.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	asr.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if v := asr.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 33.33333 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev3.ID)
	if v := asr.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 50.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	asr.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	asr.RemEvent(ev.ID)
	if v := asr.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 66.666670 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev2.ID)
	if v := asr.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 100.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev4.ID)
	if v := asr.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev5.ID)
	if v := asr.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong asr value: %f", v)
	}
}

func TestASRCompress(t *testing.T) {
	asr := &StatASR{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatASR{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 1, CompressFactor: 1},
			"EVENT_2": {Stat: 0, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Answered:  1,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := asr.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	expected = &StatASR{
		Events: map[string]*StatWithCompress{
			"EVENT_3": {Stat: 0.5, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Answered:  1,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := asr.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = 0.25
	v.CompressFactor = 4
	expected.Count = 4
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if rply := asr.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "25%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
}

func TestASRGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	asr, _ := NewASR(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = asr.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = asr.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = asr.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestASRAddOneEvent(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	asr.AddOneEvent(utils.MapStorage{utils.MetaReq: ev.Event})
	asr.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "33.33333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "25%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestACDGetStringValue(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			utils.Usage:  10 * time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	if err := acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event}); err != nil {
		t.Error(err)
	}
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	if err := acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err == nil || err.Error() != "NOT_FOUND:Usage" {
		t.Error(err)
	}
	if err := acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err == nil || err.Error() != "NOT_FOUND:Usage" {
		t.Error(err)
	}
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev.ID)
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      478433753 * time.Nanosecond,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":      30*time.Second + 982433452*time.Nanosecond,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.73043s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.73043s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev5.ID)
	acd.RemEvent(ev4.ID)
	acd.RemEvent(ev5.ID)
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
}

func TestACDGetStringValue2(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}
	if err := acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Usage": time.Minute}}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m30s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m15s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m20s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
}

func TestACDGetStringValue3(t *testing.T) {
	acd := &StatACD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatACD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Usage": time.Minute}}
	if err := acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = 3*time.Minute + 30*time.Second
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	acd.RemEvent(ev1.ID)
	acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
}

func TestACDCompress(t *testing.T) {
	acd := &StatACD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatACD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Usage": time.Minute}}
	acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := acd.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
	expected = &StatACD{
		Events: map[string]*DurationWithCompress{
			"EVENT_3": {Duration: 2 * time.Minute, CompressFactor: 3},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)

	expIDs = []string{"EVENT_3"}
	if rply := acd.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
}

func TestACDGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	acd, _ := NewACD(2, "", []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Usage": time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Usage": time.Minute}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}

	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = acd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = acd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = acd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestACDGetFloat64Value(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong acd value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong acd value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 35.0*1e9 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	// by default rounding decimal is 5
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 53.33333*1e9 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	// test for other rounding decimals
	config.CgrConfig().GeneralCfg().RoundingDecimals = 0
	acd.(*StatACD).val = nil
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 53*1e9 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	config.CgrConfig().GeneralCfg().RoundingDecimals = 1
	acd.(*StatACD).val = nil
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 53.3*1e9 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	config.CgrConfig().GeneralCfg().RoundingDecimals = 9
	acd.(*StatACD).val = nil
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 53.333333333*1e9 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	config.CgrConfig().GeneralCfg().RoundingDecimals = -1
	acd.(*StatACD).val = nil
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 50*1e9 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	//change back the rounding decimals to default value
	config.CgrConfig().GeneralCfg().RoundingDecimals = 5
	acd.(*StatACD).val = nil
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 53.33333*1e9 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev4.ID)
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 50.0*1e9 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev.ID)
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev5.ID)
	if strVal := acd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
}

func TestACDGetValue(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := acd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      8 * time.Second}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if v := acd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 9*time.Second {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev.ID)
	if v := acd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev2.ID)
	if v := acd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":      4*time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if v := acd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 2*time.Minute+45*time.Second {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev5.ID)
	acd.RemEvent(ev4.ID)
	if v := acd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev3.ID)
	if v := acd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong acd value: %+v", v)
	}
}

func TestACDAddOneEvent(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}
	if err := acd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Usage": time.Minute}}
	acd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m30s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	acd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m15s" {
		t.Errorf("wrong acd value: %s", strVal)
	}

}

func TestTCDGetStringValue(t *testing.T) {
	tcd, _ := NewTCD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"Usage":      10 * time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{
			"Usage":      10 * time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "20s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev.ID)
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2m30s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev4.ID)
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev5.ID)
	tcd.RemEvent(ev3.ID)
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestTCDGetStringValue2(t *testing.T) {
	tcd, _ := NewTCD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}
	if err := tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Usage": time.Minute}}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "3m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "5m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "4m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestTCDGetStringValue3(t *testing.T) {
	tcd := &StatTCD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatTCD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Usage": time.Minute}}
	if err := tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = 3*time.Minute + 30*time.Second
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	tcd.RemEvent(ev1.ID)
	tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
}

func TestTCDGetFloat64Value(t *testing.T) {
	tcd, _ := NewTCD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := tcd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := tcd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := tcd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 70.0*1e9 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := tcd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 160.0*1e9 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 160.0*1e9 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev4.ID)
	if strVal := tcd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 100.0*1e9 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev.ID)
	if strVal := tcd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev5.ID)
	if strVal := tcd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
}

func TestTCDGetValue(t *testing.T) {
	tcd, _ := NewTCD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := tcd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      5 * time.Second}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if v := tcd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 15*time.Second {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev.ID)
	if v := tcd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev2.ID)
	if v := tcd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if v := tcd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 2*time.Minute+30*time.Second {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev5.ID)
	tcd.RemEvent(ev4.ID)
	if v := tcd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev3.ID)
	if v := tcd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong tcd value: %+v", v)
	}
}

func TestTCDCompress(t *testing.T) {
	tcd := &StatTCD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatTCD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Usage": time.Minute}}
	tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := tcd.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
	expected = &StatTCD{
		Events: map[string]*DurationWithCompress{
			"EVENT_3": {Duration: 2 * time.Minute, CompressFactor: 3},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)

	expIDs = []string{"EVENT_3"}
	if rply := tcd.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
}

func TestTCDGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	tcd, _ := NewTCD(2, "", []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Usage": time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Usage": time.Minute}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}

	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = tcd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = tcd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = tcd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestTCDAddOneEvent(t *testing.T) {
	tcd, _ := NewTCD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Usage: 2 * time.Minute}}
	if err := tcd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Usage": time.Minute}}
	tcd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "3m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "5m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestACCGetStringValue(t *testing.T) {
	acc, _ := NewACC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "12.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev3.ID)
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.6}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	acc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	acc.RemEvent(ev.ID)
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "3.4" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "3.4" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev4.ID)
	acc.RemEvent(ev5.ID)
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	expErr := "NEGATIVE:Cost"
	if err := acc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: utils.MapStorage{
		utils.Cost: -1,
	}}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s received %v", expErr, err)
	}
}

func TestACCGetStringValue2(t *testing.T) {
	acc, _ := NewACC(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 12.3}}
	if err := acc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.3}}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.8" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
}

func TestACCGetStringValue3(t *testing.T) {
	acc := &StatACC{Events: make(map[string]*StatWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatACC{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 12.2, CompressFactor: 2},
			"EVENT_3": {Stat: 18.3, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       42.7,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Cost": 18.3}}
	if err := acc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = expected.Sum - 12.2
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	acc.RemEvent(ev1.ID)
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
}

func TestACCGetValue(t *testing.T) {
	acc, _ := NewACC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	if strVal := acc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := acc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := acc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev3.ID)
	if strVal := acc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "5.6"}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "1.2"}}
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	acc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	acc.RemEvent(ev.ID)
	if strVal := acc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 3.4 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 3.4 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev4.ID)
	acc.RemEvent(ev5.ID)
	if strVal := acc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
}

func TestACCCompress(t *testing.T) {
	acc := &StatACC{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatACC{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 18.2, CompressFactor: 1},
			"EVENT_2": {Stat: 6.2, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.3}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := acc.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	expected = &StatACC{
		Events: map[string]*StatWithCompress{
			"EVENT_3": {Stat: 12.2, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := acc.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = 12.225
	v.CompressFactor = 4
	expected.Count = 4
	expected.Sum = 48.9
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if rply := acc.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
}

func TestACCGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	acc, _ := NewACC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = acc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = acc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = acc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}
func TestACCAddOneEvent(t *testing.T) {
	acc, _ := NewACC(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 12.3}}
	if err := acc.AddOneEvent(utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.3}}
	acc.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.8" {
		t.Errorf("wrong acc value: %s", strVal)
	}
}

func TestTCCGetStringValue(t *testing.T) {
	tcc, _ := NewTCC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.7}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "18" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev3.ID)
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.6}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	tcc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	tcc.RemEvent(ev.ID)
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "6.8" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "6.8" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev4.ID)
	tcc.RemEvent(ev5.ID)
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}

	expErr := "NEGATIVE:Cost"
	if err := tcc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: utils.MapStorage{
		utils.Cost: -1,
	}}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s received %v", expErr, err)
	}
}

func TestTCCGetStringValue2(t *testing.T) {
	tcc, _ := NewTCC(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 12.3}}
	if err := tcc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.3}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "30.6" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "67.2" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "48.9" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
}

func TestTCCGetStringValue3(t *testing.T) {
	tcc := &StatTCC{Events: make(map[string]*StatWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatTCC{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 12.2, CompressFactor: 2},
			"EVENT_3": {Stat: 18.3, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       42.7,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Cost": 18.3}}
	if err := tcc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = expected.Sum - 12.2
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	tcc.RemEvent(ev1.ID)
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
}

func TestTCCGetValue(t *testing.T) {
	tcc, _ := NewTCC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	if strVal := tcc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := tcc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := tcc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 13.5 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev3.ID)
	if strVal := tcc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "5.6"}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "1.2"}}
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	tcc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	tcc.RemEvent(ev.ID)
	if strVal := tcc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 6.8 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 6.8 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev4.ID)
	tcc.RemEvent(ev5.ID)
	if strVal := tcc.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
}

func TestTCCCompress(t *testing.T) {
	tcc := &StatTCC{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatTCC{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 18.2, CompressFactor: 1},
			"EVENT_2": {Stat: 6.2, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.3}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := tcc.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	expected = &StatTCC{
		Events: map[string]*StatWithCompress{
			"EVENT_3": {Stat: 12.2, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := tcc.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = 12.225
	v.CompressFactor = 4
	expected.Count = 4
	expected.Sum = 48.9
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if rply := tcc.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
}

func TestTCCGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	tcc, _ := NewTCC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = tcc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = tcc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = tcc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestTCCAddOneEvent(t *testing.T) {
	tcc, _ := NewTCC(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 12.3}}
	if err := tcc.AddOneEvent(utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.3}}
	tcc.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "30.6" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "67.2" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
}

func TestPDDGetStringValue(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			utils.Usage:  10 * time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    5 * time.Second,
		}}
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev3.ID)
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev.ID)
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    10 * time.Second,
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			utils.PDD: 10 * time.Second,
		},
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "10s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "10s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev5.ID)
	pdd.RemEvent(ev4.ID)
	pdd.RemEvent(ev5.ID)
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestPDDGetStringValue2(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.PDD: 2 * time.Minute}}
	if err := pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.PDD: time.Minute}}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m30s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m15s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m20s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestPDDGetStringValue3(t *testing.T) {
	pdd := &StatPDD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatPDD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.PDD: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.PDD: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{utils.PDD: time.Minute}}
	if err := pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = 3*time.Minute + 30*time.Second
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	pdd.RemEvent(ev1.ID)
	pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
}

func TestPDDGetFloat64Value(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second,
			utils.PDD:    5 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := pdd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong pdd value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := pdd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong pdd value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    10 * time.Second,
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := pdd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 7.5*1e9 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := pdd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 7.5*1e9 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 7.5*1e9 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev4.ID)
	if strVal := pdd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev.ID)
	if strVal := pdd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev5.ID)
	if strVal := pdd.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
}

func TestPDDGetValue(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second,
			utils.PDD:    9 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := pdd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong pdd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      8 * time.Second,
			utils.PDD:    10 * time.Second}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	if err := pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err == nil || err.Error() != "NOT_FOUND:PDD" {
		t.Error(err)
	}
	if v := pdd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != 9*time.Second+500*time.Millisecond {
		t.Errorf("wrong pdd value: %+v", v)
	}
	pdd.RemEvent(ev.ID)
	if v := pdd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong pdd value: %+v", v)
	}
	pdd.RemEvent(ev2.ID)
	if v := pdd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong pdd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    8 * time.Second,
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":      4*time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event}); err == nil || err.Error() != "NOT_FOUND:PDD" {
		t.Error(err)
	}
	if v := pdd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong pdd value: %+v", v)
	}
	pdd.RemEvent(ev5.ID)
	pdd.RemEvent(ev4.ID)
	if v := pdd.GetValue(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -time.Nanosecond {
		t.Errorf("wrong pdd value: %+v", v)
	}
}

func TestPDDCompress(t *testing.T) {
	pdd := &StatPDD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatPDD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.PDD: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.PDD: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{utils.PDD: time.Minute}}
	pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := pdd.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
	expected = &StatPDD{
		Events: map[string]*DurationWithCompress{
			"EVENT_3": {Duration: 2 * time.Minute, CompressFactor: 3},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)

	expIDs = []string{"EVENT_3"}
	if rply := pdd.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
}

func TestPDDGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	pdd, _ := NewPDD(2, "", []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.PDD: time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.PDD: time.Minute}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.PDD: 2 * time.Minute}}

	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = pdd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = pdd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = pdd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestPDDAddOneEvent(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.PDD: 2 * time.Minute}}
	if err := pdd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.PDD: time.Minute}}
	pdd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m30s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	pdd.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m15s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestDDCGetStringValue(t *testing.T) {
	ddc, _ := NewDDC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}

	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1001"}}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2" {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev.ID)
	if strVal := ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2" {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev2.ID)
	if strVal := ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev3.ID)
	if strVal := ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}
}

func TestDDCGetFloat64Value(t *testing.T) {
	ddc, _ := NewDDC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := ddc.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong ddc value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := ddc.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong ddc value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage":           time.Minute,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         10 * time.Second,
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage":           time.Minute + 30*time.Second,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := ddc.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 2 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := ddc.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 3 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev2.ID)
	if strVal := ddc.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 3 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	ddc.RemEvent(ev4.ID)
	if strVal := ddc.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 2 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev.ID)
	if strVal := ddc.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev5.ID)
	if strVal := ddc.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
}

func TestDDCGetStringValue2(t *testing.T) {
	statDistinct, _ := NewDDC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1001"}}
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1002"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestDDCCompress(t *testing.T) {
	ddc := &StatDDC{
		Events:      make(map[string]map[string]int64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    2,
		FilterIDs:   []string{},
	}
	expected := &StatDDC{
		Events: map[string]map[string]int64{
			"EVENT_1": {
				"1001": 2,
			},
			"EVENT_3": {
				"1002": 1,
			},
		},
		FieldValues: map[string]utils.StringSet{
			"1001": {
				"EVENT_1": {},
			},
			"1002": {
				"EVENT_3": {},
			},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1001"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1001"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{utils.Destination: "1002"}}
	ddc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := ddc.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
	rply = ddc.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
}

func TestDDCGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	ddc, _ := NewDDC(2, "", []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1002"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.Destination: "1001"}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.Destination: "1001"}}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = ddc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = ddc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = ddc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestDDCAddOneEvent(t *testing.T) {
	statDistinct, _ := NewDDC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1001"}}
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddOneEvent(utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1002"}}
	statDistinct.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestStatSumGetFloat64Value(t *testing.T) {
	statSum, _ := NewStatSum(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := statSum.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong statSum value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	if err := statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err == nil || err.Error() != "NOT_FOUND:~*req.Cost" {
		t.Error(err)
	}
	if v := statSum.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong statSum value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Cost":            "20",
			"Usage":           time.Minute,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         10 * time.Second,
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Cost":            "20",
			"Usage":           time.Minute + 30*time.Second,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	statSum.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := statSum.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 40 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := statSum.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 60 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 60 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev4.ID)
	if strVal := statSum.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 40 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev.ID)
	if strVal := statSum.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev5.ID)
	if strVal := statSum.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
}

func TestStatSumGetStringValue(t *testing.T) {
	statSum, _ := NewStatSum(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}

	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1001"}}
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statSum.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "60" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev.ID)
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "40" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev3.ID)
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statSum value: %s", strVal)
	}
}

func TestStatSumGetStringValue2(t *testing.T) {
	statSum, _ := NewStatSum(2, "~*req.Cost", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 12.3}}
	if err := statSum.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.3}}
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "30.6" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "67.2" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "48.9" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
}

func TestStatSumGetStringValue3(t *testing.T) {
	statSum := &StatSum{Events: make(map[string]*StatWithCompress), MinItems: 2, FilterIDs: []string{}, FieldName: "~*req.Cost"}
	expected := &StatSum{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 12.2, CompressFactor: 2},
			"EVENT_3": {Stat: 18.3, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: "~*req.Cost",
		Count:     3,
		Sum:       42.7,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Cost": 18.3}}
	if err := statSum.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := statSum.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *statSum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statSum))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = expected.Sum - 12.2
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	statSum.RemEvent(ev1.ID)
	statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *statSum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statSum))
	}
}

func TestStatSumCompress(t *testing.T) {
	sum := &StatSum{Events: make(map[string]*StatWithCompress), FieldName: "~*req.Cost",
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatSum{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 18.2, CompressFactor: 1},
			"EVENT_2": {Stat: 6.2, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		FieldName: "~*req.Cost",
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.3}}
	sum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := sum.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
	expected = &StatSum{
		Events: map[string]*StatWithCompress{
			"EVENT_3": {Stat: 12.2, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: "~*req.Cost",
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := sum.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	sum.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = 12.225
	v.CompressFactor = 4
	expected.Count = 4
	expected.Sum = 48.9
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if rply := sum.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
}

func TestStatSumGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	sum, _ := NewStatSum(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	sum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = sum.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = sum.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	sum.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = sum.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatSumAddOneEvent(t *testing.T) {
	statSum, _ := NewStatSum(2, "~*req.Cost", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 12.3}}
	if err := statSum.AddOneEvent(utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.3}}
	statSum.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "30.6" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	statSum.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "67.2" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
}

func TestStatAverageGetFloat64Value(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := statAvg.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong statAvg value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := statAvg.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong statAvg value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Cost":            "30",
			"Usage":           time.Minute,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         10 * time.Second,
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Cost":            "20",
			"Usage":           time.Minute + 30*time.Second,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	statAvg.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := statAvg.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 25 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := statAvg.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 23.33333 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 23.33333 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev4.ID)
	if strVal := statAvg.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 20 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev.ID)
	if strVal := statAvg.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev5.ID)
	if strVal := statAvg.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1.0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
}

func TestStatAverageGetStringValue(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}

	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1001"}}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statAvg.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "20" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev.ID)
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "20" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev3.ID)
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
}

func TestStatAverageGetStringValue2(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "~*req.Cost", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 12.3}}
	if err := statAvg.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.3}}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.8" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
}

func TestStatAverageGetStringValue3(t *testing.T) {
	statAvg := &StatAverage{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}, FieldName: "~*req.Cost"}
	expected := &StatAverage{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 12.2, CompressFactor: 2},
			"EVENT_3": {Stat: 18.3, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: "~*req.Cost",
		Count:     3,
		Sum:       42.7,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Cost": 18.3}}
	if err := statAvg.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := statAvg.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *statAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statAvg))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = expected.Sum - 12.2
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	statAvg.RemEvent(ev1.ID)
	statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *statAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statAvg))
	}
}

func TestStatAverageCompress(t *testing.T) {
	avg := &StatAverage{Events: make(map[string]*StatWithCompress), FieldName: "~*req.Cost",
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatAverage{
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 18.2, CompressFactor: 1},
			"EVENT_2": {Stat: 6.2, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		FieldName: "~*req.Cost",
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.3}}
	avg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := avg.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
	expected = &StatAverage{
		Events: map[string]*StatWithCompress{
			"EVENT_3": {Stat: 12.2, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: "~*req.Cost",
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := avg.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	avg.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = 12.225
	v.CompressFactor = 4
	expected.Count = 4
	expected.Sum = 48.9
	expected.val = nil
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if rply := avg.Compress(1, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
}

func TestStatAverageGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	avg, _ := NewStatAverage(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 18.2}}
	avg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = avg.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = avg.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	avg.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = avg.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatAverageAddOneEvent(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "~*req.Cost", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": 12.3}}
	if err := statAvg.AddOneEvent(utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": 18.3}}
	statAvg.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	statAvg.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.8" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
}

func TestStatDistinctGetFloat64Value(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "~*req.Usage", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Usage": 10 * time.Second}}
	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := statDistinct.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong statDistinct value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := statDistinct.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); v != -1.0 {
		t.Errorf("wrong statDistinct value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]any{
			"Usage": time.Minute,
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]any{
			"Usage": time.Minute + 30*time.Second,
		},
	}
	if err := statDistinct.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event}); err != nil {
		t.Error(err)
	}
	if strVal := statDistinct.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 2 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := statDistinct.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 3 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev2.ID)
	if strVal := statDistinct.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 3 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev4.ID)
	if strVal := statDistinct.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != 2 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev5.ID)
	if strVal := statDistinct.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != -1 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
}

func TestStatDistinctGetStringValue(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": "20"}}
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{"Cost": "20"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{"Cost": "40"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statDistinct.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev2.ID)
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev3.ID)
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestStatDistinctGetStringValue2(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": "20"}}
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": "40"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestStatDistinctCompress(t *testing.T) {
	ddc := &StatDistinct{
		Events:      make(map[string]map[string]int64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    2,
		FilterIDs:   []string{},
		FieldName:   utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
	}
	expected := &StatDistinct{
		Events: map[string]map[string]int64{
			"EVENT_1": {
				"1001": 2,
			},
			"EVENT_3": {
				"1002": 1,
			},
		},
		FieldValues: map[string]utils.StringSet{
			"1001": {
				"EVENT_1": {},
			},
			"1002": {
				"EVENT_3": {},
			},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
		Count:     3,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1001"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1001"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]any{utils.Destination: "1002"}}
	ddc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := ddc.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
	rply = ddc.Compress(10, "EVENT_3", config.CgrConfig().GeneralCfg().RoundingDecimals)
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
}

func TestStatDistinctGetCompressFactor(t *testing.T) {
	var CF map[string]int
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	ddc, _ := NewStatDistinct(2, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Destination, []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{utils.Destination: "1002"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.Destination: "1001"}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]any{utils.Destination: "1001"}}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = ddc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = ddc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = ddc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatDistinctAddOneEvent(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": "20"}}
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddOneEvent(utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{"Cost": "40"}}
	statDistinct.AddOneEvent(utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

var jMarshaler JSONMarshaler

func TestASRMarshal(t *testing.T) {
	asr, _ := NewASR(2, "", []string{"*string:Account:1001"})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nasr StatASR
	expected := []byte(`{"FilterIDs":["*string:Account:1001"],"Answered":1,"Count":1,"Events":{"EVENT_1":{"Stat":1,"CompressFactor":1}},"MinItems":2}`)
	if b, err := asr.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := nasr.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(asr, nasr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(asr), utils.ToJSON(nasr))
	}
}

func TestACDMarshal(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nacd StatACD
	expected := []byte(`{"FilterIDs":[],"Sum":10000000000,"Count":1,"Events":{"EVENT_1":{"Duration":10000000000,"CompressFactor":1}},"MinItems":2}`)
	if b, err := acd.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := nacd.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(acd, nacd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(acd), utils.ToJSON(nacd))
	}
}

func TestTCDMarshal(t *testing.T) {
	tcd, _ := NewTCD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var ntcd StatTCD
	expected := []byte(`{"FilterIDs":[],"Sum":10000000000,"Count":1,"Events":{"EVENT_1":{"Duration":10000000000,"CompressFactor":1}},"MinItems":2}`)
	if b, err := tcd.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := ntcd.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(tcd, ntcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(tcd), utils.ToJSON(ntcd))
	}
}

func TestACCMarshal(t *testing.T) {
	acc, _ := NewACC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nacc StatACC
	expected := []byte(`{"FilterIDs":[],"Sum":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2}`)
	if b, err := acc.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := nacc.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(acc, nacc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(acc), utils.ToJSON(nacc))
	}
}

func TestTCCMarshal(t *testing.T) {
	tcc, _ := NewTCC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var ntcc StatTCC
	expected := []byte(`{"FilterIDs":[],"Sum":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2}`)
	if b, err := tcc.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := ntcc.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(tcc, ntcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(tcc), utils.ToJSON(ntcc))
	}
}

func TestPDDMarshal(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second,
			utils.PDD:    5 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var ntdd StatPDD
	expected := []byte(`{"FilterIDs":[],"Sum":5000000000,"Count":1,"Events":{"EVENT_1":{"Duration":5000000000,"CompressFactor":1}},"MinItems":2}`)
	if b, err := pdd.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := ntdd.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(pdd, ntdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(pdd), utils.ToJSON(ntdd))
	}
}

func TestDCCMarshal(t *testing.T) {
	ddc, _ := NewDDC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nddc StatDDC
	expected := []byte(`{"FilterIDs":[],"FieldValues":{"1002":{"EVENT_1":{}}},"Events":{"EVENT_1":{"1002":1}},"MinItems":2,"Count":1}`)
	if b, err := ddc.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := nddc.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(ddc, nddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(ddc), utils.ToJSON(nddc))
	}
}

func TestStatSumMarshal(t *testing.T) {
	statSum, _ := NewStatSum(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nstatSum StatSum
	expected := []byte(`{"FilterIDs":[],"Sum":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FieldName":"~*req.Cost"}`)
	if b, err := statSum.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := nstatSum.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(statSum, nstatSum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(statSum), utils.ToJSON(nstatSum))
	}
}

func TestStatAverageMarshal(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "~*req.Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nstatAvg StatAverage
	expected := []byte(`{"FilterIDs":[],"Sum":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FieldName":"~*req.Cost"}`)
	if b, err := statAvg.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := nstatAvg.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(statAvg, nstatAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(statAvg), utils.ToJSON(nstatAvg))
	}
}

func TestStatDistrictMarshal(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "~*req.Usage", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nStatDistinct StatDistinct
	expected := []byte(`{"FilterIDs":[],"FieldValues":{"10s":{"EVENT_1":{}}},"Events":{"EVENT_1":{"10s":1}},"MinItems":2,"FieldName":"~*req.Usage","Count":1}`)
	if b, err := statDistinct.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := nStatDistinct.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(statDistinct, nStatDistinct) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(statDistinct), utils.ToJSON(nStatDistinct))
	}
}

func TestStatMetricsNewStatMetricError(t *testing.T) {
	_, err := NewStatMetric("", 0, []string{})
	if err == nil || err.Error() != "unsupported metric type <>" {
		t.Errorf("\nExpecting <unsupported metric type>,\nRecevied  <%+v>", err)
	}

}

func TestStatMetricsGetMinItems(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
	result := asr.GetMinItems()
	if !reflect.DeepEqual(result, 2) {
		t.Errorf("\n Expecting <2>,\nRecevied  <%+v>", result)
	}

}
func TestStatMetricsStatDistinctGetCompressFactor(t *testing.T) {
	dst := &StatDistinct{
		FilterIDs:   []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]int64{
			"Event1": {
				"1": 10000000000,
			},
			"Event2": {
				"2": 20000000000,
			},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	eventsMap := map[string]int{
		"Event1": 1,
	}
	expected := map[string]int{
		"Event1": 10000000000,
		"Event2": 20000000000,
	}
	result := dst.GetCompressFactor(eventsMap)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, result)
	}

}

func TestStatMetricsStatDistinctGetMinItems(t *testing.T) {
	dst := &StatDistinct{
		FilterIDs:   []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]int64{
			"Event1": {
				"1": 10000000000,
			},
			"Event2": {
				"2": 20000000000,
			},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	result := dst.GetMinItems()
	if !reflect.DeepEqual(result, 3) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", 3, result)
	}
}

func TestStatMetricsStatDistinctGetFilterIDs(t *testing.T) {
	dst := &StatDistinct{
		FilterIDs:   []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]int64{
			"Event1": {
				"1": 10000000000,
			},
			"Event2": {
				"2": 20000000000,
			},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	result := dst.GetFilterIDs()
	if !reflect.DeepEqual(result, dst.FilterIDs) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", dst.FilterIDs, result)
	}
}

func TestStatMetricsStatDistinctRemEvent(t *testing.T) {
	dst := &StatDistinct{
		FilterIDs:   []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	expected := &StatDistinct{
		FilterIDs:   []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]int64{
			"Event1": {},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     2,
	}
	dst.RemEvent("Event1")

	if !reflect.DeepEqual(expected, dst) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, dst)
	}
}

func TestStatMetricsStatDistinctRemEvent2(t *testing.T) {
	dst := &StatDistinct{
		FilterIDs: []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	expected := &StatDistinct{
		FilterIDs: []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     2,
	}
	dst.RemEvent("Event1")
	if !reflect.DeepEqual(expected, dst) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, dst)
	}
}

func TestStatMetricsStatDistinctAddEventErr(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]any{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
	dst := &StatDistinct{
		FilterIDs: []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	err := dst.AddEvent("Event1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "invalid format for field <Test_Field_Name>" {
		t.Errorf("\nExpecting <invalid format for field <Test_Field_Name>>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatDistinctGetValue(t *testing.T) {
	dst := &StatDistinct{
		FilterIDs: []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	result := dst.GetValue(10)
	if !reflect.DeepEqual(result, 1.0) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", 1.0, result)
	}
}

func TestStatMetricsStatAverageGetMinItems(t *testing.T) {
	avg := &StatAverage{
		FilterIDs: []string{"Test_Filter_ID"},
		Sum:       10.0,
		Count:     20,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems:  10,
		FieldName: "Test_Field_Name",
		val:       nil,
	}
	result := avg.GetMinItems()
	if !reflect.DeepEqual(result, 10) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", 10, result)
	}
}

func TestStatMetricsStatAverageGetFilterIDs(t *testing.T) {
	avg := &StatAverage{
		FilterIDs: []string{"Test_Filter_ID"},
		Sum:       10.0,
		Count:     20,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems:  10,
		FieldName: "Test_Field_Name",
		val:       nil,
	}
	result := avg.GetFilterIDs()
	if !reflect.DeepEqual(result, avg.FilterIDs) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", avg.FilterIDs, result)
	}
}

func TestStatMetricsStatAverageGetValue(t *testing.T) {
	avg := &StatAverage{
		FilterIDs: []string{"Test_Filter_ID"},
		Sum:       10.0,
		Count:     20,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems:  10,
		FieldName: "Test_Field_Name",
		val:       nil,
	}
	result := avg.GetValue(10)
	if !reflect.DeepEqual(result, 0.5) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", 0.5, result)
	}

}

func TestStatMetricsStatSumGetMinItems(t *testing.T) {
	sum := &StatSum{
		FilterIDs: []string{"Test_Filter_ID"},
		Sum:       10.0,
		Count:     15,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems:  20,
		FieldName: "Field_Name",
		val:       nil,
	}
	result := sum.GetMinItems()
	if !reflect.DeepEqual(result, 20) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", 20, result)
	}
}

func TestStatMetricsStatSumGetFilterIDs(t *testing.T) {
	sum := &StatSum{
		FilterIDs: []string{"Test_Filter_ID"},
		Sum:       10.0,
		Count:     15,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems:  20,
		FieldName: "Field_Name",
		val:       nil,
	}
	result := sum.GetFilterIDs()
	if !reflect.DeepEqual(result, sum.FilterIDs) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", sum.FilterIDs, result)
	}
}

func TestStatMetricsStatSumGetValue(t *testing.T) {
	sum := &StatSum{
		FilterIDs: []string{"Test_Filter_ID"},
		Sum:       10.0,
		Count:     15,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems:  20,
		FieldName: "Field_Name",
		val:       nil,
	}
	result := sum.GetValue(50)
	if !reflect.DeepEqual(result, float64(-1)) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", -1, float64(-1))
	}
}

func TestStatMetricsStatDDCGetFilterIDs(t *testing.T) {
	ddc := &StatDDC{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     15,
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems: 20,
	}
	result := ddc.GetFilterIDs()
	if !reflect.DeepEqual(result, ddc.FilterIDs) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", ddc.FilterIDs, result)
	}
}

func TestStatMetricsStatDDCGetMinItems(t *testing.T) {
	ddc := &StatDDC{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     15,
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems: 20,
	}
	result := ddc.GetMinItems()
	if !reflect.DeepEqual(result, ddc.MinItems) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", ddc.MinItems, result)
	}
}

func TestStatMetricsStatDDCRemEvent(t *testing.T) {
	ddc := &StatDDC{
		FilterIDs:   []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    3,
	}
	expected := &StatDDC{
		FilterIDs:   []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]int64{
			"Event1": {},
			"Event2": {},
		},
		MinItems: 3,
		Count:    2,
	}
	ddc.RemEvent("Event1")
	if !reflect.DeepEqual(expected, ddc) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, ddc)
	}
}

func TestStatMetricsStatDDCRemEvent2(t *testing.T) {
	ddc := &StatDDC{
		FilterIDs: []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    3,
	}
	expected := &StatDDC{
		FilterIDs: []string{"Test_Filter_ID"},
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    2,
	}
	ddc.RemEvent("Event1")

	if !reflect.DeepEqual(expected, ddc) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, ddc)
	}
}

func TestStatMetricsStatACDGetFilterIDs(t *testing.T) {
	timeStruct := &DurationWithCompress{
		Duration:       time.Second,
		CompressFactor: 2,
	}
	acd := &StatACD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"Event1": timeStruct,
		},
		MinItems: 3,
		Count:    3,
	}
	result := acd.GetFilterIDs()
	if !reflect.DeepEqual(acd.FilterIDs, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", acd.FilterIDs, result)
	}
}

func TestStatMetricsStatACDGetMinItems(t *testing.T) {
	timeStruct := &DurationWithCompress{
		Duration:       time.Second,
		CompressFactor: 2,
	}
	acd := &StatACD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"Event1": timeStruct,
		},
		MinItems: 3,
		Count:    3,
	}
	result := acd.GetMinItems()
	if !reflect.DeepEqual(acd.MinItems, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", acd.MinItems, result)
	}
}

func TestStatMetricsStatTCDGetFilterIDs(t *testing.T) {
	timeStruct := &DurationWithCompress{
		Duration:       time.Second,
		CompressFactor: 2,
	}
	tcd := &StatTCD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"Event1": timeStruct,
		},
		MinItems: 3,
		Count:    3,
	}
	result := tcd.GetFilterIDs()
	if !reflect.DeepEqual(tcd.FilterIDs, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", tcd.FilterIDs, result)
	}
}

func TestStatMetricsStatTCDGetMinItems(t *testing.T) {
	timeStruct := &DurationWithCompress{
		Duration:       time.Second,
		CompressFactor: 2,
	}
	tcd := &StatTCD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"Event1": timeStruct,
		},
		MinItems: 3,
		Count:    3,
	}
	result := tcd.GetMinItems()
	if !reflect.DeepEqual(tcd.MinItems, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", tcd.MinItems, result)
	}
}

func TestStatMetricsStatACCGetFloat64Value(t *testing.T) {
	acc := &StatACC{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 3,
		Count:    3,
	}
	result := acc.GetFloat64Value(2)
	if !reflect.DeepEqual(0.0, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", 0.0, result)
	}
}

func TestStatMetricsStatACCGetFilterIDs(t *testing.T) {
	acc := &StatACC{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 3,
		Count:    3,
	}
	result := acc.GetFilterIDs()
	if !reflect.DeepEqual(acc.FilterIDs, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", acc.FilterIDs, result)
	}
}

func TestStatMetricsStatACCGetMinItems(t *testing.T) {
	acc := &StatACC{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 3,
		Count:    3,
	}
	result := acc.GetMinItems()
	if !reflect.DeepEqual(acc.MinItems, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", acc.MinItems, result)
	}
}

func TestStatMetricsStatTCCGetMinItems(t *testing.T) {
	tcc := &StatTCC{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 3,
		Count:    3,
	}
	result := tcc.GetMinItems()
	if !reflect.DeepEqual(tcc.MinItems, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", tcc.MinItems, result)
	}
}

func TestStatMetricsStatTCCGetFilterIDs(t *testing.T) {
	tcc := &StatTCC{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 3,
		Count:    3,
	}
	result := tcc.GetFilterIDs()
	if !reflect.DeepEqual(tcc.FilterIDs, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", tcc.FilterIDs, result)
	}
}

func TestStatMetricsStatTCCGetFloat64Value(t *testing.T) {
	tcc := &StatTCC{
		Sum:       2.0,
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     3,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 3,
		val:      nil,
	}
	result := tcc.GetFloat64Value(2.0)
	if !reflect.DeepEqual(2.0, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", 2.0, result)
	}
}

func TestStatMetricsStatPDDGetMinItems(t *testing.T) {
	pdd := &StatPDD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems: 3,
		Count:    3,
	}
	result := pdd.GetMinItems()
	if !reflect.DeepEqual(pdd.MinItems, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", pdd.MinItems, result)
	}
}

func TestStatMetricsStatPDDGetFilterIDs(t *testing.T) {
	pdd := &StatPDD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems: 3,
		Count:    3,
	}
	result := pdd.GetFilterIDs()
	if !reflect.DeepEqual(pdd.FilterIDs, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", pdd.FilterIDs, result)
	}
}

func TestStatMetricsStatDDCGetValue(t *testing.T) {
	ddc := &StatDDC{
		FieldValues: map[string]utils.StringSet{
			"Field_Value1": {},
		},
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]map[string]int64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    3,
	}
	result := ddc.GetValue(10000)
	if !reflect.DeepEqual(1.0, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", 1.0, result)
	}
}

func TestStatMetricsStatASRAddEventErr1(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			utils.AnswerTime: "10",
		},
	}
	asr := &StatASR{
		FilterIDs: []string{"Test_Filter_ID"},
		Answered:  1.0,
		Count:     2,
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 1, CompressFactor: 1},
			"EVENT_2": {Stat: 0, CompressFactor: 1},
		},
		MinItems: 2,
		val:      nil,
	}
	err := asr.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatASRAddEventErr2(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			"AnswerTime": false,
		},
	}
	asr := &StatASR{
		FilterIDs: []string{"Test_Filter_ID"},
		Answered:  1.0,
		Count:     2,
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 1, CompressFactor: 1},
			"EVENT_2": {Stat: 0, CompressFactor: 1},
		},
		MinItems: 2,
		val:      nil,
	}
	err := asr.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "cannot convert field: false to time.Time" {
		t.Errorf("\nExpecting <cannot convert field: false to time.Time>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatACDAddEventErr(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			"Usage": false,
		},
	}
	acd := &StatACD{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     2,
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems: 2,
		val:      nil,
	}
	err := acd.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "cannot convert field: false to time.Duration" {
		t.Errorf("\nExpecting <cannot convert field: false to time.Duration>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatACDGetCompressFactor(t *testing.T) {
	eventMap := map[string]int{
		"Event1": 1000000,
	}
	timeStruct := &DurationWithCompress{
		Duration:       time.Second,
		CompressFactor: 200000000,
	}
	acd := &StatACD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"Event1": timeStruct,
		},
		MinItems: 3,
		Count:    3,
	}
	expected := map[string]int{
		"Event1": 200000000,
	}
	result := acd.GetCompressFactor(eventMap)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, result)
	}
}

func TestStatMetricsStatTCDAddEventErr(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			"Usage": false,
		},
	}
	tcd := &StatTCD{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     2,
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems: 2,
		val:      nil,
	}
	err := tcd.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "cannot convert field: false to time.Duration" {
		t.Errorf("\nExpecting <cannot convert field: false to time.Duration>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatTCDGetCompressFactor(t *testing.T) {
	eventMap := map[string]int{
		"Event1": 1000000,
	}
	timeStruct := &DurationWithCompress{
		Duration:       time.Second,
		CompressFactor: 200000000,
	}
	tcd := &StatTCD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"Event1": timeStruct,
		},
		MinItems: 3,
		Count:    3,
	}
	expected := map[string]int{
		"Event1": 200000000,
	}
	result := tcd.GetCompressFactor(eventMap)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, result)
	}
}

func TestStatMetricsStatACCAddEventErr(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			"Cost": false,
		},
	}
	acc := &StatACC{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     2,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 2,
		val:      nil,
	}
	err := acc.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "cannot convert field: false to float64" {
		t.Errorf("\nExpecting <cannot convert field: false to float64>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatTCCAddEventErr(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			"Cost": false,
		},
	}
	tcc := &StatTCC{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     2,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 2,
		val:      nil,
	}
	err := tcc.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "cannot convert field: false to float64" {
		t.Errorf("\nExpecting <cannot convert field: false to float64>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatPDDAddEventErr(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			"PDD": false,
		},
	}
	pdd := &StatPDD{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     2,
		Events: map[string]*DurationWithCompress{
			"EVENT_1": {Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": {Duration: time.Minute, CompressFactor: 1},
		},
		MinItems: 2,
		val:      nil,
	}
	err := pdd.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "cannot convert field: false to time.Duration" {
		t.Errorf("\nExpecting <cannot convert field: false to time.Duration>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatPDDGetCompressFactor(t *testing.T) {
	eventMap := map[string]int{
		"Event1": 1000000,
	}
	timeStruct := &DurationWithCompress{
		Duration:       time.Second,
		CompressFactor: 200000000,
	}
	pdd := &StatPDD{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]*DurationWithCompress{
			"Event1": timeStruct,
		},
		MinItems: 3,
		Count:    3,
	}
	expected := map[string]int{
		"Event1": 200000000,
	}
	result := pdd.GetCompressFactor(eventMap)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, result)
	}
}

func TestStatMetricsStatDDCGetCompressFactor(t *testing.T) {
	eventMap := map[string]int{
		"Event1": 1000000,
	}
	ddc := &StatDDC{
		FilterIDs: []string{"Test_Filter_ID"},
		Events: map[string]map[string]int64{
			"Event1": {
				"Event1": 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}
	expected := map[string]int{
		"Event1": 200000000,
	}
	result := ddc.GetCompressFactor(eventMap)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", expected, result)
	}
}

func TestStatMetricsStatSumAddEventErr(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			"Cost": false,
		},
	}
	sum := &StatSum{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     2,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 2,
		val:      nil,
	}
	err := sum.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "strconv.ParseFloat: parsing \"\": invalid syntax" {
		t.Errorf("\nExpecting <strconv.ParseFloat: parsing \"\": invalid syntax>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatAverageAddEventErr(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_1",
		Event: map[string]any{
			"Cost": false,
		},
	}
	avg := &StatAverage{
		FilterIDs: []string{"Test_Filter_ID"},
		Count:     2,
		Events: map[string]*StatWithCompress{
			"Event1": {
				Stat:           5,
				CompressFactor: 6,
			},
		},
		MinItems: 2,
		val:      nil,
	}
	err := avg.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: ev.Event})
	if err == nil || err.Error() != "strconv.ParseFloat: parsing \"\": invalid syntax" {
		t.Errorf("\nExpecting <strconv.ParseFloat: parsing \"\": invalid syntax>,\n Recevied <%+v>", err)
	}
}

type mockDP struct{}

func (mockDP) String() string {
	return ""
}

func (mockDP) FieldAsInterface(fldPath []string) (any, error) {
	return nil, utils.ErrAccountNotFound
}

func (mockDP) FieldAsString(fldPath []string) (string, error) {
	return "", nil
}
func (mockDP) RemoteHost() net.Addr {
	return nil
}

func TestStatMetricsStatASRAddEventErr3(t *testing.T) {
	asr := &StatASR{
		FilterIDs: []string{"Test_Filter_ID"},
		Answered:  1.0,
		Count:     2,
		Events: map[string]*StatWithCompress{
			"EVENT_1": {Stat: 1, CompressFactor: 1},
			"EVENT_2": {Stat: 0, CompressFactor: 1},
		},
		MinItems: 2,
		val:      nil,
	}
	err := asr.AddEvent("EVENT_1", new(mockDP))
	if err == nil || err.Error() != utils.ErrAccountNotFound.Error() {
		t.Errorf("\nExpecting <%+v>,\n Recevied <%+v>", utils.ErrAccountNotFound, err)
	}
}

func TestDurationWithCompressClone(t *testing.T) {
	original := &DurationWithCompress{
		Duration:       10 * time.Second,
		CompressFactor: 3,
	}
	cloned := original.Clone()

	if cloned == nil {
		t.Errorf("Cloned result is nil, expected non-nil")
	}
	if cloned.Duration != original.Duration {
		t.Errorf("Duration mismatch: expected %v, got %v", original.Duration, cloned.Duration)
	}
	if cloned.CompressFactor != original.CompressFactor {
		t.Errorf("CompressFactor mismatch: expected %d, got %d", original.CompressFactor, cloned.CompressFactor)
	}
	if cloned == original {
		t.Errorf("Expected a different memory address for cloned object")
	}

	var nilOriginal *DurationWithCompress
	nilCloned := nilOriginal.Clone()

	if nilCloned != nil {
		t.Errorf("Expected nil clone from nil receiver, got: %+v", nilCloned)
	}
}

func TestStatWithCompressClone(t *testing.T) {
	original := &StatWithCompress{
		Stat:           100.45,
		CompressFactor: 5,
	}
	cloned := original.Clone()

	if cloned == nil {
		t.Errorf("Expected non-nil clone, got nil")
	}
	if cloned.Stat != original.Stat {
		t.Errorf("Stat mismatch: expected %v, got %v", original.Stat, cloned.Stat)
	}
	if cloned.CompressFactor != original.CompressFactor {
		t.Errorf("CompressFactor mismatch: expected %d, got %d", original.CompressFactor, cloned.CompressFactor)
	}
	if cloned == original {
		t.Errorf("Expected different memory address for clone, got same")
	}

	var nilOriginal *StatWithCompress
	nilCloned := nilOriginal.Clone()

	if nilCloned != nil {
		t.Errorf("Expected nil clone from nil receiver, got: %+v", nilCloned)
	}
}

func TestStatASRClone(t *testing.T) {
	val := 0.876
	original := &StatASR{
		FilterIDs: []string{"*string:~*req.Tenant:cgrates.org"},
		Answered:  90.0,
		Count:     100,
		MinItems:  10,
		val:       &val,
		Events: map[string]*StatWithCompress{
			"event1": {Stat: 45.5, CompressFactor: 2},
			"event2": nil,
		},
	}

	cloned := original.Clone()
	if cloned == nil {
		t.Errorf("Expected non-nil clone, got nil")
	}

	clone, ok := cloned.(*StatASR)
	if !ok {
		t.Errorf("Expected *StatASR type, got different type")
		return
	}

	if clone.Answered != original.Answered {
		t.Errorf("Answered mismatch: expected %v, got %v", original.Answered, clone.Answered)
	}
	if clone.Count != original.Count {
		t.Errorf("Count mismatch: expected %v, got %v", original.Count, clone.Count)
	}
	if clone.MinItems != original.MinItems {
		t.Errorf("MinItems mismatch: expected %v, got %v", original.MinItems, clone.MinItems)
	}

	if len(clone.FilterIDs) != len(original.FilterIDs) {
		t.Errorf("FilterIDs length mismatch")
	}
	for i := range original.FilterIDs {
		if clone.FilterIDs[i] != original.FilterIDs[i] {
			t.Errorf("FilterIDs[%d] mismatch: expected %s, got %s", i, original.FilterIDs[i], clone.FilterIDs[i])
		}
	}
	if &clone.FilterIDs[0] == &original.FilterIDs[0] {
		t.Errorf("Expected deep copy of FilterIDs slice, got shallow copy")
	}

	if len(clone.Events) != 1 {
		t.Errorf("Events length mismatch: expected 1, got %d", len(clone.Events))
	}
	if e1 := clone.Events["event1"]; e1 == nil || e1.Stat != 45.5 || e1.CompressFactor != 2 {
		t.Errorf("Cloned Events[event1] mismatch")
	}
	if _, exists := clone.Events["event2"]; exists {
		t.Errorf("Expected 'event2' to be skipped in clone due to nil value")
	}
	if clone.Events["event1"] == original.Events["event1"] {
		t.Errorf("Expected deep copy of Events[event1], got same pointer")
	}

	if clone.val == nil || *clone.val != *original.val {
		t.Errorf("val mismatch or nil")
	}
	if clone.val == original.val {
		t.Errorf("Expected val to be deep copied, got same pointer")
	}

	var nilOriginal *StatASR
	nilCloned := nilOriginal.Clone()
	if nilCloned != nil {
		t.Errorf("Expected nil clone from nil receiver, got: %+v", nilCloned)
	}
}

func TestStatHighestClone(t *testing.T) {
	var nilStat *StatHighest
	if result := nilStat.Clone(); result != nil {
		t.Error("Expected nil for nil StatHighest, got non-nil")
	}

	cachedValue := 3600.5
	original := &StatHighest{
		FilterIDs: []string{"*string:~*req.Account:1001", "*prefix:~*req.Destination:+49"},
		FieldName: "*duration",
		MinItems:  10,
		Highest:   3600.5,
		Count:     25,
		Events: map[string]float64{
			"call001": 1800.0,
			"call002": 3600.5,
		},
		cachedVal: &cachedValue,
	}

	result := original.Clone()
	if result == nil {
		t.Fatal("Clone returned nil")
	}

	cloned := result.(*StatHighest)

	if !reflect.DeepEqual(cloned.FilterIDs, original.FilterIDs) {
		t.Errorf("FilterIDs mismatch: expected %v, got %v", original.FilterIDs, cloned.FilterIDs)
	}
	if cloned.FieldName != original.FieldName {
		t.Errorf("FieldName mismatch: expected %s, got %s", original.FieldName, cloned.FieldName)
	}
	if cloned.MinItems != original.MinItems {
		t.Errorf("MinItems mismatch: expected %d, got %d", original.MinItems, cloned.MinItems)
	}
	if cloned.Highest != original.Highest {
		t.Errorf("Highest mismatch: expected %f, got %f", original.Highest, cloned.Highest)
	}
	if cloned.Count != original.Count {
		t.Errorf("Count mismatch: expected %d, got %d", original.Count, cloned.Count)
	}
	if !reflect.DeepEqual(cloned.Events, original.Events) {
		t.Errorf("Events mismatch: expected %v, got %v", original.Events, cloned.Events)
	}
	if *cloned.cachedVal != *original.cachedVal {
		t.Errorf("cachedVal mismatch: expected %f, got %f", *original.cachedVal, *cloned.cachedVal)
	}

	cloned.FilterIDs[0] = "modified"
	cloned.Events["call001"] = 999.0
	*cloned.cachedVal = 999.0

	if original.FilterIDs[0] == "modified" {
		t.Error("Original FilterIDs was affected by clone modification")
	}
	if original.Events["call001"] == 999.0 {
		t.Error("Original Events was affected by clone modification")
	}
	if *original.cachedVal == 999.0 {
		t.Error("Original cachedVal was affected by clone modification")
	}
}

func TestStatHighestGetStringValue(t *testing.T) {
	statNA := &StatHighest{
		MinItems: 10,
		Count:    5,
	}

	result := statNA.GetStringValue(2)
	if result != utils.NotAvailable {
		t.Errorf("Expected %s, got %s", utils.NotAvailable, result)
	}

	statValid := &StatHighest{
		MinItems: 5,
		Count:    10,
		Highest:  3600.75,
	}

	result = statValid.GetStringValue(2)
	expected := "3600.75"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	statZero := &StatHighest{
		MinItems: 1,
		Count:    5,
		Highest:  0.0,
	}

	result = statZero.GetStringValue(2)
	expected = "0"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestStatHighestGetValue(t *testing.T) {
	statNA := &StatHighest{
		MinItems: 10,
		Count:    5,
	}

	result := statNA.GetValue(2)
	if result != utils.StatsNA {
		t.Errorf("Expected %v, got %v", utils.StatsNA, result)
	}

	statValid := &StatHighest{
		MinItems: 5,
		Count:    10,
		Highest:  3600.75,
	}

	result = statValid.GetValue(2)
	expected := 3600.75
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	statZero := &StatHighest{
		MinItems: 1,
		Count:    5,
		Highest:  0.0,
	}

	result = statZero.GetValue(2)
	expected = 0.0
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestStatHighestGetFloat64Value(t *testing.T) {
	statNA := &StatHighest{
		MinItems: 10,
		Count:    5,
	}

	result := statNA.GetFloat64Value(2)
	if result != utils.StatsNA {
		t.Errorf("Expected %v, got %v", utils.StatsNA, result)
	}

	statValid := &StatHighest{
		MinItems: 5,
		Count:    10,
		Highest:  3600.75,
	}

	result = statValid.GetFloat64Value(2)
	expected := 3600.75
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	statZero := &StatHighest{
		MinItems: 1,
		Count:    5,
		Highest:  0.0,
	}

	result = statZero.GetFloat64Value(2)
	expected = 0.0
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestStatHighestRemEvent(t *testing.T) {
	stat := &StatHighest{
		FieldName: "Cost",
		MinItems:  1,
		Count:     3,
		Highest:   30.0,
		Events: map[string]float64{
			"ev1": 10.0,
			"ev2": 20.0,
			"ev3": 30.0,
		},
		cachedVal: new(float64),
	}

	stat.RemEvent("ev2")
	if stat.Count != 2 {
		t.Errorf("Expected count = 2 after removing ev2, got %d", stat.Count)
	}
	if stat.Highest != 30.0 {
		t.Errorf("Expected highest to remain 30.0 after removing ev2, got %.2f", stat.Highest)
	}
	if _, exists := stat.Events["ev2"]; exists {
		t.Errorf("ev2 should have been deleted")
	}
	if stat.cachedVal != nil {
		t.Errorf("Expected cachedVal to be nil after removal")
	}

	stat.RemEvent("ev3")
	if stat.Count != 1 {
		t.Errorf("Expected count = 1 after removing ev3, got %d", stat.Count)
	}
	if stat.Highest != 10.0 {
		t.Errorf("Expected highest to update to 10.0 after removing ev3, got %.2f", stat.Highest)
	}
	if _, exists := stat.Events["ev3"]; exists {
		t.Errorf("ev3 should have been deleted")
	}

	stat.RemEvent("ev1")
	if stat.Count != 0 {
		t.Errorf("Expected count = 0 after removing ev1, got %d", stat.Count)
	}
	if stat.Highest != 0.0 {
		t.Errorf("Expected highest = 0.0 after removing all events, got %.2f", stat.Highest)
	}
	if len(stat.Events) != 0 {
		t.Errorf("Expected Events map to be empty, got length %d", len(stat.Events))
	}

	stat.RemEvent("nonexistent")
	if stat.Count != 0 {
		t.Errorf("Count should remain 0 after removing nonexistent event, got %d", stat.Count)
	}
	if stat.Highest != 0.0 {
		t.Errorf("Highest should remain 0.0 after removing nonexistent event, got %.2f", stat.Highest)
	}
}

func TestStatHighestMarshal(t *testing.T) {
	tests := []struct {
		name string
		stat *StatHighest
	}{
		{
			name: "valid data",
			stat: &StatHighest{
				FilterIDs: []string{"filter1", "filter2"},
				FieldName: "usage",
				MinItems:  5,
				Highest:   99.99,
				Count:     10,
				Events: map[string]float64{
					"event1": 50.0,
					"event2": 99.99,
				},
			},
		},
		{
			name: "empty struct",
			stat: &StatHighest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaler := &JSONMarshaler{}
			result, err := tt.stat.Marshal(marshaler)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) == 0 {
				t.Error("Expected non-empty result")
			}

			var unmarshaled StatHighest
			if err := json.Unmarshal(result, &unmarshaled); err != nil {
				t.Errorf("Result is not valid JSON: %v", err)
			}

			if tt.stat.FieldName != "" && unmarshaled.FieldName != tt.stat.FieldName {
				t.Errorf("Expected FieldName %s, got %s", tt.stat.FieldName, unmarshaled.FieldName)
			}
		})
	}
}

func TestStatHighestUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		inputJSON string
		want      *StatHighest
		wantErr   bool
	}{
		{
			name:      "valid json",
			inputJSON: `{"FilterIDs":["filter1","filter2"],"FieldName":"usage","MinItems":5,"Highest":99.99,"Count":10,"Events":{"event1":50.0,"event2":99.99}}`,
			want: &StatHighest{
				FilterIDs: []string{"filter1", "filter2"},
				FieldName: "usage",
				MinItems:  5,
				Highest:   99.99,
				Count:     10,
				Events: map[string]float64{
					"event1": 50.0,
					"event2": 99.99,
				},
			},
			wantErr: false,
		},
		{
			name:      "empty json",
			inputJSON: `{}`,
			want:      &StatHighest{},
			wantErr:   false,
		},
		{
			name:      "invalid json",
			inputJSON: `{"FilterIDs": "not an array"}`,
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &JSONMarshaler{}
			got := &StatHighest{}
			err := got.LoadMarshaled(ms, []byte(tt.inputJSON))

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if len(got.FilterIDs) != len(tt.want.FilterIDs) {
					t.Errorf("FilterIDs length mismatch, got %d want %d", len(got.FilterIDs), len(tt.want.FilterIDs))
				}
				if got.FieldName != tt.want.FieldName {
					t.Errorf("FieldName mismatch, got %s want %s", got.FieldName, tt.want.FieldName)
				}
				if got.MinItems != tt.want.MinItems {
					t.Errorf("MinItems mismatch, got %d want %d", got.MinItems, tt.want.MinItems)
				}
				if got.Highest != tt.want.Highest {
					t.Errorf("Highest mismatch, got %f want %f", got.Highest, tt.want.Highest)
				}
				if got.Count != tt.want.Count {
					t.Errorf("Count mismatch, got %d want %d", got.Count, tt.want.Count)
				}
				if len(got.Events) != len(tt.want.Events) {
					t.Errorf("Events map length mismatch, got %d want %d", len(got.Events), len(tt.want.Events))
				}
			}
		})
	}
}

func TestStatHighestGetFilterIDs(t *testing.T) {
	tests := []struct {
		name       string
		stat       *StatHighest
		wantFilter []string
	}{
		{
			name: "non-empty FilterIDs",
			stat: &StatHighest{
				FilterIDs: []string{"fltr1", "fltr2"},
			},
			wantFilter: []string{"fltr1", "fltr2"},
		},
		{
			name:       "empty FilterIDs",
			stat:       &StatHighest{},
			wantFilter: nil,
		},
		{
			name:       "nil receiver",
			stat:       nil,
			wantFilter: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			if tt.stat != nil {
				got = tt.stat.GetFilterIDs()
			} else {
				got = nil
			}

			if len(got) != len(tt.wantFilter) {
				t.Errorf("GetFilterIDs() length = %d, want %d", len(got), len(tt.wantFilter))
				return
			}
			for i := range got {
				if got[i] != tt.wantFilter[i] {
					t.Errorf("GetFilterIDs()[%d] = %s, want %s", i, got[i], tt.wantFilter[i])
				}
			}
		})
	}
}

func TestStatHighestGetMinItems(t *testing.T) {
	tests := []struct {
		name    string
		stat    *StatHighest
		wantMin int
	}{
		{
			name:    "MinItems set",
			stat:    &StatHighest{MinItems: 10},
			wantMin: 10,
		},
		{
			name:    "MinItems zero",
			stat:    &StatHighest{MinItems: 0},
			wantMin: 0,
		},
		{
			name:    "Nil receiver",
			stat:    nil,
			wantMin: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			if tt.stat != nil {
				got = tt.stat.GetMinItems()
			} else {
				got = 0
			}

			if got != tt.wantMin {
				t.Errorf("GetMinItems() = %d, want %d", got, tt.wantMin)
			}
		})
	}
}

func TestStatACCClone(t *testing.T) {
	t.Run("clone nil receiver", func(t *testing.T) {
		var nilStat *StatACC
		cloned := nilStat.Clone()
		if cloned != nil {
			t.Error("Expected nil from Clone() on nil receiver, got non-nil")
		}
	})

	t.Run("clone populated struct", func(t *testing.T) {
		originalVal := 3.1415

		original := &StatACC{
			FilterIDs: []string{"*string:~*req.Account:1001", "*prefix:~*req.Destination:+44"},
			Sum:       125.50,
			Count:     5,
			MinItems:  3,
			Events: map[string]*StatWithCompress{
				"call001": {Stat: 25.10, CompressFactor: 1},
				"call002": {Stat: 50.20, CompressFactor: 2},
			},
			val: &originalVal,
		}

		cloned := original.Clone()
		if cloned == nil {
			t.Fatal("Clone returned nil")
		}

		clonedStat, ok := cloned.(*StatACC)
		if !ok {
			t.Fatalf("Clone returned unexpected type: %T", cloned)
		}

		if clonedStat.Sum != original.Sum {
			t.Errorf("Expected Sum %.2f, got %.2f", original.Sum, clonedStat.Sum)
		}
		if clonedStat.Count != original.Count {
			t.Errorf("Expected Count %d, got %d", original.Count, clonedStat.Count)
		}
		if clonedStat.MinItems != original.MinItems {
			t.Errorf("Expected MinItems %d, got %d", original.MinItems, clonedStat.MinItems)
		}

		if !reflect.DeepEqual(clonedStat.FilterIDs, original.FilterIDs) {
			t.Errorf("Expected FilterIDs %v, got %v", original.FilterIDs, clonedStat.FilterIDs)
		}

		if !reflect.DeepEqual(clonedStat.Events, original.Events) {
			t.Errorf("Expected Events %v, got %v", original.Events, clonedStat.Events)
		}

		original.FilterIDs[0] = "MODIFIED"
		original.Events["call001"].Stat = 999.99
		*original.val = 0

		if clonedStat.FilterIDs[0] == "MODIFIED" {
			t.Error("FilterIDs not deeply copied")
		}
		if clonedStat.Events["call001"].Stat == 999.99 {
			t.Error("Events not deeply copied")
		}
		if *clonedStat.val == 0 {
			t.Error("val not deeply copied")
		}
	})
}

func TestStatPDDClone(t *testing.T) {
	t.Run("clone nil receiver", func(t *testing.T) {
		var nilStat *StatPDD
		cloned := nilStat.Clone()
		if cloned != nil {
			t.Error("Expected nil from Clone() on nil receiver, got non-nil")
		}
	})

	t.Run("clone populated struct", func(t *testing.T) {
		durationVal := 1500 * time.Millisecond

		original := &StatPDD{
			FilterIDs: []string{"*string:~*req.Account:1002", "*prefix:~*req.Destination:+33"},
			Sum:       10 * time.Second,
			Count:     4,
			MinItems:  2,
			Events: map[string]*DurationWithCompress{
				"call001": {Duration: 2 * time.Second, CompressFactor: 1},
				"call002": {Duration: 3 * time.Second, CompressFactor: 2},
			},
			val: &durationVal,
		}

		cloned := original.Clone()
		if cloned == nil {
			t.Fatal("Clone returned nil")
		}

		clonedStat, ok := cloned.(*StatPDD)
		if !ok {
			t.Fatalf("Expected *StatPDD from Clone, got %T", cloned)
		}

		if clonedStat.Sum != original.Sum {
			t.Errorf("Expected Sum %v, got %v", original.Sum, clonedStat.Sum)
		}
		if clonedStat.Count != original.Count {
			t.Errorf("Expected Count %d, got %d", original.Count, clonedStat.Count)
		}
		if clonedStat.MinItems != original.MinItems {
			t.Errorf("Expected MinItems %d, got %d", original.MinItems, clonedStat.MinItems)
		}

		if !reflect.DeepEqual(clonedStat.FilterIDs, original.FilterIDs) {
			t.Errorf("Expected FilterIDs %v, got %v", original.FilterIDs, clonedStat.FilterIDs)
		}

		if !reflect.DeepEqual(clonedStat.Events, original.Events) {
			t.Errorf("Expected Events %v, got %v", original.Events, clonedStat.Events)
		}

		original.FilterIDs[0] = "MODIFIED"
		original.Events["call001"].Duration = 99 * time.Second
		*original.val = 0

		if clonedStat.FilterIDs[0] == "MODIFIED" {
			t.Error("FilterIDs not deeply copied")
		}
		if clonedStat.Events["call001"].Duration == 99*time.Second {
			t.Error("Events not deeply copied")
		}
		if *clonedStat.val == 0 {
			t.Error("val not deeply copied")
		}
	})
}

func TestStatDDCClone(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var original *StatDDC
		if original.Clone() != nil {
			t.Error("Expected nil Clone result for nil receiver")
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		original := &StatDDC{
			FilterIDs:   []string{"*req.Account:1001"},
			MinItems:    5,
			Count:       10,
			FieldValues: map[string]utils.StringSet{"account": utils.NewStringSet([]string{"subject1", "subject2"})},
			Events: map[string]map[string]int64{
				"evt1": {"subject1": 1, "subject2": 2},
			},
		}

		cloned := original.Clone()
		clonedStat, ok := cloned.(*StatDDC)
		if !ok {
			t.Fatal("Cloned object is not of type *StatDDC")
		}

		if !reflect.DeepEqual(original, clonedStat) {
			t.Error("Cloned object is not deeply equal to original")
		}

		original.FilterIDs[0] = "MODIFIED"
		original.FieldValues["account"].Add("newSubject")
		original.Events["evt1"]["subject1"] = 999

		if clonedStat.FilterIDs[0] == "MODIFIED" {
			t.Error("FilterIDs not deeply copied")
		}

		found := false
		for _, v := range clonedStat.FieldValues["account"].AsSlice() {
			if v == "newSubject" {
				found = true
				break
			}
		}
		if found {
			t.Error("FieldValues not deeply copied")
		}

		if clonedStat.Events["evt1"]["subject1"] == 999 {
			t.Error("Events not deeply copied")
		}
	})
}

func TestStatSumClone(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var original *StatSum
		if original.Clone() != nil {
			t.Error("Expected nil Clone result for nil receiver")
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		val := 42.0
		original := &StatSum{
			FilterIDs: []string{"*req.Account:1001"},
			Sum:       123.45,
			Count:     5,
			MinItems:  3,
			FieldName: "*cost",
			val:       &val,
			Events: map[string]*StatWithCompress{
				"event1": {Stat: 55.5, CompressFactor: 2},
				"event2": {Stat: 67.8, CompressFactor: 1},
			},
		}

		cloned := original.Clone()
		clonedStat, ok := cloned.(*StatSum)
		if !ok {
			t.Fatal("Cloned object is not of type *StatSum")
		}

		if !reflect.DeepEqual(original, clonedStat) {
			t.Error("Cloned object is not deeply equal to original")
		}

		original.FilterIDs[0] = "MODIFIED"
		original.Events["event1"].Stat = 999.99
		*original.val = 0

		if clonedStat.FilterIDs[0] == "MODIFIED" {
			t.Error("FilterIDs not deeply copied")
		}
		if clonedStat.Events["event1"].Stat == 999.99 {
			t.Error("Events not deeply copied")
		}
		if *clonedStat.val == 0 {
			t.Error("Cached val not deeply copied")
		}
	})
}

func TestStatAverageClone(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var original *StatAverage
		if original.Clone() != nil {
			t.Error("Expected nil Clone result for nil receiver")
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		val := 7.89
		original := &StatAverage{
			FilterIDs: []string{"*req.Account:2001"},
			Sum:       150.0,
			Count:     6,
			MinItems:  4,
			FieldName: "*duration",
			val:       &val,
			Events: map[string]*StatWithCompress{
				"event1": {Stat: 75.0, CompressFactor: 1},
				"event2": {Stat: 75.0, CompressFactor: 1},
			},
		}

		cloned := original.Clone()
		clonedStat, ok := cloned.(*StatAverage)
		if !ok {
			t.Fatal("Cloned object is not of type *StatAverage")
		}

		if !reflect.DeepEqual(original, clonedStat) {
			t.Error("Cloned object is not deeply equal to original")
		}

		original.FilterIDs[0] = "CHANGED"
		original.Events["event1"].Stat = 999.9
		*original.val = 0.0

		if clonedStat.FilterIDs[0] == "CHANGED" {
			t.Error("FilterIDs not deeply copied")
		}
		if clonedStat.Events["event1"].Stat == 999.9 {
			t.Error("Events not deeply copied")
		}
		if *clonedStat.val == 0.0 {
			t.Error("Cached val not deeply copied")
		}
	})
}

func TestNewStatHighest(t *testing.T) {
	filterIDs := []string{"filter1", "filter2"}
	fieldName := "cost"
	minItems := 3

	statMetric, err := NewStatHighest(minItems, fieldName, filterIDs)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	statHighest, ok := statMetric.(*StatHighest)
	if !ok {
		t.Fatalf("Expected type *StatHighest, got: %T", statMetric)
	}

	if !reflect.DeepEqual(statHighest.FilterIDs, filterIDs) {
		t.Errorf("FilterIDs mismatch, got: %v, want: %v", statHighest.FilterIDs, filterIDs)
	}
	if statHighest.FieldName != fieldName {
		t.Errorf("FieldName mismatch, got: %s, want: %s", statHighest.FieldName, fieldName)
	}
	if statHighest.MinItems != minItems {
		t.Errorf("MinItems mismatch, got: %d, want: %d", statHighest.MinItems, minItems)
	}
	if len(statHighest.Events) != 0 {
		t.Errorf("Expected Events to be empty, got: %v", statHighest.Events)
	}
	if statHighest.cachedVal != nil {
		t.Errorf("Expected cachedVal to be nil, got: %v", statHighest.cachedVal)
	}
}

func TestStatHighestCompress(t *testing.T) {
	s := &StatHighest{
		Events: map[string]float64{
			"evt1": 1.45,
			"evt2": 4.85,
			"evt3": 7.81,
		},
	}

	got := s.Compress(10, "defaultID", 2)

	expected := []string{"evt1", "evt2", "evt3"}
	sort.Strings(got)
	sort.Strings(expected)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Compress() returned %v, expected %v", got, expected)
	}
}

func TestStatHighestGetCompressFactor(t *testing.T) {
	s := &StatHighest{
		FilterIDs: []string{"*req.Account:1001"},
		Events: map[string]float64{
			"id1": 10.0,
			"id2": 20.0,
		},
	}

	t.Run("empty events map", func(t *testing.T) {
		input := make(map[string]int)
		got := s.GetCompressFactor(input)
		expected := map[string]int{
			"id1": 1,
			"id2": 1,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	})

	t.Run("map with existing ID", func(t *testing.T) {
		input := map[string]int{
			"id1": 5,
		}
		got := s.GetCompressFactor(input)
		expected := map[string]int{
			"id1": 5,
			"id2": 1,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	})
}

func TestNewStatLowest(t *testing.T) {
	filterIDs := []string{"f1", "f2"}
	fieldName := "duration"
	minItems := 4

	statMetric, err := NewStatLowest(minItems, fieldName, filterIDs)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	statLowest, ok := statMetric.(*StatLowest)
	if !ok {
		t.Fatalf("Expected type *StatLowest, got: %T", statMetric)
	}

	if !reflect.DeepEqual(statLowest.FilterIDs, filterIDs) {
		t.Errorf("FilterIDs mismatch, got: %v, want: %v", statLowest.FilterIDs, filterIDs)
	}
	if statLowest.FieldName != fieldName {
		t.Errorf("FieldName mismatch, got: %s, want: %s", statLowest.FieldName, fieldName)
	}
	if statLowest.MinItems != minItems {
		t.Errorf("MinItems mismatch, got: %d, want: %d", statLowest.MinItems, minItems)
	}
	if statLowest.Lowest != math.MaxFloat64 {
		t.Errorf("Lowest mismatch, got: %v, want: %v", statLowest.Lowest, math.MaxFloat64)
	}
	if len(statLowest.Events) != 0 {
		t.Errorf("Expected Events to be empty, got: %v", statLowest.Events)
	}
}

func TestStatLowestClone(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var s *StatLowest
		if got := s.Clone(); got != nil {
			t.Errorf("Expected nil clone for nil receiver, got: %#v", got)
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		cachedVal := 5.55
		original := &StatLowest{
			FilterIDs: []string{"f1", "f2"},
			FieldName: "duration",
			MinItems:  3,
			Lowest:    1.23,
			Count:     10,
			Events: map[string]float64{
				"e1": 1.23,
				"e2": 4.56,
			},
			cachedVal: &cachedVal,
		}

		cloneMetric := original.Clone()
		clone, ok := cloneMetric.(*StatLowest)
		if !ok {
			t.Fatalf("Expected *StatLowest, got %T", cloneMetric)
		}

		if !reflect.DeepEqual(original, clone) {
			t.Errorf("Clone mismatch.\nGot: %#v\nWant: %#v", clone, original)
		}

		clone.FilterIDs[0] = "changed"
		clone.Events["e1"] = 9.99
		clone.Lowest = 99.99
		if reflect.DeepEqual(original, clone) {
			t.Errorf("Expected deep copy, but changes to clone affected original")
		}
	})
}

func TestStatLowestGetStringValue(t *testing.T) {
	t.Run("not enough items", func(t *testing.T) {
		s := &StatLowest{
			MinItems: 2,
			Count:    1,
			Lowest:   5.55,
			Events:   make(map[string]float64),
		}
		got := s.GetStringValue(2)
		if got != utils.NotAvailable {
			t.Errorf("Expected %q, got %q", utils.NotAvailable, got)
		}
	})

	t.Run("valid lowest value", func(t *testing.T) {
		s := &StatLowest{
			MinItems: 1,
			Count:    2,
			Lowest:   5.55555,
			Events:   make(map[string]float64),
		}
		got := s.GetStringValue(2)
		expected := strconv.FormatFloat(utils.Round(5.55555, 2, utils.MetaRoundingMiddle), 'f', -1, 64)
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})
}

func TestStatLowestGetValue(t *testing.T) {
	t.Run("not enough items", func(t *testing.T) {
		s := &StatLowest{
			MinItems: 3,
			Count:    2,
			Lowest:   1.11,
			Events:   make(map[string]float64),
		}

		got := s.GetValue(2)
		if got != utils.StatsNA {
			t.Errorf("Expected %v, got %v", utils.StatsNA, got)
		}
	})

	t.Run("enough items", func(t *testing.T) {
		s := &StatLowest{
			MinItems: 1,
			Count:    3,
			Lowest:   5.5555,
			Events:   make(map[string]float64),
		}

		got := s.GetValue(2)
		expected := utils.Round(5.5555, 2, utils.MetaRoundingMiddle)
		if got != expected {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	})
}

func TestStatLowestGetFloat64Value(t *testing.T) {
	t.Run("not enough items", func(t *testing.T) {
		s := &StatLowest{
			MinItems: 3,
			Count:    2,
			Lowest:   1.45,
			Events:   make(map[string]float64),
		}

		got := s.GetFloat64Value(2)
		if got != utils.StatsNA {
			t.Errorf("Expected %v, got %v", utils.StatsNA, got)
		}
	})

	t.Run("enough items", func(t *testing.T) {
		s := &StatLowest{
			MinItems: 1,
			Count:    5,
			Lowest:   7.77777,
			Events:   make(map[string]float64),
		}

		got := s.GetFloat64Value(3)
		expected := utils.Round(7.77777, 3, utils.MetaRoundingMiddle)
		if got != expected {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	})
}

func TestStatLowestRemEvent(t *testing.T) {
	t.Run("remove non-existent event", func(t *testing.T) {
		s := &StatLowest{
			Lowest: math.MaxFloat64,
			Count:  1,
			Events: map[string]float64{"a": 10},
		}
		s.RemEvent("missing")

		if len(s.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(s.Events))
		}
		if s.Count != 1 {
			t.Errorf("Expected Count=1, got %d", s.Count)
		}
	})

	t.Run("remove event not lowest", func(t *testing.T) {
		s := &StatLowest{
			Lowest: 5,
			Count:  2,
			Events: map[string]float64{"a": 5, "b": 10},
		}
		s.RemEvent("b")

		if s.Count != 1 {
			t.Errorf("Expected Count=1, got %d", s.Count)
		}
		if s.Lowest != 5 {
			t.Errorf("Expected Lowest=5, got %v", s.Lowest)
		}
	})

	t.Run("remove event that is lowest", func(t *testing.T) {
		s := &StatLowest{
			Lowest: 5,
			Count:  3,
			Events: map[string]float64{"a": 5, "b": 10, "c": 7},
		}
		s.RemEvent("a")

		if s.Count != 2 {
			t.Errorf("Expected Count=2, got %d", s.Count)
		}
		if s.Lowest != 7 {
			t.Errorf("Expected Lowest=7, got %v", s.Lowest)
		}
		if s.cachedVal != nil {
			t.Errorf("Expected cachedVal to be nil after removal, got %v", s.cachedVal)
		}
	})
}

func TestStatLowestCompress(t *testing.T) {
	s := &StatLowest{
		Events: map[string]float64{
			"evt1": 1.85,
			"evt2": 4.55,
			"evt3": 0.5,
		},
	}

	got := s.Compress(10, "default", 2)

	expected := []string{"evt1", "evt2", "evt3"}

	sort.Strings(got)
	sort.Strings(expected)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Compress() returned %v, want %v", got, expected)
	}
}

func TestStatDistinctClone(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var s *StatDistinct
		if got := s.Clone(); got != nil {
			t.Errorf("Expected nil clone for nil receiver, got: %#v", got)
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		original := &StatDistinct{
			FilterIDs: []string{"fltr1", "fltr2"},
			MinItems:  5,
			FieldName: "Account",
			Count:     10,
			FieldValues: map[string]utils.StringSet{
				"1001": utils.NewStringSet([]string{"event1", "event2"}),
				"1002": utils.NewStringSet([]string{"event3"}),
			},
			Events: map[string]map[string]int64{
				"cgrates.org:event1": {"1001": 1, "1002": 2},
				"cgrates.org:event2": {"1001": 3},
			},
		}

		cloneMetric := original.Clone()
		clone, ok := cloneMetric.(*StatDistinct)
		if !ok {
			t.Fatalf("Expected *StatDistinct, got %T", cloneMetric)
		}

		if !reflect.DeepEqual(original, clone) {
			t.Errorf("Clone mismatch.\nGot: %#v\nWant: %#v", clone, original)
		}

		clone.FilterIDs[0] = "changed"
		clone.FieldValues["1001"].Add("newEvent")
		clone.Events["cgrates.org:event1"]["1001"] = 99
		clone.FieldName = "ChangedField"
		if reflect.DeepEqual(original, clone) {
			t.Errorf("Expected deep copy, but changes to clone affected original")
		}
	})
}

func TestStatLowestGetFilterIDs(t *testing.T) {
	t.Run("with filter IDs", func(t *testing.T) {
		s := &StatLowest{
			FilterIDs: []string{"filter1", "filter2"},
		}
		got := s.GetFilterIDs()
		want := []string{"filter1", "filter2"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetFilterIDs() = %v; want %v", got, want)
		}
	})

	t.Run("no filter IDs", func(t *testing.T) {
		s := &StatLowest{}
		got := s.GetFilterIDs()
		if len(got) != 0 {
			t.Errorf("Expected empty slice, got %v", got)
		}
	})
}

func TestNewStatREPFC(t *testing.T) {
	t.Run("create with valid params", func(t *testing.T) {
		filterIDs := []string{"Tenant:cgrates.org", "*string:Account:1001"}
		minItems := 5
		errorType := "Error"

		metric, err := NewStatREPFC(minItems, errorType, filterIDs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stat, ok := metric.(*StatREPFC)
		if !ok {
			t.Fatalf("expected *StatREPFC, got %T", metric)
		}

		if !reflect.DeepEqual(stat.FilterIDs, filterIDs) {
			t.Errorf("FilterIDs mismatch: got %v, want %v", stat.FilterIDs, filterIDs)
		}
		if stat.MinItems != minItems {
			t.Errorf("MinItems mismatch: got %d, want %d", stat.MinItems, minItems)
		}
		if stat.ErrorType != errorType {
			t.Errorf("ErrorType mismatch: got %s, want %s", stat.ErrorType, errorType)
		}
		if stat.Events == nil {
			t.Error("Events map should be initialized")
		}
		if len(stat.Events) != 0 {
			t.Errorf("Events map should be empty, got %d entries", len(stat.Events))
		}
	})

	t.Run("empty params", func(t *testing.T) {
		metric, err := NewStatREPFC(0, "", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		stat := metric.(*StatREPFC)

		if stat.MinItems != 0 {
			t.Errorf("MinItems mismatch: got %d, want 0", stat.MinItems)
		}
		if stat.ErrorType != "" {
			t.Errorf("ErrorType mismatch: got %s, want empty", stat.ErrorType)
		}
		if stat.FilterIDs != nil {
			t.Errorf("FilterIDs mismatch: got %v, want nil", stat.FilterIDs)
		}
		if stat.Events == nil {
			t.Error("Events map should be initialized")
		}
	})
}

func TestStatREPFCClone(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var s *StatREPFC
		if got := s.Clone(); got != nil {
			t.Errorf("Expected nil clone for nil receiver, got %#v", got)
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		cached := 3.14
		original := &StatREPFC{
			FilterIDs: []string{"*string:Account:1001", "*destination:1002"},
			MinItems:  2,
			ErrorType: "ERROR",
			Count:     7,
			Events: map[string]struct{}{
				"event_1001": {},
				"event_1002": {},
			},
			cachedVal: &cached,
		}

		cloneMetric := original.Clone()
		clone, ok := cloneMetric.(*StatREPFC)
		if !ok {
			t.Fatalf("Expected *StatREPFC, got %T", cloneMetric)
		}

		if !reflect.DeepEqual(original, clone) {
			t.Errorf("Clone mismatch.\nGot: %#v\nWant: %#v", clone, original)
		}

		clone.FilterIDs[0] = "changed"
		clone.Events["event_1001"] = struct{}{}
		clone.Count = 99
		if reflect.DeepEqual(original, clone) {
			t.Errorf("Expected deep copy, but changes to clone affected original")
		}

		if clone.cachedVal == original.cachedVal {
			t.Error("cachedVal pointer should be different in deep copy")
		}
		if *clone.cachedVal != *original.cachedVal {
			t.Errorf("cachedVal value mismatch: got %v, want %v", *clone.cachedVal, *original.cachedVal)
		}
	})
}

func TestStatREPFCGetStringValue(t *testing.T) {
	t.Run("cached value already set", func(t *testing.T) {
		cached := 42.0
		s := &StatREPFC{cachedVal: &cached}
		got := s.GetStringValue(2)
		want := strconv.FormatFloat(cached, 'f', -1, 64)
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("count is zero", func(t *testing.T) {
		s := &StatREPFC{MinItems: 1, Count: 0}
		got := s.GetStringValue(2)
		if got != utils.NotAvailable {
			t.Errorf("expected %q, got %q", utils.NotAvailable, got)
		}
	})

	t.Run("count below minItems", func(t *testing.T) {
		s := &StatREPFC{MinItems: 5, Count: 3}
		got := s.GetStringValue(2)
		if got != utils.NotAvailable {
			t.Errorf("expected %q, got %q", utils.NotAvailable, got)
		}
	})

	t.Run("count meets minItems", func(t *testing.T) {
		s := &StatREPFC{MinItems: 3, Count: 3}
		got := s.GetStringValue(2)
		want := strconv.FormatFloat(float64(3), 'f', -1, 64)
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})
}

func TestStatREPFCGetStringValueAndGetValue(t *testing.T) {
	t.Run("Count below MinItems returns NA", func(t *testing.T) {
		s := &StatREPFC{
			MinItems: 5,
			Count:    3,
		}
		gotStr := s.GetStringValue(2)
		if gotStr != utils.NotAvailable {
			t.Errorf("GetStringValue() = %v, want %v", gotStr, utils.NotAvailable)
		}
		if gotVal := s.GetValue(2); gotVal != utils.StatsNA {
			t.Errorf("GetValue() = %v, want %v", gotVal, utils.StatsNA)
		}
	})

	t.Run("Count meets MinItems returns Count", func(t *testing.T) {
		s := &StatREPFC{
			MinItems: 3,
			Count:    3,
		}
		expected := float64(3)
		wantStr := strconv.FormatFloat(expected, 'f', -1, 64)
		if gotStr := s.GetStringValue(0); gotStr != wantStr {
			t.Errorf("GetStringValue() = %v, want %v", gotStr, wantStr)
		}
		if gotVal := s.GetValue(0); gotVal != expected {
			t.Errorf("GetValue() = %v, want %v", gotVal, expected)
		}
	})

	t.Run("Uses cachedVal if set", func(t *testing.T) {
		cached := float64(42)
		s := &StatREPFC{
			MinItems:  5,
			Count:     1,
			cachedVal: &cached,
		}
		wantStr := strconv.FormatFloat(cached, 'f', -1, 64)
		if gotStr := s.GetStringValue(0); gotStr != wantStr {
			t.Errorf("GetStringValue() = %v, want %v", gotStr, wantStr)
		}
		if gotVal := s.GetValue(0); gotVal != cached {
			t.Errorf("GetValue() = %v, want %v", gotVal, cached)
		}
	})
}

func TestStatREPFCRemEvent(t *testing.T) {
	t.Run("event does not exist", func(t *testing.T) {
		s := &StatREPFC{
			Count:  5,
			Events: map[string]struct{}{"existingID": {}},
		}

		s.cachedVal = utils.Float64Pointer(123.45)

		s.RemEvent("nonExistingID")

		if s.Count != 5 {
			t.Errorf("Expected Count to remain 5, got %d", s.Count)
		}
		if _, exists := s.Events["existingID"]; !exists {
			t.Errorf("Expected existingID to remain in Events")
		}
		if s.cachedVal == nil || *s.cachedVal != 123.45 {
			t.Errorf("Expected cachedVal to remain unchanged, got %v", s.cachedVal)
		}
	})

	t.Run("event exists", func(t *testing.T) {
		s := &StatREPFC{
			Count:  3,
			Events: map[string]struct{}{"ev1": {}, "ev2": {}},
		}

		s.cachedVal = utils.Float64Pointer(999.99)

		s.RemEvent("ev1")

		if _, exists := s.Events["ev1"]; exists {
			t.Errorf("Expected ev1 to be removed from Events")
		}
		if s.Count != 2 {
			t.Errorf("Expected Count to decrement to 2, got %d", s.Count)
		}
		if s.cachedVal != nil {
			t.Errorf("Expected cachedVal to be nil after removal, got %v", s.cachedVal)
		}
	})
}

func TestStatREPFCGetMinItems(t *testing.T) {
	s := &StatREPFC{
		MinItems: 7,
	}
	if got := s.GetMinItems(); got != 7 {
		t.Errorf("GetMinItems() = %v, want %v", got, 7)
	}
}

func TestStatREPFCGetFilterIDs(t *testing.T) {
	s := &StatREPFC{
		FilterIDs: []string{"filter1", "filter2"},
	}

	got := s.GetFilterIDs()
	if !reflect.DeepEqual(got, []string{"filter1", "filter2"}) {
		t.Errorf("GetFilterIDs() = %v, want %v", got, []string{"filter1", "filter2"})
	}
}

func TestStatREPSCRemEvent(t *testing.T) {
	s := &StatREPSC{
		Count: 2,
		Events: map[string]struct{}{
			"ev1": {},
			"ev2": {},
		},
		cachedVal: func() *float64 {
			v := 42.0
			return &v
		}(),
	}

	s.RemEvent("ev1")

	if _, exists := s.Events["ev1"]; exists {
		t.Errorf("RemEvent() did not delete the event")
	}
	if s.Count != 1 {
		t.Errorf("RemEvent() Count = %d, want %d", s.Count, 1)
	}
	if s.cachedVal != nil {
		t.Errorf("RemEvent() cachedVal not reset, got %v", *s.cachedVal)
	}

	s.Count = 1
	s.RemEvent("nonexistent")
	if s.Count != 1 {
		t.Errorf("RemEvent() changed Count when removing non-existing event")
	}
}

func TestStatREPSCCompress(t *testing.T) {
	s := &StatREPSC{
		Events: map[string]struct{}{
			"ev1": {},
			"ev2": {},
			"ev3": {},
		},
	}

	result := s.Compress(10, "default", 2)

	expected := map[string]bool{"ev1": true, "ev2": true, "ev3": true}
	if len(result) != len(expected) {
		t.Fatalf("Compress() returned %d items, want %d", len(result), len(expected))
	}
	for _, id := range result {
		if !expected[id] {
			t.Errorf("Compress() returned unexpected event ID: %s", id)
		}
	}

	s.Events = map[string]struct{}{}
	result = s.Compress(5, "default", 0)
	if len(result) != 0 {
		t.Errorf("Compress() with no events returned %d items, want 0", len(result))
	}
}

func TestStatREPSCGetCompressFactor(t *testing.T) {
	s := &StatREPSC{
		Events: map[string]struct{}{
			"ev1": {},
			"ev2": {},
			"ev3": {},
		},
	}

	initialMap := make(map[string]int)
	result := s.GetCompressFactor(initialMap)

	expected := map[string]int{
		"ev1": 1,
		"ev2": 1,
		"ev3": 1,
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d entries, got %d", len(expected), len(result))
	}
	for k, v := range expected {
		if got, exists := result[k]; !exists || got != v {
			t.Errorf("Expected %s -> %d, got %d", k, v, got)
		}
	}

	initialMap = map[string]int{"ev2": 5}
	result = s.GetCompressFactor(initialMap)

	if result["ev2"] != 5 {
		t.Errorf("Expected ev2 to remain 5, got %d", result["ev2"])
	}

	if result["ev1"] != 1 || result["ev3"] != 1 {
		t.Errorf("Expected ev1 and ev3 to be added with 1, got %+v", result)
	}
}

func TestStatREPFCGetFloat64Value(t *testing.T) {
	t.Run("Returns StatsNA when Count is below MinItems", func(t *testing.T) {
		s := &StatREPFC{
			MinItems: 5,
			Count:    3,
			Events: map[string]struct{}{
				"Evt1001": {},
				"Evt1002": {},
				"Evt1003": {},
			},
			cachedVal: nil,
		}

		got := s.GetFloat64Value(2)
		if got != utils.StatsNA {
			t.Errorf("expected %v, got %v", utils.StatsNA, got)
		}
	})

	t.Run("Returns Count when Count >= MinItems", func(t *testing.T) {
		s := &StatREPFC{
			MinItems: 2,
			Count:    4,
			Events: map[string]struct{}{
				"Evt001": {},
				"Evt002": {},
				"Evt003": {},
				"Evt004": {},
			},
			cachedVal: nil,
		}

		got := s.GetFloat64Value(2)
		if got != float64(s.Count) {
			t.Errorf("expected %v, got %v", float64(s.Count), got)
		}
	})

	t.Run("Uses cached value if available", func(t *testing.T) {
		cached := 42.5
		s := &StatREPFC{
			MinItems: 0,
			Count:    10,
			Events: map[string]struct{}{
				"EevtACC_1001": {},
				"EevtACC_1002": {},
			},
			cachedVal: &cached,
		}

		got := s.GetFloat64Value(2)
		if got != cached {
			t.Errorf("expected cached value %v, got %v", cached, got)
		}
	})
}

func TestStatREPFCCompress(t *testing.T) {
	t.Run("Returns all event IDs in Events map", func(t *testing.T) {
		s := &StatREPFC{
			Events: map[string]struct{}{
				"Evt1001": {},
				"EvtE01":  {},
				"EvtR01":  {},
			},
		}

		got := s.Compress(10, "default", 2)

		expected := []string{"Evt1001", "EvtE01", "EvtR01"}

		if len(got) != len(expected) {
			t.Fatalf("expected %d IDs, got %d", len(expected), len(got))
		}

		expectedSet := make(map[string]struct{})
		for _, id := range expected {
			expectedSet[id] = struct{}{}
		}
		for _, id := range got {
			if _, ok := expectedSet[id]; !ok {
				t.Errorf("unexpected ID in result: %v", id)
			}
		}
	})

	t.Run("Returns empty slice when no events", func(t *testing.T) {
		s := &StatREPFC{
			Events: map[string]struct{}{},
		}

		got := s.Compress(5, "default", 2)

		if len(got) != 0 {
			t.Errorf("expected empty slice, got %v", got)
		}
	})
}

func TestStatREPFCGetCompressFactor(t *testing.T) {
	t.Run("Adds missing event IDs to the map with value 1", func(t *testing.T) {
		s := &StatREPFC{
			Events: map[string]struct{}{
				"Evt1001": {},
				"EvtE01":  {},
			},
		}

		input := map[string]int{
			"ExistingID": 5,
		}

		got := s.GetCompressFactor(input)

		expected := map[string]int{
			"ExistingID": 5,
			"Evt1001":    1,
			"EvtE01":     1,
		}

		if len(got) != len(expected) {
			t.Fatalf("expected map length %d, got %d", len(expected), len(got))
		}

		for key, val := range expected {
			if gotVal, ok := got[key]; !ok || gotVal != val {
				t.Errorf("expected key %s to have value %d, got %d", key, val, gotVal)
			}
		}
	})

	t.Run("Does not modify values for existing event IDs", func(t *testing.T) {
		s := &StatREPFC{
			Events: map[string]struct{}{
				"Evt2001": {},
			},
		}

		input := map[string]int{
			"Evt2001": 99,
		}

		got := s.GetCompressFactor(input)

		if got["Evt2001"] != 99 {
			t.Errorf("expected value for Evt2001 to remain 99, got %d", got["Evt2001"])
		}
	})

	t.Run("Handles empty Events map", func(t *testing.T) {
		s := &StatREPFC{
			Events: map[string]struct{}{},
		}

		input := map[string]int{"ANY": 42}

		got := s.GetCompressFactor(input)

		if len(got) != 1 || got["ANY"] != 42 {
			t.Errorf("unexpected result for empty Events map: %v", got)
		}
	})
}

func TestStatREPSCGetMinItems(t *testing.T) {
	t.Run("Returns correct MinItems value", func(t *testing.T) {
		s := &StatREPSC{
			MinItems: 5,
		}

		if got := s.GetMinItems(); got != 5 {
			t.Errorf("expected MinItems to be 5, got %d", got)
		}
	})

	t.Run("Returns zero when MinItems is default", func(t *testing.T) {
		s := &StatREPSC{}

		if got := s.GetMinItems(); got != 0 {
			t.Errorf("expected MinItems to be 0, got %d", got)
		}
	})
}

func TestStatREPSCGetFilterIDs(t *testing.T) {
	t.Run("Returns correct FilterIDs", func(t *testing.T) {
		expected := []string{"FLTR_ACCOUNT_1001", "FLTR_STAT_1_1"}
		s := &StatREPSC{
			FilterIDs: expected,
		}

		got := s.GetFilterIDs()
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("expected FilterIDs %v, got %v", expected, got)
		}
	})

	t.Run("Returns empty slice when no FilterIDs set", func(t *testing.T) {
		s := &StatREPSC{}

		got := s.GetFilterIDs()
		if len(got) != 0 {
			t.Errorf("expected empty FilterIDs slice, got %v", got)
		}
	})
}

func TestStatREPSCGetValue(t *testing.T) {
	t.Run("Returns cached value if set", func(t *testing.T) {
		val := 42.0
		s := &StatREPSC{
			cachedVal: &val,
		}

		got := s.getValue(0)
		if got != val {
			t.Errorf("expected %v, got %v", val, got)
		}
	})

	t.Run("Returns StatsNA if Count < MinItems", func(t *testing.T) {
		s := &StatREPSC{
			MinItems: 5,
			Count:    3,
			Events:   make(map[string]struct{}),
		}

		got := s.getValue(0)
		if got != utils.StatsNA {
			t.Errorf("expected StatsNA (%v), got %v", utils.StatsNA, got)
		}
	})

	t.Run("Returns StatsNA if Count == 0", func(t *testing.T) {
		s := &StatREPSC{
			MinItems: 1,
			Count:    0,
			Events:   make(map[string]struct{}),
		}

		got := s.getValue(0)
		if got != utils.StatsNA {
			t.Errorf("expected StatsNA (%v), got %v", utils.StatsNA, got)
		}
	})

	t.Run("Returns Count when Count >= MinItems", func(t *testing.T) {
		s := &StatREPSC{
			MinItems: 2,
			Count:    5,
			Events:   make(map[string]struct{}),
		}

		got := s.getValue(0)
		if got != 5.0 {
			t.Errorf("expected 5.0, got %v", got)
		}
	})
}

func TestStatREPSCClone(t *testing.T) {
	t.Run("Nil receiver returns nil", func(t *testing.T) {
		var s *StatREPSC
		if got := s.Clone(); got != nil {
			t.Errorf("expected nil, got %#v", got)
		}
	})

	t.Run("Clone returns deep copy", func(t *testing.T) {
		cached := 42.0
		original := &StatREPSC{
			FilterIDs: []string{"f1", "f2"},
			MinItems:  3,
			Count:     10,
			Events: map[string]struct{}{
				"ev1": {},
				"ev2": {},
			},
			cachedVal: &cached,
		}

		cloned := original.Clone().(*StatREPSC)

		if !reflect.DeepEqual(original.FilterIDs, cloned.FilterIDs) {
			t.Errorf("FilterIDs not equal: got %v, want %v", cloned.FilterIDs, original.FilterIDs)
		}
		if original.MinItems != cloned.MinItems {
			t.Errorf("MinItems not equal: got %v, want %v", cloned.MinItems, original.MinItems)
		}
		if original.Count != cloned.Count {
			t.Errorf("Count not equal: got %v, want %v", cloned.Count, original.Count)
		}
		if !reflect.DeepEqual(original.Events, cloned.Events) {
			t.Errorf("Events not equal: got %v, want %v", cloned.Events, original.Events)
		}
		if original.cachedVal == cloned.cachedVal {
			t.Errorf("cachedVal pointers should be different, got same address")
		}
		if *original.cachedVal != *cloned.cachedVal {
			t.Errorf("cachedVal values not equal: got %v, want %v", *cloned.cachedVal, *original.cachedVal)
		}

		cloned.FilterIDs[0] = "changed"
		cloned.Events["newEv"] = struct{}{}
		*cloned.cachedVal = 99.0

		if reflect.DeepEqual(original.FilterIDs, cloned.FilterIDs) {
			t.Errorf("FilterIDs slice was not deep copied")
		}
		if _, exists := original.Events["newEv"]; exists {
			t.Errorf("Events map was not deep copied")
		}
		if *original.cachedVal == 99.0 {
			t.Errorf("cachedVal was not deep copied")
		}
	})
}

func TestStatsREPSCGetValue(t *testing.T) {
	t.Run("Returns StatsNA when count below MinItems", func(t *testing.T) {
		s := &StatREPSC{
			MinItems: 5,
			Count:    3,
			Events:   map[string]struct{}{},
		}
		got := s.GetValue(2)
		want := utils.StatsNA
		if got != want {
			t.Errorf("expected %v, got %v", want, got)
		}
	})

	t.Run("Returns count as float64 when above MinItems", func(t *testing.T) {
		s := &StatREPSC{
			MinItems: 3,
			Count:    5,
			Events:   map[string]struct{}{},
		}
		got := s.GetValue(2)
		want := float64(5)
		if got != want {
			t.Errorf("expected %v, got %v", want, got)
		}
	})

	t.Run("Uses cached value if available", func(t *testing.T) {
		val := 123.45
		s := &StatREPSC{
			cachedVal: &val,
			Events:    map[string]struct{}{},
		}
		got := s.GetValue(2)
		if got != val {
			t.Errorf("expected cached value %v, got %v", val, got)
		}
	})
}

func TestStatREPSCGetFloat64Value(t *testing.T) {
	t.Run("Returns StatsNA when count below MinItems", func(t *testing.T) {
		s := &StatREPSC{
			MinItems: 5,
			Count:    3,
			Events:   map[string]struct{}{},
		}
		got := s.GetFloat64Value(2)
		want := utils.StatsNA
		if got != want {
			t.Errorf("expected %v, got %v", want, got)
		}
	})

	t.Run("Returns count as float64 when above MinItems", func(t *testing.T) {
		s := &StatREPSC{
			MinItems: 3,
			Count:    5,
			Events:   map[string]struct{}{},
		}
		got := s.GetFloat64Value(2)
		want := float64(5)
		if got != want {
			t.Errorf("expected %v, got %v", want, got)
		}
	})

	t.Run("Uses cached value if available", func(t *testing.T) {
		val := 77.77
		s := &StatREPSC{
			cachedVal: &val,
			Events:    map[string]struct{}{},
		}
		got := s.GetFloat64Value(2)
		if got != val {
			t.Errorf("expected cached value %v, got %v", val, got)
		}
	})
}

func TestNewStatREPSC(t *testing.T) {
	filterIDs := []string{"fltr1", "fltr2"}
	minItems := 5

	metric, err := NewStatREPSC(minItems, "ignored", filterIDs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stat, ok := metric.(*StatREPSC)
	if !ok {
		t.Fatalf("expected type *StatREPSC, got %T", metric)
	}

	if !reflect.DeepEqual(stat.FilterIDs, filterIDs) {
		t.Errorf("expected FilterIDs %v, got %v", filterIDs, stat.FilterIDs)
	}

	if stat.MinItems != minItems {
		t.Errorf("expected MinItems %d, got %d", minItems, stat.MinItems)
	}

	if stat.Events == nil {
		t.Error("expected Events map to be initialized, got nil")
	} else if len(stat.Events) != 0 {
		t.Errorf("expected Events to be empty, got length %d", len(stat.Events))
	}

	if stat.Count != 0 {
		t.Errorf("expected Count 0, got %d", stat.Count)
	}

	if stat.cachedVal != nil {
		t.Errorf("expected cachedVal to be nil, got %v", stat.cachedVal)
	}
}

func TestStatREPSCGetStringValue(t *testing.T) {
	t.Run("returns NotAvailable when StatsNA", func(t *testing.T) {
		s := &StatREPSC{}
		na := utils.StatsNA
		s.cachedVal = &na
		got := s.GetStringValue(2)
		if got != utils.NotAvailable {
			t.Errorf("expected %q, got %q", utils.NotAvailable, got)
		}
	})

	t.Run("returns formatted float when value is not StatsNA", func(t *testing.T) {
		s := &StatREPSC{}
		val := 123.456
		s.cachedVal = &val

		got := s.GetStringValue(2)
		expected := strconv.FormatFloat(val, 'f', -1, 64)
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}

func TestStatLowestGetCompressFactor(t *testing.T) {
	t.Run("adds missing IDs with value 1", func(t *testing.T) {
		s := &StatLowest{
			Events: map[string]float64{
				"id1": 10.0,
				"id2": 20.0,
			},
		}
		events := map[string]int{
			"id2": 5,
		}

		got := s.GetCompressFactor(events)

		expected := map[string]int{
			"id1": 1,
			"id2": 5,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("no changes when all IDs exist", func(t *testing.T) {
		s := &StatLowest{
			Events: map[string]float64{
				"id1": 10.0,
			},
		}
		events := map[string]int{
			"id1": 2,
		}

		got := s.GetCompressFactor(events)
		expected := map[string]int{"id1": 2}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})
}
