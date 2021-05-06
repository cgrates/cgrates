/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package utils

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cron"
	"github.com/ericlagergren/decimal"
)

// RateProfile represents the configuration of a Rate profile
type RateProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weights            DynamicWeights
	MinCost            *Decimal
	MaxCost            *Decimal
	MaxCostStrategy    string
	Rates              map[string]*Rate
}

func (rp *RateProfile) TenantID() string {
	return ConcatenatedKey(rp.Tenant, rp.ID)
}

func (rp *RateProfile) Compile() (err error) {
	for _, rtP := range rp.Rates {
		rtP.uID = ConcatenatedKey(rp.Tenant, rp.ID, rtP.ID)
		if err = rtP.Compile(); err != nil {
			return
		}
	}
	return
}

// Rate defines rate related information used within a RateProfile
type Rate struct {
	ID              string         // RateID
	FilterIDs       []string       // RateFilterIDs
	ActivationTimes string         // ActivationTimes is a cron formatted time interval
	Weights         DynamicWeights // RateWeight will decide the winner per interval start
	Blocker         bool           // RateBlocker will make this rate recurrent, deactivating further intervals
	IntervalRates   []*IntervalRate

	sched cron.Schedule // compiled version of activation times as cron.Schedule interface
	uID   string
}

type ExtRate struct {
	ID              string         // RateID
	FilterIDs       []string       // RateFilterIDs
	ActivationTimes string         // ActivationTimes is a cron formatted time interval
	Weights         DynamicWeights // RateWeight will decide the winner per interval start
	Blocker         bool           // RateBlocker will make this rate recurrent, deactivating further intervals
	IntervalRates   []*ExtIntervalRate

	sched cron.Schedule // compiled version of activation times as cron.Schedule interface
	uID   string
}

// Equals returns the equality between two ExtRate
func (eRt *ExtRate) Equals(extRT *ExtRate) (eq bool) {
	if (eRt.ID != extRT.ID ||
		eRt.ActivationTimes != extRT.ActivationTimes ||
		eRt.Blocker != extRT.Blocker) ||
		(eRt.FilterIDs == nil && extRT.FilterIDs != nil ||
			eRt.FilterIDs != nil && extRT.FilterIDs == nil) ||
		(eRt.Weights == nil && extRT.Weights != nil ||
			eRt.Weights != nil && extRT.Weights == nil ||
			len(eRt.Weights) != len(extRT.Weights)) ||
		(eRt.IntervalRates == nil && extRT.IntervalRates != nil ||
			eRt.IntervalRates != nil && extRT.IntervalRates == nil ||
			len(eRt.IntervalRates) != len(extRT.IntervalRates)) {
		return
	}
	for idx, val := range eRt.FilterIDs {
		if val != extRT.FilterIDs[idx] {
			return
		}
	}
	if eRt.Weights != nil && extRT.Weights != nil {
		for idx, val := range eRt.Weights {
			if ok := val.Equals(extRT.Weights[idx]); !ok {
				return
			}
		}
	}
	if eRt.IntervalRates != nil && extRT.IntervalRates != nil {
		for idx, val := range eRt.IntervalRates {
			if ok := val.Equals(extRT.IntervalRates[idx]); !ok {
				return
			}
		}
	}
	return true
}

// Equals returns the equality between two Rate
func (rT *Rate) Equals(rte *Rate) (eq bool) {
	if (rT.ID != rte.ID ||
		rT.ActivationTimes != rte.ActivationTimes ||
		rT.Blocker != rte.Blocker) ||
		(rT.FilterIDs == nil && rte.FilterIDs != nil ||
			rT.FilterIDs != nil && rte.FilterIDs == nil) ||
		(rT.Weights == nil && rte.Weights != nil ||
			rT.Weights != nil && rte.Weights == nil ||
			len(rT.Weights) != len(rte.Weights)) ||
		(rT.IntervalRates == nil && rte.IntervalRates != nil ||
			rT.IntervalRates != nil && rte.IntervalRates == nil ||
			len(rT.IntervalRates) != len(rte.IntervalRates)) {
		return
	}
	for idx, val := range rT.FilterIDs {
		if val != rte.FilterIDs[idx] {
			return
		}
	}
	if rT.Weights != nil && rte.Weights != nil {
		for idx, val := range rT.Weights {
			if ok := val.Equals(rte.Weights[idx]); !ok {
				return
			}
		}
	}
	if rT.IntervalRates != nil && rte.IntervalRates != nil {
		for idx, val := range rT.IntervalRates {
			if ok := val.Equals(rte.IntervalRates[idx]); !ok {
				return
			}
		}
	}
	return true
}

