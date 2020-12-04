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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/structmatcher"
	"github.com/cgrates/cgrates/utils"
)

// Account structure containing information about user's credit (minutes, cents, sms...).'
// This can represent a user or a shared group.
type Account struct {
	ID                string
	BalanceMap        map[string]Balances
	UnitCounters      UnitCounters
	ActionTriggers    ActionTriggers
	AllowNegative     bool
	Disabled          bool
	UpdateTime        time.Time
	executingTriggers bool
}

type AccountWithOpts struct {
	*Account
	Opts map[string]interface{}
}

// User's available minutes for the specified destination
func (acc *Account) getCreditForPrefix(cd *CallDescriptor) (duration time.Duration, credit float64, balances Balances) {
	creditBalances := acc.getBalancesForPrefix(cd.Destination, cd.Category, utils.MONETARY, "", cd.TimeStart)

	unitBalances := acc.getBalancesForPrefix(cd.Destination, cd.Category, cd.ToR, "", cd.TimeStart)
	// gather all balances from shared groups
	var extendedCreditBalances Balances
	for _, cb := range creditBalances {
		if len(cb.SharedGroups) > 0 {
			for sg := range cb.SharedGroups {
				if sharedGroup, _ := dm.GetSharedGroup(sg, false, utils.NonTransactional); sharedGroup != nil {
					sgb := sharedGroup.GetBalances(cd.Destination, cd.Category, utils.MONETARY, acc, cd.TimeStart)
					sgb = sharedGroup.SortBalancesByStrategy(cb, sgb)
					extendedCreditBalances = append(extendedCreditBalances, sgb...)
				}
			}
		} else {
			extendedCreditBalances = append(extendedCreditBalances, cb)
		}
	}
	var extendedMinuteBalances Balances
	for _, mb := range unitBalances {
		if len(mb.SharedGroups) > 0 {
			for sg := range mb.SharedGroups {
				if sharedGroup, _ := dm.GetSharedGroup(sg, false, utils.NonTransactional); sharedGroup != nil {
					sgb := sharedGroup.GetBalances(cd.Destination, cd.Category, cd.ToR, acc, cd.TimeStart)
					sgb = sharedGroup.SortBalancesByStrategy(mb, sgb)
					extendedMinuteBalances = append(extendedMinuteBalances, sgb...)
				}
			}
		} else {
			extendedMinuteBalances = append(extendedMinuteBalances, mb)
		}
	}
	credit = extendedCreditBalances.GetTotalValue()
	balances = extendedMinuteBalances
	for _, b := range balances {
		d, c := b.GetMinutesForCredit(cd, credit)
		credit = c
		duration += d
	}
	return
}

// sets all the fields of the balance
func (acc *Account) setBalanceAction(a *Action) error {
	if a == nil {
		return errors.New("nil action")
	}
	if acc.BalanceMap == nil {
		acc.BalanceMap = make(map[string]Balances)
	}
	var balance *Balance
	var found bool
	var previousSharedGroups utils.StringMap            // kept for comparison
	if a.Balance.Uuid != nil && *a.Balance.Uuid != "" { // balance uuid match
		for balanceType := range acc.BalanceMap {
			for _, b := range acc.BalanceMap[balanceType] {
				if b.Uuid == *a.Balance.Uuid && !b.IsExpiredAt(time.Now()) {
					previousSharedGroups = b.SharedGroups
					balance = b
					found = true
					break // only set one balance
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("cannot find balance with uuid: <%s>", *a.Balance.Uuid)
		}
	} else { // balance id match
		for balanceType := range acc.BalanceMap {
			for _, b := range acc.BalanceMap[balanceType] {
				if a.Balance.ID != nil && b.ID == *a.Balance.ID && !b.IsExpiredAt(time.Now()) {
					previousSharedGroups = b.SharedGroups
					balance = b
					found = true
					break // only set one balance
				}
			}
			if found {
				break
			}
		}
		// if it is not found then we create it
		if !found {
			if a.Balance.Type == nil { // cannot create the entry in the balance map without this info
				return errors.New("missing balance type")
			}
			balance = &Balance{}
			balance.Uuid = utils.GenUUID() // alway overwrite the uuid for consistency
			acc.BalanceMap[*a.Balance.Type] = append(acc.BalanceMap[*a.Balance.Type], balance)
		}
	}
	if a.Balance.ID != nil && *a.Balance.ID == utils.MetaDefault { // treat it separately since modifyBalance sets expiry and others parameters, not specific for *default
		if a.Balance.Value != nil {
			balance.ID = *a.Balance.ID
			balance.Value = a.Balance.GetValue()
			balance.SetDirty() // Mark the balance as dirty since we have modified and it should be checked by action triggers
		}
	} else {
		a.Balance.ModifyBalance(balance)
	}
	// modify if necessary the shared groups here
	if !found || !previousSharedGroups.Equal(balance.SharedGroups) {
		_, err := guardian.Guardian.Guard(func() (interface{}, error) {
			i := 0
			for sgID := range balance.SharedGroups {
				// add shared group member
				sg, err := dm.GetSharedGroup(sgID, false, utils.NonTransactional)
				if err != nil || sg == nil {
					//than is problem
					utils.Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgID))
				} else {
					if _, found := sg.MemberIds[acc.ID]; !found {
						// add member and save
						if sg.MemberIds == nil {
							sg.MemberIds = make(utils.StringMap)
						}
						sg.MemberIds[acc.ID] = true
						dm.SetSharedGroup(sg, utils.NonTransactional)
					}
				}
				i++
			}
			return 0, nil
		}, config.CgrConfig().GeneralCfg().LockingTimeout, balance.SharedGroups.Slice()...)
		if err != nil {
			return err
		}
	}
	acc.InitCounters()
	acc.ExecuteActionTriggers(nil)
	return nil
}

