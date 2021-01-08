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
	apierCfgPath    string
	apierCfg        *config.CGRConfig
	apierRPC        *rpc.Client
	dm              *engine.DataManager // share db connection here so we can check data we set through APIs
	APIerSv2ConfDIR string

	sTestsv2it = []func(t *testing.T){
		testAPIerSv2itLoadConfig,
		testAPIerSv2itResetDataDb,
		testAPIerSv2itResetStorDb,
		testAPIerSv2itConnectDataDB,
		testAPIerSv2itStartEngine,
		testAPIerSv2itRpcConn,
		testAPIerSv2itAddBalance,
		testAPIerSv2itSetAction,
		testAPIerSv2itSetAccountActionTriggers,
		testAPIerSv2itFraudMitigation,
		testAPIerSv2itSetAccountWithAP,
		testAPIerSv2itSetActionWithCategory,
		testAPIerSv2itSetActionPlanWithWrongTiming,
		testAPIerSv2itSetActionPlanWithWrongTiming2,
		testAPIerSv2itBackwardsCompatible,
		testAPIerSv2itGetAccountsCount,
		testAPIerSv2itGetActionsCount,
		testAPIerSv2itKillEngine,
	}
)

func TestV2IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.Skip()
	case utils.MetaMySQL:
		APIerSv2ConfDIR = "tutmysql"
	case utils.MetaMongo:
		APIerSv2ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsv2it {
		t.Run(APIerSv2ConfDIR, stest)
	}
}

func testAPIerSv2itLoadConfig(t *testing.T) {
	apierCfgPath = path.Join(*dataDir, "conf", "samples", APIerSv2ConfDIR)
	if apierCfg, err = config.NewCGRConfigFromPath(apierCfgPath); err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testAPIerSv2itResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAPIerSv2itResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

func testAPIerSv2itConnectDataDB(t *testing.T) {
	rdsITdb, err := engine.NewDataDBConn(apierCfg.DataDbCfg().DataDbType,
		apierCfg.DataDbCfg().DataDbHost, apierCfg.DataDbCfg().DataDbPort,
		apierCfg.DataDbCfg().DataDbName, apierCfg.DataDbCfg().DataDbUser,
		apierCfg.DataDbCfg().DataDbPass, apierCfg.GeneralCfg().DBDataEncoding,
		apierCfg.DataDbCfg().Opts)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	dm = engine.NewDataManager(rdsITdb, config.CgrConfig().CacheCfg(), nil)
}

// Start CGR Engine
func testAPIerSv2itStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(apierCfgPath, 200); err != nil { // Mongo requires more time to start
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAPIerSv2itRpcConn(t *testing.T) {
	apierRPC, err = newRPCClient(apierCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAPIerSv2itAddBalance(t *testing.T) {
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "dan",
		BalanceType: utils.MetaMonetary,
		Value:       5.0,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	var reply string
	if err := apierRPC.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt engine.Account
	if err := apierRPC.Call(utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary][0].Value != 5.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MetaMonetary][0])
	}
}

func testAPIerSv2itSetAction(t *testing.T) {
	attrs := utils.AttrSetActions{ActionsId: "DISABLE_ACCOUNT", Actions: []*utils.TPAction{
		{Identifier: utils.DISABLE_ACCOUNT, Weight: 10.0},
	}}
	var reply string
	if err := apierRPC.Call(utils.APIerSv2SetActions, &attrs, &reply); err != nil {
		t.Error(err)
	}
	var acts map[string]engine.Actions
	if err := apierRPC.Call(utils.APIerSv2GetActions, &AttrGetActions{ActionIDs: []string{attrs.ActionsId}}, &acts); err != nil {
		t.Error(err)
	} else if len(acts) != 1 {
		t.Errorf("Received actions: %+v", acts)
	}
}

func testAPIerSv2itSetAccountActionTriggers(t *testing.T) {
	attrs := v1.AttrSetAccountActionTriggers{
		Tenant:  "cgrates.org",
		Account: "dan",
		AttrSetActionTrigger: v1.AttrSetActionTrigger{
			GroupID: "MONITOR_MAX_BALANCE",
			ActionTrigger: map[string]interface{}{
				utils.ThresholdType:  utils.TriggerMaxBalance,
				utils.ThresholdValue: 50,
				utils.BalanceType:    utils.MetaMonetary,
				utils.ActionsID:      "DISABLE_ACCOUNT",
			},
		},
	}
	var reply string
	if err := apierRPC.Call(utils.APIerSv2SetAccountActionTriggers, attrs, &reply); err != nil {
		t.Error(err)
	}
	var ats engine.ActionTriggers
	if err := apierRPC.Call(utils.APIerSv2GetAccountActionTriggers, utils.TenantAccount{Tenant: "cgrates.org", Account: "dan"}, &ats); err != nil {
		t.Error(err)
	} else if len(ats) != 1 || ats[0].ID != attrs.GroupID || ats[0].ThresholdValue != 50.0 {
		t.Errorf("Received: %+v", ats)
	}
	attrs.ActionTrigger[utils.ThresholdValue] = 55 // Change the threshold
	if err := apierRPC.Call(utils.APIerSv2SetAccountActionTriggers, attrs, &reply); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetAccountActionTriggers, utils.TenantAccount{Tenant: "cgrates.org", Account: "dan"}, &ats); err != nil {
		t.Error(err)
	} else if len(ats) != 1 || ats[0].ID != attrs.GroupID || ats[0].ThresholdValue != 55.0 {
		t.Errorf("Received: %+v", ats)
	}
}

