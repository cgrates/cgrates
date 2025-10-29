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
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
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
	if len(cd.RatingInfos) != 3 || !cd.continuousRatingInfos() {
		t.Logf("0: %+v", cd.RatingInfos[0])
		t.Logf("1: %+v", cd.RatingInfos[1])
		t.Logf("2: %+v", cd.RatingInfos[2])
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continuousRatingInfos())
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
	if len(cd.RatingInfos) != 4 || !cd.continuousRatingInfos() {
		t.Logf("0: %+v", cd.RatingInfos[0])
		t.Logf("1: %+v", cd.RatingInfos[1])
		t.Logf("2: %+v", cd.RatingInfos[2])
		t.Logf("3: %+v", cd.RatingInfos[3])
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continuousRatingInfos())
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
	if len(cd.RatingInfos) != 1 || !cd.continuousRatingInfos() {
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continuousRatingInfos())
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
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continuousRatingInfos())
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

func TestRatingProfileEqual(t *testing.T) {
	rpa := &RatingPlanActivation{
		ActivationTime: time.Date(2021, 7, 2, 4, 23, 3, 1234, time.UTC),
		RatingPlanId:   "test",
		FallbackKeys:   []string{"test"},
	}
	orpa := &RatingPlanActivation{
		ActivationTime: time.Date(2021, 7, 2, 4, 23, 3, 1234, time.UTC),
		RatingPlanId:   "test",
		FallbackKeys:   []string{"test"},
	}

	rcv := rpa.Equal(orpa)

	if rcv != true {
		t.Error(rcv)
	}
}

func TestRatingProfileSwap(t *testing.T) {
	rpa := &RatingPlanActivation{
		ActivationTime: time.Date(2021, 7, 2, 4, 23, 3, 1234, time.UTC),
		RatingPlanId:   "test",
		FallbackKeys:   []string{"test"},
	}
	orpa := &RatingPlanActivation{
		ActivationTime: time.Date(2021, 7, 2, 4, 23, 3, 1234, time.UTC),
		RatingPlanId:   "val1",
		FallbackKeys:   []string{"val1"},
	}
	rpas := RatingPlanActivations{rpa, orpa}

	rpas.Swap(0, 1)

	if !reflect.DeepEqual(rpas[0], orpa) {
		t.Error("didn't swap")
	}
}

func TestRatingProfileSwapInfos(t *testing.T) {
	ri := &RatingInfo{
		MatchedSubject: str,
		RatingPlanId:   str,
		MatchedPrefix:  str,
		MatchedDestId:  str,
		ActivationTime: time.Date(2021, 7, 2, 4, 23, 3, 1234, time.UTC),
		RateIntervals:  RateIntervalList{},
		FallbackKeys:   []string{str},
	}
	ri2 := &RatingInfo{
		MatchedSubject: "val1",
		RatingPlanId:   "val2",
		MatchedPrefix:  str,
		MatchedDestId:  str,
		ActivationTime: time.Date(2021, 7, 2, 4, 23, 3, 1234, time.UTC),
		RateIntervals:  RateIntervalList{},
		FallbackKeys:   []string{"val1"},
	}
	ris := RatingInfos{ri, ri2}

	ris.Swap(1, 0)

	if !reflect.DeepEqual(ris[1], ri) {
		t.Error("didn't swap")
	}
}

func TestRatingProfileLess(t *testing.T) {
	ri := &RatingInfo{
		MatchedSubject: str,
		RatingPlanId:   str,
		MatchedPrefix:  str,
		MatchedDestId:  str,
		ActivationTime: time.Date(2021, 7, 2, 4, 23, 3, 1234, time.UTC),
		RateIntervals:  RateIntervalList{},
		FallbackKeys:   []string{str},
	}
	ri2 := &RatingInfo{
		MatchedSubject: "val1",
		RatingPlanId:   "val2",
		MatchedPrefix:  str,
		MatchedDestId:  str,
		ActivationTime: time.Date(2021, 5, 2, 4, 23, 3, 1234, time.UTC),
		RateIntervals:  RateIntervalList{},
		FallbackKeys:   []string{"val1"},
	}
	ris := RatingInfos{ri, ri2}

	rcv := ris.Less(0, 1)

	if rcv != false {
		t.Error(rcv)
	}
}

func TestRatingProfileString(t *testing.T) {
	ri := &RatingInfo{
		MatchedSubject: str,
		RatingPlanId:   str,
		MatchedPrefix:  str,
		MatchedDestId:  str,
		ActivationTime: time.Date(2021, 7, 2, 4, 23, 3, 1234, time.UTC),
		RateIntervals:  RateIntervalList{},
		FallbackKeys:   []string{str},
	}
	ri2 := &RatingInfo{
		MatchedSubject: "val1",
		RatingPlanId:   "val2",
		MatchedPrefix:  str,
		MatchedDestId:  str,
		ActivationTime: time.Date(2021, 5, 2, 4, 23, 3, 1234, time.UTC),
		RateIntervals:  RateIntervalList{},
		FallbackKeys:   []string{"val1"},
	}
	ris := RatingInfos{ri, ri2}

	rcv := ris.String()
	b, _ := json.MarshalIndent(ris, "", " ")
	exp := string(b)

	if rcv != exp {
		t.Error(rcv)
	}
}
