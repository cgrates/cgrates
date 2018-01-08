/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

	"sort"
	"strings"
	"time"

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	// these might be better in the confs under optimizations section
	RECURSION_MAX_DEPTH = 3
	MIN_PREFIX_MATCH    = 1
	FALLBACK_SUBJECT    = utils.ANY
	DB                  = "map"
)

func init() {
	var err error
	var data DataDB
	switch DB {
	case "map":
		if cgrCfg := config.CgrConfig(); cgrCfg == nil {
			cgrCfg, _ = config.NewDefaultCGRConfig()
			config.SetCgrConfig(cgrCfg)
		}
		data, _ = NewMapStorage()
	case utils.MONGO:
		data, err = NewMongoStorage("127.0.0.1", "27017", "cgrates_data_test", "", "", utils.DataDB, nil, config.CacheConfig{utils.CacheRatingPlans: &config.CacheParamConfig{Precache: true}}, 10)
		if err != nil {
			log.Fatal(err)
		}
	case utils.REDIS:
		data, _ = NewRedisStorage("127.0.0.1:6379", 12, "", utils.MSGPACK, utils.REDIS_MAX_CONNS, config.CacheConfig{utils.CacheRatingPlans: &config.CacheParamConfig{Precache: true}}, 10)
		if err != nil {
			log.Fatal(err)
		}
	}
	dm = NewDataManager(data)
}

var (
	dm                       *DataManager
	cdrStorage               CdrStorage
	debitPeriod              = 10 * time.Second
	globalRoundingDecimals   = 6
	thresholdS               rpcclient.RpcClientConnection // used by RALs to communicate with ThresholdS
	pubSubServer             rpcclient.RpcClientConnection
	userService              rpcclient.RpcClientConnection
	aliasService             rpcclient.RpcClientConnection
	rpSubjectPrefixMatching  bool
	lcrSubjectPrefixMatching bool
)

// Exported method to set the storage getter.
func SetDataStorage(dm2 *DataManager) {
	dm = dm2
}

func SetThresholdS(thdS rpcclient.RpcClientConnection) {
	thresholdS = thdS
}

// Sets the global rounding method and decimal precision for GetCost method
func SetRoundingDecimals(rd int) {
	globalRoundingDecimals = rd
}

func SetRpSubjectPrefixMatching(flag bool) {
	rpSubjectPrefixMatching = flag
}

func SetLcrSubjectPrefixMatching(flag bool) {
	lcrSubjectPrefixMatching = flag
}

/*
Sets the database for CDR storing, used by *cdrlog in first place
*/
func SetCdrStorage(cStorage CdrStorage) {
	cdrStorage = cStorage
}

func SetPubSub(ps rpcclient.RpcClientConnection) {
	pubSubServer = ps
}

func SetUserService(us rpcclient.RpcClientConnection) {
	userService = us
}

func SetAliasService(as rpcclient.RpcClientConnection) {
	aliasService = as
}

func Publish(event CgrEvent) {
	if pubSubServer != nil {
		var s string
		pubSubServer.Call("PubSubV1.Publish", event, &s)
	}
}

/*
The input stucture that contains call information.
*/
type CallDescriptor struct {
	Direction                             string
	Category                              string
	Tenant, Subject, Account, Destination string
	TimeStart, TimeEnd                    time.Time
	LoopIndex                             float64       // indicates the position of this segment in a cost request loop
	DurationIndex                         time.Duration // the call duration so far (till TimeEnd)
	FallbackSubject                       string        // the subject to check for destination if not found on primary subject
	RatingInfos                           RatingInfos
	Increments                            Increments
	TOR                                   string            // used unit balances selector
	ExtraFields                           map[string]string // Extra fields, mostly used for user profile matching
	// session limits
	MaxRate             float64
	MaxRateUnit         time.Duration
	MaxCostSoFar        float64
	CgrID               string
	RunID               string
	ForceDuration       bool // for Max debit if less than duration return err
	PerformRounding     bool // flag for rating info rounding
	DryRun              bool
	DenyNegativeAccount bool // prevent account going on negative during debit
	account             *Account
	testCallcost        *CallCost // testing purpose only!
}

// AsCGREvent converts the CallDescriptor into CGREvent
func (cd *CallDescriptor) AsCGREvent() *utils.CGREvent {
	cgrEv := &utils.CGREvent{
		Tenant:  cd.Tenant,
		ID:      utils.UUIDSha1Prefix(), // make it unique
		Context: utils.StringPointer(utils.MetaRating),
		Event:   make(map[string]interface{}),
	}
	for k, v := range cd.ExtraFields {
		cgrEv.Event[k] = v
	}
	cgrEv.Event[utils.TOR] = cd.TOR
	cgrEv.Event[utils.Tenant] = cd.Tenant
	cgrEv.Event[utils.Category] = cd.Category
	cgrEv.Event[utils.Account] = cd.Account
	cgrEv.Event[utils.Subject] = cd.Subject
	cgrEv.Event[utils.Destination] = cd.Destination
	cgrEv.Event[utils.AnswerTime] = cd.TimeStart
	cgrEv.Event[utils.Usage] = cd.TimeEnd.Sub(cd.TimeStart)
	return cgrEv
}

// UpdateFromCGREvent will update CallDescriptor with fields from CGREvent
// cgrEv contains both fields and their values
// fields represent fields needing update
func (cd *CallDescriptor) UpdateFromCGREvent(cgrEv *utils.CGREvent, fields []string) (err error) {
	for _, fldName := range fields {
		switch fldName {
		case utils.TOR:
			if cd.TOR, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.Tenant:
			if cd.Tenant, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.Category:
			if cd.Category, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.Account:
			if cd.Account, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.Subject:
			if cd.Subject, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.Destination:
			if cd.Destination, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.AnswerTime:
			if cd.TimeStart, err = cgrEv.FieldAsTime(fldName, config.CgrConfig().DefaultTimezone); err != nil {
				return
			}
		case utils.Usage:
			usage, err := cgrEv.FieldAsDuration(fldName)
			if err != nil {
				return err
			}
			cd.TimeEnd = cd.TimeStart.Add(usage)
		default:
			fldVal, err := cgrEv.FieldAsString(fldName)
			if err != nil {
				return err
			}
			cd.ExtraFields[fldName] = fldVal
		}
	}
	return
}

