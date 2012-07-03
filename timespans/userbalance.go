/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	// "log"
	"sync"
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

var (
	storageGetter       StorageGetter
	userBalancesRWMutex sync.RWMutex
)

/*
Structure containing information about user's credit (minutes, cents, sms...).'
*/
type UserBalance struct {
	Id             string
	Type           string // prepaid-postpaid
	BalanceMap     map[string]float64
	MinuteBuckets  []*MinuteBucket
	UnitsCounters  []*UnitsCounter
	ActionTriggers ActionTriggerPriotityList
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
	if len(ub.MinuteBuckets) == 0 {
		// log.Print("There are no minute buckets to check for user: ", ub.Id)
		return
	}
	for _, mb := range ub.MinuteBuckets {
		d, err := GetDestination(mb.DestinationId)
		if err != nil {
			continue
		}
		contains, precision := d.containsPrefix(prefix)
		if contains {
			mb.precision = precision
			if mb.Seconds > 0 {
				bucketList = append(bucketList, mb)
			}
		}
	}
	bucketList.Sort() // sorts the buckets according to priority, precision or price
	credit = ub.BalanceMap[CREDIT]
	for _, mb := range bucketList {
		s := mb.GetSecondsForCredit(credit)
		credit -= s * mb.Price
		seconds += s
	}
	return
}

/*
Debits some amount of user's money credit. Returns the remaining credit in user's balance.
*/
func (ub *UserBalance) debitMoneyBalance(amount float64) float64 {
	ub.BalanceMap[CREDIT] -= amount
	storageGetter.SetUserBalance(ub)
	return ub.BalanceMap[CREDIT]
}

/*
Debits the received amount of seconds from user's minute buckets.
All the appropriate buckets will be debited until all amount of minutes is consumed.
If the amount is bigger than the sum of all seconds in the minute buckets than nothing will be
debited and an error will be returned.
*/
func (ub *UserBalance) debitMinutesBalance(amount float64, prefix string) error {
	avaliableNbSeconds, _, bucketList := ub.getSecondsForPrefix(prefix)
	if avaliableNbSeconds < amount {
		return new(AmountTooBig)
	}
	credit := ub.BalanceMap[CREDIT]
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
	ub.BalanceMap[CREDIT] = credit // credit is > 0

	for _, mb := range bucketList {
		if mb.Seconds < amount {
			amount -= mb.Seconds
			mb.Seconds = 0
		} else {
			mb.Seconds -= amount
			break
		}
	}
	storageGetter.SetUserBalance(ub)
	return nil
}

/*
Debits some amount of user's SMS balance. Returns the remaining SMS in user's balance.
If the amount is bigger than the balance than nothing wil be debited and an error will be returned
*/
func (ub *UserBalance) debitSMSBuget(amount float64) (float64, error) {
	if ub.BalanceMap[SMS] < amount {
		return ub.BalanceMap[SMS], new(AmountTooBig)
	}
	ub.BalanceMap[SMS] -= amount

	storageGetter.SetUserBalance(ub)
	return ub.BalanceMap[SMS], nil
}

// Adds the minutes from the received minute bucket to an existing bucket if the destination
// is the same or ads the minutye bucket to the list if none matches.
func (ub *UserBalance) addMinuteBucket(newMb *MinuteBucket) {
	found := false
	for _, mb := range ub.MinuteBuckets {
		if mb.DestinationId == newMb.DestinationId {
			mb.Seconds += newMb.Seconds
			found = true
			break
		}
	}
	if !found {
		ub.MinuteBuckets = append(ub.MinuteBuckets, newMb)
	}
}

// Scans the action trigers and execute the actions for which trigger is met
func (ub *UserBalance) executeActionTriggers() {
	ub.ActionTriggers.Sort()
	for _, at := range ub.ActionTriggers {
		if at.executed {
			// trigger is marked as executed, so skipp it until
			// the next reset (see RESET_TRIGGERS action type)
			continue
		}
		for _, uc := range ub.UnitsCounters {
			if uc.BalanceId == at.BalanceId {
				if at.BalanceId == MINUTES && at.DestinationId != "" { // last check adds safty
					for _, mb := range uc.MinuteBuckets {
						if mb.DestinationId == at.DestinationId && mb.Seconds >= at.ThresholdValue {
							// run the actions
							at.Execute(ub)
						}
					}
				}
				if uc.Units >= at.ThresholdValue {
					// run the actions					
					at.Execute(ub)
				}
			}
		}
	}
}

// Mark all action trigers as ready for execution
func (ub *UserBalance) resetActionTriggers() {
	for _, at := range ub.ActionTriggers {
		at.executed = false
	}
}

/*
Adds the specified amount of seconds.
*/
// func (ub *UserBalance) addReceivedCallSeconds(direction, tor, destination string, amount float64) error {
// 	ub.ReceivedCallSeconds += amount
// 	if tariffPlan, err := ub.getTariffPlan(); tariffPlan != nil && err == nil {
// 		if ub.ReceivedCallSeconds >= tariffPlan.ReceivedCallSecondsLimit {
// 			ub.ReceivedCallSeconds -= tariffPlan.ReceivedCallSecondsLimit
// 			if tariffPlan.RecivedCallBonus != nil { // apply the bonus
// 				ub.BalanceMap[CREDIT] += tariffPlan.RecivedCallBonus.Credit
// 				ub.BalanceMap[SMS] += tariffPlan.RecivedCallBonus.SmsCredit
// 				ub.BalanceMap[TRAFFIC] += tariffPlan.RecivedCallBonus.Traffic
// 				if tariffPlan.RecivedCallBonus.MinuteBucket != nil {
// 					for _, mb := range ub.MinuteBuckets {
// 						if mb.DestinationId == tariffPlan.RecivedCallBonus.MinuteBucket.DestinationId {
// 							mb.Seconds += tariffPlan.RecivedCallBonus.MinuteBucket.Seconds
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return storageGetter.SetUserBalance(ub)
// }

/*
Resets the user balance items to their tariff plan values.
*/
/*func (ub *UserBalance) resetUserBalance() (err error) {
	if tp, err := ub.getAccountActions(); err == nil {
		for k, _ := range ub.BalanceMap {
			ub.BalanceMap[k] = tp.BalanceMap[k]
		}
		ub.MinuteBuckets = make([]*MinuteBucket, 0)
		for _, bucket := range tp.MinuteBuckets {
			mb := &MinuteBucket{Seconds: bucket.Seconds,
				Weight:        bucket.Weight,
				Price:         bucket.Price,
				DestinationId: bucket.DestinationId}
			ub.MinuteBuckets = append(ub.MinuteBuckets, mb)
		}
		err = storageGetter.SetUserBalance(ub)
	}
	return
}
*/
// Amount of a trafic of a certain type
type UnitsCounter struct {
	Direction     string
	BalanceId     string
	Units         float64
	Weight        float64
	MinuteBuckets []*MinuteBucket
}

// Structure to store actions according to weight
type countersorter []*UnitsCounter

func (s countersorter) Len() int {
	return len(s)
}

func (s countersorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s countersorter) Less(i, j int) bool {
	return s[i].Weight < s[j].Weight
}
