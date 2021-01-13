/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

const (
	// these might be better in the confs under optimizations section
	RECURSION_MAX_DEPTH = 3
	MIN_PREFIX_MATCH    = 1
	FALLBACK_SUBJECT    = utils.MetaAny
)

var (
	debitPeriod                  = 10 * time.Second
	globalRoundingDecimals       = 6
	rpSubjectPrefixMatching      bool
	rpSubjectPrefixMatchingMutex sync.RWMutex // used to reload rpSubjectPrefixMatching
)

// SetRoundingDecimals sets the global rounding method and decimal precision for GetCost method
func SetRoundingDecimals(rd int) {
	globalRoundingDecimals = rd
}

// SetRpSubjectPrefixMatching sets rpSubjectPrefixMatching (is thread safe)
func SetRpSubjectPrefixMatching(flag bool) {
	rpSubjectPrefixMatchingMutex.Lock()
	rpSubjectPrefixMatching = flag
	rpSubjectPrefixMatchingMutex.Unlock()
}

// getRpSubjectPrefixMatching returns rpSubjectPrefixMatching (is thread safe)
func getRpSubjectPrefixMatching() (flag bool) {
	rpSubjectPrefixMatchingMutex.RLock()
	flag = rpSubjectPrefixMatching
	rpSubjectPrefixMatchingMutex.RUnlock()
	return
}

// NewCallDescriptorFromCGREvent converts a CGREvent into CallDescriptor
func NewCallDescriptorFromCGREvent(cgrEv *utils.CGREvent,
	timezone string) (cd *CallDescriptor, err error) {
	cd = &CallDescriptor{Tenant: cgrEv.Tenant}
	if _, has := cgrEv.Event[utils.Category]; has {
		if cd.Category, err = cgrEv.FieldAsString(utils.Category); err != nil {
			return nil, err
		}
	}
	if cd.Account, err = cgrEv.FieldAsString(utils.AccountField); err != nil {
		return
	}
	if cd.Subject, err = cgrEv.FieldAsString(utils.Subject); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		cd.Subject = cd.Account
	}
	if cd.Destination, err = cgrEv.FieldAsString(utils.Destination); err != nil {
		return nil, err
	}
	if cd.TimeStart, err = cgrEv.FieldAsTime(utils.SetupTime,
		timezone); err != nil {
		return nil, err
	}
	if _, has := cgrEv.Event[utils.AnswerTime]; has { // AnswerTime takes precendence for TimeStart
		if aTime, err := cgrEv.FieldAsTime(utils.AnswerTime,
			timezone); err != nil {
			return nil, err
		} else if !aTime.IsZero() {
			cd.TimeStart = aTime
		}
	}
	if usage, err := cgrEv.FieldAsDuration(utils.Usage); err != nil {
		return nil, err
	} else {
		cd.TimeEnd = cd.TimeStart.Add(usage)
	}
	if _, has := cgrEv.Event[utils.ToR]; has {
		if cd.ToR, err = cgrEv.FieldAsString(utils.ToR); err != nil {
			return nil, err
		}
	}
	return
}

