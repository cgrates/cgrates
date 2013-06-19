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

package rater

import (
	"fmt"
	"sort"
)

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type Action struct {
	Id           string
	ActionType   string
	BalanceId    string
	Direction    string
	Units        float64
	Weight       float64
	MinuteBucket *MinuteBucket
}

const (
	LOG            = "LOG"
	RESET_TRIGGERS = "RESET_TRIGGERS"
	SET_POSTPAID   = "SET_POSTPAID"
	RESET_POSTPAID = "RESET_POSTPAID"
	SET_PREPAID    = "SET_PREPAID"
	RESET_PREPAID  = "RESET_PREPAID"
	TOPUP_RESET    = "TOPUP_RESET"
	TOPUP          = "TOPUP"
	DEBIT          = "DEBIT"
	RESET_COUNTER  = "RESET_COUNTER"
	RESET_COUNTERS = "RESET_COUNTERS"
)

type actionTypeFunc func(*UserBalance, *Action) error

func getActionFunc(typ string) (actionTypeFunc, bool) {
	switch typ {
	case LOG:
		return logAction, true
	case RESET_TRIGGERS:
		return resetTriggersAction, true
	case SET_POSTPAID:
		return setPostpaidAction, true
	case RESET_POSTPAID:
		return resetPostpaidAction, true
	case SET_PREPAID:
		return setPrepaidAction, true
	case RESET_PREPAID:
		return resetPrepaidAction, true
	case TOPUP_RESET:
		return topupResetAction, true
	case TOPUP:
		return topupAction, true
	case DEBIT:
		return debitAction, true
	case RESET_COUNTER:
		return resetCounterAction, true
	case RESET_COUNTERS:
		return resetCountersAction, true
	}
	return nil, false
}

func logAction(ub *UserBalance, a *Action) (err error) {
	Logger.Info(fmt.Sprintf("%v %v %v", a.BalanceId, a.Units, a.MinuteBucket))
	return
}

func resetTriggersAction(ub *UserBalance, a *Action) (err error) {
	ub.resetActionTriggers()
	return
}

func setPostpaidAction(ub *UserBalance, a *Action) (err error) {
	ub.Type = UB_TYPE_POSTPAID
	return
}

func resetPostpaidAction(ub *UserBalance, a *Action) (err error) {
	genericReset(ub)
	return setPostpaidAction(ub, a)
}

func setPrepaidAction(ub *UserBalance, a *Action) (err error) {
	ub.Type = UB_TYPE_PREPAID
	return
}

func resetPrepaidAction(ub *UserBalance, a *Action) (err error) {
	genericReset(ub)
	return setPrepaidAction(ub, a)
}

func topupResetAction(ub *UserBalance, a *Action) (err error) {
	if a.BalanceId == MINUTES {
		ub.MinuteBuckets = make([]*MinuteBucket, 0)
	} else {
		ub.BalanceMap[a.BalanceId+a.Direction] = 0
	}
	genericMakeNegative(a)
	genericDebit(ub, a)
	return
}

func topupAction(ub *UserBalance, a *Action) (err error) {
	genericMakeNegative(a)
	genericDebit(ub, a)
	return
}

func debitAction(ub *UserBalance, a *Action) (err error) {
	return genericDebit(ub, a)
}

func resetCounterAction(ub *UserBalance, a *Action) (err error) {
	uc := ub.getUnitCounter(a)
	if uc == nil {
		uc = &UnitsCounter{BalanceId: MINUTES, Direction: a.Direction}
		ub.UnitCounters = append(ub.UnitCounters, uc)
	}
	if a.BalanceId == MINUTES {
		uc.initMinuteBuckets(ub.ActionTriggers)
	} else {
		uc.Units = 0
	}
	return
}

func resetCountersAction(ub *UserBalance, a *Action) (err error) {
	ub.UnitCounters = make([]*UnitsCounter, 0)
	ub.initMinuteCounters()
	return
}

func genericMakeNegative(a *Action) {
	if a.Units > 0 { // only apply if not allready negative
		a.Units = -a.Units
	}
	if a.MinuteBucket != nil && a.MinuteBucket.Seconds > 0 {
		a.MinuteBucket.Seconds = -a.MinuteBucket.Seconds
	}
}

func genericDebit(ub *UserBalance, a *Action) (err error) {
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]float64)
	}
	switch a.BalanceId {
	case CREDIT:
		ub.debitBalance(CREDIT, a.Units, false)
	case SMS:
		ub.debitBalance(SMS, a.Units, false)
	case MINUTES:
		ub.debitMinuteBucket(a.MinuteBucket)
	case TRAFFIC:
		ub.debitBalance(TRAFFIC, a.Units, false)
	case TRAFFIC_TIME:
		ub.debitBalance(TRAFFIC_TIME, a.Units, false)
	}
	return
}

func genericReset(ub *UserBalance) {
	for k, _ := range ub.BalanceMap {
		ub.BalanceMap[k] = 0
	}
	ub.MinuteBuckets = make([]*MinuteBucket, 0)
	ub.UnitCounters = make([]*UnitsCounter, 0)
	ub.resetActionTriggers()
}

// Structure to store actions according to weight
type Actions []*Action

func (apl Actions) Len() int {
	return len(apl)
}

func (apl Actions) Swap(i, j int) {
	apl[i], apl[j] = apl[j], apl[i]
}

func (apl Actions) Less(i, j int) bool {
	return apl[i].Weight < apl[j].Weight
}

func (apl Actions) Sort() {
	sort.Sort(apl)
}
