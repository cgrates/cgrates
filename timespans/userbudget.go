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
	"strconv"
	"strings"
	"sync"
)

const (
	UB_TYPE_POSTPAID = "postpaid"
	UB_TYPE_PREPAID  = "prepaid"
)

/*
Structure conatining information about user's credit (minutes, cents, sms...).'
*/
type UserBudget struct {
	Id                 string
	Type               string // prepaid-postpaid
	BalanceMap           map[string]float64
	OutboundVolumes    []*TrafficVolume
	InboundVolumes     []*TrafficVolume
	ResetDayOfTheMonth int
	TariffPlanId       string
	tariffPlan         *TariffPlan
	MinuteBuckets      []*MinuteBucket
	mux                sync.RWMutex
}

/*
Error type for overflowed debit methods.
*/
type AmountTooBig byte

func (a AmountTooBig) Error() string {
	return "Amount excedes budget!"
}

/*
Structure to store minute buckets according to priority, precision or price.
*/
type bucketsorter []*MinuteBucket

func (bs bucketsorter) Len() int {
	return len(bs)
}

func (bs bucketsorter) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs bucketsorter) Less(j, i int) bool {
	return bs[i].Priority < bs[j].Priority ||
		bs[i].precision < bs[j].precision ||
		bs[i].Price > bs[j].Price
}

