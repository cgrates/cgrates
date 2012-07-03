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
	"errors"
	"fmt"
	"log"
	"math"
	"time"
)

const (
	// the minimum length for a destination prefix to be matched.
	MinPrefixLength     = 2
	RecursionMaxDepth   = 4
	FallbackDestination = "fallback" // the string to be used to mark the fallback destination
)

/*
Utility function for rounding a float to a certain number of decimals (not present in math).
*/
func round(val float64, prec int) float64 {

	var rounder float64
	intermed := val * math.Pow(10, float64(prec))

	if val >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / math.Pow(10, float64(prec))
}

/*
The input stucture that contains call information.
*/
type CallDescriptor struct {
	Direction                    string
	TOR                          string
	Tenant, Subject, Destination string
	TimeStart, TimeEnd           time.Time
	Amount                       float64
	FallbackSubject              string // the subject to check for destination if not found on primary subject
	ActivationPeriods            []*ActivationPeriod
	FallbackKey                  string
	userBalance                  *UserBalance
}

// Adds an activation period that applyes to current call descriptor.
func (cd *CallDescriptor) AddActivationPeriod(aps ...*ActivationPeriod) {
	cd.ActivationPeriods = append(cd.ActivationPeriods, aps...)
}

// Adds an activation period that applyes to current call descriptor if not already present.
func (cd *CallDescriptor) AddActivationPeriodIfNotPresent(aps ...*ActivationPeriod) {
	for _, ap := range aps {
		found := false
		for _, eap := range cd.ActivationPeriods {
			if ap.Equal(eap) {
				found = true
				break
			}
		}
		if !found {
			cd.ActivationPeriods = append(cd.ActivationPeriods, ap)
		}
	}
}

/*
Gets and caches the user balance information.
*/
func (cd *CallDescriptor) getUserBalance() (ub *UserBalance, err error) {
	if cd.userBalance == nil {
		key := fmt.Sprintf("%s:%s:%s", cd.Direction, cd.Tenant, cd.Subject)
		cd.userBalance, err = storageGetter.GetUserBalance(key)
	}
	return cd.userBalance, err
}

/*
Exported method to set the storage getter.
*/
func SetStorageGetter(sg StorageGetter) {
	storageGetter = sg
}

/*
Restores the activation periods for the specified prefix from storage.
*/
func (cd *CallDescriptor) SearchStorageForPrefix() (destPrefix string, err error) {
	cd.ActivationPeriods = make([]*ActivationPeriod, 0)
	base := fmt.Sprintf("%s:%s:%s:%s:", cd.Direction, cd.Tenant, cd.TOR, cd.Subject)
	destPrefix = cd.Destination
	key := base + destPrefix
	values, err := cd.getActivationPeriodsOrFallback(key, base, destPrefix, 1)
	if err != nil {
		key := base + FallbackDestination
		values, err = cd.getActivationPeriodsOrFallback(key, base, destPrefix, 1)
	}
	//load the activation preriods
	if err == nil && len(values) > 0 {
		cd.ActivationPeriods = values
	}
	return
}

func (cd *CallDescriptor) getActivationPeriodsOrFallback(key, base, destPrefix string, recursionDepth int) (values []*ActivationPeriod, err error) {
	if recursionDepth > RecursionMaxDepth {
		err = errors.New("Max fallback recursion depth reached!" + key)
		return
	}
	values, fallbackKey, err := storageGetter.GetActivationPeriodsOrFallback(key)
	if fallbackKey != "" {
		base = fallbackKey + ":"
		key = base + destPrefix
		recursionDepth++
		return cd.getActivationPeriodsOrFallback(key, base, destPrefix, recursionDepth)
	}
	//get for a smaller prefix if the orignal one was not found		
	for i := len(cd.Destination); err != nil || fallbackKey != ""; {
		if fallbackKey != "" {
			base = fallbackKey + ":"
			key = base + destPrefix
			recursionDepth++
			return cd.getActivationPeriodsOrFallback(key, base, destPrefix, recursionDepth)
		}
		i--
		if i >= MinPrefixLength {
			destPrefix = cd.Destination[:i]
			key = base + destPrefix
		} else {
			break
		}
		values, fallbackKey, err = storageGetter.GetActivationPeriodsOrFallback(key)
	}
	return
}

