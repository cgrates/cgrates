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
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	chargerCfgPath   string
	chargerCfg       *config.CGRConfig
	chargerRPC       *rpc.Client
	chargerProfile   *ChargerWithCache
	chargerConfigDIR string //run tests for specific configuration

	chargerEvent = []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			Opts: map[string]interface{}{utils.OptsContext: "simpleauth"},
		},

		{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.AccountField: "1010",
				"DistinctMatch":    "cgrates",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1007",
			},
			Opts: map[string]interface{}{utils.OptsContext: "simpleauth"},
		},
	}

	sTestsCharger = []func(t *testing.T){
		testChargerSInitCfg,
		testChargerSInitDataDb,
		testChargerSResetStorDb,
		testChargerSStartEngine,
		testChargerSRPCConn,
		testChargerSLoadAddCharger,
		testChargerSGetChargersForEvent,
		testChargerSGetChargersForEvent2,
		testChargerSProcessEvent,
		testChargerSSetChargerProfile,
		testChargerSGetChargerProfileIDs,
		testChargerSUpdateChargerProfile,
		testChargerSRemChargerProfile,
		testChargerSPing,
		testChargerSProcessWithNotFoundAttribute,
		testChargerSProccessEventWithProcceSRunS,
		testChargerSSetChargerProfileWithoutTenant,
		testChargerSRemChargerProfileWithoutTenant,
		testChargerSKillEngine,
	}
)

//Test start here
func TestChargerSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		chargerConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		chargerConfigDIR = "tutmysql"
	case utils.MetaMongo:
		chargerConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCharger {
		t.Run(chargerConfigDIR, stest)
	}
}

func testChargerSInitCfg(t *testing.T) {
	var err error
	chargerCfgPath = path.Join(*dataDir, "conf", "samples", chargerConfigDIR)
	chargerCfg, err = config.NewCGRConfigFromPath(chargerCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testChargerSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(chargerCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testChargerSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(chargerCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testChargerSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(chargerCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testChargerSRPCConn(t *testing.T) {
	var err error
	chargerRPC, err = newRPCClient(chargerCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testChargerSLoadAddCharger(t *testing.T) {
	chargerProfile := &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
			Weight:       20,
		},
	}

	var result string
	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	chargerProfile = &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "Charger2",
			FilterIDs: []string{"*string:~*req.Account:1007"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*constant:*req.RequestType:*rated;*constant:*req.Category:call"},
			Weight:       20,
		},
	}

	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	chargerProfile = &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "ChargerNotMatching",
			FilterIDs: []string{"*string:~*req.Account:1015", "*gt:~*req.Usage:10"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}

	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:   "cgrates.org",
			ID:       "ATTR_1001_SIMPLEAUTH",
			Contexts: []string{"simpleauth"},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "Password",
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "CGRateS.org",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	if err := chargerRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testChargerSGetChargersForEvent(t *testing.T) {
	chargerProfiles := &engine.ChargerProfiles{
		&engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
			Weight:       20,
		},
	}
	var result *engine.ChargerProfiles
	if err := chargerRPC.Call(utils.ChargerSv1GetChargersForEvent,
		chargerEvent[1], &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := chargerRPC.Call(utils.ChargerSv1GetChargersForEvent,
		chargerEvent[0], &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, chargerProfiles) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(chargerProfiles), utils.ToJSON(result))
	}

	chargerProfiles = &engine.ChargerProfiles{
		&engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
			Weight:       20,
		},
	}
	chargerEvent[0].Tenant = utils.EmptyString
	if err := chargerRPC.Call(utils.ChargerSv1GetChargersForEvent,
		chargerEvent[0], &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, chargerProfiles) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(chargerProfiles), utils.ToJSON(result))
	}
}

func testChargerSGetChargersForEvent2(t *testing.T) {
	var result *engine.ChargerProfiles
	if err := chargerRPC.Call(utils.ChargerSv1GetChargersForEvent,
		&utils.CGREvent{ // matching Charger1
			Tenant: utils.EmptyString,
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1015",
				utils.Usage:        1,
			},
		}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
		t.Error(utils.ToJSON(result))
	}
}

func testChargerSProcessEvent(t *testing.T) {
	processedEv := []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile:    "Charger1",
			AttributeSProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
			AlteredFields:      []string{utils.MetaReqRunID, "*req.Password"},
			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.AccountField: "1001",
					"Password":         "CGRateS.org",
					"RunID":            utils.MetaDefault,
				},
				Opts: map[string]interface{}{utils.OptsContext: "simpleauth", utils.Subsys: utils.MetaChargers},
			},
		},
	}
	var result []*engine.ChrgSProcessEventReply
	if err := chargerRPC.Call(utils.ChargerSv1ProcessEvent, chargerEvent[1], &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := chargerRPC.Call(utils.ChargerSv1ProcessEvent, chargerEvent[0], &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, processedEv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(processedEv), utils.ToJSON(result))
	}
	result = []*engine.ChrgSProcessEventReply{}
	processedEv = []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile:    "Charger2",
			AttributeSProfiles: []string{"*constant:*req.RequestType:*rated;*constant:*req.Category:call"},
			AlteredFields:      []string{"*req.Category", "*req.RequestType", utils.MetaReqRunID},
			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.AccountField: "1007",
					utils.RequestType:  "*rated",
					utils.Category:     "call",
					utils.RunID:        utils.MetaDefault,
				},
				Opts: map[string]interface{}{utils.OptsContext: "simpleauth", utils.Subsys: utils.MetaChargers},
			},
		},
	}
	if err := chargerRPC.Call(utils.ChargerSv1ProcessEvent, chargerEvent[2], &result); err != nil {
		t.Fatal(err)
	}
	sort.Strings(result[0].AlteredFields)
	if !reflect.DeepEqual(result, processedEv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(processedEv), utils.ToJSON(result))
	}

	chargerEvent[2].Tenant = utils.EmptyString
	if err := chargerRPC.Call(utils.ChargerSv1ProcessEvent, chargerEvent[2], &result); err != nil {
		t.Fatal(err)
	}
	sort.Strings(result[0].AlteredFields)
	if !reflect.DeepEqual(result, processedEv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(processedEv), utils.ToJSON(result))
	}
}