func (cd *CallDescriptor) ValidateCallData() error {
	if cd.TimeStart.After(cd.TimeEnd) || cd.TimeStart.Equal(cd.TimeEnd) {
		return errors.New("TimeStart must be strctly before TimeEnd")
	}
	if cd.TimeEnd.Sub(cd.TimeStart) < cd.DurationIndex {
		return errors.New("DurationIndex must be equal or greater than TimeEnd - TimeStart")
	}
	return nil
}

// Adds a rating plan that applyes to current call descriptor.
func (cd *CallDescriptor) AddRatingInfo(ris ...*RatingInfo) {
	cd.RatingInfos = append(cd.RatingInfos, ris...)
}

// Gets and caches the user balance information.
func (cd *CallDescriptor) getAccount() (ub *Account, err error) {
	if cd.account == nil {
		cd.account, err = dm.DataDB().GetAccount(cd.GetAccountKey())
	}
	if cd.account != nil && cd.account.Disabled {
		return nil, utils.ErrAccountDisabled
	}
	if err != nil || cd.account == nil {
		utils.Logger.Warning(fmt.Sprintf("Account: %s, not found (%v)", cd.GetAccountKey(), err))
		return nil, utils.ErrAccountNotFound
	}
	return cd.account, err
}

/*
Restores the activation periods for the specified prefix from storage.
*/
func (cd *CallDescriptor) LoadRatingPlans() (err error) {
	var rec int
	err, rec = cd.getRatingPlansForPrefix(cd.GetKey(cd.Subject), 1)
	if err == utils.ErrNotFound && rec == 1 {
		//if err != nil || !cd.continousRatingInfos() {
		// use the default subject only if the initial one was not found
		err, _ = cd.getRatingPlansForPrefix(cd.GetKey(FALLBACK_SUBJECT), 1)
	}
	//load the rating plans
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Rating plan not found for destination %s and account: %s, subject: %s", cd.Destination, cd.GetAccountKey(), cd.GetKey(cd.Subject)))
		return utils.ErrRatingPlanNotFound

	}
	if !cd.continousRatingInfos() {
		utils.Logger.Err(fmt.Sprintf("Destination %s not authorized for account: %s, subject: %s", cd.Destination, cd.GetAccountKey(), cd.GetKey(cd.Subject)))
		return utils.ErrUnauthorizedDestination
	}
	return
}

// FIXME: this method is not exhaustive but will cover 99% of cases just good
// it will not cover very long calls with very short activation periods for rates
func (cd *CallDescriptor) getRatingPlansForPrefix(key string, recursionDepth int) (error, int) {
	if recursionDepth > RECURSION_MAX_DEPTH {
		return utils.ErrMaxRecursionDepth, recursionDepth
	}
	rpf, err := RatingProfileSubjectPrefixMatching(key)
	if err != nil || rpf == nil {
		return utils.ErrNotFound, recursionDepth
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
					Category:    cd.Category,
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
					if err, _ := tempCD.getRatingPlansForPrefix(fbk, recursionDepth); err != nil {
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
	return nil, recursionDepth
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
			for i, newRi := range ris {
				_ = i
				_ = newRi
			}
		}
	}
	return true
}

// GetKey constructs the key for the storage lookup.
// The prefixLen is limiting the length of the destination prefix.
func (cd *CallDescriptor) GetKey(subject string) string {
	return utils.ConcatenatedKey(cd.Direction, cd.Tenant, cd.Category, subject)
}

// GetAccountKey returns the key used to retrive the user balance involved in this call
func (cd *CallDescriptor) GetAccountKey() string {
	subj := cd.Subject
	if cd.Account != "" {
		subj = cd.Account
	}
	return utils.ConcatenatedKey(cd.Tenant, subj)
}

