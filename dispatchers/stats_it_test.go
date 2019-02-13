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

var sTestsDspSts = []func(t *testing.T){
	testDspStsPingFailover,
	testDspStsGetStatFailover,

	testDspStsPing,
	testDspStsTestAuthKey,
	testDspStsTestAuthKey2,
	testDspStsTestAuthKey3,
}

//Test start here
func TestDspStatS(t *testing.T) {
	engine.KillEngine(0)
	allEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all"), true, true)
	allEngine2 = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all2"), true, true)
	attrEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "attributes"), true, true)
	dispEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "dispatchers"), true, true)
	allEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "tutorial"))
	allEngine2.loadData(t, path.Join(dspDataDir, "tariffplans", "oldtutorial"))
	attrEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "dispatchers"))
	time.Sleep(500 * time.Millisecond)
	for _, stest := range sTestsDspSts {
		t.Run("", stest)
	}
	attrEngine.stopEngine(t)
	dispEngine.stopEngine(t)
	allEngine.stopEngine(t)
	allEngine2.stopEngine(t)
	engine.KillEngine(0)
}

func testDspStsPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.StatSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "stat12345",
		},
	}
	if err := dispEngine.RCP.Call(utils.StatSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.StatSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.StatSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspStsGetStatFailover(t *testing.T) {
	var reply []string
	var metrics map[string]string
	expected := []string{"Stats1"}
	args := ArgsStatProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "stat12345",
		},
		StatsArgsProcessEvent: engine.StatsArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.EVENT_NAME:  "Event1",
					utils.Account:     "1001",
					utils.AnswerTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:       time.Duration(135 * time.Second),
					utils.COST:        123.0,
					utils.RunID:       utils.DEFAULT_RUNID,
					utils.Destination: "1002"},
			},
		},
	}
	if err := dispEngine.RCP.Call(utils.StatSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	args2 := TntIDWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "stat12345",
		},
		TenantID: utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats1",
		},
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	}

	allEngine.startEngine(t)
	allEngine2.stopEngine(t)

	if err := dispEngine.RCP.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error NOT_FOUND but recived %v and reply %v\n", err, reply)
	}
	allEngine2.startEngine(t)
}

func testDspStsPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.StatSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.StatSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "stat12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspStsTestAuthKey(t *testing.T) {
	var reply []string
	args := ArgsStatProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "12345",
		},
		StatsArgsProcessEvent: engine.StatsArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account:    "1001",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:      time.Duration(135 * time.Second),
					utils.COST:       123.0,
					utils.PDD:        time.Duration(12 * time.Second)}},
		}}
	if err := dispEngine.RCP.Call(utils.StatSv1ProcessEvent,
		args, &reply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}

	args2 := TntIDWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "12345",
		},
		TenantID: utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		},
	}

	var metrics map[string]string
	if err := dispEngine.RCP.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspStsTestAuthKey2(t *testing.T) {
	var reply []string
	var metrics map[string]string
	expected := []string{"Stats2"}
	args := ArgsStatProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "stat12345",
		},
		StatsArgsProcessEvent: engine.StatsArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.AnswerTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:       time.Duration(135 * time.Second),
					utils.COST:        123.0,
					utils.RunID:       utils.DEFAULT_RUNID,
					utils.Destination: "1002"},
			},
		},
	}
	if err := dispEngine.RCP.Call(utils.StatSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	args2 := TntIDWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "stat12345",
		},
		TenantID: utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		},
	}
	expectedMetrics := map[string]string{
		utils.MetaTCC: "123",
		utils.MetaTCD: "2m15s",
	}

	if err := dispEngine.RCP.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}

	args = ArgsStatProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "stat12345",
		},
		StatsArgsProcessEvent: engine.StatsArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account:     "1002",
					utils.AnswerTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:       time.Duration(45 * time.Second),
					utils.RunID:       utils.DEFAULT_RUNID,
					utils.COST:        10.0,
					utils.Destination: "1001",
				},
			},
		},
	}
	if err := dispEngine.RCP.Call(utils.StatSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expectedMetrics = map[string]string{
		utils.MetaTCC: "133",
		utils.MetaTCD: "3m0s",
	}
	if err := dispEngine.RCP.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testDspStsTestAuthKey3(t *testing.T) {
	var reply []string
	var metrics map[string]float64

	args2 := TntIDWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "stat12345",
		},
		TenantID: utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		},
	}
	expectedMetrics := map[string]float64{
		utils.MetaTCC: 133,
		utils.MetaTCD: 180,
	}

	if err := dispEngine.RCP.Call(utils.StatSv1GetQueueFloatMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %v", expectedMetrics, metrics)
	}

	estats := []string{"Stats2", "Stats2_1"}
	if err := dispEngine.RCP.Call(utils.StatSv1GetQueueIDs,
		&TntWithApiKey{
			TenantArg: utils.TenantArg{
				Tenant: "cgrates.org",
			},
			DispatcherResource: DispatcherResource{
				APIKey: "stat12345",
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(estats, reply) {
		t.Errorf("expecting: %+v, received reply: %v", estats, reply)
	}

}
