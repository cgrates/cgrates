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

package engine

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Can hold different units as seconds or monetary
type Balance struct {
	Uuid           string // system wide unique
	ID             string // account wide unique
	Value          float64
	ExpirationDate time.Time
	Weight         float64
	DestinationIDs utils.StringMap
	RatingSubject  string
	Categories     utils.StringMap
	SharedGroups   utils.StringMap
	Timings        []*RITiming
	TimingIDs      utils.StringMap
	Disabled       bool
	Factors        ValueFactors
	Blocker        bool
	precision      int
	account        *Account // used to store ub reference for shared balances
	dirty          bool
}

func (b *Balance) Equal(o *Balance) bool {
	if len(b.DestinationIDs) == 0 {
		b.DestinationIDs = utils.StringMap{utils.MetaAny: true}
	}
	if len(o.DestinationIDs) == 0 {
		o.DestinationIDs = utils.StringMap{utils.MetaAny: true}
	}
	return b.Uuid == o.Uuid &&
		b.ID == o.ID &&
		b.ExpirationDate.Equal(o.ExpirationDate) &&
		b.Weight == o.Weight &&
		b.DestinationIDs.Equal(o.DestinationIDs) &&
		b.RatingSubject == o.RatingSubject &&
		b.Categories.Equal(o.Categories) &&
		b.SharedGroups.Equal(o.SharedGroups) &&
		b.Disabled == o.Disabled &&
		b.Blocker == o.Blocker
}

func (b *Balance) MatchFilter(o *BalanceFilter, bType string, skipIds, skipExpiry bool) bool {
	if o == nil {
		return true
	}
	if bType != "" && o.Type != nil && *o.Type != "" {
		if bType != *o.Type {
			return false
		}
	}
	if !skipIds && o.Uuid != nil && *o.Uuid != "" {
		return b.Uuid == *o.Uuid
	}
	if !skipIds && o.ID != nil && *o.ID != "" {
		return b.ID == *o.ID
	}
	if !skipExpiry {
		if o.ExpirationDate != nil && !b.ExpirationDate.Equal(*o.ExpirationDate) {
			return false
		}
	}
	return (o.Weight == nil || b.Weight == *o.Weight) &&
		(o.Blocker == nil || b.Blocker == *o.Blocker) &&
		(o.Disabled == nil || b.Disabled == *o.Disabled) &&
		(o.DestinationIDs == nil || b.DestinationIDs.Includes(*o.DestinationIDs)) &&
		(o.Categories == nil || b.Categories.Includes(*o.Categories)) &&
		(o.TimingIDs == nil || b.TimingIDs.Includes(*o.TimingIDs)) &&
		(o.SharedGroups == nil || b.SharedGroups.Includes(*o.SharedGroups)) &&
		(o.RatingSubject == nil || b.RatingSubject == *o.RatingSubject)
}

func (b *Balance) HardMatchFilter(o *BalanceFilter, skipIds bool) bool {
	if o == nil {
		return true
	}
	if !skipIds && o.Uuid != nil && *o.Uuid != "" {
		return b.Uuid == *o.Uuid
	}
	if !skipIds && o.ID != nil && *o.ID != "" {
		return b.ID == *o.ID
	}
	return (o.ExpirationDate == nil || b.ExpirationDate.Equal(*o.ExpirationDate)) &&
		(o.Weight == nil || b.Weight == *o.Weight) &&
		(o.Blocker == nil || b.Blocker == *o.Blocker) &&
		(o.Disabled == nil || b.Disabled == *o.Disabled) &&
		(o.DestinationIDs == nil || b.DestinationIDs.Equal(*o.DestinationIDs)) &&
		(o.Categories == nil || b.Categories.Equal(*o.Categories)) &&
		(o.TimingIDs == nil || b.TimingIDs.Equal(*o.TimingIDs)) &&
		(o.SharedGroups == nil || b.SharedGroups.Equal(*o.SharedGroups)) &&
		(o.RatingSubject == nil || b.RatingSubject == *o.RatingSubject)
}

// the default balance has standard Id
func (b *Balance) IsDefault() bool {
	return b.ID == utils.MetaDefault
}

