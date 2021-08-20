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
	"errors"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// NewBareEventCost will intialize the EventCost with minimum information
func NewBareEventCost() *EventCost {
	return &EventCost{
		Rating:        make(Rating),
		Accounting:    make(Accounting),
		RatingFilters: make(RatingFilters),
		Rates:         make(ChargedRates),
		Timings:       make(ChargedTimings),
		Charges:       make([]*ChargingInterval, 0),
		cache:         utils.MapStorage{},
	}
}

// NewEventCostFromCallCost will initilaize the EventCost from a CallCost
func NewEventCostFromCallCost(cc *CallCost, cgrID, runID string) (ec *EventCost) {
	ec = NewBareEventCost()
	ec.CGRID = cgrID
	ec.RunID = runID
	ec.AccountSummary = cc.AccountSummary
	if len(cc.Timespans) != 0 {
		ec.Charges = make([]*ChargingInterval, len(cc.Timespans))
		ec.StartTime = cc.Timespans[0].TimeStart
	}
	for i, ts := range cc.Timespans {
		cIl := &ChargingInterval{CompressFactor: ts.CompressFactor}
		rf := RatingMatchedFilters{
			utils.DestinationID:         ts.MatchedDestId,
			utils.DestinationPrefixName: ts.MatchedPrefix,
			utils.RatingPlanID:          ts.RatingPlanId,
			utils.Subject:               ts.MatchedSubject,
		}
		isPause := ts.RatingPlanId == utils.MetaPause
		cIl.RatingID = ec.ratingIDForRateInterval(ts.RateInterval, rf, isPause)
		if len(ts.Increments) != 0 {
			cIl.Increments = make([]*ChargingIncrement, 0, len(ts.Increments)+1)
		}
		for _, incr := range ts.Increments {
			cIl.Increments = append(cIl.Increments, ec.newChargingIncrement(incr, rf, false, isPause))
		}
		if ts.RoundIncrement != nil {
			rIncr := ec.newChargingIncrement(ts.RoundIncrement, rf, true, false)
			rIncr.Cost = -rIncr.Cost
			cIl.Increments = append(cIl.Increments, rIncr)
		}
		ec.Charges[i] = cIl
	}
	return
}

// newChargingIncrement creates ChargingIncrement from a Increment
// special case if is the roundIncrement the rateID is *rounding
func (ec *EventCost) newChargingIncrement(incr *Increment, rf RatingMatchedFilters, roundedIncrement, isPause bool) (cIt *ChargingIncrement) {
	cIt = &ChargingIncrement{
		Usage:          incr.Duration,
		Cost:           incr.Cost,
		CompressFactor: incr.CompressFactor,
	}
	if incr.BalanceInfo == nil {
		return
	}
	if roundedIncrement {
		isPause = false
	}
	rateID := utils.MetaRounding
	//AccountingID
	if incr.BalanceInfo.Unit != nil {
		// 2 balances work-around
		ecUUID := utils.MetaNone // populate no matter what due to Unit not nil
		if incr.BalanceInfo.Monetary != nil {
			if !roundedIncrement {
				rateID = ec.ratingIDForRateInterval(incr.BalanceInfo.Monetary.RateInterval, rf, isPause)
			}
			bc := &BalanceCharge{
				AccountID:   incr.BalanceInfo.AccountID,
				BalanceUUID: incr.BalanceInfo.Monetary.UUID,
				Units:       incr.Cost,
				RatingID:    rateID,
			}
			if isPause {
				ecUUID = utils.MetaPause
				ec.Accounting[ecUUID] = bc
			} else {
				ecUUID = ec.Accounting.GetIDWithSet(bc)
			}
		}
		if !roundedIncrement {
			rateID = ec.ratingIDForRateInterval(incr.BalanceInfo.Unit.RateInterval, rf, isPause)
		}
		bc := &BalanceCharge{
			AccountID:     incr.BalanceInfo.AccountID,
			BalanceUUID:   incr.BalanceInfo.Unit.UUID,
			Units:         incr.BalanceInfo.Unit.Consumed,
			RatingID:      rateID,
			ExtraChargeID: ecUUID,
		}
		if isPause {
			cIt.AccountingID = utils.MetaPause
			ec.Accounting[utils.MetaPause] = bc
		} else {
			cIt.AccountingID = ec.Accounting.GetIDWithSet(bc)
		}
	} else if incr.BalanceInfo.Monetary != nil { // Only monetary
		if !roundedIncrement {
			rateID = ec.ratingIDForRateInterval(incr.BalanceInfo.Monetary.RateInterval, rf, isPause)
		}
		bc := &BalanceCharge{
			AccountID:   incr.BalanceInfo.AccountID,
			BalanceUUID: incr.BalanceInfo.Monetary.UUID,
			Units:       incr.Cost,
			RatingID:    rateID,
		}
		if isPause {
			cIt.AccountingID = utils.MetaPause
			ec.Accounting[utils.MetaPause] = bc
		} else {
			cIt.AccountingID = ec.Accounting.GetIDWithSet(bc)
		}
	}
	return
}

// EventCost stores cost for an Event
type EventCost struct {
	CGRID          string
	RunID          string
	StartTime      time.Time
	Usage          *time.Duration
	Cost           *float64 // pointer so we can nil it when dirty
	Charges        []*ChargingInterval
	AccountSummary *AccountSummary // Account summary at the end of the event calculation
	Rating         Rating
	Accounting     Accounting
	RatingFilters  RatingFilters
	Rates          ChargedRates
	Timings        ChargedTimings

	cache utils.MapStorage
}

func (ec *EventCost) initCache() {
	if ec != nil {
		ec.cache = utils.MapStorage{}
	}
}

