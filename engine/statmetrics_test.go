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
	asr := NewASR(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := asr.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := asr.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(); strVal != "50.0%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := asr.GetStringValue(); strVal != "33.33333333333333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev3.ID)
	if strVal := asr.GetStringValue(); strVal != "50.0%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	asr.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	asr.RemEvent(ev.ID)
	if strVal := asr.GetStringValue(); strVal != "66.66666666666667%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev2.ID)
	if strVal := asr.GetStringValue(); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev4.ID)
	asr.RemEvent(ev5.ID)
	if strVal := asr.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestASRGetStringValue2(t *testing.T) {
	asr := NewASR(2, "")
	ev := &utils.CGREvent{ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := "EVENT_2"
	ev4 := "EVENT_4"
	if err := asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event}); err != nil {
		t.Error(err)
	}
	if err := asr.AddEvent(ev2, utils.MapStorage{utils.MetaReq: utils.MapStorage{}}); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(); strVal != "50.0%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if err := asr.AddEvent(ev2, utils.MapStorage{utils.MetaReq: utils.MapStorage{}}); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(); strVal != "33.33333333333333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if err := asr.AddEvent(ev4, utils.MapStorage{utils.MetaReq: utils.MapStorage{}}); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(); strVal != "25.00%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if err := asr.RemEvent(ev4); err != nil {
		t.Error(err)
	}
	if err := asr.RemEvent(ev2); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(); strVal != "50.0%" {
		t.Errorf("wrong asr value: %s", strVal)
		t.Error(utils.ToJSON(asr))
	}
	if err := asr.RemEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event}); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestASRGetStringValue3(t *testing.T) {
	asr := &StatASR{Metric: NewMetric(2)}
	expected := &StatASR{
		Metric: &Metric{
			Value:    utils.NewDecimal(1, 0),
			Count:    2,
			MinItems: 2,
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
			},
		},
	}
	expected.GetStringValue()
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := asr.GetStringValue(); strVal != "50.0%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_1"]
	v.Stat = utils.NewDecimalFromFloat64(0.5)
	v.CompressFactor = 2
	v = expected.Events["EVENT_2"]
	v.Stat = utils.NewDecimalFromFloat64(0)
	v.CompressFactor = 2
	expected.Count = 4
	if strVal := asr.GetStringValue(); strVal != "25.00%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.RemEvent(ev4.ID)
	asr.RemEvent(ev2.ID)
	v = expected.Events["EVENT_1"]
	v.Stat = utils.NewDecimalFromFloat64(1)
	v.CompressFactor = 1
	v = expected.Events["EVENT_2"]
	v.Stat = utils.NewDecimalFromFloat64(0)
	v.CompressFactor = 1
	expected.Count = 2
	if strVal := asr.GetStringValue(); strVal != "50.0%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
}

func TestASRGetValue(t *testing.T) {
	asr := NewASR(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := asr.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong asr value: %f", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	asr.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if v := asr.GetValue(); v.Compare(utils.NewDecimalFromFloat64(33.33333)) != 0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev3.ID)
	if v := asr.GetValue(); v.Compare(utils.NewDecimalFromFloat64(50.0)) != 0 {
		t.Errorf("wrong asr value: %f", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	asr.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	asr.RemEvent(ev.ID)
	if v := asr.GetValue(); v.Compare(utils.NewDecimalFromFloat64(66.666670)) != 0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev2.ID)
	if v := asr.GetValue(); v.Compare(utils.NewDecimalFromFloat64(100.0)) != 0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev4.ID)
	if v := asr.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev5.ID)
	if v := asr.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong asr value: %f", v)
	}
}

func TestASRCompress(t *testing.T) {
	asr := &StatASR{Metric: NewMetric(2)}
	expected := &StatASR{
		Metric: &Metric{
			Value:    utils.NewDecimal(1, 0),
			Count:    2,
			MinItems: 2,
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
			},
		},
	}
	expected.GetStringValue()
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := asr.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(); strVal != "50.0%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	expected = &StatASR{
		Metric: &Metric{
			Value:    utils.NewDecimal(1, 0),
			Count:    2,
			MinItems: 2,
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 2},
			},
		},
	}
	expected.GetStringValue()
	expIDs = []string{"EVENT_3"}
	if rply := asr.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(); strVal != "50.0%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(0.25)
	v.CompressFactor = 4
	expected.Count = 4
	if rply := asr.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(); strVal != "25.00%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
}

func TestASRGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	asr := NewASR(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = asr.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = asr.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
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

func TestACDGetStringValue(t *testing.T) {
	acd := NewACD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			utils.Usage:  10 * time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := acd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	if err := acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event}); err != nil {
		t.Error(err)
	}
	if strVal := acd.GetStringValue(); strVal != utils.NotAvailable {
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
	if strVal := acd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	if strVal := acd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev.ID)
	if strVal := acd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      478433753 * time.Nanosecond,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      30*time.Second + 982433452*time.Nanosecond,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := acd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := acd.GetStringValue(); strVal != "15.73043s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetStringValue(); strVal != "15.73043s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev5.ID)
	acd.RemEvent(ev4.ID)
	acd.RemEvent(ev5.ID)
	if strVal := acd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
}

func TestACDGetStringValue2(t *testing.T) {
	acd := NewACD(2, "")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 2 * time.Minute}}
	if err := acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Usage": time.Minute}}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acd.GetStringValue(); strVal != "1m30s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acd.GetStringValue(); strVal != "1m15s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetStringValue(); strVal != "1m20s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
}

