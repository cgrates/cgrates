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

package dispatchers

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspCpp = []func(t *testing.T){
	testDspCppPingFailover,
	testDspCppGetChtgFailover,
	testDspCppGetChtgRoundRobin,

	testDspCppPing,
	testDspCppTestAuthKey,
	testDspCppTestAuthKey2,
}

//Test start here
func TestDspChargerST(t *testing.T) {
	var config1, config2, config3 string
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspCpp, "TestDspChargerS", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspCppPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.ChargerSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREvent{
		Tenant: "cgrates.org",

		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chrg12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ChargerSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.ChargerSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.ChargerSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspCppGetChtgFailover(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventName:    "Event1",
			utils.AccountField: "1001",
		},

		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chrg12345",
		},
	}
	eChargers := &engine.ChargerProfiles{
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DEFAULT",
			FilterIDs:    []string{},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       0,
		},
	}
	if *encoding == utils.MetaGOB {
		(*eChargers)[0].FilterIDs = nil // empty slice are nil in gob
	}
	var reply *engine.ChargerProfiles
	if err := dispEngine.RPC.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eChargers, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eChargers), utils.ToJSON(reply))
	}

	allEngine2.stopEngine(t)
	*eChargers = append(*eChargers,
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Raw",
			FilterIDs:    []string{},
			RunID:        utils.MetaRaw,
			AttributeIDs: []string{"*constant:*req.RequestType:*none"},
			Weight:       0,
		},
	)
	if *encoding == utils.MetaGOB {
		(*eChargers)[1].FilterIDs = nil // empty slice are nil in gob
	}
	if err := dispEngine.RPC.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Slice(*reply, func(i, j int) bool {
		return strings.Compare((*reply)[i].ID, (*reply)[j].ID) < 0
	})
	if !reflect.DeepEqual(eChargers, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eChargers), utils.ToJSON(reply))
	}

	allEngine2.startEngine(t)
}

func testDspCppPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.ChargerSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.ChargerSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chrg12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCppTestAuthKey(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "12345",
		},
	}
	var reply *engine.ChargerProfiles
	if err := dispEngine.RPC.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspCppTestAuthKey2(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},

		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chrg12345",
		},
	}
	eChargers := &engine.ChargerProfiles{
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DEFAULT",
			FilterIDs:    []string{},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       0,
		},
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Raw",
			FilterIDs:    []string{},
			RunID:        utils.MetaRaw,
			AttributeIDs: []string{"*constant:*req.RequestType:*none"},
			Weight:       0,
		},
	}
	if *encoding == utils.MetaGOB {
		(*eChargers)[0].FilterIDs = nil // empty slice are nil in gob
		(*eChargers)[1].FilterIDs = nil // empty slice are nil in gob
	}
	var reply *engine.ChargerProfiles
	if err := dispEngine.RPC.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Slice(*reply, func(i, j int) bool {
		return strings.Compare((*reply)[i].ID, (*reply)[j].ID) < 0
	})
	if !reflect.DeepEqual(eChargers, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eChargers), utils.ToJSON(reply))
	}
}

func testDspCppGetChtgRoundRobin(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventName:    "RoundRobin",
			utils.AccountField: "1001",
		},
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chrg12345",
		},
	}
	eChargers := &engine.ChargerProfiles{
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DEFAULT",
			FilterIDs:    []string{},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       0,
		},
	}
	if *encoding == utils.MetaGOB {
		(*eChargers)[0].FilterIDs = nil // empty slice are nil in gob
	}
	var reply *engine.ChargerProfiles
	// To ALL2
	if err := dispEngine.RPC.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eChargers, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eChargers), utils.ToJSON(reply))
	}
	// To ALL
	*eChargers = append(*eChargers,
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Raw",
			FilterIDs:    []string{},
			RunID:        utils.MetaRaw,
			AttributeIDs: []string{"*constant:*req.RequestType:*none"},
			Weight:       0,
		},
	)
	if *encoding == utils.MetaGOB {
		(*eChargers)[1].FilterIDs = nil // empty slice are nil in gob
	}
	if err := dispEngine.RPC.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Slice(*reply, func(i, j int) bool {
		return strings.Compare((*reply)[i].ID, (*reply)[j].ID) < 0
	})
	if !reflect.DeepEqual(eChargers, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eChargers), utils.ToJSON(reply))
	}

}