// Debits some amount of user's specified balance adding the balance if it does not exists.
// Returns the remaining credit in user's balance.
func (acc *Account) debitBalanceAction(a *Action, reset, resetIfNegative bool) error {
	if a == nil {
		return errors.New("nil action")
	}
	bClone := a.Balance.CreateBalance()
	//log.Print("Bclone: ", utils.ToJSON(a.Balance))
	if bClone == nil {
		return errors.New("nil balance in action")
	}
	if acc.BalanceMap == nil {
		acc.BalanceMap = make(map[string]Balances)
	}
	found := false
	balanceType := a.Balance.GetType()
	for _, b := range acc.BalanceMap[balanceType] {
		if b.IsExpiredAt(time.Now()) {
			continue // just to be safe (cleaned expired balances above)
		}
		b.account = acc
		if b.MatchFilter(a.Balance, false, false) {
			if reset || (resetIfNegative && b.Value < 0) {
				b.SetValue(0)
			}
			b.SubstractValue(bClone.GetValue())
			b.dirty = true
			found = true
			a.balanceValue = b.GetValue()
		}
	}
	// if it is not found then we add it to the list
	if !found {
		// check if the Id is *default (user trying to create the default balance)
		// use only it's value value
		if bClone.ID == utils.MetaDefault {
			bClone = &Balance{
				ID:    utils.MetaDefault,
				Value: -bClone.GetValue(),
			}
		} else {
			if bClone.GetValue() != 0 {
				bClone.SetValue(-bClone.GetValue())
			}
		}
		bClone.dirty = true // Mark the balance as dirty since we have modified and it should be checked by action triggers
		a.balanceValue = bClone.GetValue()
		bClone.Uuid = utils.GenUUID() // alway overwrite the uuid for consistency
		// load ValueFactor if defined in extra parametrs
		if a.ExtraParameters != "" {
			vf := ValueFactor{}
			err := json.Unmarshal([]byte(a.ExtraParameters), &vf)
			if err == nil {
				bClone.Factor = vf
			} else {
				utils.Logger.Warning(fmt.Sprintf("Could load value factor from actions: extra parametrs: %s", a.ExtraParameters))
			}
		}
		acc.BalanceMap[balanceType] = append(acc.BalanceMap[balanceType], bClone)
		_, err := guardian.Guardian.Guard(func() (interface{}, error) {
			sgs := make([]string, len(bClone.SharedGroups))
			i := 0
			for sgID := range bClone.SharedGroups {
				// add shared group member
				sg, err := dm.GetSharedGroup(sgID, false, utils.NonTransactional)
				if err != nil || sg == nil {
					//than is problem
					utils.Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgID))
				} else {
					if _, found := sg.MemberIds[acc.ID]; !found {
						// add member and save
						if sg.MemberIds == nil {
							sg.MemberIds = make(utils.StringMap)
						}
						sg.MemberIds[acc.ID] = true
						dm.SetSharedGroup(sg, utils.NonTransactional)
					}
				}
				i++
			}
			dm.CacheDataFromDB(utils.SHARED_GROUP_PREFIX, sgs, true)
			return 0, nil
		}, config.CgrConfig().GeneralCfg().LockingTimeout, bClone.SharedGroups.Slice()...)
		if err != nil {
			return err
		}
	}
	acc.InitCounters()
	acc.ExecuteActionTriggers(nil)
	return nil
}

func (acc *Account) getBalancesForPrefix(prefix, category, tor,
	sharedGroup string, aTime time.Time) Balances {
	var balances Balances
	balances = append(balances, acc.BalanceMap[tor]...)
	if tor != utils.MONETARY && tor != utils.GENERIC {
		balances = append(balances, acc.BalanceMap[utils.GENERIC]...)
	}

	var usefulBalances Balances
	for _, b := range balances {
		if b.Disabled {
			continue
		}
		if b.IsExpiredAt(aTime) || (len(b.SharedGroups) == 0 && b.GetValue() <= 0 && !b.Blocker) {
			continue
		}
		if sharedGroup != "" && b.SharedGroups[sharedGroup] == false {
			continue
		}
		if !b.MatchCategory(category) {
			continue
		}
		b.account = acc

		if len(b.DestinationIDs) > 0 && b.DestinationIDs[utils.ANY] == false {
			for _, p := range utils.SplitPrefix(prefix, MIN_PREFIX_MATCH) {
				if destIDs, err := dm.GetReverseDestination(p, true, true, utils.NonTransactional); err == nil {
					foundResult := false
					allInclude := true // whether it is excluded or included
					for _, dID := range destIDs {
						inclDest, found := b.DestinationIDs[dID]
						if found {
							foundResult = true
							allInclude = allInclude && inclDest
						}
					}
					// check wheter all destination ids in the balance were exclusions
					allExclude := true
					for _, inclDest := range b.DestinationIDs {
						if inclDest {
							allExclude = false
							break
						}
					}
					if foundResult || allExclude {
						if allInclude {
							b.precision = len(p)
							usefulBalances = append(usefulBalances, b)
						} else {
							b.precision = 1 // fake to exit the outer loop
						}
					}
				}
				if b.precision > 0 {
					break
				}
			}
		} else {
			usefulBalances = append(usefulBalances, b)
		}
	}

	// resort by precision
	usefulBalances.Sort()
	// clear precision
	for _, b := range usefulBalances {
		b.precision = 0
	}
	return usefulBalances
}

