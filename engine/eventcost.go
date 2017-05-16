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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type ChargedIntervalDetails map[string]*ChrgIntervDetail // so we can define search methods

// GetWithSet attempts to retrieve the UUID of a matching data or create a new one
func (cids ChargedIntervalDetails) GetUUIDWithSet(cid *ChrgIntervDetail) string {
	for k, v := range cids {
		if v.Equals(cid) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.GenUUID()
	cids[uuid] = cid
	return uuid
}

type ChargedRatingUnits map[string]*RatingUnit

// GetUUIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (crus ChargedRatingUnits) GetUUIDWithSet(cru *RatingUnit) string {
	for k, v := range crus {
		if v.Equals(cru) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.GenUUID()
	crus[uuid] = cru
	return uuid
}

type ChargedRates map[string]RateGroups

// GetUUIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (crs ChargedRates) GetUUIDWithSet(rg RateGroups) string {
	for k, v := range crs {
		if v.Equals(rg) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.GenUUID()
	crs[uuid] = rg
	return uuid
}

type ChargedTimings map[string]*ChargedTiming

// GetUUIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (cts ChargedTimings) GetUUIDWithSet(ct *ChargedTiming) string {
	for k, v := range cts {
		if v.Equals(ct) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.GenUUID()
	cts[uuid] = ct
	return uuid
}

type ChargedBalances map[string]*BalanceCharge

// GetUUIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (cbs ChargedBalances) GetUUIDWithSet(cb *BalanceCharge) string {
	for k, v := range cbs {
		if v.Equals(cb) {
			return k
		}
	}
	// not found, set it here
	uuid := utils.GenUUID()
	cbs[uuid] = cb
	return uuid
}

func NewEventCostFromCallCost(cc *CallCost, cgrID, runID string) (ec *EventCost) {
	ec = &EventCost{CGRID: cgrID, RunID: runID,
		IntervalDetails: make(ChargedIntervalDetails),
		RatingUnits:     make(ChargedRatingUnits),
		Rates:           make(ChargedRates),
		Timings:         make(ChargedTimings),
		Balances:        make(ChargedBalances),
	}
	if len(cc.Timespans) != 0 {
		ec.Charges = make([]*ChargingInterval, len(cc.Timespans))
	}
	for i, ts := range cc.Timespans {
		cIl := &ChargingInterval{StartTime: ts.TimeStart, CompressFactor: ts.CompressFactor}
		cIl.IntervalDetailsUUID = ec.IntervalDetails.GetUUIDWithSet(
			&ChrgIntervDetail{Subject: ts.MatchedSubject, DestinationPrefix: ts.MatchedPrefix,
				DestinationID: ts.MatchedDestId, RatingPlanID: ts.RatingPlanId})
		cIl.RatingUUID = ec.ratingUUIDForRateInterval(ts.RateInterval)
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
					if uuid := ec.Balances.GetUUIDWithSet(
						&BalanceCharge{
							AccountID:   incr.BalanceInfo.AccountID,
							BalanceUUID: incr.BalanceInfo.Monetary.UUID,
							Units:       incr.Cost,
							RatingUUID:  ec.ratingUUIDForRateInterval(incr.BalanceInfo.Monetary.RateInterval),
						}); uuid != "" {
						ecUUID = uuid
					}
				}
				cIt.BalanceChargeUUID = ec.Balances.GetUUIDWithSet(
					&BalanceCharge{
						AccountID:       incr.BalanceInfo.AccountID,
						BalanceUUID:     incr.BalanceInfo.Unit.UUID,
						Units:           incr.BalanceInfo.Unit.Consumed,
						RatingUUID:      ec.ratingUUIDForRateInterval(incr.BalanceInfo.Unit.RateInterval),
						ExtraChargeUUID: ecUUID})
			} else { // Only monetary
				cIt.BalanceChargeUUID = ec.Balances.GetUUIDWithSet(
					&BalanceCharge{
						AccountID:   incr.BalanceInfo.AccountID,
						BalanceUUID: incr.BalanceInfo.Monetary.UUID,
						Units:       incr.Cost,
						RatingUUID:  ec.ratingUUIDForRateInterval(incr.BalanceInfo.Unit.RateInterval)})
			}
			cIl.Increments[j] = cIt
		}
		ec.Charges[i] = cIl
	}
	return
}

