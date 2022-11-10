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
	"fmt"
	"reflect"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

/*
A unit in which a call will be split that has a specific price related interval attached to it.
*/
type TimeSpan struct {
	TimeStart      time.Time
	TimeEnd        time.Time
	Cost           float64
	RateInterval   *RateInterval
	DurationIndex  time.Duration // the call duration so far till TimeEnd
	Increments     Increments
	RoundIncrement *Increment
	MatchedSubject string
	MatchedPrefix  string
	MatchedDestId  string
	RatingPlanId   string
	CompressFactor int
	ratingInfo     *RatingInfo
}

type Increment struct {
	Duration       time.Duration
	Cost           float64
	BalanceInfo    *DebitInfo // need more than one for units with cost
	CompressFactor int
}

// Holds information about the balance that made a specific payment
type DebitInfo struct {
	Unit      *UnitInfo
	Monetary  *MonetaryInfo
	AccountID string // used when debited from shared balance
}

func (di *DebitInfo) Equal(other *DebitInfo) bool {
	return di.Unit.Equal(other.Unit) &&
		di.Monetary.Equal(other.Monetary) &&
		di.AccountID == other.AccountID
}

func (di *DebitInfo) Clone() *DebitInfo {
	nDi := &DebitInfo{
		AccountID: di.AccountID,
	}
	if di.Unit != nil {
		nDi.Unit = di.Unit.Clone()
	}
	if di.Monetary != nil {
		nDi.Monetary = di.Monetary.Clone()
	}
	return nDi
}

type MonetaryInfo struct {
	UUID         string
	ID           string
	Value        float64
	RateInterval *RateInterval
}

func (mi *MonetaryInfo) Clone() *MonetaryInfo {
	newMi := *mi
	return &newMi
}

func (mi *MonetaryInfo) Equal(other *MonetaryInfo) bool {
	if mi == nil && other == nil {
		return true
	}
	if mi == nil || other == nil {
		return false
	}
	return mi.UUID == other.UUID &&
		reflect.DeepEqual(mi.RateInterval, other.RateInterval)
}

type UnitInfo struct {
	UUID          string
	ID            string
	Value         float64
	DestinationID string
	Consumed      float64
	ToR           string
	RateInterval  *RateInterval
}

func (ui *UnitInfo) Clone() *UnitInfo {
	newUi := *ui
	return &newUi
}

