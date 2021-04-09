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

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestUnitCounterClone(t *testing.T) {
	var uc *UnitCounter
	if rcv := uc.Clone(); rcv != nil {
		t.Errorf("Expecting nil, received: %s", utils.ToJSON(rcv))
	}
	uc = &UnitCounter{}
	eOut := &UnitCounter{}
	if rcv := uc.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	uc = &UnitCounter{
		CounterType: "testCounterType",
		Counters: []*CounterFilter{
			{Value: 0.7},
			{Value: 0.8},
		},
	}
	eOut = &UnitCounter{
		CounterType: "testCounterType",
		Counters: []*CounterFilter{
			{Value: 0.7},
			{Value: 0.8},
		},
	}
	if rcv := uc.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestCounterFilterClone(t *testing.T) {
	var cfs *CounterFilter
	if rcv := cfs.Clone(); rcv != nil {
		t.Errorf("Expecting nil, received: %s", utils.ToJSON(rcv))
	}
	cfs = &CounterFilter{}
	eOut := &CounterFilter{}
	if rcv := cfs.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	cfs = &CounterFilter{
		Value: 0.7,
		Filter: &BalanceFilter{
			Uuid: utils.StringPointer("testUuid"),
		},
	}
	eOut = &CounterFilter{
		Value: 0.7,
		Filter: &BalanceFilter{
			Uuid: utils.StringPointer("testUuid"),
		},
	}
	if rcv := cfs.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestUnitCountersClone(t *testing.T) {
	var ucs UnitCounters
	if rcv := ucs.Clone(); rcv != nil {
		t.Errorf("Expecting nil, received: %s", utils.ToJSON(rcv))
	}
	ucs = UnitCounters{}
	eOut := UnitCounters{}
	if rcv := ucs.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	ucs = UnitCounters{
		"string1": []*UnitCounter{
			{
				CounterType: "testCounterType1.1",
			},
			{
				CounterType: "testCounterType1.2",
			},
		},
		"string2": []*UnitCounter{
			{
				CounterType: "testCounterType2.1",
			},
			{
				CounterType: "testCounterType2.2",
			},
		},
	}
	eOut = UnitCounters{
		"string1": []*UnitCounter{
			{
				CounterType: "testCounterType1.1",
			},
			{
				CounterType: "testCounterType1.2",
			},
		},
		"string2": []*UnitCounter{
			{
				CounterType: "testCounterType2.1",
			},
			{
				CounterType: "testCounterType2.2",
			},
		},
	}
	rcv := ucs.Clone()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	rcv["string1"][0].CounterType = "modified"
	if ucs["string1"][0].CounterType == "modified" {
		t.Error("Original struct was modified")
	}
}

func TestUnitsCounterAddBalance(t *testing.T) {
	uc := &UnitCounter{
		Counters: CounterFilters{&CounterFilter{Value: 1},
			&CounterFilter{Filter: &BalanceFilter{Weight: utils.Float64Pointer(20),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT"))}},
			&CounterFilter{Filter: &BalanceFilter{Weight: utils.Float64Pointer(10),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET"))}}},
	}
	UnitCounters{utils.MetaSMS: []*UnitCounter{uc}}.addUnits(20, utils.MetaSMS,
		&CallCost{Destination: "test"}, nil)
	if len(uc.Counters) != 3 {
		t.Error("Error adding minute bucket: ", uc.Counters)
	}
}

func TestUnitsCounterAddBalanceExists(t *testing.T) {
	uc := &UnitCounter{
		Counters: CounterFilters{&CounterFilter{Value: 1}, &CounterFilter{Value: 10,
			Filter: &BalanceFilter{Weight: utils.Float64Pointer(20),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT"))}},
			&CounterFilter{Filter: &BalanceFilter{Weight: utils.Float64Pointer(10),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET"))}}},
	}
	UnitCounters{utils.MetaSMS: []*UnitCounter{uc}}.addUnits(5,
		utils.MetaSMS, &CallCost{Destination: "0723"}, nil)
	if len(uc.Counters) != 3 || uc.Counters[1].Value != 15 {
		t.Error("Error adding minute bucket!")
	}
}

func TestUnitCountersCountAllMonetary(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TriggerMaxBalance,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.MetaMonetary, &CallCost{}, nil)

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaMonetary][0].Counters) != 2 ||
		a.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 10 ||
		a.UnitCounters[utils.MetaMonetary][0].Counters[1].Value != 10 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error Initializing adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersCountAllMonetaryId(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(20),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TriggerMaxBalance,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.MetaMonetary, nil, &Balance{Weight: 20})
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaMonetary][0].Counters) != 2 ||
		a.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 0 ||
		a.UnitCounters[utils.MetaMonetary][0].Counters[1].Value != 10 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersCountAllVoiceDestinationEvent(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(20),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:           utils.StringPointer(utils.MetaVoice),
					DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
					Weight:         utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR22",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:           utils.StringPointer(utils.MetaVoice),
					DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET")),
					Weight:         utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TriggerMaxBalance,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.MetaVoice, &CallCost{Destination: "0723045326"}, nil)

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaVoice][0].Counters) != 2 ||
		a.UnitCounters[utils.MetaVoice][0].Counters[0].Value != 10 ||
		a.UnitCounters[utils.MetaVoice][0].Counters[1].Value != 10 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersKeepValuesAfterInit(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(20),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:           utils.StringPointer(utils.MetaVoice),
					DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
					Weight:         utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR22",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:           utils.StringPointer(utils.MetaVoice),
					DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET")),
					Weight:         utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TriggerMaxBalance,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.MetaVoice, &CallCost{Destination: "0723045326"}, nil)

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaVoice][0].Counters) != 2 ||
		a.UnitCounters[utils.MetaVoice][0].Counters[0].Value != 10 ||
		a.UnitCounters[utils.MetaVoice][0].Counters[1].Value != 10 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error adding unit counters: %v", len(a.UnitCounters))
	}
	a.InitCounters()

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaVoice][0].Counters) != 2 ||
		a.UnitCounters[utils.MetaVoice][0].Counters[0].Value != 10 ||
		a.UnitCounters[utils.MetaVoice][0].Counters[1].Value != 10 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error keeping counter values after init: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersResetCounterById(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TriggerMaxBalance,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(10, utils.MetaMonetary, &CallCost{}, nil)

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaMonetary][0].Counters) != 2 ||
		a.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 10 ||
		a.UnitCounters[utils.MetaMonetary][0].Counters[1].Value != 10 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error Initializing adding unit counters: %v", len(a.UnitCounters))
	}
	a.UnitCounters.resetCounters(&Action{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
			ID:   utils.StringPointer("TestTR11"),
		},
	})
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaMonetary][0].Counters) != 2 ||
		a.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 10 ||
		a.UnitCounters[utils.MetaMonetary][0].Counters[1].Value != 0 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error Initializing adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCounterHasCounter(t *testing.T) {
	cfs := CounterFilters{
		{
			Value: 15,
			Filter: &BalanceFilter{
				ID:     utils.StringPointer("testID1"),
				Type:   utils.StringPointer("testType1"),
				Weight: utils.Float64Pointer(15),
			},
		},
		{
			Value: 10,
			Filter: &BalanceFilter{
				ID:     utils.StringPointer("testID2"),
				Type:   utils.StringPointer("testType2"),
				Weight: utils.Float64Pointer(10),
			},
		},
	}
	cf := &CounterFilter{
		Value: 10,
		Filter: &BalanceFilter{
			ID:     utils.StringPointer("testID2"),
			Type:   utils.StringPointer("testType2"),
			Weight: utils.Float64Pointer(10),
		},
	}

	rcv := cfs.HasCounter(cf)

	if rcv != true {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", true, rcv)
	}
}

