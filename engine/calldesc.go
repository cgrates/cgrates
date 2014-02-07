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
	"log"
	"log/syslog"
	"time"
	//"encoding/json"
	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	var err error
	Logger, err = syslog.New(syslog.LOG_INFO, "CGRateS")
	if err != nil {
		Logger = new(utils.StdLogger)
		Logger.Err(fmt.Sprintf("Could not connect to syslog: %v", err))
	}
	DEBUG := true
	if DEBUG {
		dataStorage, _ = NewMapStorage()
		accountingStorage, _ = NewMapStorage()
	} else {
		//dataStorage, _ = NewMongoStorage(db_server, "27017", "cgrates_test", "", "")
		dataStorage, _ = NewRedisStorage("127.0.0.1:6379", 12, "", utils.MSGPACK)
		accountingStorage, _ = NewRedisStorage("127.0.0.1:6379", 13, "", utils.MSGPACK)
	}
	storageLogger = dataStorage.(LogStorage)
}

const (
	// these might be better in the confs under optimizations section
	RECURSION_MAX_DEPTH = 3
	MIN_PREFIX_MATCH    = 1
	FALLBACK_SUBJECT    = utils.ANY
)

var (
	Logger            utils.LoggerInterface
	dataStorage       RatingStorage
	accountingStorage AccountingStorage
	storageLogger     LogStorage
	debitPeriod       = 10 * time.Second
	roundingMethod    = "*middle"
	roundingDecimals  = 4
	historyScribe     history.Scribe
	//historyScribe, _ = history.NewMockScribe()
)

// Exported method to set the storage getter.
func SetRatingStorage(sg RatingStorage) {
	dataStorage = sg
}

func SetAccountingStorage(ag AccountingStorage) {
	accountingStorage = ag
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
	CallDuration                          time.Duration // the call duration so far (till TimeEnd)
	Amount                                float64
	FallbackSubject                       string // the subject to check for destination if not found on primary subject
	RatingInfos                           RatingInfos
	Increments                            Increments
	userBalance                           *UserBalance
}

func (cd *CallDescriptor) ValidateCallData() error {
	if cd.TimeStart.After(cd.TimeEnd) || cd.TimeStart.Equal(cd.TimeEnd) {
		return errors.New("TimeStart must be strctly before TimeEnd")
	}
	if cd.TimeEnd.Sub(cd.TimeStart) < cd.CallDuration {
		return errors.New("CallDuration must be equal or grater than TimeEnd - TimeStart")
	}
	return nil
}

// Adds a rating plan that applyes to current call descriptor.
func (cd *CallDescriptor) AddRatingInfo(ris ...*RatingInfo) {
	cd.RatingInfos = append(cd.RatingInfos, ris...)
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
		cd.userBalance, err = accountingStorage.GetUserBalance(cd.GetUserBalanceKey())
	}
	if cd.userBalance != nil && cd.userBalance.Disabled {
		return nil, fmt.Errorf("User %s is disabled", ub.Id)
	}
	return cd.userBalance, err
}

/*
Restores the activation periods for the specified prefix from storage.
*/
func (cd *CallDescriptor) LoadRatingPlans() (err error) {
	err = cd.getRatingPlansForPrefix(cd.GetKey(cd.Subject), 1)
	if err != nil || !cd.continousRatingInfos() {
		// use the default subject
		err = cd.getRatingPlansForPrefix(cd.GetKey(FALLBACK_SUBJECT), 1)
	}
	//load the rating plans
	if err != nil || !cd.continousRatingInfos() {
		err = errors.New("Could not determine rating plans for call")
		return
	}
	return
}

