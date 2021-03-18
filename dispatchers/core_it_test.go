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
	"net/rpc"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var sTestsDspCore = []func(t *testing.T){
	testDspCoreLoad,
}

//Test start here
func TestDspCoreIT(t *testing.T) {
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
	testDsp(t, sTestsDspCore, "TestDspCoreIT", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspCoreLoad(t *testing.T) {
	var status map[string]interface{}
	statusTnt := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey:  "core12345",
			utils.OptsRouteID: "core1",
			"EventType":       "LoadDispatcher",
		},
	}
	expNodeID := "ALL"
	if err := dispEngine.RPC.Call(utils.CoreSv1Status, statusTnt, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID] == "ALL2" {
		expNodeID = "ALL2"
	}
	dur := &utils.DurationArgs{
		Duration: 500 * time.Millisecond,
		Tenant:   "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey:  "core12345",
			utils.OptsRouteID: "core1",
			"EventType":       "LoadDispatcher",
		},
	}
	var rply string
	statusTnt2 := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "core12345",
			"EventType":      "LoadDispatcher",
		},
	}
	call := dispEngine.RPC.Go(utils.CoreSv1Sleep, dur, &rply, make(chan *rpc.Call, 1))
	if err := dispEngine.RPC.Call(utils.CoreSv1Status, statusTnt2, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID] != expNodeID {
		t.Errorf("Expected status to be called on node <%s> but it was called on <%s>", expNodeID, status[utils.NodeID])
	}
	if ans := <-call.Done; ans.Error != nil {
		t.Fatal(ans.Error)
	} else if rply != utils.OK {
		t.Errorf("Expected: %q ,received: %q", utils.OK, rply)
	}

}
