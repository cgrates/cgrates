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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

func TestGetRatingProfileForPrefix(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 11, 18, 13, 45, 1, 0, time.UTC),
		TimeEnd:     time.Date(2013, 11, 18, 13, 47, 30, 0, time.UTC),
		Tenant:      "vdf",
		Category:    "0",
		Subject:     "fallback1",
		Destination: "0256098",
	}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 3 || !cd.continousRatingInfos() {
		t.Logf("0: %+v", cd.RatingInfos[0])
		t.Logf("1: %+v", cd.RatingInfos[1])
		t.Logf("2: %+v", cd.RatingInfos[2])
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continousRatingInfos())
	}
}

func TestGetRatingProfileForPrefixFirstEmpty(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 11, 18, 13, 44, 1, 0, time.UTC),
		TimeEnd:     time.Date(2013, 11, 18, 13, 47, 30, 0, time.UTC),
		Tenant:      "vdf",
		Category:    "0",
		Subject:     "fallback1",
		Destination: "0256098",
	}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 4 || !cd.continousRatingInfos() {
		t.Logf("0: %+v", cd.RatingInfos[0])
		t.Logf("1: %+v", cd.RatingInfos[1])
		t.Logf("2: %+v", cd.RatingInfos[2])
		t.Logf("3: %+v", cd.RatingInfos[3])
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continousRatingInfos())
	}
}

func TestGetRatingProfileNotFound(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2015, 8, 18, 22, 05, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 8, 18, 22, 06, 30, 0, time.UTC),
		Tenant:      "vdf",
		Category:    "0",
		Subject:     "no_rating_profile",
		Destination: "0256098",
	}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 || !cd.continousRatingInfos() {
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continousRatingInfos())
	}
}

func TestGetRatingProfileFoundButNoDestination(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2015, 8, 18, 22, 05, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 8, 18, 22, 06, 30, 0, time.UTC),
		Tenant:      "cgrates.org",
		Category:    "call",
		Subject:     "nt",
		Destination: "447956",
	}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continousRatingInfos())
	}
}

func TestRatingProfileRISorter(t *testing.T) {
	ris := RateIntervalList{
		&RateInterval{
			Timing: &RITiming{
				StartTime: "09:00:00",
			},
		},
		&RateInterval{
			Timing: &RITiming{
				StartTime: "00:00:00",
			},
		},
		&RateInterval{
			Timing: &RITiming{
				StartTime: "19:00:00",
			},
		},
	}
	sorter := &RateIntervalTimeSorter{referenceTime: time.Date(2016, 1, 6, 19, 0, 0, 0, time.UTC), ris: ris}
	rIntervals := sorter.Sort()
	if len(rIntervals) != 3 ||
		rIntervals[0].Timing.StartTime != "00:00:00" ||
		rIntervals[1].Timing.StartTime != "09:00:00" ||
		rIntervals[2].Timing.StartTime != "19:00:00" {
		t.Error("Wrong interval list: ", utils.ToIJSON(rIntervals))
	}
}

func TestRatingProfileRIforTSOne(t *testing.T) {
	ri := &RatingInfo{
		RateIntervals: RateIntervalList{
			&RateInterval{
				Timing: &RITiming{
					StartTime: "09:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "00:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "19:00:00",
				},
			},
		},
	}
	ts := &TimeSpan{
		TimeStart: time.Date(2016, 1, 6, 19, 0, 0, 0, time.UTC),
		TimeEnd:   time.Date(2016, 1, 6, 19, 1, 30, 0, time.UTC),
	}
	rIntervals := ri.SelectRatingIntevalsForTimespan(ts)
	if len(rIntervals) != 1 || rIntervals[0].Timing.StartTime != "19:00:00" {
		t.Error("Wrong interval list: ", utils.ToIJSON(rIntervals))
	}
}

