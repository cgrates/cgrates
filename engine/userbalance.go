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
	"time"
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
	Groups         GroupLinks // user info about groups
	// group information
	UserIds []string // group info about users
}

// Returns user's available minutes for the specified destination
func (ub *UserBalance) getSecondsForPrefix(cd *CallDescriptor) (seconds, credit float64, balances BalanceChain) {
	credit = ub.getBalancesForPrefix(cd.Destination, ub.BalanceMap[CREDIT+cd.Direction]).GetTotalValue()
	balances = ub.getBalancesForPrefix(cd.Destination, ub.BalanceMap[MINUTES+cd.Direction])
	for _, b := range balances {
		s := b.GetSecondsForCredit(cd, credit)
		cc, err := b.GetCost(cd)
		if err != nil {
			Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
			continue
		}
		if cc.Cost > 0 && cc.GetDuration() > 0 {
			// TODO: fix this
			secondCost := cc.Cost / cc.GetDuration().Seconds()
			credit -= s * secondCost
		}
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
	return nil //ub.BalanceMap[id].GetTotalValue()
}

func (ub *UserBalance) getBalancesForPrefix(prefix string, balances BalanceChain) BalanceChain {
	var usefulBalances BalanceChain
	for _, b := range balances {
		if b.IsExpired() || (ub.Type != UB_TYPE_POSTPAID && b.Value <= 0) {
			continue
		}
		if b.DestinationId != "" {
			precision, err := storageGetter.DestinationContainsPrefix(b.DestinationId, prefix)
			if err != nil {
				continue
			}
			if precision > 0 {
				b.precision = precision
				usefulBalances = append(usefulBalances, b)
			}
		} else {
			usefulBalances = append(usefulBalances, b)
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
	minuteBalances := ub.BalanceMap[MINUTES+cc.Direction]
	moneyBalances := ub.BalanceMap[CREDIT+cc.Direction]
	usefulMinuteBalances := ub.getBalancesForPrefix(cc.Destination, minuteBalances)
	usefulMoneyBalances := ub.getBalancesForPrefix(cc.Destination, moneyBalances)
	// debit connect fee
	if cc.ConnectFee > 0 {
		amount := cc.ConnectFee
		paid := false
		for _, b := range usefulMoneyBalances {
			if b.Value >= amount {
				b.Value -= amount
				// the conect fee is not refoundable!
				if count {
					ub.countUnits(&Action{BalanceId: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
				}
				paid = true
				break
			}
		}
		if !paid {
			// there are no money for the connect fee; abort mission
			cc.Timespans = make([]*TimeSpan, 0)
			return nil
		}
	}
	// debit minutes
	for tsIndex := 0; tsIndex < len(cc.Timespans); tsIndex++ {
		ts := cc.Timespans[tsIndex]
		ts.createIncrementsSlice()
		tsWasSplited := false
		for incrementIndex, increment := range ts.Increments {
			if tsWasSplited {
				break
			}
			paid := false
			for _, b := range usefulMinuteBalances {
				// check standard subject tags
				if b.RateSubject == ZEROSECOND || b.RateSubject == "" {
					amount := increment.Duration.Seconds()
					if b.Value >= amount {
						b.Value -= amount
						increment.BalanceUuid = b.Uuid
						increment.MinuteInfo = &MinuteInfo{b.DestinationId, amount, 0}
						paid = true
						if count {
							ub.countUnits(&Action{BalanceId: MINUTES, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
						}
						break
					}
				}
				if b.RateSubject == ZEROMINUTE {
					amount := time.Minute.Seconds()
					if b.Value >= amount { // balance has at least 60 seconds
						newTs := ts
						if incrementIndex != 0 {
							// if increment it's not at the begining we must split the timespan
							newTs = ts.SplitByIncrement(incrementIndex, increment)
						}
						newTs.RoundToDuration(time.Minute)
						newTs.RateInterval = &RateInterval{
							Rates: RateGroups{
								&Rate{
									GroupIntervalStart: 0,
									Value:              0,
									RateIncrement:      time.Minute,
									RateUnit:           time.Minute,
								},
							},
						}
						newTs.createIncrementsSlice()
						// overlap the rest of the timespans
						for i := tsIndex + 1; i < len(cc.Timespans); i++ {
							if cc.Timespans[i].TimeEnd.Before(newTs.TimeEnd) || cc.Timespans[i].TimeEnd.Equal(newTs.TimeEnd) {
								cc.Timespans[i].overlapped = true
							} else if cc.Timespans[i].TimeStart.Before(newTs.TimeEnd) {
								cc.Timespans[i].TimeStart = ts.TimeEnd
							}
						}
						// insert the new timespan
						if newTs != ts {
							tsIndex++
							cc.Timespans = append(cc.Timespans, nil)
							copy(cc.Timespans[tsIndex+1:], cc.Timespans[tsIndex:])
							cc.Timespans[tsIndex] = newTs
							tsWasSplited = true
						}

						var newTimespans []*TimeSpan
						// remove overlapped
						for _, ots := range cc.Timespans {
							if !ots.overlapped {
								newTimespans = append(newTimespans, ots)
							}
						}
						cc.Timespans = newTimespans
						b.Value -= amount
						newTs.Increments[0].BalanceUuid = b.Uuid
						newTs.Increments[0].MinuteInfo = &MinuteInfo{b.DestinationId, amount, 0}
						paid = true
						if count {
							ub.countUnits(&Action{BalanceId: MINUTES, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
						}
						break
					}
				}
				// newTs.SplitByIncrement()
				// get the new rate
				cd := cc.CreateCallDescriptor()
				cd.TimeStart = ts.GetTimeStartForIncrement(incrementIndex, increment)
				cd.TimeEnd = cc.Timespans[len(cc.Timespans)-1].TimeEnd
				cd.CallDuration = cc.Timespans[len(cc.Timespans)-1].CallDuration
				newCC, err := b.GetCost(cd)
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
			if paid {
				continue
			} else {
				// Split if some increments were processed by minutes
				if incrementIndex > 0 && ts.Increments[incrementIndex-1].MinuteInfo != nil {
					newTs := ts.SplitByIncrement(incrementIndex, increment)
					idx := tsIndex + 1
					cc.Timespans = append(cc.Timespans, nil)
					copy(cc.Timespans[idx+1:], cc.Timespans[idx:])
					cc.Timespans[idx] = newTs
					newTs.createIncrementsSlice()
					tsWasSplited = true
					break
				}
			}
			// debit monetary
			for _, b := range usefulMoneyBalances {
				// check standard subject tags
				if b.RateSubject == "" {
					amount := increment.Cost
					if b.Value >= amount {
						b.Value -= amount
						increment.BalanceUuid = b.Uuid
						paid = true
						if count {
							ub.countUnits(&Action{BalanceId: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
						}
						break
					}
				} else {
					// get the new rate
					cd := cc.CreateCallDescriptor()
					cd.TimeStart = ts.GetTimeStartForIncrement(incrementIndex, increment)
					cd.TimeEnd = cc.Timespans[len(cc.Timespans)-1].TimeEnd
					cd.CallDuration = cc.Timespans[len(cc.Timespans)-1].CallDuration
					newCC, err := b.GetCost(cd)
					if err != nil {
						Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
						continue
					}
					//debit new callcost
					for _, nts := range newCC.Timespans {
						for _, nIncrement := range nts.Increments {
							_ = nIncrement
						}
					}
				}
			}
			if !paid {
				// no balance was attached to this increment: cut the rest of increments/timespans
				ts.SplitByIncrement(incrementIndex, increment)
				if len(ts.Increments) == 0 {
					// if there are no increments left in the ts leav it out
					cc.Timespans = cc.Timespans[:tsIndex]
				} else {
					cc.Timespans = cc.Timespans[:tsIndex+1]
				}
				return nil
			}
		}
	}

	return nil
}

func (ub *UserBalance) refoundIncrements(increments Increments, count bool) {
	for _, increment := range increments {
		var balance *Balance
		for _, balanceChain := range ub.BalanceMap {
			if balance = balanceChain.GetBalance(increment.BalanceUuid); balance != nil {
				break
			}
		}
		if balance != nil {
			balance.Value += increment.Cost
			if count {
				ub.countUnits(&Action{BalanceId: increment.BalanceType, Direction: OUTBOUND, Balance: &Balance{Value: increment.Cost}})
			}
		} else {
			// TODO: where should put the money?
		}
	}
}

/*
Debits some amount of user's specified balance. Returns the remaining credit in user's balance.
*/
func (ub *UserBalance) debitGenericBalance(balanceId string, amount float64, count bool) float64 {
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
