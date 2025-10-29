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
package migrator

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestV1AccountAsAccount(t *testing.T) {
	d1b := &v1Balance{
		Value:          100000,
		Weight:         10,
		DestinationIds: "NAT",
		Timings: []*engine.RITiming{
			{
				StartTime: "00:00:00",
			},
		},
	}
	v1b := &v1Balance{
		Value:          100000,
		Weight:         10,
		DestinationIds: "NAT",
		Timings: []*engine.RITiming{
			{
				StartTime: "00:00:00",
			},
		},
	}
	v1Acc := &v1Account{
		Id: "*OUT:CUSTOMER_1:rif",
		BalanceMap: map[string]v1BalanceChain{
			utils.DATA:  {d1b},
			utils.VOICE: {v1b},
			utils.MONETARY: {&v1Balance{
				Value: 21,
				Timings: []*engine.RITiming{
					{
						StartTime: "00:00:00",
					},
				},
			}},
		},
	}
	d2 := &engine.Balance{
		Uuid:           "",
		ID:             "",
		Value:          100000,
		Weight:         10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "",
		Categories:     utils.NewStringMap(""),
		SharedGroups:   utils.NewStringMap(""),
		Timings:        []*engine.RITiming{{StartTime: "00:00:00"}},
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{},
	}
	v2 := &engine.Balance{
		Uuid:           "",
		ID:             "",
		Value:          0.0001,
		Weight:         10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "",
		Categories:     utils.NewStringMap(""),
		SharedGroups:   utils.NewStringMap(""),
		Timings:        []*engine.RITiming{{StartTime: "00:00:00"}},
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{},
	}
	m2 := &engine.Balance{
		Uuid:           "",
		ID:             "",
		Value:          21,
		DestinationIDs: utils.NewStringMap(""),
		RatingSubject:  "",
		Categories:     utils.NewStringMap(""),
		SharedGroups:   utils.NewStringMap(""),
		Timings:        []*engine.RITiming{{StartTime: "00:00:00"}},
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{},
	}
	testAccount := &engine.Account{
		ID: "CUSTOMER_1:rif",
		BalanceMap: map[string]engine.Balances{
			utils.DATA:     {d2},
			utils.VOICE:    {v2},
			utils.MONETARY: {m2},
		},
		UnitCounters:   engine.UnitCounters{},
		ActionTriggers: engine.ActionTriggers{},
	}
	if def := v1b.IsDefault(); def != false {
		t.Errorf("Expecting: false, received: true")
	}
	newAcc := v1Acc.V1toV3Account()
	if !reflect.DeepEqual(testAccount.BalanceMap["*monetary"][0], newAcc.BalanceMap["*monetary"][0]) {
		t.Errorf("Expecting: %+v, received: %+v", testAccount.BalanceMap["*monetary"][0], newAcc.BalanceMap["*monetary"][0])
	} else if !reflect.DeepEqual(testAccount.BalanceMap["*voice"][0], newAcc.BalanceMap["*voice"][0]) {
		t.Errorf("Expecting: %+v, received: %+v", testAccount.BalanceMap["*voice"][0], newAcc.BalanceMap["*voice"][0])
	} else if !reflect.DeepEqual(testAccount.BalanceMap["*data"][0], newAcc.BalanceMap["*data"][0]) {
		t.Errorf("Expecting: %+v, received: %+v", testAccount.BalanceMap["*data"][0], newAcc.BalanceMap["*data"][0])
	}
}
func TestV2toV3Account(t *testing.T) {
	v2Acc := &v2Account{
		ID: "account1",
		BalanceMap: map[string]engine.Balances{
			"balance1": {
				{},
			},
		},
		UnitCounters: engine.UnitCounters{
			"counter1": {
				&engine.UnitCounter{
					Counters: []*engine.CounterFilter{
						{
							Value:  100.0,
							Filter: &engine.BalanceFilter{},
						},
					},
				},
			},
		},
		ActionTriggers: engine.ActionTriggers{},
		AllowNegative:  false,
		Disabled:       false,
	}
	v3Acc := v2Acc.V2toV3Account()
	if v3Acc.ID != v2Acc.ID {
		t.Errorf("expected ID %s, got %s", v2Acc.ID, v3Acc.ID)
	}
	for key, v2Balances := range v2Acc.BalanceMap {
		v3Balances, exists := v3Acc.BalanceMap[key]
		if !exists {
			t.Errorf("balance map key %s not found in v3 account", key)
			continue
		}
		if len(v3Balances) != len(v2Balances) {
			t.Errorf("expected %d balances for key %s, got %d", len(v2Balances), key, len(v3Balances))
			continue
		}
		for i, v2Bal := range v2Balances {
			v3Bal := v3Balances[i]
			if v3Bal.ID != v2Bal.ID || v3Bal.Value != v2Bal.Value {
				t.Errorf("expected balance %+v, got %+v", v2Bal, v3Bal)
			}
		}
	}

	for key, v2UnitCounters := range v2Acc.UnitCounters {
		v3UnitCounters, exists := v3Acc.UnitCounters[key]
		if !exists {
			t.Errorf("unit counters key %s not found in v3 account", key)
			continue
		}
		if len(v3UnitCounters) != len(v2UnitCounters) {
			t.Errorf("expected %d unit counters for key %s, got %d", len(v2UnitCounters), key, len(v3UnitCounters))
			continue
		}
		for i, v2Uc := range v2UnitCounters {
			v3Uc := v3UnitCounters[i]
			if len(v3Uc.Counters) != len(v2Uc.Counters) {
				t.Errorf("expected %d counters for unit counter %d, got %d", len(v2Uc.Counters), i, len(v3Uc.Counters))
			}
		}
	}

}
