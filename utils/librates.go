/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"fmt"
	"sort"
	"strconv"
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

// Clone clones *RateProfile
func (rp *RateProfile) Clone() *RateProfile {
	if rp == nil {
		return nil
	}
	cloned := &RateProfile{
		Tenant:          rp.Tenant,
		ID:              rp.ID,
		MaxCostStrategy: rp.MaxCostStrategy,
	}
	if rp.FilterIDs != nil {
		cloned.FilterIDs = make([]string, len(rp.FilterIDs))
		copy(cloned.FilterIDs, rp.FilterIDs)
	}
	if rp.MinCost != nil {
		cloned.MinCost = rp.MinCost.Clone()
	}
	if rp.MaxCost != nil {
		cloned.MaxCost = rp.MaxCost.Clone()
	}
	if rp.Weights != nil {
		cloned.Weights = rp.Weights.Clone()
	}
	if rp.Rates != nil {
		cloned.Rates = make(map[string]*Rate)
		for k, v := range rp.Rates {
			if v != nil {
				cloned.Rates[k] = v.Clone()
			}
		}
	}
	return cloned
}

// CacheClone returns a clone of RateProfile used by ltcache CacheCloner
func (rp *RateProfile) CacheClone() any {
	return rp.Clone()
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

// Clone returns a copy of rt
func (rt *Rate) Clone() *Rate {
	if rt == nil {
		return nil
	}
	cln := &Rate{
		ID:              rt.ID,
		ActivationTimes: rt.ActivationTimes,
		Blocker:         rt.Blocker,
		uID:             rt.uID,
		sched:           rt.sched,
	}
	if rt.FilterIDs != nil {
		cln.FilterIDs = make([]string, len(rt.FilterIDs))
		copy(cln.FilterIDs, rt.FilterIDs)
	}
	if rt.Weights != nil {
		cln.Weights = rt.Weights.Clone()
	}
	if rt.IntervalRates != nil {
		cln.IntervalRates = make([]*IntervalRate, len(rt.IntervalRates))
		for i, value := range rt.IntervalRates {
			cln.IntervalRates[i] = value.Clone()
		}
	}
	return cln
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

// Clone returns a copy of iR
func (iR *IntervalRate) Clone() *IntervalRate {
	cln := new(IntervalRate)
	if iR.IntervalStart != nil {
		cln.IntervalStart = iR.IntervalStart.Clone()
	}
	if iR.FixedFee != nil {
		cln.FixedFee = iR.FixedFee.Clone()
	}
	if iR.RecurrentFee != nil {
		cln.RecurrentFee = iR.RecurrentFee.Clone()
	}
	if iR.Unit != nil {
		cln.Unit = iR.Unit.Clone()
	}
	if iR.Increment != nil {
		cln.Increment = iR.Increment.Clone()
	}
	return cln
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

	return !(iR.RecurrentFee == nil && inRt.RecurrentFee != nil ||
		iR.RecurrentFee != nil && inRt.RecurrentFee == nil ||
		iR.FixedFee == nil && inRt.FixedFee != nil ||
		iR.FixedFee != nil && inRt.FixedFee == nil ||
		iR.Increment == nil && inRt.Increment != nil ||
		iR.Increment != nil && inRt.Increment == nil ||
		iR.Unit == nil && inRt.Unit != nil ||
		iR.Unit != nil && inRt.Unit == nil ||
		iR.IntervalStart == nil && inRt.IntervalStart != nil ||
		iR.IntervalStart != nil && inRt.IntervalStart == nil ||
		iR.RecurrentFee != nil && inRt.RecurrentFee != nil &&
			iR.RecurrentFee.Compare(inRt.RecurrentFee) != 0 ||
		iR.FixedFee != nil && inRt.FixedFee != nil &&
			iR.FixedFee.Compare(inRt.FixedFee) != 0 ||
		iR.Increment != nil && inRt.Increment != nil &&
			iR.Increment.Compare(inRt.Increment) != 0 ||
		iR.Unit != nil && inRt.Unit != nil &&
			iR.Unit.Compare(inRt.Unit) != 0 ||
		iR.IntervalStart != nil && inRt.IntervalStart != nil &&
			iR.IntervalStart.Compare(inRt.IntervalStart) != 0)
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
	APIOpts map[string]any
}

// RateSInterval is used by RateS to integrate Rate info for one charging interval
type RateSInterval struct {
	IntervalStart  *Decimal
	Increments     []*RateSIncrement
	CompressFactor int64

	cost *decimal.Big // unexported total interval cost
}

// Clone returns a copy of rI
func (rI *RateSInterval) Clone() *RateSInterval {
	cln := &RateSInterval{
		CompressFactor: rI.CompressFactor,
	}
	if rI.IntervalStart != nil {
		cln.IntervalStart = rI.IntervalStart.Clone()
	}
	if rI.Increments != nil {
		cln.Increments = make([]*RateSIncrement, len(rI.Increments))
		for i, value := range rI.Increments {
			cln.Increments[i] = value.Clone()
		}
	}
	if rI.cost != nil {
		tmp := &decimal.Big{}
		cln.cost = tmp.Copy(rI.cost)
	}
	return cln
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

func (rI *RateSInterval) FieldAsInterface(fldPath []string) (any, error) {
	switch fldPath[0] {
	case IntervalStart:
		if len(fldPath) != 1 {
			return nil, ErrNotFound
		}
		return rI.IntervalStart, nil
	case CompressFactor:
		if len(fldPath) != 1 {
			return nil, ErrNotFound
		}
		return rI.CompressFactor, nil
	case Increments:
		if len(fldPath) != 1 {
			return nil, ErrNotFound
		}
		return rI.Increments, nil
	}

	opath, indx := GetPathIndex(fldPath[0])
	if opath != Increments {
		return nil, fmt.Errorf("unsupported field prefix: <%s>", opath)
	}
	if indx == nil {
		return nil, fmt.Errorf("invalid index for '%s' field", JoinedCharge)
	}
	if len(rI.Increments) <= *indx {
		return nil, ErrNotFound
	}
	incr := rI.Increments[*indx]
	if len(fldPath) == 1 {
		return incr, nil
	}
	return incr.FieldAsInterface(fldPath[1:])
}

type RateSIncrement struct {
	IncrementStart    *Decimal
	RateIntervalIndex int
	RateID            string
	CompressFactor    int64
	Usage             *Decimal

	cost *decimal.Big // unexported total increment cost
}

// Clone returns a copy of rI
func (rI *RateSIncrement) Clone() *RateSIncrement {
	cln := &RateSIncrement{
		RateIntervalIndex: rI.RateIntervalIndex,
		RateID:            rI.RateID,
		CompressFactor:    rI.CompressFactor,
	}
	if rI.IncrementStart != nil {
		cln.IncrementStart = rI.IncrementStart.Clone()
	}
	if rI.Usage != nil {
		cln.Usage = rI.Usage.Clone()
	}
	if rI.cost != nil {
		tmp := &decimal.Big{}
		cln.cost = tmp.Copy(rI.cost)
	}
	return cln
}

// Equals compares two RateSIntervals
func (rIl *RateSInterval) Equals(nRil *RateSInterval, rIlRef, nRilRef map[string]*IntervalRate) (eq bool) {
	if rIl == nil && nRil == nil {
		return true
	}
	if rIl.IntervalStart == nil && nRil.IntervalStart != nil ||
		rIl.IntervalStart != nil && nRil.IntervalStart == nil ||
		rIl.IntervalStart != nil && nRil.IntervalStart != nil &&
			rIl.IntervalStart.Compare(nRil.IntervalStart) != 0 ||
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
	return !(rI.Usage == nil && rtIn.Usage != nil ||
		rI.Usage != nil && rtIn.Usage == nil ||
		rI.Usage != nil && rtIn.Usage != nil &&
			rI.Usage.Compare(rtIn.Usage) != 0 ||
		(rI.IncrementStart == nil && rtIn.IncrementStart != nil ||
			rI.IncrementStart != nil && rtIn.IncrementStart == nil ||
			rI.IncrementStart != nil && rtIn.IncrementStart != nil &&
				rI.IncrementStart.Compare(rtIn.IncrementStart) != 0) ||
		rI.CompressFactor != rtIn.CompressFactor ||
		rI.RateIntervalIndex != rtIn.RateIntervalIndex ||
		rIRef != nil && rtInRef != nil &&
			rI.RateID != EmptyString && rtIn.RateID != EmptyString &&
			!rIRef[rI.RateID].Equals(rtInRef[rtIn.RateID]))
}

func (rI *RateSIncrement) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	case IncrementStart:
		return rI.IncrementStart, nil
	case RateIntervalIndex:
		return rI.RateIntervalIndex, nil
	case RateID:
		return rI.RateID, nil
	case CompressFactor:
		return rI.CompressFactor, nil
	case Usage:
		return rI.Usage, nil
	}
	return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
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
	if rIC.Increments != nil && nRIc.Increments == nil ||
		rIC.Increments == nil && nRIc.Increments != nil ||
		len(rIC.Increments) != len(nRIc.Increments) || rIC.CompressFactor != nRIc.CompressFactor {
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
	return !(rIncrC.Usage == nil && nRi.Usage != nil ||
		rIncrC.Usage != nil && nRi.Usage == nil ||
		rIncrC.Usage != nil && nRi.Usage != nil &&
			rIncrC.Usage.Compare(nRi.Usage) != 0 ||
		rIncrC.CompressFactor != nRi.CompressFactor ||
		rIncrC.RateIntervalIndex != nRi.RateIntervalIndex ||
		rIRef == nil && rtInRef != nil ||
		rIRef != nil && rtInRef == nil ||
		rIRef != nil && rtInRef != nil &&
			!rIRef[rIncrC.RateID].Equals(rtInRef[nRi.RateID]))
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
		rpC.Cost == nil && nRpCt.Cost != nil ||
		rpC.Cost != nil && nRpCt.Cost == nil ||
		rpC.Cost != nil && nRpCt.Cost != nil &&
			rpC.Cost.Compare(nRpCt.Cost) != 0 ||
		rpC.MinCost == nil && nRpCt.MinCost != nil ||
		rpC.MinCost != nil && nRpCt.MinCost == nil ||
		rpC.MinCost != nil && nRpCt.MinCost != nil &&
			rpC.MinCost.Compare(nRpCt.MinCost) != 0 ||
		rpC.MaxCost == nil && nRpCt.MaxCost != nil ||
		rpC.MaxCost != nil && nRpCt.MaxCost == nil ||
		rpC.MaxCost != nil && nRpCt.MaxCost != nil &&
			rpC.MaxCost.Compare(nRpCt.MaxCost) != 0 ||
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
					decimal.WithContext(DecimalContext).SetUint64(uint64(rIcr.CompressFactor)))
			}
		}
	}
	return rIcr.cost
}