// AsExtRate converts Rate to ExtRate
func (rT *Rate) AsExtRate() (eRt *ExtRate, err error) {
	eRt = &ExtRate{
		ID:              rT.ID,
		ActivationTimes: rT.ActivationTimes,
		sched:           rT.sched,
		uID:             rT.uID,
		Blocker:         rT.Blocker,
	}
	if rT.FilterIDs != nil {
		eRt.FilterIDs = make([]string, len(rT.FilterIDs))
		for idx, val := range rT.FilterIDs {
			eRt.FilterIDs[idx] = val
		}
	}
	if rT.Weights != nil {
		eRt.Weights = rT.Weights
	}
	if rT.IntervalRates != nil {
		eRt.IntervalRates = make([]*ExtIntervalRate, len(rT.IntervalRates))
		for idx, val := range rT.IntervalRates {
			if rcvIntv, err := val.AsExtIntervalRate(); err != nil {
				return nil, err
			} else {
				eRt.IntervalRates[idx] = rcvIntv
			}
		}
	}
	return
}

// UID returns system wide unique identifier
func (rt *Rate) UID() string {
	return rt.uID
}

type IntervalRate struct {
	IntervalStart *Decimal // Starting point when the Rate kicks in
	FixedFee      *Decimal
	RecurrentFee  *Decimal
	Unit          *Decimal // RateUnit
	Increment     *Decimal // RateIncrement
}

type ExtIntervalRate struct {
	IntervalStart *float64 // Starting point when the Rate kicks in
	FixedFee      *float64
	RecurrentFee  *float64
	Unit          *float64 // RateUnit
	Increment     *float64 // RateIncrement
}

// AsExtIntervalRate converts IntervalRate to ExtIntervalRate
func (iR *IntervalRate) AsExtIntervalRate() (eIr *ExtIntervalRate, err error) {
	eIr = new(ExtIntervalRate)
	if iR.IntervalStart != nil {
		if fltIntSt, ok := iR.IntervalStart.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal IntervalStart to float64")
		} else {
			eIr.IntervalStart = &fltIntSt
		}
	}
	if iR.FixedFee != nil {
		if fltFxdFee, ok := iR.FixedFee.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal FixedFee to float64")
		} else {
			eIr.FixedFee = &fltFxdFee
		}
	}
	if iR.RecurrentFee != nil {
		if fltRecFee, ok := iR.RecurrentFee.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal RecurrentFee to float64")
		} else {
			eIr.RecurrentFee = &fltRecFee
		}
	}
	if iR.Unit != nil {
		if fltUnit, ok := iR.Unit.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Unit to float64")
		} else {
			eIr.Unit = &fltUnit
		}
	}
	if iR.Increment != nil {
		if fltIncr, ok := iR.Increment.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Increment to float64")
		} else {
			eIr.Increment = &fltIncr
		}
	}
	return
}

// Equals returns the equality between two ExtIntervalRate
func (eIr *ExtIntervalRate) Equals(extIr *ExtIntervalRate) (eq bool) {
	if !((eIr.IntervalStart == nil && extIr.IntervalStart == nil) ||
		(eIr.IntervalStart != nil && extIr.IntervalStart != nil && *eIr.IntervalStart == *extIr.IntervalStart)) ||
		!((eIr.FixedFee == nil && extIr.FixedFee == nil) ||
			(eIr.FixedFee != nil && extIr.FixedFee != nil && *eIr.FixedFee == *extIr.FixedFee)) ||
		!((eIr.RecurrentFee == nil && extIr.RecurrentFee == nil) ||
			(eIr.RecurrentFee != nil && extIr.RecurrentFee != nil && *eIr.RecurrentFee == *extIr.RecurrentFee)) ||
		!((eIr.Unit == nil && extIr.Unit == nil) ||
			(eIr.Unit != nil && extIr.Unit != nil && *eIr.Unit == *extIr.Unit)) ||
		!((eIr.Increment == nil && extIr.Increment == nil) ||
			(eIr.Increment != nil && extIr.Increment != nil && *eIr.Increment == *extIr.Increment)) {
		return
	}
	return true
}

