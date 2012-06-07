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
	"sort"
	"bytes"
	"encoding/gob"
	"sync"
)

const (
	UB_TYPE_POSTPAID = "postpaid"
	UB_TYPE_PREPAID  = "prepaid"
)

var (
	storageGetter StorageGetter
)

/*
Structure containing information about user's credit (minutes, cents, sms...).'
*/
type UserBalance struct {
	Id            string
	Type          string // prepaid-postpaid
	BalanceMap    map[string]float64
	UnitsCounters []*UnitsCounter
	TariffPlanId  string
	tariffPlan    *TariffPlan
	MinuteBuckets []*MinuteBucket
	mux           sync.RWMutex
}

/*
Error type for overflowed debit methods.
*/
type AmountTooBig byte

func (a AmountTooBig) Error() string {
	return "Amount excedes balance!"
}

/*
Structure to store minute buckets according to weight, precision or price.
*/
type bucketsorter []*MinuteBucket

func (bs bucketsorter) Len() int {
	return len(bs)
}

func (bs bucketsorter) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs bucketsorter) Less(j, i int) bool {
	return bs[i].Weight < bs[j].Weight ||
		bs[i].precision < bs[j].precision ||
		bs[i].Price > bs[j].Price
}

/*
Serializes the user balance for the storage. Used for key-value storage.
*/
func (ub *UserBalance) store() (result string) {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(ub)
	return buf.String()
}

/*
De-serializes the user balance for the storage. Used for key-value storage.
*/
func (ub *UserBalance) restore(input string) {
	gob.NewDecoder(bytes.NewBuffer([]byte(input))).Decode(ub)
}

/*
Returns the tariff plan loading it from the storage if necessary.
*/
func (ub *UserBalance) getTariffPlan() (tp *TariffPlan, err error) {
	if ub.tariffPlan == nil && ub.TariffPlanId != "" {
		ub.tariffPlan, err = storageGetter.GetTariffPlan(ub.TariffPlanId)
	}
	return ub.tariffPlan, err
}

/*
Returns user's available minutes for the specified destination
*/
func (ub *UserBalance) getSecondsForPrefix(prefix string) (seconds float64, bucketList bucketsorter) {
	if len(ub.MinuteBuckets) == 0 {
		// log.Print("There are no minute buckets to check for user: ", ub.Id)
		return
	}
	for _, mb := range ub.MinuteBuckets {
		d := mb.getDestination()
		if d == nil {
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
	sort.Sort(bucketList) // sorts the buckets according to priority, precision or price
	credit := ub.BalanceMap[CREDIT]
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
	ub.mux.Lock()
	defer ub.mux.Unlock()
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
	ub.mux.Lock()
	defer ub.mux.Unlock()
	avaliableNbSeconds, bucketList := ub.getSecondsForPrefix(prefix)
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
	ub.mux.Lock()
	defer ub.mux.Unlock()
	if ub.BalanceMap[SMS] < amount {
		return ub.BalanceMap[SMS], new(AmountTooBig)
	}
	ub.BalanceMap[SMS] -= amount

	storageGetter.SetUserBalance(ub)
	return ub.BalanceMap[SMS], nil
}

/*
Adds the specified amount of seconds.
*/
func (ub *UserBalance) addReceivedCallSeconds(direction, tor, destination string, amount float64) error {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	for 
	ub.ReceivedCallSeconds += amount
	if tariffPlan, err := ub.getTariffPlan(); tariffPlan != nil && err == nil {
		if ub.ReceivedCallSeconds >= tariffPlan.ReceivedCallSecondsLimit {
			ub.ReceivedCallSeconds -= tariffPlan.ReceivedCallSecondsLimit
			if tariffPlan.RecivedCallBonus != nil { // apply the bonus
				ub.BalanceMap[CREDIT] += tariffPlan.RecivedCallBonus.Credit
				ub.BalanceMap[SMS] += tariffPlan.RecivedCallBonus.SmsCredit
				ub.BalanceMap[TRAFFIC] += tariffPlan.RecivedCallBonus.Traffic
				if tariffPlan.RecivedCallBonus.MinuteBucket != nil {
					for _, mb := range ub.MinuteBuckets {
						if mb.DestinationId == tariffPlan.RecivedCallBonus.MinuteBucket.DestinationId {
							mb.Seconds += tariffPlan.RecivedCallBonus.MinuteBucket.Seconds
						}
					}
				}
			}
		}
	}
	return storageGetter.SetUserBalance(ub)
}

/*
Resets the user balance items to their tariff plan values.
*/
func (ub *UserBalance) resetUserBalance() (err error) {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	if tp, err := ub.getTariffPlan(storageGetter); err == nil {
		ub.SmsCredit = tp.SmsCredit
		ub.Traffic = tp.Traffic
		ub.MinuteBuckets = make([]*MinuteBucket, 0)
		for _, bucket := range tp.MinuteBuckets {
			mb := &MinuteBucket{Seconds: bucket.Seconds,
				Priority:      bucket.Priority,
				Price:         bucket.Price,
				DestinationId: bucket.DestinationId}
			ub.MinuteBuckets = append(ub.MinuteBuckets, mb)
		}
		err = storageGetter.SetUserBalance(ub)
	}
	return
}
