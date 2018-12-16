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
	rlsV1CfgPath string
	rlsV1Cfg     *config.CGRConfig
	rlsV1Rpc     *rpc.Client
	rlsV1ConfDIR string //run tests for specific configuration
	rlsConfig    *engine.ResourceProfile
)

var sTestsRLSV1 = []func(t *testing.T){
	testV1RsLoadConfig,
	testV1RsInitDataDb,
	testV1RsResetStorDb,
	testV1RsStartEngine,
	testV1RsRpcConn,
	testV1RsFromFolder,
	testV1RsGetResourcesForEvent,
	testV1RsTTL0,
	testV1RsAllocateResource,
	testV1RsAuthorizeResources,
	testV1RsReleaseResource,
	testV1RsDBStore,
	testV1RsGetResourceProfileBeforeSet,
	testV1RsSetResourceProfile,
	testV1RsGetResourceProfileIDs,
	testV1RsGetResourceProfileAfterSet,
	testV1RsUpdateResourceProfile,
	testV1RsGetResourceProfileAfterUpdate,
	testV1RsRemResourceProfile,
	testV1RsGetResourceProfileAfterDelete,
	testV1RsResourcePing,
	testV1RsStopEngine,
}

//Test start here
func TestRsV1ITMySQL(t *testing.T) {
	rlsV1ConfDIR = "tutmysql"
	for _, stest := range sTestsRLSV1 {
		t.Run(rlsV1ConfDIR, stest)
	}
}

func TestRsV1ITMongo(t *testing.T) {
	rlsV1ConfDIR = "tutmongo"
	for _, stest := range sTestsRLSV1 {
		t.Run(rlsV1ConfDIR, stest)
	}
}

func testV1RsLoadConfig(t *testing.T) {
	var err error
	rlsV1CfgPath = path.Join(*dataDir, "conf", "samples", rlsV1ConfDIR)
	if rlsV1Cfg, err = config.NewCGRConfigFromFolder(rlsV1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1RsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rlsV1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RsRpcConn(t *testing.T) {
	var err error
	rlsV1Rpc, err = jsonrpc.Dial("tcp", rlsV1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1RsFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := rlsV1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)

}

func testV1RsGetResourcesForEvent(t *testing.T) {
	var reply *[]*engine.ResourceProfile
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event1",
			Event:  map[string]interface{}{"Unknown": "unknown"},
		},
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	args.CGREvent.Event = map[string]interface{}{"Destination": "10", "Account": "1001"}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &reply); err != nil {
		t.Error(err)
	}
	if reply == nil {
		t.Errorf("Expecting reply to not be nil")
		// reply shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	if len(*reply) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(*reply))
	}
	if (*reply)[0].ID != "ResGroup2" {
		t.Errorf("Expecting: %+v, received: %+v", "ResGroup2", (*reply)[0].ID)
	}

	args.CGREvent.Event = map[string]interface{}{"Destination": "20"}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	args.CGREvent.Event = map[string]interface{}{"Account": "1002", "Subject": "test", "Destination": "1002"}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &reply); err != nil {
		t.Error(err)
	}
	if len(*reply) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 2, len(*reply))
	}

	args.CGREvent.Event = map[string]interface{}{"Account": "1002", "Subject": "test", "Destination": "1001"}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &reply); err != nil {
		t.Error(err)
	}
	if len(*reply) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(*reply))
	}
	if (*reply)[0].ID != "ResGroup2" {
		t.Errorf("Expecting: %+v, received: %+v", "ResGroup2", (*reply)[0].ID)
	}
}

func testV1RsTTL0(t *testing.T) {
	// only matching Resource3
	argsRU := utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "3001",
				"Destination": "3002"},
		},
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
		Units:   1,
	}
	var reply string
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	// overwrite the first allocation
	argsRU = utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "3001",
				"Destination": "3002"},
		},
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
		Units:   2,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources, argsRU, &reply); err != nil {
		t.Error(err)
	}
	// too many units should be rejected
	argsRU = utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "3001",
				"Destination": "3002"},
		},
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e22",
		Units:   4,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources, argsRU, &reply); err == nil ||
		err.Error() != utils.ErrResourceUnavailable.Error() {
		t.Error(err)
	}
	// check the record
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event2",
			Event: map[string]interface{}{
				"Account":     "3001",
				"Destination": "3002"},
		},
	}
	expiryTime, err := utils.ParseTimeDetectLayout("0001-01-01T00:00:00Z", "")
	if err != nil {
		t.Error(err)
	}
	expectedResources := &engine.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup3",
		Usages: map[string]*engine.ResourceUsage{
			"651a8db2-4f67-4cf8-b622-169e8a482e21": {
				Tenant:     "cgrates.org",
				ID:         "651a8db2-4f67-4cf8-b622-169e8a482e21",
				ExpiryTime: expiryTime,
				Units:      2,
			},
		},
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent,
		args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 1 {
		t.Errorf("Resources: %+v", rs)
	} else {
		res := *rs
		if !reflect.DeepEqual(expectedResources.Tenant, res[0].Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", expectedResources.Tenant, res[0].Tenant)
		} else if !reflect.DeepEqual(expectedResources.ID, res[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", expectedResources.ID, res[0].ID)
		} else if !reflect.DeepEqual(expectedResources.Usages, res[0].Usages) {
			t.Errorf("Expecting: %+v, received: %+v", expectedResources.Usages, res[0].Usages)
		}
	}
	// release should not give out errors
	var releaseReply string
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e25", // same ID should be accepted by first group since the previous resource should be expired
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "3001",
				"Destination": "3002"},
		},
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1ReleaseResources,
		argsRU, &releaseReply); err != nil {
		t.Error(err)
	}
}

