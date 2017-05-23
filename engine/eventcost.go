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
	"time"

	"github.com/cgrates/cgrates/utils"
)

type RatingFilters map[string]RatingMatchedFilters // so we can define search methods

// GetWithSet attempts to retrieve the UUID of a matching data or create a new one
func (rfs RatingFilters) GetUUIDWithSet(rmf RatingMatchedFilters) string {
	if rmf == nil || len(rmf) == 0 {
		return ""
	}
	for k, v := range rfs {
		if v.Equals(rmf) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.UUIDSha1Prefix()
	rfs[uuid] = rmf
	return uuid
}

type Rating map[string]*RatingUnit

// GetUUIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (crus Rating) GetUUIDWithSet(cru *RatingUnit) string {
	if cru == nil {
		return ""
	}
	for k, v := range crus {
		if v.Equals(cru) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.UUIDSha1Prefix()
	crus[uuid] = cru
	return uuid
}

type ChargedRates map[string]RateGroups

// GetUUIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (crs ChargedRates) GetUUIDWithSet(rg RateGroups) string {
	if rg == nil || len(rg) == 0 {
		return ""
	}
	for k, v := range crs {
		if v.Equals(rg) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.UUIDSha1Prefix()
	crs[uuid] = rg
	return uuid
}

type ChargedTimings map[string]*ChargedTiming

// GetUUIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (cts ChargedTimings) GetUUIDWithSet(ct *ChargedTiming) string {
	if ct == nil {
		return ""
	}
	for k, v := range cts {
		if v.Equals(ct) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.UUIDSha1Prefix()
	cts[uuid] = ct
	return uuid
}

type Accounting map[string]*BalanceCharge

// GetUUIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (cbs Accounting) GetUUIDWithSet(cb *BalanceCharge) string {
	if cb == nil {
		return ""
	}
	for k, v := range cbs {
		if v.Equals(cb) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.UUIDSha1Prefix()
	cbs[uuid] = cb
	return uuid
}

func NewEventCostFromCallCost(cc *CallCost, cgrID, runID string) (ec *EventCost) {
	ec = &EventCost{CGRID: cgrID, RunID: runID,
		AccountSummary: cc.AccountSummary,
		RatingFilters:  make(RatingFilters),
		Rating:         make(Rating),
		Rates:          make(ChargedRates),
		Timings:        make(ChargedTimings),
		Accounting:     make(Accounting),
	}
	if len(cc.Timespans) != 0 {
		ec.Charges = make([]*ChargingInterval, len(cc.Timespans))
		ec.StartTime = cc.Timespans[0].TimeStart
	}
	for i, ts := range cc.Timespans {
		cIl := &ChargingInterval{CompressFactor: ts.CompressFactor}
		rf := RatingMatchedFilters{"Subject": ts.MatchedSubject, "DestinationPrefix": ts.MatchedPrefix,
			"DestinationID": ts.MatchedDestId, "RatingPlanID": ts.RatingPlanId}
		cIl.RatingUUID = ec.ratingUUIDForRateInterval(ts.RateInterval, rf)
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
			//BalanceChargeUUID
			if incr.BalanceInfo.Unit != nil {
				// 2 balances work-around
				ecUUID := utils.META_NONE // populate no matter what due to Unit not nil
				if incr.BalanceInfo.Monetary != nil {
					if uuid := ec.Accounting.GetUUIDWithSet(
						&BalanceCharge{
							AccountID:   incr.BalanceInfo.AccountID,
							BalanceUUID: incr.BalanceInfo.Monetary.UUID,
							Units:       incr.Cost,
							RatingUUID:  ec.ratingUUIDForRateInterval(incr.BalanceInfo.Monetary.RateInterval, rf),
						}); uuid != "" {
						ecUUID = uuid
					}
				}
				cIt.BalanceChargeUUID = ec.Accounting.GetUUIDWithSet(
					&BalanceCharge{
						AccountID:       incr.BalanceInfo.AccountID,
						BalanceUUID:     incr.BalanceInfo.Unit.UUID,
						Units:           incr.BalanceInfo.Unit.Consumed,
						RatingUUID:      ec.ratingUUIDForRateInterval(incr.BalanceInfo.Unit.RateInterval, rf),
						ExtraChargeUUID: ecUUID})
			} else if incr.BalanceInfo.Monetary != nil { // Only monetary
				cIt.BalanceChargeUUID = ec.Accounting.GetUUIDWithSet(
					&BalanceCharge{
						AccountID:   incr.BalanceInfo.AccountID,
						BalanceUUID: incr.BalanceInfo.Monetary.UUID,
						Units:       incr.Cost,
						RatingUUID:  ec.ratingUUIDForRateInterval(incr.BalanceInfo.Monetary.RateInterval, rf)})
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
	Cost           *float64 // pointer so we can nil it when dirty
	StartTime      time.Time
	Usage          *time.Duration
	Charges        []*ChargingInterval
	AccountSummary *AccountSummary // Account summary at the end of the event calculation
	Rating         Rating
	Accounting     Accounting
	RatingFilters  RatingFilters
	Rates          ChargedRates
	Timings        ChargedTimings
}

func (ec *EventCost) ratingUUIDForRateInterval(ri *RateInterval, rf RatingMatchedFilters) string {
	if ri == nil || ri.Rating == nil {
		return ""
	}
	var rfUUID string
	if rf != nil {
		rfUUID = ec.RatingFilters.GetUUIDWithSet(rf)
	}
	var tmID string
	if ri.Timing != nil {
		tmID = ec.Timings.GetUUIDWithSet(
			&ChargedTiming{
				Years:     ri.Timing.Years,
				Months:    ri.Timing.Months,
				MonthDays: ri.Timing.MonthDays,
				WeekDays:  ri.Timing.WeekDays,
				StartTime: ri.Timing.StartTime})
	}
	var rtUUID string
	if len(ri.Rating.Rates) != 0 {
		rtUUID = ec.Rates.GetUUIDWithSet(ri.Rating.Rates)
	}
	return ec.Rating.GetUUIDWithSet(
		&RatingUnit{
			ConnectFee:        ri.Rating.ConnectFee,
			RoundingMethod:    ri.Rating.RoundingMethod,
			RoundingDecimals:  ri.Rating.RoundingDecimals,
			MaxCost:           ri.Rating.MaxCost,
			MaxCostStrategy:   ri.Rating.MaxCostStrategy,
			TimingUUID:        tmID,
			RatesUUID:         rtUUID,
			RatingFiltersUUID: rfUUID})
}

func (ec *EventCost) rateIntervalForRatingUUID(ratingUUID string) (ri *RateInterval) {
	if ratingUUID == "" {
		return
	}
	cIlRU := ec.Rating[ratingUUID]
	ri = new(RateInterval)
	ri.Rating = &RIRate{ConnectFee: cIlRU.ConnectFee,
		RoundingMethod:   cIlRU.RoundingMethod,
		RoundingDecimals: cIlRU.RoundingDecimals,
		MaxCost:          cIlRU.MaxCost, MaxCostStrategy: cIlRU.MaxCostStrategy}
	if cIlRU.RatesUUID != "" {
		ri.Rating.Rates = ec.Rates[cIlRU.RatesUUID]
	}
	if cIlRU.TimingUUID != "" {
		cIlTm := ec.Timings[cIlRU.TimingUUID]
		ri.Timing = &RITiming{Years: cIlTm.Years, Months: cIlTm.Months, MonthDays: cIlTm.MonthDays,
			WeekDays: cIlTm.WeekDays, StartTime: cIlTm.StartTime}
	}
	return
}

// Compute aggregates all the compute methods on EventCost
func (ec *EventCost) Compute() {
	ec.ComputeUsage()
	ec.ComputeUsageIndexes()
	ec.ComputeCost()
}

// ResetCounters will reset all the computed cached values
func (ec *EventCost) ResetCounters() {
	ec.Cost = nil
	ec.Usage = nil
	for _, cIl := range ec.Charges {
		cIl.cost = nil
		cIl.usage = nil
		cIl.totalUsageIndex = nil
	}
}

// ComputeCost iterates through Charges, computing EventCost.Cost
func (ec *EventCost) ComputeCost() float64 {
	if ec.Cost == nil {
		var cost float64
		for _, ci := range ec.Charges {
			cost += ci.Cost() * float64(ci.CompressFactor)
		}
		cost = utils.Round(cost, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
		ec.Cost = &cost
	}
	return *ec.Cost
}

// ComputeUsage iterates through Charges, computing EventCost.Usage
func (ec *EventCost) ComputeUsage() time.Duration {
	if ec.Usage == nil {
		var usage time.Duration
		for _, ci := range ec.Charges {
			usage += time.Duration(ci.Usage().Nanoseconds() * int64(ci.CompressFactor))
		}
		ec.Usage = &usage
	}
	return *ec.Usage
}

// ComputeUsageIndexes will iterate through Chargers and populate their totalUsageIndex
func (ec *EventCost) ComputeUsageIndexes() {
	var totalUsage time.Duration
	for _, cIl := range ec.Charges {
		if cIl.totalUsageIndex == nil {
			cIl.totalUsageIndex = utils.DurationPointer(totalUsage)
		}
		totalUsage += time.Duration(cIl.Usage().Nanoseconds() * int64(cIl.CompressFactor))
	}
}

func (ec *EventCost) AsCallCost() *CallCost {
	cc := &CallCost{
		Cost: ec.ComputeCost(), RatedUsage: ec.ComputeUsage().Seconds(),
		AccountSummary: ec.AccountSummary}
	cc.Timespans = make(TimeSpans, len(ec.Charges))
	for i, cIl := range ec.Charges {
		ts := &TimeSpan{Cost: cIl.Cost(),
			DurationIndex: *cIl.Usage(), CompressFactor: cIl.CompressFactor}
		if cIl.totalUsageIndex == nil { // index was not populated yet
			ec.ComputeUsageIndexes()
		}
		ts.TimeStart = ec.StartTime.Add(*cIl.totalUsageIndex)
		ts.TimeEnd = ts.TimeStart.Add(
			time.Duration(cIl.Usage().Nanoseconds() * int64(cIl.CompressFactor)))
		if cIl.RatingUUID != "" {
			if ec.Rating[cIl.RatingUUID].RatingFiltersUUID != "" {
				rfs := ec.RatingFilters[ec.Rating[cIl.RatingUUID].RatingFiltersUUID]
				ts.MatchedSubject = rfs["Subject"].(string)
				ts.MatchedPrefix = rfs["DestinationPrefix"].(string)
				ts.MatchedDestId = rfs["DestinationID"].(string)
				ts.RatingPlanId = rfs["RatingPlanID"].(string)
			}
		}
		ts.RateInterval = ec.rateIntervalForRatingUUID(cIl.RatingUUID)
		if len(cIl.Increments) != 0 {
			ts.Increments = make(Increments, len(cIl.Increments))
		}
		for j, cInc := range cIl.Increments {
			incr := &Increment{Duration: cInc.Usage, Cost: cInc.Cost, CompressFactor: cInc.CompressFactor}
			if cInc.BalanceChargeUUID != "" {
				cBC := ec.Accounting[cInc.BalanceChargeUUID]
				incr.BalanceInfo = &DebitInfo{AccountID: cBC.AccountID}
				if cBC.ExtraChargeUUID != "" { // have both monetary and data
					// Work around, enforce logic with 2 balances for *voice/*monetary combination
					// so we can stay compatible with CallCost
					incr.BalanceInfo.Unit = &UnitInfo{UUID: cBC.BalanceUUID, Consumed: cBC.Units}
					incr.BalanceInfo.Unit.RateInterval = ec.rateIntervalForRatingUUID(cBC.RatingUUID)
					if cBC.ExtraChargeUUID != utils.META_NONE {
						cBC = ec.Accounting[cBC.ExtraChargeUUID] // overwrite original balance so we can process it in one place
					}
				}
				if cBC.ExtraChargeUUID != utils.META_NONE {
					incr.BalanceInfo.Monetary = &MonetaryInfo{UUID: cBC.BalanceUUID}
					incr.BalanceInfo.Monetary.RateInterval = ec.rateIntervalForRatingUUID(cBC.RatingUUID)
				}
			}
			ts.Increments[j] = incr
		}
		cc.Timespans[i] = ts
	}
	return cc
}

// ratingGetUUIDFomEventCost retrieves UUID based on data from another EventCost
func (ec *EventCost) ratingGetUUIDFomEventCost(oEC *EventCost, oRatingUUID string) string {
	oCIlRating := oEC.Rating[oRatingUUID].Clone() // clone so we don't influence the original data
	oCIlRating.TimingUUID = ec.Timings.GetUUIDWithSet(oEC.Timings[oCIlRating.TimingUUID])
	oCIlRating.RatingFiltersUUID = ec.RatingFilters.GetUUIDWithSet(oEC.RatingFilters[oCIlRating.RatingFiltersUUID])
	oCIlRating.RatesUUID = ec.Rates.GetUUIDWithSet(oEC.Rates[oCIlRating.RatesUUID])
	return ec.Rating.GetUUIDWithSet(oCIlRating)
}

// accountingGetUUIDFromEventCost retrieves UUID based on data from another EventCost
func (ec *EventCost) accountingGetUUIDFromEventCost(oEC *EventCost, oBalanceChargeUUID string) string {
	oBC := oEC.Accounting[oBalanceChargeUUID].Clone()
	oBC.RatingUUID = ec.ratingGetUUIDFomEventCost(oEC, oBC.RatingUUID)
	if oBC.ExtraChargeUUID != "" {
		oBC.ExtraChargeUUID = ec.accountingGetUUIDFromEventCost(oEC, oBC.ExtraChargeUUID)
	}
	return ec.Accounting.GetUUIDWithSet(oBC)
}

// appendCIl appends a ChargingInterval to existing chargers, no compression done
func (ec *EventCost) appendCIlFromEC(oEC *EventCost, cIlIdx int) {
	cIl := oEC.Charges[cIlIdx]
	cIl.RatingUUID = ec.ratingGetUUIDFomEventCost(oEC, cIl.RatingUUID)
	for _, cIt := range cIl.Increments {
		cIt.BalanceChargeUUID = ec.accountingGetUUIDFromEventCost(oEC, cIt.BalanceChargeUUID)
	}
	ec.Charges = append(ec.Charges, cIl)
}

// AppendChargingInterval appends or compresses a &ChargingInterval to existing ec.Chargers
func (ec *EventCost) AppendChargingIntervalFromEventCost(oEC *EventCost, cIlIdx int) {
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
			ec.AppendChargingIntervalFromEventCost(newEC, cIlIdx)
		}
	}
	ec.Usage = nil // Reset them
	ec.Cost = nil
}

/*
// Cut will cut the EventCost on specifiedTime at ChargingIncrement level, returning the surplus
func (ec *EventCost) Trim(atTime time.Time) (surplus *EventCost) {
	var limitIndex int
	for i, cIl := range ec.Charges {
		if cIl.StartTime >
	}
}
*/

// ChargingInterval represents one interval out of Usage providing charging info
// eg: PEAK vs OFFPEAK
type ChargingInterval struct {
	RatingUUID      string               // reference to RatingUnit
	Increments      []*ChargingIncrement // specific increments applied to this interval
	CompressFactor  int
	usage           *time.Duration // cache usage computation for this interval
	totalUsageIndex *time.Duration // computed value of totalUsage at the starting of the interval
	cost            *float64       // cache cost calculation on this interval

}

// PartiallyEquals does not compare CompressFactor, usefull for Merge
func (cIl *ChargingInterval) PartiallyEquals(oCIl *ChargingInterval) (equals bool) {
	if equals = cIl.RatingUUID == oCIl.RatingUUID &&
		len(cIl.Increments) == len(oCIl.Increments); !equals {
		return
	}
	for i := range cIl.Increments {
		if !cIl.Increments[i].Equals(oCIl.Increments[i]) {
			equals = false
			break
		}
	}
	return
}

// Usage computes the total usage of this ChargingInterval, ignoring CompressFactor
func (cIl *ChargingInterval) Usage() *time.Duration {
	if cIl.usage == nil {
		var usage time.Duration
		for _, incr := range cIl.Increments {
			usage += time.Duration(incr.Usage.Nanoseconds() * int64(incr.CompressFactor))
		}
		cIl.usage = &usage
	}
	return cIl.usage
}

// TotalUsageIndex publishes the value of totalUsageIndex
func (cIl *ChargingInterval) TotalUsageIndex() *time.Duration {
	return cIl.totalUsageIndex
}

// StartTime computes a StartTime based on EventCost.Start time and totalUsageIndex
func (cIl *ChargingInterval) StartTime(ecST time.Time) (st time.Time) {
	if cIl.totalUsageIndex != nil {
		st = ecST.Add(*cIl.totalUsageIndex)
	}
	return
}

// EndTime computes an EndTime based on ChargingInterval StartTime value and usage
func (cIl *ChargingInterval) EndTime(cIlST time.Time) (et time.Time) {
	return cIlST.Add(time.Duration(cIl.Usage().Nanoseconds() * int64(cIl.CompressFactor)))
}

// Cost computes the total cost on this ChargingInterval
func (cIl *ChargingInterval) Cost() float64 {
	if cIl.cost == nil {
		var cost float64
		for _, incr := range cIl.Increments {
			cost += incr.Cost * float64(incr.CompressFactor)
		}
		cost = utils.Round(cost, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
		cIl.cost = &cost
	}
	return *cIl.cost
}

// ChargingIncrement represents one unit charged inside an interval
type ChargingIncrement struct {
	Usage             time.Duration
	Cost              float64
	BalanceChargeUUID string
	CompressFactor    int
}

func (cIt *ChargingIncrement) Equals(oCIt *ChargingIncrement) bool {
	return cIt.Usage == oCIt.Usage &&
		cIt.Cost == oCIt.Cost &&
		cIt.BalanceChargeUUID == oCIt.BalanceChargeUUID &&
		cIt.CompressFactor == oCIt.CompressFactor
}

// BalanceCharge represents one unit charged to a balance
type BalanceCharge struct {
	AccountID       string  // keep reference for shared balances
	BalanceUUID     string  // balance charged
	RatingUUID      string  // special price applied on this balance
	Units           float64 // number of units charged
	ExtraChargeUUID string  // used in cases when paying *voice with *monetary
}

func (bc *BalanceCharge) Equals(oBC *BalanceCharge) bool {
	return bc.AccountID == oBC.AccountID &&
		bc.BalanceUUID == oBC.BalanceUUID &&
		bc.RatingUUID == oBC.RatingUUID &&
		bc.Units == oBC.Units &&
		bc.ExtraChargeUUID == oBC.ExtraChargeUUID
}

func (bc *BalanceCharge) Clone() *BalanceCharge {
	clnBC := new(BalanceCharge)
	*clnBC = *bc
	return clnBC
}

type RatingMatchedFilters map[string]interface{}

func (rf RatingMatchedFilters) Equals(oRF RatingMatchedFilters) (equals bool) {
	equals = true
	for k := range rf {
		if rf[k] != oRF[k] {
			equals = false
			break
		}
	}
	return
}

// ChargedTiming represents one timing attached to a charge
type ChargedTiming struct {
	Years     utils.Years
	Months    utils.Months
	MonthDays utils.MonthDays
	WeekDays  utils.WeekDays
	StartTime string
}

func (ct *ChargedTiming) Equals(oCT *ChargedTiming) bool {
	return ct.Years.Equals(oCT.Years) &&
		ct.Months.Equals(oCT.Months) &&
		ct.MonthDays.Equals(oCT.MonthDays) &&
		ct.WeekDays.Equals(oCT.WeekDays) &&
		ct.StartTime == oCT.StartTime
}

// RatingUnit represents one unit out of RatingPlan matching for an event
type RatingUnit struct {
	ConnectFee        float64
	RoundingMethod    string
	RoundingDecimals  int
	MaxCost           float64
	MaxCostStrategy   string
	TimingUUID        string // This RatingUnit is bounded to specific timing profile
	RatesUUID         string
	RatingFiltersUUID string
}

func (ru *RatingUnit) Equals(oRU *RatingUnit) bool {
	return ru.ConnectFee == oRU.ConnectFee &&
		ru.RoundingMethod == oRU.RoundingMethod &&
		ru.RoundingDecimals == oRU.RoundingDecimals &&
		ru.MaxCost == oRU.MaxCost &&
		ru.MaxCostStrategy == oRU.MaxCostStrategy &&
		ru.TimingUUID == oRU.TimingUUID &&
		ru.RatesUUID == oRU.RatesUUID &&
		ru.RatingFiltersUUID == oRU.RatingFiltersUUID
}

func (ru *RatingUnit) Clone() *RatingUnit {
	clnRU := new(RatingUnit)
	*clnRU = *ru
	return clnRU
}
