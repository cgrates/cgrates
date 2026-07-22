//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Test start here
func TestGuardianSIT(t *testing.T) {
	var err error
	var guardianConfigDIR string
	switch *utils.DBType {
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
	guardianCfgPath := path.Join(*utils.DataDir, "conf", "samples", guardianConfigDIR)
	guardianCfg, err := config.NewCGRConfigFromPath(guardianCfgPath)
	if err != nil {
		t.Error(err)
	}

	if err = engine.InitDataDB(guardianCfg); err != nil {
		t.Fatal(err)
	}

	if err = engine.InitStorDb(guardianCfg); err != nil {
		t.Fatal(err)
	}

	// start engine
	if _, err = engine.StopStartEngine(guardianCfgPath, *utils.WaitRater); err != nil {
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
	if err = guardianRPC.Call(context.Background(), utils.GuardianSv1RemoteLock, &args, &reply); err != nil {
		t.Error(err)
	}
	var unlockReply []string
	if err = guardianRPC.Call(context.Background(), utils.GuardianSv1RemoteUnlock, &dispatchers.AttrRemoteUnlockWithAPIOpts{RefID: reply}, &unlockReply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(args.LockIDs, unlockReply) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(args.LockIDs), utils.ToJSON(unlockReply))
	}

	// ping
	var resp string
	if err = guardianRPC.Call(context.Background(), utils.GuardianSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}

	// stop engine
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
