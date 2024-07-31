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
			utils.MetaData:  {d1b},
			utils.MetaVoice: {v1b},
			utils.MetaMonetary: {&v1Balance{
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
		Factors:        engine.ValueFactors{},
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
		Factors:        engine.ValueFactors{},
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
		Factors:        engine.ValueFactors{},
	}
	testAccount := &engine.Account{
		ID: "CUSTOMER_1:rif",
		BalanceMap: map[string]engine.Balances{
			utils.MetaData:     {d2},
			utils.MetaVoice:    {v2},
			utils.MetaMonetary: {m2},
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
	v2Acc := v2Account{
		ID:            "test-account",
		AllowNegative: true,
		Disabled:      false,
	}
	result := v2Acc.V2toV3Account()
	if result.ID != v2Acc.ID {
		t.Errorf("expected ID %v, got %v", v2Acc.ID, result.ID)
	}
	if result.AllowNegative != v2Acc.AllowNegative {
		t.Errorf("expected AllowNegative %v, got %v", v2Acc.AllowNegative, result.AllowNegative)
	}
	if result.Disabled != v2Acc.Disabled {
		t.Errorf("expected Disabled %v, got %v", v2Acc.Disabled, result.Disabled)
	}
	if len(result.BalanceMap) != len(v2Acc.BalanceMap) {
		t.Errorf("expected %d balances, got %d", len(v2Acc.BalanceMap), len(result.BalanceMap))
	}
	if resultBalance, ok := result.BalanceMap["balance1"]; ok {
		if len(resultBalance) != 1 {
			t.Errorf("expected 1 balance in balance1, got %d", len(resultBalance))
		}
		if resultBalance[0].ID != v2Acc.BalanceMap["balance1"][0].ID {
			t.Errorf("expected Balance ID %v, got %v", v2Acc.BalanceMap["balance1"][0].ID, resultBalance[0].ID)
		}
		if resultBalance[0].Value != v2Acc.BalanceMap["balance1"][0].Value {
			t.Errorf("expected Balance Value %v, got %v", v2Acc.BalanceMap["balance1"][0].Value, resultBalance[0].Value)
		}
	}
}
