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

package engine

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

func (rp *RatingPlan) RateIntervalList(dId string) (ril RateIntervalList) {
	for _, rpr := range rp.DestinationRates[dId] {
		ril = append(ril, &RateInterval{
			Timing: rp.Timings[rpr.Timing],
			Rating: rp.Ratings[rpr.Rating],
			Weight: rpr.Weight,
		})
	}
	return
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
