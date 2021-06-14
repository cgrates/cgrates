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

package apis

import (
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	chgrsPrfCfgPath   string
	chgrsPrfCfg       *config.CGRConfig
	chgrsSRPC         *birpc.Client
	chgrsPrfConfigDIR string //run tests for specific configuration

	sTestsChgrsPrf = []func(t *testing.T){
		testChgrsSInitCfg,
		testChgrsSInitDataDb,
		testChgrsSResetStorDb,
		testChgrsSStartEngine,
		testChgrsSRPCConn,
		testGetChgrsProfileBeforeSet,
		testChgrsSetGetChargerProfile,
		testChgrsGetChargerProfileIDs,
		testGetChgrsProfileBeforeSet2,
		testChgrsSetGetChargerProfile2,
		testChgrsGetChargerProfileIDs2,
		testGetChgrsProfileBeforeSet3,
		testChgrsSetGetChargerProfile3,
		testChgrsGetChargerProfileIDs3,
		testChgrsRmvChargerProfile,
		testChgrsRmvChargerProfile2,
		testChgrsRmvChargerProfile3,
		testGetChgrsProfileBeforeSet,
		testChgrsSetGetChargerProfileEvent,
		testChgrsGetChargersForEvent,
		testChgrsProcessEvent,
		testChgrsSKillEngine,
	}
)

func TestChgrsSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		chgrsPrfConfigDIR = "apis_chargers_internal"
	case utils.MetaMongo:
		chgrsPrfConfigDIR = "apis_chargers_mongo"
	case utils.MetaMySQL:
		chgrsPrfConfigDIR = "apis_chargers_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsChgrsPrf {
		t.Run(chgrsPrfConfigDIR, stest)
	}
}

func testChgrsSInitCfg(t *testing.T) {
	var err error
	chgrsPrfCfgPath = path.Join(*dataDir, "conf", "samples", chgrsPrfConfigDIR)
	chgrsPrfCfg, err = config.NewCGRConfigFromPath(chgrsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testChgrsSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(chgrsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

func testChgrsSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(chgrsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testChgrsSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(chgrsPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testChgrsSRPCConn(t *testing.T) {
	var err error
	chgrsSRPC, err = newRPCClient(chgrsPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testGetChgrsProfileBeforeSet(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST",
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testChgrsSetGetChargerProfile(t *testing.T) {
	chgrsPrf := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "TEST_CHARGERS_IT_TEST",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		APIOpts: nil,
	}
	var reply string
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chgrsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedChargerPrf := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "TEST_CHARGERS_IT_TEST",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result *engine.ChargerProfile
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedChargerPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedChargerPrf), utils.ToJSON(result))
	}
}

func testChgrsGetChargerProfileIDs(t *testing.T) {
	var reply []string
	args := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_CHARGERS_IT_TEST"}
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testGetChgrsProfileBeforeSet2(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST2",
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testChgrsSetGetChargerProfile2(t *testing.T) {
	chgrsPrf := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "TEST_CHARGERS_IT_TEST2",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		APIOpts: nil,
	}
	var reply string
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chgrsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedChargerPrf := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "TEST_CHARGERS_IT_TEST2",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result *engine.ChargerProfile
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST2",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedChargerPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedChargerPrf), utils.ToJSON(result))
	}
}

func testChgrsGetChargerProfileIDs2(t *testing.T) {
	var reply []string
	args := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_CHARGERS_IT_TEST", "TEST_CHARGERS_IT_TEST2"}
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testGetChgrsProfileBeforeSet3(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST3",
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testChgrsSetGetChargerProfile3(t *testing.T) {
	chgrsPrf := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "TEST_CHARGERS_IT_TEST3",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		APIOpts: nil,
	}
	var reply string
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chgrsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedChargerPrf := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "TEST_CHARGERS_IT_TEST3",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result *engine.ChargerProfile
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST3",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedChargerPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedChargerPrf), utils.ToJSON(result))
	}
}

func testChgrsGetChargerProfileIDs3(t *testing.T) {
	var reply []string
	args := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_CHARGERS_IT_TEST", "TEST_CHARGERS_IT_TEST2", "TEST_CHARGERS_IT_TEST3"}
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testChgrsRmvChargerProfile(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST3",
		},
		APIOpts: nil,
	}
	expected := "OK"
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
	var reply2 *utils.TenantIDWithAPIOpts
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST3",
		}, &reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var reply3 []string
	args2 := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	expected2 := []string{"TEST_CHARGERS_IT_TEST", "TEST_CHARGERS_IT_TEST2"}
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		args2, &reply3); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected2, reply3)
	}
}

func testChgrsRmvChargerProfile2(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST2",
		},
		APIOpts: nil,
	}
	expected := "OK"
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
	var reply2 *utils.TenantIDWithAPIOpts
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST2",
		}, &reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var reply3 []string
	args2 := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	expected2 := []string{"TEST_CHARGERS_IT_TEST"}
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		args2, &reply3); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected2, reply3)
	}
}

func testChgrsRmvChargerProfile3(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST",
		},
		APIOpts: nil,
	}
	expected := "OK"
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
	var reply2 *utils.TenantIDWithAPIOpts
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST",
		}, &reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var reply3 []string
	args2 := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		args2, &reply3); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testChgrsSetGetChargerProfileEvent(t *testing.T) {
	chgrsPrf := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "TEST_CHARGERS_IT_TEST",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		APIOpts: nil,
	}
	var reply string
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chgrsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedChargerPrf := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "TEST_CHARGERS_IT_TEST",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result *engine.ChargerProfile
	if err := chgrsSRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedChargerPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedChargerPrf), utils.ToJSON(result))
	}
}
func testChgrsGetChargersForEvent(t *testing.T) {

	expected := &engine.ChargerProfiles{
		{
			Tenant:       "cgrates.org",
			ID:           "TEST_CHARGERS_IT_TEST",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "eventCharger",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]interface{}{},
	}
	reply := &engine.ChargerProfiles{}
	if err := chgrsSRPC.Call(context.Background(), utils.ChargerSv1GetChargersForEvent,
		cgrEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("\nExpected %+v, \nreceived %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testChgrsProcessEvent(t *testing.T) {
	expected := &[]*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile: "TEST_CHARGERS_IT_TEST",
			AlteredFields:   []string{"*req.RunID"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "eventCharger",
				Event: map[string]interface{}{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.RunID:        utils.MetaDefault,
				},
				APIOpts: map[string]interface{}{
					"*subsys": "*chargers",
				},
			},
		},
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "eventCharger",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]interface{}{},
	}
	reply := &[]*engine.ChrgSProcessEventReply{}
	if err := chgrsSRPC.Call(context.Background(), utils.ChargerSv1ProcessEvent,
		cgrEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("\nExpected %+v, \nreceived %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

//Kill the engine when it is about to be finished
func testChgrsSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