/*
The input stucture that contains call information.
*/
type CallDescriptor struct {
	Category        string
	Tenant          string
	Subject         string
	Account         string
	Destination     string
	TimeStart       time.Time
	TimeEnd         time.Time
	LoopIndex       float64       // indicates the position of this segment in a cost request loop
	DurationIndex   time.Duration // the call duration so far (till TimeEnd)
	FallbackSubject string        // the subject to check for destination if not found on primary subject
	RatingInfos     RatingInfos
	Increments      Increments
	ToR             string            // used unit balances selector
	ExtraFields     map[string]string // Extra fields, mostly used for user profile matching
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
func (cd *CallDescriptor) AsCGREvent(opts map[string]interface{}) *utils.CGREvent {
	cgrEv := &utils.CGREvent{
		Tenant: cd.Tenant,
		ID:     utils.UUIDSha1Prefix(), // make it unique
		Event:  make(map[string]interface{}),
		Opts:   opts,
	}
	for k, v := range cd.ExtraFields {
		cgrEv.Event[k] = v
	}
	cgrEv.Event[utils.ToR] = cd.ToR
	cgrEv.Event[utils.Tenant] = cd.Tenant
	cgrEv.Event[utils.Category] = cd.Category
	cgrEv.Event[utils.AccountField] = cd.Account
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
		case utils.ToR:
			if cd.ToR, err = cgrEv.FieldAsString(fldName); err != nil {
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
		case utils.AccountField:
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
			if cd.TimeStart, err = cgrEv.FieldAsTime(fldName,
				config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
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
		cd.account, err = dm.GetAccount(cd.GetAccountKey())
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
	return utils.ConcatenatedKey(utils.MetaOut, cd.Tenant, cd.Category, subject)
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
	firstSpan := &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd,
		DurationIndex: cd.DurationIndex}

	timespans = append(timespans, firstSpan)
	if len(cd.RatingInfos) == 0 {
		return
	}
	firstSpan.setRatingInfo(cd.RatingInfos[0])
	if cd.ToR == utils.MetaVoice {
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
			newTs := timespans[i].SplitByRateInterval(interval, cd.ToR != utils.MetaVoice)
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
		if cd.LoopIndex == 0 && i == 0 && ts.RateInterval != nil {
			//Add the ConnectFee increment at the beggining
			ts.Increments = append(Increments{&Increment{
				Duration:       0,
				Cost:           ts.RateInterval.Rating.ConnectFee,
				CompressFactor: 1,
				BalanceInfo: &DebitInfo{
					Monetary:  nil,
					Unit:      nil,
					AccountID: "",
				},
			}}, ts.Increments...)
			//Add the cost from ConnectFee to TimeSpan
			ts.Cost = ts.Cost + ts.RateInterval.Rating.ConnectFee
		}
		// handle max cost
		maxCost, strategy := ts.RateInterval.GetMaxCost()

		ts.Cost = ts.CalculateCost()
		cost += ts.Cost
		cd.MaxCostSoFar += cost

		if strategy != "" && maxCost > 0 {
			//log.Print("HERE: ", strategy, maxCost)
			if strategy == utils.MetaMaxCostFree && cd.MaxCostSoFar >= maxCost {
				cost = maxCost
				cd.MaxCostSoFar = maxCost
			}
		}
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
	if cd.ToR == "" {
		cd.ToR = utils.MetaVoice
	}
	err := cd.LoadRatingPlans()
	if err != nil {
		return &CallCost{Cost: -1}, err
	}
	timespans := cd.splitInTimeSpans()
	cost := 0.0

	for i, ts := range timespans {
		ts.createIncrementsSlice()
		// only add connect fee if this is the first/only call cost request
		if cd.LoopIndex == 0 && i == 0 && ts.RateInterval != nil {
			cost += ts.RateInterval.Rating.ConnectFee
		}
		cost += ts.CalculateCost()
	}

	cc := cd.CreateCallCost()
	cc.Cost = cost
	cc.Timespans = timespans

	// global rounding
	roundingDecimals, roundingMethod := cc.GetLongestRounding()
	cc.Cost = utils.Round(cc.Cost, roundingDecimals, roundingMethod)
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
	account := origAcc.Clone()
	if account.AllowNegative {
		return -1, nil
	}
	// for zero duration index
	if origCD.DurationIndex < origCD.TimeEnd.Sub(origCD.TimeStart) {
		origCD.DurationIndex = origCD.TimeEnd.Sub(origCD.TimeStart)
	}
	if origCD.ToR == "" {
		origCD.ToR = utils.MetaVoice
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
				lkIDs = append(lkIDs, utils.AccountPrefix+acntID)
			}
		}
		_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
			duration, err = cd.getMaxSessionDuration(account)
			return
		}, config.CgrConfig().GeneralCfg().LockingTimeout, lkIDs...)
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+cd.GetAccountKey())
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
	if cd.ToR == "" {
		cd.ToR = utils.MetaVoice
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
		dm.SetAccount(account)
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
		initialAcnt := account.AsAccountSummary()
		acntIDs, sgerr := account.GetUniqueSharedGroupMembers(cd)
		if sgerr != nil {
			return nil, sgerr
		}
		var lkIDs []string
		for acntID := range acntIDs {
			if acntID != cd.GetAccountKey() {
				lkIDs = append(lkIDs, utils.AccountPrefix+acntID)
			}
		}
		_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
			cc, err = cd.debit(account, cd.DryRun, !cd.DenyNegativeAccount)
			if err == nil {
				cc.AccountSummary = cd.AccountSummary(initialAcnt)
			}
			return
		}, config.CgrConfig().GeneralCfg().LockingTimeout, lkIDs...)
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+cd.GetAccountKey())
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
		initialAcnt := account.AsAccountSummary()
		acntIDs, err := account.GetUniqueSharedGroupMembers(cd)
		if err != nil {
			return nil, err
		}
		var lkIDs []string
		for acntID := range acntIDs {
			if acntID != cd.GetAccountKey() {
				lkIDs = append(lkIDs, utils.AccountPrefix+acntID)
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
			if err != nil || remainingDuration == 0 {
				cc = cd.CreateCallCost()
				cc.AccountSummary = cd.AccountSummary(initialAcnt)
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
			if remainingDuration > 0 { // for postpaying client returns -1
				initialDuration := cd.GetDuration()
				cd.TimeEnd = cd.TimeStart.Add(remainingDuration)
				cd.DurationIndex -= initialDuration - remainingDuration
			}
			cc, err = cd.debit(account, cd.DryRun, !cd.DenyNegativeAccount)
			if err == nil {
				cc.AccountSummary = cd.AccountSummary(initialAcnt)
			}
			return
		}, config.CgrConfig().GeneralCfg().LockingTimeout, lkIDs...)
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+cd.GetAccountKey())
	return cc, err
}

