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
	actPrf          *engine.ActionProfileWithAPIOpts
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
		//cache test
		testActionSInitCfg,
		testActionSInitDataDb,
		testActionSResetStorDb,
		testActionSStartEngine,
		testActionSRPCConn,
		testActionSCacheTestGetNotFound,
		testActionSCacheTestSet,
		testActionSCacheTestGetNotFound,
		testActionSCacheReload,
		testActionSCacheTestGetFound,
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
	actPrfCfgPath = path.Join(*dataDir, "conf", "samples", actPrfConfigDIR)
	actPrfCfg, err = config.NewCGRConfigFromPath(actPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(actPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testActionSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(actPrfCfg); err != nil {
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
	time.Sleep(100 * time.Millisecond)
}

func testActionSGetActionProfile(t *testing.T) {
	expected := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ONE_TIME_ACT",
		FilterIDs: []string{},
		Weight:    10,
		Schedule:  utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: {"1001": {}, "1002": {}},
		},
		Actions: []*engine.APAction{
			{
				ID:   "TOPUP",
				Type: "*add_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestBalance.Units",
					Value: "10",
				}},
			},
			{
				ID:   "SET_BALANCE_TEST_DATA",
				Type: "*set_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestDataBalance.Type",
					Value: "*data",
				}},
			},
			{
				ID:   "TOPUP_TEST_DATA",
				Type: "*add_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestDataBalance.Units",
					Value: "1024",
				}},
			},
			{
				ID:   "SET_BALANCE_TEST_VOICE",
				Type: "*set_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestVoiceBalance.Type",
					Value: "*voice",
				}},
			},
			{
				ID:   "TOPUP_TEST_VOICE",
				Type: "*add_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestVoiceBalance.Units",
					Value: "15m15s",
				}},
			},
			{
				ID:   "SET_BALANCE_TEST_FILTERS",
				Type: "*set_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestVoiceBalance.Filters",
					Value: "*string:~*req.CustomField:500",
				}},
			},
			{
				ID:   "TOPUP_REM_VOICE",
				Type: "*rem_balance",
				Diktats: []*engine.APDiktat{{
					Path: "TestVoiceBalance2",
				}},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		expected.FilterIDs = nil
		for i := range expected.Actions {
			expected.Actions[i].FilterIDs = nil
		}
	}
	var reply *engine.ActionProfile
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expecting : %+v \n received: %+v", utils.ToJSON(expected), utils.ToJSON(reply))
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
	actPrf = &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "tenant_test",
			ID:     "id_test",
			Actions: []*engine.APAction{
				{
					ID:      "test_action_id",
					Diktats: []*engine.APDiktat{{}},
				},
				{
					ID:      "test_action_id2",
					Diktats: []*engine.APDiktat{{}},
				},
			},
		},
		APIOpts: map[string]interface{}{},
	}
	var result string
	expErr := utils.ErrNotFound.Error()
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}}, &result); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %v received: %v", expErr, err)
	}
	var reply string
	if err := actSRPC.Call(utils.APIerSv1SetActionProfile, actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var reply2 *engine.ActionProfile
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}}, &reply2); err != nil {
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
		&utils.TenantWithAPIOpts{Tenant: "tenant_test"}, &reply); err != nil {
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
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrf.ActionProfile, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", actPrf.ActionProfile, reply2)
	}
}

func testActionSRemoveActionProfile(t *testing.T) {
	var reply string
	if err := actSRPC.Call(utils.APIerSv1RemoveActionProfile, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var reply2 *engine.ActionProfile
	expErr := utils.ErrNotFound.Error()
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}}, &reply2); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %v received: %v", expErr, err)
	}
	if err := actSRPC.Call(utils.APIerSv1RemoveActionProfile, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant_test", ID: "id_test"}}, &reply2); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %v received: %v", expErr, err)
	}
}

func testActionSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testActionSCacheTestGetNotFound(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ACTION_CACHE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
}

func testActionSCacheTestGetFound(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ACTION_CACHE"}, &reply); err != nil {
		t.Fatal(err)
	}
}

func testActionSCacheTestSet(t *testing.T) {
	actPrf = &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "ACTION_CACHE",
			Actions: []*engine.APAction{
				{
					ID:      "ACTION_CACHE",
					Diktats: []*engine.APDiktat{{}},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}
	var reply string
	if err := actSRPC.Call(utils.APIerSv1SetActionProfile, actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testActionSCacheReload(t *testing.T) {
	cache := &utils.AttrReloadCacheWithAPIOpts{
		ArgsCache: map[string][]string{
			utils.ActionProfileIDs: {"cgrates.org:ACTION_CACHE"},
		},
	}
	var reply string
	if err := actSRPC.Call(utils.CacheSv1ReloadCache, cache, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}
