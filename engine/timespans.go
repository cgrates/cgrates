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
	//"fmt"

	"reflect"
	"time"

	"github.com/cgrates/cgrates/utils"
)

/*
A unit in which a call will be split that has a specific price related interval attached to it.
*/
type TimeSpan struct {
	TimeStart, TimeEnd                                         time.Time
	Cost                                                       float64
	ratingInfo                                                 *RatingInfo
	RateInterval                                               *RateInterval
	DurationIndex                                              time.Duration // the call duration so far till TimeEnd
	Increments                                                 Increments
	MatchedSubject, MatchedPrefix, MatchedDestId, RatingPlanId string
}

type Increment struct {
	Duration            time.Duration
	Cost                float64
	BalanceInfo         *BalanceInfo // need more than one for units with cost
	BalanceRateInterval *RateInterval
	UnitInfo            *UnitInfo
	CompressFactor      int
	paid                bool
}

// Holds the minute information related to a specified timespan
type UnitInfo struct {
	DestinationId string
	Quantity      float64
	TOR           string
	//Price         float64
}

func (mi *UnitInfo) Equal(other *UnitInfo) bool {
	return mi.DestinationId == other.DestinationId &&
		mi.Quantity == other.Quantity
}

// Holds information about the balance that made a specific payment
type BalanceInfo struct {
	UnitBalanceUuid  string
	MoneyBalanceUuid string
	AccountId        string // used when debited from shared balance
}

func (bi *BalanceInfo) Equal(other *BalanceInfo) bool {
	return bi.UnitBalanceUuid == other.UnitBalanceUuid &&
		bi.MoneyBalanceUuid == other.MoneyBalanceUuid &&
		bi.AccountId == other.AccountId
}

type TimeSpans []*TimeSpan

// Will delete all timespans that are `under` the timespan at index
func (timespans *TimeSpans) RemoveOverlapedFromIndex(index int) {
	tsList := *timespans
	ts := tsList[index]
	endOverlapIndex := index
	for i := index + 1; i < len(tsList); i++ {
		if tsList[i].TimeEnd.Before(ts.TimeEnd) || tsList[i].TimeEnd.Equal(ts.TimeEnd) {
			endOverlapIndex = i
		} else if tsList[i].TimeStart.Before(ts.TimeEnd) {
			tsList[i].TimeStart = ts.TimeEnd
			break
		}
	}
	if endOverlapIndex > index {
		newSliceEnd := len(tsList) - (endOverlapIndex - index)
		// delete overlapped
		copy(tsList[index+1:], tsList[endOverlapIndex+1:])
		for i := newSliceEnd; i < len(tsList); i++ {
			tsList[i] = nil
		}
		*timespans = tsList[:newSliceEnd]
		return
	}
	*timespans = tsList
}

// The paidTs will replace the timespans that are exactly `under` them
// from the reciver list
func (timespans *TimeSpans) OverlapWithTimeSpans(paidTs TimeSpans, newTs *TimeSpan, index int) bool {
	tsList := *timespans
	// calculate overlaped timespans
	var paidDuration time.Duration
	for _, pts := range paidTs {
		paidDuration += pts.GetDuration()
	}
	if paidDuration > 0 {
		// we must add the rest of the current ts to the remaingTs
		var remainingTs []*TimeSpan
		overlapStartIndex := index
		if newTs != nil {
			remainingTs = append(remainingTs, newTs)
			overlapStartIndex += 1
		}
		for tsi := overlapStartIndex; tsi < len(tsList); tsi++ {
			remainingTs = append(remainingTs, tsList[tsi])
		}
		overlapEndIndex := 0
		for i, rts := range remainingTs {
			if paidDuration >= rts.GetDuration() {
				paidDuration -= rts.GetDuration()
			} else {
				if paidDuration > 0 {
					// this ts was not fully paid
					fragment := rts.SplitByDuration(paidDuration)
					paidTs = append(paidTs, fragment)
				}
				// find the end position in tsList
				overlapEndIndex = overlapStartIndex + i
				break
			}
			// find the end position in tsList
			overlapEndIndex = overlapStartIndex + i
		}
		// delete from index to current
		if overlapEndIndex == len(tsList)-1 {
			tsList = tsList[:overlapStartIndex]
		} else {
			if overlapEndIndex+1 < len(tsList) {
				tsList = append(tsList[:overlapStartIndex], tsList[overlapEndIndex+1:]...)
			}
		}
		// append the timespans to outer tsList
		for i, pts := range paidTs {
			tsList = append(tsList, nil)
			copy(tsList[overlapStartIndex+i+1:], tsList[overlapStartIndex+i:])
			tsList[overlapStartIndex+i] = pts
		}
		*timespans = tsList
		return true
	}
	*timespans = tsList
	return false
}

