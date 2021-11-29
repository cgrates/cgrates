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

package utils

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/cron"
	"github.com/ericlagergren/decimal"
)

// RateProfile represents the configuration of a Rate profile
type RateProfile struct {
	Tenant          string
	ID              string
	FilterIDs       []string
	Weights         DynamicWeights
	MinCost         *Decimal
	MaxCost         *Decimal
	MaxCostStrategy string
	Rates           map[string]*Rate
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
	Weights         DynamicWeights // RateWeights will decide the winner per interval start
	Blocker         bool           // RateBlocker will make this rate recurrent, deactivating further intervals
	IntervalRates   []*IntervalRate

	sched cron.Schedule // compiled version of activation times as cron.Schedule interface
	uID   string
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

// Equals returns the equality between two IntervalRate
func (iR *IntervalRate) Equals(inRt *IntervalRate) (eq bool) {
	if iR == nil && inRt == nil {
		return true
	}
	if iR == nil && inRt != nil ||
		iR != nil && inRt == nil {
		return
	}

	return !( //((iR != nil) != (inRt != nil)) ||
	//	(iR == nil || inRt == nil) ||
	(iR.RecurrentFee == nil && inRt.RecurrentFee != nil) ||
		(iR.RecurrentFee != nil && inRt.RecurrentFee == nil) ||
		(iR.FixedFee == nil && inRt.FixedFee != nil) ||
		(iR.FixedFee != nil && inRt.FixedFee == nil) ||
		(iR.Increment == nil && inRt.Increment != nil) ||
		(iR.Increment != nil && inRt.Increment == nil) ||
		(iR.Unit == nil && inRt.Unit != nil) ||
		(iR.Unit != nil && inRt.Unit == nil) ||
		(iR.IntervalStart == nil && inRt.IntervalStart != nil) ||
		(iR.IntervalStart != nil && inRt.IntervalStart == nil) ||
		(iR.RecurrentFee != nil && inRt.RecurrentFee != nil &&
			iR.RecurrentFee.Compare(inRt.RecurrentFee) != 0) ||
		(iR.FixedFee != nil && inRt.FixedFee != nil &&
			iR.FixedFee.Compare(inRt.FixedFee) != 0) ||
		(iR.Increment != nil && inRt.Increment != nil &&
			iR.Increment.Compare(inRt.Increment) != 0) ||
		(iR.Unit != nil && inRt.Unit != nil &&
			iR.Unit.Compare(inRt.Unit) != 0) ||
		(iR.IntervalStart != nil && inRt.IntervalStart != nil &&
			iR.IntervalStart.Compare(inRt.IntervalStart) != 0))
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

// AsRatesIntervalsCost converts RateSInterval to RateSIntervalCost
// The difference between this 2 is that RateSIntervalCost don't need IntervalStart
func (rI *RateSInterval) AsRatesIntervalsCost() (rIc *RateSIntervalCost) {
	rIc = &RateSIntervalCost{
		CompressFactor: rI.CompressFactor,
	}
	if rI.Increments != nil {
		rIc.Increments = make([]*RateSIncrementCost, len(rI.Increments))
		for idx, incr := range rI.Increments {
			rIc.Increments[idx] = incr.AsRateSIncrementCost()
		}
	}
	return
}

type RateSIncrement struct {
	IncrementStart    *Decimal
	RateIntervalIndex int
	RateID            string
	CompressFactor    int64
	Usage             *Decimal

	cost *decimal.Big // unexported total increment cost
}

// Equals compares two RateSIntervals
func (rIl *RateSInterval) Equals(nRil *RateSInterval, rIlRef, nRilRef map[string]*IntervalRate) (eq bool) {
	if rIl == nil && nRil == nil {
		return true
	}
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
			if !rtIn.Equals(nRil.Increments[i], rIlRef, nRilRef) {
				return
			}
		}
	}
	return true
}