// Equals returns the equality between two IntervalRate
func (iR *IntervalRate) Equals(inRt *IntervalRate) (eq bool) {
	if iR.RecurrentFee == nil && inRt.RecurrentFee != nil ||
		iR.RecurrentFee != nil && inRt.RecurrentFee == nil ||
		iR.FixedFee == nil && inRt.FixedFee != nil ||
		iR.FixedFee != nil && inRt.FixedFee == nil ||
		iR.Increment == nil && inRt.Increment != nil ||
		iR.Increment != nil && inRt.Increment == nil ||
		iR.Unit == nil && inRt.Unit != nil ||
		iR.Unit != nil && inRt.Unit == nil ||
		iR.IntervalStart == nil && inRt.IntervalStart != nil ||
		iR.IntervalStart != nil && inRt.IntervalStart == nil {
		return
	}
	if iR.RecurrentFee != nil && inRt.RecurrentFee != nil &&
		iR.RecurrentFee.Compare(inRt.RecurrentFee) != 0 {
		return
	}
	if iR.FixedFee != nil && inRt.FixedFee != nil &&
		iR.FixedFee.Compare(inRt.FixedFee) != 0 {
		return
	}
	if iR.Increment != nil && inRt.Increment != nil &&
		iR.Increment.Compare(inRt.Increment) != 0 {
		return
	}
	if iR.Unit != nil && inRt.Unit != nil &&
		iR.Unit.Compare(inRt.Unit) != 0 {
		return
	}
	if iR.IntervalStart != nil && inRt.IntervalStart != nil &&
		iR.IntervalStart.Compare(inRt.IntervalStart) != 0 {
		return
	}
	return true
}

func (rt *Rate) Compile() (err error) {
	aTime := rt.ActivationTimes
	if aTime == EmptyString {
		aTime = "* * * * *"
	}
	if rt.sched, err = cron.ParseStandard(aTime); err != nil {
		return
	}
	return
}

// RunTimes returns the set of activation and deactivation times for this rate on the interval between >=sTime and <eTime
// aTimes is in the form of [][]
func (rt *Rate) RunTimes(sTime, eTime time.Time, verbosity int) (aTimes [][]time.Time, err error) {
	sTime = sTime.Add(-time.Minute) // to make sure we can cover startTime
	for i := 0; i < verbosity; i++ {
		aTime := rt.sched.Next(sTime)
		if aTime.IsZero() || !aTime.Before(eTime) { // #TestMe
			return
		}
		iTime := rt.sched.NextInactive(aTime)
		aTimes = append(aTimes, []time.Time{aTime, iTime})
		if iTime.IsZero() || !eTime.After(iTime) { // #TestMe
			return
		}
		sTime = iTime
	}
	// protect from memory leak
	Logger.Warning(
		fmt.Sprintf(
			"maximum runTime iterations reached for Rate: <%+v>, sTime: <%+v>, eTime: <%+v>",
			rt, sTime, eTime))
	return nil, ErrMaxIterationsReached
}

// RateProfileWithAPIOpts is used in replicatorV1 for dispatcher
type RateProfileWithAPIOpts struct {
	*RateProfile
	APIOpts map[string]interface{}
}

// RateSInterval is used by RateS to integrate Rate info for one charging interval
type RateSInterval struct {
	IntervalStart  *Decimal
	Increments     []*RateSIncrement
	CompressFactor int64

	cost *decimal.Big // unexported total interval cost
}

type ExtRateSInterval struct {
	IntervalStart  *float64
	Increments     []*ExtRateSIncrement
	CompressFactor int64

	cost *float64 // unexported total interval cost
}

// AsExtRateSInterval converts RateSInterval to ExtRateSInterval
func (rI *RateSInterval) AsExtRateSInterval() (eRi *ExtRateSInterval, err error) {
	eRi = &ExtRateSInterval{
		CompressFactor: rI.CompressFactor,
	}
	if rI.Increments != nil {
		eRi.Increments = make([]*ExtRateSIncrement, len(rI.Increments))
		for idx, val := range rI.Increments {
			if rcv, err := val.AsExtRateSIncrement(); err != nil {
				return nil, err
			} else {
				eRi.Increments[idx] = rcv
			}
		}
	}
	if rI.IntervalStart != nil {
		if fltIntStart, ok := rI.IntervalStart.Big.Float64(); !ok {
			return nil, errors.New("Cannot convert decimal IntervalStart into float64 ")
		} else {
			eRi.IntervalStart = &fltIntStart
		}
	}
	if rI.cost != nil {
		if fltCost, ok := rI.cost.Float64(); !ok {
			return nil, errors.New("Cannot convert decimal cost into float64 ")
		} else {
			eRi.cost = &fltCost
		}
	}
	return
}