// like getBalancesForPrefix but expanding shared balances
func (acc *Account) getAlldBalancesForPrefix(destination, category,
	balanceType string, aTime time.Time) (bc Balances) {
	balances := acc.getBalancesForPrefix(destination, category, balanceType, "", aTime)
	for _, b := range balances {
		if len(b.SharedGroups) > 0 {
			for sgID := range b.SharedGroups {
				sharedGroup, err := dm.GetSharedGroup(sgID, false, utils.NonTransactional)
				if err != nil || sharedGroup == nil {
					utils.Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgID))
					continue
				}
				sharedBalances := sharedGroup.GetBalances(destination, category, balanceType, acc, aTime)
				sharedBalances = sharedGroup.SortBalancesByStrategy(b, sharedBalances)
				bc = append(bc, sharedBalances...)
			}
		} else {
			bc = append(bc, b)
		}
	}
	return
}

func (acc *Account) debitCreditBalance(cd *CallDescriptor, count bool, dryRun bool, goNegative bool) (cc *CallCost, err error) {
	usefulUnitBalances := acc.getAlldBalancesForPrefix(cd.Destination, cd.Category, cd.ToR, cd.TimeStart)
	usefulMoneyBalances := acc.getAlldBalancesForPrefix(cd.Destination, cd.Category, utils.MONETARY, cd.TimeStart)
	var leftCC *CallCost
	cc = cd.CreateCallCost()
	var hadBalanceSubj bool
	generalBalanceChecker := true
	for generalBalanceChecker {
		generalBalanceChecker = false

		// debit minutes
		unitBalanceChecker := true
		for unitBalanceChecker {
			// try every balance multiple times in case one becomes active or ratig changes
			unitBalanceChecker = false
			for _, balance := range usefulUnitBalances {
				partCC, debitErr := balance.debitUnits(cd, balance.account,
					usefulMoneyBalances, count, dryRun, len(cc.Timespans) == 0)
				if debitErr != nil {
					return nil, debitErr
				}
				if balance.RatingSubject != "" &&
					!strings.HasPrefix(balance.RatingSubject, utils.ZERO_RATING_SUBJECT_PREFIX) {
					hadBalanceSubj = true
				}
				if partCC != nil {
					cc.Timespans = append(cc.Timespans, partCC.Timespans...)
					cc.negativeConnectFee = partCC.negativeConnectFee
					cd.TimeStart = cc.GetEndTime()
					// check if the calldescriptor is covered
					if cd.GetDuration() <= 0 {
						goto COMMIT
					}
					unitBalanceChecker = true
					generalBalanceChecker = true
					// check for max cost disconnect
					if dryRun && partCC.maxCostDisconect {
						// only return if we are in dry run (max call duration)
						return
					}
				}
				// check for blocker
				if dryRun && balance.Blocker {
					return // don't go to next balances
				}
			}
		}
		// debit money
		moneyBalanceChecker := true
		for moneyBalanceChecker {
			// try every balance multiple times in case one becomes active or ratig changes
			moneyBalanceChecker = false
			for _, balance := range usefulMoneyBalances {
				partCC, debitErr := balance.debitMoney(cd, balance.account,
					usefulMoneyBalances, count, dryRun, len(cc.Timespans) == 0)
				if debitErr != nil {
					return nil, debitErr
				}
				if partCC != nil {
					cc.Timespans = append(cc.Timespans, partCC.Timespans...)
					cc.negativeConnectFee = partCC.negativeConnectFee

					cd.TimeStart = cc.GetEndTime()
					// check if the calldescriptor is covered
					if cd.GetDuration() <= 0 {
						goto COMMIT
					}
					moneyBalanceChecker = true
					generalBalanceChecker = true
					if dryRun && partCC.maxCostDisconect {
						// only return if we are in dry run (max call duration)
						return
					}
				}
				// check for blocker
				if dryRun && balance.Blocker {
					return // don't go to next balances
				}
			}
		}
	}
	if hadBalanceSubj {
		cd.RatingInfos = nil
	}
	leftCC, err = cd.getCost()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
	}
	if leftCC.Cost == 0 && len(leftCC.Timespans) > 0 {
		// put AccountID ubformation in increments
		for _, ts := range leftCC.Timespans {
			for _, inc := range ts.Increments {
				if inc.BalanceInfo == nil {
					inc.BalanceInfo = &DebitInfo{}
				}
				inc.BalanceInfo.AccountID = acc.ID
			}
		}
		cc.Timespans = append(cc.Timespans, leftCC.Timespans...)
	}

	if leftCC.Cost > 0 && goNegative {
		initialLength := len(cc.Timespans)
		cc.Timespans = append(cc.Timespans, leftCC.Timespans...)

		var debitedConnectFeeBalance Balance
		var ok bool

		if initialLength == 0 {
			// this is the first add, debit the connect fee
			ok, debitedConnectFeeBalance = acc.DebitConnectionFee(cc, usefulMoneyBalances, count, true)
		}
		// get the default money balance
		// and go negative on it with the amount still unpaid
		if len(leftCC.Timespans) > 0 && leftCC.Cost > 0 && !acc.AllowNegative && !dryRun {
			utils.Logger.Warning(fmt.Sprintf("<Rater> Going negative on account %s with AllowNegative: false", cd.GetAccountKey()))
		}
		leftCC.Timespans.Decompress()
		for tsIndex, ts := range leftCC.Timespans {
			if ts.Increments == nil {
				ts.createIncrementsSlice()
			}

			if tsIndex == 0 && ts.RateInterval.Rating.ConnectFee > 0 && cc.deductConnectFee && ok {

				inc := &Increment{
					Duration: 0,
					Cost:     ts.RateInterval.Rating.ConnectFee,
					BalanceInfo: &DebitInfo{
						Monetary: &MonetaryInfo{
							UUID:  debitedConnectFeeBalance.Uuid,
							ID:    debitedConnectFeeBalance.ID,
							Value: debitedConnectFeeBalance.Value,
						},
						AccountID: acc.ID,
					},
				}

				incs := []*Increment{inc}
				ts.Increments = append(incs, ts.Increments...)
			}

			for incIndex, increment := range ts.Increments {
				// connect fee was processed and skip it
				if tsIndex == 0 && incIndex == 0 && ts.RateInterval.Rating.ConnectFee > 0 && cc.deductConnectFee && ok {
					continue
				}
				cost := increment.Cost
				defaultBalance := acc.GetDefaultMoneyBalance()
				defaultBalance.SubstractValue(cost)

				increment.BalanceInfo.Monetary = &MonetaryInfo{
					UUID:  defaultBalance.Uuid,
					ID:    defaultBalance.ID,
					Value: defaultBalance.Value,
				}
				increment.BalanceInfo.AccountID = acc.ID
				increment.paid = true
				if count {
					acc.countUnits(
						cost,
						utils.MONETARY,
						leftCC,
						&Balance{
							Value:          cost,
							DestinationIDs: utils.NewStringMap(leftCC.Destination),
						})
				}
			}
		}

		// in case of going to negative we send the default balance to thresholdS to be processed
		if len(config.CgrConfig().RalsCfg().ThresholdSConns) != 0 {
			defaultBalance := acc.GetDefaultMoneyBalance()
			acntTnt := utils.NewTenantID(acc.ID)
			thEv := &ThresholdsArgsProcessEvent{
				CGREventWithOpts: &utils.CGREventWithOpts{
					CGREvent: &utils.CGREvent{
						Tenant: acntTnt.Tenant,
						ID:     utils.GenUUID(),
						Event: map[string]interface{}{
							utils.EventType:   utils.BalanceUpdate,
							utils.EventSource: utils.AccountService,
							utils.Account:     acntTnt.ID,
							utils.BalanceID:   defaultBalance.ID,
							utils.Units:       defaultBalance.Value,
						},
					},
					Opts: map[string]interface{}{
						utils.MetaEventType: utils.BalanceUpdate,
					},
				},
			}
			var tIDs []string
			if err := connMgr.Call(config.CgrConfig().RalsCfg().ThresholdSConns, nil,
				utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<AccountS> error: <%s> processing balance event <%+v> with ThresholdS.",
						err.Error(), utils.ToJSON(thEv)))
			}
		}
	}

