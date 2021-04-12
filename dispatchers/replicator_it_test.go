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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspRpl = []func(t *testing.T){
	testDspRplPingFailover,
	testDspRplSupplierProfile,
	testDspRplAttributeProfile,
	testDspRplChargerProfile,
	testDspRplDispatcherProfile,
	testDspRplDispatcherHost,
	testDspRplFilter,
	testDspRplThreshold,
	testDspRplThresholdProfile,
	testDspRplStatQueue,
	testDspRplStatQueueProfile,
	testDspRplResource,
	testDspRplResourceProfile,
	testDspRplTiming,
	testDspRplDestination,
	testDspRplRateProfile,
	testDspRplAccount,
	testDspRplActionProfile,
}

//Test start here
func TestDspReplicator(t *testing.T) {
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
	testDsp(t, sTestsDspRpl, "TestDspReplicator", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspRplPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.ReplicatorSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = utils.EmptyString
	if err := allEngine2.RPC.Call(utils.ReplicatorSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = utils.EmptyString
	ev := utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	reply = utils.EmptyString
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	reply = utils.EmptyString
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
	reply = utils.EmptyString
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspRplSupplierProfile(t *testing.T) {
	// Set RouteProfile
	var replyStr string
	argSetSupplierProfile := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}

	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetRouteProfile, argSetSupplierProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetSupplierProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get RouteProfile
	var reply *engine.RouteProfile
	argRouteProfile := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRouteProfile, argRouteProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetSupplierProfile: ", err)
	} else if reply.ID != argSetSupplierProfile.ID {
		t.Errorf("Expecting: %+v, received: %+v", argSetSupplierProfile.ID, reply.ID)
	} else if reply.Tenant != argSetSupplierProfile.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argSetSupplierProfile.Tenant, reply.Tenant)
	}

	// Stop engine 1
	allEngine.stopEngine(t)

	// Get RouteProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRouteProfile, argRouteProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove SupplierProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveRouteProfile, argRouteProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get RouteProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRouteProfile, argRouteProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplAttributeProfile(t *testing.T) {
	// Set AttributeProfile
	var replyStr string
	setAttributeProfile := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID:     "id",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetAttributeProfile, setAttributeProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetAttributeProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get AttributeProfile
	var reply engine.AttributeProfile
	argAttributeProfile := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "id",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAttributeProfile, argAttributeProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetAttributeProfile: ", err)
	} else if reply.ID != setAttributeProfile.ID {
		t.Errorf("Expecting: %+v, received: %+v", setAttributeProfile.ID, reply.ID)
	} else if reply.Tenant != setAttributeProfile.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", setAttributeProfile.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get AttributeProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAttributeProfile, argAttributeProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove AttributeProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveAttributeProfile, argAttributeProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get AttributeProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAttributeProfile, argAttributeProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplChargerProfile(t *testing.T) {
	// Set ChargerProfile
	var replyStr string
	setChargerProfile := &engine.ChargerProfileWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:     "id",
			Tenant: "cgrates.org",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetChargerProfile, setChargerProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetChargerProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get ChargerProfile
	var reply engine.ChargerProfile
	argsChargerProfile := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "id",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetChargerProfile, argsChargerProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetChargerProfile: ", err)
	} else if reply.ID != argsChargerProfile.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsChargerProfile.ID, reply.ID)
	} else if reply.Tenant != argsChargerProfile.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsChargerProfile.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get ChargerProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetChargerProfile, argsChargerProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove ChargerProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveChargerProfile, argsChargerProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get ChargerProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetChargerProfile, argsChargerProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplDispatcherProfile(t *testing.T) {
	// Set DispatcherProfile
	var replyStr string
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get DispatcherProfile
	var reply engine.DispatcherProfile
	argsDispatcherProfile := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDispatcherProfile, argsDispatcherProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetDispatcherProfile: ", err)
	} else if reply.ID != argsDispatcherProfile.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsDispatcherProfile.ID, reply.ID)
	} else if reply.Tenant != argsDispatcherProfile.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsDispatcherProfile.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get DispatcherProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDispatcherProfile, argsDispatcherProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove DispatcherProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveDispatcherProfile, argsDispatcherProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get DispatcherProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDispatcherProfile, argsDispatcherProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplDispatcherHost(t *testing.T) {
	// Set DispatcherHost
	var replyStr string
	setDispatcherHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID: "ID",
			},
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetDispatcherHost: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get DispatcherHost
	var reply engine.DispatcherHost
	argsDispatcherHost := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDispatcherHost, argsDispatcherHost, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetDispatcherHost: ", err)
	} else if reply.ID != argsDispatcherHost.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsDispatcherHost.ID, reply.ID)
	} else if reply.Tenant != argsDispatcherHost.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsDispatcherHost.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get DispatcherHost
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDispatcherHost, argsDispatcherHost, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove DispatcherHost
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveDispatcherHost, argsDispatcherHost, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get DispatcherHost
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDispatcherHost, argsDispatcherHost, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplFilter(t *testing.T) {
	// Set Filter
	var replyStr string
	setFilter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetFilter, setFilter, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetFilter: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get Filter
	var reply engine.Filter
	argsFilter := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetFilter, argsFilter, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetFilter: ", err)
	} else if reply.ID != argsFilter.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsFilter.ID, reply.ID)
	} else if reply.Tenant != argsFilter.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsFilter.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get Filter
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetFilter, argsFilter, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove Filter
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveFilter, argsFilter, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get Filter
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetFilter, argsFilter, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplThreshold(t *testing.T) {
	// Set Threshold
	var replyStr string
	setThreshold := &engine.ThresholdWithAPIOpts{
		Threshold: &engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetThreshold, setThreshold, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetThreshold: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get Threshold
	var reply engine.Threshold
	argsThreshold := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetThreshold, argsThreshold, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetThreshold: ", err)
	} else if reply.ID != argsThreshold.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsThreshold.ID, reply.ID)
	} else if reply.Tenant != argsThreshold.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsThreshold.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get Threshold
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetThreshold, argsThreshold, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove Threshold
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveThreshold, argsThreshold, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get Threshold
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetThreshold, argsThreshold, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplThresholdProfile(t *testing.T) {
	// Set ThresholdProfile
	var replyStr string
	setThresholdProfile := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetThresholdProfile, setThresholdProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetThresholdProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get ThresholdProfile
	var reply engine.ThresholdProfile
	argsThresholdProfile := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetThresholdProfile, argsThresholdProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetThresholdProfile: ", err)
	} else if reply.ID != argsThresholdProfile.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsThresholdProfile.ID, reply.ID)
	} else if reply.Tenant != argsThresholdProfile.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsThresholdProfile.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get ThresholdProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetThresholdProfile, argsThresholdProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove ThresholdProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveThresholdProfile, argsThresholdProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get ThresholdProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetThresholdProfile, argsThresholdProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplStatQueue(t *testing.T) {
	// Set StatQueue
	var replyStr string
	setStatQueue := &engine.StatQueueWithAPIOpts{
		StatQueue: &engine.StatQueue{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetStatQueue, setStatQueue, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetStatQueue: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get StatQueue
	var reply engine.StatQueue
	argsStatQueue := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetStatQueue, argsStatQueue, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetStatQueue: ", err)
	} else if reply.ID != argsStatQueue.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsStatQueue.ID, reply.ID)
	} else if reply.Tenant != argsStatQueue.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsStatQueue.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get StatQueue
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetStatQueue, argsStatQueue, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove StatQueue
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveStatQueue, argsStatQueue, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get StatQueue
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetStatQueue, argsStatQueue, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplStatQueueProfile(t *testing.T) {
	// Set StatQueueProfile
	var replyStr string
	setStatQueueProfile := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetStatQueueProfile, setStatQueueProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetStatQueueProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get StatQueueProfile
	var reply engine.StatQueueProfile
	argsStatQueueProfile := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetStatQueueProfile, argsStatQueueProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetStatQueueProfile: ", err)
	} else if reply.ID != argsStatQueueProfile.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsStatQueueProfile.ID, reply.ID)
	} else if reply.Tenant != argsStatQueueProfile.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsStatQueueProfile.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get StatQueueProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetStatQueueProfile, argsStatQueueProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove StatQueueProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveStatQueueProfile, argsStatQueueProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get StatQueueProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetStatQueueProfile, argsStatQueueProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplResource(t *testing.T) {
	// Set Resource
	var replyStr string
	setResource := &engine.ResourceWithAPIOpts{
		Resource: &engine.Resource{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetResource, setResource, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetResource: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get Resource
	var reply engine.Resource
	argsResource := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetResource, argsResource, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetResource: ", err)
	} else if reply.ID != argsResource.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsResource.ID, reply.ID)
	} else if reply.Tenant != argsResource.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsResource.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get Resource
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetResource, argsResource, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove Resource
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveResource, argsResource, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get Resource
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetResource, argsResource, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplResourceProfile(t *testing.T) {
	// Set ResourceProfile
	var replyStr string
	setResourceProfile := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetResourceProfile, setResourceProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetResourceProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get ResourceProfile
	var reply engine.ResourceProfile
	argsResourceProfile := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetResourceProfile, argsResourceProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetResourceProfile: ", err)
	} else if reply.ID != argsResourceProfile.ID {
		t.Errorf("Expecting: %+v, received: %+v", argsResourceProfile.ID, reply.ID)
	} else if reply.Tenant != argsResourceProfile.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argsResourceProfile.Tenant, reply.Tenant)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get ResourceProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetResourceProfile, argsResourceProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove ResourceProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveResourceProfile, argsResourceProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get ResourceProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetResourceProfile, argsResourceProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplTiming(t *testing.T) {
	// Set Timing
	var replyStr string
	setTiming := &utils.TPTimingWithAPIOpts{
		TPTiming: &utils.TPTiming{
			ID:    "testTimings",
			Years: utils.Years{1999},
		},
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetTiming, setTiming, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetTiming: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get Timing
	var reply utils.TPTiming
	argsTiming := &utils.StringWithAPIOpts{
		Arg:    "testTimings",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetTiming, argsTiming, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetTiming: ", err)
	} else if reply.ID != argsTiming.Arg {
		t.Errorf("Expecting: %+v, received: %+v", argsTiming.Arg, reply.ID)
	} else if reply.Years[0] != 1999 {
		t.Errorf("Expecting: %+v, received: %+v", utils.Years{1999}, reply.Years)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get Timing
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetTiming, argsTiming, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove Timing
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveTiming, argsTiming, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get Timing
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetTiming, argsTiming, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

//Destination
func testDspRplDestination(t *testing.T) {
	// Set Destination
	var replyStr string
	setDestination := &engine.DestinationWithAPIOpts{
		Destination: &engine.Destination{
			ID: "idDestination"},
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetDestination, setDestination, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetDestination: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get Destination
	var reply engine.Destination
	argsDestination := &utils.StringWithAPIOpts{
		Arg:    "idDestination",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDestination, argsDestination, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetDestination: ", err)
	} else if reply.ID != setDestination.ID {
		t.Errorf("Expecting: %+v, received: %+v", setDestination.ID, reply.ID)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get Destination
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDestination, argsDestination, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove Destination
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveDestination, argsDestination, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get Destination
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetDestination, argsDestination, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplRateProfile(t *testing.T) {
	// Set RateProfile
	var replyStr string
	rPrf := &utils.RateProfileWithAPIOpts{
		RateProfile: &utils.RateProfile{
			Tenant:    "cgrates.org",
			ID:        "RP1",
			FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002"},
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"FIRST_GI": {
					ID:        "FIRST_GI",
					FilterIDs: []string{"*gi:~*req.Usage:0"},
					Weights: utils.DynamicWeights{
						{
							Weight: 0,
						},
					},
					Blocker: false,
				},
				"SECOND_GI": {
					ID:        "SECOND_GI",
					FilterIDs: []string{"*gi:~*req.Usage:1m"},
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Blocker: false,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetRateProfile, rPrf, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetRateProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get RateProfile
	var reply *utils.RateProfile
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RP1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRateProfile, args, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetRateProfile: ", err)
	} else if !reflect.DeepEqual(rPrf.RateProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v, ", rPrf.RateProfile, reply)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRateProfile, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveRateProfile, args, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRateProfile, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}
func testDspRplAccount(t *testing.T) {
	// Set Account
	var replyStr string
	rPrf := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "RP1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetAccount, rPrf, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetAccount: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get RateProfile
	var reply *utils.Account
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RP1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccount, args, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetAccount: ", err)
	} else if !reflect.DeepEqual(rPrf.Account, reply) {
		t.Errorf("Expecting: %+v, received: %+v, ", rPrf.Account, reply)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccount, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveAccount, args, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccount, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

}

func testDspRplActionProfile(t *testing.T) {
	// Set RateProfile
	var replyStr string
	rPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "RP1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetActionProfile, rPrf, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetActionProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get RateProfile
	var reply *engine.ActionProfile
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RP1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionProfile, args, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetActionProfile: ", err)
	} else if !reflect.DeepEqual(rPrf.ActionProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v, ", rPrf.ActionProfile, reply)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionProfile, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveActionProfile, args, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionProfile, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

}

func TestDspReplicatorSv1PingNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var reply *string
	result := dspSrv.ReplicatorSv1Ping(nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1PingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1Ping(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1PingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1Ping(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetDestinationNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}

	var reply *engine.Destination
	result := dspSrv.ReplicatorSv1GetDestination(nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetDestinationNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *engine.Destination
	result := dspSrv.ReplicatorSv1GetDestination(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetDestinationErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *engine.Destination
	result := dspSrv.ReplicatorSv1GetDestination(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetReverseDestinationNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var reply *[]string
	result := dspSrv.ReplicatorSv1GetReverseDestination(nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetReverseDestinationNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string
	result := dspSrv.ReplicatorSv1GetReverseDestination(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetReverseDestinationErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string
	result := dspSrv.ReplicatorSv1GetReverseDestination(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetStatQueueNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.StatQueue
	result := dspSrv.ReplicatorSv1GetStatQueue(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetStatQueueErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.StatQueue
	result := dspSrv.ReplicatorSv1GetStatQueue(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetFilterNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.Filter
	result := dspSrv.ReplicatorSv1GetFilter(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetFilterErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.Filter
	result := dspSrv.ReplicatorSv1GetFilter(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetThresholdNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.Threshold
	result := dspSrv.ReplicatorSv1GetThreshold(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetThresholdErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.Threshold
	result := dspSrv.ReplicatorSv1GetThreshold(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetThresholdProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.ThresholdProfile
	result := dspSrv.ReplicatorSv1GetThresholdProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetThresholdProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.ThresholdProfile
	result := dspSrv.ReplicatorSv1GetThresholdProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetStatQueueProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.StatQueueProfile
	result := dspSrv.ReplicatorSv1GetStatQueueProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetStatQueueProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.StatQueueProfile
	result := dspSrv.ReplicatorSv1GetStatQueueProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetTimingNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var reply *utils.TPTiming
	result := dspSrv.ReplicatorSv1GetTiming(nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetTimingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *utils.TPTiming
	result := dspSrv.ReplicatorSv1GetTiming(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetTimingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *utils.TPTiming
	result := dspSrv.ReplicatorSv1GetTiming(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetResourceNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.Resource
	result := dspSrv.ReplicatorSv1GetResource(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetResourceErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.Resource
	result := dspSrv.ReplicatorSv1GetResource(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetResourceProfileReplicatorSv1GetResourceProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.ResourceProfile
	result := dspSrv.ReplicatorSv1GetResourceProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetResourceProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.ResourceProfile
	result := dspSrv.ReplicatorSv1GetResourceProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetRouteProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.RouteProfile
	result := dspSrv.ReplicatorSv1GetRouteProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetRouteProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.RouteProfile
	result := dspSrv.ReplicatorSv1GetRouteProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetAttributeProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.AttributeProfile
	result := dspSrv.ReplicatorSv1GetAttributeProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetAttributeProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.AttributeProfile
	result := dspSrv.ReplicatorSv1GetAttributeProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetChargerProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.ChargerProfile
	result := dspSrv.ReplicatorSv1GetChargerProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetChargerProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.ChargerProfile
	result := dspSrv.ReplicatorSv1GetChargerProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetDispatcherProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.DispatcherProfile
	result := dspSrv.ReplicatorSv1GetDispatcherProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetDispatcherProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.DispatcherProfile
	result := dspSrv.ReplicatorSv1GetDispatcherProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetDispatcherHostNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.DispatcherHost
	result := dspSrv.ReplicatorSv1GetDispatcherHost(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetDispatcherHostErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *engine.DispatcherHost
	result := dspSrv.ReplicatorSv1GetDispatcherHost(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetItemLoadIDsNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var reply *map[string]int64
	result := dspSrv.ReplicatorSv1GetItemLoadIDs(nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetItemLoadIDsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *map[string]int64
	result := dspSrv.ReplicatorSv1GetItemLoadIDs(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetItemLoadIDsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *map[string]int64
	result := dspSrv.ReplicatorSv1GetItemLoadIDs(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetThresholdProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetThresholdProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetThresholdProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetThresholdProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveFilterNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveFilter(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveFilterErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveFilter(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveFilterNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveFilter(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveThresholdProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveThresholdProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveThresholdProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveThresholdProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveThresholdProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveThresholdProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveStatQueueProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveStatQueueProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveStatQueueProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveStatQueueProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveStatQueueProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveStatQueueProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveResourceNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveResource(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveResourceErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveResource(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveResourceNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveResource(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveResourceProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveResourceProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveResourceProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveResourceProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveResourceProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveResourceProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveTimingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveTiming(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveTimingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveTiming(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveTimingNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveTiming(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveActionsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveActions(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveActionsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveActions(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveActionsEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveActions(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveRouteProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveRouteProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveRouteProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveRouteProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveRouteProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveRouteProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveAttributeProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveAttributeProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveAttributeProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveAttributeProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveAttributeProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveAttributeProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveChargerProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveChargerProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveChargerProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveChargerProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveChargerProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveChargerProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveDispatcherHostNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDispatcherHost(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveDispatcherHostErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDispatcherHost(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveDispatcherHostNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDispatcherHost(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveDispatcherProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDispatcherProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveDispatcherProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDispatcherProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveDispatcherProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDispatcherProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetIndexesNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.GetIndexesArg{
		Tenant: "tenant",
	}
	var reply *map[string]utils.StringSet
	result := dspSrv.ReplicatorSv1GetIndexes(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetIndexesErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.GetIndexesArg{
		Tenant: "tenant",
	}
	var reply *map[string]utils.StringSet
	result := dspSrv.ReplicatorSv1GetIndexes(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1GetIndexesNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *map[string]utils.StringSet
	result := dspSrv.ReplicatorSv1GetIndexes(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetIndexesNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.SetIndexesArg{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetIndexes(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetIndexesErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.SetIndexesArg{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetIndexes(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetIndexesNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetIndexes(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveIndexesNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.GetIndexesArg{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveIndexes(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveDestinationNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDestination(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveDestinationNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDestination(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestReplicatorSv1RemoveDestinationErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.StringWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveDestination(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetLoadIDsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.LoadIDsWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetLoadIDs(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetLoadIDsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.LoadIDsWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetLoadIDs(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetLoadIDsNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetLoadIDs(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveAccountNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveAccount(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveAccountErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveAccount(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveAccountNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveAccount(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveStatQueueNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveStatQueue(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveStatQueueErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveStatQueue(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveStatQueueNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveStatQueue(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveIndexesErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.GetIndexesArg{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveIndexes(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveIndexesNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveIndexes(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetThresholdProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetThresholdProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetThresholdNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ThresholdWithAPIOpts{
		Threshold: &engine.Threshold{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetThreshold(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetThresholdErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ThresholdWithAPIOpts{
		Threshold: &engine.Threshold{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetThreshold(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetThresholdNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetThreshold(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetDestinationNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.DestinationWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetDestination(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetDestinationErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.DestinationWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetDestination(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetDestinationNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetDestination(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetAccountNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.AccountWithAPIOpts{
		Account: &utils.Account{},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetAccount(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetAccountErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			ID: "testID",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetAccount(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetAccountNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetAccount(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetReverseDestinationNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.DestinationWithAPIOpts{
		Destination: &engine.Destination{
			ID: "testID",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetReverseDestination(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetReverseDestinationErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.DestinationWithAPIOpts{
		Destination: &engine.Destination{
			ID: "testID",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetReverseDestination(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetReverseDestinationNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetReverseDestination(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetStatQueueNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.StatQueueWithAPIOpts{
		StatQueue: &engine.StatQueue{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetStatQueue(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetStatQueueErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.StatQueueWithAPIOpts{
		StatQueue: &engine.StatQueue{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetStatQueue(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetStatQueueNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetStatQueue(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetFilterNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetFilter(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetFilterErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetFilter(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetFilterNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetFilter(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetStatQueueProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetStatQueueProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetStatQueueProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetStatQueueProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetStatQueueProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetStatQueueProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetTimingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TPTimingWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetTiming(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetTimingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TPTimingWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetTiming(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetTimingNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetTiming(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetResourceNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ResourceWithAPIOpts{
		Resource: &engine.Resource{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetResource(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetResourceErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ResourceWithAPIOpts{
		Resource: &engine.Resource{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetResource(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetResourceNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetResource(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetResourceProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetResourceProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestReplicatorSv1SetResourceProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetResourceProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetResourceProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetResourceProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetRouteProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetRouteProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetRouteProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetRouteProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetRouteProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetRouteProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetAttributeProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetAttributeProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetAttributeProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetAttributeProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetAttributeProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetAttributeProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetChargerProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ChargerProfileWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetChargerProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetChargerProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ChargerProfileWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetChargerProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetChargerProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetChargerProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetDispatcherProfileNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetDispatcherProfile(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetDispatcherProfileErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetDispatcherProfile(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetDispatcherProfileNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetDispatcherProfile(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetDispatcherHostNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetDispatcherHost(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestReplicatorSv1SetDispatcherHostErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1SetDispatcherHost(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1SetDispatcherHostNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1SetDispatcherHost(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveThresholdNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveThreshold(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestReplicatorSv1RemoveThresholdErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveThreshold(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspReplicatorSv1RemoveThresholdNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.ReplicatorSv1RemoveThreshold(nil, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
