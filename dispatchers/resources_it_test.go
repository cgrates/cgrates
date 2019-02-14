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

var sTestsDspRes = []func(t *testing.T){
	testDspResPingFailover,

	testDspResPing,
	testDspResTestAuthKey,
	testDspResTestAuthKey2,
	testDspResTestAuthKey3,
}

//Test start here
func TestDspResourceS(t *testing.T) {
	engine.KillEngine(0)
	allEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all"), true, true)
	allEngine2 = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all2"), true, true)
	attrEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "attributes"), true, true)
	dispEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "dispatchers"), true, true)
	allEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "tutorial"))
	allEngine2.loadData(t, path.Join(dspDataDir, "tariffplans", "oldtutorial"))
	attrEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "dispatchers"))
	time.Sleep(500 * time.Millisecond)
	for _, stest := range sTestsDspRes {
		t.Run("", stest)
	}
	attrEngine.stopEngine(t)
	dispEngine.stopEngine(t)
	allEngine.stopEngine(t)
	allEngine2.stopEngine(t)
	engine.KillEngine(0)
}

func testDspResPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ResourceSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "res12345",
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ResourceSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ResourceSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspResPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ResourceSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "res12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspResTestAuthKey(t *testing.T) {
	var rs *engine.Resources
	args := &ArgsV1ResUsageWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "12345",
		},
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "1002",
				},
			},
		},
	}

	if err := dispEngine.RCP.Call(utils.ResourceSv1GetResourcesForEvent,
		args, &rs); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspResTestAuthKey2(t *testing.T) {
	var rs *engine.Resources
	args := &ArgsV1ResUsageWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "res12345",
		},
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "1002",
				},
			},
		},
	}
	eRs := &engine.Resources{
		&engine.Resource{
			Tenant: "cgrates.org",
			ID:     "ResGroup1",
			Usages: map[string]*engine.ResourceUsage{},
		},
	}

	if err := dispEngine.RCP.Call(utils.ResourceSv1GetResourcesForEvent,
		args, &rs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRs, rs) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRs), utils.ToJSON(rs))
	}
}

func testDspResTestAuthKey3(t *testing.T) {
	// first event matching Resource1
	var reply string
	argsRU := ArgsV1ResUsageWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "res12345",
		},
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					"Account":     "1002",
					"Subject":     "1001",
					"Destination": "1002"},
			},
			Units: 1,
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg := "ResGroup1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

	if err := dispEngine.RCP.Call(utils.ResourceSv1AuthorizeResources, argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != eAllocationMsg { // already 3 usages active before allow call, we should have now more than allowed
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	argsRU = ArgsV1ResUsageWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "res12345",
		},
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					"Account":     "1002",
					"Subject":     "1001",
					"Destination": "1002"},
			},
			Units: 17,
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1AuthorizeResources,
		argsRU, &reply); err == nil || err.Error() != utils.ErrResourceUnauthorized.Error() {
		t.Error(err)
	}

	// relase the only resource active for Resource1
	argsRU = ArgsV1ResUsageWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "res12345",
		},
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e55",
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					"Account":     "1002",
					"Subject":     "1001",
					"Destination": "1002"},
			},
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1ReleaseResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	// try reserving with full units for Resource1, case which did not work in previous test
	// only match Resource1 since we don't want for storing of the resource2 bellow
	argsRU = ArgsV1ResUsageWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "res12345",
		},
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					"Account":     "1002",
					"Subject":     "1001",
					"Destination": "1002"},
			},
			Units: 7,
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1AuthorizeResources, argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "ResGroup1" {
		t.Error("Unexpected reply returned", reply)
	}
	var rs *engine.Resources
	args := &ArgsV1ResUsageWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "res12345",
		},
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Event5",
				Event: map[string]interface{}{
					"Account":     "1002",
					"Subject":     "1001",
					"Destination": "1002"},
			},
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 1 {
		t.Errorf("Resources: %+v", utils.ToJSON(rs))
	}
	if rs == nil {
		t.Errorf("Expecting rs to not be nil")
		// rs shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	// make sure Resource1 have no more active resources
	for _, r := range *rs {
		if r.ID == "ResGroup1" &&
			(len(r.Usages) != 0 || len(r.TTLIdx) != 0) {
			t.Errorf("Unexpected resource: %+v", r)
		}
	}
}