COMMIT:
	if !dryRun {
		// save darty shared balances
		usefulMoneyBalances.SaveDirtyBalances(acc)
		usefulUnitBalances.SaveDirtyBalances(acc)
	}
	//log.Printf("Final CC: %+v", cc)
	return
}

// GetDefaultMoneyBalance returns the defaultmoney balance
func (acc *Account) GetDefaultMoneyBalance() *Balance {
	for _, balance := range acc.BalanceMap[utils.MONETARY] {
		if balance.IsDefault() {
			return balance
		}
	}
	// create default balance
	defaultBalance := &Balance{
		Uuid: utils.GenUUID(),
		ID:   utils.MetaDefault,
	} // minimum weight
	if acc.BalanceMap == nil {
		acc.BalanceMap = make(map[string]Balances)
	}
	acc.BalanceMap[utils.MONETARY] = append(acc.BalanceMap[utils.MONETARY], defaultBalance)
	return defaultBalance
}

// ExecuteActionTriggers scans the action triggers and execute the actions for which trigger is met
func (acc *Account) ExecuteActionTriggers(a *Action) {
	if acc.executingTriggers {
		return
	}
	acc.executingTriggers = true
	defer func() {
		acc.executingTriggers = false
	}()

	acc.ActionTriggers.Sort()
	for _, at := range acc.ActionTriggers {
		// check is effective
		if at.IsExpired(time.Now()) || !at.IsActive(time.Now()) {
			continue
		}

		// sanity check
		if !strings.Contains(at.ThresholdType, "counter") && !strings.Contains(at.ThresholdType, "balance") {
			continue
		}
		if at.Executed {
			// trigger is marked as executed, so skipp it until
			// the next reset (see RESET_TRIGGERS action type)
			continue
		}
		if !at.Match(a) {
			continue
		}
		if strings.Contains(at.ThresholdType, "counter") {
			if (at.Balance.ID == nil || *at.Balance.ID != "") && at.UniqueID != "" {
				at.Balance.ID = utils.StringPointer(at.UniqueID)
			}
			for _, counters := range acc.UnitCounters {
				for _, uc := range counters {
					if strings.Contains(at.ThresholdType, uc.CounterType[1:]) {
						for _, c := range uc.Counters {
							//log.Print("C: ", utils.ToJSON(c))
							if strings.HasPrefix(at.ThresholdType, "*max") {
								if c.Filter.Equal(at.Balance) && c.Value >= at.ThresholdValue {
									//log.Print("HERE")
									at.Execute(acc)
								}
							} else { //MIN
								if c.Filter.Equal(at.Balance) && c.Value <= at.ThresholdValue {
									at.Execute(acc)
								}
							}
						}
					}
				}
			}
		} else { // BALANCE
			for _, b := range acc.BalanceMap[at.Balance.GetType()] {
				if !b.dirty && at.ThresholdType != utils.TRIGGER_BALANCE_EXPIRED { // do not check clean balances
					continue
				}
				switch at.ThresholdType {
				case utils.TRIGGER_MAX_BALANCE:
					if b.MatchActionTrigger(at) && b.GetValue() >= at.ThresholdValue {
						at.Execute(acc)
					}
				case utils.TRIGGER_MIN_BALANCE:
					if b.MatchActionTrigger(at) && b.GetValue() <= at.ThresholdValue {
						at.Execute(acc)
					}
				case utils.TRIGGER_BALANCE_EXPIRED:
					if b.MatchActionTrigger(at) && b.IsExpiredAt(time.Now()) {
						at.Execute(acc)
					}
				}
			}
		}
	}
	acc.CleanExpiredStuff()
}

