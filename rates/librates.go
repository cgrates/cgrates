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
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/engine"
)

func newRatesWithWinner(rIt *rateWithTimes) *ratesWithWinner {
	return &ratesWithWinner{
		rts: map[string]*rateWithTimes{
			rIt.id(): rIt,
		},
		wnr: rIt,
	}
}

func initRatesWithWinner() *ratesWithWinner {
	return &ratesWithWinner{
		rts: make(map[string]*rateWithTimes),
	}
}

// ratesWithWinner computes always the winner based on highest Weight
type ratesWithWinner struct {
	rts map[string]*rateWithTimes
	wnr *rateWithTimes
}

//add will add the rate to the rates
func (rs *ratesWithWinner) add(rWt *rateWithTimes) {
	rs.rts[rWt.id()] = rWt
	if rs.wnr == nil || rs.wnr.rt.Weight < rWt.rt.Weight {
		rs.wnr = rWt
	}
}

// winner returns the rate with the highest Weight
func (rs *ratesWithWinner) winner() *rateWithTimes {
	return rs.wnr
}

// has tests if the rateID is present in rates
func (rs *ratesWithWinner) has(rtID string) (has bool) {
	_, has = rs.rts[rtID]
	return
}

// rateWithTimes keeps the aTime and iTime attached to the rate
type rateWithTimes struct {
	uId string
	rt  *engine.Rate
	aTime,
	iTime time.Time
}

func (rWt *rateWithTimes) id() string {
	if rWt.uId == "" {
		rWt.uId = fmt.Sprintf("%s_%d", rWt.rt.ID, rWt.aTime.Unix())
	}
	return rWt.uId
}

// orderRatesOnIntervals will order the rates based on ActivationInterval and intervalStart of each Rate
// there can be only one winning Rate for each interval, prioritized by the Weight
func orderRatesOnIntervals(aRts []*engine.Rate, sTime time.Time, usage time.Duration,
	isDuration bool, verbosity int) (ordRts []*engine.RateSInterval, err error) {

	endTime := sTime.Add(usage)

	// index the received rates based on unique times they run
	rtIdx := make(map[time.Time]*ratesWithWinner) // map[ActivationTimes]*ratesWithWinner
	allRates := make(map[string]*rateWithTimes)
	for _, rt := range aRts {
		var rTimes [][]time.Time
		if rTimes, err = rt.RunTimes(sTime, endTime, verbosity); err != nil {
			return
		}
		for _, rTimeSet := range rTimes {
			rIt := &rateWithTimes{
				rt:    rt,
				aTime: rTimeSet[0],
				iTime: rTimeSet[1],
			}
			allRates[rIt.id()] = rIt
			if _, hasKey := rtIdx[rTimeSet[0]]; !hasKey {
				rtIdx[rTimeSet[0]] = initRatesWithWinner()
			}
			rtIdx[rTimeSet[0]].add(rIt)
			if rTimeSet[1].IsZero() { // the rate will always be active
				continue
			}
			if _, hasKey := rtIdx[rTimeSet[1]]; !hasKey {
				rtIdx[rTimeSet[1]] = initRatesWithWinner()
			}
		}
	}
	// add the active rates to all time samples
	for tm, rWw := range rtIdx {
		for _, rIt := range allRates {
			if rWw.has(rIt.id()) ||
				rIt.aTime.After(tm) ||
				(!rIt.iTime.IsZero() && !rIt.iTime.Before(tm)) {
				continue
			}
			rWw.add(rIt)
		}
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
		break
	}
	// get the real list of interesting ordered Rates based on their activationTime
	var oRts []*rateWithTimes
	if isDuration {
		// add all the possible ActivationTimes from cron expressions
		for _, aTime := range sortedATimes {

			if !endTime.After(aTime) {
				break // we are not interested about further rates
			}
			wnr := rtIdx[aTime].winner()
			if wnr == nil {
				continue
			}
			if len(oRts) == 0 || wnr.rt.ID != oRts[len(oRts)-1].rt.ID { // only add the winner if not already active
				oRts = append(oRts, wnr)
			}
		}
	} else { // only first rate is considered for units
		oRts = []*rateWithTimes{rtIdx[sortedATimes[0]].winner()}
	}
	// add the Intervals and Increments
	var usageSIdx, usageEIdx time.Duration
	for i, rWt := range oRts {
		if sTime.Before(rWt.aTime) {
			usageSIdx = rWt.aTime.Sub(sTime)
		}
		if i != len(sortedATimes)-1 { // not the last one
			usageEIdx = sortedATimes[i+1].Sub(sTime)
		} else {
			usageEIdx = usage
		}
		// append the valid increments
		var rtIcmts []*engine.RateSIncrement
		rt := rtIdx[rWt.aTime].winner().rt
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
