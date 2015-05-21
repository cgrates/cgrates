/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"math"

	"github.com/cgrates/cgrates/history"
)

/*
The struture that is saved to storage.
*/
type RatingPlan struct {
	Id               string
	Timings          map[string]*RITiming
	Ratings          map[string]*RIRate
	DestinationRates map[string]RPRateList
}

type RPRate struct {
	Timing string
	Rating string
	Weight float64
}

func (rpr *RPRate) Equal(orpr *RPRate) bool {
	return rpr.Timing == orpr.Timing && rpr.Rating == orpr.Rating && rpr.Weight == orpr.Weight
}

type RPRateList []*RPRate

func (rp *RatingPlan) RateIntervalList(dId string) RateIntervalList {
	ril := make(RateIntervalList, len(rp.DestinationRates[dId]))
	for i, rpr := range rp.DestinationRates[dId] {
		ril[i] = &RateInterval{
			Timing: rp.Timings[rpr.Timing],
			Rating: rp.Ratings[rpr.Rating],
			Weight: rpr.Weight,
		}
	}
	return ril
}

/*
type xCachedRatingPlan struct {
	rp *RatingPlan
	*cache2go.XEntry
}
*/

/*
Adds one ore more intervals to the internal interval list only if it is not allready in the list.
*/
func (rp *RatingPlan) AddRateInterval(dId string, ris ...*RateInterval) {
	if rp.DestinationRates == nil {
		rp.Timings = make(map[string]*RITiming)
		rp.Ratings = make(map[string]*RIRate)
		rp.DestinationRates = make(map[string]RPRateList, 1)
	}
	for _, ri := range ris {
		rpr := &RPRate{Weight: ri.Weight}
		if ri.Timing != nil {
			timingTag := ri.Timing.Stringify()
			rp.Timings[timingTag] = ri.Timing
			rpr.Timing = timingTag
		}
		if ri.Rating != nil {
			ratingTag := ri.Rating.Stringify()
			rp.Ratings[ratingTag] = ri.Rating
			rpr.Rating = ratingTag
		}
		found := false
		for _, erpr := range rp.DestinationRates[dId] {
			if erpr.Equal(rpr) {
				found = true
				break
			}
		}
		if !found {
			rp.DestinationRates[dId] = append(rp.DestinationRates[dId], rpr)
		}
	}
}

func (rp *RatingPlan) Equal(o *RatingPlan) bool {
	return rp.Id == o.Id
}

// history record method
func (rp *RatingPlan) GetHistoryRecord() history.Record {
	js, _ := json.Marshal(rp)
	return history.Record{
		Id:       rp.Id,
		Filename: history.RATING_PLANS_FN,
		Payload:  js,
	}
}

// IsValid determines if the rating plan covers a continous period of time
func (rp *RatingPlan) isContinous() bool {
	weekdays := make([]int, 7)
	for _, tm := range rp.Timings {
		// if it is a blank timing than it will match all
		if tm.IsBlank() {
			return true
		}
		// skip the special timings (for specific dates)
		if len(tm.Years) != 0 || len(tm.Months) != 0 || len(tm.MonthDays) != 0 {
			continue
		}
		// if the startime is not midnight than is an extra time
		if tm.StartTime != "00:00:00" {
			continue
		}
		//check if all weekdays are covered
		for _, wd := range tm.WeekDays {
			weekdays[wd] = 1
		}
		allWeekdaysCovered := true
		for _, wd := range weekdays {
			if wd != 1 {
				allWeekdaysCovered = false
				break
			}
		}
		if allWeekdaysCovered {
			return true
		}
	}
	return false
}

func (rp *RatingPlan) areRatesSane() bool {
	for _, rating := range rp.Ratings {
		rating.Rates.Sort()
		for i, rate := range rating.Rates {
			if i < (len(rating.Rates) - 1) {
				nextRate := rating.Rates[i+1]
				if nextRate.GroupIntervalStart <= rate.GroupIntervalStart {
					return false
				}
				if math.Mod(nextRate.GroupIntervalStart.Seconds(), rate.RateIncrement.Seconds()) != 0 {
					return false
				}
			}
		}
	}
	return true
}

func (rp *RatingPlan) areTimingsSane() bool {
	for _, timing := range rp.Timings {
		if (len(timing.Years) != 0 || len(timing.Months) != 0 || len(timing.MonthDays) != 0) &&
			len(timing.WeekDays) != 0 {
			return false
		}
	}
	return true
}
