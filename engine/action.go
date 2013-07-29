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
	"fmt"
	"sort"
	"time"
)

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type Action struct {
	Id             string
	ActionType     string
	BalanceId      string
	Direction      string
	ExpirationDate time.Time
	Units          float64
	Weight         float64
	MinuteBucket   *MinuteBucket
}

const (
	LOG            = "*log"
	RESET_TRIGGERS = "*reset_triggers"
	SET_POSTPAID   = "*set_postpaid"
	RESET_POSTPAID = "*reset_postpaid"
	SET_PREPAID    = "*set_prepaid"
	RESET_PREPAID  = "*reset_prepaid"
	TOPUP_RESET    = "*topup_reset"
	TOPUP          = "*topup"
	DEBIT          = "*debit"
	RESET_COUNTER  = "*reset_counter"
	RESET_COUNTERS = "*reset_counters"
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
	ub.resetActionTriggers(a)
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
		ub.BalanceMap[a.BalanceId+a.Direction] = BalanceChain{&Balance{Value: 0}} // ToDo: can ub be empty here?
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
		ub.BalanceMap = make(map[string]BalanceChain)
	}
	if a.BalanceId == MINUTES {
		ub.debitMinuteBucket(a.MinuteBucket)
	} else {
		ub.debitBalanceAction(a)
	}
	return
}

func genericReset(ub *UserBalance) {
	for k, _ := range ub.BalanceMap {
		ub.BalanceMap[k] = BalanceChain{&Balance{Value: 0}}
	}
	ub.MinuteBuckets = make([]*MinuteBucket, 0)
	ub.UnitCounters = make([]*UnitsCounter, 0)
	ub.resetActionTriggers(nil)
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
