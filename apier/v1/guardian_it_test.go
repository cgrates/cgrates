//go:build integration
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
package v1

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//Test start here
func TestGuardianSIT(t *testing.T) {
	var err error
	var guardianConfigDIR string
	switch *dbType {
	case utils.MetaInternal:
		guardianConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		guardianConfigDIR = "tutmysql"
	case utils.MetaMongo:
		guardianConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	guardianCfgPath := path.Join(*dataDir, "conf", "samples", guardianConfigDIR)
	guardianCfg, err := config.NewCGRConfigFromPath(guardianCfgPath)
	if err != nil {
		t.Error(err)
	}

	if err = engine.InitDataDb(guardianCfg); err != nil {
		t.Fatal(err)
	}

	if err = engine.InitStorDb(guardianCfg); err != nil {
		t.Fatal(err)
	}

	// start engine
	if _, err = engine.StopStartEngine(guardianCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}

	// start RPC
	guardianRPC, err := newRPCClient(guardianCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}

	// lock
	args := utils.AttrRemoteLock{
		ReferenceID: "",
		LockIDs:     []string{"lock1"},
		Timeout:     500 * time.Millisecond,
	}
	var reply string
	if err = guardianRPC.Call(utils.GuardianSv1RemoteLock, &args, &reply); err != nil {
		t.Error(err)
	}
	var unlockReply []string
	if err = guardianRPC.Call(utils.GuardianSv1RemoteUnlock, &dispatchers.AttrRemoteUnlockWithAPIOpts{RefID: reply}, &unlockReply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(args.LockIDs, unlockReply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(args.LockIDs), utils.ToJSON(unlockReply))
	}

	// ping
	var resp string
	if err = guardianRPC.Call(utils.GuardianSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}

	// stop engine
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