func testAPIerSv2itFraudMitigation(t *testing.T) {
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "dan",
		BalanceType: utils.MetaMonetary,
		Value:       60.0,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	var reply string
	if err := apierRPC.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt engine.Account
	if err := apierRPC.Call(utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MetaMonetary][0].Value != 60.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MetaMonetary][0])
	} else if !acnt.Disabled {
		t.Fatalf("Received account: %+v", acnt)
	}
	attrSetAcnt := &AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "dan",
		ExtraOptions: map[string]bool{
			utils.Disabled: false,
		},
	}
	if err := apierRPC.Call(utils.APIerSv2SetAccount, attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	acnt = engine.Account{} // gob doesn't update the fields with default values
	if err := apierRPC.Call(utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MetaMonetary][0].Value != 60.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MetaMonetary][0])
	} else if acnt.Disabled {
		t.Fatalf("Received account: %+v", acnt)
	}
}

func testAPIerSv2itSetAccountWithAP(t *testing.T) {
	argActs1 := utils.AttrSetActions{ActionsId: "TestAPIerSv2itSetAccountWithAP_ACT_1",
		Actions: []*utils.TPAction{
			{Identifier: utils.TOPUP_RESET,
				BalanceType: utils.MetaMonetary, Units: "5.0", Weight: 20.0},
		}}
	var reply string
	if err := apierRPC.Call(utils.APIerSv2SetActions, &argActs1, &reply); err != nil {
		t.Error(err)
	}
	tNow := time.Now().Add(time.Minute)
	argAP1 := &v1.AttrSetActionPlan{Id: "TestAPIerSv2itSetAccountWithAP_AP_1",
		ActionPlan: []*v1.AttrActionPlan{
			{ActionsId: argActs1.ActionsId,
				Time:   fmt.Sprintf("%v:%v:%v", tNow.Hour(), tNow.Minute(), tNow.Second()), // 10:4:12
				Weight: 20.0}}}
	if _, err := dm.GetActionPlan(argAP1.Id, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv1SetActionPlan, &argAP1, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply)
	}
	argSetAcnt1 := AttrSetAccount{
		Tenant:        "cgrates.org",
		Account:       "TestAPIerSv2itSetAccountWithAP1",
		ActionPlanIDs: []string{argAP1.Id},
	}
	acntID := utils.ConcatenatedKey(argSetAcnt1.Tenant, argSetAcnt1.Account)
	if _, err := dm.GetAccountActionPlans(acntID, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2SetAccount, &argSetAcnt1, &reply); err != nil {
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
	argAP2 := &v1.AttrSetActionPlan{Id: "TestAPIerSv2itSetAccountWithAP_AP_2",
		ActionPlan: []*v1.AttrActionPlan{
			{ActionsId: argActs1.ActionsId, MonthDays: "1", Time: "00:00:00", Weight: 20.0}}}
	if _, err := dm.GetActionPlan(argAP2.Id, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2SetActionPlan, argAP2, &reply); err != nil {
		t.Error("Got error on APIerSv2.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActionPlan received: %s", reply)
	}
	// Test adding new AP
	argSetAcnt2 := AttrSetAccount{
		Tenant:        "cgrates.org",
		Account:       "TestAPIerSv2itSetAccountWithAP1",
		ActionPlanIDs: []string{argAP2.Id},
	}
	if err := apierRPC.Call(utils.APIerSv2SetAccount, &argSetAcnt2, &reply); err != nil {
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
		Account:              "TestAPIerSv2itSetAccountWithAP1",
		ActionPlanIDs:        []string{argAP2.Id},
		ActionPlansOverwrite: true,
	}
	if err := apierRPC.Call(utils.APIerSv2SetAccount, &argSetAcnt2, &reply); err != nil {
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

func testAPIerSv2itSetActionWithCategory(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "TestAPIerSv2itSetActionWithCategory"}
	if err := apierRPC.Call(utils.APIerSv1SetAccount, attrsSetAccount, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}

	argActs1 := utils.AttrSetActions{ActionsId: "TestAPIerSv2itSetActionWithCategory_ACT",
		Actions: []*utils.TPAction{
			{Identifier: utils.TOPUP_RESET,
				BalanceType: utils.MetaMonetary, Categories: "test", Units: "5.0", Weight: 20.0},
		}}

	if err := apierRPC.Call(utils.APIerSv2SetActions, &argActs1, &reply); err != nil {
		t.Error(err)
	}

	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: argActs1.ActionsId}
	if err := apierRPC.Call(utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	var acnt engine.Account
	if err := apierRPC.Call(utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org",
		Account: "TestAPIerSv2itSetActionWithCategory"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MetaMonetary][0].Value != 5.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MetaMonetary][0])
	} else if len(acnt.BalanceMap[utils.MetaMonetary][0].Categories) != 1 &&
		acnt.BalanceMap[utils.MetaMonetary][0].Categories["test"] != true {
		t.Fatalf("Unexpected category received: %+v", utils.ToJSON(acnt))
	}
}

