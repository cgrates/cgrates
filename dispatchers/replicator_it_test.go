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
	testDspRplAccountProfile,
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
func testDspRplAccountProfile(t *testing.T) {
	// Set RateProfile
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
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetAccountProfile, rPrf, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetAccountProfile: ", err)
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
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccountProfile, args, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetAccountProfile: ", err)
	} else if !reflect.DeepEqual(rPrf.Account, reply) {
		t.Errorf("Expecting: %+v, received: %+v, ", rPrf.Account, reply)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccountProfile, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveAccountProfile, args, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get RateProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccountProfile, args, &reply); err == nil ||
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
