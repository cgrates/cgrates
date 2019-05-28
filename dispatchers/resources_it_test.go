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
	"testing"

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
func TestDspResourceSITMySQL(t *testing.T) {
	testDsp(t, sTestsDspRes, "TestDspResourceS", "all", "all2", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspResourceSITMongo(t *testing.T) {
	testDsp(t, sTestsDspRes, "TestDspResourceS", "all", "all2", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

func testDspResPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ResourceSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
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
	if err := dispEngine.RCP.Call(utils.ResourceSv1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspResTestAuthKey(t *testing.T) {
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Destination: "1002",
			},
		},
		UsageID: utils.UUIDSha1Prefix(),
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("12345"),
		},
	}

	if err := dispEngine.RCP.Call(utils.ResourceSv1GetResourcesForEvent,
		args, &rs); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspResTestAuthKey2(t *testing.T) {
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Destination: "1002",
			},
		},
		UsageID: utils.UUIDSha1Prefix(),
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
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
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 1,
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
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
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 17,
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1AuthorizeResources,
		argsRU, &reply); err == nil || err.Error() != utils.ErrResourceUnauthorized.Error() {
		t.Error(err)
	}

	// relase the only resource active for Resource1
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e55",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1ReleaseResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	// try reserving with full units for Resource1, case which did not work in previous test
	// only match Resource1 since we don't want for storing of the resource2 bellow
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 6,
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.ResourceSv1AuthorizeResources, argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "ResGroup1" {
		t.Error("Unexpected reply returned", reply)
	}
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event5",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
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
			(len(r.Usages) != 1 || len(r.TTLIdx) != 0) {
			t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
		}
	}
}