func (ec *EventCost) ratingIDForRateInterval(ri *RateInterval, rf RatingMatchedFilters, isPause bool) string {
	if ri == nil || ri.Rating == nil {
		return utils.EmptyString
	}
	var rfUUID string
	if rf != nil {
		if isPause {
			rfUUID = utils.MetaPause
			ec.RatingFilters[rfUUID] = rf
		} else {
			rfUUID = ec.RatingFilters.GetIDWithSet(rf)
		}
	}
	var tmID string
	if ri.Timing != nil {
		// timingID can have random UUID to be reused by other Rates from EventCost
		tmID = ec.Timings.GetIDWithSet(&ChargedTiming{
			Years:     ri.Timing.Years,
			Months:    ri.Timing.Months,
			MonthDays: ri.Timing.MonthDays,
			WeekDays:  ri.Timing.WeekDays,
			StartTime: ri.Timing.StartTime,
		})
	}
	var rtUUID string
	if len(ri.Rating.Rates) != 0 {
		if isPause {
			rtUUID = utils.MetaPause
			ec.Rates[rtUUID] = ri.Rating.Rates
		} else {
			rtUUID = ec.Rates.GetIDWithSet(ri.Rating.Rates)
		}
	}
	ru := &RatingUnit{
		ConnectFee:       ri.Rating.ConnectFee,
		RoundingMethod:   ri.Rating.RoundingMethod,
		RoundingDecimals: ri.Rating.RoundingDecimals,
		MaxCost:          ri.Rating.MaxCost,
		MaxCostStrategy:  ri.Rating.MaxCostStrategy,
		TimingID:         tmID,
		RatesID:          rtUUID,
		RatingFiltersID:  rfUUID,
	}
	if isPause {
		ec.Rating[utils.MetaPause] = ru
		return utils.MetaPause
	}
	return ec.Rating.GetIDWithSet(ru)
}

func (ec *EventCost) rateIntervalForRatingID(ratingID string) (ri *RateInterval) {
	if ratingID == "" {
		return
	}
	cIlRU := ec.Rating[ratingID]
	ri = new(RateInterval)
	ri.Rating = &RIRate{ConnectFee: cIlRU.ConnectFee,
		RoundingMethod:   cIlRU.RoundingMethod,
		RoundingDecimals: cIlRU.RoundingDecimals,
		MaxCost:          cIlRU.MaxCost, MaxCostStrategy: cIlRU.MaxCostStrategy}
	if cIlRU.RatesID != "" {
		ri.Rating.Rates = ec.Rates[cIlRU.RatesID]
	}
	if cIlRU.TimingID != "" {
		cIlTm := ec.Timings[cIlRU.TimingID]
		ri.Timing = &RITiming{ID: cIlRU.TimingID, Years: cIlTm.Years, Months: cIlTm.Months, MonthDays: cIlTm.MonthDays,
			WeekDays: cIlTm.WeekDays, StartTime: cIlTm.StartTime}
	}
	return
}

// Clone will create a clone of the object
func (ec *EventCost) Clone() (cln *EventCost) {
	if ec == nil {
		return
	}
	cln = new(EventCost)
	cln.initCache()
	cln.CGRID = ec.CGRID
	cln.RunID = ec.RunID
	cln.StartTime = ec.StartTime
	if ec.Usage != nil {
		cln.Usage = utils.DurationPointer(*ec.Usage)
	}
	if ec.Cost != nil {
		cln.Cost = utils.Float64Pointer(*ec.Cost)
	}
	if ec.Charges != nil {
		cln.Charges = make([]*ChargingInterval, len(ec.Charges))
		for i, cIl := range ec.Charges {
			cln.Charges[i] = cIl.Clone()
		}
	}
	if ec.AccountSummary != nil {
		cln.AccountSummary = ec.AccountSummary.Clone()
	}
	if ec.Rating != nil {
		cln.Rating = ec.Rating.Clone()
	}
	if ec.Accounting != nil {
		cln.Accounting = ec.Accounting.Clone()
	}
	if ec.RatingFilters != nil {
		cln.RatingFilters = ec.RatingFilters.Clone()
	}
	if ec.Rates != nil {
		cln.Rates = ec.Rates.Clone()
	}
	if ec.Timings != nil {
		cln.Timings = ec.Timings.Clone()
	}
	return
}

// Compute aggregates all the compute methods on EventCost
func (ec *EventCost) Compute() {
	ec.GetUsage()
	ec.ComputeEventCostUsageIndexes()
	ec.GetCost()
}

// ResetCounters will reset all the computed cached values
func (ec *EventCost) ResetCounters() {
	ec.Cost = nil
	ec.Usage = nil
	for _, cIl := range ec.Charges {
		cIl.cost = nil
		cIl.usage = nil
		cIl.ecUsageIdx = nil
	}
}

// GetCost iterates through Charges, computing EventCost.Cost
func (ec *EventCost) GetCost() float64 {
	if ec.Cost == nil {
		var cost float64
		for _, ci := range ec.Charges {
			cost += ci.TotalCost(ec.Accounting)
		}
		cost = utils.Round(cost, globalRoundingDecimals, utils.MetaRoundingMiddle)
		ec.Cost = &cost
	}
	return *ec.Cost
}

// GetUsage iterates through Charges, computing EventCost.Usage
func (ec *EventCost) GetUsage() time.Duration {
	if ec.Usage == nil {
		var usage time.Duration
		for _, ci := range ec.Charges {
			usage += time.Duration(ci.Usage().Nanoseconds() * int64(ci.CompressFactor))
		}
		ec.Usage = &usage
	}
	return *ec.Usage
}

