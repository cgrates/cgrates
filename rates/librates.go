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

func newRatesWithWinner(rt *engine.Rate) (rts *ratesWithWinner) {
	rts = &ratesWithWinner{
		rts: make(map[string]*engine.Rate),
	}
	rts.add(rt)
	return
}

// ratesWithWinner computes always the winner based on highest Weight
type ratesWithWinner struct {
	rts map[string]*engine.Rate
	rt  *engine.Rate
}

//add will add the rate to the rates
func (rs *ratesWithWinner) add(rt *engine.Rate) {
	rs.rts[rt.ID] = rt
	if rs.rt == nil || rs.rt.Weight < rt.Weight {
		rs.rt = rt
	}
}

// winner returns the rate with the highest Weight
func (rs *ratesWithWinner) winner() *engine.Rate {
	return rs.rt
}

// has tests if the rateID is present in rates
func (rs *ratesWithWinner) has(rtID string) (has bool) {
	_, has = rs.rts[rtID]
	return
}

// orderRatesOnIntervals will order the rates based on ActivationInterval and intervalStart of each Rate
// there can be only one winning Rate for each interval, prioritized by the Weight
func orderRatesOnIntervals(aRts []*engine.Rate, sTime time.Time, usage time.Duration, isDuration bool, aTimeVerbosity int) (ordRts []*engine.RateSInterval) {
	cronSTime := sTime.Add(time.Duration(-1 * time.Minute)) // cron min verbosity is minute
	endTime := sTime.Add(usage)                             // cover also the last unit used
	// index the received rates
	rtIdx := make(map[time.Time]*ratesWithWinner) // map[ActivationTimes]*ratesWithWinner
	for _, rt := range aRts {
		nextRunTime := rt.NextActivationTime(cronSTime)
		if nextRunTime.After(endTime) {
			continue
		}
		if _, hasKey := rtIdx[nextRunTime]; !hasKey {
			rtIdx[nextRunTime] = newRatesWithWinner(rt)
			continue
		}
		rtIdx[nextRunTime].add(rt)
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
	// finalize the sortedATimes
	if isDuration {
		// add all the possible ActivationTimes from cron expressions
		for i := 0; i < aTimeVerbosity; i++ {
			//fmt.Printf("aTimeVerbosity run: %d, sortedATimes: %+v\n", i, sortedATimes)
			var altered bool
			for _, aTime := range sortedATimes {
				//fmt.Printf("checking aTime: %+v\n", aTime)
				// each of the rates will be checked against activationTime
				for _, rt := range aRts {
					nextRunTime := rt.NextActivationTime(aTime)
					//fmt.Printf("rate: %s, rt.ActivationStart: %s, on aTime: %+v nextRunTime: %+v\n", rt.ID, rt.ActivationStart, aTime, nextRunTime)
					if nextRunTime.After(endTime) {
						//fmt.Println("after endTime")
						continue
					}
					if rtMp, hasRunTime := rtIdx[nextRunTime]; hasRunTime {
						if rtMp.has(rt.ID) { // rate was already captured
							//fmt.Printf("already captured!\n")
							continue
						}
					}
					// try to see if it fits in the activation time list
					for sTimeIdx, sATime := range sortedATimes {
						//fmt.Printf("sTimeIdx: %d, sAtime: %+v\n", sTimeIdx, sATime)
						if sTimeIdx == 0 { // ignore the first one
							//fmt.Println("first interval")
							continue
						}
						if nextRunTime.After(sATime) {
							if sTimeIdx != len(sortedATimes)-1 {
								//fmt.Println("not the last one and higher")
								continue
							}
							// last one and higher
							if rtIdx[sortedATimes[sTimeIdx]].winner().ID == rt.ID { // activated in last slot
								//fmt.Println("activated in last slot")
								continue
							}
						}
						if rtIdx[sortedATimes[sTimeIdx-1]].has(rt.ID) { // activated in previous slot
							//fmt.Println("activated in previous slot")
							continue
						} else {
							//fmt.Printf("rtIdx: %+v,  has no rt with id: %s\n", rtIdx, rt.ID)
						}
						// Anything passing here should be a winner
						if nextRunTime == sATime {
							//fmt.Println("already in time, adding to rates")
							rtIdx[nextRunTime].add(rt)
							altered = true
							break
						}
						if _, has := rtIdx[nextRunTime]; !has {
							rtIdx[nextRunTime] = newRatesWithWinner(rt)
						} else {
							rtIdx[nextRunTime].add(rt)
						}
						if sTimeIdx == len(sortedATimes)-1 { // last index, higher than the last ActivationTimes
							//fmt.Println("adding as last")
							sortedATimes = append(sortedATimes, nextRunTime)
						} else { // Insert before current index
							//fmt.Printf("insert in betwee, have before: %+v", sortedATimes)
							sortedATimes = append(sortedATimes, time.Time{})
							copy(sortedATimes[sTimeIdx+1:], sortedATimes[sTimeIdx:])
							sortedATimes[sTimeIdx] = nextRunTime
							//fmt.Printf(" have after: %+v\n", sortedATimes)
						}
						altered = true
						break
					}

				}
				if altered { // start again from first aTime
					break
				}

			}
			if !altered { // no change was done in a complete iteration, no need of further lookups
				break
			}
		}

	} else { // only for duration we will have multiple activationTimes
		sortedATimes = sortedATimes[:1] // only first ordered activationTime is considered for units
	}
	//fmt.Printf("sortedTimes: %+v\n", sortedATimes)
	// add the Intervals and Increments
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
		// append the valid increments
		var rtIcmts []*engine.RateSIncrement
		rt := rtIdx[at].winner()
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
		ordRts = append(ordRts, &engine.RateSInterval{Increments: rtIcmts})
	}
	return
}