func TestRatingProfileRIforTSTwo(t *testing.T) {
	ri := &RatingInfo{
		RateIntervals: RateIntervalList{
			&RateInterval{
				Timing: &RITiming{
					StartTime: "09:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "00:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "19:00:00",
				},
			},
		},
	}
	ts := &TimeSpan{
		TimeStart: time.Date(2016, 1, 6, 10, 0, 0, 0, time.UTC),
		TimeEnd:   time.Date(2016, 1, 6, 19, 1, 30, 0, time.UTC),
	}
	rIntervals := ri.SelectRatingIntevalsForTimespan(ts)
	if len(rIntervals) != 2 ||
		rIntervals[0].Timing.StartTime != "09:00:00" ||
		rIntervals[1].Timing.StartTime != "19:00:00" {
		t.Error("Wrong interval list: ", utils.ToIJSON(rIntervals))
	}
}

func TestRatingProfileRIforTSThree(t *testing.T) {
	ri := &RatingInfo{
		RateIntervals: RateIntervalList{
			&RateInterval{
				Timing: &RITiming{
					StartTime: "09:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "00:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "19:00:00",
				},
			},
		},
	}
	ts := &TimeSpan{
		TimeStart: time.Date(2016, 1, 6, 8, 0, 0, 0, time.UTC),
		TimeEnd:   time.Date(2016, 1, 6, 19, 1, 30, 0, time.UTC),
	}
	rIntervals := ri.SelectRatingIntevalsForTimespan(ts)
	if len(rIntervals) != 3 ||
		rIntervals[0].Timing.StartTime != "00:00:00" ||
		rIntervals[1].Timing.StartTime != "09:00:00" ||
		rIntervals[2].Timing.StartTime != "19:00:00" {
		t.Error("Wrong interval list: ", utils.ToIJSON(rIntervals))
	}
}

func TestRatingProfileRIforTSMidnight(t *testing.T) {
	ri := &RatingInfo{
		RateIntervals: RateIntervalList{
			&RateInterval{
				Timing: &RITiming{
					StartTime: "09:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "00:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "19:00:00",
				},
			},
		},
	}
	ts := &TimeSpan{
		TimeStart: time.Date(2016, 1, 6, 23, 40, 0, 0, time.UTC),
		TimeEnd:   time.Date(2016, 1, 7, 1, 1, 30, 0, time.UTC),
	}
	rIntervals := ri.SelectRatingIntevalsForTimespan(ts)
	if len(rIntervals) != 1 ||
		rIntervals[0].Timing.StartTime != "19:00:00" {
		t.Error("Wrong interval list: ", utils.ToIJSON(rIntervals))
	}
}

func TestRatingProfileYearMonthDay(t *testing.T) {
	ri := &RatingInfo{
		RateIntervals: RateIntervalList{
			&RateInterval{
				Timing: &RITiming{
					StartTime: "09:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "00:00:00",
				},
			},
			&RateInterval{
				Timing: &RITiming{
					Years:     utils.Years{2016},
					Months:    utils.Months{1},
					MonthDays: utils.MonthDays{6, 7},
					WeekDays:  utils.WeekDays{},
					StartTime: "19:00:00",
				},
			},
		},
	}
	ts := &TimeSpan{
		TimeStart: time.Date(2016, 1, 6, 23, 40, 0, 0, time.UTC),
		TimeEnd:   time.Date(2016, 1, 7, 1, 1, 30, 0, time.UTC),
	}
	rIntervals := ri.SelectRatingIntevalsForTimespan(ts)
	if len(rIntervals) != 1 ||
		rIntervals[0].Timing.StartTime != "19:00:00" {
		t.Error("Wrong interval list: ", utils.ToIJSON(rIntervals))
	}
}

func TestRatingProfileWeighted(t *testing.T) {
	ri := &RatingInfo{
		RateIntervals: RateIntervalList{
			&RateInterval{
				Timing: &RITiming{
					StartTime: "09:00:00",
				},
				Weight: 10,
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "00:00:00",
				},
				Weight: 10,
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "19:00:00",
				},
				Weight: 10,
			},
			&RateInterval{
				Timing: &RITiming{
					Years:     utils.Years{2016},
					Months:    utils.Months{1},
					MonthDays: utils.MonthDays{6},
					WeekDays:  utils.WeekDays{},
					StartTime: "00:00:00",
				},
				Weight: 11,
			},
		},
	}
	ts := &TimeSpan{
		TimeStart: time.Date(2016, 1, 6, 23, 40, 0, 0, time.UTC),
		TimeEnd:   time.Date(2016, 1, 6, 23, 45, 30, 0, time.UTC),
	}
	rIntervals := ri.SelectRatingIntevalsForTimespan(ts)
	if len(rIntervals) != 1 ||
		rIntervals[0].Timing.StartTime != "00:00:00" ||
		rIntervals[0].Weight != 11 {
		t.Error("Wrong interval list: ", utils.ToIJSON(rIntervals))
	}
}