// ResetActionTriggers marks all action trigers as ready for execution
// If the action is not nil it acts like a filter
func (acc *Account) ResetActionTriggers(a *Action) {
	for _, at := range acc.ActionTriggers {
		if !at.Match(a) {
			continue
		}
		at.Executed = false
	}
	acc.ExecuteActionTriggers(a)
}

// SetRecurrent sets/unsets recurrent flag for action triggers
func (acc *Account) SetRecurrent(a *Action, recurrent bool) {
	for _, at := range acc.ActionTriggers {
		if !at.Match(a) {
			continue
		}
		at.Recurrent = recurrent
	}
}

// Increments the counter for the type
func (acc *Account) countUnits(amount float64, kind string, cc *CallCost, b *Balance) {
	acc.UnitCounters.addUnits(amount, kind, cc, b)
	acc.ExecuteActionTriggers(nil)
}

// InitCounters creates counters for all triggered actions
func (acc *Account) InitCounters() {
	oldUcs := acc.UnitCounters
	acc.UnitCounters = make(UnitCounters)
	ucTempMap := make(map[string]*UnitCounter)
	for _, at := range acc.ActionTriggers {
		//log.Print("AT: ", utils.ToJSON(at))
		if !strings.Contains(at.ThresholdType, "counter") {
			continue
		}
		ct := utils.COUNTER_EVENT //default
		if strings.Contains(at.ThresholdType, "balance") {
			ct = utils.COUNTER_BALANCE
		}
		uc, exists := ucTempMap[at.Balance.GetType()+ct]
		//log.Print("CT: ", at.Balance.GetType()+ct)
		if !exists {
			uc = &UnitCounter{
				CounterType: ct,
			}
			ucTempMap[at.Balance.GetType()+ct] = uc
			uc.Counters = make(CounterFilters, 0)
			acc.UnitCounters[at.Balance.GetType()] = append(acc.UnitCounters[at.Balance.GetType()], uc)
		}

		c := &CounterFilter{Filter: at.Balance.Clone()}
		if (c.Filter.ID == nil || *c.Filter.ID == "") && at.UniqueID != "" {
			c.Filter.ID = utils.StringPointer(at.UniqueID)
		}
		//log.Print("C: ", utils.ToJSON(c))
		if !uc.Counters.HasCounter(c) {
			uc.Counters = append(uc.Counters, c)
		}
	}
	// copy old counter values
	for key, counters := range acc.UnitCounters {
		oldCounters, found := oldUcs[key]
		if !found {
			continue
		}
		for _, uc := range counters {
			for _, oldUc := range oldCounters {
				if uc.CopyCounterValues(oldUc) {
					break
				}
			}
		}
	}
	if len(acc.UnitCounters) == 0 {
		acc.UnitCounters = nil // leave it nil if empty
	}
}

