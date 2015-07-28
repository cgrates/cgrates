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
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestUnitsCounterAddBalance(t *testing.T) {
	uc := &UnitsCounter{
		Direction:   OUTBOUND,
		BalanceType: utils.SMS,
		Balances:    BalanceChain{&Balance{Value: 1}, &Balance{Weight: 20, DestinationIds: "NAT"}, &Balance{Weight: 10, DestinationIds: "RET"}},
	}
	uc.addUnits(20, "test")
	if len(uc.Balances) != 3 {
		t.Error("Error adding minute bucket: ", uc.Balances)
	}
}

func TestUnitsCounterAddBalanceExists(t *testing.T) {
	uc := &UnitsCounter{
		Direction:   OUTBOUND,
		BalanceType: utils.SMS,
		Balances:    BalanceChain{&Balance{Value: 1}, &Balance{Value: 10, Weight: 20, DestinationIds: "NAT"}, &Balance{Weight: 10, DestinationIds: "RET"}},
	}
	uc.addUnits(5, "0723")
	if len(uc.Balances) != 3 || uc.Balances[1].GetValue() != 15 {
		t.Error("Error adding minute bucket!")
	}
}
