/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package engine

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestASRGetStringValue(t *testing.T) {
	asr := NewASR(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "33.333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(ev3.ID)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	asr.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
	asr.RemEvent(ev.ID)
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "66.667%" {
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
	asr := NewASR(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := "EVENT_2"
	ev4 := "EVENT_4"
	if err := asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := asr.AddEvent(ev2, utils.MapStorage{utils.MetaOpts: utils.MapStorage{}}); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if err := asr.AddEvent(ev2, utils.MapStorage{utils.MetaOpts: utils.MapStorage{}}); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "33.333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if err := asr.AddEvent(ev4, utils.MapStorage{utils.MetaOpts: utils.MapStorage{}}); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "25%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if err := asr.RemEvent(ev4); err != nil {
		t.Error(err)
	}
	if err := asr.RemEvent(ev2); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
		t.Error(utils.ToJSON(asr))
	}
	if err := asr.RemEvent(ev2); err != nil {
		t.Error(err)
	}
	if err := asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts}); err != nil {
		t.Error(err)
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestASRGetStringValue3(t *testing.T) {
	asr := &StatASR{Metric: NewMetric(2, nil)}
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev4 := &utils.CGREvent{ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	v := expected.Events["EVENT_1"]
	v.Stat = utils.NewDecimalFromFloat64(0.5)
	v.CompressFactor = 2
	v = expected.Events["EVENT_2"]
	v.Stat = utils.NewDecimalFromFloat64(0)
	v.CompressFactor = 2
	expected.Count = 4
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "25%" {
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
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
}

func TestASRGetValue(t *testing.T) {
	asr := NewASR(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := asr.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong asr value: %f", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	asr.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	if v := asr.GetValue(); v.Compare(utils.NewDecimalFromFloat64(33.33333333333333)) != 0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(ev3.ID)
	if v := asr.GetValue(); v.Compare(utils.NewDecimalFromFloat64(50.0)) != 0 {
		t.Errorf("wrong asr value: %f", v)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	asr.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
	asr.RemEvent(ev.ID)
	if v := asr.GetValue(); v.Compare(utils.NewDecimalFromFloat64(66.66666666666667)) != 0 {
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
	asr := &StatASR{Metric: NewMetric(2, nil)}
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev4 := &utils.CGREvent{ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := asr.Compress(10, "EVENT_3")
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
		Metric: &Metric{
			Value:    utils.NewDecimal(1, 0),
			Count:    2,
			MinItems: 2,
			Events: map[string]*DecimalWithCompress{
				"EVENT_3": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 2},
			},
		},
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := asr.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	if !reflect.DeepEqual(*expected, *asr) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(0.25)
	v.CompressFactor = 4
	expected.Count = 4
	if rply := asr.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
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
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	asr := NewASR(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev4 := &utils.CGREvent{ID: "EVENT_1"}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = asr.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	asr.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = asr.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	asr.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = asr.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestACDGetStringValue(t *testing.T) {
	acd := NewACD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaUsage:     10 * time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	if err := acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts}); err != nil {
		t.Error(err)
	}
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	if err := acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err == nil || err.Error() != "NOT_FOUND:*usage" {
		t.Error(err)
	}
	if err := acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err == nil || err.Error() != "NOT_FOUND:*usage" {
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
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaUsage:     478433753 * time.Nanosecond,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage:     30*time.Second + 982433452*time.Nanosecond,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.73s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.73s" {
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
	acd := NewACD(2, "", nil)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 2 * time.Minute}}
	if err := acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m30s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m15s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
	acd.RemEvent(ev2.ID)
	if strVal := acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m20s" {
		t.Errorf("wrong acd value: %s", strVal)
	}
}

func TestACDGetStringValue3(t *testing.T) {
	acd := &StatACD{Metric: NewMetric(2, nil)}
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	if err := acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err != nil {
		t.Error(err)
	}
	acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.NewDecimal(int64(3*time.Minute+30*time.Second), 0)
	acd.RemEvent(ev1.ID)
	acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
}

func TestACDCompress(t *testing.T) {
	acd := &StatACD{Metric: NewMetric(2, nil)}
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	acd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := acd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acd.Metric) {
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)

	expIDs = []string{"EVENT_3"}
	if rply := acd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
	}
}

func TestACDGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	acd := NewACD(2, "", nil)

	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	ev4 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaUsage: 2 * time.Minute}}

	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = acd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = acd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = acd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestACDGetFloat64Value(t *testing.T) {
	acd := NewACD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %v", v)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute,
			utils.MetaStartTime: time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute + 30*time.Second,
			utils.MetaStartTime: time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	if strVal := acd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(35.0*1e9)) != 0 {
		t.Errorf("wrong acd value: %v", strVal)
	}
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
	// by default rounding decimal is 5
	if strVal := acd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(53.33333333333333*1e9)) != 0 {
		t.Errorf("wrong acd value: %v", strVal)
	}

	acd.RemEvent(ev2.ID)
	if strVal := acd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(53.33333333333333*1e9)) != 0 {
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
	acd := NewACD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := acd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong acd value: %+v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     8 * time.Second}}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	acd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	acd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
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
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage:     4*time.Minute + 30*time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	acd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	acd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	tcd := NewTCD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaUsage:     10 * time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{
			utils.MetaUsage:     10 * time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
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
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute + 30*time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	tcd := NewTCD(2, "", nil)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 2 * time.Minute}}
	if err := tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "3m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "5m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
	tcd.RemEvent(ev2.ID)
	if strVal := tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "4m0s" {
		t.Errorf("wrong tcd value: %s", strVal)
	}
}

