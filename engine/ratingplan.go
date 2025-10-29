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
	"math"
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

// Clone returns a clone of RatingPlan
func (rp *RatingPlan) Clone() *RatingPlan {
	if rp == nil {
		return nil
	}
	result := &RatingPlan{
		Id: rp.Id,
	}
	if rp.Timings != nil {
		result.Timings = make(map[string]*RITiming, len(rp.Timings))
		for k, v := range rp.Timings {
			if v != nil {
				result.Timings[k] = v.Clone()
			}
		}
	}
	if rp.Ratings != nil {
		result.Ratings = make(map[string]*RIRate, len(rp.Ratings))
		for k, v := range rp.Ratings {
			if v != nil {
				result.Ratings[k] = v.Clone()
			}
		}
	}
	if rp.DestinationRates != nil {
		result.DestinationRates = make(map[string]RPRateList, len(rp.DestinationRates))
		for k, v := range rp.DestinationRates {
			result.DestinationRates[k] = v.Clone()
		}
	}
	return result
}

// CacheClone returns a clone of RatingPlan used by ltcache CacheCloner
func (rp *RatingPlan) CacheClone() any {
	return rp.Clone()
}

// RatingPlanWithOpts is used in replicatorV1 for dispatcher
type RatingPlanWithAPIOpts struct {
	*RatingPlan
	Tenant  string
	APIOpts map[string]any
}

type RPRate struct {
	Timing string
	Rating string
	Weight float64
}

// Clone returns a clone of RPRate
func (rpr *RPRate) Clone() *RPRate {
	if rpr == nil {
		return nil
	}
	return &RPRate{
		Timing: rpr.Timing,
		Rating: rpr.Rating,
		Weight: rpr.Weight,
	}
}

func (rpr *RPRate) Equal(orpr *RPRate) bool {
	return rpr.Timing == orpr.Timing && rpr.Rating == orpr.Rating && rpr.Weight == orpr.Weight
}

type RPRateList []*RPRate

// Clone returns a clone of RPRateList
func (rprl RPRateList) Clone() RPRateList {
	if rprl == nil {
		return nil
	}
	result := make(RPRateList, len(rprl))
	for i, rate := range rprl {
		result[i] = rate.Clone()
	}
	return result
}

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

// no sorter because it's sorted with RateIntervalTimeSorter

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

func (rp *RatingPlan) getFirstUnsaneRating() string {
	for _, rating := range rp.Ratings {
		rating.Rates.Sort()
		for i, rate := range rating.Rates {
			if i < (len(rating.Rates) - 1) {
				nextRate := rating.Rates[i+1]
				if nextRate.GroupIntervalStart <= rate.GroupIntervalStart {
					return rating.tag
				}
				if math.Mod(float64(nextRate.GroupIntervalStart.Nanoseconds()),
					float64(rate.RateIncrement.Nanoseconds())) != 0 {
					return rating.tag
				}
				if rate.RateUnit == 0 || rate.RateIncrement == 0 {
					return rating.tag
				}
			}
		}
	}
	return ""
}

func (rp *RatingPlan) getFirstUnsaneTiming() string {
	for _, timing := range rp.Timings {
		if (len(timing.Years) != 0 || len(timing.Months) != 0 || len(timing.MonthDays) != 0) &&
			len(timing.WeekDays) != 0 {
			return timing.tag
		}
	}
	return ""
}