// Splits the received timespan into sub time spans according to the activation periods intervals.
func (cd *CallDescriptor) splitInTimeSpans() (timespans []*TimeSpan) {
	firstSpan := &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd, DurationIndex: cd.DurationIndex}

	timespans = append(timespans, firstSpan)
	if len(cd.RatingInfos) == 0 {
		return
	}
	firstSpan.setRatingInfo(cd.RatingInfos[0])
	if cd.TOR == utils.VOICE {
		// split on rating plans
		afterStart, afterEnd := false, false //optimization for multiple activation periods
		for _, rp := range cd.RatingInfos {
			//log.Print("RP: ", utils.ToJSON(rp))
			if !afterStart && !afterEnd && rp.ActivationTime.Before(cd.TimeStart) {
				firstSpan.setRatingInfo(rp)
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
		//log.Printf("After SplitByRatingPlan: %+v", utils.ToJSON(timespans))
		// split on days

		for i := 0; i < len(timespans); i++ {
			rp := timespans[i].ratingInfo
			newTs := timespans[i].SplitByDay()
			if newTs != nil {
				//log.Print("NEW TS: ", newTs.TimeStart, newTs.TimeEnd)
				newTs.setRatingInfo(rp)
				// insert the new timespan
				index := i + 1
				timespans = append(timespans, nil)
				copy(timespans[index+1:], timespans[index:])
				timespans[index] = newTs
			}
		}
	}
	//log.Printf("After SplitByDay: %+v", utils.ToJSON(timespans))
	// split on rate intervals

	for i := 0; i < len(timespans); i++ {
		//log.Printf("==============%v==================", i)
		//log.Printf("TS: %+v", timespans[i])
		rp := timespans[i].ratingInfo
		//timespans[i].RatingPlan = nil
		rateIntervals := rp.SelectRatingIntevalsForTimespan(timespans[i])
		//log.Print("RIs: ", utils.ToJSON(rateIntervals))
		/*for _, interval := range rp.RateIntervals {
			if !timespans[i].hasBetterRateIntervalThan(interval) {
				timespans[i].SetRateInterval(interval)
			}
		}*/
		//log.Print("ORIG TS: ", timespans[i].TimeStart, timespans[i].TimeEnd)
		//log.Print(timespans[i].RateInterval)
		for _, interval := range rateIntervals {
			//log.Printf("\tINTERVAL: %+v", interval.Timing)
			newTs := timespans[i].SplitByRateInterval(interval, cd.TOR != utils.VOICE)
			//utils.PrintFull(timespans[i])
			//utils.PrintFull(newTs)
			if newTs != nil {
				//log.Print("NEW TS: ", newTs.TimeStart, newTs.TimeEnd)
				newTs.setRatingInfo(rp)
				// insert the new timespan
				index := i + 1
				timespans = append(timespans, nil)
				copy(timespans[index+1:], timespans[index:])
				timespans[index] = newTs
				if timespans[i].RateInterval != nil {
					break
				}
			}
		}
		//log.Print("TS: ", timespans[i].TimeStart, timespans[i].TimeEnd)
		//log.Print(timespans[i].RateInterval.Timing)
	}

	//log.Printf("After SplitByRateInterval: %+v", timespans[0].RateInterval.Timing)
	timespans = cd.roundTimeSpansToIncrement(timespans)
	//log.Printf("After round: %+v", timespans[0].RateInterval.Timing)
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
				rateIncrement = utils.RoundDuration(rateIncrement, ts.GetDuration())
			}
			if rateIncrement > ts.GetDuration() {
				initialDuration := ts.GetDuration()
				ts.TimeEnd = ts.TimeStart.Add(rateIncrement)
				ts.DurationIndex = ts.DurationIndex + (rateIncrement - initialDuration)
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
	cd.account = nil // make sure it's not cached
	cc, err := cd.getCost()
	if err != nil || cd.GetDuration() == 0 {
		return cc, err
	}

	cost := 0.0
	for i, ts := range cc.Timespans {
		// only add connect fee if this is the first/only call cost request
		//log.Printf("Interval: %+v", ts.RateInterval.Timing)
		if cd.LoopIndex == 0 && i == 0 && ts.RateInterval != nil {
			cost += ts.RateInterval.Rating.ConnectFee
		}
		//log.Printf("TS: %+v", ts)
		// handle max cost
		maxCost, strategy := ts.RateInterval.GetMaxCost()

		ts.Cost = ts.CalculateCost()
		cost += ts.Cost
		cd.MaxCostSoFar += cost
		//log.Print("Before: ", cost)
		if strategy != "" && maxCost > 0 {
			//log.Print("HERE: ", strategy, maxCost)
			if strategy == utils.MAX_COST_FREE && cd.MaxCostSoFar >= maxCost {
				cost = maxCost
				cd.MaxCostSoFar = maxCost
			}

		}
		//log.Print("Cost: ", cost)
	}
	cc.Cost = cost
	// global rounding
	roundingDecimals, roundingMethod := cc.GetLongestRounding()
	cc.Cost = utils.Round(cc.Cost, roundingDecimals, roundingMethod)

	return cc, nil
}

func (cd *CallDescriptor) getCost() (*CallCost, error) {
	// check for 0 duration
	if cd.GetDuration() == 0 {
		cc := cd.CreateCallCost()
		// add RatingInfo
		err := cd.LoadRatingPlans()
		if err == nil && len(cd.RatingInfos) > 0 {
			ts := &TimeSpan{
				TimeStart: cd.TimeStart,
				TimeEnd:   cd.TimeEnd,
			}
			ts.setRatingInfo(cd.RatingInfos[0])
			cc.Timespans = append(cc.Timespans, ts)
		}
		return cc, nil
	}
	if cd.DurationIndex < cd.TimeEnd.Sub(cd.TimeStart) {
		cd.DurationIndex = cd.TimeEnd.Sub(cd.TimeStart)
	}
	if cd.TOR == "" {
		cd.TOR = utils.VOICE
	}
	err := cd.LoadRatingPlans()
	//log.Print("ERR: ", err)
	//log.Print("RI: ", utils.ToJSON(cd.RatingInfos))
	if err != nil {
		//utils.Logger.Err(fmt.Sprintf("error getting cost for key <%s>: %s", cd.GetKey(cd.Subject), err.Error()))
		return &CallCost{Cost: -1}, err
	}
	timespans := cd.splitInTimeSpans()
	cost := 0.0

	for i, ts := range timespans {
		ts.createIncrementsSlice()
		// only add connect fee if this is the first/only call cost request
		//log.Printf("Interval: %+v", ts.RateInterval.Timing)
		if cd.LoopIndex == 0 && i == 0 && ts.RateInterval != nil {
			cost += ts.RateInterval.Rating.ConnectFee
		}
		cost += ts.CalculateCost()
	}

	//startIndex := len(fmt.Sprintf("%s:%s:%s:", cd.Direction, cd.Tenant, cd.Category))
	cc := cd.CreateCallCost()
	cc.Cost = cost
	cc.Timespans = timespans

	// global rounding
	roundingDecimals, roundingMethod := cc.GetLongestRounding()
	cc.Cost = utils.Round(cc.Cost, roundingDecimals, roundingMethod)
	//utils.Logger.Info(fmt.Sprintf("<Rater> Get Cost: %s => %v", cd.GetKey(), cc))
	cc.Timespans.Compress()
	cc.UpdateRatedUsage()
	return cc, err
}

/*
Returns the approximate max allowed session for user balance. It will try the max amount received in the call descriptor
If the user has no credit then it will return 0.
If the user has postpayed plan it returns -1.
*/
func (origCD *CallDescriptor) getMaxSessionDuration(origAcc *Account) (time.Duration, error) {
	// clone the account for discarding chenges on debit dry run
	//log.Printf("ORIG CD: %+v", origCD)
	account := origAcc.Clone()
	if account.AllowNegative {
		return -1, nil
	}
	// for zero duration index
	if origCD.DurationIndex < origCD.TimeEnd.Sub(origCD.TimeStart) {
		origCD.DurationIndex = origCD.TimeEnd.Sub(origCD.TimeStart)
	}
	if origCD.TOR == "" {
		origCD.TOR = utils.VOICE
	}
	cd := origCD.Clone()
	initialDuration := cd.TimeEnd.Sub(cd.TimeStart)
	defaultBalance := account.GetDefaultMoneyBalance()

	//use this to check what increment was payed with debt
	initialDefaultBalanceValue := defaultBalance.GetValue()

	cc, err := cd.debit(account, true, false)
	if err != nil {
		return 0, err
	}

	// not enough credit for connect fee
	if cc.negativeConnectFee == true {
		return 0, nil
	}

	var totalCost float64
	var totalDuration time.Duration
	cc.Timespans.Decompress()
	for _, ts := range cc.Timespans {
		if cd.MaxRate > 0 && cd.MaxRateUnit > 0 {
			rate, _, rateUnit := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
			if rate/float64(rateUnit.Nanoseconds()) > cd.MaxRate/float64(cd.MaxRateUnit.Nanoseconds()) {
				return utils.MinDuration(initialDuration, totalDuration), nil
			}
		}
		if ts.Increments == nil {
			ts.createIncrementsSlice()
		}
		for _, incr := range ts.Increments {
			totalCost += incr.Cost
			if incr.BalanceInfo.Monetary != nil && incr.BalanceInfo.Monetary.UUID == defaultBalance.Uuid {
				initialDefaultBalanceValue -= incr.Cost
				if initialDefaultBalanceValue < 0 {
					// this increment was payed with debt
					// TODO: improve this check
					return utils.MinDuration(initialDuration, totalDuration), nil

				}
			}
			totalDuration += incr.Duration
			//log.Print("INC: ", utils.ToJSON(incr))
			if totalDuration >= initialDuration {
				// we have enough, return
				return initialDuration, nil
			}
		}
	}
	return utils.MinDuration(initialDuration, totalDuration), nil
}

func (cd *CallDescriptor) GetMaxSessionDuration() (duration time.Duration, err error) {
	cd.account = nil // make sure it's not cached
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
		account, err := cd.getAccount()
		if err != nil {
			return 0, err
		}
		acntIDs, err := account.GetUniqueSharedGroupMembers(cd)
		if err != nil {
			return nil, err
		}
		var lkIDs []string
		for acntID := range acntIDs {
			if acntID != cd.GetAccountKey() {
				lkIDs = append(lkIDs, utils.ACCOUNT_PREFIX+acntID)
			}
		}
		_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
			duration, err = cd.getMaxSessionDuration(account)
			return
		}, 0, lkIDs...)
		return
	}, 0, utils.ACCOUNT_PREFIX+cd.GetAccountKey())
	return
}