// CompressIntervals will compress intervals which equal
// func CompressIntervals(rtIvls []*RateSInterval) {
// }

type APIRateProfile struct {
	*RateProfile
	APIOpts map[string]any
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
	APIOpts map[string]any
}

func (rp *RateProfile) Set(path []string, val any, newBranch bool) (err error) {
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
		if val != EmptyString {
			rp.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
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

func (rt *Rate) Set(path []string, val any, newBranch bool) (err error) {
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
			if val != EmptyString {
				rt.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
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

func (rp *RateProfile) Merge(v2 any) {
	vi := v2.(*RateProfile)
	if len(vi.Tenant) != 0 {
		rp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		rp.ID = vi.ID
	}
	if len(vi.MaxCostStrategy) != 0 {
		rp.MaxCostStrategy = vi.MaxCostStrategy
	}
	rp.FilterIDs = append(rp.FilterIDs, vi.FilterIDs...)
	rp.Weights = append(rp.Weights, vi.Weights...)
	for k, v := range vi.Rates {
		rt, has := rp.Rates[k]
		if !has {
			rp.Rates[k] = v
			continue
		}
		rt.Merge(v)
	}
	o := decimal.New(0, 0)
	if vi.MinCost != nil && vi.MinCost.Cmp(o) != 0 {
		rp.MinCost = vi.MinCost
	}
	if vi.MaxCost != nil && vi.MinCost.Cmp(o) != 0 {
		rp.MaxCost = vi.MaxCost
	}
}

func (rt *Rate) Merge(vi *Rate) {
	if len(vi.ID) != 0 {
		rt.ID = vi.ID
	}
	if len(vi.ActivationTimes) != 0 {
		rt.ActivationTimes = vi.ActivationTimes
	}
	if vi.Blocker {
		rt.Blocker = vi.Blocker
	}
	rt.FilterIDs = append(rt.FilterIDs, vi.FilterIDs...)
	rt.Weights = append(rt.Weights, vi.Weights...)
	var equal bool
	for _, ivalRateV2 := range vi.IntervalRates {
		for _, ivalRate := range rt.IntervalRates {
			if ivalRate.IntervalStart.Compare(ivalRateV2.IntervalStart) == 0 {
				ivalRate.Merge(ivalRateV2)
				equal = true
				break
			}
		}
		if !equal {
			rt.IntervalRates = append(rt.IntervalRates, ivalRateV2)
		}
		equal = false
	}
}

func (ivalRate *IntervalRate) Merge(v2 *IntervalRate) {
	if v2.IntervalStart.Compare(NewDecimal(0, 0)) != 0 {
		ivalRate.IntervalStart = v2.IntervalStart
	}
	if v2.FixedFee.Compare(NewDecimal(0, 0)) != 0 {
		ivalRate.FixedFee = v2.FixedFee
	}
	if v2.RecurrentFee.Compare(NewDecimal(0, 0)) != 0 {
		ivalRate.RecurrentFee = v2.RecurrentFee
	}
	if v2.Unit.Compare(NewDecimal(0, 0)) != 0 {
		ivalRate.Unit = v2.Unit
	}
	if v2.Increment.Compare(NewDecimal(0, 0)) != 0 {
		ivalRate.Increment = v2.Increment
	}
}

func (rp *RateProfile) String() string { return ToJSON(rp) }
func (rp *RateProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = rp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (rp *RateProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idxStr := GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(rp.FilterIDs) {
						return rp.FilterIDs[idx], nil
					}
				case Rates:
					if rt, has := rp.Rates[*idxStr]; has {
						return rt, nil
					}
				}
			}
			return nil, ErrNotFound
		case Tenant:
			return rp.Tenant, nil
		case ID:
			return rp.ID, nil
		case FilterIDs:
			return rp.FilterIDs, nil
		case Weights:
			return rp.Weights.String(InfieldSep, ANDSep), nil
		case MinCost:
			return rp.MinCost, nil
		case MaxCost:
			return rp.MaxCost, nil
		case MaxCostStrategy:
			return rp.MaxCostStrategy, nil
		case Rates:
			return rp.Rates, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, ErrNotFound
	}
	fld, idxStr := GetPathIndexString(fldPath[0])
	if fld != Rates {
		return nil, ErrNotFound
	}

	if idxStr == nil {

		idxStr = &fldPath[1]
		fldPath = fldPath[1:]
	}
	rt, has := rp.Rates[*idxStr]
	if !has {
		return nil, ErrNotFound
	}
	if len(fldPath) == 1 {
		return rt, nil
	}
	return rt.FieldAsInterface(fldPath[1:])
}

func (rt *Rate) String() string { return ToJSON(rt) }
func (rt *Rate) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = rt.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (rt *Rate) FieldAsInterface(fldPath []string) (_ any, err error) {
	switch len(fldPath) {
	default:
		return nil, ErrNotFound
	case 1:
		switch fldPath[0] {
		default:
			fld, idx := GetPathIndex(fldPath[0])
			if idx != nil {
				switch fld {
				case IntervalRates:
					if *idx < len(rt.IntervalRates) {
						return rt.IntervalRates[*idx], nil
					}
				case FilterIDs:
					if *idx < len(rt.FilterIDs) {
						return rt.FilterIDs[*idx], nil
					}
				}
			}
			return nil, ErrNotFound
		case ID:
			return rt.ID, nil
		case FilterIDs:
			return rt.FilterIDs, nil
		case Weights:
			return rt.Weights.String(InfieldSep, ANDSep), nil
		case IntervalRates:
			return rt.IntervalRates, nil
		case Blocker:
			return rt.Blocker, nil
		case ActivationTimes:
			return rt.ActivationTimes, nil
		}
	case 2:
		fld, idx := GetPathIndex(fldPath[0])
		if fld != IntervalRates ||
			idx == nil ||
			*idx >= len(rt.IntervalRates) {
			return nil, ErrNotFound
		}
		return rt.IntervalRates[*idx].FieldAsInterface(fldPath[1:])
	}
}

func (iR *IntervalRate) String() string { return ToJSON(iR) }
func (iR *IntervalRate) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = iR.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (iR *IntervalRate) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, ErrNotFound
	case IntervalStart:
		return iR.IntervalStart, nil
	case FixedFee:
		return iR.FixedFee, nil
	case RecurrentFee:
		return iR.RecurrentFee, nil
	case Unit:
		return iR.Unit, nil
	case Increment:
		return iR.Increment, nil
	}
}

