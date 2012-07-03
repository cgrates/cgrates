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
		"LOG":            logAction,
		"RESET_TRIGGERS": resetTriggersAction,
		"SET_POSTPAID":   setPostpaidAction,
		"RESET_POSTPAID": resetPostpaidAction,
		"SET_PREPAID":    setPrepaidAction,
		"RESET_PREPAID":  resetPrepaidAction,
		"TOPUP_RESET":    topupResetAction,
		"TOPUP":          topupAction,
		"DEBIT":          debitAction,
		"RESET_COUNTERS": resetCountersAction,
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
	ub.Type = UB_TYPE_POSTPAID
	return
}

func setPrepaidAction(ub *UserBalance, a *Action) (err error) {
	ub.Type = UB_TYPE_PREPAID
	return
}

func resetPrepaidAction(ub *UserBalance, a *Action) (err error) {
	ub.Type = UB_TYPE_PREPAID
	return
}

func topupResetAction(ub *UserBalance, a *Action) (err error) {
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]float64)
	}
	ub.BalanceMap[a.BalanceId] = a.Units
	return
}

func topupAction(ub *UserBalance, a *Action) (err error) {
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]float64)
	}
	ub.BalanceMap[a.BalanceId] += a.Units
	ub.addMinuteBucket(a.MinuteBucket)
	return
}

func debitAction(ub *UserBalance, a *Action) (err error) {
	return
}

func resetCountersAction(ub *UserBalance, a *Action) (err error) {
	//ub.UnitsCounters
	return
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
