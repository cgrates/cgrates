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

package engine

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/utils"
	"log/syslog"
	"strings"
	"time"
)

func init() {
	var err error
	Logger, err = syslog.New(syslog.LOG_INFO, "CGRateS")
	if err != nil {
		Logger = new(utils.StdLogger)
		Logger.Err(fmt.Sprintf("Could not connect to syslog: %v", err))
	}
	//db_server := "127.0.0.1"
	//db_server := "192.168.0.17"
	m, _ := NewMapStorage()
	//m, _ := NewMongoStorage(db_server, "27017", "cgrates_test", "", "")
	//m, _ := NewRedisStorage(db_server+":6379", 11, "")
	//m, _ := NewRedigoStorage(db_server+":6379", 11, "")
	//m, _ := NewRadixStorage(db_server+":6379", 11, "")
	storageGetter, _ = m.(DataStorage)

	storageLogger = storageGetter.(LogStorage)
}

const (
	RECURSION_MAX_DEPTH = 10
	FALLBACK_SUBJECT    = "*any"
	FALLBACK_SEP        = ";"
)

var (
	Logger           utils.LoggerInterface
	storageGetter    DataStorage
	storageLogger    LogStorage
	debitPeriod      = 10 * time.Second
	roundingMethod   = "*middle"
	roundingDecimals = 4
	historyScribe    history.Scribe
	//historyScribe, _ = history.NewMockScribe()
)

// Exported method to set the storage getter.
func SetDataStorage(sg DataStorage) {
	storageGetter = sg
}

// Sets the global rounding method and decimal precision for GetCost method
func SetRoundingMethodAndDecimals(rm string, rd int) {
	roundingMethod = rm
	roundingDecimals = rd
}

/*
Sets the database for logging (can be de same  as storage getter or different db)
*/
func SetStorageLogger(sg LogStorage) {
	storageLogger = sg
}

/*
Exported method to set the debit period for caching purposes.
*/
func SetDebitPeriod(d time.Duration) {
	debitPeriod = d
}

// Exported method to set the history scribe.
func SetHistoryScribe(scribe history.Scribe) {
	historyScribe = scribe
}

/*
The input stucture that contains call information.
*/
type CallDescriptor struct {
	Direction                             string
	TOR                                   string
	Tenant, Subject, Account, Destination string
	TimeStart, TimeEnd                    time.Time
	LoopIndex                             float64       // indicates the position of this segment in a cost request loop
	CallDuration                          time.Duration // the call duration so far (partial or final)
	Amount                                float64
	FallbackSubject                       string // the subject to check for destination if not found on primary subject
	RatingPlans                           []*RatingPlan
	userBalance                           *UserBalance
}

// Adds an activation period that applyes to current call descriptor.
func (cd *CallDescriptor) AddRatingPlan(aps ...*RatingPlan) {
	cd.RatingPlans = append(cd.RatingPlans, aps...)
}

// Returns the key used to retrive the user balance involved in this call
func (cd *CallDescriptor) GetUserBalanceKey() string {
	subj := cd.Subject
	if cd.Account != "" {
		subj = cd.Account
	}
	return fmt.Sprintf("%s:%s:%s", cd.Direction, cd.Tenant, subj)
}

// Gets and caches the user balance information.
func (cd *CallDescriptor) getUserBalance() (ub *UserBalance, err error) {
	if cd.userBalance == nil {
		cd.userBalance, err = storageGetter.GetUserBalance(cd.GetUserBalanceKey())
	}
	return cd.userBalance, err
}

/*
Restores the activation periods for the specified prefix from storage.
*/
func (cd *CallDescriptor) LoadRatingPlans() (destPrefix string, err error) {
	if val, err := cache2go.GetXCached(cd.GetKey() + cd.Destination); err == nil {
		xaps := val.(xCachedRatingPlans)
		cd.RatingPlans = xaps.aps
		return xaps.destPrefix, nil
	}
	destPrefix, values, err := cd.getRatingPlansForPrefix(cd.GetKey(), 1)
	if err != nil {
		fallbackKey := fmt.Sprintf("%s:%s:%s:%s", cd.Direction, cd.Tenant, cd.TOR, FALLBACK_SUBJECT)
		// use the default subject
		destPrefix, values, err = cd.getRatingPlansForPrefix(fallbackKey, 1)
	}
	//load the activation preriods
	if err == nil && len(values) > 0 {
		xaps := xCachedRatingPlans{destPrefix, values, new(cache2go.XEntry)}
		xaps.XCache(cd.GetKey()+cd.Destination, debitPeriod+5*time.Second, xaps)
		cd.RatingPlans = values
	}
	return
}

