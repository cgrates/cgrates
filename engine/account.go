/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"

	"strings"
)

/*
Structure containing information about user's credit (minutes, cents, sms...).'
This can represent a user or a shared group.
*/
type Account struct {
	Id             string
	BalanceMap     map[string]BalanceChain
	UnitCounters   UnitCounters
	ActionTriggers ActionTriggers
	AllowNegative  bool
	Disabled       bool
}

// User's available minutes for the specified destination
func (ub *Account) getCreditForPrefix(cd *CallDescriptor) (duration time.Duration, credit float64, balances BalanceChain) {
	creditBalances := ub.getBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, utils.MONETARY, "")

	unitBalances := ub.getBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, cd.TOR, "")
	// gather all balances from shared groups
	var extendedCreditBalances BalanceChain
	for _, cb := range creditBalances {
		if len(cb.SharedGroups) > 0 {
			for sg := range cb.SharedGroups {
				if sharedGroup, _ := ratingStorage.GetSharedGroup(sg, false); sharedGroup != nil {
					sgb := sharedGroup.GetBalances(cd.Destination, cd.Category, cd.Direction, utils.MONETARY, ub)
					sgb = sharedGroup.SortBalancesByStrategy(cb, sgb)
					extendedCreditBalances = append(extendedCreditBalances, sgb...)
				}
			}
		} else {
			extendedCreditBalances = append(extendedCreditBalances, cb)
		}
	}
	var extendedMinuteBalances BalanceChain
	for _, mb := range unitBalances {
		if len(mb.SharedGroups) > 0 {
			for sg := range mb.SharedGroups {
				if sharedGroup, _ := ratingStorage.GetSharedGroup(sg, false); sharedGroup != nil {
					sgb := sharedGroup.GetBalances(cd.Destination, cd.Category, cd.Direction, cd.TOR, ub)
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

// Debits some amount of user's specified balance adding the balance if it does not exists.
// Returns the remaining credit in user's balance.
func (ub *Account) debitBalanceAction(a *Action, reset bool) error {
	if a == nil {
		return errors.New("nil action")
	}
	bClone := a.Balance.Clone()
	if bClone == nil {
		return errors.New("nil balance")
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain, 1)
	}
	found := false
	id := a.BalanceType
	for _, b := range ub.BalanceMap[id] {
		if b.IsExpired() {
			continue // just to be safe (cleaned expired balances above)
		}
		b.account = ub
		if b.MatchFilter(a.Balance, false) {
			if reset {
				b.SetValue(0)
			}
			b.SubstractValue(bClone.GetValue())
			found = true
		}
	}
	// if it is not found then we add it to the list
	if !found {
		// check if the Id is *default (user trying to create the default balance)
		// use only it's value value
		if bClone.Id == utils.META_DEFAULT {
			bClone = &Balance{
				Id:    utils.META_DEFAULT,
				Value: -bClone.GetValue(),
			}
		} else {
			if bClone.GetValue() != 0 {
				bClone.SetValue(-bClone.GetValue())
			}
		}
		bClone.dirty = true // Mark the balance as dirty since we have modified and it should be checked by action triggers
		if bClone.Uuid == "" {
			bClone.Uuid = utils.GenUUID()
		}
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
		ub.BalanceMap[id] = append(ub.BalanceMap[id], bClone)
	}
	for sgId := range a.Balance.SharedGroups {
		// add shared group member
		sg, err := ratingStorage.GetSharedGroup(sgId, false)
		if err != nil || sg == nil {
			//than is problem
			utils.Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgId))
		} else {
			if _, found := sg.MemberIds[ub.Id]; !found {
				// add member and save
				if sg.MemberIds == nil {
					sg.MemberIds = make(utils.StringMap)
				}
				sg.MemberIds[ub.Id] = true
				ratingStorage.SetSharedGroup(sg)
			}
		}
	}
	ub.executeActionTriggers(nil)
	return nil //ub.BalanceMap[id].GetTotalValue()
}

func (ub *Account) enableDisableBalanceAction(a *Action) error {
	if a == nil {
		return errors.New("nil action")
	}

	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain)
	}
	found := false
	id := a.BalanceType
	disabled := a.Balance.Disabled
	a.Balance.Disabled = !disabled // match for the opposite
	for _, b := range ub.BalanceMap[id] {
		if b.MatchFilter(a.Balance, false) {
			b.Disabled = disabled
			b.dirty = true
			found = true
		}
	}
	if !found {
		return utils.ErrNotFound
	}
	return nil
}