// Interface method used to add/substract an amount of cents or bonus seconds (as returned by GetCost method)
// from user's money balance.
func (cd *CallDescriptor) debit(account *Account, dryRun bool, goNegative bool) (cc *CallCost, err error) {
	if cd.GetDuration() == 0 {
		cc = cd.CreateCallCost()
		// add RatingInfo
		err := cd.LoadRatingPlans()
		if err == nil && len(cd.RatingInfos) > 0 {
			ts := &TimeSpan{
				TimeStart: cd.TimeStart,
				TimeEnd:   cd.TimeEnd,
			}
			ts.setRatingInfo(cd.RatingInfos[0])
			cc.Timespans = append(cc.Timespans, ts)
		}
		return cc, nil
	}
	if cd.TOR == "" {
		cd.TOR = utils.VOICE
	}
	//log.Printf("Debit CD: %+v", cd)
	cc, err = account.debitCreditBalance(cd, !dryRun, dryRun, goNegative)
	//log.Printf("HERE: %+v %v", cc, err)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<Rater> Error getting cost for account key <%s>: %s", cd.GetAccountKey(), err.Error()))
		return nil, err
	}
	cc.updateCost()
	cc.UpdateRatedUsage()
	cc.Timespans.Compress()
	if !dryRun {
		dm.DataDB().SetAccount(account)
	}
	if cd.PerformRounding {
		cc.Round()
		roundIncrements := cc.GetRoundIncrements()
		if len(roundIncrements) != 0 {
			rcd := cc.CreateCallDescriptor()
			rcd.Increments = roundIncrements
			rcd.refundRounding()
		}
	}
	//log.Printf("OUT CC: ", cc)
	return
}

func (cd *CallDescriptor) Debit() (cc *CallCost, err error) {
	cd.account = nil // make sure it's not cached
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
		// lock all group members
		account, err := cd.getAccount()
		if err != nil {
			return nil, err
		}
		acntIDs, sgerr := account.GetUniqueSharedGroupMembers(cd)
		if sgerr != nil {
			return nil, sgerr
		}
		var lkIDs []string
		for acntID := range acntIDs {
			if acntID != cd.GetAccountKey() {
				lkIDs = append(lkIDs, utils.ACCOUNT_PREFIX+acntID)
			}
		}
		_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
			cc, err = cd.debit(account, cd.DryRun, !cd.DenyNegativeAccount)
			if err == nil {
				cc.AccountSummary = cd.AccountSummary()
			}
			return
		}, 0, lkIDs...)
		return
	}, 0, utils.ACCOUNT_PREFIX+cd.GetAccountKey())
	return
}

