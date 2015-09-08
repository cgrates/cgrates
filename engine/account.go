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
	"errors"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"

	"strings"
)

const (
	// Direction type
	INBOUND  = "*in"
	OUTBOUND = "*out"
	// action trigger threshold types
	TRIGGER_MIN_COUNTER = "*min_counter"
	TRIGGER_MAX_COUNTER = "*max_counter"
	TRIGGER_MIN_BALANCE = "*min_balance"
	TRIGGER_MAX_BALANCE = "*max_balance"
)

/*
Structure containing information about user's credit (minutes, cents, sms...).'
This can represent a user or a shared group.
*/
type Account struct {
	Id             string
	BalanceMap     map[string]BalanceChain
	UnitCounters   []*UnitsCounter
	ActionTriggers ActionTriggerPriotityList
	AllowNegative  bool
	Disabled       bool
}

// User's available minutes for the specified destination
func (ub *Account) getCreditForPrefix(cd *CallDescriptor) (duration time.Duration, credit float64, balances BalanceChain) {
	creditBalances := ub.getBalancesForPrefix(cd.Destination, cd.Category, ub.BalanceMap[utils.MONETARY+cd.Direction], "")
	unitBalances := ub.getBalancesForPrefix(cd.Destination, cd.Category, ub.BalanceMap[cd.TOR+cd.Direction], "")
	// gather all balances from shared groups
	var extendedCreditBalances BalanceChain
	for _, cb := range creditBalances {
		if cb.SharedGroup != "" {
			if sharedGroup, _ := ratingStorage.GetSharedGroup(cb.SharedGroup, false); sharedGroup != nil {
				sgb := sharedGroup.GetBalances(cd.Destination, cd.Category, utils.MONETARY+cd.Direction, ub)
				sgb = sharedGroup.SortBalancesByStrategy(cb, sgb)
				extendedCreditBalances = append(extendedCreditBalances, sgb...)
			}
		} else {
			extendedCreditBalances = append(extendedCreditBalances, cb)
		}
	}
	var extendedMinuteBalances BalanceChain
	for _, mb := range unitBalances {
		if mb.SharedGroup != "" {
			if sharedGroup, _ := ratingStorage.GetSharedGroup(mb.SharedGroup, false); sharedGroup != nil {
				sgb := sharedGroup.GetBalances(cd.Destination, cd.Category, cd.TOR+cd.Direction, ub)
				sgb = sharedGroup.SortBalancesByStrategy(mb, sgb)
				extendedMinuteBalances = append(extendedMinuteBalances, sgb...)
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
		return errors.New("nil minute action")
	}
	if a.Balance.Uuid == "" {
		a.Balance.Uuid = utils.GenUUID()
	}
	bClone := a.Balance.Clone()

	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain, 1)
	}
	found := false
	id := a.BalanceType + a.Direction
	ub.CleanExpiredBalances()
	for _, b := range ub.BalanceMap[id] {
		if b.IsExpired() {
			continue // just to be safe (cleaned expired balances above)
		}
		if b.MatchFilter(a.Balance) {
			if reset {
				b.SetValue(0)
			}
			b.SubstractValue(bClone.GetValue())
			found = true
		}
	}
	// if it is not found then we add it to the list
	if !found {
		if bClone.GetValue() != 0 {
			bClone.SetValue(-bClone.GetValue())
		}
		bClone.dirty = true // Mark the balance as dirty since we have modified and it should be checked by action triggers
		ub.BalanceMap[id] = append(ub.BalanceMap[id], bClone)
	}
	if a.Balance.SharedGroup != "" {
		// add shared group member
		sg, err := ratingStorage.GetSharedGroup(a.Balance.SharedGroup, false)
		if err != nil || sg == nil {
			//than problem
			Logger.Warning(fmt.Sprintf("Could not get shared group: %v", a.Balance.SharedGroup))
		} else {
			if !utils.IsSliceMember(sg.MemberIds, ub.Id) {
				// add member and save
				sg.MemberIds = append(sg.MemberIds, ub.Id)
				ratingStorage.SetSharedGroup(sg)
			}
		}
	}
	ub.executeActionTriggers(nil)
	return nil //ub.BalanceMap[id].GetTotalValue()
}

