/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package rater

import (
	"errors"
	"github.com/cgrates/cgrates/utils"
	"sort"
	"strings"
	"time"
)

const (
	UB_TYPE_POSTPAID = "postpaid"
	UB_TYPE_PREPAID  = "prepaid"
	// Direction type
	INBOUND  = "IN"
	OUTBOUND = "OUT"
	// Balance types
	CREDIT       = "MONETARY"
	SMS          = "SMS"
	TRAFFIC      = "INTERNET"
	TRAFFIC_TIME = "INTERNET_TIME"
	MINUTES      = "MINUTES"
)

/*
Structure containing information about user's credit (minutes, cents, sms...).'
*/
type UserBalance struct {
	Id             string
	Type           string // prepaid-postpaid
	BalanceMap     map[string]BalanceChain
	MinuteBuckets  []*MinuteBucket
	UnitCounters   []*UnitsCounter
	ActionTriggers ActionTriggerPriotityList
}

type Balance struct {
	Id             string
	Value          float64
	ExpirationDate time.Time
	Weight         float64
}

func (b *Balance) Equal(o *Balance) bool {
	return b.ExpirationDate.Equal(o.ExpirationDate) ||
		b.Weight == o.Weight
}

func (b *Balance) IsExpired() bool {
	return !b.ExpirationDate.IsZero() && b.ExpirationDate.Before(time.Now())
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
	return bc[i].Weight < bc[j].Weight
}

func (bc BalanceChain) Sort() {
	sort.Sort(bc)
}

