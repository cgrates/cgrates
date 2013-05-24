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
	result, err := Marshal(ap)
	expected := []byte("1328106601000000000|;2;1;3,4;14:30:00;15:00:00;0;0;0;0;0")
	if err != nil || !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := &ActivationPeriod{}
	err = Unmarshal(result, ap1)
	if err != nil || !reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v: %v", ap, ap1, err)
	}
}

func TestSimpleMarshallerApRestoreFromString(t *testing.T) {
	s := []byte("1325376000000000000|;1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5,6,0;00:00:00;;10;0;0.2;60;1\n")
	ap := &ActivationPeriod{}
	err := Unmarshal(s, ap)
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
	result, err := Marshal(rp)
	expected := []byte("test>0723=1328106601000000000|;2;1;3,4;14:30:00;15:00:00;0;0;0;0;0")
	if err != nil || !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %q was %q", expected, result)
	}
	rp1 := &RatingProfile{}
	err = Unmarshal(result, rp1)
	if err != nil || !reflect.DeepEqual(rp, rp1) {
		t.Errorf("Expected %v was %v", rp, rp1)
	}
}
