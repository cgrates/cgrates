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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func NewBareEventCost() *EventCost {
	return &EventCost{
		Rating:        make(Rating),
		Accounting:    make(Accounting),
		RatingFilters: make(RatingFilters),
		Rates:         make(ChargedRates),
		Timings:       make(ChargedTimings),
		Charges:       make([]*ChargingInterval, 0),
	}
}

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
		rf := RatingMatchedFilters{"Subject": ts.MatchedSubject, "DestinationPrefix": ts.MatchedPrefix,
			"DestinationID": ts.MatchedDestId, "RatingPlanID": ts.RatingPlanId}
		cIl.RatingID = ec.ratingIDForRateInterval(ts.RateInterval, rf)
		if len(ts.Increments) != 0 {
			cIl.Increments = make([]*ChargingIncrement, len(ts.Increments))
		}
		for j, incr := range ts.Increments {
			cIt := &ChargingIncrement{
				Usage:          incr.Duration,
				Cost:           incr.Cost,
				CompressFactor: incr.CompressFactor}
			if incr.BalanceInfo == nil {
				continue
			}
			//AccountingID
			if incr.BalanceInfo.Unit != nil {
				// 2 balances work-around
				ecUUID := utils.META_NONE // populate no matter what due to Unit not nil
				if incr.BalanceInfo.Monetary != nil {
					if uuid := ec.Accounting.GetIDWithSet(
						&BalanceCharge{
							AccountID:   incr.BalanceInfo.AccountID,
							BalanceUUID: incr.BalanceInfo.Monetary.UUID,
							Units:       incr.Cost,
							RatingID:    ec.ratingIDForRateInterval(incr.BalanceInfo.Monetary.RateInterval, rf),
						}); uuid != "" {
						ecUUID = uuid
					}
				}
				cIt.AccountingID = ec.Accounting.GetIDWithSet(
					&BalanceCharge{
						AccountID:     incr.BalanceInfo.AccountID,
						BalanceUUID:   incr.BalanceInfo.Unit.UUID,
						Units:         incr.BalanceInfo.Unit.Consumed,
						RatingID:      ec.ratingIDForRateInterval(incr.BalanceInfo.Unit.RateInterval, rf),
						ExtraChargeID: ecUUID})
			} else if incr.BalanceInfo.Monetary != nil { // Only monetary
				cIt.AccountingID = ec.Accounting.GetIDWithSet(
					&BalanceCharge{
						AccountID:   incr.BalanceInfo.AccountID,
						BalanceUUID: incr.BalanceInfo.Monetary.UUID,
						Units:       incr.Cost,
						RatingID:    ec.ratingIDForRateInterval(incr.BalanceInfo.Monetary.RateInterval, rf)})
			}
			cIl.Increments[j] = cIt
		}
		ec.Charges[i] = cIl
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
}

func (ec *EventCost) ratingIDForRateInterval(ri *RateInterval, rf RatingMatchedFilters) string {
	if ri == nil || ri.Rating == nil {
		return ""
	}
	var rfUUID string
	if rf != nil {
		rfUUID = ec.RatingFilters.GetIDWithSet(rf)
	}
	var tmID string
	if ri.Timing != nil {
		tmID = ec.Timings.GetIDWithSet(
			&ChargedTiming{
				Years:     ri.Timing.Years,
				Months:    ri.Timing.Months,
				MonthDays: ri.Timing.MonthDays,
				WeekDays:  ri.Timing.WeekDays,
				StartTime: ri.Timing.StartTime})
	}
	var rtUUID string
	if len(ri.Rating.Rates) != 0 {
		rtUUID = ec.Rates.GetIDWithSet(ri.Rating.Rates)
	}
	return ec.Rating.GetIDWithSet(
		&RatingUnit{
			ConnectFee:       ri.Rating.ConnectFee,
			RoundingMethod:   ri.Rating.RoundingMethod,
			RoundingDecimals: ri.Rating.RoundingDecimals,
			MaxCost:          ri.Rating.MaxCost,
			MaxCostStrategy:  ri.Rating.MaxCostStrategy,
			TimingID:         tmID,
			RatesID:          rtUUID,
			RatingFiltersID:  rfUUID})
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
		ri.Timing = &RITiming{Years: cIlTm.Years, Months: cIlTm.Months, MonthDays: cIlTm.MonthDays,
			WeekDays: cIlTm.WeekDays, StartTime: cIlTm.StartTime}
	}
	return
}