// Equals returns the equality between two RateSIncrement
func (rI *RateSIncrement) Equals(rtIn *RateSIncrement, rIRef, rtInRef map[string]*IntervalRate) (eq bool) {
	return !((rI.Usage == nil && rtIn.Usage != nil) ||
		(rI.Usage != nil && rtIn.Usage == nil) ||
		(rI.Usage != nil && rtIn.Usage != nil &&
			rI.Usage.Compare(rtIn.Usage) != 0) ||
		(rI.IncrementStart == nil && rtIn.IncrementStart != nil ||
			rI.IncrementStart != nil && rtIn.IncrementStart == nil ||
			(rI.IncrementStart != nil && rtIn.IncrementStart != nil &&
				rI.IncrementStart.Compare(rtIn.IncrementStart) != 0)) ||
		rI.CompressFactor != rtIn.CompressFactor ||
		rI.RateIntervalIndex != rtIn.RateIntervalIndex ||
		(rIRef != nil && rtInRef != nil &&
			rI.RateID != EmptyString && rtIn.RateID != EmptyString &&
			!rIRef[rI.RateID].Equals(rtInRef[rtIn.RateID])))
}

// RateProfileCost is the cost returned by RateS at cost queries
type RateProfileCost struct {
	ID              string // RateProfileID
	Cost            *Decimal
	MinCost         *Decimal
	MaxCost         *Decimal
	MaxCostStrategy string
	CostIntervals   []*RateSIntervalCost
	Rates           map[string]*IntervalRate
	Altered         []string
}

// RateSIntervalCost is used in the RateProfileCost to reflect the RateSInterval used
type RateSIntervalCost struct {
	Increments     []*RateSIncrementCost
	CompressFactor int64
}

// RateSIncrementCost is used in the RateProfileCost to reflect RateSIncrement
type RateSIncrementCost struct {
	Usage             *Decimal
	RateID            string
	RateIntervalIndex int
	CompressFactor    int64
}

// AsRateSIncrementCost converts RateSIncrement to RateSIncrementCost
// The difference between this 2 is that RateSIncrementCost don't need IncrementStart
func (rI *RateSIncrement) AsRateSIncrementCost() (rIc *RateSIncrementCost) {
	rIc = &RateSIncrementCost{
		RateIntervalIndex: rI.RateIntervalIndex,
		CompressFactor:    rI.CompressFactor,
		RateID:            rI.RateID,
	}
	if rI.Usage != nil {
		rIc.Usage = rI.Usage
	}
	return

}

// Equals returns the equality between two RateSIntervalCost
func (rIC *RateSIntervalCost) Equals(nRIc *RateSIntervalCost, rIlRef, nRilRef map[string]*IntervalRate) (eq bool) {
	if (rIC.Increments != nil && nRIc.Increments == nil ||
		rIC.Increments == nil && nRIc.Increments != nil ||
		len(rIC.Increments) != len(nRIc.Increments)) || rIC.CompressFactor != nRIc.CompressFactor {
		return
	}
	if rIC.Increments != nil && nRIc.Increments != nil {
		for i, rtIn := range rIC.Increments {
			if !rtIn.Equals(nRIc.Increments[i], rIlRef, nRilRef) {
				return
			}
		}
	}
	return true
}

// Equals returns the equality between two RateSIncrementCost
func (rIncrC *RateSIncrementCost) Equals(nRi *RateSIncrementCost, rIRef, rtInRef map[string]*IntervalRate) (eq bool) {
	return !((rIncrC.Usage == nil && nRi.Usage != nil) ||
		(rIncrC.Usage != nil && nRi.Usage == nil) ||
		(rIncrC.Usage != nil && nRi.Usage != nil &&
			rIncrC.Usage.Compare(nRi.Usage) != 0) ||
		rIncrC.CompressFactor != nRi.CompressFactor ||
		rIncrC.RateIntervalIndex != nRi.RateIntervalIndex ||
		(rIRef == nil && rtInRef != nil) ||
		(rIRef != nil && rtInRef == nil) ||
		(rIRef != nil && rtInRef != nil &&
			!rIRef[rIncrC.RateID].Equals(rtInRef[nRi.RateID])))
}

/*
func (rpC *RateProfileCost) SynchronizeRateKeys(nRpCt *RateProfileCost) {
	rts := make(map[string]*IntervalRate)
	reverse := make(map[string]string)
	for key, val := range rpC.Rates {
		reverseKey := key
		for newKey, newVal := range nRpCt.Rates {
			if val.Equals(newVal) {
				reverseKey = newKey
				break
			}
		}
		rts[reverseKey] = val
		reverse[key] = reverseKey
	}
	rpC.Rates = rts
	for _, val := range rpC.RateSIntervals {
		for _, incrVal := range val.Increments {
			incrVal.RateID = reverse[incrVal.RateID]
		}
	}
}

*/