func TestACDGetStringValue3(t *testing.T) {
	acd := &StatACD{Metric: NewMetric(2)}
	expected := &StatACD{
		Metric: &Metric{
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Usage": time.Minute}}
	if err := acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	acd.GetStringValue()
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.NewDecimal(int64(3*time.Minute+30*time.Second), 0)
	acd.RemEvent(ev1.ID)
	acd.GetStringValue()
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
}

func TestACDCompress(t *testing.T) {
	acd := &StatACD{Metric: NewMetric(2)}
	expected := &StatACD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
			},
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
		},
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Usage": time.Minute}}
	acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := acd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acd.GetStringValue()
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
	expected = &StatACD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 3},
			},
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
		},
	}
	expected.GetStringValue()

	expIDs = []string{"EVENT_3"}
	if rply := acd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acd.GetStringValue()
	if !reflect.DeepEqual(*expected, *acd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
}

func TestACDGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	acd := NewACD(2, "")

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Usage": time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Usage": time.Minute}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Usage: 2 * time.Minute}}

	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = acd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = acd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
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
	acd := NewACD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := acd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(35.0*1e9)) != 0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	// by default rounding decimal is 5
	if strVal := acd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(53.33333*1e9)) != 0 {
		t.Errorf("wrong acd value: %v", strVal)
	}

	acd.RemEvent(ev2.ID)
	if strVal := acd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(53.33333*1e9)) != 0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev4.ID)
	if strVal := acd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(50.0*1e9)) != 0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev.ID)
	if strVal := acd.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.RemEvent(ev5.ID)
	if strVal := acd.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acd value: %v", strVal)
	}
}

func TestACDGetValue(t *testing.T) {
	acd := NewACD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      8 * time.Second}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if v := acd.GetValue(); v.Compare(utils.NewDecimalFromFloat64(float64(9*time.Second))) != 0 {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev.ID)
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev2.ID)
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      4*time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if v := acd.GetValue(); v.Compare(utils.NewDecimalFromFloat64(float64(2*time.Minute+45*time.Second))) != 0 {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev5.ID)
	acd.RemEvent(ev4.ID)
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %+v", v)
	}
	acd.RemEvent(ev3.ID)
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %+v", v)
	}
}

func TestTCDGetStringValue(t *testing.T) {
	tcd := NewTCD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Usage":      10 * time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := tcd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := tcd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"Usage":      10 * time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := tcd.GetStringValue(); strVal != "20s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev.ID)
	if strVal := tcd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := tcd.GetStringValue(); strVal != "2m30s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev4.ID)
	if strVal := tcd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev5.ID)
	tcd.RemEvent(ev3.ID)
	if strVal := tcd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestTCDGetStringValue2(t *testing.T) {
	tcd := NewTCD(2, "")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 2 * time.Minute}}
	if err := tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Usage": time.Minute}}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcd.GetStringValue(); strVal != "3m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcd.GetStringValue(); strVal != "5m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetStringValue(); strVal != "4m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestTCDGetStringValue3(t *testing.T) {
	tcd := &StatTCD{Metric: NewMetric(2)}
	expected := &StatTCD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
			},
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
		},
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Usage": time.Minute}}
	if err := tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	tcd.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.NewDecimalFromFloat64(float64(3*time.Minute + 30*time.Second))
	tcd.RemEvent(ev1.ID)
	tcd.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
}

func TestTCDGetFloat64Value(t *testing.T) {
	tcd := NewTCD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := tcd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(70.0*1e9)) != 0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := tcd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(160.0*1e9)) != 0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(160.0*1e9)) != 0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev4.ID)
	if strVal := tcd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(100.0*1e9)) != 0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev.ID)
	if strVal := tcd.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.RemEvent(ev5.ID)
	if strVal := tcd.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %f", strVal)
	}
}

func TestTCDGetValue(t *testing.T) {
	tcd := NewTCD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      5 * time.Second}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if v := tcd.GetValue(); v.Compare(utils.NewDecimalFromFloat64(float64(15*time.Second))) != 0 {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev.ID)
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev2.ID)
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if v := tcd.GetValue(); v.Compare(utils.NewDecimalFromFloat64(float64(2*time.Minute+30*time.Second))) != 0 {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev5.ID)
	tcd.RemEvent(ev4.ID)
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %+v", v)
	}
	tcd.RemEvent(ev3.ID)
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %+v", v)
	}
}

func TestTCDCompress(t *testing.T) {
	tcd := &StatTCD{Metric: NewMetric(2)}
	expected := &StatTCD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
			},
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
		},
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Usage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Usage": time.Minute}}
	tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := tcd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcd.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
	expected = &StatTCD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 3},
			},
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
		},
	}
	expected.GetStringValue()

	expIDs = []string{"EVENT_3"}
	if rply := tcd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcd.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
}

func TestTCDGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	tcd := NewTCD(2, "")

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Usage": time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Usage": time.Minute}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Usage: 2 * time.Minute}}

	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = tcd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = tcd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = tcd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestACCGetStringValue(t *testing.T) {
	acc := NewACC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	if strVal := acc.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := acc.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := acc.GetStringValue(); strVal != "12.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev3.ID)
	if strVal := acc.GetStringValue(); strVal != utils.NotAvailable {
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
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	acc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	acc.RemEvent(ev.ID)
	if strVal := acc.GetStringValue(); strVal != "3.4" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetStringValue(); strVal != "3.4" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev4.ID)
	acc.RemEvent(ev5.ID)
	if strVal := acc.GetStringValue(); strVal != utils.NotAvailable {
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
	acc := NewACC(2, "")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 12.3}}
	if err := acc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.3}}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acc.GetStringValue(); strVal != "15.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := acc.GetStringValue(); strVal != "16.8" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetStringValue(); strVal != "16.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
}

func TestACCGetStringValue3(t *testing.T) {
	acc := &StatACC{Metric: NewMetric(2)}
	expected := &StatACC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(18.3), CompressFactor: 1},
			},
			MinItems: 2,
			Count:    3,
			Value:    utils.NewDecimalFromFloat64(42.7),
		},
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": 18.3}}
	if err := acc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	acc.GetStringValue()
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.SubstractDecimal(expected.Value, utils.NewDecimalFromFloat64(12.2))
	acc.RemEvent(ev1.ID)
	acc.GetStringValue()
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
}