// CleanExpiredStuff removed expired balances and actiontriggers
func (acc *Account) CleanExpiredStuff() {
	if config.CgrConfig().RalsCfg().RemoveExpired {
		for key, bm := range acc.BalanceMap {
			for i := 0; i < len(bm); i++ {
				if bm[i].IsExpiredAt(time.Now()) {
					// delete it
					bm = append(bm[:i], bm[i+1:]...)
				}
			}
			acc.BalanceMap[key] = bm
		}
	}

	for i := 0; i < len(acc.ActionTriggers); i++ {
		if acc.ActionTriggers[i].IsExpired(time.Now()) {
			acc.ActionTriggers = append(acc.ActionTriggers[:i], acc.ActionTriggers[i+1:]...)
		}
	}
}

func (acc *Account) allBalancesExpired() bool {
	for _, bm := range acc.BalanceMap {
		for i := 0; i < len(bm); i++ {
			if !bm[i].IsExpiredAt(time.Now()) {
				return false
			}
		}
	}
	return true
}

// GetSharedGroups returns the shared groups that this user balance belnongs to
func (acc *Account) GetSharedGroups() (groups []string) {
	for _, balanceChain := range acc.BalanceMap {
		for _, b := range balanceChain {
			for sg := range b.SharedGroups {
				groups = append(groups, sg)
			}
		}
	}
	return
}

// GetUniqueSharedGroupMembers returns the acounts from the group
func (acc *Account) GetUniqueSharedGroupMembers(cd *CallDescriptor) (utils.StringMap, error) { // ToDo: make sure we return accountIDs
	var balances []*Balance
	balances = append(balances, acc.getBalancesForPrefix(cd.Destination, cd.Category, utils.MONETARY, "", cd.TimeStart)...)
	balances = append(balances, acc.getBalancesForPrefix(cd.Destination, cd.Category, cd.ToR, "", cd.TimeStart)...)
	// gather all shared group ids
	var sharedGroupIds []string
	for _, b := range balances {
		for sg := range b.SharedGroups {
			sharedGroupIds = append(sharedGroupIds, sg)
		}
	}
	memberIds := make(utils.StringMap)
	for _, sgID := range sharedGroupIds {
		sharedGroup, err := dm.GetSharedGroup(sgID, false, utils.NonTransactional)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgID))
			return nil, err
		}
		for memberID := range sharedGroup.MemberIds {
			memberIds[memberID] = true
		}
	}
	return memberIds, nil
}

// Clone creates a copy of the account
func (acc *Account) Clone() *Account {
	newAcc := &Account{
		ID:            acc.ID,
		UnitCounters:  acc.UnitCounters.Clone(),
		AllowNegative: acc.AllowNegative,
		Disabled:      acc.Disabled,
	}
	if acc.BalanceMap != nil {
		newAcc.BalanceMap = make(map[string]Balances, len(acc.BalanceMap))
		for key, balanceChain := range acc.BalanceMap {
			newAcc.BalanceMap[key] = balanceChain.Clone()
		}
	}
	if acc.ActionTriggers != nil {
		newAcc.ActionTriggers = make(ActionTriggers, len(acc.ActionTriggers))
		for key, actionTrigger := range acc.ActionTriggers {
			newAcc.ActionTriggers[key] = actionTrigger.Clone()
		}
	}
	return newAcc
}

// DebitConnectionFee debits the connection fee
func (acc *Account) DebitConnectionFee(cc *CallCost, usefulMoneyBalances Balances, count bool, block bool) (bool, Balance) {
	var debitedBalance Balance

	if cc.deductConnectFee {
		connectFee := cc.GetConnectFee()
		//log.Print("CONNECT FEE: %f", connectFee)
		connectFeePaid := false
		for _, b := range usefulMoneyBalances {
			if b.GetValue() >= connectFee {
				b.SubstractValue(connectFee)
				// the conect fee is not refundable!
				if count {
					acc.countUnits(connectFee, utils.MONETARY, cc, b)
				}
				connectFeePaid = true
				debitedBalance = *b
				break
			}
			if b.Blocker && block { // stop here
				return false, debitedBalance
			}
		}
		// debit connect fee
		if connectFee > 0 && !connectFeePaid {
			cc.negativeConnectFee = true
			// there are no money for the connect fee; go negative
			b := acc.GetDefaultMoneyBalance()
			b.SubstractValue(connectFee)
			debitedBalance = *b
			// the conect fee is not refundable!
			if count {
				acc.countUnits(connectFee, utils.MONETARY, cc, b)
			}
		}
	}
	return true, debitedBalance
}

func (acc *Account) matchActionFilter(condition string) (bool, error) {
	sm, err := structmatcher.NewStructMatcher(condition)
	if err != nil {
		return false, err
	}
	for balanceType, balanceChain := range acc.BalanceMap {
		for _, b := range balanceChain {
			check, err := sm.Match(&struct {
				Type string
				*Balance
			}{
				Type:    balanceType,
				Balance: b,
			})
			if err != nil {
				return false, err
			}
			if check {
				return true, nil
			}
		}
	}
	return false, nil
}

// GetID returns the account ID
func (acc *Account) GetID() string {
	split := strings.Split(acc.ID, utils.CONCATENATED_KEY_SEP)
	if len(split) != 2 {
		return ""
	}
	return split[1]
}

