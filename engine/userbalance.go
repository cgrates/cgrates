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
			// TODO: improve this
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
		if b.DestinationId != "" && b.DestinationId != utils.ANY {
			dest, err := storageGetter.GetDestination(b.DestinationId, false)
			if err != nil {
				continue
			}
			precision := dest.containsPrefix(prefix)
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
		tsWasSplit := false
		for incrementIndex, increment := range ts.Increments {
			if tsWasSplit {
				break
			}
			paid := false
			for _, b := range usefulMinuteBalances {
				// check standard subject tags
				if b.RateSubject == ZEROSECOND || b.RateSubject == "" {
					amount := increment.Duration.Seconds()
					if b.Value >= amount {
						b.Value -= amount
						increment.BalanceUuids = append(increment.BalanceUuids, b.Uuid)
						increment.MinuteInfo = &MinuteInfo{cc.Destination, amount, 0}
						paid = true
						if count {
							ub.countUnits(&Action{BalanceId: MINUTES, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
						}
						break
					}
					continue
				}
				if b.RateSubject == ZEROMINUTE {
					amount := time.Minute.Seconds()
					if b.Value >= amount { // balance has at least 60 seconds
						newTs := ts
						if incrementIndex != 0 {
							// if increment it's not at the begining we must split the timespan
							newTs = ts.SplitByIncrement(incrementIndex)
						}
						newTs.RoundToDuration(time.Minute)
						newTs.RateInterval = &RateInterval{
							Rating: &RIRate{
								Rates: RateGroups{
									&Rate{
										GroupIntervalStart: 0,
										Value:              0,
										RateIncrement:      time.Minute,
										RateUnit:           time.Minute,
									},
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
							tsWasSplit = true
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
						newTs.Increments[0].BalanceUuids = append(newTs.Increments[0].BalanceUuids, b.Uuid)
						newTs.Increments[0].MinuteInfo = &MinuteInfo{cc.Destination, amount, 0}
						paid = true
						if count {
							ub.countUnits(&Action{BalanceId: MINUTES, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
						}
						break
					}
					continue
				}

				// get the new rate
				cd := cc.CreateCallDescriptor()
				cd.Subject = b.RateSubject
				cd.TimeStart = ts.GetTimeStartForIncrement(incrementIndex)
				cd.TimeEnd = cc.Timespans[len(cc.Timespans)-1].TimeEnd
				cd.CallDuration = cc.Timespans[len(cc.Timespans)-1].CallDuration
				newCC, err := b.GetCost(cd)
				if err != nil {
					Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
					continue
				}
				//debit new callcost
				var paidTs []*TimeSpan
				for _, nts := range newCC.Timespans {
					nts.createIncrementsSlice()
					paidTs = append(paidTs, nts)
					for nIdx, nInc := range nts.Increments {
						// debit minutes and money
						seconds := nInc.Duration.Seconds()
						cost := nInc.Cost
						var moneyBal *Balance
						for _, mb := range moneyBalances {
							if mb.Value >= cost {
								mb.Value -= cost
								moneyBal = mb
							}
						}
						if moneyBal != nil && b.Value >= seconds {
							b.Value -= seconds
							nInc.BalanceUuids = append(nInc.BalanceUuids, b.Uuid)
							nInc.BalanceUuids = append(nInc.BalanceUuids, moneyBal.Uuid)
							nInc.MinuteInfo = &MinuteInfo{newCC.Destination, seconds, 0}
							paid = true
							if count {
								ub.countUnits(&Action{BalanceId: MINUTES, Direction: newCC.Direction, Balance: &Balance{Value: seconds, DestinationId: newCC.Destination}})
								ub.countUnits(&Action{BalanceId: CREDIT, Direction: newCC.Direction, Balance: &Balance{Value: cost, DestinationId: newCC.Destination}})
							}
						} else {
							paid = false
							nts.SplitByIncrement(nIdx)
						}
					}
				}

				// calculate overlaped timespans
				var paidDuration time.Duration
				for _, pts := range paidTs {
					paidDuration += pts.GetDuration()
				}
				if paidDuration > 0 {
					// split from current increment
					newTs := ts.SplitByIncrement(incrementIndex)
					var remainingTs []*TimeSpan
					if newTs != nil {
						remainingTs = append(remainingTs, newTs)
					} else {
						// nothing was paied form current ts so remove it
						cc.Timespans = append(cc.Timespans[:tsIndex], cc.Timespans[tsIndex+1:]...)
						tsIndex--
					}
					for tsi := tsIndex + 1; tsi < len(cc.Timespans); tsi++ {
						remainingTs = append(remainingTs, cc.Timespans[tsi])
					}
					for remainingIndex, rts := range remainingTs {
						if paidDuration >= rts.GetDuration() {
							paidDuration -= rts.GetDuration()
						} else {
							if paidDuration > 0 {
								// this ts was not fully paid
								fragment := rts.SplitByDuration(paidDuration)
								paidTs = append(paidTs, fragment)
							}
							// delete from tsIndex to current
							cc.Timespans = append(cc.Timespans[:tsIndex], cc.Timespans[remainingIndex:]...)
							break
						}
					}

					// append the timpespans to outer timespans
					for _, pts := range paidTs {
						tsIndex++
						cc.Timespans = append(cc.Timespans, nil)
						copy(cc.Timespans[tsIndex+1:], cc.Timespans[tsIndex:])
						cc.Timespans[tsIndex] = pts
					}
					paid = true
					tsWasSplit = true
				}
			}
			if paid {
				continue
			} else {
				// Split if some increments were processed by minutes
				if incrementIndex > 0 && ts.Increments[incrementIndex-1].MinuteInfo != nil {
					newTs := ts.SplitByIncrement(incrementIndex)
					if newTs != nil {
						idx := tsIndex + 1
						cc.Timespans = append(cc.Timespans, nil)
						copy(cc.Timespans[idx+1:], cc.Timespans[idx:])
						cc.Timespans[idx] = newTs
						newTs.createIncrementsSlice()
						tsWasSplit = true
					}
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
						increment.BalanceUuids = append(increment.BalanceUuids, b.Uuid)
						paid = true
						if count {
							ub.countUnits(&Action{BalanceId: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
						}
						break
					}
				} else {
					// get the new rate
					cd := cc.CreateCallDescriptor()
					cd.Subject = b.RateSubject
					cd.TimeStart = ts.GetTimeStartForIncrement(incrementIndex)
					cd.TimeEnd = cc.Timespans[len(cc.Timespans)-1].TimeEnd
					cd.CallDuration = cc.Timespans[len(cc.Timespans)-1].CallDuration
					newCC, err := b.GetCost(cd)
					if err != nil {
						Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
						continue
					}
					//debit new callcost
					var paidTs []*TimeSpan
					for _, nts := range newCC.Timespans {
						nts.createIncrementsSlice()
						paidTs = append(paidTs, nts)
						for nIdx, nInc := range nts.Increments {
							// debit money
							amount := nInc.Cost
							if b.Value >= amount {
								b.Value -= amount
								nInc.BalanceUuids = append(nInc.BalanceUuids, b.Uuid)
								if count {
									ub.countUnits(&Action{BalanceId: CREDIT, Direction: newCC.Direction, Balance: &Balance{Value: amount, DestinationId: newCC.Destination}})
								}
							} else {
								nts.SplitByIncrement(nIdx)
							}
						}
					}
					// calculate overlaped timespans
					var paidDuration time.Duration
					for _, pts := range paidTs {
						paidDuration += pts.GetDuration()
					}
					if paidDuration > 0 {
						// split from current increment
						newTs := ts.SplitByIncrement(incrementIndex)
						var remainingTs []*TimeSpan
						if newTs != nil {
							remainingTs = append(remainingTs, newTs)
						} else {
							// nothing was paied form current ts so remove it
							cc.Timespans = append(cc.Timespans[:tsIndex], cc.Timespans[tsIndex+1:]...)
							tsIndex--
						}

						for tsi := tsIndex + 1; tsi < len(cc.Timespans); tsi++ {
							remainingTs = append(remainingTs, cc.Timespans[tsi])
						}
						for remainingIndex, rts := range remainingTs {
							if paidDuration >= rts.GetDuration() {
								paidDuration -= rts.GetDuration()
							} else {
								if paidDuration > 0 {
									// this ts was not fully paid
									fragment := rts.SplitByDuration(paidDuration)
									paidTs = append(paidTs, fragment)
								}
								// delete from tsIndex to current
								cc.Timespans = append(cc.Timespans[:tsIndex], cc.Timespans[remainingIndex:]...)
								break
							}
						}

						// append the timpespans to outer timespans
						for _, pts := range paidTs {
							tsIndex++
							cc.Timespans = append(cc.Timespans, nil)
							copy(cc.Timespans[tsIndex+1:], cc.Timespans[tsIndex:])
							cc.Timespans[tsIndex] = pts
						}
						paid = true
						tsWasSplit = true
					}
				}
			}
			if !paid {
				// no balance was attached to this increment: cut the rest of increments/timespans
				if incrementIndex == 0 {
					// if we are right at the begining in the ts leave it out
					cc.Timespans = cc.Timespans[:tsIndex]
				} else {
					ts.SplitByIncrement(incrementIndex)
					cc.Timespans = cc.Timespans[:tsIndex+1]
				}
				return errors.New("Not enough credit")
				//return nil
			}
		}
	}

	return nil
}

func (ub *UserBalance) refoundIncrements(increments Increments, count bool) {
	for _, increment := range increments {
		var balance *Balance
		if increment.MinuteInfo != nil {
			if balance = ub.BalanceMap[MINUTES+OUTBOUND].GetBalance(increment.BalanceUuids[0]); balance != nil {
				break
			}
			if balance != nil {
				balance.Value += increment.Duration.Seconds()
				if count {
					ub.countUnits(&Action{BalanceId: MINUTES, Direction: OUTBOUND, Balance: &Balance{Value: -increment.Duration.Seconds()}})
				}
			} else {
				// TODO: where should put the minutes?
			}
		}
		// check money too
		if len(increment.BalanceUuids) == 2 && increment.BalanceUuids[1] != "" {
			if balance = ub.BalanceMap[CREDIT+OUTBOUND].GetBalance(increment.BalanceUuids[1]); balance != nil {
				break
			}
			if balance != nil {
				balance.Value += increment.Cost
				if count {
					ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: -increment.Cost}})
				}
			} else {
				// TODO: where should put the money?
			}
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
		acs, err := storageGetter.GetActions(at.ActionsId)
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