// IsExpiredAt check if ExpirationDate is before time t
func (b *Balance) IsExpiredAt(t time.Time) bool {
	return !b.ExpirationDate.IsZero() && b.ExpirationDate.Before(t)
}

func (b *Balance) IsActive() bool {
	return b.IsActiveAt(time.Now())
}

func (b *Balance) IsActiveAt(t time.Time) bool {
	if b.Disabled {
		return false
	}
	if len(b.Timings) == 0 {
		return true
	}
	for _, tim := range b.Timings {
		if tim.IsActiveAt(t) {
			return true
		}
	}
	return false
}

func (b *Balance) MatchCategory(category string) bool {
	return len(b.Categories) == 0 || b.Categories[category]
}

func (b *Balance) HasDestination() bool {
	return len(b.DestinationIDs) > 0 && !b.DestinationIDs[utils.MetaAny]
}

func (b *Balance) MatchDestination(destinationID string) bool {
	return !b.HasDestination() || b.DestinationIDs[destinationID]
}

func (b *Balance) MatchActionTrigger(at *ActionTrigger) bool {
	return b.HardMatchFilter(at.Balance, false)
}

func (b *Balance) Clone() *Balance {
	if b == nil {
		return nil
	}
	n := &Balance{
		Uuid:           b.Uuid,
		ID:             b.ID,
		Value:          b.Value, // this value is in seconds
		ExpirationDate: b.ExpirationDate,
		Weight:         b.Weight,
		RatingSubject:  b.RatingSubject,
		Categories:     b.Categories,
		SharedGroups:   b.SharedGroups,
		TimingIDs:      b.TimingIDs,
		Timings:        b.Timings, // should not be a problem with aliasing
		Blocker:        b.Blocker,
		Disabled:       b.Disabled,
		DestinationIDs: b.DestinationIDs,
		Factors:        b.Factors,
		dirty:          b.dirty,
	}
	if b.DestinationIDs != nil {
		n.DestinationIDs = b.DestinationIDs.Clone()
	}
	return n
}

func (b *Balance) getMatchingPrefixAndDestID(dest string) (prefix, destID string) {
	if len(b.DestinationIDs) != 0 && !b.DestinationIDs[utils.MetaAny] {
		for _, p := range utils.SplitPrefix(dest, MIN_PREFIX_MATCH) {
			if destIDs, err := dm.GetReverseDestination(p, true, true, utils.NonTransactional); err == nil {
				for _, dID := range destIDs {
					if b.DestinationIDs[dID] {
						return p, dID
					}
				}
			}
		}
	}
	return
}

// Returns the available number of seconds for a specified credit
func (b *Balance) GetMinutesForCredit(origCD *CallDescriptor, initialCredit float64) (duration time.Duration, credit float64) {
	cd := origCD.Clone()
	availableDuration := time.Duration(b.GetValue()) * time.Second
	duration = availableDuration
	credit = initialCredit
	cc, err := b.GetCost(cd, false)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
		return 0, credit
	}
	if cc.deductConnectFee {
		connectFee := cc.GetConnectFee()
		if connectFee <= credit {
			credit -= connectFee
			// remove connect fee from the total cost
			cc.Cost -= connectFee
		} else {
			return 0, credit
		}
	}
	if cc.Cost > 0 {
		duration = 0
		for _, ts := range cc.Timespans {
			ts.createIncrementsSlice()
			if cd.MaxRate > 0 && cd.MaxRateUnit > 0 {
				rate, _, rateUnit := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
				if rate/float64(rateUnit.Nanoseconds()) > cd.MaxRate/float64(cd.MaxRateUnit.Nanoseconds()) {
					return
				}
			}
			for _, incr := range ts.Increments {
				if incr.Cost <= credit && availableDuration-incr.Duration >= 0 {
					credit -= incr.Cost
					duration += incr.Duration
					availableDuration -= incr.Duration
				} else {
					return
				}
			}
		}
	}
	return
}