func TestTCDGetStringValue3(t *testing.T) {
	tcd := &StatTCD{Metric: NewMetric(2, nil)}
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	if err := tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err != nil {
		t.Error(err)
	}
	tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.NewDecimalFromFloat64(float64(3*time.Minute + 30*time.Second))
	tcd.RemEvent(ev1.ID)
	tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
}

func TestTCDGetFloat64Value(t *testing.T) {
	tcd := NewTCD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %f", v)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute + 30*time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	if strVal := tcd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(70.0*1e9)) != 0 {
		t.Errorf("wrong tcd value: %f", strVal)
	}
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	tcd := NewTCD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := tcd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong tcd value: %+v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     5 * time.Second}}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
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
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute + 30*time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	tcd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	tcd := &StatTCD{Metric: NewMetric(2, nil)}
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 2 * time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 3 * time.Minute}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	tcd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	tcd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := tcd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcd.Metric) {
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)

	expIDs = []string{"EVENT_3"}
	if rply := tcd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
	}
}

func TestTCDGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	tcd := NewTCD(2, "", nil)

	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaUsage: time.Minute}}
	ev4 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaUsage: 2 * time.Minute}}

	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = tcd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = tcd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = tcd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestACCGetStringValue(t *testing.T) {
	acc := NewACC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      12.3}}
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      12.3}}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "12.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev3.ID)
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong acc value: %s", strVal)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      5.6}}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      1.2}}
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	acc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	expErr := "NEGATIVE:*cost"
	if err := acc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: utils.MapStorage{
		utils.MetaCost: -1,
	}}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s received %v", expErr, err)
	}
}

func TestACCGetStringValue2(t *testing.T) {
	acc := NewACC(2, "", nil)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 12.3}}
	if err := acc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.8" {
		t.Errorf("wrong acc value: %s", strVal)
	}
	acc.RemEvent(ev2.ID)
	if strVal := acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.3" {
		t.Errorf("wrong acc value: %s", strVal)
	}
}

func TestACCGetStringValue3(t *testing.T) {
	acc := &StatACC{Metric: NewMetric(2, nil)}
	expected := &StatACC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.2"), CompressFactor: 2},
				"EVENT_3": {Stat: &utils.Decimal{Big: decimal.WithContext(decimal.Context{Precision: 3}).SetFloat64(18.3)}, CompressFactor: 1},
			},
			MinItems: 2,
			Count:    3,
			Value:    utils.NewDecimalFromStringIgnoreError("42.7"),
		},
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 6.2}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	if err := acc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err != nil {
		t.Error(err)
	}
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acc.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.SubstractDecimal(expected.Value, utils.NewDecimalFromFloat64(12.2))
	acc.RemEvent(ev1.ID)
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acc.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
}

func TestACCGetValue(t *testing.T) {
	acc := NewACC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "12.3"}}
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	acc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acc value: %v", strVal)
	}
	acc.RemEvent(ev3.ID)
	if strVal := acc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong acc value: %v", strVal)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "5.6"}}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "1.2"}}
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	acc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	acc := &StatACC{Metric: NewMetric(2, nil)}
	expected := &StatACC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("18.199999999999999289457264239899814128875732421875"), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimalFromStringIgnoreError("6.20000000000000017763568394002504646778106689453125"), CompressFactor: 1},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromStringIgnoreError("24.4"),
			Count:    2,
		},
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 6.2}}
	ev4 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := acc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acc.Metric) {
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := acc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acc.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(12.22500000000000)
	v.CompressFactor = 4
	expected.Count = 4
	expected.Value = utils.NewDecimalFromFloat64(48.9)
	if rply := acc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	acc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(acc.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
	}
}

func TestACCGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	acc := NewACC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev4 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = acc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = acc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	acc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = acc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestTCCGetStringValue(t *testing.T) {
	tcc := NewTCC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      12.3}}
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      5.7}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "18" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev3.ID)
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      5.6}}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      1.2}}
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	tcc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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

	expErr := "NEGATIVE:*cost"
	if err := tcc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: utils.MapStorage{
		utils.MetaCost: -1,
	}}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s received %v", expErr, err)
	}
}

func TestTCCGetStringValue2(t *testing.T) {
	tcc := NewTCC(2, "", nil)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 12.3}}
	if err := tcc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "30.6" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "67.2" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
	tcc.RemEvent(ev2.ID)
	if strVal := tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "48.9" {
		t.Errorf("wrong tcc value: %s", strVal)
	}
}

func TestTCCGetStringValue3(t *testing.T) {
	tcc := &StatTCC{Metric: NewMetric(2, nil)}
	expected := &StatTCC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.20000000000000"), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.300000000000000710542735760100185871124267578125"), CompressFactor: 1},
			},
			MinItems: 2,
			Count:    3,
			Value:    utils.NewDecimalFromStringIgnoreError("42.700"),
		},
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 6.2}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	if err := tcc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err != nil {
		t.Error(err)
	}
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcc.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.SubstractDecimal(expected.Value, utils.NewDecimalFromFloat64(12.2))
	tcc.RemEvent(ev1.ID)
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcc.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
}