func (ec *EventCost) Clone() (cln *EventCost) {
	if ec == nil {
		return
	}
	cln = new(EventCost)
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
	cln.Rating = ec.Rating.Clone()
	cln.Accounting = ec.Accounting.Clone()
	cln.RatingFilters = ec.RatingFilters.Clone()
	cln.Rates = ec.Rates.Clone()
	cln.Timings = ec.Timings.Clone()
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

// ComputeCost iterates through Charges, computing EventCost.Cost
func (ec *EventCost) GetCost() float64 {
	if ec.Cost == nil {
		var cost float64
		for _, ci := range ec.Charges {
			cost += ci.TotalCost()
		}
		cost = utils.Round(cost, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
		ec.Cost = &cost
	}
	return *ec.Cost
}

// ComputeUsage iterates through Charges, computing EventCost.Usage
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

func (ec *EventCost) AsCallCost() *CallCost {
	cc := &CallCost{
		Cost: ec.GetCost(), RatedUsage: float64(ec.GetUsage().Nanoseconds()),
		AccountSummary: ec.AccountSummary}
	cc.Timespans = make(TimeSpans, len(ec.Charges))
	for i, cIl := range ec.Charges {
		ts := &TimeSpan{Cost: cIl.Cost(),
			DurationIndex: *cIl.Usage(), CompressFactor: cIl.CompressFactor}
		if cIl.ecUsageIdx == nil { // index was not populated yet
			ec.ComputeEventCostUsageIndexes()
		}
		ts.TimeStart = ec.StartTime.Add(*cIl.ecUsageIdx)
		ts.TimeEnd = ts.TimeStart.Add(
			time.Duration(cIl.Usage().Nanoseconds() * int64(cIl.CompressFactor)))
		if cIl.RatingID != "" {
			if ec.Rating[cIl.RatingID].RatingFiltersID != "" {
				rfs := ec.RatingFilters[ec.Rating[cIl.RatingID].RatingFiltersID]
				ts.MatchedSubject = rfs["Subject"].(string)
				ts.MatchedPrefix = rfs["DestinationPrefix"].(string)
				ts.MatchedDestId = rfs["DestinationID"].(string)
				ts.RatingPlanId = rfs["RatingPlanID"].(string)
			}
		}
		ts.RateInterval = ec.rateIntervalForRatingID(cIl.RatingID)
		if len(cIl.Increments) != 0 {
			ts.Increments = make(Increments, len(cIl.Increments))
		}
		for j, cInc := range cIl.Increments {
			incr := &Increment{Duration: cInc.Usage, Cost: cInc.Cost, CompressFactor: cInc.CompressFactor, BalanceInfo: new(DebitInfo)}
			if cInc.AccountingID != "" {
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
				if utils.IsSliceMember([]string{utils.VOICE, utils.DATA}, balanceType) && cBC.ExtraChargeID == "" {
					cBC.ExtraChargeID = utils.META_NONE // mark the balance to be exported as Unit type
				}
				if cBC.ExtraChargeID != "" { // have both monetary and data
					// Work around, enforce logic with 2 balances for *voice/*monetary combination
					// so we can stay compatible with CallCost
					incr.BalanceInfo.Unit = &UnitInfo{UUID: cBC.BalanceUUID, Consumed: cBC.Units}
					incr.BalanceInfo.Unit.RateInterval = ec.rateIntervalForRatingID(cBC.RatingID)
					if cBC.ExtraChargeID != utils.META_NONE {
						cBC = ec.Accounting[cBC.ExtraChargeID] // overwrite original balance so we can process it in one place
					}
				}
				if cBC.ExtraChargeID != utils.META_NONE {
					incr.BalanceInfo.Monetary = &MonetaryInfo{UUID: cBC.BalanceUUID}
					incr.BalanceInfo.Monetary.RateInterval = ec.rateIntervalForRatingID(cBC.RatingID)
				}
			}
			ts.Increments[j] = incr
		}
		cc.Timespans[i] = ts
	}
	return cc
}

// ratingGetIDFomEventCost retrieves UUID based on data from another EventCost
func (ec *EventCost) ratingGetIDFomEventCost(oEC *EventCost, oRatingID string) string {
	if oRatingID == "" {
		return ""
	}
	oCIlRating := oEC.Rating[oRatingID].Clone() // clone so we don't influence the original data
	oCIlRating.TimingID = ec.Timings.GetIDWithSet(oEC.Timings[oCIlRating.TimingID])
	oCIlRating.RatingFiltersID = ec.RatingFilters.GetIDWithSet(oEC.RatingFilters[oCIlRating.RatingFiltersID])
	oCIlRating.RatesID = ec.Rates.GetIDWithSet(oEC.Rates[oCIlRating.RatesID])
	return ec.Rating.GetIDWithSet(oCIlRating)
}

// accountingGetIDFromEventCost retrieves UUID based on data from another EventCost
func (ec *EventCost) accountingGetIDFromEventCost(oEC *EventCost, oAccountingID string) string {
	if oAccountingID == "" || oAccountingID == utils.META_NONE {
		return ""
	}
	oBC := oEC.Accounting[oAccountingID].Clone()
	oBC.RatingID = ec.ratingGetIDFomEventCost(oEC, oBC.RatingID)
	oBC.ExtraChargeID = ec.accountingGetIDFromEventCost(oEC, oBC.ExtraChargeID)
	return ec.Accounting.GetIDWithSet(oBC)
}

// appendCIl appends a ChargingInterval to existing chargers
// no compression done at ChargingInterval level, attempted on ChargingIncrement level
func (ec *EventCost) appendCIlFromEC(oEC *EventCost, cIlIdx int) {
	cIl := oEC.Charges[cIlIdx]
	cIlCln := cIl.Clone() // add/modify data on clone instead of original
	cIlCln.RatingID = ec.ratingGetIDFomEventCost(oEC, cIl.RatingID)
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
			ec.Charges = append(ec.Charges, cIlCln)
		}
	}
}

// AppendChargingInterval appends or compresses a &ChargingInterval to existing ec.Chargers
func (ec *EventCost) appendChargingIntervalFromEventCost(oEC *EventCost, cIlIdx int) {
	lenChargers := len(ec.Charges)
	if lenChargers != 0 && ec.Charges[lenChargers-1].PartiallyEquals(oEC.Charges[cIlIdx]) {
		ec.Charges[lenChargers-1].CompressFactor += 1
	} else {
		ec.appendCIlFromEC(oEC, cIlIdx)
	}
}

// Merge will merge a list of EventCosts into this one
func (ec *EventCost) Merge(ecs ...*EventCost) {
	for _, newEC := range ecs {
		ec.AccountSummary = newEC.AccountSummary // updated AccountSummary information
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
		srplusEC = ec
		ec = NewBareEventCost()
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
				lastActiveCIl.CompressFactor -= 1
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
		cIl.RatingID = srplusEC.ratingGetIDFomEventCost(ec, cIl.RatingID)
		for _, incr := range cIl.Increments {
			incr.AccountingID = srplusEC.accountingGetIDFromEventCost(ec, incr.AccountingID)
		}
	}
	ec.RemoveStaleReferences() // data should be transferred by now, can clean the old one
	return
}