// Gets the cost using balance RatingSubject if present otherwize
// retuns a callcost obtained using standard rating
func (b *Balance) GetCost(cd *CallDescriptor, getStandardIfEmpty bool) (*CallCost, error) {
	// testing only
	if cd.testCallcost != nil {
		return cd.testCallcost, nil
	}
	if b.RatingSubject != "" && !strings.HasPrefix(b.RatingSubject, utils.MetaRatingSubjectPrefix) {
		origSubject := cd.Subject
		cd.Subject = b.RatingSubject
		origAccount := cd.Account
		cd.Account = cd.Subject
		cd.RatingInfos = nil
		cc, err := cd.getCost()
		// restor orig values
		cd.Subject = origSubject
		cd.Account = origAccount
		return cc, err
	}
	if getStandardIfEmpty {
		cd.RatingInfos = nil
		return cd.getCost()
	}
	cc := cd.CreateCallCost()
	cc.Cost = 0
	return cc, nil
}

func (b *Balance) GetValue() float64 {
	return b.Value
}

func (b *Balance) AddValue(amount float64) {
	b.SetValue(b.GetValue() + amount)
}

func (b *Balance) SubtractValue(amount float64) {
	b.SetValue(b.GetValue() - amount)
}

func (b *Balance) SetValue(amount float64) {
	b.Value = amount
	b.Value = utils.Round(b.GetValue(), globalRoundingDecimals, utils.MetaRoundingMiddle)
	b.dirty = true
}

func (b *Balance) SetDirty() {
	b.dirty = true
}

// AsBalanceSummary converts the balance towards compressed information to be displayed
func (b *Balance) AsBalanceSummary(typ string) *BalanceSummary {
	bd := &BalanceSummary{
		UUID:     b.Uuid,
		ID:       b.ID,
		Type:     typ,
		Value:    b.Value,
		Weight:   b.Weight,
		Disabled: b.Disabled,
		Factors:  b.Factors,
	}
	if bd.ID == "" {
		bd.ID = b.Uuid
	}
	return bd
}

/*
Structure to store minute buckets according to weight, precision or price.
*/
type Balances []*Balance

func (bc Balances) Len() int {
	return len(bc)
}

func (bc Balances) Swap(i, j int) {
	bc[i], bc[j] = bc[j], bc[i]
}

// we need the better ones at the beginning
func (bc Balances) Less(j, i int) bool {
	return bc[i].precision < bc[j].precision ||
		(bc[i].precision == bc[j].precision && bc[i].Weight < bc[j].Weight)
}

func (bc Balances) Sort() {
	sort.Sort(bc)
}

func (bc Balances) GetTotalValue() (total float64) {
	for _, b := range bc {
		if !b.IsExpiredAt(time.Now()) && b.IsActive() {
			total += b.GetValue()
		}
	}
	total = utils.Round(total, globalRoundingDecimals, utils.MetaRoundingMiddle)
	return
}

func (bc Balances) Equal(o Balances) bool {
	if len(bc) != len(o) {
		return false
	}
	bc.Sort()
	o.Sort()
	for i := 0; i < len(bc); i++ {
		if !bc[i].Equal(o[i]) {
			return false
		}
	}
	return true
}

func (bc Balances) Clone() Balances {
	var newChain Balances
	for _, b := range bc {
		newChain = append(newChain, b.Clone())
	}
	return newChain
}

func (bc Balances) GetBalance(uuid string) *Balance {
	for _, balance := range bc {
		if balance.Uuid == uuid {
			return balance
		}
	}
	return nil
}

func (bc Balances) HasBalance(balance *Balance) bool {
	for _, b := range bc {
		if b.Equal(balance) {
			return true
		}
	}
	return false
}

func (bc Balances) SaveDirtyBalances(acc *Account, initBal map[string]float64) {
	savedAccounts := utils.StringSet{}
	for _, b := range bc {
		if b.account == nil || !b.dirty || savedAccounts.Has(b.account.ID) {
			continue
		}
		savedAccounts.Add(b.account.ID)
		if b.account != acc {
			dm.SetAccount(b.account)
		}
		b.account.Publish(initBal)
	}
}

type ValueFactors map[string]float64

