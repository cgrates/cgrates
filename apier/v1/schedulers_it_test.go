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
package v1

import (
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
)

var (
	schedCfgPath string
	schedCfg     *config.CGRConfig
	schedRpc     *rpc.Client
	schedConfDIR string //run tests for specific configuration
)

var sTestsSchedFiltered = []func(t *testing.T){
	testSchedLoadConfig,
	testSchedInitDataDb,
	testSchedResetStorDb,
	testSchedStartEngine,
	testSchedRpcConn,
	testSchedFromFolder,
	testSchedVeifyAllAccounts,
	testSchedVeifyAccount1001,
	testSchedVeifyAccount1002and1003,
	testSchedExecuteAction,
	testSchedStopEngine,
}

//TestSchedWithoutFilters will execute action for all accounts
func TestSchedWithoutFilters(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		schedConfDIR = "tutinternal"
	case utils.MetaMySQL:
		schedConfDIR = "tutmysql"
	case utils.MetaMongo:
		schedConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsSchedFiltered {
		t.Run(schedConfDIR, stest)
	}
}

//TestSchedWithFiltersSingleAccount will execute actions only for account 1001
func TestSchedWithFiltersSingleAccount(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		schedConfDIR = "filtered_scheduler_internal"
	case utils.MetaMySQL:
		schedConfDIR = "filtered_scheduler_mysql"
	case utils.MetaMongo:
		schedConfDIR = "filtered_scheduler_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsSchedFiltered {
		t.Run(schedConfDIR, stest)
	}
}

//TestSchedWithFilters2 will execute actions for accounts 1002 and 1003 ( 1001 will be skiped )
func TestSchedWithFilters2(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		schedConfDIR = "filtered_scheduler2_internal"
	case utils.MetaMySQL:
		schedConfDIR = "filtered_scheduler2_mysql"
	case utils.MetaMongo:
		schedConfDIR = "filtered_scheduler2_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsSchedFiltered {
		t.Run(schedConfDIR, stest)
	}
}

func testSchedLoadConfig(t *testing.T) {
	var err error
	schedCfgPath = path.Join(*dataDir, "conf", "samples", schedConfDIR)
	if schedCfg, err = config.NewCGRConfigFromPath(schedCfgPath); err != nil {
		t.Error(err)
	}
}

func testSchedInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(schedCfg); err != nil {
		t.Fatal(err)
	}
}

func testSchedResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(schedCfg); err != nil {
		t.Fatal(err)
	}
}

func testSchedStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(schedCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSchedRpcConn(t *testing.T) {
	var err error
	schedRpc, err = newRPCClient(schedCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testSchedFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := schedRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSchedVeifyAllAccounts(t *testing.T) {
	if !(schedConfDIR == "tutinternal" || schedConfDIR == "tutmysql" || schedConfDIR == "tutmongo") {
		t.SkipNow()
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}
	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}
	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1003",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}
}

func testSchedVeifyAccount1001(t *testing.T) {
	if !(schedConfDIR == "filtered_scheduler_internal" || schedConfDIR == "filtered_scheduler_mysql" || schedConfDIR == "filtered_scheduler_mongo") {
		t.SkipNow()
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}

	acnt = nil // in case of gob ( it doesn't update the empty fields)
	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if lenBal := len(acnt.BalanceMap[utils.MONETARY]); lenBal != 0 {
		t.Errorf("Expecting: %v, received: %v",
			0, lenBal)
	}

	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1003",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if lenBal := len(acnt.BalanceMap[utils.MONETARY]); lenBal != 0 {
		t.Errorf("Expecting: %v, received: %v",
			0, lenBal)
	}

}

func testSchedVeifyAccount1002and1003(t *testing.T) {
	if !(schedConfDIR == "filtered_scheduler2_internal" || schedConfDIR == "filtered_scheduler2_mysql" || schedConfDIR == "filtered_scheduler2_mongo") {
		t.SkipNow()
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if lenBal := len(acnt.BalanceMap[utils.MONETARY]); lenBal != 0 {
		t.Errorf("Expecting: %v, received: %v",
			0, lenBal)
	}

	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}

	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1003",
	}
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}
}

func testSchedExecuteAction(t *testing.T) {
	if !(schedConfDIR == "tutinternal" || schedConfDIR == "tutmysql" || schedConfDIR == "tutmongo") {
		t.SkipNow()
	}
	// set a new ActionPlan
	var reply1 string
	if err := schedRpc.Call(utils.APIerSv1SetActionPlan, &AttrSetActionPlan{
		Id: "CustomAP",
		ActionPlan: []*AttrActionPlan{
			{
				ActionsId: "ACT_TOPUP_RST_10",
				Time:      utils.MetaHourly,
				Weight:    20.0},
		},
	}, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply1)
	}
	var reply string
	if err := schedRpc.Call(utils.APIerSv1SetAccount, utils.AttrSetAccount{
		Tenant:       "cgrates.org",
		Account:      "CustomAccount",
		ActionPlanID: "CustomAP",
	}, &reply); err != nil {
		t.Fatal(err)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "CustomAccount",
	}
	expected := 0.0
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}

	if err := schedRpc.Call(utils.SchedulerSv1ExecuteActions, &utils.AttrsExecuteActions{ActionPlanID: "CustomAP"}, &reply); err != nil {
		t.Error(err)
	}
	expected = 10.0
	if err := schedRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testSchedStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
