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
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"

	"strings"
)

const (
	// Direction type
	INBOUND  = "*in"
	OUTBOUND = "*out"
	// Balance types
	CREDIT  = utils.MONETARY
	SMS     = utils.SMS
	DATA    = utils.DATA
	MINUTES = utils.VOICE
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
	creditBalances := ub.getBalancesForPrefix(cd.Destination, ub.BalanceMap[CREDIT+cd.Direction], "")
	unitBalances := ub.getBalancesForPrefix(cd.Destination, ub.BalanceMap[cd.TOR+cd.Direction], "")
	// gather all balances from shared groups
	var extendedCreditBalances BalanceChain
	for _, cb := range creditBalances {
		if cb.SharedGroup != "" {
			if sharedGroup, _ := accountingStorage.GetSharedGroup(cb.SharedGroup, false); sharedGroup != nil {
				sgb := sharedGroup.GetBalances(cd.Destination, CREDIT+cd.Direction, ub)
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
			if sharedGroup, _ := accountingStorage.GetSharedGroup(mb.SharedGroup, false); sharedGroup != nil {
				sgb := sharedGroup.GetBalances(cd.Destination, cd.TOR+cd.Direction, ub)
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
func (ub *Account) debitBalanceAction(a *Action) error {
	if a == nil {
		return errors.New("nil minute action!")
	}
	if a.Balance.Uuid == "" {
		a.Balance.Uuid = utils.GenUUID()
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain, 1)
	}
	found := false
	id := a.BalanceType + a.Direction
	for _, b := range ub.BalanceMap[id] {
		if b.IsExpired() {
			continue // we can clean expired balances balances here
		}
		if b.Equal(a.Balance) {
			b.SubstractAmount(a.Balance.Value)
			found = true
			break
		}
	}
	// if it is not found then we add it to the list
	if !found {
		a.Balance.Value = -a.Balance.Value
		ub.BalanceMap[id] = append(ub.BalanceMap[id], a.Balance)
		if a.Balance.SharedGroup != "" {
			// add shared group member
			sg, err := accountingStorage.GetSharedGroup(a.Balance.SharedGroup, false)
			if err != nil || sg == nil {
				//than problem
				Logger.Warning(fmt.Sprintf("Could not get shared group: %v", a.Balance.SharedGroup))
			} else {
				// add member and save
				sg.MemberIds = append(sg.MemberIds, ub.Id)
				accountingStorage.SetSharedGroup(sg)
			}
		}
	}
	ub.executeActionTriggers(nil)
	return nil //ub.BalanceMap[id].GetTotalValue()
}

func (ub *Account) getBalancesForPrefix(prefix string, balances BalanceChain, sharedGroup string) BalanceChain {
	var usefulBalances BalanceChain
	for _, b := range balances {
		if b.IsExpired() || (ub.AllowNegative == false && b.SharedGroup == "" && b.Value <= 0) {
			continue
		}
		if sharedGroup != "" && sharedGroup != "" && b.SharedGroup != sharedGroup {
			continue
		}
		b.account = ub
		if b.DestinationId != "" && b.DestinationId != utils.ANY {
			for _, p := range utils.SplitPrefix(prefix, MIN_PREFIX_MATCH) {
				if x, err := cache2go.GetCached(DESTINATION_PREFIX + p); err == nil {
					destIds := x.([]string)
					for _, dId := range destIds {
						if dId == b.DestinationId {
							b.precision = len(p)
							usefulBalances = append(usefulBalances, b)
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
func (account *Account) getAlldBalancesForPrefix(destination, balanceType string) (bc BalanceChain) {
	balances := account.getBalancesForPrefix(destination, account.BalanceMap[balanceType], "")
	for _, b := range balances {
		if b.SharedGroup != "" {
			sharedGroup, err := accountingStorage.GetSharedGroup(b.SharedGroup, false)
			if err != nil {
				Logger.Warning(fmt.Sprintf("Could not get shared group: %v", b.SharedGroup))
				continue
			}
			sharedBalances := sharedGroup.GetBalances(destination, balanceType, account)
			sharedBalances = sharedGroup.SortBalancesByStrategy(b, sharedBalances)
			bc = append(bc, sharedBalances...)
		} else {
			bc = append(bc, b)
		}
	}
	return
}

func (ub *Account) debitCreditBalance(cc *CallCost, count bool) (err error) {
	usefulUnitBalances := ub.getAlldBalancesForPrefix(cc.Destination, cc.TOR+cc.Direction)
	usefulMoneyBalances := ub.getAlldBalancesForPrefix(cc.Destination, CREDIT+cc.Direction)
	// debit minutes
	for _, balance := range usefulUnitBalances {
		balance.DebitUnits(cc, count, balance.account, usefulMoneyBalances)
		if cc.IsPaid() {
			goto CONNECT_FEE
		}
	}
	for tsIndex := 0; tsIndex < len(cc.Timespans); tsIndex++ {
		ts := cc.Timespans[tsIndex]
		if paid, incrementIndex := ts.IsPaid(); !paid {
			newTs := ts.SplitByIncrement(incrementIndex)
			if newTs != nil {
				idx := tsIndex + 1
				cc.Timespans = append(cc.Timespans, nil)
				copy(cc.Timespans[idx+1:], cc.Timespans[idx:])
				cc.Timespans[idx] = newTs
			}
		}
	}
	// debit money
	for _, balance := range usefulMoneyBalances {
		balance.DebitMoney(cc, count, balance.account)
		if cc.IsPaid() {
			goto CONNECT_FEE
		}
	}
	// get the default money balanance
	// and go negative on it with the amount still unpaid
	for tsIndex := 0; tsIndex < len(cc.Timespans); tsIndex++ {
		ts := cc.Timespans[tsIndex]
		if ts.Increments == nil {
			ts.createIncrementsSlice()
		}
		if paid, incrementIndex := ts.IsPaid(); !paid {
			newTs := ts.SplitByIncrement(incrementIndex)
			if newTs != nil {
				idx := tsIndex + 1
				cc.Timespans = append(cc.Timespans, nil)
				copy(cc.Timespans[idx+1:], cc.Timespans[idx:])
				cc.Timespans[idx] = newTs
				continue
			}
			for _, increment := range ts.Increments {
				cost := increment.Cost
				ub.GetDefaultMoneyBalance(cc.Direction).SubstractAmount(cost)
				if count {
					ub.countUnits(&Action{BalanceType: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: cost, DestinationId: cc.Destination}})
				}
				err = errors.New("not enough credit")
			}
		}
	}
CONNECT_FEE:
	if cc.deductConnectFee {
		connectFee := cc.GetConnectFee()
		connectFeePaid := false
		for _, b := range usefulMoneyBalances {
			if b.Value >= connectFee {
				b.SubstractAmount(connectFee)
				// the conect fee is not refundable!
				if count {
					ub.countUnits(&Action{BalanceType: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: connectFee, DestinationId: cc.Destination}})
				}
				connectFeePaid = true
				break
			}
		}
		// debit connect fee
		if connectFee > 0 && !connectFeePaid {
			// there are no money for the connect fee; go negative
			ub.GetDefaultMoneyBalance(cc.Direction).Value -= connectFee
			// the conect fee is not refundable!
			if count {
				ub.countUnits(&Action{BalanceType: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: connectFee, DestinationId: cc.Destination}})
			}
		}
	}
	// save darty shared balances
	usefulMoneyBalances.SaveDirtyBalances(ub)
	usefulUnitBalances.SaveDirtyBalances(ub)
	return
}

func (ub *Account) GetDefaultMoneyBalance(direction string) *Balance {
	for _, balance := range ub.BalanceMap[CREDIT+direction] {
		if balance.IsDefault() {
			return balance
		}
	}
	// create default balance
	defaultBalance := &Balance{Weight: 0} // minimum weight
	ub.BalanceMap[CREDIT+direction] = append(ub.BalanceMap[CREDIT+direction], defaultBalance)
	return defaultBalance
}

func (ub *Account) refundIncrement(increment *Increment, direction, unitType string, count bool) {
	var balance *Balance
	if increment.BalanceInfo.UnitBalanceUuid != "" {
		if balance = ub.BalanceMap[unitType+direction].GetBalance(increment.BalanceInfo.UnitBalanceUuid); balance == nil {
			return
		}
		balance.Value += increment.Duration.Seconds()
		if count {
			ub.countUnits(&Action{BalanceType: unitType, Direction: direction, Balance: &Balance{Value: -increment.Duration.Seconds()}})
		}
	}
	// check money too
	if increment.BalanceInfo.MoneyBalanceUuid != "" {
		if balance = ub.BalanceMap[CREDIT+direction].GetBalance(increment.BalanceInfo.MoneyBalanceUuid); balance == nil {
			return
		}
		balance.Value += increment.Cost
		if count {
			ub.countUnits(&Action{BalanceType: CREDIT, Direction: direction, Balance: &Balance{Value: -increment.Cost}})
		}
	}
}

// Scans the action trigers and execute the actions for which trigger is met
func (ub *Account) executeActionTriggers(a *Action) {
	ub.ActionTriggers.Sort()
	for _, at := range ub.ActionTriggers {
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
							if mb.MatchDestination(at.DestinationId) && mb.Value >= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						} else { //MIN
							if mb.MatchDestination(at.DestinationId) && mb.Value <= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						}
					}
				}
			}
		} else { // BALANCE
			for _, b := range ub.BalanceMap[at.BalanceType+at.Direction] {
				if strings.Contains(at.ThresholdType, "*max") {
					if b.MatchDestination(at.DestinationId) && b.Value >= at.ThresholdValue {
						// run the actions
						at.Execute(ub)
					}
				} else { //MIN
					if b.MatchDestination(at.DestinationId) && b.Value <= at.ThresholdValue {
						// run the actions
						at.Execute(ub)
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

	unitsCounter.addUnits(a.Balance.Value, a.Balance.DestinationId) // DestinationId is actually a destination (number or prefix)
	ub.executeActionTriggers(nil)
}

// Create minute counters for all triggered actions that have actions opertating on balances
func (ub *Account) initCounters() {
	ucTempMap := make(map[string]*UnitsCounter, 2)
	for _, at := range ub.ActionTriggers {
		acs, err := accountingStorage.GetActions(at.ActionsId, false)
		if err != nil {
			continue
		}
		for _, a := range acs {
			if a.Balance != nil {
				direction := at.Direction
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
				b.Value = 0
				uc.Balances = append(uc.Balances, b)
				uc.Balances.Sort()
			}
		}
	}
}

func (ub *Account) CleanExpiredBalancesAndBuckets() {
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

func (account *Account) GetUniqueSharedGroupMembers(destination, direction, unitType string) ([]string, error) {
	creditBalances := account.getBalancesForPrefix(destination, account.BalanceMap[CREDIT+direction], "")
	unitBalances := account.getBalancesForPrefix(destination, account.BalanceMap[unitType+direction], "")
	// gather all shared group ids
	var sharedGroupIds []string
	for _, cb := range creditBalances {
		if cb.SharedGroup != "" {
			sharedGroupIds = append(sharedGroupIds, cb.SharedGroup)
		}
	}
	for _, mb := range unitBalances {
		if mb.SharedGroup != "" {
			sharedGroupIds = append(sharedGroupIds, mb.SharedGroup)
		}
	}
	var memberIds []string
	for _, sgID := range sharedGroupIds {
		sharedGroup, err := accountingStorage.GetSharedGroup(sgID, false)
		if err != nil {
			Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgID))
			return nil, err
		}
		for _, memberId := range sharedGroup.GetMembersExceptUser(account.Id) {
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