// ComputeEventCostUsageIndexes will iterate through Chargers and populate their ecUsageIdx
func (ec *EventCost) ComputeEventCostUsageIndexes() {
	var totalUsage time.Duration
	for _, cIl := range ec.Charges {
		if cIl.ecUsageIdx == nil {
			cIl.ecUsageIdx = utils.DurationPointer(totalUsage)
		}
		totalUsage += time.Duration(cIl.Usage().Nanoseconds() * int64(cIl.CompressFactor))
	}
}

// AsRefundIncrements converts an EventCost into a CallDescriptor
func (ec *EventCost) AsRefundIncrements(tor string) (cd *CallDescriptor) {
	cd = &CallDescriptor{
		CgrID:         ec.CGRID,
		RunID:         ec.RunID,
		ToR:           tor,
		TimeStart:     ec.StartTime,
		TimeEnd:       ec.StartTime.Add(ec.GetUsage()),
		DurationIndex: ec.GetUsage(),
	}
	if len(ec.Charges) == 0 {
		return
	}
	var nrIcrms int
	for _, cIl := range ec.Charges {
		nrIcrms += (cIl.CompressFactor * len(cIl.Increments))
	}
	cd.Increments = make(Increments, nrIcrms)
	var iIdx int
	for _, cIl := range ec.Charges {
		for i := 0; i < cIl.CompressFactor; i++ {
			for _, cIcrm := range cIl.Increments {
				cd.Increments[iIdx] = &Increment{
					Cost:           cIcrm.Cost,
					Duration:       cIcrm.Usage,
					CompressFactor: cIcrm.CompressFactor,
				}
				if cIcrm.AccountingID != utils.EmptyString {
					cd.Increments[iIdx].BalanceInfo = &DebitInfo{
						AccountID: ec.Accounting[cIcrm.AccountingID].AccountID,
					}
					blncSmry := ec.AccountSummary.BalanceSummaries.BalanceSummaryWithUUD(ec.Accounting[cIcrm.AccountingID].BalanceUUID)
					if blncSmry.Type == utils.MetaMonetary {
						cd.Increments[iIdx].BalanceInfo.Monetary = &MonetaryInfo{UUID: blncSmry.UUID}
					} else if utils.NonMonetaryBalances.Has(blncSmry.Type) {
						cd.Increments[iIdx].BalanceInfo.Unit = &UnitInfo{UUID: blncSmry.UUID}
					}
					if ec.Accounting[cIcrm.AccountingID].ExtraChargeID == utils.MetaNone ||
						ec.Accounting[cIcrm.AccountingID].ExtraChargeID == utils.EmptyString {
						iIdx++
						continue
					}
					// extra charges, ie: non-free *voice
					extraSmry := ec.AccountSummary.BalanceSummaries.BalanceSummaryWithUUD(
						ec.Accounting[ec.Accounting[cIcrm.AccountingID].ExtraChargeID].BalanceUUID)
					if extraSmry.Type == utils.MetaMonetary {
						cd.Increments[iIdx].BalanceInfo.Monetary = &MonetaryInfo{UUID: extraSmry.UUID}
					} else if utils.NonMonetaryBalances.Has(blncSmry.Type) {
						cd.Increments[iIdx].BalanceInfo.Unit = &UnitInfo{UUID: extraSmry.UUID}
					}
				}
				iIdx++
			}
		}
	}
	return
}

// AsCallCost converts an EventCost into a CallCost
func (ec *EventCost) AsCallCost(tor string) *CallCost {
	cc := &CallCost{
		ToR:            utils.FirstNonEmpty(tor, utils.MetaVoice),
		Cost:           ec.GetCost(),
		RatedUsage:     float64(ec.GetUsage().Nanoseconds()),
		AccountSummary: ec.AccountSummary,
	}
	cc.Timespans = make(TimeSpans, len(ec.Charges))
	for i, cIl := range ec.Charges {
		ts := &TimeSpan{
			Cost:           cIl.Cost(ec.Accounting),
			DurationIndex:  *cIl.Usage(),
			CompressFactor: cIl.CompressFactor,
		}
		if cIl.ecUsageIdx == nil { // index was not populated yet
			ec.ComputeEventCostUsageIndexes()
		}
		ts.TimeStart = ec.StartTime.Add(*cIl.ecUsageIdx)
		ts.TimeEnd = ts.TimeStart.Add(
			time.Duration(cIl.Usage().Nanoseconds() * int64(cIl.CompressFactor)))
		if cIl.RatingID != "" &&
			ec.Rating[cIl.RatingID].RatingFiltersID != "" {
			rfs := ec.RatingFilters[ec.Rating[cIl.RatingID].RatingFiltersID]
			ts.MatchedSubject = rfs[utils.Subject].(string)
			ts.MatchedPrefix = rfs[utils.DestinationPrefixName].(string)
			ts.MatchedDestId = rfs[utils.DestinationID].(string)
			ts.RatingPlanId = rfs[utils.RatingPlanID].(string)
		}
		ts.RateInterval = ec.rateIntervalForRatingID(cIl.RatingID)

		incrs := cIl.Increments
		if l := len(cIl.Increments); l != 0 {
			if cIl.Increments[l-1].Cost != 0 &&
				ec.Accounting[cIl.Increments[l-1].AccountingID].RatingID == utils.MetaRounding {
				// special case: if the last increment has the ratingID equal to *rounding
				// we consider it as the roundIncrement
				l--
				incrs = incrs[:l]
				ts.RoundIncrement = ec.newIntervalFromCharge(cIl.Increments[l-1])
				ts.RoundIncrement.Cost = -ts.RoundIncrement.Cost
			}
			ts.Increments = make(Increments, l)
		}
		for j, cInc := range incrs {
			ts.Increments[j] = ec.newIntervalFromCharge(cInc)
		}
		cc.Timespans[i] = ts
	}
	return cc
}

