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
	"fmt"
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	apierCfgPath   string
	apierCfg       *config.CGRConfig
	apierRPC       *rpc.Client
	dm             *engine.DataManager // share db connection here so we can check data we set through APIs
	apierv2ConfDIR string

	sTestsv2it = []func(t *testing.T){
		testApierV2itLoadConfig,
		testApierV2itResetDataDb,
		testApierV2itResetStorDb,
		testApierV2itConnectDataDB,
		testApierV2itStartEngine,
		testApierV2itRpcConn,
		testApierV2itAddBalance,
		testApierV2itSetAction,
		testApierV2itSetAccountActionTriggers,
		testApierV2itFraudMitigation,
		testApierV2itSetAccountWithAP,
		testApierV2itSetActionWithCategory,
		testApierV2itSetActionPlanWithWrongTiming,
		testApierV2itSetActionPlanWithWrongTiming2,
		testApierV2itKillEngine,
	}
)

func TestV2IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.Skip()
	case utils.MetaSQL:
		apierv2ConfDIR = "tutmysql"
	case utils.MetaMongo:
		apierv2ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsv2it {
		t.Run(apierv2ConfDIR, stest)
	}
}

func testApierV2itLoadConfig(t *testing.T) {
	apierCfgPath = path.Join(*dataDir, "conf", "samples", apierv2ConfDIR)
	if apierCfg, err = config.NewCGRConfigFromPath(apierCfgPath); err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testApierV2itResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testApierV2itResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

func testApierV2itConnectDataDB(t *testing.T) {
	rdsITdb, err := engine.NewDataDBConn(apierCfg.DataDbCfg().DataDbType,
		apierCfg.DataDbCfg().DataDbHost, apierCfg.DataDbCfg().DataDbPort,
		apierCfg.DataDbCfg().DataDbName, apierCfg.DataDbCfg().DataDbUser,
		apierCfg.DataDbCfg().DataDbPass, apierCfg.GeneralCfg().DBDataEncoding,
		apierCfg.DataDbCfg().DataDbSentinelName, apierCfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	dm = engine.NewDataManager(rdsITdb, config.CgrConfig().CacheCfg(), nil)
}

// Start CGR Engine
func testApierV2itStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(apierCfgPath, 200); err != nil { // Mongo requires more time to start
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testApierV2itRpcConn(t *testing.T) {
	apierRPC, err = newRPCClient(apierCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testApierV2itAddBalance(t *testing.T) {
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "dan",
		BalanceType: utils.MONETARY,
		Value:       5.0,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	var reply string
	if err := apierRPC.Call(utils.ApierV2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt engine.Account
	if err := apierRPC.Call(utils.ApierV2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY][0].Value != 5.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MONETARY][0])
	}
}

func testApierV2itSetAction(t *testing.T) {
	attrs := utils.AttrSetActions{ActionsId: "DISABLE_ACCOUNT", Actions: []*utils.TPAction{
		{Identifier: utils.DISABLE_ACCOUNT, Weight: 10.0},
	}}
	var reply string
	if err := apierRPC.Call(utils.ApierV2SetActions, attrs, &reply); err != nil {
		t.Error(err)
	}
	var acts map[string]engine.Actions
	if err := apierRPC.Call(utils.ApierV2GetActions, AttrGetActions{ActionIDs: []string{attrs.ActionsId}}, &acts); err != nil {
		t.Error(err)
	} else if len(acts) != 1 {
		t.Errorf("Received actions: %+v", acts)
	}
}

func testApierV2itSetAccountActionTriggers(t *testing.T) {
	attrs := v1.AttrSetAccountActionTriggers{
		Tenant:  "cgrates.org",
		Account: "dan",
		AttrSetActionTrigger: v1.AttrSetActionTrigger{
			GroupID: "MONITOR_MAX_BALANCE",
			ActionTrigger: map[string]interface{}{
				utils.ThresholdType:  utils.TRIGGER_MAX_BALANCE,
				utils.ThresholdValue: 50,
				utils.BalanceType:    utils.MONETARY,
				utils.ActionsID:      "DISABLE_ACCOUNT",
			},
		},
	}
	var reply string
	if err := apierRPC.Call(utils.ApierV2SetAccountActionTriggers, attrs, &reply); err != nil {
		t.Error(err)
	}
	var ats engine.ActionTriggers
	if err := apierRPC.Call(utils.ApierV2GetAccountActionTriggers, utils.TenantAccount{Tenant: "cgrates.org", Account: "dan"}, &ats); err != nil {
		t.Error(err)
	} else if len(ats) != 1 || ats[0].ID != attrs.GroupID || ats[0].ThresholdValue != 50.0 {
		t.Errorf("Received: %+v", ats)
	}
	attrs.ActionTrigger[utils.ThresholdValue] = 55 // Change the threshold
	if err := apierRPC.Call(utils.ApierV2SetAccountActionTriggers, attrs, &reply); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.ApierV2GetAccountActionTriggers, utils.TenantAccount{Tenant: "cgrates.org", Account: "dan"}, &ats); err != nil {
		t.Error(err)
	} else if len(ats) != 1 || ats[0].ID != attrs.GroupID || ats[0].ThresholdValue != 55.0 {
		t.Errorf("Received: %+v", ats)
	}
}

func testApierV2itFraudMitigation(t *testing.T) {
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "dan",
		BalanceType: utils.MONETARY,
		Value:       60.0,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	var reply string
	if err := apierRPC.Call(utils.ApierV2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt engine.Account
	if err := apierRPC.Call(utils.ApierV2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MONETARY][0].Value != 60.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MONETARY][0])
	} else if !acnt.Disabled {
		t.Fatalf("Received account: %+v", acnt)
	}
	attrSetAcnt := AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "dan",
		ExtraOptions: map[string]bool{
			utils.Disabled: false,
		},
	}
	if err := apierRPC.Call(utils.ApierV2SetAccount, attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	acnt = engine.Account{} // gob doesn't update the fields with default values
	if err := apierRPC.Call(utils.ApierV2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MONETARY][0].Value != 60.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MONETARY][0])
	} else if acnt.Disabled {
		t.Fatalf("Received account: %+v", acnt)
	}
}

