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
	"net/rpc/jsonrpc"
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
	chargerProfile   *engine.ChargerProfile
	chargerConfigDIR string //run tests for specific configuration
)

var chargerEvent = []*utils.CGREvent{
	{ // matching Charger1
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.Account: "1001",
		},
	},
	{ // no matching
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.Account:   "1010",
			"DistinctMatch": "cgrates",
		},
	},
}

var sTestsCharger = []func(t *testing.T){
	testChargerSInitCfg,
	testChargerSInitDataDb,
	testChargerSResetStorDb,
	testChargerSStartEngine,
	testChargerSRPCConn,
	testChargerSLoadAddCharger,
	testChargerSGetChargersForEvent,
	testChargerSProcessEvent,
	testChargerSSetChargerProfile,
	testChargerSGetChargerProfileIDs,
	testChargerSUpdateChargerProfile,
	testChargerSRemChargerProfile,
	testChargerSPing,
	testChargerSKillEngine,
}

//Test start here
func TestChargerSITMySql(t *testing.T) {
	chargerConfigDIR = "tutmysql"
	for _, stest := range sTestsCharger {
		t.Run(chargerConfigDIR, stest)
	}
}

func TestChargerSITMongo(t *testing.T) {
	chargerConfigDIR = "tutmongo"
	for _, stest := range sTestsCharger {
		t.Run(chargerConfigDIR, stest)
	}
}

func testChargerSInitCfg(t *testing.T) {
	var err error
	chargerCfgPath = path.Join(*dataDir, "conf", "samples", chargerConfigDIR)
	chargerCfg, err = config.NewCGRConfigFromFolder(chargerCfgPath)
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
	chargerRPC, err = jsonrpc.Dial("tcp", chargerCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testChargerSLoadAddCharger(t *testing.T) {
	chargerProfile := &engine.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"*string:Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		RunID:        "*default",
		AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
		Weight:       20,
	}
	var result string
	if err := chargerRPC.Call("ApierV1.SetChargerProfile", chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf = &engine.AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR_1001_SIMPLEAUTH",
		Contexts: []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FieldName: "Password",
				Initial:   utils.ANY,
				Substitute: config.RSRParsers{
					&config.RSRParser{
						Rules:           "CGRateS.org",
						AllFiltersMatch: true,
					},
				},
				Append: true,
			},
		},
		Blocker: false,
		Weight:  10,
	}
	if err := chargerRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
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
			FilterIDs: []string{"*string:Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			RunID:        "*default",
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

func testChargerSProcessEvent(t *testing.T) {
	processedEv := &[]*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile:    "Charger1",
			AttributeSProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
			AlteredFields:      []string{"Password"},
			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant:  "cgrates.org",
				ID:      "event1",
				Context: utils.StringPointer(utils.MetaChargers),
				Event: map[string]interface{}{
					utils.Account: "1001",
					"Password":    "CGRateS.org",
					"RunID":       "*default",
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
	chargerProfile = &engine.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		FilterIDs: []string{"*string:Account:1001", "*string:Account:1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		RunID:        "*default",
		AttributeIDs: []string{"Attr1", "Attr2"},
		Weight:       20,
	}
	var result string
	if err := chargerRPC.Call("ApierV1.SetChargerProfile", chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ChargerProfile
	if err := chargerRPC.Call("ApierV1.GetChargerProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile, reply)
	}
}

func testChargerSGetChargerProfileIDs(t *testing.T) {
	expected := []string{"Charger1", "ApierTest"}
	var result []string
	if err := chargerRPC.Call("ApierV1.GetChargerProfileIDs", "cgrates.org", &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testChargerSUpdateChargerProfile(t *testing.T) {
	chargerProfile.RunID = "*rated"
	var result string
	if err := chargerRPC.Call("ApierV1.SetChargerProfile", chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ChargerProfile
	if err := chargerRPC.Call("ApierV1.GetChargerProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile, reply)
	}
}

func testChargerSRemChargerProfile(t *testing.T) {
	var resp string
	if err := chargerRPC.Call("ApierV1.RemoveChargerProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.AttributeProfile
	if err := chargerRPC.Call("ApierV1.GetChargerProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testChargerSPing(t *testing.T) {
	var resp string
	if err := chargerRPC.Call(utils.ChargerSv1Ping, "", &resp); err != nil {
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
