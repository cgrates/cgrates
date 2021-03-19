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
	testDspRplAccount,
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
	testDspRplActionTriggers,
	testDspRplSharedGroup,
	testDspRplActions,
	testDspRplActionPlan,
	// testDspRplAccountActionPlans,
	testDspRplRatingPlan,
	testDspRplRatingProfile,
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

		Opts: map[string]interface{}{
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

func testDspRplAccount(t *testing.T) {
	// Set
	var replyStr string
	attrSetAccount := &engine.AccountWithAPIOpts{
		Account: &engine.Account{
			ID:            "cgrates.org:1008",
			AllowNegative: true,
			Disabled:      true,
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetAccount, attrSetAccount, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetAccount: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get
	var reply *engine.Account
	argsGetAccount := &utils.StringWithAPIOpts{
		Arg: "cgrates.org:1008",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccount, argsGetAccount, &reply); err != nil {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if reply.ID != attrSetAccount.Account.ID {
		t.Errorf("Expecting: %+v, received: %+v", attrSetAccount.Account.ID, reply.ID)
	} else if reply.AllowNegative != true {
		t.Errorf("Expecting: true, received: %+v", reply.AllowNegative)
	} else if reply.Disabled != true {
		t.Errorf("Expecting: true, received: %+v", reply.Disabled)
	}
	// Stop engine 1
	allEngine.stopEngine(t)
	// Get
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccount, argsGetAccount, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
	// Start engine 1
	allEngine.startEngine(t)
	// Remove Account
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveAccount, &utils.StringWithAPIOpts{
		Arg: "cgrates.org:1008",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get Account
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccount, argsGetAccount, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplSupplierProfile(t *testing.T) {
	// Set RouteProfile
	var replyStr string
	argSetSupplierProfile := &engine.RouteProfileWithOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		Opts: map[string]interface{}{
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
	setAttributeProfile := &engine.AttributeProfileWithOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID:     "id",
		},
		Opts: map[string]interface{}{
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
	setChargerProfile := &engine.ChargerProfileWithOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:     "id",
			Tenant: "cgrates.org",
		},
		Opts: map[string]interface{}{
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
	setDispatcherProfile := &engine.DispatcherProfileWithOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		Opts: map[string]interface{}{
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
	setDispatcherHost := &engine.DispatcherHostWithOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID: "ID",
			},
		},
		Opts: map[string]interface{}{
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
	setStatQueueProfile := &engine.StatQueueProfileWithOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		Opts: map[string]interface{}{
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
	setResourceProfile := &engine.ResourceProfileWithOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		Opts: map[string]interface{}{
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

func testDspRplActionTriggers(t *testing.T) {
	// Set ActionTriggers
	var replyStr string
	setActionTriggers := &engine.SetActionTriggersArgWithOpts{
		Key: "testActionTriggers",
		Attrs: engine.ActionTriggers{
			&engine.ActionTrigger{ID: "testActionTriggers"}},
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetActionTriggers, setActionTriggers, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetActionTriggers: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get ActionTriggers
	var reply engine.ActionTriggers
	argsActionTriggers := &utils.StringWithAPIOpts{
		Arg:    "testActionTriggers",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionTriggers, argsActionTriggers, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetActionTriggers: ", err)
	} else if reply[0].ID != argsActionTriggers.Arg {
		t.Errorf("Expecting: %+v, received: %+v", argsActionTriggers.Arg, reply[0].ID)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get ActionTriggers
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionTriggers, argsActionTriggers, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove ActionTriggers
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveActionTriggers, argsActionTriggers, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get ActionTriggers
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionTriggers, argsActionTriggers, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplSharedGroup(t *testing.T) {
	// Set SharedGroup
	var replyStr string
	setSharedGroup := &engine.SharedGroupWithOpts{
		SharedGroup: &engine.SharedGroup{
			Id: "IDSharedGroup",
		},
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetSharedGroup, setSharedGroup, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetSharedGroup: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get SharedGroup
	var reply engine.SharedGroup
	argsSharedGroup := &utils.StringWithAPIOpts{
		Arg:    "IDSharedGroup",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetSharedGroup, argsSharedGroup, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetSharedGroup: ", err)
	} else if reply.Id != setSharedGroup.Id {
		t.Errorf("Expecting: %+v, received: %+v", setSharedGroup.Id, reply.Id)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get SharedGroup
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetSharedGroup, argsSharedGroup, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove SharedGroup
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveSharedGroup, argsSharedGroup, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get SharedGroup
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetSharedGroup, argsSharedGroup, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplActions(t *testing.T) {
	// Set Actions
	var replyStr string
	setActions := &engine.SetActionsArgsWithOpts{
		Acs: engine.Actions{
			&engine.Action{
				Id:         "Action1",
				ActionType: utils.MetaLog,
			},
		},
		Key:    "KeyActions",
		Tenant: "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetActions, setActions, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetActions: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get Actions
	var reply engine.Actions
	argsActions := &utils.StringWithAPIOpts{
		Arg:    "KeyActions",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActions, argsActions, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetActions: ", err)
	} else if reply[0].Id != setActions.Acs[0].Id {
		t.Errorf("Expecting: %+v, received: %+v", setActions.Acs[0].Id, reply[0].Id)
	} else if len(reply) != len(setActions.Acs) {
		t.Errorf("Expecting: %+v, received: %+v", len(setActions.Acs), len(reply))
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get Actions
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActions, argsActions, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove Actions
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveActions, argsActions, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get Actions
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActions, argsActions, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplActionPlan(t *testing.T) {
	// Set ActionPlan
	var replyStr string
	setActionPlan := &engine.SetActionPlanArgWithOpts{
		Ats: &engine.ActionPlan{
			Id: "idtas",
			AccountIDs: utils.StringMap{
				"AccountTest": true,
			},
			ActionTimings: []*engine.ActionTiming{
				{
					ActionsID: "ActionsID",
				},
			},
		},
		Key:    "KeyActionPlan",
		Tenant: "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetActionPlan, setActionPlan, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetActionPlan: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get ActionPlan
	var reply engine.ActionPlan
	argsActionPlan := &utils.StringWithAPIOpts{
		Arg:    "KeyActionPlan",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionPlan, argsActionPlan, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetActionPlan: ", err)
	} else if reply.Id != setActionPlan.Ats.Id {
		t.Errorf("Expecting: %+v, received: %+v", setActionPlan.Ats.Id, reply.Id)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get ActionPlan
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionPlan, argsActionPlan, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove ActionPlan
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveActionPlan, argsActionPlan, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get ActionPlan
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetActionPlan, argsActionPlan, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplAccountActionPlans(t *testing.T) {
	// Set AccountActionPlans
	var replyStr string
	setAccountActionPlans := &engine.SetAccountActionPlansArgWithOpts{
		AplIDs:    []string{"KeyAccountActionPlans"},
		Overwrite: true,
		AcntID:    "KeyAccountActionPlans",
		Tenant:    "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetAccountActionPlans, setAccountActionPlans, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetAccountActionPlans: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get AccountActionPlans
	var reply []string
	argsAccountActionPlans := &utils.StringWithAPIOpts{
		Arg:    "KeyAccountActionPlans",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccountActionPlans, argsAccountActionPlans, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetAccountActionPlans: ", err)
	} else if reply[0] != setAccountActionPlans.AcntID {
		t.Errorf("Expecting: %+v, received: %+v", setAccountActionPlans.AcntID, reply[0])
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get AccountActionPlans
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccountActionPlans, argsAccountActionPlans, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove AccountActionPlans
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemAccountActionPlans, argsAccountActionPlans, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get AccountActionPlans
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetAccountActionPlans, argsAccountActionPlans, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplRatingPlan(t *testing.T) {
	// Set RatingPlan
	var replyStr string
	setRatingPlan := &engine.RatingPlanWithOpts{
		RatingPlan: &engine.RatingPlan{
			Id: "id",
			DestinationRates: map[string]engine.RPRateList{
				"DestinationRates": {&engine.RPRate{Rating: "Rating"}}},
			Ratings: map[string]*engine.RIRate{"Ratings": {ConnectFee: 0.7}},
			Timings: map[string]*engine.RITiming{"Timings": {Months: utils.Months{4}}},
		},
		Tenant: "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetRatingPlan, setRatingPlan, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetRatingPlan: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get RatingPlan
	var reply engine.RatingPlan
	argsRatingPlan := &utils.StringWithAPIOpts{
		Arg:    "id",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRatingPlan, argsRatingPlan, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetRatingPlan: ", err)
	} else if reply.Id != setRatingPlan.Id {
		t.Errorf("Expecting: %+v, received: %+v", setRatingPlan.Id, reply.Id)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get RatingPlan
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRatingPlan, argsRatingPlan, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove RatingPlan
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveRatingPlan, argsRatingPlan, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get RatingPlan
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRatingPlan, argsRatingPlan, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplRatingProfile(t *testing.T) {
	// Set RatingProfile
	var replyStr string
	setRatingProfile := &engine.RatingProfileWithOpts{
		RatingProfile: &engine.RatingProfile{
			Id: "idRatingProfile",
			RatingPlanActivations: engine.RatingPlanActivations{
				&engine.RatingPlanActivation{RatingPlanId: "RatingPlanId"}},
		},
		Tenant: "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetRatingProfile, setRatingProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetRatingProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get RatingProfile
	var reply engine.RatingProfile
	argsRatingProfile := &utils.StringWithAPIOpts{
		Arg:    "idRatingProfile",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRatingProfile, argsRatingProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetRatingProfile: ", err)
	} else if reply.Id != setRatingProfile.Id {
		t.Errorf("Expecting: %+v, received: %+v", setRatingProfile.Id, reply.Id)
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get RatingProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRatingProfile, argsRatingProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove RatingProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveRatingProfile, argsRatingProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get RatingProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetRatingProfile, argsRatingProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

//Destination
func testDspRplDestination(t *testing.T) {
	// Set Destination
	var replyStr string
	setDestination := &engine.DestinationWithAPIOpts{
		Destination: &engine.Destination{
			Id: "idDestination"},
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
	} else if reply.Id != setDestination.Id {
		t.Errorf("Expecting: %+v, received: %+v", setDestination.Id, reply.Id)
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

func testDspRplLoadIDs(t *testing.T) {
	// Set LoadIDs
	var replyStr string
	setLoadIDs := &utils.LoadIDsWithOpts{
		LoadIDs: map[string]int64{
			"LoadID1": 1,
			"LoadID2": 2},
		Tenant: "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetLoadIDs, setLoadIDs, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetLoadIDs: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get LoadIDs
	var reply map[string]int64
	argsLoadIDs := &utils.StringWithAPIOpts{
		Arg:    "idLoadIDs",
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "repl12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetItemLoadIDs, argsLoadIDs, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetItemLoadIDs: ", err)
	} else if reflect.DeepEqual(reply, setLoadIDs) {
		t.Errorf("Expecting: %+v, received: %+v", setLoadIDs, reply)
	} else if len(reply) != len(setLoadIDs.LoadIDs) {
		t.Errorf("Expecting: %+v, received: %+v, ", len(setLoadIDs.LoadIDs), len(reply))
	} else if reply["LoadID1"] != setLoadIDs.LoadIDs["LoadID1"] {
		t.Errorf("Expecting: %+v, received: %+v, ", setLoadIDs.LoadIDs["LoadID1"], reply["LoadID1"])
	}
	// Stop engine 1
	allEngine.stopEngine(t)

	// Get LoadIDs
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetItemLoadIDs, argsLoadIDs, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)
}

func testDspRplRateProfile(t *testing.T) {
	// Set RateProfile
	var replyStr string
	rPrf := &utils.RateProfileWithOpts{
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
		Opts: map[string]interface{}{
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
	rPrf := &utils.AccountProfileWithAPIOpts{
		AccountProfile: &utils.AccountProfile{
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
	var reply *utils.AccountProfile
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
	} else if !reflect.DeepEqual(rPrf.AccountProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v, ", rPrf.AccountProfile, reply)
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