func (ub *Account) getBalancesForPrefix(prefix, category, direction, tor string, sharedGroup string) BalanceChain {
	var balances BalanceChain
	balances = append(balances, ub.BalanceMap[tor]...)
	if tor != utils.MONETARY && tor != utils.GENERIC {
		balances = append(balances, ub.BalanceMap[utils.GENERIC]...)
	}
	var usefulBalances BalanceChain
	for _, b := range balances {
		if b.Disabled {
			continue
		}
		if b.IsExpired() || (len(b.SharedGroups) == 0 && b.GetValue() <= 0) {
			continue
		}
		if sharedGroup != "" && b.SharedGroups[sharedGroup] == false {
			continue
		}
		if !b.MatchCategory(category) {
			continue
		}
		if b.HasDirection() && b.Directions[direction] == false {
			continue
		}
		b.account = ub
		if len(b.DestinationIds) > 0 && b.DestinationIds[utils.ANY] == false {
			for _, p := range utils.SplitPrefix(prefix, MIN_PREFIX_MATCH) {
				if x, err := cache2go.Get(utils.DESTINATION_PREFIX + p); err == nil {
					destIds := x.(map[interface{}]struct{})
					for dId, _ := range destIds {
						if b.DestinationIds[dId.(string)] == true {
							b.precision = len(p)
							usefulBalances = append(usefulBalances, b)
							break
						}
						if b.precision > 0 {
							break
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
func (account *Account) getAlldBalancesForPrefix(destination, category, direction, balanceType string) (bc BalanceChain) {
	balances := account.getBalancesForPrefix(destination, category, direction, balanceType, "")
	for _, b := range balances {
		if len(b.SharedGroups) > 0 {
			for sgId := range b.SharedGroups {
				sharedGroup, err := ratingStorage.GetSharedGroup(sgId, false)
				if err != nil {
					utils.Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgId))
					continue
				}
				sharedBalances := sharedGroup.GetBalances(destination, category, direction, balanceType, account)
				sharedBalances = sharedGroup.SortBalancesByStrategy(b, sharedBalances)
				bc = append(bc, sharedBalances...)
			}
		} else {
			bc = append(bc, b)
		}
	}
	return
}

func (ub *Account) debitCreditBalance(cd *CallDescriptor, count bool, dryRun bool, goNegative bool) (cc *CallCost, err error) {
	usefulUnitBalances := ub.getAlldBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, cd.TOR)
	usefulMoneyBalances := ub.getAlldBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, utils.MONETARY)
	//utils.Logger.Info(fmt.Sprintf("%+v, %+v", usefulMoneyBalances, usefulUnitBalances))
	//utils.Logger.Info(fmt.Sprintf("STARTCD: %+v", cd))
	var leftCC *CallCost
	cc = cd.CreateCallCost()

	generalBalanceChecker := true
	for generalBalanceChecker {
		generalBalanceChecker = false

		// debit minutes
		unitBalanceChecker := true
		for unitBalanceChecker {
			// try every balance multiple times in case one becomes active or ratig changes
			unitBalanceChecker = false
			//log.Printf("InitialCD: %+v", cd)
			for _, balance := range usefulUnitBalances {
				//utils.Logger.Info(fmt.Sprintf("Unit balance: %+v", balance))
				//utils.Logger.Info(fmt.Sprintf("CD BEFORE UNIT: %+v", cd))

				partCC, debitErr := balance.debitUnits(cd, balance.account, usefulMoneyBalances, count, dryRun, len(cc.Timespans) == 0)
				if debitErr != nil {
					return nil, debitErr
				}
				//utils.Logger.Info(fmt.Sprintf("CD AFTER UNIT: %+v", cd))
				if partCC != nil {
					//log.Printf("partCC: %+v", partCC.Timespans[0])
					cc.Timespans = append(cc.Timespans, partCC.Timespans...)
					cc.negativeConnectFee = partCC.negativeConnectFee
					// for i, ts := range cc.Timespans {
					//  log.Printf("cc.times[an[%d]: %+v\n", i, ts)
					// }
					cd.TimeStart = cc.GetEndTime()
					//log.Printf("CD: %+v", cd)
					//log.Printf("CD: %+v - %+v", cd.TimeStart, cd.TimeEnd)
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
					//log.Print("BLOCKER!")
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
				//utils.Logger.Info(fmt.Sprintf("Money balance: %+v", balance))
				//utils.Logger.Info(fmt.Sprintf("CD BEFORE MONEY: %+v", cd))
				partCC, debitErr := balance.debitMoney(cd, balance.account, usefulMoneyBalances, count, dryRun, len(cc.Timespans) == 0)
				if debitErr != nil {
					return nil, debitErr
				}
				//utils.Logger.Info(fmt.Sprintf("CD AFTER MONEY: %+v", cd))
				if partCC != nil {
					cc.Timespans = append(cc.Timespans, partCC.Timespans...)
					cc.negativeConnectFee = partCC.negativeConnectFee

					/*for i, ts := range cc.Timespans {
						log.Printf("cc.times[an[%d]: %+v\n", i, ts)
					}*/
					cd.TimeStart = cc.GetEndTime()
					//log.Printf("CD: %+v", cd)
					//log.Printf("CD: %+v - %+v", cd.TimeStart, cd.TimeEnd)
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
					//log.Print("BLOCKER!")
					return // don't go to next balances
				}
			}
		}
		//log.Printf("END CD: %+v", cd)
		//log.Print("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	}
	//log.Printf("After balances CD: %+v", cd)
	leftCC, err = cd.getCost()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
	}
	if leftCC.Cost == 0 && len(leftCC.Timespans) > 0 {
		cc.Timespans = append(cc.Timespans, leftCC.Timespans...)
	}

	//log.Printf("HERE: %+v", leftCC)
	if leftCC.Cost > 0 && goNegative {
		initialLength := len(cc.Timespans)
		cc.Timespans = append(cc.Timespans, leftCC.Timespans...)
		if initialLength == 0 {
			// this is the first add, debit the connect fee
			ub.DebitConnectionFee(cc, usefulMoneyBalances, count)
		}
		//log.Printf("Left CC: %+v ", leftCC)
		// get the default money balanance
		// and go negative on it with the amount still unpaid
		if len(leftCC.Timespans) > 0 && leftCC.Cost > 0 && !ub.AllowNegative && !dryRun {
			utils.Logger.Err(fmt.Sprintf("<Rater> Going negative on account %s with AllowNegative: false", cd.GetAccountKey()))
		}
		leftCC.Timespans.Decompress()
		for _, ts := range leftCC.Timespans {
			if ts.Increments == nil {
				ts.createIncrementsSlice()
			}
			for _, increment := range ts.Increments {
				cost := increment.Cost
				defaultBalance := ub.GetDefaultMoneyBalance()
				defaultBalance.SubstractValue(cost)
				increment.BalanceInfo.MoneyBalanceUuid = defaultBalance.Uuid
				increment.BalanceInfo.AccountId = ub.Id
				increment.paid = true
				if count {
					ub.countUnits(
						cost,
						utils.MONETARY,
						leftCC,
						&Balance{
							Directions:     utils.StringMap{leftCC.Direction: true},
							Value:          cost,
							DestinationIds: utils.NewStringMap(leftCC.Destination),
						})
				}
			}
		}
	}

COMMIT:
	if !dryRun {
		// save darty shared balances
		usefulMoneyBalances.SaveDirtyBalances(ub)
		usefulUnitBalances.SaveDirtyBalances(ub)
	}
	//log.Printf("Final CC: %+v", cc)
	return
}

func (ub *Account) GetDefaultMoneyBalance() *Balance {
	for _, balance := range ub.BalanceMap[utils.MONETARY] {
		if balance.IsDefault() {
			return balance
		}
	}
	// create default balance
	defaultBalance := &Balance{
		Uuid: utils.GenUUID(),
		Id:   utils.META_DEFAULT,
	} // minimum weight
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain)
	}
	ub.BalanceMap[utils.MONETARY] = append(ub.BalanceMap[utils.MONETARY], defaultBalance)
	return defaultBalance
}

func (ub *Account) refundIncrement(increment *Increment, cd *CallDescriptor, count bool) {
	var balance *Balance
	unitType := cd.TOR
	cc := cd.CreateCallCost()
	if increment.BalanceInfo.UnitBalanceUuid != "" {
		if balance = ub.BalanceMap[unitType].GetBalance(increment.BalanceInfo.UnitBalanceUuid); balance == nil {
			return
		}
		balance.AddValue(increment.Duration.Seconds())
		if count {
			ub.countUnits(-increment.Duration.Seconds(), unitType, cc, balance)
		}
	}
	// check money too
	if increment.BalanceInfo.MoneyBalanceUuid != "" {
		if balance = ub.BalanceMap[utils.MONETARY].GetBalance(increment.BalanceInfo.MoneyBalanceUuid); balance == nil {
			return
		}
		balance.AddValue(increment.Cost)
		if count {
			ub.countUnits(-increment.Cost, utils.MONETARY, cc, balance)
		}
	}
}

// Scans the action trigers and execute the actions for which trigger is met
func (ub *Account) executeActionTriggers(a *Action) {
	ub.ActionTriggers.Sort()
	for _, at := range ub.ActionTriggers {
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
			for _, uc := range ub.UnitCounters {
				if uc.BalanceType == at.BalanceType &&
					strings.Contains(at.ThresholdType, uc.CounterType[1:]) {
					for _, mb := range uc.Balances {
						if strings.HasPrefix(at.ThresholdType, "*max") {
							if mb.MatchActionTrigger(at) && mb.GetValue() >= at.ThresholdValue {
								at.Execute(ub, nil)
							}
						} else { //MIN
							if mb.MatchActionTrigger(at) && mb.GetValue() <= at.ThresholdValue {
								at.Execute(ub, nil)
							}
						}
					}
				}
			}
		} else { // BALANCE
			for _, b := range ub.BalanceMap[at.BalanceType] {
				if !b.dirty && at.ThresholdType != utils.TRIGGER_BALANCE_EXPIRED { // do not check clean balances
					continue
				}
				switch at.ThresholdType {
				case utils.TRIGGER_MAX_BALANCE:
					if b.MatchActionTrigger(at) && b.GetValue() >= at.ThresholdValue {
						at.Execute(ub, nil)
					}
				case utils.TRIGGER_MIN_BALANCE:
					if b.MatchActionTrigger(at) && b.GetValue() <= at.ThresholdValue {
						at.Execute(ub, nil)
					}
				case utils.TRIGGER_BALANCE_EXPIRED:
					if b.MatchActionTrigger(at) && b.IsExpired() {
						at.Execute(ub, nil)
					}
				}
			}
		}
	}
	ub.CleanExpiredBalances()
}

// Mark all action trigers as ready for execution
// If the action is not nil it acts like a filter
func (acc *Account) ResetActionTriggers(a *Action) {
	for _, at := range acc.ActionTriggers {
		if !at.Match(a) {
			continue
		}
		at.Executed = false
	}
	acc.executeActionTriggers(a)
}

// Sets/Unsets recurrent flag for action triggers
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
	acc.executeActionTriggers(nil)
}

