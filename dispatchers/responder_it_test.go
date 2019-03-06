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
	"fmt"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var sTestsDspRsp = []func(t *testing.T){
	testDspResponderStatus,
	testDspResponderShutdown,

	testDspResponderRandom,
}

//Test start here
func TestDspResponderTMySQL(t *testing.T) {
	testDsp(t, sTestsDspRsp, "TestDspAttributeS", "all", "all2", "attributes", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspResponderMongo(t *testing.T) {
	testDsp(t, sTestsDspRsp, "TestDspAttributeS", "all", "all2", "attributes_mongo", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

func testDspResponderStatus(t *testing.T) {
	var reply map[string]interface{}
	if err := allEngine.RCP.Call(utils.ResponderStatus, "", &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "ALL" {
		t.Errorf("Received: %s", reply)
	}
	ev := TntWithApiKey{
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "rsp12345",
		},
	}
	if err := dispEngine.RCP.Call(utils.ResponderStatus, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "ALL" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ResponderStatus, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "ALL2" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	allEngine.startEngine(t)
}

func getNodeWithRoute(route string, t *testing.T) string {
	var reply map[string]interface{}
	var pingReply string
	pingEv := CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.EVENT_NAME: "Random",
			},
		},
		DispatcherResource: DispatcherResource{
			APIKey:  "attr12345",
			RouteID: &route,
		},
	}
	ev := TntWithApiKey{
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey:  "rsp12345",
			RouteID: &route,
		},
	}

	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, pingEv, &pingReply); err != nil {
		t.Error(err)
	} else if pingReply != utils.Pong {
		t.Errorf("Received: %s", pingReply)
	}
	if err := dispEngine.RCP.Call(utils.ResponderStatus, &ev, &reply); err != nil {
		t.Error(err)
	}
	return reply[utils.NodeID].(string)
}

func testDspResponderRandom(t *testing.T) {
	node := getNodeWithRoute("r_init", t)
	for i := 0; i < 10; i++ {
		if node != getNodeWithRoute(fmt.Sprintf("R_%v", i), t) {
			return
		}
	}
	t.Errorf("Random strategy fail with 0.0009765625%% probability")
}

func testDspResponderShutdown(t *testing.T) {
	var reply string
	var statusReply map[string]interface{}
	ev := TntWithApiKey{
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "rsp12345",
		},
	}
	if err := dispEngine.RCP.Call(utils.ResponderShutdown, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != "Done!" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	if err := dispEngine.RCP.Call(utils.ResponderStatus, &ev, &statusReply); err != nil {
		t.Error(err)
	} else if statusReply[utils.NodeID] != "ALL2" {
		t.Errorf("Received: %s", utils.ToJSON(statusReply))
	}
	if err := dispEngine.RCP.Call(utils.ResponderShutdown, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != "Done!" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}
