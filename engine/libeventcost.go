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

// ChargingInterval represents one interval out of Usage providing charging info
// eg: PEAK vs OFFPEAK
type ChargingInterval struct {
	RatingID       string               // reference to RatingUnit
	Increments     []*ChargingIncrement // specific increments applied to this interval
	CompressFactor int
	usage          *time.Duration // cache usage computation for this interval
	ecUsageIdx     *time.Duration // computed value of totalUsage at the starting of the interval
	cost           *float64       // cache cost calculation on this interval
}

// PartiallyEquals does not compare CompressFactor, usefull for Merge
func (cIl *ChargingInterval) PartiallyEquals(oCIl *ChargingInterval) bool {
	if equals := cIl.RatingID == oCIl.RatingID &&
		len(cIl.Increments) == len(oCIl.Increments); !equals {
		return false
	}
	for i := range cIl.Increments {
		if !cIl.Increments[i].Equals(oCIl.Increments[i]) {
			return false
		}
	}
	return true
}

// Usage computes the total usage of this ChargingInterval, ignoring CompressFactor
func (cIl *ChargingInterval) Usage() *time.Duration {
	if cIl.usage == nil {
		var usage time.Duration
		for _, incr := range cIl.Increments {
			usage += incr.TotalUsage()
		}
		cIl.usage = &usage
	}
	return cIl.usage
}

// TotalUsage returns the total usage of this interval, considering compress factor
func (cIl *ChargingInterval) TotalUsage() (tu *time.Duration) {
	usage := cIl.Usage()
	if usage == nil {
		return
	}
	tu = new(time.Duration)
	*tu = time.Duration(usage.Nanoseconds() * int64(cIl.CompressFactor))
	return
}

// EventCostUsageIndex publishes the value of ecUsageIdx
func (cIl *ChargingInterval) EventCostUsageIndex() *time.Duration {
	return cIl.ecUsageIdx
}