// FIXME: this method is not exhaustive but will cover 99% of cases just good
// it will not cover very long calls with very short activation periods for rates
func (cd *CallDescriptor) getRatingPlansForPrefix(key string, recursionDepth int) (err error) {
	if recursionDepth > RECURSION_MAX_DEPTH {
		err = errors.New("Max fallback recursion depth reached!" + key)
		return
	}
	rpf, err := dataStorage.GetRatingProfile(key, false)
	if err != nil || rpf == nil {
		return err
	}
	if err = rpf.GetRatingPlansForPrefix(cd); err != nil || !cd.continousRatingInfos() {
		// try rating profile fallback
		recursionDepth++
		for index := 0; index < len(cd.RatingInfos); index++ {
			ri := cd.RatingInfos[index]
			if len(ri.RateIntervals) > 0 {
				// go to next rate info
				continue
			}
			if len(ri.FallbackKeys) > 0 {
				tempCD := &CallDescriptor{
					TOR:         cd.TOR,
					Direction:   cd.Direction,
					Tenant:      cd.Tenant,
					Destination: cd.Destination,
				}
				if index == 0 {
					tempCD.TimeStart = cd.TimeStart
				} else {
					tempCD.TimeStart = ri.ActivationTime
				}
				if index == len(cd.RatingInfos)-1 {
					tempCD.TimeEnd = cd.TimeEnd
				} else {
					tempCD.TimeEnd = cd.RatingInfos[index+1].ActivationTime
				}
				for _, fbk := range ri.FallbackKeys {
					if err := tempCD.getRatingPlansForPrefix(fbk, recursionDepth); err != nil {
						continue
					}
					// extract the rate infos and break
					for newIndex, newRI := range tempCD.RatingInfos {
						// check if the new ri is filled
						if len(newRI.RateIntervals) == 0 {
							continue
						}
						if newIndex == 0 {
							cd.RatingInfos[index] = newRI
						} else {
							// insert extra data
							i := index + newIndex
							cd.RatingInfos = append(cd.RatingInfos, nil)
							copy(cd.RatingInfos[i+1:], cd.RatingInfos[i:])
							cd.RatingInfos[i] = newRI
						}
					}
					// if this fallbackey covered the interval than skip
					// the other fallback keys
					if tempCD.continousRatingInfos() {
						break
					}
				}
			}
		}
	}
	return
}

// checks if there is rating info for the entire call duration
func (cd *CallDescriptor) continousRatingInfos() bool {
	if len(cd.RatingInfos) == 0 || cd.RatingInfos[0].ActivationTime.After(cd.TimeStart) {
		return false
	}
	for _, ri := range cd.RatingInfos {
		if ri.RateIntervals == nil {
			return false
		}
	}
	return true
}

// adds a rating infos only if that call period is not already covered
// returns true if added
func (cd *CallDescriptor) addRatingInfos(ris RatingInfos) bool {
	if len(cd.RatingInfos) == 0 {
		cd.RatingInfos = append(cd.RatingInfos, ris...)
		return true
	}
	cd.RatingInfos.Sort()
	// check if we dont have the start covered
	if cd.RatingInfos[0].ActivationTime.After(cd.TimeStart) {
		if ris[0].ActivationTime.Before(cd.RatingInfos[0].ActivationTime) {
			cd.RatingInfos = append(cd.RatingInfos, ris[0])
			cd.RatingInfos.Sort()
		}
	}
	for _, ri := range cd.RatingInfos {
		if ri.RateIntervals == nil {
			for i, new_ri := range ris {
				_ = i
				_ = new_ri
			}
		}
	}
	return true
}

// Constructs the key for the storage lookup.
// The prefixLen is limiting the length of the destination prefix.
func (cd *CallDescriptor) GetKey(subject string) string {
	return fmt.Sprintf("%s:%s:%s:%s", cd.Direction, cd.Tenant, cd.TOR, subject)
}