func (bc BalanceChain) GetTotalValue() (total float64) {
	for _, b := range bc {
		if !b.IsExpired() {
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
			b.Value -= amount
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

/*
Error type for overflowed debit methods.
*/
type AmountTooBig struct{}

func (a AmountTooBig) Error() string {
	return "Amount excedes balance!"
}

/*
Returns user's available minutes for the specified destination
*/
func (ub *UserBalance) getSecondsForPrefix(prefix string) (seconds, credit float64, bucketList bucketsorter) {
	credit = ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue()
	if len(ub.MinuteBuckets) == 0 {
		// Logger.Debug("There are no minute buckets to check for user: ", ub.Id)
		return
	}
	for _, mb := range ub.MinuteBuckets {
		if mb.IsExpired() {
			continue
		}
		d, err := GetDestination(mb.DestinationId)
		if err != nil {
			continue
		}
		if precision, ok := d.containsPrefix(prefix); ok {
			mb.precision = precision
			if mb.Seconds > 0 {
				bucketList = append(bucketList, mb)
			}
		}
	}
	bucketList.Sort() // sorts the buckets according to priority, precision or price
	for _, mb := range bucketList {
		s := mb.GetSecondsForCredit(credit)
		credit -= s * mb.Price
		seconds += s
	}
	return
}

// Debit seconds from specified minute bucket
func (ub *UserBalance) debitMinuteBucket(newMb *MinuteBucket) error {
	if newMb == nil {
		return errors.New("Nil minute bucket!")
	}
	found := false
	for _, mb := range ub.MinuteBuckets {
		if mb.IsExpired() {
			continue
		}
		if mb.Equal(newMb) {
			mb.Seconds -= newMb.Seconds
			found = true
			break
		}
	}
	// if it is not found and the Seconds are negative (topup)
	// then we add it to the list
	if !found && newMb.Seconds <= 0 {
		newMb.Seconds = -newMb.Seconds
		ub.MinuteBuckets = append(ub.MinuteBuckets, newMb)
	}
	return nil
}

/*
Debits the received amount of seconds from user's minute buckets.
All the appropriate buckets will be debited until all amount of minutes is consumed.
If the amount is bigger than the sum of all seconds in the minute buckets than nothing will be
debited and an error will be returned.
*/
func (ub *UserBalance) debitMinutesBalance(amount float64, prefix string, count bool) error {
	if count {
		ub.countUnits(&Action{BalanceId: MINUTES, Direction: OUTBOUND, MinuteBucket: &MinuteBucket{Seconds: amount, DestinationId: prefix}})
	}
	avaliableNbSeconds, _, bucketList := ub.getSecondsForPrefix(prefix)
	if avaliableNbSeconds < amount {
		return new(AmountTooBig)
	}
	credit := ub.BalanceMap[CREDIT+OUTBOUND]
	// calculating money debit
	// this is needed because if the credit is less then the amount needed to be debited
	// we need to keep everything in place and return an error
	for _, mb := range bucketList {
		if mb.Seconds < amount {
			if mb.Price > 0 { // debit the money if the bucket has price
				credit.Debit(mb.Seconds * mb.Price)
			}
		} else {
			if mb.Price > 0 { // debit the money if the bucket has price
				credit.Debit(amount * mb.Price)
			}
			break
		}
		if credit.GetTotalValue() < 0 {
			break
		}
	}
	if credit.GetTotalValue() < 0 {
		return new(AmountTooBig)
	}
	ub.BalanceMap[CREDIT+OUTBOUND] = credit // credit is > 0

	for _, mb := range bucketList {
		if mb.Seconds < amount {
			amount -= mb.Seconds
			mb.Seconds = 0
		} else {
			mb.Seconds -= amount
			break
		}
	}
	return nil
}

// Debits some amount of user's specified balance adding the balance if it does not exists.
// Returns the remaining credit in user's balance.
func (ub *UserBalance) debitBalanceAction(a *Action) float64 {
	newBalance := &Balance{
		Id:             utils.GenUUID(),
		ExpirationDate: a.ExpirationDate,
		Weight:         a.Weight,
	}
	found := false
	id := a.BalanceId + a.Direction
	for _, b := range ub.BalanceMap[id] {
		if b.Equal(newBalance) {
			b.Value -= a.Units
			found = true
		}
	}
	if !found {
		newBalance.Value -= a.Units
		ub.BalanceMap[id] = append(ub.BalanceMap[id], newBalance)
	}
	return ub.BalanceMap[a.BalanceId+OUTBOUND].GetTotalValue()
}

/*
Debits some amount of user's specified balance. Returns the remaining credit in user's balance.
*/
func (ub *UserBalance) debitBalance(balanceId string, amount float64, count bool) float64 {
	if count {
		ub.countUnits(&Action{BalanceId: balanceId, Direction: OUTBOUND, Units: amount})
	}
	ub.BalanceMap[balanceId+OUTBOUND].Debit(amount)
	return ub.BalanceMap[balanceId+OUTBOUND].GetTotalValue()
}

// Scans the action trigers and execute the actions for which trigger is met
func (ub *UserBalance) executeActionTriggers(a *Action) {
	ub.ActionTriggers.Sort()
	for _, at := range ub.ActionTriggers {
		if at.Executed {
			// trigger is marked as executed, so skipp it until
			// the next reset (see RESET_TRIGGERS action type)
			continue
		}
		if a != nil && (at.BalanceId != a.BalanceId ||
			at.Direction != a.Direction ||
			(a.MinuteBucket != nil &&
				(at.ThresholdType != a.MinuteBucket.PriceType ||
					at.ThresholdValue != a.MinuteBucket.Price))) {
			continue
		}
		if strings.Contains(at.ThresholdType, "COUNTER") {
			for _, uc := range ub.UnitCounters {
				if uc.BalanceId == at.BalanceId {
					if at.BalanceId == MINUTES && at.DestinationId != "" { // last check adds safety
						for _, mb := range uc.MinuteBuckets {
							if strings.Contains(at.ThresholdType, "MAX") {
								if mb.DestinationId == at.DestinationId && mb.Seconds >= at.ThresholdValue {
									// run the actions
									at.Execute(ub)
								}
							} else { //MIN
								if mb.DestinationId == at.DestinationId && mb.Seconds <= at.ThresholdValue {
									// run the actions
									at.Execute(ub)
								}
							}
						}
					} else {
						if strings.Contains(at.ThresholdType, "MAX") {
							if uc.Units >= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						} else { //MIN
							if uc.Units <= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						}
					}
				}
			}
		} else { // BALANCE
			for _, b := range ub.BalanceMap[at.BalanceId] {
				if at.BalanceId == MINUTES && at.DestinationId != "" { // last check adds safety
					for _, mb := range ub.MinuteBuckets {
						if strings.Contains(at.ThresholdType, "MAX") {
							if mb.DestinationId == at.DestinationId && mb.Seconds >= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						} else { //MIN
							if mb.DestinationId == at.DestinationId && mb.Seconds <= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						}
					}
				} else {
					if strings.Contains(at.ThresholdType, "MAX") {
						if b.Value >= at.ThresholdValue {
							// run the actions
							at.Execute(ub)
						}
					} else { //MIN
						if b.Value <= at.ThresholdValue {
							// run the actions
							at.Execute(ub)
						}
					}
				}
			}
		}
	}
}

// Mark all action trigers as ready for execution
// If the action is not nil it acts like a filter
func (ub *UserBalance) resetActionTriggers(a *Action) {
	for _, at := range ub.ActionTriggers {
		if a != nil && (at.BalanceId != a.BalanceId ||
			at.Direction != a.Direction ||
			(a.MinuteBucket != nil &&
				(at.ThresholdType != a.MinuteBucket.PriceType ||
					at.ThresholdValue != a.MinuteBucket.Price))) {
			continue
		}
		at.Executed = false
	}
	ub.executeActionTriggers(a)
}

// Returns the unit counter that matches the specified action type
func (ub *UserBalance) getUnitCounter(a *Action) *UnitsCounter {
	for _, uc := range ub.UnitCounters {
		direction := a.Direction
		if direction == "" {
			direction = OUTBOUND
		}
		if uc.BalanceId == a.BalanceId && uc.Direction == direction {
			return uc
		}
	}
	return nil
}

// Increments the counter for the type specified in the received Action
// with the actions values
func (ub *UserBalance) countUnits(a *Action) {
	unitsCounter := ub.getUnitCounter(a)
	// if not found add the counter
	if unitsCounter == nil {
		direction := a.Direction
		if direction == "" {
			direction = OUTBOUND
		}
		unitsCounter = &UnitsCounter{BalanceId: a.BalanceId, Direction: direction}
		ub.UnitCounters = append(ub.UnitCounters, unitsCounter)
	}
	if a.BalanceId == MINUTES && a.MinuteBucket != nil {
		unitsCounter.addMinutes(a.MinuteBucket.Seconds, a.MinuteBucket.DestinationId)
	} else {
		unitsCounter.Units += a.Units
	}
	ub.executeActionTriggers(nil)
}

// Create minute counters for all triggered actions that have actions operating on minute buckets
func (ub *UserBalance) initMinuteCounters() {
	ucTempMap := make(map[string]*UnitsCounter, 2)
	for _, at := range ub.ActionTriggers {
		acs, err := storageGetter.GetActions(at.ActionsId)
		if err != nil {
			continue
		}
		for _, a := range acs {
			if a.MinuteBucket != nil {
				direction := at.Direction
				if direction == "" {
					direction = OUTBOUND
				}
				uc, exists := ucTempMap[direction]
				if !exists {
					uc = &UnitsCounter{BalanceId: MINUTES, Direction: direction}
					ucTempMap[direction] = uc
					uc.MinuteBuckets = bucketsorter{}
					ub.UnitCounters = append(ub.UnitCounters, uc)
				}
				uc.MinuteBuckets = append(uc.MinuteBuckets, a.MinuteBucket.Clone())
				uc.MinuteBuckets.Sort()
			}
		}
	}
}

func (ub *UserBalance) CleanExpiredBalancesAndBuckets() {
	for key, _ := range ub.BalanceMap {
		bm := ub.BalanceMap[key]
		for i := 0; i < len(bm); i++ {
			if bm[i].IsExpired() {
				// delete it
				bm = append(bm[:i], bm[i+1:]...)
			}
		}
		ub.BalanceMap[key] = bm
	}
	for i := 0; i < len(ub.MinuteBuckets); i++ {
		if ub.MinuteBuckets[i].IsExpired() {
			ub.MinuteBuckets = append(ub.MinuteBuckets[:i], ub.MinuteBuckets[i+1:]...)
		}
	}
}