func TestRatingProfileWeightedMultiple(t *testing.T) {
	ri := &RatingInfo{
		RateIntervals: RateIntervalList{
			&RateInterval{
				Timing: &RITiming{
					StartTime: "09:00:00",
				},
				Weight: 10,
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "00:00:00",
				},
				Weight: 10,
			},
			&RateInterval{
				Timing: &RITiming{
					StartTime: "19:00:00",
				},
				Weight: 10,
			},
			&RateInterval{
				Timing: &RITiming{
					Years:     utils.Years{2016},
					Months:    utils.Months{1},
					MonthDays: utils.MonthDays{6},
					WeekDays:  utils.WeekDays{},
					StartTime: "00:00:00",
				},
				Weight: 11,
			},
			&RateInterval{
				Timing: &RITiming{
					Years:     utils.Years{2016},
					Months:    utils.Months{1},
					MonthDays: utils.MonthDays{6},
					WeekDays:  utils.WeekDays{},
					StartTime: "18:00:00",
				},
				Weight: 11,
			},
		},
	}
	ts := &TimeSpan{
		TimeStart: time.Date(2016, 1, 6, 17, 40, 0, 0, time.UTC),
		TimeEnd:   time.Date(2016, 1, 6, 23, 45, 30, 0, time.UTC),
	}
	rIntervals := ri.SelectRatingIntevalsForTimespan(ts)
	if len(rIntervals) != 2 ||
		rIntervals[0].Timing.StartTime != "00:00:00" ||
		rIntervals[0].Weight != 11 ||
		rIntervals[1].Timing.StartTime != "18:00:00" ||
		rIntervals[1].Weight != 11 {
		t.Error("Wrong interval list: ", utils.ToIJSON(rIntervals))
	}
}

func TestRatingProfileSubjectPrefixMatching(t *testing.T) {
	rpSubjectPrefixMatching = true
	rp, err := RatingProfileSubjectPrefixMatching("*out:cgrates.org:data:rif")
	if rp == nil || err != nil {
		t.Errorf("Error getting rating profile by prefix: %+v (%v)", rp, err)
	}

	rp, err = RatingProfileSubjectPrefixMatching("*out:cgrates.org:data:rifescu")
	if rp == nil || err != nil {
		t.Errorf("Error getting rating profile by prefix: %+v (%v)", rp, err)
	}
	rpSubjectPrefixMatching = false
}

func TestRatingProfileGetRatingPlansPrefixAny(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	defer func() {
		dm = tmpDm
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	rpf := &RatingProfile{
		Id: "*out:cgrates.org:call:1001",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:   "RP_2CNT",
		}},
	}
	cd := &CallDescriptor{
		ToR:      "*prepaid",
		Tenant:   "cgrates.org",
		Category: "call", Account: "1001",
		Subject: "1001", Destination: "1003",
		TimeStart: time.Date(2023, 11, 7, 8, 42, 26, 0, time.UTC),
		TimeEnd:   time.Date(2023, 11, 7, 8, 42, 26, 0, time.UTC).Add(50 * time.Second),
	}
	dm.SetRatingPlan(&RatingPlan{
		Id: "RP_2CNT",
		Timings: map[string]*RITiming{
			"30eab300": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": {
				Rates: []*RGRate{
					{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      60 * time.Second,
						RateUnit:           60 * time.Second,
					},
				},
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"*any": []*RPRate{
				{
					Timing: "30eab300",
					Rating: "b457f86d",
					Weight: 10,
				},
			},
		},
	})

	dm.SetReverseDestination("DEST", []string{"1003"}, "")
	SetDataStorage(dm)
	if err := rpf.GetRatingPlansForPrefix(cd); err != nil {
		t.Error("Error getting rating plans for prefix: ", err)
	}

}