// Create counters for all triggered actions
func (acc *Account) InitCounters() {
	acc.UnitCounters = nil
	ucTempMap := make(map[string]*UnitCounter)
	for _, at := range acc.ActionTriggers {
		if !strings.Contains(at.ThresholdType, "counter") {
			continue
		}
		ct := utils.COUNTER_EVENT //default
		if strings.Contains(at.ThresholdType, "balance") {
			ct = utils.COUNTER_BALANCE
		}

		uc, exists := ucTempMap[at.BalanceType+ct]
		if !exists {
			uc = &UnitCounter{
				BalanceType: at.BalanceType,
				CounterType: ct,
			}
			ucTempMap[at.BalanceType+ct] = uc
			uc.Balances = BalanceChain{}
			acc.UnitCounters = append(acc.UnitCounters, uc)
		}
		b := at.CreateBalance()
		if !uc.Balances.HasBalance(b) {
			uc.Balances = append(uc.Balances, b)
		}
	}
}

func (acc *Account) CleanExpiredBalances() {
	for key, bm := range acc.BalanceMap {
		for i := 0; i < len(bm); i++ {
			if bm[i].IsExpired() {
				// delete it
				bm = append(bm[:i], bm[i+1:]...)
			}
		}
		acc.BalanceMap[key] = bm
	}
}