// Splits the received timespan into sub time spans according to the activation periods intervals.
func (cd *CallDescriptor) splitInTimeSpans(firstSpan *TimeSpan) (timespans []*TimeSpan) {
	if firstSpan == nil {
		firstSpan = &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd, CallDuration: cd.CallDuration}
	}
	timespans = append(timespans, firstSpan)
	if len(cd.RatingInfos) == 0 {
		return
	}

	firstSpan.ratingInfo = cd.RatingInfos[0]
	// split on rating plans
	afterStart, afterEnd := false, false //optimization for multiple activation periods
	for _, rp := range cd.RatingInfos {
		if !afterStart && !afterEnd && rp.ActivationTime.Before(cd.TimeStart) {
			firstSpan.ratingInfo = rp
			firstSpan.MatchedSubject = rp.MatchedSubject
			firstSpan.MatchedPrefix = rp.MatchedPrefix
		} else {
			afterStart = true
			for i := 0; i < len(timespans); i++ {
				newTs := timespans[i].SplitByRatingPlan(rp)
				if newTs != nil {
					timespans = append(timespans, newTs)
				} else {
					afterEnd = true
					break
				}
			}
		}
	}
	// Logger.Debug(fmt.Sprintf("After SplitByRatingPlan: %+v", timespans))
	// split on price intervals
	for i := 0; i < len(timespans); i++ {
		//log.Printf("==============%v==================", i)
		//log.Printf("TS: %+v", timespans[i])
		rp := timespans[i].ratingInfo
		// Logger.Debug(fmt.Sprintf("rp: %+v", rp))
		//timespans[i].RatingPlan = nil
		rp.RateIntervals.Sort()
		for _, interval := range rp.RateIntervals {
			//log.Printf("\tINTERVAL: %+v %v", interval, len(rp.RateIntervals))
			if timespans[i].RateInterval != nil && timespans[i].RateInterval.Weight < interval.Weight {
				continue // if the timespan has an interval than it already has a heigher weight
			}
			newTs := timespans[i].SplitByRateInterval(interval)
			if newTs != nil {
				newTs.ratingInfo = rp
				// insert the new timespan
				index := i + 1
				timespans = append(timespans, nil)
				copy(timespans[index+1:], timespans[index:])
				timespans[index] = newTs
				break
			}
		}
	}
	//Logger.Debug(fmt.Sprintf("After SplitByRateInterval: %+v", timespans))
	//log.Printf("After SplitByRateInterval: %+v", timespans)
	timespans = cd.roundTimeSpansToIncrement(timespans)
	// Logger.Debug(fmt.Sprintf("After round: %+v", timespans))
	//log.Printf("After round: %+v", timespans)
	return
}

// if the rate interval for any timespan has a RatingIncrement larger than the timespan duration
// the timespan must expand potentially overlaping folowing timespans and may exceed call
// descriptor's initial duration
func (cd *CallDescriptor) roundTimeSpansToIncrement(timespans TimeSpans) []*TimeSpan {
	for i := 0; i < len(timespans); i++ {
		ts := timespans[i]
		if ts.RateInterval != nil {
			_, rateIncrement, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
			// if the timespan duration is larger than the rate increment make sure it is a multiple of it
			if rateIncrement < ts.GetDuration() {
				rateIncrement = utils.RoundTo(rateIncrement, ts.GetDuration())
			}
			if rateIncrement > ts.GetDuration() {
				initialDuration := ts.GetDuration()
				ts.TimeEnd = ts.TimeStart.Add(rateIncrement)
				ts.CallDuration = ts.CallDuration + (rateIncrement - initialDuration)
				timespans.RemoveOverlapedFromIndex(i)
			}
		}
	}

	return timespans
}

// Returns call descripor's total duration
func (cd *CallDescriptor) GetDuration() time.Duration {
	return cd.TimeEnd.Sub(cd.TimeStart)
}

/*
Creates a CallCost structure with the cost information calculated for the received CallDescriptor.
*/
func (cd *CallDescriptor) GetCost() (*CallCost, error) {
	if cd.CallDuration < cd.TimeEnd.Sub(cd.TimeStart) {
		cd.CallDuration = cd.TimeEnd.Sub(cd.TimeStart)
	}
	err := cd.LoadRatingPlans()
	if err != nil {
		Logger.Err(fmt.Sprintf("error getting cost for key %v: %v", cd.GetUserBalanceKey(), err))
		return &CallCost{Cost: -1}, err
	}
	timespans := cd.splitInTimeSpans(nil)
	cost := 0.0

	for i, ts := range timespans {
		// only add connect fee if this is the first/only call cost request
		if cd.LoopIndex == 0 && i == 0 && ts.RateInterval != nil {
			cost += ts.RateInterval.Rating.ConnectFee
		}
		cost += ts.getCost()
	}
	// global rounding
	cost = utils.Round(cost, roundingDecimals, roundingMethod)
	//startIndex := len(fmt.Sprintf("%s:%s:%s:", cd.Direction, cd.Tenant, cd.TOR))
	cc := &CallCost{
		Direction:        cd.Direction,
		TOR:              cd.TOR,
		Tenant:           cd.Tenant,
		Account:          cd.Account,
		Destination:      cd.Destination,
		Subject:          cd.Subject,
		Cost:             cost,
		Timespans:        timespans,
		deductConnectFee: cd.LoopIndex == 0,
	}
	//Logger.Info(fmt.Sprintf("<Rater> Get Cost: %s => %v", cd.GetKey(), cc))
	return cc, err
}

