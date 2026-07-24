//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var sTestsDspSched = []func(t *testing.T){
	testDspSchedPing,
	testDspSchedPingEmptyCGREventWIthArgDispatcher,
}

// Test start here
func TestDspSchedulerS(t *testing.T) {
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
	testDsp(t, sTestsDspSched, "TestDspSchedulerSTMySQL", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspSchedPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.SchedulerSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.SchedulerSv1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sched12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspSchedPingEmptyCGREventWIthArgDispatcher(t *testing.T) {
	expected := "MANDATORY_IE_MISSING: [APIKey]"
	var reply string
	if err := dispEngine.RPC.Call(utils.SchedulerSv1Ping,
		&utils.CGREventWithArgDispatcher{}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}