func TestACCGetValue(t *testing.T) {
	acc := NewACC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev3.ID)
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
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
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	acc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	acc.RemEvent(ev.ID)
	if strVal := acc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(3.4)) != 0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(3.4)) != 0 {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev4.ID)
	acc.RemEvent(ev5.ID)
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acc value: %v", strVal)
	}
}

func TestACCCompress(t *testing.T) {
	acc := &StatACC{Metric: NewMetric(2)}
	expected := &StatACC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromFloat64(18.2), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimalFromFloat64(6.2), CompressFactor: 1},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromFloat64(24.4),
			Count:    2,
		},
	}
	expected.GetStringValue()
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.3}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := acc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue()
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	expected = &StatACC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromFloat64(24.4),
			Count:    2,
		},
	}
	expected.GetStringValue()
	expIDs = []string{"EVENT_3"}
	if rply := acc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue()
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(12.225)
	v.CompressFactor = 4
	expected.Count = 4
	expected.Value = utils.NewDecimalFromFloat64(48.9)
	if rply := acc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue()
	if !reflect.DeepEqual(*expected, *acc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
}

func TestACCGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	acc := NewACC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = acc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = acc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
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

func TestTCCGetStringValue(t *testing.T) {
	tcc := NewTCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       12.3}}
	if strVal := tcc.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := tcc.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       5.7}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := tcc.GetStringValue(); strVal != "18" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev3.ID)
	if strVal := tcc.GetStringValue(); strVal != utils.NotAvailable {
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
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	tcc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	tcc.RemEvent(ev.ID)
	if strVal := tcc.GetStringValue(); strVal != "6.8" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetStringValue(); strVal != "6.8" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev4.ID)
	tcc.RemEvent(ev5.ID)
	if strVal := tcc.GetStringValue(); strVal != utils.NotAvailable {
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
	tcc := NewTCC(2, "")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 12.3}}
	if err := tcc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.3}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcc.GetStringValue(); strVal != "30.6" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := tcc.GetStringValue(); strVal != "67.2" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetStringValue(); strVal != "48.9" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
}

func TestTCCGetStringValue3(t *testing.T) {
	tcc := &StatTCC{Metric: NewMetric(2)}
	expected := &StatTCC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(18.3), CompressFactor: 1},
			},
			MinItems: 2,
			Count:    3,
			Value:    utils.NewDecimalFromFloat64(42.7),
		},
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": 18.3}}
	if err := tcc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	tcc.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.SubstractDecimal(expected.Value, utils.NewDecimalFromFloat64(12.2))
	tcc.RemEvent(ev1.ID)
	tcc.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
}

func TestTCCGetValue(t *testing.T) {
	tcc := NewTCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	if strVal := tcc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := tcc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       1.2}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := tcc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(13.5)) != 0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev3.ID)
	if strVal := tcc.GetValue(); strVal != utils.DecimalNaN {
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
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	tcc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	tcc.RemEvent(ev.ID)
	if strVal := tcc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(6.8)) != 0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(6.8)) != 0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev4.ID)
	tcc.RemEvent(ev5.ID)
	if strVal := tcc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong tcc value: %v", strVal)
	}
}

func TestTCCCompress(t *testing.T) {
	tcc := &StatTCC{Metric: NewMetric(2)}
	expected := &StatTCC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromFloat64(18.2), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimalFromFloat64(6.2), CompressFactor: 1},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromFloat64(24.4),
			Count:    2,
		},
	}
	expected.GetStringValue()
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.3}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := tcc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	expected = &StatTCC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromFloat64(24.4),
			Count:    2,
		},
	}
	expected.GetStringValue()
	expIDs = []string{"EVENT_3"}
	if rply := tcc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(12.225)
	v.CompressFactor = 4
	expected.Count = 4
	expected.Value = utils.NewDecimalFromFloat64(48.9)
	if rply := tcc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue()
	if !reflect.DeepEqual(*expected, *tcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
}

func TestTCCGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	tcc := NewTCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = tcc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = tcc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
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

func TestPDDGetStringValue(t *testing.T) {
	pdd := NewPDD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			utils.Usage:  10 * time.Second,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    5 * time.Second,
		}}
	if strVal := pdd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := pdd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3"}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := pdd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev3.ID)
	if strVal := pdd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev.ID)
	if strVal := pdd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    10 * time.Second,
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			utils.PDD: 10 * time.Second,
		},
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := pdd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := pdd.GetStringValue(); strVal != "10s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetStringValue(); strVal != "10s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev5.ID)
	pdd.RemEvent(ev4.ID)
	pdd.RemEvent(ev5.ID)
	if strVal := pdd.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestPDDGetStringValue2(t *testing.T) {
	pdd := NewPDD(2, "")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: 2 * time.Minute}}
	if err := pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.PDD: time.Minute}}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := pdd.GetStringValue(); strVal != "1m30s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := pdd.GetStringValue(); strVal != "1m15s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetStringValue(); strVal != "1m20s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestPDDGetStringValue3(t *testing.T) {
	pdd := &StatPDD{Metric: NewMetric(2)}
	expected := &StatPDD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
			},
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
		},
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{utils.PDD: time.Minute}}
	if err := pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	pdd.GetStringValue()
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.NewDecimalFromFloat64(float64(3*time.Minute + 30*time.Second))
	pdd.RemEvent(ev1.ID)
	pdd.GetStringValue()
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
}

