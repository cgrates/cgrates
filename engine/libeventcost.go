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
	"encoding/json"
	"fmt"
	"reflect"
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
func (cIl *ChargingInterval) Cost(ac Accounting) float64 {
	if cIl.cost == nil {
		var cost float64
		for _, incr := range cIl.Increments {
			if ac[incr.AccountingID] == nil || // ignore the rounding increment
				ac[incr.AccountingID].RatingID != utils.MetaRounding { // only used to justify the diference between the debited price and the final CDR cost
				cost += incr.Cost * float64(incr.CompressFactor)
			}
		}
		cost = utils.Round(cost, globalRoundingDecimals, utils.MetaRoundingMiddle)
		cIl.cost = &cost
	}
	return *cIl.cost
}

// TotalCost returns the cost of charges
func (cIl *ChargingInterval) TotalCost(ac Accounting) float64 {
	return utils.Round((cIl.Cost(ac) * float64(cIl.CompressFactor)),
		globalRoundingDecimals, utils.MetaRoundingMiddle)
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

// Equals returns if the structure has the same value
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

// Clone creates a copy of ChargingIncrement
func (cIt *ChargingIncrement) Clone() (cln *ChargingIncrement) {
	cln = new(ChargingIncrement)
	*cln = *cIt
	return
}

// TotalUsage returns the total usage of the increment, considering compress factor
func (cIt *ChargingIncrement) TotalUsage() time.Duration {
	return time.Duration(cIt.Usage.Nanoseconds() * int64(cIt.CompressFactor))
}

// TotalCost returns the cost of the increment
func (cIt *ChargingIncrement) TotalCost() float64 {
	return cIt.Cost * float64(cIt.CompressFactor)
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (cIt *ChargingIncrement) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.Usage:
		return cIt.Usage, nil
	case utils.Cost:
		return cIt.Cost, nil
	case utils.AccountingID:
		return cIt.AccountingID, nil
	case utils.CompressFactor:
		return cIt.CompressFactor, nil
	}
}

// BalanceCharge represents one unit charged to a balance
type BalanceCharge struct {
	AccountID     string  // keep reference for shared balances
	BalanceUUID   string  // balance charged
	RatingID      string  // special price applied on this balance
	Units         float64 // number of units charged
	ExtraChargeID string  // used in cases when paying *voice with *monetary
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (bc *BalanceCharge) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if bc == nil || len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.AccountID:
		return bc.AccountID, nil
	case utils.BalanceUUID:
		return bc.BalanceUUID, nil
	case utils.RatingID:
		return bc.RatingID, nil
	case utils.Units:
		return bc.Units, nil
	case utils.ExtraChargeID:
		return bc.ExtraChargeID, nil
	}
}

// Equals returns if the structure have the same fields
func (bc *BalanceCharge) Equals(oBC *BalanceCharge) bool {
	bcExtraChargeID := bc.ExtraChargeID
	if bcExtraChargeID == "" {
		bcExtraChargeID = utils.MetaNone
	}
	oBCExtraChargerID := oBC.ExtraChargeID
	if oBCExtraChargerID == "" { // so we can compare them properly
		oBCExtraChargerID = utils.MetaNone
	}
	return bc.AccountID == oBC.AccountID &&
		bc.BalanceUUID == oBC.BalanceUUID &&
		bc.RatingID == oBC.RatingID &&
		bc.Units == oBC.Units &&
		bcExtraChargeID == oBCExtraChargerID
}

// Clone creates a copy of BalanceCharge
func (bc *BalanceCharge) Clone() *BalanceCharge {
	clnBC := new(BalanceCharge)
	*clnBC = *bc
	return clnBC
}

// RatingMatchedFilters a rating filter
type RatingMatchedFilters map[string]interface{}

// Equals returns if the RatingMatchedFilters are equal
func (rf RatingMatchedFilters) Equals(oRF RatingMatchedFilters) bool {
	for k := range rf {
		if rf[k] != oRF[k] {
			return false
		}
	}
	return true
}

