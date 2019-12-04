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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var sTestsDspRALs = []func(t *testing.T){
	testDspRALsPing,
	testDspRALsGetRatingPlanCost,
}

//Test start here
func TestDspRALsITMySQL(t *testing.T) {
	testDsp(t, sTestsDspRALs, "TestDspRALsITMySQL", "all", "all2", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func testDspRALsPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.RALsV1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.RALsV1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("rals12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspRALsGetRatingPlanCost(t *testing.T) {
	arg := &utils.RatingPlanCostArg{
		Destination:   "1002",
		RatingPlanIDs: []string{"RP_1001", "RP_1002"},
		SetupTime:     utils.META_NOW,
		Usage:         "1h",
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("rals12345"),
		},
	}
	var reply RatingPlanCost
	if err := dispEngine.RPC.Call(utils.RALsV1GetRatingPlansCost, arg, &reply); err != nil {
		t.Error(err)
	} else if reply.RatingPlanID != "RP_1001" {
		t.Error("Unexpected RatingPlanID: ", reply.RatingPlanID)
	} else if *reply.EventCost.Cost != 6.5118 {
		t.Error("Unexpected Cost: ", *reply.EventCost.Cost)
	} else if *reply.EventCost.Usage != time.Duration(time.Hour) {
		t.Error("Unexpected Usage: ", *reply.EventCost.Usage)
	}
}