func TestPDDGetFloat64Value(t *testing.T) {
	pdd := NewPDD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second,
			utils.PDD:    5 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    10 * time.Second,
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":      time.Minute + 30*time.Second,
			"AnswerTime": time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := pdd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(7.5*1e9)) != 0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := pdd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(7.5*1e9)) != 0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(7.5*1e9)) != 0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev4.ID)
	if strVal := pdd.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev.ID)
	if strVal := pdd.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.RemEvent(ev5.ID)
	if strVal := pdd.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %v", strVal)
	}
}

func TestPDDGetValue(t *testing.T) {
	pdd := NewPDD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second,
			utils.PDD:    9 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %+v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{
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
	if v := pdd.GetValue(); v.Compare(utils.NewDecimalFromFloat64(float64(9*time.Second+500*time.Millisecond))) != 0 {
		t.Errorf("wrong pdd value: %+v", v)
	}
	if err := pdd.RemEvent(ev.ID); err != nil {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %+v", v)
	}
	if err := pdd.RemEvent(ev2.ID); err != nil {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %+v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":      time.Minute,
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:    8 * time.Second,
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
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
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %+v", v)
	}
	if err := pdd.RemEvent(ev5.ID); err == nil || err.Error() != "NOT_FOUND" {
		t.Error(err)
	}
	if err := pdd.RemEvent(ev4.ID); err != nil {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %+v", v)
	}
}

func TestPDDCompress(t *testing.T) {
	pdd := &StatPDD{Metric: NewMetric(2)}
	expected := &StatPDD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
			},
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
		},
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: 2 * time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: 3 * time.Minute}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{utils.PDD: time.Minute}}
	pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := pdd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	pdd.GetStringValue()
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
	expected = &StatPDD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 3},
			},
			Value:    utils.NewDecimal(6*int64(time.Minute), 0),
			Count:    3,
			MinItems: 2,
		},
	}
	expected.GetStringValue()

	expIDs = []string{"EVENT_3"}
	if rply := pdd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	pdd.GetStringValue()
	if !reflect.DeepEqual(*expected, *pdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
}

func TestPDDGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	pdd := NewPDD(2, "")

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.PDD: time.Minute}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.PDD: time.Minute}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.PDD: 2 * time.Minute}}

	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = pdd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = pdd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = pdd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestDDCGetStringValue(t *testing.T) {
	ddc := NewDDC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := ddc.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := ddc.GetStringValue(); strVal != utils.NotAvailable {
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
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := ddc.GetStringValue(); strVal != "2" {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev.ID)
	if strVal := ddc.GetStringValue(); strVal != "2" {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev2.ID)
	if strVal := ddc.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ddc.RemEvent(ev3.ID)
	if strVal := ddc.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}
}

func TestDDCGetFloat64Value(t *testing.T) {
	ddc := NewDDC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := ddc.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong ddc value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := ddc.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong ddc value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage":           time.Minute,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         10 * time.Second,
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage":           time.Minute + 30*time.Second,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := ddc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(2)) != 0 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := ddc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(3)) != 0 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev2.ID)
	if strVal := ddc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(3)) != 0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	ddc.RemEvent(ev4.ID)
	if strVal := ddc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(2)) != 0 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev.ID)
	if strVal := ddc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.RemEvent(ev5.ID)
	if strVal := ddc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong ddc value: %v", strVal)
	}
}

func TestDDCGetStringValue2(t *testing.T) {
	statDistinct := NewDDC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statDistinct.GetStringValue(); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestDDCCompress(t *testing.T) {
	ddc := &StatDDC{
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    2,
	}
	expected := &StatDDC{
		Events: map[string]map[string]uint64{
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
		MinItems: 2,
		Count:    3,
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	ddc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue()
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
	rply = ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue()
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
}

func TestDDCGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	ddc := NewDDC(2, "")

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Destination: "1001"}}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = ddc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = ddc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = ddc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatSumGetFloat64Value(t *testing.T) {
	statSum := NewStatSum(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := statSum.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statSum value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	if err := statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err == nil || err.Error() != "NOT_FOUND:~*req.Cost" {
		t.Error(err)
	}
	if v := statSum.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statSum value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Cost":            "20",
			"Usage":           time.Minute,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         10 * time.Second,
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Cost":            "20",
			"Usage":           time.Minute + 30*time.Second,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	statSum.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := statSum.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(40)) != 0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := statSum.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(60)) != 0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(60)) != 0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev4.ID)
	if strVal := statSum.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(40)) != 0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev.ID)
	if strVal := statSum.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.RemEvent(ev5.ID)
	if strVal := statSum.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong statSum value: %v", strVal)
	}
}

func TestStatSumGetStringValue(t *testing.T) {
	statSum := NewStatSum(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := statSum.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statSum.GetStringValue(); strVal != utils.NotAvailable {
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
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statSum.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := statSum.GetStringValue(); strVal != "60" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev.ID)
	if strVal := statSum.GetStringValue(); strVal != "40" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev3.ID)
	if strVal := statSum.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statSum value: %s", strVal)
	}
}

func TestStatSumGetStringValue2(t *testing.T) {
	statSum := NewStatSum(2, "~*req.Cost")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 12.3}}
	if err := statSum.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.3}}
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statSum.GetStringValue(); strVal != "30.6" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statSum.GetStringValue(); strVal != "67.2" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetStringValue(); strVal != "48.9" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
}

func TestStatSumGetStringValue3(t *testing.T) {
	statSum := &StatSum{Metric: NewMetric(2), FieldName: "~*req.Cost"}
	expected := &StatSum{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(18.3), CompressFactor: 1},
			},
			MinItems: 2,
			Count:    3,
			Value:    utils.NewDecimalFromFloat64(42.7),
		},
		FieldName: "~*req.Cost",
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": 18.3}}
	if err := statSum.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := statSum.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	statSum.GetStringValue()
	if !reflect.DeepEqual(*expected, *statSum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statSum))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.SubstractDecimal(expected.Value, utils.NewDecimalFromFloat64(12.2))
	statSum.RemEvent(ev1.ID)
	statSum.GetStringValue()
	if !reflect.DeepEqual(*expected, *statSum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statSum))
	}
}

