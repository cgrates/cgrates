//go:build integration
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
	cdrsOfflineCfgPath string
	cdrsOfflineCfg     *config.CGRConfig
	cdrsOfflineRpc     *rpc.Client
	cdrsOfflineConfDIR string // run the tests for specific configuration

	// subtests to be executed for each confDIR
	sTestsCDRsOfflineIT = []func(t *testing.T){
		testV2CDRsOfflineInitConfig,
		testV2CDRsOfflineInitDataDb,
		testV2CDRsOfflineInitCdrDb,
		testV2CDRsOfflineStartEngine,
		testV2cdrsOfflineRpcConn,
		testV2CDRsOfflineLoadData,
		testV2CDRsOfflineBalanceUpdate,
		testV2CDRsOfflineExpiryBalance,
		testV2CDRsBalancesWithSameWeight,
		testV2CDRsOfflineKillEngine,
	}
)

// Tests starting here
func TestCDRsOfflineIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		cdrsOfflineConfDIR = "cdrsv2internal"
	case utils.MetaMySQL:
		cdrsOfflineConfDIR = "cdrsv2mysql"
	case utils.MetaMongo:
		cdrsOfflineConfDIR = "cdrsv2mongo"
	case utils.MetaPostgres:
		cdrsOfflineConfDIR = "cdrsv2psql"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCDRsOfflineIT {
		t.Run(cdrsOfflineConfDIR, stest)
	}
}

func testV2CDRsOfflineInitConfig(t *testing.T) {
	var err error
	cdrsOfflineCfgPath = path.Join(*dataDir, "conf", "samples", cdrsOfflineConfDIR)
	if cdrsOfflineCfg, err = config.NewCGRConfigFromPath(cdrsOfflineCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV2CDRsOfflineInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdrsOfflineCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func testV2CDRsOfflineInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdrsOfflineCfg); err != nil {
		t.Fatal(err)
	}
}