func (ub *Account) getBalancesForPrefix(prefix, category string, balances BalanceChain, sharedGroup string) BalanceChain {
	var usefulBalances BalanceChain
	for _, b := range balances {
		if b.IsExpired() || (b.SharedGroup == "" && b.GetValue() <= 0) {
			continue
		}
		if sharedGroup != "" && b.SharedGroup != sharedGroup {
			continue
		}
		if !b.MatchCategory(category) {
			continue
		}
		b.account = ub
		if b.DestinationIds != "" && b.DestinationIds != utils.ANY {
			for _, p := range utils.SplitPrefix(prefix, MIN_PREFIX_MATCH) {
				if x, err := cache2go.Get(utils.DESTINATION_PREFIX + p); err == nil {
					destIds := x.(map[interface{}]struct{})
					for dId, _ := range destIds {
						balDestIds := strings.Split(b.DestinationIds, utils.INFIELD_SEP)
						for _, balDestID := range balDestIds {
							if dId == balDestID {
								b.precision = len(p)
								usefulBalances = append(usefulBalances, b)
								break
							}
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
func (account *Account) getAlldBalancesForPrefix(destination, category, balanceType string) (bc BalanceChain) {
	balances := account.getBalancesForPrefix(destination, category, account.BalanceMap[balanceType], "")
	for _, b := range balances {
		if b.SharedGroup != "" {
			sharedGroup, err := ratingStorage.GetSharedGroup(b.SharedGroup, false)
			if err != nil {
				Logger.Warning(fmt.Sprintf("Could not get shared group: %v", b.SharedGroup))
				continue
			}
			sharedBalances := sharedGroup.GetBalances(destination, category, balanceType, account)
			sharedBalances = sharedGroup.SortBalancesByStrategy(b, sharedBalances)
			bc = append(bc, sharedBalances...)
		} else {
			bc = append(bc, b)
		}
	}
	return
}

func (ub *Account) debitCreditBalance(cd *CallDescriptor, count bool, dryRun bool, goNegative bool) (cc *CallCost, err error) {
	usefulUnitBalances := ub.getAlldBalancesForPrefix(cd.Destination, cd.Category, cd.TOR+cd.Direction)
	usefulMoneyBalances := ub.getAlldBalancesForPrefix(cd.Destination, cd.Category, utils.MONETARY+cd.Direction)
	//log.Print(usefulMoneyBalances, usefulUnitBalances)
	//log.Print("STARTCD: ", cd)
	var leftCC *CallCost
	var initialLength int
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
				//log.Printf("Unit balance: %+v", balance)
				// log.Printf("CD BEFORE UNIT: %+v", cd)

				partCC, debitErr := balance.DebitUnits(cd, balance.account, usefulMoneyBalances, count, dryRun)
				if debitErr != nil {
					return nil, debitErr
				}
				//log.Printf("CD AFTER UNIT: %+v", cd)
				if partCC != nil {
					//log.Printf("partCC: %+v", partCC.Timespans[0])
					initialLength = len(cc.Timespans)
					cc.Timespans = append(cc.Timespans, partCC.Timespans...)
					if initialLength == 0 {
						// this is the first add, debit the connect fee
						ub.DebitConnectionFee(cc, usefulMoneyBalances, count)
					}
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
			}
		}

		// debit money
		moneyBalanceChecker := true
		for moneyBalanceChecker {
			// try every balance multiple times in case one becomes active or ratig changes
			moneyBalanceChecker = false
			for _, balance := range usefulMoneyBalances {
				//log.Printf("Money balance: %+v", balance)
				//log.Printf("CD BEFORE MONEY: %+v", cd)
				partCC, debitErr := balance.DebitMoney(cd, balance.account, count, dryRun)
				if debitErr != nil {
					return nil, debitErr
				}
				//log.Printf("CD AFTER MONEY: %+v", cd)
				//log.Printf("partCC: %+v", partCC)
				if partCC != nil {
					initialLength = len(cc.Timespans)
					cc.Timespans = append(cc.Timespans, partCC.Timespans...)
					if initialLength == 0 {
						// this is the first add, debit the connect fee
						ub.DebitConnectionFee(cc, usefulMoneyBalances, count)
					}
					//for i, ts := range cc.Timespans {
					//log.Printf("cc.times[an[%d]: %+v\n", i, ts)
					//}
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
			}
		}
		//log.Printf("END CD: %+v", cd)
		//log.Print("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	}
	//log.Printf("After balances CD: %+v", cd)
	leftCC, err = cd.getCost()
	if err != nil {
		Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
	}
	if leftCC.Cost == 0 && len(leftCC.Timespans) > 0 {
		cc.Timespans = append(cc.Timespans, leftCC.Timespans...)
	}

	//log.Printf("HERE: %+v %d", leftCC)
	if leftCC.Cost > 0 && goNegative {
		initialLength = len(cc.Timespans)
		cc.Timespans = append(cc.Timespans, leftCC.Timespans...)
		if initialLength == 0 {
			// this is the first add, debit the connect fee
			ub.DebitConnectionFee(cc, usefulMoneyBalances, count)
		}
		//log.Printf("Left CC: %+v ", leftCC)
		// get the default money balanance
		// and go negative on it with the amount still unpaid
		if len(leftCC.Timespans) > 0 && leftCC.Cost > 0 && !ub.AllowNegative && !dryRun {
			Logger.Err(fmt.Sprintf("<Rater> Going negative on account %s with AllowNegative: false", cd.GetAccountKey()))
		}
		for _, ts := range leftCC.Timespans {
			if ts.Increments == nil {
				ts.createIncrementsSlice()
			}
			for _, increment := range ts.Increments {
				cost := increment.Cost
				defaultBalance := ub.GetDefaultMoneyBalance(leftCC.Direction)
				defaultBalance.SubstractValue(cost)
				increment.BalanceInfo.MoneyBalanceUuid = defaultBalance.Uuid
				increment.BalanceInfo.AccountId = ub.Id
				increment.paid = true
				if count {
					ub.countUnits(&Action{BalanceType: utils.MONETARY, Direction: leftCC.Direction, Balance: &Balance{Value: cost, DestinationIds: leftCC.Destination}})
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

func (ub *Account) GetDefaultMoneyBalance(direction string) *Balance {
	for _, balance := range ub.BalanceMap[utils.MONETARY+direction] {
		if balance.IsDefault() {
			return balance
		}
	}
	// create default balance
	defaultBalance := &Balance{
		Uuid:   "DEFAULT" + utils.GenUUID(),
		Weight: 0,
	} // minimum weight
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain)
	}
	ub.BalanceMap[utils.MONETARY+direction] = append(ub.BalanceMap[utils.MONETARY+direction], defaultBalance)
	return defaultBalance
}

func (ub *Account) refundIncrement(increment *Increment, direction, unitType string, count bool) {
	var balance *Balance
	if increment.BalanceInfo.UnitBalanceUuid != "" {
		if balance = ub.BalanceMap[unitType+direction].GetBalance(increment.BalanceInfo.UnitBalanceUuid); balance == nil {
			return
		}
		balance.AddValue(increment.Duration.Seconds())
		if count {
			ub.countUnits(&Action{BalanceType: unitType, Direction: direction, Balance: &Balance{Value: -increment.Duration.Seconds()}})
		}
	}
	// check money too
	if increment.BalanceInfo.MoneyBalanceUuid != "" {
		if balance = ub.BalanceMap[utils.MONETARY+direction].GetBalance(increment.BalanceInfo.MoneyBalanceUuid); balance == nil {
			return
		}
		balance.AddValue(increment.Cost)
		if count {
			ub.countUnits(&Action{BalanceType: utils.MONETARY, Direction: direction, Balance: &Balance{Value: -increment.Cost}})
		}
	}
}

// Scans the action trigers and execute the actions for which trigger is met
func (ub *Account) executeActionTriggers(a *Action) {
	ub.ActionTriggers.Sort()
	for _, at := range ub.ActionTriggers {
		// sanity check
		if !strings.Contains(at.ThresholdType, "counter") &&
			!strings.Contains(at.ThresholdType, "balance") {
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
				if uc.BalanceType == at.BalanceType {
					for _, mb := range uc.Balances {
						if strings.Contains(at.ThresholdType, "*max") {
							if mb.MatchActionTrigger(at) && mb.GetValue() >= at.ThresholdValue {
								// run the actions
								at.Execute(ub, nil)
							}
						} else { //MIN
							if mb.MatchActionTrigger(at) && mb.GetValue() <= at.ThresholdValue {
								// run the actions
								at.Execute(ub, nil)
							}
						}
					}
				}
			}
		} else { // BALANCE
			for _, b := range ub.BalanceMap[at.BalanceType+at.BalanceDirection] {
				if !b.dirty { // do not check clean balances
					continue
				}
				if strings.Contains(at.ThresholdType, "*max") {
					if b.MatchActionTrigger(at) && b.GetValue() >= at.ThresholdValue {
						// run the actions
						at.Execute(ub, nil)
					}
				} else { //MIN
					if b.MatchActionTrigger(at) && b.GetValue() <= at.ThresholdValue {
						// run the actions
						at.Execute(ub, nil)
					}
				}
			}
		}
	}
}

// Mark all action trigers as ready for execution
// If the action is not nil it acts like a filter
func (ub *Account) ResetActionTriggers(a *Action) {
	for _, at := range ub.ActionTriggers {
		if !at.Match(a) {
			continue
		}
		at.Executed = false
	}
	ub.executeActionTriggers(a)
}

// Sets/Unsets recurrent flag for action triggers
func (ub *Account) SetRecurrent(a *Action, recurrent bool) {
	for _, at := range ub.ActionTriggers {
		if !at.Match(a) {
			continue
		}
		at.Recurrent = recurrent
	}
}

// Returns the unit counter that matches the specified action type
func (ub *Account) getUnitCounter(a *Action) *UnitsCounter {
	for _, uc := range ub.UnitCounters {
		direction := a.Direction
		if direction == "" {
			direction = OUTBOUND
		}
		if uc.BalanceType == a.BalanceType && uc.Direction == direction {
			return uc
		}
	}
	return nil
}

// Increments the counter for the type specified in the received Action
// with the actions values
func (ub *Account) countUnits(a *Action) {
	unitsCounter := ub.getUnitCounter(a)
	// if not found add the counter
	if unitsCounter == nil {
		direction := a.Direction
		if direction == "" {
			direction = OUTBOUND
		}
		unitsCounter = &UnitsCounter{BalanceType: a.BalanceType, Direction: direction}
		ub.UnitCounters = append(ub.UnitCounters, unitsCounter)
	}

	unitsCounter.addUnits(a.Balance.GetValue(), a.Balance.DestinationIds) // DestinationIds is actually a destination (number or prefix)
	ub.executeActionTriggers(nil)
}

// Create minute counters for all triggered actions that have actions opertating on balances
func (ub *Account) initCounters() {
	ucTempMap := make(map[string]*UnitsCounter, 2)
	for _, at := range ub.ActionTriggers {
		acs, err := ratingStorage.GetActions(at.ActionsId, false)
		if err != nil {
			continue
		}
		for _, a := range acs {
			if a.Balance != nil {
				direction := at.BalanceDirection
				if direction == "" {
					direction = OUTBOUND
				}
				uc, exists := ucTempMap[direction]
				if !exists {
					uc = &UnitsCounter{BalanceType: a.BalanceType, Direction: direction}
					ucTempMap[direction] = uc
					uc.Balances = BalanceChain{}
					ub.UnitCounters = append(ub.UnitCounters, uc)
				}
				b := a.Balance.Clone()
				b.SetValue(0)
				uc.Balances = append(uc.Balances, b)
				uc.Balances.Sort()
			}
		}
	}
}

func (ub *Account) CleanExpiredBalances() {
	for key, bm := range ub.BalanceMap {
		for i := 0; i < len(bm); i++ {
			if bm[i].IsExpired() {
				// delete it
				bm = append(bm[:i], bm[i+1:]...)
			}
		}
		ub.BalanceMap[key] = bm
	}
}

func (ub *Account) allBalancesExpired() bool {
	for _, bm := range ub.BalanceMap {
		for i := 0; i < len(bm); i++ {
			if !bm[i].IsExpired() {
				return false
			}
		}
	}
	return true
}

// returns the shared groups that this user balance belnongs to
func (ub *Account) GetSharedGroups() (groups []string) {
	for _, balanceChain := range ub.BalanceMap {
		for _, b := range balanceChain {
			if b.SharedGroup != "" {
				groups = append(groups, b.SharedGroup)
			}
		}
	}
	return
}

func (account *Account) GetUniqueSharedGroupMembers(cd *CallDescriptor) ([]string, error) {
	var balances []*Balance
	balances = append(balances, account.getBalancesForPrefix(cd.Destination, cd.Category, account.BalanceMap[utils.MONETARY+cd.Direction], "")...)
	balances = append(balances, account.getBalancesForPrefix(cd.Destination, cd.Category, account.BalanceMap[cd.TOR+cd.Direction], "")...)
	// gather all shared group ids
	var sharedGroupIds []string
	for _, b := range balances {
		if b.SharedGroup != "" {
			sharedGroupIds = append(sharedGroupIds, b.SharedGroup)
		}
	}
	var memberIds []string
	for _, sgID := range sharedGroupIds {
		sharedGroup, err := ratingStorage.GetSharedGroup(sgID, false)
		if err != nil {
			Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgID))
			return nil, err
		}
		for _, memberId := range sharedGroup.MemberIds {
			if !utils.IsSliceMember(memberIds, memberId) {
				memberIds = append(memberIds, memberId)
			}
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
					acc.countUnits(&Action{BalanceType: utils.MONETARY, Direction: cc.Direction, Balance: &Balance{Value: connectFee, DestinationIds: cc.Destination}})
				}
				connectFeePaid = true
				break
			}
		}
		// debit connect fee
		if connectFee > 0 && !connectFeePaid {
			// there are no money for the connect fee; go negative
			acc.GetDefaultMoneyBalance(cc.Direction).SubstractValue(connectFee)
			// the conect fee is not refundable!
			if count {
				acc.countUnits(&Action{BalanceType: utils.MONETARY, Direction: cc.Direction, Balance: &Balance{Value: connectFee, DestinationIds: cc.Destination}})
			}
		}
	}
}
