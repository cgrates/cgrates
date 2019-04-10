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

var sTestsDspDspv1 = []func(t *testing.T){
	testDspDspv1GetProfileForEvent,
}

//Test start here
func TestDspDspv1SMySQL(t *testing.T) {
	engine.KillEngine(0)
	dispEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "dispatchers"), true, true)
	dispEngine.loadData2(t, path.Join(dspDataDir, "tariffplans", "dispatchers"))
	time.Sleep(500 * time.Millisecond)
	for _, stest := range sTestsDspDspv1 {
		t.Run("TestDspDspv1", stest)
	}
	dispEngine.stopEngine(t)
	engine.KillEngine(0)
}

func TestDspDspv1SMongo(t *testing.T) {
	engine.KillEngine(0)
	dispEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "dispatchers_mongo"), true, true)
	dispEngine.loadData2(t, path.Join(dspDataDir, "tariffplans", "dispatchers"))
	time.Sleep(500 * time.Millisecond)
	for _, stest := range sTestsDspDspv1 {
		t.Run("TestDspDspv1", stest)
	}
	dispEngine.stopEngine(t)
	engine.KillEngine(0)
}

func testDspDspv1GetProfileForEvent(t *testing.T) {
	arg := DispatcherEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testDspv1",
			Event: map[string]interface{}{
				utils.EVENT_NAME: "Event1",
			},
		},
		Subsystem: utils.META_ANY,
	}
	var reply engine.DispatcherProfile
	expected := engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "EVENT1",
		Subsystems:     []string{utils.META_ANY},
		FilterIDs:      []string{"*string:~EventName:Event1"},
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         30,
		Hosts: engine.DispatcherHostProfiles{
			&engine.DispatcherHostProfile{
				ID:        "ALL2",
				FilterIDs: []string{},
				Weight:    20,
				Params:    make(map[string]interface{}),
			},
			&engine.DispatcherHostProfile{
				ID:        "ALL",
				FilterIDs: []string{},
				Weight:    10,
				Params:    make(map[string]interface{}),
			},
		},
	}
	if err := dispEngine.RCP.Call(utils.DispatcherSv1GetProfileForEvent, &arg, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}
