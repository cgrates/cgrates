/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
)

// Can hold different units as seconds or monetary
type Balance struct {
	Uuid           string //system wide unique
	ID             string // account wide unique
	Value          float64
	Directions     utils.StringMap
	ExpirationDate time.Time
	Weight         float64
	DestinationIDs utils.StringMap
	RatingSubject  string
	Categories     utils.StringMap
	SharedGroups   utils.StringMap
	Timings        []*RITiming
	TimingIDs      utils.StringMap
	Disabled       bool
	Factor         ValueFactor
	Blocker        bool
	precision      int
	account        *Account // used to store ub reference for shared balances
	dirty          bool
}

func (b *Balance) Equal(o *Balance) bool {
	if len(b.DestinationIDs) == 0 {
		b.DestinationIDs = utils.StringMap{utils.ANY: true}
	}
	if len(o.DestinationIDs) == 0 {
		o.DestinationIDs = utils.StringMap{utils.ANY: true}
	}
	return b.Uuid == o.Uuid &&
		b.ID == o.ID &&
		b.ExpirationDate.Equal(o.ExpirationDate) &&
		b.Weight == o.Weight &&
		b.DestinationIDs.Equal(o.DestinationIDs) &&
		b.Directions.Equal(o.Directions) &&
		b.RatingSubject == o.RatingSubject &&
		b.Categories.Equal(o.Categories) &&
		b.SharedGroups.Equal(o.SharedGroups) &&
		b.Disabled == o.Disabled &&
		b.Blocker == o.Blocker
}

func (b *Balance) MatchFilter(o *BalanceFilter, skipIds bool) bool {
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
		(o.DestinationIDs == nil || b.DestinationIDs.Includes(*o.DestinationIDs)) &&
		(o.Directions == nil || b.Directions.Includes(*o.Directions)) &&
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
		(o.Directions == nil || b.Directions.Equal(*o.Directions)) &&
		(o.Categories == nil || b.Categories.Equal(*o.Categories)) &&
		(o.TimingIDs == nil || b.TimingIDs.Equal(*o.TimingIDs)) &&
		(o.SharedGroups == nil || b.SharedGroups.Equal(*o.SharedGroups)) &&
		(o.RatingSubject == nil || b.RatingSubject == *o.RatingSubject)
}

// the default balance has standard Id
func (b *Balance) IsDefault() bool {
	return b.ID == utils.META_DEFAULT
}

func (b *Balance) IsExpired() bool {
	// check if it expires in the next second
	return !b.ExpirationDate.IsZero() && b.ExpirationDate.Before(time.Now().Add(1*time.Second))
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
	return len(b.Categories) == 0 || b.Categories[category] == true
}

func (b *Balance) HasDestination() bool {
	return len(b.DestinationIDs) > 0 && b.DestinationIDs[utils.ANY] == false
}

func (b *Balance) HasDirection() bool {
	return len(b.Directions) > 0
}

func (b *Balance) MatchDestination(destinationID string) bool {
	return !b.HasDestination() || b.DestinationIDs[destinationID] == true
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
		dirty:          b.dirty,
	}
	if b.DestinationIDs != nil {
		n.DestinationIDs = b.DestinationIDs.Clone()
	}
	if b.Directions != nil {
		n.Directions = b.Directions.Clone()
	}
	return n
}