// AsDataDBMap is used to is a convert method in order to properly set trough a hasmap in redis server our rate profile
func (rp *RateProfile) AsDataDBMap(ms Marshaler) (mp map[string]any, err error) {
	mp = map[string]any{
		MaxCostStrategy: rp.MaxCostStrategy,
	}
	if len(rp.FilterIDs) != 0 {
		var fltrs string
		for idx, fltr := range rp.FilterIDs {
			fltrs += fltr
			if idx != len(rp.FilterIDs)-1 {
				fltrs += InfieldSep
			}
		}
		mp[FilterIDs] = fltrs
	}
	if rp.Weights != nil {
		mp[Weights] = rp.Weights.String(InfieldSep, ANDSep)
	}
	if rp.MinCost != nil {
		mp[MinCost] = rp.MinCost.String()
	}
	if rp.MaxCost != nil {
		mp[MaxCost] = rp.MaxCost.String()
	}
	for rateID, rt := range rp.Rates {
		var result []byte
		if result, err = ms.Marshal(rt); err != nil {
			return nil, err
		}
		fldKey := ConcatenatedKey(Rates, rateID)
		mp[fldKey] = string(result)
	}
	return mp, nil
}

// NewRateProfileFromMapDataDBMap will convert a RateProfile map into a RatePRofile struct. This is used when we get the map from redis database
func NewRateProfileFromMapDataDBMap(tnt, id string, mapRP map[string]any, ms Marshaler) (rp *RateProfile, err error) {
	rp = &RateProfile{
		ID:              id,
		Tenant:          tnt,
		MaxCostStrategy: IfaceAsString(mapRP[MaxCostStrategy]),
		Rates:           make(map[string]*Rate),
	}
	if fltrsIDs, has := mapRP[FilterIDs]; has {
		fltrs := strings.Split(IfaceAsString(fltrsIDs), InfieldSep)
		rp.FilterIDs = make([]string, len(fltrs))
		copy(rp.FilterIDs, fltrs)
	}
	if weights, has := mapRP[Weights]; has {
		rp.Weights, err = NewDynamicWeightsFromString(IfaceAsString(weights), InfieldSep, ANDSep)
		if err != nil {
			return nil, err
		}
	}
	if minCost, has := mapRP[MinCost]; has {
		rp.MinCost, err = NewDecimalFromString(IfaceAsString(minCost))
		if err != nil {
			return nil, err
		}
	}
	if maxCost, has := mapRP[MaxCost]; has {
		rp.MaxCost, err = NewDecimalFromString(IfaceAsString(maxCost))
		if err != nil {
			return nil, err
		}
	}
	for keyID, rateStr := range mapRP {
		if strings.HasPrefix(keyID, Rates+ConcatenatedKeySep) {
			var rate *Rate
			if err := ms.Unmarshal([]byte(IfaceAsString(rateStr)), &rate); err != nil {
				return nil, err
			}
			rp.Rates[strings.TrimPrefix(keyID, Rates+ConcatenatedKeySep)] = rate
		}
	}
	return rp, err
}