func TestStatSumCompress(t *testing.T) {
	sum := &StatSum{Metric: NewMetric(2), FieldName: "~*req.Cost"}
	expected := &StatSum{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromFloat64(18.2), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimalFromFloat64(6.2), CompressFactor: 1},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromFloat64(24.4),
			Count:    2,
		},
		FieldName: "~*req.Cost",
	}
	expected.GetStringValue()
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.3}}
	sum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := sum.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue()
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
	expected = &StatSum{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromFloat64(24.4),
			Count:    2,
		},
		FieldName: "~*req.Cost",
	}
	expected.GetStringValue()
	expIDs = []string{"EVENT_3"}
	if rply := sum.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue()
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	sum.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(12.225)
	v.CompressFactor = 4
	expected.Count = 4
	expected.Value = utils.NewDecimalFromFloat64(48.9)
	if rply := sum.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue()
	if !reflect.DeepEqual(*expected, *sum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
}

func TestStatSumGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	sum := NewStatSum(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	sum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = sum.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = sum.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
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

func TestStatAverageGetFloat64Value(t *testing.T) {
	statAvg := NewStatAverage(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := statAvg.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statAvg value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := statAvg.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statAvg value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Cost":            "30",
			"Usage":           time.Minute,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.PDD:         10 * time.Second,
			utils.Destination: "1001",
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Cost":            "20",
			"Usage":           time.Minute + 30*time.Second,
			"AnswerTime":      time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1003",
		},
	}
	statAvg.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	if strVal := statAvg.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(25)) != 0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := statAvg.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(23.33333)) != 0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(23.33333)) != 0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev4.ID)
	if strVal := statAvg.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(20)) != 0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev.ID)
	if strVal := statAvg.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev5.ID)
	if strVal := statAvg.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
}

func TestStatAverageGetStringValue(t *testing.T) {
	statAvg := NewStatAverage(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Destination: "1002"}}
	if strVal := statAvg.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statAvg.GetStringValue(); strVal != utils.NotAvailable {
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
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statAvg.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := statAvg.GetStringValue(); strVal != "20" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev.ID)
	if strVal := statAvg.GetStringValue(); strVal != "20" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev3.ID)
	if strVal := statAvg.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
}

func TestStatAverageGetStringValue2(t *testing.T) {
	statAvg := NewStatAverage(2, "~*req.Cost")
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 12.3}}
	if err := statAvg.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.3}}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statAvg.GetStringValue(); strVal != "15.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statAvg.GetStringValue(); strVal != "16.8" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetStringValue(); strVal != "16.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
}

func TestStatAverageGetStringValue3(t *testing.T) {
	statAvg := &StatAverage{Metric: NewMetric(2), FieldName: "~*req.Cost"}
	expected := &StatAverage{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(18.3), CompressFactor: 1},
			},
			MinItems: 2,
			Count:    3,
			Value:    utils.NewDecimalFromFloat64(42.7),
		},
		FieldName: "~*req.Cost",
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": 18.3}}
	if err := statAvg.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event}); err != nil {
		t.Error(err)
	}
	if err := statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event}); err != nil {
		t.Error(err)
	}
	if err := statAvg.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event}); err != nil {
		t.Error(err)
	}
	statAvg.GetStringValue()
	if !reflect.DeepEqual(*expected, *statAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statAvg))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.SubstractDecimal(expected.Value, utils.NewDecimalFromFloat64(12.2))
	statAvg.RemEvent(ev1.ID)
	statAvg.GetStringValue()
	if !reflect.DeepEqual(*expected, *statAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statAvg))
	}
}

func TestStatAverageCompress(t *testing.T) {
	avg := &StatAverage{Metric: NewMetric(2), FieldName: "~*req.Cost"}
	expected := &StatAverage{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromFloat64(18.2), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimalFromFloat64(6.2), CompressFactor: 1},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromFloat64(24.4),
			Count:    2,
		},
		FieldName: "~*req.Cost",
	}
	expected.GetStringValue()
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 6.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.3}}
	avg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := avg.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue()
	if !reflect.DeepEqual(*expected, *avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
	expected = &StatAverage{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromFloat64(24.4),
			Count:    2,
		},
		FieldName: "~*req.Cost",
	}
	expected.GetStringValue()
	expIDs = []string{"EVENT_3"}
	if rply := avg.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue()
	if !reflect.DeepEqual(*expected, *avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	avg.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(12.22500000000000)
	v.CompressFactor = 4
	expected.Count = 4
	expected.Value = utils.NewDecimalFromFloat64(48.90000000000000)
	if rply := avg.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if !reflect.DeepEqual(expected, avg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
}

func TestStatAverageGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	avg := NewStatAverage(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": 18.2}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": 18.2}}
	avg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = avg.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = avg.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
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

func TestStatDistinctGetFloat64Value(t *testing.T) {
	statDistinct := NewStatDistinct(2, "~*req.Usage")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Usage": 10 * time.Second}}
	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if v := statDistinct.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statDistinct value: %v", v)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2"}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if v := statDistinct.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statDistinct value: %v", v)
	}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_4",
		Event: map[string]interface{}{
			"Usage": time.Minute,
		},
	}
	ev5 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_5",
		Event: map[string]interface{}{
			"Usage": time.Minute + 30*time.Second,
		},
	}
	if err := statDistinct.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event}); err != nil {
		t.Error(err)
	}
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(2)) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.AddEvent(ev5.ID, utils.MapStorage{utils.MetaReq: ev5.Event})
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(3)) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev2.ID)
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(3)) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev4.ID)
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(2)) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(-1)) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev5.ID)
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(-1)) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
}