func (f ValueFactors) GetValue(category string) float64 {
	if f == nil {
		return 1.0
	}
	if value, ok := f[category]; ok {
		return value
	}
	return 1.0
}

// BalanceSummary represents compressed information about a balance
type BalanceSummary struct {
	UUID     string  // Balance UUID
	ID       string  // Balance ID  if not defined
	Type     string  // *voice, *data, etc
	Initial  float64 // initial value before the debit operation
	Value    float64
	Weight   float64 `json:",omitempty"`
	Disabled bool
	Factors  ValueFactors `json:",omitempty"`
}

// BalanceSummaries is a list of BalanceSummaries
type BalanceSummaries []*BalanceSummary

// BalanceSummaryWithUUD returns a BalanceSummary based on an UUID
func (bs BalanceSummaries) BalanceSummaryWithUUD(bsUUID string) (b *BalanceSummary) {
	for _, blc := range bs {
		if blc.UUID == bsUUID {
			b = blc
			break
		}
	}
	return
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (bl *BalanceSummary) FieldAsInterface(fldPath []string) (val any, err error) {
	if bl == nil || len(fldPath) < 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.UUID:
		return bl.UUID, nil
	case utils.ID:
		return bl.ID, nil
	case utils.Type:
		return bl.Type, nil
	case utils.Value:
		return bl.Value, nil
	case utils.Weight:
		return bl.Weight, nil
	case utils.Disabled:
		return bl.Disabled, nil
	case utils.Initial:
		return bl.Initial, nil
	case utils.Factors:
		if len(fldPath) == 1 {
			return bl.Factors, nil
		}
		return bl.Factors.FieldAsInterface(fldPath[1:])
	}
}