// StartTime computes a StartTime based on EventCost.Start time and ecUsageIdx
func (cIl *ChargingInterval) StartTime(ecST time.Time) (st time.Time) {
	if cIl.ecUsageIdx != nil {
		st = ecST.Add(*cIl.ecUsageIdx)
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

func (cIl *ChargingInterval) TotalCost() float64 {
	return utils.Round((cIl.Cost() * float64(cIl.CompressFactor)),
		globalRoundingDecimals, utils.ROUNDING_MIDDLE)
}

// Clone returns a new instance of ChargingInterval with independent data
func (cIl *ChargingInterval) Clone() (cln *ChargingInterval) {
	cln = new(ChargingInterval)
	cln.RatingID = cIl.RatingID
	cln.CompressFactor = cIl.CompressFactor
	cln.Increments = make([]*ChargingIncrement, len(cIl.Increments))
	for i, cIt := range cIl.Increments {
		cln.Increments[i] = cIt.Clone()
	}
	return
}

// ChargingIncrement represents one unit charged inside an interval
type ChargingIncrement struct {
	Usage          time.Duration
	Cost           float64
	AccountingID   string
	CompressFactor int
}

func (cIt *ChargingIncrement) Equals(oCIt *ChargingIncrement) bool {
	return cIt.Usage == oCIt.Usage &&
		cIt.Cost == oCIt.Cost &&
		cIt.AccountingID == oCIt.AccountingID &&
		cIt.CompressFactor == oCIt.CompressFactor
}

// PartiallyEquals ignores the CompressFactor when comparing
func (cIt *ChargingIncrement) PartiallyEquals(oCIt *ChargingIncrement) bool {
	return cIt.Usage == oCIt.Usage &&
		cIt.Cost == oCIt.Cost &&
		cIt.AccountingID == oCIt.AccountingID
}

func (cIt *ChargingIncrement) Clone() (cln *ChargingIncrement) {
	cln = new(ChargingIncrement)
	*cln = *cIt
	return
}

// TotalUsage returns the total usage of the increment, considering compress factor
func (cIt *ChargingIncrement) TotalUsage() time.Duration {
	return time.Duration(cIt.Usage.Nanoseconds() * int64(cIt.CompressFactor))
}

func (cIt *ChargingIncrement) TotalCost() float64 {
	return cIt.Cost * float64(cIt.CompressFactor)
}

// BalanceCharge represents one unit charged to a balance
type BalanceCharge struct {
	AccountID     string  // keep reference for shared balances
	BalanceUUID   string  // balance charged
	RatingID      string  // special price applied on this balance
	Units         float64 // number of units charged
	ExtraChargeID string  // used in cases when paying *voice with *monetary
}

func (bc *BalanceCharge) Equals(oBC *BalanceCharge) bool {
	bcExtraChargeID := bc.ExtraChargeID
	if bcExtraChargeID == "" {
		bcExtraChargeID = utils.META_NONE
	}
	oBCExtraChargerID := oBC.ExtraChargeID
	if oBCExtraChargerID == "" { // so we can compare them properly
		oBCExtraChargerID = utils.META_NONE
	}
	return bc.AccountID == oBC.AccountID &&
		bc.BalanceUUID == oBC.BalanceUUID &&
		bc.RatingID == oBC.RatingID &&
		bc.Units == oBC.Units &&
		bcExtraChargeID == oBCExtraChargerID
}

func (bc *BalanceCharge) Clone() *BalanceCharge {
	clnBC := new(BalanceCharge)
	*clnBC = *bc
	return clnBC
}

type RatingMatchedFilters map[string]interface{}

func (rf RatingMatchedFilters) Equals(oRF RatingMatchedFilters) bool {
	for k := range rf {
		if rf[k] != oRF[k] {
			return false
		}
	}
	return true
}

func (rf RatingMatchedFilters) Clone() (cln map[string]interface{}) {
	cln = make(map[string]interface{})
	utils.Clone(rf, &cln)
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

func (ct *ChargedTiming) Clone() (cln *ChargedTiming) {
	cln = new(ChargedTiming)
	*cln = *ct
	return
}

// RatingUnit represents one unit out of RatingPlan matching for an event
type RatingUnit struct {
	ConnectFee       float64
	RoundingMethod   string
	RoundingDecimals int
	MaxCost          float64
	MaxCostStrategy  string
	TimingID         string // This RatingUnit is bounded to specific timing profile
	RatesID          string
	RatingFiltersID  string
}

func (ru *RatingUnit) Equals(oRU *RatingUnit) bool {
	return ru.ConnectFee == oRU.ConnectFee &&
		ru.RoundingMethod == oRU.RoundingMethod &&
		ru.RoundingDecimals == oRU.RoundingDecimals &&
		ru.MaxCost == oRU.MaxCost &&
		ru.MaxCostStrategy == oRU.MaxCostStrategy &&
		ru.TimingID == oRU.TimingID &&
		ru.RatesID == oRU.RatesID &&
		ru.RatingFiltersID == oRU.RatingFiltersID
}

func (ru *RatingUnit) Clone() (cln *RatingUnit) {
	cln = new(RatingUnit)
	*cln = *ru
	return
}

type RatingFilters map[string]RatingMatchedFilters // so we can define search methods

// GetWithSet attempts to retrieve the UUID of a matching data or create a new one
func (rfs RatingFilters) GetIDWithSet(rmf RatingMatchedFilters) string {
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

func (rfs RatingFilters) Clone() (cln RatingFilters) {
	cln = make(RatingFilters, len(rfs))
	for k, v := range rfs {
		cln[k] = v.Clone()
	}
	return
}

type Rating map[string]*RatingUnit

// GetIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (crus Rating) GetIDWithSet(cru *RatingUnit) string {
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

func (crus Rating) Clone() (cln Rating) {
	cln = make(Rating, len(crus))
	for k, v := range crus {
		cln[k] = v.Clone()
	}
	return
}

type ChargedRates map[string]RateGroups

// GetIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (crs ChargedRates) GetIDWithSet(rg RateGroups) string {
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

func (crs ChargedRates) Clone() (cln ChargedRates) {
	cln = make(ChargedRates, len(crs))
	for k, v := range crs {
		cln[k] = v.Clone()
	}
	return
}

type ChargedTimings map[string]*ChargedTiming

// GetIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (cts ChargedTimings) GetIDWithSet(ct *ChargedTiming) string {
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

func (cts ChargedTimings) Clone() (cln ChargedTimings) {
	cln = make(ChargedTimings, len(cts))
	for k, v := range cts {
		cln[k] = v.Clone()
	}
	return
}

type Accounting map[string]*BalanceCharge

// GetIDWithSet attempts to retrieve the UUID of a matching data or create a new one
func (cbs Accounting) GetIDWithSet(cb *BalanceCharge) string {
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

func (cbs Accounting) Clone() (cln Accounting) {
	cln = make(Accounting, len(cbs))
	for k, v := range cbs {
		cln[k] = v.Clone()
	}
	return
}