func TestUnitsCounteraddUnits(t *testing.T) {
	var uc *UnitCounter
	ucs := UnitCounters{
		"kind": []*UnitCounter{
			{
				Counters: CounterFilters{
					{
						Value: 10,
					},
					{
						Value: 15,
					},
				},
				CounterType: "",
			},
			uc,
			{
				Counters: CounterFilters{
					{
						Value: 20,
					},
					{
						Value: 25,
					},
				},
				CounterType: utils.MetaBalance,
			},
		},
	}
	cc := &CallCost{}
	b := &Balance{}

	exp := []*UnitCounter{
		{
			Counters: CounterFilters{
				{
					Value: 15,
				},
				{
					Value: 20,
				},
			},
			CounterType: utils.MetaCounterEvent,
		},
		uc,
		{
			Counters: CounterFilters{
				{
					Value: 25,
				},
				{
					Value: 30,
				},
			},
			CounterType: utils.MetaBalance,
		},
	}
	ucs.addUnits(5, "kind", cc, b)

	if !reflect.DeepEqual(ucs["kind"], exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(ucs["kind"]))
	}
}

func TestUnitsCounterresetCounters(t *testing.T) {
	var emptyctr *UnitCounter
	ucs := UnitCounters{
		"kind": []*UnitCounter{
			{
				Counters: CounterFilters{
					{
						Value: 10,
						Filter: &BalanceFilter{
							ID:     utils.StringPointer("testID1"),
							Type:   utils.StringPointer("kind"),
							Weight: utils.Float64Pointer(10),
						},
					},

					{
						Value: 15,
						Filter: &BalanceFilter{
							ID:     utils.StringPointer("testID2"),
							Type:   utils.StringPointer("kind"),
							Weight: utils.Float64Pointer(15),
						},
					},
				},
				CounterType: "",
			},
			emptyctr,
			{
				Counters: CounterFilters{
					{
						Value: 20,
						Filter: &BalanceFilter{
							ID:     utils.StringPointer("testID2"),
							Type:   utils.StringPointer("kind"),
							Weight: utils.Float64Pointer(15),
						},
					},
					{
						Value: 25,
						Filter: &BalanceFilter{
							ID:     utils.StringPointer("testID1"),
							Type:   utils.StringPointer("kind"),
							Weight: utils.Float64Pointer(10),
						},
					},
				},
				CounterType: utils.MetaBalance,
			},
		},
	}
	a := &Action{
		Balance: &BalanceFilter{
			ID:     utils.StringPointer("testID1"),
			Type:   utils.StringPointer("kind"),
			Weight: utils.Float64Pointer(10),
		},
	}

	exp := []*UnitCounter{
		{
			Counters: CounterFilters{
				{
					Value: 0,
					Filter: &BalanceFilter{
						ID:     utils.StringPointer("testID1"),
						Type:   utils.StringPointer("kind"),
						Weight: utils.Float64Pointer(10),
					},
				},
				{
					Value: 15,
					Filter: &BalanceFilter{
						ID:     utils.StringPointer("testID2"),
						Type:   utils.StringPointer("kind"),
						Weight: utils.Float64Pointer(15),
					},
				},
			},
			CounterType: "",
		},
		emptyctr,
		{
			Counters: CounterFilters{
				{
					Value: 20,
					Filter: &BalanceFilter{
						ID:     utils.StringPointer("testID2"),
						Type:   utils.StringPointer("kind"),
						Weight: utils.Float64Pointer(15),
					},
				},
				{
					Value: 0,
					Filter: &BalanceFilter{
						ID:     utils.StringPointer("testID1"),
						Type:   utils.StringPointer("kind"),
						Weight: utils.Float64Pointer(10),
					},
				},
			},
			CounterType: utils.MetaBalance,
		},
	}
	ucs.resetCounters(a)

	if !reflect.DeepEqual(ucs["kind"], exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(ucs["kind"]))
	}
}