func TestStatDistinctGetStringValue(t *testing.T) {
	statDistinct := NewStatDistinct(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": "20"}}
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{"Cost": "20"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{"Cost": "40"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	statDistinct.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	if strVal := statDistinct.GetStringValue(); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev2.ID)
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev3.ID)
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestStatDistinctGetStringValue2(t *testing.T) {
	statDistinct := NewStatDistinct(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": "20"}}
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{"Cost": "40"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if strVal := statDistinct.GetStringValue(); strVal != "2" {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	statDistinct.RemEvent(ev.ID)
	if strVal := statDistinct.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
}

func TestStatDistinctCompress(t *testing.T) {
	ddc := &StatDistinct{
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    2,
		FieldName:   utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
	}
	expected := &StatDistinct{
		Events: map[string]map[string]uint64{
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
		FieldName: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
		Count:     3,
	}
	expected.GetStringValue()
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_3",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	ddc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue()
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
	rply = ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue()
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
}

func TestStatDistinctGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	ddc := NewStatDistinct(2, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Destination)

	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{utils.Destination: "1002"}}
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Destination: "1001"}}
	ev4 := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_2",
		Event: map[string]interface{}{utils.Destination: "1001"}}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	if CF = ddc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
	expectedCF["EVENT_2"] = 2
	if CF = ddc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaReq: ev4.Event})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = ddc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

var jMarshaler JSONMarshaler

func TestASRMarshal(t *testing.T) {
	asr, err := NewStatMetric(utils.MetaASR, 2, []string{"*string:Account:1001"})
	if err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nasr StatASR
	expected := []byte(`{"StatMetric":{"Value":1,"Count":1,"Events":{"EVENT_1":{"Stat":1,"CompressFactor":1}},"MinItems":2},"FilterIDs":["*string:Account:1001"]}`)
	if b, err := jMarshaler.Marshal(asr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &nasr); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(asr, nasr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(asr), utils.ToJSON(nasr))
	}
}

func TestACDMarshal(t *testing.T) {
	acd := NewACD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nacd StatACD
	expected := []byte(`{"Value":10000000000,"Count":1,"Events":{"EVENT_1":{"Stat":10000000000,"CompressFactor":1}},"MinItems":2}`)
	if b, err := jMarshaler.Marshal(acd); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &nacd); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(acd, nacd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(acd), utils.ToJSON(nacd))
	}
}

func TestTCDMarshal(t *testing.T) {
	tcd := NewTCD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var ntcd StatTCD
	expected := []byte(`{"Value":10000000000,"Count":1,"Events":{"EVENT_1":{"Stat":10000000000,"CompressFactor":1}},"MinItems":2}`)
	if b, err := jMarshaler.Marshal(tcd); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &ntcd); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(tcd, ntcd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(tcd), utils.ToJSON(ntcd))
	}
}

func TestACCMarshal(t *testing.T) {
	acc := NewACC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nacc StatACC
	expected := []byte(`{"Value":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2}`)
	if b, err := jMarshaler.Marshal(acc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &nacc); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(acc, nacc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(acc), utils.ToJSON(nacc))
	}
}

func TestTCCMarshal(t *testing.T) {
	tcc := NewTCC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Cost":       "12.3"}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var ntcc StatTCC
	expected := []byte(`{"Value":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2}`)
	if b, err := jMarshaler.Marshal(tcc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &ntcc); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(tcc, ntcc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(tcc), utils.ToJSON(ntcc))
	}
}

func TestPDDMarshal(t *testing.T) {
	pdd := NewPDD(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":      10 * time.Second,
			utils.PDD:    5 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var ntdd StatPDD
	expected := []byte(`{"Value":5000000000,"Count":1,"Events":{"EVENT_1":{"Stat":5000000000,"CompressFactor":1}},"MinItems":2}`)
	if b, err := jMarshaler.Marshal(pdd); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &ntdd); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(pdd, ntdd) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(pdd), utils.ToJSON(ntdd))
	}
}

func TestDCCMarshal(t *testing.T) {
	ddc := NewDDC(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nddc StatDDC
	expected := []byte(`{"FieldValues":{"1002":{"EVENT_1":{}}},"Events":{"EVENT_1":{"1002":1}},"MinItems":2,"Count":1}`)
	if b, err := jMarshaler.Marshal(ddc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &nddc); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(ddc, nddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(ddc), utils.ToJSON(nddc))
	}
}

func TestStatSumMarshal(t *testing.T) {
	statSum := NewStatSum(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nstatSum StatSum
	expected := []byte(`{"Value":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FieldName":"~*req.Cost"}`)
	if b, err := jMarshaler.Marshal(statSum); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &nstatSum); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(statSum, nstatSum) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(statSum), utils.ToJSON(nstatSum))
	}
}

func TestStatAverageMarshal(t *testing.T) {
	statAvg := NewStatAverage(2, "~*req.Cost")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"Cost":            "20",
			"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			"Usage":           10 * time.Second,
			utils.PDD:         5 * time.Second,
			utils.Destination: "1002"}}
	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaReq: ev.Event})
	var nstatAvg StatAverage
	expected := []byte(`{"Value":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FieldName":"~*req.Cost"}`)
	if b, err := jMarshaler.Marshal(statAvg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &nstatAvg); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(statAvg, nstatAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(statAvg), utils.ToJSON(nstatAvg))
	}
}

