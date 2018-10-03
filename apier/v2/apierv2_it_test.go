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
package v2

import (
	"flag"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	waitRater = flag.Int("wait_rater", 1500, "Number of miliseconds to wait for rater to start and cache")
)

var apierCfgPath string
var apierCfg *config.CGRConfig
var apierRPC *rpc.Client
var dm *engine.DataManager // share db connection here so we can check data we set through APIs

func TestApierV2itLoadConfig(t *testing.T) {
	apierCfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	if apierCfg, err = config.NewCGRConfigFromFolder(apierCfgPath); err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func TestApierV2itResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestApierV2itResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

func TestApierV2itConnectDataDB(t *testing.T) {
	rdsDb, _ := strconv.Atoi(apierCfg.DataDbName)
	if rdsITdb, err := engine.NewRedisStorage(fmt.Sprintf("%s:%s", apierCfg.DataDbHost, apierCfg.DataDbPort),
		rdsDb, apierCfg.DataDbPass, apierCfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, ""); err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	} else {
		dm = engine.NewDataManager(rdsITdb)
	}
}

// Start CGR Engine
func TestApierV2itStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(apierCfgPath, 200); err != nil { // Mongo requires more time to start
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestApierV2itRpcConn(t *testing.T) {
	apierRPC, err = jsonrpc.Dial("tcp", apierCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func TestApierV2itAddBalance(t *testing.T) {
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

func TestApierV2itSetAccountWithAP(t *testing.T) {
	argActs1 := utils.AttrSetActions{ActionsId: "TestApierV2itSetAccountWithAP_ACT_1",
		Actions: []*utils.TPAction{
			&utils.TPAction{Identifier: engine.TOPUP_RESET, BalanceType: utils.MONETARY, Directions: utils.OUT, Units: "5.0", Weight: 20.0},
		}}
	var reply string
	if err := apierRPC.Call("ApierV2.SetActions", argActs1, &reply); err != nil {
		t.Error(err)
	}
	argAP1 := &v1.AttrSetActionPlan{Id: "TestApierV2itSetAccountWithAP_AP_1",
		ActionPlan: []*v1.AttrActionPlan{
			&v1.AttrActionPlan{ActionsId: argActs1.ActionsId,
				Time:   time.Now().Add(time.Duration(time.Minute)).String(),
				Weight: 20.0}}}
	if _, err := dm.DataDB().GetActionPlan(argAP1.Id, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call("ApierV1.SetActionPlan", argAP1, &reply); err != nil {
		t.Error("Got error on ApierV1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.SetActionPlan received: %s", reply)
	}
	argSetAcnt1 := AttrSetAccount{
		Tenant:        "cgrates.org",
		Account:       "TestApierV2itSetAccountWithAP1",
		ActionPlanIDs: &[]string{argAP1.Id},
	}
	acntID := utils.AccountKey(argSetAcnt1.Tenant, argSetAcnt1.Account)
	if _, err := dm.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call("ApierV2.SetAccount", argSetAcnt1, &reply); err != nil {
		t.Fatal(err)
	}
	if ap, err := dm.DataDB().GetActionPlan(argAP1.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; !hasIt {
		t.Errorf("ActionPlan does not contain the accountID: %+v", ap)
	}
	eAAPids := []string{argAP1.Id}
	if aapIDs, err := dm.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAAPids, aapIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eAAPids, aapIDs)
	}
	// Set second AP so we can see the proper indexing done
	argAP2 := &v1.AttrSetActionPlan{Id: "TestApierV2itSetAccountWithAP_AP_2",
		ActionPlan: []*v1.AttrActionPlan{
			&v1.AttrActionPlan{ActionsId: argActs1.ActionsId, MonthDays: "1", Time: "00:00:00", Weight: 20.0}}}
	if _, err := dm.DataDB().GetActionPlan(argAP2.Id, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call("ApierV2.SetActionPlan", argAP2, &reply); err != nil {
		t.Error("Got error on ApierV2.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActionPlan received: %s", reply)
	}
	// Test adding new AP
	argSetAcnt2 := AttrSetAccount{
		Tenant:        "cgrates.org",
		Account:       "TestApierV2itSetAccountWithAP1",
		ActionPlanIDs: &[]string{argAP2.Id},
	}
	if err := apierRPC.Call("ApierV2.SetAccount", argSetAcnt2, &reply); err != nil {
		t.Fatal(err)
	}
	if ap, err := dm.DataDB().GetActionPlan(argAP2.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; !hasIt {
		t.Errorf("ActionPlan does not contain the accountID: %+v", ap)
	}
	if ap, err := dm.DataDB().GetActionPlan(argAP1.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; !hasIt {
		t.Errorf("ActionPlan does not contain the accountID: %+v", ap)
	}
	eAAPids = []string{argAP1.Id, argAP2.Id}
	if aapIDs, err := dm.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAAPids, aapIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eAAPids, aapIDs)
	}
	// test remove and overwrite
	argSetAcnt2 = AttrSetAccount{
		Tenant:               "cgrates.org",
		Account:              "TestApierV2itSetAccountWithAP1",
		ActionPlanIDs:        &[]string{argAP2.Id},
		ActionPlansOverwrite: true,
	}
	if err := apierRPC.Call("ApierV2.SetAccount", argSetAcnt2, &reply); err != nil {
		t.Fatal(err)
	}
	if ap, err := dm.DataDB().GetActionPlan(argAP1.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; hasIt {
		t.Errorf("ActionPlan does contain the accountID: %+v", ap)
	}
	if ap, err := dm.DataDB().GetActionPlan(argAP2.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; !hasIt {
		t.Errorf("ActionPlan does not contain the accountID: %+v", ap)
	}
	eAAPids = []string{argAP2.Id}
	if aapIDs, err := dm.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAAPids, aapIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eAAPids, aapIDs)
	}
}

func TestApierV2itKillEngine(t *testing.T) {
	if err := engine.KillEngine(delay); err != nil {
		t.Error(err)
	}
}