// debit will debit units for call descriptor.
// returns the amount debited within cc
func (b *Balance) debit(cd *CallDescriptor, ub *Account, moneyBalances Balances,
	count, dryRun, debitConnectFee, isUnitBal bool, fltrS *FilterS) (cc *CallCost, err error) {
	if !b.IsActiveAt(cd.TimeStart) || b.GetValue() <= 0 {
		return
	}
	tor := cd.ToR
	if !isUnitBal {
		tor = utils.MetaMonetary
	}

	// Compute the balance factor ahead of time. If the map is nil, or
	// the factor is not found, it will default to 1.
	bFactor := b.Factors.GetValue(cd.ExtraFields[utils.BalanceFactorID])

	if duration, err_ := utils.ParseZeroRatingSubject(tor, b.RatingSubject,
		config.CgrConfig().RalsCfg().BalanceRatingSubject, isUnitBal); err_ == nil {
		// we have *zero based units
		cc = cd.CreateCallCost()
		ts := &TimeSpan{
			TimeStart: cd.TimeStart,
			TimeEnd:   cd.TimeEnd,
		}
		cc.Timespans = TimeSpans{ts}
		ts.RoundToDuration(duration)
		ts.RateInterval = &RateInterval{
			Rating: &RIRate{
				Rates: RateGroups{
					&RGRate{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      duration,
						RateUnit:           duration,
					},
				},
			},
		}
		prefix, destid := b.getMatchingPrefixAndDestID(cd.Destination)
		if prefix == utils.EmptyString {
			prefix = cd.Destination
		}
		if destid == utils.EmptyString {
			destid = utils.MetaAny
		}
		ts.setRatingInfo(&RatingInfo{
			MatchedSubject: b.Uuid,
			MatchedPrefix:  prefix,
			MatchedDestId:  destid,
			RatingPlanId:   utils.MetaNone,
		})
		ts.createIncrementsSlice()
		//log.Printf("CC: %+v", ts)
		for incIndex, inc := range ts.Increments {
			//log.Printf("INCREMENET: %+v", inc)
			amount := float64(inc.Duration)
			if bFactor != 1 {
				amount = utils.Round(amount*bFactor, globalRoundingDecimals, utils.MetaRoundingUp)
			}
			if b.GetValue() >= amount {
				b.SubtractValue(amount)
				inc.BalanceInfo.Unit = &UnitInfo{
					UUID:          b.Uuid,
					ID:            b.ID,
					Value:         b.Value,
					DestinationID: cc.Destination,
					Consumed:      amount,
					Factor:        bFactor,
					Category:      cd.Category,
					ToR:           tor,
					RateInterval:  nil,
				}
				inc.BalanceInfo.AccountID = ub.ID
				inc.Cost = 0
				if count {
					ub.countUnits(amount, tor, cc, b, fltrS)
				}
				continue
			}
			// delete the rest of the unpiad increments/timespans
			if incIndex == 0 {
				// cut the entire current timespan
				return nil, nil
			}
			ts.SplitByIncrement(incIndex)
			if len(cc.Timespans) == 0 {
				cc = nil
			}
			return
		}
		return
	}
	// no rating subject
	//log.Print("B: ", utils.ToJSON(b))
	//log.Printf("}}}}}}} %+v", cd.testCallcost)
	cc, err = b.GetCost(cd, true)
	if err != nil {
		return nil, err
	}

	var debitedConnectFeeBalance Balance
	var connectFeeDebited bool
	//log.Print("cc: " + utils.ToJSON(cc))
	if debitConnectFee {
		// this is the first add, debit the connect fee
		if connectFeeDebited, debitedConnectFeeBalance = ub.DebitConnectionFee(cc, moneyBalances, count, true, fltrS); !connectFeeDebited {
			// balance is blocker
			return nil, nil
		}
	}

	cc.Timespans.Decompress()
	//log.Printf("CallCost In Debit: %+v", cc)
	//for _, ts := range cc.Timespans {
	//	log.Printf("CC_TS: %+v", ts.RateInterval.Rating.Rates[0])
	//}
	for tsIndex, ts := range cc.Timespans {
		if ts.RateInterval == nil {
			utils.Logger.Err(fmt.Sprintf("Nil RateInterval ERROR on TS: %+v, CC: %+v, from CD: %+v", ts, cc, cd))
			return nil, errors.New("timespan with no rate interval assigned")
		}
		if ts.Increments == nil {
			ts.createIncrementsSlice()
		}
		//log.Printf("TS: %+v", ts)

		if tsIndex == 0 &&
			ts.RateInterval.Rating.ConnectFee > 0 &&
			debitConnectFee &&
			cc.deductConnectFee &&
			connectFeeDebited {

			ts.Increments = append([]*Increment{{
				Duration: 0,
				Cost:     ts.RateInterval.Rating.ConnectFee,
				BalanceInfo: &DebitInfo{
					Monetary: &MonetaryInfo{
						UUID:  debitedConnectFeeBalance.Uuid,
						ID:    debitedConnectFeeBalance.ID,
						Value: debitedConnectFeeBalance.Value,
					},
					AccountID: ub.ID,
				},
			}}, ts.Increments...)
		}

		maxCost, strategy := ts.RateInterval.GetMaxCost()
		//log.Printf("Timing: %+v", ts.RateInterval.Timing)
		//log.Printf("RGRate: %+v", ts.RateInterval.Rating)
		for incIndex, inc := range ts.Increments {
			// check standard subject tags
			//log.Printf("INC: %+v", inc)

			if tsIndex == 0 &&
				incIndex == 0 &&
				ts.RateInterval.Rating.ConnectFee > 0 &&
				cc.deductConnectFee &&
				connectFeeDebited {
				// go to nextincrement
				continue
			}
			if cd.MaxCostSoFar >= maxCost {
				if strategy == utils.MetaMaxCostFree {
					inc.Cost = 0.0
					inc.BalanceInfo.Monetary = &MonetaryInfo{
						UUID:  b.Uuid,
						ID:    b.ID,
						Value: b.Value,
					}
					inc.BalanceInfo.AccountID = ub.ID
					if b.RatingSubject != utils.EmptyString || isUnitBal {
						inc.BalanceInfo.Monetary.RateInterval = ts.RateInterval
					}
					if count {
						ub.countUnits(inc.Cost, utils.MetaMonetary, cc, b, fltrS)
					}

					//log.Printf("TS: %+v", cc.Cost)
					// go to nextincrement
					continue
				} else if cc.maxCostDisconect = strategy == utils.MetaMaxCostDisconnect; cc.maxCostDisconect && dryRun {
					// cut the entire current timespan
					if incIndex == 0 {
						// cut the entire current timespan
						cc.Timespans = cc.Timespans[:tsIndex]
					} else {
						ts.SplitByIncrement(incIndex)
						cc.Timespans = cc.Timespans[:tsIndex+1]
					}
					return
				}
			}

			// debit minutes and money
			amount := float64(inc.Duration)
			cost := inc.Cost

			canDebitCost := b.GetValue() >= cost
			var moneyBal *Balance
			if isUnitBal {
				if bFactor != 1 {
					amount = utils.Round(amount*bFactor, globalRoundingDecimals, utils.MetaRoundingUp)
				}
				for _, mb := range moneyBalances {
					if mb.GetValue() >= cost {
						moneyBal = mb
						break
					}
				}
				if cost != 0 && moneyBal == nil && (!dryRun || ub.AllowNegative) { // Fix for issue #685
					if !ub.AllowNegative {
						utils.Logger.Warning(fmt.Sprintf("<RALs> Going negative on account %s with AllowNegative: false", cd.GetAccountKey()))
					}
					moneyBal = ub.GetDefaultMoneyBalance()
				}
				canDebitCost = b.GetValue() >= amount && (moneyBal != nil || cost == 0)
			}
			if !canDebitCost {
				// delete the rest of the unpaid increments/timespans
				if incIndex == 0 {
					// cut the entire current timespan
					cc.Timespans = cc.Timespans[:tsIndex]
				} else {
					ts.SplitByIncrement(incIndex)
					cc.Timespans = cc.Timespans[:tsIndex+1]
				}
				if len(cc.Timespans) == 0 {
					cc = nil
				}
				return
			}

			if isUnitBal { // unit balance
				b.SubtractValue(amount)
				inc.BalanceInfo.Unit = &UnitInfo{
					UUID:          b.Uuid,
					ID:            b.ID,
					Value:         b.Value,
					DestinationID: cc.Destination,
					Consumed:      amount,
					Factor:        bFactor,
					Category:      cc.Category,
					ToR:           cc.ToR,
					RateInterval:  ts.RateInterval,
				}
				inc.BalanceInfo.AccountID = ub.ID
				if cost != 0 {
					moneyBal.SubtractValue(cost)
					inc.BalanceInfo.Monetary = &MonetaryInfo{
						UUID:  moneyBal.Uuid,
						ID:    moneyBal.ID,
						Value: moneyBal.Value,
					}
					cd.MaxCostSoFar += cost
				}
				if count {
					ub.countUnits(amount, cc.ToR, cc, b, fltrS)
					if cost != 0 {
						ub.countUnits(cost, utils.MetaMonetary, cc, moneyBal, fltrS)
					}
				}
			} else { // monetary balance
				b.SubtractValue(cost)
				cd.MaxCostSoFar += cost
				inc.BalanceInfo.Monetary = &MonetaryInfo{
					UUID:  b.Uuid,
					ID:    b.ID,
					Value: b.Value,
				}
				inc.BalanceInfo.AccountID = ub.ID
				if b.RatingSubject != "" {
					inc.BalanceInfo.Monetary.RateInterval = ts.RateInterval
				}
				if count {
					ub.countUnits(cost, utils.MetaMonetary, cc, b, fltrS)
				}
			}
		}
	}
	if !isUnitBal && len(cc.Timespans) == 0 {
		cc = nil
	}
	return
}