// Equals returns the equality between two RateProfileCost
func (rpC *RateProfileCost) Equals(nRpCt *RateProfileCost) (eq bool) {
	if rpC.ID != nRpCt.ID ||
		(rpC.Cost == nil && nRpCt.Cost != nil) ||
		(rpC.Cost != nil && nRpCt.Cost == nil) ||
		(rpC.Cost != nil && nRpCt.Cost != nil &&
			rpC.Cost.Compare(nRpCt.Cost) != 0) ||
		(rpC.MinCost == nil && nRpCt.MinCost != nil) ||
		(rpC.MinCost != nil && nRpCt.MinCost == nil) ||
		(rpC.MinCost != nil && nRpCt.MinCost != nil &&
			rpC.MinCost.Compare(nRpCt.MinCost) != 0) ||
		(rpC.MaxCost == nil && nRpCt.MaxCost != nil) ||
		(rpC.MaxCost != nil && nRpCt.MaxCost == nil) ||
		(rpC.MaxCost != nil && nRpCt.MaxCost != nil &&
			rpC.MaxCost.Compare(nRpCt.MaxCost) != 0) ||
		rpC.MaxCostStrategy != nRpCt.MaxCostStrategy ||
		(rpC.CostIntervals != nil && nRpCt.CostIntervals == nil ||
			rpC.CostIntervals == nil && nRpCt.CostIntervals != nil ||
			len(rpC.CostIntervals) != len(nRpCt.CostIntervals)) ||
		(rpC.Rates != nil && nRpCt.Rates == nil ||
			rpC.Rates == nil && nRpCt.Rates != nil ||
			len(rpC.Rates) != len(nRpCt.Rates)) ||
		(rpC.Altered != nil && nRpCt.Altered == nil ||
			rpC.Altered == nil && nRpCt.Altered != nil ||
			len(rpC.Altered) != len(nRpCt.Altered)) {
		return
	}
	for idx, val := range rpC.CostIntervals {
		if ok := val.Equals(nRpCt.CostIntervals[idx], rpC.Rates, nRpCt.Rates); !ok {
			return
		}
	}
	for idx, val := range rpC.Altered {
		if val != nRpCt.Altered[idx] {
			return
		}
	}
	return true
}