// refundIncrements has no locks
// returns the updated account referenced by the CallDescriptor
func (cd *CallDescriptor) refundIncrements() (acnt *Account, err error) {
	accountsCache := make(map[string]*Account)
	for _, increment := range cd.Increments {
		// work around for the refund from CDRServer:
		// for the calls with Cost 0 but with at least a TimeSpan it will make the information
		// from BalanceInfo nil so here we can ignore all increments with BalanceInfo nil
		if increment.BalanceInfo == nil {
			continue
		}

		account, found := accountsCache[increment.BalanceInfo.AccountID]
		if !found {
			if acc, err := dm.GetAccount(increment.BalanceInfo.AccountID); err == nil && acc != nil {
				account = acc
				accountsCache[increment.BalanceInfo.AccountID] = account
				// will save the account only once at the end of the function
				defer dm.SetAccount(account)
			}
		}
		if account == nil {
			utils.Logger.Warning(fmt.Sprintf("Could not get the account to be refunded: %s", increment.BalanceInfo.AccountID))
			continue
		}
		//utils.Logger.Info(fmt.Sprintf("Refunding increment %+v", increment))
		var balance *Balance
		unitType := cd.ToR
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
			if balance = account.BalanceMap[utils.MetaMonetary].GetBalance(increment.BalanceInfo.Monetary.UUID); balance == nil {
				return
			}
			balance.AddValue(increment.Cost)
			account.countUnits(-increment.Cost, utils.MetaMonetary, cc, balance)
		}
	}
	acnt = accountsCache[utils.ConcatenatedKey(cd.Tenant, cd.Account)]
	return

}

func (cd *CallDescriptor) RefundIncrements() (acnt *Account, err error) {
	// get account list for locking
	// all must be locked in order to use cache
	cd.Increments.Decompress()
	accMap := make(utils.StringMap)
	for _, increment := range cd.Increments {
		if increment.BalanceInfo == nil {
			continue
		}
		if increment.BalanceInfo.Monetary != nil || increment.BalanceInfo.Unit != nil {
			accMap[utils.AccountPrefix+increment.BalanceInfo.AccountID] = true
		}
	}
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
		acnt, err = cd.refundIncrements()
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, accMap.Slice()...)
	return
}

