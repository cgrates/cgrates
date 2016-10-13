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
package v2

import (
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var testIT = flag.Bool("integration", false, "Perform the tests only on local test environment, not by default.")

var apierCfgPath string
var apierCfg *config.CGRConfig
var apierRPC *rpc.Client

func TestApierV2itLoadConfig(t *testing.T) {
	if !*testIT {
		return
	}
	apierCfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	if apierCfg, err = config.NewCGRConfigFromFolder(tpCfgPath); err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func TestApierV2itResetDataDb(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.InitDataDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestApierV2itResetStorDb(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.InitStorDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestApierV2itStartEngine(t *testing.T) {
	if !*testIT {
		return
	}
	if _, err := engine.StopStartEngine(apierCfgPath, 200); err != nil { // Mongo requires more time to start
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestApierV2itRpcConn(t *testing.T) {
	if !*testIT {
		return
	}
	apierRPC, err = jsonrpc.Dial("tcp", apierCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func TestApierV2itAddBalance(t *testing.T) {
	if !*testIT {
		return
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "dan",
		BalanceType: utils.MONETARY,
		BalanceID:   utils.StringPointer(utils.META_DEFAULT),
		Value:       utils.Float64Pointer(5.0),
		Weight:      utils.Float64Pointer(10.0),
	}
	var reply string
	if err := apierRPC.Call("ApierV2.SetBalance", attrs, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt engine.Account
	if err := apierRPC.Call("ApierV2.GetAccount", &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY][0].Value != 5.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MONETARY][0])
	}
}

func TestApierV2itSetAction(t *testing.T) {
	if !*testIT {
		return
	}
	attrs := utils.AttrSetActions{ActionsId: "DISABLE_ACCOUNT", Actions: []*utils.TPAction{
		&utils.TPAction{Identifier: engine.DISABLE_ACCOUNT, Weight: 10.0},
	}}
	var reply string
	if err := apierRPC.Call("ApierV2.SetActions", attrs, &reply); err != nil {
		t.Error(err)
	}
	var acts map[string]engine.Actions
	if err := apierRPC.Call("ApierV2.GetActions", AttrGetActions{ActionIDs: []string{attrs.ActionsId}}, &acts); err != nil {
		t.Error(err)
	} else if len(acts) != 1 {
		t.Errorf("Received actions: %+v", acts)
	}
}

func TestApierV2itSetAccountActionTriggers(t *testing.T) {
	if !*testIT {
		return
	}
	attrs := AttrSetAccountActionTriggers{
		Tenant:         "cgrates.org",
		Account:        "dan",
		GroupID:        utils.StringPointer("MONITOR_MAX_BALANCE"),
		ThresholdType:  utils.StringPointer(utils.TRIGGER_MAX_BALANCE),
		ThresholdValue: utils.Float64Pointer(50),
		BalanceType:    utils.StringPointer(utils.MONETARY),
		ActionsID:      utils.StringPointer("DISABLE_ACCOUNT"),
	}
	var reply string
	if err := apierRPC.Call("ApierV2.SetAccountActionTriggers", attrs, &reply); err != nil {
		t.Error(err)
	}
	var ats engine.ActionTriggers
	if err := apierRPC.Call("ApierV2.GetAccountActionTriggers", v1.AttrAcntAction{Tenant: "cgrates.org", Account: "dan"}, &ats); err != nil {
		t.Error(err)
	} else if len(ats) != 1 || ats[0].ID != *attrs.GroupID || ats[0].ThresholdValue != 50.0 {
		t.Errorf("Received: %+v", ats)
	}
	attrs.ThresholdValue = utils.Float64Pointer(55) // Change the threshold
	if err := apierRPC.Call("ApierV2.SetAccountActionTriggers", attrs, &reply); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call("ApierV2.GetAccountActionTriggers", v1.AttrAcntAction{Tenant: "cgrates.org", Account: "dan"}, &ats); err != nil {
		t.Error(err)
	} else if len(ats) != 1 || ats[0].ID != *attrs.GroupID || ats[0].ThresholdValue != 55.0 {
		t.Errorf("Received: %+v", ats)
	}
}

func TestApierV2itFraudMitigation(t *testing.T) {
	if !*testIT {
		return
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "dan",
		BalanceType: utils.MONETARY,
		BalanceID:   utils.StringPointer(utils.META_DEFAULT),
		Value:       utils.Float64Pointer(60.0),
		Weight:      utils.Float64Pointer(10.0),
	}
	var reply string
	if err := apierRPC.Call("ApierV2.SetBalance", attrs, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt engine.Account
	if err := apierRPC.Call("ApierV2.GetAccount", &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MONETARY][0].Value != 60.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MONETARY][0])
	} else if !acnt.Disabled {
		t.Fatalf("Received account: %+v", acnt)
	}
	attrSetAcnt := AttrSetAccount{
		Tenant:   "cgrates.org",
		Account:  "dan",
		Disabled: utils.BoolPointer(false),
	}
	if err := apierRPC.Call("ApierV2.SetAccount", attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	if err := apierRPC.Call("ApierV2.GetAccount", &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MONETARY][0].Value != 60.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MONETARY][0])
	} else if acnt.Disabled {
		t.Fatalf("Received account: %+v", acnt)
	}
}

func TestApierV2itKillEngine(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.KillEngine(delay); err != nil {
		t.Error(err)
	}
}