func testApierV2itSetAccountWithAP(t *testing.T) {
	argActs1 := utils.AttrSetActions{ActionsId: "TestApierV2itSetAccountWithAP_ACT_1",
		Actions: []*utils.TPAction{
			{Identifier: utils.TOPUP_RESET,
				BalanceType: utils.MONETARY, Units: "5.0", Weight: 20.0},
		}}
	var reply string
	if err := apierRPC.Call(utils.ApierV2SetActions, argActs1, &reply); err != nil {
		t.Error(err)
	}
	tNow := time.Now().Add(time.Duration(time.Minute))
	argAP1 := &v1.AttrSetActionPlan{Id: "TestApierV2itSetAccountWithAP_AP_1",
		ActionPlan: []*v1.AttrActionPlan{
			{ActionsId: argActs1.ActionsId,
				Time:   fmt.Sprintf("%v:%v:%v", tNow.Hour(), tNow.Minute(), tNow.Second()), // 10:4:12
				Weight: 20.0}}}
	if _, err := dm.GetActionPlan(argAP1.Id, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.ApierV1SetActionPlan, argAP1, &reply); err != nil {
		t.Error("Got error on ApierV1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.SetActionPlan received: %s", reply)
	}
	argSetAcnt1 := AttrSetAccount{
		Tenant:        "cgrates.org",
		Account:       "TestApierV2itSetAccountWithAP1",
		ActionPlanIDs: []string{argAP1.Id},
	}
	acntID := utils.ConcatenatedKey(argSetAcnt1.Tenant, argSetAcnt1.Account)
	if _, err := dm.GetAccountActionPlans(acntID, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.ApierV2SetAccount, argSetAcnt1, &reply); err != nil {
		t.Fatal(err)
	}
	if ap, err := dm.GetActionPlan(argAP1.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; !hasIt {
		t.Errorf("ActionPlan does not contain the accountID: %+v", ap)
	}
	eAAPids := []string{argAP1.Id}
	if aapIDs, err := dm.GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAAPids, aapIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eAAPids, aapIDs)
	}
	// Set second AP so we can see the proper indexing done
	argAP2 := &v1.AttrSetActionPlan{Id: "TestApierV2itSetAccountWithAP_AP_2",
		ActionPlan: []*v1.AttrActionPlan{
			{ActionsId: argActs1.ActionsId, MonthDays: "1", Time: "00:00:00", Weight: 20.0}}}
	if _, err := dm.GetActionPlan(argAP2.Id, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.ApierV2SetActionPlan, argAP2, &reply); err != nil {
		t.Error("Got error on ApierV2.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActionPlan received: %s", reply)
	}
	// Test adding new AP
	argSetAcnt2 := AttrSetAccount{
		Tenant:        "cgrates.org",
		Account:       "TestApierV2itSetAccountWithAP1",
		ActionPlanIDs: []string{argAP2.Id},
	}
	if err := apierRPC.Call(utils.ApierV2SetAccount, argSetAcnt2, &reply); err != nil {
		t.Fatal(err)
	}
	if ap, err := dm.GetActionPlan(argAP2.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; !hasIt {
		t.Errorf("ActionPlan does not contain the accountID: %+v", ap)
	}
	if ap, err := dm.GetActionPlan(argAP1.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; !hasIt {
		t.Errorf("ActionPlan does not contain the accountID: %+v", ap)
	}
	eAAPids = []string{argAP1.Id, argAP2.Id}
	if aapIDs, err := dm.GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAAPids, aapIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eAAPids, aapIDs)
	}
	// test remove and overwrite
	argSetAcnt2 = AttrSetAccount{
		Tenant:               "cgrates.org",
		Account:              "TestApierV2itSetAccountWithAP1",
		ActionPlanIDs:        []string{argAP2.Id},
		ActionPlansOverwrite: true,
	}
	if err := apierRPC.Call(utils.ApierV2SetAccount, argSetAcnt2, &reply); err != nil {
		t.Fatal(err)
	}
	if ap, err := dm.GetActionPlan(argAP1.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; hasIt {
		t.Errorf("ActionPlan does contain the accountID: %+v", ap)
	}
	if ap, err := dm.GetActionPlan(argAP2.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if _, hasIt := ap.AccountIDs[acntID]; !hasIt {
		t.Errorf("ActionPlan does not contain the accountID: %+v", ap)
	}
	eAAPids = []string{argAP2.Id}
	if aapIDs, err := dm.GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAAPids, aapIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eAAPids, aapIDs)
	}
}