// EventCost stores cost for an Event
type EventCost struct {
	CGRID           string
	RunID           string
	Cost            float64
	Usage           time.Duration
	Charges         []*ChargingInterval
	IntervalDetails ChargedIntervalDetails
	RatingUnits     ChargedRatingUnits
	Rates           ChargedRates
	Timings         ChargedTimings
	Balances        ChargedBalances
	dirty           bool // mark need of recomputation of the Cost and Usage
}

func (ec *EventCost) ratingUUIDForRateInterval(ri *RateInterval) string {
	if ri == nil || ri.Rating == nil {
		return ""
	}
	var tmID string
	if ri.Timing != nil {
		tmID = ec.Timings.GetUUIDWithSet(
			&ChargedTiming{Years: ri.Timing.Years,
				Months:    ri.Timing.Months,
				MonthDays: ri.Timing.MonthDays,
				WeekDays:  ri.Timing.WeekDays,
				StartTime: ri.Timing.StartTime})
	}
	var rtUUID string
	if len(ri.Rating.Rates) != 0 {
		rtUUID = ec.Rates.GetUUIDWithSet(ri.Rating.Rates)
	}
	return ec.RatingUnits.GetUUIDWithSet(
		&RatingUnit{
			ConnectFee:       ri.Rating.ConnectFee,
			RoundingMethod:   ri.Rating.RoundingMethod,
			RoundingDecimals: ri.Rating.RoundingDecimals,
			MaxCost:          ri.Rating.MaxCost,
			MaxCostStrategy:  ri.Rating.MaxCostStrategy,
			TimingUUID:       tmID, RatesUUID: rtUUID})
}