// CorrectCost should be called in final phase of cost calculation
// in order to apply further correction like Min/MaxCost or rounding
func (rPc *RateProfileCost) CorrectCost(rndDec *int, rndMtd string) {
	if rPc.MinCost != nil && rPc.Cost.Compare(rPc.MinCost) < 0 {
		rPc.Cost = rPc.MinCost
		rPc.Altered = append(rPc.Altered, MinCost)
	}
	if rPc.MaxCost != nil && rPc.Cost.Compare(rPc.MaxCost) > 0 {
		rPc.Cost = rPc.MaxCost
		rPc.Altered = append(rPc.Altered, MaxCost)
	}
	if rndDec != nil {
		rPc.Cost = rPc.Cost.Round(*rndDec)
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

func (rIv *RateSInterval) Cost(rts map[string]*IntervalRate) (cost *decimal.Big) {
	if rIv.cost == nil {
		rIv.cost = new(decimal.Big)
		for _, incrm := range rIv.Increments {
			rIv.cost = SumBig(rIv.cost, incrm.Cost(rts))
		}
	}
	return rIv.cost
}

// CompressEquals compares two RateSIncrement for Compress function
func (rIcr *RateSIncrement) CompressEquals(rIcr2 *RateSIncrement) (eq bool) {
	return rIcr.RateID == rIcr2.RateID &&
		rIcr.RateIntervalIndex == rIcr2.RateIntervalIndex &&
		rIcr.Usage.Big.Cmp(rIcr2.Usage.Big) == 0
}

// Cost computes the Cost on RateSIncrement
func (rIcr *RateSIncrement) Cost(rts map[string]*IntervalRate) (cost *decimal.Big) {
	if rIcr.cost == nil {
		icrRt, has := rts[rIcr.RateID]
		if !has {
			// return nil, fmt.Errorf("Cannot get the IntervalRate with this RateID: %s", rIcr.RateID)
			return
		}
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

// CompressIntervals will compress intervals which equal
// func CompressIntervals(rtIvls []*RateSInterval) {
// }

// AsRateProfile converts APIRateProfile to RateProfile
func (ext *APIRateProfile) AsRateProfile() (rp *RateProfile, err error) {
	rp = &RateProfile{
		Tenant:          ext.Tenant,
		ID:              ext.ID,
		FilterIDs:       ext.FilterIDs,
		MaxCostStrategy: ext.MaxCostStrategy,
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
	err = rp.Compile()
	return
}

type APIRateProfile struct {
	Tenant          string
	ID              string
	FilterIDs       []string
	Weights         string
	MinCost         *float64
	MaxCost         *float64
	MaxCostStrategy string
	Rates           map[string]*APIRate
	APIOpts         map[string]interface{}
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
	Weights         string   // RateWeights will decide the winner per interval start
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

func (rp *RateProfile) Set(path []string, val interface{}, newBranch bool, _ string) (err error) {
	if len(path) == 0 {
		return ErrWrongPath
	}
	var rtID string
	if len(path) != 1 && path[0] == Rates {
		rtID = path[1]
		path = path[1:]
	} else if strings.HasPrefix(path[0], Rates) &&
		path[0][5] == '[' && path[0][len(path[0])-1] == ']' {
		rtID = path[0][6 : len(path[0])-1]
	}
	if rtID != EmptyString {
		if _, has := rp.Rates[rtID]; !has {
			rp.Rates[rtID] = &Rate{
				ID: rtID,
			}
		}
		return rp.Rates[rtID].Set(path[1:], val, newBranch)
	}
	if len(path) != 1 {
		return ErrWrongPath
	}
	switch path[0] {
	default:
		return ErrWrongPath
	case Tenant:
		rp.Tenant = IfaceAsString(val)
	case ID:
		rp.ID = IfaceAsString(val)
	case FilterIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		rp.FilterIDs = append(rp.FilterIDs, valA...)
	case Weights:
		rp.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
	case MinCost:
		if val != EmptyString {
			var valB *decimal.Big
			valB, err = IfaceAsBig(val)
			rp.MinCost = &Decimal{valB}
		}
	case MaxCost:
		if val != EmptyString {
			var valB *decimal.Big
			valB, err = IfaceAsBig(val)
			rp.MaxCost = &Decimal{valB}
		}
	case MaxCostStrategy:
		rp.MaxCostStrategy = IfaceAsString(val)
	}
	return

}

func (rt *Rate) Set(path []string, val interface{}, newBranch bool) (err error) {
	switch len(path) {
	default:
		return ErrWrongPath
	case 1:
		switch path[0] {
		default:
			return ErrWrongPath
		case ID:
			rt.ID = IfaceAsString(val)
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			rt.FilterIDs = append(rt.FilterIDs, valA...)
		case Weights:
			rt.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		case ActivationTimes:
			rt.ActivationTimes = IfaceAsString(val)
		case Blocker:
			rt.Blocker, err = IfaceAsBool(val)
		}
	case 2:
		if path[0] != IntervalRates {
			return ErrWrongPath
		}
		if len(rt.IntervalRates) == 0 || newBranch {
			rt.IntervalRates = append(rt.IntervalRates, &IntervalRate{IntervalStart: NewDecimal(0, 0), FixedFee: NewDecimal(0, 0)})
		}
		switch path[1] {
		case IntervalStart:
			var valB *decimal.Big
			valB, err = IfaceAsBig(val)
			rt.IntervalRates[len(rt.IntervalRates)-1].IntervalStart = &Decimal{valB}
		case FixedFee:
			var valB *decimal.Big
			valB, err = IfaceAsBig(val)
			rt.IntervalRates[len(rt.IntervalRates)-1].FixedFee = &Decimal{valB}
		case RecurrentFee:
			var valB *decimal.Big
			valB, err = IfaceAsBig(val)
			rt.IntervalRates[len(rt.IntervalRates)-1].RecurrentFee = &Decimal{valB}
		case Unit:
			var valB *decimal.Big
			valB, err = IfaceAsBig(val)
			rt.IntervalRates[len(rt.IntervalRates)-1].Unit = &Decimal{valB}
		case Increment:
			var valB *decimal.Big
			valB, err = IfaceAsBig(val)
			rt.IntervalRates[len(rt.IntervalRates)-1].Increment = &Decimal{valB}
		default:
			return ErrWrongPath
		}
	}
	return
}