func TestEngineSwapRpasIndex(t *testing.T) {
	activationTime1 := time.Now()
	activationTime2 := activationTime1.Add(time.Hour)
	plan1 := &RatingPlanActivation{
		ActivationTime: activationTime1,
		RatingPlanId:   "planTest1",
		FallbackKeys:   []string{"key1", "key2"},
	}
	plan2 := &RatingPlanActivation{
		ActivationTime: activationTime2,
		RatingPlanId:   "planTest2",
		FallbackKeys:   []string{"key3"},
	}
	initialRPAs := RatingPlanActivations{plan1, plan2}
	want := RatingPlanActivations{plan2, plan1}
	rpas := make(RatingPlanActivations, len(initialRPAs))
	copy(rpas, initialRPAs)
	rpas.Swap(0, 1)
	diff := cmp.Diff(rpas, want)
	if diff != "" {
		t.Errorf("Expected RatingPlanActivations after swap to be equal to want: (-got, +want)\n%s", diff)
	}
}

func TestEngineEqualRpaOrpa(t *testing.T) {
	activationTime1 := time.Now()
	activationTime2 := activationTime1.Add(time.Hour)
	plan1 := &RatingPlanActivation{
		ActivationTime: activationTime1,
		RatingPlanId:   "planId1",
		FallbackKeys:   []string{"key1", "key2"},
	}
	plan2 := &RatingPlanActivation{
		ActivationTime: activationTime2,
		RatingPlanId:   "planId2",
		FallbackKeys:   []string{"key3"},
	}
	tests := []struct {
		name     string
		rpa      *RatingPlanActivation
		orpa     *RatingPlanActivation
		expected bool
	}{
		{name: "Equal plans", rpa: plan1, orpa: plan1, expected: true},
		{name: "Different activation time", rpa: plan1, orpa: plan2, expected: false},
		{name: "Nil arguments (both nil)", rpa: nil, orpa: nil, expected: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result bool
			if tc.rpa == nil && tc.orpa == nil {
				result = tc.expected
			} else {
				result = tc.rpa.Equal(tc.orpa)
			}
			if result != tc.expected {
				t.Errorf("Expected Equal(%v, %v) to be %v, got %v", tc.rpa, tc.orpa, tc.expected, result)
			}
		})
	}
}

func TestEngineSwapIndexRis(t *testing.T) {
	ris := RatingInfos{
		&RatingInfo{MatchedSubject: "Cgrates1"},
		&RatingInfo{MatchedSubject: "Cgrates2"},
		&RatingInfo{MatchedSubject: "Cgrates3"},
	}
	i, j := 0, 2
	ris.Swap(i, j)
	if ris[i].MatchedSubject != "Cgrates3" || ris[j].MatchedSubject != "Cgrates1" {
		t.Errorf("Swap did not swap elements correctly")
	}
}

func TestEngineLessRis(t *testing.T) {
	ris := RatingInfos{
		&RatingInfo{ActivationTime: time.Now()},
		&RatingInfo{ActivationTime: time.Now().Add(time.Hour)},
		&RatingInfo{ActivationTime: time.Now().Add(-time.Hour)},
	}
	if ris.Less(0, 2) {
		t.Errorf("Expected first element not to be less than the third one")
	}
}

func TestRatingInfosString(t *testing.T) {
	ratingInfos := RatingInfos{
		{MatchedSubject: "Info1", RatingPlanId: "Value1"},
		{MatchedSubject: "Info2", RatingPlanId: "Value2"},
	}
	expectedJSON, err := json.MarshalIndent(ratingInfos, "", " ")
	if err != nil {
		t.Fatalf("Failed to marshal expected JSON: %v", err)
	}
	result := ratingInfos.String()
	if result != string(expectedJSON) {
		t.Errorf("Expected %s, but got %s", string(expectedJSON), result)
	}
}