// newIntervalFromCharge creates Increment from a ChargingIncrement
func (ec *EventCost) newIntervalFromCharge(cInc *ChargingIncrement) (incr *Increment) {
	incr = &Increment{
		Duration:       cInc.Usage,
		Cost:           cInc.Cost,
		CompressFactor: cInc.CompressFactor,
		BalanceInfo:    new(DebitInfo),
	}
	if len(cInc.AccountingID) == 0 {
		return
	}
	cBC := ec.Accounting[cInc.AccountingID]
	incr.BalanceInfo.AccountID = cBC.AccountID
	var balanceType string
	if cBC.BalanceUUID != "" {
		if ec.AccountSummary != nil {
			for _, b := range ec.AccountSummary.BalanceSummaries {
				if b.UUID == cBC.BalanceUUID {
					balanceType = b.Type
					break
				}
			}
		}
	}
	if utils.SliceHasMember([]string{utils.MetaData, utils.MetaVoice}, balanceType) && cBC.ExtraChargeID == "" {
		cBC.ExtraChargeID = utils.MetaNone // mark the balance to be exported as Unit type
	}
	if cBC.ExtraChargeID != "" { // have both monetary and data
		// Work around, enforce logic with 2 balances for *voice/*monetary combination
		// so we can stay compatible with CallCost
		incr.BalanceInfo.Unit = &UnitInfo{UUID: cBC.BalanceUUID, Consumed: cBC.Units}
		incr.BalanceInfo.Unit.RateInterval = ec.rateIntervalForRatingID(cBC.RatingID)
		if cBC.ExtraChargeID != utils.MetaNone {
			cBC = ec.Accounting[cBC.ExtraChargeID] // overwrite original balance so we can process it in one place
		}
	}
	if cBC.ExtraChargeID != utils.MetaNone {
		incr.BalanceInfo.Monetary = &MonetaryInfo{UUID: cBC.BalanceUUID}
		incr.BalanceInfo.Monetary.RateInterval = ec.rateIntervalForRatingID(cBC.RatingID)
	}
	return
}

// ratingGetIDFromEventCost retrieves UUID based on data from another EventCost
func (ec *EventCost) ratingGetIDFromEventCost(oEC *EventCost, oRatingID string) string {
	if oRatingID == utils.EmptyString {
		return utils.EmptyString
	} else if oRatingID == utils.MetaPause {
		oCIlRating := oEC.Rating[oRatingID].Clone() // clone so we don't influence the original data
		oCIlRating.TimingID = utils.MetaPause
		ec.Timings[utils.MetaPause] = oEC.Timings[oCIlRating.TimingID]
		oCIlRating.RatingFiltersID = utils.MetaPause
		ec.RatingFilters[utils.MetaPause] = oEC.RatingFilters[oCIlRating.RatingFiltersID]
		oCIlRating.RatesID = utils.MetaPause
		ec.Rates[utils.MetaPause] = oEC.Rates[oCIlRating.RatesID]
		ec.Rating[utils.MetaPause] = oCIlRating
		return utils.MetaPause
	}
	oCIlRating := oEC.Rating[oRatingID].Clone() // clone so we don't influence the original data
	oCIlRating.TimingID = ec.Timings.GetIDWithSet(oEC.Timings[oCIlRating.TimingID])
	oCIlRating.RatingFiltersID = ec.RatingFilters.GetIDWithSet(oEC.RatingFilters[oCIlRating.RatingFiltersID])
	oCIlRating.RatesID = ec.Rates.GetIDWithSet(oEC.Rates[oCIlRating.RatesID])
	return ec.Rating.GetIDWithSet(oCIlRating)
}

// accountingGetIDFromEventCost retrieves UUID based on data from another EventCost
func (ec *EventCost) accountingGetIDFromEventCost(oEC *EventCost, oAccountingID string) string {
	if oAccountingID == "" || oAccountingID == utils.MetaNone {
		return ""
	} else if oAccountingID == utils.MetaPause { // *pause represent a pause in debited session
		oBC := oEC.Accounting[oAccountingID].Clone()
		oBC.RatingID = ec.ratingGetIDFromEventCost(oEC, oBC.RatingID)
		ec.Accounting[utils.MetaPause] = oBC
		return utils.MetaPause
	}
	oBC := oEC.Accounting[oAccountingID].Clone()
	oBC.RatingID = ec.ratingGetIDFromEventCost(oEC, oBC.RatingID)
	oBC.ExtraChargeID = ec.accountingGetIDFromEventCost(oEC, oBC.ExtraChargeID)
	return ec.Accounting.GetIDWithSet(oBC)
}

