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

	"github.com/cgrates/cgrates/utils"
)

var sTestsDspCDRs = []func(t *testing.T){
	testDspCDRsPing,
}

//Test start here
func TestDspCDRsITMySQL(t *testing.T) {
	testDsp(t, sTestsDspCDRs, "TestDspCDRs", "all", "all2", "attributes", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspCDRsITMongo(t *testing.T) {
	testDsp(t, sTestsDspCDRs, "TestDspCDRs", "all", "all2", "attributes_mongo", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
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
