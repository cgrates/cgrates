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
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspRpl = []func(t *testing.T){
	testDspRplPingFailover,
	testDspRplAccount,
	testDspRplSupplierProfile,
	testDspRplAttributeProfile,
	testDspRplChargerProfile,
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
	ev := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345"),
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
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
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
	attrSetAccount := &engine.AccountWithArgDispatcher{
		Account: &engine.Account{
			ID:            "cgrates.org:1008",
			AllowNegative: true,
			Disabled:      true,
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345"),
		},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetAccount, attrSetAccount, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetAccount: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get
	var reply *engine.Account
	argsGetAccount := &utils.StringWithApiKey{
		Arg: "cgrates.org:1008",
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345"),
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
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveAccount, &utils.StringWithApiKey{
		Arg: "cgrates.org:1008",
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345")},
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
	// Set SupplierProfile
	var replyStr string
	argSetSupplierProfile := &engine.SupplierProfileWithArgDispatcher{
		SupplierProfile: &engine.SupplierProfile{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345")},
	}

	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetSupplierProfile, argSetSupplierProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetSupplierProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get SupplierProfile
	var reply *engine.SupplierProfile
	argSupplierProfile := &utils.TenantIDWithArgDispatcher{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ID",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345")},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetSupplierProfile, argSupplierProfile, &reply); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.GetSupplierProfile: ", err)
	} else if reply.ID != argSetSupplierProfile.ID {
		t.Errorf("Expecting: %+v, received: %+v", argSetSupplierProfile.ID, reply.ID)
	} else if reply.Tenant != argSetSupplierProfile.Tenant {
		t.Errorf("Expecting: %+v, received: %+v", argSetSupplierProfile.Tenant, reply.Tenant)
	}

	// Stop engine 1
	allEngine.stopEngine(t)

	// Get SupplierProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetSupplierProfile, argSupplierProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}

	// Start engine 1
	allEngine.startEngine(t)

	// Remove SupplierProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1RemoveSupplierProfile, argSupplierProfile, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get SupplierProfile
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1GetSupplierProfile, argSupplierProfile, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v, received: %+v, ", utils.ErrNotFound, err)
	}
}

func testDspRplAttributeProfile(t *testing.T) {
	// Set AttributeProfile
	var replyStr string
	setAttributeProfile := &engine.AttributeProfileWithArgDispatcher{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID:     "id",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345")},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetAttributeProfile, setAttributeProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetAttributeProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Get AttributeProfile
	var reply engine.AttributeProfile
	argAttributeProfile := &utils.TenantIDWithArgDispatcher{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "id",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345")},
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
	setChargerProfile := &engine.ChargerProfileWithArgDispatcher{
		ChargerProfile: &engine.ChargerProfile{
			ID:     "id",
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345")},
	}
	if err := dispEngine.RPC.Call(utils.ReplicatorSv1SetChargerProfile, setChargerProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling ReplicatorSv1.SetChargerProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	// Get ChargerProfile
	var reply engine.ChargerProfile
	argsChargerProfile := &utils.TenantIDWithArgDispatcher{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "id",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("repl12345")},
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