/*
Constructs the key for the storage lookup.
The prefixLen is limiting the length of the destination prefix.
*/
func (cd *CallDescriptor) GetKey() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", cd.Direction, cd.Tenant, cd.TOR, cd.Subject, cd.Destination)
}

/*
Splits the call descriptor timespan into sub time spans according to the activation periods intervals.
*/
func (cd *CallDescriptor) splitInTimeSpans() (timespans []*TimeSpan) {
	return cd.splitTimeSpan(&TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd})
}

/*
Splits the received timespan into sub time spans according to the activation periods intervals.
*/
func (cd *CallDescriptor) splitTimeSpan(firstSpan *TimeSpan) (timespans []*TimeSpan) {
	userBalancesRWMutex.RLock()
	defer userBalancesRWMutex.RUnlock()
	timespans = append(timespans, firstSpan)
	// split on (free) minute buckets	
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		_, _, bucketList := userBalance.getSecondsForPrefix(cd.Destination)
		for _, mb := range bucketList {
			for i := 0; i < len(timespans); i++ {
				if timespans[i].MinuteInfo != nil {
					continue
				}
				newTs := timespans[i].SplitByMinuteBucket(mb)
				if newTs != nil {
					timespans = append(timespans, newTs)
					firstSpan = newTs // we move the firstspan to the newly created one for further spliting
					break
				}
			}
		}
	}
	if firstSpan.MinuteInfo != nil {
		return // all the timespans are on minutes
	}
	if len(cd.ActivationPeriods) == 0 {
		return
	}
	firstSpan.ActivationPeriod = cd.ActivationPeriods[0]

	// split on activation periods
	afterStart, afterEnd := false, false //optimization for multiple activation periods
	for _, ap := range cd.ActivationPeriods {
		if !afterStart && !afterEnd && ap.ActivationTime.Before(cd.TimeStart) {
			firstSpan.ActivationPeriod = ap
		} else {
			afterStart = true
			for i := 0; i < len(timespans); i++ {
				if timespans[i].MinuteInfo != nil {
					continue
				}
				newTs := timespans[i].SplitByActivationPeriod(ap)
				if newTs != nil {
					timespans = append(timespans, newTs)
				} else {
					afterEnd = true
					break
				}
			}
		}
	}
	// split on price intervals
	for i := 0; i < len(timespans); i++ {
		if timespans[i].MinuteInfo != nil {
			continue
		}
		ap := timespans[i].ActivationPeriod
		//timespans[i].ActivationPeriod = nil
		for _, interval := range ap.Intervals {
			newTs := timespans[i].SplitByInterval(interval)
			if newTs != nil {
				newTs.ActivationPeriod = ap
				timespans = append(timespans, newTs)
			}
		}
	}
	return
}

/*
Creates a CallCost structure with the cost information calculated for the received CallDescriptor.
*/
func (cd *CallDescriptor) GetCost() (*CallCost, error) {
	destPrefix, err := cd.SearchStorageForPrefix()

	timespans := cd.splitInTimeSpans()
	cost := 0.0
	connectionFee := 0.0
	for i, ts := range timespans {
		if i == 0 && ts.MinuteInfo == nil && ts.Interval != nil {
			connectionFee = ts.Interval.ConnectFee
		}
		cost += ts.getCost(cd)
	}
	cc := &CallCost{TOR: cd.TOR,
		Tenant:      cd.Tenant,
		Subject:     cd.Subject,
		Destination: destPrefix,
		Cost:        cost,
		ConnectFee:  connectionFee,
		Timespans:   timespans}

	return cc, err
}

/*
Returns the cost of a second in the present time conditions.
*/
func (cd *CallDescriptor) getPresentSecondCost() (cost float64, err error) {
	// TODO: remove this method if if not still used
	_, err = cd.SearchStorageForPrefix()
	now := time.Now()
	oneSecond, _ := time.ParseDuration("1s")
	ts := &TimeSpan{TimeStart: now, TimeEnd: now.Add(oneSecond)}
	timespans := cd.splitTimeSpan(ts)

	if len(timespans) > 0 {
		cost = round(timespans[0].getCost(cd), 3)
	}
	return
}

