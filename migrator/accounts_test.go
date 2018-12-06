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
			&engine.RITiming{
				StartTime: "00:00:00",
			},
		},
	}
	v1b := &v1Balance{
		Value:          100000,
		Weight:         10,
		DestinationIds: "NAT",
		Timings: []*engine.RITiming{
			&engine.RITiming{
				StartTime: "00:00:00",
			},
		},
	}
	v1Acc := &v1Account{
		Id: "*OUT:CUSTOMER_1:rif",
		BalanceMap: map[string]v1BalanceChain{
			utils.DATA:  v1BalanceChain{d1b},
			utils.VOICE: v1BalanceChain{v1b},
			utils.MONETARY: v1BalanceChain{&v1Balance{
				Value: 21,
				Timings: []*engine.RITiming{
					&engine.RITiming{
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
		Directions:     utils.StringMap{"*OUT": true},
		Weight:         10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "",
		Categories:     utils.NewStringMap(""),
		SharedGroups:   utils.NewStringMap(""),
		Timings:        []*engine.RITiming{&engine.RITiming{StartTime: "00:00:00"}},
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{},
	}
	v2 := &engine.Balance{
		Uuid:           "",
		ID:             "",
		Value:          0.0001,
		Directions:     utils.StringMap{"*OUT": true},
		Weight:         10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "",
		Categories:     utils.NewStringMap(""),
		SharedGroups:   utils.NewStringMap(""),
		Timings:        []*engine.RITiming{&engine.RITiming{StartTime: "00:00:00"}},
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{},
	}
	m2 := &engine.Balance{
		Uuid:           "",
		ID:             "",
		Value:          21,
		Directions:     utils.StringMap{"*OUT": true},
		DestinationIDs: utils.NewStringMap(""),
		RatingSubject:  "",
		Categories:     utils.NewStringMap(""),
		SharedGroups:   utils.NewStringMap(""),
		Timings:        []*engine.RITiming{&engine.RITiming{StartTime: "00:00:00"}},
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{},
	}
	testAccount := &engine.Account{
		ID: "CUSTOMER_1:rif",
		BalanceMap: map[string]engine.Balances{
			utils.DATA:     engine.Balances{d2},
			utils.VOICE:    engine.Balances{v2},
			utils.MONETARY: engine.Balances{m2},
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
