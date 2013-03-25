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
	"encoding/json"
	"reflect"
	"testing"
	"time"
	//"log"
)

func TestApStoreRestore(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{
		Months:    Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	result := ap.store()
	expected := "1328106601000000000|;2;1;3,4;14:30:00;15:00:00;0;0;0;0;0"
	if result != expected {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := ActivationPeriod{}
	ap1.restore(result)
	if reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestApRestoreFromString(t *testing.T) {
	s := "1325376000000000000|;1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5,6,0;00:00:00;;10;0;0.2;60;1\n"
	ap := ActivationPeriod{}
	ap.restore(s)
	if len(ap.Intervals) != 1 {
		t.Error("Error restoring activation period from string", ap)
	}
}

func TestApRestoreFromStorage(t *testing.T) {
	cd := &CallDescriptor{
		Direction:   "OUT",
		TOR:         "0",
		Tenant:      "CUSTOMER_1",
		Subject:     "rif:from:tm",
		Destination: "49"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 2 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestApStoreRestoreJson(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	result, _ := json.Marshal(ap)
	expected := "{\"ActivationTime\":\"2012-02-01T14:30:01Z\",\"Intervals\":[{\"Years\":null,\"Months\":[2],\"MonthDays\":[1],\"WeekDays\":[3,4],\"StartTime\":\"14:30:00\",\"EndTime\":\"15:00:00\",\"Weight\":0,\"ConnectFee\":0,\"Price\":0,\"PricedUnits\":0,\"RateIncrements\":0}]}"
	if string(result) != expected {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := ActivationPeriod{}
	json.Unmarshal(result, &ap1)
	if reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestApStoreRestoreBlank(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	result, _ := json.Marshal(ap)
	expected := "{\"ActivationTime\":\"2012-02-01T14:30:01Z\",\"Intervals\":[{\"Years\":null,\"Months\":null,\"MonthDays\":null,\"WeekDays\":null,\"StartTime\":\"\",\"EndTime\":\"\",\"Weight\":0,\"ConnectFee\":0,\"Price\":0,\"PricedUnits\":0,\"RateIncrements\":0}]}"
	if string(result) != expected {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := ActivationPeriod{}
	json.Unmarshal(result, &ap1)
	if reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestFallbackDirect(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: "OUT", Tenant: "CUSTOMER_2", Subject: "danb:87.139.12.167", Destination: "41"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestFallbackWithBackTrace(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: "OUT", Tenant: "CUSTOMER_2", Subject: "danb:87.139.12.167", Destination: "4123"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestFallbackDefault(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: "OUT", Tenant: "CUSTOMER_2", Subject: "danb:87.139.12.167", Destination: "4123"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestFallbackNoInfiniteLoop(t *testing.T) {
	cd := &CallDescriptor{Tenant: "vdf", Subject: "rif", Destination: "0721"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestApAddIntervalIfNotPresent(t *testing.T) {
	i1 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	i2 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	i3 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{}
	ap.AddIntervalIfNotPresent(i1)
	ap.AddIntervalIfNotPresent(i2)
	if len(ap.Intervals) != 1 {
		t.Error("Wronfully appended interval ;)")
	}
	ap.AddIntervalIfNotPresent(i3)
	if len(ap.Intervals) != 2 {
		t.Error("Wronfully not appended interval ;)")
	}
}

/**************************** Benchmarks *************************************/

func BenchmarkActivationPeriodStoreRestoreJson(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)

	ap1 := ActivationPeriod{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		result, _ := json.Marshal(ap)
		json.Unmarshal(result, &ap1)
	}
}

func BenchmarkActivationPeriodRestore(b *testing.B) {
	ap := ActivationPeriod{}
	for i := 0; i < b.N; i++ {
		ap.restore("1328106601000000000|;2;1;3,4;14:30:00;15:00:00;0;0;0;0;0")
	}
}

func BenchmarkActivationPeriodStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)

	ap1 := ActivationPeriod{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		result := ap.store()
		ap1.restore(result)
	}
}
