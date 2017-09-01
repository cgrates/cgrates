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
	rlsConfig    *engine.ResourceCfg
	resDelay     int
)

var sTestsRLSV1 = []func(t *testing.T){
	testV1RLSLoadConfig,
	testV1RLSInitDataDb,
	testV1RLSResetStorDb,
	testV1RLSStartEngine,
	testV1RLSRpcConn,
	testV1RLSFromFolder,
	testV1RLSGetResourcesFromEvent,
	testV1RLSAllocateResource,
	testV1RLSAllowUsage,
	testV1RLSReleaseResource,
	testV1RLSGetResourceConfigBeforeSet,
	testV1RLSSetResourceConfig,
	testV1RLSGetResourceConfigAfterSet,
	testV1RLSUpdateResourceConfig,
	testV1RLSGetResourceConfigAfterUpdate,
	testV1RLSRemResourceCOnfig,
	testV1RLSGetResourceConfigAfterDelete,
	testV1RLSStopEngine,
}

//Test start here
func TestRLSV1ITMySQL(t *testing.T) {
	rlsV1ConfDIR = "tutmysql"
	for _, stest := range sTestsRLSV1 {
		t.Run(rlsV1ConfDIR, stest)
	}
}

func TestRLSV1ITMongo(t *testing.T) {
	rlsV1ConfDIR = "tutmongo"
	for _, stest := range sTestsRLSV1 {
		t.Run(rlsV1ConfDIR, stest)
	}
}

func testV1RLSLoadConfig(t *testing.T) {
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

func testV1RLSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1RLSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RLSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rlsV1CfgPath, resDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1RLSRpcConn(t *testing.T) {
	var err error
	rlsV1Rpc, err = jsonrpc.Dial("tcp", rlsV1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1RLSFromFolder(t *testing.T) {
	var reply string
	time.Sleep(time.Duration(2000) * time.Millisecond)
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := rlsV1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(1000) * time.Millisecond)
}

func testV1RLSGetResourcesFromEvent(t *testing.T) {
	var reply *[]*engine.ResourceCfg
	ev := map[string]interface{}{"Unknown": "unknown"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", ev, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	ev = map[string]interface{}{"Destination": "10"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", ev, &reply); err != nil {
		t.Error(err)
	}
	if len(*reply) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(*reply))
	}
	if (*reply)[0].ID != "ResGroup2" {
		t.Errorf("Expecting: %+v, received: %+v", "ResGroup2", (*reply)[0].ID)
	}

	ev = map[string]interface{}{"Destination": "20"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", ev, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	ev = map[string]interface{}{"Account": "1002", "Subject": "test", "Destination": "1002"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", ev, &reply); err != nil {
		t.Error(err)
	}
	if len(*reply) != 2 {
		t.Errorf("Expecting: %+v, received: %+v", 2, len(*reply))
	}

	ev = map[string]interface{}{"Account": "1002", "Subject": "test", "Destination": "1001"}
	if err := rlsV1Rpc.Call("ResourceSV1.GetResourcesForEvent", ev, &reply); err != nil {
		t.Error(err)
	}
	if len(*reply) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(*reply))
	}
	if (*reply)[0].ID != "ResGroup2" {
		t.Errorf("Expecting: %+v, received: %+v", "ResGroup2", (*reply)[0].ID)
	}
}

func testV1RLSAllocateResource(t *testing.T) {
	var reply string

	attrRU := utils.AttrRLsResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
		Event:   map[string]interface{}{"Account": "1002", "Subject": "1001", "Destination": "1002"},
		Units:   3,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", attrRU, &reply); err != nil {
		t.Error(err)
	}
	if reply != "ResGroup1" {
		t.Errorf("Expecting: %+v, received: %+v", "ResGroup1", reply)
	}

	time.Sleep(time.Duration(1000) * time.Millisecond)

	attrRU = utils.AttrRLsResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e52",
		Event:   map[string]interface{}{"Destination": "100"},
		Units:   5,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", attrRU, &reply); err != nil {
		t.Error(err)
	}
	if reply != "ResGroup2" {
		t.Errorf("Expecting: %+v, received: %+v", "ResGroup2", reply)
	}

	time.Sleep(time.Duration(1000) * time.Millisecond)

	attrRU = utils.AttrRLsResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e53",
		Event:   map[string]interface{}{"Account": "1002", "Subject": "1001", "Destination": "1002"},
		Units:   3,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllocateResource", attrRU, &reply); err != nil {
		t.Error(err)
	}
	if reply != "ResGroup1" {
		t.Errorf("Expecting: %+v, received: %+v", "ResGroup1", reply)
	}

}

