/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package rater

import (
	"reflect"
	"testing"
	"time"
)

func TestSimpleMarshallerApStoreRestore(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{
		Months:    Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	result, err := ap.Store()
	expected := "1328106601000000000|;2;1;3,4;14:30:00;15:00:00;0;0;0;0;0"
	if err != nil || !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := &ActivationPeriod{}
	err = ap1.Restore(result)
	if err != nil || !reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v: %v", ap, ap1, err)
	}
}

func TestSimpleMarshallerApRestoreFromString(t *testing.T) {
	s := "1325376000000000000|;1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5,6,0;00:00:00;;10;0;0.2;60;1\n"
	ap := &ActivationPeriod{}
	err := ap.Restore(s)
	if err != nil || len(ap.Intervals) != 1 {
		t.Error("Error restoring activation period from string", ap)
	}
}

func TestRpStoreRestore(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{
		Months:    Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	rp := &RatingProfile{FallbackKey: "test"}
	rp.AddActivationPeriodIfNotPresent("0723", ap)
	result, err := rp.Store()
	expected := "test>0723=1328106601000000000|;2;1;3,4;14:30:00;15:00:00;0;0;0;0;0"
	if err != nil || !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %q was %q", expected, result)
	}
	rp1 := &RatingProfile{}
	err = rp1.Restore(result)
	if err != nil || !reflect.DeepEqual(rp, rp1) {
		t.Errorf("Expected %v was %v", rp, rp1)
	}
}

func TestActionTimingStoreRestore(t *testing.T) {
	i := &Interval{
		Months:         Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays:      MonthDays{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		WeekDays:       WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartTime:      "18:00:00",
		EndTime:        "00:00:00",
		Weight:         10.0,
		ConnectFee:     0.0,
		Price:          1.0,
		PricedUnits:    60,
		RateIncrements: 1,
	}
	at := &ActionTiming{
		Id:             "some uuid",
		Tag:            "test",
		UserBalanceIds: []string{"one", "two", "three"},
		Timing:         i,
		Weight:         10.0,
		ActionsId:      "Commando",
	}
	r, err := at.Store()
	if err != nil || r != "some uuid|test|one,two,three|;1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5;18:00:00;00:00:00;10;0;1;60;1|10|Commando" {
		t.Errorf("Error serializing action timing: %v", string(r))
	}
	o := &ActionTiming{}
	err = o.Restore(r)
	if err != nil || !reflect.DeepEqual(o, at) {
		t.Errorf("Expected %v was  %v", at, o)
	}
}

func TestActionTriggerStoreRestore(t *testing.T) {
	at := &ActionTrigger{
		Id:             "some_uuid",
		BalanceId:      CREDIT,
		Direction:      OUTBOUND,
		ThresholdValue: 100.0,
		DestinationId:  "NAT",
		Weight:         10.0,
		ActionsId:      "Commando",
	}
	r, err := at.Store()
	if err != nil || r != "some_uuid;MONETARY;OUT;NAT;Commando;100;10;false" {
		t.Errorf("Error serializing action trigger: %v", string(r))
	}
	o := &ActionTrigger{}
	err = o.Restore(r)
	if err != nil || !reflect.DeepEqual(o, at) {
		t.Errorf("Expected %v was  %v", at, o)
	}
}

func TestIntervalStoreRestore(t *testing.T) {
	i := &Interval{
		Months:         Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays:      MonthDays{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		WeekDays:       WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartTime:      "18:00:00",
		EndTime:        "00:00:00",
		Weight:         10.0,
		ConnectFee:     0.0,
		Price:          1.0,
		PricedUnits:    60,
		RateIncrements: 1,
	}
	r, err := i.Store()
	if err != nil || r != ";1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5;18:00:00;00:00:00;10;0;1;60;1" {
		t.Errorf("Error serializing interval: %v", string(r))
	}
	o := &Interval{}
	err = o.Restore(r)
	if err != nil || !reflect.DeepEqual(o, i) {
		t.Errorf("Expected %v was  %v", i, o)
	}
}

func TestIntervalRestoreFromString(t *testing.T) {
	s := ";1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5,6,0;00:00:00;;10;0;0.2;60;1"
	i := Interval{}
	err := i.Restore(s)
	if err != nil || i.Price != 0.2 {
		t.Errorf("Error restoring inteval period from string %+v", i)
	}
}

func TestMonthYearStoreRestore(t *testing.T) {
	y := Years{2010, 2011, 2012}
	r, err := y.Store()
	if err != nil || r != "2010,2011,2012" {
		t.Errorf("Error serializing years: %v", string(r))
	}
	o := Years{}
	err = o.Restore(r)
	if err != nil || !reflect.DeepEqual(o, y) {
		t.Errorf("Expected %v was  %v", y, o)
	}
}

func TestMonthStoreRestore(t *testing.T) {
	m := Months{5, 6, 7, 8}
	r, err := m.Store()
	if err != nil || r != "5,6,7,8" {
		t.Errorf("Error serializing months: %v", string(r))
	}
	o := Months{}
	err = o.Restore(r)
	if err != nil || !reflect.DeepEqual(o, m) {
		t.Errorf("Expected %v was  %v", m, o)
	}
}

func TestMonthDayStoreRestore(t *testing.T) {
	md := MonthDays{24, 25, 26}
	r, err := md.Store()
	if err != nil || r != "24,25,26" {
		t.Errorf("Error serializing month days: %v", string(r))
	}
	o := MonthDays{}
	err = o.Restore(r)
	if err != nil || !reflect.DeepEqual(o, md) {
		t.Errorf("Expected %v was  %v", md, o)
	}
}

func TestWeekDayStoreRestore(t *testing.T) {
	wd := WeekDays{time.Saturday, time.Sunday}
	r, err := wd.Store()
	if err != nil || r != "6,0" {
		t.Errorf("Error serializing week days: %v", string(r))
	}
	o := WeekDays{}
	err = o.Restore(r)
	if err != nil || !reflect.DeepEqual(o, wd) {
		t.Errorf("Expected %v was  %v", wd, o)
	}
}
