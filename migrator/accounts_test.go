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
	v1b := &v1Balance{Value: 10, Weight: 10, DestinationIds: "NAT", Timings: []*engine.RITiming{&engine.RITiming{StartTime: "00:00:00"}}}
	v1Acc := &v1Account{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]v1BalanceChain{utils.VOICE: v1BalanceChain{v1b}, utils.MONETARY: v1BalanceChain{&v1Balance{Value: 21, Timings: []*engine.RITiming{&engine.RITiming{StartTime: "00:00:00"}}}}}}
	v2 := &engine.Balance{Uuid: "", ID: "", Value: 10, Directions: utils.StringMap{"*OUT": true}, Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "", Categories: utils.NewStringMap(""), SharedGroups: utils.NewStringMap(""), Timings: []*engine.RITiming{&engine.RITiming{StartTime: "00:00:00"}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	m2 := &engine.Balance{Uuid: "", ID: "", Value: 21, Directions: utils.StringMap{"*OUT": true}, DestinationIDs: utils.NewStringMap(""), RatingSubject: "", Categories: utils.NewStringMap(""), SharedGroups: utils.NewStringMap(""), Timings: []*engine.RITiming{&engine.RITiming{StartTime: "00:00:00"}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	testAccount := &engine.Account{ID: "CUSTOMER_1:rif", BalanceMap: map[string]engine.Balances{utils.VOICE: engine.Balances{v2}, utils.MONETARY: engine.Balances{m2}}, UnitCounters: engine.UnitCounters{}, ActionTriggers: engine.ActionTriggers{}}
	if def := v1b.IsDefault(); def != false {
		t.Errorf("Expecting: false, received: true")
	}
	newAcc := v1Acc.AsAccount()
	if !reflect.DeepEqual(testAccount, newAcc) {
		t.Errorf("Expecting: %+v, received: %+v", testAccount, newAcc)
	}
}
