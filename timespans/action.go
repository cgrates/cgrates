/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"log"
	"sort"
	"strconv"
	"strings"
)

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type Action struct {
	ActionType   string
	BalanceId    string
	Units        float64
	Weight       float64
	MinuteBucket *MinuteBucket
}

type actionTypeFunc func(*UserBalance, *Action) error

var (
	actionTypeFuncMap = map[string]actionTypeFunc{
		"LOG":                logAction,
		"RESET_TRIGGERS":     resetTriggersAction,
		"SET_POSTPAID":       setPostpaidAction,
		"RESET_POSTPAID":     resetPostpaidAction,
		"SET_PREPAID":        setPrepaidAction,
		"RESET_PREPAID":      resetPrepaidAction,
		"TOPUP_RESET":        topupResetAction,
		"TOPUP":              topupAction,
		"DEBIT":              debitAction,
		"RESET_COUNTER":      resetCounterAction,
		"RESET_ALL_COUNTERS": resetAllCountersAction,
	}
)

func logAction(ub *UserBalance, a *Action) (err error) {
	log.Printf("%v %v %v", a.BalanceId, a.Units, a.MinuteBucket)
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
		ub.BalanceMap[a.BalanceId] = 0
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
		uc = &UnitsCounter{BalanceId: MINUTES}
		ub.UnitCounters = append(ub.UnitCounters, uc)
	}
	if a.BalanceId == MINUTES {
		uc.initMinuteBuckets(ub.ActionTriggers)
	} else {
		uc.Units = 0
	}
	return
}

func resetAllCountersAction(ub *UserBalance, a *Action) (err error) {
	ub.UnitCounters = make([]*UnitsCounter, 0)
	uc := &UnitsCounter{BalanceId: MINUTES}
	uc.initMinuteBuckets(ub.ActionTriggers)
	ub.UnitCounters = append(ub.UnitCounters, uc)
	return
}

func genericMakeNegative(a *Action) {
	a.Units = -a.Units
	if a.MinuteBucket != nil {
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
type ActionPriotityList []*Action

func (apl ActionPriotityList) Len() int {
	return len(apl)
}

func (apl ActionPriotityList) Swap(i, j int) {
	apl[i], apl[j] = apl[j], apl[i]
}

func (apl ActionPriotityList) Less(i, j int) bool {
	return apl[i].Weight < apl[j].Weight
}

func (apl ActionPriotityList) Sort() {
	sort.Sort(apl)
}

/*
Serializes the action for the storage. Used for key-value storages.
*/
func (a *Action) store() (result string) {
	result += a.ActionType + "|"
	result += a.BalanceId + "|"
	result += strconv.FormatFloat(a.Units, 'f', -1, 64) + "|"
	result += strconv.FormatFloat(a.Weight, 'f', -1, 64)
	if a.MinuteBucket != nil {
		result += "|"
		result += a.MinuteBucket.store()
	}
	return
}

/*
De-serializes the action for the storage. Used for key-value storages.
*/
func (a *Action) restore(input string) {
	elements := strings.Split(input, "|")
	a.ActionType = elements[0]
	a.BalanceId = elements[1]
	a.Units, _ = strconv.ParseFloat(elements[2], 64)
	a.Weight, _ = strconv.ParseFloat(elements[3], 64)
	if len(elements) == 5 {
		a.MinuteBucket = &MinuteBucket{}
		a.MinuteBucket.restore(elements[4])
	}
}
