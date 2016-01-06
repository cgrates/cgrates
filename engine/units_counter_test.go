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
	uc := &UnitCounter{
		BalanceType: utils.SMS,
		Balances:    BalanceChain{&Balance{Value: 1}, &Balance{Weight: 20, DestinationIds: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIds: utils.NewStringMap("RET")}},
	}
	UnitCounters{uc}.addUnits(20, utils.SMS, &CallCost{Destination: "test"}, nil)
	if len(uc.Balances) != 3 {
		t.Error("Error adding minute bucket: ", uc.Balances)
	}
}

func TestUnitsCounterAddBalanceExists(t *testing.T) {
	uc := &UnitCounter{
		BalanceType: utils.SMS,
		Balances:    BalanceChain{&Balance{Value: 1}, &Balance{Value: 10, Weight: 20, DestinationIds: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIds: utils.NewStringMap("RET")}},
	}
	UnitCounters{uc}.addUnits(5, utils.SMS, &CallCost{Destination: "0723"}, nil)
	if len(uc.Balances) != 3 || uc.Balances[1].GetValue() != 15 {
		t.Error("Error adding minute bucket!")
	}
}

func TestUnitCountersCountAllMonetary(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:          "TestTR1",
				ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:       utils.MONETARY,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR11",
				ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:       utils.MONETARY,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR2",
				ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:       utils.VOICE,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR3",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.VOICE,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR4",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.SMS,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR5",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE,
				BalanceType:       utils.SMS,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.MONETARY, &CallCost{}, nil)

	if len(a.UnitCounters) != 4 ||
		len(a.UnitCounters[0].Balances) != 2 ||
		a.UnitCounters[0].Balances[0].Value != 10 ||
		a.UnitCounters[0].Balances[1].Value != 10 {
		for _, uc := range a.UnitCounters {
			t.Logf("UC: %+v", uc)
			for _, b := range uc.Balances {
				t.Logf("B: %+v", b)
			}
		}
		t.Errorf("Error Initializing adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersCountAllMonetaryId(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:          "TestTR1",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.MONETARY,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR11",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.MONETARY,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     20,
			},
			&ActionTrigger{
				UniqueID:          "TestTR2",
				ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:       utils.VOICE,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR3",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.VOICE,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR4",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.SMS,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR5",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE,
				BalanceType:       utils.SMS,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.MONETARY, nil, &Balance{Weight: 20, Directions: utils.NewStringMap(utils.OUT)})

	if len(a.UnitCounters) != 4 ||
		len(a.UnitCounters[0].Balances) != 2 ||
		a.UnitCounters[0].Balances[0].Value != 0 ||
		a.UnitCounters[0].Balances[1].Value != 10 {
		for _, uc := range a.UnitCounters {
			t.Logf("UC: %+v", uc)
			for _, b := range uc.Balances {
				t.Logf("B: %+v", b)
			}
		}
		t.Errorf("Error adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersCountAllVoiceDestinationEvent(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:          "TestTR1",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.MONETARY,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR11",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.MONETARY,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     20,
			},
			&ActionTrigger{
				UniqueID:              "TestTR2",
				ThresholdType:         utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:           utils.VOICE,
				BalanceDirections:     utils.NewStringMap(utils.OUT),
				BalanceDestinationIds: utils.NewStringMap("NAT"),
				BalanceWeight:         10,
			},
			&ActionTrigger{
				UniqueID:              "TestTR22",
				ThresholdType:         utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:           utils.VOICE,
				BalanceDestinationIds: utils.NewStringMap("RET"),
				BalanceWeight:         10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR3",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.VOICE,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR4",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.SMS,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR5",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE,
				BalanceType:       utils.SMS,
				BalanceDirections: utils.NewStringMap(utils.OUT),
				BalanceWeight:     10,
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.VOICE, &CallCost{Destination: "0723045326"}, nil)

	if len(a.UnitCounters) != 4 ||
		len(a.UnitCounters[1].Balances) != 2 ||
		a.UnitCounters[1].Balances[0].Value != 10 ||
		a.UnitCounters[1].Balances[1].Value != 10 {
		for _, uc := range a.UnitCounters {
			t.Logf("UC: %+v", uc)
			for _, b := range uc.Balances {
				t.Logf("B: %+v", b)
			}
		}
		t.Errorf("Error adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersResetCounterById(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:          "TestTR1",
				ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:       utils.MONETARY,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR11",
				ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:       utils.MONETARY,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR2",
				ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
				BalanceType:       utils.VOICE,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR3",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.VOICE,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR4",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE_COUNTER,
				BalanceType:       utils.SMS,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
			&ActionTrigger{
				UniqueID:          "TestTR5",
				ThresholdType:     utils.TRIGGER_MAX_BALANCE,
				BalanceType:       utils.SMS,
				BalanceDirections: utils.NewStringMap(utils.OUT, utils.IN),
				BalanceWeight:     10,
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.MONETARY, &CallCost{}, nil)

	if len(a.UnitCounters) != 4 ||
		len(a.UnitCounters[0].Balances) != 2 ||
		a.UnitCounters[0].Balances[0].Value != 10 ||
		a.UnitCounters[0].Balances[1].Value != 10 {
		for _, uc := range a.UnitCounters {
			t.Logf("UC: %+v", uc)
			for _, b := range uc.Balances {
				t.Logf("B: %+v", b)
			}
		}
		t.Errorf("Error Initializing adding unit counters: %v", len(a.UnitCounters))
	}
	a.UnitCounters.resetCounters(&Action{
		BalanceType: utils.MONETARY,
		Balance: &Balance{
			Id: "TestTR11",
		},
	})
	if len(a.UnitCounters) != 4 ||
		len(a.UnitCounters[0].Balances) != 2 ||
		a.UnitCounters[0].Balances[0].Value != 10 ||
		a.UnitCounters[0].Balances[1].Value != 0 {
		for _, uc := range a.UnitCounters {
			t.Logf("UC: %+v", uc)
			for _, b := range uc.Balances {
				t.Logf("B: %+v", b)
			}
		}
		t.Errorf("Error Initializing adding unit counters: %v", len(a.UnitCounters))
	}
}
