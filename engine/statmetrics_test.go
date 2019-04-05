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
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestASRGetStringValue(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
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

func TestASRGetStringValue2(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev)
	asr.AddEvent(ev2)
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev2)
	if strVal := asr.GetStringValue(""); strVal != "33.33333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev4)
	if strVal := asr.GetStringValue(""); strVal != "25%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev4.ID)
	asr.RemEvent(ev2.ID)
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev2.ID)
	asr.AddEvent(ev)
	if strVal := asr.GetStringValue(""); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestASRGetStringValue3(t *testing.T) {
	asr := &StatASR{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatASR{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 1, CompressFactor: 1},
			"EVENT_2": &StatWithCompress{Stat: 0, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Answered:  1,
		Count:     2,
	}
	expected.GetStringValue("")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev)
	asr.AddEvent(ev2)
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.AddEvent(ev2)
	asr.AddEvent(ev4)
	v := expected.Events["EVENT_1"]
	v.Stat = 0.5
	v.CompressFactor = 2
	v = expected.Events["EVENT_2"]
	v.Stat = 0
	v.CompressFactor = 2
	expected.Count = 4
	expected.val = nil
	expected.GetStringValue("")
	if strVal := asr.GetStringValue(""); strVal != "25%" {
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
	expected.GetStringValue("")
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
}

func TestASRGetValue(t *testing.T) {
	asr, _ := NewASR(2, "", []string{})
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

func TestASRCompress(t *testing.T) {
	asr := &StatASR{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatASR{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 1, CompressFactor: 1},
			"EVENT_2": &StatWithCompress{Stat: 0, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Answered:  1,
		Count:     2,
	}
	expected.GetStringValue("")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev)
	asr.AddEvent(ev2)
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := asr.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	expected = &StatASR{
		Events: map[string]*StatWithCompress{
			"EVENT_3": &StatWithCompress{Stat: 0.5, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Answered:  1,
		Count:     2,
	}
	expected.GetStringValue("")
	expIDs = []string{"EVENT_3"}
	if rply := asr.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.AddEvent(ev2)
	asr.AddEvent(ev4)
	v := expected.Events["EVENT_3"]
	v.Stat = 0.25
	v.CompressFactor = 4
	expected.Count = 4
	expected.val = nil
	expected.GetStringValue("")
	if rply := asr.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(""); strVal != "25%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
}

func TestASRGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	asr, _ := NewASR(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev)
	asr.AddEvent(ev2)
	if CF = asr.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	asr.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = asr.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	asr.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = asr.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestACDGetStringValue(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			utils.Usage:  time.Duration(10 * time.Second),
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	if err := acd.AddEvent(ev); err != nil {
		t.Error(err)
	}
	if strVal := acd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	if err := acd.AddEvent(ev2); err == nil || err.Error() != "NOT_FOUND:Usage" {
		t.Error(err)
	}
	if err := acd.AddEvent(ev3); err == nil || err.Error() != "NOT_FOUND:Usage" {
		t.Error(err)
	}
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

func TestACDGetStringValue2(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(2 * time.Minute)}}
	if err := acd.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	acd.AddEvent(ev2)
	if strVal := acd.GetStringValue(""); strVal != "1m30s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev2)
	acd.AddEvent(ev2)
	if strVal := acd.GetStringValue(""); strVal != "1m15s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetStringValue(""); strVal != "1m20s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
}

func TestACDGetStringValue3(t *testing.T) {
	acd := &StatACD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatACD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": &DurationWithCompress{Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": &DurationWithCompress{Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(2 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(3 * time.Minute)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	if err := acd.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	if err := acd.AddEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := acd.AddEvent(ev3); err != nil {
		t.Error(err)
	}
	acd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = 3*time.Minute + 30*time.Second
	expected.val = nil
	expected.GetStringValue("")
	acd.RemEvent(ev1.ID)
	acd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
}

func TestACDCompress(t *testing.T) {
	acd := &StatACD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatACD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": &DurationWithCompress{Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": &DurationWithCompress{Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(2 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(3 * time.Minute)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	acd.AddEvent(ev1)
	acd.AddEvent(ev2)
	acd.AddEvent(ev3)
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := acd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
	expected = &StatACD{
		Events: map[string]*DurationWithCompress{
			"EVENT_3": &DurationWithCompress{Duration: 2 * time.Minute, CompressFactor: 3},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")

	expIDs = []string{"EVENT_3"}
	if rply := acd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
}

func TestACDGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	acd, _ := NewACD(2, "", []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Usage: time.Duration(2 * time.Minute)}}

	acd.AddEvent(ev)
	acd.AddEvent(ev2)
	if CF = acd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acd.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = acd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acd.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = acd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestACDGetFloat64Value(t *testing.T) {
	acd, _ := NewACD(2, "", []string{})
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
	acd, _ := NewACD(2, "", []string{})
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
	tcd, _ := NewTCD(2, "", []string{})
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

func TestTCDGetStringValue2(t *testing.T) {
	tcd, _ := NewTCD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(2 * time.Minute)}}
	if err := tcd.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	tcd.AddEvent(ev2)
	if strVal := tcd.GetStringValue(""); strVal != "3m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev2)
	tcd.AddEvent(ev2)
	if strVal := tcd.GetStringValue(""); strVal != "5m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetStringValue(""); strVal != "4m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestTCDGetStringValue3(t *testing.T) {
	tcd := &StatTCD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatTCD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": &DurationWithCompress{Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": &DurationWithCompress{Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(2 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(3 * time.Minute)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	if err := tcd.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	if err := tcd.AddEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := tcd.AddEvent(ev3); err != nil {
		t.Error(err)
	}
	tcd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = 3*time.Minute + 30*time.Second
	expected.val = nil
	expected.GetStringValue("")
	tcd.RemEvent(ev1.ID)
	tcd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
}

func TestTCDGetFloat64Value(t *testing.T) {
	tcd, _ := NewTCD(2, "", []string{})
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
	tcd, _ := NewTCD(2, "", []string{})
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

func TestTCDCompress(t *testing.T) {
	tcd := &StatTCD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatTCD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": &DurationWithCompress{Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": &DurationWithCompress{Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(2 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: time.Duration(3 * time.Minute)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	tcd.AddEvent(ev1)
	tcd.AddEvent(ev2)
	tcd.AddEvent(ev3)
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := tcd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
	expected = &StatTCD{
		Events: map[string]*DurationWithCompress{
			"EVENT_3": &DurationWithCompress{Duration: 2 * time.Minute, CompressFactor: 3},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")

	expIDs = []string{"EVENT_3"}
	if rply := tcd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
}

func TestTCDGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	tcd, _ := NewTCD(2, "", []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Usage": time.Duration(1 * time.Minute)}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Usage: time.Duration(2 * time.Minute)}}

	tcd.AddEvent(ev)
	tcd.AddEvent(ev2)
	if CF = tcd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcd.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = tcd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcd.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = tcd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestACCGetStringValue(t *testing.T) {
	acc, _ := NewACC(2, "", []string{})
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

func TestACCGetStringValue2(t *testing.T) {
	acc, _ := NewACC(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 12.3}}
	if err := acc.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.3}}
	acc.AddEvent(ev2)
	if strVal := acc.GetStringValue(""); strVal != "15.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev2)
	acc.AddEvent(ev2)
	if strVal := acc.GetStringValue(""); strVal != "16.8" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetStringValue(""); strVal != "16.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
}

func TestACCGetStringValue3(t *testing.T) {
	acc := &StatACC{Events: make(map[string]*StatWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatACC{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 12.2, CompressFactor: 2},
			"EVENT_3": &StatWithCompress{Stat: 18.3, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       42.7,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": 18.3}}
	if err := acc.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	if err := acc.AddEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := acc.AddEvent(ev3); err != nil {
		t.Error(err)
	}
	acc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = expected.Sum - 12.2
	expected.val = nil
	expected.GetStringValue("")
	acc.RemEvent(ev1.ID)
	acc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
}

func TestACCGetValue(t *testing.T) {
	acc, _ := NewACC(2, "", []string{})
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

func TestACCCompress(t *testing.T) {
	acc := &StatACC{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatACC{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 18.2, CompressFactor: 1},
			"EVENT_2": &StatWithCompress{Stat: 6.2, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue("")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.3}}
	acc.AddEvent(ev)
	acc.AddEvent(ev2)
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := acc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	expected = &StatACC{
		Events: map[string]*StatWithCompress{
			"EVENT_3": &StatWithCompress{Stat: 12.2, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue("")
	expIDs = []string{"EVENT_3"}
	if rply := acc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	acc.AddEvent(ev2)
	acc.AddEvent(ev4)
	v := expected.Events["EVENT_3"]
	v.Stat = 12.225
	v.CompressFactor = 4
	expected.Count = 4
	expected.Sum = 48.9
	expected.val = nil
	expected.GetStringValue("")
	if rply := acc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
}

func TestACCGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	acc, _ := NewACC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	acc.AddEvent(ev)
	acc.AddEvent(ev2)
	if CF = acc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acc.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = acc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acc.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = acc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestTCCGetStringValue(t *testing.T) {
	tcc, _ := NewTCC(2, "", []string{})
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

func TestTCCGetStringValue2(t *testing.T) {
	tcc, _ := NewTCC(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 12.3}}
	if err := tcc.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.3}}
	tcc.AddEvent(ev2)
	if strVal := tcc.GetStringValue(""); strVal != "30.6" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev2)
	tcc.AddEvent(ev2)
	if strVal := tcc.GetStringValue(""); strVal != "67.2" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetStringValue(""); strVal != "48.9" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
}

func TestTCCGetStringValue3(t *testing.T) {
	tcc := &StatTCC{Events: make(map[string]*StatWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatTCC{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 12.2, CompressFactor: 2},
			"EVENT_3": &StatWithCompress{Stat: 18.3, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       42.7,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": 18.3}}
	if err := tcc.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	if err := tcc.AddEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := tcc.AddEvent(ev3); err != nil {
		t.Error(err)
	}
	tcc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = expected.Sum - 12.2
	expected.val = nil
	expected.GetStringValue("")
	tcc.RemEvent(ev1.ID)
	tcc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
}

func TestTCCGetValue(t *testing.T) {
	tcc, _ := NewTCC(2, "", []string{})
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

func TestTCCCompress(t *testing.T) {
	tcc := &StatTCC{Events: make(map[string]*StatWithCompress),
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatTCC{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 18.2, CompressFactor: 1},
			"EVENT_2": &StatWithCompress{Stat: 6.2, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue("")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.3}}
	tcc.AddEvent(ev)
	tcc.AddEvent(ev2)
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := tcc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	expected = &StatTCC{
		Events: map[string]*StatWithCompress{
			"EVENT_3": &StatWithCompress{Stat: 12.2, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue("")
	expIDs = []string{"EVENT_3"}
	if rply := tcc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	tcc.AddEvent(ev2)
	tcc.AddEvent(ev4)
	v := expected.Events["EVENT_3"]
	v.Stat = 12.225
	v.CompressFactor = 4
	expected.Count = 4
	expected.Sum = 48.9
	expected.val = nil
	expected.GetStringValue("")
	if rply := tcc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
}

func TestTCCGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	tcc, _ := NewTCC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	tcc.AddEvent(ev)
	tcc.AddEvent(ev2)
	if CF = tcc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcc.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = tcc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcc.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = tcc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestPDDGetStringValue(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
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
	if strVal := pdd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev3.ID)
	if strVal := pdd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
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
	if strVal := pdd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev5)
	if strVal := pdd.GetStringValue(""); strVal != "10s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetStringValue(""); strVal != "10s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev5.ID)
	pdd.RemEvent(ev4.ID)
	pdd.RemEvent(ev5.ID)
	if strVal := pdd.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestPDDGetStringValue2(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: time.Duration(2 * time.Minute)}}
	if err := pdd.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.PDD: time.Duration(1 * time.Minute)}}
	pdd.AddEvent(ev2)
	if strVal := pdd.GetStringValue(""); strVal != "1m30s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev2)
	pdd.AddEvent(ev2)
	if strVal := pdd.GetStringValue(""); strVal != "1m15s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetStringValue(""); strVal != "1m20s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestPDDGetStringValue3(t *testing.T) {
	pdd := &StatPDD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatPDD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": &DurationWithCompress{Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": &DurationWithCompress{Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: time.Duration(2 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: time.Duration(3 * time.Minute)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{utils.PDD: time.Duration(1 * time.Minute)}}
	if err := pdd.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev3); err != nil {
		t.Error(err)
	}
	pdd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = 3*time.Minute + 30*time.Second
	expected.val = nil
	expected.GetStringValue("")
	pdd.RemEvent(ev1.ID)
	pdd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
}

func TestPDDGetFloat64Value(t *testing.T) {
	pdd, _ := NewPDD(2, "", []string{})
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
	if v := pdd.GetFloat64Value(); v != -1.0 {
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
	if strVal := pdd.GetFloat64Value(); strVal != 7.5 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.AddEvent(ev5)
	if strVal := pdd.GetFloat64Value(); strVal != 7.5 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetFloat64Value(); strVal != 7.5 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev4.ID)
	if strVal := pdd.GetFloat64Value(); strVal != -1.0 {
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
	pdd, _ := NewPDD(2, "", []string{})
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
	if err := pdd.AddEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev3); err == nil || err.Error() != "NOT_FOUND:PDD" {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != time.Duration(9*time.Second+500*time.Millisecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	if err := pdd.RemEvent(ev.ID); err != nil {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	if err := pdd.RemEvent(ev2.ID); err != nil {
		t.Error(err)
	}
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
	if err := pdd.AddEvent(ev4); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev5); err == nil || err.Error() != "NOT_FOUND:PDD" {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
	if err := pdd.RemEvent(ev5.ID); err == nil || err.Error() != "NOT_FOUND" {
		t.Error(err)
	}
	if err := pdd.RemEvent(ev4.ID); err != nil {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != time.Duration((-1)*time.Nanosecond) {
		t.Errorf("wrong pdd value: %+v", v)
	}
}

func TestPDDCompress(t *testing.T) {
	pdd := &StatPDD{Events: make(map[string]*DurationWithCompress), MinItems: 2, FilterIDs: []string{}}
	expected := &StatPDD{
		Events: map[string]*DurationWithCompress{
			"EVENT_1": &DurationWithCompress{Duration: 2*time.Minute + 30*time.Second, CompressFactor: 2},
			"EVENT_3": &DurationWithCompress{Duration: time.Minute, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: time.Duration(2 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: time.Duration(3 * time.Minute)}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{utils.PDD: time.Duration(1 * time.Minute)}}
	pdd.AddEvent(ev1)
	pdd.AddEvent(ev2)
	pdd.AddEvent(ev3)
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := pdd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	pdd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
	expected = &StatPDD{
		Events: map[string]*DurationWithCompress{
			"EVENT_3": &DurationWithCompress{Duration: 2 * time.Minute, CompressFactor: 3},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
		Sum:       6 * time.Minute,
	}
	expected.GetStringValue("")

	expIDs = []string{"EVENT_3"}
	if rply := pdd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	pdd.GetStringValue("")
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
}

func TestPDDGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	pdd, _ := NewPDD(2, "", []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: time.Duration(1 * time.Minute)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.PDD: time.Duration(1 * time.Minute)}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.PDD: time.Duration(2 * time.Minute)}}

	pdd.AddEvent(ev)
	pdd.AddEvent(ev2)
	if CF = pdd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	pdd.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = pdd.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	pdd.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = pdd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestDDCGetStringValue(t *testing.T) {
	ddc, _ := NewDDC(2, "", []string{})
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
	ddc, _ := NewDDC(2, "", []string{})
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

func TestDDCGetStringValue2(t *testing.T) {
	statDistinct, _ := NewDDC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev)
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	statDistinct.AddEvent(ev2)
	if strVal := statDistinct.GetStringValue(""); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestDDCCompress(t *testing.T) {
	ddc := &StatDDC{
		Events:      make(map[string]map[string]int64),
		FieldValues: make(map[string]map[string]struct{}),
		MinItems:    2,
		FilterIDs:   []string{},
	}
	expected := &StatDDC{
		Events: map[string]map[string]int64{
			"EVENT_1": map[string]int64{
				"1001": 2,
			},
			"EVENT_3": map[string]int64{
				"1002": 1,
			},
		},
		FieldValues: map[string]map[string]struct{}{
			"1001": map[string]struct{}{
				"EVENT_1": struct{}{},
			},
			"1002": map[string]struct{}{
				"EVENT_3": struct{}{},
			},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Count:     3,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	ddc.AddEvent(ev1)
	ddc.AddEvent(ev2)
	ddc.AddEvent(ev3)
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
	rply = ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
}

func TestDDCGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	ddc, _ := NewDDC(2, "", []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Destination: "1001"}}

	ddc.AddEvent(ev)
	ddc.AddEvent(ev2)
	if CF = ddc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = ddc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = ddc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatSumGetFloat64Value(t *testing.T) {
	statSum, _ := NewStatSum(2, "Cost", []string{})
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
	if err := statSum.AddEvent(ev2); err == nil || err.Error() != "NOT_FOUND:Cost" {
		t.Error(err)
	}
	if v := statSum.GetFloat64Value(); v != -1.0 {
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
	statSum, _ := NewStatSum(2, "Cost", []string{})
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

func TestStatSumGetStringValue2(t *testing.T) {
	statSum, _ := NewStatSum(2, "Cost", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 12.3}}
	if err := statSum.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.3}}
	statSum.AddEvent(ev2)
	if strVal := statSum.GetStringValue(""); strVal != "30.6" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.AddEvent(ev2)
	statSum.AddEvent(ev2)
	if strVal := statSum.GetStringValue(""); strVal != "67.2" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetStringValue(""); strVal != "48.9" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
}

func TestStatSumGetStringValue3(t *testing.T) {
	statSum := &StatSum{Events: make(map[string]*StatWithCompress), MinItems: 2, FilterIDs: []string{}, FieldName: "Cost"}
	expected := &StatSum{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 12.2, CompressFactor: 2},
			"EVENT_3": &StatWithCompress{Stat: 18.3, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: "Cost",
		Count:     3,
		Sum:       42.7,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": 18.3}}
	if err := statSum.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	if err := statSum.AddEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := statSum.AddEvent(ev3); err != nil {
		t.Error(err)
	}
	statSum.GetStringValue("")
	if !reflect.DeepEqual(*expected, *statSum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statSum))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = expected.Sum - 12.2
	expected.val = nil
	expected.GetStringValue("")
	statSum.RemEvent(ev1.ID)
	statSum.GetStringValue("")
	if !reflect.DeepEqual(*expected, *statSum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statSum))
	}
}

func TestStatSumCompress(t *testing.T) {
	sum := &StatSum{Events: make(map[string]*StatWithCompress), FieldName: "Cost",
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatSum{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 18.2, CompressFactor: 1},
			"EVENT_2": &StatWithCompress{Stat: 6.2, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		FieldName: "Cost",
		Count:     2,
	}
	expected.GetStringValue("")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.3}}
	sum.AddEvent(ev)
	sum.AddEvent(ev2)
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := sum.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue("")
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
	expected = &StatSum{
		Events: map[string]*StatWithCompress{
			"EVENT_3": &StatWithCompress{Stat: 12.2, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: "Cost",
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue("")
	expIDs = []string{"EVENT_3"}
	if rply := sum.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue("")
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
	sum.AddEvent(ev2)
	sum.AddEvent(ev4)
	v := expected.Events["EVENT_3"]
	v.Stat = 12.225
	v.CompressFactor = 4
	expected.Count = 4
	expected.Sum = 48.9
	expected.val = nil
	expected.GetStringValue("")
	if rply := sum.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue("")
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
}

func TestStatSumGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	sum, _ := NewStatSum(2, "Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	sum.AddEvent(ev)
	sum.AddEvent(ev2)
	if CF = sum.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	sum.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = sum.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	sum.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = sum.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatAverageGetFloat64Value(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "Cost", []string{})
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
	statAvg, _ := NewStatAverage(2, "Cost", []string{})
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

func TestStatAverageGetStringValue2(t *testing.T) {
	statAvg, _ := NewStatAverage(2, "Cost", []string{})
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 12.3}}
	if err := statAvg.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.3}}
	statAvg.AddEvent(ev2)
	if strVal := statAvg.GetStringValue(""); strVal != "15.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.AddEvent(ev2)
	statAvg.AddEvent(ev2)
	if strVal := statAvg.GetStringValue(""); strVal != "16.8" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetStringValue(""); strVal != "16.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
}

func TestStatAverageGetStringValue3(t *testing.T) {
	statAvg := &StatAverage{Events: make(map[string]*StatWithCompress), MinItems: 2, FilterIDs: []string{}, FieldName: "Cost"}
	expected := &StatAverage{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 12.2, CompressFactor: 2},
			"EVENT_3": &StatWithCompress{Stat: 18.3, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: "Cost",
		Count:     3,
		Sum:       42.7,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": 18.3}}
	if err := statAvg.AddEvent(ev1); err != nil {
		t.Error(err)
	}
	if err := statAvg.AddEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := statAvg.AddEvent(ev3); err != nil {
		t.Error(err)
	}
	statAvg.GetStringValue("")
	if !reflect.DeepEqual(*expected, *statAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statAvg))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Sum = expected.Sum - 12.2
	expected.val = nil
	expected.GetStringValue("")
	statAvg.RemEvent(ev1.ID)
	statAvg.GetStringValue("")
	if !reflect.DeepEqual(*expected, *statAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statAvg))
	}
}

func TestStatAverageCompress(t *testing.T) {
	avg := &StatAverage{Events: make(map[string]*StatWithCompress), FieldName: "Cost",
		MinItems: 2, FilterIDs: []string{}}
	expected := &StatAverage{
		Events: map[string]*StatWithCompress{
			"EVENT_1": &StatWithCompress{Stat: 18.2, CompressFactor: 1},
			"EVENT_2": &StatWithCompress{Stat: 6.2, CompressFactor: 1},
		},
		MinItems:  2,
		FilterIDs: []string{},
		Sum:       24.4,
		FieldName: "Cost",
		Count:     2,
	}
	expected.GetStringValue("")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.3}}
	avg.AddEvent(ev)
	avg.AddEvent(ev2)
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := avg.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue("")
	if !reflect.DeepEqual(*expected, *avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
	expected = &StatAverage{
		Events: map[string]*StatWithCompress{
			"EVENT_3": &StatWithCompress{Stat: 12.2, CompressFactor: 2},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: "Cost",
		Sum:       24.4,
		Count:     2,
	}
	expected.GetStringValue("")
	expIDs = []string{"EVENT_3"}
	if rply := avg.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue("")
	if !reflect.DeepEqual(*expected, *avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
	avg.AddEvent(ev2)
	avg.AddEvent(ev4)
	v := expected.Events["EVENT_3"]
	v.Stat = 12.225
	v.CompressFactor = 4
	expected.Count = 4
	expected.Sum = 48.9
	expected.val = nil
	expected.GetStringValue("")
	if rply := avg.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue("")
	if !reflect.DeepEqual(*expected, *avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
}

func TestStatAverageGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	avg, _ := NewStatAverage(2, "Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	avg.AddEvent(ev)
	avg.AddEvent(ev2)
	if CF = avg.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	avg.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = avg.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	avg.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = avg.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatDistinctGetFloat64Value(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "Usage", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Usage": time.Duration(10 * time.Second)}}
	statDistinct.AddEvent(ev)
	if v := statDistinct.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong statDistinct value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	statDistinct.AddEvent(ev2)
	if v := statDistinct.GetFloat64Value(); v != -1.0 {
		t.Errorf("wrong statDistinct value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage": time.Duration(1 * time.Minute),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage": time.Duration(1*time.Minute + 30*time.Second),
		},
	}
	if err := statDistinct.AddEvent(ev4); err != nil {
		t.Error(err)
	}
	if strVal := statDistinct.GetFloat64Value(); strVal != 2 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.AddEvent(ev5)
	if strVal := statDistinct.GetFloat64Value(); strVal != 3 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev2.ID)
	if strVal := statDistinct.GetFloat64Value(); strVal != 3 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev4.ID)
	if strVal := statDistinct.GetFloat64Value(); strVal != 2 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetFloat64Value(); strVal != -1 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev5.ID)
	if strVal := statDistinct.GetFloat64Value(); strVal != -1 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
}

func TestStatDistinctGetStringValue(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": "20"}}
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev)
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": "20"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": "40"}}
	statDistinct.AddEvent(ev2)
	statDistinct.AddEvent(ev3)
	if strVal := statDistinct.GetStringValue(""); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(""); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev2.ID)
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev3.ID)
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestStatDistinctGetStringValue2(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": "20"}}
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev)
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": "40"}}
	statDistinct.AddEvent(ev2)
	if strVal := statDistinct.GetStringValue(""); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestStatDistinctCompress(t *testing.T) {
	ddc := &StatDistinct{
		Events:      make(map[string]map[string]int64),
		FieldValues: make(map[string]map[string]struct{}),
		MinItems:    2,
		FilterIDs:   []string{},
		FieldName:   utils.Destination,
	}
	expected := &StatDistinct{
		Events: map[string]map[string]int64{
			"EVENT_1": map[string]int64{
				"1001": 2,
			},
			"EVENT_3": map[string]int64{
				"1002": 1,
			},
		},
		FieldValues: map[string]map[string]struct{}{
			"1001": map[string]struct{}{
				"EVENT_1": struct{}{},
			},
			"1002": map[string]struct{}{
				"EVENT_3": struct{}{},
			},
		},
		MinItems:  2,
		FilterIDs: []string{},
		FieldName: utils.Destination,
		Count:     3,
	}
	expected.GetStringValue("")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	ddc.AddEvent(ev1)
	ddc.AddEvent(ev2)
	ddc.AddEvent(ev3)
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
	rply = ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue("")
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
}

func TestStatDistinctGetCompressFactor(t *testing.T) {
	CF := make(map[string]int)
	expectedCF := map[string]int{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	ddc, _ := NewStatDistinct(2, utils.Destination, []string{})

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Destination: "1001"}}

	ddc.AddEvent(ev)
	ddc.AddEvent(ev2)
	if CF = ddc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev2)
	expectedCF["EVENT_2"] = 2
	if CF = ddc.GetCompressFactor(make(map[string]int)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev4)
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = ddc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

var jMarshaler JSONMarshaler

func TestASRMarshal(t *testing.T) {
	asr, _ := NewASR(2, "", []string{"*string:Account:1001"})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev)
	var nasr StatASR
	expected := []byte(`{"FilterIDs":["*string:Account:1001"],"Answered":1,"Count":1,"Events":{"EVENT_1":{"Stat":1,"CompressFactor":1}},"MinItems":2}`)
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
	acd, _ := NewACD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	acd.AddEvent(ev)
	var nacd StatACD
	expected := []byte(`{"FilterIDs":[],"Sum":10000000000,"Count":1,"Events":{"EVENT_1":{"Duration":10000000000,"CompressFactor":1}},"MinItems":2}`)
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
	tcd, _ := NewTCD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second)}}
	tcd.AddEvent(ev)
	var ntcd StatTCD
	expected := []byte(`{"FilterIDs":[],"Sum":10000000000,"Count":1,"Events":{"EVENT_1":{"Duration":10000000000,"CompressFactor":1}},"MinItems":2}`)
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
	acc, _ := NewACC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	acc.AddEvent(ev)
	var nacc StatACC
	expected := []byte(`{"FilterIDs":[],"Sum":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2}`)
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
	tcc, _ := NewTCC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	tcc.AddEvent(ev)
	var ntcc StatTCC
	expected := []byte(`{"FilterIDs":[],"Sum":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2}`)
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
	pdd, _ := NewPDD(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      time.Duration(10 * time.Second),
			utils.PDD:    time.Duration(5 * time.Second)}}
	pdd.AddEvent(ev)
	var ntdd StatPDD
	expected := []byte(`{"FilterIDs":[],"Sum":5000000000,"Count":1,"Events":{"EVENT_1":{"Duration":5000000000,"CompressFactor":1}},"MinItems":2}`)
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
	ddc, _ := NewDDC(2, "", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	ddc.AddEvent(ev)
	var nddc StatDDC
	expected := []byte(`{"FilterIDs":[],"FieldValues":{"1002":{"EVENT_1":{}}},"Events":{"EVENT_1":{"1002":1}},"MinItems":2,"Count":1}`)
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
	statSum, _ := NewStatSum(2, "Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	statSum.AddEvent(ev)
	var nstatSum StatSum
	expected := []byte(`{"FilterIDs":[],"Sum":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FieldName":"Cost"}`)
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
	statAvg, _ := NewStatAverage(2, "Cost", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	statAvg.AddEvent(ev)
	var nstatAvg StatAverage
	expected := []byte(`{"FilterIDs":[],"Sum":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FieldName":"Cost"}`)
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

func TestStatDistrictMarshal(t *testing.T) {
	statDistinct, _ := NewStatDistinct(2, "Usage", []string{})
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           time.Duration(10 * time.Second),
			utils.PDD:         time.Duration(5 * time.Second),
			utils.Destination: "1002"}}
	statDistinct.AddEvent(ev)
	var nStatDistinct StatDistinct
	expected := []byte(`{"FilterIDs":[],"FieldValues":{"10s":{"EVENT_1":{}}},"Events":{"EVENT_1":{"10s":1}},"MinItems":2,"FieldName":"Usage","Count":1}`)
	if b, err := statDistinct.Marshal(&jMarshaler); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , recived: %s", string(expected), string(b))
	} else if err := nStatDistinct.LoadMarshaled(&jMarshaler, b); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(statDistinct, nStatDistinct) {
		t.Errorf("Expected: %s , recived: %s", utils.ToJSON(statDistinct), utils.ToJSON(nStatDistinct))
	}
}