func TestStatDistrictMarshal(t *testing.T) {
	statDistinct := NewStatDistinct(2, "~*req.Usage")
	statDistinct.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: map[string]interface{}{
		"Cost":            "20",
		"AnswerTime":      time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		"Usage":           10 * time.Second,
		utils.PDD:         5 * time.Second,
		utils.Destination: "1002"}})
	var nStatDistinct StatDistinct
	expected := []byte(`{"FieldValues":{"10s":{"EVENT_1":{}}},"Events":{"EVENT_1":{"10s":1}},"MinItems":2,"FieldName":"~*req.Usage","Count":1}`)
	if b, err := jMarshaler.Marshal(statDistinct); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, b) {
		t.Errorf("Expected: %s , received: %s", string(expected), string(b))
	} else if err := jMarshaler.Unmarshal(b, &nStatDistinct); err != nil {
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
	asr := NewASR(2, "")
	result := asr.GetMinItems()
	if result != 2 {
		t.Errorf("\n Expecting <2>,\nRecevied  <%+v>", result)
	}

}
func TestStatMetricsStatDistinctGetCompressFactor(t *testing.T) {
	dst := &StatDistinct{
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]uint64{
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
	eventsMap := map[string]uint64{
		"Event1": 1,
	}
	expected := map[string]uint64{
		"Event1": 10000000000,
		"Event2": 20000000000,
	}
	result := dst.GetCompressFactor(eventsMap)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, result)
	}

}

func TestStatMetricsStatDistinctGetMinItems(t *testing.T) {
	dst := &StatDistinct{
		FieldValues: map[string]utils.StringSet{},
		Events:      map[string]map[string]uint64{},
		MinItems:    3,
	}
	result := dst.GetMinItems()
	if result != 3 {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", 3, result)
	}
}

func TestStatMetricsStatDistinctRemEventErr2(t *testing.T) {
	dst := &StatDistinct{
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	err := dst.RemEvent("Event2")
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatMetricsStatDistinctRemEvent(t *testing.T) {
	dst := &StatDistinct{
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]uint64{
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
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]uint64{
			"Event1": {},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     2,
	}
	err := dst.RemEvent("Event1")
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Recevied <%+v>", err)
	}
	if !reflect.DeepEqual(expected, dst) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, dst)
	}
}

func TestStatMetricsStatDistinctRemEvent2(t *testing.T) {
	dst := &StatDistinct{
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]uint64{
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
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     2,
	}
	err := dst.RemEvent("Event1")
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Recevied <%+v>", err)
	}
	if !reflect.DeepEqual(expected, dst) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, dst)
	}
}

func TestStatMetricsStatDistinctAddEventErr(t *testing.T) {
	asr := NewASR(2, "")
	ev := &utils.CGREvent{Tenant: "cgrates.org", ID: "EVENT_1",
		Event: map[string]interface{}{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := asr.GetStringValue(); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
	dst := &StatDistinct{
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]uint64{
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
	if err == nil || err.Error() != "Invalid format for field <Test_Field_Name>" {
		t.Errorf("\nExpecting <Invalid format for field <Test_Field_Name>>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatDistinctGetValue(t *testing.T) {
	dst := &StatDistinct{
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems:  3,
		FieldName: "Test_Field_Name",
		Count:     3,
	}
	result := dst.GetValue()
	if result.Compare(utils.NewDecimal(1, 0)) != 0 {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", 1.0, result)
	}
}

func TestStatMetricsStatAverageGetMinItems(t *testing.T) {
	avg := &StatAverage{
		Metric:    NewMetric(10),
		FieldName: "Test_Field_Name",
	}
	result := avg.GetMinItems()
	if result != 10 {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", 10, result)
	}
}

func TestStatMetricsStatAverageGetValue(t *testing.T) {
	avg := &StatAverage{
		Metric: &Metric{
			Value:    utils.NewDecimal(10, 0),
			Count:    20,
			MinItems: 10,
			Events: map[string]*DecimalWithCompress{
				"Event1": {},
			},
		},
	}
	result := avg.GetValue()
	if result.Compare(utils.NewDecimalFromFloat64(0.5)) != 0 {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", 0.5, result)
	}

}

func TestStatMetricsStatSumGetMinItems(t *testing.T) {
	sum := &StatSum{
		Metric: NewMetric(20),
	}
	result := sum.GetMinItems()
	if result != 20 {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", 20, result)
	}
}

func TestStatMetricsStatSumGetValue(t *testing.T) {
	sum := &StatSum{
		Metric: &Metric{
			Value:    utils.NewDecimal(10, 0),
			Count:    15,
			MinItems: 20,
			Events: map[string]*DecimalWithCompress{
				"Event1": {},
			},
		},
	}
	result := sum.GetValue()
	if result != utils.DecimalNaN {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.DecimalNaN, result)
	}
}

func TestStatMetricsStatDDCGetMinItems(t *testing.T) {
	ddc := &StatDDC{
		Count: 15,
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems: 20,
	}
	result := ddc.GetMinItems()
	if !reflect.DeepEqual(result, ddc.MinItems) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", ddc.MinItems, result)
	}
}

func TestStatMetricsStatDDCRemEventErr2(t *testing.T) {
	ddc := &StatDDC{
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    3,
	}
	err := ddc.RemEvent("Event2")
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatMetricsStatDDCRemEvent(t *testing.T) {
	ddc := &StatDDC{
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    3,
	}
	expected := &StatDDC{
		FieldValues: map[string]utils.StringSet{},
		Events: map[string]map[string]uint64{
			"Event1": {},
			"Event2": {},
		},
		MinItems: 3,
		Count:    2,
	}
	err := ddc.RemEvent("Event1")
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Recevied <%+v>", err)
	}
	if !reflect.DeepEqual(expected, ddc) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, ddc)
	}
}

func TestStatMetricsStatDDCRemEvent2(t *testing.T) {
	ddc := &StatDDC{
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 2,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    3,
	}
	expected := &StatDDC{
		FieldValues: map[string]utils.StringSet{
			"FieldValue1": {},
		},
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    2,
	}
	err := ddc.RemEvent("Event1")
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Recevied <%+v>", err)
	}
	if !reflect.DeepEqual(expected, ddc) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, ddc)
	}
}

func TestStatMetricsStatACDGetMinItems(t *testing.T) {
	acd := &StatACD{Metric: NewMetric(2)}
	result := acd.GetMinItems()
	if !reflect.DeepEqual(acd.MinItems, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", acd.MinItems, result)
	}
}

