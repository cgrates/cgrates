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
	"github.com/cgrates/cgrates/utils"
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

// rateWithTimes activates a rate on an interval
type rateWithTimes struct {
	uId string
	rt  *engine.Rate
	aTime,
	iTime time.Time
}

// id is used to provide an unique identifier for a rateWithTimes
func (rWt *rateWithTimes) id() string {
	if rWt.uId == "" {
		rWt.uId = fmt.Sprintf("%s_%d", rWt.rt.ID, rWt.aTime.Unix())
	}
	return rWt.uId
}

type orderedRate struct {
	time.Duration
	*engine.Rate
}

// orderRatesOnIntervals will order the rates based on ActivationInterval and intervalStart of each Rate
// there can be only one winning Rate for each interval, prioritized by the Weight
func orderRatesOnIntervals(aRts []*engine.Rate, sTime time.Time, usage time.Duration,
	isDuration bool, verbosity int) (ordRts []*orderedRate, err error) {

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
				(!rIt.iTime.IsZero() && !tm.Before(rIt.iTime)) {
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

	// compute the list of returned rates together with their index interval
	if isDuration {
		// add all the possible ActivationTimes from cron expressions
		var usageIndx time.Duration // the difference between setup and activation time of the rate
		for _, aTime := range sortedATimes {
			if !endTime.After(aTime) {
				break // we are not interested about further rates
			}
			wnr := rtIdx[aTime].winner()
			if wnr == nil {
				continue
			}
			if sTime.Before(aTime) {
				usageIndx = aTime.Sub(sTime)
			}
			if len(ordRts) == 0 || wnr.rt.ID != ordRts[len(ordRts)-1].Rate.ID { // only add the winner if not already active
				ordRts = append(ordRts, &orderedRate{usageIndx, rtIdx[aTime].winner().rt})
			}
		}
	} else { // only first rate is considered for units
		ordRts = []*orderedRate{{time.Duration(0), rtIdx[sortedATimes[0]].winner().rt}}
	}
	return
}

// computeRateSIntervals will give out the cost projection for the given orderedRates and usage
func computeRateSIntervals(rts []*orderedRate, intervalStart, usage time.Duration) (rtIvls []*engine.RateSInterval, err error) {
	totalUsage := usage
	if intervalStart != 0 {
		totalUsage = usage + intervalStart
	}
	for i, rt := range rts {
		isLastRt := i == len(rts)-1
		var rtUsageEIdx time.Duration
		if !isLastRt {
			rtUsageEIdx = rts[i+1].Duration
		} else {
			rtUsageEIdx = totalUsage
		}
		var rIcmts []*engine.RateSIncrement
		iRtUsageSIdx := intervalStart
		iRtUsageEIdx := rtUsageEIdx
		for j, iRt := range rt.IntervalRates {
			//fmt.Printf("ivalStart: %v, ivalEnd: %v, rtID: %s, increment idx: %d, iRtUsageSIdx: %+v, iRtUsageEIdx: %+v, iRt: %s\n",
			//	intervalStart, rtUsageEIdx, rt.ID, j, iRtUsageSIdx, iRtUsageEIdx, utils.ToIJSON(iRt))
			if iRtUsageSIdx >= rtUsageEIdx { // charged enough for interval
				break
			}
			// make sure we bill from start
			if iRt.IntervalStart > iRtUsageSIdx {
				if j == 0 {
					return nil, fmt.Errorf("intervalStart for rate: <%s> higher than usage: %v",
						rt.UID(), iRtUsageSIdx)
				}
				break // we are pass the start
			}
			isLastIRt := j == len(rt.IntervalRates)-1
			if !isLastIRt && rt.IntervalRates[j+1].IntervalStart <= iRtUsageSIdx {
				continue // the next interval changes the rating
			}
			if isLastIRt {
				iRtUsageEIdx = rtUsageEIdx
			} else if rt.IntervalRates[j+1].IntervalStart > rtUsageEIdx {
				iRtUsageEIdx = rtUsageEIdx
			} else {
				iRtUsageEIdx = rt.IntervalRates[j+1].IntervalStart
			}
			if iRtUsageEIdx == time.Duration(0) {
				return nil, fmt.Errorf("zero usage to be charged with rate: <%s>", rt.UID())
			}
			if iRt.Increment == time.Duration(0) {
				return nil, fmt.Errorf("zero increment to be charged within rate: <%s>", rt.UID())
			}
			if rt.IntervalRates[j].FixedFee != 0 { // Add FixedFee
				rIcmts = append(rIcmts, &engine.RateSIncrement{
					UsageStart:        iRtUsageSIdx,
					Rate:              rt.Rate,
					IntervalRateIndex: j,
					CompressFactor:    1,
					Usage:             utils.InvalidDuration,
				})
			}
			iRtUsage := iRtUsageEIdx - iRtUsageSIdx
			intUsage := int64(iRtUsage)
			intIncrm := int64(iRt.Increment)
			cmpFactor := intUsage / intIncrm
			if intUsage%intIncrm != 0 {
				cmpFactor++ // int division has used math.Floor, need Ceil
			}
			rIcrm := &engine.RateSIncrement{
				UsageStart:        iRtUsageSIdx,
				Rate:              rt.Rate,
				IntervalRateIndex: j,
				CompressFactor:    cmpFactor,
				Usage:             iRtUsage,
			}
			if len(rIcmts) != 0 && rIcrm.CompressEquals(rIcmts[len(rIcmts)-1], false) {
				rIcmts[len(rIcmts)-1].CompressFactor += rIcrm.CompressFactor
			} else {
				rIcmts = append(rIcmts, rIcrm)
			}
			iRtUsageSIdx += iRtUsage

		}
		usageStart := intervalStart
		intervalStart = rtUsageEIdx // continue for the next interval
		if len(rIcmts) == 0 {       // no match found
			continue
		}
		rIvl := &engine.RateSInterval{
			UsageStart:     usageStart,
			Increments:     rIcmts,
			CompressFactor: 1}
		if len(rtIvls) != 0 && rIvl.CompressEquals(rtIvls[len(rtIvls)-1]) {
			rtIvls[len(rtIvls)-1].CompressFactor += rIvl.CompressFactor
		} else {
			rtIvls = append(rtIvls, rIvl)
		}
		if iRtUsageSIdx >= totalUsage { // charged enough for the usage
			break
		}

	}
	return
}