func (ec *EventCost) rateIntervalForRatingUUID(ratingUUID string) (ri *RateInterval) {
	if ratingUUID == "" {
		return
	}
	cIlRU := ec.RatingUnits[ratingUUID]
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

func (ec *EventCost) AsCallCost(ToR, Tenant, Direction, Category, Account, Subject, Destination string) *CallCost {
	cc := &CallCost{Direction: Direction, Category: Category, Tenant: Tenant,
		Subject: Subject, Account: Account, Destination: Destination, TOR: ToR}
	cc.Timespans = make(TimeSpans, len(ec.Charges))
	for i, cIl := range ec.Charges {
		ts := &TimeSpan{TimeStart: cIl.StartTime, TimeEnd: cIl.StartTime.Add(cIl.Usage()),
			Cost: cIl.Cost(), DurationIndex: cIl.Usage(), CompressFactor: cIl.CompressFactor}
		if cIl.IntervalDetailsUUID != "" {
			iDtls := ec.IntervalDetails[cIl.IntervalDetailsUUID]
			ts.MatchedSubject = iDtls.Subject
			ts.MatchedPrefix = iDtls.DestinationPrefix
			ts.MatchedDestId = iDtls.DestinationID
			ts.RatingPlanId = iDtls.RatingPlanID
		}
		ts.RateInterval = ec.rateIntervalForRatingUUID(cIl.RatingUUID)
		if len(cIl.Increments) != 0 {
			ts.Increments = make(Increments, len(cIl.Increments))
		}
		for j, cInc := range cIl.Increments {
			incr := &Increment{Duration: cInc.Usage, Cost: cInc.Cost, CompressFactor: cInc.CompressFactor}
			if cInc.BalanceChargeUUID != "" {
				cBC := ec.Balances[cInc.BalanceChargeUUID]
				incr.BalanceInfo = &DebitInfo{AccountID: cBC.AccountID}
				if cBC.ExtraChargeUUID != "" { // have both monetary and data
					// Work around, enforce logic with 2 balances for *voice/*monetary combination
					// so we can stay compatible with CallCost
					incr.BalanceInfo.Unit = &UnitInfo{UUID: cBC.BalanceUUID, Consumed: cBC.Units}
					incr.BalanceInfo.Unit.RateInterval = ec.rateIntervalForRatingUUID(cBC.RatingUUID)
					cBC = ec.Balances[cBC.ExtraChargeUUID] // overwrite original balance so we can process it in one place
				}
				incr.BalanceInfo.Monetary = &MonetaryInfo{UUID: cBC.BalanceUUID}
				incr.BalanceInfo.Monetary.RateInterval = ec.rateIntervalForRatingUUID(cBC.RatingUUID)
			}
			ts.Increments[j] = incr
		}
		cc.Timespans[i] = ts
	}
	return cc
}

// ChargingInterval represents one interval out of Usage providing charging info
// eg: PEAK vs OFFPEAK
type ChargingInterval struct {
	StartTime           time.Time
	IntervalDetailsUUID string               // reference to CIntervDetails
	RatingUUID          string               // reference to RatingUnit
	Increments          []*ChargingIncrement // specific increments applied to this interval
	CompressFactor      int
	usage               *time.Duration // cache usage computation for this interval
	cost                *float64       // cache cost calculation on this interval
}

func (cIl *ChargingInterval) Equals(oCIl *ChargingInterval) (equals bool) {
	if equals = cIl.StartTime.Equal(oCIl.StartTime) &&
		cIl.IntervalDetailsUUID == oCIl.IntervalDetailsUUID &&
		cIl.RatingUUID == oCIl.RatingUUID &&
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
func (cIl *ChargingInterval) Usage() time.Duration {
	if cIl.usage == nil {
		var usage time.Duration
		for _, incr := range cIl.Increments {
			usage += time.Duration(incr.Usage.Nanoseconds() * int64(incr.CompressFactor))
		}
		cIl.usage = &usage
	}
	return *cIl.usage
}

// Cost computes the total cost on this ChargingInterval
func (cIl *ChargingInterval) Cost() float64 {
	if cIl.cost == nil {
		var cost float64
		for _, incr := range cIl.Increments {
			cost += incr.Cost * float64(incr.CompressFactor)
		}
		cost = utils.Round(cost, config.CgrConfig().RoundingDecimals, utils.ROUNDING_MIDDLE)
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
		cIt.BalanceChargeUUID == oCIt.BalanceChargeUUID
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

type ChrgIntervDetail struct {
	Subject           string // matched subject
	DestinationPrefix string // matched destination prefix
	DestinationID     string // matched destinationID
	RatingPlanID      string // matched ratingPlanID
}

func (cid *ChrgIntervDetail) Equals(oCID *ChrgIntervDetail) bool {
	return cid.Subject == oCID.Subject &&
		cid.DestinationPrefix == oCID.DestinationPrefix &&
		cid.DestinationID == oCID.DestinationID &&
		cid.RatingPlanID == oCID.RatingPlanID
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
	ConnectFee       float64
	RoundingMethod   string
	RoundingDecimals int
	MaxCost          float64
	MaxCostStrategy  string
	TimingUUID       string // This RatingUnit is bounded to specific timing profile
	RatesUUID        string
}

func (ru *RatingUnit) Equals(oRU *RatingUnit) bool {
	return ru.ConnectFee == oRU.ConnectFee &&
		ru.RoundingMethod == oRU.RoundingMethod &&
		ru.RoundingDecimals == oRU.RoundingDecimals &&
		ru.MaxCost == oRU.MaxCost &&
		ru.MaxCostStrategy == oRU.MaxCostStrategy &&
		ru.TimingUUID == oRU.TimingUUID &&
		ru.RatesUUID == oRU.RatesUUID
}
