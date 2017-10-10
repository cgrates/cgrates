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
	resDelay     int
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
	testV1RsAllowUsage,
	testV1RsReleaseResource,
	testV1RsDBStore,
	testV1RsGetResourceProfileBeforeSet,
	testV1RsSetResourceProfile,
	testV1RsGetResourceProfileAfterSet,
	testV1RsUpdateResourceProfile,
	testV1RsGetResourceProfileAfterUpdate,
	testV1RsRemResourceProfile,
	testV1RsGetResourceProfileAfterDelete,
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
	switch rlsV1ConfDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		resDelay = 4000
	default:
		resDelay = 1000
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
	if _, err := engine.StopStartEngine(rlsV1CfgPath, resDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1RsRpcConn(t *testing.T) {
	var err error
	rlsV1Rpc, err = jsonrpc.Dial("tcp", rlsV1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1RsFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := rlsV1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(1000) * time.Millisecond)

}

func testV1RsGetResourcesForEvent(t *testing.T) {
	var reply *[]*engine.ResourceProfile
	args := &utils.ArgRSv1ResourceUsage{
		Tenant: "cgrates.org",
		Event:  map[string]interface{}{"Unknown": "unknown"}}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	time.Sleep(time.Duration(500) * time.Millisecond)
	args.Event = map[string]interface{}{"Destination": "10"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(500) * time.Millisecond)
	if len(*reply) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(*reply))
	}
	if (*reply)[0].ID != "ResGroup2" {
		t.Errorf("Expecting: %+v, received: %+v", "ResGroup2", (*reply)[0].ID)
	}

	args.Event = map[string]interface{}{"Destination": "20"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	args.Event = map[string]interface{}{"Account": "1002", "Subject": "test", "Destination": "1002"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &reply); err != nil {
		t.Error(err)
	}
	if len(*reply) != 2 {
		t.Errorf("Expecting: %+v, received: %+v", 2, len(*reply))
	}

	args.Event = map[string]interface{}{"Account": "1002", "Subject": "test", "Destination": "1001"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &reply); err != nil {
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
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
		Event: map[string]interface{}{
			"Account":     "3001",
			"Destination": "3002"},
		Units: 1,
	}
	var reply string
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err != nil {
		t.Error(err)
	}
	// second allocation should be also allowed
	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
		Event: map[string]interface{}{
			"Account":     "3001",
			"Destination": "3002"},
		Units: 1,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err != nil {
		t.Error(err)
	}
	// too many units should be rejected
	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e22",
		Event: map[string]interface{}{
			"Account":     "3001",
			"Destination": "3002"},
		Units: 2,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err == nil ||
		err.Error() != utils.ErrResourceUnavailable.Error() {
		t.Error(err)
	}
	// make sure no usage was recorded
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "3001",
			"Destination": "3002"}}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 1 {
		t.Errorf("Resources: %+v", rs)
	} else {
		res := *rs
		if len(res[0].Usages) != 0 || len(res[0].TTLIdx) != 0 {
			t.Errorf("Resource should have no usage records in: %+v", res[0])
		}
	}
	// release should not give out errors
	var releaseReply string
	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e25", // same ID should be accepted by first group since the previous resource should be expired
		Event: map[string]interface{}{
			"Account":     "3001",
			"Destination": "3002"},
	}
	if err := rlsV1Rpc.Call("ResourceSV1.ReleaseResource", argsRU, &releaseReply); err != nil {
		t.Error(err)
	}
}

func testV1RsAllocateResource(t *testing.T) {
	// first event matching Resource1
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"},
		Units: 3,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg := "ResGroup1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	// Second event to test matching of exact limit of first resource
	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e52",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"},
		Units: 4,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResGroup1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	// Third event testing overflow to second resource which still has one resource available
	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e53",
		Event: map[string]interface{}{
			"Account":     "dan",
			"Subject":     "dan",
			"Destination": "1002"},
		Units: 1,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResGroup2"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	// Test resource unavailable
	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e54", // same ID should be accepted by first group since the previous resource should be expired
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"},
		Units: 1,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err == nil || err.Error() != utils.ErrResourceUnavailable.Error() {
		t.Error(err)
	}
	eAllocationMsg = "ResGroup1"
	time.Sleep(time.Duration(1000) * time.Millisecond) // Give time for allocations on first resource to expire

	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e55", // same ID should be accepted by first group since the previous resource should be expired
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"},
		Units: 1,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResGroup1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
}

