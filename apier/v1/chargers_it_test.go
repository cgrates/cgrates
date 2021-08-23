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
	chargerCfgPath   string
	chargerCfg       *config.CGRConfig
	chargerRPC       *rpc.Client
	chargerProfile   *ChargerWithCache
	chargerConfigDIR string //run tests for specific configuration

	chargerEvent = []*utils.CGREventWithArgDispatcher{
		{
			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account: "1001",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // no matching
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account:   "1010",
					"DistinctMatch": "cgrates",
				},
			},
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
	chargerCfg.DataFolderPath = *dataDir
	config.SetCgrConfig(chargerCfg)
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
							Rules:           "CGRateS.org",
							AllFiltersMatch: true,
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
	if err := chargerRPC.Call(utils.ChargerSv1GetChargersForEvent, chargerEvent[1], &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := chargerRPC.Call(utils.ChargerSv1GetChargersForEvent, chargerEvent[0], &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, chargerProfiles) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(chargerProfiles), utils.ToJSON(result))
	}
}

func testChargerSGetChargersForEvent2(t *testing.T) {
	var result *engine.ChargerProfiles
	if err := chargerRPC.Call(utils.ChargerSv1GetChargersForEvent,
		&utils.CGREventWithArgDispatcher{
			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account: "1015",
					utils.Usage:   1,
				},
			},
		}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
		t.Error(utils.ToJSON(result))
	}
}

func testChargerSProcessEvent(t *testing.T) {
	processedEv := &[]*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile:    "Charger1",
			AttributeSProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
			AlteredFields:      []string{utils.MetaReqRunID, "*req.Password"},
			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account: "1001",
					"Password":    "CGRateS.org",
					"RunID":       utils.MetaDefault,
				},
			},
		},
	}
	var result *[]*engine.ChrgSProcessEventReply
	if err := chargerRPC.Call(utils.ChargerSv1ProcessEvent, chargerEvent[1], &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := chargerRPC.Call(utils.ChargerSv1ProcessEvent, chargerEvent[0], &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, processedEv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(processedEv), utils.ToJSON(result))
	}
}

func testChargerSSetChargerProfile(t *testing.T) {
	chargerProfile = &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			FilterIDs: []string{"*string:~*req.Account:1001", "*string:~Account:1002"},
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

func testChargerSGetChargerProfileIDs(t *testing.T) {
	expected := []string{"Charger1", "ApierTest", "ChargerNotMatching"}
	var result []string
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfileIDs, utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{Tenant: "cgrates.org"}}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := chargerRPC.Call(utils.APIerSv1GetChargerProfileIDs, utils.TenantArgWithPaginator{
		TenantArg: utils.TenantArg{Tenant: "cgrates.org"},
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
		t.Errorf("Expected error: %v recived: %v", utils.ErrNotFound, err)
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

func testChargerSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
