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

type ChargedRatingUnits map[string]*RatingUnit

type ChargedRates map[string]RateGroups

type ChargedTimings map[string]*ChargedTiming

type ChargedBalances map[string]*BalanceCharge

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
		if cIl.RatingUUID != "" {
			cIlRU := ec.RatingUnits[cIl.RatingUUID]
			ri := new(RateInterval)
			ri.Rating = &RIRate{ConnectFee: cIlRU.ConnectFee,
				RoundingMethod: cIlRU.RoundingMethod, RoundingDecimals: cIlRU.RoundingDecimals,
				MaxCost: cIlRU.MaxCost, MaxCostStrategy: cIlRU.MaxCostStrategy}
			if cIlRU.RatesUUID != "" {
				ri.Rating.Rates = ec.Rates[cIlRU.RatesUUID]
			}
			if cIlRU.TimingUUID != "" {
				cIlTm := ec.Timings[cIlRU.TimingUUID]
				ri.Timing = &RITiming{Years: cIlTm.Years, Months: cIlTm.Months, MonthDays: cIlTm.MonthDays,
					WeekDays: cIlTm.WeekDays, StartTime: cIlTm.StartTime}
			}
			ts.RateInterval = ri
		}
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
					if cBC.RatingUUID != "" {
						cBCRU := ec.RatingUnits[cBC.RatingUUID]
						ri := new(RateInterval)
						ri.Rating = &RIRate{ConnectFee: cBCRU.ConnectFee,
							RoundingMethod: cBCRU.RoundingMethod, RoundingDecimals: cBCRU.RoundingDecimals,
							MaxCost: cBCRU.MaxCost, MaxCostStrategy: cBCRU.MaxCostStrategy}
						if cBCRU.RatesUUID != "" {
							ri.Rating.Rates = ec.Rates[cBCRU.RatesUUID]
						}
						incr.BalanceInfo.Unit.RateInterval = ri
					}
					cBC = ec.Balances[cBC.ExtraChargeUUID] // overwrite original balance so we can process it in one place
				}
				incr.BalanceInfo.Monetary = &MonetaryInfo{UUID: cBC.BalanceUUID}
				if cBC.RatingUUID != "" {
					cBCRU := ec.RatingUnits[cBC.RatingUUID]
					ri := new(RateInterval)
					ri.Rating = &RIRate{ConnectFee: cBCRU.ConnectFee,
						RoundingMethod: cBCRU.RoundingMethod, RoundingDecimals: cBCRU.RoundingDecimals,
						MaxCost: cBCRU.MaxCost, MaxCostStrategy: cBCRU.MaxCostStrategy}
					if cBCRU.RatesUUID != "" {
						ri.Rating.Rates = ec.Rates[cBCRU.RatesUUID]
					}
					incr.BalanceInfo.Monetary.RateInterval = ri
				}
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