// Interface method used to add/substract an amount of cents or bonus seconds (as returned by GetCost method)
// from user's money balance.
// This methods combines the Debit and GetMaxSessionDuration and will debit the max available time as returned
// by the GetMaxSessionDuration method. The amount filed has to be filled in call descriptor.
func (cd *CallDescriptor) MaxDebit() (cc *CallCost, err error) {
	cd.account = nil // make sure it's not cached
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
		account, err := cd.getAccount()
		if err != nil {
			return nil, err
		}
		//log.Printf("ACC: %+v", account)
		acntIDs, err := account.GetUniqueSharedGroupMembers(cd)
		if err != nil {
			return nil, err
		}
		var lkIDs []string
		for acntID := range acntIDs {
			if acntID != cd.GetAccountKey() {
				lkIDs = append(lkIDs, utils.ACCOUNT_PREFIX+acntID)
			}
		}
		_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
			remainingDuration, err := cd.getMaxSessionDuration(account)
			if err != nil && cd.GetDuration() > 0 {
				return nil, err
			}
			// check ForceDuartion
			if cd.ForceDuration && !account.AllowNegative && remainingDuration < cd.GetDuration() {
				return nil, utils.ErrInsufficientCredit
			}
			//log.Print("AFTER MAX SESSION: ", cd)
			if err != nil || remainingDuration == 0 {
				cc = cd.CreateCallCost()
				cc.AccountSummary = cd.AccountSummary()
				if cd.GetDuration() == 0 {
					// add RatingInfo
					err = cd.LoadRatingPlans()
					if err == nil && len(cd.RatingInfos) > 0 {
						ts := &TimeSpan{
							TimeStart: cd.TimeStart,
							TimeEnd:   cd.TimeEnd,
						}
						ts.setRatingInfo(cd.RatingInfos[0])
						cc.Timespans = append(cc.Timespans, ts)
					}
					return
				}
				return
			}
			//log.Print("Remaining: ", remainingDuration)
			if remainingDuration > 0 { // for postpaying client returns -1
				initialDuration := cd.GetDuration()
				cd.TimeEnd = cd.TimeStart.Add(remainingDuration)
				cd.DurationIndex -= initialDuration - remainingDuration
			}
			//log.Print("Remaining duration: ", remainingDuration)
			cc, err = cd.debit(account, cd.DryRun, !cd.DenyNegativeAccount)
			if err == nil {
				cc.AccountSummary = cd.AccountSummary()
			}
			//log.Print(balanceMap[0].Value, balanceMap[1].Value)
			return
		}, 0, lkIDs...)
		return
	}, 0, utils.ACCOUNT_PREFIX+cd.GetAccountKey())
	return cc, err
}

// refundIncrements has no locks
func (cd *CallDescriptor) refundIncrements() (err error) {
	accountsCache := make(map[string]*Account)
	for _, increment := range cd.Increments {
		account, found := accountsCache[increment.BalanceInfo.AccountID]
		if !found {
			if acc, err := dm.DataDB().GetAccount(increment.BalanceInfo.AccountID); err == nil && acc != nil {
				account = acc
				accountsCache[increment.BalanceInfo.AccountID] = account
				// will save the account only once at the end of the function
				defer dm.DataDB().SetAccount(account)
			}
		}
		if account == nil {
			utils.Logger.Warning(fmt.Sprintf("Could not get the account to be refunded: %s", increment.BalanceInfo.AccountID))
			continue
		}
		//utils.Logger.Info(fmt.Sprintf("Refunding increment %+v", increment))
		var balance *Balance
		unitType := cd.TOR
		cc := cd.CreateCallCost()
		if increment.BalanceInfo.Unit != nil && increment.BalanceInfo.Unit.UUID != "" {
			if balance = account.BalanceMap[unitType].GetBalance(increment.BalanceInfo.Unit.UUID); balance == nil {
				return
			}
			balance.AddValue(float64(increment.Duration.Nanoseconds()))
			account.countUnits(-float64(increment.Duration.Nanoseconds()), unitType, cc, balance)
		}
		// check money too
		if increment.BalanceInfo.Monetary != nil && increment.BalanceInfo.Monetary.UUID != "" {
			if balance = account.BalanceMap[utils.MONETARY].GetBalance(increment.BalanceInfo.Monetary.UUID); balance == nil {
				return
			}
			balance.AddValue(increment.Cost)
			account.countUnits(-increment.Cost, utils.MONETARY, cc, balance)
		}
	}
	return

}

func (cd *CallDescriptor) RefundIncrements() (err error) {
	// get account list for locking
	// all must be locked in order to use cache
	cd.Increments.Decompress()
	accMap := make(utils.StringMap)
	for _, increment := range cd.Increments {
		if increment.BalanceInfo == nil {
			continue
		}
		if increment.BalanceInfo.Monetary != nil || increment.BalanceInfo.Unit != nil {
			accMap[utils.ACCOUNT_PREFIX+increment.BalanceInfo.AccountID] = true
		}
	}
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
		err = cd.refundIncrements()
		return
	}, 0, accMap.Slice()...)
	return
}

func (cd *CallDescriptor) refundRounding() (err error) {
	// get account list for locking
	// all must be locked in order to use cache
	accountsCache := make(map[string]*Account)
	for _, increment := range cd.Increments {
		account, found := accountsCache[increment.BalanceInfo.AccountID]
		if !found {
			if acc, err := dm.DataDB().GetAccount(increment.BalanceInfo.AccountID); err == nil && acc != nil {
				account = acc
				accountsCache[increment.BalanceInfo.AccountID] = account
				// will save the account only once at the end of the function
				defer dm.DataDB().SetAccount(account)
			}
		}
		if account == nil {
			utils.Logger.Warning(fmt.Sprintf("Could not get the account to be refunded: %s", increment.BalanceInfo.AccountID))
			continue
		}
		cc := cd.CreateCallCost()
		if increment.BalanceInfo.Monetary != nil {
			var balance *Balance
			if balance = account.BalanceMap[utils.MONETARY].GetBalance(increment.BalanceInfo.Monetary.UUID); balance == nil {
				return
			}
			balance.AddValue(-increment.Cost)
			account.countUnits(increment.Cost, utils.MONETARY, cc, balance)
		}
	}
	return
}