func (bc Balances) String() string {
	return utils.ToJSON(bc)
}

func (bc Balances) FieldAsInterface(fldPath []string) (val any, err error) {
	if bc == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}

	if fldPath[0] == utils.GetTotalValue {
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return bc.GetTotalValue(), nil
	}
	for _, at := range bc {
		if at.ID == fldPath[0] {
			if len(fldPath) == 1 {
				return at, nil
			}
			return at.FieldAsInterface(fldPath[1:])
		}
	}
	var indx int
	if indx, err = strconv.Atoi(fldPath[0]); err != nil {
		return
	}
	if len(bc) <= indx {
		return nil, utils.ErrNotFound
	}
	c := bc[indx]
	if len(fldPath) == 1 {
		return c, nil
	}
	return c.FieldAsInterface(fldPath[1:])
}

func (bc Balances) FieldAsString(fldPath []string) (val string, err error) {
	var iface any
	iface, err = bc.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

func (b *Balance) String() string {
	return utils.ToJSON(b)
}

func (b *Balance) FieldAsInterface(fldPath []string) (val any, err error) {
	if b == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		opath, indx := utils.GetPathIndexString(fldPath[0])
		if indx != nil {
			switch opath {
			case utils.DestinationIDs:
				val, has := b.DestinationIDs[*indx]
				if !has || len(fldPath) != 1 {
					return nil, utils.ErrNotFound
				}
				return val, nil
			case utils.Categories:
				val, has := b.Categories[*indx]
				if !has || len(fldPath) != 1 {
					return nil, utils.ErrNotFound
				}
				return val, nil
			case utils.SharedGroups:
				val, has := b.SharedGroups[*indx]
				if !has || len(fldPath) != 1 {
					return nil, utils.ErrNotFound
				}
				return val, nil
			case utils.TimingIDs:
				val, has := b.TimingIDs[*indx]
				if !has || len(fldPath) != 1 {
					return nil, utils.ErrNotFound
				}
				return val, nil
			case utils.Timings:
				var idx int
				if idx, err = strconv.Atoi(*indx); err != nil {
					return
				}
				if len(b.Timings) <= idx {
					return nil, utils.ErrNotFound
				}
				tm := b.Timings[idx]
				if len(fldPath) == 1 {
					return tm, nil
				}
				return tm.FieldAsInterface(fldPath[1:])
			case utils.Factors:
				val, has := b.Factors[*indx]
				if !has || len(fldPath) != 1 {
					return nil, utils.ErrNotFound
				}
				return val, nil
			}
		}
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.Uuid:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return b.Uuid, nil
	case utils.ID:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return b.ID, nil
	case utils.Value:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return b.Value, nil
	case utils.ExpirationDate:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return b.ExpirationDate, nil
	case utils.Weight:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return b.Weight, nil
	case utils.DestinationIDs:
		if len(fldPath) == 1 {
			return b.DestinationIDs, nil
		}
		return b.DestinationIDs.FieldAsInterface(fldPath[1:])
	case utils.RatingSubject:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return b.RatingSubject, nil
	case utils.Categories:
		if len(fldPath) == 1 {
			return b.Categories, nil
		}
		return b.Categories.FieldAsInterface(fldPath[1:])
	case utils.SharedGroups:
		if len(fldPath) == 1 {
			return b.SharedGroups, nil
		}
		return b.SharedGroups.FieldAsInterface(fldPath[1:])
	case utils.Timings:
		if len(fldPath) == 1 {
			return b.Timings, nil
		}
		for _, tm := range b.Timings {
			if tm.ID == fldPath[1] {
				if len(fldPath) == 2 {
					return tm, nil
				}
				return tm.FieldAsInterface(fldPath[2:])
			}
		}
		return nil, utils.ErrNotFound
	case utils.TimingIDs:
		if len(fldPath) == 1 {
			return b.TimingIDs, nil
		}
		return b.TimingIDs.FieldAsInterface(fldPath[1:])
	case utils.Disabled:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return b.Disabled, nil
	case utils.Factors:
		if len(fldPath) == 1 {
			return b.Factors, nil
		}
		return b.Factors.FieldAsInterface(fldPath[1:])
	case utils.Blocker:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return b.Blocker, nil
	}
}

func (b *Balance) FieldAsString(fldPath []string) (val string, err error) {
	var iface any
	iface, err = b.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

func (f ValueFactors) String() string {
	return utils.ToJSON(f)
}

func (f ValueFactors) FieldAsInterface(fldPath []string) (val any, err error) {
	if f == nil || len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	c, has := f[fldPath[0]]
	if !has {
		return nil, utils.ErrNotFound
	}
	return c, nil
}

func (f ValueFactors) FieldAsString(fldPath []string) (val string, err error) {
	var iface any
	iface, err = f.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}