func TestTCCGetValue(t *testing.T) {
	tcc := NewTCC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "12.3"}}
	if strVal := tcc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := tcc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      1.2}}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	tcc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	if strVal := tcc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(13.5)) != 0 {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	tcc.RemEvent(ev3.ID)
	if strVal := tcc.GetValue(); strVal != utils.DecimalNaN {
		t.Errorf("wrong tcc value: %v", strVal)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "5.6"}}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "1.2"}}
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	tcc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	tcc := &StatTCC{Metric: NewMetric(2, nil)}
	expected := &StatTCC{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("18.199999999999999289457264239899814128875732421875"), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimalFromStringIgnoreError("6.20000000000000017763568394002504646778106689453125"), CompressFactor: 1},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromStringIgnoreError("24.400"),
			Count:    2,
		},
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 6.2}}
	ev4 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := tcc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcc.Metric) {
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := tcc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcc.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(12.225)
	v.CompressFactor = 4
	expected.Count = 4
	expected.Value = utils.NewDecimalFromFloat64(48.9)
	if rply := tcc.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	tcc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(tcc.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
	}
}

func TestTCCGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	tcc := NewTCC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev4 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = tcc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = tcc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	tcc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = tcc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestPDDGetStringValue(t *testing.T) {
	pdd := NewPDD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaPDD:       5 * time.Second,
			utils.MetaUsage:     10 * time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		}}
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
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
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaPDD:       10 * time.Second,
			utils.MetaUsage:     time.Minute,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{utils.MetaPDD: 10 * time.Second},
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts, utils.MetaReq: ev4.Event})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	pdd := NewPDD(2, "", nil)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaPDD: 2 * time.Minute}}
	if err := pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaPDD: time.Minute}}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m30s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m15s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
	pdd.RemEvent(ev2.ID)
	if strVal := pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "1m20s" {
		t.Errorf("wrong pdd value: %s", strVal)
	}
}

func TestPDDGetStringValue3(t *testing.T) {
	pdd := &StatPDD{Metric: NewMetric(2, nil)}
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaPDD: 2 * time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaPDD: 3 * time.Minute}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaPDD: time.Minute}}
	if err := pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err != nil {
		t.Error(err)
	}
	pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(pdd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.NewDecimalFromFloat64(float64(3*time.Minute + 30*time.Second))
	pdd.RemEvent(ev1.ID)
	pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(pdd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
}

func TestPDDGetFloat64Value(t *testing.T) {
	pdd := NewPDD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaPDD:       5 * time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %v", v)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaPDD:       10 * time.Second,
			utils.MetaUsage:     time.Minute,
			utils.MetaStartTime: time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage:     time.Minute + 30*time.Second,
			utils.MetaStartTime: time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	if strVal := pdd.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(7.5*1e9)) != 0 {
		t.Errorf("wrong pdd value: %v", strVal)
	}
	pdd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	pdd := NewPDD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaPDD:       9 * time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong pdd value: %+v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{
			utils.MetaPDD:       10 * time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     8 * time.Second}}
	ev3 := &utils.CGREvent{ID: "EVENT_3"}
	if err := pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err == nil || err.Error() != "NOT_FOUND:*pdd" {
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
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaPDD:       8 * time.Second,
			utils.MetaUsage:     time.Minute,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage:     4*time.Minute + 30*time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := pdd.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts}); err == nil || err.Error() != "NOT_FOUND:*pdd" {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong *pdd value: %+v", v)
	}
	if err := pdd.RemEvent(ev5.ID); err == nil || err.Error() != "NOT_FOUND" {
		t.Error(err)
	}
	if err := pdd.RemEvent(ev4.ID); err != nil {
		t.Error(err)
	}
	if v := pdd.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong *pdd value: %+v", v)
	}
}

func TestPDDCompress(t *testing.T) {
	pdd := &StatPDD{Metric: NewMetric(2, nil)}
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaPDD: 2 * time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaPDD: 3 * time.Minute}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaPDD: time.Minute}}
	pdd.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	pdd.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := pdd.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(pdd.Metric) {
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)

	expIDs = []string{"EVENT_3"}
	if rply := pdd.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	pdd.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(pdd.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
	}
}

func TestPDDGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	pdd := NewPDD(2, "", nil)

	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaPDD: time.Minute}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaPDD: time.Minute}}
	ev4 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaPDD: 2 * time.Minute}}

	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = pdd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	pdd.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = pdd.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	pdd.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = pdd.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestDDCGetStringValue(t *testing.T) {
	ddc := NewDDC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}

	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{
			utils.MetaDestination: "1001",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
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
	ddc := NewDDC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaPDD:         5 * time.Second,
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:       10 * time.Second}}
	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := ddc.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong ddc value: %v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if v := ddc.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong ddc value: %v", v)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaDestination: "1001",
			utils.MetaPDD:         10 * time.Second,
			utils.MetaUsage:       time.Minute,
			utils.MetaStartTime:   time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaDestination: "1003",
			utils.MetaUsage:       time.Minute + 30*time.Second,
			utils.MetaStartTime:   time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	if strVal := ddc.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(2)) != 0 {
		t.Errorf("wrong ddc value: %v", strVal)
	}
	ddc.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	statDistinct := NewDDC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaDestination: "1002"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
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
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaDestination: "1002"}}
	ddc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
	rply = ddc.Compress(10, "EVENT_3")
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
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	ddc := NewDDC(2, "", nil)

	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaDestination: "1002"}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}
	ev4 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = ddc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = ddc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = ddc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatSumGetFloat64Value(t *testing.T) {
	statSum := NewStatSum(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaPDD:         5 * time.Second,
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:       10 * time.Second}}
	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := statSum.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statSum value: %v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	if err := statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err == nil ||
		err != utils.ErrNotFound {
		t.Error(err)
	}
	if v := statSum.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statSum value: %v", v)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaDestination: "1001",
			utils.MetaPDD:         10 * time.Second,
			utils.MetaCost:        "20",
			utils.MetaUsage:       time.Minute,
			utils.MetaStartTime:   time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaDestination: "1003",
			utils.MetaCost:        "20",
			utils.MetaUsage:       time.Minute + 30*time.Second,
			utils.MetaStartTime:   time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	statSum.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	if strVal := statSum.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(40)) != 0 {
		t.Errorf("wrong statSum value: %v", strVal)
	}
	statSum.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	statSum := NewStatSum(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}

	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{
			utils.MetaDestination: "1001",
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	statSum.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
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
	statSum := NewStatSum(2, "~*opts.*cost", nil)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 12.3}}
	if err := statSum.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "30.6" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "67.2" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
	statSum.RemEvent(ev2.ID)
	if strVal := statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "48.9" {
		t.Errorf("wrong statSum value: %s", strVal)
	}
}

func TestStatSumGetStringValue3(t *testing.T) {
	statSum := &StatSum{
		Metric: NewMetric(2, nil),
		Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
	}
	expected := &StatSum{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.2"), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.3"), CompressFactor: 1},
			},
			MinItems: 2,
			Count:    3,
			Value:    utils.NewDecimalFromStringIgnoreError("42.7"),
		},
		Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{
		ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaCost: utils.NewDecimal(182, 1),
		},
	}
	ev2 := &utils.CGREvent{
		ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaCost: utils.NewDecimal(62, 1),
		},
	}
	ev3 := &utils.CGREvent{
		ID: "EVENT_3",
		APIOpts: map[string]any{
			utils.MetaCost: utils.NewDecimal(183, 1),
		},
	}
	if err := statSum.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := statSum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := statSum.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err != nil {
		t.Error(err)
	}
	statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(statSum.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statSum))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.SubstractDecimal(expected.Value, utils.NewDecimalFromFloat64(12.2))
	statSum.RemEvent(ev1.ID)
	statSum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(statSum.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statSum))
	}
}

func TestStatSumCompress(t *testing.T) {
	sum := &StatSum{
		Metric: NewMetric(2, nil),
		Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
	}
	expected := &StatSum{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("18.2"), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimalFromStringIgnoreError("6.2"), CompressFactor: 1},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromStringIgnoreError("24.4"),
			Count:    2,
		},
		Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{
		ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaCost: utils.NewDecimal(182, 1),
		},
	}
	ev2 := &utils.CGREvent{
		ID: "EVENT_2",
		APIOpts: map[string]any{
			utils.MetaCost: utils.NewDecimal(62, 1),
		},
	}
	ev4 := &utils.CGREvent{
		ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaCost: utils.NewDecimal(183, 1),
		},
	}
	sum.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := sum.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(sum.Metric) {
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
		Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := sum.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(sum.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	sum.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(12.225)
	v.CompressFactor = 4
	expected.Count = 4
	expected.Value = utils.NewDecimalFromFloat64(48.9)
	if rply := sum.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	sum.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(sum.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
	}
}

func TestStatSumGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	sum := NewStatSum(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev4 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	sum.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = sum.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	sum.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = sum.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	sum.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = sum.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatAverageGetFloat64Value(t *testing.T) {
	statAvg := NewStatAverage(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:       10 * time.Second,
			utils.MetaPDD:         5 * time.Second,
			utils.MetaDestination: "1002"}}
	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := statAvg.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statAvg value: %v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if v := statAvg.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statAvg value: %v", v)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaCost:        "30",
			utils.MetaUsage:       time.Minute,
			utils.MetaStartTime:   time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaPDD:         10 * time.Second,
			utils.MetaDestination: "1001",
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaCost:        "20",
			utils.MetaUsage:       time.Minute + 30*time.Second,
			utils.MetaStartTime:   time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaDestination: "1003",
		},
	}
	statAvg.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	if strVal := statAvg.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(25)) != 0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
	if strVal := statAvg.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(23.33333333333333)) != 0 {
		t.Errorf("wrong statAvg value: %v", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(23.33333333333333)) != 0 {
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
	statAvg := NewStatAverage(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaDestination: "1002"}}
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong ddc value: %s", strVal)
	}

	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaDestination: "1002"}}

	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaDestination: "1001"}}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	statAvg.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
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
	statAvg := NewStatAverage(2, "~*opts.*cost", nil)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 12.3}}
	if err := statAvg.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "15.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.8" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
	statAvg.RemEvent(ev2.ID)
	if strVal := statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != "16.3" {
		t.Errorf("wrong statAvg value: %s", strVal)
	}
}