func (cd *CallDescriptor) RefundRounding() (err error) {
	accMap := make(utils.StringMap)
	for _, inc := range cd.Increments {
		accMap[utils.ACCOUNT_PREFIX+inc.BalanceInfo.AccountID] = true
	}
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
		err = cd.refundRounding()
		return
	}, 0, accMap.Slice()...)
	return
}

// Creates a CallCost structure copying related data from CallDescriptor
func (cd *CallDescriptor) CreateCallCost() *CallCost {
	return &CallCost{
		Direction:        cd.Direction,
		Category:         cd.Category,
		Tenant:           cd.Tenant,
		Subject:          cd.Subject,
		Account:          cd.Account,
		Destination:      cd.Destination,
		TOR:              cd.TOR,
		deductConnectFee: cd.LoopIndex == 0,
	}
}

func (cd *CallDescriptor) Clone() *CallDescriptor {
	return &CallDescriptor{
		Direction:       cd.Direction,
		Category:        cd.Category,
		Tenant:          cd.Tenant,
		Subject:         cd.Subject,
		Account:         cd.Account,
		Destination:     cd.Destination,
		TimeStart:       cd.TimeStart,
		TimeEnd:         cd.TimeEnd,
		LoopIndex:       cd.LoopIndex,
		DurationIndex:   cd.DurationIndex,
		MaxRate:         cd.MaxRate,
		MaxRateUnit:     cd.MaxRateUnit,
		MaxCostSoFar:    cd.MaxCostSoFar,
		FallbackSubject: cd.FallbackSubject,
		//RatingInfos:     cd.RatingInfos,
		//Increments:      cd.Increments,
		TOR:             cd.TOR,
		ForceDuration:   cd.ForceDuration,
		PerformRounding: cd.PerformRounding,
		DryRun:          cd.DryRun,
		CgrID:           cd.CgrID,
		RunID:           cd.RunID,
	}
}

func (cd *CallDescriptor) GetLCRFromStorage() (*LCR, error) {
	keyVariants := []string{
		utils.LCRKey(cd.Direction, cd.Tenant, cd.Category, cd.Account, cd.Subject),
		utils.LCRKey(cd.Direction, cd.Tenant, cd.Category, cd.Account, utils.ANY),
		utils.LCRKey(cd.Direction, cd.Tenant, cd.Category, utils.ANY, utils.ANY),
		utils.LCRKey(cd.Direction, cd.Tenant, utils.ANY, utils.ANY, utils.ANY),
		utils.LCRKey(utils.ANY, utils.ANY, cd.Category, utils.ANY, utils.ANY),
		utils.LCRKey(utils.ANY, utils.ANY, utils.ANY, utils.ANY, utils.ANY),
	}
	if lcrSubjectPrefixMatching {
		var partialSubjects []string
		lenSubject := len(cd.Subject)
		for i := 1; i < lenSubject; i++ {
			partialSubjects = append(partialSubjects, utils.LCRKey(cd.Direction, cd.Tenant, cd.Category, cd.Account, cd.Subject[:lenSubject-i]))
		}
		// insert partialsubjects into keyVariants
		keyVariants = append(keyVariants[:1], append(partialSubjects, keyVariants[1:]...)...)
	}
	for _, key := range keyVariants {
		if lcr, err := dm.GetLCR(key, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return nil, err
		} else if err == nil && lcr != nil {
			return lcr, nil
		}
	}
	return nil, utils.ErrNotFound
}

