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
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	actPrfCfgPath   string
	actPrfCfg       *config.CGRConfig
	actSRPC         *rpc.Client
	actPrfDataDir   = "/usr/share/cgrates"
	actPrf          *ActionProfileWithCache
	actPrfConfigDIR string //run tests for specific configuration

	sTestsActPrf = []func(t *testing.T){
		testActionSInitCfg,
		testActionSInitDataDb,
		testActionSResetStorDb,
		testActionSStartEngine,
		testActionSRPCConn,
		testActionSLoadFromFolder,
		testActionSGetActionProfile,
		testActionSPing,
		testActionSSettActionProfile,
		testActionSGetActionProfileIDs,
		testActionSGetActionProfileIDsCount,
		testActionSUpdateActionProfile,
		testActionSRemoveActionProfile,
		testActionSKillEngine,
	}
)

//Test start here
func TestActionSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		actPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		actPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		actPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsActPrf {
		t.Run(actPrfConfigDIR, stest)
	}
}

func testActionSInitCfg(t *testing.T) {
	var err error
	actPrfCfgPath = path.Join(actPrfDataDir, "conf", "samples", actPrfConfigDIR)
	actPrfCfg, err = config.NewCGRConfigFromPath(actPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(actPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testActionSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(actPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testActionSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(actPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testActionSRPCConn(t *testing.T) {
	var err error
	actSRPC, err = newRPCClient(actPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testActionSLoadFromFolder(t *testing.T) {
	var reply string
	acts := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutactions")}
	if err := actSRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, acts, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testActionSGetActionProfile(t *testing.T) {
	expected := &engine.ActionProfile{
		Tenant:     "cgrates.org",
		ID:         "ONE_TIME_ACT",
		FilterIDs:  []string{},
		Weight:     10,
		Schedule:   utils.ASAP,
		AccountIDs: map[string]struct{}{"1001": struct{}{}, "1002": struct{}{}},
		Actions: []*engine.APAction{
			&engine.APAction{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestBalance.Value",
				Value:     config.NewRSRParsersMustCompile("10", actPrfCfg.GeneralCfg().RSRSep),
			},
			&engine.APAction{
				ID:        "SET_BALANCE_TEST_DATA",
				FilterIDs: []string{},
				Type:      "*set_balance",
				Path:      "~*balance.TestDataBalance.Type",
				Value:     config.NewRSRParsersMustCompile("*data", actPrfCfg.GeneralCfg().RSRSep),
			},
			&engine.APAction{
				ID:        "TOPUP_TEST_DATA",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestDataBalance.Value",
				Value:     config.NewRSRParsersMustCompile("1024", actPrfCfg.GeneralCfg().RSRSep),
			},
			&engine.APAction{
				ID:        "SET_BALANCE_TEST_VOICE",
				FilterIDs: []string{},
				Type:      "*set_balance",
				Path:      "~*balance.TestVoiceBalance.Type",
				Value:     config.NewRSRParsersMustCompile("*voice", actPrfCfg.GeneralCfg().RSRSep),
			},
			&engine.APAction{
				ID:        "TOPUP_TEST_VOICE",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestVoiceBalance.Value",
				Value:     config.NewRSRParsersMustCompile("15m15s", actPrfCfg.GeneralCfg().RSRSep),
			},
		},
	}
	var reply *engine.ActionProfile
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}}, &reply); err != nil {
		t.Fatal(err)
	} else {
		for _, act := range reply.Actions { // the path variable from RSRParsers is with lower letter and need to be compiled manually in tests to pass reflect.DeepEqual
			act.Value.Compile()
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expecting : %+v \n received: %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func testActionSPing(t *testing.T) {
	var resp string
	if err := actSRPC.Call(utils.ActionSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testActionSSettActionProfile(t *testing.T) {
	actPrf = &ActionProfileWithCache{
		ActionProfileWithOpts: &engine.ActionProfileWithOpts{
			&engine.ActionProfile{
				Tenant:             "tenant_test",
				ID:                 "id_test",
				FilterIDs:          nil,
				ActivationInterval: nil,
				Weight:             0,
				Schedule:           "",
				AccountIDs:         utils.StringSet{},
				Actions: []*engine.APAction{
					{
						ID:        "test_action_id",
						FilterIDs: nil,
						Blocker:   false,
						TTL:       0,
						Type:      "",
						Opts:      nil,
						Path:      "",
						Value:     nil,
					},
					{
						ID:        "test_action_id2",
						FilterIDs: nil,
						Blocker:   false,
						TTL:       0,
						Type:      "",
						Opts:      nil,
						Path:      "",
						Value:     nil,
					},
				},
			},
			map[string]interface{}{},
		},
	}
	var result string
	expErr := utils.ErrNotFound.Error()
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile, &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}, &result); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %v received: %v", expErr, err)
	}
	var reply string
	if err := actSRPC.Call(utils.APIerSv1SetActionProfile, actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var reply2 *engine.ActionProfile
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile, &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrf.ActionProfile, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", actPrf.ActionProfile, reply2)
	}

}

func testActionSGetActionProfileIDs(t *testing.T) {

	expected := []string{"id_test"}
	var result []string
	if err := actSRPC.Call(utils.APIerSv1GetActionProfileIDs, utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := actSRPC.Call(utils.APIerSv1GetActionProfileIDs, utils.PaginatorWithTenant{Tenant: "tenant_test"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := actSRPC.Call(utils.APIerSv1GetActionProfileIDs, utils.PaginatorWithTenant{
		Tenant:    "tenant_test",
		Paginator: utils.Paginator{Limit: utils.IntPointer(1)},
	}, &result); err != nil {
		t.Error(err)
	} else if 1 != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}

}

func testActionSGetActionProfileIDsCount(t *testing.T) {
	var reply int
	if err := actSRPC.Call(utils.APIerSv1GetActionProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "tenant_test"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expecting: 1, received: %+v", reply)
	}

}

func testActionSUpdateActionProfile(t *testing.T) {
	var reply string
	actPrf.Weight = 2
	if err := actSRPC.Call(utils.APIerSv1SetActionProfile, actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var reply2 *engine.ActionProfile
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile, &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrf.ActionProfile, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", actPrf.ActionProfile, reply2)
	}
}

func testActionSRemoveActionProfile(t *testing.T) {
	var reply string
	if err := actSRPC.Call(utils.APIerSv1RemoveActionProfile, &utils.TenantIDWithCache{Tenant: "tenant_test", ID: "id_test"}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var reply2 *engine.ActionProfile
	expErr := utils.ErrNotFound.Error()
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile, &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}, &reply2); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %v received: %v", expErr, err)
	}
	if err := actSRPC.Call(utils.APIerSv1RemoveActionProfile, &utils.TenantIDWithCache{Tenant: "tenant_test", ID: "id_test"}, &reply2); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %v received: %v", expErr, err)
	}
}

func testActionSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