func testV1RLSAllowUsage(t *testing.T) {
	var reply bool
	attrRU := utils.AttrRLsResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
		Event:   map[string]interface{}{"Account": "1002", "Subject": "1001", "Destination": "1002"},
		Units:   1,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllowUsage", attrRU, &reply); err != nil {
		t.Error(err)
	} else if reply != true {
		t.Errorf("Expecting: %+v, received: %+v", true, reply)
	}

	attrRU = utils.AttrRLsResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
		Event:   map[string]interface{}{"Account": "1002", "Subject": "1001", "Destination": "1002"},
		Units:   2,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllowUsage", attrRU, &reply); err != nil {
		t.Error(err)
	}
}

func testV1RLSReleaseResource(t *testing.T) {
	var reply interface{}

	attrRU := utils.AttrRLsResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e52",
		Event:   map[string]interface{}{"Destination": "100"},
		Units:   2,
	}
	if err := rlsV1Rpc.Call("ResourceSV1.ReleaseResource", attrRU, &reply); err != nil {
		t.Error(err)
	}
	if err := rlsV1Rpc.Call("ResourceSV1.AllowUsage", attrRU, &reply); err != nil {
		t.Error(err)
	} else {
		if reply != true {
			t.Errorf("Expecting: %+v, received: %+v", true, reply)
		}
	}

	attrRU.Units += 7
	if err := rlsV1Rpc.Call("ResourceSV1.AllowUsage", attrRU, &reply); err == nil {
		t.Errorf("Expecting: %+v, received: %+v", false, reply)
	}

}

func testV1RLSGetResourceConfigBeforeSet(t *testing.T) {
	var reply *string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceConfig", &AttrGetResCfg{ID: "RCFG1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RLSSetResourceConfig(t *testing.T) {
	rlsConfig = &engine.ResourceCfg{
		ID: "RCFG1",
		Filters: []*engine.RequestFilter{
			&engine.RequestFilter{
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
	if err := rlsV1Rpc.Call("ApierV1.SetResourceConfig", rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1RLSGetResourceConfigAfterSet(t *testing.T) {
	var reply *engine.ResourceCfg
	if err := rlsV1Rpc.Call("ApierV1.GetResourceConfig", &AttrGetResCfg{ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", rlsConfig, reply)
	}
}

func testV1RLSUpdateResourceConfig(t *testing.T) {
	var result string
	rlsConfig.Filters = []*engine.RequestFilter{
		&engine.RequestFilter{
			Type:      "type",
			FieldName: "Name",
			Values:    []string{"FilterValue1", "FilterValue2"},
		},
		&engine.RequestFilter{
			Type:      "*string",
			FieldName: "Accout",
			Values:    []string{"1001", "1002"},
		},
	}
	if err := rlsV1Rpc.Call("ApierV1.SetResourceConfig", rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1RLSGetResourceConfigAfterUpdate(t *testing.T) {
	var reply *engine.ResourceCfg
	if err := rlsV1Rpc.Call("ApierV1.GetResourceConfig", &AttrGetResCfg{ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", rlsConfig, reply)
	}
}

func testV1RLSRemResourceCOnfig(t *testing.T) {
	var resp string
	if err := rlsV1Rpc.Call("ApierV1.RemResourceConfig", &AttrGetResCfg{ID: rlsConfig.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1RLSGetResourceConfigAfterDelete(t *testing.T) {
	var reply *string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceConfig", &AttrGetResCfg{ID: "RCFG1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RLSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