func (cd *CallDescriptor) GetLCR(stats rpcclient.RpcClientConnection, lcrFltr *LCRFilter, p *utils.Paginator) (*LCRCost, error) {
	cd.account = nil // make sure it's not cached
	lcr, err := cd.GetLCRFromStorage()
	if err != nil {
		return nil, err
	}
	// sort by activation time
	lcr.Sort()
	// find if one ore more entries apply to this cd (create lcr timespans)
	// create timespans and attach lcr entries to them
	lcrCost := &LCRCost{}
	for _, lcrActivation := range lcr.Activations {
		lcrEntry := lcrActivation.GetLCREntryForPrefix(cd.Destination)
		if lcrActivation.ActivationTime.Before(cd.TimeStart) ||
			lcrActivation.ActivationTime.Equal(cd.TimeStart) {
			lcrCost.Entry = lcrEntry
		} else {
			// because lcr is sorted the folowing ones will
			// only activate later than cd.Timestart
			break
		}
	}
	if lcrCost.Entry == nil {
		return lcrCost, nil
	}
	if lcrCost.Entry.Strategy == LCR_STRATEGY_STATIC {
		for _, supplier := range lcrCost.Entry.GetParams() {
			lcrCD := cd.Clone()
			lcrCD.Account = supplier
			lcrCD.Subject = supplier
			lcrCD.Category = lcrCost.Entry.RPCategory
			fullSupplier := utils.ConcatenatedKey(lcrCD.Direction, lcrCD.Tenant, lcrCD.Category, lcrCD.Subject)
			var cc *CallCost
			var err error
			if cd.account, err = dm.DataDB().GetAccount(lcrCD.GetAccountKey()); err == nil {
				if cd.account.Disabled {
					lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
						Supplier: fullSupplier,
						Error:    fmt.Sprintf("supplier %s is disabled", supplier),
					})
					continue
				}
				cc, err = lcrCD.debit(cd.account, true, true)
			} else {
				cc, err = lcrCD.GetCost()
			}
			//log.Printf("CC: %+v", cc.Timespans[0].ratingInfo.RateIntervals[0].Rating.Rates[0])
			if err != nil || cc == nil {
				lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
					Supplier: fullSupplier,
					Error:    err.Error(),
				})
			} else {
				lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
					Supplier: fullSupplier,
					Cost:     cc.Cost,
					Duration: cc.GetDuration(),
				})
			}
		}
	} else {
		// find rating profiles
		category := lcrCost.Entry.RPCategory
		if category == utils.META_DEFAULT {
			category = lcr.Category
		}
		ratingProfileSearchKey := utils.ConcatenatedKey(lcr.Direction, lcr.Tenant, lcrCost.Entry.RPCategory)
		searchKey := utils.RATING_PROFILE_PREFIX + ratingProfileSearchKey
		suppliers := cache.GetEntryKeys(searchKey)
		if len(suppliers) == 0 { // Most probably the data was not cached, do it here, #ToDo: move logic in RAL service
			suppliers, err = dm.DataDB().GetKeysForPrefix(searchKey)
			if err != nil {
				return nil, err
			}
			transID := utils.GenUUID()
			for _, dbKey := range suppliers {
				if _, err := dm.GetRatingProfile(dbKey[len(utils.RATING_PROFILE_PREFIX):], true, transID); err != nil { // cache the keys here
					cache.RollbackTransaction(transID)
					return nil, err
				}
			}
			cache.CommitTransaction(transID)
		}
		for _, supplier := range suppliers {
			split := strings.Split(supplier, ":")
			supplier = split[len(split)-1]
			lcrCD := cd.Clone()
			lcrCD.Category = category
			lcrCD.Account = supplier
			lcrCD.Subject = supplier
			fullSupplier := utils.ConcatenatedKey(lcrCD.Direction, lcrCD.Tenant, lcrCD.Category, lcrCD.Subject)
			var qosSortParams []string
			var asrValues sort.Float64Slice
			var pddValues sort.Float64Slice
			var acdValues sort.Float64Slice
			var tcdValues sort.Float64Slice
			var accValues sort.Float64Slice
			var tccValues sort.Float64Slice
			var ddcValues sort.Float64Slice
			// track if one value is never calculated
			asrNeverConsidered := true
			pddNeverConsidered := true
			acdNeverConsidered := true
			tcdNeverConsidered := true
			accNeverConsidered := true
			tccNeverConsidered := true
			ddcNeverConsidered := true
			if utils.IsSliceMember([]string{LCR_STRATEGY_QOS, LCR_STRATEGY_QOS_THRESHOLD, LCR_STRATEGY_LOAD}, lcrCost.Entry.Strategy) {
				if stats == nil {
					lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
						Supplier: fullSupplier,
						Error:    fmt.Sprintf("Cdr stats service not configured"),
					})
					continue
				}
				rpfKey := utils.ConcatenatedKey(ratingProfileSearchKey, supplier)
				if rpf, err := RatingProfileSubjectPrefixMatching(rpfKey); err != nil {
					lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
						Supplier: fullSupplier,
						Error:    fmt.Sprintf("Rating plan error: %s", err.Error()),
					})
					continue
				} else if rpf != nil {
					rpf.RatingPlanActivations.Sort()
					activeRas := rpf.RatingPlanActivations.GetActiveForCall(cd)
					var cdrCDRStatsQueueIds []string
					for _, ra := range activeRas {
						for _, qId := range ra.CdrStatQueueIds {
							if qId != "" {
								cdrCDRStatsQueueIds = append(cdrCDRStatsQueueIds, qId)
							}
						}
					}

					statsErr := false
					var supplierQueues []*CDRStatsQueue
					for _, qId := range cdrCDRStatsQueueIds {
						if lcrCost.Entry.Strategy == LCR_STRATEGY_LOAD {
							for _, qId := range cdrCDRStatsQueueIds {
								sq := &CDRStatsQueue{}
								if err := stats.Call("CDRStatsV1.GetQueue", qId, sq); err == nil {
									if sq.conf.QueueLength == 0 { //only add qeues that don't have fixed length
										supplierQueues = append(supplierQueues, sq)
									}
								}
							}
						} else {
							statValues := make(map[string]float64)
							if err := stats.Call("CDRStatsV1.GetValues", qId, &statValues); err != nil {
								lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
									Supplier: fullSupplier,
									Error:    fmt.Sprintf("Get stats values for queue id %s, error %s", qId, err.Error()),
								})
								statsErr = true
								break
							}
							if asr, exists := statValues[ASR]; exists {
								if asr > STATS_NA {
									asrValues = append(asrValues, asr)
								}
								asrNeverConsidered = false
							}
							if pdd, exists := statValues[PDD]; exists {
								if pdd > STATS_NA {
									pddValues = append(pddValues, pdd)
								}
								pddNeverConsidered = false
							}
							if acd, exists := statValues[ACD]; exists {
								if acd > STATS_NA {
									acdValues = append(acdValues, acd)
								}
								acdNeverConsidered = false
							}
							if tcd, exists := statValues[TCD]; exists {
								if tcd > STATS_NA {
									tcdValues = append(tcdValues, tcd)
								}
								tcdNeverConsidered = false
							}
							if acc, exists := statValues[ACC]; exists {
								if acc > STATS_NA {
									accValues = append(accValues, acc)
								}
								accNeverConsidered = false
							}
							if tcc, exists := statValues[TCC]; exists {
								if tcc > STATS_NA {
									tccValues = append(tccValues, tcc)
								}
								tccNeverConsidered = false
							}
							if ddc, exists := statValues[TCC]; exists {
								if ddc > STATS_NA {
									ddcValues = append(ddcValues, ddc)
								}
								ddcNeverConsidered = false
							}

						}
					}
					if lcrCost.Entry.Strategy == LCR_STRATEGY_LOAD {
						if len(supplierQueues) > 0 {
							lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
								Supplier:       fullSupplier,
								supplierQueues: supplierQueues,
							})
						}
						continue // next supplier
					}

					if statsErr { // Stats error in loop, to go next supplier
						continue
					}
					asrValues.Sort()
					pddValues.Sort()
					acdValues.Sort()
					tcdValues.Sort()
					accValues.Sort()
					tccValues.Sort()
					ddcValues.Sort()

					//log.Print(asrValues, acdValues)
					if utils.IsSliceMember([]string{LCR_STRATEGY_QOS_THRESHOLD, LCR_STRATEGY_QOS}, lcrCost.Entry.Strategy) {
						qosSortParams = lcrCost.Entry.GetParams()
					}
					if lcrCost.Entry.Strategy == LCR_STRATEGY_QOS_THRESHOLD {
						// filter suppliers by qos thresholds
						asrMin, asrMax, pddMin, pddMax, acdMin, acdMax, tcdMin, tcdMax, accMin, accMax, tccMin, tccMax, ddcMin, ddcMax := lcrCost.Entry.GetQOSLimits()
						//log.Print(asrMin, asrMax, acdMin, acdMax)
						// skip current supplier if off limits
						if asrMin > 0 && len(asrValues) != 0 && asrValues[0] < asrMin {
							continue
						}
						if asrMax > 0 && len(asrValues) != 0 && asrValues[len(asrValues)-1] > asrMax {
							continue
						}
						if pddMin > 0 && len(pddValues) != 0 && pddValues[0] < pddMin.Seconds() {
							continue
						}
						if pddMax > 0 && len(pddValues) != 0 && pddValues[len(pddValues)-1] > pddMax.Seconds() {
							continue
						}
						if acdMin > 0 && len(acdValues) != 0 && acdValues[0] < acdMin.Seconds() {
							continue
						}
						if acdMax > 0 && len(acdValues) != 0 && acdValues[len(acdValues)-1] > acdMax.Seconds() {
							continue
						}
						if tcdMin > 0 && len(tcdValues) != 0 && tcdValues[0] < tcdMin.Seconds() {
							continue
						}
						if tcdMax > 0 && len(tcdValues) != 0 && tcdValues[len(tcdValues)-1] > tcdMax.Seconds() {
							continue
						}
						if accMin > 0 && len(accValues) != 0 && accValues[0] < accMin {
							continue
						}
						if accMax > 0 && len(accValues) != 0 && accValues[len(accValues)-1] > accMax {
							continue
						}
						if tccMin > 0 && len(tccValues) != 0 && tccValues[0] < tccMin {
							continue
						}
						if tccMax > 0 && len(tccValues) != 0 && tccValues[len(tccValues)-1] > tccMax {
							continue
						}
						if ddcMin > 0 && len(ddcValues) != 0 && ddcValues[0] < ddcMin {
							continue
						}
						if ddcMax > 0 && len(ddcValues) != 0 && ddcValues[len(ddcValues)-1] > ddcMax {
							continue
						}
					}
				}
			}

			var cc *CallCost
			var err error
			if cd.account, err = dm.DataDB().GetAccount(lcrCD.GetAccountKey()); err == nil {
				if cd.account.Disabled {
					lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
						Supplier: fullSupplier,
						Error:    fmt.Sprintf("supplier %s is disabled", supplier),
					})
					continue
				}
				cc, err = lcrCD.debit(cd.account, true, true)
			} else {
				cc, err = lcrCD.GetCost()
			}
			if err != nil || cc == nil {
				lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, &LCRSupplierCost{
					Supplier: fullSupplier,
					Error:    err.Error(),
				})
				continue
			} else {
				if lcrFltr != nil {
					if lcrFltr.MinCost != nil && cc.Cost < *lcrFltr.MinCost {
						continue // MinCost not reached, ignore the supplier
					}
					if lcrFltr.MaxCost != nil && cc.Cost >= *lcrFltr.MaxCost {
						continue // Equal or higher than MaxCost allowed, ignore the supplier
					}
				}
				supplCost := &LCRSupplierCost{
					Supplier: fullSupplier,
					Cost:     cc.Cost,
					Duration: cc.GetDuration(),
				}
				qos := make(map[string]float64, 5)
				if !asrNeverConsidered {
					qos[ASR] = utils.AvgNegative(asrValues)
				}
				if !pddNeverConsidered {
					qos[PDD] = utils.AvgNegative(pddValues)
				}
				if !acdNeverConsidered {
					qos[ACD] = utils.AvgNegative(acdValues)
				}
				if !tcdNeverConsidered {
					qos[TCD] = utils.AvgNegative(tcdValues)
				}
				if !accNeverConsidered {
					qos[ACC] = utils.AvgNegative(accValues)
				}
				if !tccNeverConsidered {
					qos[TCC] = utils.AvgNegative(tccValues)
				}
				if !ddcNeverConsidered {
					qos[DDC] = utils.AvgNegative(ddcValues)
				}
				if utils.IsSliceMember([]string{LCR_STRATEGY_QOS, LCR_STRATEGY_QOS_THRESHOLD}, lcrCost.Entry.Strategy) {
					supplCost.QOS = qos
					supplCost.qosSortParams = qosSortParams
				}
				lcrCost.SupplierCosts = append(lcrCost.SupplierCosts, supplCost)
			}
		}
		// sort according to strategy
		lcrCost.Sort()
	}
	if p != nil {
		if p.Offset != nil && *p.Offset > 0 && *p.Offset < len(lcrCost.SupplierCosts) {
			lcrCost.SupplierCosts = lcrCost.SupplierCosts[*p.Offset:]
		}
		if p.Limit != nil && *p.Limit > 0 && *p.Limit < len(lcrCost.SupplierCosts) {
			lcrCost.SupplierCosts = lcrCost.SupplierCosts[:*p.Limit]
		}
	}
	return lcrCost, nil
}

// AccountSummary returns the AccountSummary for cached account
func (cd *CallDescriptor) AccountSummary() *AccountSummary {
	if cd.account == nil {
		return nil
	}
	return cd.account.AsAccountSummary()
}
