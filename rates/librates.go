/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package rates

/*
// orderRatesOnIntervals will order the rates based on ActivationInterval and intervalStart of each Rate
// there can be only one winning Rate for each interval, prioritized by the Weight
func orderRatesOnIntervals(aRts []*engine.Rate, sTime time.Time, isDuration bool) (ordRts []*engine.RateSInterval) {
	// index the received rates
	rtIdx := make(map[time.Time]map[time.Duration][]*engine.Rate) // map[ActivationTime]map[IntervalStart][]*engine.Rate
	for _, rt := range aRts {
		var rtATimesKey time.Time
		if rt.ActivationInterval != nil {
			rtATimesKey = rt.ActivationInterval.ActivationTime
		}
		if _, has := rtIdx[rtATimesKey]; !has {
			rtIdx[rtATimesKey] = make(map[time.Duration][]*engine.Rate)
		}
		rtIdx[rtATimesKey][rt.IntervalStart] = append(rtIdx[rtATimesKey][rt.IntervalStart], rt)
	}

	// sort the rates within the duration map
	for _, durMp := range rtIdx {
		for _, rts := range durMp {
			sort.Slice(rts, func(i, j int) bool {
				return rts[i].Weight > rts[j].Weight
			})
		}
	}

	// sort the IntervalStarts
	sortedStarts := make(map[time.Time][]time.Duration) // map[ActivationTime]
	for aTime, durMp := range rtIdx {
		sortedStarts[aTime] = make([]time.Duration, len(durMp))
		idxDur := 0
		for dur := range durMp {
			sortedStarts[aTime] = append(sortedStarts[aTime], dur)
			idxDur++
		}
		sort.Slice(sortedStarts[aTime], func(i, j int) bool {
			return sortedStarts[aTime][i] < sortedStarts[aTime][j]
		})
	}

	// sort the activation times
	sortedATimes := make([]time.Time, len(rtIdx))
	idxATimes := 0
	for aTime := range rtIdx {
		sortedATimes[idxATimes] = aTime
		idxATimes++
	}
	sort.Slice(sortedATimes, func(i, j int) bool {
		return sortedATimes[i].Before(sortedATimes[j])
	})

	return
}
*/