func testAPIerSv2itSetActionPlanWithWrongTiming(t *testing.T) {
	var reply string
	tNow := time.Now().Add(time.Minute).String()
	argAP1 := &v1.AttrSetActionPlan{Id: "TestAPIerSv2itSetAccountWithAPWithWrongTiming",
		ActionPlan: []*v1.AttrActionPlan{
			&v1.AttrActionPlan{
				ActionsId: "TestAPIerSv2itSetAccountWithAP_ACT_1",
				Time:      tNow,
				Weight:    20.0,
			},
		},
	}

	if err := apierRPC.Call(utils.APIerSv1SetActionPlan, &argAP1, &reply); err == nil ||
		err.Error() != fmt.Sprintf("UNSUPPORTED_FORMAT:%s", tNow) {
		t.Error("Expecting error ", err)
	}
}

func testAPIerSv2itSetActionPlanWithWrongTiming2(t *testing.T) {
	var reply string
	argAP1 := &v1.AttrSetActionPlan{Id: "TestAPIerSv2itSetAccountWithAPWithWrongTiming",
		ActionPlan: []*v1.AttrActionPlan{
			&v1.AttrActionPlan{
				ActionsId: "TestAPIerSv2itSetAccountWithAP_ACT_1",
				Time:      "aa:bb:cc",
				Weight:    20.0,
			},
		},
	}

	if err := apierRPC.Call(utils.APIerSv1SetActionPlan, &argAP1, &reply); err == nil ||
		err.Error() != fmt.Sprintf("UNSUPPORTED_FORMAT:aa:bb:cc") {
		t.Error("Expecting error ", err)
	}
}

func testAPIerSv2itBackwardsCompatible(t *testing.T) {
	var reply string
	if err := apierRPC.Call("ApierV2.Ping", new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Expecting : %+v, received: %+v", utils.Pong, reply)
	}
}