func (tss TimeSpans) Compress() {
	for _, ts := range tss {
		var cIncrs Increments
		for _, incr := range ts.Increments {
			if len(cIncrs) == 0 || !cIncrs[len(cIncrs)-1].Equal(incr) {
				incr.GetCompressFactor() // sideefect
				cIncrs = append(cIncrs, incr)
			} else {
				cIncrs[len(cIncrs)-1].CompressFactor++
			}
		}
		ts.Increments = cIncrs
	}
}

func (tss TimeSpans) Decompress() {
	for _, ts := range tss {
		var incrs Increments
		for _, cIncr := range ts.Increments {
			for i := 0; i < cIncr.GetCompressFactor(); i++ {
				incrs = append(incrs, cIncr.Clone())
			}
		}
		ts.Increments = incrs
	}
}

func (incr *Increment) Clone() *Increment {
	nIncr := &Increment{
		Duration:            incr.Duration,
		Cost:                incr.Cost,
		BalanceRateInterval: incr.BalanceRateInterval,
		UnitInfo:            incr.UnitInfo,
		BalanceInfo:         incr.BalanceInfo,
	}
	return nIncr
}

func (incr *Increment) Equal(other *Increment) bool {
	return incr.Duration == other.Duration &&
		incr.Cost == other.Cost &&
		((incr.BalanceInfo == nil && other.BalanceInfo == nil) || incr.BalanceInfo.Equal(other.BalanceInfo)) &&
		((incr.BalanceRateInterval == nil && other.BalanceRateInterval == nil) || reflect.DeepEqual(incr.BalanceRateInterval, other.BalanceRateInterval)) &&
		((incr.UnitInfo == nil && other.UnitInfo == nil) || incr.UnitInfo.Equal(other.UnitInfo))
}

func (incr *Increment) GetCompressFactor() int {
	if incr.CompressFactor == 0 {
		incr.CompressFactor = 1
	}
	return incr.CompressFactor
}

type Increments []*Increment

func (incs Increments) GetTotalCost() float64 {
	cost := 0.0
	for _, increment := range incs {
		cost += (float64(increment.GetCompressFactor()) * increment.Cost)
	}
	return cost
}

func (incs Increments) Length() (length int) {
	for _, incr := range incs {
		length += incr.GetCompressFactor()
	}
	return
}

// Returns the duration of the timespan
func (ts *TimeSpan) GetDuration() time.Duration {
	return ts.TimeEnd.Sub(ts.TimeStart)
}

// Returns true if the given time is inside timespan range.
func (ts *TimeSpan) Contains(t time.Time) bool {
	return t.After(ts.TimeStart) && t.Before(ts.TimeEnd)
}

func (ts *TimeSpan) SetRateInterval(interval *RateInterval) {
	if interval == nil {
		return
	}
	if !ts.hasBetterRateIntervalThan(interval) {
		ts.RateInterval = interval
	}
}