func (cd *CallDescriptor) refundRounding() (err error) {
	// get account list for locking
	// all must be locked in order to use cache
	accountsCache := make(map[string]*Account)
	for _, increment := range cd.Increments {
		account, found := accountsCache[increment.BalanceInfo.AccountID]
		if !found {
			if acc, err := dm.GetAccount(increment.BalanceInfo.AccountID); err == nil && acc != nil {
				account = acc
				accountsCache[increment.BalanceInfo.AccountID] = account
				// will save the account only once at the end of the function
				defer dm.SetAccount(account)
			}
		}
		if account == nil {
			utils.Logger.Warning(fmt.Sprintf("Could not get the account to be refunded: %s", increment.BalanceInfo.AccountID))
			continue
		}
		cc := cd.CreateCallCost()
		if increment.BalanceInfo.Monetary != nil {
			var balance *Balance
			if balance = account.BalanceMap[utils.MetaMonetary].GetBalance(increment.BalanceInfo.Monetary.UUID); balance == nil {
				return
			}
			balance.AddValue(-increment.Cost)
			account.countUnits(increment.Cost, utils.MetaMonetary, cc, balance)
		}
	}
	return
}

func (cd *CallDescriptor) RefundRounding() (err error) {
	accMap := make(utils.StringMap)
	for _, inc := range cd.Increments {
		accMap[utils.AccountPrefix+inc.BalanceInfo.AccountID] = true
	}
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) {
		err = cd.refundRounding()
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, accMap.Slice()...)
	return
}

// Creates a CallCost structure copying related data from CallDescriptor
func (cd *CallDescriptor) CreateCallCost() *CallCost {
	return &CallCost{
		Category:         cd.Category,
		Tenant:           cd.Tenant,
		Subject:          cd.Subject,
		Account:          cd.Account,
		Destination:      cd.Destination,
		ToR:              cd.ToR,
		deductConnectFee: cd.LoopIndex == 0,
	}
}

func (cd *CallDescriptor) Clone() *CallDescriptor {
	return &CallDescriptor{
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
		ToR:             cd.ToR,
		ForceDuration:   cd.ForceDuration,
		PerformRounding: cd.PerformRounding,
		DryRun:          cd.DryRun,
		CgrID:           cd.CgrID,
		RunID:           cd.RunID,
	}

}

// AccountSummary returns the AccountSummary for cached account
func (cd *CallDescriptor) AccountSummary(initialAcnt *AccountSummary) *AccountSummary {
	if cd.account == nil {
		return nil
	}
	acntSummary := cd.account.AsAccountSummary()
	for _, initialBal := range initialAcnt.BalanceSummaries {
		for _, currentBal := range acntSummary.BalanceSummaries {
			if currentBal.UUID == initialBal.UUID {
				currentBal.Initial = initialBal.Value
				break
			}
		}
	}
	return acntSummary
}

// FieldAsInterface is part of utils.DataProvider
func (cd *CallDescriptor) FieldAsInterface(fldPath []string) (fldVal interface{}, err error) {
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	return utils.ReflectFieldInterface(cd, fldPath[0], utils.ExtraFields)
}

// FieldAsString is part of utils.DataProvider
func (cd *CallDescriptor) FieldAsString(fldPath []string) (fldVal string, err error) {
	if len(fldPath) == 0 {
		return "", utils.ErrNotFound
	}
	return utils.ReflectFieldAsString(cd, fldPath[0], utils.ExtraFields)
}

// String is part of utils.DataProvider
func (cd *CallDescriptor) String() string {
	return utils.ToJSON(cd)
}

// RemoteHost is part of utils.DataProvider
func (cd *CallDescriptor) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

type CallDescriptorWithOpts struct {
	*CallDescriptor
	Opts map[string]interface{}
}
