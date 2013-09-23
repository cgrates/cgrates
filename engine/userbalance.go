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
	// minute subjects
	ZEROSECOND = "*zerosecond"
	ZEROMINUTE = "*zerominute"
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

	Groups GroupLinks // user info about groups
	// group information
	UserIds []string // group info about users
}

// Returns user's available minutes for the specified destination
func (ub *UserBalance) getSecondsForPrefix(prefix string) (seconds, credit float64, balances BalanceChain) {
	credit = ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue()
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) == 0 {
		// Logger.Debug("There are no minute buckets to check for user: ", ub.Id)
		return
	}
	for _, b := range ub.BalanceMap[MINUTES+OUTBOUND] {
		if b.IsExpired() {
			continue
		}
		d, err := GetDestination(b.DestinationId)
		if err != nil {
			continue
		}
		if precision, ok := d.containsPrefix(prefix); ok {
			b.precision = precision
			if b.Value > 0 {
				balances = append(balances, b)
			}
		}
	}
	balances.Sort() // sorts the buckets according to priority, precision or price
	for _, b := range balances {
		s := b.GetSecondsForCredit(credit)
		credit -= s * b.SpecialPrice
		seconds += s
	}
	return
}

// Debits some amount of user's specified balance adding the balance if it does not exists.
// Returns the remaining credit in user's balance.
func (ub *UserBalance) debitBalanceAction(a *Action) error {
	if a == nil {
		return errors.New("nil minute action!")
	}
	if a.Balance.Id == "" {
		a.Balance.Id = utils.GenUUID()
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
	return nil //ub.BalanceMap[id].GetTotalValue()
}

func (ub *UserBalance) getBalanceForPrefix(prefix string, balances BalanceChain) BalanceChain {
	var usefulBalances BalanceChain
	for _, b := range balances {
		if b.IsExpired() {
			continue
		}
		if b.DestinationId != "" {
			precision, err := storageGetter.DestinationContainsPrefix(b.DestinationId, prefix)
			if err != nil {
				continue
			}
			if precision > 0 {
				b.precision = precision
				if b.Value > 0 {
					balances = append(balances, b)
				}
			}
		}
	}
	// resort by precision
	usefulBalances.Sort()
	return usefulBalances
}

/*
This method is the core of userbalance debiting: don't panic just follow the branches
*/
func (ub *UserBalance) debitCreditBalance(cc *CallCost, count bool) error {
	// debit minutes first
	minuteBalances := ub.BalanceMap[MINUTES+cc.Direction]

	usefulBalances := ub.getBalanceForPrefix(cc.Destination, minuteBalances)

	for _, ts := range cc.Timespans {
		ts.createIncrementsSlice()
		for incrementIndex, increment := range ts.Increments {
			for _, b := range usefulBalances {
				if b.Value == 0 {
					continue
				}

				// check standard subject tags
				if b.RateSubject == ZEROSECOND || b.RateSubject == "" {
					if b.Value >= increment.Duration.Seconds() {
						b.Value -= increment.Duration.Seconds()
						increment.BalanceId = b.Id
						break
					}
				}
				if b.RateSubject == ZEROMINUTE {
					if b.Value >= 60 {
						// TODO: round to minute (consume the rest of the timespans)
						nts := ts.SplitByIncrement(incrementIndex, increment)
						nts.RoundToDuration(increment.Duration)

						b.Value -= 60
						increment.BalanceId = b.Id
						break
					}
				}
				// nts.SplitByIncrement()
				// get the new rate
				cd := cc.CreateCallDescriptor()
				cd.TimeStart = ts.GetTimeStartForIncrement(incrementIndex, increment)
				cd.TimeEnd = cc.Timespans[len(cc.Timespans)-1].TimeEnd
				cd.CallDuration = cc.Timespans[len(cc.Timespans)-1].CallDuration
				newCC, err := cd.GetCost()
				if err != nil {
					Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
					continue
				}
				//debit new callcost
				for _, nts := range newCC.Timespans {
					for _, nIncrement := range nts.Increments {
						// debit minutes and money
						_ = nIncrement
					}
				}
			}
			if increment.BalanceId == "" {
				// no balance was attached to this increment: cut the rest of increments/timespans

			}
		}
	}

	if count {
		ub.countUnits(&Action{BalanceId: MINUTES, Direction: OUTBOUND, Balance: &Balance{Value: cc.Cost + cc.ConnectFee, DestinationId: cc.Destination}})
	}
	return nil
}

/*
Debits some amount of user's specified balance. Returns the remaining credit in user's balance.
*/
func (ub *UserBalance) debitBalance(balanceId string, amount float64, count bool) float64 {
	if count {
		ub.countUnits(&Action{BalanceId: balanceId, Direction: OUTBOUND, Balance: &Balance{Value: amount}})
	}
	ub.BalanceMap[balanceId+OUTBOUND].Debit(amount)
	return ub.BalanceMap[balanceId+OUTBOUND].GetTotalValue()
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
					if at.BalanceId == MINUTES {
						for _, mb := range uc.MinuteBalances {
							if strings.Contains(at.ThresholdType, "*max") {
								if mb.DestinationId == at.DestinationId && mb.Value >= at.ThresholdValue {
									// run the actions
									at.Execute(ub)
								}
							} else { //MIN
								if mb.DestinationId == at.DestinationId && mb.Value <= at.ThresholdValue {
									// run the actions
									at.Execute(ub)
								}
							}
						}
					} else {
						if strings.Contains(at.ThresholdType, "*max") {
							if uc.Units >= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						} else { //MIN
							if uc.Units <= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						}
					}
				}
			}
		} else { // BALANCE
			for _, b := range ub.BalanceMap[at.BalanceId] {
				if at.BalanceId == MINUTES {
					for _, mb := range ub.BalanceMap[MINUTES+OUTBOUND] {
						if strings.Contains(at.ThresholdType, "*max") {
							if mb.DestinationId == at.DestinationId && mb.Value >= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						} else { //MIN
							if mb.DestinationId == at.DestinationId && mb.Value <= at.ThresholdValue {
								// run the actions
								at.Execute(ub)
							}
						}
					}
				} else {
					if strings.Contains(at.ThresholdType, "*max") {
						if b.Value >= at.ThresholdValue {
							// run the actions
							at.Execute(ub)
						}
					} else { //MIN
						if b.Value <= at.ThresholdValue {
							// run the actions
							at.Execute(ub)
						}
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
	if a.BalanceId == MINUTES && a.Balance != nil {
		unitsCounter.addMinutes(a.Balance.Value, a.Balance.DestinationId)
	} else {
		unitsCounter.Units += a.Balance.Value
	}
	ub.executeActionTriggers(nil)
}

// Create minute counters for all triggered actions that have actions operating on minute buckets
func (ub *UserBalance) initMinuteCounters() {
	ucTempMap := make(map[string]*UnitsCounter, 2)
	for _, at := range ub.ActionTriggers {
		acs, err := storageGetter.GetActions(at.ActionsId)
		if err != nil {
			continue
		}
		for _, a := range acs {
			if a.BalanceId == MINUTES && a.Balance != nil {
				direction := at.Direction
				if direction == "" {
					direction = OUTBOUND
				}
				uc, exists := ucTempMap[direction]
				if !exists {
					uc = &UnitsCounter{BalanceId: MINUTES, Direction: direction}
					ucTempMap[direction] = uc
					uc.MinuteBalances = BalanceChain{}
					ub.UnitCounters = append(ub.UnitCounters, uc)
				}
				b := a.Balance.Clone()
				b.Value = 0
				uc.MinuteBalances = append(uc.MinuteBalances, b)
				uc.MinuteBalances.Sort()
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
