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
	"testing"
)

func TestUnitsCounterAddBalance(t *testing.T) {
	uc := &UnitsCounter{
		Direction:      OUTBOUND,
		BalanceId:      SMS,
		Units:          100,
		MinuteBalances: []*Balance{&Balance{Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}},
	}
	uc.addMinutes(20, "test")
	if len(uc.MinuteBalances) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestUnitsCounterAddBalanceExists(t *testing.T) {
	uc := &UnitsCounter{
		Direction:      OUTBOUND,
		BalanceId:      SMS,
		Units:          100,
		MinuteBalances: []*Balance{&Balance{Value: 10, Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}},
	}
	uc.addMinutes(5, "0723")
	if len(uc.MinuteBalances) != 2 || uc.MinuteBalances[0].Value != 15 {
		t.Error("Error adding minute bucket!")
	}
}
