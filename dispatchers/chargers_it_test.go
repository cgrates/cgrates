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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspCpp = []func(t *testing.T){
	testDspCppPingFailover,
	testDspCppGetChtgFailover,

	testDspCppPing,
	testDspCppTestAuthKey,
	testDspCppTestAuthKey2,
}

//Test start here
func TestDspChargerS(t *testing.T) {
	allEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all"), true, true)
	allEngine2 = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all2"), true, true)
	attrEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "attributes"), true, true)
	dispEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "dispatchers"), true, true)
	allEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "tutorial"))
	allEngine2.loadData(t, path.Join(dspDataDir, "tariffplans", "oldtutorial"))
	attrEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "dispatchers"))
	time.Sleep(500 * time.Millisecond)
	for _, stest := range sTestsDspCpp {
		t.Run("", stest)
	}
	attrEngine.stopEngine(t)
	dispEngine.stopEngine(t)
	allEngine.stopEngine(t)
	allEngine2.stopEngine(t)
}

func testDspCppPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ChargerSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		APIKey: "chrg12345",
	}
	if err := dispEngine.RCP.Call(utils.ChargerSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ChargerSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ChargerSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspCppGetChtgFailover(t *testing.T) {
	args := CGREvWithApiKey{
		APIKey: "chrg12345",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.EVENT_NAME: "Event1",
				utils.Account:    "1001",
			},
		},
	}
	eChargers := &engine.ChargerProfiles{
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DEFAULT",
			FilterIDs:    []string{},
			RunID:        "*default",
			AttributeIDs: []string{"*none"},
			Weight:       0,
		},
	}
	var reply *engine.ChargerProfiles
	if err := dispEngine.RCP.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error NOT_FOUND but recived %v and reply %v\n", err, reply)
	}

	allEngine2.stopEngine(t)

	if err := dispEngine.RCP.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eChargers, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eChargers), utils.ToJSON(reply))
	}

	allEngine2.startEngine(t)
}

func testDspCppPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ChargerSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.ChargerSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		APIKey: "chrg12345",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCppTestAuthKey(t *testing.T) {
	args := CGREvWithApiKey{
		APIKey: "12345",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	var reply *engine.ChargerProfiles
	if err := dispEngine.RCP.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspCppTestAuthKey2(t *testing.T) {
	args := CGREvWithApiKey{
		APIKey: "chrg12345",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	eChargers := &engine.ChargerProfiles{
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DEFAULT",
			FilterIDs:    []string{},
			RunID:        "*default",
			AttributeIDs: []string{"*none"},
			Weight:       0,
		},
	}
	var reply *engine.ChargerProfiles
	if err := dispEngine.RCP.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eChargers, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eChargers), utils.ToJSON(reply))
	}
}