// Equals compares two ExtRateSInterval
func (rIl *ExtRateSInterval) Equals(nRil *ExtRateSInterval) (eq bool) {
	if !((rIl.IntervalStart == nil && nRil.IntervalStart == nil) ||
		(rIl.IntervalStart != nil && nRil.IntervalStart != nil && *rIl.IntervalStart == *nRil.IntervalStart)) ||
		(rIl.Increments == nil && nRil.Increments != nil ||
			rIl.Increments != nil && nRil.Increments == nil ||
			len(rIl.Increments) != len(nRil.Increments)) ||
		(rIl.CompressFactor != nRil.CompressFactor) {
		return
	}
	for i, rtIn := range rIl.Increments {
		if !rtIn.Equals(nRil.Increments[i]) {
			return
		}
	}
	return true
}

type RateSIncrement struct {
	IncrementStart    *Decimal
	Rate              *Rate
	IntervalRateIndex int
	CompressFactor    int64
	Usage             *Decimal

	cost *decimal.Big // unexported total increment cost
}

type ExtRateSIncrement struct {
	IncrementStart    *float64
	Rate              *ExtRate
	IntervalRateIndex int
	CompressFactor    int64
	Usage             *float64

	cost *float64 // unexported total increment cost
}

// AsExtRateSIncrement converts RateSIncrement to ExtRateSIncrement
func (rI *RateSIncrement) AsExtRateSIncrement() (eRi *ExtRateSIncrement, err error) {
	eRi = &ExtRateSIncrement{
		IntervalRateIndex: rI.IntervalRateIndex,
		CompressFactor:    rI.CompressFactor,
	}
	if rI.Rate != nil {
		eRi.Rate = new(ExtRate)
		if rt, err := rI.Rate.AsExtRate(); err != nil {
			return nil, err
		} else {
			eRi.Rate = rt
		}
	}
	if rI.IncrementStart != nil {
		if fltIncrStart, ok := rI.IncrementStart.Big.Float64(); !ok {
			return nil, errors.New("Cannot convert decimal IncrementStart into float64 ")
		} else {
			eRi.IncrementStart = &fltIncrStart
		}
	}
	if rI.Usage != nil {
		if fltUsage, ok := rI.Usage.Big.Float64(); !ok {
			return nil, errors.New("Cannot convert decimal Usage into float64 ")
		} else {
			eRi.Usage = &fltUsage
		}
	}
	if rI.cost != nil {
		if fltCost, ok := rI.cost.Float64(); !ok {
			return nil, errors.New("Cannot convert decimal cost into float64 ")
		} else {
			eRi.cost = &fltCost
		}
	}
	return
}

// Equals returns the equality between twoExt RateSIncrement
func (eRI *ExtRateSIncrement) Equals(extRI *ExtRateSIncrement) (eq bool) {
	if !((eRI.Usage == nil && extRI.Usage == nil) ||
		(eRI.Usage != nil && extRI.Usage != nil && *eRI.Usage == *extRI.Usage)) ||
		!((eRI.IncrementStart == nil && extRI.IncrementStart == nil) ||
			(eRI.IncrementStart != nil && extRI.IncrementStart != nil && *eRI.IncrementStart == *extRI.IncrementStart)) ||
		(((eRI.Rate != nil && extRI.Rate == nil) ||
			(eRI.Rate == nil && extRI.Rate != nil)) ||
			(eRI.Rate != nil && extRI.Rate != nil && !eRI.Rate.Equals(extRI.Rate))) {
		return
	}
	return eRI.CompressFactor == extRI.CompressFactor &&
		eRI.IntervalRateIndex == extRI.IntervalRateIndex
}

// Equals compares two RateSIntervals
func (rIl *RateSInterval) Equals(nRil *RateSInterval) (eq bool) {
	if rIl.IntervalStart == nil && nRil.IntervalStart != nil ||
		rIl.IntervalStart != nil && nRil.IntervalStart == nil ||
		(rIl.IntervalStart != nil && nRil.IntervalStart != nil &&
			rIl.IntervalStart.Compare(nRil.IntervalStart) != 0) ||
		(rIl.Increments != nil && rIl.Increments == nil ||
			rIl.Increments == nil && nRil.Increments != nil ||
			len(rIl.Increments) != len(nRil.Increments)) ||
		rIl.CompressFactor != nRil.CompressFactor {
		return
	}
	if rIl.Increments != nil && nRil.Increments != nil {
		for i, rtIn := range rIl.Increments {
			if !rtIn.Equals(nRil.Increments[i]) {
				return
			}
		}
	}
	return true
}

