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
	"sort"
	"time"
)

type RatingInfo struct {
	MatchedSubject string
	RatingPlanId   string
	MatchedPrefix  string
	MatchedDestId  string
	ActivationTime time.Time
	RateIntervals  RateIntervalList
	FallbackKeys   []string
}

// SelectRatingIntevalsForTimespan orders rate intervals in time preserving only those which aply to the specified timestamp
func (ri RatingInfo) SelectRatingIntevalsForTimespan(ts *TimeSpan) (result RateIntervalList) {
	sorter := &RateIntervalTimeSorter{referenceTime: ts.TimeStart, ris: ri.RateIntervals}
	rateIntervals := sorter.Sort()
	// get the rating interval closest to begining of timespan
	var delta time.Duration = -1
	var bestRateIntervalIndex int
	var bestIntervalWeight float64
	for index, rateInterval := range rateIntervals {
		if !rateInterval.Contains(ts.TimeStart, false) {
			continue
		}
		if rateInterval.Weight < bestIntervalWeight {
			break // don't consider lower weights'
		}
		startTime := rateInterval.Timing.getLeftMargin(ts.TimeStart)
		tmpDelta := ts.TimeStart.Sub(startTime)
		if (startTime.Before(ts.TimeStart) ||
			startTime.Equal(ts.TimeStart)) &&
			(delta == -1 || tmpDelta < delta) {
			bestRateIntervalIndex = index
			bestIntervalWeight = rateInterval.Weight
			delta = tmpDelta
		}
	}
	result = append(result, rateIntervals[bestRateIntervalIndex])
	// check if later rating intervals influence this timespan
	//log.Print("RIS: ", utils.ToIJSON(rateIntervals))
	for i := bestRateIntervalIndex + 1; i < len(rateIntervals); i++ {
		if rateIntervals[i].Weight < bestIntervalWeight {
			break // don't consider lower weights'
		}
		startTime := rateIntervals[i].Timing.getLeftMargin(ts.TimeStart)
		if startTime.Before(ts.TimeEnd) {
			result = append(result, rateIntervals[i])
		}
	}
	return
}

type RatingInfos []*RatingInfo

func (ris RatingInfos) Len() int {
	return len(ris)
}

func (ris RatingInfos) Swap(i, j int) {
	ris[i], ris[j] = ris[j], ris[i]
}

func (ris RatingInfos) Less(i, j int) bool {
	return ris[i].ActivationTime.Before(ris[j].ActivationTime)
}

func (ris RatingInfos) Sort() {
	sort.Sort(ris)
}

func (ris RatingInfos) String() string {
	b, _ := json.MarshalIndent(ris, "", " ")
	return string(b)
}
