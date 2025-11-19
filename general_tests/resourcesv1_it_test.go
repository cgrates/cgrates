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
package general_tests

import (
	"os"
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/resources"
	"github.com/cgrates/cgrates/utils"
)

var (
	rlsV1CfgPath string
	rlsV1Cfg     *config.CGRConfig
	rlsV1Rpc     *birpc.Client
	rlsV1ConfDIR string //run tests for specific configuration

	sTestsRLSV1 = []func(t *testing.T){
		testV1RsLoadConfig,
		testV1RsFlushDBs,

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
		rlsV1ConfDIR = "resources_internal"
		if err := os.MkdirAll("/tmp/internal_db/db", 0755); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll("/tmp/internal_db"); err != nil {
				t.Error(err)
			}
		})
	case utils.MetaRedis:
		rlsV1ConfDIR = "tutredis"
	case utils.MetaMySQL:
		rlsV1ConfDIR = "resources_mysql"
	case utils.MetaMongo:
		rlsV1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		rlsV1ConfDIR = "resources_postgres"
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
	if rlsV1Cfg, err = config.NewCGRConfigFromPath(context.Background(), rlsV1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RsFlushDBs(t *testing.T) {
	if err := engine.InitDB(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rlsV1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RsRpcConn(t *testing.T) {
	rlsV1Rpc = engine.NewRPCClient(t, rlsV1Cfg.ListenCfg(), *utils.Encoding)
}

func testV1RsSetProfile(t *testing.T) {
	rls := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES_GR_TEST",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			UsageTTL:          -1,
			Limit:             2,
			AllocationMessage: "Account1Channels",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
			Stored:       true,
		},
	}
	var result string
	if err := rlsV1Rpc.Call(context.Background(), utils.AdminSv1SetResourceProfile, rls, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1RsAllocate(t *testing.T) {
	argsRU := &utils.CGREvent{
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
		argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "Account1Channels" {
		t.Error("Unexpected reply returned", reply)
	}

	argsRU2 := &utils.CGREvent{
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
		argsRU2, &reply); err != nil {
		t.Error(err)
	} else if reply != "Account1Channels" {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1RsAuthorize(t *testing.T) {
	var reply *resources.Resources
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
	if (*reply)[0].Resource.ID != "RES_GR_TEST" {
		t.Errorf("Expecting: %+v, received: %+v", "RES_GR_TEST", (*reply)[0].Resource.ID)
	}
	if len((*reply)[0].Resource.Usages) != 2 {
		t.Errorf("Expecting: %+v, received: %+v", 2, len((*reply)[0].Resource.Usages))
	}

	var reply2 string
	argsRU := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002"},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "chan_1",
			utils.OptsResourcesUnits:   1,
		},
	}
	if err := rlsV1Rpc.Call(context.Background(), utils.ResourceSv1AuthorizeResources,
		&argsRU, &reply2); err.Error() != "RESOURCE_UNAUTHORIZED" {
		t.Error(err)
	}
}

func testV1RsStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