// Equals returns the equality between two RateSIncrement
func (rI *RateSIncrement) Equals(rtIn *RateSIncrement) (eq bool) {
	if (rI.Usage == nil && rtIn.Usage != nil ||
		rI.Usage != nil && rtIn.Usage == nil ||
		(rI.Usage != nil && rtIn.Usage != nil &&
			rI.Usage.Compare(rtIn.Usage) != 0)) ||
		(rI.IncrementStart == nil && rtIn.IncrementStart != nil ||
			rI.IncrementStart != nil && rtIn.IncrementStart == nil ||
			(rI.IncrementStart != nil && rtIn.IncrementStart != nil &&
				rI.IncrementStart.Compare(rtIn.IncrementStart) != 0)) ||
		(rI.Rate == nil && rtIn.Rate != nil ||
			rI.Rate != nil && rtIn.Rate == nil ||
			(rI.Rate != nil && rtIn.Rate != nil &&
				!rI.Rate.Equals(rtIn.Rate))) ||
		rI.CompressFactor != rtIn.CompressFactor ||
		rI.IntervalRateIndex != rtIn.IntervalRateIndex {
		return
	}
	return true
}

// RateProfileCost is the cost returned by RateS at cost queries
type RateProfileCost struct {
	ID              string // RateProfileID
	Cost            float64
	MinCost         float64
	MaxCost         float64
	MaxCostStrategy string
	RateSIntervals  []*RateSInterval
	Altered         []string
}

// CorrectCost should be called in final phase of cost calculation
// in order to apply further correction like Min/MaxCost or rounding
func (rPc *RateProfileCost) CorrectCost(rndDec *int, rndMtd string) {
	if rPc.MinCost != 0 && rPc.Cost < rPc.MinCost {
		rPc.Cost = rPc.MinCost
		rPc.Altered = append(rPc.Altered, MinCost)
	}
	if rPc.MaxCost != 0 && rPc.Cost > rPc.MaxCost {
		rPc.Cost = rPc.MaxCost
		rPc.Altered = append(rPc.Altered, MaxCost)
	}
	if rndDec != nil {
		rPc.Cost = Round(rPc.Cost, *rndDec, rndMtd)
		rPc.Altered = append(rPc.Altered, RoundingDecimals)
	}
}

// Sort will sort the IntervalRates from each Rate based on IntervalStart
func (rpp *RateProfile) Sort() {
	for _, rate := range rpp.Rates {
		sort.Slice(rate.IntervalRates, func(i, j int) bool {
			return rate.IntervalRates[i].IntervalStart.Compare(rate.IntervalRates[j].IntervalStart) == -1
		})
	}
}

// CompressEquals compares two RateSIntervals for Compress function
func (rIv *RateSInterval) CompressEquals(rIv2 *RateSInterval) (eq bool) {
	if len(rIv.Increments) != len(rIv2.Increments) {
		return
	}
	for i, rIcr := range rIv.Increments {
		if !rIcr.CompressEquals(rIv2.Increments[i]) {
			return
		}
	}
	return true
}

func (rIv *RateSInterval) Cost() *decimal.Big {
	if rIv.cost == nil {
		rIv.cost = new(decimal.Big)
		for _, incrm := range rIv.Increments {
			rIv.cost = SumBig(rIv.cost, incrm.Cost())
		}
	}
	return rIv.cost
}

// CompressEquals compares two RateSIncrement for Compress function
func (rIcr *RateSIncrement) CompressEquals(rIcr2 *RateSIncrement) (eq bool) {
	return rIcr.Rate.UID() == rIcr2.Rate.UID() &&
		rIcr.IntervalRateIndex == rIcr2.IntervalRateIndex &&
		rIcr.Usage.Big.Cmp(rIcr2.Usage.Big) == 0
}

// Cost computes the Cost on RateSIncrement
func (rIcr *RateSIncrement) Cost() *decimal.Big {
	if rIcr.cost == nil {
		icrRt := rIcr.Rate.IntervalRates[rIcr.IntervalRateIndex]
		if rIcr.Usage.Compare(NewDecimal(-1, 0)) == 0 { // FixedFee
			rIcr.cost = icrRt.FixedFee.Big
		} else {
			rIcr.cost = icrRt.RecurrentFee.Big
			if icrRt.Unit != icrRt.Increment {
				rIcr.cost = DivideBig(
					MultiplyBig(rIcr.cost, icrRt.Increment.Big),
					icrRt.Unit.Big)
			}
			if rIcr.CompressFactor != 1 {
				rIcr.cost = MultiplyBig(
					rIcr.cost,
					new(decimal.Big).SetUint64(uint64(rIcr.CompressFactor)))
			}
		}
	}
	return rIcr.cost
}

