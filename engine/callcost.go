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
	"errors"
	"reflect"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// The output structure that will be returned with the call cost information.
type CallCost struct {
	Direction, Category, Tenant, Subject, Account, Destination, TOR string
	Cost                                                            float64
	Timespans                                                       TimeSpans
	deductConnectFee                                                bool
	maxCostDisconect                                                bool
}

// Merges the received timespan if they are similar (same activation period, same interval, same minute info.
func (cc *CallCost) Merge(other *CallCost) {
	if len(cc.Timespans)-1 < 0 || len(other.Timespans) == 0 {
		return
	}
	ts := cc.Timespans[len(cc.Timespans)-1]
	otherTs := other.Timespans[0]
	if reflect.DeepEqual(ts.ratingInfo, otherTs.ratingInfo) &&
		reflect.DeepEqual(ts.RateInterval, otherTs.RateInterval) {
		// extend the last timespan with
		ts.TimeEnd = ts.TimeEnd.Add(otherTs.GetDuration())
		// add the rest of the timspans
		cc.Timespans = append(cc.Timespans, other.Timespans[1:]...)
	} else {
		// just add all timespans
		cc.Timespans = append(cc.Timespans, other.Timespans...)
	}
	cc.Cost += other.Cost
}

func (cc *CallCost) GetStartTime() time.Time {
	if len(cc.Timespans) == 0 {
		return time.Now()
	}
	return cc.Timespans[0].TimeStart
}

func (cc *CallCost) GetEndTime() time.Time {
	if len(cc.Timespans) == 0 {
		return time.Now()
	}
	return cc.Timespans[len(cc.Timespans)-1].TimeEnd
}

func (cc *CallCost) GetDuration() (td time.Duration) {
	for _, ts := range cc.Timespans {
		td += ts.GetDuration()
	}
	return
}

func (cc *CallCost) GetConnectFee() float64 {
	if len(cc.Timespans) == 0 ||
		cc.Timespans[0].RateInterval == nil ||
		cc.Timespans[0].RateInterval.Rating == nil {
		return 0
	}
	return cc.Timespans[0].RateInterval.Rating.ConnectFee
}

// Creates a CallDescriptor structure copying related data from CallCost
func (cc *CallCost) CreateCallDescriptor() *CallDescriptor {
	return &CallDescriptor{
		Direction:   cc.Direction,
		Category:    cc.Category,
		Tenant:      cc.Tenant,
		Subject:     cc.Subject,
		Account:     cc.Account,
		Destination: cc.Destination,
	}
}

func (cc *CallCost) IsPaid() bool {
	for _, ts := range cc.Timespans {
		if paid, _ := ts.IsPaid(); !paid {
			return false
		}
	}
	return true
}

func (cc *CallCost) ToDataCost() (*DataCost, error) {
	if cc.TOR == utils.VOICE {
		return nil, errors.New("Not a data call!")
	}
	dc := &DataCost{
		Direction:        cc.Direction,
		Category:         cc.Category,
		Tenant:           cc.Tenant,
		Subject:          cc.Subject,
		Account:          cc.Account,
		Destination:      cc.Destination,
		TOR:              cc.TOR,
		Cost:             cc.Cost,
		deductConnectFee: cc.deductConnectFee,
	}
	dc.DataSpans = make([]*DataSpan, len(cc.Timespans))
	for i, ts := range cc.Timespans {
		length := ts.TimeEnd.Sub(ts.TimeStart).Seconds()
		callDuration := ts.DurationIndex.Seconds()
		dc.DataSpans[i] = &DataSpan{
			DataStart:      callDuration - length,
			DataEnd:        callDuration,
			Cost:           ts.Cost,
			ratingInfo:     ts.ratingInfo,
			RateInterval:   ts.RateInterval,
			DataIndex:      callDuration,
			MatchedSubject: ts.MatchedSubject,
			MatchedPrefix:  ts.MatchedPrefix,
			MatchedDestId:  ts.MatchedDestId,
			RatingPlanId:   ts.RatingPlanId,
		}
		dc.DataSpans[i].Increments = make([]*DataIncrement, len(ts.Increments))
		for j, incr := range ts.Increments {
			dc.DataSpans[i].Increments[j] = &DataIncrement{
				Amount:              incr.Duration.Seconds(),
				Cost:                incr.Cost,
				BalanceInfo:         incr.BalanceInfo,
				BalanceRateInterval: incr.BalanceRateInterval,
				UnitInfo:            incr.UnitInfo,
				CompressFactor:      incr.CompressFactor,
				paid:                incr.paid,
			}
		}
	}
	return dc, nil
}

func (cc *CallCost) GetLongestRounding() (roundingDecimals int, roundingMethod string) {
	for _, ts := range cc.Timespans {
		if ts.RateInterval != nil && ts.RateInterval.Rating.RoundingDecimals > roundingDecimals { //ToDo: When will ts.RateInterval be empty?
			roundingDecimals = ts.RateInterval.Rating.RoundingDecimals
			roundingMethod = ts.RateInterval.Rating.RoundingMethod
		}
	}
	return
}