func testAPIerSv2itGetAccountsCount(t *testing.T) {
	var reply1 int
	if err := apierRPC.Call(utils.APIerSv2GetAccountsCount, &utils.AttrGetAccountsCount{
		Tenant: "cgrates.org"}, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != 3 {
		t.Errorf("Expecting: 3, received: %+v", reply1)
	}
	var reply string
	if err := apierRPC.Call(utils.APIerSv1RemoveAccount, &utils.AttrRemoveAccount{
		Account: "dan", Tenant: "cgrates.org"}, &reply); err != nil {
		t.Errorf("Unexpected error : %+v", err)
	}
	if err := apierRPC.Call(utils.APIerSv1RemoveAccount, &utils.AttrRemoveAccount{
		Account: "TestAPIerSv2itSetAccountWithAP1", Tenant: "cgrates.org"}, &reply); err != nil {
		t.Errorf("Unexpected error : %+v", err)
	}
	if err := apierRPC.Call(utils.APIerSv1RemoveAccount, &utils.AttrRemoveAccount{
		Account: "TestAPIerSv2itSetActionWithCategory", Tenant: "cgrates.org"}, &reply); err != nil {
		t.Errorf("Unexpected error : %+v", err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetAccountsCount, &utils.AttrGetAccountsCount{
		Tenant: "cgrates.org"}, &reply1); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting %+v, received: %+v", utils.ErrNotFound, err)
	}
	argSetAccount := AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "TestAPIerSv2CountAccounts",
	}
	if err := apierRPC.Call(utils.APIerSv2SetAccount, &argSetAccount, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt engine.Account
	if err := apierRPC.Call(utils.APIerSv2GetAccount, &utils.AttrGetAccount{
		Tenant: "cgrates.org", Account: "TestAPIerSv2CountAccounts"}, &acnt); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetAccountsCount, &utils.AttrGetAccountsCount{Tenant: "cgrates.org"}, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != 1 {
		t.Errorf("Expecting: 1, received: %+v", reply1)
	}
	argSetAccount = AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "TestAPIerSv2CountAccounts2",
	}
	if err := apierRPC.Call(utils.APIerSv2SetAccount, &argSetAccount, &reply); err != nil {
		t.Fatal(err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetAccount, &utils.AttrGetAccount{
		Tenant: "cgrates.org", Account: "TestAPIerSv2CountAccounts2"}, &acnt); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetAccountsCount, &utils.AttrGetAccountsCount{Tenant: "cgrates.org"}, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != 2 {
		t.Errorf("Expecting: 2, received: %+v", reply1)
	}
	if err := apierRPC.Call(utils.APIerSv1RemoveAccount, &utils.AttrRemoveAccount{
		Account: "TestAPIerSv2CountAccounts2", Tenant: "cgrates.org"}, &reply); err != nil {
		t.Errorf("Unexpected error : %+v", err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetAccountsCount, &utils.AttrGetAccountsCount{Tenant: "cgrates.org"}, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != 1 {
		t.Errorf("Expecting: 1, received: %+v", reply1)
	}
	if err := apierRPC.Call(utils.APIerSv1RemoveAccount, &utils.AttrRemoveAccount{
		Account: "TestAPIerSv2CountAccounts", Tenant: "cgrates.org"}, &reply); err != nil {
		t.Errorf("Unexpected error : %+v", err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetAccountsCount, &utils.AttrGetAccountsCount{
		Tenant: "cgrates.org"}, &reply1); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting %+v, received: %+v", utils.ErrNotFound, err)
	}
}

func testAPIerSv2itGetActionsCount(t *testing.T) {
	var reply1 int
	if err := apierRPC.Call(utils.APIerSv2GetActionsCount, &AttrGetActionsCount{}, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != 3 {
		t.Errorf("Expecting: 3, received : %+v", reply1)
	}
	attrs := utils.AttrSetActions{ActionsId: "DISABLE_ACCOUNT2", Actions: []*utils.TPAction{
		{Identifier: utils.DISABLE_ACCOUNT, Weight: 0.7},
	}}
	var reply string
	if err := apierRPC.Call(utils.APIerSv2SetActions, &attrs, &reply); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetActionsCount, &AttrGetActionsCount{}, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != 4 {
		t.Errorf("Expecting: 4, received : %+v", reply1)
	}

	attrRemoveActions := &v1.AttrRemoveActions{
		ActionIDs: []string{"DISABLE_ACCOUNT", "DISABLE_ACCOUNT2", "TestAPIerSv2itSetAccountWithAP_ACT_1"},
	}
	if err := apierRPC.Call(utils.APIerSv2RemoveActions, &attrRemoveActions, &reply); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetActionsCount, &AttrGetActionsCount{}, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != 1 {
		t.Errorf("Expecting: 1, received : %+v", reply1)
	}
	attrRemoveActions = &v1.AttrRemoveActions{
		ActionIDs: []string{"TestAPIerSv2itSetActionWithCategory_ACT"},
	}
	if err := apierRPC.Call(utils.APIerSv2RemoveActions, &attrRemoveActions, &reply); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetActionsCount, &AttrGetActionsCount{}, &reply1); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting %+v, received: %+v", utils.ErrNotFound, err)
	}
	attrs = utils.AttrSetActions{ActionsId: "Test", Actions: []*utils.TPAction{
		{Identifier: utils.DISABLE_ACCOUNT, Weight: 0.7},
	}}
	if err := apierRPC.Call(utils.APIerSv2SetActions, &attrs, &reply); err != nil {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.APIerSv2GetActionsCount, &AttrGetActionsCount{}, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != 1 {
		t.Errorf("Expecting: 1, received : %+v", reply1)
	}
}

func testAPIerSv2itKillEngine(t *testing.T) {
	if err := engine.KillEngine(delay); err != nil {
		t.Error(err)
	}
}