// appendCIl appends a ChargingInterval to existing chargers
// no compression done at ChargingInterval level, attempted on ChargingIncrement level
func (ec *EventCost) appendCIlFromEC(oEC *EventCost, cIlIdx int) {
	cIl := oEC.Charges[cIlIdx]
	cIlCln := cIl.Clone() // add/modify data on clone instead of original
	cIlCln.RatingID = ec.ratingGetIDFromEventCost(oEC, cIl.RatingID)
	lastCIl := ec.Charges[len(ec.Charges)-1]
	lastCIt := lastCIl.Increments[len(lastCIl.Increments)-1]
	appendChargingIncrement := lastCIl.CompressFactor == 1 &&
		lastCIl.RatingID == cIlCln.RatingID // attempt compressing of the ChargingIncrements
	var idxFirstCIt *int // keep here the reference towards last not appended charging increment so we can create separate ChargingInterval
	var idxLastCF *int   // reference towards last compress not absorbed by ec.Charges
	for cF := cIl.CompressFactor; cF > 0; cF-- {
		for i, cIt := range cIl.Increments {
			cIlCln.Increments[i].AccountingID = ec.accountingGetIDFromEventCost(oEC, cIt.AccountingID)
			if idxFirstCIt != nil {
				continue
			}
			if !appendChargingIncrement ||
				!lastCIt.PartiallyEquals(cIlCln.Increments[i]) {
				idxFirstCIt = utils.IntPointer(i)
				idxLastCF = utils.IntPointer(cF)
				continue
			}
			lastCIt.CompressFactor += cIt.CompressFactor // compress the iterated ChargingIncrement
		}
	}
	if lastCIl.PartiallyEquals(cIlCln) { // the two CIls are equal, compress the original one
		lastCIl.CompressFactor += cIlCln.CompressFactor
		return
	}
	if idxFirstCIt != nil { // CIt was not completely absorbed
		cIl.RatingID = cIlCln.RatingID // reuse cIl so we don't clone again
		cIl.CompressFactor = 1
		cIl.Increments = cIlCln.Increments[*idxFirstCIt:]
		ec.Charges = append(ec.Charges, cIl)
		if *idxLastCF > 1 { // add the remaining part out of original ChargingInterval
			cIlCln.CompressFactor = *idxLastCF - 1
			// append the cloned charge in order to not keep refrences
			// of the increments used for previous charge
			ec.Charges = append(ec.Charges, cIlCln.Clone())
		}
	}
}

// AppendChargingInterval appends or compresses a &ChargingInterval to existing ec.Chargers
func (ec *EventCost) appendChargingIntervalFromEventCost(oEC *EventCost, cIlIdx int) {
	lenChargers := len(ec.Charges)
	if lenChargers != 0 && ec.Charges[lenChargers-1].PartiallyEquals(oEC.Charges[cIlIdx]) {
		ec.Charges[lenChargers-1].CompressFactor++
	} else {
		ec.appendCIlFromEC(oEC, cIlIdx)
	}
}

// SyncKeys will sync the keys present into ec with the ones in refEC
func (ec *EventCost) SyncKeys(refEC *EventCost) {
	// sync RatingFilters
	sncedRFilterIDs := make(map[string]string)
	for key, rf := range ec.RatingFilters {
		for refKey, refRf := range refEC.RatingFilters {
			if rf.Equals(refRf) {
				delete(ec.RatingFilters, key)
				sncedRFilterIDs[key] = refKey
				ec.RatingFilters[refKey] = rf
				break
			}
		}
	}
	// sync Rates
	sncedRateIDs := make(map[string]string)
	for key, rt := range ec.Rates {
		for refKey, refRt := range refEC.Rates {
			if rt.Equals(refRt) {
				delete(ec.Rates, key)
				sncedRateIDs[key] = refKey
				ec.Rates[refKey] = rt
				break
			}
		}
	}
	// sync Timings
	sncedTimingIDs := make(map[string]string)
	for key, tm := range ec.Timings {
		for refKey, refTm := range refEC.Timings {
			if tm.Equals(refTm) {
				delete(ec.Timings, key)
				sncedTimingIDs[key] = refKey
				ec.Timings[refKey] = tm
				break
			}
		}
	}
	// sync Rating
	sncedRatingIDs := make(map[string]string)
	for key, ru := range ec.Rating {
		if tmRefKey, has := sncedTimingIDs[ru.TimingID]; has {
			ru.TimingID = tmRefKey
		}
		if rtRefID, has := sncedRateIDs[ru.RatesID]; has {
			ru.RatesID = rtRefID
		}
		if rfRefID, has := sncedRFilterIDs[ru.RatingFiltersID]; has {
			ru.RatingFiltersID = rfRefID
		}
		for refKey, refRU := range refEC.Rating {
			if ru.Equals(refRU) {
				delete(ec.Rating, key)
				sncedRatingIDs[key] = refKey
				ec.Rating[refKey] = ru
				break
			}
		}
	}
	// sync Accounting
	sncedAcntIDs := make(map[string]string)
	for key, acnt := range ec.Accounting {
		if rtRefKey, has := sncedRatingIDs[acnt.RatingID]; has {
			acnt.RatingID = rtRefKey
		}
		for refKey, refAcnt := range refEC.Accounting {
			if acnt.Equals(refAcnt) {
				delete(ec.Accounting, key)
				sncedAcntIDs[key] = refKey
				ec.Accounting[refKey] = acnt
				break
			}
		}
	}
	// correct the ExtraCharge
	for _, acnt := range ec.Accounting {
		if acntRefID, has := sncedAcntIDs[acnt.ExtraChargeID]; has {
			acnt.ExtraChargeID = acntRefID
		}
	}
	// need another sync for the corrected ExtraChargeIDs
	for key, acnt := range ec.Accounting {
		for refKey, refAcnt := range refEC.Accounting {
			if acnt.Equals(refAcnt) {
				delete(ec.Accounting, key)
				sncedAcntIDs[key] = refKey
				ec.Accounting[refKey] = acnt
				break
			}
		}
	}
	// sync Charges
	for _, ci := range ec.Charges {
		if refRatingID, has := sncedRatingIDs[ci.RatingID]; has {
			ci.RatingID = refRatingID
		}
		for _, cIcrmt := range ci.Increments {
			if refAcntID, has := sncedAcntIDs[cIcrmt.AccountingID]; has {
				cIcrmt.AccountingID = refAcntID
			}
		}
	}
}

