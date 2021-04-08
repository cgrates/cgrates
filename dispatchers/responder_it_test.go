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
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var sTestsDspRsp = []func(t *testing.T){
	testDspResponderStatus,
	testDspResponderShutdown,

	testDspResponderRandom,
	testDspResponderBroadcast,
	testDspResponderInternal,
}

//Test start here
func TestDspResponder(t *testing.T) {
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
	testDsp(t, sTestsDspRsp, "TestDspResponder", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspResponderStatus(t *testing.T) {
	var reply map[string]interface{}
	if err := allEngine.RPC.Call(utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "ALL" {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "rsp12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "ALL" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "ALL2" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	allEngine.startEngine(t)
}

func getNodeWithRoute(route string, t *testing.T) string {
	var reply map[string]interface{}
	var pingReply string
	pingEv := utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.EventName: "Random",
		},

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey:  "rsp12345",
			utils.OptsRouteID: route,
		},
	}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey:  "rsp12345",
			utils.OptsRouteID: route,
		},
	}

	if err := dispEngine.RPC.Call(utils.CoreSv1Ping, pingEv, &pingReply); err != nil {
		t.Error(err)
	} else if pingReply != utils.Pong {
		t.Errorf("Received: %s", pingReply)
	}
	if err := dispEngine.RPC.Call(utils.CoreSv1Status, ev, &reply); err != nil {
		t.Error(err)
	}
	if reply[utils.NodeID] == nil {
		return ""
	}
	return reply[utils.NodeID].(string)
}

func testDspResponderRandom(t *testing.T) {
	node := getNodeWithRoute("r_init", t)
	for i := 0; i < 10; i++ {
		if node != getNodeWithRoute(fmt.Sprintf("R_%v", i), t) {
			return
		}
	}
	t.Errorf("Random strategy fail with 0.0009765625%% probability")
}

func testDspResponderShutdown(t *testing.T) {
	var reply string
	var statusReply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "rsp12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ResponderShutdown, ev, &reply); err != nil {
		t.Error(err)
	} else if reply != "Done!" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	if err := dispEngine.RPC.Call(utils.CoreSv1Status, &ev, &statusReply); err != nil {
		t.Error(err)
	} else if statusReply[utils.NodeID] != "ALL2" {
		t.Errorf("Received: %s", utils.ToJSON(statusReply))
	}
	if err := dispEngine.RPC.Call(utils.ResponderShutdown, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != "Done!" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspResponderBroadcast(t *testing.T) {
	var pingReply string
	pingEv := utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.EventName: "Broadcast",
		},

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "rsp12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ResponderPing, pingEv, &pingReply); err != nil {
		t.Error(err)
	} else if pingReply != utils.Pong {
		t.Errorf("Received: %s", pingReply)
	}

	allEngine2.stopEngine(t)
	pingReply = ""
	if err := dispEngine.RPC.Call(utils.ResponderPing, pingEv, &pingReply); err == nil ||
		err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("Expected error: %s received error: %v	 and reply %q", utils.ErrPartiallyExecuted.Error(), err, pingReply)
	}
	allEngine.stopEngine(t)
	time.Sleep(10 * time.Millisecond)
	pingReply = ""
	if err := dispEngine.RPC.Call(utils.ResponderPing, pingEv, &pingReply); err == nil ||
		!rpcclient.IsNetworkError(err) {
		t.Errorf("Expected error: %s received error: %v	 and reply %q", utils.ErrPartiallyExecuted.Error(), err, pingReply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspResponderInternal(t *testing.T) {
	var reply map[string]interface{}
	var pingReply string
	route := "internal"
	pingEv := utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.EventName: "Internal",
		},

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey:  "rsp12345",
			utils.OptsRouteID: route,
		},
	}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey:  "rsp12345",
			utils.OptsRouteID: route,
		},
	}
	if err := dispEngine.RPC.Call(utils.CoreSv1Ping, pingEv, &pingReply); err != nil {
		t.Error(err)
	} else if pingReply != utils.Pong {
		t.Errorf("Received: %s", pingReply)
	}
	if err := dispEngine.RPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	}
	if reply[utils.NodeID] == nil {
		return
	}
	if strRply := reply[utils.NodeID].(string); strRply != "DispatcherS1" {
		t.Errorf("Expected: DispatcherS1 , received: %s", strRply)
	}
}

func TestDspResponderPingEventNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var reply *string
	result := dspSrv.ResponderPing(nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderPingEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ResponderPing(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderPingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ResponderPing(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderDebitNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderDebit(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderDebitErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderDebit(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetCostNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderGetCost(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetCostErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderGetCost(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderMaxDebitNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderMaxDebit(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderMaxDebitErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderMaxDebit(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderRefundIncrementsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.Account
	result := dspSrv.ResponderRefundIncrements(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderRefundIncrementsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.Account
	result := dspSrv.ResponderRefundIncrements(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderRefundRoundingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *float64
	result := dspSrv.ResponderRefundRounding(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderRefundRoundingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *float64
	result := dspSrv.ResponderRefundRounding(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetMaxSessionTimeNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *time.Duration
	result := dspSrv.ResponderGetMaxSessionTime(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetMaxSessionTimeErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *time.Duration
	result := dspSrv.ResponderGetMaxSessionTime(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderShutdownNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ResponderShutdown(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderShutdownErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ResponderShutdown(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