// CostForIntervals sums the costs for all intervals
func CostForIntervals(rtIvls []*RateSInterval) (cost *decimal.Big) {
	cost = new(decimal.Big)
	for _, rtIvl := range rtIvls {
		cost = SumBig(cost, rtIvl.Cost())
	}
	return
}

// CompressIntervals will compress intervals which equal
func CompressIntervals(rtIvls []*RateSInterval) {
}

// AsRateProfile converts APIRateProfile to RateProfile
func (ext *APIRateProfile) AsRateProfile() (rp *RateProfile, err error) {
	rp = &RateProfile{
		Tenant:             ext.Tenant,
		ID:                 ext.ID,
		FilterIDs:          ext.FilterIDs,
		ActivationInterval: ext.ActivationInterval,
		MaxCostStrategy:    ext.MaxCostStrategy,
	}
	if ext.Weights != EmptyString {
		if rp.Weights, err = NewDynamicWeightsFromString(ext.Weights, ";", "&"); err != nil {
			return nil, err
		}
	}
	if ext.MinCost != nil {
		rp.MinCost = NewDecimalFromFloat64(*ext.MinCost)
	}
	if ext.MaxCost != nil {
		rp.MaxCost = NewDecimalFromFloat64(*ext.MaxCost)
	}
	if len(ext.Rates) != 0 {
		rp.Rates = make(map[string]*Rate)
		for key, extRate := range ext.Rates {
			if rp.Rates[key], err = extRate.AsRate(); err != nil {
				return
			}
		}
	}
	if err = rp.Compile(); err != nil {
		return
	}
	return
}

type APIRateProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weights            string
	MinCost            *float64
	MaxCost            *float64
	MaxCostStrategy    string
	Rates              map[string]*APIRate
	APIOpts            map[string]interface{}
}

// AsRate converts APIRate to Rate
func (ext *APIRate) AsRate() (rate *Rate, err error) {
	rate = &Rate{
		ID:              ext.ID,
		FilterIDs:       ext.FilterIDs,
		ActivationTimes: ext.ActivationTimes,
		Blocker:         ext.Blocker,
	}
	if ext.Weights != EmptyString {
		if rate.Weights, err = NewDynamicWeightsFromString(ext.Weights, ";", "&"); err != nil {
			return nil, err
		}
	}
	if len(ext.IntervalRates) != 0 {
		rate.IntervalRates = make([]*IntervalRate, len(ext.IntervalRates))
		for i, iRate := range ext.IntervalRates {
			if rate.IntervalRates[i], err = iRate.AsIntervalRate(); err != nil {
				return
			}
		}
	}
	return
}

type APIRate struct {
	ID              string   // RateID
	FilterIDs       []string // RateFilterIDs
	ActivationTimes string   // ActivationTimes is a cron formatted time interval
	Weights         string   // RateWeight will decide the winner per interval start
	Blocker         bool     // RateBlocker will make this rate recurrent, deactivating further intervals
	IntervalRates   []*APIIntervalRate
}

// AsIntervalRate converts APIIntervalRate to IntervalRate
func (ext *APIIntervalRate) AsIntervalRate() (iRate *IntervalRate, err error) {
	iRate = new(IntervalRate)
	if iRate.IntervalStart, err = NewDecimalFromUsage(ext.IntervalStart); err != nil {
		return
	}
	if ext.FixedFee != nil {
		iRate.FixedFee = NewDecimalFromFloat64(*ext.FixedFee)
	}
	if ext.RecurrentFee != nil {
		iRate.RecurrentFee = NewDecimalFromFloat64(*ext.RecurrentFee)
	}
	if ext.Unit != nil {
		iRate.Unit = NewDecimalFromFloat64(*ext.Unit)
	}
	if ext.Increment != nil {
		iRate.Increment = NewDecimalFromFloat64(*ext.Increment)
	}
	return
}

type APIIntervalRate struct {
	IntervalStart string
	FixedFee      *float64
	RecurrentFee  *float64
	Unit          *float64 // RateUnit
	Increment     *float64 // RateIncrement
}

type RemoveRPrfRates struct {
	Tenant  string
	ID      string
	RateIDs []string
	APIOpts map[string]interface{}
}
