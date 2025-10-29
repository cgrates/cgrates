/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"strconv"

	"github.com/cgrates/cgrates/utils"
)

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

func (ucs UnitCounters) addUnits(amount float64, kind string, cc *CallCost, b *Balance) {
	counters, found := ucs[kind]
	if !found {
		return
	}
	for _, uc := range counters {
		if uc == nil { // safeguard
			continue
		}
		if uc.CounterType == "" {
			uc.CounterType = utils.MetaCounterEvent
		}
		for _, c := range uc.Counters {
			if uc.CounterType == utils.MetaCounterEvent && cc != nil && cc.MatchCCFilter(c.Filter) {
				c.Value += amount
				continue
			}

			if uc.CounterType == utils.MetaBalance && b != nil && b.MatchFilter(c.Filter, "", true, false) {
				c.Value += amount
				continue
			}
		}

	}
}

func (ucs UnitCounters) resetCounters(a *Action) {
	for key, counters := range ucs {
		if a != nil && a.Balance.Type != nil && a.Balance.GetType() != key {
			continue
		}
		for _, c := range counters {
			if c == nil { // safeguard
				continue
			}
			for _, cf := range c.Counters {
				if a == nil || a.Balance == nil || cf.Filter.Equal(a.Balance) {
					cf.Value = 0
				}
			}
		}
	}
}

func (uc *UnitCounter) String() string {
	return utils.ToJSON(uc)
}

func (uc *UnitCounter) FieldAsInterface(fldPath []string) (val any, err error) {
	if uc == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		opath, indx := utils.GetPathIndex(fldPath[0])
		if opath == utils.Counters && indx != nil {
			if len(uc.Counters) <= *indx {
				return nil, utils.ErrNotFound
			}
			c := uc.Counters[*indx]
			if len(fldPath) == 1 {
				return c, nil
			}
			return c.FieldAsInterface(fldPath[1:])
		}
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.CounterType:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return uc.CounterType, nil
	case utils.Counters:
		if len(fldPath) == 1 {
			return uc.Counters, nil
		}
		var indx int
		if indx, err = strconv.Atoi(fldPath[1]); err != nil {
			return
		}
		if len(uc.Counters) <= indx {
			return nil, utils.ErrNotFound
		}
		c := uc.Counters[indx]
		if len(fldPath) == 2 {
			return c, nil
		}
		return c.FieldAsInterface(fldPath[2:])
	}
}

func (uc *UnitCounter) FieldAsString(fldPath []string) (val string, err error) {
	var iface any
	iface, err = uc.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

func (cfs *CounterFilter) String() string {
	return utils.ToJSON(cfs)
}

func (cfs *CounterFilter) FieldAsInterface(fldPath []string) (val any, err error) {
	if cfs == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.Value:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return cfs.Value, nil
	case utils.Filter:
		if len(fldPath) == 1 {
			return cfs.Filter, nil
		}
		return cfs.Filter.FieldAsInterface(fldPath[1:])
	}
}

func (cfs *CounterFilter) FieldAsString(fldPath []string) (val string, err error) {
	var iface any
	iface, err = cfs.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}
