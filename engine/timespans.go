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

import (
	"fmt"
	"time"
)

/*
A unit in which a call will be split that has a specific price related interval attached to it.
*/
type TimeSpan struct {
	TimeStart, TimeEnd time.Time
	Cost               float64
	RatingPlan         *RatingPlan
	RateInterval       *RateInterval
	MinuteInfo         *MinuteInfo
	CallDuration       time.Duration // the call duration so far till TimeEnd
	overlapped         bool          // mark a timespan as overlapped by an expanded one
}

// Holds the bonus minute information related to a specified timespan
type MinuteInfo struct {
	DestinationId string
	Quantity      float64
	Price         float64
}

// Returns the duration of the timespan
func (ts *TimeSpan) GetDuration() time.Duration {
	return ts.TimeEnd.Sub(ts.TimeStart)
}

// Returns the cost of the timespan according to the relevant cost interval.
// It also sets the Cost field of this timespan (used for refound on session
// manager debit loop where the cost cannot be recalculated)
func (ts *TimeSpan) getCost(cd *CallDescriptor) (cost float64) {
	if ts.MinuteInfo != nil {
		return ts.GetDuration().Seconds() * ts.MinuteInfo.Price
	}
	if ts.RateInterval == nil {
		return 0
	}
	i := ts.RateInterval
	cost = i.GetCost(ts.GetDuration(), ts.GetGroupStart())
	ts.Cost = cost
	return
}

/*
Returns true if the given time is inside timespan range.
*/
func (ts *TimeSpan) Contains(t time.Time) bool {
	return t.After(ts.TimeStart) && t.Before(ts.TimeEnd)
}

/*
Will set the interval as spans's interval if new Weight is lower then span's interval Weight
or if the Weights are equal and new price is lower then spans's interval price
*/
func (ts *TimeSpan) SetRateInterval(i *RateInterval) {
	if ts.RateInterval == nil || ts.RateInterval.Weight < i.Weight {
		ts.RateInterval = i
		return
	}
	iPrice, _, _ := i.GetRateParameters(ts.GetGroupStart())
	tsPrice, _, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
	if ts.RateInterval.Weight == i.Weight && iPrice < tsPrice {
		ts.RateInterval = i
	}
}

/*
Splits the given timespan according to how it relates to the interval.
It will modify the endtime of the received timespan and it will return
a new timespan starting from the end of the received one.
The interval will attach itself to the timespan that overlaps the interval.
*/
func (ts *TimeSpan) SplitByRateInterval(i *RateInterval) (nts *TimeSpan) {

	//Logger.Debug("here: ", ts, " +++ ", i)
	// if the span is not in interval return nil
	if !(i.Contains(ts.TimeStart) || i.Contains(ts.TimeEnd)) {
		//Logger.Debug("Not in interval")
		return
	}
	// split by GroupStart
	i.Rates.Sort()
	for _, price := range i.Rates {
		if ts.GetGroupStart() < price.GroupIntervalStart && ts.GetGroupEnd() >= price.GroupIntervalStart {
			ts.SetRateInterval(i)
			splitTime := ts.TimeStart.Add(price.GroupIntervalStart - ts.GetGroupStart())
			nts = &TimeSpan{TimeStart: splitTime, TimeEnd: ts.TimeEnd}
			ts.TimeEnd = splitTime
			nts.SetRateInterval(i)
			nts.CallDuration = ts.CallDuration
			ts.SetNewCallDuration(nts)

			return
		}
	}

	// if the span is enclosed in the interval try to set as new interval and return nil
	if i.Contains(ts.TimeStart) && i.Contains(ts.TimeEnd) {
		//Logger.Debug("All in interval")
		ts.SetRateInterval(i)
		return
	}
	// if only the start time is in the interval split the interval to the right
	if i.Contains(ts.TimeStart) {
		//Logger.Debug("Start in interval")
		splitTime := i.getRightMargin(ts.TimeStart)
		ts.SetRateInterval(i)
		if splitTime == ts.TimeStart {
			return
		}
		nts = &TimeSpan{TimeStart: splitTime, TimeEnd: ts.TimeEnd}
		ts.TimeEnd = splitTime
		nts.CallDuration = ts.CallDuration
		ts.SetNewCallDuration(nts)

		return
	}
	// if only the end time is in the interval split the interval to the left
	if i.Contains(ts.TimeEnd) {
		//Logger.Debug("End in interval")
		splitTime := i.getLeftMargin(ts.TimeEnd)
		if splitTime == ts.TimeEnd {
			return
		}
		nts = &TimeSpan{TimeStart: splitTime, TimeEnd: ts.TimeEnd}
		ts.TimeEnd = splitTime

		nts.SetRateInterval(i)
		nts.CallDuration = ts.CallDuration
		ts.SetNewCallDuration(nts)

		return
	}
	return
}

/*
Splits the given timespan on activation period's activation time.
*/
func (ts *TimeSpan) SplitByRatingPlan(ap *RatingPlan) (newTs *TimeSpan) {
	if !ts.Contains(ap.ActivationTime) {
		return nil
	}
	newTs = &TimeSpan{TimeStart: ap.ActivationTime, TimeEnd: ts.TimeEnd, RatingPlan: ap}
	newTs.CallDuration = ts.CallDuration
	ts.TimeEnd = ap.ActivationTime
	ts.SetNewCallDuration(newTs)
	return
}

/*
Splits the given timespan on minute bucket's duration.
*/
func (ts *TimeSpan) SplitByMinuteBalance(mb *Balance) (newTs *TimeSpan) {
	// if mb expired skip it
	if !mb.ExpirationDate.IsZero() && (ts.TimeStart.Equal(mb.ExpirationDate) || ts.TimeStart.After(mb.ExpirationDate)) {
		return nil
	}

	// expiring before time spans end

	if !mb.ExpirationDate.IsZero() && ts.TimeEnd.After(mb.ExpirationDate) {
		newTs = &TimeSpan{TimeStart: mb.ExpirationDate, TimeEnd: ts.TimeEnd}
		newTs.CallDuration = ts.CallDuration
		ts.TimeEnd = mb.ExpirationDate
		ts.SetNewCallDuration(newTs)
	}

	s := ts.GetDuration().Seconds()
	ts.MinuteInfo = &MinuteInfo{mb.DestinationId, s, mb.SpecialPrice}
	if s <= mb.Value {
		mb.Value -= s
		return newTs
	}
	secDuration, _ := time.ParseDuration(fmt.Sprintf("%vs", mb.Value))

	newTimeEnd := ts.TimeStart.Add(secDuration)
	newTs = &TimeSpan{TimeStart: newTimeEnd, TimeEnd: ts.TimeEnd}
	ts.TimeEnd = newTimeEnd
	newTs.CallDuration = ts.CallDuration
	ts.MinuteInfo.Quantity = mb.Value
	ts.SetNewCallDuration(newTs)
	mb.Value = 0

	return
}

// Returns the starting time of this timespan
func (ts *TimeSpan) GetGroupStart() time.Duration {
	s := ts.CallDuration - ts.GetDuration()
	if s < 0 {
		s = 0
	}
	return s
}

func (ts *TimeSpan) GetGroupEnd() time.Duration {
	return ts.CallDuration
}

// sets the CallDuration attribute to reflect new timespan
func (ts *TimeSpan) SetNewCallDuration(nts *TimeSpan) {
	d := ts.CallDuration - nts.GetDuration()
	if d < 0 {
		d = 0
	}
	ts.CallDuration = d
}