func (b *Balance) getMatchingPrefixAndDestID(dest string) (prefix, destId string) {
	if len(b.DestinationIDs) != 0 && b.DestinationIDs[utils.ANY] == false {
		for _, p := range utils.SplitPrefix(dest, MIN_PREFIX_MATCH) {
			if x, err := cache2go.Get(utils.DESTINATION_PREFIX + p); err == nil {
				destIDs := x.(map[interface{}]struct{})
				for dID := range destIDs {
					if b.DestinationIDs[dID.(string)] == true {
						return p, dID.(string)
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
				if rate/rateUnit.Seconds() > cd.MaxRate/cd.MaxRateUnit.Seconds() {
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
	if b.RatingSubject != "" && !strings.HasPrefix(b.RatingSubject, utils.ZERO_RATING_SUBJECT_PREFIX) {
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
	} else {
		cc := cd.CreateCallCost()
		cc.Cost = 0
		return cc, nil
	}
}

func (b *Balance) GetValue() float64 {
	return b.Value
}

func (b *Balance) AddValue(amount float64) {
	b.SetValue(b.GetValue() + amount)
}

func (b *Balance) SubstractValue(amount float64) {
	b.SetValue(b.GetValue() - amount)
}

func (b *Balance) SetValue(amount float64) {
	b.Value = amount
	b.Value = utils.Round(b.GetValue(), globalRoundingDecimals, utils.ROUNDING_MIDDLE)
	b.dirty = true
}

func (b *Balance) SetDirty() {
	b.dirty = true
}

func (b *Balance) debitUnits(cd *CallDescriptor, ub *Account, moneyBalances BalanceChain, count bool, dryRun, debitConnectFee bool) (cc *CallCost, err error) {
	if !b.IsActiveAt(cd.TimeStart) || b.GetValue() <= 0 {
		return
	}
	if duration, err := utils.ParseZeroRatingSubject(b.RatingSubject); err == nil {
		// we have *zero based units
		cc = cd.CreateCallCost()
		cc.Timespans = append(cc.Timespans, &TimeSpan{
			TimeStart: cd.TimeStart,
			TimeEnd:   cd.TimeEnd,
		})

		ts := cc.Timespans[0]
		ts.RoundToDuration(duration)
		ts.RateInterval = &RateInterval{
			Rating: &RIRate{
				Rates: RateGroups{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      duration,
						RateUnit:           duration,
					},
				},
			},
		}
		prefix, destid := b.getMatchingPrefixAndDestID(cd.Destination)
		if prefix == "" {
			prefix = cd.Destination
		}
		if destid == "" {
			destid = utils.ANY
		}
		ts.setRatingInfo(&RatingInfo{
			MatchedSubject: b.Uuid,
			MatchedPrefix:  prefix,
			MatchedDestId:  destid,
			RatingPlanId:   utils.META_NONE,
		})
		ts.createIncrementsSlice()
		//log.Printf("CC: %+v", ts)
		for incIndex, inc := range ts.Increments {
			//log.Printf("INCREMENET: %+v", inc)

			amount := inc.Duration.Seconds()
			if b.Factor != nil {
				amount = utils.Round(amount/b.Factor.GetValue(cd.TOR), globalRoundingDecimals, utils.ROUNDING_UP)
			}
			if b.GetValue() >= amount {
				b.SubstractValue(amount)
				inc.BalanceInfo.UnitBalanceUuid = b.Uuid
				inc.BalanceInfo.AccountId = ub.Id
				inc.UnitInfo = &UnitInfo{cc.Destination, amount, cc.TOR}
				inc.Cost = 0
				inc.paid = true
				if count {
					ub.countUnits(amount, cc.TOR, cc, b)
				}
			} else {
				inc.paid = false
				// delete the rest of the unpiad increments/timespans
				if incIndex == 0 {
					// cat the entire current timespan
					cc.Timespans = nil
				} else {
					ts.SplitByIncrement(incIndex)
				}
				if len(cc.Timespans) == 0 {
					cc = nil
				}
				return cc, nil
			}
		}
	} else {
		// get the cost from balance
		//log.Printf("::::::: %+v", cd)
		cc, err = b.GetCost(cd, true)
		if err != nil {
			return nil, err
		}
		if debitConnectFee {
			// this is the first add, debit the connect fee
			ub.DebitConnectionFee(cc, moneyBalances, count)
		}
		cc.Timespans.Decompress()
		//log.Printf("CC: %+v", cc)

		for tsIndex, ts := range cc.Timespans {
			if ts.Increments == nil {
				ts.createIncrementsSlice()
			}

			if ts.RateInterval == nil {
				utils.Logger.Err(fmt.Sprintf("Nil RateInterval ERROR on TS: %+v, CC: %+v, from CD: %+v", ts, cc, cd))
				return nil, errors.New("timespan with no rate interval assigned")
			}
			maxCost, strategy := ts.RateInterval.GetMaxCost()
			for incIndex, inc := range ts.Increments {
				// debit minutes and money
				amount := inc.Duration.Seconds()
				if b.Factor != nil {
					amount = utils.Round(amount/b.Factor.GetValue(cd.TOR), globalRoundingDecimals, utils.ROUNDING_UP)
				}
				cost := inc.Cost
				inc.paid = false
				if strategy == utils.MAX_COST_DISCONNECT && cd.MaxCostSoFar >= maxCost {
					// cat the entire current timespan
					cc.maxCostDisconect = true
					if dryRun {
						if incIndex == 0 {
							// cat the entire current timespan
							cc.Timespans = cc.Timespans[:tsIndex]
						} else {
							ts.SplitByIncrement(incIndex)
							cc.Timespans = cc.Timespans[:tsIndex+1]
						}
						return cc, nil
					}
				}
				if strategy == utils.MAX_COST_FREE && cd.MaxCostSoFar >= maxCost {
					cost, inc.Cost = 0.0, 0.0
					inc.BalanceInfo.MoneyBalanceUuid = b.Uuid
					inc.BalanceInfo.AccountId = ub.Id
					inc.paid = true
					if count {
						ub.countUnits(cost, utils.MONETARY, cc, b)
					}
					// go to nextincrement
					continue
				}
				var moneyBal *Balance
				for _, mb := range moneyBalances {
					if mb.GetValue() >= cost {
						moneyBal = mb
						break
					}
				}
				if (cost == 0 || moneyBal != nil) && b.GetValue() >= amount {
					b.SubstractValue(amount)
					inc.BalanceInfo.UnitBalanceUuid = b.Uuid
					inc.BalanceInfo.AccountId = ub.Id
					inc.UnitInfo = &UnitInfo{cc.Destination, amount, cc.TOR}
					if cost != 0 {
						inc.BalanceInfo.MoneyBalanceUuid = moneyBal.Uuid
						moneyBal.SubstractValue(cost)
						cd.MaxCostSoFar += cost
					}
					inc.paid = true
					if count {
						ub.countUnits(amount, cc.TOR, cc, b)
						if cost != 0 {
							ub.countUnits(cost, utils.MONETARY, cc, moneyBal)
						}
					}
				} else {
					inc.paid = false
					// delete the rest of the unpiad increments/timespans
					if incIndex == 0 {
						// cat the entire current timespan
						cc.Timespans = cc.Timespans[:tsIndex]
					} else {
						ts.SplitByIncrement(incIndex)
						cc.Timespans = cc.Timespans[:tsIndex+1]
					}
					if len(cc.Timespans) == 0 {
						cc = nil
					}
					return cc, nil
				}
			}
		}
	}
	return
}

func (b *Balance) debitMoney(cd *CallDescriptor, ub *Account, moneyBalances BalanceChain, count bool, dryRun, debitConnectFee bool) (cc *CallCost, err error) {
	if !b.IsActiveAt(cd.TimeStart) || b.GetValue() <= 0 {
		return
	}
	//log.Printf("}}}}}}} %+v", cd.testCallcost)
	cc, err = b.GetCost(cd, true)
	if err != nil {
		return nil, err
	}

	if debitConnectFee {
		// this is the first add, debit the connect fee
		ub.DebitConnectionFee(cc, moneyBalances, count)
	}

	cc.Timespans.Decompress()
	//log.Printf("CallCost In Debit: %+v", cc)
	//for _, ts := range cc.Timespans {
	//	log.Printf("CC_TS: %+v", ts.RateInterval.Rating.Rates[0])
	//}
	for tsIndex, ts := range cc.Timespans {
		if ts.Increments == nil {
			ts.createIncrementsSlice()
		}
		//log.Printf("TS: %+v", ts)
		if ts.RateInterval == nil {
			utils.Logger.Err(fmt.Sprintf("Nil RateInterval ERROR on TS: %+v, CC: %+v, from CD: %+v", ts, cc, cd))
			return nil, errors.New("timespan with no rate interval assigned")
		}
		maxCost, strategy := ts.RateInterval.GetMaxCost()
		//log.Printf("Timing: %+v", ts.RateInterval.Timing)
		//log.Printf("Rate: %+v", ts.RateInterval.Rating)
		for incIndex, inc := range ts.Increments {
			// check standard subject tags
			//log.Printf("INC: %+v", inc)
			amount := inc.Cost
			inc.paid = false
			if strategy == utils.MAX_COST_DISCONNECT && cd.MaxCostSoFar >= maxCost {
				// cat the entire current timespan
				cc.maxCostDisconect = true
				if dryRun {
					if incIndex == 0 {
						// cat the entire current timespan
						cc.Timespans = cc.Timespans[:tsIndex]
					} else {
						ts.SplitByIncrement(incIndex)
						cc.Timespans = cc.Timespans[:tsIndex+1]
					}
					return cc, nil
				}
			}
			if strategy == utils.MAX_COST_FREE && cd.MaxCostSoFar >= maxCost {
				amount, inc.Cost = 0.0, 0.0
				inc.BalanceInfo.MoneyBalanceUuid = b.Uuid
				inc.BalanceInfo.AccountId = ub.Id
				inc.paid = true
				if count {
					ub.countUnits(amount, utils.MONETARY, cc, b)
				}

				//log.Printf("TS: %+v", cc.Cost)
				// go to nextincrement
				continue
			}

			if b.GetValue() >= amount {
				b.SubstractValue(amount)
				cd.MaxCostSoFar += amount
				inc.BalanceInfo.MoneyBalanceUuid = b.Uuid
				inc.BalanceInfo.AccountId = ub.Id
				inc.paid = true
				if count {
					ub.countUnits(amount, utils.MONETARY, cc, b)
				}
			} else {
				inc.paid = false
				// delete the rest of the unpiad increments/timespans
				if incIndex == 0 {
					// cat the entire current timespan
					cc.Timespans = cc.Timespans[:tsIndex]
				} else {
					ts.SplitByIncrement(incIndex)
					cc.Timespans = cc.Timespans[:tsIndex+1]
				}
				if len(cc.Timespans) == 0 {
					cc = nil
				}
				return cc, nil
			}
		}
	}
	//log.Printf("END: %+v", cd.testCallcost)
	if len(cc.Timespans) == 0 {
		cc = nil
	}
	return cc, nil
}

/*
Structure to store minute buckets according to weight, precision or price.
*/
type BalanceChain []*Balance

func (bc BalanceChain) Len() int {
	return len(bc)
}

func (bc BalanceChain) Swap(i, j int) {
	bc[i], bc[j] = bc[j], bc[i]
}

// we need the better ones at the beginning
func (bc BalanceChain) Less(j, i int) bool {
	return bc[i].precision < bc[j].precision ||
		(bc[i].precision == bc[j].precision && bc[i].Weight < bc[j].Weight)

}

func (bc BalanceChain) Sort() {
	sort.Sort(bc)
}

func (bc BalanceChain) GetTotalValue() (total float64) {
	for _, b := range bc {
		if !b.IsExpired() && b.IsActive() {
			total += b.GetValue()
		}
	}
	return
}

func (bc BalanceChain) Equal(o BalanceChain) bool {
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

func (bc BalanceChain) Clone() BalanceChain {
	var newChain BalanceChain
	for _, b := range bc {
		newChain = append(newChain, b.Clone())
	}
	return newChain
}

func (bc BalanceChain) GetBalance(uuid string) *Balance {
	for _, balance := range bc {
		if balance.Uuid == uuid {
			return balance
		}
	}
	return nil
}

func (bc BalanceChain) HasBalance(balance *Balance) bool {
	for _, b := range bc {
		if b.Equal(balance) {
			return true
		}
	}
	return false
}

func (bc BalanceChain) SaveDirtyBalances(acc *Account) {
	savedAccounts := make(map[string]bool)
	for _, b := range bc {
		if b.dirty {
			// publish event
			accountId := ""
			allowNegative := ""
			disabled := ""
			if b.account != nil { // only publish modifications for balances with account set
				//utils.LogStack()
				accountId = b.account.Id
				allowNegative = strconv.FormatBool(b.account.AllowNegative)
				disabled = strconv.FormatBool(b.account.Disabled)
				Publish(CgrEvent{
					"EventName":            utils.EVT_ACCOUNT_BALANCE_MODIFIED,
					"Uuid":                 b.Uuid,
					"Id":                   b.ID,
					"Value":                strconv.FormatFloat(b.Value, 'f', -1, 64),
					"ExpirationDate":       b.ExpirationDate.String(),
					"Weight":               strconv.FormatFloat(b.Weight, 'f', -1, 64),
					"DestinationIDs":       b.DestinationIDs.String(),
					"Directions":           b.Directions.String(),
					"RatingSubject":        b.RatingSubject,
					"Categories":           b.Categories.String(),
					"SharedGroups":         b.SharedGroups.String(),
					"TimingIDs":            b.TimingIDs.String(),
					"Account":              accountId,
					"AccountAllowNegative": allowNegative,
					"AccountDisabled":      disabled,
				})
			}
		}
		if b.account != nil && b.account != acc && b.dirty && savedAccounts[b.account.Id] == false {
			accountingStorage.SetAccount(b.account)
			savedAccounts[b.account.Id] = true
		}
	}
}

type ValueFactor map[string]float64

func (f ValueFactor) GetValue(tor string) float64 {
	if value, ok := f[tor]; ok {
		return value
	}
	return 1.0
}
