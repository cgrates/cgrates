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
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"

	"strings"
)

const (
	UB_TYPE_POSTPAID = "*postpaid"
	UB_TYPE_PREPAID  = "*prepaid"
	// Direction type
	INBOUND  = "*in"
	OUTBOUND = "*out"
	// Balance types
	CREDIT       = "*monetary"
	SMS          = "*sms"
	TRAFFIC      = "*internet"
	TRAFFIC_TIME = "*internet_time"
	MINUTES      = "*minutes"
	// action price type
	PRICE_PERCENT  = "*percent"
	PRICE_ABSOLUTE = "*absolute"
	// action trigger threshold types
	TRIGGER_MIN_COUNTER = "*min_counter"
	TRIGGER_MAX_COUNTER = "*max_counter"
	TRIGGER_MIN_BALANCE = "*min_balance"
	TRIGGER_MAX_BALANCE = "*max_balance"
)

var (
	AMOUNT_TOO_BIG = errors.New("Amount excedes balance!")
)

/*
Structure containing information about user's credit (minutes, cents, sms...).'
This can represent a user or a shared group.
*/
type UserBalance struct {
	Id             string
	Type           string // prepaid-postpaid
	BalanceMap     map[string]BalanceChain
	UnitCounters   []*UnitsCounter
	ActionTriggers ActionTriggerPriotityList
	Groups         GroupLinks // user info about groups
	// group information
	UserIds  []string // group info about users
	Disabled bool
}

// Returns user's available minutes for the specified destination
func (ub *UserBalance) getCreditForPrefix(cd *CallDescriptor) (duration time.Duration, credit float64, balances BalanceChain) {
	credit = ub.getBalancesForPrefix(cd.Destination, ub.BalanceMap[CREDIT+cd.Direction]).GetTotalValue()
	balances = ub.getBalancesForPrefix(cd.Destination, ub.BalanceMap[MINUTES+cd.Direction])

	for _, b := range balances {
		d, c := b.GetMinutesForCredit(cd, credit)
		credit = c
		duration += d
	}
	return
}

// Debits some amount of user's specified balance adding the balance if it does not exists.
// Returns the remaining credit in user's balance.
func (ub *UserBalance) debitBalanceAction(a *Action) error {
	if a == nil {
		return errors.New("nil minute action!")
	}
	if a.Balance.Uuid == "" {
		a.Balance.Uuid = utils.GenUUID()
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain, 0)
	}
	found := false
	id := a.BalanceId + a.Direction
	for _, b := range ub.BalanceMap[id] {
		if b.IsExpired() {
			continue // we can clean expired balances balances here
		}
		if b.Equal(a.Balance) {
			b.Value -= a.Balance.Value
			found = true
			break
		}
	}
	// if it is not found and the Seconds are negative (topup)
	// then we add it to the list
	if !found && a.Balance.Value <= 0 {
		a.Balance.Value = -a.Balance.Value
		ub.BalanceMap[id] = append(ub.BalanceMap[id], a.Balance)
	}
	ub.executeActionTriggers(nil)
	return nil //ub.BalanceMap[id].GetTotalValue()
}

func (ub *UserBalance) getBalancesForPrefix(prefix string, balances BalanceChain) BalanceChain {
	var usefulBalances BalanceChain
	for _, b := range balances {
		if b.IsExpired() || (ub.Type != UB_TYPE_POSTPAID && b.Value <= 0) {
			continue
		}
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
	return usefulBalances
}

func (ub *UserBalance) debitCreditBalance(cc *CallCost, count bool) error {
	usefulMinuteBalances := ub.getBalancesForPrefix(cc.Destination, ub.BalanceMap[MINUTES+cc.Direction])
	usefulMoneyBalances := ub.getBalancesForPrefix(cc.Destination, ub.BalanceMap[CREDIT+cc.Direction])
	// debit minutes
	for _, balance := range usefulMinuteBalances {
		balance.DebitMinutes(cc, count, ub, usefulMoneyBalances)
	}
	allPaidWithMinutes := true
	for tsIndex := 0; tsIndex < len(cc.Timespans); tsIndex++ {
		ts := cc.Timespans[tsIndex]
		if paid, incrementIndex := ts.IsPaid(); !paid {
			allPaidWithMinutes = false
			newTs := ts.SplitByIncrement(incrementIndex)
			if newTs != nil {
				idx := tsIndex + 1
				cc.Timespans = append(cc.Timespans, nil)
				copy(cc.Timespans[idx+1:], cc.Timespans[idx:])
				cc.Timespans[idx] = newTs
			}
		}
	}
	var returnError error
	insuficientCreditError := errors.New("not enough credit")
	moneyBalance := ub.GetDefaultMoneyBalance(cc.Direction)
	if allPaidWithMinutes {
		goto CONNECT_FEE
	}
	// debit money
	for _, balance := range usefulMoneyBalances {
		balance.DebitMoney(cc, count, ub)
	}
	// get the highest priority money balanance
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
				moneyBalance.Value -= cost
				if count {
					ub.countUnits(&Action{BalanceId: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: cost, DestinationId: cc.Destination}})
				}
				returnError = insuficientCreditError
			}
		}
	}