/*
Returns the approximate max allowed session for user balance. It will try the max amount received in the call descriptor 
and will decrease it by 10% for nine times. So if the user has little credit it will still allow 10% of the initial amount.
If the user has no credit then it will return 0.
*/
func (cd *CallDescriptor) GetMaxSessionTime() (seconds float64, err error) {
	userBalancesRWMutex.RLock()
	defer userBalancesRWMutex.RUnlock()
	_, err = cd.SearchStorageForPrefix()
	now := time.Now()
	availableCredit, availableSeconds := 0.0, 0.0
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		if userBalance.Type == UB_TYPE_POSTPAID {
			return -1, nil
		} else {
			availableSeconds, availableCredit, _ = userBalance.getSecondsForPrefix(cd.Destination)
		}
	} else {
		return cd.Amount, err
	}
	// check for zero balance	
	if availableCredit == 0 {
		return availableSeconds, nil
	}
	maxSessionSeconds := cd.Amount
	for i := 0; i < 10; i++ {
		maxDuration, _ := time.ParseDuration(fmt.Sprintf("%vs", maxSessionSeconds-availableSeconds))
		ts := &TimeSpan{TimeStart: now, TimeEnd: now.Add(maxDuration)}
		timespans := cd.splitTimeSpan(ts)

		cost := 0.0
		for i, ts := range timespans {
			if i == 0 && ts.MinuteInfo == nil && ts.Interval != nil {
				cost += ts.Interval.ConnectFee
			}
			cost += ts.getCost(cd)
		}
		//log.Print(availableCredit, availableSeconds, cost)
		if cost < availableCredit {
			return maxSessionSeconds, nil
		} else { //decrease the period by 10% and try again
			maxSessionSeconds -= cd.Amount * 0.1
		}
	}
	return 0, nil
}

// Interface method used to add/substract an amount of cents or bonus seconds (as returned by GetCost method)
// from user's money balance.
func (cd *CallDescriptor) Debit() (cc *CallCost, err error) {
	userBalancesRWMutex.Lock()
	defer userBalancesRWMutex.Unlock()
	cc, err = cd.GetCost()
	if err != nil {
		log.Printf("error getting cost %v", err)
	}
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer storageGetter.SetUserBalance(userBalance)
		if cc.Cost != 0 || cc.ConnectFee != 0 {
			userBalance.debitMoneyBalance(cc.Cost + cc.ConnectFee)
		}
		for _, ts := range cc.Timespans {
			if ts.MinuteInfo != nil {
				userBalance.debitMinutesBalance(ts.MinuteInfo.Quantity, cd.Destination)
			}
		}
	}
	return
}

/*
Interface method used to add/substract an amount of cents from user's money balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitCents() (left float64, err error) {
	userBalancesRWMutex.Lock()
	defer userBalancesRWMutex.Unlock()
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer storageGetter.SetUserBalance(userBalance)
		return userBalance.debitMoneyBalance(cd.Amount), nil
	}
	return 0.0, err
}

/*
Interface method used to add/substract an amount of units from user's sms balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitSMS() (left float64, err error) {
	userBalancesRWMutex.Lock()
	defer userBalancesRWMutex.Unlock()
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer storageGetter.SetUserBalance(userBalance)
		return userBalance.debitSMSBalance(cd.Amount)
	}
	return 0, err
}

/*
Interface method used to add/substract an amount of seconds from user's minutes balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitSeconds() (err error) {
	userBalancesRWMutex.Lock()
	defer userBalancesRWMutex.Unlock()
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer storageGetter.SetUserBalance(userBalance)
		return userBalance.debitMinutesBalance(cd.Amount, cd.Destination)
	}
	return err
}

/*
Adds the specified amount of seconds to the received call seconds. When the threshold specified
in the user's tariff plan is reached then the received call balance is reseted and the bonus
specified in the tariff plan is applied.
The amount filed has to be filled in call descriptor.
*/
// func (cd *CallDescriptor) AddRecievedCallSeconds() (err error) {
// userBalancesRWMutex.Lock()
// defer userBalancesRWMutex.Unlock()
// 	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
// 		return userBalance.addReceivedCallSeconds(INBOUND, cd.TOR, cd.Destination, cd.Amount)
// 	}
// 	return err
// }

/*
Resets user balances value to the amounts specified in the tariff plan.
*/
/*func (cd *CallDescriptor) ResetUserBalance() (err error) {
	userBalancesRWMutex.Lock()
	defer userBalancesRWMutex.Unlock()
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		return userBalance.resetUserBalance()
	}
	return err
}*/
