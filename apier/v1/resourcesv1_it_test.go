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
)

func TestRLsV1LoadConfig(t *testing.T) {
	var err error
	rlsV1CfgPath = path.Join(*dataDir, "conf", "samples", "reslimiter")
	if rlsV1Cfg, err = config.NewCGRConfigFromFolder(rlsV1CfgPath); err != nil {
		t.Error(err)
	}
}

func TestRLsV1InitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rlsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestRLsV1StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rlsV1CfgPath, 1000); err != nil {
		t.Fatal(err)
	}
}

func TestRLsV1RpcConn(t *testing.T) {
	var err error
	rlsV1Rpc, err = jsonrpc.Dial("tcp", rlsV1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestRLsV1TPFromFolder(t *testing.T) {
	var reply string
	time.Sleep(time.Duration(2000) * time.Millisecond)
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := rlsV1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(1000) * time.Millisecond)
}

func TestRLsV1GetResourcesForEvent(t *testing.T) {
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

func TestRLsV1AllocateResource(t *testing.T) {
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

func TestRLsV1AllowUsage(t *testing.T) {
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

func TestRLsV1ReleaseResource(t *testing.T) {
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

var resConfig = &engine.ResourceCfg{
	ID: "RCFG1",
	Filters: []*engine.RequestFilter{
		&engine.RequestFilter{
			Type:      "type",
			FieldName: "Name",
			Values:    []string{"FilterValue1", "FilterValue2"},
		},
	},
	ActivationInterval: &utils.ActivationInterval{
		ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
	},
	UsageTTL:          time.Duration(10) * time.Microsecond,
	Limit:             10,
	AllocationMessage: "MessageAllocation",
	Blocker:           true,
	Stored:            true,
	Weight:            20,
	Thresholds:        []string{"Val1", "Val2"},
}

func TestRLsV1GetResourceConfigBeforeSet(t *testing.T) {
	var reply *string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceConfig", &AttrGetResCfg{ID: "RCFG1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestRLsV1SetResourceConfig(t *testing.T) {
	var result string
	if err := rlsV1Rpc.Call("ApierV1.SetResourceConfig", resConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func TestRLsV1GetResourceConfigAfterSet(t *testing.T) {
	var reply *engine.ResourceCfg
	if err := rlsV1Rpc.Call("ApierV1.GetResourceConfig", &AttrGetResCfg{ID: "RCFG1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, resConfig) {
		t.Errorf("Expecting: %+v, received: %+v", resConfig, reply)
	}
}

func TestRLsV1UpdateResourceConfig(t *testing.T) {
	var result string
	resConfig.Filters = []*engine.RequestFilter{
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
		&engine.RequestFilter{
			Type:      "*string_prefix",
			FieldName: "Destination",
			Values:    []string{"10", "20"},
		},
	}
	if err := rlsV1Rpc.Call("ApierV1.SetResourceConfig", resConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func TestRLsV1GetResourceConfigAfterUpdate(t *testing.T) {
	var reply *engine.ResourceCfg
	if err := rlsV1Rpc.Call("ApierV1.GetResourceConfig", &AttrGetResCfg{ID: "RCFG1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, resConfig) {
		t.Errorf("Expecting: %+v, received: %+v", resConfig, reply)
	}
}

func TestRLsV1RemResourceCOnfig(t *testing.T) {
	var resp string
	if err := rlsV1Rpc.Call("ApierV1.RemResourceConfig", &AttrGetResCfg{ID: resConfig.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func TestRLsV1GetResourceConfigAfterDelete(t *testing.T) {
	var reply *string
	if err := rlsV1Rpc.Call("ApierV1.GetResourceConfig", &AttrGetResCfg{ID: resConfig.ID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestRLsV1StopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