CONNECT_FEE:
	if cc.deductConnectFee {
		amount := cc.GetConnectFee()
		connectFeePaid := false
		for _, b := range usefulMoneyBalances {
			if b.Value >= amount {
				b.Value -= amount
				// the conect fee is not refundable!
				if count {
					ub.countUnits(&Action{BalanceId: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
				}
				connectFeePaid = true
				break
			}
		}
		// debit connect fee
		if cc.GetConnectFee() > 0 && !connectFeePaid {
			// there are no money for the connect fee; go negative
			moneyBalance.Value -= amount
			// the conect fee is not refundable!
			if count {
				ub.countUnits(&Action{BalanceId: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
			}
		}
	}
	return returnError
}

func (ub *UserBalance) GetDefaultMoneyBalance(direction string) *Balance {
	for _, balance := range ub.BalanceMap[CREDIT+direction] {
		if balance.IsDefault() {
			return balance
		}
	}

	// create default balance
	defaultBalance := &Balance{Weight: 999}
	ub.BalanceMap[CREDIT+direction] = append(ub.BalanceMap[CREDIT+direction], defaultBalance)
	return defaultBalance
}

func (ub *UserBalance) refundIncrements(increments Increments, direction string, count bool) {
	for _, increment := range increments {
		var balance *Balance
		if increment.GetMinuteBalance() != "" {
			if balance = ub.BalanceMap[MINUTES+direction].GetBalance(increment.GetMinuteBalance()); balance == nil {
				continue
			}
			balance.Value += increment.Duration.Seconds()
			if count {
				ub.countUnits(&Action{BalanceId: MINUTES, Direction: direction, Balance: &Balance{Value: -increment.Duration.Seconds()}})
			}
		}
		// check money too
		if increment.GetMoneyBalance() != "" {
			if balance = ub.BalanceMap[CREDIT+direction].GetBalance(increment.GetMoneyBalance()); balance == nil {
				continue
			}
			balance.Value += increment.Cost
			if count {
				ub.countUnits(&Action{BalanceId: CREDIT, Direction: direction, Balance: &Balance{Value: -increment.Cost}})
			}
		}
	}
}

/*
Debits some amount of user's specified balance. Returns the remaining credit in user's balance.
*/
func (ub *UserBalance) debitGenericBalance(balanceId string, direction string, amount float64, count bool) float64 {
	if count {
		ub.countUnits(&Action{BalanceId: balanceId, Direction: direction, Balance: &Balance{Value: amount}})
	}
	ub.BalanceMap[balanceId+direction].Debit(amount)
	return ub.BalanceMap[balanceId+direction].GetTotalValue()
}

// Scans the action trigers and execute the actions for which trigger is met
func (ub *UserBalance) executeActionTriggers(a *Action) {
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
				if uc.BalanceId == at.BalanceId {
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
			for _, b := range ub.BalanceMap[at.BalanceId+at.Direction] {
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
func (ub *UserBalance) resetActionTriggers(a *Action) {
	for _, at := range ub.ActionTriggers {
		if !at.Match(a) {
			continue
		}
		at.Executed = false
	}
	ub.executeActionTriggers(a)
}

// Returns the unit counter that matches the specified action type
func (ub *UserBalance) getUnitCounter(a *Action) *UnitsCounter {
	for _, uc := range ub.UnitCounters {
		direction := a.Direction
		if direction == "" {
			direction = OUTBOUND
		}
		if uc.BalanceId == a.BalanceId && uc.Direction == direction {
			return uc
		}
	}
	return nil
}

// Increments the counter for the type specified in the received Action
// with the actions values
func (ub *UserBalance) countUnits(a *Action) {
	unitsCounter := ub.getUnitCounter(a)
	// if not found add the counter
	if unitsCounter == nil {
		direction := a.Direction
		if direction == "" {
			direction = OUTBOUND
		}
		unitsCounter = &UnitsCounter{BalanceId: a.BalanceId, Direction: direction}
		ub.UnitCounters = append(ub.UnitCounters, unitsCounter)
	}

	unitsCounter.addUnits(a.Balance.Value, a.Balance.DestinationId) // DestinationId is actually a destination (number or prefix)
	ub.executeActionTriggers(nil)
}

// Create minute counters for all triggered actions that have actions operating on minute buckets
func (ub *UserBalance) initCounters() {
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
					uc = &UnitsCounter{BalanceId: a.BalanceId, Direction: direction}
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

func (ub *UserBalance) CleanExpiredBalancesAndBuckets() {
	for key, _ := range ub.BalanceMap {
		bm := ub.BalanceMap[key]
		for i := 0; i < len(bm); i++ {
			if bm[i].IsExpired() {
				// delete it
				bm = append(bm[:i], bm[i+1:]...)
			}
		}
		ub.BalanceMap[key] = bm
	}
	for i := 0; i < len(ub.BalanceMap[MINUTES+OUTBOUND]); i++ {
		if ub.BalanceMap[MINUTES+OUTBOUND][i].IsExpired() {
			ub.BalanceMap[MINUTES+OUTBOUND] = append(ub.BalanceMap[MINUTES+OUTBOUND][:i], ub.BalanceMap[MINUTES+OUTBOUND][i+1:]...)
		}
	}
}
