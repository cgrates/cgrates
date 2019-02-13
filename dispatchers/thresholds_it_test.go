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
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspTh = []func(t *testing.T){
	testDspThPingFailover,
	testDspThProcessEventFailover,

	testDspThPing,
	testDspThTestAuthKey,
	testDspThTestAuthKey2,
	testDspThTestAuthKey3,
}

//Test start here
func TestDspThresholdS(t *testing.T) {
	engine.KillEngine(0)
	allEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all"), true, true)
	allEngine2 = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all2"), true, true)
	attrEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "attributes"), true, true)
	dispEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "dispatchers"), true, true)
	allEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "tutorial"))
	allEngine2.loadData(t, path.Join(dspDataDir, "tariffplans", "oldtutorial"))
	attrEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "dispatchers"))
	time.Sleep(500 * time.Millisecond)
	for _, stest := range sTestsDspTh {
		t.Run("", stest)
	}
	attrEngine.stopEngine(t)
	dispEngine.stopEngine(t)
	allEngine.stopEngine(t)
	allEngine2.stopEngine(t)
	engine.KillEngine(0)
}

func testDspThPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ThresholdSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "thr12345",
		},
	}
	if err := dispEngine.RCP.Call(utils.ThresholdSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ThresholdSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ThresholdSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspThProcessEventFailover(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1001"}
	nowTime := time.Now()
	args := &ArgsProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "thr12345",
		},
		ArgsProcessEvent: engine.ArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Time:   &nowTime,
				Event: map[string]interface{}{
					utils.EVENT_NAME: "Event1",
					utils.Account:    "1001"},
			},
		},
	}

	if err := dispEngine.RCP.Call(utils.ThresholdSv1ProcessEvent, args,
		&ids); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error NOT_FOUND but recived %v and reply %v\n", err, ids)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
	allEngine2.startEngine(t)
}

func testDspThPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ThresholdSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.ThresholdSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "thr12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspThTestAuthKey(t *testing.T) {
	var ids []string
	nowTime := time.Now()
	args := &ArgsProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "12345",
		},
		ArgsProcessEvent: engine.ArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Time:   &nowTime,
				Event: map[string]interface{}{
					utils.Account: "1002"},
			},
		},
	}

	if err := dispEngine.RCP.Call(utils.ThresholdSv1ProcessEvent,
		args, &ids); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
	var th *engine.Thresholds
	if err := dispEngine.RCP.Call(utils.ThresholdSv1GetThresholdsForEvent, args,
		&th); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspThTestAuthKey2(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1002"}
	nowTime := time.Now()
	args := &ArgsProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "thr12345",
		},
		ArgsProcessEvent: engine.ArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Time:   &nowTime,
				Event: map[string]interface{}{
					utils.Account: "1002"},
			},
		},
	}

	if err := dispEngine.RCP.Call(utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
	var th *engine.Thresholds
	eTh := &engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
			Hits:   1,
		},
	}
	if err := dispEngine.RCP.Call(utils.ThresholdSv1GetThresholdsForEvent, args, &th); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual((*eTh)[0].Tenant, (*th)[0].Tenant) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].Tenant, (*th)[0].Tenant)
	} else if !reflect.DeepEqual((*eTh)[0].ID, (*th)[0].ID) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].ID, (*th)[0].ID)
	} else if !reflect.DeepEqual((*eTh)[0].Hits, (*th)[0].Hits) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].Hits, (*th)[0].Hits)
	}
}

func testDspThTestAuthKey3(t *testing.T) {
	var th *engine.Threshold
	eTh := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
		Hits:   1,
	}
	if err := dispEngine.RCP.Call(utils.ThresholdSv1GetThreshold, &TntIDWithApiKey{
		TenantID: utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "thr12345",
		},
	}, &th); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual((*eTh).Tenant, (*th).Tenant) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).Tenant, (*th).Tenant)
	} else if !reflect.DeepEqual((*eTh).ID, (*th).ID) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).ID, (*th).ID)
	} else if !reflect.DeepEqual((*eTh).Hits, (*th).Hits) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).Hits, (*th).Hits)
	}

	var ids []string
	eIDs := []string{"THD_ACNT_1002"}

	if err := dispEngine.RCP.Call(utils.ThresholdSv1GetThresholdIDs, &TntWithApiKey{
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "thr12345",
		},
	}, &ids); err != nil {
		t.Fatal(err)
	}
	sort.Strings(ids)
	if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
}
