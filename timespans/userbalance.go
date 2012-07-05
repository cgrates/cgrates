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
	"strconv"
	"strings"
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
	UnitCounters   []*UnitsCounter
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

	return ub.BalanceMap[CREDIT]
}

// Debit seconds from specified minute bucket
func (ub *UserBalance) debitMinuteBucket(newMb *MinuteBucket) error {
	for _, mb := range ub.MinuteBuckets {
		if mb.Equal(newMb) {
			mb.Seconds -= newMb.Seconds
		}
	}
	return nil
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
	return nil
}

/*
Debits some amount of user's SMS balance. Returns the remaining SMS in user's balance.
If the amount is bigger than the balance than nothing wil be debited and an error will be returned
*/
func (ub *UserBalance) debitSMSBalance(amount float64) (float64, error) {
	if ub.BalanceMap[SMS] < amount {
		return ub.BalanceMap[SMS], new(AmountTooBig)
	}
	ub.BalanceMap[SMS] -= amount

	return ub.BalanceMap[SMS], nil
}

func (ub *UserBalance) debitTrafficBalance(amount float64) (float64, error) {
	if ub.BalanceMap[SMS] < amount {
		return ub.BalanceMap[SMS], new(AmountTooBig)
	}
	ub.BalanceMap[SMS] -= amount

	return ub.BalanceMap[SMS], nil
}

func (ub *UserBalance) debitTrafficTimeBalance(amount float64) (float64, error) {
	if ub.BalanceMap[SMS] < amount {
		return ub.BalanceMap[SMS], new(AmountTooBig)
	}
	ub.BalanceMap[SMS] -= amount

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
		for _, uc := range ub.UnitCounters {
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

func (ub *UserBalance) CountUnits(a *Action) {
	var unitsCounter *UnitsCounter
	for _, uc := range ub.UnitCounters {
		if uc.BalanceId == a.BalanceId {
			unitsCounter = uc
			break
		}
	}
	if unitsCounter != nil {
		if unitsCounter.BalanceId == MINUTES && a.MinuteBucket != nil {
			unitsCounter.addMinuteBucket(a.MinuteBucket)
			goto TRIGGERS
		}
		unitsCounter.Units += a.Units
	}
TRIGGERS:
	ub.executeActionTriggers()
}

/*
Serializes the user balance for the storage. Used for key-value storages.
*/
func (ub *UserBalance) store() (result string) {
	result += ub.Id + "|"
	result += ub.Type + "|"
	for k, v := range ub.BalanceMap {
		result += k + ":" + strconv.FormatFloat(v, 'f', -1, 64) + "#"
	}
	result = strings.TrimRight(result, "#") + "|"
	for _, mb := range ub.MinuteBuckets {
		result += mb.store() + "#"
	}
	result = strings.TrimRight(result, "#") + "|"
	for _, uc := range ub.UnitCounters {
		result += uc.store() + "#"
	}
	result = strings.TrimRight(result, "#") + "|"
	for _, at := range ub.ActionTriggers {
		result += at.store() + "#"
	}
	result = strings.TrimRight(result, "#")
	return
}

/*
De-serializes the user balance for the storage. Used for key-value storages.
*/
func (ub *UserBalance) restore(input string) {
	elements := strings.Split(input, "|")
	ub.Id = elements[0]
	ub.Type = elements[1]
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]float64, 0)
	}
	for _, maps := range strings.Split(elements[2], "#") {
		kv := strings.Split(maps, ":")
		if len(kv) != 2 {
			continue
		}
		value, _ := strconv.ParseFloat(kv[1], 64)
		ub.BalanceMap[kv[0]] = value
	}
	for _, mbs := range strings.Split(elements[3], "#") {
		if mbs == "" {
			continue
		}
		mb := &MinuteBucket{}
		mb.restore(mbs)
		ub.MinuteBuckets = append(ub.MinuteBuckets, mb)
	}
	for _, ucs := range strings.Split(elements[4], "#") {
		if ucs == "" {
			continue
		}
		uc := &UnitsCounter{}
		uc.restore(ucs)
		ub.UnitCounters = append(ub.UnitCounters, uc)
	}
	for _, ats := range strings.Split(elements[5], "#") {
		if ats == "" {
			continue
		}
		at := &ActionTrigger{}
		at.restore(ats)
		ub.ActionTriggers = append(ub.ActionTriggers, at)
	}
}