func TestStatAverageGetStringValue3(t *testing.T) {
	statAvg := &StatAverage{Metric: NewMetric(2, nil), FieldName: "~*opts.*cost"}
	expected := &StatAverage{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.20000000000000"), CompressFactor: 2},
				"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.300000000000000710542735760100185871124267578125"), CompressFactor: 1},
			},
			MinItems: 2,
			Count:    3,
			Value:    utils.NewDecimalFromStringIgnoreError("42.70000000000000"),
		},
		FieldName: "~*opts.*cost",
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 6.2}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	if err := statAvg.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := statAvg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts}); err != nil {
		t.Error(err)
	}
	if err := statAvg.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts}); err != nil {
		t.Error(err)
	}
	statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(statAvg.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statAvg))
	}
	v := expected.Events[ev1.ID]
	v.CompressFactor = 1
	expected.Count = 2
	expected.Value = utils.SubstractDecimal(expected.Value, utils.NewDecimalFromFloat64(12.2))
	statAvg.RemEvent(ev1.ID)
	statAvg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *statAvg) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(statAvg))
	}
}

func TestStatAverageCompress(t *testing.T) {
	avg := &StatAverage{Metric: NewMetric(2, nil), FieldName: "~*opts.*cost"}
	expected := &StatAverage{
		Metric: &Metric{
			Events: map[string]*DecimalWithCompress{
				"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("18.199999999999999289457264239899814128875732421875"), CompressFactor: 1},
				"EVENT_2": {Stat: utils.NewDecimalFromStringIgnoreError("6.20000000000000017763568394002504646778106689453125"), CompressFactor: 1},
			},
			MinItems: 2,
			Value:    utils.NewDecimalFromStringIgnoreError("24.40000000000000"),
			Count:    2,
		},
		FieldName: "~*opts.*cost",
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 6.2}}
	ev4 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.3}}
	avg.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_2"}
	rply := avg.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(avg.Metric) {
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
		FieldName: "~*opts.*cost",
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	expIDs = []string{"EVENT_3"}
	if rply := avg.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	avg.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !expected.Equal(avg.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	avg.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	v := expected.Events["EVENT_3"]
	v.Stat = utils.NewDecimalFromFloat64(12.22500000000000)
	v.CompressFactor = 4
	expected.Count = 4
	expected.Value = utils.NewDecimalFromFloat64(48.90000000000000)
	if rply := avg.Compress(1, "EVENT_3"); !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	if !expected.Equal(avg.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
	}
}

func TestStatAverageGetCompressFactor(t *testing.T) {
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	avg := NewStatAverage(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	ev4 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: 18.2}}
	avg.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = avg.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	avg.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = avg.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	avg.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	expectedCF["EVENT_1"] = 2
	CF["EVENT_2"] = 3
	if CF = avg.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

func TestStatDistinctGetFloat64Value(t *testing.T) {
	statDistinct := NewStatDistinct(2, "~*opts.*usage", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaUsage: 10 * time.Second}}
	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if v := statDistinct.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statDistinct value: %v", v)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2"}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if v := statDistinct.GetValue(); v != utils.DecimalNaN {
		t.Errorf("wrong statDistinct value: %v", v)
	}
	ev4 := &utils.CGREvent{ID: "EVENT_4",
		APIOpts: map[string]any{
			utils.MetaUsage: time.Minute,
		},
	}
	ev5 := &utils.CGREvent{ID: "EVENT_5",
		APIOpts: map[string]any{
			utils.MetaUsage: time.Minute + 30*time.Second,
		},
	}
	if err := statDistinct.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts}); err != nil {
		t.Error(err)
	}
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.NewDecimalFromFloat64(2)) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.AddEvent(ev5.ID, utils.MapStorage{utils.MetaOpts: ev5.APIOpts})
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
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.DecimalNaN) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
	statDistinct.RemEvent(ev5.ID)
	if strVal := statDistinct.GetValue(); strVal.Compare(utils.DecimalNaN) != 0 {
		t.Errorf("wrong statDistinct value: %v", strVal)
	}
}