/*
Returns the approximate max allowed session for user balance. It will try the max amount received in the call descriptor
If the user has no credit then it will return 0.
If the user has postpayed plan it returns -1.
*/
func (origCd *CallDescriptor) GetMaxSessionDuration() (time.Duration, error) {
	cd := origCd.Clone()
	if cd.CallDuration == 0 {
		cd.CallDuration = cd.TimeEnd.Sub(cd.TimeStart)
	}
	//Logger.Debug(fmt.Sprintf("MAX SESSION cd: %+v", cd))
	err := cd.LoadRatingPlans()
	if err != nil {
		Logger.Err(fmt.Sprintf("error getting cost for key %v: %v", cd.GetUserBalanceKey(), err))
		return 0, err
	}
	var availableDuration time.Duration
	availableCredit := 0.0
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		if userBalance.Type == UB_TYPE_POSTPAID {
			return -1, nil
		} else {
			availableDuration, availableCredit, _ = userBalance.getCreditForPrefix(cd)
			// Logger.Debug(fmt.Sprintf("available sec: %v credit: %v", availableSeconds, availableCredit))
		}
	} else {
		Logger.Err(fmt.Sprintf("Could not get user balance for %s: %s.", cd.GetUserBalanceKey(), err.Error()))
		return 0, err
	}
	//Logger.Debug(fmt.Sprintf("availableDuration: %v, availableCredit: %v", availableDuration, availableCredit))
	initialDuration := cd.TimeEnd.Sub(cd.TimeStart)
	if initialDuration <= availableDuration {
		// there are enough minutes for requested interval
		return initialDuration, nil
	}
	// check for zero balance
	if availableCredit == 0 {
		return utils.MinDuration(initialDuration, availableDuration), nil
	}
	//Logger.Debug(fmt.Sprintf("initial Duration: %v", initialDuration))
	// we must move the timestart for the interval with the available duration because
	// that was already checked
	cd.TimeStart = cd.TimeStart.Add(availableDuration)
	// substract the connect fee
	cc, err := cd.GetCost()
	availableCredit -= cc.GetConnectFee()
	if err != nil {
		Logger.Err(fmt.Sprintf("Could not get cost for %s: %s.", cd.GetKey(cd.Subject), err.Error()))
		return 0, err
	}
	// now let's check how many increments are covered with the avilableCredit
	for _, ts := range cc.Timespans {
		ts.createIncrementsSlice()
		//Logger.Debug(fmt.Sprintf("TS: %+v", ts))
		for _, incr := range ts.Increments {
			if incr.Cost <= availableCredit {
				availableCredit -= incr.Cost
				availableDuration += incr.Duration
			} else {
				return availableDuration, nil
			}
		}
	}
	log.Print(initialDuration, availableDuration, initialDuration < availableDuration)
	if initialDuration < availableDuration {
		return initialDuration, nil
	}
	return utils.MinDuration(initialDuration, availableDuration), nil
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
		// Logger.Debug(fmt.Sprintf("<Rater> No user balance defined: %v", cd.GetUserBalanceKey()))
	} else {
		//Logger.Debug(fmt.Sprintf("<Rater> Attempting to debit from %v, value: %v", cd.GetUserBalanceKey(), cc.Cost+cc.ConnectFee))
		defer accountingStorage.SetUserBalance(userBalance)
		//ub, _ := json.Marshal(userBalance)
		//Logger.Debug(fmt.Sprintf("UserBalance: %s", ub))
		//cCost, _ := json.Marshal(cc)
		//Logger.Debug(fmt.Sprintf("CallCost: %s", cCost))
		if cc.Cost != 0 || cc.GetConnectFee() != 0 {
			userBalance.debitCreditBalance(cc, true)
		}
		cost := 0.0
		// re-calculate call cost after balances
		for _, ts := range cc.Timespans {
			cost += ts.getCost()
			cost = utils.Round(cost, roundingDecimals, utils.ROUNDING_MIDDLE) // just get rid of the extra decimals
		}
		cc.Cost = cost
	}
	return
}

