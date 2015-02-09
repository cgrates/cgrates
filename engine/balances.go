/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Can hold different units as seconds or monetary
type Balance struct {
	Uuid           string //system wide unique
	Id             string // account wide unique
	Value          float64
	ExpirationDate time.Time
	Weight         float64
	DestinationId  string
	RatingSubject  string
	Category       string
	SharedGroup    string
	Timings        []*RITiming
	TimingIDs      string
	precision      int
	account        *Account // used to store ub reference for shared balances
	dirty          bool
}

func (b *Balance) Equal(o *Balance) bool {
	if b.DestinationId == "" {
		b.DestinationId = utils.ANY
	}
	if o.DestinationId == "" {
		o.DestinationId = utils.ANY
	}
	return b.Id == o.Id &&
		b.ExpirationDate.Equal(o.ExpirationDate) &&
		b.Weight == o.Weight &&
		b.DestinationId == o.DestinationId &&
		b.RatingSubject == o.RatingSubject &&
		b.Category == o.Category &&
		b.SharedGroup == o.SharedGroup
}

func (b *Balance) MatchFilter(o *Balance) bool {
	if o.Id != "" {
		return b.Id == o.Id
	}
	if b.DestinationId == "" {
		b.DestinationId = utils.ANY
	}
	if o.DestinationId == "" {
		o.DestinationId = utils.ANY
	}
	return (o.ExpirationDate.IsZero() || b.ExpirationDate.Equal(o.ExpirationDate)) &&
		(o.Weight == 0 || b.Weight == o.Weight) &&
		(o.DestinationId == "" || b.DestinationId == o.DestinationId) &&
		(o.RatingSubject == "" || b.RatingSubject == o.RatingSubject) &&
		(o.Category == "" || b.Category == o.Category) &&
		(o.SharedGroup == "" || b.SharedGroup == o.SharedGroup)
}

// the default balance has no destinationid, Expirationdate or ratesubject
func (b *Balance) IsDefault() bool {
	return (b.DestinationId == "" || b.DestinationId == utils.ANY) &&
		b.RatingSubject == "" &&
		b.Category == "" &&
		b.ExpirationDate.IsZero() &&
		b.SharedGroup == "" &&
		b.Weight == 0
}

func (b *Balance) IsExpired() bool {
	return !b.ExpirationDate.IsZero() && b.ExpirationDate.Before(time.Now())
}

func (b *Balance) IsActive() bool {
	return b.IsActiveAt(time.Now())
}

func (b *Balance) IsActiveAt(t time.Time) bool {
	if len(b.Timings) == 0 {
		return true
	}
	for _, tim := range b.Timings {
		if tim.IsActiveAt(t, false) {
			return true
		}
	}
	return false
}

func (b *Balance) MatchCategory(category string) bool {
	return b.Category == "" || b.Category == category
}

func (b *Balance) HasDestination() bool {
	return b.DestinationId != "" && b.DestinationId != utils.ANY
}

func (b *Balance) MatchDestination(destinationId string) bool {
	return !b.HasDestination() || b.DestinationId == destinationId
}

func (b *Balance) MatchActionTrigger(at *ActionTrigger) bool {
	if at.BalanceId != "" {
		return b.Id == at.BalanceId
	}
	matchesDestination := true
	if at.BalanceDestinationId != "" {
		matchesDestination = (at.BalanceDestinationId == b.DestinationId)
	}
	matchesExpirationDate := true
	if !at.BalanceExpirationDate.IsZero() {
		matchesExpirationDate = (at.BalanceExpirationDate.Equal(b.ExpirationDate))
	}
	matchesWeight := true
	if at.BalanceWeight > 0 {
		matchesWeight = (at.BalanceWeight == b.Weight)
	}
	matchesRatingSubject := true
	if at.BalanceRatingSubject != "" {
		matchesRatingSubject = (at.BalanceRatingSubject == b.RatingSubject)
	}

	matchesSharedGroup := true
	if at.BalanceSharedGroup != "" {
		matchesSharedGroup = (at.BalanceSharedGroup == b.SharedGroup)
	}

	return matchesDestination &&
		matchesExpirationDate &&
		matchesWeight &&
		matchesRatingSubject &&
		matchesSharedGroup
}

func (b *Balance) Clone() *Balance {
	return &Balance{
		Uuid:           b.Uuid,
		Id:             b.Id,
		Value:          b.Value, // this value is in seconds
		DestinationId:  b.DestinationId,
		ExpirationDate: b.ExpirationDate,
		Weight:         b.Weight,
		RatingSubject:  b.RatingSubject,
		Category:       b.Category,
		SharedGroup:    b.SharedGroup,
		TimingIDs:      b.TimingIDs,
		Timings:        b.Timings, // should not be a problem with aliasing
	}
}