// AsOldStructure used in some api for transition
func (acc *Account) AsOldStructure() interface{} {
	type Balance struct {
		Uuid           string //system wide unique
		Id             string // account wide unique
		Value          float64
		ExpirationDate time.Time
		Weight         float64
		DestinationIds string
		RatingSubject  string
		Category       string
		SharedGroup    string
		Timings        []*RITiming
		TimingIDs      string
		Disabled       bool
		precision      int
		account        *Account
		dirty          bool
	}
	type Balances []*Balance
	type UnitsCounter struct {
		BalanceType string
		//	Units     float64
		Balances Balances // first balance is the general one (no destination)
	}
	type ActionTrigger struct {
		Id                    string
		ThresholdType         string
		ThresholdValue        float64
		Recurrent             bool
		MinSleep              time.Duration
		BalanceId             string
		BalanceType           string
		BalanceDestinationIds string
		BalanceWeight         float64
		BalanceExpirationDate time.Time
		BalanceTimingTags     string
		BalanceRatingSubject  string
		BalanceCategory       string
		BalanceSharedGroup    string
		BalanceDisabled       bool
		Weight                float64
		ActionsId             string
		MinQueuedItems        int
		Executed              bool
	}
	type ActionTriggers []*ActionTrigger
	type Account struct {
		Id             string
		BalanceMap     map[string]Balances
		UnitCounters   []*UnitsCounter
		ActionTriggers ActionTriggers
		AllowNegative  bool
		Disabled       bool
	}

	result := &Account{
		Id:             utils.META_OUT + utils.CONCATENATED_KEY_SEP + acc.ID,
		BalanceMap:     make(map[string]Balances, len(acc.BalanceMap)),
		UnitCounters:   make([]*UnitsCounter, len(acc.UnitCounters)),
		ActionTriggers: make(ActionTriggers, len(acc.ActionTriggers)),
		AllowNegative:  acc.AllowNegative,
		Disabled:       acc.Disabled,
	}
	for balanceType, counters := range acc.UnitCounters {
		for i, uc := range counters {
			if uc == nil {
				continue
			}
			result.UnitCounters[i] = &UnitsCounter{
				BalanceType: balanceType,
				Balances:    make(Balances, len(uc.Counters)),
			}
			if len(uc.Counters) > 0 {
				for j, c := range uc.Counters {
					result.UnitCounters[i].Balances[j] = &Balance{
						Uuid:           c.Filter.GetUuid(),
						Id:             c.Filter.GetID(),
						Value:          c.Filter.GetValue(),
						ExpirationDate: c.Filter.GetExpirationDate(),
						Weight:         c.Filter.GetWeight(),
						DestinationIds: c.Filter.GetDestinationIDs().String(),
						RatingSubject:  c.Filter.GetRatingSubject(),
						Category:       c.Filter.GetCategories().String(),
						SharedGroup:    c.Filter.GetSharedGroups().String(),
						Timings:        c.Filter.Timings,
						TimingIDs:      c.Filter.GetTimingIDs().String(),
						Disabled:       c.Filter.GetDisabled(),
					}
				}
			}
		}
	}
	for i, at := range acc.ActionTriggers {
		b := at.Balance.CreateBalance()
		result.ActionTriggers[i] = &ActionTrigger{
			Id:                    at.ID,
			ThresholdType:         at.ThresholdType,
			ThresholdValue:        at.ThresholdValue,
			Recurrent:             at.Recurrent,
			MinSleep:              at.MinSleep,
			BalanceType:           at.Balance.GetType(),
			BalanceId:             b.ID,
			BalanceDestinationIds: b.DestinationIDs.String(),
			BalanceWeight:         b.Weight,
			BalanceExpirationDate: b.ExpirationDate,
			BalanceTimingTags:     b.TimingIDs.String(),
			BalanceRatingSubject:  b.RatingSubject,
			BalanceCategory:       b.Categories.String(),
			BalanceSharedGroup:    b.SharedGroups.String(),
			BalanceDisabled:       b.Disabled,
			Weight:                at.Weight,
			ActionsId:             at.ActionsID,
			MinQueuedItems:        at.MinQueuedItems,
			Executed:              at.Executed,
		}
	}
	for key, values := range acc.BalanceMap {
		if len(values) > 0 {
			key += utils.META_OUT
			result.BalanceMap[key] = make(Balances, len(values))
			for i, b := range values {
				result.BalanceMap[key][i] = &Balance{
					Uuid:           b.Uuid,
					Id:             b.ID,
					Value:          b.Value,
					ExpirationDate: b.ExpirationDate,
					Weight:         b.Weight,
					DestinationIds: b.DestinationIDs.String(),
					RatingSubject:  b.RatingSubject,
					Category:       b.Categories.String(),
					SharedGroup:    b.SharedGroups.String(),
					Timings:        b.Timings,
					TimingIDs:      b.TimingIDs.String(),
					Disabled:       b.Disabled,
				}
			}
		}
	}
	return result
}