func testV2CDRsOfflineStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsOfflineCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testV2cdrsOfflineRpcConn(t *testing.T) {
	cdrsOfflineRpc, err = newRPCClient(cdrsOfflineCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV2CDRsOfflineLoadData(t *testing.T) {
	var loadInst utils.LoadInstance
	if err := cdrsOfflineRpc.Call(utils.APIerSv2LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testV2CDRsOfflineBalanceUpdate(t *testing.T) {
	//add a test account with balance type monetary and value 10
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "test",
		BalanceType: utils.MetaMonetary,
		Value:       10.0,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	var reply string
	if err := cdrsOfflineRpc.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt *engine.Account
	if err := cdrsOfflineRpc.Call(utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "test"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MetaMonetary][0].Value != 10.0 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MetaMonetary][0])
	}

	var thReply *engine.ThresholdProfile
	var result string

	//create a log action
	attrsAA := &utils.AttrSetActions{ActionsId: "ACT_LOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaLog},
	}}
	if err := cdrsOfflineRpc.Call(utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	//make sure that the threshold don't exit
	if err := cdrsOfflineRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &thReply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//create a threshold that match out account
	tPrfl := engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_Test",
			FilterIDs: []string{"*string:Account:test"},
			MaxHits:   -1,
			MinSleep:  time.Second,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_LOG"},
			Async:     false,
		},
	}
	if err := cdrsOfflineRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := cdrsOfflineRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &thReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, thReply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, thReply)
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.OriginID:     "testV2CDRsOfflineProcessCDR2",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "testV2CDRsOfflineProcessCDR",
			utils.RequestType:  utils.MetaPostpaid,
			utils.Category:     "call",
			utils.AccountField: "test",
			utils.Subject:      "test",
			utils.Destination:  "1002",
			utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.Usage:        time.Minute,
		},
	}
	mapEv := engine.NewMapEvent(cgrEv.Event)
	cdr, err := mapEv.AsCDR(nil, "cgrates.org", "")
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	//process cdr should trigger balance update event
	if err := cdrsOfflineRpc.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithAPIOpts{CDR: cdr}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsOfflineExpiryBalance(t *testing.T) {
	var reply string
	acc := &utils.AttrSetActions{ActionsId: "ACT_TOPUP_TEST2", Actions: []*utils.TPAction{
		{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BalanceExpired1", Units: "5",
			ExpiryTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC).String(), BalanceWeight: "10", Weight: 20.0},
		{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BalanceExpired2", Units: "10",
			ExpiryTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC).String(), BalanceWeight: "10", Weight: 20.0},
		{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "NewBalance", Units: "10",
			ExpiryTime: utils.MetaUnlimited, BalanceWeight: "10", Weight: 20.0},
	}}
	if err := cdrsOfflineRpc.Call(utils.APIerSv2SetActions, acc, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	atm1 := &v1.AttrActionPlan{ActionsId: "ACT_TOPUP_TEST2", Time: "*asap", Weight: 20.0}
	atms1 := &v1.AttrSetActionPlan{Id: "AP_TEST2", ActionPlan: []*v1.AttrActionPlan{atm1}}
	if err := cdrsOfflineRpc.Call(utils.APIerSv1SetActionPlan, &atms1, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply)
	}

	if err := cdrsOfflineRpc.Call(utils.APIerSv2SetAccount,
		&AttrSetAccount{Tenant: "cgrates.org", Account: "test2",
			ActionPlanIDs: []string{"AP_TEST2"}, ReloadScheduler: true},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetAccount received: %s", reply)
	}

	var acnt *engine.Account
	//verify if the third balance was added
	if err := cdrsOfflineRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "test2"}, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].Len() != 1 {
		t.Errorf("Unexpected balance received: %+v", utils.ToIJSON(acnt))
	}

	var thReply *engine.ThresholdProfile
	var result string

	//create a log action
	attrsA := &utils.AttrSetActions{ActionsId: "ACT_LOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaLog},
	}}
	if err := cdrsOfflineRpc.Call(utils.APIerSv2SetActions, attrsA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	//make sure that the threshold don't exit
	if err := cdrsOfflineRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test2"}, &thReply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//create a threshold that match out account
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_Test2",
			FilterIDs: []string{"*string:Account:test2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  0,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_LOG"},
			Async:     false,
		},
	}
	if err := cdrsOfflineRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := cdrsOfflineRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test2"}, &thReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, thReply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, thReply)
	}

	args := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:     "testV2CDRsOfflineProcessCDR1",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testV2CDRsOfflineProcessCDR",
				utils.RequestType:  utils.MetaPostpaid,
				utils.Category:     "call",
				utils.AccountField: "test2",
				utils.Subject:      "test2",
				utils.Destination:  "1002",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
			},
		},
	}
	//process cdr should trigger balance update event
	if err := cdrsOfflineRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsBalancesWithSameWeight(t *testing.T) {
	//add a test account with balance type monetary and value 10
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "specialTest",
		BalanceType: utils.MetaMonetary,
		Value:       10.0,
		Balance: map[string]interface{}{
			utils.ID:     "SpecialBalance1",
			utils.Weight: 10.0,
		},
	}
	var reply string
	if err := cdrsOfflineRpc.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	attrs.Balance[utils.ID] = "SpecialBalance2"
	if err := cdrsOfflineRpc.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	var acnt *engine.Account
	if err := cdrsOfflineRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "specialTest"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MetaMonetary])
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.OriginID:     "testV2CDRsBalancesWithSameWeight",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "testV2CDRsBalancesWithSameWeight",
			utils.RequestType:  utils.MetaPostpaid,
			utils.Category:     "call",
			utils.AccountField: "specialTest",
			utils.Subject:      "specialTest",
			utils.Destination:  "1002",
			utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.Usage:        time.Minute,
		},
	}
	mapEv := engine.NewMapEvent(cgrEv.Event)
	cdr, err := mapEv.AsCDR(nil, "cgrates.org", "")
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	//process cdr should trigger balance update event
	if err := cdrsOfflineRpc.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithAPIOpts{CDR: cdr}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsOfflineKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
