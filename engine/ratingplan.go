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
	DestinationRates map[string]RateIntervalList
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
		rp.DestinationRates = make(map[string]RateIntervalList, 1)
	}
	for _, ri := range ris {
		found := false
		for _, eri := range rp.DestinationRates[dId] {
			if ri.Equal(eri) {
				found = true
				break
			}
		}
		if !found {
			rp.DestinationRates[dId] = append(rp.DestinationRates[dId], ri)
		}
	}
}

func (rp *RatingPlan) Equal(o *RatingPlan) bool {
	return rp.Id == o.Id
}