// AsAccountSummary converts the account into AccountSummary
func (acc *Account) AsAccountSummary() *AccountSummary {
	idSplt := strings.Split(acc.ID, utils.CONCATENATED_KEY_SEP)
	ad := &AccountSummary{AllowNegative: acc.AllowNegative, Disabled: acc.Disabled}
	if len(idSplt) == 1 {
		ad.ID = idSplt[0]
	} else if len(idSplt) == 2 {
		ad.Tenant = idSplt[0]
		ad.ID = idSplt[1]
	}

	for _, balanceType := range []string{utils.DATA, utils.SMS, utils.MMS, utils.VOICE, utils.GENERIC, utils.MONETARY} {
		balances, has := acc.BalanceMap[balanceType]
		if !has {
			continue
		}
		for _, balance := range balances {
			ad.BalanceSummaries = append(ad.BalanceSummaries, balance.AsBalanceSummary(balanceType))
		}
	}
	return ad
}

// Publish sends the account to stats and threshold
func (acc *Account) Publish() {
	acntSummary := acc.AsAccountSummary()
	cgrEv := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: acntSummary.Tenant,
			ID:     utils.GenUUID(),
			Time:   utils.TimePointer(time.Now()),
			Event:  acntSummary.AsMapInterface(),
		},
		Opts: map[string]interface{}{
			utils.MetaEventType: utils.AccountUpdate,
		},
	}
	if len(config.CgrConfig().RalsCfg().ThresholdSConns) != 0 {
		var tIDs []string
		if err := connMgr.Call(config.CgrConfig().RalsCfg().ThresholdSConns, nil,
			utils.ThresholdSv1ProcessEvent, &ThresholdsArgsProcessEvent{
				CGREventWithOpts: cgrEv,
			}, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<AccountS> error: %s processing account event %+v with ThresholdS.", err.Error(), cgrEv))
		}
	}
	if len(config.CgrConfig().RalsCfg().StatSConns) != 0 {
		var stsIDs []string
		if err := connMgr.Call(config.CgrConfig().RalsCfg().StatSConns, nil,
			utils.StatSv1ProcessEvent, &StatsArgsProcessEvent{
				CGREventWithOpts: cgrEv,
			}, &stsIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<AccountS> error: %s processing account event %+v with StatS.", err.Error(), cgrEv))
		}
	}

}

// NewAccountSummaryFromJSON creates a new AcccountSummary from a json string
func NewAccountSummaryFromJSON(jsn string) (acntSummary *AccountSummary, err error) {
	if !utils.SliceHasMember([]string{"", "null"}, jsn) { // Unmarshal only when content
		err = json.Unmarshal([]byte(jsn), &acntSummary)
	}
	return
}

// AccountSummary contains compressed information about an Account
type AccountSummary struct {
	Tenant           string
	ID               string
	BalanceSummaries BalanceSummaries
	AllowNegative    bool
	Disabled         bool
}

// Clone creates a copy of the structure
func (as *AccountSummary) Clone() (cln *AccountSummary) {
	cln = new(AccountSummary)
	cln.Tenant = as.Tenant
	cln.ID = as.ID
	cln.AllowNegative = as.AllowNegative
	cln.Disabled = as.Disabled
	if as.BalanceSummaries != nil {
		cln.BalanceSummaries = make(BalanceSummaries, len(as.BalanceSummaries))
		for i, bs := range as.BalanceSummaries {
			cln.BalanceSummaries[i] = new(BalanceSummary)
			*cln.BalanceSummaries[i] = *bs
		}
	}
	return
}

// GetBalanceWithID returns a Balance given balance type and balance ID
func (acc *Account) GetBalanceWithID(blcType, blcID string) (blc *Balance) {
	for _, blc = range acc.BalanceMap[blcType] {
		if blc.ID == blcID {
			return
		}
	}
	return nil
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (as *AccountSummary) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		opath, indx := utils.GetPathIndex(fldPath[0])
		if opath == utils.BalanceSummaries && indx != nil {
			if len(as.BalanceSummaries) <= *indx {
				return nil, utils.ErrNotFound
			}
			bl := as.BalanceSummaries[*indx]
			if len(fldPath) == 1 {
				return bl, nil
			}
			return bl.FieldAsInterface(fldPath[1:])
		}
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.Tenant:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return as.Tenant, nil
	case utils.ID:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return as.ID, nil
	case utils.BalanceSummaries:
		if len(fldPath) == 1 {
			return as.BalanceSummaries, nil
		}
		for _, bs := range as.BalanceSummaries {
			if bs.ID == fldPath[1] {
				if len(fldPath) == 2 {
					return bs, nil
				}
				return bs.FieldAsInterface(fldPath[2:])
			}
		}
		return nil, utils.ErrNotFound
	case utils.AllowNegative:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return as.AllowNegative, nil
	case utils.Disabled:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return as.Disabled, nil
	}
}

func (as *AccountSummary) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	iface, err = as.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

// String implements utils.DataProvider
func (as *AccountSummary) String() string {
	return utils.ToIJSON(as)
}

// RemoteHost implements utils.DataProvider
func (ar *AccountSummary) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

func (as *AccountSummary) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.Tenant:           as.Tenant,
		utils.ID:               as.ID,
		utils.AllowNegative:    as.AllowNegative,
		utils.Disabled:         as.Disabled,
		utils.BalanceSummaries: as.BalanceSummaries,
	}
}