// Merge will merge a list of EventCosts into this one
func (ec *EventCost) Merge(ecs ...*EventCost) {
	for _, newEC := range ecs {
		// updated AccountSummary information
		newEC.AccountSummary.UpdateInitialValue(ec.AccountSummary)
		ec.AccountSummary = newEC.AccountSummary
		for cIlIdx := range newEC.Charges {
			ec.appendChargingIntervalFromEventCost(newEC, cIlIdx)
		}
	}
	ec.ResetCounters()
}

// RemoveStaleReferences iterates through cached data and makes sure it is still referenced from Charging
func (ec *EventCost) RemoveStaleReferences() {
	// RatingIDs
	for key := range ec.Rating {
		var keyUsed bool
		for _, cIl := range ec.Charges {
			if cIl.RatingID == key {
				keyUsed = true
				break
			}
		}
		if !keyUsed { // look also in accounting for references
			for _, bc := range ec.Accounting {
				if bc.RatingID == key {
					keyUsed = true
					break
				}
			}
		}
		if !keyUsed { // not used, remove it
			delete(ec.Rating, key)
		}
	}
	for key := range ec.Accounting {
		var keyUsed bool
		for _, cIl := range ec.Charges {
			for _, cIt := range cIl.Increments {
				if cIt.AccountingID == key {
					keyUsed = true
					break
				}
			}
			if !keyUsed {
				for _, bCharge := range ec.Accounting {
					if bCharge.ExtraChargeID == key {
						keyUsed = true
						break
					}
				}
			}
		}
		if !keyUsed {
			delete(ec.Accounting, key)
		}
	}
	for key := range ec.RatingFilters {
		var keyUsed bool
		for _, ru := range ec.Rating {
			if ru.RatingFiltersID == key {
				keyUsed = true
				break
			}
		}
		if !keyUsed {
			delete(ec.RatingFilters, key)
		}
	}
	for key := range ec.Rates {
		var keyUsed bool
		for _, ru := range ec.Rating {
			if ru.RatesID == key {
				keyUsed = true
				break
			}
		}
		if !keyUsed {
			delete(ec.Rates, key)
		}
	}
	for key := range ec.Timings {
		var keyUsed bool
		for _, ru := range ec.Rating {
			if ru.TimingID == key {
				keyUsed = true
				break
			}

		}
		if !keyUsed {
			delete(ec.Rates, key)
		}
	}
}