func testV1RsAllocateResource(t *testing.T) {
	// first event matching Resource1
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 3,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg := "ResGroup1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	// Second event to test matching of exact limit of first resource
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e52",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 4,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResGroup1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	// Third event testing overflow to second resource which still has one resource available
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e53",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "dan",
				"Subject":     "dan",
				"Destination": "1002"},
		},
		Units: 1,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "SPECIAL_1002"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	// Test resource unavailable
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e54", // same ID should be accepted by first group since the previous resource should be expired
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		}, Units: 1,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err == nil || err.Error() != utils.ErrResourceUnavailable.Error() {
		t.Error(err)
	}
	eAllocationMsg = "ResGroup1"
	time.Sleep(time.Second) // Give time for allocations on first resource to expire

	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e55", // same ID should be accepted by first group since the previous resource should be expired
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 1,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResGroup1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
}

func testV1RsAuthorizeResources(t *testing.T) {
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 6,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AuthorizeResources, argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "ResGroup1" { // already 3 usages active before allow call, we should have now more than allowed
		t.Error("Unexpected reply returned", reply)
	}
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 7,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AuthorizeResources,
		argsRU, &reply); err.Error() != utils.ErrResourceUnauthorized.Error() {
		t.Error(err)
	}
}

func testV1RsReleaseResource(t *testing.T) {
	// relase the only resource active for Resource1
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e55", // same ID should be accepted by first group since the previous resource should be expired
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1ReleaseResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	// try reserving with full units for Resource1, case which did not work in previous test
	// only match Resource1 since we don't want for storing of the resource2 bellow
	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 7,
	}
	if err := rlsV1Rpc.Call(utils.ResourceSv1AuthorizeResources, argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "ResGroup1" {
		t.Error("Unexpected reply returned", reply)
	}
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event5",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		}}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 2 {
		t.Errorf("Resources: %+v", rs)
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

func testV1RsDBStore(t *testing.T) {
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e71",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 1,
	}
	var reply string
	eAllocationMsg := "ResGroup1"
	if err := rlsV1Rpc.Call(utils.ResourceSv1AllocateResources, argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event3",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		}}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 2 {
		t.Errorf("Resources: %+v", rs)
	}
	if rs == nil {
		t.Errorf("Expecting rs to not be nil")
		// rs shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	// count resources before restart
	for _, r := range *rs {
		switch r.ID {
		case "ResGroup1":
			if len(r.Usages) != 1 || len(r.TTLIdx) != 1 {
				t.Errorf("Unexpected resource: %+v", r)
			}
		case "ResGroup2":
			if len(r.Usages) != 4 || len(r.TTLIdx) != 4 {
				t.Errorf("Unexpected resource: %+v", r)
			}
		}
	}
	if _, err := engine.StopStartEngine(rlsV1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	var err error
	rlsV1Rpc, err = jsonrpc.Dial("tcp", rlsV1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	rs = new(engine.Resources)
	args = &utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event4",
			Event: map[string]interface{}{
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		}}
	if err := rlsV1Rpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 2 {
		t.Errorf("Resources: %+v", rs)
	}
	// count resources after restart
	for _, r := range *rs {
		switch r.ID {
		case "ResGroup1":
			if len(r.Usages) != 0 || len(r.TTLIdx) != 0 {
				t.Errorf("Unexpected resource: %+v", r)
			}
		case "ResGroup2":
			if len(r.Usages) != 3 || len(r.TTLIdx) != 3 {
				t.Errorf("Unexpected resource: %s", utils.ToJSON(r))
			}
		}
	}
}

func testV1RsGetResourceProfileBeforeSet(t *testing.T) {
	var reply *string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "RCFG1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RsSetResourceProfile(t *testing.T) {
	rlsConfig = &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_GR_TEST",
		FilterIDs: []string{"*string:Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(1) * time.Nanosecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{"Val1"},
	}
	var result string

	if err := rlsV1Rpc.Call("ApierV1.SetResourceProfile", rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1RsGetResourceProfileIDs(t *testing.T) {
	expected := []string{"ResGroup2", "ResGroup1", "ResGroup3", "RES_GR_TEST"}
	var result []string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceProfileIDs", "cgrates.org", &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testV1RsGetResourceProfileAfterSet(t *testing.T) {
	var reply *engine.ResourceProfile
	if err := rlsV1Rpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig), utils.ToJSON(reply))
	}
}

func testV1RsUpdateResourceProfile(t *testing.T) {
	var result string
	rlsConfig.FilterIDs = []string{"*string:Account:1001", "*prefix:DST:10"}
	if err := rlsV1Rpc.Call("ApierV1.SetResourceProfile", rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1RsGetResourceProfileAfterUpdate(t *testing.T) {
	var reply *engine.ResourceProfile
	if err := rlsV1Rpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig), utils.ToJSON(reply))
	}
}

func testV1RsRemResourceProfile(t *testing.T) {
	var resp string
	if err := rlsV1Rpc.Call("ApierV1.RemoveResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1RsGetResourceProfileAfterDelete(t *testing.T) {
	var reply *string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "RCFG1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RsResourcePing(t *testing.T) {
	var resp string
	if err := rlsV1Rpc.Call(utils.ResourceSv1Ping, "", &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1RsStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
