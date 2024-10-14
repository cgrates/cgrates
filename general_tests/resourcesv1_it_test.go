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
package general_tests

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rlsV1CfgPath string
	rlsV1Cfg     *config.CGRConfig
	rlsV1Rpc     *birpc.Client
	rlsV1ConfDIR string //run tests for specific configuration

	sTestsRLSV1 = []func(t *testing.T){
		testV1RsLoadConfig,
		testV1RsInitDataDb,
		testV1RsResetStorDb,
		testV1RsStartEngine,
		testV1RsRpcConn,
		testV1RsSetProfile,
		testV1RsAllocate,
		testV1RsAuthorize,
		testV1RsStopEngine,
	}
)

// Test start here
func TestRsV1IT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		rlsV1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		rlsV1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		rlsV1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRLSV1 {
		t.Run(rlsV1ConfDIR, stest)
	}
}

func testV1RsLoadConfig(t *testing.T) {
	var err error
	rlsV1CfgPath = path.Join(*utils.DataDir, "conf", "samples", rlsV1ConfDIR)
	if rlsV1Cfg, err = config.NewCGRConfigFromPath(rlsV1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1RsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rlsV1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RsRpcConn(t *testing.T) {
	rlsV1Rpc = engine.NewRPCClient(t, rlsV1Cfg.ListenCfg())
}

func testV1RsSetProfile(t *testing.T) {
	rls := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RES_GR_TEST",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          -1,
			Limit:             2,
			AllocationMessage: "Account1Channels",
			Weight:            20,
			ThresholdIDs:      []string{utils.MetaNone},
		},
	}
	var result string
	if err := rlsV1Rpc.Call(context.Background(), utils.APIerSv1SetResourceProfile, rls, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1RsAllocate(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "chan_1",
			utils.OptsResourcesUnits:   1,
		},
	}
	var reply string
	if err := rlsV1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		cgrEv, &reply); err != nil {
		t.Error(err)
	} else if reply != "Account1Channels" {
		t.Error("Unexpected reply returned", reply)
	}

	cgrEv2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "chan_2",
			utils.OptsResourcesUnits:   1,
		},
	}
	if err := rlsV1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		cgrEv2, &reply); err != nil {
		t.Error(err)
	} else if reply != "Account1Channels" {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1RsAuthorize(t *testing.T) {
	var reply *engine.Resources
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RandomUsageID",
		},
	}
	if err := rlsV1Rpc.Call(context.Background(), utils.ResourceSv1GetResourcesForEvent,
		args, &reply); err != nil {
		t.Error(err)
	}
	if reply == nil {
		t.Errorf("Expecting reply to not be nil")
		// reply shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	if len(*reply) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(*reply))
	}
	if (*reply)[0].ID != "RES_GR_TEST" {
		t.Errorf("Expecting: %+v, received: %+v", "RES_GR_TEST", (*reply)[0].ID)
	}
	if len((*reply)[0].Usages) != 2 {
		t.Errorf("Expecting: %+v, received: %+v", 2, len((*reply)[0].Usages))
	}

	var reply2 string
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "chan_1",
			utils.OptsResourcesUnits:   1,
		},
	}
	if err := rlsV1Rpc.Call(context.Background(), utils.ResourceSv1AuthorizeResources,
		&cgrEv, &reply2); err.Error() != "RESOURCE_UNAUTHORIZED" {
		t.Error(err)
	}
}

func testV1RsStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