func testChargerSSetChargerProfile(t *testing.T) {
	chargerProfile = &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			FilterIDs: []string{"*wrong:inline"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"Attr1", "Attr2"},
			Weight:       20,
		},
	}
	var result string
	expErr := "SERVER_ERROR: broken reference to filter: *wrong:inline for item with ID: cgrates.org:ApierTest"
	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}
	var reply *engine.ChargerProfile
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	chargerProfile.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~Account:1002"}
	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply)
	}
}

func testChargerSGetChargerProfileIDs(t *testing.T) {
	expected := []string{"Charger1", "Charger2", "ApierTest", "ChargerNotMatching"}
	var result []string
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfileIDs, utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfileIDs, utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfileIDs, utils.PaginatorWithTenant{
		Tenant:    "cgrates.org",
		Paginator: utils.Paginator{Limit: utils.IntPointer(1)},
	}, &result); err != nil {
		t.Error(err)
	} else if 1 != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testChargerSUpdateChargerProfile(t *testing.T) {
	chargerProfile.RunID = "*rated"
	var result string
	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ChargerProfile
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply)
	}
}

func testChargerSRemChargerProfile(t *testing.T) {
	var resp string
	if err := chargerRPC.Call(utils.APIerSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.AttributeProfile
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := chargerRPC.Call(utils.APIerSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &resp); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v received: %v", utils.ErrNotFound, err)
	}
}

func testChargerSPing(t *testing.T) {
	var resp string
	if err := chargerRPC.Call(utils.ChargerSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testChargerSProcessWithNotFoundAttribute(t *testing.T) {
	var result string
	chargerProfile = &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "ChargerWithoutAttribute",
			FilterIDs: []string{"*string:~*req.CustomField:WithoutAttributes"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			RunID:  "CustomRun",
			Weight: 20,
		},
	}

	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CustomEvent",
		Event: map[string]interface{}{
			utils.AccountField: "Random",
			"CustomField":      "WithoutAttributes",
		},
	}
	processedEv := []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile:    "ChargerWithoutAttribute",
			AttributeSProfiles: []string{},
			AlteredFields:      []string{utils.MetaReqRunID},
			CGREvent: &utils.CGREvent{ // matching ChargerWithoutAttribute
				Tenant: "cgrates.org",
				ID:     "CustomEvent",
				Event: map[string]interface{}{
					utils.AccountField: "Random",
					"CustomField":      "WithoutAttributes",
					"RunID":            "CustomRun",
				},
				Opts: map[string]interface{}{utils.Subsys: utils.MetaChargers},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		processedEv[0].AttributeSProfiles = nil
	}
	var rply []*engine.ChrgSProcessEventReply
	if err := chargerRPC.Call(utils.ChargerSv1ProcessEvent, ev, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, processedEv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(processedEv), utils.ToJSON(rply))
	}

}
func testChargerSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testChargerSProccessEventWithProcceSRunS(t *testing.T) {
	chargerProfile = &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			FilterIDs: []string{"*string:~*req.Account:1010"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*constant:*req.Account:1002", "*constant:*req.Account:1003"},
			Weight:       20,
		},
	}
	var result string
	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ChargerProfile
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply)
	}

	processedEv := []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile:    "ApierTest",
			AttributeSProfiles: []string{"*constant:*req.Account:1002"},
			AlteredFields:      []string{utils.MetaReqRunID, "*req.Account"},
			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.AccountField: "1002",
					utils.RunID:        "*default",
				},
				Opts: map[string]interface{}{
					utils.Subsys:                    utils.MetaChargers,
					utils.OptsAttributesProcessRuns: 1.,
				},
			},
		},
	}
	cgrEv := &utils.CGREvent{ // matching Charger1
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.AccountField: "1010",
		},
		Opts: map[string]interface{}{utils.OptsAttributesProcessRuns: 1.},
	}
	var result2 []*engine.ChrgSProcessEventReply
	if err := chargerRPC.Call(utils.ChargerSv1ProcessEvent, cgrEv, &result2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result2, processedEv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(processedEv), utils.ToJSON(result2))
	}
}

func testChargerSSetChargerProfileWithoutTenant(t *testing.T) {
	chargerProfile = &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			ID:        "randomID",
			FilterIDs: []string{"*string:~*req.Account:1010"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"Attr1", "Attr2"},
			Weight:       20,
		},
	}
	var reply string
	if err := chargerRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	chargerProfile.ChargerProfile.Tenant = "cgrates.org"
	var result *engine.ChargerProfile
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{ID: "randomID"},
		&result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, result) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(chargerProfile.ChargerProfile), utils.ToJSON(result))
	}
}

func testChargerSRemChargerProfileWithoutTenant(t *testing.T) {
	var reply string
	if err := chargerRPC.Call(utils.APIerSv1RemoveChargerProfile,
		&utils.TenantIDWithCache{ID: "randomID"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *engine.ChargerProfile
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{ID: "randomID"},
		&result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}