func (acc *Account) allBalancesExpired() bool {
	for _, bm := range acc.BalanceMap {
		for i := 0; i < len(bm); i++ {
			if !bm[i].IsExpired() {
				return false
			}
		}
	}
	return true
}

// returns the shared groups that this user balance belnongs to
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

func (account *Account) GetUniqueSharedGroupMembers(cd *CallDescriptor) (utils.StringMap, error) {
	var balances []*Balance
	balances = append(balances, account.getBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, utils.MONETARY, "")...)
	balances = append(balances, account.getBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, cd.TOR, "")...)
	// gather all shared group ids
	var sharedGroupIds []string
	for _, b := range balances {
		for sg := range b.SharedGroups {
			sharedGroupIds = append(sharedGroupIds, sg)
		}
	}
	memberIds := make(utils.StringMap)
	for _, sgID := range sharedGroupIds {
		sharedGroup, err := ratingStorage.GetSharedGroup(sgID, false)
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

type TenantAccount struct {
	Tenant, Account string
}

func (acc *Account) Clone() *Account {
	newAcc := &Account{
		Id:             acc.Id,
		BalanceMap:     make(map[string]BalanceChain, len(acc.BalanceMap)),
		UnitCounters:   nil, // not used when cloned (dryRun)
		ActionTriggers: nil, // not used when cloned (dryRun)
		AllowNegative:  acc.AllowNegative,
		Disabled:       acc.Disabled,
	}
	for key, balanceChain := range acc.BalanceMap {
		newAcc.BalanceMap[key] = balanceChain.Clone()
	}
	return newAcc
}

func (acc *Account) DebitConnectionFee(cc *CallCost, usefulMoneyBalances BalanceChain, count bool) {
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
				break
			}
		}
		// debit connect fee
		if connectFee > 0 && !connectFeePaid {
			cc.negativeConnectFee = true
			// there are no money for the connect fee; go negative
			b := acc.GetDefaultMoneyBalance()
			b.SubstractValue(connectFee)
			// the conect fee is not refundable!
			if count {
				acc.countUnits(connectFee, utils.MONETARY, cc, b)
			}
		}
	}
}

