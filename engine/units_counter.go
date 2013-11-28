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
)

// Amount of a trafic of a certain type
type UnitsCounter struct {
	Direction      string
	BalanceId      string
	Units          float64
	MinuteBalances BalanceChain
}

func (uc *UnitsCounter) initMinuteBalances(ats []*ActionTrigger) {
	uc.MinuteBalances = make(BalanceChain, 0)
	for _, at := range ats {
		acs, err := storageGetter.GetActions(at.ActionsId)
		if err != nil {
			continue
		}
		for _, a := range acs {
			if a.BalanceId == MINUTES && a.Balance != nil {
				b := a.Balance.Clone()
				b.Value = 0
				uc.MinuteBalances = append(uc.MinuteBalances, b)
			}
		}
	}
	uc.MinuteBalances.Sort()
}

// Adds the minutes from the received minute balance to an existing bucket if the destination
// is the same or ads the minute balance to the list if none matches.
func (uc *UnitsCounter) addMinutes(amount float64, prefix string) {
	for _, mb := range uc.MinuteBalances {
		dest, err := storageGetter.GetDestination(mb.DestinationId, false)
		if err != nil {
			Logger.Err(fmt.Sprintf("Minutes counter: unknown destination: %v", mb.DestinationId))
			continue
		}
		precision := dest.containsPrefix(prefix)
		if precision > 0 {
			mb.Value += amount
			break
		}
	}
}

func (uc *UnitsCounter) String() string {
	return fmt.Sprintf("%s %s %v", uc.BalanceId, uc.Direction, uc.Units)
}