// Returns the cost of the timespan according to the relevant cost interval.
// It also sets the Cost field of this timespan (used for refund on session
// manager debit loop where the cost cannot be recalculated)
func (ts *TimeSpan) getCost() float64 {
	if ts.Increments.Length() == 0 {
		if ts.RateInterval == nil {
			return 0
		}
		cost := ts.RateInterval.GetCost(ts.GetDuration(), ts.GetGroupStart())
		ts.Cost = utils.Round(cost, ts.RateInterval.Rating.RoundingDecimals, ts.RateInterval.Rating.RoundingMethod)
		return ts.Cost
	} else {
		cost := ts.Increments.GetTotalCost()
		if ts.RateInterval != nil && ts.RateInterval.Rating != nil {
			return utils.Round(cost, ts.RateInterval.Rating.RoundingDecimals, ts.RateInterval.Rating.RoundingMethod)
		} else {
			return utils.Round(cost, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
		}
	}
}

func (ts *TimeSpan) setRatingInfo(rp *RatingInfo) {
	ts.ratingInfo = rp
	ts.MatchedSubject = rp.MatchedSubject
	ts.MatchedPrefix = rp.MatchedPrefix
	ts.MatchedDestId = rp.MatchedDestId
	ts.RatingPlanId = rp.RatingPlanId
}

func (ts *TimeSpan) createIncrementsSlice() {
	if ts.RateInterval == nil {
		return
	}
	ts.Increments = make([]*Increment, 0)
	// create rated units series
	_, rateIncrement, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
	// we will use the cost calculated cost and devide by nb of increments
	// because ts cost is rounded
	//incrementCost := rate / rateUnit.Seconds() * rateIncrement.Seconds()
	nbIncrements := int(ts.GetDuration() / rateIncrement)
	incrementCost := ts.getCost() / float64(nbIncrements)
	incrementCost = utils.Round(incrementCost, globalRoundingDecimals, utils.ROUNDING_MIDDLE) // just get rid of the extra decimals
	for s := 0; s < nbIncrements; s++ {
		inc := &Increment{
			Duration:    rateIncrement,
			Cost:        incrementCost,
			BalanceInfo: &BalanceInfo{},
		}
		ts.Increments = append(ts.Increments, inc)
	}
	// put the rounded cost back in timespan
	ts.Cost = incrementCost * float64(nbIncrements)
}

// returns whether the timespan has all increments marked as paid and if not
// it also returns the first unpaied increment
func (ts *TimeSpan) IsPaid() (bool, int) {
	if ts.Increments.Length() == 0 {
		return false, 0
	}
	for incrementIndex, increment := range ts.Increments {
		if !increment.paid {
			return false, incrementIndex
		}
	}
	return true, len(ts.Increments)
}

/*
Splits the given timespan according to how it relates to the interval.
It will modify the endtime of the received timespan and it will return
a new timespan starting from the end of the received one.
The interval will attach itself to the timespan that overlaps the interval.
*/
func (ts *TimeSpan) SplitByRateInterval(i *RateInterval, data bool) (nts *TimeSpan) {
	// if the span is not in interval return nil
	//log.Printf("Checking: %+v (%v,%v)", i.Timing, ts.TimeStart, ts.TimeEnd)
	if !(i.Contains(ts.TimeStart, false) || i.Contains(ts.TimeEnd, true)) {
		//log.Print("Not in interval")
		return
	}
	//Logger.Debug(fmt.Sprintf("TS: %+v", ts))
	// split by GroupStart
	if i.Rating != nil {
		i.Rating.Rates.Sort()
		for _, rate := range i.Rating.Rates {
			//Logger.Debug(fmt.Sprintf("Rate: %+v", rate))
			if ts.GetGroupStart() < rate.GroupIntervalStart && ts.GetGroupEnd() > rate.GroupIntervalStart {
				//log.Print("Splitting")
				ts.SetRateInterval(i)
				splitTime := ts.TimeStart.Add(rate.GroupIntervalStart - ts.GetGroupStart())
				nts = &TimeSpan{
					TimeStart: splitTime,
					TimeEnd:   ts.TimeEnd,
				}
				nts.copyRatingInfo(ts)
				ts.TimeEnd = splitTime
				nts.SetRateInterval(i)
				nts.DurationIndex = ts.DurationIndex
				ts.SetNewDurationIndex(nts)
				// Logger.Debug(fmt.Sprintf("Group splitting: %+v %+v", ts, nts))
				return
			}
		}
	}
	if data {
		if i.Contains(ts.TimeStart, false) {
			ts.SetRateInterval(i)
		}
		return
	}
	// if the span is enclosed in the interval try to set as new interval and return nil
	//log.Printf("Timing: %+v", i.Timing)
	if i.Contains(ts.TimeStart, false) && i.Contains(ts.TimeEnd, true) {
		//log.Print("All in interval")
		ts.SetRateInterval(i)
		return
	}
	// if only the start time is in the interval split the interval to the right
	if i.Contains(ts.TimeStart, false) {
		//log.Print("Start in interval")
		splitTime := i.Timing.getRightMargin(ts.TimeStart)
		ts.SetRateInterval(i)
		if splitTime == ts.TimeStart || splitTime.Equal(ts.TimeEnd) {
			return
		}
		nts = &TimeSpan{
			TimeStart: splitTime,
			TimeEnd:   ts.TimeEnd,
		}
		nts.copyRatingInfo(ts)
		ts.TimeEnd = splitTime
		nts.DurationIndex = ts.DurationIndex
		ts.SetNewDurationIndex(nts)
		// Logger.Debug(fmt.Sprintf("right: %+v %+v", ts, nts))
		return
	}
	// if only the end time is in the interval split the interval to the left
	if i.Contains(ts.TimeEnd, true) {
		//log.Print("End in interval")
		//tmpTime := time.Date(ts.TimeStart.)
		splitTime := i.Timing.getLeftMargin(ts.TimeEnd)
		splitTime = utils.CopyHour(splitTime, ts.TimeStart)
		if splitTime.Equal(ts.TimeEnd) {
			return
		}
		nts = &TimeSpan{
			TimeStart: splitTime,
			TimeEnd:   ts.TimeEnd,
		}
		nts.copyRatingInfo(ts)
		ts.TimeEnd = splitTime
		nts.SetRateInterval(i)
		nts.DurationIndex = ts.DurationIndex
		ts.SetNewDurationIndex(nts)
		// Logger.Debug(fmt.Sprintf("left: %+v %+v", ts, nts))
		return
	}
	return
}

// Split the timespan at the given increment start
func (ts *TimeSpan) SplitByIncrement(index int) *TimeSpan {
	if index <= 0 || index >= len(ts.Increments) {
		return nil
	}
	timeStart := ts.GetTimeStartForIncrement(index)
	newTs := &TimeSpan{
		RateInterval: ts.RateInterval,
		TimeStart:    timeStart,
		TimeEnd:      ts.TimeEnd,
	}
	newTs.copyRatingInfo(ts)
	newTs.DurationIndex = ts.DurationIndex
	ts.TimeEnd = timeStart
	newTs.Increments = ts.Increments[index:]
	ts.Increments = ts.Increments[:index]
	ts.SetNewDurationIndex(newTs)
	return newTs
}

// Split the timespan at the given second
func (ts *TimeSpan) SplitByDuration(duration time.Duration) *TimeSpan {
	if duration <= 0 || duration >= ts.GetDuration() {
		return nil
	}
	timeStart := ts.TimeStart.Add(duration)
	newTs := &TimeSpan{
		RateInterval: ts.RateInterval,
		TimeStart:    timeStart,
		TimeEnd:      ts.TimeEnd,
	}
	newTs.copyRatingInfo(ts)
	newTs.DurationIndex = ts.DurationIndex
	ts.TimeEnd = timeStart
	// split the increment
	for incrIndex, incr := range ts.Increments {
		if duration-incr.Duration >= 0 {
			duration -= incr.Duration
		} else {

			splitIncrement := ts.Increments[incrIndex].Clone()
			splitIncrement.Duration -= duration
			ts.Increments[incrIndex].Duration = duration
			newTs.Increments = Increments{splitIncrement}
			if incrIndex < len(ts.Increments)-1 {
				newTs.Increments = append(newTs.Increments, ts.Increments[incrIndex+1:]...)
			}
			ts.Increments = ts.Increments[:incrIndex+1]
			break
		}
	}
	ts.SetNewDurationIndex(newTs)
	return newTs
}

// Splits the given timespan on activation period's activation time.
func (ts *TimeSpan) SplitByRatingPlan(rp *RatingInfo) (newTs *TimeSpan) {
	activationTime := rp.ActivationTime.In(ts.TimeStart.Location())
	if !ts.Contains(activationTime) {
		return nil
	}
	newTs = &TimeSpan{
		TimeStart: activationTime,
		TimeEnd:   ts.TimeEnd,
	}
	newTs.copyRatingInfo(ts)
	newTs.DurationIndex = ts.DurationIndex
	ts.TimeEnd = activationTime
	ts.SetNewDurationIndex(newTs)
	// Logger.Debug(fmt.Sprintf("RP SPLITTING: %+v %+v", ts, newTs))
	return
}

// Returns the starting time of this timespan
func (ts *TimeSpan) GetGroupStart() time.Duration {
	s := ts.DurationIndex - ts.GetDuration()
	if s < 0 {
		s = 0
	}
	return s
}

func (ts *TimeSpan) GetGroupEnd() time.Duration {
	return ts.DurationIndex
}

// sets the DurationIndex attribute to reflect new timespan
func (ts *TimeSpan) SetNewDurationIndex(nts *TimeSpan) {
	d := ts.DurationIndex - nts.GetDuration()
	if d < 0 {
		d = 0
	}
	ts.DurationIndex = d
}

func (nts *TimeSpan) copyRatingInfo(ts *TimeSpan) {
	if ts.ratingInfo == nil {
		return
	}
	nts.setRatingInfo(ts.ratingInfo)
}

// returns a time for the specified second in the time span
func (ts *TimeSpan) GetTimeStartForIncrement(index int) time.Time {
	return ts.TimeStart.Add(time.Duration(int64(index) * ts.Increments[0].Duration.Nanoseconds()))
}

func (ts *TimeSpan) RoundToDuration(duration time.Duration) {
	if duration < ts.GetDuration() {
		duration = utils.RoundDuration(duration, ts.GetDuration())
	}
	if duration > ts.GetDuration() {
		initialDuration := ts.GetDuration()
		ts.TimeEnd = ts.TimeStart.Add(duration)
		ts.DurationIndex = ts.DurationIndex + (duration - initialDuration)
	}
}

func (ts *TimeSpan) AddIncrement(inc *Increment) {
	ts.Increments = append(ts.Increments, inc)
	ts.TimeEnd.Add(inc.Duration)
}

func (ts *TimeSpan) hasBetterRateIntervalThan(interval *RateInterval) bool {
	if interval.Timing == nil {
		return false
	}
	otherLeftMargin := interval.Timing.getLeftMargin(ts.TimeStart)
	otherDistance := ts.TimeStart.Sub(otherLeftMargin)
	//log.Print("OTHER LEFT: ", otherLeftMargin)
	//log.Print("OTHER DISTANCE: ", otherDistance)
	// if the distance is negative it's not usable
	if otherDistance < 0 {
		return true
	}
	//log.Print("RI: ", ts.RateInterval)
	if ts.RateInterval == nil {
		return false
	}

	// the higher the weight the better
	if ts.RateInterval != nil &&
		ts.RateInterval.Weight < interval.Weight {
		return false
	}
	// check interval is closer than the new one
	ownLeftMargin := ts.RateInterval.Timing.getLeftMargin(ts.TimeStart)
	ownDistance := ts.TimeStart.Sub(ownLeftMargin)

	//log.Print("OWN LEFT: ", otherLeftMargin)
	//log.Print("OWN DISTANCE: ", otherDistance)
	//endOtherDistance := ts.TimeEnd.Sub(otherLeftMargin)

	// if own interval is closer than its better
	//log.Print(ownDistance)
	if ownDistance > otherDistance {
		return false
	}
	ownPrice, _, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
	otherPrice, _, _ := interval.GetRateParameters(ts.GetGroupStart())
	// if own price is smaller than it's better
	//log.Print(ownPrice, otherPrice)
	if ownPrice < otherPrice {
		return true
	}
	return true
}
