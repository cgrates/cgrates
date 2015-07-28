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

import (
	"strings"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
)

// Amount of a trafic of a certain type
type UnitsCounter struct {
	Direction   string
	BalanceType string
	//	Units     float64
	Balances BalanceChain // first balance is the general one (no destination)
}

// clears balances for this counter
// makes sure there are balances for all action triggers
func (uc *UnitsCounter) initBalances(ats []*ActionTrigger) {
	uc.Balances = BalanceChain{&Balance{}} // general balance
	for _, at := range ats {
		if !strings.Contains(at.ThresholdType, "counter") {
			// only get actions fo counter type action triggers
			continue
		}
		acs, err := ratingStorage.GetActions(at.ActionsId, false)
		if err != nil {
			continue
		}
		for _, a := range acs {
			if a.Balance != nil {
				b := a.Balance.Clone()
				b.SetValue(0)
				if !uc.Balances.HasBalance(b) {
					uc.Balances = append(uc.Balances, b)
				}
			}
		}
	}
	uc.Balances.Sort()
}

// returns the first balance that has no destination attached
func (uc *UnitsCounter) GetGeneralBalance() *Balance {
	if len(uc.Balances) == 0 { // general balance not present for some reson
		uc.Balances = append(uc.Balances, &Balance{})
	}
	return uc.Balances[0]
}

// Adds the units from the received balance to an existing balance if the destination
// is the same or ads the balance to the list if none matches.
func (uc *UnitsCounter) addUnits(amount float64, prefix string) {
	counted := false
	if prefix != "" {
		for _, mb := range uc.Balances {
			if !mb.HasDestination() {
				continue
			}
			for _, p := range utils.SplitPrefix(prefix, MIN_PREFIX_MATCH) {
				if x, err := cache2go.GetCached(utils.DESTINATION_PREFIX + p); err == nil {
					destIds := x.(map[interface{}]struct{})
					if _, found := destIds[mb.DestinationIds]; found {
						mb.AddValue(amount)
						counted = true
						break
					}
				}
				if counted {
					break
				}
			}
		}
	}
	if !counted {
		// use general balance
		b := uc.GetGeneralBalance()
		b.AddValue(amount)
	}
}

/*func (uc *UnitsCounter) String() string {
	return fmt.Sprintf("%s %s %v", uc.BalanceId, uc.Direction, uc.Units)
}*/