func TestStatDistinctGetStringValue(t *testing.T) {
	statDistinct := NewStatDistinct(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: "20"}}
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaCost: "20"}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaCost: "40"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	statDistinct.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
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
	statDistinct := NewStatDistinct(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: "20"}}
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}

	statDistinct.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if strVal := statDistinct.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
		t.Errorf("wrong statDistinct value: %s", strVal)
	}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaCost: "40"}}
	statDistinct.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
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
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    2,
		FieldName:   utils.DynamicDataPrefix + utils.MetaOpts + utils.NestingSep + utils.MetaDestination,
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
		FieldName: utils.DynamicDataPrefix + utils.MetaOpts + utils.NestingSep + utils.MetaDestination,
		Count:     3,
	}
	expected.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	ev1 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}
	ev2 := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}
	ev3 := &utils.CGREvent{ID: "EVENT_3",
		APIOpts: map[string]any{utils.MetaDestination: "1002"}}
	ddc.AddEvent(ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	ddc.AddEvent(ev3.ID, utils.MapStorage{utils.MetaOpts: ev3.APIOpts})
	expIDs := []string{"EVENT_1", "EVENT_3"}
	rply := ddc.Compress(10, "EVENT_3")
	sort.Strings(rply)
	if !reflect.DeepEqual(expIDs, rply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expIDs), utils.ToJSON(rply))
	}
	ddc.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals)
	if !reflect.DeepEqual(*expected, *ddc) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
	}
	rply = ddc.Compress(10, "EVENT_3")
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
	var CF map[string]uint64
	expectedCF := map[string]uint64{
		"EVENT_1": 1,
		"EVENT_2": 1,
	}
	ddc := NewStatDistinct(2, utils.DynamicDataPrefix+utils.MetaOpts+utils.NestingSep+utils.MetaDestination, nil)

	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{utils.MetaDestination: "1002"}}
	ev2 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}
	ev4 := &utils.CGREvent{ID: "EVENT_2",
		APIOpts: map[string]any{utils.MetaDestination: "1001"}}

	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	if CF = ddc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev2.ID, utils.MapStorage{utils.MetaOpts: ev2.APIOpts})
	expectedCF["EVENT_2"] = 2
	if CF = ddc.GetCompressFactor(make(map[string]uint64)); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
	ddc.AddEvent(ev4.ID, utils.MapStorage{utils.MetaOpts: ev4.APIOpts})
	expectedCF["EVENT_2"] = 3
	CF["EVENT_2"] = 3
	if CF = ddc.GetCompressFactor(CF); !reflect.DeepEqual(expectedCF, CF) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedCF), utils.ToJSON(CF))
	}
}

var jMarshaler utils.JSONMarshaler

func TestASRMarshal(t *testing.T) {
	asr, err := NewStatMetric(utils.MetaASR, 2, []string{"*string:Account:1001"})
	if err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	asr.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var nasr StatASR
	expected := []byte(`{"Value":1,"Count":1,"Events":{"EVENT_1":{"Stat":1,"CompressFactor":1}},"MinItems":2,"FilterIDs":["*string:Account:1001"]}`)
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
	acd := NewACD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	acd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var nacd StatACD
	expected := []byte(`{"Value":10000000000,"Count":1,"Events":{"EVENT_1":{"Stat":10000000000,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`)
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
	tcd := NewTCD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	tcd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var ntcd StatTCD
	expected := []byte(`{"Value":10000000000,"Count":1,"Events":{"EVENT_1":{"Stat":10000000000,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`)
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
	acc := NewACC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "12.3"}}
	acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var nacc StatACC
	expected := []byte(`{"Value":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`)
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
	tcc := NewTCC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "12.3"}}
	tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var ntcc StatTCC
	expected := []byte(`{"Value":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`)
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
	pdd := NewPDD(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaPDD:       5 * time.Second,
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     10 * time.Second}}
	pdd.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var ntdd StatPDD
	expected := []byte(`{"Value":5000000000,"Count":1,"Events":{"EVENT_1":{"Stat":5000000000,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`)
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
	ddc := NewDDC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaPDD:         5 * time.Second,
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:       10 * time.Second}}
	ddc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var nddc StatDDC
	expected := []byte(`{"FieldValues":{"1002":{"EVENT_1":{}}},"Events":{"EVENT_1":{"1002":1}},"MinItems":2,"Count":1,"FilterIDs":null}`)
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
	statSum := NewStatSum(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaPDD:         5 * time.Second,
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:       10 * time.Second}}
	statSum.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var nstatSum StatSum
	expected := []byte(`{"Value":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FilterIDs":null,"Fields":[{"Rules":"~*opts.*cost","Path":"~*opts.*cost"}]}`)
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
	statAvg := NewStatAverage(2, "~*opts.*cost", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaPDD:         5 * time.Second,
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:       10 * time.Second}}
	statAvg.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	var nstatAvg StatAverage
	expected := []byte(`{"Value":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FilterIDs":null,"FieldName":"~*opts.*cost"}`)
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
	statDistinct := NewStatDistinct(2, "~*opts.*usage", nil)
	statDistinct.AddEvent("EVENT_1", utils.MapStorage{
		utils.MetaOpts: map[string]any{
			utils.MetaDestination: "1002",
			utils.MetaPDD:         5 * time.Second,
			utils.MetaCost:        "20",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:       10 * time.Second}})
	var nStatDistinct StatDistinct
	expected := []byte(`{"FieldValues":{"10s":{"EVENT_1":{}}},"Events":{"EVENT_1":{"10s":1}},"MinItems":2,"FieldName":"~*opts.*usage","Count":1,"FilterIDs":null}`)
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
	asr := NewASR(2, "", nil)
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
	asr := NewASR(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}}
	if strVal := asr.GetStringValue(config.CgrConfig().GeneralCfg().RoundingDecimals); strVal != utils.NotAvailable {
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
	err := dst.AddEvent("Event1", utils.MapStorage{utils.MetaOpts: ev.APIOpts})
	if err == nil || err.Error() != "invalid format for field <Test_Field_Name>" {
		t.Errorf("\nExpecting <invalid format for field <Test_Field_Name>>,\n Recevied <%+v>", err)
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
		Metric:    NewMetric(10, nil),
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
		Metric: NewMetric(20, nil),
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
	acd := &StatACD{Metric: NewMetric(2, nil)}
	result := acd.GetMinItems()
	if !reflect.DeepEqual(acd.MinItems, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", acd.MinItems, result)
	}
}

func TestStatMetricsStatTCDGetMinItems(t *testing.T) {
	tcd := &StatTCD{Metric: NewMetric(2, nil)}
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
	acc := &StatACC{Metric: NewMetric(2, nil)}
	result := acc.GetMinItems()
	if !reflect.DeepEqual(acc.MinItems, result) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", acc.MinItems, result)
	}
}

