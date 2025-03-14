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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestApRestoreFromStorage(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
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
		Tenant:      "vdf",
		Subject:     "one",
		Destination: "0723"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackNoInfiniteLoop(t *testing.T) {
	cd := &CallDescriptor{Category: "0",
		Tenant: "vdf", Subject: "rif", Destination: "0721"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestFallbackNoInfiniteLoopSelf(t *testing.T) {
	cd := &CallDescriptor{Category: "0",
		Tenant: "vdf", Subject: "inf", Destination: "0721"}
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
		Rating: &RIRate{Rates: RateGroups{&RGRate{0, 1, time.Second, time.Second}}},
	}
	i2 := &RateInterval{
		Rating: &RIRate{Rates: RateGroups{&RGRate{30 * time.Second, 2, time.Second, time.Second}}},
	}
	i3 := &RateInterval{
		Rating: &RIRate{Rates: RateGroups{&RGRate{30 * time.Second, 2, time.Second, time.Second}}},
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
			"blank": {StartTime: "00:00:00"},
			"other": {WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousGood(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"first":  {WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second": {WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":  {WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanisContinousBad(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"first":  {WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second": {WeekDays: utils.WeekDays{4, 5, 0}, StartTime: "00:00:00"},
		},
	}
	if rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousSpecial(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special": {Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first":   {WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second":  {WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":   {WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousMultiple(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special":  {Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first":    {WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"first_08": {WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "08:00:00"},
			"second":   {WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":    {WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanIsContinousMissing(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special":  {Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first_08": {WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "08:00:00"},
			"second":   {WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":    {WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRatingPlanSaneTimingsBad(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"one": {Years: utils.Years{2015}, WeekDays: utils.WeekDays{time.Monday}, tag: "first"},
		},
	}
	if crazyTiming := rpl.getFirstUnsaneTiming(); crazyTiming == "" {
		t.Errorf("Error detecting bad timings in rating profile: %+v", rpl)
	}
}

func TestRatingPlanSaneTimingsGood(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"one": {Years: utils.Years{2015}, tag: "first"},
			"two": {WeekDays: utils.WeekDays{0, 1, 2, 3, 4}, StartTime: "00:00:00", tag: "second"},
		},
	}
	if crazyTiming := rpl.getFirstUnsaneTiming(); crazyTiming != "" {
		t.Errorf("Error detecting bad timings in rating profile: %+v", rpl)
	}
}

func TestRatingPlanSaneRatingsEqual(t *testing.T) {
	rpl := &RatingPlan{
		Ratings: map[string]*RIRate{
			"one": {
				tag: "first",
				Rates: RateGroups{
					&RGRate{
						GroupIntervalStart: 0,
						RateIncrement:      30 * time.Second,
					},
					&RGRate{
						GroupIntervalStart: 0,
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
			"one": {
				tag: "first",
				Rates: RateGroups{
					&RGRate{
						GroupIntervalStart: 0,
						RateIncrement:      30 * time.Second,
					},
					&RGRate{
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
			"one": {
				tag: "first",
				Rates: RateGroups{
					&RGRate{
						GroupIntervalStart: 60 * time.Second,
						RateIncrement:      30 * time.Second,
						RateUnit:           time.Second,
					},
					&RGRate{
						GroupIntervalStart: 0,
						RateIncrement:      30 * time.Second,
						RateUnit:           time.Second,
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
	dm.SetRatingPlan(rp)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dm.GetRatingPlan(rp.Id, true, utils.NonTransactional)
	}
}

func TestEqual(t *testing.T) {
	rp1 := &RatingPlan{Id: "1"}
	rp2 := &RatingPlan{Id: "1"}
	if !rp1.Equal(rp2) {
		t.Errorf("expected RatingPlan instances with the same Id to be equal")
	}
	rp1 = &RatingPlan{Id: "1"}
	rp2 = &RatingPlan{Id: "2"}
	if rp1.Equal(rp2) {
		t.Errorf("expected RatingPlan instances with different Ids to be not equal")
	}
}

func TestRatingPlanIsContinous(t *testing.T) {
	tests := []struct {
		name     string
		rp       *RatingPlan
		expected bool
	}{
		{
			name: "Test with all weekdays covered and midnight start",
			rp: &RatingPlan{
				Timings: map[string]*RITiming{
					"timing1": {
						WeekDays:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday},
						StartTime: "00:00:00",
					},
				},
			},
			expected: true,
		},
		{
			name: "Test with not all weekdays covered",
			rp: &RatingPlan{
				Timings: map[string]*RITiming{
					"timing1": {
						WeekDays:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday},
						StartTime: "00:00:00",
					},
				},
			},
			expected: false,
		},
		{
			name: "Test with non-midnight start time",
			rp: &RatingPlan{
				Timings: map[string]*RITiming{
					"timing1": {
						WeekDays:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday},
						StartTime: "01:00:00",
					},
				},
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rp.isContinous()
			if result != tt.expected {
				t.Errorf("Expected isContinous to be %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestRatingPlanEqual(t *testing.T) {
	tests := []struct {
		name     string
		rp1      *RatingPlan
		rp2      *RatingPlan
		expected bool
	}{
		{
			name: "Equal Id",
			rp1: &RatingPlan{
				Id: "plan1",
			},
			rp2: &RatingPlan{
				Id: "plan1",
			},
			expected: true,
		},
		{
			name: "Different Id",
			rp1: &RatingPlan{
				Id: "plan1",
			},
			rp2: &RatingPlan{
				Id: "plan2",
			},
			expected: false,
		},
		{
			name: "Empty Ids",
			rp1: &RatingPlan{
				Id: "",
			},
			rp2: &RatingPlan{
				Id: "",
			},
			expected: true,
		},
		{
			name: "One empty Id",
			rp1: &RatingPlan{
				Id: "plan1",
			},
			rp2: &RatingPlan{
				Id: "",
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rp1.Equal(tt.rp2)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestRateGroupsSwap(t *testing.T) {
	tests := []struct {
		name   string
		i, j   int
		before RateGroups
		after  RateGroups
	}{
		{
			name: "Swap first and second element",
			i:    0,
			j:    1,
			before: RateGroups{
				&RGRate{GroupIntervalStart: 10, Value: 100},
				&RGRate{GroupIntervalStart: 20, Value: 200},
			},
			after: RateGroups{
				&RGRate{GroupIntervalStart: 20, Value: 200},
				&RGRate{GroupIntervalStart: 10, Value: 100},
			},
		},
		{
			name: "Swap same elements (no change)",
			i:    0,
			j:    0,
			before: RateGroups{
				&RGRate{GroupIntervalStart: 10, Value: 100},
				&RGRate{GroupIntervalStart: 20, Value: 200},
			},
			after: RateGroups{
				&RGRate{GroupIntervalStart: 10, Value: 100},
				&RGRate{GroupIntervalStart: 20, Value: 200},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := append(RateGroups(nil), tt.before...)
			pg.Swap(tt.i, tt.j)
			for idx, expected := range tt.after {
				if pg[idx].GroupIntervalStart != expected.GroupIntervalStart || pg[idx].Value != expected.Value {
					t.Errorf("After Swap at index %d: expected %v, got %v", idx, expected, pg[idx])
				}
			}
		})
	}
}