func (cd *CallDescriptor) getRatingPlansForPrefix(key string, recursionDepth int) (foundPrefix string, aps []*RatingPlan, err error) {
	if recursionDepth > RECURSION_MAX_DEPTH {
		err = errors.New("Max fallback recursion depth reached!" + key)
		return
	}
	rp, err := storageGetter.GetRatingProfile(key)
	if err != nil {
		return "", nil, err
	}
	foundPrefix, aps, err = rp.GetRatingPlansForPrefix(cd.Destination)
	if err != nil {
		if rp.FallbackKey != "" {
			recursionDepth++
			for _, fbk := range strings.Split(rp.FallbackKey, FALLBACK_SEP) {
				if destPrefix, values, err := cd.getRatingPlansForPrefix(fbk, recursionDepth); err == nil {
					return destPrefix, values, err
				}
			}
		}
	}

	return
}

/*
Constructs the key for the storage lookup.
The prefixLen is limiting the length of the destination prefix.
*/
func (cd *CallDescriptor) GetKey() string {
	return fmt.Sprintf("%s:%s:%s:%s", cd.Direction, cd.Tenant, cd.TOR, cd.Subject)
}

/*
Splits the received timespan into sub time spans according to the activation periods intervals.
*/
func (cd *CallDescriptor) splitInTimeSpans(firstSpan *TimeSpan) (timespans []*TimeSpan) {
	if firstSpan == nil {
		firstSpan = &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd, CallDuration: cd.CallDuration}
	}
	timespans = append(timespans, firstSpan)
	// split on (free) minute buckets
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		_, _, minuteBalances := userBalance.getSecondsForPrefix(cd.Destination)
		for _, b := range minuteBalances {
			for i := 0; i < len(timespans); i++ {
				if timespans[i].MinuteInfo != nil {
					continue
				}
				newTs := timespans[i].SplitByMinuteBalance(b)
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
	if len(cd.RatingPlans) == 0 {
		return
	}
	firstSpan.RatingPlan = cd.RatingPlans[0]

	// split on activation periods
	afterStart, afterEnd := false, false //optimization for multiple activation periods
	for _, ap := range cd.RatingPlans {
		if !afterStart && !afterEnd && ap.ActivationTime.Before(cd.TimeStart) {
			firstSpan.RatingPlan = ap
		} else {
			afterStart = true
			for i := 0; i < len(timespans); i++ {
				if timespans[i].MinuteInfo != nil {
					continue
				}
				newTs := timespans[i].SplitByRatingPlan(ap)
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
			continue // cont try to split timespans payed with minutes
		}
		ap := timespans[i].RatingPlan
		//timespans[i].RatingPlan = nil
		ap.RateIntervals.Sort()
		for _, interval := range ap.RateIntervals {
			if timespans[i].RateInterval != nil && timespans[i].RateInterval.Weight < interval.Weight {
				continue // if the timespan has an interval than it already has a heigher weight
			}
			newTs := timespans[i].SplitByRateInterval(interval)
			if newTs != nil {
				newTs.RatingPlan = ap
				timespans = append(timespans, newTs)
			}
		}
	}
	timespans = cd.expandTimeSpans(timespans)
	return
}

// if the rate interval for any timespan has a RatingIncrement larger than the timespan duration
// the timespan must expand potentially overlaping folowing timespans and may exceed call
// descriptor's initial duration
func (cd *CallDescriptor) expandTimeSpans(timespans []*TimeSpan) []*TimeSpan {
	for i, ts := range timespans {
		if ts.RateInterval != nil {
			_, rateIncrement, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
			// if the timespan duration is larger than the rate increment make sure it is a multiple of it
			if rateIncrement < ts.GetDuration() {
				rateIncrement = utils.RoundTo(rateIncrement, ts.GetDuration())
			}
			if rateIncrement > ts.GetDuration() {
				ts.TimeEnd = ts.TimeStart.Add(rateIncrement)
				ts.SetNewCallDuration(ts) // set new call duration for this timespan
				// overlap the rest of the timespans
				for ; i < len(timespans); i++ {
					if timespans[i].TimeEnd.Before(ts.TimeEnd) {
						timespans[i].overlapped = true
					} else if timespans[i].TimeStart.Before(ts.TimeEnd) {
						timespans[i].TimeStart = ts.TimeEnd
					}
				}
				break
			}
		}
	}
	// remove overlapped
	for i, ts := range timespans {
		if ts.overlapped {
			timespans = timespans[:i]
			break
		}
	}
	return timespans
}

/*
Creates a CallCost structure with the cost information calculated for the received CallDescriptor.
*/
func (cd *CallDescriptor) GetCost() (*CallCost, error) {
	destPrefix, err := cd.LoadRatingPlans()
	if err != nil {
		Logger.Err(fmt.Sprintf("error getting cost for key %v: %v", cd.GetUserBalanceKey(), err))
		return &CallCost{Cost: -1}, err
	}
	timespans := cd.splitInTimeSpans(nil)
	cost := 0.0
	connectionFee := 0.0

	for i, ts := range timespans {
		// only add connect fee if this is the first/only call cost request
		if cd.LoopIndex == 0 && i == 0 && ts.MinuteInfo == nil && ts.RateInterval != nil {
			connectionFee = ts.RateInterval.ConnectFee
		}
		cost += ts.getCost(cd)
	}
	cost = utils.Round(cost, roundingDecimals, roundingMethod)
	cc := &CallCost{
		Direction:   cd.Direction,
		TOR:         cd.TOR,
		Tenant:      cd.Tenant,
		Subject:     cd.Subject,
		Account:     cd.Account,
		Destination: destPrefix,
		Cost:        cost,
		ConnectFee:  connectionFee,
		Timespans:   timespans}
	//Logger.Info(fmt.Sprintf("<Rater> Get Cost: %s => %v", cd.GetKey(), cc))
	return cc, err
}

/*
Returns the approximate max allowed session for user balance. It will try the max amount received in the call descriptor
and will decrease it by 10% for nine times. So if the user has little credit it will still allow 10% of the initial amount.
If the user has no credit then it will return 0.
If the user has postpayed plan it returns -1.
*/
func (cd *CallDescriptor) GetMaxSessionTime(startTime time.Time) (seconds float64, err error) {
	_, err = cd.LoadRatingPlans()
	if err != nil {
		Logger.Err(fmt.Sprintf("error getting cost for key %v: %v", cd.GetUserBalanceKey(), err))
		return 0, err
	}
	availableCredit, availableSeconds := 0.0, 0.0
	Logger.Debug(fmt.Sprintf("cd: %+v", cd))
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		if userBalance.Type == UB_TYPE_POSTPAID {
			return -1, nil
		} else {
			availableSeconds, availableCredit, _ = userBalance.getSecondsForPrefix(cd.Destination)
			Logger.Debug(fmt.Sprintf("available sec: %v credit: %v", availableSeconds, availableCredit))
		}
	} else {
		Logger.Err(fmt.Sprintf("Could not get user balance for %s: %s.", cd.GetUserBalanceKey(), err.Error()))
		return cd.Amount, err
	}
	// check for zero balance
	if availableCredit == 0 {
		return availableSeconds, nil
	}
	// the price of a second cannot be determined because all the seconds can have a different cost.
	// therfore we get the cost for the whole period and then if there are not enough money we backout in steps of 10%.
	maxSessionSeconds := cd.Amount
	for i := 0; i < 10; i++ {
		maxDuration, _ := time.ParseDuration(fmt.Sprintf("%vs", maxSessionSeconds-availableSeconds))
		ts := &TimeSpan{TimeStart: startTime, TimeEnd: startTime.Add(maxDuration)}
		timespans := cd.splitInTimeSpans(ts)

		cost := 0.0
		for i, ts := range timespans {
			if i == 0 && ts.MinuteInfo == nil && ts.RateInterval != nil {
				cost += ts.RateInterval.ConnectFee
			}
			cost += ts.getCost(cd)
		}
		//logger.Print(availableCredit, availableSeconds, cost)
		if cost < availableCredit {
			return maxSessionSeconds, nil
		} else { //decrease the period by 10% and try again
			maxSessionSeconds -= cd.Amount * 0.1
		}
	}
	Logger.Debug("Even 10% of the max session time is too much!")
	return 0, nil
}

// Interface method used to add/substract an amount of cents or bonus seconds (as returned by GetCost method)
// from user's money balance.
func (cd *CallDescriptor) Debit() (cc *CallCost, err error) {
	cc, err = cd.GetCost()
	if err != nil {
		Logger.Err(fmt.Sprintf("<Rater> Error getting cost for account key %v: %v", cd.GetUserBalanceKey(), err))
		return
	}
	if userBalance, err := cd.getUserBalance(); err != nil {
		Logger.Err(fmt.Sprintf("<Rater> Error retrieving user balance: %v", err))
	} else if userBalance == nil {
		Logger.Debug(fmt.Sprintf("<Rater> No user balance defined: %v", cd.GetUserBalanceKey()))
	} else {
		Logger.Debug(fmt.Sprintf("<Rater> Attempting to debit from %v, value: %v", cd.GetUserBalanceKey(), cc.Cost+cc.ConnectFee))
		defer storageGetter.SetUserBalance(userBalance)
		if cc.Cost != 0 || cc.ConnectFee != 0 {
			userBalance.debitBalance(CREDIT, cc.Cost+cc.ConnectFee, true)
		}
		for _, ts := range cc.Timespans {
			if ts.MinuteInfo != nil {
				if err = userBalance.debitMinutesBalance(ts.MinuteInfo.Quantity, cd.Destination, true); err != nil {
					return cc, err
				}
			}
		}
	}
	return
}

// Interface method used to add/substract an amount of cents or bonus seconds (as returned by GetCost method)
// from user's money balance.
// This methods combines the Debit and GetMaxSessionTime and will debit the max available time as returned
// by the GetMaxSessionTime method. The amount filed has to be filled in call descriptor.
func (cd *CallDescriptor) MaxDebit(startTime time.Time) (cc *CallCost, err error) {
	remainingSeconds, err := cd.GetMaxSessionTime(startTime)
	Logger.Debug(fmt.Sprintf("In MaxDebitd remaining seconds: %v", remainingSeconds))
	if err != nil || remainingSeconds == 0 {
		return new(CallCost), errors.New("no more credit")
	}
	if remainingSeconds > 0 { // for postpaying client returns -1
		cd.TimeEnd = cd.TimeStart.Add(time.Duration(remainingSeconds) * time.Second)
	}
	return cd.Debit()
}

/*
Interface method used to add/substract an amount of cents from user's money balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitCents() (left float64, err error) {
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer storageGetter.SetUserBalance(userBalance)
		return userBalance.debitBalance(CREDIT, cd.Amount, true), nil
	}
	return 0.0, err
}

/*
Interface method used to add/substract an amount of units from user's sms balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitSMS() (left float64, err error) {
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer storageGetter.SetUserBalance(userBalance)
		return userBalance.debitBalance(SMS, cd.Amount, true), nil
	}
	return 0, err
}

/*
Interface method used to add/substract an amount of seconds from user's minutes balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitSeconds() (err error) {
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer storageGetter.SetUserBalance(userBalance)
		return userBalance.debitMinutesBalance(cd.Amount, cd.Destination, true)
	}
	return err
}

/*
Adds the specified amount of seconds to the received call seconds. When the threshold specified
in the user's tariff plan is reached then the received call balance is reseted and the bonus
specified in the tariff plan is applied.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) AddRecievedCallSeconds() (err error) {
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		a := &Action{
			Direction: INBOUND,
			Balance:   &Balance{Value: cd.Amount, DestinationId: cd.Destination},
		}
		userBalance.countUnits(a)
		return nil
	}
	return err
}

// Cleans all chached data
func (cd *CallDescriptor) FlushCache() (err error) {
	cache2go.XFlush()
	cache2go.Flush()
	return nil

}
