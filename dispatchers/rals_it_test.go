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
func TestDspRALsIT(t *testing.T) {
	var config1, config2, config3 string
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspRALs, "TestDspRALsITMySQL", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspRALsPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.RALsV1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.RALsV1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		Opts: map[string]interface{}{
			utils.OptsAPIKey: "rals12345",
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
		SetupTime:     utils.MetaNow,
		Usage:         "1h",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "rals12345",
		},
	}
	var reply RatingPlanCost
	if err := dispEngine.RPC.Call(utils.RALsV1GetRatingPlansCost, arg, &reply); err != nil {
		t.Error(err)
	} else if reply.RatingPlanID != "RP_1001" {
		t.Error("Unexpected RatingPlanID: ", reply.RatingPlanID)
	} else if *reply.EventCost.Cost != 6.5118 {
		t.Error("Unexpected Cost: ", *reply.EventCost.Cost)
	} else if *reply.EventCost.Usage != time.Hour {
		t.Error("Unexpected Usage: ", *reply.EventCost.Usage)
	}
}
