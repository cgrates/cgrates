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
	testDsp(t, sTestsDspSts, "TestDspStatS", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspStsPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.StatSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "stat12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.StatSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.StatSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.StatSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspStsGetStatFailover(t *testing.T) {
	var reply []string
	var metrics map[string]string
	expected := []string{"Stats1"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.EventName:    "Event1",
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        135 * time.Second,
				utils.Cost:         123.0,
				utils.RunID:        utils.MetaDefault,
				utils.Destination:  "1002",
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "stat12345",
			},
		},
	}
	if err := dispEngine.RPC.Call(utils.StatSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	args2 := utils.TenantIDWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "stat12345",
		},
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats1",
		},
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	}

	allEngine.startEngine(t)
	allEngine2.stopEngine(t)

	if err := dispEngine.RPC.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error NOT_FOUND but received %v and reply %v\n", err, reply)
	}
	allEngine2.startEngine(t)
}

func testDspStsPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.StatSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.StatSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "stat12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspStsTestAuthKey(t *testing.T) {
	var reply []string
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        135 * time.Second,
				utils.Cost:         123.0,
				utils.PDD:          12 * time.Second},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "12345",
			},
		},
	}
	if err := dispEngine.RPC.Call(utils.StatSv1ProcessEvent,
		args, &reply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}

	args2 := utils.TenantIDWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "12345",
		},
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		},
	}

	var metrics map[string]string
	if err := dispEngine.RPC.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspStsTestAuthKey2(t *testing.T) {
	var reply []string
	var metrics map[string]string
	expected := []string{"Stats2"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        135 * time.Second,
				utils.Cost:         123.0,
				utils.RunID:        utils.MetaDefault,
				utils.Destination:  "1002"},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "stat12345",
			},
		},
	}
	if err := dispEngine.RPC.Call(utils.StatSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	args2 := utils.TenantIDWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "stat12345",
		},
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		},
	}
	expectedMetrics := map[string]string{
		utils.MetaTCC: "123",
		utils.MetaTCD: "2m15s",
	}

	if err := dispEngine.RPC.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}

	args = engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        45 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         10.0,
				utils.Destination:  "1001",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "stat12345",
			},
		},
	}
	if err := dispEngine.RPC.Call(utils.StatSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expectedMetrics = map[string]string{
		utils.MetaTCC: "133",
		utils.MetaTCD: "3m0s",
	}
	if err := dispEngine.RPC.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testDspStsTestAuthKey3(t *testing.T) {
	var reply []string
	var metrics map[string]float64

	args2 := utils.TenantIDWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "stat12345",
		},
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		},
	}
	expectedMetrics := map[string]float64{
		utils.MetaTCC: 133,
		utils.MetaTCD: 180,
	}

	if err := dispEngine.RPC.Call(utils.StatSv1GetQueueFloatMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %v", expectedMetrics, metrics)
	}

	estats := []string{"Stats2", "Stats2_1"}
	if err := dispEngine.RPC.Call(utils.StatSv1GetQueueIDs,
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "stat12345",
			},
		}, &reply); err != nil {
		t.Error(err)
	}
	sort.Strings(estats)
	sort.Strings(reply)
	if !reflect.DeepEqual(estats, reply) {
		t.Errorf("expecting: %+v, received reply: %v", estats, reply)
	}

	estats = []string{"Stats2"}
	if err := dispEngine.RPC.Call(utils.StatSv1GetStatQueuesForEvent,
		&engine.StatsArgsProcessEvent{

			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "GetStats",
				Event: map[string]interface{}{
					utils.AccountField: "1002",
					utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:        45 * time.Second,
					utils.RunID:        utils.MetaDefault,
					utils.Cost:         10.0,
					utils.Destination:  "1001",
				},

				APIOpts: map[string]interface{}{
					utils.OptsAPIKey: "stat12345",
				},
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(estats, reply) {
		t.Errorf("expecting: %+v, received reply: %v", estats, reply)
	}
}

func TestDspStatSv1PingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.StatSv1Ping(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1PingNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var reply *string
	result := dspSrv.StatSv1Ping(nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1PingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.StatSv1Ping(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1GetStatQueuesForEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]string
	result := dspSrv.StatSv1GetStatQueuesForEvent(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1GetStatQueuesForEventErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]string
	result := dspSrv.StatSv1GetStatQueuesForEvent(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1GetQueueStringMetricsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *map[string]string
	result := dspSrv.StatSv1GetQueueStringMetrics(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1GetQueueStringMetricsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *map[string]string
	result := dspSrv.StatSv1GetQueueStringMetrics(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1ProcessEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]string
	result := dspSrv.StatSv1ProcessEvent(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1ProcessEventErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]string
	result := dspSrv.StatSv1ProcessEvent(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1GetQueueFloatMetricsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *map[string]float64
	result := dspSrv.StatSv1GetQueueFloatMetrics(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1GetQueueFloatMetricsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *map[string]float64
	result := dspSrv.StatSv1GetQueueFloatMetrics(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1GetQueueIDsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string
	result := dspSrv.StatSv1GetQueueIDs(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspStatSv1GetQueueIDsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string
	result := dspSrv.StatSv1GetQueueIDs(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