// Returns the available number of seconds for a specified credit
func (b *Balance) GetMinutesForCredit(origCD *CallDescriptor, initialCredit float64) (duration time.Duration, credit float64) {
	cd := origCD.Clone()
	availableDuration := time.Duration(b.Value) * time.Second
	duration = availableDuration
	credit = initialCredit
	cc, err := b.GetCost(cd, false)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
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
func (b *Balance) GetCost(cd *CallDescriptor, getStandarIfEmpty bool) (*CallCost, error) {
	// testing only
	if cd.test_callcost != nil {
		return cd.test_callcost, nil
	}
	if b.RatingSubject != "" && !strings.HasPrefix(b.RatingSubject, utils.ZERO_RATING_SUBJECT_PREFIX) {
		origSubject := cd.Subject
		cd.Subject = b.RatingSubject
		origAccount := cd.Account
		cd.Account = cd.Subject
		cd.RatingInfos = nil
		cc, err := cd.GetCost()
		// restor orig values
		cd.Subject = origSubject
		cd.Account = origAccount
		return cc, err
	}
	if getStandarIfEmpty {
		cd.RatingInfos = nil
		return cd.GetCost()
	} else {
		cc := cd.CreateCallCost()
		cc.Cost = 0
		return cc, nil
	}
}

func (b *Balance) SubstractAmount(amount float64) {
	b.Value -= amount
	b.Value = utils.Round(b.Value, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
	b.dirty = true
}

func (b *Balance) DebitUnits(cd *CallDescriptor, count bool, ub *Account, moneyBalances BalanceChain) (cc *CallCost, err error) {
	if !b.IsActiveAt(cd.TimeStart) || b.Value <= 0 {
		return
	}
	if duration, err := utils.ParseZeroRatingSubject(b.RatingSubject); err == nil {
		// we have *zero based units
		cc = cd.CreateCallCost()
		cc.Timespans = append(cc.Timespans, &TimeSpan{
			TimeStart: cd.TimeStart,
			TimeEnd:   cd.TimeEnd,
		})

		seconds := duration.Seconds()
		amount := seconds
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
		ts.createIncrementsSlice()
		//log.Printf("CC: %+v", ts)
		for incIndex, inc := range ts.Increments {
			//log.Printf("INCREMENET: %+v", inc)
			if seconds == 1 {
				amount = inc.Duration.Seconds()
			}
			if b.Value >= amount {
				b.SubstractAmount(amount)
				inc.BalanceInfo.UnitBalanceUuid = b.Uuid
				inc.BalanceInfo.AccountId = ub.Id
				inc.UnitInfo = &UnitInfo{cc.Destination, amount, cc.TOR}
				inc.Cost = 0
				inc.paid = true
				if count {
					ub.countUnits(&Action{BalanceType: cc.TOR, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
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
		cc.Timespans.Decompress()
		//log.Printf("CC: %+v", cc)
		if err != nil {
			return nil, fmt.Errorf("Error getting new cost for balance subject: %v", err)
		}
		for tsIndex, ts := range cc.Timespans {
			if ts.Increments == nil {
				ts.createIncrementsSlice()
			}
			//log.Printf("TS: %+v", ts)
			for incIndex, inc := range ts.Increments {
				// debit minutes and money
				seconds := inc.Duration.Seconds()
				cost := inc.Cost
				//log.Printf("INC: %+v", inc)
				var moneyBal *Balance
				for _, mb := range moneyBalances {
					if mb.Value >= cost {
						moneyBal = mb
						break
					}
				}
				if (cost == 0 || moneyBal != nil) && b.Value >= seconds {
					b.SubstractAmount(seconds)
					inc.BalanceInfo.UnitBalanceUuid = b.Uuid
					inc.BalanceInfo.AccountId = ub.Id
					inc.UnitInfo = &UnitInfo{cc.Destination, seconds, cc.TOR}
					if cost != 0 {
						inc.BalanceInfo.MoneyBalanceUuid = moneyBal.Uuid
						moneyBal.SubstractAmount(cost)
					}
					inc.paid = true
					if count {
						ub.countUnits(&Action{BalanceType: cc.TOR, Direction: cc.Direction, Balance: &Balance{Value: seconds, DestinationId: cc.Destination}})
						if cost != 0 {
							ub.countUnits(&Action{BalanceType: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: cost, DestinationId: cc.Destination}})
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

func (b *Balance) DebitMoney(cd *CallDescriptor, count bool, ub *Account) (cc *CallCost, err error) {
	if !b.IsActiveAt(cd.TimeStart) || b.Value <= 0 {
		return
	}
	//log.Printf("}}}}}}} %+v", cd.test_callcost)
	cc, err = b.GetCost(cd, true)
	cc.Timespans.Decompress()
	//log.Printf("CallCost In Debit: %+v", cc)
	//for _, ts := range cc.Timespans {
	//	log.Printf("CC_TS: %+v", ts.RateInterval.Rating.Rates[0])
	//}
	if err != nil {
		return nil, fmt.Errorf("Error getting new cost for balance subject: %v", err)
	}
	for tsIndex, ts := range cc.Timespans {
		if ts.Increments == nil {
			ts.createIncrementsSlice()
		}
		//log.Printf("TS: %+v", ts)
		for incIndex, inc := range ts.Increments {
			// check standard subject tags
			//log.Printf("INC: %+v", inc)
			amount := inc.Cost
			if b.Value >= amount {
				b.SubstractAmount(amount)
				inc.BalanceInfo.MoneyBalanceUuid = b.Uuid
				inc.BalanceInfo.AccountId = ub.Id
				inc.paid = true
				if count {
					ub.countUnits(&Action{BalanceType: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
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
	//log.Printf("END: %+v", cd.test_callcost)
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
			total += b.Value
		}
	}
	return
}

func (bc BalanceChain) Debit(amount float64) float64 {
	bc.Sort()
	for i, b := range bc {
		if b.IsExpired() {
			continue
		}
		if b.Value >= amount || i == len(bc)-1 { // if last one go negative
			b.SubstractAmount(amount)
			break
		}
		b.Value = 0
		amount -= b.Value
	}
	return bc.GetTotalValue()
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
	for _, b := range bc {
		// TODO: check if the account was not already saved ?
		if b.account != nil && b.account != acc && b.dirty {
			accountingStorage.SetAccount(b.account)
		}
	}
}