// Clone creates a copy of RatingMatchedFilters
func (rf RatingMatchedFilters) Clone() (cln map[string]interface{}) {
	if rf == nil {
		return nil
	}
	cln = make(map[string]interface{})
	for key, value := range rf {
		cln[key] = value
	}
	return
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (rf RatingMatchedFilters) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if rf == nil || len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	ct, has := rf[fldPath[0]]
	if !has || ct == nil {
		return nil, utils.ErrNotFound
	}
	return ct, nil
}

// ChargedTiming represents one timing attached to a charge
type ChargedTiming struct {
	Years     utils.Years
	Months    utils.Months
	MonthDays utils.MonthDays
	WeekDays  utils.WeekDays
	StartTime string
}

// Equals returns if the timings are equal
func (ct *ChargedTiming) Equals(oCT *ChargedTiming) bool {
	return ct.Years.Equals(oCT.Years) &&
		ct.Months.Equals(oCT.Months) &&
		ct.MonthDays.Equals(oCT.MonthDays) &&
		ct.WeekDays.Equals(oCT.WeekDays) &&
		ct.StartTime == oCT.StartTime
}

// Clone creates a copy of ChargedTiming
func (ct *ChargedTiming) Clone() (cln *ChargedTiming) {
	cln = new(ChargedTiming)
	*cln = *ct
	return
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (ct ChargedTiming) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.YearsFieldName:
		return ct.Years, nil
	case utils.MonthsFieldName:
		return ct.Months, nil
	case utils.MonthDaysFieldName:
		return ct.MonthDays, nil
	case utils.WeekDaysFieldName:
		return ct.WeekDays, nil
	case utils.StartTime:
		return ct.StartTime, nil
	}
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

// Equals returns if RatingUnit is equal to the other
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

// Clone creates a copy of RatingUnit
func (ru *RatingUnit) Clone() (cln *RatingUnit) {
	cln = new(RatingUnit)
	*cln = *ru
	return
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (ru RatingUnit) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.ConnectFee:
		return ru.ConnectFee, nil
	case utils.RoundingMethod:
		return ru.RoundingMethod, nil
	case utils.RoundingDecimals:
		return ru.RoundingDecimals, nil
	case utils.MaxCost:
		return ru.MaxCost, nil
	case utils.MaxCostStrategy:
		return ru.MaxCostStrategy, nil
	case utils.TimingID:
		return ru.TimingID, nil
	case utils.RatesID:
		return ru.RatesID, nil
	case utils.RatingFiltersID:
		return ru.RatingFiltersID, nil
	}
}

// RatingFilters the map of rating filters
type RatingFilters map[string]RatingMatchedFilters // so we can define search methods

// GetIDWithSet attempts to retrieve the UUID of a matching data or create a new one
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

// Clone creates a copy of RatingFilters
func (rfs RatingFilters) Clone() (cln RatingFilters) {
	cln = make(RatingFilters, len(rfs))
	for k, v := range rfs {
		cln[k] = v.Clone()
	}
	return
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (rfs RatingFilters) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if rfs == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	ct, has := rfs[fldPath[0]]
	if !has || ct == nil {
		return nil, utils.ErrNotFound
	}
	if len(fldPath) == 1 {
		return ct, nil
	}
	return ct.FieldAsInterface(fldPath[1:])
}

// Rating the map of rating units
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

// Clone creates a copy of Rating
func (crus Rating) Clone() (cln Rating) {
	cln = make(Rating, len(crus))
	for k, v := range crus {
		cln[k] = v.Clone()
	}
	return
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (crus Rating) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if crus == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	rt, has := crus[fldPath[0]]
	if !has || rt == nil {
		return nil, utils.ErrNotFound
	}
	if len(fldPath) == 1 {
		return rt, nil
	}
	return rt.FieldAsInterface(fldPath[1:])
}

// ChargedRates the map with rateGroups
type ChargedRates map[string]RateGroups

// FieldAsInterface func to help EventCost FieldAsInterface
func (crs ChargedRates) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if crs == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	opath, indx := utils.GetPathIndex(fldPath[0])
	cr, has := crs[opath]
	if !has || cr == nil {
		return nil, utils.ErrNotFound
	}
	if indx != nil {
		if len(cr) <= *indx {
			return nil, utils.ErrNotFound
		}
		rg := cr[*indx]
		if len(fldPath) == 1 {
			return rg, nil
		}
		return rg.FieldAsInterface(fldPath[1:])
	}
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	return cr, nil
}

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