func testV1RsAllowUsage(t *testing.T) {
	var allowed bool
	argsRU := utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"},
		Units: 6,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllowUsage", argsRU, &allowed); err != nil {
		t.Error(err)
	} else if !allowed { // already 3 usages active before allow call, we should have now more than allowed
		t.Error("resource is not allowed")
	}
	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"},
		Units: 7,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllowUsage", argsRU, &allowed); err != nil {
		t.Error(err)
	} else if allowed { // already 3 usages active before allow call, we should have now more than allowed
		t.Error("resource should not be allowed")
	}
}

func testV1RsReleaseResource(t *testing.T) {
	// relase the only resource active for Resource1
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e55", // same ID should be accepted by first group since the previous resource should be expired
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"},
	}
	if err := rlsV1Rpc.Call("ResourceSV1.ReleaseResource", argsRU, &reply); err != nil {
		t.Error(err)
	}
	// try reserving with full units for Resource1, case which did not work in previous test
	// only match Resource1 since we don't want for storing of the resource2 bellow
	argsRU = utils.ArgRSv1ResourceUsage{
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "2002"},
		Units: 7,
	}
	var allowed bool
	if err := rlsV1Rpc.Call("ResourceSV1.AllowUsage", argsRU, &allowed); err != nil {
		t.Error(err)
	} else if !allowed {
		t.Error("resource should be allowed")
	}
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"}}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 2 {
		t.Errorf("Resources: %+v", rs)
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
		Tenant:  "cgrates.org",
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e71",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"},
		Units: 1,
	}
	var reply string
	eAllocationMsg := "ResGroup1"
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"}}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 2 {
		t.Errorf("Resources: %+v", rs)
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
	if _, err := engine.StopStartEngine(rlsV1CfgPath, resDelay); err != nil {
		t.Fatal(err)
	}
	var err error
	rlsV1Rpc, err = jsonrpc.Dial("tcp", rlsV1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	rs = new(engine.Resources)
	args = &utils.ArgRSv1ResourceUsage{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "1002",
			"Subject":     "1001",
			"Destination": "1002"}}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", args, &rs); err != nil {
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
				t.Errorf("Unexpected resource: %+v", r)
			}
		}
	}
	time.Sleep(time.Duration(1) * time.Second)
}

func testV1RsGetResourceProfileBeforeSet(t *testing.T) {
	var reply *string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "RCFG1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RsSetResourceProfile(t *testing.T) {
	rlsConfig = &engine.ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "RCFG1",
		Filters: []*engine.Filter{
			&engine.Filter{
				Type:      "type",
				FieldName: "Name",
				Values:    []string{"FilterValue1", "FilterValue2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
		UsageTTL:          time.Duration(10) * time.Microsecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		Thresholds:        []string{"Val1", "Val2"},
	}
	var result string
	if err := rlsV1Rpc.Call("ApierV1.SetResourceProfile", rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1RsGetResourceProfileAfterSet(t *testing.T) {
	var reply *engine.ResourceProfile
	if err := rlsV1Rpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", rlsConfig, reply)
	}
}

func testV1RsUpdateResourceProfile(t *testing.T) {
	var result string
	rlsConfig.Filters = []*engine.Filter{
		&engine.Filter{
			Type:      "type",
			FieldName: "Name",
			Values:    []string{"FilterValue1", "FilterValue2"},
		},
		&engine.Filter{
			Type:      "*string",
			FieldName: "Accout",
			Values:    []string{"1001", "1002"},
		},
	}
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
		t.Errorf("Expecting: %+v, received: %+v", rlsConfig, reply)
	}
}

func testV1RsRemResourceProfile(t *testing.T) {
	var resp string
	if err := rlsV1Rpc.Call("ApierV1.RemResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1RsGetResourceProfileAfterDelete(t *testing.T) {
	var reply *string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "RCFG1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RsStopEngine(t *testing.T) {
	if err := engine.KillEngine(resDelay); err != nil {
		t.Error(err)
	}
}