// Trim will cut the EventCost at specific duration
// returns the srplusEC as separate EventCost
func (ec *EventCost) Trim(atUsage time.Duration) (srplusEC *EventCost, err error) {
	if ec.Usage == nil {
		ec.GetUsage()
	}
	origECUsage := ec.GetUsage()
	if atUsage >= *ec.Usage {
		return // no trim
	}
	if atUsage == 0 {
		//clone the event because we need to overwrite ec
		srplusEC = ec.Clone()
		// modify the value of ec
		*ec = *NewBareEventCost()
		ec.CGRID = srplusEC.CGRID
		ec.RunID = srplusEC.RunID
		ec.StartTime = srplusEC.StartTime
		ec.AccountSummary = srplusEC.AccountSummary.Clone()
		return // trim all, fresh EC with 0 usage
	}

	srplusEC = NewBareEventCost()
	srplusEC.CGRID = ec.CGRID
	srplusEC.RunID = ec.RunID
	srplusEC.StartTime = ec.StartTime
	srplusEC.AccountSummary = ec.AccountSummary.Clone()

	var lastActiveCIlIdx *int // mark last index which should stay with ec
	for i, cIl := range ec.Charges {
		if cIl.ecUsageIdx == nil {
			ec.ComputeEventCostUsageIndexes()
		}
		if cIl.usage == nil {
			ec.GetUsage()
		}
		if *cIl.ecUsageIdx+*cIl.TotalUsage() >= atUsage {
			lastActiveCIlIdx = utils.IntPointer(i)
			break
		}
	}
	if lastActiveCIlIdx == nil {
		return nil, errors.New("cannot find last active ChargingInterval")
	}
	lastActiveCIl := ec.Charges[*lastActiveCIlIdx]
	if *lastActiveCIl.ecUsageIdx >= atUsage {
		return nil, errors.New("failed detecting last active ChargingInterval")
	} else if lastActiveCIl.CompressFactor == 0 {
		return nil, errors.New("ChargingInterval with 0 compressFactor")
	}
	srplusEC.Charges = append(srplusEC.Charges, ec.Charges[*lastActiveCIlIdx+1:]...) // direct assignment will wrongly reference later
	ec.Charges = ec.Charges[:*lastActiveCIlIdx+1]
	ec.Usage = nil
	ec.Cost = nil
	if lastActiveCIl.CompressFactor != 1 &&
		*lastActiveCIl.ecUsageIdx+*lastActiveCIl.TotalUsage() > atUsage { // Split based on compress factor if needed
		var laCF int
		for ciCnt := 1; ciCnt <= lastActiveCIl.CompressFactor; ciCnt++ {
			if *lastActiveCIl.ecUsageIdx+
				time.Duration(lastActiveCIl.usage.Nanoseconds()*int64(ciCnt)) >= atUsage {
				laCF = ciCnt
				break
			}
		}
		if laCF == 0 {
			return nil, errors.New("cannot detect last active CompressFactor in ChargingInterval")
		}
		if laCF != lastActiveCIl.CompressFactor {
			srplsCIl := lastActiveCIl.Clone()
			srplsCIl.CompressFactor = lastActiveCIl.CompressFactor - laCF
			srplusEC.Charges = append([]*ChargingInterval{srplsCIl}, srplusEC.Charges...) // prepend surplus CIl
			lastActiveCIl.CompressFactor = laCF                                           // correct compress factor
			ec.Usage = nil
			ec.Cost = nil
		}
	}
	if atUsage != ec.GetUsage() { // lastInterval covering more than needed, need split
		atUsage -= (ec.GetUsage() - *lastActiveCIl.Usage()) // remaining duration to cover in increments of the last charging interval
		// find out last increment covering duration
		var lastActiveCItIdx *int
		var incrementsUsage time.Duration
		for i, cIt := range lastActiveCIl.Increments {
			incrementsUsage += cIt.TotalUsage()
			if incrementsUsage >= atUsage {
				lastActiveCItIdx = utils.IntPointer(i)
				break
			}
		}
		if lastActiveCItIdx == nil { // bug in increments
			return nil, errors.New("no active increment found")
		}
		lastActiveCIts := lastActiveCIl.Increments // so we can modify the reference in case we have surplus
		lastIncrement := lastActiveCIts[*lastActiveCItIdx]
		if lastIncrement.CompressFactor == 0 {
			return nil, errors.New("empty compress factor in increment")
		}
		var srplsIncrements []*ChargingIncrement
		if *lastActiveCItIdx < len(lastActiveCIl.Increments)-1 { // less that complete increments, have surplus
			srplsIncrements = lastActiveCIts[*lastActiveCItIdx+1:]
			lastActiveCIts = lastActiveCIts[:*lastActiveCItIdx+1]
			ec.Usage = nil
			ec.Cost = nil
		}
		var laItCF int
		if lastIncrement.CompressFactor != 1 && atUsage != incrementsUsage {
			// last increment compress factor is higher that we need to cover
			incrementsUsage -= lastIncrement.TotalUsage()
			for cnt := 1; cnt <= lastIncrement.CompressFactor; cnt++ {
				incrementsUsage += lastIncrement.Usage
				if incrementsUsage >= atUsage {
					laItCF = cnt
					break
				}
			}
			if laItCF == 0 {
				return nil, errors.New("cannot detect last active CompressFactor in ChargingIncrement")
			}
			if laItCF != lastIncrement.CompressFactor {
				srplsIncrement := lastIncrement.Clone()
				srplsIncrement.CompressFactor = srplsIncrement.CompressFactor - laItCF
				srplsIncrements = append([]*ChargingIncrement{srplsIncrement}, srplsIncrements...) // prepend the surplus out of compress
			}
		}
		if len(srplsIncrements) != 0 { // partially covering, need trim
			if lastActiveCIl.CompressFactor > 1 { // ChargingInterval not covering in full, need to split it
				lastActiveCIl.CompressFactor--
				ec.Charges = append(ec.Charges, lastActiveCIl.Clone())
				lastActiveCIl = ec.Charges[len(ec.Charges)-1]
				lastActiveCIl.CompressFactor = 1
				ec.Usage = nil
				ec.Cost = nil
			}
			srplsCIl := lastActiveCIl.Clone()
			srplsCIl.Increments = srplsIncrements
			srplusEC.Charges = append([]*ChargingInterval{srplsCIl}, srplusEC.Charges...)
			lastActiveCIl.Increments = make([]*ChargingIncrement, len(lastActiveCIts))
			for i, incr := range lastActiveCIts {
				lastActiveCIl.Increments[i] = incr.Clone() // avoid pointer references to the other interval
			}
			if laItCF != 0 {
				lastActiveCIl.Increments[len(lastActiveCIl.Increments)-1].CompressFactor = laItCF // correct the compressFactor for the last increment
				ec.Usage = nil
				ec.Cost = nil
			}
		}
	}
	ec.ResetCounters()
	if usage := ec.GetUsage(); usage < atUsage {
		return nil, errors.New("usage of EventCost smaller than requested")
	}
	srplusEC.ResetCounters()
	srplusEC.StartTime = ec.StartTime.Add(ec.GetUsage())
	if srplsUsage := srplusEC.GetUsage(); srplsUsage > origECUsage-atUsage {
		return nil, errors.New("surplus EventCost too big")
	}
	// close surplus with missing cache
	for _, cIl := range srplusEC.Charges {
		cIl.RatingID = srplusEC.ratingGetIDFromEventCost(ec, cIl.RatingID)
		for _, incr := range cIl.Increments {
			incr.AccountingID = srplusEC.accountingGetIDFromEventCost(ec, incr.AccountingID)
		}
	}
	ec.RemoveStaleReferences() // data should be transferred by now, can clean the old one
	return
}

// FieldAsInterface func to implement DataProvider
func (ec *EventCost) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if ec.cache == nil {
		ec.cache = utils.MapStorage{} // fix gob deserialization
	}
	if val, err = ec.cache.FieldAsInterface(fldPath); err != nil {
		if err != utils.ErrNotFound { // item found in cache
			return
		}
		err = nil // cancel previous err
	} else if val == nil {
		return nil, utils.ErrNotFound
	} else {
		return // data found in cache
	}
	val, err = ec.fieldAsInterface(fldPath)
	if err == nil {
		ec.cache.Set(fldPath, val)
	} else if err == utils.ErrNotFound {
		ec.cache.Set(fldPath, nil)
	}
	return
}

