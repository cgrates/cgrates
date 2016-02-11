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

import "github.com/cgrates/cgrates/utils"

// Amount of a trafic of a certain type
type UnitCounter struct {
	BalanceType string       // *monetary/*voice/*sms/etc
	CounterType string       // *event or *balance
	Balances    BalanceChain // first balance is the general one (no destination)
}

// Returns true if the counters were of the same type
// Copies the value from old balances
func (uc *UnitCounter) CopyCounterValues(oldUc *UnitCounter) bool {
	if uc.BalanceType+uc.CounterType != oldUc.BalanceType+oldUc.CounterType { // type check
		return false
	}
	for _, b := range uc.Balances {
		for _, oldB := range oldUc.Balances {
			if b.Equal(oldB) {
				b.Value = oldB.Value
				break
			}
		}
	}
	return true
}

type UnitCounters []*UnitCounter

func (ucs UnitCounters) addUnits(amount float64, kind string, cc *CallCost, b *Balance) {
	for _, uc := range ucs {
		if uc == nil { // safeguard
			continue
		}
		if uc.BalanceType != kind {
			continue
		}
		if uc.CounterType == "" {
			uc.CounterType = utils.COUNTER_EVENT
		}
		for _, bal := range uc.Balances {
			if uc.CounterType == utils.COUNTER_EVENT && cc != nil && bal.MatchCCFilter(cc) {
				bal.AddValue(amount)
				continue
			}
			bp := &BalancePointer{}
			bp.LoadFromBalance(bal)
			if uc.CounterType == utils.COUNTER_BALANCE && b != nil && b.MatchFilter(bp, true) {
				bal.AddValue(amount)
				continue
			}
		}

	}
}

func (ucs UnitCounters) resetCounters(a *Action) {
	for _, uc := range ucs {
		if uc == nil { // safeguard
			continue
		}
		if a != nil && a.Balance.Type != nil && a.Balance.GetType() != uc.BalanceType {
			continue
		}
		for _, b := range uc.Balances {
			if a == nil || a.Balance == nil || b.MatchFilter(a.Balance, false) {
				b.Value = 0
			}
		}
	}
}
