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
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"

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
	var config1, config2, config3 string
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspTh, "TestDspThresholdS", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspThPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.ThresholdSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "thr12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ThresholdSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.ThresholdSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.ThresholdSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspThProcessEventFailover(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1001"}
	nowTime := time.Now()
	args := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.EventName:    "Event1",
				utils.AccountField: "1001"},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "thr12345",
			},
		},
	}

	if err := dispEngine.RPC.Call(utils.ThresholdSv1ProcessEvent, args,
		&ids); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error NOT_FOUND but received %v and reply %v\n", err, ids)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
	allEngine2.startEngine(t)
}

func testDspThPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.ThresholdSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.ThresholdSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "thr12345",
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
	args := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.AccountField: "1002"},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "12345",
			},
		},
	}

	if err := dispEngine.RPC.Call(utils.ThresholdSv1ProcessEvent,
		args, &ids); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
	var th *engine.Thresholds
	if err := dispEngine.RPC.Call(utils.ThresholdSv1GetThresholdsForEvent, args,
		&th); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspThTestAuthKey2(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1002"}
	nowTime := time.Now()
	args := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.AccountField: "1002"},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "thr12345",
			},
		},
	}

	if err := dispEngine.RPC.Call(utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
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
	if err := dispEngine.RPC.Call(utils.ThresholdSv1GetThresholdsForEvent, args, &th); err != nil {
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
	if err := dispEngine.RPC.Call(utils.ThresholdSv1GetThreshold, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "thr12345",
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

	if err := dispEngine.RPC.Call(utils.ThresholdSv1GetThresholdIDs, &utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "thr12345",
		},
	}, &ids); err != nil {
		t.Fatal(err)
	}
	sort.Strings(ids)
	if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
}

func TestDspThresholdSv1PingNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ThresholdSv1Ping(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1PingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ThresholdSv1Ping(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1PingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{}
	var reply *string
	result := dspSrv.ThresholdSv1Ping(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1GetThresholdsForEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *engine.Thresholds
	result := dspSrv.ThresholdSv1GetThresholdsForEvent(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1GetThresholdsEvnil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *engine.Thresholds
	result := dspSrv.ThresholdSv1GetThresholdsForEvent(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1ProcessEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]string
	result := dspSrv.ThresholdSv1ProcessEvent(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1ProcessEventnNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]string
	result := dspSrv.ThresholdSv1ProcessEvent(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1GetThresholdIDsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string
	result := dspSrv.ThresholdSv1GetThresholdIDs(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1GetThresholdIDErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string
	result := dspSrv.ThresholdSv1GetThresholdIDs(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1GetThresholdNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.Threshold
	result := dspSrv.ThresholdSv1GetThreshold(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspThresholdSv1GetThresholdErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.Threshold
	result := dspSrv.ThresholdSv1GetThreshold(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