func TestStatMetricsStatTCCGetMinItems(t *testing.T) {
	tcc := &StatTCC{Metric: NewMetric(2, nil)}
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
	pdd := &StatPDD{Metric: NewMetric(2, nil)}
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
	asr := &StatASR{Metric: NewMetric(2, nil)}
	err := asr.AddEvent("EVENT_1", utils.MapStorage{utils.MetaOpts: map[string]any{utils.MetaStartTime: "10"}})
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatASRAddEventErr2(t *testing.T) {
	asr := &StatASR{Metric: NewMetric(2, nil)}
	err := asr.AddEvent("EVENT_1", utils.MapStorage{utils.MetaOpts: utils.MapStorage{utils.MetaStartTime: false}})
	if err == nil || err.Error() != "cannot convert field: false to time.Time" {
		t.Errorf("\nExpecting <cannot convert field: false to time.Time>,\n Recevied <%+v>", err)
	}
}

func TestStatMetricsStatACDAddEventErr(t *testing.T) {
	acd := NewMetric(2, nil)
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

func (mockDP) FieldAsInterface(fldPath []string) (any, error) {
	return nil, utils.ErrAccountNotFound
}

func (mockDP) FieldAsString([]string) (string, error) {
	return "", nil
}

func TestStatMetricsStatASRAddEventErr3(t *testing.T) {
	asr := &StatASR{Metric: NewMetric(2, nil)}
	err := asr.AddEvent("EVENT_1", new(mockDP))
	if err == nil || err.Error() != utils.ErrAccountNotFound.Error() {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.ErrAccountNotFound, err)
	}
}

func TestStatASRClone(t *testing.T) {

	asr := &StatASR{Metric: NewMetric(2, nil)}

	if rcv := asr.Clone(); !reflect.DeepEqual(rcv, asr) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", asr, rcv)
	}
}

func TestStatACDClone(t *testing.T) {

	acd := &StatACD{Metric: NewMetric(2, nil)}

	if rcv := acd.Clone(); !reflect.DeepEqual(rcv, acd) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", acd, rcv)
	}
}

func TestStatACCClone(t *testing.T) {

	acc := &StatACC{Metric: NewMetric(2, nil)}

	if rcv := acc.Clone(); !reflect.DeepEqual(rcv, acc) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", acc, rcv)
	}
}

func TestStatTCCClone(t *testing.T) {

	tcc := &StatTCC{Metric: NewMetric(2, nil)}

	if rcv := tcc.Clone(); !reflect.DeepEqual(rcv, tcc) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", tcc, rcv)
	}
}

func TestStatPDDClone(t *testing.T) {

	pdd := &StatPDD{Metric: NewMetric(2, nil)}

	if rcv := pdd.Clone(); !reflect.DeepEqual(rcv, pdd) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", pdd, rcv)
	}
}

func TestStatSumClone(t *testing.T) {
	sum := &StatSum{
		Metric: NewMetric(2, nil),
		Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
	}
	if rcv := sum.Clone(); !reflect.DeepEqual(rcv, sum) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", sum, rcv)
	}
}

func TestStatAverageClone(t *testing.T) {

	avg := &StatAverage{Metric: NewMetric(2, nil), FieldName: "~*opts.*cost"}

	if rcv := avg.Clone(); !reflect.DeepEqual(rcv, avg) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", avg, rcv)
	}
}

func TestStatTCDClone(t *testing.T) {

	sum := &StatTCD{Metric: NewMetric(2, nil)}

	if rcv := sum.Clone(); !reflect.DeepEqual(rcv, sum) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", sum, rcv)
	}
}

func TestStatDDCClone(t *testing.T) {

	ddc := &StatDDC{
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

	if rcv := ddc.Clone(); !reflect.DeepEqual(rcv, ddc) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", ddc, rcv)
	}
}

func TestStatDistinctClone(t *testing.T) {

	dst := &StatDistinct{
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

	if rcv := dst.Clone(); !reflect.DeepEqual(rcv, dst) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", dst, rcv)
	}
}

func TestACCAddEventErr(t *testing.T) {
	acc := NewACC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "wrong input"}}

	expErr := "can't convert <wrong input> to decimal"
	if err := acc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: <%v> received <%v>", expErr, err)
	}

}

func TestTCCAddEventErr(t *testing.T) {
	tcc := NewTCC(2, "", nil)
	ev := &utils.CGREvent{ID: "EVENT_1",
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaCost:      "wrong input"}}

	expErr := "can't convert <wrong input> to decimal"
	if err := tcc.AddEvent(ev.ID, utils.MapStorage{utils.MetaOpts: ev.APIOpts}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: <%v> received <%v>", expErr, err)
	}

}

func TestDDCGetFilterIDs(t *testing.T) {

	ddc := NewDDC(2, "", []string{"flt1", "flt2"})

	exp := &StatDDC{
		FilterIDs: []string{"flt1", "flt2"},
	}

	if rcv := ddc.GetFilterIDs(); !reflect.DeepEqual(rcv, exp.FilterIDs) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.ToJSON(exp.FilterIDs), utils.ToJSON(rcv))
	}

}