func testApierV2itSetActionWithCategory(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "TestApierV2itSetActionWithCategory"}
	if err := apierRPC.Call(utils.ApierV1SetAccount, attrsSetAccount, &reply); err != nil {
		t.Error("Got error on ApierV1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.SetAccount received: %s", reply)
	}

	argActs1 := utils.AttrSetActions{ActionsId: "TestApierV2itSetActionWithCategory_ACT",
		Actions: []*utils.TPAction{
			{Identifier: utils.TOPUP_RESET,
				BalanceType: utils.MONETARY, Categories: "test", Units: "5.0", Weight: 20.0},
		}}

	if err := apierRPC.Call(utils.ApierV2SetActions, argActs1, &reply); err != nil {
		t.Error(err)
	}

	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: argActs1.ActionsId}
	if err := apierRPC.Call(utils.ApierV1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}

	var acnt engine.Account
	if err := apierRPC.Call(utils.ApierV2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org",
		Account: "TestApierV2itSetActionWithCategory"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MONETARY][0].Value != 5.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MONETARY][0])
	} else if len(acnt.BalanceMap[utils.MONETARY][0].Categories) != 1 &&
		acnt.BalanceMap[utils.MONETARY][0].Categories["test"] != true {
		t.Fatalf("Unexpected category received: %+v", utils.ToJSON(acnt))
	}
}

func testApierV2itSetActionPlanWithWrongTiming(t *testing.T) {
	var reply string
	tNow := time.Now().Add(time.Duration(time.Minute)).String()
	argAP1 := &v1.AttrSetActionPlan{Id: "TestApierV2itSetAccountWithAPWithWrongTiming",
		ActionPlan: []*v1.AttrActionPlan{
			&v1.AttrActionPlan{
				ActionsId: "TestApierV2itSetAccountWithAP_ACT_1",
				Time:      tNow,
				Weight:    20.0,
			},
		},
	}

	if err := apierRPC.Call(utils.ApierV1SetActionPlan, argAP1, &reply); err == nil ||
		err.Error() != fmt.Sprintf("UNSUPPORTED_FORMAT:%s", tNow) {
		t.Error("Expecting error ", err)
	}
}

func testApierV2itSetActionPlanWithWrongTiming2(t *testing.T) {
	var reply string
	argAP1 := &v1.AttrSetActionPlan{Id: "TestApierV2itSetAccountWithAPWithWrongTiming",
		ActionPlan: []*v1.AttrActionPlan{
			&v1.AttrActionPlan{
				ActionsId: "TestApierV2itSetAccountWithAP_ACT_1",
				Time:      "aa:bb:cc",
				Weight:    20.0,
			},
		},
	}

	if err := apierRPC.Call(utils.ApierV1SetActionPlan, argAP1, &reply); err == nil ||
		err.Error() != fmt.Sprintf("UNSUPPORTED_FORMAT:aa:bb:cc") {
		t.Error("Expecting error ", err)
	}
}

func testApierV2itKillEngine(t *testing.T) {
	if err := engine.KillEngine(delay); err != nil {
		t.Error(err)
	}
}
