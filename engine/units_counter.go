/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

// Amount of a trafic of a certain type
type UnitCounter struct {
	CounterType string         // *event or *balance
	Counters    CounterFilters // first balance is the general one (no destination)
}

type CounterFilter struct {
	Value  float64
	Filter *BalanceFilter
}

type CounterFilters []*CounterFilter

func (cfs CounterFilters) HasCounter(cf *CounterFilter) bool {
	for _, c := range cfs {
		if c.Filter.Equal(cf.Filter) {
			return true
		}
	}
	return false
}

// Clone clones *UnitCounter
func (uc *UnitCounter) Clone() (newUnit *UnitCounter) {
	if uc == nil {
		return
	}
	newUnit = &UnitCounter{
		CounterType: uc.CounterType,
	}
	if uc.Counters != nil {
		newUnit.Counters = make(CounterFilters, len(uc.Counters))
		for i, counter := range uc.Counters {
			newUnit.Counters[i] = counter.Clone()
		}
	}
	return newUnit
}

// Clone clones *CounterFilter
func (cfs *CounterFilter) Clone() *CounterFilter {
	if cfs == nil {
		return nil
	}
	return &CounterFilter{
		Value:  cfs.Value,
		Filter: cfs.Filter.Clone(),
	}
}

// Clone clones UnitCounters
func (ucs UnitCounters) Clone() UnitCounters {
	if ucs == nil {
		return nil
	}
	newUnitCounters := make(UnitCounters)
	for key, unitCounter := range ucs {
		newal := make([]*UnitCounter, len(unitCounter))
		for i, uc := range unitCounter {
			newal[i] = uc.Clone()
		}
		newUnitCounters[key] = newal
	}
	return newUnitCounters
}

// Returns true if the counters were of the same type
// Copies the value from old balances
func (uc *UnitCounter) CopyCounterValues(oldUc *UnitCounter) bool {
	if uc.CounterType != oldUc.CounterType { // type check
		return false
	}
	for _, c := range uc.Counters {
		for _, oldC := range oldUc.Counters {
			if c.Filter.Equal(oldC.Filter) {
				c.Value = oldC.Value
				break
			}
		}
	}
	return true
}

type UnitCounters map[string][]*UnitCounter