func TestMetricClone(t *testing.T) {

	sum := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"Event1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	if rcv := sum.Clone(); !reflect.DeepEqual(rcv, sum) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", sum, rcv)
	}
}

func TestMetricEqualFalse(t *testing.T) {

	sum := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"Event1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	sum2 := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"Event1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
			"Event2": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	if rcv := sum.Equal(sum2); rcv {
		t.Errorf("Expecting to not be equal, Recevied equal <%v>", rcv)
	}
}

func TestMetricEqualEventFalse(t *testing.T) {

	sum := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"even1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	sum2 := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"even1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	if rcv := sum.Equal(sum2); rcv {
		t.Errorf("Expecting to not be equal, Recevied equal <%v>", rcv)
	}
}

func TestStatDistinctGetFilterIDs(t *testing.T) {

	dst := NewStatDistinct(2, "", []string{"flt1", "flt2"})

	exp := &StatDistinct{
		FilterIDs: []string{"flt1", "flt2"},
	}

	if rcv := dst.GetFilterIDs(); !reflect.DeepEqual(rcv, exp.FilterIDs) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.ToJSON(exp.FilterIDs), utils.ToJSON(rcv))
	}

}

func TestMetricAddOneEvent(t *testing.T) {
	tests := []struct {
		name        string
		initialVal  *utils.Decimal
		initialCnt  uint64
		input       any
		expectErr   bool
		expectVal   *decimal.Big
		expectCount uint64
	}{
		{
			name:        "Int input",
			input:       42,
			expectErr:   false,
			expectVal:   utils.NewDecimal(42, 0).Big,
			expectCount: 1,
		},
		{
			name:        "Duration input",
			input:       time.Duration(5),
			expectErr:   false,
			expectVal:   utils.NewDecimal(5, 0).Big,
			expectCount: 1,
		},
		{
			name:        "Add to existing value",
			initialVal:  &utils.Decimal{Big: utils.NewDecimal(10, 0).Big},
			initialCnt:  1,
			input:       15,
			expectErr:   false,
			expectVal:   utils.NewDecimal(25, 0).Big,
			expectCount: 2,
		},
		{
			name:        "Invalid type input",
			input:       struct{}{},
			expectErr:   true,
			expectVal:   nil,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				Value: tt.initialVal,
				Count: tt.initialCnt,
			}

			err := m.addOneEvent(tt.input)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectVal == nil {
				if m.Value != nil {
					t.Errorf("Expected nil Value, got: %v", m.Value.Big)
				}
			} else {
				if m.Value == nil || m.Value.Big.Cmp(tt.expectVal) != 0 {
					t.Errorf("Expected Value: %v, got: %v", tt.expectVal, m.Value.Big)
				}
			}

			if m.Count != tt.expectCount {
				t.Errorf("Expected Count: %d, got: %d", tt.expectCount, m.Count)
			}
		})
	}
}

func TestStatDDCClones(t *testing.T) {
	original := &StatDDC{
		FieldValues: map[string]utils.StringSet{
			"field1": utils.NewStringSet([]string{"ID", "ID1"}),
		},
		Events: map[string]map[string]uint64{
			"cgrates.org": {
				"val1": 5,
			},
		},
		MinItems:  2,
		Count:     10,
		FilterIDs: []string{"f1", "f2"},
	}

	cloned := original.Clone().(*StatDDC)

	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Cloned StatDDC is not equal to original\nOriginal: %+v\nCloned: %+v", original, cloned)
	}

	cloned.Count = 20
	cloned.Events["cgrates.org"]["val1"] = 99
	cloned.FieldValues["field1"].Add("c")
	cloned.FilterIDs[0] = "modified"

	if reflect.DeepEqual(original, cloned) {
		t.Error("Original StatDDC changed after modifying clone")
	}

	t.Run("nil receiver returns nil", func(t *testing.T) {
		var ddc *StatDDC
		if ddc.Clone() != nil {
			t.Error("Expected nil Clone result from nil receiver, got non-nil")
		}
	})
}

func TestStatDistinctClones(t *testing.T) {
	original := &StatDistinct{
		FieldValues: map[string]utils.StringSet{
			"field1": utils.NewStringSet([]string{"ID", "ID1"}),
		},
		Events: map[string]map[string]uint64{
			"cgrates.org": {
				"val1": 5,
			},
		},
		MinItems:  2,
		Count:     10,
		FieldName: "testField",
		FilterIDs: []string{"f1", "f2"},
	}

	cloned := original.Clone().(*StatDistinct)

	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Cloned StatDistinct is not equal to original\nOriginal: %+v\nCloned: %+v", original, cloned)
	}

	cloned.Count = 20
	cloned.Events["cgrates.org"]["val1"] = 99
	cloned.FieldValues["field1"].Add("c")
	cloned.FilterIDs[0] = "modified"
	cloned.FieldName = "modifiedField"

	if reflect.DeepEqual(original, cloned) {
		t.Error("Original StatDistinct changed after modifying clone")
	}

	t.Run("nil receiver returns nil", func(t *testing.T) {
		var dst *StatDistinct
		if dst.Clone() != nil {
			t.Error("Expected nil Clone result from nil receiver, got non-nil")
		}
	})
}