// Interface method used to add/substract an amount of cents or bonus seconds (as returned by GetCost method)
// from user's money balance.
// This methods combines the Debit and GetMaxSessionDuration and will debit the max available time as returned
// by the GetMaxSessionTime method. The amount filed has to be filled in call descriptor.
func (cd *CallDescriptor) MaxDebit() (cc *CallCost, err error) {
	remainingDuration, err := cd.GetMaxSessionDuration()
	if err != nil || remainingDuration == 0 {
		return new(CallCost), errors.New("no more credit")
	}
	log.Print("REM_DUR: ", remainingDuration)
	if remainingDuration > 0 { // for postpaying client returns -1
		cd.TimeEnd = cd.TimeStart.Add(remainingDuration)
	}
	return cd.Debit()
}

func (cd *CallDescriptor) RefundIncrements() (left float64, err error) {
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer accountingStorage.SetUserBalance(userBalance)
		userBalance.refundIncrements(cd.Increments, cd.Direction, true)
	}
	return 0.0, err
}

/*
Interface method used to add/substract an amount of cents from user's money balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitCents() (left float64, err error) {
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer accountingStorage.SetUserBalance(userBalance)
		return userBalance.debitGenericBalance(CREDIT, cd.Direction, cd.Amount, true), nil
	}
	return 0.0, err
}

/*
Interface method used to add/substract an amount of units from user's sms balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitSMS() (left float64, err error) {
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer accountingStorage.SetUserBalance(userBalance)
		return userBalance.debitGenericBalance(SMS, cd.Direction, cd.Amount, true), nil
	}
	return 0, err
}

/*
Interface method used to add/substract an amount of seconds from user's minutes balance.
The amount filed has to be filled in call descriptor.
*/
func (cd *CallDescriptor) DebitSeconds() (err error) {
	if userBalance, err := cd.getUserBalance(); err == nil && userBalance != nil {
		defer accountingStorage.SetUserBalance(userBalance)
		return userBalance.debitCreditBalance(cd.CreateCallCost(), true)
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

func (cd *CallDescriptor) FlushCache() (err error) {
	cache2go.XFlush()
	cache2go.Flush()
	dataStorage.CacheRating(nil, nil, nil)
	accountingStorage.CacheAccounting(nil)
	return nil

}

// Creates a CallCost structure copying related data from CallDescriptor
func (cd *CallDescriptor) CreateCallCost() *CallCost {
	return &CallCost{
		Direction:   cd.Direction,
		TOR:         cd.TOR,
		Tenant:      cd.Tenant,
		Subject:     cd.Subject,
		Account:     cd.Account,
		Destination: cd.Destination,
	}
}

func (cd *CallDescriptor) Clone() *CallDescriptor {
	return &CallDescriptor{
		Direction:       cd.Direction,
		TOR:             cd.TOR,
		Tenant:          cd.Tenant,
		Subject:         cd.Subject,
		Account:         cd.Account,
		Destination:     cd.Destination,
		TimeStart:       cd.TimeStart,
		TimeEnd:         cd.TimeEnd,
		LoopIndex:       cd.LoopIndex,
		CallDuration:    cd.CallDuration,
		Amount:          cd.Amount,
		FallbackSubject: cd.FallbackSubject,
		RatingInfos:     cd.RatingInfos,
		Increments:      cd.Increments,
	}
}