/*
Serializes the user budget for the storage. Used for key-value storages.
*/
func (ub *UserBudget) store() (result string) {
	result += ub.Type + ";"
	result += strconv.FormatFloat(ub.Credit, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(ub.SmsCredit, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(ub.Traffic, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(ub.VolumeDiscountSeconds, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(ub.ReceivedCallSeconds, 'f', -1, 64) + ";"
	result += strconv.Itoa(ub.ResetDayOfTheMonth) + ";"
	result += ub.TariffPlanId
	if ub.MinuteBuckets != nil {
		result += ";"
	}
	for i, mb := range ub.MinuteBuckets {
		if i > 0 {
			result += ","
		}
		result += mb.store()
	}
	return
}

/*
De-serializes the user budget for the storage. Used for key-value storages.
*/
func (ub *UserBudget) restore(input string) {
	elements := strings.Split(input, ";")
	ub.Type = elements[0]
	ub.Credit, _ = strconv.ParseFloat(elements[1], 64)
	ub.SmsCredit, _ = strconv.ParseFloat(elements[2], 64)
	ub.Traffic, _ = strconv.ParseFloat(elements[3], 64)
	ub.VolumeDiscountSeconds, _ = strconv.ParseFloat(elements[4], 64)
	ub.ReceivedCallSeconds, _ = strconv.ParseFloat(elements[5], 64)
	ub.ResetDayOfTheMonth, _ = strconv.Atoi(elements[6])
	ub.TariffPlanId = elements[7]
	if len(elements) > 8 {
		for _, mbs := range strings.Split(elements[8], ",") {
			mb := &MinuteBucket{}
			mb.restore(mbs)
			ub.MinuteBuckets = append(ub.MinuteBuckets, mb)
		}
	}
}

/*
Returns the tariff plan loading it from the storage if necessary.
*/
func (ub *UserBudget) getTariffPlan(storage StorageGetter) (tp *TariffPlan, err error) {
	if ub.tariffPlan == nil && ub.TariffPlanId != "" {
		ub.tariffPlan, err = storage.GetTariffPlan(ub.TariffPlanId)
	}
	return ub.tariffPlan, err
}

/*
Returns thevolume discount procentage according to the nuber of acumulated volume discount seconds.
*/
func (ub *UserBudget) getVolumeDiscount(storage StorageGetter) (float64, error) {
	tariffPlan, err := ub.getTariffPlan(storage)
	if err != nil || tariffPlan == nil {
		return 0.0, err
	}
	thresholds := len(tariffPlan.VolumeDiscountThresholds)
	for i, vd := range tariffPlan.VolumeDiscountThresholds {
		if ub.VolumeDiscountSeconds >= vd.Volume &&
			(i > thresholds-2 || ub.VolumeDiscountSeconds < tariffPlan.VolumeDiscountThresholds[i+1].Volume) {
			return vd.Discount, nil
		}
	}
	return 0, nil
}

/*
Returns user's avaliable minutes for the specified destination
*/
func (ub *UserBudget) getSecondsForPrefix(sg StorageGetter, prefix string) (seconds float64, bucketList bucketsorter) {
	if len(ub.MinuteBuckets) == 0 {
		// log.Print("There are no minute buckets to check for user: ", ub.Id)
		return
	}
	for _, mb := range ub.MinuteBuckets {
		d := mb.getDestination(sg)
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
	credit := ub.Credit
	for _, mb := range bucketList {
		s := mb.GetSecondsForCredit(credit)
		credit -= s * mb.Price
		seconds += s
	}
	return
}

/*
Debits some amount of user's money credit. Returns the remaining credit in user's budget.
*/
func (ub *UserBudget) debitMoneyBudget(sg StorageGetter, amount float64) float64 {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	ub.Credit -= amount
	sg.SetUserBudget(ub)
	return ub.Credit
}

/*
Debits the recived amount of seconds from user's minute buckets.
All the appropriate buckets will be debited until all amount of minutes is consumed.
If the amount is bigger than the sum of all seconds in the minute buckets than nothing will be
debited and an error will be returned.
*/
func (ub *UserBudget) debitMinutesBudget(sg StorageGetter, amount float64, prefix string) error {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	avaliableNbSeconds, bucketList := ub.getSecondsForPrefix(sg, prefix)
	if avaliableNbSeconds < amount {
		return new(AmountTooBig)
	}
	credit := ub.Credit
	// calculating money debit
	// this is needed because if the credit is less then the amount needed to be debited
	// we need to keep everithing in place and return an error
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
	ub.Credit = credit // credit is > 0
	for _, mb := range bucketList {
		if mb.Seconds < amount {
			amount -= mb.Seconds
			mb.Seconds = 0
		} else {
			mb.Seconds -= amount
			break
		}
	}
	sg.SetUserBudget(ub)
	return nil
}

/*
Debits some amount of user's SMS budget. Returns the remaining SMS in user's budget.
If the amount is bigger than the budget than nothing wil be debited and an error will be returned
*/
func (ub *UserBudget) debitSMSBuget(sg StorageGetter, amount float64) (float64, error) {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	if ub.SmsCredit < amount {
		return ub.SmsCredit, new(AmountTooBig)
	}
	ub.SmsCredit -= amount

	sg.SetUserBudget(ub)
	return ub.SmsCredit, nil
}

/*
Adds the the specified amount to volume discount seconds budget.
*/
func (ub *UserBudget) addVolumeDiscountSeconds(sg StorageGetter, amount float64) error {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	ub.VolumeDiscountSeconds += amount
	return sg.SetUserBudget(ub)
}

/*
Resets the volume discounts seconds (sets zero value).
*/
func (ub *UserBudget) resetVolumeDiscountSeconds(sg StorageGetter) error {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	ub.VolumeDiscountSeconds = 0
	return sg.SetUserBudget(ub)
}

/*
Adds the spcifeied amount of seconds to the reci.
*/
func (ub *UserBudget) addReceivedCallSeconds(sg StorageGetter, amount float64) error {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	ub.ReceivedCallSeconds += amount
	if tariffPlan, err := ub.getTariffPlan(sg); tariffPlan != nil && err == nil {
		if ub.ReceivedCallSeconds >= tariffPlan.ReceivedCallSecondsLimit {
			ub.ReceivedCallSeconds -= tariffPlan.ReceivedCallSecondsLimit
			if tariffPlan.RecivedCallBonus != nil { // apply the bonus
				ub.Credit += tariffPlan.RecivedCallBonus.Credit
				ub.SmsCredit += tariffPlan.RecivedCallBonus.SmsCredit
				ub.Traffic += tariffPlan.RecivedCallBonus.Traffic
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
	return sg.SetUserBudget(ub)
}

/*
Resets the user budget items to their tariff plan values.
*/
func (ub *UserBudget) resetUserBudget(sg StorageGetter) (err error) {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	if tp, err := ub.getTariffPlan(sg); err == nil {
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
		err = sg.SetUserBudget(ub)
	}
	return
}