func (acc *Account) matchConditions(condition string) (bool, error) {
	cl := &utils.CondLoader{}
	if err := cl.Parse(condition); err != nil {
		return false, err
	}
	for balanceType, balanceChain := range acc.BalanceMap {
		for _, b := range balanceChain {
			check, err := cl.Check(&struct {
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

// used in some api for transition
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
	type BalanceChain []*Balance
	type UnitsCounter struct {
		Direction   string
		BalanceType string
		//	Units     float64
		Balances BalanceChain // first balance is the general one (no destination)
	}
	type ActionTrigger struct {
		Id                    string
		ThresholdType         string
		ThresholdValue        float64
		Recurrent             bool
		MinSleep              time.Duration
		BalanceId             string
		BalanceType           string
		BalanceDirection      string
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
		BalanceMap     map[string]BalanceChain
		UnitCounters   []*UnitsCounter
		ActionTriggers ActionTriggers
		AllowNegative  bool
		Disabled       bool
	}

	result := &Account{
		Id:             utils.OUT + ":" + acc.Id,
		BalanceMap:     make(map[string]BalanceChain, len(acc.BalanceMap)),
		UnitCounters:   make([]*UnitsCounter, len(acc.UnitCounters)),
		ActionTriggers: make(ActionTriggers, len(acc.ActionTriggers)),
		AllowNegative:  acc.AllowNegative,
		Disabled:       acc.Disabled,
	}
	for i, uc := range acc.UnitCounters {
		if uc == nil {
			continue
		}
		result.UnitCounters[i] = &UnitsCounter{
			BalanceType: uc.BalanceType,
			Balances:    make(BalanceChain, len(uc.Balances)),
		}
		if len(uc.Balances) > 0 {
			result.UnitCounters[i].Direction = uc.Balances[0].Directions.String()
			for j, b := range uc.Balances {
				result.UnitCounters[i].Balances[j] = &Balance{
					Uuid:           b.Uuid,
					Id:             b.Id,
					Value:          b.Value,
					ExpirationDate: b.ExpirationDate,
					Weight:         b.Weight,
					DestinationIds: b.DestinationIds.String(),
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
	for i, at := range acc.ActionTriggers {
		result.ActionTriggers[i] = &ActionTrigger{
			Id:                    at.ID,
			ThresholdType:         at.ThresholdType,
			ThresholdValue:        at.ThresholdValue,
			Recurrent:             at.Recurrent,
			MinSleep:              at.MinSleep,
			BalanceId:             at.BalanceId,
			BalanceType:           at.BalanceType,
			BalanceDirection:      at.BalanceDirections.String(),
			BalanceDestinationIds: at.BalanceDestinationIds.String(),
			BalanceWeight:         at.BalanceWeight,
			BalanceExpirationDate: at.BalanceExpirationDate,
			BalanceTimingTags:     at.BalanceTimingTags.String(),
			BalanceRatingSubject:  at.BalanceRatingSubject,
			BalanceCategory:       at.BalanceCategories.String(),
			BalanceSharedGroup:    at.BalanceSharedGroups.String(),
			BalanceDisabled:       at.BalanceDisabled,
			Weight:                at.Weight,
			ActionsId:             at.ActionsId,
			MinQueuedItems:        at.MinQueuedItems,
			Executed:              at.Executed,
		}
	}
	for key, values := range acc.BalanceMap {
		if len(values) > 0 {
			key += utils.OUT
			result.BalanceMap[key] = make(BalanceChain, len(values))
			for i, b := range values {
				result.BalanceMap[key][i] = &Balance{
					Uuid:           b.Uuid,
					Id:             b.Id,
					Value:          b.Value,
					ExpirationDate: b.ExpirationDate,
					Weight:         b.Weight,
					DestinationIds: b.DestinationIds.String(),
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
