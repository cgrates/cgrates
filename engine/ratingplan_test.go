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

	"github.com/cgrates/cgrates/utils"

	"reflect"
	"testing"
	"time"
)

func TestApRestoreFromStorage(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   utils.OUT,
		Category:    "0",
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
		Category:    "0",
		Direction:   utils.OUT,
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
		Category:    "0",
		Direction:   utils.OUT,
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
		Category:    "0",
		Direction:   utils.OUT,
		Tenant:      "CUSTOMER_2",
		Subject:     "danb:87.139.12.167",
		Destination: "4123"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackNoDefault(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Category:    "0",
		Direction:   utils.OUT,
		Tenant:      "vdf",
		Subject:     "one",
		Destination: "0723"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackNoInfiniteLoop(t *testing.T) {
	cd := &CallDescriptor{Category: "0", Direction: utils.OUT, Tenant: "vdf", Subject: "rif", Destination: "0721"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackNoInfiniteLoopSelf(t *testing.T) {
	cd := &CallDescriptor{Category: "0", Direction: utils.OUT, Tenant: "vdf", Subject: "inf", Destination: "0721"}
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

func TestRatingPlanIsContinousEmpty(t *testing.T) {
	rpl := &RatingPlan{}
	if rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousBlank(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"blank": &RITiming{StartTime: "00:00:00"},
			"other": &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousGood(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"first":  &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second": &RITiming{WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":  &RITiming{WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanisContinousBad(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"first":  &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second": &RITiming{WeekDays: utils.WeekDays{4, 5, 0}, StartTime: "00:00:00"},
		},
	}
	if rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousSpecial(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special": &RITiming{Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first":   &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second":  &RITiming{WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":   &RITiming{WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousMultiple(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special":  &RITiming{Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first":    &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"first_08": &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "08:00:00"},
			"second":   &RITiming{WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":    &RITiming{WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousMissing(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special":  &RITiming{Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first_08": &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "08:00:00"},
			"second":   &RITiming{WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":    &RITiming{WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanSaneTimingsBad(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"one": &RITiming{Years: utils.Years{2015}, WeekDays: utils.WeekDays{time.Monday}, tag: "first"},
		},
	}
	if crazyTiming := rpl.getFirstUnsaneTiming(); crazyTiming == "" {
		t.Errorf("Error detecting bad timings in rating profile: %+v", rpl)
	}
}

func TestRatingPlanSaneTimingsGood(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"one": &RITiming{Years: utils.Years{2015}, tag: "first"},
			"two": &RITiming{WeekDays: utils.WeekDays{0, 1, 2, 3, 4}, StartTime: "00:00:00", tag: "second"},
		},
	}
	if crazyTiming := rpl.getFirstUnsaneTiming(); crazyTiming != "" {
		t.Errorf("Error detecting bad timings in rating profile: %+v", rpl)
	}
}

func TestRatingPlanSaneRatingsEqual(t *testing.T) {
	rpl := &RatingPlan{
		Ratings: map[string]*RIRate{
			"one": &RIRate{
				tag: "first",
				Rates: RateGroups{
					&Rate{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      30 * time.Second,
					},
					&Rate{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      30 * time.Second,
					},
				},
			},
		},
	}
	if crazyRating := rpl.getFirstUnsaneRating(); crazyRating == "" {
		t.Errorf("Error detecting bad rate groups in rating profile: %+v", rpl)
	}
}

func TestRatingPlanSaneRatingsNotMultiple(t *testing.T) {
	rpl := &RatingPlan{
		Ratings: map[string]*RIRate{
			"one": &RIRate{
				tag: "first",
				Rates: RateGroups{
					&Rate{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      30 * time.Second,
					},
					&Rate{
						GroupIntervalStart: 15 * time.Second,
						RateIncrement:      30 * time.Second,
					},
				},
			},
		},
	}
	if crazyRating := rpl.getFirstUnsaneRating(); crazyRating == "" {
		t.Errorf("Error detecting bad rate groups in rating profile: %+v", rpl)
	}
}

func TestRatingPlanSaneRatingsGoot(t *testing.T) {
	rpl := &RatingPlan{
		Ratings: map[string]*RIRate{
			"one": &RIRate{
				tag: "first",
				Rates: RateGroups{
					&Rate{
						GroupIntervalStart: 60 * time.Second,
						RateIncrement:      30 * time.Second,
						RateUnit:           1 * time.Second,
					},
					&Rate{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      30 * time.Second,
						RateUnit:           1 * time.Second,
					},
				},
			},
		},
	}
	if crazyRating := rpl.getFirstUnsaneRating(); crazyRating != "" {
		t.Errorf("Error detecting bad rate groups in rating profile: %+v", rpl)
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
	dm.SetRatingPlan(rp, utils.NonTransactional)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dm.GetRatingPlan(rp.Id, true, utils.NonTransactional)
	}
}