func (ui *UnitInfo) Equal(other *UnitInfo) bool {
	if ui == nil && other == nil {
		return true
	}
	if ui == nil || other == nil {
		return false
	}
	return ui.UUID == other.UUID &&
		ui.DestinationID == other.DestinationID &&
		ui.Consumed == other.Consumed &&
		ui.ToR == other.ToR &&
		reflect.DeepEqual(ui.RateInterval, other.RateInterval)
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

func (tss *TimeSpans) Compress() { // must be pointer receiver
	for _, ts := range *tss {
		ts.Increments.Compress()
	}
	var cTss TimeSpans
	for _, ts := range *tss {
		if len(cTss) == 0 || !cTss[len(cTss)-1].Equal(ts) {
			ts.GetCompressFactor() // sideefect
			cTss = append(cTss, ts)
		} else {
			cTs := cTss[len(cTss)-1]
			cTs.CompressFactor++
			cTs.Cost += ts.Cost
			cTs.TimeEnd = ts.TimeEnd
			cTs.DurationIndex = ts.DurationIndex
		}
	}
	*tss = cTss
}

func (tss *TimeSpans) Decompress() { // must be pointer receiver
	for _, ts := range *tss {
		ts.Increments.Decompress()
	}
	var cTss TimeSpans
	for _, cTs := range *tss {
		var duration time.Duration
		if cTs.GetCompressFactor() > 1 {
			duration = cTs.GetUnitDuration()
		}
		for i := cTs.GetCompressFactor(); i > 1; i-- {
			uTs := &TimeSpan{}
			*uTs = *cTs // cloned by copy
			uTs.TimeEnd = cTs.TimeStart.Add(duration)
			uTs.DurationIndex = cTs.DurationIndex - time.Duration((i-1)*int(duration))
			uTs.CompressFactor = 1
			uTs.Cost = cTs.Cost / float64(cTs.GetCompressFactor())
			cTs.TimeStart = uTs.TimeEnd
			cTss = append(cTss, uTs)
		}
		cTs.Cost = cTs.GetUnitCost()
		cTs.CompressFactor = 1
		cTss = append(cTss, cTs)
	}
	*tss = cTss
}

func (tss *TimeSpans) Merge() { // Merge whenever possible
	tssVal := *tss
	if len(tssVal) > 2 { // Optimization for faster merge
		middle := len(tssVal) / 2
		tssVal1 := tssVal[:middle]
		tssVal2 := tssVal[middle:]
		tssVal1.Merge()
		tssVal2.Merge()
		tssVal = append(tssVal1, tssVal2...)
	}
	for i := 1; i < len(tssVal); i++ {
		if tssVal[i-1].Merge(tssVal[i]) {
			tssVal = append(tssVal[:i], tssVal[i+1:]...)
			i-- // Reschedule checking of last index since slice will decrease
		}
	}
	*tss = tssVal
}

func (incr *Increment) Clone() *Increment {
	nInc := &Increment{
		Duration:       incr.Duration,
		Cost:           incr.Cost,
		CompressFactor: incr.CompressFactor,
	}
	if incr.BalanceInfo != nil {
		nInc.BalanceInfo = incr.BalanceInfo.Clone()
	}
	return nInc
}

func (incr *Increment) Equal(other *Increment) bool {
	return incr.Duration == other.Duration &&
		incr.Cost == other.Cost &&
		((incr.BalanceInfo == nil && other.BalanceInfo == nil) || incr.BalanceInfo.Equal(other.BalanceInfo))
}

func (incr *Increment) GetCompressFactor() int {
	if incr.CompressFactor == 0 {
		incr.CompressFactor = 1
	}
	return incr.CompressFactor
}

func (incr *Increment) GetCost() float64 {
	return float64(incr.GetCompressFactor()) * incr.Cost
}

type Increments []*Increment

func (incs Increments) Clone() (cln Increments) {
	if incs == nil {
		return nil
	}
	cln = make(Increments, len(incs))
	for index, increment := range incs {
		cln[index] = increment.Clone()
	}
	return cln
}

func (incs Increments) Equal(other Increments) bool {
	if len(other) < len(incs) { // Protect index in case of not being the same size
		return false
	}
	for index, i := range incs {
		if !i.Equal(other[index]) || i.GetCompressFactor() != other[index].GetCompressFactor() {
			return false
		}
	}
	return true
}

func (incs *Increments) Compress() { // must be pointer receiver
	var cIncrs Increments
	for _, incr := range *incs {
		if len(cIncrs) == 0 || !cIncrs[len(cIncrs)-1].Equal(incr) {
			incr.GetCompressFactor() // sideefect
			cIncrs = append(cIncrs, incr)
		} else {
			cIncrs[len(cIncrs)-1].CompressFactor++
			if cIncrs[len(cIncrs)-1].BalanceInfo != nil && incr.BalanceInfo != nil {
				if cIncrs[len(cIncrs)-1].BalanceInfo.Monetary != nil && incr.BalanceInfo.Monetary != nil {
					cIncrs[len(cIncrs)-1].BalanceInfo.Monetary.Value = incr.BalanceInfo.Monetary.Value
				}
				if cIncrs[len(cIncrs)-1].BalanceInfo.Unit != nil && incr.BalanceInfo.Unit != nil {
					cIncrs[len(cIncrs)-1].BalanceInfo.Unit.Value = incr.BalanceInfo.Unit.Value
				}
			}
		}
	}
	*incs = cIncrs
}

func (incs *Increments) Decompress() { // must be pointer receiver
	var cIncrs Increments
	for _, cIncr := range *incs {
		cf := cIncr.GetCompressFactor()
		for i := 0; i < cf; i++ {
			incr := cIncr.Clone()
			incr.CompressFactor = 1
			// set right Values
			if incr.BalanceInfo != nil {
				if incr.BalanceInfo.Monetary != nil {
					incr.BalanceInfo.Monetary.Value += (float64(cf-(i+1)) * incr.Cost)
				}
				if incr.BalanceInfo.Unit != nil {
					incr.BalanceInfo.Unit.Value += (float64(cf-(i+1)) * incr.BalanceInfo.Unit.Consumed)
				}
			}
			cIncrs = append(cIncrs, incr)
		}
	}
	*incs = cIncrs
}

// Estimate whether the increments are the same ignoring the CompressFactor
func (incs Increments) SharingSignature(other Increments) bool {
	otherCloned := other.Clone()
	thisCloned := incs.Clone()
	otherCloned.Compress()
	thisCloned.Compress()
	if len(otherCloned) < len(thisCloned) { // Protect index in case of not being the same size
		return false
	}
	for index, i := range thisCloned {
		if !i.Equal(otherCloned[index]) {
			return false
		}
	}
	return true
}

func (incs Increments) GetTotalCost() float64 {
	cost := 0.0
	for _, increment := range incs {
		cost += increment.GetCost()
	}
	return utils.Round(cost, globalRoundingDecimals, utils.MetaRoundingMiddle)
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

// Returns the duration of a unitary timespan in a compressed set
func (ts *TimeSpan) GetUnitDuration() time.Duration {
	return time.Duration(int(ts.TimeEnd.Sub(ts.TimeStart)) / ts.GetCompressFactor())
}

func (ts *TimeSpan) GetUnitCost() float64 {
	return ts.Cost / float64(ts.GetCompressFactor())
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
func (ts *TimeSpan) CalculateCost() float64 {
	if ts.Increments.Length() == 0 {
		if ts.RateInterval == nil {
			return 0
		}
		return ts.RateInterval.GetCost(ts.GetDuration(), ts.GetGroupStart())
	}
	return ts.Increments.GetTotalCost() * float64(ts.GetCompressFactor())
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
	// create rated units series
	_, rateIncrement, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
	// we will use the calculated cost and devide by nb of increments
	// because ts cost is rounded
	//incrementCost := rate / rateUnit.Seconds() * rateIncrement.Seconds()
	nbIncrements := int(ts.GetDuration() / rateIncrement)
	if nbIncrements > config.CgrConfig().RalsCfg().MaxIncrements {
		utils.Logger.Warning(fmt.Sprintf("error: <%s with %+v>, when creating increments slice, TimeSpan: %s", utils.ErrMaxIncrementsExceeded, nbIncrements, utils.ToJSON(ts)))
		ts.Increments = make([]*Increment, 0)
		return
	}
	incrementCost := ts.CalculateCost() / float64(nbIncrements)
	incrementCost = utils.Round(incrementCost, globalRoundingDecimals, utils.MetaRoundingMiddle)
	ts.Increments = make([]*Increment, nbIncrements)
	for i := range ts.Increments {
		ts.Increments[i] = &Increment{
			Duration:    rateIncrement,
			Cost:        incrementCost,
			BalanceInfo: &DebitInfo{},
		}
	}
	// put the rounded cost back in timespan
	ts.Cost = incrementCost * float64(nbIncrements)
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
	// split by GroupStart
	if i.Rating != nil {
		i.Rating.Rates.Sort()
		for _, rate := range i.Rating.Rates {
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
		return
	}
	// if only the end time is in the interval split the interval to the left
	if i.Contains(ts.TimeEnd, true) {
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
		return
	}
	return
}

// Split the timespan at the given increment start
func (ts *TimeSpan) SplitByIncrement(index int) (newTs *TimeSpan) {
	if index <= 0 || index >= len(ts.Increments) {
		return
	}
	timeStart := ts.GetTimeStartForIncrement(index)
	newTs = &TimeSpan{
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
	return
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
	return
}

// Splits the given timespan on activation period's activation time.
func (ts *TimeSpan) SplitByDay() (newTs *TimeSpan) {
	if ts.TimeStart.Day() == ts.TimeEnd.Day() || utils.TimeIs0h(ts.TimeEnd) {
		return
	}

	splitDate := ts.TimeStart.AddDate(0, 0, 1)
	splitDate = time.Date(splitDate.Year(), splitDate.Month(), splitDate.Day(), 0, 0, 0, 0, splitDate.Location())
	newTs = &TimeSpan{
		TimeStart: splitDate,
		TimeEnd:   ts.TimeEnd,
	}
	newTs.copyRatingInfo(ts)
	newTs.DurationIndex = ts.DurationIndex
	ts.TimeEnd = splitDate
	ts.SetNewDurationIndex(newTs)
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

	start := ts.TimeStart
	for incIndex, inc := range ts.Increments {
		if incIndex < index {
			start = start.Add(time.Duration(inc.Duration.Nanoseconds()))
		}
	}
	return start
	//return ts.TimeStart.Add(time.Duration(int64(index) * ts.Increments[0].Duration.Nanoseconds()))
}

func (ts *TimeSpan) RoundToDuration(duration time.Duration) {
	tsDur := ts.GetDuration()
	if duration < tsDur {
		duration = utils.RoundDuration(duration, tsDur)
	}
	if duration > tsDur {
		initialDuration := tsDur
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

	// if own interval is closer than its better
	if ownDistance > otherDistance {
		return false
	}
	ownPrice, _, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
	otherPrice, _, _ := interval.GetRateParameters(ts.GetGroupStart())
	// if own price is smaller than it's better
	if ownPrice < otherPrice {
		return true
	}
	return true
}

func (ts *TimeSpan) Equal(other *TimeSpan) bool {
	return ts.Increments.Equal(other.Increments) &&
		ts.RateInterval.Equal(other.RateInterval) &&
		ts.GetUnitCost() == other.GetUnitCost() &&
		ts.GetUnitDuration() == other.GetUnitDuration() &&
		ts.MatchedSubject == other.MatchedSubject &&
		ts.MatchedPrefix == other.MatchedPrefix &&
		ts.MatchedDestId == other.MatchedDestId &&
		ts.RatingPlanId == other.RatingPlanId
}

// Estimate if they share charging signature
func (ts *TimeSpan) SharingSignature(other *TimeSpan) bool {
	if ts.GetCompressFactor() != other.GetCompressFactor() ||
		!ts.Increments.SharingSignature(other.Increments) ||
		!ts.RateInterval.Equal(other.RateInterval) ||
		ts.MatchedSubject != other.MatchedSubject ||
		ts.MatchedPrefix != other.MatchedPrefix ||
		ts.MatchedDestId != other.MatchedDestId ||
		ts.RatingPlanId != other.RatingPlanId {
		return false
	}
	return true
}

func (ts *TimeSpan) GetCompressFactor() int {
	if ts.CompressFactor == 0 {
		ts.CompressFactor = 1
	}
	return ts.CompressFactor
}

// Merges timespans if they share the same charging signature, useful to run in SM before compressing
func (ts *TimeSpan) Merge(other *TimeSpan) bool {
	if !ts.SharingSignature(other) {
		return false
	} else if !ts.TimeEnd.Equal(other.TimeStart) { // other needs to continue ts for merge to be possible
		return false
	}
	ts.TimeEnd = other.TimeEnd
	ts.Cost += other.Cost
	ts.DurationIndex = other.DurationIndex
	ts.Increments = append(ts.Increments, other.Increments...)
	return true
}
