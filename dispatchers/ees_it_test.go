//go:build integration
// +build integration

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

package dispatchers

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspEEs = []func(t *testing.T){
	testDspEEsPingFailover,
	testDspEEsProcessEventFailover,
	testDspEEsProcessEventRoundRobin,

	testDspEEsPing,
	testDspEEsTestAuthKey,
	testDspEEsTestAuthKey2,
}

func TestDspEEsIT(t *testing.T) {
	var config1, config2, config3 string
	switch *utils.DBType {
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
	if *utils.Encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspEEs, "TestDspEEs", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspEEsPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.EeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply: %s", reply)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsAPIKey: "ees12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1Ping, ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1Ping, ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1Ping, ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspEEsProcessEventFailover(t *testing.T) {
	args := &engine.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.EventName:    "Event1",
				utils.AccountField: "1001",
			},

			APIOpts: map[string]any{
				utils.OptsAPIKey: "ees12345",
			},
		},
	}
	var reply map[string]map[string]any
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1ProcessEvent, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1ProcessEvent,
		args, &reply); err != nil {
		t.Fatal(err)
	}
	allEngine2.startEngine(t)
}

func testDspEEsPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.EeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsAPIKey: "ees12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspEEsTestAuthKey(t *testing.T) {
	args := &engine.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsAPIKey: "12345",
			},
		},
	}
	var reply map[string]map[string]any
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1ProcessEvent,
		args, &reply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ErrUnauthorizedApi.Error(), err)
	}
}

func testDspEEsTestAuthKey2(t *testing.T) {
	args := &engine.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsAPIKey: "ees12345",
			},
		},
	}
	var reply map[string]map[string]any
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1ProcessEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if _, ok := reply[utils.MetaDefault]; !ok {
		t.Error("expected to match the *default exporter")
	}
}

func testDspEEsProcessEventRoundRobin(t *testing.T) {
	args := &engine.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.EventName:    "RoundRobin",
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsAPIKey: "ees12345",
			},
		},
	}
	var reply map[string]map[string]any
	// To ALL2
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1ProcessEvent,
		args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
	// To ALL
	if err := dispEngine.RPC.Call(context.Background(), utils.EeSv1ProcessEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if _, ok := reply[utils.MetaDefault]; !ok {
		t.Error("expected to match the *default exporter")
	}
}