// Clone creates a copy of ChargedRates
func (crs ChargedRates) Clone() (cln ChargedRates) {
	cln = make(ChargedRates, len(crs))
	for k, v := range crs {
		cln[k] = v.Clone()
	}
	return
}

// ChargedTimings the map of ChargedTiming
type ChargedTimings map[string]*ChargedTiming

// FieldAsInterface func to help EventCost FieldAsInterface
func (cts ChargedTimings) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if cts == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	ct, has := cts[fldPath[0]]
	if !has || ct == nil {
		return nil, utils.ErrNotFound
	}
	if len(fldPath) == 1 {
		return ct, nil
	}
	return ct.FieldAsInterface(fldPath[1:])
}

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

// Clone creates a copy of ChargedTimings
func (cts ChargedTimings) Clone() (cln ChargedTimings) {
	cln = make(ChargedTimings, len(cts))
	for k, v := range cts {
		cln[k] = v.Clone()
	}
	return
}

// Accounting the map of debited balances
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

// Clone creates a copy of Accounting
func (cbs Accounting) Clone() (cln Accounting) {
	cln = make(Accounting, len(cbs))
	for k, v := range cbs {
		cln[k] = v.Clone()
	}
	return
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (cbs Accounting) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if cbs == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	ac, has := cbs[fldPath[0]]
	if !has || ac == nil {
		return nil, utils.ErrNotFound
	}
	if len(fldPath) == 1 {
		return ac, nil
	}
	return ac.FieldAsInterface(fldPath[1:])
}

// IfaceAsEventCost converts an interface to EventCost
func IfaceAsEventCost(itm interface{}) (ec *EventCost, err error) {
	switch otm := itm.(type) {
	case nil:
	case *EventCost:
		ec = otm
	case string:
		var rawEC EventCost
		if errUnmarshal := json.Unmarshal([]byte(otm), &rawEC); errUnmarshal != nil {
			return nil, fmt.Errorf("JSON cannot unmarshal to *EventCost, err: %s", errUnmarshal.Error())
		}
		ec = &rawEC
	case map[string]interface{}:
		ec, err = IfaceAsEventCost(utils.ToJSON(otm))
	default:
		err = utils.ErrNotConvertibleTF(reflect.TypeOf(otm).String(), "*EventCost")
	}
	return
}

// NewFreeEventCost returns an EventCost of given duration that it's free
func NewFreeEventCost(cgrID, runID, account string, tStart time.Time, usage time.Duration) *EventCost {
	return &EventCost{
		CGRID:     cgrID,
		RunID:     runID,
		StartTime: tStart,
		Cost:      utils.Float64Pointer(0),
		Charges: []*ChargingInterval{{
			RatingID: utils.MetaPause,
			Increments: []*ChargingIncrement{
				{
					Usage:          usage,
					AccountingID:   utils.MetaPause,
					CompressFactor: 1,
				},
			},
			CompressFactor: 1,
		}},

		Rating: Rating{
			utils.MetaPause: {
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				RatesID:          utils.MetaPause,
				RatingFiltersID:  utils.MetaPause,
				TimingID:         utils.MetaPause,
			},
		},
		Accounting: Accounting{
			utils.MetaPause: {
				AccountID: account,
				// BalanceUUID: "",
				RatingID: utils.MetaPause,
			},
		},
		RatingFilters: RatingFilters{
			utils.MetaPause: {
				utils.Subject:               "",
				utils.DestinationPrefixName: "",
				utils.DestinationID:         "",
				utils.RatingPlanID:          utils.MetaPause,
			},
		},
		Rates: ChargedRates{
			utils.MetaPause: {
				{
					RateIncrement: 1,
					RateUnit:      1,
				},
			},
		},
		Timings: ChargedTimings{
			utils.MetaPause: {

				StartTime: "00:00:00",
			},
		},
		cache: utils.MapStorage{},
	}
}