func TestStatMetricsStatTCDGetMinItems(t *testing.T) {
	tcd := &StatTCD{Metric: NewMetric(2)}
	result := tcd.GetMinItems()
	if !reflect.DeepEqual(tcd.MinItems, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", tcd.MinItems, result)
	}
}

func TestStatMetricsStatACCGetFloat64Value(t *testing.T) {
	acc := &StatACC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"Event1": {},
			},
			MinItems: 3,
			Value:    utils.NewDecimal(0, 0),
			Count:    3,
		},
	}
	result := acc.GetValue()
	if result.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", 0.0, result)
	}
}

func TestStatMetricsStatACCGetMinItems(t *testing.T) {
	acc := &StatACC{Metric: NewMetric(2)}
	result := acc.GetMinItems()
	if !reflect.DeepEqual(acc.MinItems, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", acc.MinItems, result)
	}
}

func TestStatMetricsStatTCCGetMinItems(t *testing.T) {
	tcc := &StatTCC{Metric: NewMetric(2)}
	result := tcc.GetMinItems()
	if !reflect.DeepEqual(tcc.MinItems, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", tcc.MinItems, result)
	}
}

func TestStatMetricsStatTCCGetFloat64Value(t *testing.T) {
	tcc := &StatTCC{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 3,
			Events: map[string]*DecimalWithCompress{
				"Event1": {},
			},
			MinItems: 3,
		},
	}
	result := tcc.GetValue()
	if result.Compare(utils.NewDecimal(2, 0)) != 0 {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", 2.0, result)
	}
}

func TestStatMetricsStatPDDGetMinItems(t *testing.T) {
	pdd := &StatPDD{Metric: NewMetric(2)}
	result := pdd.GetMinItems()
	if !reflect.DeepEqual(pdd.MinItems, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", pdd.MinItems, result)
	}
}

func TestStatMetricsStatDDCGetValue(t *testing.T) {
	ddc := &StatDDC{
		FieldValues: map[string]utils.StringSet{
			"Field_Value1": {},
		},
		Events: map[string]map[string]uint64{
			"Event1": {
				"FieldValue1": 1,
			},
			"Event2": {},
		},
		MinItems: 3,
		Count:    3,
	}
	result := ddc.GetValue()
	if result.Compare(utils.NewDecimal(1, 0)) != 0 {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", 1.0, result)
	}
}

func TestStatMetricsStatASRAddEventErr1(t *testing.T) {
	asr := &StatASR{Metric: NewMetric(2)}
	err := asr.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: map[string]interface{}{utils.AnswerTime: "10"}})
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatASRAddEventErr2(t *testing.T) {
	asr := &StatASR{Metric: NewMetric(2)}
	err := asr.AddEvent("EVENT_1", utils.MapStorage{utils.MetaReq: utils.MapStorage{"AnswerTime": false}})
	if err == nil || err.Error() != "cannot convert field: false to time.Time" {
		t.Errorf("\nExpecting <cannot convert field: false to time.Time>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatACDAddEventErr(t *testing.T) {
	acd := NewMetric(2)
	err := acd.addEvent("EVENT_1", false)
	if err == nil || err.Error() != "cannot convert field: bool to decimal.Big" {
		t.Errorf("\nExpecting <cannot convert field: false to time.Duration>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatACDGetCompressFactor(t *testing.T) {
	acd := &StatACD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"Event1": {
					Stat:           utils.NewDecimal(int64(time.Second), 0),
					CompressFactor: 200000000,
				},
			},
			MinItems: 3,
			Count:    3,
		},
	}
	expected := map[string]uint64{
		"Event1": 200000000,
	}
	result := acd.GetCompressFactor(map[string]uint64{
		"Event1": 1000000,
	})
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, result)
	}
}

func TestStatMetricsStatTCDGetCompressFactor(t *testing.T) {
	eventMap := map[string]uint64{
		"Event1": 1000000,
	}
	tcd := &StatTCD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"Event1": {
					Stat:           utils.NewDecimal(int64(time.Second), 0),
					CompressFactor: 200000000,
				},
			},
			MinItems: 3,
			Count:    3,
		},
	}
	expected := map[string]uint64{
		"Event1": 200000000,
	}
	result := tcd.GetCompressFactor(eventMap)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, result)
	}
}

func TestStatMetricsStatPDDGetCompressFactor(t *testing.T) {
	eventMap := map[string]uint64{
		"Event1": 1000000,
	}
	pdd := &StatPDD{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"Event1": {
					Stat:           utils.NewDecimal(int64(time.Second), 0),
					CompressFactor: 200000000,
				},
			},
			MinItems: 3,
			Count:    3,
		},
	}
	expected := map[string]uint64{
		"Event1": 200000000,
	}
	result := pdd.GetCompressFactor(eventMap)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, result)
	}
}

func TestStatMetricsStatDDCGetCompressFactor(t *testing.T) {
	eventMap := map[string]uint64{
		"Event1": 1000000,
	}
	ddc := &StatDDC{
		Events: map[string]map[string]uint64{
			"Event1": {
				"Event1": 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}
	expected := map[string]uint64{
		"Event1": 200000000,
	}
	result := ddc.GetCompressFactor(eventMap)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", expected, result)
	}
}

type mockDP struct{}

func (mockDP) String() string {
	return ""
}

func (mockDP) FieldAsInterface(fldPath []string) (interface{}, error) {
	return nil, utils.ErrAccountNotFound
}

func (mockDP) FieldAsString([]string) (string, error) {
	return "", nil
}

func TestStatMetricsStatASRAddEventErr3(t *testing.T) {
	asr := &StatASR{Metric: NewMetric(2)}
	err := asr.AddEvent("EVENT_1", new(mockDP))
	if err == nil || err.Error() != utils.ErrAccountNotFound.Error() {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.ErrAccountNotFound, err)
	}
}
