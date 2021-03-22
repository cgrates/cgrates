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

	"github.com/cgrates/cgrates/utils"
)

var sTestsDspGrd = []func(t *testing.T){
	testDspGrdPing,
	testDspGrdLock,
}

//Test start here
func TestDspGuardianST(t *testing.T) {
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
	testDsp(t, sTestsDspGrd, "TestDspGuardianS", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspGrdPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.GuardianSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.GuardianSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "grd12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspGrdLock(t *testing.T) {
	// lock
	args := utils.AttrRemoteLock{
		ReferenceID: "",
		LockIDs:     []string{"lock1"},
		Timeout:     500 * time.Millisecond,
	}
	var reply string
	if err := dispEngine.RPC.Call(utils.GuardianSv1RemoteLock, &AttrRemoteLockWithOpts{
		AttrRemoteLock: args,
		Tenant:         "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "grd12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	}

	var unlockReply []string
	if err := dispEngine.RPC.Call(utils.GuardianSv1RemoteUnlock, &AttrRemoteUnlockWithOpts{
		RefID:  reply,
		Tenant: "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "grd12345",
		},
	}, &unlockReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(args.LockIDs, unlockReply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(args.LockIDs), utils.ToJSON(unlockReply))
	}
}
