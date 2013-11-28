/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"encoding/json"

	"github.com/cgrates/cgrates/utils"

	"reflect"
	"testing"
	"time"
)

func TestApRestoreFromStorage(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   OUTBOUND,
		TOR:         "0",
		Tenant:      "CUSTOMER_1",
		Subject:     "rif:from:tm",
		Destination: "49"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Errorf("Error restoring activation periods: %+v", cd.RatingInfos[0])
	}
}

func TestApStoreRestoreJson(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{
		Months:    []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)
	result, _ := json.Marshal(ap)
	ap1 := &RatingPlan{}
	json.Unmarshal(result, ap1)
	if !reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestApStoreRestoreBlank(t *testing.T) {
	i := &RateInterval{}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)
	result, _ := json.Marshal(ap)
	ap1 := RatingPlan{}
	json.Unmarshal(result, &ap1)
	if reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestFallbackDirect(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		TOR:         "0",
		Direction:   OUTBOUND,
		Tenant:      "CUSTOMER_2",
		Subject:     "danb:87.139.12.167",
		Destination: "41"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackMultiple(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		TOR:         "0",
		Direction:   OUTBOUND,
		Tenant:      "vdf",
		Subject:     "fall",
		Destination: "0723045"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Errorf("Error restoring rating plans: %+v", cd.RatingInfos)
	}
}

func TestFallbackWithBackTrace(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		TOR:         "0",
		Direction:   OUTBOUND,
		Tenant:      "CUSTOMER_2",
		Subject:     "danb:87.139.12.167",
		Destination: "4123"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackDefault(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		TOR:         "0",
		Direction:   OUTBOUND,
		Tenant:      "vdf",
		Subject:     "one",
		Destination: "0723"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackNoInfiniteLoop(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: OUTBOUND, Tenant: "vdf", Subject: "rif", Destination: "0721"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackNoInfiniteLoopSelf(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: OUTBOUND, Tenant: "vdf", Subject: "inf", Destination: "0721"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestApAddIntervalIfNotPresent(t *testing.T) {
	i1 := &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{time.February},
			MonthDays: utils.MonthDays{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	i2 := &RateInterval{Timing: &RITiming{
		Months:    utils.Months{time.February},
		MonthDays: utils.MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}}
	i3 := &RateInterval{Timing: &RITiming{
		Months:    utils.Months{time.February},
		MonthDays: utils.MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}}
	rp := &RatingPlan{}
	rp.AddRateInterval("NAT", i1)
	rp.AddRateInterval("NAT", i2)
	if len(rp.DestinationRates["NAT"]) != 1 {
		t.Error("Wronfullyrppended interval ;)")
	}
	t.Log()
	rp.AddRateInterval("NAT", i3)
	if len(rp.DestinationRates["NAT"]) != 2 {
		t.Error("Wronfully not appended interval ;)", rp.DestinationRates)
	}
}

func TestApAddRateIntervalGroups(t *testing.T) {
	i1 := &RateInterval{
		Rating: &RIRate{Rates: RateGroups{&Rate{0, 1, 1 * time.Second, 1 * time.Second}}},
	}
	i2 := &RateInterval{
		Rating: &RIRate{Rates: RateGroups{&Rate{30 * time.Second, 2, 1 * time.Second, 1 * time.Second}}},
	}
	i3 := &RateInterval{
		Rating: &RIRate{Rates: RateGroups{&Rate{30 * time.Second, 2, 1 * time.Second, 1 * time.Second}}},
	}
	ap := &RatingPlan{}
	ap.AddRateInterval("NAT", i1)
	ap.AddRateInterval("NAT", i2)
	ap.AddRateInterval("NAT", i3)
	if len(ap.DestinationRates) != 1 {
		t.Error("Wronfully appended interval ;)")
	}
	if len(ap.RateIntervalList("NAT")[0].Rating.Rates) != 1 {
		t.Errorf("Group prices not formed: %#v", ap.RateIntervalList("NAT")[0].Rating.Rates[0])
	}
}

func TestGetActiveForCall(t *testing.T) {
	rpas := RatingPlanActivations{
		&RatingPlanActivation{ActivationTime: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC)},
		&RatingPlanActivation{ActivationTime: time.Date(2013, 11, 12, 11, 40, 0, 0, time.UTC)},
		&RatingPlanActivation{ActivationTime: time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC)},
	}
	cd := &CallDescriptor{
		TimeStart: time.Date(2013, 11, 12, 11, 39, 0, 0, time.UTC),
		TimeEnd:   time.Date(2013, 11, 12, 11, 45, 0, 0, time.UTC),
	}
	active := rpas.GetActiveForCall(cd)
	if len(active) != 2 {
		t.Errorf("Error getting active rating plans: %+v", active)
	}
}

/**************************** Benchmarks *************************************/

func BenchmarkRatingPlanMarshalJson(b *testing.B) {
	b.StopTimer()
	i := &RateInterval{
		Timing: &RITiming{
			Months:    []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)

	ap1 := RatingPlan{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		result, _ := json.Marshal(ap)
		json.Unmarshal(result, &ap1)
	}
}

func BenchmarkRatingPlanMarshal(b *testing.B) {
	b.StopTimer()
	i := &RateInterval{
		Timing: &RITiming{Months: []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)

	ap1 := &RatingPlan{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		result, _ := marsh.Marshal(ap)
		marsh.Unmarshal(result, ap1)
	}
}

func BenchmarkRatingPlanRestore(b *testing.B) {
	i := &RateInterval{
		Timing: &RITiming{Months: []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	rp := &RatingPlan{Id: "test"}
	rp.AddRateInterval("NAT", i)
	storageGetter.SetRatingPlan(rp)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storageGetter.GetRatingPlan(rp.Id, true)
	}
}
