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
	"net/rpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rlsV1CfgPath string
	rlsV1Cfg     *config.CGRConfig
	rlsV1Rpc     *rpc.Client
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

//Test start here
func TestRsV1IT(t *testing.T) {
	switch *dbType {
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
	rlsV1CfgPath = path.Join(*dataDir, "conf", "samples", rlsV1ConfDIR)
	if rlsV1Cfg, err = config.NewCGRConfigFromPath(rlsV1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RsInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1RsResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rlsV1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RsRpcConn(t *testing.T) {
	var err error
	rlsV1Rpc, err = newRPCClient(rlsV1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1RsSetProfile(t *testing.T) {
	rls := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES_GR_TEST",
			FilterIDs:         []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z;2014-07-14T14:26:00Z"},
			UsageTTL:          -1,
			Limit:             2,
			AllocationMessage: "Account1Channels",
			Weight:            20,
			ThresholdIDs:      []string{utils.MetaNone},
		},
	}
	var result string
	if err := rlsV1Rpc.Call(utils.APIerSv1SetResourceProfile, rls, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1RsAllocate(t *testing.T) {
	argsRU := utils.ArgRSv1ResourceUsage{

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Destination": "1002",
			},
		},

		UsageID: "chan_1",
		Units:   1,
	}
	var reply string
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "Account1Channels" {
		t.Error("Unexpected reply returned", reply)
	}

	argsRU2 := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Destination": "1002",
			},
		},

		UsageID: "chan_2",
		Units:   1,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU2, &reply); err != nil {
		t.Error(err)
	} else if reply != "Account1Channels" {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1RsAuthorize(t *testing.T) {
	var reply *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Destination": "1002",
			},
		},

		UsageID: "RandomUsageID",
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent,
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
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "chan_1",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Destination": "1002"},
		},

		Units: 1,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AuthorizeResources,
		&argsRU, &reply2); err.Error() != "RESOURCE_UNAUTHORIZED" {
		t.Error(err)
	}
}

func testV1RsStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
