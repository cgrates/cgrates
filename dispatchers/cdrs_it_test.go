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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspCDRs = []func(t *testing.T){
	testDspCDRsPing,
	testDspCDRsProcessEvent,
	testDspCDRsCountCDR,
	testDspCDRsGetCDR,
}

//Test start here
func TestDspCDRsITMySQL(t *testing.T) {
	testDsp(t, sTestsDspCDRs, "TestDspCDRs", "all", "all2", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspCDRsITMongo(t *testing.T) {
	testDsp(t, sTestsDspCDRs, "TestDspCDRs", "all", "all2", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

func testDspCDRsPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.CDRsV1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.CDRsV1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("cdrs12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsProcessEvent(t *testing.T) {
	var reply string
	args := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsProcessCDR1",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testDspCDRsProcessEvent",
				utils.RequestType: utils.META_RATED,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Duration(1) * time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("cdrs12345"),
		},
	}

	if err := dispEngine.RCP.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testDspCDRsCountCDR(t *testing.T) {
	var reply int64
	args := &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
		TenantWithArgDispatcher: &utils.TenantWithArgDispatcher{
			TenantArg: &utils.TenantArg{
				Tenant: "cgrates.org",
			},
			ArgDispatcher: &utils.ArgDispatcher{
				APIKey: utils.StringPointer("cdrs12345"),
			},
		},
	}

	if err := dispEngine.RCP.Call(utils.CDRsV1CountCDRs, args, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Received: %+v", reply)
	}
}

func testDspCDRsGetCDR(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
		TenantWithArgDispatcher: &utils.TenantWithArgDispatcher{
			TenantArg: &utils.TenantArg{
				Tenant: "cgrates.org",
			},
			ArgDispatcher: &utils.ArgDispatcher{
				APIKey: utils.StringPointer("cdrs12345"),
			},
		},
	}

	if err := dispEngine.RCP.Call(utils.CDRsV1GetCDRs, args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "1888cd1ba89f9654ca22b45787db217182f9da0d" {
		t.Errorf("Expected: 1888cd1ba89f9654ca22b45787db217182f9da0d , received:%v", reply[0].CGRID)
	}
}
