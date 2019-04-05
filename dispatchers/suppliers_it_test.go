// +build integration

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

package dispatchers

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspSup = []func(t *testing.T){
	testDspSupPingFailover,
	testDspSupGetSupFailover,
	testDspSupGetSupRoundRobin,

	testDspSupPing,
	testDspSupTestAuthKey,
	testDspSupTestAuthKey2,
}

//Test start here
func TestDspSupplierSTMySQL(t *testing.T) {
	testDsp(t, sTestsDspSup, "TestDspSupplierS", "all", "all2", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspSupplierSMongo(t *testing.T) {
	testDsp(t, sTestsDspSup, "TestDspSupplierS", "all", "all2", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

func testDspSupPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.SupplierSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.SupplierSv1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspSupPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.SupplierSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.SupplierSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.SupplierSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.SupplierSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspSupGetSupFailover(t *testing.T) {
	var rpl *engine.SortedSuppliers
	eRpl1 := &engine.SortedSuppliers{
		ProfileID: "SPL_WEIGHT_2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID:         "supplier1",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	eRpl := &engine.SortedSuppliers{
		ProfileID: "SPL_ACNT_1002",
		Sorting:   utils.MetaLeastCost,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID:         "supplier1",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Cost:         0.3166,
					utils.RatingPlanID: "RP_1002_LOW",
					utils.Weight:       10.0,
				},
			},
			{
				SupplierID:         "supplier2",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Cost:         0.6334,
					utils.RatingPlanID: "RP_1002",
					utils.Weight:       20.0,
				},
			},
		},
	}
	args := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "Event1",
				utils.Account:     "1002",
				utils.Subject:     "1002",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.SupplierSv1GetSuppliers,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl1, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl1), utils.ToJSON(rpl))
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.SupplierSv1GetSuppliers,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl), utils.ToJSON(rpl))
	}
	allEngine2.startEngine(t)
}

func testDspSupTestAuthKey(t *testing.T) {
	var rpl *engine.SortedSuppliers
	args := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.Account:     "1002",
				utils.Subject:     "1002",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.SupplierSv1GetSuppliers,
		args, &rpl); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspSupTestAuthKey2(t *testing.T) {
	var rpl *engine.SortedSuppliers
	eRpl := &engine.SortedSuppliers{
		ProfileID: "SPL_ACNT_1002",
		Sorting:   utils.MetaLeastCost,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID:         "supplier1",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Cost:         0.3166,
					utils.RatingPlanID: "RP_1002_LOW",
					utils.Weight:       10.0,
				},
			},
			{
				SupplierID:         "supplier2",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Cost:         0.6334,
					utils.RatingPlanID: "RP_1002",
					utils.Weight:       20.0,
				},
			},
		},
	}
	args := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.Account:     "1002",
				utils.Subject:     "1002",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.SupplierSv1GetSuppliers,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl), utils.ToJSON(rpl))
	}
}

func testDspSupGetSupRoundRobin(t *testing.T) {
	var rpl *engine.SortedSuppliers
	eRpl1 := &engine.SortedSuppliers{
		ProfileID: "SPL_WEIGHT_2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID:         "supplier1",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	eRpl := &engine.SortedSuppliers{
		ProfileID: "SPL_ACNT_1002",
		Sorting:   utils.MetaLeastCost,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID:         "supplier1",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Cost:         0.3166,
					utils.RatingPlanID: "RP_1002_LOW",
					utils.Weight:       10.0,
				},
			},
			{
				SupplierID:         "supplier2",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Cost:         0.6334,
					utils.RatingPlanID: "RP_1002",
					utils.Weight:       20.0,
				},
			},
		},
	}
	args := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "RoundRobin",
				utils.Account:     "1002",
				utils.Subject:     "1002",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.SupplierSv1GetSuppliers,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl1, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl1), utils.ToJSON(rpl))
	}
	if err := dispEngine.RCP.Call(utils.SupplierSv1GetSuppliers,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl), utils.ToJSON(rpl))
	}
}
