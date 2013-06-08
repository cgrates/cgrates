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
	//"log"
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
	// Price types
	PERCENT  = "PERCENT"
	ABSOLUTE = "ABSOLUTE"
)

/*
Structure containing information about user's credit (minutes, cents, sms...).'
*/
type UserBalance struct {
	Id                                  string
	Type                                string // prepaid-postpaid
	BalanceMap                          map[string]float64
	MinuteBuckets                       []*MinuteBucket
	UnitCounters                        []*UnitsCounter
	ActionTriggers                      ActionTriggerPriotityList
	actionTriggersTag, actionTimingsTag string // used only for loading
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
	credit = ub.BalanceMap[CREDIT+OUTBOUND]
	if len(ub.MinuteBuckets) == 0 {
		// Logger.Debug("There are no minute buckets to check for user: ", ub.Id)
		return
	}
	for _, mb := range ub.MinuteBuckets {
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
				credit -= mb.Seconds * mb.Price
			}
		} else {
			if mb.Price > 0 { // debit the money if the bucket has price
				credit -= amount * mb.Price
			}
			break
		}
		if credit < 0 {
			break
		}
	}
	if credit < 0 {
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

/*
Debits some amount of user's specified balance. Returns the remaining credit in user's balance.
*/
func (ub *UserBalance) debitBalance(balanceId string, amount float64, count bool) float64 {
	if count {
		ub.countUnits(&Action{BalanceId: balanceId, Direction: OUTBOUND, Units: amount})
	}
	ub.BalanceMap[balanceId+OUTBOUND] -= amount
	return ub.BalanceMap[balanceId+OUTBOUND]
}

// Scans the action trigers and execute the actions for which trigger is met
func (ub *UserBalance) executeActionTriggers() {
	ub.ActionTriggers.Sort()
	for _, at := range ub.ActionTriggers {
		if at.Executed {
			// trigger is marked as executed, so skipp it until
			// the next reset (see RESET_TRIGGERS action type)
			continue
		}
		for _, uc := range ub.UnitCounters {
			if uc.BalanceId == at.BalanceId {
				if at.BalanceId == MINUTES && at.DestinationId != "" { // last check adds safty
					for _, mb := range uc.MinuteBuckets {
						if mb.DestinationId == at.DestinationId && mb.Seconds >= at.ThresholdValue {
							// run the actions
							at.Execute(ub)
						}
					}
				} else {
					if uc.Units >= at.ThresholdValue {
						// run the actions
						at.Execute(ub)
					}
				}
			}
		}
	}
}

// Mark all action trigers as ready for execution
func (ub *UserBalance) resetActionTriggers() {
	for _, at := range ub.ActionTriggers {
		at.Executed = false
	}
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
	ub.executeActionTriggers()
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
