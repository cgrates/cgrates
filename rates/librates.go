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

import (
	"sort"
	"time"

	"github.com/cgrates/cgrates/engine"
)

// orderRatesOnIntervals will order the rates based on ActivationInterval and intervalStart of each Rate
// there can be only one winning Rate for each interval, prioritized by the Weight
func orderRatesOnIntervals(aRts []*engine.Rate, sTime time.Time, usage time.Duration, isDuration bool) (ordRts []*engine.RateSInterval) {
	cronSTime := sTime.Add(time.Duration(-1 * time.Minute)) // cron min verbosity is minute
	endTime := sTime.Add(usage)                             // cover also the last unit used
	// index the received rates
	rtIdx := make(map[time.Time][]*engine.Rate) // map[ActivationTime][]*engine.Rate
	for _, rt := range aRts {
		nextRunTime := rt.NextActivationTime(cronSTime)
		if nextRunTime.After(endTime) {
			continue
		}
		rtIdx[nextRunTime] = append(rtIdx[nextRunTime], rt)
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
	// start with most recent activationTime lower or equal to sTime
	for i, aT := range sortedATimes {
		if !aT.After(sTime) || i == 0 {
			continue
		}
		sortedATimes = sortedATimes[i-1:]
	}
	// only for duration we will have multiple activationTimes
	if !isDuration {
		sortedATimes = sortedATimes[:1]
	}
	var usageSIdx, usageEIdx time.Duration
	for i, at := range sortedATimes {
		if sTime.Before(at) {
			usageSIdx = at.Sub(sTime)
		}
		if i != len(sortedATimes)-1 { // not the last one
			usageEIdx = sortedATimes[i+1].Sub(sTime)
		} else {
			usageEIdx = usage
		}
		// Sort the rates based on their Weight
		sort.Slice(rtIdx[at], func(i, j int) bool {
			return rtIdx[at][i].Weight < rtIdx[at][j].Weight
		})
		var rtIcmts []*engine.RateSIncrement
		for _, rt := range rtIdx[at] {

			for j, ivlRt := range rt.IntervalRates {
				if ivlRt.IntervalStart > usageEIdx {
					break
				}
				if j != len(rt.IntervalRates)-1 &&
					rt.IntervalRates[j+1].IntervalStart <= usageSIdx { // not the last one
					// the next intervalStat is still good for usageSIdx, no need of adding the current one
					continue
				}
				// ready to add the increment
				rtIcmts = append(rtIcmts,
					&engine.RateSIncrement{Rate: rt, IntervalRateIndex: j})
			}
		}
		ordRts = append(ordRts, &engine.RateSInterval{Increments: rtIcmts})
	}
	return
}