// fieldAsInterface the implementation of FieldAsInterface
func (ec *EventCost) fieldAsInterface(fldPath []string) (val interface{}, err error) {
	switch fldPath[0] {
	default: // "Charges[1]"
		opath, indx := utils.GetPathIndex(fldPath[0])
		if opath != utils.Charges {
			return nil, fmt.Errorf("unsupported field prefix: <%s>", opath)
		}
		if indx != nil {
			if len(ec.Charges) <= *indx {
				return nil, utils.ErrNotFound
			}
			return ec.getChargesForPath(fldPath[1:], ec.Charges[*indx])
		}
	case utils.Charges:
		if len(fldPath) != 1 { // slice has no members
			return nil, utils.ErrNotFound
		}
		return ec.Charges, nil
	case utils.CGRID:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return ec.CGRID, nil
	case utils.RunID:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return ec.RunID, nil
	case utils.StartTime:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return ec.StartTime, nil
	case utils.Usage:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		if ec.Usage == nil {
			return nil, nil
		}
		return *ec.Usage, nil
	case utils.Cost:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		if ec.Cost == nil {
			return nil, nil
		}
		return *ec.Cost, nil
	case utils.AccountSummary:
		if len(fldPath) == 1 {
			return ec.AccountSummary, nil
		}
		return ec.AccountSummary.FieldAsInterface(fldPath[1:])
	case utils.Timings:
		if len(fldPath) == 1 {
			return ec.Timings, nil
		}
		return ec.Timings.FieldAsInterface(fldPath[1:])
	case utils.Rates:
		if len(fldPath) == 1 {
			return ec.Rates, nil
		}
		return ec.Rates.FieldAsInterface(fldPath[1:])
	case utils.RatingFilters:
		if len(fldPath) == 1 {
			return ec.RatingFilters, nil
		}
		return ec.RatingFilters.FieldAsInterface(fldPath[1:])
	case utils.Accounting:
		if len(fldPath) == 1 {
			return ec.Accounting, nil
		}
		return ec.Accounting.FieldAsInterface(fldPath[1:])
	case utils.Rating:
		if len(fldPath) == 1 {
			return ec.Rating, nil
		}
		return ec.Rating.FieldAsInterface(fldPath[1:])
	}
	return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
}

func (ec *EventCost) getChargesForPath(fldPath []string, chr *ChargingInterval) (val interface{}, err error) {
	if chr == nil {
		return nil, utils.ErrNotFound
	}
	if len(fldPath) == 0 {
		return chr, nil
	}
	if fldPath[0] == utils.CompressFactor {
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return chr.CompressFactor, nil
	}
	if fldPath[0] == utils.Rating {
		return ec.getRatingForPath(fldPath[1:], ec.Rating[chr.RatingID])
	}
	opath, indx := utils.GetPathIndex(fldPath[0])
	if opath != utils.Increments {
		return nil, fmt.Errorf("unsupported field prefix: <%s>", opath)
	}
	if indx == nil {
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return chr.Increments, nil
	}
	if len(chr.Increments) <= *indx {
		return nil, utils.ErrNotFound
	}
	incr := chr.Increments[*indx]
	if len(fldPath) == 1 {
		return incr, nil
	}
	if fldPath[1] == utils.Accounting {
		return ec.getAcountingForPath(fldPath[2:], ec.Accounting[incr.AccountingID])
	}
	return incr.FieldAsInterface(fldPath)
}

func (ec *EventCost) getRatingForPath(fldPath []string, rating *RatingUnit) (val interface{}, err error) {
	if rating == nil {
		return nil, utils.ErrNotFound
	}
	if len(fldPath) == 0 {
		return rating, nil
	}

	switch fldPath[0] {
	default:
		opath, indx := utils.GetPathIndex(fldPath[0])
		if opath != utils.Rates {
			return nil, fmt.Errorf("unsupported field prefix: <%s>", opath)
		}
		rts, has := ec.Rates[rating.RatesID]
		if !has || rts == nil {
			return nil, utils.ErrNotFound
		}
		if indx != nil {
			if len(rts) <= *indx {
				return nil, utils.ErrNotFound
			}
			rt := rts[*indx]
			if len(fldPath) == 1 {
				return rt, nil
			}
			return rt.FieldAsInterface(fldPath[1:])
		}
	case utils.Rates:
		rts, has := ec.Rates[rating.RatesID]
		if !has || rts == nil {
			return nil, utils.ErrNotFound
		}
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound // no field on slice
		}
		return rts, nil
	case utils.Timing:
		tmg, has := ec.Timings[rating.TimingID]
		if !has || tmg == nil {
			return nil, utils.ErrNotFound
		}
		if len(fldPath) == 1 {
			return tmg, nil
		}
		return tmg.FieldAsInterface(fldPath[1:])
	case utils.RatingFilter:
		rtFltr, has := ec.RatingFilters[rating.RatingFiltersID]
		if !has || rtFltr == nil {
			return nil, utils.ErrNotFound
		}
		if len(fldPath) == 1 {
			return rtFltr, nil
		}
		return rtFltr.FieldAsInterface(fldPath[1:])
	}
	return rating.FieldAsInterface(fldPath)
}

func (ec *EventCost) getAcountingForPath(fldPath []string, bc *BalanceCharge) (val interface{}, err error) {
	if bc == nil {
		return nil, utils.ErrNotFound
	}
	if len(fldPath) == 0 {
		return bc, nil
	}

	if fldPath[0] == utils.BalanceField {
		bl := ec.AccountSummary.BalanceSummaries.BalanceSummaryWithUUD(bc.BalanceUUID)
		if bl == nil {
			return nil, utils.ErrNotFound
		}
		if len(fldPath) == 1 {
			return bl, nil
		}
		return bl.FieldAsInterface(fldPath[1:])

	}
	return bc.FieldAsInterface(fldPath)
}

// String to implement Dataprovider
func (ec *EventCost) String() string {
	return utils.ToJSON(ec)
}

// FieldAsString to implement Dataprovider
func (ec *EventCost) FieldAsString(fldPath []string) (string, error) {
	ival, err := ec.FieldAsInterface(fldPath)
	if err != nil {
		return utils.EmptyString, err
	}
	return utils.IfaceAsString(ival), nil
}
