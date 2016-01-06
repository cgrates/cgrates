/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

This program is free software: you can Storagetribute it and/or modify
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
		Direction:   utils.OUT,
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
		Direction:   utils.OUT,
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
		Direction:   utils.OUT,
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
		Direction:   utils.OUT,
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

func TestRatingProfileRIforTSThre(t *testing.T) {
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
